package task_management_workitem_schema

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"

	rc "terraform-provider-genesyscloud/genesyscloud/resource_cache"

	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

/*
The genesyscloud_task_management_workitem_schema_proxy.go file contains the proxy structures and methods that interact
with the Genesys Cloud SDK. We use composition here for each function on the proxy so individual functions can be stubbed
out during testing.
*/

// internalProxy holds a proxy instance that can be used throughout the package
var internalProxy *taskManagementProxy

// Type definitions for each func on our proxy so we can easily mock them out later
type createTaskManagementWorkitemSchemaFunc func(ctx context.Context, p *taskManagementProxy, schema *platformclientv2.Dataschema) (*platformclientv2.Dataschema, *platformclientv2.APIResponse, error)
type getAllTaskManagementWorkitemSchemaFunc func(ctx context.Context, p *taskManagementProxy) (*[]platformclientv2.Dataschema, *platformclientv2.APIResponse, error)
type getTaskManagementWorkitemSchemasByNameFunc func(ctx context.Context, p *taskManagementProxy, name string) (schemas *[]platformclientv2.Dataschema, retryable bool, resp *platformclientv2.APIResponse, err error)
type getTaskManagementWorkitemSchemaByIdFunc func(ctx context.Context, p *taskManagementProxy, id string) (schema *platformclientv2.Dataschema, response *platformclientv2.APIResponse, err error)
type updateTaskManagementWorkitemSchemaFunc func(ctx context.Context, p *taskManagementProxy, id string, schema *platformclientv2.Dataschema) (*platformclientv2.Dataschema, *platformclientv2.APIResponse, error)
type deleteTaskManagementWorkitemSchemaFunc func(ctx context.Context, p *taskManagementProxy, id string) (response *platformclientv2.APIResponse, err error)
type getTaskManagementWorkitemSchemaDeletedStatusFunc func(ctx context.Context, p *taskManagementProxy, schemaId string) (isDeleted bool, resp *platformclientv2.APIResponse, err error)

// taskManagementProxy contains all of the methods that call genesys cloud APIs.
type taskManagementProxy struct {
	clientConfig                                     *platformclientv2.Configuration
	taskManagementApi                                *platformclientv2.TaskManagementApi
	createTaskManagementWorkitemSchemaAttr           createTaskManagementWorkitemSchemaFunc
	getAllTaskManagementWorkitemSchemaAttr           getAllTaskManagementWorkitemSchemaFunc
	getTaskManagementWorkitemSchemasByNameAttr       getTaskManagementWorkitemSchemasByNameFunc
	getTaskManagementWorkitemSchemaByIdAttr          getTaskManagementWorkitemSchemaByIdFunc
	updateTaskManagementWorkitemSchemaAttr           updateTaskManagementWorkitemSchemaFunc
	deleteTaskManagementWorkitemSchemaAttr           deleteTaskManagementWorkitemSchemaFunc
	getTaskManagementWorkitemSchemaDeletedStatusAttr getTaskManagementWorkitemSchemaDeletedStatusFunc
	workitemSchemaCache                              rc.CacheInterface[platformclientv2.Dataschema]
}

// newTaskManagementProxy initializes the task management proxy with all of the data needed to communicate with Genesys Cloud
func newTaskManagementProxy(clientConfig *platformclientv2.Configuration) *taskManagementProxy {
	api := platformclientv2.NewTaskManagementApiWithConfig(clientConfig)
	workitemSchemaCache := rc.NewResourceCache[platformclientv2.Dataschema]()

	return &taskManagementProxy{
		clientConfig:                                     clientConfig,
		taskManagementApi:                                api,
		createTaskManagementWorkitemSchemaAttr:           createTaskManagementWorkitemSchemaFn,
		getAllTaskManagementWorkitemSchemaAttr:           getAllTaskManagementWorkitemSchemaFn,
		getTaskManagementWorkitemSchemasByNameAttr:       getTaskManagementWorkitemSchemasByNameFn,
		getTaskManagementWorkitemSchemaByIdAttr:          getTaskManagementWorkitemSchemaByIdFn,
		updateTaskManagementWorkitemSchemaAttr:           updateTaskManagementWorkitemSchemaFn,
		deleteTaskManagementWorkitemSchemaAttr:           deleteTaskManagementWorkitemSchemaFn,
		getTaskManagementWorkitemSchemaDeletedStatusAttr: getTaskManagementWorkitemSchemaDeletedStatusFn,
		workitemSchemaCache:                              workitemSchemaCache,
	}
}

