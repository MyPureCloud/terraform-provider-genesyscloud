package task_management_worktype_flow_datebased_rule

import (
	"context"
	"fmt"

	rc "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_cache"
	taskManagementWorktype "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/task_management_worktype"

	"github.com/mypurecloud/platform-client-sdk-go/v165/platformclientv2"
)

/*
The genesyscloud_task_management_datebased_rule_proxy.go file contains the proxy structures and methods that interact
with the Genesys Cloud SDK. We use composition here for each function on the proxy so individual functions can be stubbed
out during testing.
*/

// internalProxy holds a proxy instance that can be used throughout the package
var internalProxy *taskManagementDateBasedRuleProxy

// Type definitions for each func on our proxy so we can easily mock them out later
type createTaskManagementDateBasedRuleFunc func(ctx context.Context, p *taskManagementDateBasedRuleProxy, worktypeId string, dateBasedRuleCreate *platformclientv2.Workitemdatebasedrulecreate) (*platformclientv2.Workitemdatebasedrule, *platformclientv2.APIResponse, error)
type getAllTaskManagementDateBasedRuleFunc func(ctx context.Context, p *taskManagementDateBasedRuleProxy, worktypeId string) (*[]platformclientv2.Workitemdatebasedrule, *platformclientv2.APIResponse, error)
type getTaskManagementDateBasedRuleIdByNameFunc func(ctx context.Context, p *taskManagementDateBasedRuleProxy, worktypeId string, name string) (id string, retryable bool, resp *platformclientv2.APIResponse, err error)
type getTaskManagementDateBasedRuleByIdFunc func(ctx context.Context, p *taskManagementDateBasedRuleProxy, worktypeId string, id string) (worktype *platformclientv2.Workitemdatebasedrule, response *platformclientv2.APIResponse, err error)
type updateTaskManagementDateBasedRuleFunc func(ctx context.Context, p *taskManagementDateBasedRuleProxy, worktypeId string, id string, dateBasedRuleUpdate *platformclientv2.Workitemdatebasedruleupdate) (*platformclientv2.Workitemdatebasedrule, *platformclientv2.APIResponse, error)
type deleteTaskManagementDateBasedRuleFunc func(ctx context.Context, p *taskManagementDateBasedRuleProxy, worktypeId string, id string) (response *platformclientv2.APIResponse, err error)

// taskManagementDateBasedRuleProxy contains all the methods that call genesys cloud APIs.
type taskManagementDateBasedRuleProxy struct {
	clientConfig                               *platformclientv2.Configuration
	taskManagementApi                          *platformclientv2.TaskManagementApi
	worktypeProxy                              *taskManagementWorktype.TaskManagementWorktypeProxy
	createTaskManagementDateBasedRuleAttr      createTaskManagementDateBasedRuleFunc
	getAllTaskManagementDateBasedRuleAttr      getAllTaskManagementDateBasedRuleFunc
	getTaskManagementDateBasedRuleIdByNameAttr getTaskManagementDateBasedRuleIdByNameFunc
	getTaskManagementDateBasedRuleByIdAttr     getTaskManagementDateBasedRuleByIdFunc
	updateTaskManagementDateBasedRuleAttr      updateTaskManagementDateBasedRuleFunc
	deleteTaskManagementDateBasedRuleAttr      deleteTaskManagementDateBasedRuleFunc
	dateBasedRuleCache                         rc.CacheInterface[platformclientv2.Workitemdatebasedrule]
}

// newTaskManagementDateBasedRuleProxy initializes the task management worktype proxy with all the data needed to communicate with Genesys Cloud
func newTaskManagementDateBasedRuleProxy(clientConfig *platformclientv2.Configuration) *taskManagementDateBasedRuleProxy {
	api := platformclientv2.NewTaskManagementApiWithConfig(clientConfig)
	dateBasedRuleCache := rc.NewResourceCache[platformclientv2.Workitemdatebasedrule]()
	taskmanagementProxy := taskManagementWorktype.GetTaskManagementWorktypeProxy(clientConfig)
	return &taskManagementDateBasedRuleProxy{
		clientConfig:                               clientConfig,
		taskManagementApi:                          api,
		worktypeProxy:                              taskmanagementProxy,
		createTaskManagementDateBasedRuleAttr:      createTaskManagementDateBasedRuleFn,
		getAllTaskManagementDateBasedRuleAttr:      getAllTaskManagementDateBasedRuleFn,
		getTaskManagementDateBasedRuleIdByNameAttr: getTaskManagementDateBasedRuleIdByNameFn,
		getTaskManagementDateBasedRuleByIdAttr:     getTaskManagementDateBasedRuleByIdFn,
		updateTaskManagementDateBasedRuleAttr:      updateTaskManagementDateBasedRuleFn,
		deleteTaskManagementDateBasedRuleAttr:      deleteTaskManagementDateBasedRuleFn,
		dateBasedRuleCache:                         dateBasedRuleCache,
	}
}

