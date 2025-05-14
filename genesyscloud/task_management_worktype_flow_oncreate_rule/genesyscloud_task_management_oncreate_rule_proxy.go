package task_management_worktype_flow_oncreate_rule

import (
	"context"
	"fmt"
	rc "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_cache"
	taskManagementWorktype "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/task_management_worktype"

	"github.com/mypurecloud/platform-client-sdk-go/v157/platformclientv2"
)

/*
The genesyscloud_task_management_oncreate_rule_proxy.go file contains the proxy structures and methods that interact
with the Genesys Cloud SDK. We use composition here for each function on the proxy so individual functions can be stubbed
out during testing.
*/

// internalProxy holds a proxy instance that can be used throughout the package
var internalProxy *taskManagementOnCreateRuleProxy

// Type definitions for each func on our proxy so we can easily mock them out later
type createTaskManagementOnCreateRuleFunc func(ctx context.Context, p *taskManagementOnCreateRuleProxy, worktypeId string, onCreateRuleCreate *platformclientv2.Workitemoncreaterulecreate) (*platformclientv2.Workitemoncreaterule, *platformclientv2.APIResponse, error)
type getAllTaskManagementOnCreateRuleFunc func(ctx context.Context, p *taskManagementOnCreateRuleProxy, worktypeId string) (*[]platformclientv2.Workitemoncreaterule, *platformclientv2.APIResponse, error)
type getTaskManagementOnCreateRuleIdByNameFunc func(ctx context.Context, p *taskManagementOnCreateRuleProxy, worktypeId string, name string) (id string, retryable bool, resp *platformclientv2.APIResponse, err error)
type getTaskManagementOnCreateRuleByIdFunc func(ctx context.Context, p *taskManagementOnCreateRuleProxy, worktypeId string, id string) (worktype *platformclientv2.Workitemoncreaterule, response *platformclientv2.APIResponse, err error)
type updateTaskManagementOnCreateRuleFunc func(ctx context.Context, p *taskManagementOnCreateRuleProxy, worktypeId string, id string, onCreateRuleUpdate *platformclientv2.Workitemoncreateruleupdate) (*platformclientv2.Workitemoncreaterule, *platformclientv2.APIResponse, error)
type deleteTaskManagementOnCreateRuleFunc func(ctx context.Context, p *taskManagementOnCreateRuleProxy, worktypeId string, id string) (response *platformclientv2.APIResponse, err error)

// taskManagementOnCreateRuleProxy contains all the methods that call genesys cloud APIs.
type taskManagementOnCreateRuleProxy struct {
	clientConfig                              *platformclientv2.Configuration
	taskManagementApi                         *platformclientv2.TaskManagementApi
	worktypeProxy                             *taskManagementWorktype.TaskManagementWorktypeProxy
	createTaskManagementOnCreateRuleAttr      createTaskManagementOnCreateRuleFunc
	getAllTaskManagementOnCreateRuleAttr      getAllTaskManagementOnCreateRuleFunc
	getTaskManagementOnCreateRuleIdByNameAttr getTaskManagementOnCreateRuleIdByNameFunc
	getTaskManagementOnCreateRuleByIdAttr     getTaskManagementOnCreateRuleByIdFunc
	updateTaskManagementOnCreateRuleAttr      updateTaskManagementOnCreateRuleFunc
	deleteTaskManagementOnCreateRuleAttr      deleteTaskManagementOnCreateRuleFunc
	onCreateRuleCache                         rc.CacheInterface[platformclientv2.Workitemoncreaterule]
}

// newTaskManagementOnCreateRuleProxy initializes the task management worktype proxy with all the data needed to communicate with Genesys Cloud
func newTaskManagementOnCreateRuleProxy(clientConfig *platformclientv2.Configuration) *taskManagementOnCreateRuleProxy {
	api := platformclientv2.NewTaskManagementApiWithConfig(clientConfig)
	onCreateRuleCache := rc.NewResourceCache[platformclientv2.Workitemoncreaterule]()
	taskmanagementProxy := taskManagementWorktype.GetTaskManagementWorktypeProxy(clientConfig)
	return &taskManagementOnCreateRuleProxy{
		clientConfig:                              clientConfig,
		taskManagementApi:                         api,
		worktypeProxy:                             taskmanagementProxy,
		createTaskManagementOnCreateRuleAttr:      createTaskManagementOnCreateRuleFn,
		getAllTaskManagementOnCreateRuleAttr:      getAllTaskManagementOnCreateRuleFn,
		getTaskManagementOnCreateRuleIdByNameAttr: getTaskManagementOnCreateRuleIdByNameFn,
		getTaskManagementOnCreateRuleByIdAttr:     getTaskManagementOnCreateRuleByIdFn,
		updateTaskManagementOnCreateRuleAttr:      updateTaskManagementOnCreateRuleFn,
		deleteTaskManagementOnCreateRuleAttr:      deleteTaskManagementOnCreateRuleFn,
		onCreateRuleCache:                         onCreateRuleCache,
	}
}

