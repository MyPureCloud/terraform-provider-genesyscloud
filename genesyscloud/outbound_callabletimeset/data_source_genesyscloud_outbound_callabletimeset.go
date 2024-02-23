package outbound_callabletimeset

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

/*
The data_source_genesyscloud_outbound_callabletimeset.go contains the data source implementation for the resource.
*/

// dataSourceOutboundCallabletimesetRead retrieves by name term the id in question
func dataSourceOutboundCallabletimesetRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sdkConfig := m.(*provider.ProviderMeta).ClientConfig
	proxy := getOutboundCallabletimesetProxy(sdkConfig)
	timesetName := d.Get("name").(string)

	return util.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		timesetId, retryable, err := proxy.getOutboundCallabletimesetByName(ctx, timesetName)

		if err != nil && !retryable {
			return retry.NonRetryableError(fmt.Errorf("error requesting callable timeset %s: %s", timesetName, err))
		}

		if retryable {
			return retry.RetryableError(fmt.Errorf("no callable timeset found with timesetName %s", timesetName))
		}

		d.SetId(timesetId)
		return nil
	})
}
