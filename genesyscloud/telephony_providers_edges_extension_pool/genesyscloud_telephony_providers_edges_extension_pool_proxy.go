package telephony_providers_edges_extension_pool

import (
	"context"

	"github.com/mypurecloud/platform-client-sdk-go/v121/platformclientv2"
)

var internalProxy *extensionPoolProxy

type getExtensionPoolFunc func(ctxctx context.Context, p *extensionPoolProxy, extensionPoolId string) (*platformclientv2.Extensionpool, *platformclientv2.APIResponse, error)
type deleteExtensionPoolFunc func(ctx context.Context, p *extensionPoolProxy, extensionPoolId string) (*platformclientv2.APIResponse, error)
type putExtensionPoolFunc func(ctx context.Context, p *extensionPoolProxy, extensionPoolId string, body platformclientv2.Extensionpool) (*platformclientv2.Extensionpool, *platformclientv2.APIResponse, error)
type postExtensionPoolFunc func(ctx context.Context, p *extensionPoolProxy, body platformclientv2.Extensionpool) (*platformclientv2.Extensionpool, *platformclientv2.APIResponse, error)
type getAllExtensionPoolsFunc func(ctx context.Context, p *extensionPoolProxy) (*[]platformclientv2.Extensionpool, error)

// ExtensionPoolProxy represents the interface required to access the extension pool custom resource
type extensionPoolProxy struct {
	clientConfig *platformclientv2.Configuration
	edgesApi     *platformclientv2.TelephonyProvidersEdgeApi

	getExtensionPoolAttr     getExtensionPoolFunc
	deleteExtensionPoolAttr  deleteExtensionPoolFunc
	putExtensionPoolAttr     putExtensionPoolFunc
	postExtensionPoolAttr    postExtensionPoolFunc
	getAllExtensionPoolsAttr getAllExtensionPoolsFunc
}

func newExtensionPoolProxy(clientConfig *platformclientv2.Configuration) *extensionPoolProxy {
	edgesApi := platformclientv2.NewTelephonyProvidersEdgeApiWithConfig(clientConfig)

	return &extensionPoolProxy{
		clientConfig: clientConfig,
		edgesApi:     edgesApi,

		getExtensionPoolAttr:     getExtensionPoolFn,
		deleteExtensionPoolAttr:  deleteExtensionPoolFn,
		putExtensionPoolAttr:     putExtensionPoolFn,
		postExtensionPoolAttr:    postExtensionPoolFn,
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

func (p *extensionPoolProxy) putExtensionPool(ctx context.Context, extensionPoolId string, body platformclientv2.Extensionpool) (*platformclientv2.Extensionpool, *platformclientv2.APIResponse, error) {
	return p.putExtensionPoolAttr(ctx, p, extensionPoolId, body)
}

func (p *extensionPoolProxy) postExtensionPool(ctx context.Context, body platformclientv2.Extensionpool) (*platformclientv2.Extensionpool, *platformclientv2.APIResponse, error) {
	return p.postExtensionPoolAttr(ctx, p, body)
}

func (p *extensionPoolProxy) getAllExtensionPools(ctx context.Context) (*[]platformclientv2.Extensionpool, error) {
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

func putExtensionPoolFn(ctx context.Context, p *extensionPoolProxy, extensionPoolId string, body platformclientv2.Extensionpool) (*platformclientv2.Extensionpool, *platformclientv2.APIResponse, error) {
	extensionPool, resp, err := p.edgesApi.PutTelephonyProvidersEdgesExtensionpool(extensionPoolId, body)
	if err != nil {
		return nil, resp, err
	}

	return extensionPool, resp, nil
}

func postExtensionPoolFn(ctx context.Context, p *extensionPoolProxy, body platformclientv2.Extensionpool) (*platformclientv2.Extensionpool, *platformclientv2.APIResponse, error) {
	extensionPool, resp, err := p.edgesApi.PostTelephonyProvidersEdgesExtensionpools(body)
	if err != nil {
		return nil, resp, err
	}

	return extensionPool, resp, nil
}
