package task_management_workitem

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
The resource_genesyscloud_task_management_workitem.go contains all of the methods that perform the core logic for a resource.
*/

// getAllAuthTaskManagementWorkitem retrieves all of the task management workitem via Terraform in the Genesys Cloud and is used for the exporter
func getAllAuthTaskManagementWorkitems(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	proxy := getTaskManagementWorkitemProxy(clientConfig)
	resources := make(resourceExporter.ResourceIDMetaMap)

	workitems, resp, err := proxy.getAllTaskManagementWorkitem(ctx)
	if err != nil {
		return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get task management workitem error: %s", err), resp)
	}

	for _, workitem := range *workitems {
		resources[*workitem.Id] = &resourceExporter.ResourceMeta{BlockLabel: *workitem.Name}
	}

	return resources, nil
}

// createTaskManagementWorkitem is used by the task_management_workitem resource to create Genesys cloud task management workitem
func createTaskManagementWorkitem(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getTaskManagementWorkitemProxy(sdkConfig)

	taskManagementWorkitem, err := getWorkitemCreateFromResourceData(d)
	if err != nil {
		return util.BuildDiagnosticError(ResourceType, "failed to build Workitem create from resource data", err)
	}

	log.Printf("Creating task management workitem %s", *taskManagementWorkitem.Name)
	workitem, resp, err := proxy.createTaskManagementWorkitem(ctx, taskManagementWorkitem)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to create task management workitem %s error: %s", *taskManagementWorkitem.Name, err), resp)
	}

	d.SetId(*workitem.Id)
	log.Printf("Created task management workitem %s", *workitem.Id)
	return readTaskManagementWorkitem(ctx, d, meta)
}

// readTaskManagementWorkitem is used by the task_management_workitem resource to read an task management workitem from genesys cloud
func readTaskManagementWorkitem(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getTaskManagementWorkitemProxy(sdkConfig)
	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceTaskManagementWorkitem(), constants.ConsistencyChecks(), ResourceType)

	log.Printf("Reading task management workitem %s", d.Id())

	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		workitem, resp, getErr := proxy.getTaskManagementWorkitemById(ctx, d.Id())
		if getErr != nil {
			if util.IsStatus404(resp) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("failed to read task management workitem %s | error: %s", d.Id(), getErr), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("failed to read task management workitem %s | error: %s", d.Id(), getErr), resp))
		}

		resourcedata.SetNillableValue(d, "name", workitem.Name)
		resourcedata.SetNillableValue(d, "description", workitem.Description)
		resourcedata.SetNillableValue(d, "priority", workitem.Priority)
		resourcedata.SetNillableTime(d, "date_due", workitem.DateDue)
		resourcedata.SetNillableTime(d, "date_expires", workitem.DateExpires)
		resourcedata.SetNillableValue(d, "duration_seconds", workitem.DurationSeconds)
		resourcedata.SetNillableValue(d, "ttl", workitem.Ttl)
		resourcedata.SetNillableValue(d, "external_tag", workitem.ExternalTag)
		resourcedata.SetNillableValue(d, "auto_status_transition", workitem.AutoStatusTransition)

		if workitem.VarType != nil {
			resourcedata.SetNillableValue(d, "worktype_id", workitem.VarType.Id)
		}
		if workitem.Language != nil {
			resourcedata.SetNillableValue(d, "language_id", workitem.Language.Id)
		}
		if workitem.Status != nil {
			resourcedata.SetNillableValue(d, "status_id", workitem.Status.Id)
		}
		if workitem.Workbin != nil {
			resourcedata.SetNillableValue(d, "workbin_id", workitem.Workbin.Id)
		}
		if workitem.Assignee != nil {
			resourcedata.SetNillableValue(d, "assignee_id", workitem.Assignee.Id)
		}
		if workitem.ExternalContact != nil {
			resourcedata.SetNillableValue(d, "external_contact_id", workitem.ExternalContact.Id)
		}
		if workitem.Queue != nil {
			resourcedata.SetNillableValue(d, "queue_id", workitem.Queue.Id)
		}

		resourcedata.SetNillableValueWithInterfaceArrayWithFunc(d, "skills_ids", workitem.Skills, flattenRoutingSkillReferences)
		resourcedata.SetNillableValueWithInterfaceArrayWithFunc(d, "preferred_agents_ids", workitem.PreferredAgents, flattenUserReferences)
		resourcedata.SetNillableValueWithInterfaceArrayWithFunc(d, "scored_agents", workitem.ScoredAgents, flattenWorkitemScoredAgents)

		if workitem.CustomFields != nil {
			cf, err := flattenCustomFields(workitem.CustomFields)
			if err != nil {
				return retry.NonRetryableError(fmt.Errorf("failed to flatten custom fields: %v", err))
			}
			d.Set("custom_fields", cf)
		} else {
			d.Set("custom_fields", "")
		}

		log.Printf("Read task management workitem %s %s", d.Id(), *workitem.Name)
		return cc.CheckState(d)
	})
}

// updateTaskManagementWorkitem is used by the task_management_workitem resource to update an task management workitem in Genesys Cloud
func updateTaskManagementWorkitem(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getTaskManagementWorkitemProxy(sdkConfig)

	taskManagementWorkitem, err := getWorkitemUpdateFromResourceData(d)
	if err != nil {
		return util.BuildDiagnosticError(ResourceType, "failed to update Workitem create from resource data", err)
	}

	log.Printf("Updating task management workitem %s", *taskManagementWorkitem.Name)
	workitem, resp, err := proxy.updateTaskManagementWorkitem(ctx, d.Id(), taskManagementWorkitem)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to update task management workitem %s error: %s", *taskManagementWorkitem.Name, err), resp)
	}

	log.Printf("Updated task management workitem %s", *workitem.Id)
	return readTaskManagementWorkitem(ctx, d, meta)
}

// deleteTaskManagementWorkitem is used by the task_management_workitem resource to delete an task management workitem from Genesys cloud
func deleteTaskManagementWorkitem(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getTaskManagementWorkitemProxy(sdkConfig)

	resp, err := proxy.deleteTaskManagementWorkitem(ctx, d.Id())
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to delete task management workitem %s error: %s", d.Id(), err), resp)
	}

	return util.WithRetries(ctx, 180*time.Second, func() *retry.RetryError {
		_, resp, err := proxy.getTaskManagementWorkitemById(ctx, d.Id())

		if err != nil {
			if util.IsStatus404(resp) {
				log.Printf("Deleted task management workitem %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("error deleting task management workitem %s | error: %s", d.Id(), err), resp))
		}
		return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("task management workitem %s still exists", d.Id()), resp))
	})
}
