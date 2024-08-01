package outbound_callanalysisresponseset

import (
	"context"
	"fmt"
	"log"

	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

/*
The genesyscloud_outbound_callanalysisresponseset_proxy.go file contains the proxy structures and methods that interact
with the Genesys Cloud SDK. We use composition here for each function on the proxy so individual functions can be stubbed
out during testing.
*/

// internalProxy holds a proxy instance that can be used throughout the package
var internalProxy *outboundCallanalysisresponsesetProxy

// Type definitions for each func on our proxy so we can easily mock them out later
type createOutboundCallanalysisresponsesetFunc func(ctx context.Context, p *outboundCallanalysisresponsesetProxy, responseSet *platformclientv2.Responseset) (*platformclientv2.Responseset, *platformclientv2.APIResponse, error)
type getAllOutboundCallanalysisresponsesetFunc func(ctx context.Context, p *outboundCallanalysisresponsesetProxy, name string) (*[]platformclientv2.Responseset, *platformclientv2.APIResponse, error)
type getOutboundCallanalysisresponsesetIdByNameFunc func(ctx context.Context, p *outboundCallanalysisresponsesetProxy, name string) (id string, retryable bool, response *platformclientv2.APIResponse, err error)
type getOutboundCallanalysisresponsesetByIdFunc func(ctx context.Context, p *outboundCallanalysisresponsesetProxy, id string) (responseSet *platformclientv2.Responseset, response *platformclientv2.APIResponse, err error)
type updateOutboundCallanalysisresponsesetFunc func(ctx context.Context, p *outboundCallanalysisresponsesetProxy, id string, responseSet *platformclientv2.Responseset) (*platformclientv2.Responseset, *platformclientv2.APIResponse, error)
type deleteOutboundCallanalysisresponsesetFunc func(ctx context.Context, p *outboundCallanalysisresponsesetProxy, id string) (response *platformclientv2.APIResponse, err error)

// outboundCallanalysisresponsesetProxy contains all of the methods that call genesys cloud APIs.
type outboundCallanalysisresponsesetProxy struct {
	clientConfig                                   *platformclientv2.Configuration
	outboundApi                                    *platformclientv2.OutboundApi
	createOutboundCallanalysisresponsesetAttr      createOutboundCallanalysisresponsesetFunc
	getAllOutboundCallanalysisresponsesetAttr      getAllOutboundCallanalysisresponsesetFunc
	getOutboundCallanalysisresponsesetIdByNameAttr getOutboundCallanalysisresponsesetIdByNameFunc
	getOutboundCallanalysisresponsesetByIdAttr     getOutboundCallanalysisresponsesetByIdFunc
	updateOutboundCallanalysisresponsesetAttr      updateOutboundCallanalysisresponsesetFunc
	deleteOutboundCallanalysisresponsesetAttr      deleteOutboundCallanalysisresponsesetFunc
}

// newOutboundCallanalysisresponsesetProxy initializes the outbound callanalysisresponseset proxy with all of the data needed to communicate with Genesys Cloud
func newOutboundCallanalysisresponsesetProxy(clientConfig *platformclientv2.Configuration) *outboundCallanalysisresponsesetProxy {
	api := platformclientv2.NewOutboundApiWithConfig(clientConfig)
	return &outboundCallanalysisresponsesetProxy{
		clientConfig: clientConfig,
		outboundApi:  api,
		createOutboundCallanalysisresponsesetAttr:      createOutboundCallanalysisresponsesetFn,
		getAllOutboundCallanalysisresponsesetAttr:      getAllOutboundCallanalysisresponsesetFn,
		getOutboundCallanalysisresponsesetIdByNameAttr: getOutboundCallanalysisresponsesetIdByNameFn,
		getOutboundCallanalysisresponsesetByIdAttr:     getOutboundCallanalysisresponsesetByIdFn,
		updateOutboundCallanalysisresponsesetAttr:      updateOutboundCallanalysisresponsesetFn,
		deleteOutboundCallanalysisresponsesetAttr:      deleteOutboundCallanalysisresponsesetFn,
	}
}

// getOutboundCallanalysisresponsesetProxy acts as a singleton to for the internalProxy.  It also ensures
// that we can still proxy our tests by directly setting internalProxy package variable
func getOutboundCallanalysisresponsesetProxy(clientConfig *platformclientv2.Configuration) *outboundCallanalysisresponsesetProxy {
	if internalProxy == nil {
		internalProxy = newOutboundCallanalysisresponsesetProxy(clientConfig)
	}
	return internalProxy
}

// createOutboundCallanalysisresponseset creates a Genesys Cloud outbound callanalysisresponseset
func (p *outboundCallanalysisresponsesetProxy) createOutboundCallanalysisresponseset(ctx context.Context, outboundCallanalysisresponseset *platformclientv2.Responseset) (*platformclientv2.Responseset, *platformclientv2.APIResponse, error) {
	return p.createOutboundCallanalysisresponsesetAttr(ctx, p, outboundCallanalysisresponseset)
}

// getOutboundCallanalysisresponseset retrieves all Genesys Cloud outbound callanalysisresponseset
func (p *outboundCallanalysisresponsesetProxy) getAllOutboundCallanalysisresponseset(ctx context.Context) (*[]platformclientv2.Responseset, *platformclientv2.APIResponse, error) {
	return p.getAllOutboundCallanalysisresponsesetAttr(ctx, p, "")
}

// getOutboundCallanalysisresponsesetIdByName returns a single Genesys Cloud outbound callanalysisresponseset by a name
func (p *outboundCallanalysisresponsesetProxy) getOutboundCallanalysisresponsesetIdByName(ctx context.Context, name string) (id string, retryable bool, response *platformclientv2.APIResponse, err error) {
	return p.getOutboundCallanalysisresponsesetIdByNameAttr(ctx, p, name)
}

// getOutboundCallanalysisresponsesetById returns a single Genesys Cloud outbound callanalysisresponseset by Id
func (p *outboundCallanalysisresponsesetProxy) getOutboundCallanalysisresponsesetById(ctx context.Context, id string) (outboundCallanalysisresponseset *platformclientv2.Responseset, response *platformclientv2.APIResponse, err error) {
	return p.getOutboundCallanalysisresponsesetByIdAttr(ctx, p, id)
}

// updateOutboundCallanalysisresponseset updates a Genesys Cloud outbound callanalysisresponseset
func (p *outboundCallanalysisresponsesetProxy) updateOutboundCallanalysisresponseset(ctx context.Context, id string, outboundCallanalysisresponseset *platformclientv2.Responseset) (*platformclientv2.Responseset, *platformclientv2.APIResponse, error) {
	return p.updateOutboundCallanalysisresponsesetAttr(ctx, p, id, outboundCallanalysisresponseset)
}

// deleteOutboundCallanalysisresponseset deletes a Genesys Cloud outbound callanalysisresponseset by Id
func (p *outboundCallanalysisresponsesetProxy) deleteOutboundCallanalysisresponseset(ctx context.Context, id string) (response *platformclientv2.APIResponse, err error) {
	return p.deleteOutboundCallanalysisresponsesetAttr(ctx, p, id)
}

// createOutboundCallanalysisresponsesetFn is an implementation function for creating a Genesys Cloud outbound callanalysisresponseset
func createOutboundCallanalysisresponsesetFn(ctx context.Context, p *outboundCallanalysisresponsesetProxy, outboundCallanalysisresponseset *platformclientv2.Responseset) (*platformclientv2.Responseset, *platformclientv2.APIResponse, error) {
	responseSet, resp, err := p.outboundApi.PostOutboundCallanalysisresponsesets(*outboundCallanalysisresponseset)
	if err != nil {
		return nil, resp, err
	}
	return responseSet, resp, nil
}

// getAllOutboundCallanalysisresponsesetFn is the implementation for retrieving all outbound callanalysisresponseset in Genesys Cloud
func getAllOutboundCallanalysisresponsesetFn(ctx context.Context, p *outboundCallanalysisresponsesetProxy, name string) (*[]platformclientv2.Responseset, *platformclientv2.APIResponse, error) {
	var allResponseSets []platformclientv2.Responseset
	const pageSize = 100

	responseSets, resp, err := p.outboundApi.GetOutboundCallanalysisresponsesets(pageSize, 1, true, "", name, "", "")
	if err != nil {
		return nil, resp, fmt.Errorf("Failed to get response set: %v", err)
	}
	if responseSets.Entities == nil || len(*responseSets.Entities) == 0 {
		return &allResponseSets, resp, nil
	}
	for _, responseSet := range *responseSets.Entities {
		allResponseSets = append(allResponseSets, responseSet)
	}

	for pageNum := 2; pageNum <= *responseSets.PageCount; pageNum++ {
		responseSets, resp, err := p.outboundApi.GetOutboundCallanalysisresponsesets(pageSize, pageNum, true, "", name, "", "")
		if err != nil {
			return nil, resp, fmt.Errorf("Failed to get response set: %v", err)
		}

		if responseSets.Entities == nil || len(*responseSets.Entities) == 0 {
			break
		}

		for _, responseSet := range *responseSets.Entities {
			allResponseSets = append(allResponseSets, responseSet)
		}
	}
	return &allResponseSets, resp, nil
}

// getOutboundCallanalysisresponsesetIdByNameFn is an implementation of the function to get a Genesys Cloud outbound callanalysisresponseset by name
func getOutboundCallanalysisresponsesetIdByNameFn(ctx context.Context, p *outboundCallanalysisresponsesetProxy, name string) (id string, retryable bool, response *platformclientv2.APIResponse, err error) {
	responseSets, resp, err := getAllOutboundCallanalysisresponsesetFn(ctx, p, name)
	if err != nil {
		return "", false, resp, err
	}

	if responseSets == nil || len(*responseSets) == 0 {
		return "", true, resp, fmt.Errorf("No outbound callanalysisresponseset found with name %s", name)
	}

	for _, responseSet := range *responseSets {
		if *responseSet.Name == name {
			log.Printf("Retrieved the outbound callanalysisresponseset id %s by name %s", *responseSet.Id, name)
			return *responseSet.Id, false, resp, nil
		}
	}
	return "", true, resp, fmt.Errorf("Unable to find outbound callanalysisresponseset with name %s", name)
}

// getOutboundCallanalysisresponsesetByIdFn is an implementation of the function to get a Genesys Cloud outbound callanalysisresponseset by Id
func getOutboundCallanalysisresponsesetByIdFn(ctx context.Context, p *outboundCallanalysisresponsesetProxy, id string) (outboundCallanalysisresponseset *platformclientv2.Responseset, response *platformclientv2.APIResponse, err error) {
	responseSet, resp, err := p.outboundApi.GetOutboundCallanalysisresponseset(id)
	if err != nil {
		return nil, resp, err
	}
	return responseSet, resp, nil
}

// updateOutboundCallanalysisresponsesetFn is an implementation of the function to update a Genesys Cloud outbound callanalysisresponseset
func updateOutboundCallanalysisresponsesetFn(ctx context.Context, p *outboundCallanalysisresponsesetProxy, id string, outboundCallanalysisresponseset *platformclientv2.Responseset) (*platformclientv2.Responseset, *platformclientv2.APIResponse, error) {
	responseSet, resp, err := getOutboundCallanalysisresponsesetByIdFn(ctx, p, id)
	if err != nil {
		return nil, resp, err
	}
	outboundCallanalysisresponseset.Version = responseSet.Version

	outboundCallanalysisresponseset, resp, err = p.outboundApi.PutOutboundCallanalysisresponseset(id, *outboundCallanalysisresponseset)
	if err != nil {
		return nil, resp, err
	}
	return outboundCallanalysisresponseset, resp, nil
}

// deleteOutboundCallanalysisresponsesetFn is an implementation function for deleting a Genesys Cloud outbound callanalysisresponseset
func deleteOutboundCallanalysisresponsesetFn(ctx context.Context, p *outboundCallanalysisresponsesetProxy, id string) (response *platformclientv2.APIResponse, err error) {
	return p.outboundApi.DeleteOutboundCallanalysisresponseset(id)
}
