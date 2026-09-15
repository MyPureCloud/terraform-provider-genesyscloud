package task_management_worktype

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/constants"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v195/platformclientv2"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/consistency_checker"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/resourcedata"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

/*
The resource_genesyscloud_task_management_worktype.go contains all methods that perform the core logic for a resource.
*/

// getAllAuthTaskManagementWorktype retrieves all task management worktype via Terraform in the Genesys Cloud and is used for the exporter
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

// Category values for the two statuses Genesys Cloud automatically creates for a new Worktype.
// These are stable, non-localized server-side enum values (see the "category" ValidateFunc on
// genesyscloud_task_management_worktype_status), unlike a status's user-facing name.
const (
	openStatusCategory   = "Open"
	closedStatusCategory = "Closed"
)

// createTaskManagementWorktype is used by the task_management_worktype resource to create Genesys cloud task management worktype
func createTaskManagementWorktype(ctx context.Context, d *schema.ResourceData, meta interface{}) (diags diag.Diagnostics) {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := GetTaskManagementWorktypeProxy(sdkConfig)

	taskManagementWorktype := getWorktypecreateFromResourceData(d)

	err := util.RetryWhen(util.IsStatus409, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		// Create the base worktype
		log.Printf("Creating task management worktype %s", *taskManagementWorktype.Name)
		worktype, resp, err := proxy.createTaskManagementWorktype(ctx, &taskManagementWorktype)
		if err != nil {
			return resp, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to create task management worktype %s error: %s", *taskManagementWorktype.Name, err), resp)
		}

		log.Printf("Created the base task management worktype %s", *worktype.Id)
		d.SetId(*worktype.Id)
		return resp, nil
	})

	if err != nil {
		diags = append(diags, err...)
	}

	if diags.HasError() {
		return diags
	}

	if diagErr := applyAutoCreatedStatusSettings(ctx, proxy, d); diagErr != nil {
		diags = append(diags, diagErr...)
	}

	return append(diags, readTaskManagementWorktype(ctx, d, meta)...)
}

// applyAutoCreatedStatusSettings configures the Open/Closed statuses that Genesys Cloud
// automatically creates for a Worktype, based on the open_status_default and
// closed_status_auto_terminate fields. It is a no-op for any field the user has not explicitly
// set (GetOkExists-style check via GetNillableBool), and for closed_status_auto_terminate on
// update it only acts when the field actually changed. open_status_default cannot be set to
// false: a Worktype must always have exactly one default status, so there is nothing meaningful
// to fall back to; the caller must set default=true on a different status resource instead.
func applyAutoCreatedStatusSettings(ctx context.Context, proxy *TaskManagementWorktypeProxy, d *schema.ResourceData) diag.Diagnostics {
	worktypeId := d.Id()

	if openDefault := resourcedata.GetNillableBool(d, "open_status_default"); openDefault != nil {
		if !*openDefault {
			return diag.Errorf("open_status_default cannot be set to false: a task management worktype must always have exactly "+
				"one default status. To change the default status, set 'default = true' on a different "+
				"genesyscloud_task_management_worktype_status resource for worktype %s instead of unsetting this field.", worktypeId)
		}

		openStatusId, diagErr := resolveStatusId(ctx, proxy, worktypeId, openStatusCategory, "Open")
		if diagErr != nil {
			return diagErr
		}

		if diagErr := setWorktypeDefaultStatus(ctx, proxy, worktypeId, openStatusId); diagErr != nil {
			return diagErr
		}
	}

	if autoTerminate := resourcedata.GetNillableBool(d, "closed_status_auto_terminate"); autoTerminate != nil {
		closedStatusId, diagErr := resolveStatusId(ctx, proxy, worktypeId, closedStatusCategory, "Closed")
		if diagErr != nil {
			return diagErr
		}

		if resp, err := proxy.patchWorktypeStatusAutoTerminate(ctx, worktypeId, closedStatusId, *autoTerminate); err != nil {
			return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to set auto_terminate_workitem on the Closed status for worktype %s error: %s", worktypeId, err), resp)
		}
	}

	return nil
}

