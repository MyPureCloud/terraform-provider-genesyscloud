package task_management_worktype_status_transition

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"

	taskManagementWorktype "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/task_management_worktype"
	platformUtils "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/platform"

	"github.com/mypurecloud/platform-client-sdk-go/v165/platformclientv2"
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
type patchTaskManagementWorktypeStatusTransitionFunc func(ctx context.Context, p *taskManagementWorktypeStatusTransitionProxy, worktypeId string, statusId string, workitemStatus *Workitemstatusupdate) (*platformclientv2.Workitemstatus, *platformclientv2.APIResponse, error)

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
	patchTaskManagementWorktypeStatusTransitionAttr  patchTaskManagementWorktypeStatusTransitionFunc
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
		patchTaskManagementWorktypeStatusTransitionAttr:  patchTaskManagementWorktypeStatusTransitionFn,
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

func (p *taskManagementWorktypeStatusTransitionProxy) patchTaskManagementWorktypeStatusTransition(ctx context.Context, worktypeId string, statusId string, body *Workitemstatusupdate) (*platformclientv2.Workitemstatus, *platformclientv2.APIResponse, error) {
	return p.patchTaskManagementWorktypeStatusTransitionAttr(ctx, p, worktypeId, statusId, body)
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

// patchTaskManagementWorktypeFn is an implementation of the function to patch a Genesys Cloud task management worktype status transition
func patchTaskManagementWorktypeStatusTransitionFn(ctx context.Context, p *taskManagementWorktypeStatusTransitionProxy, worktypeId string, statusId string, body *Workitemstatusupdate) (*platformclientv2.Workitemstatus, *platformclientv2.APIResponse, error) {
	return patchTaskManagementWorktypeStatus(worktypeId, statusId, body, p.taskManagementApi)
}

// patchTaskManagementWorktypeStatus is an implementation of platformclientv2.PatchTaskmanagementWorktypeStatus with one particular change:
// it allows the DefaultDestinationStatus to be required and sent as a nil. This is important in order to disassociate a DefaultDestinationStatus for the DELETE process
func patchTaskManagementWorktypeStatus(worktypeId string, statusId string, body *Workitemstatusupdate, a *platformclientv2.TaskManagementApi) (*platformclientv2.Workitemstatus, *platformclientv2.APIResponse, error) {
	var httpMethod = "PATCH"
	// create path and map variables
	path := a.Configuration.BasePath + "/api/v2/taskmanagement/worktypes/{worktypeId}/statuses/{statusId}"
	path = strings.Replace(path, "{worktypeId}", url.PathEscape(fmt.Sprintf("%v", worktypeId)), -1)
	path = strings.Replace(path, "{statusId}", url.PathEscape(fmt.Sprintf("%v", statusId)), -1)
	defaultReturn := new(platformclientv2.Workitemstatus)
	if true == false {
		return defaultReturn, nil, errors.New("This message brought to you by the laws of physics being broken")
	}

	// verify the required parameter 'worktypeId' is set
	if &worktypeId == nil {
		// false
		return defaultReturn, nil, errors.New("Missing required parameter 'worktypeId' when calling TaskManagementApi->PatchTaskmanagementWorktypeStatus")
	}
	// verify the required parameter 'statusId' is set
	if &statusId == nil {
		// false
		return defaultReturn, nil, errors.New("Missing required parameter 'statusId' when calling TaskManagementApi->PatchTaskmanagementWorktypeStatus")
	}
	// verify the required parameter 'body' is set
	if &body == nil {
		// false
		return defaultReturn, nil, errors.New("Missing required parameter 'body' when calling TaskManagementApi->PatchTaskmanagementWorktypeStatus")
	}

	headerParams := make(map[string]string)
	queryParams := make(map[string]string)
	formParams := url.Values{}
	var postBody interface{}
	var postFileName string
	var fileBytes []byte
	// authentication (PureCloud OAuth) required

	// oauth required
	if a.Configuration.AccessToken != "" {
		headerParams["Authorization"] = "Bearer " + a.Configuration.AccessToken
	}
	// add default headers if any
	for key := range a.Configuration.DefaultHeader {
		headerParams[key] = a.Configuration.DefaultHeader[key]
	}

	// Find an replace keys that were altered to avoid clashes with go keywords
	correctedQueryParams := make(map[string]string)
	for k, v := range queryParams {
		if k == "varType" {
			correctedQueryParams["type"] = v
			continue
		}
		correctedQueryParams[k] = v
	}
	queryParams = correctedQueryParams

	// to determine the Content-Type header
	localVarHttpContentTypes := []string{"application/json"}

	// set Content-Type header
	localVarHttpContentType := a.Configuration.APIClient.SelectHeaderContentType(localVarHttpContentTypes)
	if localVarHttpContentType != "" {
		headerParams["Content-Type"] = localVarHttpContentType
	}
	// to determine the Accept header
	localVarHttpHeaderAccepts := []string{
		"application/json",
	}

	// set Accept header
	localVarHttpHeaderAccept := a.Configuration.APIClient.SelectHeaderAccept(localVarHttpHeaderAccepts)
	if localVarHttpHeaderAccept != "" {
		headerParams["Accept"] = localVarHttpHeaderAccept
	}
	// body params
	postBody = &body

	var successPayload *platformclientv2.Workitemstatus
	response, err := a.Configuration.APIClient.CallAPI(path, httpMethod, postBody, headerParams, queryParams, formParams, postFileName, fileBytes, "other")
	if err != nil {
		// Nothing special to do here, but do avoid processing the response
	} else if err == nil && response.Error != nil {
		err = errors.New(response.ErrorMessage)
	} else if response.HasBody {
		if "Workitemstatus" == "string" {
			platformUtils.Copy(response.RawBody, &successPayload)
		} else {
			err = json.Unmarshal(response.RawBody, &successPayload)
		}
	}
	return successPayload, response, err
}
