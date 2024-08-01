package telephony_providers_edges_extension_pool

import (
	"context"

	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

var internalProxy *extensionPoolProxy

type getExtensionPoolFunc func(ctxctx context.Context, p *extensionPoolProxy, extensionPoolId string) (*platformclientv2.Extensionpool, *platformclientv2.APIResponse, error)
type deleteExtensionPoolFunc func(ctx context.Context, p *extensionPoolProxy, extensionPoolId string) (*platformclientv2.APIResponse, error)
type updateExtensionPoolFunc func(ctx context.Context, p *extensionPoolProxy, extensionPoolId string, body platformclientv2.Extensionpool) (*platformclientv2.Extensionpool, *platformclientv2.APIResponse, error)
type createExtensionPoolFunc func(ctx context.Context, p *extensionPoolProxy, body platformclientv2.Extensionpool) (*platformclientv2.Extensionpool, *platformclientv2.APIResponse, error)
type getAllExtensionPoolsFunc func(ctx context.Context, p *extensionPoolProxy) (*[]platformclientv2.Extensionpool, *platformclientv2.APIResponse, error)

// ExtensionPoolProxy represents the interface required to access the extension pool custom resource
type extensionPoolProxy struct {
	clientConfig *platformclientv2.Configuration
	edgesApi     *platformclientv2.TelephonyProvidersEdgeApi

	getExtensionPoolAttr     getExtensionPoolFunc
	deleteExtensionPoolAttr  deleteExtensionPoolFunc
	updateExtensionPoolAttr  updateExtensionPoolFunc
	createExtensionPoolAttr  createExtensionPoolFunc
	getAllExtensionPoolsAttr getAllExtensionPoolsFunc
}

func newExtensionPoolProxy(clientConfig *platformclientv2.Configuration) *extensionPoolProxy {
	edgesApi := platformclientv2.NewTelephonyProvidersEdgeApiWithConfig(clientConfig)

	return &extensionPoolProxy{
		clientConfig: clientConfig,
		edgesApi:     edgesApi,

		getExtensionPoolAttr:     getExtensionPoolFn,
		deleteExtensionPoolAttr:  deleteExtensionPoolFn,
		updateExtensionPoolAttr:  updateExtensionPoolFn,
		createExtensionPoolAttr:  createExtensionPoolFn,
		getAllExtensionPoolsAttr: getAllExtensionPoolsFn,
	}
}

func getExtensionPoolProxy(clientConfig *platformclientv2.Configuration) *extensionPoolProxy {
	if internalProxy == nil {
		internalProxy = newExtensionPoolProxy(clientConfig)
	}
	return internalProxy
}

func (p *extensionPoolProxy) getExtensionPool(ctx context.Context, extensionPoolId string) (*platformclientv2.Extensionpool, *platformclientv2.APIResponse, error) {
	return p.getExtensionPoolAttr(ctx, p, extensionPoolId)
}

func (p *extensionPoolProxy) deleteExtensionPool(ctx context.Context, extensionPoolId string) (*platformclientv2.APIResponse, error) {
	return p.deleteExtensionPoolAttr(ctx, p, extensionPoolId)
}

func (p *extensionPoolProxy) updateExtensionPool(ctx context.Context, extensionPoolId string, body platformclientv2.Extensionpool) (*platformclientv2.Extensionpool, *platformclientv2.APIResponse, error) {
	return p.updateExtensionPoolAttr(ctx, p, extensionPoolId, body)
}

func (p *extensionPoolProxy) createExtensionPool(ctx context.Context, body platformclientv2.Extensionpool) (*platformclientv2.Extensionpool, *platformclientv2.APIResponse, error) {
	return p.createExtensionPoolAttr(ctx, p, body)
}

func (p *extensionPoolProxy) getAllExtensionPools(ctx context.Context) (*[]platformclientv2.Extensionpool, *platformclientv2.APIResponse, error) {
	return p.getAllExtensionPoolsAttr(ctx, p)
}

func getExtensionPoolFn(ctx context.Context, p *extensionPoolProxy, extensionPoolId string) (*platformclientv2.Extensionpool, *platformclientv2.APIResponse, error) {
	extensionPool, resp, err := p.edgesApi.GetTelephonyProvidersEdgesExtensionpool(extensionPoolId)
	if err != nil {
		return nil, resp, err
	}

	return extensionPool, resp, nil
}

func deleteExtensionPoolFn(ctx context.Context, p *extensionPoolProxy, extensionPoolId string) (*platformclientv2.APIResponse, error) {
	resp, err := p.edgesApi.DeleteTelephonyProvidersEdgesExtensionpool(extensionPoolId)
	return resp, err
}

func updateExtensionPoolFn(ctx context.Context, p *extensionPoolProxy, extensionPoolId string, body platformclientv2.Extensionpool) (*platformclientv2.Extensionpool, *platformclientv2.APIResponse, error) {
	extensionPool, resp, err := p.edgesApi.PutTelephonyProvidersEdgesExtensionpool(extensionPoolId, body)
	if err != nil {
		return nil, resp, err
	}

	return extensionPool, resp, nil
}

func createExtensionPoolFn(ctx context.Context, p *extensionPoolProxy, body platformclientv2.Extensionpool) (*platformclientv2.Extensionpool, *platformclientv2.APIResponse, error) {
	extensionPool, resp, err := p.edgesApi.PostTelephonyProvidersEdgesExtensionpools(body)
	if err != nil {
		return nil, resp, err
	}

	return extensionPool, resp, nil
}

func getAllExtensionPoolsFn(ctx context.Context, p *extensionPoolProxy) (*[]platformclientv2.Extensionpool, *platformclientv2.APIResponse, error) {

	const pageSize = 100
	var (
		allExtensionPools []platformclientv2.Extensionpool
		pageNum           = 1
	)
	//Checking First Page
	extensionPools, resp, err := p.edgesApi.GetTelephonyProvidersEdgesExtensionpools(pageSize, pageNum, "", "")
	if err != nil {
		return nil, resp, err
	}
	if extensionPools.Entities != nil && len(*extensionPools.Entities) > 0 {
		for _, extensionPool := range *extensionPools.Entities {
			if extensionPool.State != nil && *extensionPool.State != "deleted" {
				allExtensionPools = append(allExtensionPools, extensionPool)
			}
		}
	}
	if *extensionPools.PageCount < 2 {
		return &allExtensionPools, resp, nil
	}

	for pageNum := 2; pageNum <= *extensionPools.PageCount; pageNum++ {
		extensionPools, resp, err := p.edgesApi.GetTelephonyProvidersEdgesExtensionpools(pageSize, pageNum, "", "")
		if err != nil {
			return nil, resp, err
		}
		if extensionPools.Entities == nil || len(*extensionPools.Entities) == 0 {
			break
		}
		for _, extensionPool := range *extensionPools.Entities {
			if extensionPool.State != nil && *extensionPool.State != "deleted" {
				allExtensionPools = append(allExtensionPools, extensionPool)
			}
		}
	}

	return &allExtensionPools, resp, nil
}
