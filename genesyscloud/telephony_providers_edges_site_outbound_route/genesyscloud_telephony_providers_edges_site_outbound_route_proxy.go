package telephony_providers_edges_site_outbound_route

import (
	"context"
	"fmt"
	rc "terraform-provider-genesyscloud/genesyscloud/resource_cache"
	telephonyProvidersEdgesSite "terraform-provider-genesyscloud/genesyscloud/telephony_providers_edges_site"

	"github.com/mypurecloud/platform-client-sdk-go/v150/platformclientv2"
)

/*
The genesyscloud_telephony_providers_edges_site_outbound_route_proxy.go file contains the proxy structures and methods that interact
with the Genesys Cloud SDK. We use composition here for each function on the proxy so individual functions can be stubbed
out during testing.
*/

// internalProxy holds a proxy instance that can be used throughout the package
var internalProxy *siteOutboundRouteProxy

// Type definitions for each func on our proxy so we can easily mock them out later
type getAllSiteOutboundRoutesFunc func(ctx context.Context, p *siteOutboundRouteProxy, siteId string) (*[]platformclientv2.Outboundroutebase, *platformclientv2.APIResponse, error)
type getSiteByIdFunc func(ctx context.Context, p *siteOutboundRouteProxy, siteId string) (*platformclientv2.Site, *platformclientv2.APIResponse, error)
type createSiteOutboundRouteFunc func(ctx context.Context, p *siteOutboundRouteProxy, siteId string, outboundRoute *platformclientv2.Outboundroutebase) (*platformclientv2.Outboundroutebase, *platformclientv2.APIResponse, error)
type getSiteOutboundRouteByIdFunc func(ctx context.Context, p *siteOutboundRouteProxy, siteId string, outboundRoute string) (*platformclientv2.Outboundroutebase, *platformclientv2.APIResponse, error)
type getSiteOutboundRouteByNameFunc func(ctx context.Context, p *siteOutboundRouteProxy, siteIdOrEmpty string, outboundRouteName string) (siteId string, outboundRouteId string, retryable bool, response *platformclientv2.APIResponse, err error)
type updateSiteOutboundRouteFunc func(ctx context.Context, p *siteOutboundRouteProxy, siteId string, outboundRouteId string, outboundRoute *platformclientv2.Outboundroutebase) (*platformclientv2.Outboundroutebase, *platformclientv2.APIResponse, error)
type deleteSiteOutboundRouteFunc func(ctx context.Context, p *siteOutboundRouteProxy, siteId string, outboundRouteId string) (*platformclientv2.APIResponse, error)

// siteOutboundRouteProxy contains all the methods that call genesys cloud APIs.
type siteOutboundRouteProxy struct {
	clientConfig *platformclientv2.Configuration
	edgesApi     *platformclientv2.TelephonyProvidersEdgeApi

	getAllSiteOutboundRoutesAttr   getAllSiteOutboundRoutesFunc
	getSiteByIdAttr                getSiteByIdFunc
	createSiteOutboundRouteAttr    createSiteOutboundRouteFunc
	getSiteOutboundRouteByIdAttr   getSiteOutboundRouteByIdFunc
	getSiteOutboundRouteByNameAttr getSiteOutboundRouteByNameFunc
	updateSiteOutboundRouteAttr    updateSiteOutboundRouteFunc
	deleteSiteOutboundRouteAttr    deleteSiteOutboundRouteFunc
	siteOutboundRouteCache         rc.CacheInterface[platformclientv2.Outboundroutebase]
	siteProxy                      *telephonyProvidersEdgesSite.SiteProxy
}

// newSiteOutboundRouteProxy initializes the Site proxy with all the data needed to communicate with Genesys Cloud
func newSiteOutboundRouteProxy(clientConfig *platformclientv2.Configuration) *siteOutboundRouteProxy {
	edgesApi := platformclientv2.NewTelephonyProvidersEdgeApiWithConfig(clientConfig)
	siteProxy := telephonyProvidersEdgesSite.GetSiteProxy(clientConfig)
	siteOutboundRouteCache := rc.NewResourceCache[platformclientv2.Outboundroutebase]()

	return &siteOutboundRouteProxy{
		clientConfig: clientConfig,
		edgesApi:     edgesApi,

		getAllSiteOutboundRoutesAttr:   getAllSiteOutboundRoutesFn,
		getSiteByIdAttr:                getSiteFn,
		createSiteOutboundRouteAttr:    createSiteOutboundRouteFn,
		getSiteOutboundRouteByIdAttr:   getSiteOutboundRouteByIdFn,
		getSiteOutboundRouteByNameAttr: getSiteOutboundRouteByNameFn,
		updateSiteOutboundRouteAttr:    updateSiteOutboundRouteFn,
		deleteSiteOutboundRouteAttr:    deleteSiteOutboundRouteFn,
		siteOutboundRouteCache:         siteOutboundRouteCache,
		siteProxy:                      siteProxy,
	}
}

