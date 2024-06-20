package task_management_worktype

import (
	"context"
	"fmt"
	"log"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"terraform-provider-genesyscloud/genesyscloud/util/constants"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v131/platformclientv2"

	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"

	"terraform-provider-genesyscloud/genesyscloud/util/lists"
	"terraform-provider-genesyscloud/genesyscloud/util/resourcedata"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

/*
The resource_genesyscloud_task_management_worktype.go contains all of the methods that perform the core logic for a resource.
*/

// getAllAuthTaskManagementWorktype retrieves all of the task management worktype via Terraform in the Genesys Cloud and is used for the exporter
func getAllAuthTaskManagementWorktypes(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	proxy := getTaskManagementWorktypeProxy(clientConfig)
	resources := make(resourceExporter.ResourceIDMetaMap)

	worktypes, resp, err := proxy.getAllTaskManagementWorktype(ctx)
	if err != nil {
		return nil, util.BuildAPIDiagnosticError(resourceName, fmt.Sprintf("Failed to get task management worktype error: %s", err), resp)
	}

	for _, worktype := range *worktypes {
		resources[*worktype.Id] = &resourceExporter.ResourceMeta{Name: *worktype.Name}
	}
	return resources, nil
}

// createTaskManagementWorktype is used by the task_management_worktype resource to create Genesys cloud task management worktype
func createTaskManagementWorktype(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getTaskManagementWorktypeProxy(sdkConfig)

	taskManagementWorktype := getWorktypecreateFromResourceData(d)

	// Create the base worktype
	log.Printf("Creating task management worktype %s", *taskManagementWorktype.Name)
	worktype, resp, err := proxy.createTaskManagementWorktype(ctx, &taskManagementWorktype)
	if err != nil {
		return util.BuildAPIDiagnosticError(resourceName, fmt.Sprintf("Failed to create task management worktype %s error: %s", *taskManagementWorktype.Name, err), resp)
	}

	log.Printf("Created the base task management worktype %s", *worktype.Id)
	d.SetId(*worktype.Id)

	// Create and update (for referencing other status) the worktype statuses
	log.Printf("Creating the task management worktype statuses of %s", *worktype.Id)
	statuses := d.Get("statuses").(*schema.Set).List()
	if _, err := createWorktypeStatuses(ctx, proxy, *worktype.Id, statuses); err != nil {
		return util.BuildDiagnosticError(resourceName, fmt.Sprintf("failed to create task management worktype statuses"), err)
	}
	log.Printf("Updating the destination statuses of the statuses of worktype %s", *worktype.Id)
	if _, err := updateWorktypeStatuses(ctx, proxy, *worktype.Id, statuses, true); err != nil {
		return util.BuildDiagnosticError(resourceName, fmt.Sprintf("failed to update task management worktype statuses"), err)
	}

	// Update the worktype if 'default_status_name' is set
	if d.HasChange("default_status_name") {
		time.Sleep(5 * time.Second)
		err := updateDefaultStatusName(ctx, proxy, d, *worktype.Id)
		if err != nil {
			return util.BuildDiagnosticError(resourceName, fmt.Sprintf("failed to update default status name of worktype"), err)
		}
	}

	log.Printf("Created the task management worktype statuses of %s", *worktype.Id)
	return readTaskManagementWorktype(ctx, d, meta)
}

// readTaskManagementWorktype is used by the task_management_worktype resource to read a task management worktype from genesys cloud
func readTaskManagementWorktype(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getTaskManagementWorktypeProxy(sdkConfig)
	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceTaskManagementWorktype(), constants.DefaultConsistencyChecks, resourceName)

	log.Printf("Reading task management worktype %s", d.Id())

	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		worktype, resp, getErr := proxy.getTaskManagementWorktypeById(ctx, d.Id())
		if getErr != nil {
			if util.IsStatus404(resp) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(resourceName, fmt.Sprintf("failed to read task management worktype %s | error: %s", d.Id(), getErr), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(resourceName, fmt.Sprintf("failed to read task management worktype %s | error: %s", d.Id(), getErr), resp))
		}

		resourcedata.SetNillableValue(d, "name", worktype.Name)
		resourcedata.SetNillableValue(d, "description", worktype.Description)
		resourcedata.SetNillableReferenceDivision(d, "division_id", worktype.Division)

		if worktype.DefaultWorkbin != nil {
			resourcedata.SetNillableValue(d, "default_workbin_id", worktype.DefaultWorkbin.Id)
		}

		// Default status can be an empty object
		if worktype.DefaultStatus != nil && worktype.DefaultStatus.Id != nil {
			if statusName := getStatusNameFromId(*worktype.DefaultStatus.Id, worktype.Statuses); statusName != nil {
				d.Set("default_status_name", statusName)
			}
		}
		resourcedata.SetNillableValueWithInterfaceArrayWithFunc(d, "statuses", worktype.Statuses, flattenWorkitemStatuses)

		resourcedata.SetNillableValue(d, "default_duration_seconds", worktype.DefaultDurationSeconds)
		resourcedata.SetNillableValue(d, "default_expiration_seconds", worktype.DefaultExpirationSeconds)
		resourcedata.SetNillableValue(d, "default_due_duration_seconds", worktype.DefaultDueDurationSeconds)
		resourcedata.SetNillableValue(d, "default_priority", worktype.DefaultPriority)
		resourcedata.SetNillableValue(d, "default_ttl_seconds", worktype.DefaultTtlSeconds)

		if worktype.DefaultLanguage != nil {
			resourcedata.SetNillableValue(d, "default_language_id", worktype.DefaultLanguage.Id)
		}
		if worktype.DefaultQueue != nil {
			resourcedata.SetNillableValue(d, "default_queue_id", worktype.DefaultQueue.Id)
		}

		resourcedata.SetNillableValueWithInterfaceArrayWithFunc(d, "default_skills_ids", worktype.DefaultSkills, flattenRoutingSkillReferences)
		resourcedata.SetNillableValue(d, "assignment_enabled", worktype.AssignmentEnabled)

		if worktype.Schema != nil {
			resourcedata.SetNillableValue(d, "schema_id", worktype.Schema.Id)
			resourcedata.SetNillableValue(d, "schema_version", worktype.Schema.Version)
		}

		log.Printf("Read task management worktype %s %s", d.Id(), *worktype.Name)
		return cc.CheckState(d)
	})
}

