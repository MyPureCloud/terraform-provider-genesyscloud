package task_management_worktype_flow_onattributechange_rule

import (
	"context"
	"fmt"
	rc "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_cache"
	taskManagementWorktype "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/task_management_worktype"

	"github.com/mypurecloud/platform-client-sdk-go/v157/platformclientv2"
)

/*
The genesyscloud_task_management_onattributechange_rule_proxy.go file contains the proxy structures and methods that interact
with the Genesys Cloud SDK. We use composition here for each function on the proxy so individual functions can be stubbed
out during testing.
*/

// internalProxy holds a proxy instance that can be used throughout the package
var internalProxy *taskManagementOnAttributeChangeRuleProxy

// Type definitions for each func on our proxy so we can easily mock them out later
type createTaskManagementOnAttributeChangeRuleFunc func(ctx context.Context, p *taskManagementOnAttributeChangeRuleProxy, worktypeId string, onAttributeChangeRuleCreate *platformclientv2.Workitemonattributechangerulecreate) (*platformclientv2.Workitemonattributechangerule, *platformclientv2.APIResponse, error)
type getAllTaskManagementOnAttributeChangeRuleFunc func(ctx context.Context, p *taskManagementOnAttributeChangeRuleProxy, worktypeId string) (*[]platformclientv2.Workitemonattributechangerule, *platformclientv2.APIResponse, error)
type getTaskManagementOnAttributeChangeRuleIdByNameFunc func(ctx context.Context, p *taskManagementOnAttributeChangeRuleProxy, worktypeId string, name string) (id string, retryable bool, resp *platformclientv2.APIResponse, err error)
type getTaskManagementOnAttributeChangeRuleByIdFunc func(ctx context.Context, p *taskManagementOnAttributeChangeRuleProxy, worktypeId string, id string) (worktype *platformclientv2.Workitemonattributechangerule, response *platformclientv2.APIResponse, err error)
type updateTaskManagementOnAttributeChangeRuleFunc func(ctx context.Context, p *taskManagementOnAttributeChangeRuleProxy, worktypeId string, id string, onAttributeChangeRuleUpdate *platformclientv2.Workitemonattributechangeruleupdate) (*platformclientv2.Workitemonattributechangerule, *platformclientv2.APIResponse, error)
type deleteTaskManagementOnAttributeChangeRuleFunc func(ctx context.Context, p *taskManagementOnAttributeChangeRuleProxy, worktypeId string, id string) (response *platformclientv2.APIResponse, err error)

// taskManagementOnAttributeChangeRuleProxy contains all the methods that call genesys cloud APIs.
type taskManagementOnAttributeChangeRuleProxy struct {
	clientConfig                                       *platformclientv2.Configuration
	taskManagementApi                                  *platformclientv2.TaskManagementApi
	worktypeProxy                                      *taskManagementWorktype.TaskManagementWorktypeProxy
	createTaskManagementOnAttributeChangeRuleAttr      createTaskManagementOnAttributeChangeRuleFunc
	getAllTaskManagementOnAttributeChangeRuleAttr      getAllTaskManagementOnAttributeChangeRuleFunc
	getTaskManagementOnAttributeChangeRuleIdByNameAttr getTaskManagementOnAttributeChangeRuleIdByNameFunc
	getTaskManagementOnAttributeChangeRuleByIdAttr     getTaskManagementOnAttributeChangeRuleByIdFunc
	updateTaskManagementOnAttributeChangeRuleAttr      updateTaskManagementOnAttributeChangeRuleFunc
	deleteTaskManagementOnAttributeChangeRuleAttr      deleteTaskManagementOnAttributeChangeRuleFunc
	onAttributeChangeRuleCache                         rc.CacheInterface[platformclientv2.Workitemonattributechangerule]
}

// newTaskManagementOnAttributeChangeRuleProxy initializes the task management worktype proxy with all the data needed to communicate with Genesys Cloud
func newTaskManagementOnAttributeChangeRuleProxy(clientConfig *platformclientv2.Configuration) *taskManagementOnAttributeChangeRuleProxy {
	api := platformclientv2.NewTaskManagementApiWithConfig(clientConfig)
	onAttributeChangeRuleCache := rc.NewResourceCache[platformclientv2.Workitemonattributechangerule]()
	taskmanagementProxy := taskManagementWorktype.GetTaskManagementWorktypeProxy(clientConfig)
	return &taskManagementOnAttributeChangeRuleProxy{
		clientConfig:      clientConfig,
		taskManagementApi: api,
		worktypeProxy:     taskmanagementProxy,
		createTaskManagementOnAttributeChangeRuleAttr:      createTaskManagementOnAttributeChangeRuleFn,
		getAllTaskManagementOnAttributeChangeRuleAttr:      getAllTaskManagementOnAttributeChangeRuleFn,
		getTaskManagementOnAttributeChangeRuleIdByNameAttr: getTaskManagementOnAttributeChangeRuleIdByNameFn,
		getTaskManagementOnAttributeChangeRuleByIdAttr:     getTaskManagementOnAttributeChangeRuleByIdFn,
		updateTaskManagementOnAttributeChangeRuleAttr:      updateTaskManagementOnAttributeChangeRuleFn,
		deleteTaskManagementOnAttributeChangeRuleAttr:      deleteTaskManagementOnAttributeChangeRuleFn,
		onAttributeChangeRuleCache:                         onAttributeChangeRuleCache,
	}
}

