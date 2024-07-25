package task_management_worktype_status

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
)

/*
   The data_source_genesyscloud_task_management_worktype_status.go contains the data source implementation
   for the resource.
*/

// dataSourceTaskManagementWorktypeStatusRead retrieves by name the id in question
func dataSourceTaskManagementWorktypeStatusRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getTaskManagementWorktypeStatusProxy(sdkConfig)

	worktypeId := d.Get("worktype_id").(string)
	name := d.Get("name").(string)

	return util.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		worktypeStatusId, resp, retryable, err := proxy.getTaskManagementWorktypeStatusIdByName(ctx, worktypeId, name)

		if err != nil && !retryable {
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(resourceName, fmt.Sprintf("Error searching task management worktype %s status %s | error: %s", worktypeId, name, err), resp))
		}

		if retryable {
			return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(resourceName, fmt.Sprintf("No task management worktype %s status found with name %s", worktypeId, name), resp))
		}

		d.SetId(worktypeStatusId)
		return nil
	})
}
