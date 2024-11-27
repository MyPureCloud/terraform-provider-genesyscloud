package conversations_messaging_integrations_open

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
)

/*
   The data_source_genesyscloud_conversations_messaging_integrations_open.go contains the data source implementation
   for the resource.
*/

// dataSourceConversationsMessagingIntegrationsOpenRead retrieves by name the id in question
func dataSourceConversationsMessagingIntegrationsOpenRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getConversationsMessagingIntegrationsOpenProxy(sdkConfig)

	name := d.Get("name").(string)

	return util.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		openIntegrationRequestId, retryable, resp, err := proxy.getConversationsMessagingIntegrationsOpenIdByName(ctx, name)

		if err != nil {
			if retryable {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("No conversations messaging integrations open found with name %s", name), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Error searching conversations messaging integrations open %s: %s", name, err), resp))
		}

		d.SetId(openIntegrationRequestId)
		return nil
	})
}