// resolveStatusId looks up the auto-created status for the given category and safely extracts its
// id. It guards against both a nil status (should not happen given the current proxy
// implementation, but the proxy's lookup function is a swappable seam, so this is not assumed)
// and a status returned with a nil Id field (the SDK model's Id is a plain pointer with no
// guarantee it is always populated by every response).
func resolveStatusId(ctx context.Context, proxy *TaskManagementWorktypeProxy, worktypeId string, category string, categoryLabel string) (string, diag.Diagnostics) {
	status, resp, err := proxy.getWorktypeStatusByCategory(ctx, worktypeId, category)
	if err != nil {
		return "", util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to find the auto-created %s status for worktype %s error: %s", categoryLabel, worktypeId, err), resp)
	}
	if status == nil || status.Id == nil {
		return "", util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("The auto-created %s status for worktype %s was found but has no id", categoryLabel, worktypeId), resp)
	}
	return *status.Id, nil
}

// setWorktypeDefaultStatus sets statusId as the worktype's default status. Genesys Cloud may
// already have this status set as the default (e.g. it automatically makes the auto-created Open
// status the default when the worktype is created), in which case the PATCH returns a
// 400 "No change for the record is obtained" error. That response means the desired state is
// already in place, so it is treated as success rather than an error.
func setWorktypeDefaultStatus(ctx context.Context, proxy *TaskManagementWorktypeProxy, worktypeId string, statusId string) diag.Diagnostics {
	worktypeUpdate := platformclientv2.Worktypeupdate{DefaultStatusId: &statusId}
	_, resp, err := proxy.UpdateTaskManagementWorktype(ctx, worktypeId, &worktypeUpdate)
	if err == nil {
		return nil
	}

	if util.IsStatus400(resp) && strings.Contains(resp.ErrorMessage, "No change for the record is obtained") {
		log.Printf("Status %s is already the default status for worktype %s; no update needed", statusId, worktypeId)
		return nil
	}

	return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to set status %s as default for worktype %s error: %s", statusId, worktypeId, err), resp)
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

		if worktype.DefaultScript != nil {
			resourcedata.SetNillableValue(d, "default_script_id", worktype.DefaultScript.Id)
		}

		if worktype.RuleSettings != nil {
			resourcedata.SetNillableValue(d, "flow_rules_enabled", worktype.RuleSettings.FlowRulesEnabled)
		}

		// disable_default_status_creation is a write-only (create-time) flag in the API/SDK.
		// Keep state aligned with config to avoid perpetual diffs/invalid plans on refresh.
		if v := resourcedata.GetNillableBool(d, "disable_default_status_creation"); v != nil {
			_ = d.Set("disable_default_status_creation", *v)
		}

		readAutoCreatedStatusSettings(ctx, proxy, d, worktype)

		log.Printf("Read task management worktype %s %s", d.Id(), *worktype.Name)
		return cc.CheckState(d)
	})
}

// readAutoCreatedStatusSettings populates open_status_default and closed_status_auto_terminate
// (both Optional+Computed) from the worktype's current state on the platform. This keeps the two
// fields accurate on refresh and ensures that when a user omits either field from their config,
// Terraform sees no drift and never attempts to change the platform's existing value.
//
// Lookup failures here are logged and skipped rather than surfaced as errors: an org that used
// disable_default_status_creation, or that has since deleted/renamed the auto-created statuses,
// may not have an Open or Closed status to find, and that should not break reads of the worktype
// itself.
func readAutoCreatedStatusSettings(ctx context.Context, proxy *TaskManagementWorktypeProxy, d *schema.ResourceData, worktype *platformclientv2.Worktype) {
	openStatus, _, err := proxy.getWorktypeStatusByCategory(ctx, d.Id(), openStatusCategory)
	if err != nil {
		log.Printf("Could not find an Open status for worktype %s while reading open_status_default: %s", d.Id(), err)
	} else if openStatus != nil && openStatus.Id != nil && worktype.DefaultStatus != nil && worktype.DefaultStatus.Id != nil {
		_ = d.Set("open_status_default", *openStatus.Id == *worktype.DefaultStatus.Id)
	}

	closedStatus, _, err := proxy.getWorktypeStatusByCategory(ctx, d.Id(), closedStatusCategory)
	if err != nil {
		log.Printf("Could not find a Closed status for worktype %s while reading closed_status_auto_terminate: %s", d.Id(), err)
	} else if closedStatus != nil && closedStatus.AutoTerminateWorkitem != nil {
		_ = d.Set("closed_status_auto_terminate", *closedStatus.AutoTerminateWorkitem)
	}
}

