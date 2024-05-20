package telephony_providers_edges_site_outbound_route

import (
	"context"
	"github.com/mypurecloud/platform-client-sdk-go/v129/platformclientv2"
)

/*
The genesyscloud_telephony_providers_edges_site_outbound_route_proxy.go file contains the proxy structures and methods that interact
with the Genesys Cloud SDK. We use composition here for each function on the proxy so individual functions can be stubbed
out during testing.
*/

// internalProxy holds a proxy instance that can be used throughout the package
var internalProxy *siteProxy

// Type definitions for each func on our proxy so we can easily mock them out later
type getSiteFunc func(ctx context.Context, p *siteProxy, siteId string) (*platformclientv2.Site, *platformclientv2.APIResponse, error)
type createSiteOutboundRouteFunc func(ctx context.Context, p *siteProxy, siteId string, outboundRoute *platformclientv2.Outboundroutebase) (*platformclientv2.Outboundroutebase, *platformclientv2.APIResponse, error)
type getSiteOutboundRoutesFunc func(ctx context.Context, p *siteProxy, siteId string) (*[]platformclientv2.Outboundroutebase, *platformclientv2.APIResponse, error)
type updateSiteOutboundRouteFunc func(ctx context.Context, p *siteProxy, siteId string, outboundRouteId string, outboundRoute *platformclientv2.Outboundroutebase) (*platformclientv2.Outboundroutebase, *platformclientv2.APIResponse, error)
type deleteSiteOutboundRouteFunc func(ctx context.Context, p *siteProxy, siteId string, outboundRouteId string) (*platformclientv2.APIResponse, error)

// siteProxy contains all of the methods that call genesys cloud APIs.
type siteProxy struct {
	clientConfig *platformclientv2.Configuration
	edgesApi     *platformclientv2.TelephonyProvidersEdgeApi

	getSiteAttr                 getSiteFunc
	createSiteOutboundRouteAttr createSiteOutboundRouteFunc
	getSiteOutboundRoutesAttr   getSiteOutboundRoutesFunc
	updateSiteOutboundRouteAttr updateSiteOutboundRouteFunc
	deleteSiteOutboundRouteAttr deleteSiteOutboundRouteFunc
}

// newSiteProxy initializes the Site proxy with all the data needed to communicate with Genesys Cloud
func newSiteProxy(clientConfig *platformclientv2.Configuration) *siteProxy {
	edgesApi := platformclientv2.NewTelephonyProvidersEdgeApiWithConfig(clientConfig)

	return &siteProxy{
		clientConfig: clientConfig,
		edgesApi:     edgesApi,

		getSiteAttr:                 getSiteFn,
		createSiteOutboundRouteAttr: createSiteOutboundRouteFn,
		getSiteOutboundRoutesAttr:   getSiteOutboundRoutesFn,
		updateSiteOutboundRouteAttr: updateSiteOutboundRouteFn,
		deleteSiteOutboundRouteAttr: deleteSiteOutboundRouteFn,
	}
}

// getSiteOutboundRouteProxy acts as a singleton for the internalProxy.  It also ensures
// that we can still proxy our tests by directly setting internalProxy package variable
func getSiteOutboundRouteProxy(clientConfig *platformclientv2.Configuration) *siteProxy {
	if internalProxy == nil {
		internalProxy = newSiteProxy(clientConfig)
	}
	return internalProxy
}

func (p *siteProxy) getSite(ctx context.Context, id string) (*platformclientv2.Site, *platformclientv2.APIResponse, error) {
	return p.getSiteAttr(ctx, p, id)
}

// createSiteOutboundRouteFunc creates an Outbound Route for a Genesys Cloud Site
func (p *siteProxy) createSiteOutboundRoute(ctx context.Context, siteId string, outboundRoute *platformclientv2.Outboundroutebase) (*platformclientv2.Outboundroutebase, *platformclientv2.APIResponse, error) {
	return p.createSiteOutboundRouteAttr(ctx, p, siteId, outboundRoute)
}

