package conversations_messaging_integrations_open

import (
	"context"
	"encoding/json"
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
The resource_genesyscloud_conversations_messaging_integrations_open.go contains all of the methods that perform the core logic for a resource.
*/

// getAllAuthConversationsMessagingIntegrationsOpen retrieves all of the conversations messaging integrations open via Terraform in the Genesys Cloud and is used for the exporter
func getAllAuthConversationsMessagingIntegrationsOpens(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	proxy := getConversationsMessagingIntegrationsOpenProxy(clientConfig)
	resources := make(resourceExporter.ResourceIDMetaMap)

	openIntegrationRequests, resp, err := proxy.getAllConversationsMessagingIntegrationsOpen(ctx)
	if err != nil {
		return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get conversations messaging integrations open: %v", err), resp)
	}

	for _, openIntegrationRequest := range *openIntegrationRequests {
		resources[*openIntegrationRequest.Id] = &resourceExporter.ResourceMeta{BlockLabel: *openIntegrationRequest.Name}
	}

	return resources, nil
}

// createConversationsMessagingIntegrationsOpen is used by the conversations_messaging_integrations_open resource to create Genesys cloud conversations messaging integrations open
func createConversationsMessagingIntegrationsOpen(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getConversationsMessagingIntegrationsOpenProxy(sdkConfig)

	conversationsMessagingIntegrationsOpen := getConversationsMessagingIntegrationsOpenFromResourceData(d)

	log.Printf("Creating conversations messaging integrations open %s", *conversationsMessagingIntegrationsOpen.Name)
	openIntegrationRequest, resp, err := proxy.createConversationsMessagingIntegrationsOpen(ctx, &conversationsMessagingIntegrationsOpen)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to create conversations messaging integrations open: %s", err), resp)
	}

	d.SetId(*openIntegrationRequest.Id)
	log.Printf("Created conversations messaging integrations open %s", *openIntegrationRequest.Id)
	return readConversationsMessagingIntegrationsOpen(ctx, d, meta)
}

// readConversationsMessagingIntegrationsOpen is used by the conversations_messaging_integrations_open resource to read an conversations messaging integrations open from genesys cloud
func readConversationsMessagingIntegrationsOpen(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getConversationsMessagingIntegrationsOpenProxy(sdkConfig)

	log.Printf("Reading conversations messaging integrations open %s", d.Id())

	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceConversationsMessagingIntegrationsOpen(), constants.ConsistencyChecks(), ResourceType)

	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		openIntegrationRequest, resp, err := proxy.getConversationsMessagingIntegrationsOpenById(ctx, d.Id())
		if err != nil {
			if util.IsStatus404(resp) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read conversations messaging integrations open %s: %s", d.Id(), err), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read conversations messaging integrations open %s: %s", d.Id(), err), resp))
		}

		resourcedata.SetNillableValue(d, "name", openIntegrationRequest.Name)

		if openIntegrationRequest.SupportedContent != nil && openIntegrationRequest.SupportedContent.Id != nil {
			_ = d.Set("supported_content_id", *openIntegrationRequest.SupportedContent.Id)
		}

		if openIntegrationRequest.MessagingSetting != nil && openIntegrationRequest.MessagingSetting.Id != nil {
			_ = d.Set("messaging_setting_id", *openIntegrationRequest.MessagingSetting.Id)
		}

		resourcedata.SetNillableValue(d, "outbound_notification_webhook_url", openIntegrationRequest.OutboundNotificationWebhookUrl)

		webhookProps, _ := json.Marshal(openIntegrationRequest.WebhookHeaders)
		var webhookPropsPtr *string
		if string(webhookProps) != util.NullValue {
			webhookPropsStr := string(webhookProps)
			webhookPropsPtr = &webhookPropsStr
		}
		_ = d.Set("webhook_headers", *webhookPropsPtr)

		log.Printf("Read conversations messaging integrations open %s %s", d.Id(), *openIntegrationRequest.Name)
		return cc.CheckState(d)
	})
}

// updateConversationsMessagingIntegrationsOpen is used by the conversations_messaging_integrations_open resource to update an conversations messaging integrations open in Genesys Cloud
func updateConversationsMessagingIntegrationsOpen(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getConversationsMessagingIntegrationsOpenProxy(sdkConfig)

	conversationsMessagingIntegrationsOpen := getConversationsMessagingIntegrationsOpenFromResourceDataForUpdate(d)

	log.Printf("Updating conversations messaging integrations open %s", *conversationsMessagingIntegrationsOpen.Name)
	openIntegrationRequest, resp, err := proxy.updateConversationsMessagingIntegrationsOpen(ctx, d.Id(), &conversationsMessagingIntegrationsOpen)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to update conversations messaging integrations open: %s", err), resp)
	}

	log.Printf("Updated conversations messaging integrations open %s", *openIntegrationRequest.Id)
	return readConversationsMessagingIntegrationsOpen(ctx, d, meta)
}

// deleteConversationsMessagingIntegrationsOpen is used by the conversations_messaging_integrations_open resource to delete an conversations messaging integrations open from Genesys cloud
func deleteConversationsMessagingIntegrationsOpen(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getConversationsMessagingIntegrationsOpenProxy(sdkConfig)

	resp, err := proxy.deleteConversationsMessagingIntegrationsOpen(ctx, d.Id())
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to delete conversations messaging integrations open %s: %s", d.Id(), err), resp)
	}

	return util.WithRetries(ctx, 180*time.Second, func() *retry.RetryError {
		_, resp, err := proxy.getConversationsMessagingIntegrationsOpenById(ctx, d.Id())

		if err != nil {
			if util.IsStatus404(resp) {
				log.Printf("Deleted conversations messaging integrations open %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Error deleting conversations messaging integrations open %s: %s", d.Id(), err), resp))
		}

		return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("conversations messaging integrations open %s still exists", d.Id()), resp))
	})
}
