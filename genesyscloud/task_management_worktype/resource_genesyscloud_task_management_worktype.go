package task_management_worktype

import (
	"context"
	"fmt"
	"log"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v115/platformclientv2"

	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"

	gcloud "terraform-provider-genesyscloud/genesyscloud"

	"terraform-provider-genesyscloud/genesyscloud/util/lists"
	"terraform-provider-genesyscloud/genesyscloud/util/resourcedata"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

/*
The resource_genesyscloud_task_management_worktype.go contains all of the methods that perform the core logic for a resource.
*/

// getAllAuthTaskManagementWorktype retrieves all of the task management worktype via Terraform in the Genesys Cloud and is used for the exporter
func getAllAuthTaskManagementWorktypes(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	proxy := newTaskManagementWorktypeProxy(clientConfig)
	resources := make(resourceExporter.ResourceIDMetaMap)

	worktypes, err := proxy.getAllTaskManagementWorktype(ctx)
	if err != nil {
		return nil, diag.Errorf("Failed to get task management worktype: %v", err)
	}

	for _, worktype := range *worktypes {
		resources[*worktype.Id] = &resourceExporter.ResourceMeta{Name: *worktype.Id}
	}

	return resources, nil
}

// createTaskManagementWorktype is used by the task_management_worktype resource to create Genesys cloud task management worktype
func createTaskManagementWorktype(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	proxy := getTaskManagementWorktypeProxy(sdkConfig)

	taskManagementWorktype := getWorktypecreateFromResourceData(d)

	log.Printf("Creating task management worktype %s", *taskManagementWorktype.Name)
	worktype, err := proxy.createTaskManagementWorktype(ctx, &taskManagementWorktype)
	if err != nil {
		return diag.Errorf("failed to create task management worktype: %s", err)
	}

	log.Printf("Created the base task management worktype %s", *worktype.Id)
	d.SetId(*worktype.Id)

	// Create the worktype statuses
	log.Printf("Creating the task management worktype statuses of %s", *worktype.Id)

	statuses := d.Get("statuses").(*schema.Set).List()
	if _, err := createWorktypeStatuses(ctx, proxy, *worktype.Id, statuses); err != nil {
		return diag.Errorf("failed to create task management worktype statuses: %v", err)
	}
	log.Printf("Updating the destination statuses of the statuses of worktype %s", *worktype.Id)
	if _, err := updateWorktypeStatuses(ctx, proxy, *worktype.Id, statuses, true); err != nil {
		return diag.Errorf("failed to update task management worktype statuses: %v", err)
	}

	// Update the worktype if 'default_status_name' is set
	if d.HasChange("default_status_name") {
		time.Sleep(5 * time.Second)
		err := updateDefaultStatusName(ctx, proxy, d, *worktype.Id)
		if err != nil {
			return diag.Errorf("failed to update default status name of worktype: %v", err)
		}
	}

	log.Printf("Created the task management worktype statuses of %s", *worktype.Id)
	return readTaskManagementWorktype(ctx, d, meta)
}

// readTaskManagementWorktype is used by the task_management_worktype resource to read a task management worktype from genesys cloud
func readTaskManagementWorktype(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	proxy := getTaskManagementWorktypeProxy(sdkConfig)

	log.Printf("Reading task management worktype %s", d.Id())

	return gcloud.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		worktype, respCode, getErr := proxy.getTaskManagementWorktypeById(ctx, d.Id())
		if getErr != nil {
			if gcloud.IsStatus404ByInt(respCode) {
				return retry.RetryableError(fmt.Errorf("failed to read task management worktype %s: %s", d.Id(), getErr))
			}
			return retry.NonRetryableError(fmt.Errorf("failed to read task management worktype %s: %s", d.Id(), getErr))
		}

		cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceTaskManagementWorktype())

		resourcedata.SetNillableValue(d, "name", worktype.Name)
		resourcedata.SetNillableValue(d, "description", worktype.Description)

		if worktype.Division != nil {
			resourcedata.SetNillableValue(d, "division_id", worktype.Division.Id)
		}
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
		return cc.CheckState()
	})
}

