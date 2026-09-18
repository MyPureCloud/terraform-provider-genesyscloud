package routing_email_route_identity_resolution

import (
	"context"

	"github.com/mypurecloud/platform-client-sdk-go/v199/platformclientv2"
	provider "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	routingEmailRoute "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/routing_email_route"
)

/*
The genesyscloud_routing_email_route_identity_resolution_proxy.go file contains the proxy structures and methods that interact
with the Genesys Cloud SDK. We use composition here for each function on the proxy so individual functions can be stubbed
out during testing.
*/

var internalProxy *routingEmailRouteIdentityResolutionProxy

type getRoutingEmailRouteIdentityResolutionFunc func(ctx context.Context, p *routingEmailRouteIdentityResolutionProxy, domainName string, routeId string) (*platformclientv2.Routeidentityresolutionconfig, *platformclientv2.APIResponse, error)
type putRoutingEmailRouteIdentityResolutionFunc func(ctx context.Context, p *routingEmailRouteIdentityResolutionProxy, domainName string, routeId string, config platformclientv2.Routeidentityresolutionconfig) (*platformclientv2.Routeidentityresolutionconfig, *platformclientv2.APIResponse, error)

type routingEmailRouteIdentityResolutionProxy struct {
	clientConfig                               *platformclientv2.Configuration
	routingApi                                 *platformclientv2.RoutingApi
	getRoutingEmailRouteIdentityResolutionAttr getRoutingEmailRouteIdentityResolutionFunc
	putRoutingEmailRouteIdentityResolutionAttr putRoutingEmailRouteIdentityResolutionFunc
	routingEmailRouteProxy                     *routingEmailRoute.RoutingEmailRouteProxy
}

func newRoutingEmailRouteIdentityResolutionProxy(clientConfig *platformclientv2.Configuration) *routingEmailRouteIdentityResolutionProxy {
	api := platformclientv2.NewRoutingApiWithConfig(clientConfig)

	return &routingEmailRouteIdentityResolutionProxy{
		clientConfig: clientConfig,
		routingApi:   api,
		getRoutingEmailRouteIdentityResolutionAttr: getRoutingEmailRouteIdentityResolutionFn,
		putRoutingEmailRouteIdentityResolutionAttr: putRoutingEmailRouteIdentityResolutionFn,
		routingEmailRouteProxy:                     routingEmailRoute.GetRoutingEmailRouteProxy(clientConfig),
	}
}

func getRoutingEmailRouteIdentityResolutionProxy(clientConfig *platformclientv2.Configuration) *routingEmailRouteIdentityResolutionProxy {
	if internalProxy == nil {
		internalProxy = newRoutingEmailRouteIdentityResolutionProxy(clientConfig)
	}
	return internalProxy
}

func (p *routingEmailRouteIdentityResolutionProxy) getRoutingEmailRouteIdentityResolution(ctx context.Context, domainName string, routeId string) (*platformclientv2.Routeidentityresolutionconfig, *platformclientv2.APIResponse, error) {
	return p.getRoutingEmailRouteIdentityResolutionAttr(ctx, p, domainName, routeId)
}

func (p *routingEmailRouteIdentityResolutionProxy) putRoutingEmailRouteIdentityResolution(ctx context.Context, domainName string, routeId string, config platformclientv2.Routeidentityresolutionconfig) (*platformclientv2.Routeidentityresolutionconfig, *platformclientv2.APIResponse, error) {
	return p.putRoutingEmailRouteIdentityResolutionAttr(ctx, p, domainName, routeId, config)
}

func getRoutingEmailRouteIdentityResolutionFn(ctx context.Context, p *routingEmailRouteIdentityResolutionProxy, domainName string, routeId string) (*platformclientv2.Routeidentityresolutionconfig, *platformclientv2.APIResponse, error) {
	// Set resource context for SDK debug logging
	ctx = provider.EnsureResourceContext(ctx, ResourceType)
	return p.routingApi.GetRoutingEmailDomainRouteIdentityresolution(domainName, routeId)
}

func putRoutingEmailRouteIdentityResolutionFn(ctx context.Context, p *routingEmailRouteIdentityResolutionProxy, domainName string, routeId string, config platformclientv2.Routeidentityresolutionconfig) (*platformclientv2.Routeidentityresolutionconfig, *platformclientv2.APIResponse, error) {
	// Set resource context for SDK debug logging
	ctx = provider.EnsureResourceContext(ctx, ResourceType)
	return p.routingApi.PutRoutingEmailDomainRouteIdentityresolution(domainName, routeId, config)
}
