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

	log.Printf("Creating the base task management worktype %s", *worktype.Id)
	d.SetId(*worktype.Id)

	// Create the worktype statuses
	log.Printf("Creating the task management worktype statuses of %s", *worktype.Id)
	sdkWorkitemStatusCreates := buildWorkitemStatusCreates(d.Get("statuses").([]interface{}))
	for _, statusCreate := range *sdkWorkitemStatusCreates {
		_, err := proxy.createTaskManagementWorktypeStatus(ctx, *worktype.Id, &statusCreate)
		if err != nil {
			return diag.Errorf("failed to create worktype status %s: %v", *statusCreate.Name, err)
		}
	}

	// Get all the worktype statuses so we'll have the new statuses for referencing
	worktype, _, err = proxy.getTaskManagementWorktypeById(ctx, *worktype.Id)
	if err != nil {
		return diag.Errorf("failed to get task management worktype %s: %v", *worktype.Name, err)
	}

	// Update the worktype statuses as they need to build the "destination status" references
	log.Printf("Updating the destination statuses of the statuses of worktype %s", *worktype.Id)
	sdkWorkitemStatusUpdates := buildWorkitemStatusUpdates(d.Get("statuses").([]interface{}), worktype.Statuses)
	for _, statusUpdate := range *sdkWorkitemStatusUpdates {
		statusId := getStatusIdFromName(*statusUpdate.Name, worktype.Statuses)

		// API does not allow updating a status with no actual change.
		// This update portion is only for resolving status references, so skip statuses where
		// "destination statuses" and "default destination id" are not set.
		if (statusUpdate.DefaultDestinationStatusId == nil || *statusUpdate.DefaultDestinationStatusId == "") &&
			(statusUpdate.DestinationStatusIds == nil || len(*statusUpdate.DestinationStatusIds) == 0) {
			continue
		}

		if statusId == nil {
			return diag.Errorf("failed to update a status %s. Not found in the worktype %s: %v", *statusUpdate.Name, *worktype.Name, err)
		}

		_, err := proxy.updateTaskManagementWorktypeStatus(ctx, *worktype.Id, *statusId, &statusUpdate)
		if err != nil {
			return diag.Errorf("failed to update worktype status %s: %v", *statusUpdate.Name, err)
		}
	}
	log.Printf("Created the task management worktype statuses of %s", *worktype.Id)

	log.Printf("Finalized creation of task management worktype %s", *worktype.Id)
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
				d.Set("default_status", statusName)
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

	taskManagementWorktype := getWorktypeupdateFromResourceData(d)

	log.Printf("Updating task management worktype %s", *taskManagementWorktype.Name)
	worktype, err := proxy.updateTaskManagementWorktype(ctx, d.Id(), &taskManagementWorktype)
	if err != nil {
		return diag.Errorf("failed to update task management worktype: %s", err)
	}

	log.Printf("Updated task management worktype %s", *worktype.Id)

	// Update the worktype statuses
	// if d.HasChange("statuses") {

	// }

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
