package genesyscloud

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v115/platformclientv2"
)

func dataSourceExtensionPool() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for Genesys Cloud Extension pool. Select an Extension pool by starting number and ending number",
		ReadContext: ReadWithPooledClient(dataSourceExtensionPoolRead),
		Schema: map[string]*schema.Schema{
			"start_number": {
				Description:      "Starting number of the Extension Pool range.",
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: validateExtensionPool,
			},
			"end_number": {
				Description:      "Ending number of the Extension Pool range.",
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: validateExtensionPool,
			},
		},
	}
}

func dataSourceExtensionPoolRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sdkConfig := m.(*ProviderMeta).ClientConfig
	telephonyAPI := platformclientv2.NewTelephonyProvidersEdgeApiWithConfig(sdkConfig)

	extensionPoolStartPhoneNumber := d.Get("start_number").(string)
	extensionPoolEndPhoneNumber := d.Get("end_number").(string)

	return WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
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