// getTaskManagementDateBasedRuleProxy acts as a singleton to for the internalProxy.  It also ensures
// that we can still proxy our tests by directly setting internalProxy package variable
func getTaskManagementDateBasedRuleProxy(clientConfig *platformclientv2.Configuration) *taskManagementDateBasedRuleProxy {
	if internalProxy == nil {
		internalProxy = newTaskManagementDateBasedRuleProxy(clientConfig)
	}
	return internalProxy
}

// createTaskManagementDateBasedRule creates a Genesys Cloud task management datebased rule
func (p *taskManagementDateBasedRuleProxy) createTaskManagementDateBasedRule(ctx context.Context, worktypeId string, dateBasedRuleCreate *platformclientv2.Workitemdatebasedrulecreate) (*platformclientv2.Workitemdatebasedrule, *platformclientv2.APIResponse, error) {
	return p.createTaskManagementDateBasedRuleAttr(ctx, p, worktypeId, dateBasedRuleCreate)
}

// GetAllTaskManagementDateBasedRule retrieves all Genesys Cloud task management datebased rule
func (p *taskManagementDateBasedRuleProxy) getAllTaskManagementDateBasedRule(ctx context.Context, worktypeId string) (*[]platformclientv2.Workitemdatebasedrule, *platformclientv2.APIResponse, error) {
	return p.getAllTaskManagementDateBasedRuleAttr(ctx, p, worktypeId)
}

// getTaskManagementDateBasedRuleIdByName returns a single Genesys Cloud task management datebased rule by a name
func (p *taskManagementDateBasedRuleProxy) getTaskManagementDateBasedRuleIdByName(ctx context.Context, worktypeId string, name string) (id string, retryable bool, resp *platformclientv2.APIResponse, err error) {
	return p.getTaskManagementDateBasedRuleIdByNameAttr(ctx, p, worktypeId, name)
}

// GetTaskManagementDateBasedRuleById returns a single Genesys Cloud task management datebased rule by Id
func (p *taskManagementDateBasedRuleProxy) getTaskManagementDateBasedRuleById(ctx context.Context, worktypeId string, id string) (taskManagementDateBasedRule *platformclientv2.Workitemdatebasedrule, resp *platformclientv2.APIResponse, err error) {
	return p.getTaskManagementDateBasedRuleByIdAttr(ctx, p, worktypeId, id)
}

// UpdateTaskManagementDateBasedRule updates a Genesys Cloud task management datebased rule
func (p *taskManagementDateBasedRuleProxy) updateTaskManagementDateBasedRule(ctx context.Context, worktypeId string, id string, dateBasedRuleUpdate *platformclientv2.Workitemdatebasedruleupdate) (*platformclientv2.Workitemdatebasedrule, *platformclientv2.APIResponse, error) {
	return p.updateTaskManagementDateBasedRuleAttr(ctx, p, worktypeId, id, dateBasedRuleUpdate)
}

// deleteTaskManagementDateBasedRule deletes a Genesys Cloud task management datebased rule by Id
func (p *taskManagementDateBasedRuleProxy) deleteTaskManagementDateBasedRule(ctx context.Context, worktypeId string, id string) (resp *platformclientv2.APIResponse, err error) {
	return p.deleteTaskManagementDateBasedRuleAttr(ctx, p, worktypeId, id)
}

// createTaskManagementDateBasedRuleFn is an implementation function for creating a Genesys Cloud task management datebased rule
func createTaskManagementDateBasedRuleFn(ctx context.Context, p *taskManagementDateBasedRuleProxy, worktypeId string, dateBasedRuleCreate *platformclientv2.Workitemdatebasedrulecreate) (*platformclientv2.Workitemdatebasedrule, *platformclientv2.APIResponse, error) {
	return p.taskManagementApi.PostTaskmanagementWorktypeFlowsDatebasedRules(worktypeId, *dateBasedRuleCreate)
}

