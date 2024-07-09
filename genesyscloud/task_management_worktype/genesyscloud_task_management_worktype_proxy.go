package task_management_worktype

import (
	"context"
	"fmt"
	"log"

	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

/*
The genesyscloud_task_management_worktype_proxy.go file contains the proxy structures and methods that interact
with the Genesys Cloud SDK. We use composition here for each function on the proxy so individual functions can be stubbed
out during testing.
*/

// internalProxy holds a proxy instance that can be used throughout the package
var internalProxy *taskManagementWorktypeProxy

// Type definitions for each func on our proxy so we can easily mock them out later

type createTaskManagementWorktypeFunc func(ctx context.Context, p *taskManagementWorktypeProxy, worktype *platformclientv2.Worktypecreate) (*platformclientv2.Worktype, *platformclientv2.APIResponse, error)
type getAllTaskManagementWorktypeFunc func(ctx context.Context, p *taskManagementWorktypeProxy) (*[]platformclientv2.Worktype, *platformclientv2.APIResponse, error)
type getTaskManagementWorktypeIdByNameFunc func(ctx context.Context, p *taskManagementWorktypeProxy, name string) (id string, retryable bool, resp *platformclientv2.APIResponse, err error)
type getTaskManagementWorktypeByNameFunc func(ctx context.Context, p *taskManagementWorktypeProxy, name string) (workItemType *platformclientv2.Worktype, retryable bool, resp *platformclientv2.APIResponse, err error)
type getTaskManagementWorktypeByIdFunc func(ctx context.Context, p *taskManagementWorktypeProxy, id string) (worktype *platformclientv2.Worktype, response *platformclientv2.APIResponse, err error)
type updateTaskManagementWorktypeFunc func(ctx context.Context, p *taskManagementWorktypeProxy, id string, worktype *platformclientv2.Worktypeupdate) (*platformclientv2.Worktype, *platformclientv2.APIResponse, error)
type deleteTaskManagementWorktypeFunc func(ctx context.Context, p *taskManagementWorktypeProxy, id string) (response *platformclientv2.APIResponse, err error)
type getAllTaskManagementWorktypeStatusesFunc func(ctx context.Context, p *taskManagementWorktypeProxy, worktypeId string) (*platformclientv2.Workitemstatuslisting, *platformclientv2.APIResponse, error)
type createTaskManagementWorktypeStatusFunc func(ctx context.Context, p *taskManagementWorktypeProxy, worktypeId string, status *platformclientv2.Workitemstatuscreate) (*platformclientv2.Workitemstatus, *platformclientv2.APIResponse, error)
type updateTaskManagementWorktypeStatusFunc func(ctx context.Context, p *taskManagementWorktypeProxy, worktypeId string, statusId string, statusUpdate *platformclientv2.Workitemstatusupdate) (*platformclientv2.Workitemstatus, *platformclientv2.APIResponse, error)
type deleteTaskManagementWorktypeStatusFunc func(ctx context.Context, p *taskManagementWorktypeProxy, worktypeId string, statusId string) (response *platformclientv2.APIResponse, err error)

// taskManagementWorktypeProxy contains all of the methods that call genesys cloud APIs.
type taskManagementWorktypeProxy struct {
	clientConfig                          *platformclientv2.Configuration
	taskManagementApi                     *platformclientv2.TaskManagementApi
	createTaskManagementWorktypeAttr      createTaskManagementWorktypeFunc
	getAllTaskManagementWorktypeAttr      getAllTaskManagementWorktypeFunc
	getTaskManagementWorktypeIdByNameAttr getTaskManagementWorktypeIdByNameFunc
	getTaskManagementWorktypeByIdAttr     getTaskManagementWorktypeByIdFunc
	getTaskManagementWorktypeByNameAttr   getTaskManagementWorktypeByNameFunc
	updateTaskManagementWorktypeAttr      updateTaskManagementWorktypeFunc
	deleteTaskManagementWorktypeAttr      deleteTaskManagementWorktypeFunc

	getAllTaskManagementWorktypeStatusesAttr getAllTaskManagementWorktypeStatusesFunc
	createTaskManagementWorktypeStatusAttr   createTaskManagementWorktypeStatusFunc
	updateTaskManagementWorktypeStatusAttr   updateTaskManagementWorktypeStatusFunc
	deleteTaskManagementWorktypeStatusAttr   deleteTaskManagementWorktypeStatusFunc
}

// newTaskManagementWorktypeProxy initializes the task management worktype proxy with all of the data needed to communicate with Genesys Cloud
func newTaskManagementWorktypeProxy(clientConfig *platformclientv2.Configuration) *taskManagementWorktypeProxy {
	api := platformclientv2.NewTaskManagementApiWithConfig(clientConfig)
	return &taskManagementWorktypeProxy{
		clientConfig:                          clientConfig,
		taskManagementApi:                     api,
		createTaskManagementWorktypeAttr:      createTaskManagementWorktypeFn,
		getAllTaskManagementWorktypeAttr:      getAllTaskManagementWorktypeFn,
		getTaskManagementWorktypeIdByNameAttr: getTaskManagementWorktypeIdByNameFn,
		getTaskManagementWorktypeByNameAttr:   getTaskManagementWorktypeByNameFn,
		getTaskManagementWorktypeByIdAttr:     getTaskManagementWorktypeByIdFn,
		updateTaskManagementWorktypeAttr:      updateTaskManagementWorktypeFn,
		deleteTaskManagementWorktypeAttr:      deleteTaskManagementWorktypeFn,

		getAllTaskManagementWorktypeStatusesAttr: getAllTaskManagementWorktypeStatusesFn,
		createTaskManagementWorktypeStatusAttr:   createTaskManagementWorktypeStatusFn,
		updateTaskManagementWorktypeStatusAttr:   updateTaskManagementWorktypeStatusFn,
		deleteTaskManagementWorktypeStatusAttr:   deleteTaskManagementWorktypeStatusFn,
	}
}

// getTaskManagementWorktypeProxy acts as a singleton to for the internalProxy.  It also ensures
// that we can still proxy our tests by directly setting internalProxy package variable
func getTaskManagementWorktypeProxy(clientConfig *platformclientv2.Configuration) *taskManagementWorktypeProxy {
	if internalProxy == nil {
		internalProxy = newTaskManagementWorktypeProxy(clientConfig)
	}
	return internalProxy
}

// createTaskManagementWorktype creates a Genesys Cloud task management worktype
func (p *taskManagementWorktypeProxy) createTaskManagementWorktype(ctx context.Context, taskManagementWorktype *platformclientv2.Worktypecreate) (*platformclientv2.Worktype, *platformclientv2.APIResponse, error) {
	return p.createTaskManagementWorktypeAttr(ctx, p, taskManagementWorktype)
}

// getTaskManagementWorktype retrieves all Genesys Cloud task management worktype
func (p *taskManagementWorktypeProxy) getAllTaskManagementWorktype(ctx context.Context) (*[]platformclientv2.Worktype, *platformclientv2.APIResponse, error) {
	return p.getAllTaskManagementWorktypeAttr(ctx, p)
}

// getTaskManagementWorktypeIdByName returns a single Genesys Cloud task management worktype by a name
func (p *taskManagementWorktypeProxy) getTaskManagementWorktypeIdByName(ctx context.Context, name string) (id string, retryable bool, resp *platformclientv2.APIResponse, err error) {
	return p.getTaskManagementWorktypeIdByNameAttr(ctx, p, name)
}

// getTaskManagementWorktypeByName returns a single Genesys Cloud task management worktype by a name
func (p *taskManagementWorktypeProxy) getTaskManagementWorktypeByName(ctx context.Context, name string) (workItemType *platformclientv2.Worktype, retryable bool, resp *platformclientv2.APIResponse, err error) {
	return p.getTaskManagementWorktypeByNameAttr(ctx, p, name)
}

// getTaskManagementWorktypeById returns a single Genesys Cloud task management worktype by Id
func (p *taskManagementWorktypeProxy) getTaskManagementWorktypeById(ctx context.Context, id string) (taskManagementWorktype *platformclientv2.Worktype, resp *platformclientv2.APIResponse, err error) {
	return p.getTaskManagementWorktypeByIdAttr(ctx, p, id)
}

// updateTaskManagementWorktype updates a Genesys Cloud task management worktype
func (p *taskManagementWorktypeProxy) updateTaskManagementWorktype(ctx context.Context, id string, taskManagementWorktype *platformclientv2.Worktypeupdate) (*platformclientv2.Worktype, *platformclientv2.APIResponse, error) {
	return p.updateTaskManagementWorktypeAttr(ctx, p, id, taskManagementWorktype)
}

// deleteTaskManagementWorktype deletes a Genesys Cloud task management worktype by Id
func (p *taskManagementWorktypeProxy) deleteTaskManagementWorktype(ctx context.Context, id string) (resp *platformclientv2.APIResponse, err error) {
	return p.deleteTaskManagementWorktypeAttr(ctx, p, id)
}

// createTaskManagementWorktypeStatus creates a Genesys Cloud task management worktype status
func (p *taskManagementWorktypeProxy) getAllTaskManagementWorktypeStatuses(ctx context.Context, worktypeId string) (*platformclientv2.Workitemstatuslisting, *platformclientv2.APIResponse, error) {
	return p.getAllTaskManagementWorktypeStatusesAttr(ctx, p, worktypeId)

} // createTaskManagementWorktypeStatus creates a Genesys Cloud task management worktype status
func (p *taskManagementWorktypeProxy) createTaskManagementWorktypeStatus(ctx context.Context, worktypeId string, status *platformclientv2.Workitemstatuscreate) (*platformclientv2.Workitemstatus, *platformclientv2.APIResponse, error) {
	return p.createTaskManagementWorktypeStatusAttr(ctx, p, worktypeId, status)
}

// updateTaskManagementWorktypeStatus updates a Genesys Cloud task management worktype status
func (p *taskManagementWorktypeProxy) updateTaskManagementWorktypeStatus(ctx context.Context, worktypeId string, statusId string, statusUpdate *platformclientv2.Workitemstatusupdate) (*platformclientv2.Workitemstatus, *platformclientv2.APIResponse, error) {
	return p.updateTaskManagementWorktypeStatusAttr(ctx, p, worktypeId, statusId, statusUpdate)
}

// deleteTaskManagementWorktypeStatus deletes a Genesys Cloud task management worktype status by Id
func (p *taskManagementWorktypeProxy) deleteTaskManagementWorktypeStatus(ctx context.Context, worktypeId string, statusId string) (response *platformclientv2.APIResponse, err error) {
	return p.deleteTaskManagementWorktypeStatusAttr(ctx, p, worktypeId, statusId)
}

// createTaskManagementWorktypeFn is an implementation function for creating a Genesys Cloud task management worktype
func createTaskManagementWorktypeFn(ctx context.Context, p *taskManagementWorktypeProxy, taskManagementWorktype *platformclientv2.Worktypecreate) (*platformclientv2.Worktype, *platformclientv2.APIResponse, error) {
	return p.taskManagementApi.PostTaskmanagementWorktypes(*taskManagementWorktype)
}

// getAllTaskManagementWorktypeFn is the implementation for retrieving all task management worktype in Genesys Cloud
func getAllTaskManagementWorktypeFn(ctx context.Context, p *taskManagementWorktypeProxy) (*[]platformclientv2.Worktype, *platformclientv2.APIResponse, error) {
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
func getWorkType(name string, p *taskManagementWorktypeProxy) (*platformclientv2.Worktype, bool, *platformclientv2.APIResponse, error) {
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
func getTaskManagementWorktypeIdByNameFn(ctx context.Context, p *taskManagementWorktypeProxy, name string) (id string, retryable bool, resp *platformclientv2.APIResponse, err error) {
	worktype, retry, resp, err := getWorkType(name, p)
	if err != nil {
		return "", retry, resp, err
	}

	log.Printf("Retrieved the task management worktype id %s by name %s", *worktype.Id, name)
	return *worktype.Id, false, resp, nil
}

// getTaskManagementWorktypeByNameFn Retrieves the full worktype item rather than just the id
func getTaskManagementWorktypeByNameFn(ctx context.Context, p *taskManagementWorktypeProxy, name string) (workItemType *platformclientv2.Worktype, retryable bool, resp *platformclientv2.APIResponse, err error) {
	worktype, retry, resp, err := getWorkType(name, p)
	if err != nil {
		return nil, retry, resp, err
	}

	log.Printf("Retrieved the task management worktype %s by name %s", *worktype.Id, name)
	return worktype, false, resp, nil
}

// getTaskManagementWorktypeByIdFn is an implementation of the function to get a Genesys Cloud task management worktype by Id
func getTaskManagementWorktypeByIdFn(ctx context.Context, p *taskManagementWorktypeProxy, id string) (taskManagementWorktype *platformclientv2.Worktype, resp *platformclientv2.APIResponse, err error) {
	return p.taskManagementApi.GetTaskmanagementWorktype(id, []string{})
}

// updateTaskManagementWorktypeFn is an implementation of the function to update a Genesys Cloud task management worktype
func updateTaskManagementWorktypeFn(ctx context.Context, p *taskManagementWorktypeProxy, id string, worktypeUpdate *platformclientv2.Worktypeupdate) (*platformclientv2.Worktype, *platformclientv2.APIResponse, error) {
	return p.taskManagementApi.PatchTaskmanagementWorktype(id, *worktypeUpdate)
}

// deleteTaskManagementWorktypeFn is an implementation function for deleting a Genesys Cloud task management worktype
func deleteTaskManagementWorktypeFn(ctx context.Context, p *taskManagementWorktypeProxy, id string) (resp *platformclientv2.APIResponse, err error) {
	return p.taskManagementApi.DeleteTaskmanagementWorktype(id)
}

// getAllTaskManagementWorktypeStatusesFn is an implementation function for getting all statues for a Genesys Cloud task management worktype
func getAllTaskManagementWorktypeStatusesFn(ctx context.Context, p *taskManagementWorktypeProxy, worktypeId string) (*platformclientv2.Workitemstatuslisting, *platformclientv2.APIResponse, error) {
	return p.taskManagementApi.GetTaskmanagementWorktypeStatuses(worktypeId)
}

// createTaskManagementWorktypeStatusFn is an implementation function for creating a Genesys Cloud task management worktype status
func createTaskManagementWorktypeStatusFn(ctx context.Context, p *taskManagementWorktypeProxy, worktypeId string, status *platformclientv2.Workitemstatuscreate) (*platformclientv2.Workitemstatus, *platformclientv2.APIResponse, error) {
	return p.taskManagementApi.PostTaskmanagementWorktypeStatuses(worktypeId, *status)
}

// updateTaskManagementWorktypeStatusFn is an implementation of the function to update a Genesys Cloud task management worktype status
func updateTaskManagementWorktypeStatusFn(ctx context.Context, p *taskManagementWorktypeProxy, worktypeId string, statusId string, statusUpdate *platformclientv2.Workitemstatusupdate) (*platformclientv2.Workitemstatus, *platformclientv2.APIResponse, error) {
	return p.taskManagementApi.PatchTaskmanagementWorktypeStatus(worktypeId, statusId, *statusUpdate)
}

// deleteTaskManagementWorktypeStatusFn is an implementation function for deleting a Genesys Cloud task management worktype status
func deleteTaskManagementWorktypeStatusFn(ctx context.Context, p *taskManagementWorktypeProxy, worktypeId string, statusId string) (response *platformclientv2.APIResponse, err error) {
	return p.taskManagementApi.DeleteTaskmanagementWorktypeStatus(worktypeId, statusId)
}
