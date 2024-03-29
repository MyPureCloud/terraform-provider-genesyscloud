package flow_logLevel

import (
	"context"
	"fmt"
	"github.com/mypurecloud/platform-client-sdk-go/v125/platformclientv2"
	"log"
)

/*
The genesyscloud_flow_logLevel_proxy.go file contains the proxy structures and methods that interact
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
type createFlowLogLevelFunc func(ctx context.Context, p *flowLogLevelProxy, flowId string, flowLogLevelRequest *platformclientv2.Flowloglevelrequest) (*platformclientv2.Flowsettingsresponse, error)
type getAllFlowLogLevelsFunc func(ctx context.Context, p *flowLogLevelProxy) (*[]platformclientv2.Flowsettingsresponse, error)
type getFlowLogLevelByIdFunc func(ctx context.Context, p *flowLogLevelProxy, flowId string) (flowLogLevel *platformclientv2.Flowsettingsresponse, responseCode int, err error)
type updateFlowLogLevelFunc func(ctx context.Context, p *flowLogLevelProxy, flowId string, flowLogLevel *platformclientv2.Flowloglevelrequest) (*platformclientv2.Flowsettingsresponse, error)
type deleteFlowLogLevelFunc func(ctx context.Context, p *flowLogLevelProxy, flowId string) (responseCode int, err error)

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

// newFlowLogLevelsContactsProxy initializes the External Contacts proxy with all the data needed to communicate with Genesys Cloud
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

// getAllFlowLogLevels retrieves all Genesys Cloud External Contacts
func (p *flowLogLevelProxy) getAllFlowLogLevels(ctx context.Context) (*[]platformclientv2.Flowsettingsresponse, error) {
	return p.getAllFlowLogLevelsAttr(ctx, p)
}

// createFlowLogLevel creates a Genesys Cloud External Contact
func (p *flowLogLevelProxy) createFlowLogLevel(ctx context.Context, flowLogLevelId string, flowLogLevelRequest *platformclientv2.Flowloglevelrequest) (*platformclientv2.Flowsettingsresponse, error) {
	return p.createFlowLogLevelAttr(ctx, p, flowLogLevelId, flowLogLevelRequest)
}

// getFlowLogLevelById returns a single Genesys Cloud External Contact by Id
func (p *flowLogLevelProxy) getFlowLogLevelById(ctx context.Context, flowId string) (*platformclientv2.Flowsettingsresponse, int, error) {
	return p.getFlowLogLevelByIdAttr(ctx, p, flowId)
}

// updateFlowLogLevel updates a Genesys Cloud External Contact
func (p *flowLogLevelProxy) updateFlowLogLevel(ctx context.Context, flowId string, flowLogLevelRequest *platformclientv2.Flowloglevelrequest) (*platformclientv2.Flowsettingsresponse, error) {
	return p.updateFlowLogLevelAttr(ctx, p, flowId, flowLogLevelRequest)
}

// DeleteFlowLogLevel deletes a Genesys Cloud External Contact by Id
func (p *flowLogLevelProxy) deleteFlowLogLevelById(ctx context.Context, flowId string) (int, error) {
	return p.deleteFlowLogLevelByIdAttr(ctx, p, flowId)
}

// createFlowLogLevelFn is an implementation function for creating a Genesys Cloud External Contact
func createFlowLogLevelFn(ctx context.Context, p *flowLogLevelProxy, flowId string, flowLogLevelRequest *platformclientv2.Flowloglevelrequest) (*platformclientv2.Flowsettingsresponse, error) {
	flowLogLevel, resp, err := p.architectApi.PostFlowInstancesSettingsLoglevels(flowId, *flowLogLevelRequest, nil)
	log.Printf("createFlowLogLevelFn flowLogLevelRequest  %v", flowLogLevelRequest)
	log.Printf("createFlowLogLevelFn flowId %s", flowId)
	log.Printf("createFlowLogLevelFn flowLogLevel %v", flowLogLevel)
	log.Printf("createFlowLogLevelFn resp %v", resp)
	log.Printf("createFlowLogLevelFn err %v", err)
	if err != nil {
		return nil, fmt.Errorf("Failed to create flow log level: %s", err)
	}

	return flowLogLevel, nil
}

// getFlowLogLevelByIdFn is an implementation of the function to get a Genesys Cloud External Contact by Id
func getFlowLogLevelByIdFn(ctx context.Context, p *flowLogLevelProxy, flowId string) (*platformclientv2.Flowsettingsresponse, int, error) {
	expandArray := []string{"logLevelCharacteristics.characteristics"}
	flowLogLevel, resp, err := p.architectApi.GetFlowInstancesSettingsLoglevels(flowId, expandArray)
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("Failed to retrieve flow log level by id %s: %s", flowId, err)
	}

	return flowLogLevel, 0, nil
}

// getAllFlowLogLevelsFn is the implementation for retrieving all flow log levels in Genesys Cloud
func getAllFlowLogLevelsFn(ctx context.Context, p *flowLogLevelProxy) (*[]platformclientv2.Flowsettingsresponse, error) {
	const pageSize = 100
	var totalFlowLogLevels []platformclientv2.Flowsettingsresponse

	flowSettingsResponse, _, err := p.architectApi.GetFlowsInstancesSettingsLoglevels(nil, 1, pageSize)
	if err != nil {
		return nil, fmt.Errorf("Failed to get page of flows: %v", err)
	}

	for _, flowLogLevel := range *flowSettingsResponse.Entities {
		totalFlowLogLevels = append(totalFlowLogLevels, flowLogLevel)
	}

	for pageNum := 2; pageNum <= *flowSettingsResponse.PageCount; pageNum++ {
		flowSettingsResponse, _, err := p.architectApi.GetFlowsInstancesSettingsLoglevels(nil, pageNum, pageSize)
		if err != nil {
			return nil, fmt.Errorf("Failed to get page %d of flow log levels: %v", pageNum, err)
		}
		for _, flowLogLevel := range *flowSettingsResponse.Entities {
			totalFlowLogLevels = append(totalFlowLogLevels, flowLogLevel)
		}
	}
	return &totalFlowLogLevels, nil
}

// updateFlowLogLevelFn is an implementation of the function to update a Genesys Cloud flow log level
func updateFlowLogLevelFn(ctx context.Context, p *flowLogLevelProxy, flowLogLevelId string, flowLogLevelRequest *platformclientv2.Flowloglevelrequest) (*platformclientv2.Flowsettingsresponse, error) {
	flowSettingsResponse, _, err := p.architectApi.PutFlowInstancesSettingsLoglevels(flowLogLevelId, *flowLogLevelRequest, nil)
	if err != nil {
		return nil, fmt.Errorf("Failed to update flow log level: %s", err)
	}
	return flowSettingsResponse, nil
}

// deleteFlowLogLevelsFn is an implementation function for deleting a Genesys Cloud External Contact
func deleteFlowLogLevelsFn(ctx context.Context, p *flowLogLevelProxy, flowLogLevelId string) (int, error) {
	resp, err := p.architectApi.DeleteFlowInstancesSettingsLoglevels(flowLogLevelId)
	if err != nil {
		return resp.StatusCode, fmt.Errorf("Failed to delete flow log level: %s", err)
	}

	return resp.StatusCode, nil
}
