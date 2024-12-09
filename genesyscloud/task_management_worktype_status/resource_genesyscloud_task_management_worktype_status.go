package task_management_worktype_status

import (
	"context"
	"fmt"
	"log"
	"strings"
	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	"terraform-provider-genesyscloud/genesyscloud/util/constants"
	"terraform-provider-genesyscloud/genesyscloud/util/lists"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v146/platformclientv2"

	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"terraform-provider-genesyscloud/genesyscloud/util/resourcedata"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

/*
	NOTE: This resource's Id is in the format <worktypeId>/<statusId> so we can persist the id of the parent worktype.
	The worktype_id field can not be used for this because attribute values are dropped during a read so they can be
	re-read.
*/

/*
The resource_genesyscloud_task_management_worktype_status.go contains all of the methods that perform the core logic for a resource.
*/

// getAllAuthTaskManagementWorktypeStatus retrieves all of the task management worktype status via Terraform in the Genesys Cloud and is used for the exporter
func getAllAuthTaskManagementWorktypeStatuss(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	proxy := getTaskManagementWorktypeStatusProxy(clientConfig)
	resources := make(resourceExporter.ResourceIDMetaMap)

	worktypes, resp, err := proxy.worktypeProxy.GetAllTaskManagementWorktype(ctx)
	if err != nil {
		return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get task management worktypes: %v", err), resp)
	}

	for _, worktype := range *worktypes {
		worktypeStatuses, resp, err := proxy.getAllTaskManagementWorktypeStatus(ctx, *worktype.Id)
		if err != nil {
			return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get task management worktype statuses: %v", err), resp)
		}

		for _, status := range *worktypeStatuses {
			resources[*worktype.Id+"/"+*status.Id] = &resourceExporter.ResourceMeta{BlockLabel: *status.Name}
		}
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
		DestinationStatusIds:         lists.BuildSdkStringListFromInterfaceArray(d, "destination_status_ids"),
		DefaultDestinationStatusId:   resourcedata.GetNillableValue[string](d, "default_destination_status_id"),
		StatusTransitionDelaySeconds: resourcedata.GetNillableValue[int](d, "status_transition_delay_seconds"),
		StatusTransitionTime:         resourcedata.GetNillableValue[string](d, "status_transition_time"),
	}

	err := validateSchema(d)
	if err != nil {
		errorMsg := fmt.Sprintf("Failed to create task management worktype %s status %s: %s", worktypeId, *taskManagementWorktypeStatus.Name, err)
		return util.BuildDiagnosticError(ResourceType, errorMsg, fmt.Errorf(errorMsg))
	}

	// If the user makes a reference to a status that is managed by terraform the id will look like this <worktypeId>/<statusId>
	// so we need to extract just the status id from any status references that look like this
	if taskManagementWorktypeStatus.DestinationStatusIds != nil && len(*taskManagementWorktypeStatus.DestinationStatusIds) > 0 {
		for i, destinationStatusId := range *taskManagementWorktypeStatus.DestinationStatusIds {
			if strings.Contains(destinationStatusId, "/") {
				_, id := SplitWorktypeStatusTerraformId(destinationStatusId)
				(*taskManagementWorktypeStatus.DestinationStatusIds)[i] = id
			}
		}
	}

	if taskManagementWorktypeStatus.DefaultDestinationStatusId != nil && strings.Contains(*taskManagementWorktypeStatus.DefaultDestinationStatusId, "/") {
		_, id := SplitWorktypeStatusTerraformId(*taskManagementWorktypeStatus.DefaultDestinationStatusId)
		taskManagementWorktypeStatus.DefaultDestinationStatusId = &id
	}

	log.Printf("Creating task management worktype %s status %s", worktypeId, *taskManagementWorktypeStatus.Name)
	var (
		workitemStatus *platformclientv2.Workitemstatus
		resp           *platformclientv2.APIResponse
	)

	diagErr := util.WithRetries(ctx, 60*time.Second, func() *retry.RetryError {
		workitemStatus, resp, err = proxy.createTaskManagementWorktypeStatus(ctx, worktypeId, &taskManagementWorktypeStatus)
		if err != nil {
			// The api can throw a 400 if we operate on statuses asynchronously. Retry if we encounter this
			if util.IsStatus400(resp) && strings.Contains(resp.ErrorMessage, "Database transaction was cancelled") {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to create task management worktype %s status %s: %s", worktypeId, *taskManagementWorktypeStatus.Name, err), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to update task management worktype %s status %s: %s", worktypeId, *taskManagementWorktypeStatus.Name, err), resp))
		}
		return nil
	})
	if diagErr != nil {
		return diagErr
	}

	d.SetId(worktypeId + "/" + *workitemStatus.Id)

	// Check if we need to set this status as the default status on the worktype
	if d.Get("default").(bool) {
		log.Printf("Setting status %s as default for worktype %s", *workitemStatus.Id, worktypeId)
		if diagErr := updateWorktypeDefaultStatus(ctx, proxy, worktypeId, *workitemStatus.Id); diagErr != nil {
			return diagErr
		}
		log.Printf("Status %s set as default for worktype %s", *workitemStatus.Id, worktypeId)
	}

	log.Printf("Created task management worktype %s status %s %s", worktypeId, *workitemStatus.Id, *workitemStatus.Name)
	return readTaskManagementWorktypeStatus(ctx, d, meta)
}

