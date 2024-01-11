package telephony_providers_edges_extension_pool

import (
	"context"
	"fmt"
	gcloud "terraform-provider-genesyscloud/genesyscloud"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v119/platformclientv2"
)

func dataSourceExtensionPoolRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sdkConfig := m.(*gcloud.ProviderMeta).ClientConfig
	telephonyAPI := platformclientv2.NewTelephonyProvidersEdgeApiWithConfig(sdkConfig)

	extensionPoolStartPhoneNumber := d.Get("start_number").(string)
	extensionPoolEndPhoneNumber := d.Get("end_number").(string)

	return gcloud.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		for pageNum := 1; ; pageNum++ {
			const pageSize = 100
			extensionPools, _, getErr := telephonyAPI.GetTelephonyProvidersEdgesExtensionpools(pageSize, pageNum, "", "")

			if getErr != nil {
				return retry.NonRetryableError(fmt.Errorf("error requesting list of extension pools: %s", getErr))
			}

			if extensionPools.Entities == nil || len(*extensionPools.Entities) == 0 {
				return retry.RetryableError(fmt.Errorf("no extension pools found with start phone number: %s and end phone number: %s", extensionPoolStartPhoneNumber, extensionPoolEndPhoneNumber))
			}

			for _, extensionPool := range *extensionPools.Entities {
				if extensionPool.StartNumber != nil && *extensionPool.StartNumber == extensionPoolStartPhoneNumber &&
					extensionPool.EndNumber != nil && *extensionPool.EndNumber == extensionPoolEndPhoneNumber &&
					extensionPool.State != nil && *extensionPool.State != "deleted" {
					d.SetId(*extensionPool.Id)
					return nil
				}
			}

		}
	})

}
