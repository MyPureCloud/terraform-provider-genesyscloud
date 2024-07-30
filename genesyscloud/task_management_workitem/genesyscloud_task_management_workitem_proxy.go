package task_management_workitem

import (
	"context"
	"fmt"
	"log"

	rc "terraform-provider-genesyscloud/genesyscloud/resource_cache"

	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

/*
The genesyscloud_task_management_workitem_proxy.go file contains the proxy structures and methods that interact
with the Genesys Cloud SDK. We use composition here for each function on the proxy so individual functions can be stubbed
out during testing.
*/

// internalProxy holds a proxy instance that can be used throughout the package
var internalProxy *taskManagementWorkitemProxy

// Type definitions for each func on our proxy so we can easily mock them out later
type createTaskManagementWorkitemFunc func(ctx context.Context, p *taskManagementWorkitemProxy, workitem *platformclientv2.Workitemcreate) (*platformclientv2.Workitem, *platformclientv2.APIResponse, error)
type getAllTaskManagementWorkitemFunc func(ctx context.Context, p *taskManagementWorkitemProxy) (*[]platformclientv2.Workitem, *platformclientv2.APIResponse, error)
type getTaskManagementWorkitemIdByNameFunc func(ctx context.Context, p *taskManagementWorkitemProxy, name string, workbinId string, worktypeId string) (id string, retryable bool, resp *platformclientv2.APIResponse, err error)
type getTaskManagementWorkitemByIdFunc func(ctx context.Context, p *taskManagementWorkitemProxy, id string) (workitem *platformclientv2.Workitem, response *platformclientv2.APIResponse, err error)
type updateTaskManagementWorkitemFunc func(ctx context.Context, p *taskManagementWorkitemProxy, id string, workitem *platformclientv2.Workitemupdate) (*platformclientv2.Workitem, *platformclientv2.APIResponse, error)
type deleteTaskManagementWorkitemFunc func(ctx context.Context, p *taskManagementWorkitemProxy, id string) (response *platformclientv2.APIResponse, err error)

// taskManagementWorkitemProxy contains all of the methods that call genesys cloud APIs.
type taskManagementWorkitemProxy struct {
	clientConfig                          *platformclientv2.Configuration
	taskManagementApi                     *platformclientv2.TaskManagementApi
	createTaskManagementWorkitemAttr      createTaskManagementWorkitemFunc
	getAllTaskManagementWorkitemAttr      getAllTaskManagementWorkitemFunc
	getTaskManagementWorkitemIdByNameAttr getTaskManagementWorkitemIdByNameFunc
	getTaskManagementWorkitemByIdAttr     getTaskManagementWorkitemByIdFunc
	updateTaskManagementWorkitemAttr      updateTaskManagementWorkitemFunc
	deleteTaskManagementWorkitemAttr      deleteTaskManagementWorkitemFunc
	workitemCache                         rc.CacheInterface[platformclientv2.Workitem]
}

// newTaskManagementWorkitemProxy initializes the task management workitem proxy with all of the data needed to communicate with Genesys Cloud
func newTaskManagementWorkitemProxy(clientConfig *platformclientv2.Configuration) *taskManagementWorkitemProxy {
	api := platformclientv2.NewTaskManagementApiWithConfig(clientConfig)
	workitemCache := rc.NewResourceCache[platformclientv2.Workitem]()
	return &taskManagementWorkitemProxy{
		clientConfig:                          clientConfig,
		taskManagementApi:                     api,
		createTaskManagementWorkitemAttr:      createTaskManagementWorkitemFn,
		getAllTaskManagementWorkitemAttr:      getAllTaskManagementWorkitemFn,
		getTaskManagementWorkitemIdByNameAttr: getTaskManagementWorkitemIdByNameFn,
		getTaskManagementWorkitemByIdAttr:     getTaskManagementWorkitemByIdFn,
		updateTaskManagementWorkitemAttr:      updateTaskManagementWorkitemFn,
		deleteTaskManagementWorkitemAttr:      deleteTaskManagementWorkitemFn,
		workitemCache:                         workitemCache,
	}
}

// getTaskManagementWorkitemProxy acts as a singleton to for the internalProxy.  It also ensures
// that we can still proxy our tests by directly setting internalProxy package variable
func getTaskManagementWorkitemProxy(clientConfig *platformclientv2.Configuration) *taskManagementWorkitemProxy {
	if internalProxy == nil {
		internalProxy = newTaskManagementWorkitemProxy(clientConfig)
	}
	return internalProxy
}

// createTaskManagementWorkitem creates a Genesys Cloud task management workitem
func (p *taskManagementWorkitemProxy) createTaskManagementWorkitem(ctx context.Context, taskManagementWorkitem *platformclientv2.Workitemcreate) (*platformclientv2.Workitem, *platformclientv2.APIResponse, error) {
	return p.createTaskManagementWorkitemAttr(ctx, p, taskManagementWorkitem)
}

// getTaskManagementWorkitem retrieves all Genesys Cloud task management workitem
func (p *taskManagementWorkitemProxy) getAllTaskManagementWorkitem(ctx context.Context) (*[]platformclientv2.Workitem, *platformclientv2.APIResponse, error) {
	return p.getAllTaskManagementWorkitemAttr(ctx, p)
}

// getTaskManagementWorkitemIdByName returns a single Genesys Cloud task management workitem by a name
func (p *taskManagementWorkitemProxy) getTaskManagementWorkitemIdByName(ctx context.Context, name string, workbinId string, worktypeId string) (id string, retryable bool, resp *platformclientv2.APIResponse, err error) {
	return p.getTaskManagementWorkitemIdByNameAttr(ctx, p, name, workbinId, worktypeId)
}

// getTaskManagementWorkitemById returns a single Genesys Cloud task management workitem by Id
func (p *taskManagementWorkitemProxy) getTaskManagementWorkitemById(ctx context.Context, id string) (taskManagementWorkitem *platformclientv2.Workitem, resp *platformclientv2.APIResponse, err error) {
	return p.getTaskManagementWorkitemByIdAttr(ctx, p, id)
}

// updateTaskManagementWorkitem updates a Genesys Cloud task management workitem
func (p *taskManagementWorkitemProxy) updateTaskManagementWorkitem(ctx context.Context, id string, taskManagementWorkitem *platformclientv2.Workitemupdate) (*platformclientv2.Workitem, *platformclientv2.APIResponse, error) {
	return p.updateTaskManagementWorkitemAttr(ctx, p, id, taskManagementWorkitem)
}

// deleteTaskManagementWorkitem deletes a Genesys Cloud task management workitem by Id
func (p *taskManagementWorkitemProxy) deleteTaskManagementWorkitem(ctx context.Context, id string) (resp *platformclientv2.APIResponse, err error) {
	return p.deleteTaskManagementWorkitemAttr(ctx, p, id)
}

// createTaskManagementWorkitemFn is an implementation function for creating a Genesys Cloud task management workitem
func createTaskManagementWorkitemFn(ctx context.Context, p *taskManagementWorkitemProxy, taskManagementWorkitem *platformclientv2.Workitemcreate) (*platformclientv2.Workitem, *platformclientv2.APIResponse, error) {
	return p.taskManagementApi.PostTaskmanagementWorkitems(*taskManagementWorkitem)
}

// getAllTaskManagementWorkitemFn is the implementation for retrieving all task management workitem in Genesys Cloud
func getAllTaskManagementWorkitemFn(ctx context.Context, p *taskManagementWorkitemProxy) (*[]platformclientv2.Workitem, *platformclientv2.APIResponse, error) {
	// Workitem query requires one of workbin, assignee, or worktype filter. We'll use workbins.

	// Get all workbins
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
		response = resp
		if err != nil {
			return nil, resp, fmt.Errorf("failed to get workbin: %v", err)
		}
		allWorkbins = append(allWorkbins, *workbins.Entities...)

		// Exit loop if there are no more 'pages'
		if workbins.After == nil || *workbins.After == "" {
			break
		}
		after = *workbins.After
	}

	// Method to query workitems on a workbin
	queryWorkitemsOnWorkbin := func(workbinId string) (*[]platformclientv2.Workitem, error) {
		var wbWorkitems []platformclientv2.Workitem
		pageSize := 200
		after := ""

		for {
			queryReq := &platformclientv2.Workitemquerypostrequest{
				PageSize: &pageSize,
				After:    &after,
				Filters: &[]platformclientv2.Workitemfilter{
					{
						Name:     platformclientv2.String("workbinId"),
						VarType:  platformclientv2.String("String"),
						Operator: platformclientv2.String("EQ"),
						Values:   &[]string{workbinId},
					},
				},
			}
			workitems, resp, err := p.taskManagementApi.PostTaskmanagementWorkitemsQuery(*queryReq)
			response = resp
			if err != nil {
				return nil, fmt.Errorf("failed to get workitems: %v %v", err, resp)
			}
			wbWorkitems = append(wbWorkitems, *workitems.Entities...)

			// Exit loop if there are no more 'pages'
			if workitems.After == nil || *workitems.After == "" {
				break
			}
			after = *workitems.After
		}
		return &wbWorkitems, nil
	}

	// Get all workitems on all workbins
	var allWorkitems []platformclientv2.Workitem
	for _, wb := range allWorkbins {
		wbWorkitems, err := queryWorkitemsOnWorkbin(*wb.Id)
		if err != nil {
			return nil, response, fmt.Errorf("failed to get workitems on workbin %s: %v", *wb.Id, err)
		}
		allWorkitems = append(allWorkitems, *wbWorkitems...)
	}
	return &allWorkitems, response, nil
}

