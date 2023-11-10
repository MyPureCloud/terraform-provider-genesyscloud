package task_management_workitem_schema

import (
	"context"
	"fmt"

	"github.com/mypurecloud/platform-client-sdk-go/v115/platformclientv2"
)

/*
The genesyscloud_task_management_workitem_schema_proxy.go file contains the proxy structures and methods that interact
with the Genesys Cloud SDK. We use composition here for each function on the proxy so individual functions can be stubbed
out during testing.
*/

// internalProxy holds a proxy instance that can be used throughout the package
var internalProxy *taskManagementProxy

// Type definitions for each func on our proxy so we can easily mock them out later
type createTaskManagementWorkitemSchemaFunc func(ctx context.Context, p *taskManagementProxy, schema *platformclientv2.Dataschema) (*platformclientv2.Dataschema, error)
type getAllTaskManagementWorkitemSchemaFunc func(ctx context.Context, p *taskManagementProxy) (*[]platformclientv2.Dataschema, error)
type getTaskManagementWorkitemSchemasByNameFunc func(ctx context.Context, p *taskManagementProxy, name string) (schemas *[]platformclientv2.Dataschema, retryable bool, err error)
type getTaskManagementWorkitemSchemaByIdFunc func(ctx context.Context, p *taskManagementProxy, id string) (schema *platformclientv2.Dataschema, responseCode int, err error)
type updateTaskManagementWorkitemSchemaFunc func(ctx context.Context, p *taskManagementProxy, id string, schema *platformclientv2.Dataschema) (*platformclientv2.Dataschema, error)
type deleteTaskManagementWorkitemSchemaFunc func(ctx context.Context, p *taskManagementProxy, id string) (responseCode int, err error)

// taskManagementProxy contains all of the methods that call genesys cloud APIs.
type taskManagementProxy struct {
	clientConfig                               *platformclientv2.Configuration
	taskManagementApi                          *platformclientv2.TaskManagementApi
	createTaskManagementWorkitemSchemaAttr     createTaskManagementWorkitemSchemaFunc
	getAllTaskManagementWorkitemSchemaAttr     getAllTaskManagementWorkitemSchemaFunc
	getTaskManagementWorkitemSchemasByNameAttr getTaskManagementWorkitemSchemasByNameFunc
	getTaskManagementWorkitemSchemaByIdAttr    getTaskManagementWorkitemSchemaByIdFunc
	updateTaskManagementWorkitemSchemaAttr     updateTaskManagementWorkitemSchemaFunc
	deleteTaskManagementWorkitemSchemaAttr     deleteTaskManagementWorkitemSchemaFunc
}

