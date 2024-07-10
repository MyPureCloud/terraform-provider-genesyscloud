package responsemanagement_responseasset

import (
	"context"
	"fmt"
	"log"
	rc "terraform-provider-genesyscloud/genesyscloud/resource_cache"

	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

/*
The genesyscloud_responsemanagement_responseasset_proxy.go file contains the proxy structures and methods that interact
with the Genesys Cloud SDK. We use composition here for each function on the proxy so individual functions can be stubbed
out during testing.
*/

// internalProxy holds a proxy instance that can be used throughout the package
var internalProxy *responsemanagementResponseassetProxy

// Type definitions for each func on our proxy so we can easily mock them out later
type getAllResponseAssetsFunc func(ctx context.Context, p *responsemanagementResponseassetProxy) (*[]platformclientv2.Responseasset, *platformclientv2.APIResponse, error)
type createRespManagementRespAssetFunc func(ctx context.Context, p *responsemanagementResponseassetProxy, respAsset *platformclientv2.Createresponseassetrequest) (*platformclientv2.Createresponseassetresponse, *platformclientv2.APIResponse, error)
type updateRespManagementRespAssetFunc func(ctx context.Context, p *responsemanagementResponseassetProxy, id string, respAsset *platformclientv2.Responseassetrequest) (*platformclientv2.Responseasset, *platformclientv2.APIResponse, error)
type getRespManagementRespAssetByIdFunc func(ctx context.Context, p *responsemanagementResponseassetProxy, id string) (*platformclientv2.Responseasset, *platformclientv2.APIResponse, error)
type getRespManagementRespAssetByNameFunc func(ctx context.Context, p *responsemanagementResponseassetProxy, name string) (string, bool, *platformclientv2.APIResponse, error)
type deleteRespManagementRespAssetFunc func(ctx context.Context, p *responsemanagementResponseassetProxy, id string) (response *platformclientv2.APIResponse, err error)

// responsemanagementResponseassetProxy contains all of the methods that call genesys cloud APIs.
type responsemanagementResponseassetProxy struct {
	clientConfig                         *platformclientv2.Configuration
	responseManagementApi                *platformclientv2.ResponseManagementApi
	getAllResponseAssetsAttr             getAllResponseAssetsFunc
	createRespManagementRespAssetAttr    createRespManagementRespAssetFunc
	updateRespManagementRespAssetAttr    updateRespManagementRespAssetFunc
	getRespManagementRespAssetByIdAttr   getRespManagementRespAssetByIdFunc
	getRespManagementRespAssetByNameAttr getRespManagementRespAssetByNameFunc
	deleteRespManagementRespAssetAttr    deleteRespManagementRespAssetFunc
	assetCache                           rc.CacheInterface[platformclientv2.Responseasset]
}

// newRespManagementRespAssetProxy initializes the responsemanagement responseasset proxy with all of the data needed to communicate with Genesys Cloud
func newRespManagementRespAssetProxy(clientConfig *platformclientv2.Configuration) *responsemanagementResponseassetProxy {
	api := platformclientv2.NewResponseManagementApiWithConfig(clientConfig)
	assetCache := rc.NewResourceCache[platformclientv2.Responseasset]()
	return &responsemanagementResponseassetProxy{
		clientConfig:                         clientConfig,
		responseManagementApi:                api,
		getAllResponseAssetsAttr:             getAllResponseAssetsFn,
		createRespManagementRespAssetAttr:    createRespManagementRespAssetFn,
		updateRespManagementRespAssetAttr:    updateRespManagementRespAssetFn,
		getRespManagementRespAssetByIdAttr:   getRespManagementRespAssetByIdFn,
		getRespManagementRespAssetByNameAttr: getRespManagementRespAssetByNameFn,
		deleteRespManagementRespAssetAttr:    deleteRespManagementRespAssetFn,
		assetCache:                           assetCache,
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

func (p *responsemanagementResponseassetProxy) getAllResponseAssets(ctx context.Context) (*[]platformclientv2.Responseasset, *platformclientv2.APIResponse, error) {
	return p.getAllResponseAssetsAttr(ctx, p)
}

// createRespManagementRespAsset creates a Genesys Cloud responsemanagement responseasset by Id
func (p *responsemanagementResponseassetProxy) createRespManagementRespAsset(ctx context.Context, respAsset *platformclientv2.Createresponseassetrequest) (*platformclientv2.Createresponseassetresponse, *platformclientv2.APIResponse, error) {
	return p.createRespManagementRespAssetAttr(ctx, p, respAsset)
}

// updateRespManagementRespAsset updates a Genesys Cloud responsemanagement responseasset by Id
func (p *responsemanagementResponseassetProxy) updateRespManagementRespAsset(ctx context.Context, id string, respAsset *platformclientv2.Responseassetrequest) (*platformclientv2.Responseasset, *platformclientv2.APIResponse, error) {
	return p.updateRespManagementRespAssetAttr(ctx, p, id, respAsset)
}

// getRespManagementRespAssetById returns a single Genesys Cloud responsemanagement responseasset by Id
func (p *responsemanagementResponseassetProxy) getRespManagementRespAssetById(ctx context.Context, id string) (*platformclientv2.Responseasset, *platformclientv2.APIResponse, error) {
	return p.getRespManagementRespAssetByIdAttr(ctx, p, id)
}
func (p *responsemanagementResponseassetProxy) getRespManagementRespAssetByName(ctx context.Context, name string) (string, bool, *platformclientv2.APIResponse, error) {
	return p.getRespManagementRespAssetByNameAttr(ctx, p, name)
}

// deleteRespManagementRespAsset deletes a Genesys Cloud responsemanagement responseasset by Id
func (p *responsemanagementResponseassetProxy) deleteRespManagementRespAsset(ctx context.Context, id string) (response *platformclientv2.APIResponse, err error) {
	return p.deleteRespManagementRespAssetAttr(ctx, p, id)
}

func getAllResponseAssetsFn(ctx context.Context, p *responsemanagementResponseassetProxy) (*[]platformclientv2.Responseasset, *platformclientv2.APIResponse, error) {
	var allResponseAssets []platformclientv2.Responseasset
	var response *platformclientv2.APIResponse
	pageSize := 100

	responseAssets, resp, err := p.responseManagementApi.PostResponsemanagementResponseassetsSearch(platformclientv2.Responseassetsearchrequest{
		PageSize:   &pageSize,
		PageNumber: platformclientv2.Int(1),
	}, []string{})
	response = resp
	if err != nil {
		return nil, resp, fmt.Errorf("Failed to get response asset search request: %v", err)
	}

	if responseAssets.Results == nil || len(*responseAssets.Results) == 0 {
		return &allResponseAssets, resp, nil
	}
	allResponseAssets = append(allResponseAssets, *responseAssets.Results...)

	for pageNum := 2; pageNum <= *responseAssets.PageCount; pageNum++ {
		responseAssets, resp, err := p.responseManagementApi.PostResponsemanagementResponseassetsSearch(platformclientv2.Responseassetsearchrequest{
			PageSize:   &pageSize,
			PageNumber: &pageNum,
		}, []string{})
		response = resp
		if err != nil {
			return nil, resp, fmt.Errorf("Failed to get response asset search request: %v", err)
		}

		if responseAssets.Results == nil || len(*responseAssets.Results) == 0 {
			break
		}
		allResponseAssets = append(allResponseAssets, *responseAssets.Results...)
	}

	for _, asset := range allResponseAssets {
		rc.SetCache(p.assetCache, *asset.Id, asset)
	}

	return &allResponseAssets, response, nil
}

// createRespManagementRespAssetFn is an implementation of the function to create a Genesys Cloud responsemanagement responseasset
func createRespManagementRespAssetFn(ctx context.Context, p *responsemanagementResponseassetProxy, respAsset *platformclientv2.Createresponseassetrequest) (*platformclientv2.Createresponseassetresponse, *platformclientv2.APIResponse, error) {
	postResponseData, resp, err := p.responseManagementApi.PostResponsemanagementResponseassetsUploads(*respAsset)
	if err != nil {
		return nil, resp, fmt.Errorf("Failed to upload response asset: %v", err)
	}
	return postResponseData, resp, nil
}

// updateRespManagementRespAssetFn is an implementation of the function to update a Genesys Cloud responsemanagement responseasset
func updateRespManagementRespAssetFn(ctx context.Context, p *responsemanagementResponseassetProxy, id string, respAsset *platformclientv2.Responseassetrequest) (*platformclientv2.Responseasset, *platformclientv2.APIResponse, error) {
	putResponseData, resp, err := p.responseManagementApi.PutResponsemanagementResponseasset(id, *respAsset)
	if err != nil {
		return nil, resp, fmt.Errorf("Failed to update Responsemanagement response asset %s: %v", id, err)
	}
	return putResponseData, resp, nil
}

// getRespManagementRespAssetByIdFn is an implementation of the function to get a Genesys Cloud responsemanagement responseasset by Id
func getRespManagementRespAssetByIdFn(ctx context.Context, p *responsemanagementResponseassetProxy, id string) (*platformclientv2.Responseasset, *platformclientv2.APIResponse, error) {
	asset := rc.GetCacheItem(p.assetCache, id)
	if asset != nil {
		return asset, nil, nil
	}

	sdkAsset, resp, getErr := p.responseManagementApi.GetResponsemanagementResponseasset(id)
	if getErr != nil {
		return nil, resp, fmt.Errorf("failed to retrieve response asset: %s", getErr)
	}
	return sdkAsset, resp, nil
}

func getRespManagementRespAssetByNameFn(ctx context.Context, p *responsemanagementResponseassetProxy, name string) (string, bool, *platformclientv2.APIResponse, error) {
	var (
		field   = "name"
		fields  = []string{field}
		varType = "TERM"
		filter  = platformclientv2.Responseassetfilter{
			Fields:  &fields,
			Value:   &name,
			VarType: &varType,
		}
		body = platformclientv2.Responseassetsearchrequest{
			Query:  &[]platformclientv2.Responseassetfilter{filter},
			SortBy: &field,
		}
	)

	respAssets, resp, err := p.responseManagementApi.PostResponsemanagementResponseassetsSearch(body, nil)
	if err != nil {
		return "", false, resp, err
	}

	if respAssets == nil || len(*respAssets.Results) == 0 {
		return "", true, resp, fmt.Errorf("No responsemanagement response asset found with name %s", name)
	}

	for _, asset := range *respAssets.Results {
		if *asset.Name == name {
			log.Printf("Retrieved the responsemanagement response asset id %s by name %s", *asset.Id, name)
			return *asset.Id, false, resp, nil
		}
	}
	return "", true, resp, fmt.Errorf("Unable to find responsemanagement response asset with name %s", name)
}

// deleteRespManagementRespAssetFn is an implementation function for deleting a Genesys Cloud responsemanagement responseasset
func deleteRespManagementRespAssetFn(ctx context.Context, p *responsemanagementResponseassetProxy, id string) (response *platformclientv2.APIResponse, err error) {
	resp, err := p.responseManagementApi.DeleteResponsemanagementResponseasset(id)
	if err != nil {
		return resp, fmt.Errorf("failed to delete response asset: %s", err)
	}
	return resp, nil
}
