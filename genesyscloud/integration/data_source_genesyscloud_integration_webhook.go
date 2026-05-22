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

		// Check if attributes exist
		if integration.Attributes == nil {
			return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("attributes is nil for integration %s, retrying...", integrationName), resp))
		}

		// Convert attributes to JSON string
		attrJSONStr, err := json.Marshal(*integration.Attributes)
		if err != nil {
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("failed to marshal integration attributes for %s | error: %s", integrationName, err), resp))
		}

		// Parse attributes JSON
		var attributes map[string]interface{}
		if err := json.Unmarshal(attrJSONStr, &attributes); err != nil {
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("failed to parse integration attributes for %s | error: %s", integrationName, err), resp))
		}

		// Extract webhookId
		webhookId, exists := attributes["webhookId"]
		if !exists {
			return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("webhookId not found in attributes for integration %s, retrying...", integrationName), resp))
		}

		webhookIdStr, ok := webhookId.(string)
		if !ok || webhookIdStr == "" {
			return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("webhookId is empty for integration %s, retrying...", integrationName), resp))
		}

		d.Set("web_hook_id", webhookIdStr)

		// Extract invocation URL
		invocationURL, exists := attributes["invocationUrl"]
		if !exists {
			return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("invocationUrl not found in attributes for integration %s, retrying...", integrationName), resp))
		}

		invocationURLStr, ok := invocationURL.(string)
		if !ok || invocationURLStr == "" {
			return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("invocationUrl is empty for integration %s, retrying...", integrationName), resp))
		}

		d.Set("invocation_url", invocationURLStr)

		return nil
	})
}