// updateTaskManagementWorktype is used by the task_management_worktype resource to update a task management worktype in Genesys Cloud
func updateTaskManagementWorktype(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	proxy := getTaskManagementWorktypeProxy(sdkConfig)

	taskManagementWorktype := getWorktypeupdateFromResourceData(d, nil)
	if d.HasChangesExcept("statuses", "default_status_name") {
		worktype, err := proxy.updateTaskManagementWorktype(ctx, d.Id(), &taskManagementWorktype)
		if err != nil {
			return diag.Errorf("failed to update task management worktype: %s", err)
		}

		log.Printf("Updated base configuration of task management worktype %s", *worktype.Id)
	}

	// Get the current state of the worktype because we will cross-check if any of the existing ones
	// need to be deleted
	oldWorktype, _, err := proxy.getTaskManagementWorktypeById(ctx, d.Id())
	if err != nil {
		return diag.Errorf("failed to get task management worktype: %s", err)
	}
	oldStatusIds := []string{}
	for _, oldStatus := range *oldWorktype.Statuses {
		oldStatusIds = append(oldStatusIds, *oldStatus.Id)
	}

	// We'll use this to keep track of the actual status ids as a result of the worktype update.
	// Any ids not here will be deleted eventually as the last step.
	statusIdsToStay := []string{}

	statuses := d.Get("statuses").(*schema.Set).List()

	// If the status still has an id that means there is no update to it and so should not be included
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

	// We will consider it the same status and update in-place if the name and the category matches.
	// else, a new status will be created.
	forCreation := make([]interface{}, 0)
	forUpdate := make([]interface{}, 0)
	for _, status := range forUpdateOrCreation {
		statusMap := status.(map[string]interface{})
		statusName, ok := statusMap["name"]
		if !ok {
			continue
		}
		statusCat, ok := statusMap["category"]
		if !ok {
			continue
		}
		toCreateNewStatus := true

		// If the status matches an existing name and same category then we'll consider
		// it the same and not create a new one.
		for _, existingStatus := range *oldWorktype.Statuses {
			if *existingStatus.Name == statusName && *existingStatus.Category == statusCat {
				toCreateNewStatus = false
				break
			}
		}

		if toCreateNewStatus {
			forCreation = append(forCreation, status)
		} else {
			forUpdate = append(forUpdate, status)
		}
	}

	// Create new statuses
	log.Printf("Creating the task management worktype statuses of %s", d.Id())
	if _, err := createWorktypeStatuses(ctx, proxy, d.Id(), forCreation); err != nil {
		return diag.Errorf("failed to create task management worktype statuses: %v", err)
	}

	// Update the newly created statuses with status refs
	log.Printf("Updating the newly created statuses of worktype %s", d.Id())
	createdStatuses, err := updateWorktypeStatuses(ctx, proxy, d.Id(), forCreation, true)
	if err != nil {
		return diag.Errorf("failed to update task management worktype statuses: %v", err)
	}
	for _, updateStat := range *createdStatuses {
		statusIdsToStay = append(statusIdsToStay, *updateStat.Id)
	}

	// Update the other already existing statuses
	log.Printf("Updating the destination statuses of the statuses of worktype %s", d.Id())
	updatedStatuses, err := updateWorktypeStatuses(ctx, proxy, d.Id(), forUpdate, false)
	if err != nil {
		return diag.Errorf("failed to update task management worktype statuses: %v", err)
	}
	for _, updateStat := range *updatedStatuses {
		statusIdsToStay = append(statusIdsToStay, *updateStat.Id)
	}

	// Delete statuses that are no longer defined in the configuration
	for _, forDeletionId := range lists.SliceDifference(oldStatusIds, statusIdsToStay) {
		if _, err := proxy.deleteTaskManagementWorktypeStatus(ctx, d.Id(), forDeletionId); err != nil {
			return diag.Errorf("failed to delete task management worktype status %s: %v", forDeletionId, err)
		}
	}

	// Update the worktype if 'default_status_name' is changed
	if d.HasChange("default_status_name") {
		time.Sleep(5 * time.Second)
		err := updateDefaultStatusName(ctx, proxy, d, d.Id())
		if err != nil {
			return diag.Errorf("failed to update default status name of worktype: %v", err)
		}
	}

	return readTaskManagementWorktype(ctx, d, meta)
}

// deleteTaskManagementWorktype is used by the task_management_worktype resource to delete a task management worktype from Genesys cloud
func deleteTaskManagementWorktype(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	proxy := getTaskManagementWorktypeProxy(sdkConfig)

	_, err := proxy.deleteTaskManagementWorktype(ctx, d.Id())
	if err != nil {
		return diag.Errorf("failed to delete task management worktype %s: %s", d.Id(), err)
	}

	return gcloud.WithRetries(ctx, 180*time.Second, func() *retry.RetryError {
		_, respCode, err := proxy.getTaskManagementWorktypeById(ctx, d.Id())

		if err != nil {
			if gcloud.IsStatus404ByInt(respCode) {
				log.Printf("Deleted task management worktype %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(fmt.Errorf("error deleting task management worktype %s: %s", d.Id(), err))
		}

		return retry.RetryableError(fmt.Errorf("task management worktype %s still exists", d.Id()))
	})
}
