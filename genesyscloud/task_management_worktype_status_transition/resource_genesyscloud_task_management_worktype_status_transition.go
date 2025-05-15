package task_management_worktype_status_transition

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/consistency_checker"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/constants"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/lists"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v157/platformclientv2"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/resourcedata"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

/*
	NOTE: This resource's Id is in the format <worktypeId>/<statusId> so we can persist the id of the parent worktype.
	The worktype_id field can not be used for this because attribute values are dropped during a read so they can be
	re-read.
*/

/*
The resource_genesyscloud_task_management_worktype_status_transition.go contains all of the methods that perform the core logic for a resource.
*/

// getAllAuthTaskManagementWorkTypeStatusTransition retrieves all of the task management worktype status via Terraform in the Genesys Cloud and is used for the exporter
func getAllAuthTaskManagementWorkTypeStatusTransition(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
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
			resources[*worktype.Id+"/"+*status.Id+" transition"] = &resourceExporter.ResourceMeta{BlockLabel: *worktype.Name + "_" + *status.Name}
		}
	}

	return resources, nil
}

// createTaskManagementWorkTypeStatusTransition is used by the task_management_worktype_status resource to create Genesys cloud task management worktype status
func createTaskManagementWorkTypeStatusTransition(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return modifyTaskManagementWorkTypeStatusTransition(ctx, d, meta, "create")
}

func modifyTaskManagementWorkTypeStatusTransition(ctx context.Context, d *schema.ResourceData, meta interface{}, operation string) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getTaskManagementWorktypeStatusProxy(sdkConfig)
	worktypeId := d.Get("worktype_id").(string)
	log.Printf("%v status during transition create", d.Get("status_id").(string))
	statusId := fetchWorktypeStatusTerraformId(d.Get("status_id").(string))
	destinationStatusIds := lists.BuildSdkStringListFromInterfaceArray(d, "destination_status_ids")
	defaultDestinationStatusId := resourcedata.GetNillableValue[string](d, "default_destination_status_id")
	statusTransitionDelaySeconds := resourcedata.GetNillableValue[int](d, "status_transition_delay_seconds")
	statusTransitionTime := resourcedata.GetNillableValue[string](d, "status_transition_time")

	err := validateSchema(d)
	if err != nil {
		errorMsg := fmt.Sprintf("Failed to %s task management worktype transition %s status: %s", operation, worktypeId, err)
		return util.BuildDiagnosticError(ResourceType, errorMsg, fmt.Errorf(errorMsg))
	}

	var (
		workitemStatus *platformclientv2.Workitemstatus
		resp           *platformclientv2.APIResponse
		getErr         error
	)

	diagErr := util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		workitemStatus, resp, getErr = proxy.getTaskManagementWorktypeStatusById(ctx, worktypeId, statusId)
		if getErr != nil {
			if util.IsStatus404(resp) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read task management worktype %s status %s: %s", worktypeId, statusId, getErr), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read task management worktype %s status %s: %s", worktypeId, statusId, getErr), resp))
		}
		return nil
	})
	if diagErr != nil {
		return diagErr
	}

	// If the user makes a reference to a status that is managed by terraform the id will look like this <worktypeId>/<statusId>
	// so we need to extract just the status id from any status references that look like this
	if destinationStatusIds != nil && len(*destinationStatusIds) > 0 {
		for i, destinationStatusId := range *destinationStatusIds {
			if strings.Contains(destinationStatusId, "/") {
				_, id := splitWorktypeStatusTerraformTransitionId(destinationStatusId)
				(*destinationStatusIds)[i] = id
			}
		}
	}

	if defaultDestinationStatusId != nil && strings.Contains(*defaultDestinationStatusId, "/") {
		_, id := splitWorktypeStatusTerraformTransitionId(*defaultDestinationStatusId)
		defaultDestinationStatusId = &id
	}

	log.Printf("%s task management worktype %s status %s %s in progress", operation, worktypeId, statusId, *workitemStatus.Name)

	taskManagementWorktypeStatus := platformclientv2.Workitemstatusupdate{
		Name:                         workitemStatus.Name,
		Description:                  workitemStatus.Description,
		DestinationStatusIds:         destinationStatusIds,
		DefaultDestinationStatusId:   defaultDestinationStatusId,
		StatusTransitionDelaySeconds: statusTransitionDelaySeconds,
		StatusTransitionTime:         statusTransitionTime,
	}

	diagErr = util.WithRetries(ctx, 60*time.Second, func() *retry.RetryError {
		workitemStatus, resp, err = proxy.updateTaskManagementWorktypeStatusTransition(ctx, worktypeId, statusId, &taskManagementWorktypeStatus)
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

	d.SetId(worktypeId + "/" + *workitemStatus.Id + " transition")

	log.Printf("%s task management worktype %s status %s %s, completed", operation, worktypeId, *workitemStatus.Id, *workitemStatus.Name)
	return readTaskManagementWorkTypeStatusTransition(ctx, d, meta)
}

// readTaskManagementWorkTypeStatusTransition is used by the task_management_worktype_status resource to read an task management worktype status from genesys cloud
func readTaskManagementWorkTypeStatusTransition(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getTaskManagementWorktypeStatusProxy(sdkConfig)
	worktypeId, statusId := splitWorktypeStatusTerraformTransitionId(d.Id())
	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceTaskManagementWorktypeStatusTransition(), constants.ConsistencyChecks(), ResourceType)

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
		_ = d.Set("status_id", *workitemStatus.Worktype.Id+"/"+*workitemStatus.Id)
		if workitemStatus.DestinationStatuses != nil {
			destinationStatuses := make([]interface{}, len(*workitemStatus.DestinationStatuses))
			for i, v := range *workitemStatus.DestinationStatuses {
				destinationStatuses[i] = *v.Id
			}
			_ = d.Set("destination_status_ids", destinationStatuses)
		}
		if workitemStatus.DefaultDestinationStatus != nil && workitemStatus.DefaultDestinationStatus.Id != nil {
			_ = d.Set("default_destination_status_id", *workitemStatus.DefaultDestinationStatus.Id)
		} else {
			_ = d.Set("default_destination_status_id", "")
		}
		resourcedata.SetNillableValue(d, "status_transition_delay_seconds", workitemStatus.StatusTransitionDelaySeconds)
		resourcedata.SetNillableValue(d, "status_transition_time", workitemStatus.StatusTransitionTime)
		log.Printf("Read task management worktype %s status transition %s %s", worktypeId, *workitemStatus.Id, *workitemStatus.Name)
		return cc.CheckState(d)
	})
}

