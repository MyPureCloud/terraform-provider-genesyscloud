package responsemanagement_responseasset

import (
	"context"
	"github.com/mypurecloud/platform-client-sdk-go/v121/platformclientv2"
)

/*
The genesyscloud_responsemanagement_responseasset_proxy.go file contains the proxy structures and methods that interact
with the Genesys Cloud SDK. We use composition here for each function on the proxy so individual functions can be stubbed
out during testing.
*/

// internalProxy holds a proxy instance that can be used throughout the package
var internalProxy *responsemanagementResponseassetProxy

// Type definitions for each func on our proxy so we can easily mock them out later
type getResponsemanagementResponseassetByIdFunc func(ctx context.Context, p *responsemanagementResponseassetProxy, id string) (responseAssetSearchRequest *platformclientv2.Responseassetsearchrequest, responseCode int, err error)
type deleteResponsemanagementResponseassetFunc func(ctx context.Context, p *responsemanagementResponseassetProxy, id string) (responseCode int, err error)

// responsemanagementResponseassetProxy contains all of the methods that call genesys cloud APIs.
type responsemanagementResponseassetProxy struct {
	clientConfig                               *platformclientv2.Configuration
	responseManagementApi                      *platformclientv2.ResponseManagementApi
	getResponsemanagementResponseassetByIdAttr getResponsemanagementResponseassetByIdFunc
	deleteResponsemanagementResponseassetAttr  deleteResponsemanagementResponseassetFunc
}

// newResponsemanagementResponseassetProxy initializes the responsemanagement responseasset proxy with all of the data needed to communicate with Genesys Cloud
func newResponsemanagementResponseassetProxy(clientConfig *platformclientv2.Configuration) *responsemanagementResponseassetProxy {
	api := platformclientv2.NewResponseManagementApiWithConfig(clientConfig)
	return &responsemanagementResponseassetProxy{
		clientConfig:          clientConfig,
		responseManagementApi: api,
		getResponsemanagementResponseassetByIdAttr: getResponsemanagementResponseassetByIdFn,
		deleteResponsemanagementResponseassetAttr:  deleteResponsemanagementResponseassetFn,
	}
}

// getResponsemanagementResponseassetProxy acts as a singleton to for the internalProxy.  It also ensures
// that we can still proxy our tests by directly setting internalProxy package variable
func getResponsemanagementResponseassetProxy(clientConfig *platformclientv2.Configuration) *responsemanagementResponseassetProxy {
	if internalProxy == nil {
		internalProxy = newResponsemanagementResponseassetProxy(clientConfig)
	}

	return internalProxy
}

// getResponsemanagementResponseassetById returns a single Genesys Cloud responsemanagement responseasset by Id
func (p *responsemanagementResponseassetProxy) getResponsemanagementResponseassetById(ctx context.Context, id string) (responsemanagementResponseasset *platformclientv2.Responseassetsearchrequest, statusCode int, err error) {
	return p.getResponsemanagementResponseassetByIdAttr(ctx, p, id)
}

// deleteResponsemanagementResponseasset deletes a Genesys Cloud responsemanagement responseasset by Id
func (p *responsemanagementResponseassetProxy) deleteResponsemanagementResponseasset(ctx context.Context, id string) (statusCode int, err error) {
	return p.deleteResponsemanagementResponseassetAttr(ctx, p, id)
}

// getResponsemanagementResponseassetByIdFn is an implementation of the function to get a Genesys Cloud responsemanagement responseasset by Id
func getResponsemanagementResponseassetByIdFn(ctx context.Context, p *responsemanagementResponseassetProxy, id string) (responsemanagementResponseasset *platformclientv2.Responseassetsearchrequest, statusCode int, err error) {
	return nil, 0, nil
}

// deleteResponsemanagementResponseassetFn is an implementation function for deleting a Genesys Cloud responsemanagement responseasset
func deleteResponsemanagementResponseassetFn(ctx context.Context, p *responsemanagementResponseassetProxy, id string) (statusCode int, err error) {
	return 0, nil
}
