package telephony_providers_edges_site

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	gcloud "terraform-provider-genesyscloud/genesyscloud"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceSiteRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sdkConfig := m.(*gcloud.ProviderMeta).ClientConfig
	sp := getSiteProxy(sdkConfig)

	name := d.Get("name").(string)
	managed := d.Get("managed").(bool)

	return gcloud.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		siteId, retryable, err := sp.getSiteIdByName(ctx, name, managed)
		if err != nil {
			if retryable {
				return retry.RetryableError(fmt.Errorf("failed to get site %s", name))
			}

			return retry.NonRetryableError(fmt.Errorf("error requesting site %s: %s", name, err))
		}

		d.SetId(siteId)
		return nil
	})
}
