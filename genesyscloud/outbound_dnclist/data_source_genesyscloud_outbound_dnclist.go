package outbound_dnclist

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	gcloud "terraform-provider-genesyscloud/genesyscloud"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceOutboundDncListRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sdkConfig := m.(*gcloud.ProviderMeta).ClientConfig
	proxy := getOutboundDnclistProxy(sdkConfig)
	name := d.Get("name").(string)

	return gcloud.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		dnclistId, retryable, getErr := proxy.getOutboundDnclistByName(ctx, name)
		if getErr != nil && !retryable {
			return retry.NonRetryableError(fmt.Errorf("error requesting dnc lists %s: %s", name, getErr))
		}
		if retryable {
			return retry.RetryableError(fmt.Errorf("no dnc lists found with name %s", name))
		}
		d.SetId(dnclistId)
		return nil
	})
}
