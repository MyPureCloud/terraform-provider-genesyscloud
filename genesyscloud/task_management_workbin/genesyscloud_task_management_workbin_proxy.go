package task_management_workbin

import (
	"context"
	"fmt"

	"github.com/mypurecloud/platform-client-sdk-go/v119/platformclientv2"
)

/*
The genesyscloud_task_management_workbin_proxy.go file contains the proxy structures and methods that interact
with the Genesys Cloud SDK. We use composition here for each function on the proxy so individual functions can be stubbed
out during testing.
*/

// internalProxy holds a proxy instance that can be used throughout the package
var internalProxy *taskManagementWorkbinProxy

// Type definitions for each func on our proxy so we can easily mock them out later
type createTaskManagementWorkbinFunc func(ctx context.Context, p *taskManagementWorkbinProxy, workbin *platformclientv2.Workbincreate) (*platformclientv2.Workbin, error)
type getAllTaskManagementWorkbinFunc func(ctx context.Context, p *taskManagementWorkbinProxy) (*[]platformclientv2.Workbin, error)
type getTaskManagementWorkbinIdByNameFunc func(ctx context.Context, p *taskManagementWorkbinProxy, name string) (id string, retryable bool, err error)
type getTaskManagementWorkbinByIdFunc func(ctx context.Context, p *taskManagementWorkbinProxy, id string) (workbin *platformclientv2.Workbin, responseCode int, err error)
type updateTaskManagementWorkbinFunc func(ctx context.Context, p *taskManagementWorkbinProxy, id string, workbin *platformclientv2.Workbinupdate) (*platformclientv2.Workbin, error)
type deleteTaskManagementWorkbinFunc func(ctx context.Context, p *taskManagementWorkbinProxy, id string) (responseCode int, err error)

// taskManagementWorkbinProxy contains all of the methods that call genesys cloud APIs.
type taskManagementWorkbinProxy struct {
	clientConfig                         *platformclientv2.Configuration
	taskManagementApi                    *platformclientv2.TaskManagementApi
	createTaskManagementWorkbinAttr      createTaskManagementWorkbinFunc
	getAllTaskManagementWorkbinAttr      getAllTaskManagementWorkbinFunc
	getTaskManagementWorkbinIdByNameAttr getTaskManagementWorkbinIdByNameFunc
	getTaskManagementWorkbinByIdAttr     getTaskManagementWorkbinByIdFunc
	updateTaskManagementWorkbinAttr      updateTaskManagementWorkbinFunc
	deleteTaskManagementWorkbinAttr      deleteTaskManagementWorkbinFunc
}

// newTaskManagementWorkbinProxy initializes the task management workbin proxy with all of the data needed to communicate with Genesys Cloud
func newTaskManagementWorkbinProxy(clientConfig *platformclientv2.Configuration) *taskManagementWorkbinProxy {
	api := platformclientv2.NewTaskManagementApiWithConfig(clientConfig)
	return &taskManagementWorkbinProxy{
		clientConfig:                         clientConfig,
		taskManagementApi:                    api,
		createTaskManagementWorkbinAttr:      createTaskManagementWorkbinFn,
		getAllTaskManagementWorkbinAttr:      getAllTaskManagementWorkbinFn,
		getTaskManagementWorkbinIdByNameAttr: getTaskManagementWorkbinIdByNameFn,
		getTaskManagementWorkbinByIdAttr:     getTaskManagementWorkbinByIdFn,
		updateTaskManagementWorkbinAttr:      updateTaskManagementWorkbinFn,
		deleteTaskManagementWorkbinAttr:      deleteTaskManagementWorkbinFn,
	}
}

// getTaskManagementWorkbinProxy acts as a singleton to for the internalProxy.  It also ensures
// that we can still proxy our tests by directly setting internalProxy package variable
func getTaskManagementWorkbinProxy(clientConfig *platformclientv2.Configuration) *taskManagementWorkbinProxy {
	if internalProxy == nil {
		internalProxy = newTaskManagementWorkbinProxy(clientConfig)
	}

	return internalProxy
}

// createTaskManagementWorkbin creates a Genesys Cloud task management workbin
func (p *taskManagementWorkbinProxy) createTaskManagementWorkbin(ctx context.Context, taskManagementWorkbin *platformclientv2.Workbincreate) (*platformclientv2.Workbin, error) {
	return p.createTaskManagementWorkbinAttr(ctx, p, taskManagementWorkbin)
}

// getTaskManagementWorkbin retrieves all Genesys Cloud task management workbin
func (p *taskManagementWorkbinProxy) getAllTaskManagementWorkbin(ctx context.Context) (*[]platformclientv2.Workbin, error) {
	return p.getAllTaskManagementWorkbinAttr(ctx, p)
}

// getTaskManagementWorkbinIdByName returns a single Genesys Cloud task management workbin by a name
func (p *taskManagementWorkbinProxy) getTaskManagementWorkbinIdByName(ctx context.Context, name string) (id string, retryable bool, err error) {
	return p.getTaskManagementWorkbinIdByNameAttr(ctx, p, name)
}

