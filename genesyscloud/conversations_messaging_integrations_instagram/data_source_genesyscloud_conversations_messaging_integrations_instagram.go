package conversations_messaging_integrations_instagram

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
)

/*
   The data_source_genesyscloud_conversations_messaging_integrations_instagram.go contains the data source implementation
   for the resource.
*/

// dataSourceConversationsMessagingIntegrationsInstagramRead retrieves by name the id in question
func dataSourceConversationsMessagingIntegrationsInstagramRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getConversationsMessagingIntegrationsInstagramProxy(sdkConfig)

	name := d.Get("name").(string)

	return util.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		instagramIntegrationRequestId, retryable, _, err := proxy.getConversationsMessagingIntegrationsInstagramIdByName(ctx, name)

		if err != nil && !retryable {
			return retry.NonRetryableError(fmt.Errorf("Error searching conversations messaging integrations instagram %s: %s", name, err))
		}

		if retryable {
			return retry.RetryableError(fmt.Errorf("No conversations messaging integrations instagram found with name %s", name))
		}

		d.SetId(instagramIntegrationRequestId)
		return nil
	})
}
