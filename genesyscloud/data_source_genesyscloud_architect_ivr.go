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

func DataSourceArchitectIvr() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for Genesys Cloud IVRs. Select an IVR by name.",
		ReadContext: ReadWithPooledClient(dataSourceIvrRead),
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "IVR name.",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}

func dataSourceIvrRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sdkConfig := m.(*ProviderMeta).ClientConfig
	archAPI := platformclientv2.NewArchitectApiWithConfig(sdkConfig)

	name := d.Get("name").(string)

	// Query ivr by name. Retry in case search has not yet indexed the ivr.
	return WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		const pageNum = 1
		const pageSize = 100
		ivrs, _, getErr := archAPI.GetArchitectIvrs(pageNum, pageSize, "", "", name, "", "")
		if getErr != nil {
			return retry.NonRetryableError(fmt.Errorf("Error requesting IVR %s: %s", name, getErr))
		}

		if ivrs.Entities == nil || len(*ivrs.Entities) == 0 {
			return retry.RetryableError(fmt.Errorf("No IVRs found with name %s", name))
		}

		ivr := (*ivrs.Entities)[0]
		d.SetId(*ivr.Id)
		return nil
	})
}
