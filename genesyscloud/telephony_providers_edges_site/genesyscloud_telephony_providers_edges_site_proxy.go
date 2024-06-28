package telephony_providers_edges_site

import (
	"context"
	"fmt"
	rc "terraform-provider-genesyscloud/genesyscloud/resource_cache"

	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

/*
The genesyscloud_telephony_providers_edges_site_proxy.go file contains the proxy structures and methods that interact
with the Genesys Cloud SDK. We use composition here for each function on the proxy so individual functions can be stubbed
out during testing.

Each proxy implementation:

1.  Should provide a private package level variable that holds a instance of a proxy class.
2.  A New... constructor function  to initialize the proxy object.  This constructor should only be used within
    the proxy.
3.  A get private constructor function that the classes in the package can be used to to retrieve
    the proxy.  This proxy should check to see if the package level proxy instance is nil and
    should initialize it, otherwise it should return the instance
4.  Type definitions for each function that will be used in the proxy.  We use composition here
    so that we can easily provide mocks for testing.
5.  A struct for the proxy that holds an attribute for each function type.
6.  Wrapper methods on each of the elements on the struct.
7.  Function implementations for each function type definition.

*/

// internalProxy holds a proxy instance that can be used throughout the package
var internalProxy *SiteProxy

// Type definitions for each func on our proxy so we can easily mock them out later
type getAllSitesFunc func(ctx context.Context, p *SiteProxy, managed bool) (*[]platformclientv2.Site, *platformclientv2.APIResponse, error)
type createSiteFunc func(ctx context.Context, p *SiteProxy, site *platformclientv2.Site) (*platformclientv2.Site, *platformclientv2.APIResponse, error)
type deleteSiteFunc func(ctx context.Context, p *SiteProxy, siteId string) (*platformclientv2.APIResponse, error)
type getSiteByIdFunc func(ctx context.Context, p *SiteProxy, siteId string) (site *platformclientv2.Site, resp *platformclientv2.APIResponse, err error)
type getSiteIdByNameFunc func(ctx context.Context, p *SiteProxy, siteName string) (siteId string, retryable bool, resp *platformclientv2.APIResponse, err error)
type updateSiteFunc func(ctx context.Context, p *SiteProxy, siteId string, site *platformclientv2.Site) (*platformclientv2.Site, *platformclientv2.APIResponse, error)

type createSiteOutboundRouteFunc func(ctx context.Context, p *SiteProxy, siteId string, outboundRoute *platformclientv2.Outboundroutebase) (*platformclientv2.Outboundroutebase, *platformclientv2.APIResponse, error)
type getSiteOutboundRoutesFunc func(ctx context.Context, p *SiteProxy, siteId string) (*[]platformclientv2.Outboundroutebase, *platformclientv2.APIResponse, error)
type updateSiteOutboundRouteFunc func(ctx context.Context, p *SiteProxy, siteId string, outboundRouteId string, outboundRoute *platformclientv2.Outboundroutebase) (*platformclientv2.Outboundroutebase, *platformclientv2.APIResponse, error)
type deleteSiteOutboundRouteFunc func(ctx context.Context, p *SiteProxy, siteId string, outboundRouteId string) (*platformclientv2.APIResponse, error)

type getSiteNumberPlansFunc func(ctx context.Context, p *SiteProxy, siteId string) (*[]platformclientv2.Numberplan, *platformclientv2.APIResponse, error)
type updateSiteNumberPlansFunc func(ctx context.Context, p *SiteProxy, siteId string, numberPlans *[]platformclientv2.Numberplan) (*[]platformclientv2.Numberplan, *platformclientv2.APIResponse, error)

type getLocationFunc func(ctx context.Context, p *SiteProxy, locationId string) (*platformclientv2.Locationdefinition, *platformclientv2.APIResponse, error)
type getTelephonyMediaregionsFunc func(ctx context.Context, p *SiteProxy) (*platformclientv2.Mediaregions, *platformclientv2.APIResponse, error)
type setDefaultSiteFunc func(ctx context.Context, p *SiteProxy, siteId string) (*platformclientv2.APIResponse, error)
type getDefaultSiteIdFunc func(ctx context.Context, p *SiteProxy) (siteId string, resp *platformclientv2.APIResponse, err error)

// SiteProxy contains all of the methods that call genesys cloud APIs.
type SiteProxy struct {
	clientConfig    *platformclientv2.Configuration
	edgesApi        *platformclientv2.TelephonyProvidersEdgeApi
	locationsApi    *platformclientv2.LocationsApi
	telephonyApi    *platformclientv2.TelephonyApi
	organizationApi *platformclientv2.OrganizationApi

	getAllSitesAttr     getAllSitesFunc
	createSiteAttr      createSiteFunc
	deleteSiteAttr      deleteSiteFunc
	getSiteByIdAttr     getSiteByIdFunc
	getSiteIdByNameAttr getSiteIdByNameFunc
	updateSiteAttr      updateSiteFunc

	createSiteOutboundRouteAttr createSiteOutboundRouteFunc
	getSiteOutboundRoutesAttr   getSiteOutboundRoutesFunc
	updateSiteOutboundRouteAttr updateSiteOutboundRouteFunc
	deleteSiteOutboundRouteAttr deleteSiteOutboundRouteFunc

	getSiteNumberPlansAttr    getSiteNumberPlansFunc
	updateSiteNumberPlansAttr updateSiteNumberPlansFunc

	getLocationAttr              getLocationFunc
	getTelephonyMediaregionsAttr getTelephonyMediaregionsFunc
	setDefaultSiteAttr           setDefaultSiteFunc
	getDefaultSiteIdAttr         getDefaultSiteIdFunc

	unmanagedSiteCache rc.CacheInterface[platformclientv2.Site]
	managedSiteCache   rc.CacheInterface[platformclientv2.Site]
}

// newSiteProxy initializes the Site proxy with all the data needed to communicate with Genesys Cloud
func newSiteProxy(clientConfig *platformclientv2.Configuration) *SiteProxy {
	edgesApi := platformclientv2.NewTelephonyProvidersEdgeApiWithConfig(clientConfig)
	locationsApi := platformclientv2.NewLocationsApiWithConfig(clientConfig)
	telephonyApi := platformclientv2.NewTelephonyApiWithConfig(clientConfig)
	organizationApi := platformclientv2.NewOrganizationApiWithConfig(clientConfig)

	unmanagedSiteCache := rc.NewResourceCache[platformclientv2.Site]()
	managedSiteCache := rc.NewResourceCache[platformclientv2.Site]()

	return &SiteProxy{
		clientConfig:    clientConfig,
		edgesApi:        edgesApi,
		locationsApi:    locationsApi,
		telephonyApi:    telephonyApi,
		organizationApi: organizationApi,

		getAllSitesAttr:     getAllSitesFn,
		createSiteAttr:      createSiteFn,
		deleteSiteAttr:      deleteSiteFn,
		getSiteByIdAttr:     getSiteByIdFn,
		getSiteIdByNameAttr: getSiteIdByNameFn,
		updateSiteAttr:      updateSiteFn,

		createSiteOutboundRouteAttr: createSiteOutboundRouteFn,
		getSiteOutboundRoutesAttr:   getSiteOutboundRoutesFn,
		updateSiteOutboundRouteAttr: updateSiteOutboundRouteFn,
		deleteSiteOutboundRouteAttr: deleteSiteOutboundRouteFn,

		getSiteNumberPlansAttr:    getSiteNumberPlansFn,
		updateSiteNumberPlansAttr: updateSiteNumberPlansFn,

		getLocationAttr:              getLocationFn,
		getTelephonyMediaregionsAttr: getTelephonyMediaregionsFn,
		setDefaultSiteAttr:           setDefaultSiteFn,
		getDefaultSiteIdAttr:         getDefaultSiteIdFn,

		unmanagedSiteCache: unmanagedSiteCache,
		managedSiteCache:   managedSiteCache,
	}
}

// GetSiteProxy acts as a singleton for the internalProxy.  It also ensures
// that we can still proxy our tests by directly setting internalProxy package variable
func GetSiteProxy(clientConfig *platformclientv2.Configuration) *SiteProxy {
	if internalProxy == nil {
		internalProxy = newSiteProxy(clientConfig)
	}
	return internalProxy
}

// GetAllSites retrieves all managed Genesys Cloud Sites
func (p *SiteProxy) GetAllSites(ctx context.Context, managed bool) (*[]platformclientv2.Site, *platformclientv2.APIResponse, error) {
	return p.getAllSitesAttr(ctx, p, managed)
}

// createSiteFunc creates a Genesys Cloud Site
func (p *SiteProxy) createSite(ctx context.Context, site *platformclientv2.Site) (*platformclientv2.Site, *platformclientv2.APIResponse, error) {
	return p.createSiteAttr(ctx, p, site)
}

// deleteSiteFunc deletes a Genesys Cloud Site by ID
func (p *SiteProxy) deleteSite(ctx context.Context, siteId string) (*platformclientv2.APIResponse, error) {
	return p.deleteSiteAttr(ctx, p, siteId)
}

// getSiteByIdFunc returns a single Genesys Cloud Site by Id
func (p *SiteProxy) getSiteById(ctx context.Context, siteId string) (site *platformclientv2.Site, resp *platformclientv2.APIResponse, err error) {
	return p.getSiteByIdAttr(ctx, p, siteId)
}

// getSiteIdByNameFunc returns a single Genesys Cloud Site by Name
func (p *SiteProxy) getSiteIdByName(ctx context.Context, siteName string) (siteId string, retryable bool, resp *platformclientv2.APIResponse, err error) {
	return p.getSiteIdByNameAttr(ctx, p, siteName)
}

// updateSiteFunc updates a Genesys Cloud Site
func (p *SiteProxy) updateSite(ctx context.Context, siteId string, site *platformclientv2.Site) (*platformclientv2.Site, *platformclientv2.APIResponse, error) {
	return p.updateSiteAttr(ctx, p, siteId, site)
}

// createSiteOutboundRouteFunc creates an Outbound Route for a Genesys Cloud Site
func (p *SiteProxy) createSiteOutboundRoute(ctx context.Context, siteId string, outboundRoute *platformclientv2.Outboundroutebase) (*platformclientv2.Outboundroutebase, *platformclientv2.APIResponse, error) {
	return p.createSiteOutboundRouteAttr(ctx, p, siteId, outboundRoute)
}

// getSiteByIdFunc returns a single Outbound Route by Id
func (p *SiteProxy) getSiteOutboundRoutes(ctx context.Context, siteId string) (*[]platformclientv2.Outboundroutebase, *platformclientv2.APIResponse, error) {
	return p.getSiteOutboundRoutesAttr(ctx, p, siteId)
}

// updateSiteFunc updates a Genesys Cloud Outbound Route for a Genesys Cloud Site
func (p *SiteProxy) updateSiteOutboundRoute(ctx context.Context, siteId string, outboundRouteId string, outboundRoute *platformclientv2.Outboundroutebase) (*platformclientv2.Outboundroutebase, *platformclientv2.APIResponse, error) {
	return p.updateSiteOutboundRouteAttr(ctx, p, siteId, outboundRouteId, outboundRoute)
}

// deleteSiteFunc deletes a Genesys Cloud Outbound Route by Id for a Genesys Cloud Site
func (p *SiteProxy) deleteSiteOutboundRoute(ctx context.Context, siteId string, outboundRouteId string) (*platformclientv2.APIResponse, error) {
	return p.deleteSiteOutboundRouteAttr(ctx, p, siteId, outboundRouteId)
}

// getSiteNumberPlansFunc retrieves all Number Plans of a Genesys Cloud Sites
func (p *SiteProxy) getSiteNumberPlans(ctx context.Context, siteId string) (*[]platformclientv2.Numberplan, *platformclientv2.APIResponse, error) {
	return p.getSiteNumberPlansAttr(ctx, p, siteId)
}

// updateSiteNumberPlansFunc updates the Number Plans for a Genesys Cloud Site
func (p *SiteProxy) updateSiteNumberPlans(ctx context.Context, siteId string, numberPlans *[]platformclientv2.Numberplan) (*[]platformclientv2.Numberplan, *platformclientv2.APIResponse, error) {
	return p.updateSiteNumberPlansAttr(ctx, p, siteId, numberPlans)
}

// getLocation retrieves a Genesys Cloud Location by Id
func (p *SiteProxy) getLocation(ctx context.Context, locationId string) (*platformclientv2.Locationdefinition, *platformclientv2.APIResponse, error) {
	return p.getLocationAttr(ctx, p, locationId)
}

// getTelephonyMediaregions retrieves the Genesys Cloud media regions
func (p *SiteProxy) getTelephonyMediaregions(ctx context.Context) (*platformclientv2.Mediaregions, *platformclientv2.APIResponse, error) {
	return p.getTelephonyMediaregionsAttr(ctx, p)
}

// setDefaultSite sets a Genesys Cloud Site as the default site for the org
func (p *SiteProxy) setDefaultSite(ctx context.Context, siteId string) (*platformclientv2.APIResponse, error) {
	return p.setDefaultSiteAttr(ctx, p, siteId)
}

// getDefaultSiteId gets the default Site for the Genesys Cloud org
func (p *SiteProxy) getDefaultSiteId(ctx context.Context) (siteId string, resp *platformclientv2.APIResponse, err error) {
	return p.getDefaultSiteIdAttr(ctx, p)
}

// getAllManagedSitesFn is an implementation function for retrieving all Genesys Cloud Outbound managed Sites
func getAllSitesFn(ctx context.Context, p *SiteProxy, managed bool) (*[]platformclientv2.Site, *platformclientv2.APIResponse, error) {
	var allSites []platformclientv2.Site
	var siteCache rc.CacheInterface[platformclientv2.Site]

	switch {
	case managed:
		siteCache = p.managedSiteCache
		break
	case !managed:
		siteCache = p.unmanagedSiteCache
		break
	}

	const pageSize = 100
	sites, resp, err := p.edgesApi.GetTelephonyProvidersEdgesSites(pageSize, 1, "", "", "", "", managed, nil)
	if err != nil {
		return nil, resp, err
	}

	// Get only sites that are not 'deleted'
	for _, site := range *sites.Entities {
		if site.State != nil && *site.State != "deleted" {
			allSites = append(allSites, site)
		}
	}

	// Check if the site cache is populated with all the data, if it is, return that instead
	// If the size of the cache is the same as the total number of queues, the cache is up-to-date
	if rc.GetCacheSize(siteCache) == *sites.Total && rc.GetCacheSize(siteCache) != 0 {
		return rc.GetCache(siteCache), nil, nil
	} else if rc.GetCacheSize(siteCache) != *sites.Total && rc.GetCacheSize(siteCache) != 0 {
		// The cache is populated but not with the right data, clear the cache so it can be re populated
		siteCache = rc.NewResourceCache[platformclientv2.Site]()
	}

	for pageNum := 2; pageNum <= *sites.PageCount; pageNum++ {
		sites, resp, err := p.edgesApi.GetTelephonyProvidersEdgesSites(pageSize, pageNum, "", "", "", "", managed, nil)
		if err != nil {
			return nil, resp, err
		}
		if sites.Entities == nil || len(*sites.Entities) == 0 {
			break
		}

		// Get only sites that are not 'deleted'
		for _, site := range *sites.Entities {
			if site.State != nil && *site.State != "deleted" {
				allSites = append(allSites, site)
			}
		}
	}

	// Populate the site cache (unmanaged site cache or managed site cache)
	for _, site := range allSites {
		rc.SetCache(siteCache, *site.Id, site)
	}

	return &allSites, resp, nil
}

// createSiteFn is an implementation function for creating a Genesys Cloud Site
func createSiteFn(ctx context.Context, p *SiteProxy, siteReq *platformclientv2.Site) (*platformclientv2.Site, *platformclientv2.APIResponse, error) {
	site, resp, err := p.edgesApi.PostTelephonyProvidersEdgesSites(*siteReq)
	if err != nil {
		return nil, resp, err
	}
	return site, resp, nil
}

// deleteSiteFn is an implementation function for deleting a Genesys Cloud Site
func deleteSiteFn(ctx context.Context, p *SiteProxy, siteId string) (*platformclientv2.APIResponse, error) {
	resp, err := p.edgesApi.DeleteTelephonyProvidersEdgesSite(siteId)
	if err != nil {
		return resp, err
	}

	return resp, nil
}

// getSiteByIdFn is an implementation function for retrieving a Genesys Cloud Site by id
func getSiteByIdFn(ctx context.Context, p *SiteProxy, siteId string) (*platformclientv2.Site, *platformclientv2.APIResponse, error) {
	var site *platformclientv2.Site

	// Query managed site cache for the site
	site = rc.GetCacheItem(p.managedSiteCache, siteId)
	if site != nil {
		return site, nil, nil
	} else {
		// Query unmanaged sites cache if not in managed site cache
		site = rc.GetCacheItem(p.unmanagedSiteCache, siteId)
		if site != nil {
			return site, nil, nil
		}
	}

	site, resp, err := p.edgesApi.GetTelephonyProvidersEdgesSite(siteId)
	if err != nil {
		return nil, resp, err
	}

	return site, resp, nil
}

// getSiteIdByNameFn is an implementation function for retrieving a Genesys Cloud Site by name
func getSiteIdByNameFn(ctx context.Context, p *SiteProxy, siteName string) (string, bool, *platformclientv2.APIResponse, error) {
	managed, resp, err := getAllSitesFn(ctx, p, true)
	if err != nil {
		return "", false, resp, err
	}

	if managed != nil {
		for _, site := range *managed {
			if (site.Name != nil && *site.Name == siteName) && (site.State != nil && *site.State != "deleted") {
				return *site.Id, false, resp, nil
			}
		}
	}

	unmanaged, resp, err := getAllSitesFn(ctx, p, false)
	if err != nil {
		return "", false, resp, err
	}

	if unmanaged != nil {
		for _, site := range *unmanaged {
			if (site.Name != nil && *site.Name == siteName) && (site.State != nil && *site.State != "deleted") {
				return *site.Id, false, resp, nil
			}
		}
	}
	return "", true, resp, fmt.Errorf("no sites found with name %s", siteName)
}

// updateSiteFn is an implementation function for updating a Genesys Cloud Site
func updateSiteFn(ctx context.Context, p *SiteProxy, siteId string, site *platformclientv2.Site) (*platformclientv2.Site, *platformclientv2.APIResponse, error) {
	updatedSite, resp, err := p.edgesApi.PutTelephonyProvidersEdgesSite(siteId, *site)
	if err != nil {
		return nil, resp, err
	}

	return updatedSite, resp, nil
}

// createSiteOutboundRouteFn is an implementation function for creating an outbound route for a Genesys Cloud Site
func createSiteOutboundRouteFn(ctx context.Context, p *SiteProxy, siteId string, outboundRoute *platformclientv2.Outboundroutebase) (*platformclientv2.Outboundroutebase, *platformclientv2.APIResponse, error) {
	obr, resp, err := p.edgesApi.PostTelephonyProvidersEdgesSiteOutboundroutes(siteId, *outboundRoute)
	if err != nil {
		return nil, resp, err
	}

	return obr, resp, nil
}

// getSiteOutboundRoutesFn is an implementation function for getting an outbound route for a Genesys Cloud Site
func getSiteOutboundRoutesFn(ctx context.Context, p *SiteProxy, siteId string) (*[]platformclientv2.Outboundroutebase, *platformclientv2.APIResponse, error) {
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
func updateSiteOutboundRouteFn(ctx context.Context, p *SiteProxy, siteId string, outboundRouteId string, outboundRoute *platformclientv2.Outboundroutebase) (*platformclientv2.Outboundroutebase, *platformclientv2.APIResponse, error) {
	obrs, resp, err := p.edgesApi.PutTelephonyProvidersEdgesSiteOutboundroute(siteId, outboundRouteId, *outboundRoute)
	if err != nil {
		return nil, resp, err
	}

	return obrs, resp, nil
}

// deleteSiteOutboundRouteFn is an implementation function for deleting an outbound route for a Genesys Cloud Site
func deleteSiteOutboundRouteFn(ctx context.Context, p *SiteProxy, siteId string, outboundRouteId string) (*platformclientv2.APIResponse, error) {
	resp, err := p.edgesApi.DeleteTelephonyProvidersEdgesSiteOutboundroute(siteId, outboundRouteId)
	if err != nil {
		return resp, err
	}

	return resp, nil
}

// getSiteNumberPlansFn is an implementation function for retrieving number plans of a Genesys Cloud Site
func getSiteNumberPlansFn(ctx context.Context, p *SiteProxy, siteId string) (*[]platformclientv2.Numberplan, *platformclientv2.APIResponse, error) {
	numberPlans, resp, err := p.edgesApi.GetTelephonyProvidersEdgesSiteNumberplans(siteId)
	if err != nil {
		return nil, resp, err
	}

	return &numberPlans, resp, nil
}

// updateSiteNumberPlansFn is an implementation function for updating number plans of a Genesys Cloud Site
func updateSiteNumberPlansFn(ctx context.Context, p *SiteProxy, siteId string, numberPlansUpdate *[]platformclientv2.Numberplan) (*[]platformclientv2.Numberplan, *platformclientv2.APIResponse, error) {
	numberPlans, resp, err := p.edgesApi.PutTelephonyProvidersEdgesSiteNumberplans(siteId, *numberPlansUpdate)
	if err != nil {
		return nil, resp, err
	}

	return &numberPlans, resp, nil
}

// getLocationFn is an implementation function for retrieving a Genesys Cloud Location
func getLocationFn(ctx context.Context, p *SiteProxy, locationId string) (*platformclientv2.Locationdefinition, *platformclientv2.APIResponse, error) {
	location, resp, err := p.locationsApi.GetLocation(locationId, nil)
	if err != nil {
		return nil, resp, err
	}
	if location.EmergencyNumber == nil {
		return nil, resp, fmt.Errorf("location with id %v does not have an emergency number", locationId)
	}

	return location, resp, nil
}

// getTelephonyMediaregionsFn is an implementation function for retrieving a Genesys Cloud Media Regions
func getTelephonyMediaregionsFn(ctx context.Context, p *SiteProxy) (*platformclientv2.Mediaregions, *platformclientv2.APIResponse, error) {
	telephonyRegions, resp, err := p.telephonyApi.GetTelephonyMediaregions()
	if err != nil {
		return nil, resp, err
	}

	return telephonyRegions, resp, nil
}

// setDefaultSiteFn is an implementation function for setting the default Site of a Genesys Cloud org
func setDefaultSiteFn(ctx context.Context, p *SiteProxy, siteId string) (*platformclientv2.APIResponse, error) {
	org, resp, err := p.organizationApi.GetOrganizationsMe()
	if err != nil {
		return resp, err
	}

	// Update org details
	*org.DefaultSiteId = siteId

	_, resp, err = p.organizationApi.PutOrganizationsMe(*org)
	if err != nil {
		return resp, fmt.Errorf("error on setting default site. Make sure only one resource has the `set_as_default_site` set to true. %v", err)
	}

	return resp, nil
}

// getDefaultSiteIdFn is an implementation function for getting the default Site of a Genesys Cloud org
func getDefaultSiteIdFn(ctx context.Context, p *SiteProxy) (string, *platformclientv2.APIResponse, error) {
	org, resp, err := p.organizationApi.GetOrganizationsMe()
	if err != nil {
		return "", resp, err
	}

	return *org.DefaultSiteId, resp, nil
}