// getSiteOutboundRouteProxy acts as a singleton for the internalProxy.  It also ensures
// that we can still proxy our tests by directly setting internalProxy package variable
func getSiteOutboundRouteProxy(clientConfig *platformclientv2.Configuration) *siteOutboundRouteProxy {
	if internalProxy == nil {
		internalProxy = newSiteOutboundRouteProxy(clientConfig)
	}
	return internalProxy
}

func (p *siteOutboundRouteProxy) getAllSiteOutboundRoutes(ctx context.Context, siteId string) (*[]platformclientv2.Outboundroutebase, *platformclientv2.APIResponse, error) {
	return p.getAllSiteOutboundRoutesAttr(ctx, p, siteId)
}

func (p *siteOutboundRouteProxy) getSite(ctx context.Context, id string) (*platformclientv2.Site, *platformclientv2.APIResponse, error) {
	return p.getSiteByIdAttr(ctx, p, id)
}

// createSiteOutboundRouteFunc creates an Outbound Route for a Genesys Cloud Site
func (p *siteOutboundRouteProxy) createSiteOutboundRoute(ctx context.Context, siteId string, outboundRoute *platformclientv2.Outboundroutebase) (*platformclientv2.Outboundroutebase, *platformclientv2.APIResponse, error) {
	return p.createSiteOutboundRouteAttr(ctx, p, siteId, outboundRoute)
}

// getSiteByIdFunc returns a single Outbound Route by Id
func (p *siteOutboundRouteProxy) getSiteOutboundRouteById(ctx context.Context, siteId string, outboundRoute string) (*platformclientv2.Outboundroutebase, *platformclientv2.APIResponse, error) {
	return p.getSiteOutboundRouteByIdAttr(ctx, p, siteId, outboundRoute)
}

// getSiteByNameFunc returns the outbound route id
func (p *siteOutboundRouteProxy) getSiteOutboundRouteByName(ctx context.Context, siteIdOrEmpty string, outboundRouteName string) (siteId string, outboundRouteId string, retryable bool, response *platformclientv2.APIResponse, err error) {
	return p.getSiteOutboundRouteByNameAttr(ctx, p, siteIdOrEmpty, outboundRouteName)
}

// updateSiteFunc updates a Genesys Cloud Outbound Route for a Genesys Cloud Site
func (p *siteOutboundRouteProxy) updateSiteOutboundRoute(ctx context.Context, siteId string, outboundRouteId string, outboundRoute *platformclientv2.Outboundroutebase) (*platformclientv2.Outboundroutebase, *platformclientv2.APIResponse, error) {
	return p.updateSiteOutboundRouteAttr(ctx, p, siteId, outboundRouteId, outboundRoute)
}

// deleteSiteFunc deletes a Genesys Cloud Outbound Route by Id for a Genesys Cloud Site
func (p *siteOutboundRouteProxy) deleteSiteOutboundRoute(ctx context.Context, siteId string, outboundRouteId string) (*platformclientv2.APIResponse, error) {
	return p.deleteSiteOutboundRouteAttr(ctx, p, siteId, outboundRouteId)
}

func getSiteFn(ctx context.Context, p *siteOutboundRouteProxy, id string) (*platformclientv2.Site, *platformclientv2.APIResponse, error) {
	return p.edgesApi.GetTelephonyProvidersEdgesSite(id)
}

func createSiteOutboundRouteFn(ctx context.Context, p *siteOutboundRouteProxy, siteId string, route *platformclientv2.Outboundroutebase) (*platformclientv2.Outboundroutebase, *platformclientv2.APIResponse, error) {
	return p.edgesApi.PostTelephonyProvidersEdgesSiteOutboundroutes(siteId, *route)
}