// newTaskManagementProxy initializes the task management proxy with all of the data needed to communicate with Genesys Cloud
func newTaskManagementProxy(clientConfig *platformclientv2.Configuration) *taskManagementProxy {
	api := platformclientv2.NewTaskManagementApiWithConfig(clientConfig)
	return &taskManagementProxy{
		clientConfig:                               clientConfig,
		taskManagementApi:                          api,
		createTaskManagementWorkitemSchemaAttr:     createTaskManagementWorkitemSchemaFn,
		getAllTaskManagementWorkitemSchemaAttr:     getAllTaskManagementWorkitemSchemaFn,
		getTaskManagementWorkitemSchemasByNameAttr: getTaskManagementWorkitemSchemasByNameFn,
		getTaskManagementWorkitemSchemaByIdAttr:    getTaskManagementWorkitemSchemaByIdFn,
		updateTaskManagementWorkitemSchemaAttr:     updateTaskManagementWorkitemSchemaFn,
		deleteTaskManagementWorkitemSchemaAttr:     deleteTaskManagementWorkitemSchemaFn,
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
func (p *taskManagementProxy) createTaskManagementWorkitemSchema(ctx context.Context, schema *platformclientv2.Dataschema) (*platformclientv2.Dataschema, error) {
	return p.createTaskManagementWorkitemSchemaAttr(ctx, p, schema)
}

// getAllTaskManagementWorkitemSchema retrieves all Genesys Cloud task management workitem schemas
func (p *taskManagementProxy) getAllTaskManagementWorkitemSchema(ctx context.Context) (*[]platformclientv2.Dataschema, error) {
	return p.getAllTaskManagementWorkitemSchemaAttr(ctx, p)
}

// getTaskManagementWorkitemSchemaIdByName returns a single Genesys Cloud task management workitem schema by a name
func (p *taskManagementProxy) getTaskManagementWorkitemSchemasByName(ctx context.Context, name string) (schemas *[]platformclientv2.Dataschema, retryable bool, err error) {
	return p.getTaskManagementWorkitemSchemasByNameAttr(ctx, p, name)
}

// getTaskManagementWorkitemSchemaById returns a single Genesys Cloud task management workitem schema by Id
func (p *taskManagementProxy) getTaskManagementWorkitemSchemaById(ctx context.Context, id string) (schema *platformclientv2.Dataschema, statusCode int, err error) {
	return p.getTaskManagementWorkitemSchemaByIdAttr(ctx, p, id)
}

// updateTaskManagementWorkitemSchema updates a Genesys Cloud task management workitem schema
func (p *taskManagementProxy) updateTaskManagementWorkitemSchema(ctx context.Context, id string, schemaUpdate *platformclientv2.Dataschema) (*platformclientv2.Dataschema, error) {
	return p.updateTaskManagementWorkitemSchemaAttr(ctx, p, id, schemaUpdate)
}

// deleteTaskManagementWorkitemSchema deletes a Genesys Cloud task management workitem schema by Id
func (p *taskManagementProxy) deleteTaskManagementWorkitemSchema(ctx context.Context, id string) (statusCode int, err error) {
	return p.deleteTaskManagementWorkitemSchemaAttr(ctx, p, id)
}

// createTaskManagementWorkitemSchemaFn is an implementation function for creating a Genesys Cloud task management workitem schema
func createTaskManagementWorkitemSchemaFn(ctx context.Context, p *taskManagementProxy, schema *platformclientv2.Dataschema) (*platformclientv2.Dataschema, error) {
	createdSchema, _, err := p.taskManagementApi.PostTaskmanagementWorkitemsSchemas(*schema)
	if err != nil {
		return nil, fmt.Errorf("failed to create task management workitem schema: %s", err)
	}

	return createdSchema, nil
}

// getAllTaskManagementWorkitemSchemaFn is the implementation for retrieving all task management workitem schemas in Genesys Cloud
func getAllTaskManagementWorkitemSchemaFn(ctx context.Context, p *taskManagementProxy) (*[]platformclientv2.Dataschema, error) {
	// NOTE: At the time of implementation (Preview API) retrieving schemas does not have any sort of pagination.
	// It seemingly will return all schemas in one call. This might have to be updated as there may be some
	// undocumented limit or if there would be changes to the API call before release.

	schemas, _, err := p.taskManagementApi.GetTaskmanagementWorkitemsSchemas()
	if err != nil {
		return nil, fmt.Errorf("failed to get all workitem schemas: %v", err)
	}
	if schemas.Entities == nil || *schemas.Total == 0 {
		return &([]platformclientv2.Dataschema{}), nil
	}

	return schemas.Entities, nil
}

// getTaskManagementWorkitemSchemasByNameFn is an implementation of the function to get a Genesys Cloud task management workitem schemas by name
func getTaskManagementWorkitemSchemasByNameFn(ctx context.Context, p *taskManagementProxy, name string) (matchingSchemas *[]platformclientv2.Dataschema, retryable bool, err error) {
	finalSchemas := []platformclientv2.Dataschema{}

	schemas, err := p.getAllTaskManagementWorkitemSchema(ctx)
	if err != nil {
		return nil, false, err
	}

	for _, schema := range *schemas {
		if schema.Name != nil && *schema.Name == name {
			finalSchemas = append(finalSchemas, schema)
		}
	}

	if len(finalSchemas) == 0 {
		return nil, true, fmt.Errorf("no task management workitem schema found with name %s", name)
	}

	return &finalSchemas, false, nil
}

// getTaskManagementWorkitemSchemaByIdFn is an implementation of the function to get a Genesys Cloud task management workitem schema by Id
func getTaskManagementWorkitemSchemaByIdFn(ctx context.Context, p *taskManagementProxy, id string) (schema *platformclientv2.Dataschema, statusCode int, err error) {
	schema, resp, err := p.taskManagementApi.GetTaskmanagementWorkitemsSchema(id)
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("failed to retrieve task management workitem schema by id %s: %v", id, err)
	}

	return schema, resp.StatusCode, nil
}

// updateTaskManagementWorkitemSchemaFn is an implementation of the function to update a Genesys Cloud task management workitem schema
func updateTaskManagementWorkitemSchemaFn(ctx context.Context, p *taskManagementProxy, id string, schemaUpdate *platformclientv2.Dataschema) (*platformclientv2.Dataschema, error) {
	schema, _, err := p.taskManagementApi.PutTaskmanagementWorkitemsSchema(id, *schemaUpdate)
	if err != nil {
		return nil, fmt.Errorf("failed to update task management workitem schema: %s", err)
	}
	return schema, nil
}

// deleteTaskManagementWorkitemSchemaFn is an implementation function for deleting a Genesys Cloud task management workitem schema
func deleteTaskManagementWorkitemSchemaFn(ctx context.Context, p *taskManagementProxy, id string) (statusCode int, err error) {
	resp, err := p.taskManagementApi.DeleteTaskmanagementWorkitemsSchema(id)
	if err != nil {
		return resp.StatusCode, fmt.Errorf("failed to delete task management workitem schema: %s", err)
	}

	return resp.StatusCode, nil
}
