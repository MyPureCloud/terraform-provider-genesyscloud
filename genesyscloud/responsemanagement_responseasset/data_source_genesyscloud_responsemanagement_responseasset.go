package responsemanagement_responseasset

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	gcloud "terraform-provider-genesyscloud/genesyscloud"
	"time"
)

func dataSourceResponseManagementResponseAssetRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sdkConfig := m.(*gcloud.ProviderMeta).ClientConfig
	proxy := getRespManagementRespAssetProxy(sdkConfig)

	name := d.Get("name").(string)
	return gcloud.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		responseId, retryable, err := proxy.getRespManagementRespAssetByName(ctx, name)

		if err != nil && !retryable {
			return retry.NonRetryableError(fmt.Errorf("Error searching responsemanagement response asset %s: %s", name, err))
		}
		if retryable {
			return retry.RetryableError(fmt.Errorf("No responsemanagement response asset found with name %s", name))
		}

		d.SetId(responseId)
		return nil
	})
}
