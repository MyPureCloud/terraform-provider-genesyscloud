package genesyscloud

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v105/platformclientv2"
)

func dataSourceDidPool() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for Genesys Cloud DID pool. Select a DID pool by starting phone number and ending phone number",
		ReadContext: ReadWithPooledClient(dataSourceDidPoolRead),
		Schema: map[string]*schema.Schema{
			"start_phone_number": {
				Description:      "Starting phone number of the DID Pool range. Must be in an E.164 number format.",
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: ValidatePhoneNumber,
			},
			"end_phone_number": {
				Description:      "Ending phone number of the DID Pool range.",
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: ValidatePhoneNumber,
			},
		},
	}
}

func dataSourceDidPoolRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sdkConfig := m.(*ProviderMeta).ClientConfig
	telephonyAPI := platformclientv2.NewTelephonyProvidersEdgeApiWithConfig(sdkConfig)

	didPoolStartPhoneNumber := d.Get("start_phone_number").(string)
	didPoolEndPhoneNumber := d.Get("end_phone_number").(string)

	return WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		for pageNum := 1; ; pageNum++ {
			const pageSize = 100
			didPools, _, getErr := telephonyAPI.GetTelephonyProvidersEdgesDidpools(pageSize, pageNum, "", nil)

			if getErr != nil {
				return retry.NonRetryableError(fmt.Errorf("error requesting list of DID pools: %s", getErr))
			}

			if didPools.Entities == nil || len(*didPools.Entities) == 0 {
				return retry.RetryableError(fmt.Errorf("no DID pools found with start phone number: %s and end phone number: %s", didPoolStartPhoneNumber, didPoolEndPhoneNumber))
			}

			for _, didPool := range *didPools.Entities {
				if didPool.StartPhoneNumber != nil && *didPool.StartPhoneNumber == didPoolStartPhoneNumber &&
					didPool.EndPhoneNumber != nil && *didPool.EndPhoneNumber == didPoolEndPhoneNumber &&
					didPool.State != nil && *didPool.State != "deleted" {
					d.SetId(*didPool.Id)
					return nil
				}
			}

		}
	})

}
