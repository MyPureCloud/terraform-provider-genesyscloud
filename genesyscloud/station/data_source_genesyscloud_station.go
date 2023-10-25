package station

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	gcloud "terraform-provider-genesyscloud/genesyscloud"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceStationRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sdkConfig := m.(*gcloud.ProviderMeta).ClientConfig
	sp := getStationProxy(sdkConfig)

	stationName := d.Get("name").(string)

	return gcloud.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		stationId, retryable, err := sp.getStationIdByName(ctx, stationName)
		if err != nil && !retryable {
			return retry.NonRetryableError(fmt.Errorf("error requesting station %s", err))
		}

		if retryable {
			return retry.RetryableError(fmt.Errorf("no stations found"))
		}

		d.SetId(stationId)
		return nil
	})
}