// getSiteByIdFunc returns a single Outbound Route by Id
func (p *siteProxy) getSiteOutboundRoutes(ctx context.Context, siteId string) (*[]platformclientv2.Outboundroutebase, *platformclientv2.APIResponse, error) {
	return p.getSiteOutboundRoutesAttr(ctx, p, siteId)
}

// updateSiteFunc updates a Genesys Cloud Outbound Route for a Genesys Cloud Site
func (p *siteProxy) updateSiteOutboundRoute(ctx context.Context, siteId string, outboundRouteId string, outboundRoute *platformclientv2.Outboundroutebase) (*platformclientv2.Outboundroutebase, *platformclientv2.APIResponse, error) {
	return p.updateSiteOutboundRouteAttr(ctx, p, siteId, outboundRouteId, outboundRoute)
}

// deleteSiteFunc deletes a Genesys Cloud Outbound Route by Id for a Genesys Cloud Site
func (p *siteProxy) deleteSiteOutboundRoute(ctx context.Context, siteId string, outboundRouteId string) (*platformclientv2.APIResponse, error) {
	return p.deleteSiteOutboundRouteAttr(ctx, p, siteId, outboundRouteId)
}

func createSiteOutboundRouteFn(ctx context.Context, p *siteProxy, id string, route *platformclientv2.Outboundroutebase) (*platformclientv2.Outboundroutebase, *platformclientv2.APIResponse, error) {
	return p.edgesApi.PostTelephonyProvidersEdgesSiteOutboundroutes(id, *route)
}

func getSiteFn(ctx context.Context, p *siteProxy, id string) (*platformclientv2.Site, *platformclientv2.APIResponse, error) {
	return p.edgesApi.GetTelephonyProvidersEdgesSite(id)
}

// getSiteOutboundRoutesFn is an implementation function for getting an outbound route for a Genesys Cloud Site
func getSiteOutboundRoutesFn(ctx context.Context, p *siteProxy, siteId string) (*[]platformclientv2.Outboundroutebase, *platformclientv2.APIResponse, error) {
	var allOutboundRoutes = []platformclientv2.Outboundroutebase{}
	const pageSize = 100
	outboundRoutes, resp, err := p.edgesApi.GetTelephonyProvidersEdgesSiteOutboundroutes(siteId, pageSize, 1, "", "", "")
	if err != nil {
		return nil, resp, err
	}
	allOutboundRoutes = append(allOutboundRoutes, *outboundRoutes.Entities...)

	for pageNum := 2; pageNum <= *outboundRoutes.PageCount; pageNum++ {
		outboundRoutes, resp, err := p.edgesApi.GetTelephonyProvidersEdgesSiteOutboundroutes(siteId, pageSize, pageNum, "", "", "")
		if err != nil {
			return nil, resp, err
		}
		if outboundRoutes.Entities == nil {
			break
		}
		allOutboundRoutes = append(allOutboundRoutes, *outboundRoutes.Entities...)
	}
	return &allOutboundRoutes, resp, nil
}

// updateSiteOutboundRouteFn is an implementation function for updating an outbound route for a Genesys Cloud Site
func updateSiteOutboundRouteFn(ctx context.Context, p *siteProxy, siteId string, outboundRouteId string, outboundRoute *platformclientv2.Outboundroutebase) (*platformclientv2.Outboundroutebase, *platformclientv2.APIResponse, error) {
	return p.edgesApi.PutTelephonyProvidersEdgesSiteOutboundroute(siteId, outboundRouteId, *outboundRoute)
}

// deleteSiteOutboundRouteFn is an implementation function for deleting an outbound route for a Genesys Cloud Site
func deleteSiteOutboundRouteFn(ctx context.Context, p *siteProxy, siteId string, outboundRouteId string) (*platformclientv2.APIResponse, error) {
	return p.edgesApi.DeleteTelephonyProvidersEdgesSiteOutboundroute(siteId, outboundRouteId)
}