// getTaskManagementWorkitemIdByNameFn is an implementation of the function to get a Genesys Cloud task management workitem by name
func getTaskManagementWorkitemIdByNameFn(ctx context.Context, p *taskManagementWorkitemProxy, name string, workbinId string, worktypeId string) (id string, retryable bool, resp *platformclientv2.APIResponse, err error) {
	pageSize := 100

	// Filter for the workitem name
	queryReq := &platformclientv2.Workitemquerypostrequest{
		PageSize: &pageSize,
		Filters: &[]platformclientv2.Workitemfilter{
			{
				Name:     platformclientv2.String("name"),
				VarType:  platformclientv2.String("String"),
				Operator: platformclientv2.String("EQ"),
				Values:   &[]string{name},
			},
		},
	}

	// Filter for the worktype id
	if worktypeId != "" {
		*queryReq.Filters = append(*queryReq.Filters, platformclientv2.Workitemfilter{
			Name:     platformclientv2.String("typeId"),
			VarType:  platformclientv2.String("String"),
			Operator: platformclientv2.String("EQ"),
			Values:   &[]string{worktypeId},
		})
	}

	// Filter for the workbin id
	if workbinId != "" {
		*queryReq.Filters = append(*queryReq.Filters, platformclientv2.Workitemfilter{
			Name:     platformclientv2.String("workbinId"),
			VarType:  platformclientv2.String("String"),
			Operator: platformclientv2.String("EQ"),
			Values:   &[]string{workbinId},
		})
	}

	workitems, resp, err := p.taskManagementApi.PostTaskmanagementWorkitemsQuery(*queryReq)
	if err != nil {
		return "", false, resp, fmt.Errorf("failed to get worktype %s: %v", name, err)
	}

	if workitems.Entities == nil || len(*workitems.Entities) == 0 {
		return "", true, resp, fmt.Errorf("no task management worktype found with name %s", name)
	}

	workitem := (*workitems.Entities)[0]

	log.Printf("Retrieved the task management worktype id %s by name %s", *workitem.Id, name)
	return *workitem.Id, false, resp, nil
}

