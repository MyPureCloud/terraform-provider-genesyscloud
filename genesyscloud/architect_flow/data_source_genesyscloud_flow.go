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
)

func dataSourceFlowRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sdkConfig := m.(*provider.ProviderMeta).ClientConfig
	p := getArchitectFlowProxy(sdkConfig)

	name := d.Get("name").(string)

	// Query flow by name. Retry in case search has not yet indexed the flow.
	return util.WithRetries(ctx, 5*time.Second, func() *retry.RetryError {
		const pageSize = 100
		for pageNum := 1; ; pageNum++ {
			flows, getErr := p.GetAllFlows(ctx)
			if getErr != nil {
				return retry.NonRetryableError(fmt.Errorf("Error requesting flow %s: %s", name, getErr))
			}

			if flows == nil || len(*flows) == 0 {
				return retry.RetryableError(fmt.Errorf("No flows found with name %s", name))
			}

			for _, entity := range *flows {
				if *entity.Name == name {
					d.SetId(*entity.Id)
					return nil
				}
			}
		}
	})
}
