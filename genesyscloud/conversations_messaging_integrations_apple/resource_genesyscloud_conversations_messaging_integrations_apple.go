package conversations_messaging_integrations_apple

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

// getAllAppleIntegrations retrieves all of the apple integration via Terraform in the Genesys Cloud and is used for the exporter
func getAllConversationsMessagingIntegrationsApple(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	proxy := getConversationsMessagingIntegrationsAppleProxy(clientConfig)
	resources := make(resourceExporter.ResourceIDMetaMap)

	appleIntegrations, resp, err := proxy.getAllConversationsMessagingIntegrationsApple(ctx)
	if err != nil {
		return nil, util.BuildAPIDiagnosticError(resourceName, fmt.Sprintf("Failed to get apple integration: %s", err), resp)
	}

	for _, appleIntegration := range *appleIntegrations {
		resources[*appleIntegration.Id] = &resourceExporter.ResourceMeta{}
	}

	return resources, nil
}

// createAppleIntegration is used by the apple_integration resource to create Genesys cloud apple integration
func createConversationsMessagingIntegrationsApple(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getConversationsMessagingIntegrationsAppleProxy(sdkConfig)

	request := getConversationsMessagingIntegrationsAppleFromResourceData(d)

	log.Printf("Creating apple integration %s", *request.Name)
	createdAppleIntegration, resp, err := proxy.createConversationsMessagingIntegrationsApple(ctx, &request)
	if err != nil {
		return util.BuildAPIDiagnosticError(resourceName, fmt.Sprintf("Failed to create apple integration %s", *request.Name), resp)
	}

	d.SetId(*createdAppleIntegration.Id)
	log.Printf("Created apple integration %s", *createdAppleIntegration.Id)
	return readConversationsMessagingIntegrationsApple(ctx, d, meta)
}

// readAppleIntegration is used by the apple_integration resource to read an apple integration from genesys cloud
func readConversationsMessagingIntegrationsApple(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getConversationsMessagingIntegrationsAppleProxy(sdkConfig)
	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceConversationsMessagingIntegrationsApple(), 5, resourceName)

	log.Printf("Reading apple integration %s", d.Id())

	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		appleIntegration, resp, getErr := proxy.getConversationsMessagingIntegrationsAppleById(ctx, d.Id())
		if getErr != nil {
			if resp != nil && resp.StatusCode == 404 {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(resourceName, fmt.Sprintf("Failed to read apple integration %s", d.Id()), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(resourceName, fmt.Sprintf("Failed to read apple integration %s", d.Id()), resp))
		}

		resourcedata.SetNillableValue(d, "name", appleIntegration.Name)
		if appleIntegration.SupportedContent != nil && appleIntegration.SupportedContent.Id != nil {
			d.Set("supported_content_id", *appleIntegration.SupportedContent.Id)
		}
		if appleIntegration.MessagingSetting != nil && appleIntegration.MessagingSetting.Id != nil {
			d.Set("messaging_setting_id", *appleIntegration.MessagingSetting.Id)
		}
		resourcedata.SetNillableValue(d, "messages_for_business_id", appleIntegration.MessagesForBusinessId)
		resourcedata.SetNillableValue(d, "business_name", appleIntegration.BusinessName)
		resourcedata.SetNillableValue(d, "logo_url", appleIntegration.LogoUrl)
		resourcedata.SetNillableValueWithInterfaceArrayWithFunc(d, "apple_i_message_app", appleIntegration.AppleIMessageApp, flattenAppleIMessageApp)
		resourcedata.SetNillableValueWithInterfaceArrayWithFunc(d, "apple_authentication", appleIntegration.AppleAuthentication, flattenAppleAuthentication)
		resourcedata.SetNillableValueWithInterfaceArrayWithFunc(d, "apple_pay", appleIntegration.ApplePay, flattenApplePay)

		log.Printf("Read apple integration %s %s", d.Id(), *appleIntegration.Name)
		return cc.CheckState(d)
	})
}

// updateAppleIntegration is used by the apple_integration resource to update an apple integration in Genesys Cloud
func updateConversationsMessagingIntegrationsApple(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getConversationsMessagingIntegrationsAppleProxy(sdkConfig)

	request := getConversationsMessagingIntegrationsAppleFromResourceDataForUpdate(d)

	log.Printf("Updating apple integration %s", *request.Name)
	updatedAppleIntegration, resp, updateErr := proxy.updateConversationsMessagingIntegrationsApple(ctx, d.Id(), &request)
	if updateErr != nil {
		return util.BuildAPIDiagnosticError(resourceName, fmt.Sprintf("Failed to update apple integration %s", d.Id()), resp)
	}

	log.Printf("Updated apple integration %s", *updatedAppleIntegration.Id)
	return readConversationsMessagingIntegrationsApple(ctx, d, meta)
}

// deleteAppleIntegration is used by the apple_integration resource to delete an apple integration from Genesys cloud
func deleteConversationsMessagingIntegrationsApple(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getConversationsMessagingIntegrationsAppleProxy(sdkConfig)

	resp, err := proxy.deleteConversationsMessagingIntegrationsApple(ctx, d.Id())
	if err != nil {
		return util.BuildAPIDiagnosticError(resourceName, fmt.Sprintf("Failed to delete apple integration %s", d.Id()), resp)
	}

	return util.WithRetries(ctx, 180*time.Second, func() *retry.RetryError {
		_, resp, err := proxy.getConversationsMessagingIntegrationsAppleById(ctx, d.Id())

		if err != nil {
			if resp != nil && resp.StatusCode == 404 {
				log.Printf("Deleted apple integration %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(resourceName, fmt.Sprintf("Error deleting apple integration %s", d.Id()), resp))
		}

		return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(resourceName, fmt.Sprintf("Apple integration %s still exists", d.Id()), resp))
	})
}
