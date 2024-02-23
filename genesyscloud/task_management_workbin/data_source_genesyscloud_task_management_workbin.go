package task_management_workbin

import (
	"context"
	"fmt"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

/*
   The data_source_genesyscloud_task_management_workbin.go contains the data source implementation
   for the resource.
*/

// dataSourceTaskManagementWorkbinRead retrieves by name the id in question
func dataSourceTaskManagementWorkbinRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := newTaskManagementWorkbinProxy(sdkConfig)

	name := d.Get("name").(string)

	return util.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		workbinId, retryable, err := proxy.getTaskManagementWorkbinIdByName(ctx, name)

		if err != nil && !retryable {
			return retry.NonRetryableError(fmt.Errorf("error searching task management workbin %s: %s", name, err))
		}

		if retryable {
			return retry.RetryableError(fmt.Errorf("no task management workbin found with name %s", name))
		}

		d.SetId(workbinId)
		return nil
	})
}
