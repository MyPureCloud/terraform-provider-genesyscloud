package task_management_worktype

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/mypurecloud/platform-client-sdk-go/v119/platformclientv2"
)

/*
The genesyscloud_task_management_worktype_proxy.go file contains the proxy structures and methods that interact
with the Genesys Cloud SDK. We use composition here for each function on the proxy so individual functions can be stubbed
out during testing.
*/

// internalProxy holds a proxy instance that can be used throughout the package
var internalProxy *taskManagementWorktypeProxy

// Type definitions for each func on our proxy so we can easily mock them out later
type createTaskManagementWorktypeFunc func(ctx context.Context, p *taskManagementWorktypeProxy, worktype *platformclientv2.Worktypecreate) (*platformclientv2.Worktype, error)
type getAllTaskManagementWorktypeFunc func(ctx context.Context, p *taskManagementWorktypeProxy) (*[]platformclientv2.Worktype, error)
type getTaskManagementWorktypeIdByNameFunc func(ctx context.Context, p *taskManagementWorktypeProxy, name string) (id string, retryable bool, err error)
type getTaskManagementWorktypeByIdFunc func(ctx context.Context, p *taskManagementWorktypeProxy, id string) (worktype *Worktype, responseCode int, err error)
type updateTaskManagementWorktypeFunc func(ctx context.Context, p *taskManagementWorktypeProxy, id string, worktype *platformclientv2.Worktypeupdate) (*platformclientv2.Worktype, error)
type deleteTaskManagementWorktypeFunc func(ctx context.Context, p *taskManagementWorktypeProxy, id string) (responseCode int, err error)

type createTaskManagementWorktypeStatusFunc func(ctx context.Context, p *taskManagementWorktypeProxy, worktypeId string, status *Workitemstatuscreate) (*Workitemstatus, error)
type updateTaskManagementWorktypeStatusFunc func(ctx context.Context, p *taskManagementWorktypeProxy, worktypeId string, statusId string, statusUpdate *Workitemstatusupdate) (*Workitemstatus, error)
type deleteTaskManagementWorktypeStatusFunc func(ctx context.Context, p *taskManagementWorktypeProxy, worktypeId string, statusId string) (responseCode int, err error)

