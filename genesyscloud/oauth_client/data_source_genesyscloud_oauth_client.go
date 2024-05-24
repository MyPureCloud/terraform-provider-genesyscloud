package oauth_client

import (
	"context"
	"fmt"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceOAuthClientRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sdkConfig := m.(*provider.ProviderMeta).ClientConfig
	oauthClientProxy := GetOAuthClientProxy(sdkConfig)

	name := d.Get("name").(string)

	// Find first non-deleted oauth client by name. Retry in case new oauth client is not yet indexed by search
	return util.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		clients, resp, getErr := oauthClientProxy.getAllOAuthClients(ctx)
		if getErr != nil {
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(resourceName, fmt.Sprintf("Error requesting oauth client %s | error: %s", name, getErr), resp))
		}

		if len(*clients) == 0 {
			return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(resourceName, fmt.Sprintf("No oauth clients found with name %s", name), resp))
		}

		for _, oauth := range *clients {
			if oauth.Name != nil && *oauth.Name == name &&
				oauth.State != nil && *oauth.State != "deleted" {
				d.SetId(*oauth.Id)
				return nil
			}
		}

		return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(resourceName, fmt.Sprintf("Unable to locate a oauth client with name %s", name), resp))
	})

}