// readTaskManagementWorktypeStatus is used by the task_management_worktype_status resource to read an task management worktype status from genesys cloud
func readTaskManagementWorktypeStatus(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getTaskManagementWorktypeStatusProxy(sdkConfig)
	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceTaskManagementWorktypeStatus(), constants.ConsistencyChecks(), ResourceType)
	worktypeId, statusId := SplitWorktypeStatusTerraformId(d.Id())

	log.Printf("Reading task management worktype %s status %s", worktypeId, statusId)

	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		workitemStatus, resp, getErr := proxy.getTaskManagementWorktypeStatusById(ctx, worktypeId, statusId)
		if getErr != nil {
			if util.IsStatus404(resp) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read task management worktype %s status %s: %s", worktypeId, statusId, getErr), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read task management worktype %s status %s: %s", worktypeId, statusId, getErr), resp))
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
		} else {
			_ = d.Set("default_destination_status_id", "")
		}
		resourcedata.SetNillableValue(d, "status_transition_delay_seconds", workitemStatus.StatusTransitionDelaySeconds)
		resourcedata.SetNillableValue(d, "status_transition_time", workitemStatus.StatusTransitionTime)

		// Check if this status is the default on the worktype
		worktype, resp, err := proxy.worktypeProxy.GetTaskManagementWorktypeById(ctx, worktypeId)
		if err != nil {
			return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("failed to read worktype %s", worktypeId), resp))
		}

		_ = d.Set("default", false)
		if worktype.DefaultStatus != nil && worktype.DefaultStatus.Id != nil && *worktype.DefaultStatus.Id == statusId {
			_ = d.Set("default", true)
		}

		log.Printf("Read task management worktype %s status %s %s", worktypeId, statusId, *workitemStatus.Name)
		return cc.CheckState(d)
	})
}

