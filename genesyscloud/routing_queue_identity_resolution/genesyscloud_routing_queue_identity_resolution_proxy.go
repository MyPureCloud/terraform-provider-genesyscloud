package routing_queue_identity_resolution

import (
	"context"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	routingQueue "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/routing_queue"

	"github.com/mypurecloud/platform-client-sdk-go/v193/platformclientv2"
)

var internalProxy *routingQueueIdentityResolutionProxy

type getRoutingQueueIdentityResolutionFunc func(ctx context.Context, p *routingQueueIdentityResolutionProxy, queueId string) (*platformclientv2.Identityresolutionqueueconfig, *platformclientv2.APIResponse, error)
type putRoutingQueueIdentityResolutionFunc func(ctx context.Context, p *routingQueueIdentityResolutionProxy, queueId string, config platformclientv2.Identityresolutionqueueconfig) (*platformclientv2.Identityresolutionqueueconfig, *platformclientv2.APIResponse, error)
type getRoutingQueueByIdFunc func(ctx context.Context, p *routingQueueIdentityResolutionProxy, queueId string) (*platformclientv2.Queue, *platformclientv2.APIResponse, error)

type routingQueueIdentityResolutionProxy struct {
	clientConfig                          *platformclientv2.Configuration
	routingApi                            *platformclientv2.RoutingApi
	getRoutingQueueIdentityResolutionAttr getRoutingQueueIdentityResolutionFunc
	putRoutingQueueIdentityResolutionAttr putRoutingQueueIdentityResolutionFunc
	getRoutingQueueByIdAttr               getRoutingQueueByIdFunc
	routingQueueProxy                     *routingQueue.RoutingQueueProxy
}

func newRoutingQueueIdentityResolutionProxy(clientConfig *platformclientv2.Configuration) *routingQueueIdentityResolutionProxy {
	api := platformclientv2.NewRoutingApiWithConfig(clientConfig)
	routingQueueProxy := routingQueue.GetRoutingQueueProxy(clientConfig)

	return &routingQueueIdentityResolutionProxy{
		clientConfig:                          clientConfig,
		routingApi:                            api,
		getRoutingQueueIdentityResolutionAttr: getRoutingQueueIdentityResolutionFn,
		putRoutingQueueIdentityResolutionAttr: putRoutingQueueIdentityResolutionFn,
		getRoutingQueueByIdAttr:               getRoutingQueueByIdFn,
		routingQueueProxy:                     routingQueueProxy,
	}
}

func getRoutingQueueIdentityResolutionProxy(clientConfig *platformclientv2.Configuration) *routingQueueIdentityResolutionProxy {
	if internalProxy == nil {
		internalProxy = newRoutingQueueIdentityResolutionProxy(clientConfig)
	}
	return internalProxy
}

func (p *routingQueueIdentityResolutionProxy) getRoutingQueueIdentityResolution(ctx context.Context, queueId string) (*platformclientv2.Identityresolutionqueueconfig, *platformclientv2.APIResponse, error) {
	return p.getRoutingQueueIdentityResolutionAttr(ctx, p, queueId)
}

func (p *routingQueueIdentityResolutionProxy) putRoutingQueueIdentityResolution(ctx context.Context, queueId string, config platformclientv2.Identityresolutionqueueconfig) (*platformclientv2.Identityresolutionqueueconfig, *platformclientv2.APIResponse, error) {
	return p.putRoutingQueueIdentityResolutionAttr(ctx, p, queueId, config)
}

func (p *routingQueueIdentityResolutionProxy) getRoutingQueueById(ctx context.Context, queueId string) (*platformclientv2.Queue, *platformclientv2.APIResponse, error) {
	return p.getRoutingQueueByIdAttr(ctx, p, queueId)
}

func getRoutingQueueIdentityResolutionFn(ctx context.Context, p *routingQueueIdentityResolutionProxy, queueId string) (*platformclientv2.Identityresolutionqueueconfig, *platformclientv2.APIResponse, error) {
	ctx = provider.EnsureResourceContext(ctx, ResourceType)
	return p.routingApi.GetRoutingQueueIdentityresolution(queueId)
}

func putRoutingQueueIdentityResolutionFn(ctx context.Context, p *routingQueueIdentityResolutionProxy, queueId string, config platformclientv2.Identityresolutionqueueconfig) (*platformclientv2.Identityresolutionqueueconfig, *platformclientv2.APIResponse, error) {
	ctx = provider.EnsureResourceContext(ctx, ResourceType)
	return p.routingApi.PutRoutingQueueIdentityresolution(queueId, config)
}

func getRoutingQueueByIdFn(ctx context.Context, p *routingQueueIdentityResolutionProxy, queueId string) (*platformclientv2.Queue, *platformclientv2.APIResponse, error) {
	ctx = provider.EnsureResourceContext(ctx, ResourceType)
	return p.routingApi.GetRoutingQueue(queueId, nil)
}