// getTaskManagementWorkbinById returns a single Genesys Cloud task management workbin by Id
func (p *taskManagementWorkbinProxy) getTaskManagementWorkbinById(ctx context.Context, id string) (taskManagementWorkbin *platformclientv2.Workbin, statusCode int, err error) {
	return p.getTaskManagementWorkbinByIdAttr(ctx, p, id)
}

// updateTaskManagementWorkbin updates a Genesys Cloud task management workbin
func (p *taskManagementWorkbinProxy) updateTaskManagementWorkbin(ctx context.Context, id string, taskManagementWorkbin *platformclientv2.Workbinupdate) (*platformclientv2.Workbin, error) {
	return p.updateTaskManagementWorkbinAttr(ctx, p, id, taskManagementWorkbin)
}

// deleteTaskManagementWorkbin deletes a Genesys Cloud task management workbin by Id
func (p *taskManagementWorkbinProxy) deleteTaskManagementWorkbin(ctx context.Context, id string) (statusCode int, err error) {
	return p.deleteTaskManagementWorkbinAttr(ctx, p, id)
}

// createTaskManagementWorkbinFn is an implementation function for creating a Genesys Cloud task management workbin
func createTaskManagementWorkbinFn(ctx context.Context, p *taskManagementWorkbinProxy, taskManagementWorkbin *platformclientv2.Workbincreate) (*platformclientv2.Workbin, error) {
	workbin, _, err := p.taskManagementApi.PostTaskmanagementWorkbins(*taskManagementWorkbin)
	if err != nil {
		return nil, fmt.Errorf("failed to create task management workbin: %s", err)
	}

	return workbin, nil
}

// getAllTaskManagementWorkbinFn is the implementation for retrieving all task management workbin in Genesys Cloud
func getAllTaskManagementWorkbinFn(ctx context.Context, p *taskManagementWorkbinProxy) (*[]platformclientv2.Workbin, error) {
	var allWorkbins []platformclientv2.Workbin
	pageSize := 200
	after := ""

	for {
		queryReq := &platformclientv2.Workbinqueryrequest{
			PageSize: &pageSize,
			After:    &after,
		}
		workbins, _, err := p.taskManagementApi.PostTaskmanagementWorkbinsQuery(*queryReq)
		if err != nil {
			return nil, fmt.Errorf("failed to get workbin: %v", err)
		}
		allWorkbins = append(allWorkbins, *workbins.Entities...)

		// Exit loop if there are no more 'pages'
		if workbins.After == nil || *workbins.After == "" {
			break
		}
		after = *workbins.After
	}

	return &allWorkbins, nil
}

// getTaskManagementWorkbinIdByNameFn is an implementation of the function to get a Genesys Cloud task management workbin by name
func getTaskManagementWorkbinIdByNameFn(ctx context.Context, p *taskManagementWorkbinProxy, name string) (id string, retryable bool, err error) {
	workbins, err := p.getAllTaskManagementWorkbin(ctx)
	if err != nil {
		return "", false, fmt.Errorf("failed to get workbin %s. failed to get all task management workbins", name)
	}

	for _, workbin := range *workbins {
		if *workbin.Name == name {
			return *workbin.Id, false, nil
		}
	}

	return "", true, fmt.Errorf("no task management workbin found with name %s", name)
}

// getTaskManagementWorkbinByIdFn is an implementation of the function to get a Genesys Cloud task management workbin by Id
func getTaskManagementWorkbinByIdFn(ctx context.Context, p *taskManagementWorkbinProxy, id string) (taskManagementWorkbin *platformclientv2.Workbin, statusCode int, err error) {
	workbin, resp, err := p.taskManagementApi.GetTaskmanagementWorkbin(id)
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("failed to retrieve task management workbin by id %s: %s", id, err)
	}

	return workbin, resp.StatusCode, nil
}

// updateTaskManagementWorkbinFn is an implementation of the function to update a Genesys Cloud task management workbin
func updateTaskManagementWorkbinFn(ctx context.Context, p *taskManagementWorkbinProxy, id string, taskManagementWorkbin *platformclientv2.Workbinupdate) (*platformclientv2.Workbin, error) {
	workbin, _, err := p.taskManagementApi.PatchTaskmanagementWorkbin(id, *taskManagementWorkbin)
	if err != nil {
		return nil, fmt.Errorf("failed to update task management workbin: %s", err)
	}
	return workbin, nil
}

// deleteTaskManagementWorkbinFn is an implementation function for deleting a Genesys Cloud task management workbin
func deleteTaskManagementWorkbinFn(ctx context.Context, p *taskManagementWorkbinProxy, id string) (statusCode int, err error) {
	resp, err := p.taskManagementApi.DeleteTaskmanagementWorkbin(id)
	if err != nil {
		return resp.StatusCode, fmt.Errorf("failed to delete task management workbin: %s", err)
	}

	return resp.StatusCode, nil
}