// updateTaskManagementWorktype is used by the task_management_worktype resource to update a task management worktype in Genesys Cloud
func updateTaskManagementWorktype(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getTaskManagementWorktypeProxy(sdkConfig)

	// Update the base configuration of the Worktype
	taskManagementWorktype := getWorktypeupdateFromResourceData(d, nil)
	if d.HasChangesExcept("statuses", "default_status_name") {
		worktype, resp, err := proxy.updateTaskManagementWorktype(ctx, d.Id(), &taskManagementWorktype)
		if err != nil {
			return util.BuildAPIDiagnosticError(resourceName, fmt.Sprintf("Failed to update task management worktype %s error: %s", *taskManagementWorktype.Name, err), resp)
		}
		log.Printf("Updated base configuration of task management worktype %s", *worktype.Id)
	}

	// Get the current state of the worktype because we will cross-check if any of the existing ones
	// need to be deleted
	oldWorktype, resp, err := proxy.getTaskManagementWorktypeById(ctx, d.Id())
	if err != nil {
		return util.BuildAPIDiagnosticError(resourceName, fmt.Sprintf("Failed to get task management worktype %s error: %s", d.Id(), err), resp)
	}
	oldStatusIds := []string{}
	for _, oldStatus := range *oldWorktype.Statuses {
		oldStatusIds = append(oldStatusIds, *oldStatus.Id)
	}

	// We'll use this to keep track of the actual status ids as a result of the worktype update.
	// Any ids not here will be deleted eventually as the last step.
	statusIdsToStay := []string{}

	statuses := d.Get("statuses").(*schema.Set).List()

	// If the status in the state still has an id that means there is no update to it and so should not be included
	// in the API update (GC API gives error if update is called but actually no diff).
	forUpdateOrCreation := make([]interface{}, 0)
	for _, status := range statuses {
		statusMap := status.(map[string]interface{})
		if statusMap["id"] == "" {
			forUpdateOrCreation = append(forUpdateOrCreation, status)
		} else {
			statusIdsToStay = append(statusIdsToStay, statusMap["id"].(string))
		}
	}
	forCreation, forUpdate := getStatusesForUpdateAndCreation(forUpdateOrCreation, oldWorktype.Statuses)

	// Create new statuses
	log.Printf("Creating the task management worktype statuses of %s", d.Id())
	if _, err := createWorktypeStatuses(ctx, proxy, d.Id(), forCreation); err != nil {
		return util.BuildDiagnosticError(resourceName, fmt.Sprintf("failed to create task management worktype statuses"), err)
	}

	// Update the newly created statuses with status refs
	log.Printf("Updating the newly created statuses of worktype %s", d.Id())
	createdStatuses, err := updateWorktypeStatuses(ctx, proxy, d.Id(), forCreation, true)
	if err != nil {
		return util.BuildDiagnosticError(resourceName, fmt.Sprintf("failed to update task management worktype statuses"), err)
	}
	for _, updateStat := range *createdStatuses {
		statusIdsToStay = append(statusIdsToStay, *updateStat.Id)
	}

	// Update the other already existing statuses for reference or other property updates
	log.Printf("Updating the destination statuses of the statuses of worktype %s", d.Id())
	updatedStatuses, err := updateWorktypeStatuses(ctx, proxy, d.Id(), forUpdate, false)
	if err != nil {
		return util.BuildDiagnosticError(resourceName, fmt.Sprintf("failed to update task management worktype statuses"), err)
	}
	for _, updateStat := range *updatedStatuses {
		statusIdsToStay = append(statusIdsToStay, *updateStat.Id)
	}

	// Delete statuses that are no longer defined in the configuration
	log.Printf("Cleaning up statuses of worktype %s", d.Id())
	forDeletionIds := lists.SliceDifference(oldStatusIds, statusIdsToStay)

	// Go through and clear the status references first to avoid dependency errors on deletion
	log.Printf("Clearing references of statuses for deletion of worktype %s", d.Id())
	for _, forDeletionId := range forDeletionIds {
		updateForCleaning := platformclientv2.Workitemstatusupdate{}

		// // Force these properties as 'null' for the API request
		updateForCleaning.SetField("DestinationStatusIds", &[]string{})
		updateForCleaning.SetField("DefaultDestinationStatusId", nil)
		updateForCleaning.SetField("StatusTransitionDelaySeconds", nil)
		updateForCleaning.SetField("StatusTransitionTime", nil)

		// We put a random description so we can ensure there is a 'change' in the status.
		// Else we'll get a 400 error if the status has no destination status /default status to begin with
		// This is simpler than checking the status fields if there are any changes.
		// Since this status is for deletion anyway we shouldn't care about this managed update.
		description := "this status is set for deletion by CX as Code " + uuid.NewString()
		updateForCleaning.SetField("Description", &description)

		if _, resp, err := proxy.updateTaskManagementWorktypeStatus(ctx, d.Id(), forDeletionId, &updateForCleaning); err != nil {
			return util.BuildAPIDiagnosticError(resourceName, fmt.Sprintf("Failed to clean up references of task management worktype status %s error: %s", d.Id(), err), resp)
		}
	}

	// Actually delete the status
	log.Printf("Deleting unused statuses of worktype %s", d.Id())
	for _, forDeletionId := range forDeletionIds {
		if resp, err := proxy.deleteTaskManagementWorktypeStatus(ctx, d.Id(), forDeletionId); err != nil {
			return util.BuildAPIDiagnosticError(resourceName, fmt.Sprintf("Failed to delete task management worktype status %s error: %s", forDeletionId, err), resp)
		}
	}

	// Update the worktype if 'default_status_name' is changed
	// We do this last so that the statuses are surely updated first
	if d.HasChange("default_status_name") {
		log.Printf("Updating default status of worktype %s", d.Id())
		time.Sleep(5 * time.Second)
		err := updateDefaultStatusName(ctx, proxy, d, d.Id())
		if err != nil {
			return util.BuildDiagnosticError(resourceName, fmt.Sprintf("failed to update default status name of worktype"), err)
		}
	}

	log.Printf("Finished updating worktype %s", d.Id())

	return readTaskManagementWorktype(ctx, d, meta)
}

// deleteTaskManagementWorktype is used by the task_management_worktype resource to delete a task management worktype from Genesys cloud
func deleteTaskManagementWorktype(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getTaskManagementWorktypeProxy(sdkConfig)

	resp, err := proxy.deleteTaskManagementWorktype(ctx, d.Id())
	if err != nil {
		return util.BuildAPIDiagnosticError(resourceName, fmt.Sprintf("Failed to delete task management worktype %s error: %s", d.Id(), err), resp)
	}

	return util.WithRetries(ctx, 180*time.Second, func() *retry.RetryError {
		_, resp, err := proxy.getTaskManagementWorktypeById(ctx, d.Id())

		if err != nil {
			if util.IsStatus404(resp) {
				log.Printf("Deleted task management worktype %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(resourceName, fmt.Sprintf("error deleting task management worktype %s | error: %s", d.Id(), err), resp))
		}
		return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(resourceName, fmt.Sprintf("task management worktype %s still exists", d.Id()), resp))
	})
}
