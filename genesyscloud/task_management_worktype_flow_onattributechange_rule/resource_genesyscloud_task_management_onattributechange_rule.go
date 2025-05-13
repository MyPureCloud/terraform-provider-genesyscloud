package task_management_worktype_flow_onattributechange_rule

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
The resource_genesyscloud_task_management_onattributechange_rule.go contains all of the methods that perform the core logic for a resource.
*/

// getAllAuthTaskManagementOnAttributeChangeRule retrieves all of the task management onattributechange Rules via Terraform in the Genesys Cloud and is used for the exporter
func getAllAuthTaskManagementOnAttributeChangeRule(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	proxy := getTaskManagementOnAttributeChangeRuleProxy(clientConfig)
	resources := make(resourceExporter.ResourceIDMetaMap)

	worktypes, resp, err := proxy.worktypeProxy.GetAllTaskManagementWorktype(ctx)
	if err != nil {
		return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get task management worktypes: %v", err), resp)
	}
	if worktypes == nil {
		return resources, nil
	}

	for _, worktype := range *worktypes {
		onAttributeChangeRules, resp, err := proxy.getAllTaskManagementOnAttributeChangeRule(ctx, *worktype.Id)
		if err != nil {
			return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get task management onattributechange rules error: %s", err), resp)
		}
		if onAttributeChangeRules == nil {
			continue
		}

		for _, onAttributeChangeRule := range *onAttributeChangeRules {
			resources[composeWorktypeBasedTerraformId(*worktype.Id, *onAttributeChangeRule.Id)] = &resourceExporter.ResourceMeta{BlockLabel: *onAttributeChangeRule.Name}
		}
	}
	return resources, nil
}

// createTaskManagementOnAttributeChangeRule is used by the task_management_worktype_flow_onattributechange_rule resource to create Genesys cloud task management onattributechange rule
func createTaskManagementOnAttributeChangeRule(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getTaskManagementOnAttributeChangeRuleProxy(sdkConfig)

	worktypeId := d.Get("worktype_id").(string)
	onAttributeChangeRuleCreate := getWorkitemonattributechangerulecreateFromResourceData(d)

	log.Printf("Creating task management onattributechange rule %s for worktype %s", *onAttributeChangeRuleCreate.Name, worktypeId)
	onAttributeChangeRule, resp, err := proxy.createTaskManagementOnAttributeChangeRule(ctx, worktypeId, &onAttributeChangeRuleCreate)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to create task management onattributechange rule %s error: %s", *onAttributeChangeRuleCreate.Name, err), resp)
	}
	log.Printf("Created the base task management onattributechange rule %s for worktype %s", *onAttributeChangeRule.Id, worktypeId)

	d.SetId(composeWorktypeBasedTerraformId(worktypeId, *onAttributeChangeRule.Id))

	return readTaskManagementOnAttributeChangeRule(ctx, d, meta)
}

// readTaskManagementOnAttributeChangeRule is used by the task_management_worktype_flow_onattributechange_rule resource to read a task management onattributechange rule from genesys cloud
func readTaskManagementOnAttributeChangeRule(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getTaskManagementOnAttributeChangeRuleProxy(sdkConfig)
	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceTaskManagementOnAttributeChangeRule(), constants.ConsistencyChecks(), ResourceType)

	worktypeId, id := splitWorktypeBasedTerraformId(d.Id())

	log.Printf("Reading task management onattributechange rule %s for worktype %s", id, worktypeId)
	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		onAttributeChangeRule, resp, getErr := proxy.getTaskManagementOnAttributeChangeRuleById(ctx, worktypeId, id)
		if getErr != nil {
			if util.IsStatus404(resp) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("failed to read task management onattributechange rule %s for worktype %s | error: %s", id, worktypeId, getErr), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("failed to read task management onattributechange rule %s for worktype %s | error: %s", id, worktypeId, getErr), resp))
		}

		resourcedata.SetNillableValue(d, "name", onAttributeChangeRule.Name)
		resourcedata.SetNillableValue(d, "worktype_id", onAttributeChangeRule.Worktype.Id)
		resourcedata.SetNillableValueWithInterfaceArrayWithFunc(d, "condition", onAttributeChangeRule, flattenSdkCondition)

		log.Printf("Read task management onattributechange rule %s for worktype %s", id, worktypeId)
		return cc.CheckState(d)
	})
}

// updateTaskManagementOnAttributeChangeRule is used by the task_management_worktype_flow_onattributechange_rule resource to update a task management onattributechange rule in Genesys Cloud
func updateTaskManagementOnAttributeChangeRule(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getTaskManagementOnAttributeChangeRuleProxy(sdkConfig)

	onAttributeChangeRuleUpdate := getWorkitemonattributechangeruleupdateFromResourceData(d)
	worktypeId, id := splitWorktypeBasedTerraformId(d.Id())

	log.Printf("Updating onattributechange rule %s for worktype %s", id, worktypeId)
	_, resp, err := proxy.updateTaskManagementOnAttributeChangeRule(ctx, worktypeId, id, &onAttributeChangeRuleUpdate)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to update task management onattributechange rule %s for worktype %s error: %s", id, worktypeId, err), resp)
	}

	log.Printf("Updated onattributechange rule %s for worktype %s", id, worktypeId)

	return readTaskManagementOnAttributeChangeRule(ctx, d, meta)
}

// deleteTaskManagementOnAttributeChangeRule is used by the task_management_worktype_flow_onattributechange_rule resource to delete a task management onattributechange rule from Genesys cloud
func deleteTaskManagementOnAttributeChangeRule(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getTaskManagementOnAttributeChangeRuleProxy(sdkConfig)

	worktypeId, id := splitWorktypeBasedTerraformId(d.Id())

	resp, err := proxy.deleteTaskManagementOnAttributeChangeRule(ctx, worktypeId, id)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to delete task management onattributechange rule %s for worktype %s error: %s", id, worktypeId, err), resp)
	}

	return util.WithRetries(ctx, 180*time.Second, func() *retry.RetryError {
		_, resp, err := proxy.getTaskManagementOnAttributeChangeRuleById(ctx, worktypeId, id)

		if err != nil {
			if util.IsStatus404(resp) {
				log.Printf("Deleted task management onattributechange rule %s for worktype %s", id, worktypeId)
				return nil
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("error deleting task management onattributechange rule %s for worktype %s | error: %s", id, worktypeId, err), resp))
		}
		return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("task management onattributechange rule %s still exists", id), resp))
	})
}
