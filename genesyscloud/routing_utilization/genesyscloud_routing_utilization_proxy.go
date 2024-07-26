package routing_utilization

import (
	"context"

	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

var internalProxy *routingUtilizationProxy

type getRoutingUtilizationFunc func(ctx context.Context, p *routingUtilizationProxy) (*platformclientv2.Utilizationresponse, *platformclientv2.APIResponse, error)
type updateRoutingUtilizationFunc func(ctx context.Context, p *routingUtilizationProxy, request *platformclientv2.Utilizationrequest) (*platformclientv2.Utilizationresponse, *platformclientv2.APIResponse, error)
type deleteRoutingUtilizationFunc func(ctx context.Context, p *routingUtilizationProxy) (*platformclientv2.APIResponse, error)

type routingUtilizationProxy struct {
	clientConfig                 *platformclientv2.Configuration
	routingApi                   *platformclientv2.RoutingApi
	getRoutingUtilizationAttr    getRoutingUtilizationFunc
	updateRoutingUtilizationAttr updateRoutingUtilizationFunc
	deleteRoutingUtilizationAttr deleteRoutingUtilizationFunc
}

func newRoutingUtilizationProxy(clientConfig *platformclientv2.Configuration) *routingUtilizationProxy {
	api := platformclientv2.NewRoutingApiWithConfig(clientConfig)
	return &routingUtilizationProxy{
		clientConfig:                 clientConfig,
		routingApi:                   api,
		getRoutingUtilizationAttr:    getRoutingUtilizationFn,
		updateRoutingUtilizationAttr: updateRoutingUtilizationFn,
		deleteRoutingUtilizationAttr: deleteRoutingUtilizationFn,
	}
}

func getRoutingUtilizationProxy(clientConfig *platformclientv2.Configuration) *routingUtilizationProxy {
	if internalProxy == nil {
		internalProxy = newRoutingUtilizationProxy(clientConfig)
	}
	return internalProxy
}

func (p *routingUtilizationProxy) getRoutingUtilization(ctx context.Context) (*platformclientv2.Utilizationresponse, *platformclientv2.APIResponse, error) {
	return p.getRoutingUtilizationAttr(ctx, p)
}

func (p *routingUtilizationProxy) updateRoutingUtilization(ctx context.Context, request *platformclientv2.Utilizationrequest) (*platformclientv2.Utilizationresponse, *platformclientv2.APIResponse, error) {
	return p.updateRoutingUtilizationAttr(ctx, p, request)
}

func (p *routingUtilizationProxy) deleteRoutingUtilization(ctx context.Context) (*platformclientv2.APIResponse, error) {
	return p.deleteRoutingUtilizationAttr(ctx, p)
}

func getRoutingUtilizationFn(ctx context.Context, p *routingUtilizationProxy) (*platformclientv2.Utilizationresponse, *platformclientv2.APIResponse, error) {
	return p.routingApi.GetRoutingUtilization()
}

func updateRoutingUtilizationFn(ctx context.Context, p *routingUtilizationProxy, utilizationRequest *platformclientv2.Utilizationrequest) (*platformclientv2.Utilizationresponse, *platformclientv2.APIResponse, error) {
	return p.routingApi.PutRoutingUtilization(*utilizationRequest)
}

func deleteRoutingUtilizationFn(ctx context.Context, p *routingUtilizationProxy) (*platformclientv2.APIResponse, error) {
	return p.routingApi.DeleteRoutingUtilization()
}
