package organization_authentication_settings

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v121/platformclientv2"
	"log"
	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"

	gcloud "terraform-provider-genesyscloud/genesyscloud"

	"terraform-provider-genesyscloud/genesyscloud/util/resourcedata"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

/*
The resource_genesyscloud_organization_authentication_settings.go contains all of the methods that perform the core logic for a resource.
*/

func getAllOrganizationAuthenticationSettings(_ context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(resourceExporter.ResourceIDMetaMap)
	resources["0"] = &resourceExporter.ResourceMeta{Name: "organization_authentication_settings"}
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
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	proxy := getOrgAuthSettingsProxy(sdkConfig)

	log.Printf("Reading organization authentication settings %s", d.Id())

	return gcloud.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		orgAuthSettings, resp, getErr := proxy.getOrgAuthSettingsById(ctx, d.Id())
		if getErr != nil {
			if gcloud.IsStatus404(resp) {
				return retry.RetryableError(fmt.Errorf("Failed to read organization authentication settings %s: %s", d.Id(), getErr))
			}
			return retry.NonRetryableError(fmt.Errorf("Failed to read organization authentication settings %s: %s", d.Id(), getErr))
		}

		cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceOrganizationAuthenticationSettings())

		resourcedata.SetNillableValue(d, "multifactor_authentication_required", orgAuthSettings.MultifactorAuthenticationRequired)
		resourcedata.SetNillableValue(d, "domain_allowlist_enabled", orgAuthSettings.DomainAllowlistEnabled)
		resourcedata.SetNillableValue(d, "domain_allowlist", orgAuthSettings.DomainAllowlist)
		resourcedata.SetNillableValue(d, "ip_address_allowlist", orgAuthSettings.IpAddressAllowlist)
		resourcedata.SetNillableValueWithInterfaceArrayWithFunc(d, "password_requirements", orgAuthSettings.PasswordRequirements, flattenPasswordRequirements)

		log.Printf("Read organization authentication settings %s", d.Id())
		return cc.CheckState()
	})
}

// updateOrganizationAuthenticationSettings is used by the organization_authentication_settings resource to update an organization authentication settings in Genesys Cloud
func updateOrganizationAuthenticationSettings(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	proxy := getOrgAuthSettingsProxy(sdkConfig)
	AuthSettings := getOrganizationAuthenticationSettingsFromResourceData(d)

	log.Printf("Updating organization authentication settings %s", d.Id())

	orgAuthSettings, _, err := proxy.updateOrgAuthSettings(ctx, &AuthSettings)
	if err != nil {
		return diag.Errorf("Failed to update organization authentication settings: %s", err)
	}

	log.Printf("Updated organization authentication settings %s %s", d.Id(), orgAuthSettings)
	return readOrganizationAuthenticationSettings(ctx, d, meta)
}

// deleteOrganizationAuthenticationSettings is used by the organization_authentication_settings resource to delete an organization authentication settings from Genesys cloud
func deleteOrganizationAuthenticationSettings(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return nil
}
