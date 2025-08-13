package task_management_worktype_flow_datebased_rule

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/constants"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v165/platformclientv2"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/consistency_checker"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/resourcedata"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

/*
The resource_genesyscloud_task_management_datebased_rule.go contains all of the methods that perform the core logic for a resource.
*/

// getAllAuthTaskManagementDateBasedRule retrieves all of the task management datebased Rules via Terraform in the Genesys Cloud and is used for the exporter
func getAllAuthTaskManagementDateBasedRule(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	proxy := getTaskManagementDateBasedRuleProxy(clientConfig)
	resources := make(resourceExporter.ResourceIDMetaMap)

	worktypes, resp, err := proxy.worktypeProxy.GetAllTaskManagementWorktype(ctx)
	if err != nil {
		return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get task management worktypes: %v", err), resp)
	}
	if worktypes == nil {
		return resources, nil
	}

	for _, worktype := range *worktypes {
		dateBasedRules, resp, err := proxy.getAllTaskManagementDateBasedRule(ctx, *worktype.Id)
		if err != nil {
			return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get task management datebased rules error: %s", err), resp)
		}
		if dateBasedRules == nil {
			continue
		}

		for _, dateBasedRule := range *dateBasedRules {
			resources[composeWorktypeBasedTerraformId(*worktype.Id, *dateBasedRule.Id)] = &resourceExporter.ResourceMeta{BlockLabel: *dateBasedRule.Name}
		}
	}
	return resources, nil
}

// createTaskManagementDateBasedRule is used by the task_management_worktype_flow_datebased_rule resource to create Genesys cloud task management datebased rule
func createTaskManagementDateBasedRule(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getTaskManagementDateBasedRuleProxy(sdkConfig)

	worktypeId := d.Get("worktype_id").(string)
	dateBasedRuleCreate := getWorkitemdatebasedrulecreateFromResourceData(d)

	log.Printf("Creating task management datebased rule %s for worktype %s", *dateBasedRuleCreate.Name, worktypeId)
	dateBasedRule, resp, err := proxy.createTaskManagementDateBasedRule(ctx, worktypeId, &dateBasedRuleCreate)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to create task management datebased rule %s error: %s", *dateBasedRuleCreate.Name, err), resp)
	}
	log.Printf("Created the base task management datebased rule %s for worktype %s", *dateBasedRule.Id, worktypeId)

	d.SetId(composeWorktypeBasedTerraformId(worktypeId, *dateBasedRule.Id))

	return readTaskManagementDateBasedRule(ctx, d, meta)
}

// readTaskManagementDateBasedRule is used by the task_management_worktype_flow_datebased_rule resource to read a task management datebased rule from genesys cloud
func readTaskManagementDateBasedRule(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getTaskManagementDateBasedRuleProxy(sdkConfig)
	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceTaskManagementDateBasedRule(), constants.ConsistencyChecks(), ResourceType)

	worktypeId, id := splitWorktypeBasedTerraformId(d.Id())

	log.Printf("Reading task management datebased rule %s for worktype %s", id, worktypeId)
	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		dateBasedRule, resp, getErr := proxy.getTaskManagementDateBasedRuleById(ctx, worktypeId, id)
		if getErr != nil {
			if util.IsStatus404(resp) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("failed to read task management datebased rule %s for worktype %s | error: %s", id, worktypeId, getErr), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("failed to read task management datebased rule %s for worktype %s | error: %s", id, worktypeId, getErr), resp))
		}

		resourcedata.SetNillableValue(d, "name", dateBasedRule.Name)
		resourcedata.SetNillableValue(d, "worktype_id", dateBasedRule.Worktype.Id)
		resourcedata.SetNillableValueWithInterfaceArrayWithFunc(d, "condition", dateBasedRule, flattenSdkCondition)

		log.Printf("Read task management datebased rule %s for worktype %s", id, worktypeId)
		return cc.CheckState(d)
	})
}

// updateTaskManagementDateBasedRule is used by the task_management_worktype_flow_datebased_rule resource to update a task management datebased rule in Genesys Cloud
func updateTaskManagementDateBasedRule(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getTaskManagementDateBasedRuleProxy(sdkConfig)

	dateBasedRuleUpdate := getWorkitemdatebasedruleupdateFromResourceData(d)
	worktypeId, id := splitWorktypeBasedTerraformId(d.Id())

	log.Printf("Updating datebased rule %s for worktype %s", id, worktypeId)
	_, resp, err := proxy.updateTaskManagementDateBasedRule(ctx, worktypeId, id, &dateBasedRuleUpdate)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to update task management datebased rule %s for worktype %s error: %s", id, worktypeId, err), resp)
	}

	log.Printf("Updated datebased rule %s for worktype %s", id, worktypeId)

	return readTaskManagementDateBasedRule(ctx, d, meta)
}

// deleteTaskManagementDateBasedRule is used by the task_management_worktype_flow_datebased_rule resource to delete a task management datebased rule from Genesys cloud
func deleteTaskManagementDateBasedRule(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getTaskManagementDateBasedRuleProxy(sdkConfig)

	worktypeId, id := splitWorktypeBasedTerraformId(d.Id())

	resp, err := proxy.deleteTaskManagementDateBasedRule(ctx, worktypeId, id)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to delete task management datebased rule %s for worktype %s error: %s", id, worktypeId, err), resp)
	}

	return util.WithRetries(ctx, 180*time.Second, func() *retry.RetryError {
		_, resp, err := proxy.getTaskManagementDateBasedRuleById(ctx, worktypeId, id)

		if err != nil {
			if util.IsStatus404(resp) {
				log.Printf("Deleted task management datebased rule %s for worktype %s", id, worktypeId)
				return nil
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("error deleting task management datebased rule %s for worktype %s | error: %s", id, worktypeId, err), resp))
		}
		return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("task management datebased rule %s still exists", id), resp))
	})
}
