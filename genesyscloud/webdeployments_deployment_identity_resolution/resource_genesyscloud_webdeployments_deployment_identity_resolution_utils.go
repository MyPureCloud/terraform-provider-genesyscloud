package webdeployments_deployment_identity_resolution

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v195/platformclientv2"
)

func buildDeploymentIdentityResolutionConfig(d *schema.ResourceData) (platformclientv2.Deploymentidentityresolutionconfig, error) {
	automergeConfig, err := buildAutomergeConfig(d.Get("automerge_config").([]interface{}))
	if err != nil {
		return platformclientv2.Deploymentidentityresolutionconfig{}, err
	}

	config := platformclientv2.Deploymentidentityresolutionconfig{
		ResolveIdentities: platformclientv2.Bool(d.Get("resolve_identities").(bool)),
		Division:          nil,
		ExternalSource:    nil,
		Automerge:         automergeConfig,
	}

	if divisionId, ok := d.Get("division_id").(string); ok && !isUnassignedDivisionId(divisionId) {
		config.Division = &platformclientv2.Writablestarrabledivision{
			Id: platformclientv2.String(divisionId),
		}
	}

	if externalSourceId, ok := d.Get("external_source_id").(string); ok && externalSourceId != "" {
		config.ExternalSource = &platformclientv2.Identityresolutionexternalsource{
			Id: platformclientv2.String(externalSourceId),
		}
	}

	return config, nil
}

func buildAutomergeConfig(automergeConfigBlock []interface{}) (*platformclientv2.Identityresolutionautomergeconfig, error) {
	result := buildDefaultAutomergeConfig()

	if len(automergeConfigBlock) > 0 {
		automergeConfig, ok := automergeConfigBlock[0].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("automerge_config must be a block")
		}

		if authenticatedWebMessaging, ok := automergeConfig["authenticated_web_messaging"].(bool); ok {
			result.AuthenticatedWebMessaging = platformclientv2.Bool(authenticatedWebMessaging)
		}

		if webTracking, ok := automergeConfig["web_tracking"].(bool); ok {
			result.WebTracking = platformclientv2.Bool(webTracking)
		}
	}

	return result, nil
}

func buildDefaultDeploymentIdentityResolutionConfig() platformclientv2.Deploymentidentityresolutionconfig {
	return platformclientv2.Deploymentidentityresolutionconfig{
		ResolveIdentities: platformclientv2.Bool(true),
		Division:          nil,
		ExternalSource:    nil,
		Automerge:         buildDefaultAutomergeConfig(),
	}
}

func buildDefaultAutomergeConfig() *platformclientv2.Identityresolutionautomergeconfig {
	return &platformclientv2.Identityresolutionautomergeconfig{
		AuthenticatedWebMessaging: platformclientv2.Bool(false),
		WebTracking:               platformclientv2.Bool(false),
	}
}

func flattenAutomergeConfig(automergeConfig *platformclientv2.Identityresolutionautomergeconfig, declared bool) []map[string]interface{} {
	if !declared && isDefaultAutomergeConfig(automergeConfig) {
		return nil
	}

	result := map[string]interface{}{
		"authenticated_web_messaging": false,
		"web_tracking":                false,
	}

	if automergeConfig != nil {
		if automergeConfig.AuthenticatedWebMessaging != nil {
			result["authenticated_web_messaging"] = *automergeConfig.AuthenticatedWebMessaging
		}
		if automergeConfig.WebTracking != nil {
			result["web_tracking"] = *automergeConfig.WebTracking
		}
	}

	return []map[string]interface{}{result}
}

// authenticatedWebMessagingDowngraded reports whether the platform silently refused a
// requested authenticated_web_messaging=true.
//
// contacts-service normalizes instead of rejecting (IdentityResolutionService): the flag
// is AND-ed with allowsAutomergingAuthenticatedWebMessaging(), which SquonkClient derives
// from the deployment's resolved configVersion.auth.enabled -- i.e. authentication_settings
// with enabled = true on the genesyscloud_webdeployments_configuration the deployment
// points at. That happens on a 200 response, so without this check the refusal surfaces
// only as a plan that can never converge.
//
// This is the only silently-normalized field on this channel: SquonkClient allows
// externalSourceId and webTracking unconditionally, and an unknown externalSourceId is a
// hard ServiceValidationException rather than a silent clear.
func authenticatedWebMessagingDowngraded(requested, stored *platformclientv2.Deploymentidentityresolutionconfig) bool {
	if stored == nil {
		// No response to compare against, so nothing was observably refused.
		return false
	}

	return authenticatedWebMessagingEnabled(requested) && !authenticatedWebMessagingEnabled(stored)
}

// authenticatedWebMessagingEnabled reports whether a config switches on automerging of
// authenticated web messaging. A missing automerge object counts as off, which is how the
// platform reports a channel that allows neither automerge flag.
func authenticatedWebMessagingEnabled(config *platformclientv2.Deploymentidentityresolutionconfig) bool {
	if config == nil || config.Automerge == nil || config.Automerge.AuthenticatedWebMessaging == nil {
		return false
	}

	return *config.Automerge.AuthenticatedWebMessaging
}

func isDefaultIdentityResolutionConfig(config *platformclientv2.Deploymentidentityresolutionconfig) bool {
	if config == nil {
		return true
	}

	if config.ResolveIdentities != nil && !*config.ResolveIdentities {
		return false
	}

	if config.Division != nil && config.Division.Id != nil && !isUnassignedDivisionId(*config.Division.Id) {
		return false
	}

	if config.ExternalSource != nil && config.ExternalSource.Id != nil {
		return false
	}

	return isDefaultAutomergeConfig(config.Automerge)
}

func isDefaultAutomergeConfig(config *platformclientv2.Identityresolutionautomergeconfig) bool {
	if config == nil {
		return true
	}
	if config.AuthenticatedWebMessaging != nil && *config.AuthenticatedWebMessaging {
		return false
	}
	if config.WebTracking != nil && *config.WebTracking {
		return false
	}
	return true
}

// isUnassignedDivisionId reports whether division_id is the unassigned (STAR) division
// sentinel (omitted in config/state, or explicit "*"). In contacts-service, "*" serializes
// a null/STAR division id — not "all divisions" and not "parent resource's division".
func isUnassignedDivisionId(divisionId string) bool {
	return divisionId == "" || divisionId == "*"
}

// suppressUnassignedDivisionIdDiff treats omitted division_id and "*" as equivalent
// (both mean the unassigned / STAR division). Read omits "*" from state, so explicit
// config "*" would otherwise show a perpetual plan diff.
func suppressUnassignedDivisionIdDiff(_, old, new string, _ *schema.ResourceData) bool {
	return isUnassignedDivisionId(old) && isUnassignedDivisionId(new)
}
