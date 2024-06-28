package telephony_providers_edges_site_outbound_route

import (
	"context"
	rc "terraform-provider-genesyscloud/genesyscloud/resource_cache"
	telephonyProvidersEdgesSite "terraform-provider-genesyscloud/genesyscloud/telephony_providers_edges_site"

	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

/*
The genesyscloud_telephony_providers_edges_site_outbound_route_proxy.go file contains the proxy structures and methods that interact
with the Genesys Cloud SDK. We use composition here for each function on the proxy so individual functions can be stubbed
out during testing.
*/

// internalProxy holds a proxy instance that can be used throughout the package
var internalProxy *siteOutboundRoutesProxy

// Type definitions for each func on our proxy so we can easily mock them out later
type getSiteFunc func(ctx context.Context, p *siteOutboundRoutesProxy, siteId string) (*platformclientv2.Site, *platformclientv2.APIResponse, error)
type createSiteOutboundRouteFunc func(ctx context.Context, p *siteOutboundRoutesProxy, siteId string, outboundRoute *platformclientv2.Outboundroutebase) (*platformclientv2.Outboundroutebase, *platformclientv2.APIResponse, error)
type getSiteOutboundRoutesFunc func(ctx context.Context, p *siteOutboundRoutesProxy, siteId string) (*[]platformclientv2.Outboundroutebase, *platformclientv2.APIResponse, error)
type updateSiteOutboundRouteFunc func(ctx context.Context, p *siteOutboundRoutesProxy, siteId string, outboundRouteId string, outboundRoute *platformclientv2.Outboundroutebase) (*platformclientv2.Outboundroutebase, *platformclientv2.APIResponse, error)
type deleteSiteOutboundRouteFunc func(ctx context.Context, p *siteOutboundRoutesProxy, siteId string, outboundRouteId string) (*platformclientv2.APIResponse, error)

// siteOutboundRoutesProxy contains all the methods that call genesys cloud APIs.
type siteOutboundRoutesProxy struct {
	clientConfig *platformclientv2.Configuration
	edgesApi     *platformclientv2.TelephonyProvidersEdgeApi

	getSiteAttr                 getSiteFunc
	createSiteOutboundRouteAttr createSiteOutboundRouteFunc
	getSiteOutboundRoutesAttr   getSiteOutboundRoutesFunc
	updateSiteOutboundRouteAttr updateSiteOutboundRouteFunc
	deleteSiteOutboundRouteAttr deleteSiteOutboundRouteFunc
	siteOutboundRouteCache      rc.CacheInterface[[]platformclientv2.Outboundroutebase]
	siteProxy                   *telephonyProvidersEdgesSite.SiteProxy
}

// newSiteOutboundRoutesProxy initializes the Site proxy with all the data needed to communicate with Genesys Cloud
func newSiteOutboundRoutesProxy(clientConfig *platformclientv2.Configuration) *siteOutboundRoutesProxy {
	edgesApi := platformclientv2.NewTelephonyProvidersEdgeApiWithConfig(clientConfig)
	siteProxy := telephonyProvidersEdgesSite.GetSiteProxy(clientConfig)
	siteOutboundRouteCache := rc.NewResourceCache[[]platformclientv2.Outboundroutebase]()

	return &siteOutboundRoutesProxy{
		clientConfig: clientConfig,
		edgesApi:     edgesApi,

		getSiteAttr:                 getSiteFn,
		createSiteOutboundRouteAttr: createSiteOutboundRouteFn,
		getSiteOutboundRoutesAttr:   getSiteOutboundRoutesFn,
		updateSiteOutboundRouteAttr: updateSiteOutboundRouteFn,
		deleteSiteOutboundRouteAttr: deleteSiteOutboundRouteFn,
		siteOutboundRouteCache:      siteOutboundRouteCache,
		siteProxy:                   siteProxy,
	}
}

// getSiteOutboundRouteProxy acts as a singleton for the internalProxy.  It also ensures
// that we can still proxy our tests by directly setting internalProxy package variable
func getSiteOutboundRouteProxy(clientConfig *platformclientv2.Configuration) *siteOutboundRoutesProxy {
	if internalProxy == nil {
		internalProxy = newSiteOutboundRoutesProxy(clientConfig)
	}
	return internalProxy
}

func (p *siteOutboundRoutesProxy) getSite(ctx context.Context, id string) (*platformclientv2.Site, *platformclientv2.APIResponse, error) {
	return p.getSiteAttr(ctx, p, id)
}

// createSiteOutboundRouteFunc creates an Outbound Route for a Genesys Cloud Site
func (p *siteOutboundRoutesProxy) createSiteOutboundRoute(ctx context.Context, siteId string, outboundRoute *platformclientv2.Outboundroutebase) (*platformclientv2.Outboundroutebase, *platformclientv2.APIResponse, error) {
	return p.createSiteOutboundRouteAttr(ctx, p, siteId, outboundRoute)
}

// getSiteByIdFunc returns a single Outbound Route by Id
func (p *siteOutboundRoutesProxy) getSiteOutboundRoutes(ctx context.Context, siteId string) (*[]platformclientv2.Outboundroutebase, *platformclientv2.APIResponse, error) {
	return p.getSiteOutboundRoutesAttr(ctx, p, siteId)
}

// updateSiteFunc updates a Genesys Cloud Outbound Route for a Genesys Cloud Site
func (p *siteOutboundRoutesProxy) updateSiteOutboundRoute(ctx context.Context, siteId string, outboundRouteId string, outboundRoute *platformclientv2.Outboundroutebase) (*platformclientv2.Outboundroutebase, *platformclientv2.APIResponse, error) {
	return p.updateSiteOutboundRouteAttr(ctx, p, siteId, outboundRouteId, outboundRoute)
}

// deleteSiteFunc deletes a Genesys Cloud Outbound Route by Id for a Genesys Cloud Site
func (p *siteOutboundRoutesProxy) deleteSiteOutboundRoute(ctx context.Context, siteId string, outboundRouteId string) (*platformclientv2.APIResponse, error) {
	return p.deleteSiteOutboundRouteAttr(ctx, p, siteId, outboundRouteId)
}

func getSiteFn(ctx context.Context, p *siteOutboundRoutesProxy, id string) (*platformclientv2.Site, *platformclientv2.APIResponse, error) {
	return p.edgesApi.GetTelephonyProvidersEdgesSite(id)
}

func createSiteOutboundRouteFn(ctx context.Context, p *siteOutboundRoutesProxy, id string, route *platformclientv2.Outboundroutebase) (*platformclientv2.Outboundroutebase, *platformclientv2.APIResponse, error) {
	return p.edgesApi.PostTelephonyProvidersEdgesSiteOutboundroutes(id, *route)
}

// getSiteOutboundRoutesFn is an implementation function for getting an outbound route for a Genesys Cloud Site
func getSiteOutboundRoutesFn(ctx context.Context, p *siteOutboundRoutesProxy, siteId string) (*[]platformclientv2.Outboundroutebase, *platformclientv2.APIResponse, error) {
	// Check if site's outbound routes exist in cache
	routes := rc.GetCacheItem(p.siteOutboundRouteCache, siteId)
	if routes != nil && len(*routes) != 0 {
		return routes, nil, nil
	}

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

	rc.SetCache(p.siteOutboundRouteCache, siteId, allOutboundRoutes)

	return &allOutboundRoutes, resp, nil
}

// updateSiteOutboundRouteFn is an implementation function for updating an outbound route for a Genesys Cloud Site
func updateSiteOutboundRouteFn(ctx context.Context, p *siteOutboundRoutesProxy, siteId string, outboundRouteId string, outboundRoute *platformclientv2.Outboundroutebase) (*platformclientv2.Outboundroutebase, *platformclientv2.APIResponse, error) {
	return p.edgesApi.PutTelephonyProvidersEdgesSiteOutboundroute(siteId, outboundRouteId, *outboundRoute)
}

// deleteSiteOutboundRouteFn is an implementation function for deleting an outbound route for a Genesys Cloud Site
func deleteSiteOutboundRouteFn(ctx context.Context, p *siteOutboundRoutesProxy, siteId string, outboundRouteId string) (*platformclientv2.APIResponse, error) {
	return p.edgesApi.DeleteTelephonyProvidersEdgesSiteOutboundroute(siteId, outboundRouteId)
}
