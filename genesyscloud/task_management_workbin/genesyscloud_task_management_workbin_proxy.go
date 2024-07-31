package task_management_workbin

import (
	"context"
	"fmt"

	rc "terraform-provider-genesyscloud/genesyscloud/resource_cache"

	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

/*
The genesyscloud_task_management_workbin_proxy.go file contains the proxy structures and methods that interact
with the Genesys Cloud SDK. We use composition here for each function on the proxy so individual functions can be stubbed
out during testing.
*/

// internalProxy holds a proxy instance that can be used throughout the package
var internalProxy *taskManagementWorkbinProxy

// Type definitions for each func on our proxy so we can easily mock them out later
type createTaskManagementWorkbinFunc func(ctx context.Context, p *taskManagementWorkbinProxy, workbin *platformclientv2.Workbincreate) (*platformclientv2.Workbin, *platformclientv2.APIResponse, error)
type getAllTaskManagementWorkbinFunc func(ctx context.Context, p *taskManagementWorkbinProxy) (*[]platformclientv2.Workbin, *platformclientv2.APIResponse, error)
type getTaskManagementWorkbinIdByNameFunc func(ctx context.Context, p *taskManagementWorkbinProxy, name string) (id string, retryable bool, resp *platformclientv2.APIResponse, err error)
type getTaskManagementWorkbinByIdFunc func(ctx context.Context, p *taskManagementWorkbinProxy, id string) (workbin *platformclientv2.Workbin, response *platformclientv2.APIResponse, err error)
type updateTaskManagementWorkbinFunc func(ctx context.Context, p *taskManagementWorkbinProxy, id string, workbin *platformclientv2.Workbinupdate) (*platformclientv2.Workbin, *platformclientv2.APIResponse, error)
type deleteTaskManagementWorkbinFunc func(ctx context.Context, p *taskManagementWorkbinProxy, id string) (response *platformclientv2.APIResponse, err error)

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
	workbinCache                         rc.CacheInterface[platformclientv2.Workbin]
}

