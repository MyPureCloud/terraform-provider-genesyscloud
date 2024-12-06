package organization_authentication_settings

import (
	"context"
	"fmt"
	"log"
	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"terraform-provider-genesyscloud/genesyscloud/util/constants"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v146/platformclientv2"

	"terraform-provider-genesyscloud/genesyscloud/util/resourcedata"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

/*
The resource_genesyscloud_organization_authentication_settings.go contains all the methods that perform the core logic for a resource.
*/

func getAllOrganizationAuthenticationSettings(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	// Although this resource typically has only a single instance,
	// we are attempting to fetch the data from the API in order to
	// verify the user's permission to access this resource's API endpoint(s).

	proxy := getOrgAuthSettingsProxy(clientConfig)
	resources := make(resourceExporter.ResourceIDMetaMap)

	_, resp, err := proxy.getOrgAuthSettings(ctx)
	if err != nil {
		if util.IsStatus404(resp) {
			// Don't export if config doesn't exist
			return resources, nil
		}
		return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get %s resource due to error: %s", ResourceType, err), resp)
	}
	resources["0"] = &resourceExporter.ResourceMeta{BlockLabel: ResourceType}
	return resources, nil
}

// createOrganizationAuthenticationSettings is used by the organization_authentication_settings resource to create Genesys cloud organization authentication settings
func createOrganizationAuthenticationSettings(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Printf("Creating Organization Authentication Settings")
	d.SetId("Settings")
	return updateOrganizationAuthenticationSettings(ctx, d, meta)
}

// readOrganizationAuthenticationSettings is used by the organization_authentication_settings resource to read an organization authentication settings from genesys cloud
func readOrganizationAuthenticationSettings(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getOrgAuthSettingsProxy(sdkConfig)
	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceOrganizationAuthenticationSettings(), constants.ConsistencyChecks(), ResourceType)

	log.Printf("Reading organization authentication settings %s", d.Id())

	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		orgAuthSettings, resp, getErr := proxy.getOrgAuthSettings(ctx)
		if getErr != nil {
			if util.IsStatus404(resp) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read organization authentication settings %s | error: %s", d.Id(), getErr), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read organization authentication settings %s | error: %s", d.Id(), getErr), resp))
		}

		resourcedata.SetNillableValue(d, "multifactor_authentication_required", orgAuthSettings.MultifactorAuthenticationRequired)
		resourcedata.SetNillableValue(d, "domain_allowlist_enabled", orgAuthSettings.DomainAllowlistEnabled)
		resourcedata.SetNillableValue(d, "domain_allowlist", orgAuthSettings.DomainAllowlist)
		resourcedata.SetNillableValue(d, "ip_address_allowlist", orgAuthSettings.IpAddressAllowlist)
		resourcedata.SetNillableValueWithInterfaceArrayWithFunc(d, "password_requirements", orgAuthSettings.PasswordRequirements, flattenPasswordRequirements)

		log.Printf("Read organization authentication settings %s", d.Id())
		return cc.CheckState(d)
	})
}

// updateOrganizationAuthenticationSettings is used by the organization_authentication_settings resource to update an organization authentication settings in Genesys Cloud
func updateOrganizationAuthenticationSettings(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getOrgAuthSettingsProxy(sdkConfig)
	authSettings := getOrganizationAuthenticationSettingsFromResourceData(d)

	log.Printf("Updating organization authentication settings %s", d.Id())

	orgAuthSettings, resp, err := proxy.updateOrgAuthSettings(ctx, &authSettings)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to update organization authentication settings: %s", err), resp)
	}

	log.Printf("Updated organization authentication settings %s %s", d.Id(), orgAuthSettings)
	return readOrganizationAuthenticationSettings(ctx, d, meta)
}

// deleteOrganizationAuthenticationSettings is used by the organization_authentication_settings resource to delete an organization authentication settings from Genesys cloud
func deleteOrganizationAuthenticationSettings(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return nil
}
