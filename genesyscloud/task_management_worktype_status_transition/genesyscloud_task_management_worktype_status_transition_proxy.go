package task_management_worktype_status_transition

import (
	"context"
	"fmt"
	taskManagementWorktype "terraform-provider-genesyscloud/genesyscloud/task_management_worktype"

	"github.com/mypurecloud/platform-client-sdk-go/v157/platformclientv2"
)

/*
The genesyscloud_task_management_worktype_status_transition_proxy.go file contains the proxy structures and methods that interact
with the Genesys Cloud SDK. We use composition here for each function on the proxy so individual functions can be stubbed
out during testing.
*/

// internalProxy holds a proxy instance that can be used throughout the package
var internalProxy *taskManagementWorktypeStatusTransitionProxy

// Type definitions for each func on our proxy so we can easily mock them out later
type getAllTaskManagementWorktypeStatusFunc func(ctx context.Context, p *taskManagementWorktypeStatusTransitionProxy, worktypeId string) (*[]platformclientv2.Workitemstatus, *platformclientv2.APIResponse, error)
type getTaskManagementWorktypeStatusIdByNameFunc func(ctx context.Context, p *taskManagementWorktypeStatusTransitionProxy, worktypeId string, name string) (string, *platformclientv2.APIResponse, bool, error)
type getTaskManagementWorktypeStatusByIdFunc func(ctx context.Context, p *taskManagementWorktypeStatusTransitionProxy, worktypeId string, statusId string) (*platformclientv2.Workitemstatus, *platformclientv2.APIResponse, error)
type updateTaskManagementWorktypeStatusTransitionFunc func(ctx context.Context, p *taskManagementWorktypeStatusTransitionProxy, worktypeId string, statusId string, workitemStatus *platformclientv2.Workitemstatusupdate) (*platformclientv2.Workitemstatus, *platformclientv2.APIResponse, error)
type getTaskManagementWorktypeFunc func(ctx context.Context, p *taskManagementWorktypeStatusTransitionProxy, worktypeId string) (*platformclientv2.Worktype, *platformclientv2.APIResponse, error)

// taskManagementWorktypeStatusTransitionProxy contains all the methods that call genesys cloud APIs.
type taskManagementWorktypeStatusTransitionProxy struct {
	clientConfig                                     *platformclientv2.Configuration
	taskManagementApi                                *platformclientv2.TaskManagementApi
	worktypeProxy                                    *taskManagementWorktype.TaskManagementWorktypeProxy
	getAllTaskManagementWorktypeStatusAttr           getAllTaskManagementWorktypeStatusFunc
	getTaskManagementWorktypeStatusIdByNameAttr      getTaskManagementWorktypeStatusIdByNameFunc
	getTaskManagementWorktypeStatusByIdAttr          getTaskManagementWorktypeStatusByIdFunc
	updateTaskManagementWorktypeStatusTransitionAttr updateTaskManagementWorktypeStatusTransitionFunc
	getTaskManagementWorktypeAttr                    getTaskManagementWorktypeFunc
}

// newTaskManagementWorktypeStatusProxy initializes the task management worktype status proxy with all of the data needed to communicate with Genesys Cloud
func newTaskManagementWorktypeStatusProxy(clientConfig *platformclientv2.Configuration) *taskManagementWorktypeStatusTransitionProxy {
	api := platformclientv2.NewTaskManagementApiWithConfig(clientConfig)
	taskmanagementProxy := taskManagementWorktype.GetTaskManagementWorktypeProxy(clientConfig)
	return &taskManagementWorktypeStatusTransitionProxy{
		clientConfig:                                     clientConfig,
		taskManagementApi:                                api,
		worktypeProxy:                                    taskmanagementProxy,
		getAllTaskManagementWorktypeStatusAttr:           getAllTaskManagementWorktypeStatusFn,
		getTaskManagementWorktypeStatusIdByNameAttr:      getTaskManagementWorktypeStatusIdByNameFn,
		getTaskManagementWorktypeStatusByIdAttr:          getTaskManagementWorktypeStatusByIdFn,
		updateTaskManagementWorktypeStatusTransitionAttr: updateTaskManagementWorktypeStatusTransitionAttr,
		getTaskManagementWorktypeAttr:                    getTaskManagementWorktypeFn,
	}
}

// getTaskManagementWorktypeStatusProxy acts as a singleton to for the internalProxy.  It also ensures
// that we can still proxy our tests by directly setting internalProxy package variable
func getTaskManagementWorktypeStatusProxy(clientConfig *platformclientv2.Configuration) *taskManagementWorktypeStatusTransitionProxy {
	if internalProxy == nil {
		internalProxy = newTaskManagementWorktypeStatusProxy(clientConfig)
	}

	return internalProxy
}

// getTaskManagementWorktypeStatus retrieves all Genesys Cloud task management worktype status
func (p *taskManagementWorktypeStatusTransitionProxy) getAllTaskManagementWorktypeStatus(ctx context.Context, worktypeId string) (*[]platformclientv2.Workitemstatus, *platformclientv2.APIResponse, error) {
	return p.getAllTaskManagementWorktypeStatusAttr(ctx, p, worktypeId)
}

