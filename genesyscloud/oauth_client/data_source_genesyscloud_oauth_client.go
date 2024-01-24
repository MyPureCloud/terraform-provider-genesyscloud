package oauth_client

import (
	"context"
	"fmt"
	"terraform-provider-genesyscloud/genesyscloud"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceOAuthClientRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sdkConfig := m.(*genesyscloud.ProviderMeta).ClientConfig
	oauthClientProxy := getOAuthClientProxy(sdkConfig)

	name := d.Get("name").(string)

	// Find first non-deleted oauth client by name. Retry in case new oauth client is not yet indexed by search
	return genesyscloud.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		clients, getErr := oauthClientProxy.getAllOAuthClients(ctx)
		if getErr != nil {
			return retry.NonRetryableError(fmt.Errorf("Error requesting oauth client %s: %s", name, getErr))
		}

		if len(*clients) == 0 {
			return retry.RetryableError(fmt.Errorf("No oauth clients found with name %s", name))
		}

		for _, oauth := range *clients {
			if oauth.Name != nil && *oauth.Name == name &&
				oauth.State != nil && *oauth.State != "deleted" {
				d.SetId(*oauth.Id)
				return nil
			}
		}

		return nil
	})

}