// GetTaskManagementOnAttributeChangeRuleProxy acts as a singleton to for the internalProxy.  It also ensures
// that we can still proxy our tests by directly setting internalProxy package variable
func getTaskManagementOnAttributeChangeRuleProxy(clientConfig *platformclientv2.Configuration) *taskManagementOnAttributeChangeRuleProxy {
	if internalProxy == nil {
		internalProxy = newTaskManagementOnAttributeChangeRuleProxy(clientConfig)
	}
	return internalProxy
}

// createTaskManagementOnAttributeChangeRule creates a Genesys Cloud task management onattributechange rule
func (p *taskManagementOnAttributeChangeRuleProxy) createTaskManagementOnAttributeChangeRule(ctx context.Context, worktypeId string, onAttributeChangeRuleCreate *platformclientv2.Workitemonattributechangerulecreate) (*platformclientv2.Workitemonattributechangerule, *platformclientv2.APIResponse, error) {
	return p.createTaskManagementOnAttributeChangeRuleAttr(ctx, p, worktypeId, onAttributeChangeRuleCreate)
}

// GetAllTaskManagementOnAttributeChangeRule retrieves all Genesys Cloud task management onattributechange rule
func (p *taskManagementOnAttributeChangeRuleProxy) getAllTaskManagementOnAttributeChangeRule(ctx context.Context, worktypeId string) (*[]platformclientv2.Workitemonattributechangerule, *platformclientv2.APIResponse, error) {
	return p.getAllTaskManagementOnAttributeChangeRuleAttr(ctx, p, worktypeId)
}

// getTaskManagementOnAttributeChangeRuleIdByName returns a single Genesys Cloud task management onattributechange rule by a name
func (p *taskManagementOnAttributeChangeRuleProxy) getTaskManagementOnAttributeChangeRuleIdByName(ctx context.Context, worktypeId string, name string) (id string, retryable bool, resp *platformclientv2.APIResponse, err error) {
	return p.getTaskManagementOnAttributeChangeRuleIdByNameAttr(ctx, p, worktypeId, name)
}

// GetTaskManagementOnAttributeChangeRuleById returns a single Genesys Cloud task management onattributechange rule by Id
func (p *taskManagementOnAttributeChangeRuleProxy) getTaskManagementOnAttributeChangeRuleById(ctx context.Context, worktypeId string, id string) (taskManagementOnAttributeChangeRule *platformclientv2.Workitemonattributechangerule, resp *platformclientv2.APIResponse, err error) {
	return p.getTaskManagementOnAttributeChangeRuleByIdAttr(ctx, p, worktypeId, id)
}

// UpdateTaskManagementOnAttributeChangeRule updates a Genesys Cloud task management onattributechange rule
func (p *taskManagementOnAttributeChangeRuleProxy) updateTaskManagementOnAttributeChangeRule(ctx context.Context, worktypeId string, id string, onAttributeChangeRuleUpdate *platformclientv2.Workitemonattributechangeruleupdate) (*platformclientv2.Workitemonattributechangerule, *platformclientv2.APIResponse, error) {
	return p.updateTaskManagementOnAttributeChangeRuleAttr(ctx, p, worktypeId, id, onAttributeChangeRuleUpdate)
}

// deleteTaskManagementOnAttributeChangeRule deletes a Genesys Cloud task management onattributechange rule by Id
func (p *taskManagementOnAttributeChangeRuleProxy) deleteTaskManagementOnAttributeChangeRule(ctx context.Context, worktypeId string, id string) (resp *platformclientv2.APIResponse, err error) {
	return p.deleteTaskManagementOnAttributeChangeRuleAttr(ctx, p, worktypeId, id)
}

// createTaskManagementOnAttributeChangeRuleFn is an implementation function for creating a Genesys Cloud task management onattributechange rule
func createTaskManagementOnAttributeChangeRuleFn(ctx context.Context, p *taskManagementOnAttributeChangeRuleProxy, worktypeId string, onAttributeChangeRuleCreate *platformclientv2.Workitemonattributechangerulecreate) (*platformclientv2.Workitemonattributechangerule, *platformclientv2.APIResponse, error) {
	return p.taskManagementApi.PostTaskmanagementWorktypeFlowsOnattributechangeRules(worktypeId, *onAttributeChangeRuleCreate)
}

