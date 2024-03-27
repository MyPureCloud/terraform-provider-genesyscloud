package outbound_dnclist

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

func dataSourceOutboundDncListRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sdkConfig := m.(*provider.ProviderMeta).ClientConfig
	proxy := getOutboundDnclistProxy(sdkConfig)
	name := d.Get("name").(string)

	return util.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		dnclistId, retryable, resp, getErr := proxy.getOutboundDnclistByName(ctx, name)
		if getErr != nil && !retryable {
			return retry.NonRetryableError(fmt.Errorf("error requesting dnc lists %s: %s %v", name, getErr, resp))
		}
		if retryable {
			return retry.RetryableError(fmt.Errorf("no dnc lists found with name %s", name))
		}
		d.SetId(dnclistId)
		return nil
	})
}
