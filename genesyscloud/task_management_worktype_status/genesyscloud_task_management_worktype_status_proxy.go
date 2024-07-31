package task_management_worktype_status

import (
	"context"
	"fmt"
	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
	taskManagementWorktype "terraform-provider-genesyscloud/genesyscloud/task_management_worktype"
)

/*
The genesyscloud_task_management_worktype_status_proxy.go file contains the proxy structures and methods that interact
with the Genesys Cloud SDK. We use composition here for each function on the proxy so individual functions can be stubbed
out during testing.
*/

// internalProxy holds a proxy instance that can be used throughout the package
var internalProxy *taskManagementWorktypeStatusProxy

// Type definitions for each func on our proxy so we can easily mock them out later
type createTaskManagementWorktypeStatusFunc func(ctx context.Context, p *taskManagementWorktypeStatusProxy, worktypeId string, workitemStatus *platformclientv2.Workitemstatuscreate) (*platformclientv2.Workitemstatus, *platformclientv2.APIResponse, error)
type getAllTaskManagementWorktypeStatusFunc func(ctx context.Context, p *taskManagementWorktypeStatusProxy, worktypeId string) (*[]platformclientv2.Workitemstatus, *platformclientv2.APIResponse, error)
type getTaskManagementWorktypeStatusIdByNameFunc func(ctx context.Context, p *taskManagementWorktypeStatusProxy, worktypeId string, name string) (string, *platformclientv2.APIResponse, bool, error)
type getTaskManagementWorktypeStatusByIdFunc func(ctx context.Context, p *taskManagementWorktypeStatusProxy, worktypeId string, statusId string) (*platformclientv2.Workitemstatus, *platformclientv2.APIResponse, error)
type updateTaskManagementWorktypeStatusFunc func(ctx context.Context, p *taskManagementWorktypeStatusProxy, worktypeId string, statusId string, workitemStatus *platformclientv2.Workitemstatusupdate) (*platformclientv2.Workitemstatus, *platformclientv2.APIResponse, error)
type deleteTaskManagementWorktypeStatusFunc func(ctx context.Context, p *taskManagementWorktypeStatusProxy, worktypeId string, statusId string) (*platformclientv2.APIResponse, error)
type getTaskManagementWorktypeFunc func(ctx context.Context, p *taskManagementWorktypeStatusProxy, worktypeId string) (*platformclientv2.Worktype, *platformclientv2.APIResponse, error)

// taskManagementWorktypeStatusProxy contains all the methods that call genesys cloud APIs.
type taskManagementWorktypeStatusProxy struct {
	clientConfig                                *platformclientv2.Configuration
	taskManagementApi                           *platformclientv2.TaskManagementApi
	worktypeProxy                               *taskManagementWorktype.TaskManagementWorktypeProxy
	createTaskManagementWorktypeStatusAttr      createTaskManagementWorktypeStatusFunc
	getAllTaskManagementWorktypeStatusAttr      getAllTaskManagementWorktypeStatusFunc
	getTaskManagementWorktypeStatusIdByNameAttr getTaskManagementWorktypeStatusIdByNameFunc
	getTaskManagementWorktypeStatusByIdAttr     getTaskManagementWorktypeStatusByIdFunc
	updateTaskManagementWorktypeStatusAttr      updateTaskManagementWorktypeStatusFunc
	deleteTaskManagementWorktypeStatusAttr      deleteTaskManagementWorktypeStatusFunc
	getTaskManagementWorktypeAttr               getTaskManagementWorktypeFunc
}

