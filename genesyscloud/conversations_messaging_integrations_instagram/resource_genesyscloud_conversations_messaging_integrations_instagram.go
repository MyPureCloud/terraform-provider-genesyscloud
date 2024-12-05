package conversations_messaging_integrations_instagram

import (
	"context"
	"fmt"
	"log"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v146/platformclientv2"

	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"

	"terraform-provider-genesyscloud/genesyscloud/util/constants"
	"terraform-provider-genesyscloud/genesyscloud/util/resourcedata"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

/*
The resource_genesyscloud_conversations_messaging_integrations_instagram.go contains all of the methods that perform the core logic for a resource.
*/

// getAllAuthConversationsMessagingIntegrationsInstagram retrieves all of the conversations messaging integrations instagram via Terraform in the Genesys Cloud and is used for the exporter
func getAllAuthConversationsMessagingIntegrationsInstagrams(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	proxy := getConversationsMessagingIntegrationsInstagramProxy(clientConfig)
	resources := make(resourceExporter.ResourceIDMetaMap)

	instagramIntegrationRequests, resp, err := proxy.getAllConversationsMessagingIntegrationsInstagram(ctx)
	if err != nil {
		return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get conversations messaging integrations instagram: %v", err), resp)
	}

	for _, instagramIntegrationRequest := range *instagramIntegrationRequests {
		resources[*instagramIntegrationRequest.Id] = &resourceExporter.ResourceMeta{BlockLabel: *instagramIntegrationRequest.Name}
	}

	return resources, nil
}

// createConversationsMessagingIntegrationsInstagram is used by the conversations_messaging_integrations_instagram resource to create Genesys cloud conversations messaging integrations instagram
func createConversationsMessagingIntegrationsInstagram(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getConversationsMessagingIntegrationsInstagramProxy(sdkConfig)

	conversationsMessagingIntegrationsInstagram := getConversationsMessagingIntegrationsInstagramFromResourceData(d)

	log.Printf("Creating conversations messaging integrations instagram %s", *conversationsMessagingIntegrationsInstagram.Name)
	instagramIntegrationRequest, resp, err := proxy.createConversationsMessagingIntegrationsInstagram(ctx, &conversationsMessagingIntegrationsInstagram)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to create conversations messaging integrations instagram: %s", err), resp)
	}

	d.SetId(*instagramIntegrationRequest.Id)
	log.Printf("Created conversations messaging integrations instagram %s", *instagramIntegrationRequest.Id)
	return readConversationsMessagingIntegrationsInstagram(ctx, d, meta)
}

// readConversationsMessagingIntegrationsInstagram is used by the conversations_messaging_integrations_instagram resource to read an conversations messaging integrations instagram from genesys cloud
func readConversationsMessagingIntegrationsInstagram(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getConversationsMessagingIntegrationsInstagramProxy(sdkConfig)

	log.Printf("Reading conversations messaging integrations instagram %s", d.Id())

	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		instagramIntegrationRequest, resp, getErr := proxy.getConversationsMessagingIntegrationsInstagramById(ctx, d.Id())
		if getErr != nil {
			if util.IsStatus404(resp) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read conversations messaging integrations instagram %s: %s", d.Id(), getErr), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read conversations messaging integrations instagram %s: %s", d.Id(), getErr), resp))
		}

		cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceConversationsMessagingIntegrationsInstagram(), constants.ConsistencyChecks(), ResourceType)

		resourcedata.SetNillableValue(d, "name", instagramIntegrationRequest.Name)

		if instagramIntegrationRequest.SupportedContent != nil && instagramIntegrationRequest.SupportedContent.Id != nil {
			_ = d.Set("supported_content_id", *instagramIntegrationRequest.SupportedContent.Id)
		}

		if instagramIntegrationRequest.MessagingSetting != nil && instagramIntegrationRequest.MessagingSetting.Id != nil {
			_ = d.Set("messaging_setting_id", *instagramIntegrationRequest.MessagingSetting.Id)
		}

		resourcedata.SetNillableValue(d, "page_id", instagramIntegrationRequest.PageId)
		resourcedata.SetNillableValue(d, "app_id", instagramIntegrationRequest.AppId)

		log.Printf("Read conversations messaging integrations instagram %s", d.Id())
		return cc.CheckState(d)
	})
}

// updateConversationsMessagingIntegrationsInstagram is used by the conversations_messaging_integrations_instagram resource to update an conversations messaging integrations instagram in Genesys Cloud
func updateConversationsMessagingIntegrationsInstagram(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getConversationsMessagingIntegrationsInstagramProxy(sdkConfig)

	conversationsMessagingIntegrationsInstagram := getConversationsMessagingIntegrationsInstagramFromResourceDataForUpdate(d)

	log.Printf("Updating conversations messaging integrations instagram %s", *conversationsMessagingIntegrationsInstagram.Name)
	instagramIntegrationRequest, resp, err := proxy.updateConversationsMessagingIntegrationsInstagram(ctx, d.Id(), &conversationsMessagingIntegrationsInstagram)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to update conversations messaging integrations instagram: %s", err), resp)
	}

	log.Printf("Updated conversations messaging integrations instagram %s", *instagramIntegrationRequest.Id)
	return readConversationsMessagingIntegrationsInstagram(ctx, d, meta)
}

// deleteConversationsMessagingIntegrationsInstagram is used by the conversations_messaging_integrations_instagram resource to delete an conversations messaging integrations instagram from Genesys cloud
func deleteConversationsMessagingIntegrationsInstagram(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getConversationsMessagingIntegrationsInstagramProxy(sdkConfig)

	resp, err := proxy.deleteConversationsMessagingIntegrationsInstagram(ctx, d.Id())
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to delete conversations messaging integrations instagram %s: %s", d.Id(), err), resp)
	}

	return util.WithRetries(ctx, 180*time.Second, func() *retry.RetryError {
		_, resp, err := proxy.getConversationsMessagingIntegrationsInstagramById(ctx, d.Id())

		if err != nil {
			if util.IsStatus404(resp) {
				log.Printf("Deleted conversations messaging integrations instagram %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Error deleting conversations messaging integrations instagram %s: %s", d.Id(), err), resp))
		}

		return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("conversations messaging integrations instagram %s still exists", d.Id()), resp))
	})
}
