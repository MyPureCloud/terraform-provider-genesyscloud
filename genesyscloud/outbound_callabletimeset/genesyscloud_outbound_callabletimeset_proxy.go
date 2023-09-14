package outbound_callabletimeset

import (
	"context"
	"fmt"
	"log"

	"github.com/mypurecloud/platform-client-sdk-go/v105/platformclientv2"
)

/*
The genesyscloud_outbound_callabletimeset_proxy.go file contains the proxy structures and methods that interact
with the Genesys Cloud SDK. We use composition here for each function on the proxy so individual functions can be stubbed
out during testing.
*/

// internalProxy holds a proxy instance that can be used throughout the package
var internalProxy *outboundCallabletimesetProxy

// Type definitions for each func on our proxy so we can easily mock them out later
type createOutboundCallabletimesetFunc func(ctx context.Context, p *outboundCallabletimesetProxy, callableTimeSet *platformclientv2.Callabletimeset) (*platformclientv2.Callabletimeset, error)
type getAllOutboundCallabletimesetFunc func(ctx context.Context, p *outboundCallabletimesetProxy) (*[]platformclientv2.Callabletimeset, error)
type getOutboundCallabletimesetByIdFunc func(ctx context.Context, p *outboundCallabletimesetProxy, id string) (callableTimeSet *platformclientv2.Callabletimeset, responseCode int, err error)
type getOutboundCallabletimesetIdByNameFunc func(ctx context.Context, p *outboundCallabletimesetProxy, name string) (id string, retryable bool, err error)
type updateOutboundCallabletimesetFunc func(ctx context.Context, p *outboundCallabletimesetProxy, id string, callableTimeSet *platformclientv2.Callabletimeset) (*platformclientv2.Callabletimeset, error)
type deleteOutboundCallabletimesetFunc func(ctx context.Context, p *outboundCallabletimesetProxy, id string) (responseCode int, err error)

// outboundCallabletimesetProxy contains all of the methods that call genesys cloud APIs.
type outboundCallabletimesetProxy struct {
	clientConfig                           *platformclientv2.Configuration
	outboundApi                            *platformclientv2.OutboundApi
	createOutboundCallabletimesetAttr      createOutboundCallabletimesetFunc
	getAllOutboundCallabletimesetAttr      getAllOutboundCallabletimesetFunc
	getOutboundCallabletimesetByIdAttr     getOutboundCallabletimesetByIdFunc
	getOutboundCallabletimesetIdByNameAttr getOutboundCallabletimesetIdByNameFunc
	updateOutboundCallabletimesetAttr      updateOutboundCallabletimesetFunc
	deleteOutboundCallabletimesetAttr      deleteOutboundCallabletimesetFunc
}

// newOutboundCallabletimesetProxy initializes the Outbound Callabletimeset proxy with all of the data needed to communicate with Genesys Cloud
func newOutboundCallabletimesetProxy(clientConfig *platformclientv2.Configuration) *outboundCallabletimesetProxy {
	api := platformclientv2.NewOutboundApiWithConfig(clientConfig)
	return &outboundCallabletimesetProxy{
		clientConfig:                           clientConfig,
		outboundApi:                            api,
		createOutboundCallabletimesetAttr:      createOutboundCallabletimesetFn,
		getAllOutboundCallabletimesetAttr:      getAllOutboundCallabletimesetFn,
		getOutboundCallabletimesetByIdAttr:     getOutboundCallabletimesetByIdFn,
		getOutboundCallabletimesetIdByNameAttr: getOutboundCallabletimesetIdByNameFn,
		updateOutboundCallabletimesetAttr:      updateOutboundCallabletimesetFn,
		deleteOutboundCallabletimesetAttr:      deleteOutboundCallabletimesetFn,
	}
}

// getOutboundCallabletimesetProxy acts as a singleton to for the internalProxy.  It also ensures
// that we can still proxy our tests by directly setting internalProxy package variable
func getOutboundCallabletimesetProxy(clientConfig *platformclientv2.Configuration) *outboundCallabletimesetProxy {
	if internalProxy == nil {
		internalProxy = newOutboundCallabletimesetProxy(clientConfig)
	}

	return internalProxy
}

// createOutboundCallabletimeset creates a Genesys Cloud Outbound Callabletimeset
func (p *outboundCallabletimesetProxy) createOutboundCallabletimeset(ctx context.Context, outboundCallabletimeset *platformclientv2.Callabletimeset) (*platformclientv2.Callabletimeset, error) {
	return p.createOutboundCallabletimesetAttr(ctx, p, outboundCallabletimeset)
}

// getOutboundCallabletimeset retrieves all Genesys Cloud Outbound Callabletimeset
func (p *outboundCallabletimesetProxy) getAllOutboundCallabletimeset(ctx context.Context) (*[]platformclientv2.Callabletimeset, error) {
	return p.getAllOutboundCallabletimesetAttr(ctx, p)
}

// getOutboundCallabletimesetById returns a single Genesys Cloud Outbound Callabletimeset by Id
func (p *outboundCallabletimesetProxy) getOutboundCallabletimesetById(ctx context.Context, id string) (outboundCallabletimeset *platformclientv2.Callabletimeset, statusCode int, err error) {
	return p.getOutboundCallabletimesetByIdAttr(ctx, p, id)
}

// getOutboundCallabletimesetIdByName returns a single Genesys Cloud Outbound Callabletimeset by a name
func (p *outboundCallabletimesetProxy) getOutboundCallabletimesetIdByName(ctx context.Context, name string) (id string, retryable bool, err error) {
	return p.getOutboundCallabletimesetIdByNameAttr(ctx, p, name)
}

