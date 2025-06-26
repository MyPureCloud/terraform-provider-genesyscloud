package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

/*
   The data_source_genesyscloud_integration_webhook.go contains the data source implementation
   for looking up webhook integrations by name.

   This data source specifically looks for integrations with type "webhook" that match the provided name.
*/

// dataSourceIntegrationWebhookRead retrieves webhook integration by name
func dataSourceIntegrationWebhookRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sdkConfig := m.(*provider.ProviderMeta).ClientConfig
	ip := getIntegrationsProxy(sdkConfig)
	integrationName := d.Get("name").(string)

	return util.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		integration, retryable, resp, err := ip.getIntegrationByName(ctx, integrationName)
		if err != nil && !retryable {
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("failed to get integration by name: %s | error: %s", integrationName, err), resp))
		}
		if retryable {
			return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("failed to get integration %s", integrationName), resp))
		}

		// Check if the integration is of type "webhook"
		if integration.IntegrationType == nil || integration.IntegrationType.Id == nil || *integration.IntegrationType.Id != "webhook" {
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("integration %s is not of type 'webhook'", integrationName), resp))
		}

		// Set the integration ID
		d.SetId(*integration.Id)

		// Extract webhookId and invocation URL from attributes
		if integration.Attributes != nil {
			// Convert attributes to JSON string first
			attrJSONStr, err := json.Marshal(*integration.Attributes)
			if err != nil {
				return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("failed to marshal integration attributes for %s | error: %s", integrationName, err), resp))
			}

			var attributes map[string]interface{}
			if err := json.Unmarshal(attrJSONStr, &attributes); err != nil {
				return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("failed to parse integration attributes for %s | error: %s", integrationName, err), resp))
			}

			// Extract webhookId
			if webhookId, exists := attributes["webhookId"]; exists {
				if webhookIdStr, ok := webhookId.(string); ok && webhookIdStr != "" {
					d.Set("web_hook_id", webhookIdStr)
				} else {
					// If webhookId exists but is empty, this might be a timing issue
					return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("webhookId is empty for integration %s, retrying...", integrationName), resp))
				}
			} else {
				// If webhookId doesn't exist in attributes, this might be a timing issue
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("webhookId not found in attributes for integration %s, retrying...", integrationName), resp))
			}

			// Extract invocation URL
			if invocationURL, exists := attributes["invocationUrl"]; exists {
				if invocationURLStr, ok := invocationURL.(string); ok && invocationURLStr != "" {
					d.Set("invocation_url", invocationURLStr)
				} else {
					// If invocationUrl exists but is empty, this might be a timing issue
					return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("invocationUrl is empty for integration %s, retrying...", integrationName), resp))
				}
			} else {
				// If invocationUrl doesn't exist in attributes, this might be a timing issue
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("invocationUrl not found in attributes for integration %s, retrying...", integrationName), resp))
			}
		} else {
			// If attributes is nil, this might be a timing issue
			return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("attributes is nil for integration %s, retrying...", integrationName), resp))
		}

		return nil
	})
}
