package apple_integration

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v171/platformclientv2"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/consistency_checker"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/resourcedata"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

/*
The resource_genesyscloud_apple_integration.go contains all of the methods that perform the core logic for a resource.
*/

// getAllAuthAppleIntegration retrieves all of the apple integration via Terraform in the Genesys Cloud and is used for the exporter
func getAllAuthAppleIntegrations(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	proxy := getAppleIntegrationProxy(clientConfig)
	resources := make(resourceExporter.ResourceIDMetaMap)

	appleIntegrations, _, err := proxy.getAllAppleIntegration(ctx)
	if err != nil {
		return nil, diag.Errorf("Failed to get apple integration: %v", err)
	}

	for _, appleIntegration := range *appleIntegrations {
		resources[*appleIntegration.Id] = &resourceExporter.ResourceMeta{}
	}

	return resources, nil
}

// createAppleIntegration is used by the apple_integration resource to create Genesys cloud apple integration
func createAppleIntegration(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getAppleIntegrationProxy(sdkConfig)

	appleIntegration := getAppleIntegrationFromResourceData(d)

	log.Printf("Creating apple integration %s", *appleIntegration.Name)
	createdAppleIntegration, _, err := proxy.createAppleIntegration(ctx, &appleIntegration)
	if err != nil {
		return diag.Errorf("Failed to create apple integration: %s", err)
	}
	appleIntegration = *createdAppleIntegration

	d.SetId(*appleIntegration.Id)
	log.Printf("Created apple integration %s", *appleIntegration.Id)
	return readAppleIntegration(ctx, d, meta)
}

// readAppleIntegration is used by the apple_integration resource to read an apple integration from genesys cloud
func readAppleIntegration(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getAppleIntegrationProxy(sdkConfig)
	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceAppleIntegration(), 5, resourceName)

	log.Printf("Reading apple integration %s", d.Id())

	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		appleIntegration, resp, getErr := proxy.getAppleIntegrationById(ctx, d.Id())
		if getErr != nil {
			if resp != nil && resp.StatusCode == 404 {
				return retry.RetryableError(fmt.Errorf("Failed to read apple integration %s: %s", d.Id(), getErr))
			}
			return retry.NonRetryableError(fmt.Errorf("Failed to read apple integration %s: %s", d.Id(), getErr))
		}

		resourcedata.SetNillableValue(d, "name", appleIntegration.Name)
		resourcedata.SetNillableValueWithInterfaceArrayWithFunc(d, "supported_content", appleIntegration.SupportedContent, flattenSupportedContentReference)
		resourcedata.SetNillableValueWithInterfaceArrayWithFunc(d, "messaging_setting", appleIntegration.MessagingSetting, flattenMessagingSettingReference)
		resourcedata.SetNillableValue(d, "messages_for_business_id", appleIntegration.MessagesForBusinessId)
		resourcedata.SetNillableValue(d, "business_name", appleIntegration.BusinessName)
		resourcedata.SetNillableValue(d, "logo_url", appleIntegration.LogoUrl)
		resourcedata.SetNillableValue(d, "status", appleIntegration.Status)
		// RecipientId field not available in v171 SDK
		resourcedata.SetNillableValue(d, "create_status", appleIntegration.CreateStatus)
		resourcedata.SetNillableValueWithInterfaceArrayWithFunc(d, "create_error", appleIntegration.CreateError, flattenErrorBody)
		resourcedata.SetNillableValueWithInterfaceArrayWithFunc(d, "apple_i_message_app", appleIntegration.AppleIMessageApp, flattenAppleIMessageApp)
		resourcedata.SetNillableValueWithInterfaceArrayWithFunc(d, "apple_authentication", appleIntegration.AppleAuthentication, flattenAppleAuthentication)
		resourcedata.SetNillableValueWithInterfaceArrayWithFunc(d, "apple_pay", appleIntegration.ApplePay, flattenApplePay)
		resourcedata.SetNillableValueWithInterfaceArrayWithFunc(d, "identity_resolution", appleIntegration.IdentityResolution, flattenAppleIdentityResolutionConfig)

		log.Printf("Read apple integration %s %s", d.Id(), *appleIntegration.Name)
		return cc.CheckState(d)
	})
}

// updateAppleIntegration is used by the apple_integration resource to update an apple integration in Genesys Cloud
func updateAppleIntegration(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getAppleIntegrationProxy(sdkConfig)

	appleIntegration := getAppleIntegrationFromResourceData(d)

	log.Printf("Updating apple integration %s", *appleIntegration.Name)
	updatedAppleIntegration, _, err := proxy.updateAppleIntegration(ctx, d.Id(), &appleIntegration)
	if err != nil {
		return diag.Errorf("Failed to update apple integration %s: %s", d.Id(), err)
	}
	appleIntegration = *updatedAppleIntegration

	log.Printf("Updated apple integration %s", *appleIntegration.Id)
	return readAppleIntegration(ctx, d, meta)
}

// deleteAppleIntegration is used by the apple_integration resource to delete an apple integration from Genesys cloud
func deleteAppleIntegration(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getAppleIntegrationProxy(sdkConfig)

	_, err := proxy.deleteAppleIntegration(ctx, d.Id())
	if err != nil {
		return diag.Errorf("Failed to delete apple integration %s: %s", d.Id(), err)
	}

	return util.WithRetries(ctx, 180*time.Second, func() *retry.RetryError {
		_, resp, err := proxy.getAppleIntegrationById(ctx, d.Id())

		if err != nil {
			if resp != nil && resp.StatusCode == 404 {
				log.Printf("Deleted apple integration %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(fmt.Errorf("Error deleting apple integration %s: %s", d.Id(), err))
		}

		return retry.RetryableError(fmt.Errorf("apple integration %s still exists", d.Id()))
	})
}