// updateOutboundCallabletimeset updates a Genesys Cloud Outbound Callabletimeset
func (p *outboundCallabletimesetProxy) updateOutboundCallabletimeset(ctx context.Context, id string, outboundCallabletimeset *platformclientv2.Callabletimeset) (*platformclientv2.Callabletimeset, error) {
	return p.updateOutboundCallabletimesetAttr(ctx, p, id, outboundCallabletimeset)
}

// deleteOutboundCallabletimeset deletes a Genesys Cloud Outbound Callabletimeset by Id
func (p *outboundCallabletimesetProxy) deleteOutboundCallabletimeset(ctx context.Context, id string) (statusCode int, err error) {
	return p.deleteOutboundCallabletimesetAttr(ctx, p, id)
}

// createOutboundCallabletimesetFn is an implementation function for creating a Genesys Cloud Outbound Callabletimeset
func createOutboundCallabletimesetFn(ctx context.Context, p *outboundCallabletimesetProxy, outboundCallabletimeset *platformclientv2.Callabletimeset) (*platformclientv2.Callabletimeset, error) {
	callableTimeSet, _, err := p.outboundApi.PostOutboundCallabletimesets(*outboundCallabletimeset)
	if err != nil {
		return nil, fmt.Errorf("Failed to create Outbound Callabletimeset: %s", err)
	}

	return callableTimeSet, nil
}

// getAllOutboundCallabletimesetFn is the implementation for retrieving all Outbound Callabletimeset in Genesys Cloud
func getAllOutboundCallabletimesetFn(ctx context.Context, p *outboundCallabletimesetProxy) (*[]platformclientv2.Callabletimeset, error) {
	var alls []platformclientv2.Callabletimeset

	for pageNum := 1; ; pageNum++ {
		const pageSize = 100

		callableTimeSets, _, err := p.outboundApi.GetOutboundCallabletimesets(pageSize, pageNum, true, "", "", "", "")
		if err != nil {
			return nil, fmt.Errorf("Failed to get Outbound Callabletimeset: %v", err)
		}

		if callableTimeSets.Entities == nil || len(*callableTimeSets.Entities) == 0 {
			break
		}

		for _, callableTimeSet := range *callableTimeSets.Entities {
			log.Printf("Dealing with callableTimeSet id : %s", *callableTimeSet.Id)
			alls = append(alls, callableTimeSet)
		}
	}

	return &alls, nil
}

// getOutboundCallabletimesetByIdFn is an implementation of the function to get a Genesys Cloud Outbound Callabletimeset by Id
func getOutboundCallabletimesetByIdFn(ctx context.Context, p *outboundCallabletimesetProxy, id string) (outboundCallabletimeset *platformclientv2.Callabletimeset, statusCode int, err error) {
	callableTimeSet, resp, err := p.outboundApi.GetOutboundCallabletimeset(id)
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("Failed to retrieve callabletimeset by id %s: %s", id, err)
	}

	return callableTimeSet, resp.StatusCode, nil
}

// getOutboundCallabletimesetIdByNameFn is an implementation of the function to get a Genesys Cloud Outbound Callabletimeset by name
func getOutboundCallabletimesetIdByNameFn(ctx context.Context, p *outboundCallabletimesetProxy, name string) (id string, retryable bool, err error) {
	const pageNum = 1
	const pageSize = 100
	callableTimeSets, _, err := p.outboundApi.GetOutboundCallabletimesets(pageSize, pageNum, true, "", name, "", "")
	if err != nil {
		return "", false, fmt.Errorf("Error searching Outbound Callabletimeset %s: %s", name, err)
	}

	if callableTimeSets.Entities == nil || len(*callableTimeSets.Entities) == 0 {
		return "", true, fmt.Errorf("No Outbound Callabletimeset found with name %s", name)
	}

	if len(*callableTimeSets.Entities) > 1 {
		return "", false, fmt.Errorf("Too many values returned in look for Outbound Callabletimesets.  Unable to choose 1 Outbound Callabletimeset.  Please refine search and continue.")
	}

	log.Printf("Retrieved the callableTimeSet id %s by name %s", *(*callableTimeSets.Entities)[0].Id, name)
	callableTimeSet := (*callableTimeSets.Entities)[0]
	return *callableTimeSet.Id, false, nil
}

// updateOutboundCallabletimesetFn is an implementation of the function to update a Genesys Cloud Outbound Callabletimeset
func updateOutboundCallabletimesetFn(ctx context.Context, p *outboundCallabletimesetProxy, id string, outboundCallabletimeset *platformclientv2.Callabletimeset) (*platformclientv2.Callabletimeset, error) {
	callableTimeSet, _, err := getOutboundCallabletimesetByIdFn(ctx, p, id)
	if err != nil {
		return nil, fmt.Errorf("Failed to get Outbound Callabletimeset by id %s: %s", id, err)
	}

	outboundCallabletimeset.Version = callableTimeSet.Version
	outboundCallabletimeset, _, err = p.outboundApi.PutOutboundCallabletimeset(id, *outboundCallabletimeset)
	if err != nil {
		return nil, fmt.Errorf("Failed to update Outbound Callabletimeset: %s", err)
	}
	return outboundCallabletimeset, nil
}

// deleteOutboundCallabletimesetFn is an implementation function for deleting a Genesys Cloud Outbound Callabletimeset
func deleteOutboundCallabletimesetFn(ctx context.Context, p *outboundCallabletimesetProxy, id string) (statusCode int, err error) {
	resp, err := p.outboundApi.DeleteOutboundCallabletimeset(id)
	if err != nil {
		return resp.StatusCode, fmt.Errorf("Failed to delete Outbound Callabletimeset: %s", err)
	}

	return resp.StatusCode, nil
}