func getAllSiteOutboundRoutesFn(ctx context.Context, p *siteOutboundRouteProxy, siteId string) (*[]platformclientv2.Outboundroutebase, *platformclientv2.APIResponse, error) {
	var allOutboundRoutes []platformclientv2.Outboundroutebase

	const pageSize = 100
	outboundRoutes, resp, err := p.edgesApi.GetTelephonyProvidersEdgesSiteOutboundroutes(siteId, pageSize, 1, "", "", "")
	if err != nil {
		return nil, resp, err
	}

	allOutboundRoutes = append(allOutboundRoutes, *outboundRoutes.Entities...)

	// Check if the site cache is populated with all the data, if it is, return that instead
	// If the size of the cache is the same as the total number of sites, the cache is up-to-date
	if rc.GetCacheSize(p.siteOutboundRouteCache) == *outboundRoutes.Total && rc.GetCacheSize(p.siteOutboundRouteCache) != 0 {
		return rc.GetCache(p.siteOutboundRouteCache), nil, nil
	} else if rc.GetCacheSize(p.siteOutboundRouteCache) != *outboundRoutes.Total && rc.GetCacheSize(p.siteOutboundRouteCache) != 0 {
		// The cache is populated but not with the right data, clear the cache so it can be re populated
		p.siteOutboundRouteCache = rc.NewResourceCache[platformclientv2.Outboundroutebase]()
	}

	for pageNum := 2; pageNum <= *outboundRoutes.PageCount; pageNum++ {
		outboundRoutes, resp, err := p.edgesApi.GetTelephonyProvidersEdgesSiteOutboundroutes(siteId, pageSize, pageNum, "", "", "")
		if err != nil {
			return nil, resp, err
		}
		if outboundRoutes.Entities == nil || len(*outboundRoutes.Entities) == 0 {
			break
		}

		allOutboundRoutes = append(allOutboundRoutes, *outboundRoutes.Entities...)
	}

	// Populate the site cache
	for _, outboundRoute := range allOutboundRoutes {
		rc.SetCache(p.siteOutboundRouteCache, *outboundRoute.Id, outboundRoute)
	}

	return &allOutboundRoutes, resp, nil

}

// getSiteOutboundRouteByIdFn is an implementation function for getting an outbound route for a Genesys Cloud Site
func getSiteOutboundRouteByIdFn(ctx context.Context, p *siteOutboundRouteProxy, siteId string, outboundRouteId string) (*platformclientv2.Outboundroutebase, *platformclientv2.APIResponse, error) {
	// Check if site's outbound route exist in cache
	route := rc.GetCacheItem(p.siteOutboundRouteCache, outboundRouteId)
	if route != nil {
		return route, nil, nil
	}

	outboundRoute, resp, err := p.edgesApi.GetTelephonyProvidersEdgesSiteOutboundroute(siteId, outboundRouteId)
	if err != nil {
		return nil, resp, err
	}

	rc.SetCache(p.siteOutboundRouteCache, outboundRouteId, *outboundRoute)

	return outboundRoute, resp, nil
}

func getSiteOutboundRouteByNameFn(ctx context.Context, p *siteOutboundRouteProxy, siteIdOrEmpty string, outboundRouteName string) (siteId string, outboundRouteId string, retryable bool, resp *platformclientv2.APIResponse, err error) {
	var allSites []platformclientv2.Site
	unmanagedSites, resp, err := p.siteProxy.GetAllSites(ctx, false)
	if err != nil {
		return "", "", false, resp, err
	}
	allSites = append(allSites, *unmanagedSites...)

	managedSites, resp, err := p.siteProxy.GetAllSites(ctx, true)
	if err != nil {
		return "", "", false, resp, err
	}
	allSites = append(allSites, *managedSites...)
	for _, site := range allSites {
		outboundRoutes, resp, err := p.getAllSiteOutboundRoutes(ctx, *site.Id)
		if err != nil {
			return "", "", false, resp, err
		}
		if siteIdOrEmpty != "" && *site.Id != siteIdOrEmpty {
			continue
		}
		for _, outboundRoute := range *outboundRoutes {
			if (outboundRoute.Name != nil && *outboundRoute.Name == outboundRouteName) &&
				(outboundRoute.State != nil && *outboundRoute.State != "deleted") {
				return *site.Id, *outboundRoute.Id, false, resp, nil
			}
		}
	}

	return "", "", true, resp, fmt.Errorf("no outbound route found with name %s", outboundRouteName)
}

// updateSiteOutboundRouteFn is an implementation function for updating an outbound route for a Genesys Cloud Site
func updateSiteOutboundRouteFn(ctx context.Context, p *siteOutboundRouteProxy, siteId string, outboundRouteId string, outboundRoute *platformclientv2.Outboundroutebase) (*platformclientv2.Outboundroutebase, *platformclientv2.APIResponse, error) {
	return p.edgesApi.PutTelephonyProvidersEdgesSiteOutboundroute(siteId, outboundRouteId, *outboundRoute)
}

// deleteSiteOutboundRouteFn is an implementation function for deleting an outbound route for a Genesys Cloud Site
func deleteSiteOutboundRouteFn(ctx context.Context, p *siteOutboundRouteProxy, siteId string, outboundRouteId string) (*platformclientv2.APIResponse, error) {
	resp, err := p.edgesApi.DeleteTelephonyProvidersEdgesSiteOutboundroute(siteId, outboundRouteId)
	if err != nil {
		return resp, err
	}

	rc.DeleteCacheItem(p.siteOutboundRouteCache, outboundRouteId)
	return resp, nil
}
