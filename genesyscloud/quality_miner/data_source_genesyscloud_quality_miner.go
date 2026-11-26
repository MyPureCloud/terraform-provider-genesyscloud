package quality_miner

import (
	"context"
	"fmt"
	"time"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceQualityMinerRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sdkConfig := m.(*provider.ProviderMeta).ClientConfig
	proxy := getQualityMinerProxy(sdkConfig)
	name := d.Get("name").(string)
	minerType := d.Get("miner_type").(string) // optional; "" means match any type

	// Find first miner by name (optionally narrowed by type). Retry in case a newly created miner is not yet indexed.
	return util.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		minerId, resp, retryable, err := proxy.getQualityMinerIdByName(ctx, name, minerType)
		if err != nil && !retryable {
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Error requesting miner %s | error: %s", name, err), resp))
		}
		if retryable {
			return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Error requesting miner %s | error: %s", name, err), resp))
		}

		d.SetId(minerId)
		return nil
	})
}
