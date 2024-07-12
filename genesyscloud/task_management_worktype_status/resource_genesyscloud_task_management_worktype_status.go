package task_management_worktype_status

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
	"log"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	"terraform-provider-genesyscloud/genesyscloud/util/lists"
	"time"

	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"

	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"terraform-provider-genesyscloud/genesyscloud/util/constants"
	"terraform-provider-genesyscloud/genesyscloud/util/resourcedata"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

/*
The resource_genesyscloud_task_management_worktype_status.go contains all of the methods that perform the core logic for a resource.
*/

// getAllAuthTaskManagementWorktypeStatus retrieves all of the task management worktype status via Terraform in the Genesys Cloud and is used for the exporter
func getAllAuthTaskManagementWorktypeStatuss(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	proxy := newTaskManagementWorktypeStatusProxy(clientConfig)
	resources := make(resourceExporter.ResourceIDMetaMap)

	workitemStatuss, resp, err := proxy.getAllTaskManagementWorktypeStatus(ctx)
	if err != nil {
		return nil, util.BuildAPIDiagnosticError(resourceName, fmt.Sprintf("Failed to get task management worktype statuses: %v", err), resp)
	}

	for _, workitemStatus := range *workitemStatuss {
		resources[*workitemStatus.Id] = &resourceExporter.ResourceMeta{Name: *workitemStatus.Name}
	}

	return resources, nil
}

// createTaskManagementWorktypeStatus is used by the task_management_worktype_status resource to create Genesys cloud task management worktype status
func createTaskManagementWorktypeStatus(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getTaskManagementWorktypeStatusProxy(sdkConfig)
	worktypeId := d.Get("worktype_id").(string)

	taskManagementWorktypeStatus := platformclientv2.Workitemstatuscreate{
		Name:                         platformclientv2.String(d.Get("name").(string)),
		Category:                     platformclientv2.String(d.Get("category").(string)),
		Description:                  resourcedata.GetNillableValue[string](d, "description"),
		DestinationStatusIds:         lists.BuildSdkStringListFromInterfaceArray(d, "default_destination_status_id"),
		DefaultDestinationStatusId:   resourcedata.GetNillableValue[string](d, "default_destination_status_id"),
		StatusTransitionDelaySeconds: resourcedata.GetNillableValue[int](d, "status_transition_delay_seconds"),
		StatusTransitionTime:         resourcedata.GetNillableValue[string](d, "status_transition_time"),
	}

	log.Printf("Creating task management worktype %s status %s", worktypeId, *taskManagementWorktypeStatus.Name)
	workitemStatus, resp, err := proxy.createTaskManagementWorktypeStatus(ctx, worktypeId, &taskManagementWorktypeStatus)
	if err != nil {
		return util.BuildAPIDiagnosticError(resourceName, fmt.Sprintf("Failed to create task management worktype %s status %s: %s", worktypeId, *taskManagementWorktypeStatus.Name, err), resp)
	}

	d.SetId(*workitemStatus.Id)
	log.Printf("Created task management worktype %s status %s %s", worktypeId, *workitemStatus.Id, *workitemStatus.Name)
	return readTaskManagementWorktypeStatus(ctx, d, meta)
}

