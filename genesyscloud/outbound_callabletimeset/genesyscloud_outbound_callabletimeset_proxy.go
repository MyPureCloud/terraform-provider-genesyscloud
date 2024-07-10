package outbound_callabletimeset

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

/*
The genesyscloud_outbound_callabletimeset_proxy.go file contains the proxy structures and methods that interact
with the Genesys Cloud SDK
*/

// internalProxy holds a proxy instance that can be used throughout the package
var internalProxy *outboundCallableTimesetProxy

// type definitions for each func on our proxy
type createOutboundCallabletimesetFunc func(ctx context.Context, p *outboundCallableTimesetProxy, timeset *platformclientv2.Callabletimeset) (*platformclientv2.Callabletimeset, *platformclientv2.APIResponse, error)
type getAllOutboundCallableTimesetFunc func(ctx context.Context, p *outboundCallableTimesetProxy) (*[]platformclientv2.Callabletimeset, *platformclientv2.APIResponse, error)
type getOutboundCallabletimesetByIdFunc func(ctx context.Context, p *outboundCallableTimesetProxy, timesetId string) (timeset *platformclientv2.Callabletimeset, response *platformclientv2.APIResponse, err error)
type getOutboundCallabletimesetByNameFunc func(ctx context.Context, p *outboundCallableTimesetProxy, name string) (timesetId string, retryable bool, response *platformclientv2.APIResponse, err error)
type updateOutboundCallabletimesetFunc func(ctx context.Context, p *outboundCallableTimesetProxy, timesetId string, timeset *platformclientv2.Callabletimeset) (*platformclientv2.Callabletimeset, *platformclientv2.APIResponse, error)
type deleteOutboundCallabletimesetFunc func(ctx context.Context, p *outboundCallableTimesetProxy, timesetId string) (response *platformclientv2.APIResponse, err error)

// outboundCallableTimesetProxy contains all of the methods that call genesys cloud APIs
type outboundCallableTimesetProxy struct {
	clientConfig                         *platformclientv2.Configuration
	outboundApi                          *platformclientv2.OutboundApi
	createOutboundCallabletimesetAttr    createOutboundCallabletimesetFunc
	getAllOutboundCallableTimesetAttr    getAllOutboundCallableTimesetFunc
	getOutboundCallabletimesetByIdAttr   getOutboundCallabletimesetByIdFunc
	getOutboundCallabletimesetByNameAttr getOutboundCallabletimesetByNameFunc
	updateOutboundCallabletimesetAttr    updateOutboundCallabletimesetFunc
	deleteOutboundCallabletimesetAttr    deleteOutboundCallabletimesetFunc
}

// newOutboundCallableTimesetProxy initializes the timeset proxy with the data needed for communication with the genesys cloud
func newOutboundCallableTimesetProxy(clientConfig *platformclientv2.Configuration) *outboundCallableTimesetProxy {
	api := platformclientv2.NewOutboundApiWithConfig(clientConfig)
	return &outboundCallableTimesetProxy{
		clientConfig:                         clientConfig,
		outboundApi:                          api,
		createOutboundCallabletimesetAttr:    createOutboundCallabletimesetFn,
		getAllOutboundCallableTimesetAttr:    getAllOutboundCallableTimesetFn,
		getOutboundCallabletimesetByIdAttr:   getOutboundCallabletimesetByIdFn,
		getOutboundCallabletimesetByNameAttr: getOutboundCallabletimesetByNameFn,
		updateOutboundCallabletimesetAttr:    updateOutboundCallabletimesetFn,
		deleteOutboundCallabletimesetAttr:    deleteOutboundCallabletimesetFn,
	}
}

func getOutboundCallabletimesetProxy(clientConfig *platformclientv2.Configuration) *outboundCallableTimesetProxy {
	if internalProxy == nil {
		internalProxy = newOutboundCallableTimesetProxy(clientConfig)
	}

	return internalProxy
}