// GetTaskManagementOnCreateRuleProxy acts as a singleton to for the internalProxy.  It also ensures
// that we can still proxy our tests by directly setting internalProxy package variable
func getTaskManagementOnCreateRuleProxy(clientConfig *platformclientv2.Configuration) *taskManagementOnCreateRuleProxy {
	if internalProxy == nil {
		internalProxy = newTaskManagementOnCreateRuleProxy(clientConfig)
	}
	return internalProxy
}

// createTaskManagementOnCreateRule creates a Genesys Cloud task management oncreate rule
func (p *taskManagementOnCreateRuleProxy) createTaskManagementOnCreateRule(ctx context.Context, worktypeId string, onCreateRuleCreate *platformclientv2.Workitemoncreaterulecreate) (*platformclientv2.Workitemoncreaterule, *platformclientv2.APIResponse, error) {
	return p.createTaskManagementOnCreateRuleAttr(ctx, p, worktypeId, onCreateRuleCreate)
}

// GetAllTaskManagementOnCreateRule retrieves all Genesys Cloud task management oncreate rule
func (p *taskManagementOnCreateRuleProxy) getAllTaskManagementOnCreateRule(ctx context.Context, worktypeId string) (*[]platformclientv2.Workitemoncreaterule, *platformclientv2.APIResponse, error) {
	return p.getAllTaskManagementOnCreateRuleAttr(ctx, p, worktypeId)
}

// getTaskManagementOnCreateRuleIdByName returns a single Genesys Cloud task management oncreate rule by a name
func (p *taskManagementOnCreateRuleProxy) getTaskManagementOnCreateRuleIdByName(ctx context.Context, worktypeId string, name string) (id string, retryable bool, resp *platformclientv2.APIResponse, err error) {
	return p.getTaskManagementOnCreateRuleIdByNameAttr(ctx, p, worktypeId, name)
}

// GetTaskManagementOnCreateRuleById returns a single Genesys Cloud task management oncreate rule by Id
func (p *taskManagementOnCreateRuleProxy) getTaskManagementOnCreateRuleById(ctx context.Context, worktypeId string, id string) (taskManagementOnCreateRule *platformclientv2.Workitemoncreaterule, resp *platformclientv2.APIResponse, err error) {
	return p.getTaskManagementOnCreateRuleByIdAttr(ctx, p, worktypeId, id)
}

// UpdateTaskManagementOnCreateRule updates a Genesys Cloud task management oncreate rule
func (p *taskManagementOnCreateRuleProxy) updateTaskManagementOnCreateRule(ctx context.Context, worktypeId string, id string, onCreateRuleUpdate *platformclientv2.Workitemoncreateruleupdate) (*platformclientv2.Workitemoncreaterule, *platformclientv2.APIResponse, error) {
	return p.updateTaskManagementOnCreateRuleAttr(ctx, p, worktypeId, id, onCreateRuleUpdate)
}

// deleteTaskManagementOnCreateRule deletes a Genesys Cloud task management oncreate rule by Id
func (p *taskManagementOnCreateRuleProxy) deleteTaskManagementOnCreateRule(ctx context.Context, worktypeId string, id string) (resp *platformclientv2.APIResponse, err error) {
	return p.deleteTaskManagementOnCreateRuleAttr(ctx, p, worktypeId, id)
}

// createTaskManagementOnCreateRuleFn is an implementation function for creating a Genesys Cloud task management oncreate rule
func createTaskManagementOnCreateRuleFn(ctx context.Context, p *taskManagementOnCreateRuleProxy, worktypeId string, onCreateRuleCreate *platformclientv2.Workitemoncreaterulecreate) (*platformclientv2.Workitemoncreaterule, *platformclientv2.APIResponse, error) {
	return p.taskManagementApi.PostTaskmanagementWorktypeFlowsOncreateRules(worktypeId, *onCreateRuleCreate)
}