// updateTaskManagementWorkTypeStatusTransition is used by the task_management_worktype_status resource to update an task management worktype status in Genesys Cloud
func updateTaskManagementWorkTypeStatusTransition(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return modifyTaskManagementWorkTypeStatusTransition(ctx, d, meta, "update")
}

// deleteTaskManagementWorkTypeStatusTransition is used by the task_management_worktype_status resource to delete an task management worktype status from Genesys cloud
func deleteTaskManagementWorkTypeStatusTransition(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getTaskManagementWorktypeStatusProxy(sdkConfig)

	worktypeId, statusId := splitWorktypeStatusTerraformTransitionId(d.Id())

	err := validateSchema(d)
	if err != nil {
		errorMsg := fmt.Sprintf("Failed to %s task management worktype transition %s status: %s", "delete", worktypeId, err)
		return util.BuildDiagnosticError(ResourceType, errorMsg, fmt.Errorf(errorMsg))
	}

	var (
		workitemStatus *platformclientv2.Workitemstatus
		resp           *platformclientv2.APIResponse
		getErr         error
	)

	diagErr := util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		workitemStatus, resp, getErr = proxy.getTaskManagementWorktypeStatusById(ctx, worktypeId, statusId)
		if getErr != nil {
			if util.IsStatus404(resp) {
				return nil
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read task management worktype %s status %s: %s", worktypeId, statusId, getErr), resp))
		}
		return nil
	})
	if diagErr != nil {
		return diagErr
	}

	if workitemStatus == nil {
		return nil
	}

	// If this resource manages the default destination status (via default_destination_status_id),
	// set it to nil. Otherwise, keep the value from the API response.
	defaultDestinationStatusId := resourcedata.GetNillableValue[string](d, "default_destination_status_id")
	if defaultDestinationStatusId == nil {
		if workitemStatus.DefaultDestinationStatus != nil && workitemStatus.DefaultDestinationStatus.Id != nil {
			defaultDestinationStatusId = workitemStatus.DefaultDestinationStatus.Id
		}
	} else {
		defaultDestinationStatusId = nil
	}

	// Build a list of destination status IDs that are managed by the API but not by Terraform.
	// This preserves any status IDs that were set outside of Terraform.
	destinationStatusIds := []string{}
	stateDestinationStatusIds := lists.BuildSdkStringListFromInterfaceArray(d, "destination_status_ids")
	apiDestinationStatusIds := workitemStatus.DestinationStatuses
	for _, v := range *apiDestinationStatusIds {
		if !lists.ItemInSlice[string](*v.Id, *stateDestinationStatusIds) {
			destinationStatusIds = append(destinationStatusIds, *v.Id)
		}
	}

	log.Printf("%s task management worktype %s status %s %s in progress", "delete", worktypeId, statusId, *workitemStatus.Name)

	taskManagementWorktypeStatus := Workitemstatusupdate{
		Name:                       workitemStatus.Name,
		Description:                workitemStatus.Description,
		DestinationStatusIds:       &destinationStatusIds,
		DefaultDestinationStatusId: defaultDestinationStatusId,
	}

	diagErr = util.WithRetries(ctx, 60*time.Second, func() *retry.RetryError {
		workitemStatus, resp, err = proxy.patchTaskManagementWorktypeStatusTransition(ctx, worktypeId, statusId, &taskManagementWorktypeStatus)
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
	log.Printf("%s task management worktype %s status  transition %s %s, completed", "delete", worktypeId, *workitemStatus.Id, *workitemStatus.Name)
	return nil
}
