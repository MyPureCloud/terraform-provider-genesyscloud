package task_management_worktype

import (
	"context"
	"fmt"
	"log"
	rc "terraform-provider-genesyscloud/genesyscloud/resource_cache"

	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

/*
The genesyscloud_task_management_worktype_proxy.go file contains the proxy structures and methods that interact
with the Genesys Cloud SDK. We use composition here for each function on the proxy so individual functions can be stubbed
out during testing.
*/

// internalProxy holds a proxy instance that can be used throughout the package
var internalProxy *TaskManagementWorktypeProxy

// Type definitions for each func on our proxy so we can easily mock them out later
type createTaskManagementWorktypeFunc func(ctx context.Context, p *TaskManagementWorktypeProxy, worktype *platformclientv2.Worktypecreate) (*platformclientv2.Worktype, *platformclientv2.APIResponse, error)
type getAllTaskManagementWorktypeFunc func(ctx context.Context, p *TaskManagementWorktypeProxy) (*[]platformclientv2.Worktype, *platformclientv2.APIResponse, error)
type getTaskManagementWorktypeIdByNameFunc func(ctx context.Context, p *TaskManagementWorktypeProxy, name string) (id string, retryable bool, resp *platformclientv2.APIResponse, err error)
type getTaskManagementWorktypeByNameFunc func(ctx context.Context, p *TaskManagementWorktypeProxy, name string) (workItemType *platformclientv2.Worktype, retryable bool, resp *platformclientv2.APIResponse, err error)
type getTaskManagementWorktypeByIdFunc func(ctx context.Context, p *TaskManagementWorktypeProxy, id string) (worktype *platformclientv2.Worktype, response *platformclientv2.APIResponse, err error)
type updateTaskManagementWorktypeFunc func(ctx context.Context, p *TaskManagementWorktypeProxy, id string, worktype *platformclientv2.Worktypeupdate) (*platformclientv2.Worktype, *platformclientv2.APIResponse, error)
type deleteTaskManagementWorktypeFunc func(ctx context.Context, p *TaskManagementWorktypeProxy, id string) (response *platformclientv2.APIResponse, err error)

// TaskManagementWorktypeProxy contains all the methods that call genesys cloud APIs.
type TaskManagementWorktypeProxy struct {
	clientConfig                          *platformclientv2.Configuration
	taskManagementApi                     *platformclientv2.TaskManagementApi
	createTaskManagementWorktypeAttr      createTaskManagementWorktypeFunc
	getAllTaskManagementWorktypeAttr      getAllTaskManagementWorktypeFunc
	getTaskManagementWorktypeIdByNameAttr getTaskManagementWorktypeIdByNameFunc
	getTaskManagementWorktypeByIdAttr     getTaskManagementWorktypeByIdFunc
	getTaskManagementWorktypeByNameAttr   getTaskManagementWorktypeByNameFunc
	updateTaskManagementWorktypeAttr      updateTaskManagementWorktypeFunc
	deleteTaskManagementWorktypeAttr      deleteTaskManagementWorktypeFunc
	worktypeCache                         rc.CacheInterface[platformclientv2.Worktype]
}

// newTaskManagementWorktypeProxy initializes the task management worktype proxy with all the data needed to communicate with Genesys Cloud
func newTaskManagementWorktypeProxy(clientConfig *platformclientv2.Configuration) *TaskManagementWorktypeProxy {
	api := platformclientv2.NewTaskManagementApiWithConfig(clientConfig)
	worktypeCache := rc.NewResourceCache[platformclientv2.Worktype]()
	return &TaskManagementWorktypeProxy{
		clientConfig:                          clientConfig,
		taskManagementApi:                     api,
		createTaskManagementWorktypeAttr:      createTaskManagementWorktypeFn,
		getAllTaskManagementWorktypeAttr:      getAllTaskManagementWorktypeFn,
		getTaskManagementWorktypeIdByNameAttr: getTaskManagementWorktypeIdByNameFn,
		getTaskManagementWorktypeByNameAttr:   getTaskManagementWorktypeByNameFn,
		getTaskManagementWorktypeByIdAttr:     getTaskManagementWorktypeByIdFn,
		updateTaskManagementWorktypeAttr:      updateTaskManagementWorktypeFn,
		deleteTaskManagementWorktypeAttr:      deleteTaskManagementWorktypeFn,
		worktypeCache:                         worktypeCache,
	}
}

// GetTaskManagementWorktypeProxy acts as a singleton to for the internalProxy.  It also ensures
// that we can still proxy our tests by directly setting internalProxy package variable
func GetTaskManagementWorktypeProxy(clientConfig *platformclientv2.Configuration) *TaskManagementWorktypeProxy {
	if internalProxy == nil {
		internalProxy = newTaskManagementWorktypeProxy(clientConfig)
	}
	return internalProxy
}

// createTaskManagementWorktype creates a Genesys Cloud task management worktype
func (p *TaskManagementWorktypeProxy) createTaskManagementWorktype(ctx context.Context, taskManagementWorktype *platformclientv2.Worktypecreate) (*platformclientv2.Worktype, *platformclientv2.APIResponse, error) {
	return p.createTaskManagementWorktypeAttr(ctx, p, taskManagementWorktype)
}

// GetAllTaskManagementWorktype retrieves all Genesys Cloud task management worktype
func (p *TaskManagementWorktypeProxy) GetAllTaskManagementWorktype(ctx context.Context) (*[]platformclientv2.Worktype, *platformclientv2.APIResponse, error) {
	return p.getAllTaskManagementWorktypeAttr(ctx, p)
}

// getTaskManagementWorktypeIdByName returns a single Genesys Cloud task management worktype by a name
func (p *TaskManagementWorktypeProxy) getTaskManagementWorktypeIdByName(ctx context.Context, name string) (id string, retryable bool, resp *platformclientv2.APIResponse, err error) {
	return p.getTaskManagementWorktypeIdByNameAttr(ctx, p, name)
}

// getTaskManagementWorktypeByName returns a single Genesys Cloud task management worktype by a name
func (p *TaskManagementWorktypeProxy) getTaskManagementWorktypeByName(ctx context.Context, name string) (workItemType *platformclientv2.Worktype, retryable bool, resp *platformclientv2.APIResponse, err error) {
	return p.getTaskManagementWorktypeByNameAttr(ctx, p, name)
}

// GetTaskManagementWorktypeById returns a single Genesys Cloud task management worktype by Id
func (p *TaskManagementWorktypeProxy) GetTaskManagementWorktypeById(ctx context.Context, id string) (taskManagementWorktype *platformclientv2.Worktype, resp *platformclientv2.APIResponse, err error) {
	return p.getTaskManagementWorktypeByIdAttr(ctx, p, id)
}

// UpdateTaskManagementWorktype updates a Genesys Cloud task management worktype
func (p *TaskManagementWorktypeProxy) UpdateTaskManagementWorktype(ctx context.Context, id string, taskManagementWorktype *platformclientv2.Worktypeupdate) (*platformclientv2.Worktype, *platformclientv2.APIResponse, error) {
	return p.updateTaskManagementWorktypeAttr(ctx, p, id, taskManagementWorktype)
}

// deleteTaskManagementWorktype deletes a Genesys Cloud task management worktype by Id
func (p *TaskManagementWorktypeProxy) deleteTaskManagementWorktype(ctx context.Context, id string) (resp *platformclientv2.APIResponse, err error) {
	return p.deleteTaskManagementWorktypeAttr(ctx, p, id)
}

// createTaskManagementWorktypeFn is an implementation function for creating a Genesys Cloud task management worktype
func createTaskManagementWorktypeFn(ctx context.Context, p *TaskManagementWorktypeProxy, taskManagementWorktype *platformclientv2.Worktypecreate) (*platformclientv2.Worktype, *platformclientv2.APIResponse, error) {
	return p.taskManagementApi.PostTaskmanagementWorktypes(*taskManagementWorktype)
}

// getAllTaskManagementWorktypeFn is the implementation for retrieving all task management worktype in Genesys Cloud
func getAllTaskManagementWorktypeFn(ctx context.Context, p *TaskManagementWorktypeProxy) (*[]platformclientv2.Worktype, *platformclientv2.APIResponse, error) {
	var allWorktypes []platformclientv2.Worktype
	pageSize := 200
	after := ""
	var response *platformclientv2.APIResponse
	for {
		queryReq := &platformclientv2.Worktypequeryrequest{
			PageSize: &pageSize,
			After:    &after,
		}
		worktypes, resp, err := p.taskManagementApi.PostTaskmanagementWorktypesQuery(*queryReq)
		response = resp
		if err != nil {
			return nil, resp, fmt.Errorf("failed to get worktypes: %v", err)
		}
		allWorktypes = append(allWorktypes, *worktypes.Entities...)

		// Exit loop if there are no more 'pages'
		if worktypes.After == nil || *worktypes.After == "" {
			break
		}
		after = *worktypes.After
	}
	return &allWorktypes, response, nil
}

// getWorkType looks up a worktype by name
func getWorkType(name string, p *TaskManagementWorktypeProxy) (*platformclientv2.Worktype, bool, *platformclientv2.APIResponse, error) {
	pageSize := 100

	filterType := "String"
	filterOperator := "EQ"
	filterNameKey := "name"
	filterNameValues := []string{name}

	queryReq := &platformclientv2.Worktypequeryrequest{
		PageSize: &pageSize,
		Filters: &[]platformclientv2.Workitemfilter{
			// Filter for the worktype name
			platformclientv2.Workitemfilter{
				Name:     &filterNameKey,
				VarType:  &filterType,
				Operator: &filterOperator,
				Values:   &filterNameValues,
			},
		},
	}

	worktypes, resp, err := p.taskManagementApi.PostTaskmanagementWorktypesQuery(*queryReq)
	if err != nil {
		return nil, false, resp, fmt.Errorf("failed to get worktype %s: %v", name, err)
	}

	if worktypes.Entities == nil || len(*worktypes.Entities) == 0 {
		return nil, true, resp, fmt.Errorf("no task management worktype found with name %s", name)
	}

	if len(*worktypes.Entities) > 1 {
		return nil, true, resp, fmt.Errorf("%d workitem types have been found with the same name: %s .  Unable to resolve to a single id", len(*worktypes.Entities), name)
	}

	workType := (*worktypes.Entities)[0]

	return &workType, false, resp, nil
}

// getTaskManagementWorktypeIdByNameFn is an implementation of the function to get a Genesys Cloud task management worktype by name
func getTaskManagementWorktypeIdByNameFn(ctx context.Context, p *TaskManagementWorktypeProxy, name string) (id string, retryable bool, resp *platformclientv2.APIResponse, err error) {
	worktype, retry, resp, err := getWorkType(name, p)
	if err != nil {
		return "", retry, resp, err
	}

	log.Printf("Retrieved the task management worktype id %s by name %s", *worktype.Id, name)
	return *worktype.Id, false, resp, nil
}

// getTaskManagementWorktypeByNameFn Retrieves the full worktype item rather than just the id
func getTaskManagementWorktypeByNameFn(ctx context.Context, p *TaskManagementWorktypeProxy, name string) (workItemType *platformclientv2.Worktype, retryable bool, resp *platformclientv2.APIResponse, err error) {
	worktype, retry, resp, err := getWorkType(name, p)
	if err != nil {
		return nil, retry, resp, err
	}

	log.Printf("Retrieved the task management worktype %s by name %s", *worktype.Id, name)
	return worktype, false, resp, nil
}

// getTaskManagementWorktypeByIdFn is an implementation of the function to get a Genesys Cloud task management worktype by Id
func getTaskManagementWorktypeByIdFn(ctx context.Context, p *TaskManagementWorktypeProxy, id string) (taskManagementWorktype *platformclientv2.Worktype, resp *platformclientv2.APIResponse, err error) {
	worktype := rc.GetCacheItem(p.worktypeCache, id)
	if worktype != nil {
		return worktype, nil, nil
	}
	
	return p.taskManagementApi.GetTaskmanagementWorktype(id, []string{})
}

// updateTaskManagementWorktypeFn is an implementation of the function to update a Genesys Cloud task management worktype
func updateTaskManagementWorktypeFn(ctx context.Context, p *TaskManagementWorktypeProxy, id string, worktypeUpdate *platformclientv2.Worktypeupdate) (*platformclientv2.Worktype, *platformclientv2.APIResponse, error) {
	return p.taskManagementApi.PatchTaskmanagementWorktype(id, *worktypeUpdate)
}

// deleteTaskManagementWorktypeFn is an implementation function for deleting a Genesys Cloud task management worktype
func deleteTaskManagementWorktypeFn(ctx context.Context, p *TaskManagementWorktypeProxy, id string) (resp *platformclientv2.APIResponse, err error) {
	return p.taskManagementApi.DeleteTaskmanagementWorktype(id)
}