// taskManagementWorktypeProxy contains all of the methods that call genesys cloud APIs.
type taskManagementWorktypeProxy struct {
	clientConfig                          *platformclientv2.Configuration
	taskManagementApi                     *platformclientv2.TaskManagementApi
	createTaskManagementWorktypeAttr      createTaskManagementWorktypeFunc
	getAllTaskManagementWorktypeAttr      getAllTaskManagementWorktypeFunc
	getTaskManagementWorktypeIdByNameAttr getTaskManagementWorktypeIdByNameFunc
	getTaskManagementWorktypeByIdAttr     getTaskManagementWorktypeByIdFunc
	updateTaskManagementWorktypeAttr      updateTaskManagementWorktypeFunc
	deleteTaskManagementWorktypeAttr      deleteTaskManagementWorktypeFunc

	createTaskManagementWorktypeStatusAttr createTaskManagementWorktypeStatusFunc
	updateTaskManagementWorktypeStatusAttr updateTaskManagementWorktypeStatusFunc
	deleteTaskManagementWorktypeStatusAttr deleteTaskManagementWorktypeStatusFunc
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
		getTaskManagementWorktypeByIdAttr:     getTaskManagementWorktypeByIdFn,
		updateTaskManagementWorktypeAttr:      updateTaskManagementWorktypeFn,
		deleteTaskManagementWorktypeAttr:      deleteTaskManagementWorktypeFn,

		createTaskManagementWorktypeStatusAttr: createTaskManagementWorktypeStatusFn,
		updateTaskManagementWorktypeStatusAttr: updateTaskManagementWorktypeStatusFn,
		deleteTaskManagementWorktypeStatusAttr: deleteTaskManagementWorktypeStatusFn,
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
func (p *taskManagementWorktypeProxy) createTaskManagementWorktype(ctx context.Context, taskManagementWorktype *platformclientv2.Worktypecreate) (*platformclientv2.Worktype, error) {
	return p.createTaskManagementWorktypeAttr(ctx, p, taskManagementWorktype)
}

// getTaskManagementWorktype retrieves all Genesys Cloud task management worktype
func (p *taskManagementWorktypeProxy) getAllTaskManagementWorktype(ctx context.Context) (*[]platformclientv2.Worktype, error) {
	return p.getAllTaskManagementWorktypeAttr(ctx, p)
}

// getTaskManagementWorktypeIdByName returns a single Genesys Cloud task management worktype by a name
func (p *taskManagementWorktypeProxy) getTaskManagementWorktypeIdByName(ctx context.Context, name string) (id string, retryable bool, err error) {
	return p.getTaskManagementWorktypeIdByNameAttr(ctx, p, name)
}

// getTaskManagementWorktypeById returns a single Genesys Cloud task management worktype by Id
func (p *taskManagementWorktypeProxy) getTaskManagementWorktypeById(ctx context.Context, id string) (taskManagementWorktype *Worktype, statusCode int, err error) {
	return p.getTaskManagementWorktypeByIdAttr(ctx, p, id)
}

// updateTaskManagementWorktype updates a Genesys Cloud task management worktype
func (p *taskManagementWorktypeProxy) updateTaskManagementWorktype(ctx context.Context, id string, taskManagementWorktype *platformclientv2.Worktypeupdate) (*platformclientv2.Worktype, error) {
	return p.updateTaskManagementWorktypeAttr(ctx, p, id, taskManagementWorktype)
}

// deleteTaskManagementWorktype deletes a Genesys Cloud task management worktype by Id
func (p *taskManagementWorktypeProxy) deleteTaskManagementWorktype(ctx context.Context, id string) (statusCode int, err error) {
	return p.deleteTaskManagementWorktypeAttr(ctx, p, id)
}

// createTaskManagementWorktypeStatus creates a Genesys Cloud task management worktype status
func (p *taskManagementWorktypeProxy) createTaskManagementWorktypeStatus(ctx context.Context, worktypeId string, status *Workitemstatuscreate) (*Workitemstatus, error) {
	return p.createTaskManagementWorktypeStatusAttr(ctx, p, worktypeId, status)
}

// updateTaskManagementWorktypeStatus updates a Genesys Cloud task management worktype status
func (p *taskManagementWorktypeProxy) updateTaskManagementWorktypeStatus(ctx context.Context, worktypeId string, statusId string, statusUpdate *Workitemstatusupdate) (*Workitemstatus, error) {
	return p.updateTaskManagementWorktypeStatusAttr(ctx, p, worktypeId, statusId, statusUpdate)
}

// deleteTaskManagementWorktypeStatus deletes a Genesys Cloud task management worktype status by Id
func (p *taskManagementWorktypeProxy) deleteTaskManagementWorktypeStatus(ctx context.Context, worktypeId string, statusId string) (responseCode int, err error) {
	return p.deleteTaskManagementWorktypeStatusAttr(ctx, p, worktypeId, statusId)
}

// createTaskManagementWorktypeFn is an implementation function for creating a Genesys Cloud task management worktype
func createTaskManagementWorktypeFn(ctx context.Context, p *taskManagementWorktypeProxy, taskManagementWorktype *platformclientv2.Worktypecreate) (*platformclientv2.Worktype, error) {
	log.Printf("Creating task management worktype: %s", *taskManagementWorktype.Name)
	worktype, resp, err := p.taskManagementApi.PostTaskmanagementWorktypes(*taskManagementWorktype)
	log.Printf("Completed call to create task management worktype %s with status code %d, correlation id %s and err %s", *taskManagementWorktype.Name, resp.StatusCode, resp.CorrelationID, err)
	if err != nil {
		return nil, fmt.Errorf("failed to create task management worktype: %s", err)
	}

	return worktype, nil
}

// getAllTaskManagementWorktypeFn is the implementation for retrieving all task management worktype in Genesys Cloud
func getAllTaskManagementWorktypeFn(ctx context.Context, p *taskManagementWorktypeProxy) (*[]platformclientv2.Worktype, error) {
	var allWorktypes []platformclientv2.Worktype
	pageSize := 200
	after := ""

	for {
		queryReq := &platformclientv2.Worktypequeryrequest{
			PageSize: &pageSize,
			After:    &after,
		}
		worktypes, _, err := p.taskManagementApi.PostTaskmanagementWorktypesQuery(*queryReq)
		if err != nil {
			return nil, fmt.Errorf("failed to get worktypes: %v", err)
		}
		allWorktypes = append(allWorktypes, *worktypes.Entities...)

		// Exit loop if there are no more 'pages'
		if worktypes.After == nil || *worktypes.After == "" {
			break
		}
		after = *worktypes.After
	}

	return &allWorktypes, nil
}

// getTaskManagementWorktypeIdByNameFn is an implementation of the function to get a Genesys Cloud task management worktype by name
func getTaskManagementWorktypeIdByNameFn(ctx context.Context, p *taskManagementWorktypeProxy, name string) (id string, retryable bool, err error) {
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

	worktypes, _, err := p.taskManagementApi.PostTaskmanagementWorktypesQuery(*queryReq)
	if err != nil {
		return "", false, fmt.Errorf("failed to get worktype %s: %v", name, err)
	}

	if worktypes.Entities == nil || len(*worktypes.Entities) == 0 {
		return "", true, fmt.Errorf("no task management worktype found with name %s", name)
	}

	worktype := (*worktypes.Entities)[0]

	log.Printf("Retrieved the task management worktype id %s by name %s", *worktype.Id, name)
	return *worktype.Id, false, nil
}

// getTaskManagementWorktypeByIdFn is an implementation of the function to get a Genesys Cloud task management worktype by Id
func getTaskManagementWorktypeByIdFn(ctx context.Context, p *taskManagementWorktypeProxy, id string) (taskManagementWorktype *Worktype, statusCode int, err error) {
	apiClient := &p.clientConfig.APIClient

	// create path and map variables
	path := p.clientConfig.BasePath + "/api/v2/taskmanagement/worktypes/" + id

	headerParams := make(map[string]string)
	queryParams := make(map[string]string)

	// oauth required
	if p.clientConfig.AccessToken != "" {
		headerParams["Authorization"] = "Bearer " + p.clientConfig.AccessToken
	}
	// add default headers if any
	for key := range p.clientConfig.DefaultHeader {
		headerParams[key] = p.clientConfig.DefaultHeader[key]
	}

	headerParams["Content-Type"] = "application/json"
	headerParams["Accept"] = "application/json"

	var successPayload Worktype
	response, err := apiClient.CallAPI(path, http.MethodGet, nil, headerParams, queryParams, nil, "", nil)
	if err != nil {
		return nil, response.StatusCode, fmt.Errorf("failed to get worktype %s: %v", id, err)
	}
	if response.Error != nil {
		return nil, response.StatusCode, fmt.Errorf("failed to get worktype %s: %v", id, errors.New(response.ErrorMessage))
	}

	err = json.Unmarshal([]byte(response.RawBody), &successPayload)
	if err != nil {
		return nil, response.StatusCode, fmt.Errorf("failed to get worktype %s: %v", id, err)
	}

	return &successPayload, response.StatusCode, nil
}

// updateTaskManagementWorktypeFn is an implementation of the function to update a Genesys Cloud task management worktype
func updateTaskManagementWorktypeFn(ctx context.Context, p *taskManagementWorktypeProxy, id string, worktypeUpdate *platformclientv2.Worktypeupdate) (*platformclientv2.Worktype, error) {
	worktype, _, err := p.taskManagementApi.PatchTaskmanagementWorktype(id, *worktypeUpdate)
	if err != nil {
		return nil, fmt.Errorf("failed to update task management worktype %s: %v", id, err)
	}
	return worktype, nil
}

// deleteTaskManagementWorktypeFn is an implementation function for deleting a Genesys Cloud task management worktype
func deleteTaskManagementWorktypeFn(ctx context.Context, p *taskManagementWorktypeProxy, id string) (statusCode int, err error) {
	resp, err := p.taskManagementApi.DeleteTaskmanagementWorktype(id)
	if err != nil {
		return resp.StatusCode, fmt.Errorf("failed to delete task management worktype %s: %v", id, err)
	}

	return resp.StatusCode, nil
}

// createTaskManagementWorktypeStatusFn is an implementation function for creating a Genesys Cloud task management worktype status
func createTaskManagementWorktypeStatusFn(ctx context.Context, p *taskManagementWorktypeProxy, worktypeId string, status *Workitemstatuscreate) (*Workitemstatus, error) {
	apiClient := &p.clientConfig.APIClient

	// create path and map variables
	path := p.clientConfig.BasePath + "/api/v2/taskmanagement/worktypes/" + worktypeId + "/statuses"

	headerParams := make(map[string]string)
	queryParams := make(map[string]string)

	// oauth required
	if p.clientConfig.AccessToken != "" {
		headerParams["Authorization"] = "Bearer " + p.clientConfig.AccessToken
	}
	// add default headers if any
	for key := range p.clientConfig.DefaultHeader {
		headerParams[key] = p.clientConfig.DefaultHeader[key]
	}

	headerParams["Content-Type"] = "application/json"
	headerParams["Accept"] = "application/json"

	var successPayload Workitemstatus
	log.Printf("Creating task management worktype status: %s", *status.Name)
	response, err := apiClient.CallAPI(path, http.MethodPost, status, headerParams, queryParams, nil, "", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create status %s for worktype %s: %v", *status.Name, worktypeId, err)
	}
	if response.Error != nil {
		return nil, fmt.Errorf("failed to create status %s for worktype %s: %v", *status.Name, worktypeId, errors.New(response.ErrorMessage))
	}
	log.Printf("Completed call to create task management worktype status %s with status code %d, correlation id %s and err %s", *status.Name, response.StatusCode, response.CorrelationID, err)

	err = json.Unmarshal([]byte(response.RawBody), &successPayload)
	if err != nil {
		return nil, fmt.Errorf("failed to get newly created worktype status %s of worktype %s: %v", *status.Name, worktypeId, err)
	}

	return &successPayload, nil
}

// updateTaskManagementWorktypeStatusFn is an implementation of the function to update a Genesys Cloud task management worktype status
func updateTaskManagementWorktypeStatusFn(ctx context.Context, p *taskManagementWorktypeProxy, worktypeId string, statusId string, statusUpdate *Workitemstatusupdate) (*Workitemstatus, error) {
	apiClient := &p.clientConfig.APIClient

	// create path and map variables
	path := p.clientConfig.BasePath + "/api/v2/taskmanagement/worktypes/" + worktypeId + "/statuses/" + statusId

	headerParams := make(map[string]string)
	queryParams := make(map[string]string)

	// oauth required
	if p.clientConfig.AccessToken != "" {
		headerParams["Authorization"] = "Bearer " + p.clientConfig.AccessToken
	}
	// add default headers if any
	for key := range p.clientConfig.DefaultHeader {
		headerParams[key] = p.clientConfig.DefaultHeader[key]
	}

	headerParams["Content-Type"] = "application/json"
	headerParams["Accept"] = "application/json"

	var successPayload Workitemstatus
	response, err := apiClient.CallAPI(path, http.MethodPatch, statusUpdate, headerParams, queryParams, nil, "", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to update status %s for worktype %s: %v", statusId, worktypeId, err)
	}
	if response.Error != nil {
		return nil, fmt.Errorf("failed to update status %s for worktype %s: %v", statusId, worktypeId, errors.New(response.ErrorMessage))
	}

	err = json.Unmarshal([]byte(response.RawBody), &successPayload)
	if err != nil {
		return nil, fmt.Errorf("failed to get newly created worktype status %s of worktype %s: %v", *statusUpdate.Name, worktypeId, err)
	}

	return &successPayload, nil
}

// deleteTaskManagementWorktypeStatusFn is an implementation function for deleting a Genesys Cloud task management worktype status
func deleteTaskManagementWorktypeStatusFn(ctx context.Context, p *taskManagementWorktypeProxy, worktypeId string, statusId string) (responseCode int, err error) {
	resp, err := p.taskManagementApi.DeleteTaskmanagementWorktypeStatus(worktypeId, statusId)
	if err != nil {
		return resp.StatusCode, fmt.Errorf("failed to delete task management worktype status %s: %v", statusId, err)
	}

	return resp.StatusCode, nil
}
