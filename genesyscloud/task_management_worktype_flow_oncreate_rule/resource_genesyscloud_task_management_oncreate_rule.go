package task_management_worktype_flow_oncreate_rule

import (
	"context"
	"fmt"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/constants"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v157/platformclientv2"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/consistency_checker"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/resourcedata"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

/*
The resource_genesyscloud_task_management_oncreate_rule.go contains all of the methods that perform the core logic for a resource.
*/

// getAllAuthTaskManagementOnCreateRule retrieves all of the task management oncreate Rules via Terraform in the Genesys Cloud and is used for the exporter
func getAllAuthTaskManagementOnCreateRule(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	proxy := getTaskManagementOnCreateRuleProxy(clientConfig)
	resources := make(resourceExporter.ResourceIDMetaMap)

	worktypes, resp, err := proxy.worktypeProxy.GetAllTaskManagementWorktype(ctx)
	if err != nil {
		return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get task management worktypes: %v", err), resp)
	}
	if worktypes == nil {
		return resources, nil
	}

	for _, worktype := range *worktypes {
		onCreateRules, resp, err := proxy.getAllTaskManagementOnCreateRule(ctx, *worktype.Id)
		if err != nil {
			return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get task management oncreate rules error: %s", err), resp)
		}
		if onCreateRules == nil {
			continue
		}

		for _, onCreateRule := range *onCreateRules {
			resources[composeWorktypeBasedTerraformId(*worktype.Id, *onCreateRule.Id)] = &resourceExporter.ResourceMeta{BlockLabel: *onCreateRule.Name}
		}
	}
	return resources, nil
}

// createTaskManagementOnCreateRule is used by the task_management_worktype_flow_oncreate_rule resource to create Genesys cloud task management oncreate rule
func createTaskManagementOnCreateRule(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getTaskManagementOnCreateRuleProxy(sdkConfig)

	worktypeId := d.Get("worktype_id").(string)
	workitemOnCreateRuleCreate := getWorkitemoncreaterulecreateFromResourceData(d)

	log.Printf("Creating task management oncreate rule %s for worktype %s", *workitemOnCreateRuleCreate.Name, worktypeId)
	onCreateRule, resp, err := proxy.createTaskManagementOnCreateRule(ctx, worktypeId, &workitemOnCreateRuleCreate)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to create task management oncreate rule %s error: %s", *workitemOnCreateRuleCreate.Name, err), resp)
	}
	log.Printf("Created the base task management oncreate rule %s for worktype %s", *onCreateRule.Id, worktypeId)

	d.SetId(composeWorktypeBasedTerraformId(worktypeId, *onCreateRule.Id))

	return readTaskManagementOnCreateRule(ctx, d, meta)
}

// readTaskManagementOnCreateRule is used by the task_management_worktype_flow_oncreate_rule resource to read a task management oncreate rule from genesys cloud
func readTaskManagementOnCreateRule(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getTaskManagementOnCreateRuleProxy(sdkConfig)
	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceTaskManagementOnCreateRule(), constants.ConsistencyChecks(), ResourceType)

	worktypeId, id := splitWorktypeBasedTerraformId(d.Id())

	log.Printf("Reading task management oncreate rule %s for worktype %s", id, worktypeId)
	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		onCreateRule, resp, getErr := proxy.getTaskManagementOnCreateRuleById(ctx, worktypeId, id)
		if getErr != nil {
			if util.IsStatus404(resp) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("failed to read task management oncreate rule %s for worktype %s | error: %s", id, worktypeId, getErr), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("failed to read task management oncreate rule %s for worktype %s | error: %s", id, worktypeId, getErr), resp))
		}

		resourcedata.SetNillableValue(d, "name", onCreateRule.Name)
		resourcedata.SetNillableValue(d, "worktype_id", onCreateRule.Worktype.Id)

		log.Printf("Read task management oncreate rule %s for worktype %s", id, worktypeId)
		return cc.CheckState(d)
	})
}

// updateTaskManagementOnCreateRule is used by the task_management_worktype_flow_oncreate_rule resource to update a task management oncreate rule in Genesys Cloud
func updateTaskManagementOnCreateRule(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getTaskManagementOnCreateRuleProxy(sdkConfig)

	onCreateRuleUpdate := getWorkitemoncreateruleupdateFromResourceData(d)
	worktypeId, id := splitWorktypeBasedTerraformId(d.Id())

	log.Printf("Updating oncreate rule %s for worktype %s", id, worktypeId)
	_, resp, err := proxy.updateTaskManagementOnCreateRule(ctx, worktypeId, id, &onCreateRuleUpdate)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to update task management oncreate rule %s for worktype %s error: %s", id, worktypeId, err), resp)
	}

	log.Printf("Updated oncreate rule %s for worktype %s", id, worktypeId)

	return readTaskManagementOnCreateRule(ctx, d, meta)
}

// deleteTaskManagementOnCreateRule is used by the task_management_worktype_flow_oncreate_rule resource to delete a task management oncreate rule from Genesys cloud
func deleteTaskManagementOnCreateRule(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getTaskManagementOnCreateRuleProxy(sdkConfig)

	worktypeId, id := splitWorktypeBasedTerraformId(d.Id())

	resp, err := proxy.deleteTaskManagementOnCreateRule(ctx, worktypeId, id)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to delete task management oncreate rule %s for worktype %s error: %s", id, worktypeId, err), resp)
	}

	return util.WithRetries(ctx, 180*time.Second, func() *retry.RetryError {
		_, resp, err := proxy.getTaskManagementOnCreateRuleById(ctx, worktypeId, id)

		if err != nil {
			if util.IsStatus404(resp) {
				log.Printf("Deleted task management oncreate rule %s for worktype %s", id, worktypeId)
				return nil
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("error deleting task management oncreate rule %s for worktype %s | error: %s", id, worktypeId, err), resp))
		}
		return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("task management oncreate rule %s still exists", id), resp))
	})
}
