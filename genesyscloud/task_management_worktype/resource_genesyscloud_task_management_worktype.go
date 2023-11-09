package task_management_worktype

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v115/platformclientv2"
	"log"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	"time"

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

	taskManagementWorktype := getTaskManagementWorktypeFromResourceData(d)

	log.Printf("Creating task management worktype %s", *taskManagementWorktype.Name)
	worktype, err := proxy.createTaskManagementWorktype(ctx, &taskManagementWorktype)
	if err != nil {
		return diag.Errorf("Failed to create task management worktype: %s", err)
	}

	d.SetId(*worktype.Id)
	log.Printf("Created task management worktype %s", *worktype.Id)
	return readTaskManagementWorktype(ctx, d, meta)
}

// readTaskManagementWorktype is used by the task_management_worktype resource to read an task management worktype from genesys cloud
func readTaskManagementWorktype(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	proxy := getTaskManagementWorktypeProxy(sdkConfig)

	log.Printf("Reading task management worktype %s", d.Id())

	return gcloud.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		worktype, respCode, getErr := proxy.getTaskManagementWorktypeById(ctx, d.Id())
		if getErr != nil {
			if gcloud.IsStatus404ByInt(respCode) {
				return retry.RetryableError(fmt.Errorf("Failed to read task management worktype %s: %s", d.Id(), getErr))
			}
			return retry.NonRetryableError(fmt.Errorf("Failed to read task management worktype %s: %s", d.Id(), getErr))
		}

		cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceTaskManagementWorktype())

		resourcedata.SetNillableValue(d, "name", worktype.Name)
		resourcedata.SetNillableReferenceWritableDivision(d, "division_id", worktype.Division)
		resourcedata.SetNillableValue(d, "description", worktype.Description)
		resourcedata.SetNillableValueWithInterfaceArrayWithFunc(d, "default_workbin", worktype.DefaultWorkbin, flattenWorkbinReference)
		resourcedata.SetNillableValueWithInterfaceArrayWithFunc(d, "default_status", worktype.DefaultStatus, flattenWorkitemStatusReference)
		resourcedata.SetNillableValueWithInterfaceArrayWithFunc(d, "statuses", worktype.Statuses, flattenWorkitemStatuss)
		resourcedata.SetNillableValue(d, "default_duration_seconds", worktype.DefaultDurationSeconds)
		resourcedata.SetNillableValue(d, "default_expiration_seconds", worktype.DefaultExpirationSeconds)
		resourcedata.SetNillableValue(d, "default_due_duration_seconds", worktype.DefaultDueDurationSeconds)
		resourcedata.SetNillableValue(d, "default_priority", worktype.DefaultPriority)
		resourcedata.SetNillableValueWithInterfaceArrayWithFunc(d, "default_language", worktype.DefaultLanguage, flattenLanguageReference)
		resourcedata.SetNillableValue(d, "default_ttl_seconds", worktype.DefaultTtlSeconds)
		resourcedata.SetNillableValueWithInterfaceArrayWithFunc(d, "default_queue", worktype.DefaultQueue, flattenQueueReference)
		resourcedata.SetNillableValueWithInterfaceArrayWithFunc(d, "default_skills", worktype.DefaultSkills, flattenRoutingSkillReferences)
		resourcedata.SetNillableValue(d, "assignment_enabled", worktype.AssignmentEnabled)
		resourcedata.SetNillableValueWithInterfaceArrayWithFunc(d, "schema", worktype.Schema, flattenWorkitemSchema)

		log.Printf("Read task management worktype %s %s", d.Id(), *worktype.Name)
		return cc.CheckState()
	})
}

// updateTaskManagementWorktype is used by the task_management_worktype resource to update an task management worktype in Genesys Cloud
func updateTaskManagementWorktype(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	proxy := getTaskManagementWorktypeProxy(sdkConfig)

	taskManagementWorktype := getTaskManagementWorktypeFromResourceData(d)

	log.Printf("Updating task management worktype %s", *taskManagementWorktype.Name)
	worktype, err := proxy.updateTaskManagementWorktype(ctx, d.Id(), &taskManagementWorktype)
	if err != nil {
		return diag.Errorf("Failed to update task management worktype: %s", err)
	}

	log.Printf("Updated task management worktype %s", *worktype.Id)
	return readTaskManagementWorktype(ctx, d, meta)
}

// deleteTaskManagementWorktype is used by the task_management_worktype resource to delete an task management worktype from Genesys cloud
func deleteTaskManagementWorktype(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	proxy := getTaskManagementWorktypeProxy(sdkConfig)

	_, err := proxy.deleteTaskManagementWorktype(ctx, d.Id())
	if err != nil {
		return diag.Errorf("Failed to delete task management worktype %s: %s", d.Id(), err)
	}

	return gcloud.WithRetries(ctx, 180*time.Second, func() *retry.RetryError {
		_, respCode, err := proxy.getTaskManagementWorktypeById(ctx, d.Id())

		if err != nil {
			if gcloud.IsStatus404ByInt(respCode) {
				log.Printf("Deleted task management worktype %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(fmt.Errorf("Error deleting task management worktype %s: %s", d.Id(), err))
		}

		return retry.RetryableError(fmt.Errorf("task management worktype %s still exists", d.Id()))
	})
}