// getAllTaskManagementOnCreateRuleFn is the implementation for retrieving all task management oncreate rules in Genesys Cloud
func getAllTaskManagementOnCreateRuleFn(ctx context.Context, p *taskManagementOnCreateRuleProxy, worktypeId string) (*[]platformclientv2.Workitemoncreaterule, *platformclientv2.APIResponse, error) {
	const pageSize = 200
	var (
		allOnCreateRules []platformclientv2.Workitemoncreaterule
		after            = ""
		response         *platformclientv2.APIResponse
	)
	for {
		onCreateRules, resp, err := p.taskManagementApi.GetTaskmanagementWorktypeFlowsOncreateRules(worktypeId, after, pageSize)
		response = resp
		if err != nil {
			return nil, resp, fmt.Errorf("failed to get oncreate rules: %v", err)
		}
		if onCreateRules.Entities != nil && len(*onCreateRules.Entities) > 0 {
			allOnCreateRules = append(allOnCreateRules, *onCreateRules.Entities...)
		}

		// Exit loop if there are no more 'pages'
		if onCreateRules.After == nil || *onCreateRules.After == "" {
			break
		}
		after = *onCreateRules.After
	}
	return &allOnCreateRules, response, nil
}

// getTaskManagementOnCreateRuleIdByNameFn is an implementation of the function to get a Genesys Cloud task management oncreate rule by name
func getTaskManagementOnCreateRuleIdByNameFn(ctx context.Context, p *taskManagementOnCreateRuleProxy, worktypeId string, name string) (id string, retryable bool, resp *platformclientv2.APIResponse, err error) {
	const pageSize = 200
	var (
		after    = ""
		response *platformclientv2.APIResponse
	)
	for {
		onCreateRules, resp, err := p.taskManagementApi.GetTaskmanagementWorktypeFlowsOncreateRules(worktypeId, after, pageSize)
		response = resp
		if err != nil {
			return "", false, resp, fmt.Errorf("failed to get oncreate rules: %v", err)
		}

		if onCreateRules.Entities != nil && len(*onCreateRules.Entities) > 0 {
			for i := 0; i < len(*onCreateRules.Entities); i++ {
				onCreateRule := (*onCreateRules.Entities)[i]
				if *onCreateRule.Name == name {
					return *onCreateRule.Id, false, resp, nil
				}
			}
		}

		// Exit loop if there are no more 'pages'
		if onCreateRules.After == nil || *onCreateRules.After == "" {
			break
		}
		after = *onCreateRules.After
	}
	return "", true, response, fmt.Errorf("no task management oncreate rules found with name %s", name)
}

// getTaskManagementOnCreateRuleByIdFn is an implementation of the function to get a Genesys Cloud task management oncreate rule by Id
func getTaskManagementOnCreateRuleByIdFn(ctx context.Context, p *taskManagementOnCreateRuleProxy, worktypeId string, id string) (taskManagementOnCreateRule *platformclientv2.Workitemoncreaterule, resp *platformclientv2.APIResponse, err error) {
	onCreateRule := rc.GetCacheItem(p.onCreateRuleCache, id)
	if onCreateRule != nil {
		return onCreateRule, nil, nil
	}

	return p.taskManagementApi.GetTaskmanagementWorktypeFlowsOncreateRule(worktypeId, id)
}

// updateTaskManagementOnCreateRuleFn is an implementation of the function to update a Genesys Cloud task management oncreate rule
func updateTaskManagementOnCreateRuleFn(ctx context.Context, p *taskManagementOnCreateRuleProxy, worktypeId string, id string, onCreateRuleUpdate *platformclientv2.Workitemoncreateruleupdate) (*platformclientv2.Workitemoncreaterule, *platformclientv2.APIResponse, error) {
	return p.taskManagementApi.PatchTaskmanagementWorktypeFlowsOncreateRule(worktypeId, id, *onCreateRuleUpdate)
}

// deleteTaskManagementOnCreateRuleFn is an implementation function for deleting a Genesys Cloud task management oncreate rule
func deleteTaskManagementOnCreateRuleFn(ctx context.Context, p *taskManagementOnCreateRuleProxy, worktypeId string, id string) (resp *platformclientv2.APIResponse, err error) {
	resp, err = p.taskManagementApi.DeleteTaskmanagementWorktypeFlowsOncreateRule(worktypeId, id)
	if err == nil {
		rc.DeleteCacheItem(p.onCreateRuleCache, id)
	}
	return resp, err
}
