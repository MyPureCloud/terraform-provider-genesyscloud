package task_management_workitem

import (
	"context"
	"fmt"
	"log"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v116/platformclientv2"

	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"

	gcloud "terraform-provider-genesyscloud/genesyscloud"

	"terraform-provider-genesyscloud/genesyscloud/util/resourcedata"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

/*
The resource_genesyscloud_task_management_workitem.go contains all of the methods that perform the core logic for a resource.
*/

// getAllAuthTaskManagementWorkitem retrieves all of the task management workitem via Terraform in the Genesys Cloud and is used for the exporter
func getAllAuthTaskManagementWorkitems(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	proxy := newTaskManagementWorkitemProxy(clientConfig)
	resources := make(resourceExporter.ResourceIDMetaMap)

	workitems, err := proxy.getAllTaskManagementWorkitem(ctx)
	if err != nil {
		return nil, diag.Errorf("Failed to get task management workitem: %v", err)
	}

	for _, workitem := range *workitems {
		resources[*workitem.Id] = &resourceExporter.ResourceMeta{Name: *workitem.Id}
	}

	return resources, nil
}

// createTaskManagementWorkitem is used by the task_management_workitem resource to create Genesys cloud task management workitem
func createTaskManagementWorkitem(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	proxy := getTaskManagementWorkitemProxy(sdkConfig)

	taskManagementWorkitem, err := getTaskManagementWorkitemFromResourceData(d)
	if err != nil {
		return diag.Errorf("failed to build Workitem create from resource data: %v", err)
	}

	log.Printf("Creating task management workitem %s", *taskManagementWorkitem.Name)
	workitem, err := proxy.createTaskManagementWorkitem(ctx, taskManagementWorkitem)
	if err != nil {
		return diag.Errorf("Failed to create task management workitem: %s", err)
	}

	d.SetId(*workitem.Id)
	log.Printf("Created task management workitem %s", *workitem.Id)
	return readTaskManagementWorkitem(ctx, d, meta)
}

// readTaskManagementWorkitem is used by the task_management_workitem resource to read an task management workitem from genesys cloud
func readTaskManagementWorkitem(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	proxy := getTaskManagementWorkitemProxy(sdkConfig)

	log.Printf("Reading task management workitem %s", d.Id())

	return gcloud.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		workitem, respCode, getErr := proxy.getTaskManagementWorkitemById(ctx, d.Id())
		if getErr != nil {
			if gcloud.IsStatus404ByInt(respCode) {
				return retry.RetryableError(fmt.Errorf("Failed to read task management workitem %s: %s", d.Id(), getErr))
			}
			return retry.NonRetryableError(fmt.Errorf("Failed to read task management workitem %s: %s", d.Id(), getErr))
		}

		cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceTaskManagementWorkitem())

		resourcedata.SetNillableValue(d, "name", workitem.Name)
		resourcedata.SetNillableReferenceWritableDivision(d, "division_id", workitem.Division)
		resourcedata.SetNillableValueWithInterfaceArrayWithFunc(d, "type", workitem.Type, flattenWorktypeReference)
		resourcedata.SetNillableValue(d, "description", workitem.Description)
		resourcedata.SetNillableValueWithInterfaceArrayWithFunc(d, "language", workitem.Language, flattenLanguageReference)
		resourcedata.SetNillableValue(d, "priority", workitem.Priority)
		resourcedata.SetNillableValue(d, "date_due", workitem.DateDue)
		resourcedata.SetNillableValue(d, "date_expires", workitem.DateExpires)
		resourcedata.SetNillableValue(d, "duration_seconds", workitem.DurationSeconds)
		resourcedata.SetNillableValue(d, "ttl", workitem.Ttl)
		resourcedata.SetNillableValueWithInterfaceArrayWithFunc(d, "status", workitem.Status, flattenWorkitemStatusReference)
		resourcedata.SetNillableValue(d, "status_category", workitem.StatusCategory)
		resourcedata.SetNillableValueWithInterfaceArrayWithFunc(d, "workbin", workitem.Workbin, flattenWorkbinReference)
		resourcedata.SetNillableValueWithInterfaceArrayWithFunc(d, "assignee", workitem.Assignee, flattenUserReferenceWithName)
		resourcedata.SetNillableValueWithInterfaceArrayWithFunc(d, "external_contact", workitem.ExternalContact, flattenExternalContactReference)
		resourcedata.SetNillableValue(d, "external_tag", workitem.ExternalTag)
		resourcedata.SetNillableValueWithInterfaceArrayWithFunc(d, "queue", workitem.Queue, flattenWorkitemQueueReference)
		resourcedata.SetNillableValueWithInterfaceArrayWithFunc(d, "skills", workitem.Skills, flattenRoutingSkillReferences)
		resourcedata.SetNillableValueWithInterfaceArrayWithFunc(d, "preferred_agents", workitem.PreferredAgents, flattenUserReferences)
		resourcedata.SetNillableValue(d, "auto_status_transition", workitem.AutoStatusTransition)
		// TODO: Handle custom_fields property
		resourcedata.SetNillableValueWithInterfaceArrayWithFunc(d, "scored_agents", workitem.ScoredAgents, flattenWorkitemScoredAgents)

		log.Printf("Read task management workitem %s %s", d.Id(), *workitem.Name)
		return cc.CheckState()
	})
}

// updateTaskManagementWorkitem is used by the task_management_workitem resource to update an task management workitem in Genesys Cloud
func updateTaskManagementWorkitem(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	proxy := getTaskManagementWorkitemProxy(sdkConfig)

	taskManagementWorkitem := getTaskManagementWorkitemFromResourceData(d)

	log.Printf("Updating task management workitem %s", *taskManagementWorkitem.Name)
	workitem, err := proxy.updateTaskManagementWorkitem(ctx, d.Id(), &taskManagementWorkitem)
	if err != nil {
		return diag.Errorf("Failed to update task management workitem: %s", err)
	}

	log.Printf("Updated task management workitem %s", *workitem.Id)
	return readTaskManagementWorkitem(ctx, d, meta)
}

// deleteTaskManagementWorkitem is used by the task_management_workitem resource to delete an task management workitem from Genesys cloud
func deleteTaskManagementWorkitem(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	proxy := getTaskManagementWorkitemProxy(sdkConfig)

	_, err := proxy.deleteTaskManagementWorkitem(ctx, d.Id())
	if err != nil {
		return diag.Errorf("Failed to delete task management workitem %s: %s", d.Id(), err)
	}

	return gcloud.WithRetries(ctx, 180*time.Second, func() *retry.RetryError {
		_, respCode, err := proxy.getTaskManagementWorkitemById(ctx, d.Id())

		if err != nil {
			if gcloud.IsStatus404ByInt(respCode) {
				log.Printf("Deleted task management workitem %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(fmt.Errorf("Error deleting task management workitem %s: %s", d.Id(), err))
		}

		return retry.RetryableError(fmt.Errorf("task management workitem %s still exists", d.Id()))
	})
}
