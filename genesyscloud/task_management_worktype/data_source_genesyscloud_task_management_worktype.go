package task_management_worktype

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
   The data_source_genesyscloud_task_management_worktype.go contains the data source implementation
   for the resource.
*/

// dataSourceTaskManagementWorktypeRead retrieves by name the id in question
func dataSourceTaskManagementWorktypeRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getTaskManagementWorktypeProxy(sdkConfig)

	name := d.Get("name").(string)

	return util.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		worktypeId, retryable, resp, err := proxy.getTaskManagementWorktypeIdByName(ctx, name)

		if err != nil && !retryable {
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(resourceName, fmt.Sprintf("error searching task management worktype %s | error: %s", name, err), resp))
		}

		if retryable {
			return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(resourceName, fmt.Sprintf("no task management worktype found with name %s", name), resp))
		}

		d.SetId(worktypeId)
		return nil
	})
}

func dataSourceTaskManagementWorktypeStatusRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getTaskManagementWorktypeProxy(sdkConfig)

	workTypeName := d.Get("worktype_name").(string)
	statusName := d.Get("worktype_status_name").(string)

	return util.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		//First retrieve the work type id by name
		workType, retryable, resp, err := proxy.getTaskManagementWorktypeByName(ctx, workTypeName)

		if err != nil && !retryable {
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(resourceName, fmt.Sprintf("error searching task management worktype %s | error: %s", workTypeName, err), resp))
		}

		if retryable {
			return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(resourceName, fmt.Sprintf("no task management worktype found with name %s", workTypeName), resp))
		}

		for _, status := range *workType.Statuses {
			if *status.Name == statusName {
				d.SetId(*status.Id)
				return nil
			}
		}

		return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(resourceName, fmt.Sprintf("No records found for management worktype %s with status name %s | error: %s", workTypeName, statusName, err), resp))
	})
}
