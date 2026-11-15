package authorization_roles_settings

import (
	"context"
	"fmt"
	"log"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/consistency_checker"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/constants"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v200/platformclientv2"
)

/*
The resource_genesyscloud_authorization_roles_settings.go contains all the methods that perform the core logic for the resource.
*/

// authorizationRolesSettingsId is the fixed ID for this singleton resource
const authorizationRolesSettingsId = "settings"

func getAllAuthorizationRolesSettings(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	// Although this resource has only a single instance, we attempt to fetch the data
	// from the API in order to verify the user's permission to access the endpoint.
	proxy := getAuthorizationRolesSettingsProxy(clientConfig)
	resources := make(resourceExporter.ResourceIDMetaMap)

	_, resp, err := proxy.getAuthorizationRolesSettings(ctx)
	if err != nil {
		if util.IsStatus404(resp) {
			// Don't export if config doesn't exist
			return resources, nil
		}
		return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get %s resource due to error: %s", ResourceType, err), resp)
	}

	resources[authorizationRolesSettingsId] = &resourceExporter.ResourceMeta{BlockLabel: ResourceType}
	return resources, nil
}

// createAuthorizationRolesSettings is used by the resource to create the settings.
//
// The API has no POST for this endpoint: the settings object always exists on an organization.
// Terraform still requires a create hook, so it assigns the singleton's fixed ID and delegates
// to update. No other HTTP verb is involved.
//
// Delegating is deliberate rather than merely convenient. A first apply has to do exactly what
// any other apply does - GET the organization's current settings, overlay the configured
// attributes, PUT the result - and that read-modify-write is what stops an initial apply from
// clearing genesys_org_policy_bypass. Create is in fact the riskiest moment for it: there is no
// prior state, so every attribute the practitioner left out reads as a zero value. Building the
// request body here instead would mean duplicating that logic in the one place it matters most.
//
// Note this calls updateAuthorizationRolesSettings directly, not the UpdateWithPooledClient
// wrapper. The create wrapper has already borrowed a client from the pool; going through the
// wrapper again would borrow a second one for the same operation.
func createAuthorizationRolesSettings(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Printf("Creating authorization role settings %s", authorizationRolesSettingsId)
	d.SetId(authorizationRolesSettingsId)
	return updateAuthorizationRolesSettings(ctx, d, meta)
}

// readAuthorizationRolesSettings is used by the resource to read the settings from Genesys Cloud
func readAuthorizationRolesSettings(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getAuthorizationRolesSettingsProxy(sdkConfig)
	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceAuthorizationRolesSettings(), constants.ConsistencyChecks(), ResourceType)

	log.Printf("Reading authorization role settings %s", d.Id())

	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		settings, resp, getErr := proxy.getAuthorizationRolesSettings(ctx)
		if getErr != nil {
			if util.IsStatus404(resp) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read authorization role settings %s | error: %s", d.Id(), getErr), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read authorization role settings %s | error: %s", d.Id(), getErr), resp))
		}

		setAuthorizationRolesSettingsToResourceData(d, settings)

		log.Printf("Read authorization role settings %s", d.Id())
		return cc.CheckState(d)
	})
}

// updateAuthorizationRolesSettings is used by the resource to update the settings in Genesys Cloud.
//
// The settings are read before they are written because the API exposes only PUT, which
// replaces the whole object. Attributes the practitioner has not configured are carried over
// from this read rather than being sent as their zero value. The read is mandatory: if it
// fails we stop instead of writing a partial object, since doing so could disable
// genesys_org_policy_bypass and with it Genesys support access.
func updateAuthorizationRolesSettings(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getAuthorizationRolesSettingsProxy(sdkConfig)

	current, resp, err := proxy.getAuthorizationRolesSettings(ctx)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to read current authorization role settings before update: %s", err), resp)
	}

	settings := buildAuthorizationRolesSettingsFromResourceData(d, current)

	log.Printf("Updating authorization role settings %s", d.Id())

	_, resp, err = proxy.updateAuthorizationRolesSettings(ctx, &settings)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to update authorization role settings: %s", err), resp)
	}

	log.Printf("Updated authorization role settings %s", d.Id())
	return readAuthorizationRolesSettings(ctx, d, meta)
}

// deleteAuthorizationRolesSettings is a no-op.
//
// The API has no DELETE for this endpoint and the settings object cannot be removed from an
// organization. Terraform requires a delete hook regardless (schema.Resource.InternalValidate
// rejects a writable resource without one), so this drops the resource from state and leaves
// the organization's settings exactly as they are.
func deleteAuthorizationRolesSettings(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Printf("Removing authorization role settings %s from state. The organization's settings are left unchanged; this endpoint supports only GET and PUT.", d.Id())
	return nil
}
