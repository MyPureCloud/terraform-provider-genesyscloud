package genesyscloud

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v115/platformclientv2"
)

func dataSourceOAuthClient() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for Genesys Cloud OAuth Clients. Select an OAuth Client by name.",
		ReadContext: ReadWithPooledClient(dataSourceOAuthClientRead),
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "OAuth Client name.",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}

func dataSourceOAuthClientRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sdkConfig := m.(*ProviderMeta).ClientConfig
	oauthAPI := platformclientv2.NewOAuthApiWithConfig(sdkConfig)

	name := d.Get("name").(string)

	// Find first non-deleted oauth client by name. Retry in case new oauth client is not yet indexed by search
	return WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		for pageNum := 1; ; pageNum++ {
			oauths, _, getErr := oauthAPI.GetOauthClients()
			if getErr != nil {
				return retry.NonRetryableError(fmt.Errorf("Error requesting oauth client %s: %s", name, getErr))
			}

			if oauths.Entities == nil || len(*oauths.Entities) == 0 {
				return retry.RetryableError(fmt.Errorf("No oauth clients found with name %s", name))
			}

			for _, oauth := range *oauths.Entities {
				if oauth.Name != nil && *oauth.Name == name &&
					oauth.State != nil && *oauth.State != "deleted" {
					d.SetId(*oauth.Id)
					return nil
				}
			}
		}
	})
}