// getTaskManagementWorktypeStatusIdByName returns a single Genesys Cloud task management worktype status by a name
func (p *taskManagementWorktypeStatusTransitionProxy) getTaskManagementWorktypeStatusIdByName(ctx context.Context, worktypeId string, name string) (string, *platformclientv2.APIResponse, bool, error) {
	return p.getTaskManagementWorktypeStatusIdByNameAttr(ctx, p, worktypeId, name)
}

// getTaskManagementWorktypeStatusById returns a single Genesys Cloud task management worktype status by Id
func (p *taskManagementWorktypeStatusTransitionProxy) getTaskManagementWorktypeStatusById(ctx context.Context, worktypeId string, statusId string) (*platformclientv2.Workitemstatus, *platformclientv2.APIResponse, error) {
	return p.getTaskManagementWorktypeStatusByIdAttr(ctx, p, worktypeId, statusId)
}

// updateTaskManagementWorktypeStatus updates a Genesys Cloud task management worktype status
func (p *taskManagementWorktypeStatusTransitionProxy) updateTaskManagementWorktypeStatusTransition(ctx context.Context, worktypeId string, statusId string, taskManagementWorktypeStatus *platformclientv2.Workitemstatusupdate) (*platformclientv2.Workitemstatus, *platformclientv2.APIResponse, error) {
	return p.updateTaskManagementWorktypeStatusTransitionAttr(ctx, p, worktypeId, statusId, taskManagementWorktypeStatus)
}

// getTaskManagementWorktype returns a single Genesys Cloud task management worktype
func (p *taskManagementWorktypeStatusTransitionProxy) getTaskManagementWorktype(ctx context.Context, worktypeId string) (*platformclientv2.Worktype, *platformclientv2.APIResponse, error) {
	return p.getTaskManagementWorktypeAttr(ctx, p, worktypeId)
}

// getAllTaskManagementWorktypeStatusFn is the implementation for retrieving all task management worktype status in Genesys Cloud
func getAllTaskManagementWorktypeStatusFn(ctx context.Context, p *taskManagementWorktypeStatusTransitionProxy, worktypeId string) (*[]platformclientv2.Workitemstatus, *platformclientv2.APIResponse, error) {
	statuses, resp, err := p.taskManagementApi.GetTaskmanagementWorktypeStatuses(worktypeId)
	if err != nil {
		return nil, resp, err
	}

	return statuses.Entities, resp, nil
}

// getTaskManagementWorktypeStatusIdByNameFn is an implementation of the function to get a Genesys Cloud task management worktype status by name
func getTaskManagementWorktypeStatusIdByNameFn(ctx context.Context, p *taskManagementWorktypeStatusTransitionProxy, worktypeId string, name string) (string, *platformclientv2.APIResponse, bool, error) {
	statuses, resp, err := getAllTaskManagementWorktypeStatusFn(ctx, p, worktypeId)
	if err != nil {
		return "", resp, false, err
	}

	if statuses == nil || len(*statuses) == 0 {
		return "", resp, true, fmt.Errorf("Unable to find task management worktype status with name %s", name)
	}

	for _, workitemStatus := range *statuses {
		if *workitemStatus.Name == name {
			return *workitemStatus.Id, resp, false, nil
		}
	}

	return "", resp, true, fmt.Errorf("Unable to find task management worktype status with name %s", name)
}

// getTaskManagementWorktypeStatusByIdFn is an implementation of the function to get a Genesys Cloud task management worktype status by Id
func getTaskManagementWorktypeStatusByIdFn(ctx context.Context, p *taskManagementWorktypeStatusTransitionProxy, worktypeId string, statusId string) (*platformclientv2.Workitemstatus, *platformclientv2.APIResponse, error) {
	return p.taskManagementApi.GetTaskmanagementWorktypeStatus(worktypeId, statusId)
}

// updateTaskManagementWorktypeStatusFn is an implementation of the function to update a Genesys Cloud task management worktype status
func updateTaskManagementWorktypeStatusTransitionAttr(ctx context.Context, p *taskManagementWorktypeStatusTransitionProxy, worktypeId string, statusId string, taskManagementWorktypeStatus *platformclientv2.Workitemstatusupdate) (*platformclientv2.Workitemstatus, *platformclientv2.APIResponse, error) {
	return p.taskManagementApi.PatchTaskmanagementWorktypeStatus(worktypeId, statusId, *taskManagementWorktypeStatus)
}

// getTaskManagementWorktypeFn is an implementation of the function to get a Genesys Cloud task management worktype
func getTaskManagementWorktypeFn(ctx context.Context, p *taskManagementWorktypeStatusTransitionProxy, worktypeId string) (*platformclientv2.Worktype, *platformclientv2.APIResponse, error) {
	return p.taskManagementApi.GetTaskmanagementWorktype(worktypeId, nil)
}