// createOutboundCallabletimeset creates a Genesys Cloud Outbound Callable Timeset
func (p *outboundCallableTimesetProxy) createOutboundCallabletimeset(ctx context.Context, timeset *platformclientv2.Callabletimeset) (*platformclientv2.Callabletimeset, *platformclientv2.APIResponse, error) {
	return p.createOutboundCallabletimesetAttr(ctx, p, timeset)
}

// getAllOutboundCallableTimeset retrieves all Genesys Cloud Outbound Callable Timesets
func (p *outboundCallableTimesetProxy) getAllOutboundCallableTimeset(ctx context.Context) (*[]platformclientv2.Callabletimeset, *platformclientv2.APIResponse, error) {
	return p.getAllOutboundCallableTimesetAttr(ctx, p)
}

// getOutboundCallabletimesetById returns a single Genesys Cloud Outbound Callable Timeset by Id
func (p *outboundCallableTimesetProxy) getOutboundCallabletimesetById(ctx context.Context, timesetId string) (timeset *platformclientv2.Callabletimeset, response *platformclientv2.APIResponse, err error) {
	return p.getOutboundCallabletimesetByIdAttr(ctx, p, timesetId)
}

// getOutboundCallabletimesetByName returns a single Genesys Cloud Outbound Callable Timeset by a name
func (p *outboundCallableTimesetProxy) getOutboundCallabletimesetByName(ctx context.Context, name string) (timesetId string, retryable bool, response *platformclientv2.APIResponse, err error) {
	return p.getOutboundCallabletimesetByNameAttr(ctx, p, name)
}

// updateOutboundCallabletimeset updates a Genesys Cloud Outbound Callable Timeset
func (p *outboundCallableTimesetProxy) updateOutboundCallabletimeset(ctx context.Context, timesetId string, timeset *platformclientv2.Callabletimeset) (*platformclientv2.Callabletimeset, *platformclientv2.APIResponse, error) {
	return p.updateOutboundCallabletimesetAttr(ctx, p, timesetId, timeset)
}

// deleteOutboundCallabletimeset deletes a Genesys Cloud Outbound Callable timeset by Id
func (p *outboundCallableTimesetProxy) deleteOutboundCallabletimeset(ctx context.Context, timesetId string) (response *platformclientv2.APIResponse, err error) {
	return p.deleteOutboundCallabletimesetAttr(ctx, p, timesetId)
}

// createOutboundCallabletimesetFn is an implementation function for creating a Genesys Cloud Outbound Callable Timeset
func createOutboundCallabletimesetFn(ctx context.Context, p *outboundCallableTimesetProxy, timeset *platformclientv2.Callabletimeset) (*platformclientv2.Callabletimeset, *platformclientv2.APIResponse, error) {
	timeset, resp, err := p.outboundApi.PostOutboundCallabletimesets(*timeset)
	if err != nil {
		return nil, resp, fmt.Errorf("Failed to create timeset: %s", err)
	}
	return timeset, resp, nil
}

// getAllOutboundCallableTimesetFn is the implementation for retrieving all outbound callable timesets in Genesys Cloud
func getAllOutboundCallableTimesetFn(ctx context.Context, p *outboundCallableTimesetProxy) (*[]platformclientv2.Callabletimeset, *platformclientv2.APIResponse, error) {
	var allCallableTimesets []platformclientv2.Callabletimeset

	timesets, resp, err := p.outboundApi.GetOutboundCallabletimesets(100, 1, true, "", "", "", "")
	if err != nil {
		return nil, resp, fmt.Errorf("Failed to get outbound timesets: %v", err)
	}
	if timesets.Entities == nil || len(*timesets.Entities) == 0 {
		return &allCallableTimesets, resp, nil
	}

	for _, timeset := range *timesets.Entities {
		allCallableTimesets = append(allCallableTimesets, timeset)
	}

	var response *platformclientv2.APIResponse
	for pageNum := 2; pageNum <= *timesets.PageCount; pageNum++ {
		const pageSize = 100

		timesets, resp, err := p.outboundApi.GetOutboundCallabletimesets(pageSize, pageNum, true, "", "", "", "")
		if err != nil {
			return nil, resp, fmt.Errorf("Failed to get outbound timesets: %v", err)
		}
		response = resp
		if timesets.Entities == nil || len(*timesets.Entities) == 0 {
			break
		}

		for _, timeset := range *timesets.Entities {
			log.Printf("Dealing with timeset id : %s", *timeset.Id)
			allCallableTimesets = append(allCallableTimesets, timeset)
		}
	}
	return &allCallableTimesets, response, nil
}

