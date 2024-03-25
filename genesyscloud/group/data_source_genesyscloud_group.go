package group

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

func dataSourceGroupRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sdkConfig := m.(*provider.ProviderMeta).ClientConfig
	gp := getGroupProxy(sdkConfig)

	nameStr := d.Get("name").(string)

	return util.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		groups, _, getErr := gp.getGroupsByName(ctx, nameStr)
		if getErr != nil {
			return retry.NonRetryableError(fmt.Errorf("Error requesting group %s: %s", nameStr, getErr))
		}

		if *groups.Total > 1 {
			return retry.NonRetryableError(fmt.Errorf("Multiple groups found with name %s ", nameStr))
		}

		if *groups.Total == 0 {
			return retry.RetryableError(fmt.Errorf("No groups found with name %s ", nameStr))
		}

		// Select first group in the list
		group := (*groups.Results)[0]
		d.SetId(*group.Id)
		return nil
	})
}