// newTaskManagementWorktypeStatusProxy initializes the task management worktype status proxy with all of the data needed to communicate with Genesys Cloud
func newTaskManagementWorktypeStatusProxy(clientConfig *platformclientv2.Configuration) *taskManagementWorktypeStatusProxy {
	api := platformclientv2.NewTaskManagementApiWithConfig(clientConfig)
	taskmanagementProxy := taskManagementWorktype.GetTaskManagementWorktypeProxy(clientConfig)
	return &taskManagementWorktypeStatusProxy{
		clientConfig:                                clientConfig,
		taskManagementApi:                           api,
		worktypeProxy:                               taskmanagementProxy,
		createTaskManagementWorktypeStatusAttr:      createTaskManagementWorktypeStatusFn,
		getAllTaskManagementWorktypeStatusAttr:      getAllTaskManagementWorktypeStatusFn,
		getTaskManagementWorktypeStatusIdByNameAttr: getTaskManagementWorktypeStatusIdByNameFn,
		getTaskManagementWorktypeStatusByIdAttr:     getTaskManagementWorktypeStatusByIdFn,
		updateTaskManagementWorktypeStatusAttr:      updateTaskManagementWorktypeStatusFn,
		deleteTaskManagementWorktypeStatusAttr:      deleteTaskManagementWorktypeStatusFn,
		getTaskManagementWorktypeAttr:               getTaskManagementWorktypeFn,
	}
}

// getTaskManagementWorktypeStatusProxy acts as a singleton to for the internalProxy.  It also ensures
// that we can still proxy our tests by directly setting internalProxy package variable
func getTaskManagementWorktypeStatusProxy(clientConfig *platformclientv2.Configuration) *taskManagementWorktypeStatusProxy {
	if internalProxy == nil {
		internalProxy = newTaskManagementWorktypeStatusProxy(clientConfig)
	}

	return internalProxy
}

// createTaskManagementWorktypeStatus creates a Genesys Cloud task management worktype status
func (p *taskManagementWorktypeStatusProxy) createTaskManagementWorktypeStatus(ctx context.Context, worktypeId string, taskManagementWorktypeStatus *platformclientv2.Workitemstatuscreate) (*platformclientv2.Workitemstatus, *platformclientv2.APIResponse, error) {
	return p.createTaskManagementWorktypeStatusAttr(ctx, p, worktypeId, taskManagementWorktypeStatus)
}

// getTaskManagementWorktypeStatus retrieves all Genesys Cloud task management worktype status
func (p *taskManagementWorktypeStatusProxy) getAllTaskManagementWorktypeStatus(ctx context.Context, worktypeId string) (*[]platformclientv2.Workitemstatus, *platformclientv2.APIResponse, error) {
	return p.getAllTaskManagementWorktypeStatusAttr(ctx, p, worktypeId)
}

// getTaskManagementWorktypeStatusIdByName returns a single Genesys Cloud task management worktype status by a name
func (p *taskManagementWorktypeStatusProxy) getTaskManagementWorktypeStatusIdByName(ctx context.Context, worktypeId string, name string) (string, *platformclientv2.APIResponse, bool, error) {
	return p.getTaskManagementWorktypeStatusIdByNameAttr(ctx, p, worktypeId, name)
}

// getTaskManagementWorktypeStatusById returns a single Genesys Cloud task management worktype status by Id
func (p *taskManagementWorktypeStatusProxy) getTaskManagementWorktypeStatusById(ctx context.Context, worktypeId string, statusId string) (*platformclientv2.Workitemstatus, *platformclientv2.APIResponse, error) {
	return p.getTaskManagementWorktypeStatusByIdAttr(ctx, p, worktypeId, statusId)
}

// updateTaskManagementWorktypeStatus updates a Genesys Cloud task management worktype status
func (p *taskManagementWorktypeStatusProxy) updateTaskManagementWorktypeStatus(ctx context.Context, worktypeId string, statusId string, taskManagementWorktypeStatus *platformclientv2.Workitemstatusupdate) (*platformclientv2.Workitemstatus, *platformclientv2.APIResponse, error) {
	return p.updateTaskManagementWorktypeStatusAttr(ctx, p, worktypeId, statusId, taskManagementWorktypeStatus)
}

// deleteTaskManagementWorktypeStatus deletes a Genesys Cloud task management worktype status by Id
func (p *taskManagementWorktypeStatusProxy) deleteTaskManagementWorktypeStatus(ctx context.Context, worktypeId string, statusId string) (*platformclientv2.APIResponse, error) {
	return p.deleteTaskManagementWorktypeStatusAttr(ctx, p, worktypeId, statusId)
}