// updateTaskManagementWorktypeStatus is used by the task_management_worktype_status resource to update an task management worktype status in Genesys Cloud
func updateTaskManagementWorktypeStatus(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getTaskManagementWorktypeStatusProxy(sdkConfig)
	worktypeId, statusId := SplitWorktypeStatusTerraformId(d.Id())

	err := validateSchema(d)
	if err != nil {
		errorMsg := fmt.Sprintf("Failed to update task management worktype %s status %s: %s", worktypeId, statusId, err)
		return util.BuildDiagnosticError(ResourceType, errorMsg, fmt.Errorf(errorMsg))
	}

	taskManagementWorktypeStatus := platformclientv2.Workitemstatusupdate{
		Name:                         platformclientv2.String(d.Get("name").(string)),
		Description:                  resourcedata.GetNillableValue[string](d, "description"),
		DestinationStatusIds:         lists.BuildSdkStringListFromInterfaceArray(d, "destination_status_ids"),
		DefaultDestinationStatusId:   resourcedata.GetNillableValue[string](d, "default_destination_status_id"),
		StatusTransitionDelaySeconds: resourcedata.GetNillableValue[int](d, "status_transition_delay_seconds"),
		StatusTransitionTime:         resourcedata.GetNillableValue[string](d, "status_transition_time"),
	}

	// If the user makes a reference to a status that is managed by terraform the id will look like this <worktypeId>/<statusId>
	// so we need to extract just the status id from any status references that look like this
	if taskManagementWorktypeStatus.DestinationStatusIds != nil && len(*taskManagementWorktypeStatus.DestinationStatusIds) > 0 {
		for i, destinationStatusId := range *taskManagementWorktypeStatus.DestinationStatusIds {
			if strings.Contains(destinationStatusId, "/") {
				_, id := SplitWorktypeStatusTerraformId(destinationStatusId)
				(*taskManagementWorktypeStatus.DestinationStatusIds)[i] = id
			}
		}
	}

	if taskManagementWorktypeStatus.DefaultDestinationStatusId != nil && strings.Contains(*taskManagementWorktypeStatus.DefaultDestinationStatusId, "/") {
		_, id := SplitWorktypeStatusTerraformId(*taskManagementWorktypeStatus.DefaultDestinationStatusId)
		taskManagementWorktypeStatus.DefaultDestinationStatusId = &id
	}

	log.Printf("Updating task management worktype %s status %s %s", worktypeId, statusId, *taskManagementWorktypeStatus.Name)

	var (
		workitemStatus *platformclientv2.Workitemstatus
		resp           *platformclientv2.APIResponse
	)
	diagErr := util.WithRetries(ctx, 60*time.Second, func() *retry.RetryError {
		workitemStatus, resp, err = proxy.updateTaskManagementWorktypeStatus(ctx, worktypeId, statusId, &taskManagementWorktypeStatus)
		if err != nil {
			// The api can throw a 400 if we operate on statuses asynchronously. Retry if we encounter this
			if util.IsStatus400(resp) && strings.Contains(resp.ErrorMessage, "Database transaction was cancelled") {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to update task management worktype %s status %s: %s", worktypeId, statusId, err), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to update task management worktype %s status %s: %s", worktypeId, statusId, err), resp))
		}
		return nil
	})
	if diagErr != nil {
		return diagErr
	}

	// Check if we need to set this status as the default status on the worktype
	if d.Get("default").(bool) {
		log.Printf("Setting status %s as default for worktype %s", statusId, worktypeId)
		if diagErr := updateWorktypeDefaultStatus(ctx, proxy, worktypeId, *workitemStatus.Id); diagErr != nil {
			return diagErr
		}
		log.Printf("Status %s set as default for worktype %s", statusId, worktypeId)
	}

	log.Printf("Updated task management worktype %s status %s %s", worktypeId, *workitemStatus.Id, *workitemStatus.Id)
	return readTaskManagementWorktypeStatus(ctx, d, meta)
}

// deleteTaskManagementWorktypeStatus is used by the task_management_worktype_status resource to delete an task management worktype status from Genesys cloud
func deleteTaskManagementWorktypeStatus(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getTaskManagementWorktypeStatusProxy(sdkConfig)
	worktypeId, statusId := SplitWorktypeStatusTerraformId(d.Id())

	// Check if worktype exists before trying to check the status. If the worktype is gone then so it the status
	_, resp, err := proxy.getTaskManagementWorktypeStatusById(ctx, worktypeId, statusId)
	if err != nil {
		if util.IsStatus404(resp) {
			log.Printf("Task management worktype %s already deleted", worktypeId)
			return nil
		}
	}

	// Can't delete the status if it's the default on the worktype
	if d.Get("default").(bool) {
		log.Printf("Unable to delete default status %s  for worktype %s", statusId, worktypeId)
		return nil
	}

	log.Printf("Deleting task management worktype %s status %s", worktypeId, statusId)

	diagErr := util.WithRetries(ctx, 60*time.Second, func() *retry.RetryError {
		resp, err = proxy.deleteTaskManagementWorktypeStatus(ctx, worktypeId, statusId)
		if err != nil {
			// The api can throw a 400 if we operate on statuses asynchronously. Retry if we encounter this
			if util.IsStatus400(resp) && strings.Contains(resp.ErrorMessage, "Database transaction was cancelled") {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to delete task management worktype %s status %s: %s", worktypeId, statusId, err), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to delete task management worktype %s status %s: %s", worktypeId, statusId, err), resp))
		}
		return nil
	})
	if diagErr != nil {
		return diagErr
	}

	return util.WithRetries(ctx, 180*time.Second, func() *retry.RetryError {
		_, resp, err = proxy.getTaskManagementWorktypeStatusById(ctx, worktypeId, statusId)
		if err != nil {
			if util.IsStatus404(resp) {
				log.Printf("Deleted task management worktype %s status %s", worktypeId, statusId)
				return nil
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Error deleting task management worktype %s status %s: %s", worktypeId, statusId, err), resp))
		}

		return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("task management worktype %s status %s still exists", worktypeId, statusId), resp))
	})
}
