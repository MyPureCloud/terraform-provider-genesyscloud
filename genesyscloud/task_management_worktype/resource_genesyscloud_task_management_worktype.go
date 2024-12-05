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

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v146/platformclientv2"

	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"

	"terraform-provider-genesyscloud/genesyscloud/util/resourcedata"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

/*
The resource_genesyscloud_task_management_worktype.go contains all of the methods that perform the core logic for a resource.
*/

// getAllAuthTaskManagementWorktype retrieves all of the task management worktype via Terraform in the Genesys Cloud and is used for the exporter
func getAllAuthTaskManagementWorktypes(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	proxy := GetTaskManagementWorktypeProxy(clientConfig)
	resources := make(resourceExporter.ResourceIDMetaMap)

	worktypes, resp, err := proxy.GetAllTaskManagementWorktype(ctx)
	if err != nil {
		return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get task management worktype error: %s", err), resp)
	}

	for _, worktype := range *worktypes {
		resources[*worktype.Id] = &resourceExporter.ResourceMeta{BlockLabel: *worktype.Name}
	}
	return resources, nil
}

// createTaskManagementWorktype is used by the task_management_worktype resource to create Genesys cloud task management worktype
func createTaskManagementWorktype(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := GetTaskManagementWorktypeProxy(sdkConfig)

	taskManagementWorktype := getWorktypecreateFromResourceData(d)

	// Create the base worktype
	log.Printf("Creating task management worktype %s", *taskManagementWorktype.Name)
	worktype, resp, err := proxy.createTaskManagementWorktype(ctx, &taskManagementWorktype)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to create task management worktype %s error: %s", *taskManagementWorktype.Name, err), resp)
	}

	log.Printf("Created the base task management worktype %s", *worktype.Id)
	d.SetId(*worktype.Id)

	return readTaskManagementWorktype(ctx, d, meta)
}

// readTaskManagementWorktype is used by the task_management_worktype resource to read a task management worktype from genesys cloud
func readTaskManagementWorktype(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := GetTaskManagementWorktypeProxy(sdkConfig)
	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceTaskManagementWorktype(), constants.ConsistencyChecks(), ResourceType)

	log.Printf("Reading task management worktype %s", d.Id())

	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		worktype, resp, getErr := proxy.GetTaskManagementWorktypeById(ctx, d.Id())
		if getErr != nil {
			if util.IsStatus404(resp) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("failed to read task management worktype %s | error: %s", d.Id(), getErr), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("failed to read task management worktype %s | error: %s", d.Id(), getErr), resp))
		}

		resourcedata.SetNillableValue(d, "name", worktype.Name)
		resourcedata.SetNillableValue(d, "description", worktype.Description)
		resourcedata.SetNillableReferenceDivision(d, "division_id", worktype.Division)

		if worktype.DefaultWorkbin != nil {
			resourcedata.SetNillableValue(d, "default_workbin_id", worktype.DefaultWorkbin.Id)
		}

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
	proxy := GetTaskManagementWorktypeProxy(sdkConfig)

	// Update the base configuration of the Worktype
	taskManagementWorktype := getWorktypeupdateFromResourceData(d)

	log.Printf("Updating worktype %s %s", d.Id(), *taskManagementWorktype.Name)
	_, resp, err := proxy.UpdateTaskManagementWorktype(ctx, d.Id(), &taskManagementWorktype)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to update task management worktype %s error: %s", *taskManagementWorktype.Name, err), resp)
	}

	log.Printf("Updated worktype %s %s", d.Id(), *taskManagementWorktype.Name)

	return readTaskManagementWorktype(ctx, d, meta)
}

// deleteTaskManagementWorktype is used by the task_management_worktype resource to delete a task management worktype from Genesys cloud
func deleteTaskManagementWorktype(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := GetTaskManagementWorktypeProxy(sdkConfig)

	resp, err := proxy.deleteTaskManagementWorktype(ctx, d.Id())
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to delete task management worktype %s error: %s", d.Id(), err), resp)
	}

	return util.WithRetries(ctx, 180*time.Second, func() *retry.RetryError {
		_, resp, err := proxy.GetTaskManagementWorktypeById(ctx, d.Id())

		if err != nil {
			if util.IsStatus404(resp) {
				log.Printf("Deleted task management worktype %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("error deleting task management worktype %s | error: %s", d.Id(), err), resp))
		}
		return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("task management worktype %s still exists", d.Id()), resp))
	})
}