// getTaskManagementWorktype returns a single Genesys Cloud task management worktype
func (p *taskManagementWorktypeStatusProxy) getTaskManagementWorktype(ctx context.Context, worktypeId string) (*platformclientv2.Worktype, *platformclientv2.APIResponse, error) {
	return p.getTaskManagementWorktypeAttr(ctx, p, worktypeId)
}

// createTaskManagementWorktypeStatusFn is an implementation function for creating a Genesys Cloud task management worktype status
func createTaskManagementWorktypeStatusFn(ctx context.Context, p *taskManagementWorktypeStatusProxy, worktypeId string, taskManagementWorktypeStatus *platformclientv2.Workitemstatuscreate) (*platformclientv2.Workitemstatus, *platformclientv2.APIResponse, error) {
	return p.taskManagementApi.PostTaskmanagementWorktypeStatuses(worktypeId, *taskManagementWorktypeStatus)
}

// getAllTaskManagementWorktypeStatusFn is the implementation for retrieving all task management worktype status in Genesys Cloud
func getAllTaskManagementWorktypeStatusFn(ctx context.Context, p *taskManagementWorktypeStatusProxy, worktypeId string) (*[]platformclientv2.Workitemstatus, *platformclientv2.APIResponse, error) {
	statuses, resp, err := p.taskManagementApi.GetTaskmanagementWorktypeStatuses(worktypeId)
	if err != nil {
		return nil, resp, err
	}

	return statuses.Entities, resp, nil
}

// getTaskManagementWorktypeStatusIdByNameFn is an implementation of the function to get a Genesys Cloud task management worktype status by name
func getTaskManagementWorktypeStatusIdByNameFn(ctx context.Context, p *taskManagementWorktypeStatusProxy, worktypeId string, name string) (string, *platformclientv2.APIResponse, bool, error) {
	statuses, resp, err := getAllTaskManagementWorktypeStatusFn(ctx, p, worktypeId)
	if err != nil {
		return "", resp, false, err
	}

	if statuses == nil || len(*statuses) == 0 {
		return "", resp, true, err
	}

	for _, workitemStatus := range *statuses {
		if *workitemStatus.Name == name {
			return *workitemStatus.Id, resp, false, nil
		}
	}

	return "", resp, true, fmt.Errorf("Unable to find task management worktype status with name %s", name)
}

// getTaskManagementWorktypeStatusByIdFn is an implementation of the function to get a Genesys Cloud task management worktype status by Id
func getTaskManagementWorktypeStatusByIdFn(ctx context.Context, p *taskManagementWorktypeStatusProxy, worktypeId string, statusId string) (*platformclientv2.Workitemstatus, *platformclientv2.APIResponse, error) {
	return p.taskManagementApi.GetTaskmanagementWorktypeStatus(worktypeId, statusId)
}

// updateTaskManagementWorktypeStatusFn is an implementation of the function to update a Genesys Cloud task management worktype status
func updateTaskManagementWorktypeStatusFn(ctx context.Context, p *taskManagementWorktypeStatusProxy, worktypeId string, statusId string, taskManagementWorktypeStatus *platformclientv2.Workitemstatusupdate) (*platformclientv2.Workitemstatus, *platformclientv2.APIResponse, error) {
	return p.taskManagementApi.PatchTaskmanagementWorktypeStatus(worktypeId, statusId, *taskManagementWorktypeStatus)
}

// deleteTaskManagementWorktypeStatusFn is an implementation function for deleting a Genesys Cloud task management worktype status
func deleteTaskManagementWorktypeStatusFn(ctx context.Context, p *taskManagementWorktypeStatusProxy, worktypeId string, statusId string) (*platformclientv2.APIResponse, error) {
	return p.taskManagementApi.DeleteTaskmanagementWorktypeStatus(worktypeId, statusId)
}

// getTaskManagementWorktypeFn is an implementation of the function to get a Genesys Cloud task management worktype
func getTaskManagementWorktypeFn(ctx context.Context, p *taskManagementWorktypeStatusProxy, worktypeId string) (*platformclientv2.Worktype, *platformclientv2.APIResponse, error) {
	return p.taskManagementApi.GetTaskmanagementWorktype(worktypeId, nil)
}