// getAllTaskManagementOnAttributeChangeRuleFn is the implementation for retrieving all task management onattributechange rules in Genesys Cloud
func getAllTaskManagementOnAttributeChangeRuleFn(ctx context.Context, p *taskManagementOnAttributeChangeRuleProxy, worktypeId string) (*[]platformclientv2.Workitemonattributechangerule, *platformclientv2.APIResponse, error) {
	const pageSize = 200
	var (
		allOnAttributeChangeRules []platformclientv2.Workitemonattributechangerule
		after                     = ""
		response                  *platformclientv2.APIResponse
	)
	for {
		onAttributeChangeRules, resp, err := p.taskManagementApi.GetTaskmanagementWorktypeFlowsOnattributechangeRules(worktypeId, after, pageSize)
		response = resp
		if err != nil {
			return nil, resp, fmt.Errorf("failed to get onattributechange rules: %v", err)
		}
		if onAttributeChangeRules.Entities != nil && len(*onAttributeChangeRules.Entities) > 0 {
			allOnAttributeChangeRules = append(allOnAttributeChangeRules, *onAttributeChangeRules.Entities...)
		}

		// Exit loop if there are no more 'pages'
		if onAttributeChangeRules.After == nil || *onAttributeChangeRules.After == "" {
			break
		}
		after = *onAttributeChangeRules.After
	}
	return &allOnAttributeChangeRules, response, nil
}

// getTaskManagementOnAttributeChangeRuleIdByNameFn is an implementation of the function to get a Genesys Cloud task management onattributechange rule by name
func getTaskManagementOnAttributeChangeRuleIdByNameFn(ctx context.Context, p *taskManagementOnAttributeChangeRuleProxy, worktypeId string, name string) (id string, retryable bool, resp *platformclientv2.APIResponse, err error) {
	const pageSize = 200
	var (
		after    = ""
		response *platformclientv2.APIResponse
	)
	for {
		onAttributeChangeRules, resp, err := p.taskManagementApi.GetTaskmanagementWorktypeFlowsOnattributechangeRules(worktypeId, after, pageSize)
		response = resp
		if err != nil {
			return "", false, resp, fmt.Errorf("failed to get onattributechange rules: %v", err)
		}

		if onAttributeChangeRules.Entities != nil && len(*onAttributeChangeRules.Entities) > 0 {
			for i := 0; i < len(*onAttributeChangeRules.Entities); i++ {
				onAttributeChangeRule := (*onAttributeChangeRules.Entities)[i]
				if *onAttributeChangeRule.Name == name {
					return *onAttributeChangeRule.Id, false, resp, nil
				}
			}
		}

		// Exit loop if there are no more 'pages'
		if onAttributeChangeRules.After == nil || *onAttributeChangeRules.After == "" {
			break
		}
		after = *onAttributeChangeRules.After
	}
	return "", true, response, fmt.Errorf("no task management onattributechange rules found with name %s", name)
}

// getTaskManagementOnAttributeChangeRuleByIdFn is an implementation of the function to get a Genesys Cloud task management onattributechange rule by Id
func getTaskManagementOnAttributeChangeRuleByIdFn(ctx context.Context, p *taskManagementOnAttributeChangeRuleProxy, worktypeId string, id string) (taskManagementOnAttributeChangeRule *platformclientv2.Workitemonattributechangerule, resp *platformclientv2.APIResponse, err error) {
	onAttributeChangeRule := rc.GetCacheItem(p.onAttributeChangeRuleCache, id)
	if onAttributeChangeRule != nil {
		return onAttributeChangeRule, nil, nil
	}

	return p.taskManagementApi.GetTaskmanagementWorktypeFlowsOnattributechangeRule(worktypeId, id)
}

// updateTaskManagementOnAttributeChangeRuleFn is an implementation of the function to update a Genesys Cloud task management onattributechange rule
func updateTaskManagementOnAttributeChangeRuleFn(ctx context.Context, p *taskManagementOnAttributeChangeRuleProxy, worktypeId string, id string, onAttributeChangeRuleUpdate *platformclientv2.Workitemonattributechangeruleupdate) (*platformclientv2.Workitemonattributechangerule, *platformclientv2.APIResponse, error) {
	return p.taskManagementApi.PatchTaskmanagementWorktypeFlowsOnattributechangeRule(worktypeId, id, *onAttributeChangeRuleUpdate)
}

// deleteTaskManagementOnAttributeChangeRuleFn is an implementation function for deleting a Genesys Cloud task management onattributechange rule
func deleteTaskManagementOnAttributeChangeRuleFn(ctx context.Context, p *taskManagementOnAttributeChangeRuleProxy, worktypeId string, id string) (resp *platformclientv2.APIResponse, err error) {
	resp, err = p.taskManagementApi.DeleteTaskmanagementWorktypeFlowsOnattributechangeRule(worktypeId, id)
	if err == nil {
		rc.DeleteCacheItem(p.onAttributeChangeRuleCache, id)
	}
	return resp, err
}
