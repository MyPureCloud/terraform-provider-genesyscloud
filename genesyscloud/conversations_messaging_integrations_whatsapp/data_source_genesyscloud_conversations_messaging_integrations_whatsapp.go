package conversations_messaging_integrations_whatsapp

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
   The data_source_genesyscloud_conversations_messaging_integrations_whatsapp.go contains the data source implementation
   for the resource.
*/

// dataSourceConversationsMessagingIntegrationsWhatsappRead retrieves by name the id in question
func dataSourceConversationsMessagingIntegrationsWhatsappRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getConversationsMessagingIntegrationsWhatsappProxy(sdkConfig)

	name := d.Get("name").(string)

	return util.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		whatsAppEmbeddedSignupIntegrationRequestId, retryable, resp, err := proxy.getConversationsMessagingIntegrationsWhatsappIdByName(ctx, name)

		if err != nil {
			if retryable {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("No conversations messaging integrations whatsapp found with name %s", name), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Error searching conversations messaging integrations whatsapp %s: %s", name, err), resp))
		}

		d.SetId(whatsAppEmbeddedSignupIntegrationRequestId)
		return nil
	})
}
