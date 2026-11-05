package authorization_roles_settings

import (
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/resourcedata"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v200/platformclientv2"
)

/*
The resource_genesyscloud_authorization_roles_settings_utils.go file contains various helper methods to marshal
and unmarshal data into formats consumable by Terraform and/or Genesys Cloud.
*/

// buildAuthorizationRolesSettingsFromResourceData produces the PUT request body by merging the
// practitioner's configuration over the organization's current settings.
//
// current is the settings object just read back from the API, and may be nil if that read
// failed; in that case only explicitly configured attributes are sent.
func buildAuthorizationRolesSettingsFromResourceData(d *schema.ResourceData, current *roleSettings) roleSettings {
	var rawConfig cty.Value
	if d != nil {
		rawConfig = d.GetRawConfig()
	}
	return mergeAuthorizationRolesSettings(rawConfig, current)
}

// mergeAuthorizationRolesSettings overlays the configured attributes in rawConfig onto current
// and returns the result.
//
// This read-modify-write is deliberate. The endpoint only supports PUT, which replaces the
// whole object, so a config that manages a single attribute would otherwise send the zero
// value for the others. Seeding the body from the settings just read back keeps the
// unconfigured attributes at their current values.
//
// genesys_org_policy_bypass is the attribute that makes this matter. Genesys Cloud enables it
// by default, and clearing it stops paired Genesys Services organizations from bypassing ABAC
// policy enforcement, which can cut off Genesys support access. A practitioner who manages only
// backfill_enabled must not have that happen as a side effect.
//
// It takes the raw cty config rather than a *schema.ResourceData so that the merge can be
// exercised directly in unit tests; schema.TestResourceDataRaw does not populate a raw config.
func mergeAuthorizationRolesSettings(rawConfig cty.Value, current *roleSettings) roleSettings {
	settings := roleSettings{}

	if current != nil {
		settings.BackfillEnabled = current.BackfillEnabled
		settings.AuthorizationGrantDivisionAware = current.AuthorizationGrantDivisionAware
		settings.GenesysOrgPolicyBypass = current.GenesysOrgPolicyBypass
	}

	if value, ok := configuredBool(rawConfig, "backfill_enabled"); ok {
		settings.BackfillEnabled = platformclientv2.Bool(value)
	}
	if value, ok := configuredBool(rawConfig, "authorization_grant_division_aware"); ok {
		settings.AuthorizationGrantDivisionAware = platformclientv2.Bool(value)
	}
	if value, ok := configuredBool(rawConfig, "genesys_org_policy_bypass"); ok {
		settings.GenesysOrgPolicyBypass = platformclientv2.Bool(value)
	}

	return settings
}

// configuredBool reports a boolean attribute as it appears in the practitioner's
// configuration. ok is false when the attribute is absent or null there, which for these
// Optional+Computed attributes means "leave whatever the organization already has".
//
// The raw config is consulted rather than d.Get or d.GetOkExists because neither can tell an
// explicit `false` apart from an unset attribute: d.Get returns the zero value for both, and
// GetOkExists reports ok=false for a zero value. Distinguishing the two matters on create,
// where there is no prior state for d.Get to fall back on and every unconfigured bool would
// otherwise look like an explicit false - which for genesys_org_policy_bypass would disable
// Genesys support access on the first apply.
//
// Null-safe for operations that run without a configuration, such as import.
func configuredBool(rawConfig cty.Value, key string) (bool, bool) {
	if rawConfig == cty.NilVal || !rawConfig.IsKnown() || rawConfig.IsNull() || !rawConfig.CanIterateElements() {
		return false, false
	}
	if t := rawConfig.Type(); !t.IsObjectType() || !t.HasAttribute(key) {
		return false, false
	}
	attr := rawConfig.GetAttr(key)
	if !attr.IsKnown() || attr.IsNull() {
		return false, false
	}
	return attr.True(), true
}

// setAuthorizationRolesSettingsToResourceData maps data from a roleSettings object into a schema
// ResourceData object. Shared by both the resource read and the data source read.
func setAuthorizationRolesSettingsToResourceData(d *schema.ResourceData, settings *roleSettings) {
	if settings == nil {
		return
	}
	resourcedata.SetNillableValue(d, "backfill_enabled", settings.BackfillEnabled)
	resourcedata.SetNillableValue(d, "authorization_grant_division_aware", settings.AuthorizationGrantDivisionAware)
	resourcedata.SetNillableValue(d, "genesys_org_policy_bypass", settings.GenesysOrgPolicyBypass)
}
