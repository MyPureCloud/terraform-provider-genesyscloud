package flow_loglevel

import (
	"context"
	"fmt"

	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

/*
The genesyscloud_flow_loglevel_proxy.go file contains the proxy structures and methods that interact
with the Genesys Cloud SDK. We use composition here for each function on the proxy so individual functions can be stubbed
out during testing.

Each proxy implementation:

1.  Should provide a private package level variable that holds a instance of a proxy class.
2.  A New... constructor function  to initialize the proxy object.  This constructor should only be used within
    the proxy.
3.  A get private constructor function that the classes in the package can be used to retrieve
    the proxy.  This proxy should check to see if the package level proxy instance is nil and
    should initialize it, otherwise it should return the instance
4.  Type definitions for each function that will be used in the proxy.  We use composition here
    so that we can easily provide mocks for testing.
5.  A struct for the proxy that holds an attribute for each function type.
6.  Wrapper methods on each of the elements on the struct.
7.  Function implementations for each function type definition.

*/

// internalProxy holds a proxy instance that can be used throughout the package
var internalProxy *flowLogLevelProxy

// Type definitions for each func on our proxy so we can easily mock them out later
type createFlowLogLevelFunc func(ctx context.Context, p *flowLogLevelProxy, flowId string, flowLogLevelRequest *platformclientv2.Flowloglevelrequest) (*platformclientv2.Flowsettingsresponse, *platformclientv2.APIResponse, error)
type getAllFlowLogLevelsFunc func(ctx context.Context, p *flowLogLevelProxy) (*[]platformclientv2.Flowsettingsresponse, *platformclientv2.APIResponse, error)
type getFlowLogLevelByIdFunc func(ctx context.Context, p *flowLogLevelProxy, flowId string) (flowLogLevel *platformclientv2.Flowsettingsresponse, apiResponse *platformclientv2.APIResponse, err error)
type updateFlowLogLevelFunc func(ctx context.Context, p *flowLogLevelProxy, flowId string, flowLogLevel *platformclientv2.Flowloglevelrequest) (*platformclientv2.Flowsettingsresponse, *platformclientv2.APIResponse, error)
type deleteFlowLogLevelFunc func(ctx context.Context, p *flowLogLevelProxy, flowId string) (*platformclientv2.APIResponse, error)

// flowLogLevelProxy contains all the methods that call genesys cloud APIs.
type flowLogLevelProxy struct {
	clientConfig               *platformclientv2.Configuration
	architectApi               *platformclientv2.ArchitectApi
	createFlowLogLevelAttr     createFlowLogLevelFunc
	getFlowLogLevelByIdAttr    getFlowLogLevelByIdFunc
	getAllFlowLogLevelsAttr    getAllFlowLogLevelsFunc
	updateFlowLogLevelAttr     updateFlowLogLevelFunc
	deleteFlowLogLevelByIdAttr deleteFlowLogLevelFunc
}

// newFlowLogLevelProxy initializes the Flow Log Level proxy with all the data needed to communicate with Genesys Cloud
func newFlowLogLevelProxy(clientConfig *platformclientv2.Configuration) *flowLogLevelProxy {
	api := platformclientv2.NewArchitectApiWithConfig(clientConfig)
	return &flowLogLevelProxy{
		clientConfig:               clientConfig,
		architectApi:               api,
		createFlowLogLevelAttr:     createFlowLogLevelFn,
		getAllFlowLogLevelsAttr:    getAllFlowLogLevelsFn,
		getFlowLogLevelByIdAttr:    getFlowLogLevelByIdFn,
		updateFlowLogLevelAttr:     updateFlowLogLevelFn,
		deleteFlowLogLevelByIdAttr: deleteFlowLogLevelsFn,
	}
}

// getFlowLogLevelProxy acts as a singleton to for the internalProxy.  It also ensures
// that we can still proxy our tests by directly setting internalProxy package variable
func getFlowLogLevelProxy(clientConfig *platformclientv2.Configuration) *flowLogLevelProxy {
	if internalProxy == nil {
		internalProxy = newFlowLogLevelProxy(clientConfig)
	}

	return internalProxy
}

// getAllFlowLogLevels retrieves all Genesys Cloud Flow Log Levels
func (p *flowLogLevelProxy) getAllFlowLogLevels(ctx context.Context) (*[]platformclientv2.Flowsettingsresponse, *platformclientv2.APIResponse, error) {
	return p.getAllFlowLogLevelsAttr(ctx, p)
}

// createFlowLogLevel creates a Genesys Cloud Flow Log Level
func (p *flowLogLevelProxy) createFlowLogLevel(ctx context.Context, flowLogLevelId string, flowLogLevelRequest *platformclientv2.Flowloglevelrequest) (*platformclientv2.Flowsettingsresponse, *platformclientv2.APIResponse, error) {
	return p.createFlowLogLevelAttr(ctx, p, flowLogLevelId, flowLogLevelRequest)
}

// getFlowLogLevelById returns a single Genesys Cloud Flow Log Level by Id
func (p *flowLogLevelProxy) getFlowLogLevelById(ctx context.Context, flowId string) (*platformclientv2.Flowsettingsresponse, *platformclientv2.APIResponse, error) {
	return p.getFlowLogLevelByIdAttr(ctx, p, flowId)
}

// updateFlowLogLevel updates a Genesys Cloud Flow Log Level
func (p *flowLogLevelProxy) updateFlowLogLevel(ctx context.Context, flowId string, flowLogLevelRequest *platformclientv2.Flowloglevelrequest) (*platformclientv2.Flowsettingsresponse, *platformclientv2.APIResponse, error) {
	return p.updateFlowLogLevelAttr(ctx, p, flowId, flowLogLevelRequest)
}

// DeleteFlowLogLevel deletes a Genesys Cloud Flow Log Level by Id
func (p *flowLogLevelProxy) deleteFlowLogLevelById(ctx context.Context, flowId string) (*platformclientv2.APIResponse, error) {
	return p.deleteFlowLogLevelByIdAttr(ctx, p, flowId)
}

// createFlowLogLevelFn is an implementation function for creating a Genesys Cloud Flow Log Level
func createFlowLogLevelFn(ctx context.Context, p *flowLogLevelProxy, flowId string, flowLogLevelRequest *platformclientv2.Flowloglevelrequest) (*platformclientv2.Flowsettingsresponse, *platformclientv2.APIResponse, error) {
	flowLogLevel, apiResponse, err := p.architectApi.PostFlowInstancesSettingsLoglevels(flowId, *flowLogLevelRequest, nil)

	if err != nil {
		return nil, apiResponse, fmt.Errorf("Failed to create flow log level: %s", err)
	}

	return flowLogLevel, apiResponse, nil
}

// getFlowLogLevelByIdFn is an implementation of the function to get a Genesys Cloud Flow Log Level by Id
func getFlowLogLevelByIdFn(ctx context.Context, p *flowLogLevelProxy, flowId string) (*platformclientv2.Flowsettingsresponse, *platformclientv2.APIResponse, error) {
	flowLogLevel, apiResponse, err := p.architectApi.GetFlowInstancesSettingsLoglevels(flowId, nil)
	if err != nil {
		return nil, apiResponse, fmt.Errorf("Failed to retrieve flow log level by id %s: %s", flowId, err)
	}

	return flowLogLevel, apiResponse, nil
}

// getAllFlowLogLevelsFn is the implementation for retrieving all flow log levels in Genesys Cloud
func getAllFlowLogLevelsFn(ctx context.Context, p *flowLogLevelProxy) (*[]platformclientv2.Flowsettingsresponse, *platformclientv2.APIResponse, error) {
	const pageSize = 100
	var totalFlowLogLevels []platformclientv2.Flowsettingsresponse

	flowSettingsResponse, apiResponse, err := p.architectApi.GetFlowsInstancesSettingsLoglevels(nil, 1, pageSize)
	if err != nil {
		return nil, apiResponse, fmt.Errorf("Failed to get page of flows: %v", err)
	}

	if flowSettingsResponse.Entities == nil || len(*flowSettingsResponse.Entities) == 0 {
		return &totalFlowLogLevels, apiResponse, nil
	}

	totalFlowLogLevels = append(totalFlowLogLevels, *flowSettingsResponse.Entities...)

	for pageNum := 2; pageNum <= *flowSettingsResponse.PageCount; pageNum++ {
		flowSettingsResponse, apiResponse, err := p.architectApi.GetFlowsInstancesSettingsLoglevels(nil, pageNum, pageSize)
		if err != nil {
			return nil, apiResponse, fmt.Errorf("Failed to get page %d of flow log levels: %v", pageNum, err)
		}
		if flowSettingsResponse.Entities == nil || len(*flowSettingsResponse.Entities) == 0 {
			return &totalFlowLogLevels, apiResponse, nil
		}

		totalFlowLogLevels = append(totalFlowLogLevels, *flowSettingsResponse.Entities...)
	}
	return &totalFlowLogLevels, apiResponse, nil
}

// updateFlowLogLevelFn is an implementation of the function to update a Genesys Cloud flow log level
func updateFlowLogLevelFn(ctx context.Context, p *flowLogLevelProxy, flowLogLevelId string, flowLogLevelRequest *platformclientv2.Flowloglevelrequest) (*platformclientv2.Flowsettingsresponse, *platformclientv2.APIResponse, error) {
	flowSettingsResponse, apiResponse, err := p.architectApi.PutFlowInstancesSettingsLoglevels(flowLogLevelId, *flowLogLevelRequest, nil)
	if err != nil {
		return nil, apiResponse, fmt.Errorf("Failed to update flow log level: %s", err)
	}
	return flowSettingsResponse, apiResponse, nil
}

// deleteFlowLogLevelsFn is an implementation function for deleting a Genesys Cloud Flow Log Level
func deleteFlowLogLevelsFn(ctx context.Context, p *flowLogLevelProxy, flowLogLevelId string) (*platformclientv2.APIResponse, error) {
	apiResponse, err := p.architectApi.DeleteFlowInstancesSettingsLoglevels(flowLogLevelId)
	if err != nil {
		return apiResponse, fmt.Errorf("Failed to delete flow log level: %s", err)
	}

	return apiResponse, nil
}
