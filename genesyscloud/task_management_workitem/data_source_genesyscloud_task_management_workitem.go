package task_management_workitem

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
   The data_source_genesyscloud_task_management_workitem.go contains the data source implementation
   for the resource.
*/

// dataSourceTaskManagementWorkitemRead retrieves by name the id in question
func dataSourceTaskManagementWorkitemRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getTaskManagementWorkitemProxy(sdkConfig)

	name := d.Get("name").(string)
	workbinId := d.Get("workbin_id").(string)
	worktypeId := d.Get("worktype_id").(string)

	return util.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		workitemId, retryable, resp, err := proxy.getTaskManagementWorkitemIdByName(ctx, name, workbinId, worktypeId)

		if err != nil && !retryable {
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(resourceName, fmt.Sprintf("error searching task management workitem %s | error: %s", name, err), resp))
		}

		if retryable {
			return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(resourceName, fmt.Sprintf("no task management workitem found with name %s", name), resp))
		}

		d.SetId(workitemId)
		return nil
	})
}