// getOutboundCallabletimesetByIdFn is an implementation of the function to get a Genesys Cloud Outbound Callabletimeset by Id
func getOutboundCallabletimesetByIdFn(ctx context.Context, p *outboundCallableTimesetProxy, timesetId string) (timeset *platformclientv2.Callabletimeset, response *platformclientv2.APIResponse, err error) {
	timeset, resp, err := p.outboundApi.GetOutboundCallabletimeset(timesetId)
	if err != nil {
		//This is an API that throws an error on a 404 instead of just returning a 404.
		if strings.Contains(fmt.Sprintf("%s", err), "API Error: 404") {
			return nil, resp, nil

		}
		return nil, resp, fmt.Errorf("Failed to retrieve timeset by id %s: %s", timesetId, err)
	}
	return timeset, resp, nil
}

// getOutboundCallabletimesetIdByNameFn is an implementation of the function to get a Genesys Cloud Outbound Callabletimeset by name
func getOutboundCallabletimesetByNameFn(ctx context.Context, p *outboundCallableTimesetProxy, name string) (timesetId string, retryable bool, response *platformclientv2.APIResponse, err error) {
	timesets, resp, err := getAllOutboundCallableTimesetFn(ctx, p)
	if err != nil {
		return "", false, resp, fmt.Errorf("Error searching outbound timeset %s: %s", name, err)
	}

	var timeset platformclientv2.Callabletimeset
	for _, timesetSdk := range *timesets {
		if *timesetSdk.Name == name {
			log.Printf("Retrieved the timeset id %s by name %s", *timesetSdk.Id, name)
			timeset = timesetSdk
			return *timeset.Id, false, resp, nil
		}
	}
	return "", true, resp, fmt.Errorf("Unable to find timeset with name %s", name)
}

// updateOutboundCallabletimesetFn is an implementation of the function to update a Genesys Cloud Outbound Callabletimesets
func updateOutboundCallabletimesetFn(ctx context.Context, p *outboundCallableTimesetProxy, timesetId string, timeset *platformclientv2.Callabletimeset) (*platformclientv2.Callabletimeset, *platformclientv2.APIResponse, error) {
	outboundCallabletimeset, resp, err := getOutboundCallabletimesetByIdFn(ctx, p, timesetId)
	if err != nil {
		return nil, resp, fmt.Errorf("Failed to timeset by id %s: %s", timesetId, err)
	}

	timeset.Version = outboundCallabletimeset.Version
	timeset, resp, err = p.outboundApi.PutOutboundCallabletimeset(timesetId, *timeset)
	if err != nil {
		return nil, resp, fmt.Errorf("Failed to update timeset: %s", err)
	}
	return timeset, resp, nil
}

// deleteOutboundCallabletimesetFn is an implementation function for deleting a Genesys Cloud Outbound Callabletimesets
func deleteOutboundCallabletimesetFn(ctx context.Context, p *outboundCallableTimesetProxy, timesetId string) (response *platformclientv2.APIResponse, err error) {
	resp, err := p.outboundApi.DeleteOutboundCallabletimeset(timesetId)
	if err != nil {
		return resp, fmt.Errorf("Failed to delete timeset: %s", err)
	}
	return resp, nil
}
