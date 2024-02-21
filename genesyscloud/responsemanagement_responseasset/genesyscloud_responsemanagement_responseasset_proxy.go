package responsemanagement_responseasset

import (
	"context"
	"fmt"
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
type createRespManagementRespAssetFunc func(ctx context.Context, p *responsemanagementResponseassetProxy, respAsset *platformclientv2.Createresponseassetrequest) (*platformclientv2.Createresponseassetresponse, int, error)
type updateRespManagementRespAssetFunc func(ctx context.Context, p *responsemanagementResponseassetProxy, id string, respAsset *platformclientv2.Responseassetrequest) (*platformclientv2.Responseasset, int, error)
type getRespManagementRespAssetByIdFunc func(ctx context.Context, p *responsemanagementResponseassetProxy, id string) (*platformclientv2.Responseasset, *platformclientv2.APIResponse, error)
type deleteRespManagementRespAssetFunc func(ctx context.Context, p *responsemanagementResponseassetProxy, id string) (response *platformclientv2.APIResponse, err error)

// responsemanagementResponseassetProxy contains all of the methods that call genesys cloud APIs.
type responsemanagementResponseassetProxy struct {
	clientConfig                       *platformclientv2.Configuration
	responseManagementApi              *platformclientv2.ResponseManagementApi
	createRespManagementRespAssetAttr  createRespManagementRespAssetFunc
	updateRespManagementRespAssetAttr  updateRespManagementRespAssetFunc
	getRespManagementRespAssetByIdAttr getRespManagementRespAssetByIdFunc
	deleteRespManagementRespAssetAttr  deleteRespManagementRespAssetFunc
}

// newRespManagementRespAssetProxy initializes the responsemanagement responseasset proxy with all of the data needed to communicate with Genesys Cloud
func newRespManagementRespAssetProxy(clientConfig *platformclientv2.Configuration) *responsemanagementResponseassetProxy {
	api := platformclientv2.NewResponseManagementApiWithConfig(clientConfig)
	return &responsemanagementResponseassetProxy{
		clientConfig:                       clientConfig,
		responseManagementApi:              api,
		createRespManagementRespAssetAttr:  createRespManagementRespAssetFn,
		updateRespManagementRespAssetAttr:  updateRespManagementRespAssetFn,
		getRespManagementRespAssetByIdAttr: getRespManagementRespAssetByIdFn,
		deleteRespManagementRespAssetAttr:  deleteRespManagementRespAssetFn,
	}
}

// getRespManagementRespAssetProxy acts as a singleton to for the internalProxy.  It also ensures
// that we can still proxy our tests by directly setting internalProxy package variable
func getRespManagementRespAssetProxy(clientConfig *platformclientv2.Configuration) *responsemanagementResponseassetProxy {
	if internalProxy == nil {
		internalProxy = newRespManagementRespAssetProxy(clientConfig)
	}
	return internalProxy
}

// createRespManagementRespAsset creates a Genesys Cloud responsemanagement responseasset by Id
func (p *responsemanagementResponseassetProxy) createRespManagementRespAsset(ctx context.Context, respAsset *platformclientv2.Createresponseassetrequest) (*platformclientv2.Createresponseassetresponse, int, error) {
	return p.createRespManagementRespAssetAttr(ctx, p, respAsset)
}

// updateRespManagementRespAsset updates a Genesys Cloud responsemanagement responseasset by Id
func (p *responsemanagementResponseassetProxy) updateRespManagementRespAsset(ctx context.Context, id string, respAsset *platformclientv2.Responseassetrequest) (*platformclientv2.Responseasset, int, error) {
	return p.updateRespManagementRespAssetAttr(ctx, p, id, respAsset)
}

// getRespManagementRespAssetById returns a single Genesys Cloud responsemanagement responseasset by Id
func (p *responsemanagementResponseassetProxy) getRespManagementRespAssetById(ctx context.Context, id string) (*platformclientv2.Responseasset, *platformclientv2.APIResponse, error) {
	return p.getRespManagementRespAssetByIdAttr(ctx, p, id)
}

// deleteRespManagementRespAsset deletes a Genesys Cloud responsemanagement responseasset by Id
func (p *responsemanagementResponseassetProxy) deleteRespManagementRespAsset(ctx context.Context, id string) (response *platformclientv2.APIResponse, err error) {
	return p.deleteRespManagementRespAssetAttr(ctx, p, id)
}

// createRespManagementRespAssetFn is an implementation of the function to create a Genesys Cloud responsemanagement responseasset
func createRespManagementRespAssetFn(ctx context.Context, p *responsemanagementResponseassetProxy, respAsset *platformclientv2.Createresponseassetrequest) (*platformclientv2.Createresponseassetresponse, int, error) {
	postResponseData, resp, err := p.responseManagementApi.PostResponsemanagementResponseassetsUploads(*respAsset)
	if err != nil {
		return nil, 0, fmt.Errorf("Failed to upload response asset: %v", err)
	}
	return postResponseData, resp.StatusCode, nil
}

// updateRespManagementRespAssetFn is an implementation of the function to update a Genesys Cloud responsemanagement responseasset
func updateRespManagementRespAssetFn(ctx context.Context, p *responsemanagementResponseassetProxy, id string, respAsset *platformclientv2.Responseassetrequest) (*platformclientv2.Responseasset, int, error) {
	putResponseData, resp, err := p.responseManagementApi.PutResponsemanagementResponseasset(id, *respAsset)
	if err != nil {
		return nil, 0, fmt.Errorf("Failed to update Responsemanagement response asset %s: %v", id, err)
	}
	return putResponseData, resp.StatusCode, nil
}

// getRespManagementRespAssetByIdFn is an implementation of the function to get a Genesys Cloud responsemanagement responseasset by Id
func getRespManagementRespAssetByIdFn(ctx context.Context, p *responsemanagementResponseassetProxy, id string) (*platformclientv2.Responseasset, *platformclientv2.APIResponse, error) {
	sdkAsset, resp, getErr := p.responseManagementApi.GetResponsemanagementResponseasset(id)
	if getErr != nil {
		return nil, nil, fmt.Errorf("failed to retrieve response asset: %s", getErr)
	}
	return sdkAsset, resp, nil
}

// deleteRespManagementRespAssetFn is an implementation function for deleting a Genesys Cloud responsemanagement responseasset
func deleteRespManagementRespAssetFn(ctx context.Context, p *responsemanagementResponseassetProxy, id string) (response *platformclientv2.APIResponse, err error) {
	resp, err := p.responseManagementApi.DeleteResponsemanagementResponseasset(id)
	if err != nil {
		return nil, fmt.Errorf("failed to delete response asset: %s", err)
	}
	return resp, nil
}