// getTaskManagementProxy acts as a singleton to for the internalProxy.  It also ensures
// that we can still proxy our tests by directly setting internalProxy package variable
func getTaskManagementProxy(clientConfig *platformclientv2.Configuration) *taskManagementProxy {
	if internalProxy == nil {
		internalProxy = newTaskManagementProxy(clientConfig)
	}
	return internalProxy
}

// createTaskManagementWorkitemSchema creates a Genesys Cloud task management workitem schema
func (p *taskManagementProxy) createTaskManagementWorkitemSchema(ctx context.Context, schema *platformclientv2.Dataschema) (*platformclientv2.Dataschema, *platformclientv2.APIResponse, error) {
	return p.createTaskManagementWorkitemSchemaAttr(ctx, p, schema)
}

// getAllTaskManagementWorkitemSchema retrieves all Genesys Cloud task management workitem schemas
func (p *taskManagementProxy) getAllTaskManagementWorkitemSchema(ctx context.Context) (*[]platformclientv2.Dataschema, *platformclientv2.APIResponse, error) {
	return p.getAllTaskManagementWorkitemSchemaAttr(ctx, p)
}

// getTaskManagementWorkitemSchemaIdByName returns a single Genesys Cloud task management workitem schema by a name
func (p *taskManagementProxy) getTaskManagementWorkitemSchemasByName(ctx context.Context, name string) (schemas *[]platformclientv2.Dataschema, retryable bool, resp *platformclientv2.APIResponse, err error) {
	return p.getTaskManagementWorkitemSchemasByNameAttr(ctx, p, name)
}

// getTaskManagementWorkitemSchemaById returns a single Genesys Cloud task management workitem schema by Id
func (p *taskManagementProxy) getTaskManagementWorkitemSchemaById(ctx context.Context, id string) (schema *platformclientv2.Dataschema, resp *platformclientv2.APIResponse, err error) {
	return p.getTaskManagementWorkitemSchemaByIdAttr(ctx, p, id)
}

// updateTaskManagementWorkitemSchema updates a Genesys Cloud task management workitem schema
func (p *taskManagementProxy) updateTaskManagementWorkitemSchema(ctx context.Context, id string, schemaUpdate *platformclientv2.Dataschema) (*platformclientv2.Dataschema, *platformclientv2.APIResponse, error) {
	return p.updateTaskManagementWorkitemSchemaAttr(ctx, p, id, schemaUpdate)
}

// deleteTaskManagementWorkitemSchema deletes a Genesys Cloud task management workitem schema by Id
func (p *taskManagementProxy) deleteTaskManagementWorkitemSchema(ctx context.Context, id string) (resp *platformclientv2.APIResponse, err error) {
	return p.deleteTaskManagementWorkitemSchemaAttr(ctx, p, id)
}

// getTaskManagementWorkitemSchemaDeletedStatus gets the deleted status of a Genesys Cloud task management workitem schema
func (p *taskManagementProxy) getTaskManagementWorkitemSchemaDeletedStatus(ctx context.Context, schemaId string) (isDeleted bool, resp *platformclientv2.APIResponse, err error) {
	return p.getTaskManagementWorkitemSchemaDeletedStatusAttr(ctx, p, schemaId)
}

// createTaskManagementWorkitemSchemaFn is an implementation function for creating a Genesys Cloud task management workitem schema
func createTaskManagementWorkitemSchemaFn(ctx context.Context, p *taskManagementProxy, schema *platformclientv2.Dataschema) (*platformclientv2.Dataschema, *platformclientv2.APIResponse, error) {
	log.Printf("Creating task management workitem schema: %s", *schema.Name)
	createdSchema, resp, err := p.taskManagementApi.PostTaskmanagementWorkitemsSchemas(*schema)
	log.Printf("Completed call to create task management workitem schema %s with status code %d, correlation id %s and err %s", *schema.Name, resp.StatusCode, resp.CorrelationID, err)
	if err != nil {
		return nil, resp, fmt.Errorf("failed to create task management workitem schema: %s", err)
	}
	return createdSchema, resp, nil
}

// getAllTaskManagementWorkitemSchemaFn is the implementation for retrieving all task management workitem schemas in Genesys Cloud
func getAllTaskManagementWorkitemSchemaFn(ctx context.Context, p *taskManagementProxy) (*[]platformclientv2.Dataschema, *platformclientv2.APIResponse, error) {
	// NOTE: At the time of implementation (Preview API) retrieving schemas does not have any sort of pagination.
	// It seemingly will return all schemas in one call. This might have to be updated as there may be some
	// undocumented limit or if there would be changes to the API call before release.

	schemas, resp, err := p.taskManagementApi.GetTaskmanagementWorkitemsSchemas()
	if err != nil {
		return nil, resp, fmt.Errorf("failed to get all workitem schemas: %v", err)
	}
	if schemas.Entities == nil || *schemas.Total == 0 {
		return &([]platformclientv2.Dataschema{}), resp, nil
	}
	return schemas.Entities, resp, nil
}