// newTaskManagementWorkbinProxy initializes the task management workbin proxy with all of the data needed to communicate with Genesys Cloud
func newTaskManagementWorkbinProxy(clientConfig *platformclientv2.Configuration) *taskManagementWorkbinProxy {
	api := platformclientv2.NewTaskManagementApiWithConfig(clientConfig)
	workbinCache := rc.NewResourceCache[platformclientv2.Workbin]()
	return &taskManagementWorkbinProxy{
		clientConfig:                         clientConfig,
		taskManagementApi:                    api,
		createTaskManagementWorkbinAttr:      createTaskManagementWorkbinFn,
		getAllTaskManagementWorkbinAttr:      getAllTaskManagementWorkbinFn,
		getTaskManagementWorkbinIdByNameAttr: getTaskManagementWorkbinIdByNameFn,
		getTaskManagementWorkbinByIdAttr:     getTaskManagementWorkbinByIdFn,
		updateTaskManagementWorkbinAttr:      updateTaskManagementWorkbinFn,
		deleteTaskManagementWorkbinAttr:      deleteTaskManagementWorkbinFn,
		workbinCache:                         workbinCache,
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
func (p *taskManagementWorkbinProxy) createTaskManagementWorkbin(ctx context.Context, taskManagementWorkbin *platformclientv2.Workbincreate) (*platformclientv2.Workbin, *platformclientv2.APIResponse, error) {
	return p.createTaskManagementWorkbinAttr(ctx, p, taskManagementWorkbin)
}

// getTaskManagementWorkbin retrieves all Genesys Cloud task management workbin
func (p *taskManagementWorkbinProxy) getAllTaskManagementWorkbin(ctx context.Context) (*[]platformclientv2.Workbin, *platformclientv2.APIResponse, error) {
	return p.getAllTaskManagementWorkbinAttr(ctx, p)
}

// getTaskManagementWorkbinIdByName returns a single Genesys Cloud task management workbin by a name
func (p *taskManagementWorkbinProxy) getTaskManagementWorkbinIdByName(ctx context.Context, name string) (id string, retryable bool, resp *platformclientv2.APIResponse, err error) {
	return p.getTaskManagementWorkbinIdByNameAttr(ctx, p, name)
}

// getTaskManagementWorkbinById returns a single Genesys Cloud task management workbin by Id
func (p *taskManagementWorkbinProxy) getTaskManagementWorkbinById(ctx context.Context, id string) (taskManagementWorkbin *platformclientv2.Workbin, resp *platformclientv2.APIResponse, err error) {
	return p.getTaskManagementWorkbinByIdAttr(ctx, p, id)
}

// updateTaskManagementWorkbin updates a Genesys Cloud task management workbin
func (p *taskManagementWorkbinProxy) updateTaskManagementWorkbin(ctx context.Context, id string, taskManagementWorkbin *platformclientv2.Workbinupdate) (*platformclientv2.Workbin, *platformclientv2.APIResponse, error) {
	return p.updateTaskManagementWorkbinAttr(ctx, p, id, taskManagementWorkbin)
}

// deleteTaskManagementWorkbin deletes a Genesys Cloud task management workbin by Id
func (p *taskManagementWorkbinProxy) deleteTaskManagementWorkbin(ctx context.Context, id string) (resp *platformclientv2.APIResponse, err error) {
	return p.deleteTaskManagementWorkbinAttr(ctx, p, id)
}

// createTaskManagementWorkbinFn is an implementation function for creating a Genesys Cloud task management workbin
func createTaskManagementWorkbinFn(ctx context.Context, p *taskManagementWorkbinProxy, taskManagementWorkbin *platformclientv2.Workbincreate) (*platformclientv2.Workbin, *platformclientv2.APIResponse, error) {
	workbin, resp, err := p.taskManagementApi.PostTaskmanagementWorkbins(*taskManagementWorkbin)
	if err != nil {
		return nil, resp, fmt.Errorf("failed to create task management workbin: %s", err)
	}
	return workbin, resp, nil
}

// getAllTaskManagementWorkbinFn is the implementation for retrieving all task management workbin in Genesys Cloud
func getAllTaskManagementWorkbinFn(ctx context.Context, p *taskManagementWorkbinProxy) (*[]platformclientv2.Workbin, *platformclientv2.APIResponse, error) {
	var allWorkbins []platformclientv2.Workbin
	pageSize := 200
	after := ""
	var response *platformclientv2.APIResponse
	for {
		queryReq := &platformclientv2.Workbinqueryrequest{
			PageSize: &pageSize,
			After:    &after,
		}
		workbins, resp, err := p.taskManagementApi.PostTaskmanagementWorkbinsQuery(*queryReq)
		if err != nil {
			return nil, resp, fmt.Errorf("failed to get workbin: %v", err)
		}
		response = resp
		allWorkbins = append(allWorkbins, *workbins.Entities...)

		// Exit loop if there are no more 'pages'
		if workbins.After == nil || *workbins.After == "" {
			break
		}
		after = *workbins.After
	}
	return &allWorkbins, response, nil
}

// getTaskManagementWorkbinIdByNameFn is an implementation of the function to get a Genesys Cloud task management workbin by name
func getTaskManagementWorkbinIdByNameFn(ctx context.Context, p *taskManagementWorkbinProxy, name string) (id string, retryable bool, resp *platformclientv2.APIResponse, err error) {
	workbins, resp, err := p.getAllTaskManagementWorkbin(ctx)
	if err != nil {
		return "", false, resp, fmt.Errorf("failed to get workbin %s. failed to get all task management workbins", name)
	}

	for _, workbin := range *workbins {
		if *workbin.Name == name {
			return *workbin.Id, false, resp, nil
		}
	}
	return "", true, resp, fmt.Errorf("no task management workbin found with name %s", name)
}

// getTaskManagementWorkbinByIdFn is an implementation of the function to get a Genesys Cloud task management workbin by Id
func getTaskManagementWorkbinByIdFn(ctx context.Context, p *taskManagementWorkbinProxy, id string) (taskManagementWorkbin *platformclientv2.Workbin, resp *platformclientv2.APIResponse, err error) {
	workbin := rc.GetCacheItem(p.workbinCache, id)
	if workbin != nil {
		return workbin, nil, nil
	}
	
	return p.taskManagementApi.GetTaskmanagementWorkbin(id)
}

// updateTaskManagementWorkbinFn is an implementation of the function to update a Genesys Cloud task management workbin
func updateTaskManagementWorkbinFn(ctx context.Context, p *taskManagementWorkbinProxy, id string, taskManagementWorkbin *platformclientv2.Workbinupdate) (*platformclientv2.Workbin, *platformclientv2.APIResponse, error) {
	workbin, resp, err := p.taskManagementApi.PatchTaskmanagementWorkbin(id, *taskManagementWorkbin)
	if err != nil {
		return nil, resp, fmt.Errorf("failed to update task management workbin: %s", err)
	}
	return workbin, resp, nil
}

// deleteTaskManagementWorkbinFn is an implementation function for deleting a Genesys Cloud task management workbin
func deleteTaskManagementWorkbinFn(ctx context.Context, p *taskManagementWorkbinProxy, id string) (resp *platformclientv2.APIResponse, err error) {
	resp, err = p.taskManagementApi.DeleteTaskmanagementWorkbin(id)
	if err != nil {
		return resp, fmt.Errorf("failed to delete task management workbin: %s", err)
	}
	return resp, nil
}