// getTaskManagementWorkitemByIdFn is an implementation of the function to get a Genesys Cloud task management workitem by Id
func getTaskManagementWorkitemByIdFn(ctx context.Context, p *taskManagementWorkitemProxy, id string) (taskManagementWorkitem *platformclientv2.Workitem, resp *platformclientv2.APIResponse, err error) {
	workitem := rc.GetCacheItem(p.workitemCache, id)
	if workitem != nil {
		return workitem, nil, nil
	}

	return p.taskManagementApi.GetTaskmanagementWorkitem(id, "")
}

// updateTaskManagementWorkitemFn is an implementation of the function to update a Genesys Cloud task management workitem
func updateTaskManagementWorkitemFn(ctx context.Context, p *taskManagementWorkitemProxy, id string, taskManagementWorkitem *platformclientv2.Workitemupdate) (*platformclientv2.Workitem, *platformclientv2.APIResponse, error) {
	return p.taskManagementApi.PatchTaskmanagementWorkitem(id, *taskManagementWorkitem)
}

// deleteTaskManagementWorkitemFn is an implementation function for deleting a Genesys Cloud task management workitem
func deleteTaskManagementWorkitemFn(ctx context.Context, p *taskManagementWorkitemProxy, id string) (resp *platformclientv2.APIResponse, err error) {
	return p.taskManagementApi.DeleteTaskmanagementWorkitem(id)
}