// getTaskManagementWorkitemSchemasByNameFn is an implementation of the function to get a Genesys Cloud task management workitem schemas by name
func getTaskManagementWorkitemSchemasByNameFn(ctx context.Context, p *taskManagementProxy, name string) (matchingSchemas *[]platformclientv2.Dataschema, retryable bool, resp *platformclientv2.APIResponse, err error) {
	finalSchemas := []platformclientv2.Dataschema{}

	schemas, resp, err := p.getAllTaskManagementWorkitemSchema(ctx)
	if err != nil {
		return nil, false, resp, err
	}

	for _, schema := range *schemas {
		if schema.Name != nil && *schema.Name == name {
			finalSchemas = append(finalSchemas, schema)
		}
	}

	if len(finalSchemas) == 0 {
		return nil, true, resp, fmt.Errorf("no task management workitem schema found with name %s", name)
	}
	return &finalSchemas, false, resp, nil
}

// getTaskManagementWorkitemSchemaByIdFn is an implementation of the function to get a Genesys Cloud task management workitem schema by Id
func getTaskManagementWorkitemSchemaByIdFn(ctx context.Context, p *taskManagementProxy, id string) (schema *platformclientv2.Dataschema, resp *platformclientv2.APIResponse, err error) {
	workitemSchema := rc.GetCacheItem(p.workitemSchemaCache, id)
	if workitemSchema != nil {
		return schema, nil, nil
	}
	return p.taskManagementApi.GetTaskmanagementWorkitemsSchema(id)
}

// updateTaskManagementWorkitemSchemaFn is an implementation of the function to update a Genesys Cloud task management workitem schema
func updateTaskManagementWorkitemSchemaFn(ctx context.Context, p *taskManagementProxy, id string, schemaUpdate *platformclientv2.Dataschema) (*platformclientv2.Dataschema, *platformclientv2.APIResponse, error) {
	schema, resp, err := p.taskManagementApi.PutTaskmanagementWorkitemsSchema(id, *schemaUpdate)
	if err != nil {
		return nil, resp, fmt.Errorf("failed to update task management workitem schema: %s", err)
	}
	return schema, resp, nil
}

// deleteTaskManagementWorkitemSchemaFn is an implementation function for deleting a Genesys Cloud task management workitem schema
func deleteTaskManagementWorkitemSchemaFn(ctx context.Context, p *taskManagementProxy, id string) (resp *platformclientv2.APIResponse, err error) {
	resp, err = p.taskManagementApi.DeleteTaskmanagementWorkitemsSchema(id)
	if err != nil {
		return resp, fmt.Errorf("failed to delete task management workitem schema: %s", err)
	}
	return resp, nil
}

// getTaskManagementWorkitemSchemaDeletedStatusFn is an implementation function to get the 'deleted' status of a Genesys Cloud task management workitem schema
func getTaskManagementWorkitemSchemaDeletedStatusFn(ctx context.Context, p *taskManagementProxy, schemaId string) (isDeleted bool, resp *platformclientv2.APIResponse, err error) {
	apiClient := &p.clientConfig.APIClient

	// create path and map variables
	path := p.clientConfig.BasePath + "/api/v2/taskmanagement/workitems/schemas/" + schemaId

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

	var successPayload map[string]interface{}
	response, err := apiClient.CallAPI(path, http.MethodGet, nil, headerParams, queryParams, nil, "", nil)
	if err != nil {
		return false, response, fmt.Errorf("failed to get workitem schema %s: %v", schemaId, err)
	}
	if response.Error != nil {
		return false, response, fmt.Errorf("failed to get workitem schema %s: %v", schemaId, errors.New(response.ErrorMessage))
	}

	err = json.Unmarshal([]byte(response.RawBody), &successPayload)
	if err != nil {
		return false, response, fmt.Errorf("failed to get deleted status of %s: %v", schemaId, err)
	}

	// Manually query for the 'deleted' property because it is removed when
	// response JSON body becomes SDK Dataschema object.
	if isDeleted, ok := successPayload["deleted"].(bool); ok {
		return isDeleted, response, nil
	}

	return false, response, fmt.Errorf("failed to get deleted status of %s: %v", schemaId, err)
}