// getAllTaskManagementDateBasedRuleFn is the implementation for retrieving all task management datebased rules in Genesys Cloud
func getAllTaskManagementDateBasedRuleFn(ctx context.Context, p *taskManagementDateBasedRuleProxy, worktypeId string) (*[]platformclientv2.Workitemdatebasedrule, *platformclientv2.APIResponse, error) {
	const pageSize = 200
	var (
		allDateBasedRules []platformclientv2.Workitemdatebasedrule
		after             = ""
		response          *platformclientv2.APIResponse
	)
	for {
		dateBasedRules, resp, err := p.taskManagementApi.GetTaskmanagementWorktypeFlowsDatebasedRules(worktypeId, after, pageSize)
		response = resp
		if err != nil {
			return nil, resp, fmt.Errorf("failed to get datebased rules: %v", err)
		}
		if dateBasedRules.Entities != nil && len(*dateBasedRules.Entities) > 0 {
			allDateBasedRules = append(allDateBasedRules, *dateBasedRules.Entities...)
		}

		// Exit loop if there are no more 'pages'
		if dateBasedRules.After == nil || *dateBasedRules.After == "" {
			break
		}
		after = *dateBasedRules.After
	}
	return &allDateBasedRules, response, nil
}

// getTaskManagementDateBasedRuleIdByNameFn is an implementation of the function to get a Genesys Cloud task management datebased rule by name
func getTaskManagementDateBasedRuleIdByNameFn(ctx context.Context, p *taskManagementDateBasedRuleProxy, worktypeId string, name string) (id string, retryable bool, resp *platformclientv2.APIResponse, err error) {
	const pageSize = 200
	var (
		after    = ""
		response *platformclientv2.APIResponse
	)
	for {
		dateBasedRules, resp, err := p.taskManagementApi.GetTaskmanagementWorktypeFlowsDatebasedRules(worktypeId, after, pageSize)
		response = resp
		if err != nil {
			return "", false, resp, fmt.Errorf("failed to get datebased rules: %v", err)
		}

		if dateBasedRules.Entities != nil && len(*dateBasedRules.Entities) > 0 {
			for i := 0; i < len(*dateBasedRules.Entities); i++ {
				dateBasedRule := (*dateBasedRules.Entities)[i]
				if *dateBasedRule.Name == name {
					return *dateBasedRule.Id, false, resp, nil
				}
			}
		}

		// Exit loop if there are no more 'pages'
		if dateBasedRules.After == nil || *dateBasedRules.After == "" {
			break
		}
		after = *dateBasedRules.After
	}
	return "", true, response, fmt.Errorf("no task management datebased rules found with name %s", name)
}

// getTaskManagementDateBasedRuleByIdFn is an implementation of the function to get a Genesys Cloud task management datebased rule by Id
func getTaskManagementDateBasedRuleByIdFn(ctx context.Context, p *taskManagementDateBasedRuleProxy, worktypeId string, id string) (taskManagementDateBasedRule *platformclientv2.Workitemdatebasedrule, resp *platformclientv2.APIResponse, err error) {
	dateBasedRule := rc.GetCacheItem(p.dateBasedRuleCache, id)
	if dateBasedRule != nil {
		return dateBasedRule, nil, nil
	}

	return p.taskManagementApi.GetTaskmanagementWorktypeFlowsDatebasedRule(worktypeId, id)
}

// updateTaskManagementDateBasedRuleFn is an implementation of the function to update a Genesys Cloud task management datebased rule
func updateTaskManagementDateBasedRuleFn(ctx context.Context, p *taskManagementDateBasedRuleProxy, worktypeId string, id string, dateBasedRuleUpdate *platformclientv2.Workitemdatebasedruleupdate) (*platformclientv2.Workitemdatebasedrule, *platformclientv2.APIResponse, error) {
	return p.taskManagementApi.PatchTaskmanagementWorktypeFlowsDatebasedRule(worktypeId, id, *dateBasedRuleUpdate)
}

// deleteTaskManagementDateBasedRuleFn is an implementation function for deleting a Genesys Cloud task management datebased rule
func deleteTaskManagementDateBasedRuleFn(ctx context.Context, p *taskManagementDateBasedRuleProxy, worktypeId string, id string) (resp *platformclientv2.APIResponse, err error) {
	resp, err = p.taskManagementApi.DeleteTaskmanagementWorktypeFlowsDatebasedRule(worktypeId, id)
	if err == nil {
		rc.DeleteCacheItem(p.dateBasedRuleCache, id)
	}
	return resp, err
}
