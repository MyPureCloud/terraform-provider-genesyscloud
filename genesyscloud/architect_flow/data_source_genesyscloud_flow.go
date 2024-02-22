package architect_flow

import (
	"context"
	"fmt"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v121/platformclientv2"
)

func dataSourceFlowRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sdkConfig := m.(*provider.ProviderMeta).ClientConfig
	archAPI := platformclientv2.NewArchitectApiWithConfig(sdkConfig)

	name := d.Get("name").(string)

	// Query flow by name. Retry in case search has not yet indexed the flow.
	return util.WithRetries(ctx, 5*time.Second, func() *retry.RetryError {
		const pageSize = 100
		for pageNum := 1; ; pageNum++ {
			flows, _, getErr := archAPI.GetFlows(nil, pageNum, pageSize, "", "", nil, name, "", "", "", "", "", "", "", false, false, "", "", nil)
			if getErr != nil {
				return retry.NonRetryableError(fmt.Errorf("Error requesting flow %s: %s", name, getErr))
			}

			if flows.Entities == nil || len(*flows.Entities) == 0 {
				return retry.RetryableError(fmt.Errorf("No flows found with name %s", name))
			}

			for _, entity := range *flows.Entities {
				if *entity.Name == name {
					d.SetId(*entity.Id)
					return nil
				}
			}
		}
	})
}