// updateTaskManagementWorktype is used by the task_management_worktype resource to update a task management worktype in Genesys Cloud
func updateTaskManagementWorktype(ctx context.Context, d *schema.ResourceData, meta interface{}) (diags diag.Diagnostics) {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := GetTaskManagementWorktypeProxy(sdkConfig)

	// Update the base configuration of the Worktype. getWorktypeupdateFromResourceData always
	// sets Name (with no HasChange guard), so SetFieldNames has at least that one entry even when
	// nothing in the base worktype actually changed - e.g. an apply that only changes
	// open_status_default/closed_status_auto_terminate, which are not base Worktypeupdate fields
	// at all. Sending a PATCH with only an unchanged Name causes a 400 "No change for the record
	// is obtained" error, so skip the base PATCH entirely when Name is the only field set.
	taskManagementWorktype := getWorktypeupdateFromResourceData(d)

	if len(taskManagementWorktype.SetFieldNames) > 1 {
		diags = append(diags, util.RetryWhen(util.IsStatus409, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
			log.Printf("Updating worktype %s %s", d.Id(), *taskManagementWorktype.Name)
			_, resp, err := proxy.UpdateTaskManagementWorktype(ctx, d.Id(), &taskManagementWorktype)
			if err != nil {
				return resp, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to update task management worktype %s error: %s", *taskManagementWorktype.Name, err), resp)
			}
			log.Printf("Updated worktype %s %s", d.Id(), *taskManagementWorktype.Name)
			return resp, nil
		})...)
	} else {
		log.Printf("No base worktype fields changed for worktype %s; skipping base update call", d.Id())
	}

	if diags.HasError() {
		return diags
	}

	if diagErr := applyAutoCreatedStatusSettingsOnUpdate(ctx, proxy, d); diagErr != nil {
		diags = append(diags, diagErr...)
	}

	if diags.HasError() {
		return diags
	}

	return append(diags, readTaskManagementWorktype(ctx, d, meta)...)
}

// applyAutoCreatedStatusSettingsOnUpdate mirrors applyAutoCreatedStatusSettings, but only acts on
// a field when its value actually changed in this apply (d.HasChange), so updates that don't
// touch these fields never re-call the status APIs.
func applyAutoCreatedStatusSettingsOnUpdate(ctx context.Context, proxy *TaskManagementWorktypeProxy, d *schema.ResourceData) diag.Diagnostics {
	worktypeId := d.Id()

	if d.HasChange("open_status_default") {
		if openDefault := resourcedata.GetNillableBool(d, "open_status_default"); openDefault != nil {
			if !*openDefault {
				return diag.Errorf("open_status_default cannot be set to false: a task management worktype must always have exactly "+
					"one default status. To change the default status, set 'default = true' on a different "+
					"genesyscloud_task_management_worktype_status resource for worktype %s instead of unsetting this field.", worktypeId)
			}

			openStatusId, diagErr := resolveStatusId(ctx, proxy, worktypeId, openStatusCategory, "Open")
			if diagErr != nil {
				return diagErr
			}

			if diagErr := setWorktypeDefaultStatus(ctx, proxy, worktypeId, openStatusId); diagErr != nil {
				return diagErr
			}
		}
	}

	if d.HasChange("closed_status_auto_terminate") {
		if autoTerminate := resourcedata.GetNillableBool(d, "closed_status_auto_terminate"); autoTerminate != nil {
			closedStatusId, diagErr := resolveStatusId(ctx, proxy, worktypeId, closedStatusCategory, "Closed")
			if diagErr != nil {
				return diagErr
			}

			if resp, err := proxy.patchWorktypeStatusAutoTerminate(ctx, worktypeId, closedStatusId, *autoTerminate); err != nil {
				return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to set auto_terminate_workitem on the Closed status for worktype %s error: %s", worktypeId, err), resp)
			}
		}
	}

	return nil
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