// readTaskManagementWorktypeStatus is used by the task_management_worktype_status resource to read an task management worktype status from genesys cloud
func readTaskManagementWorktypeStatus(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getTaskManagementWorktypeStatusProxy(sdkConfig)
	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceTaskManagementWorktypeStatus(), constants.DefaultConsistencyChecks, resourceName)
	worktypeId := d.Get("worktype_id").(string)

	log.Printf("Reading task management worktype %s status %s", worktypeId, d.Id())

	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		workitemStatus, resp, getErr := proxy.getTaskManagementWorktypeStatusById(ctx, worktypeId, d.Id())
		if getErr != nil {
			if util.IsStatus404(resp) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(resourceName, fmt.Sprintf("Failed to read task management worktype %s status %s: %s", worktypeId, d.Id(), getErr), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(resourceName, fmt.Sprintf("Failed to read task management worktype %s status %s: %s", worktypeId, d.Id(), getErr), resp))
		}

		resourcedata.SetNillableValue(d, "worktype_id", workitemStatus.Worktype.Id)
		resourcedata.SetNillableValue(d, "name", workitemStatus.Name)
		resourcedata.SetNillableValue(d, "category", workitemStatus.Category)
		if workitemStatus.DestinationStatuses != nil {
			destinationStatuses := make([]interface{}, len(*workitemStatus.DestinationStatuses))
			for i, v := range *workitemStatus.DestinationStatuses {
				destinationStatuses[i] = *v.Id
			}
			_ = d.Set("destination_status_ids", destinationStatuses)
		}
		resourcedata.SetNillableValue(d, "description", workitemStatus.Description)
		if workitemStatus.DefaultDestinationStatus != nil && workitemStatus.DefaultDestinationStatus.Id != nil {
			_ = d.Set("default_destination_status_id", *workitemStatus.DefaultDestinationStatus.Id)
		}
		resourcedata.SetNillableValue(d, "status_transition_delay_seconds", workitemStatus.StatusTransitionDelaySeconds)
		resourcedata.SetNillableValue(d, "status_transition_time", workitemStatus.StatusTransitionTime)

		log.Printf("Read task management worktype %s status %s %s", worktypeId, d.Id(), *workitemStatus.Name)
		return cc.CheckState(d)
	})
}

// updateTaskManagementWorktypeStatus is used by the task_management_worktype_status resource to update an task management worktype status in Genesys Cloud
func updateTaskManagementWorktypeStatus(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getTaskManagementWorktypeStatusProxy(sdkConfig)
	worktypeId := d.Get("worktype_id").(string)

	taskManagementWorktypeStatus := platformclientv2.Workitemstatusupdate{
		Name:                         platformclientv2.String(d.Get("name").(string)),
		Description:                  resourcedata.GetNillableValue[string](d, "description"),
		DestinationStatusIds:         lists.BuildSdkStringListFromInterfaceArray(d, "default_destination_status_id"),
		DefaultDestinationStatusId:   resourcedata.GetNillableValue[string](d, "default_destination_status_id"),
		StatusTransitionDelaySeconds: resourcedata.GetNillableValue[int](d, "status_transition_delay_seconds"),
		StatusTransitionTime:         resourcedata.GetNillableValue[string](d, "status_transition_time"),
	}

	log.Printf("Updating task management worktype %s status %s %s", worktypeId, d.Id(), *taskManagementWorktypeStatus.Name)
	workitemStatus, resp, err := proxy.updateTaskManagementWorktypeStatus(ctx, worktypeId, d.Id(), &taskManagementWorktypeStatus)
	if err != nil {
		return util.BuildAPIDiagnosticError(resourceName, fmt.Sprintf("Failed to update task management worktype %s status %s: %s", worktypeId, d.Id(), err), resp)
	}

	log.Printf("Updated task management worktype %s status %s %s", worktypeId, *workitemStatus.Id, *workitemStatus.Id)
	return readTaskManagementWorktypeStatus(ctx, d, meta)
}

// deleteTaskManagementWorktypeStatus is used by the task_management_worktype_status resource to delete an task management worktype status from Genesys cloud
func deleteTaskManagementWorktypeStatus(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getTaskManagementWorktypeStatusProxy(sdkConfig)
	worktypeId := d.Get("worktype_id").(string)

	resp, err := proxy.deleteTaskManagementWorktypeStatus(ctx, worktypeId, d.Id())
	if err != nil {
		return util.BuildAPIDiagnosticError(resourceName, fmt.Sprintf("Failed to delete task management worktype %s status %s: %s", worktypeId, d.Id(), err), resp)
	}

	return util.WithRetries(ctx, 180*time.Second, func() *retry.RetryError {
		_, resp, err := proxy.getTaskManagementWorktypeStatusById(ctx, worktypeId, d.Id())

		if err != nil {
			if util.IsStatus404(resp) {
				log.Printf("Deleted task management worktype %s status %s", worktypeId, d.Id())
				return nil
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(resourceName, fmt.Sprintf("Error deleting task management worktype %s status %s: %s", worktypeId, d.Id(), err), resp))
		}

		return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(resourceName, fmt.Sprintf("task management worktype %s status %s still exists", worktypeId, d.Id()), resp))
	})
}
