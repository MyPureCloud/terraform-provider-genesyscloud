package architect_schedulegroups

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

/*
   The data_source_genesyscloud_architect_schedulegroups.go contains the data source implementation
   for the resource.
*/

// dataSourceArchitectSchedulegroupsRead retrieves by name the id in question
func dataSourceArchitectSchedulegroupsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := newArchitectSchedulegroupsProxy(sdkConfig)

	name := d.Get("name").(string)

	return util.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		scheduleGroupId, retryable, err := proxy.getArchitectSchedulegroupsIdByName(ctx, name)

		if err != nil && !retryable {
			return retry.NonRetryableError(fmt.Errorf("Error searching architect schedulegroups %s: %s", name, err))
		}

		if retryable {
			return retry.RetryableError(fmt.Errorf("No architect schedulegroups found with name %s", name))
		}

		d.SetId(scheduleGroupId)
		return nil
	})
}
