package telephony_providers_edges_site

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

func dataSourceSiteRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sdkConfig := m.(*provider.ProviderMeta).ClientConfig
	sp := getSiteProxy(sdkConfig)

	name := d.Get("name").(string)
	managed := d.Get("managed").(bool)

	return util.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		siteId, retryable, resp, err := sp.getSiteIdByName(ctx, name, managed)
		if err != nil {
			if retryable {
				return retry.RetryableError(fmt.Errorf("failed to get site %s %v", name, resp))
			}

			return retry.NonRetryableError(fmt.Errorf("error requesting site %s: %s", name, err))
		}

		d.SetId(siteId)
		return nil
	})
}
