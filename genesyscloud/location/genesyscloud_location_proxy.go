package location

import (
	"context"
	"fmt"
	rc "terraform-provider-genesyscloud/genesyscloud/resource_cache"

	"github.com/mypurecloud/platform-client-sdk-go/v149/platformclientv2"
)

var internalProxy *locationProxy

type getAllLocationFunc func(ctx context.Context, p *locationProxy) (*[]platformclientv2.Locationdefinition, *platformclientv2.APIResponse, error)
type createLocationFunc func(ctx context.Context, p *locationProxy, locationCreateDefinition *platformclientv2.Locationcreatedefinition) (*platformclientv2.Locationdefinition, *platformclientv2.APIResponse, error)
type getLocationByIdFunc func(ctx context.Context, p *locationProxy, id string, expand []string) (*platformclientv2.Locationdefinition, *platformclientv2.APIResponse, error)
type getLocationBySearchFunc func(ctx context.Context, p *locationProxy, body *platformclientv2.Locationsearchrequest) (*platformclientv2.Locationdefinition, *platformclientv2.APIResponse, error)
type updateLocationFunc func(ctx context.Context, p *locationProxy, id string, updateReq *platformclientv2.Locationupdatedefinition) (*platformclientv2.Locationdefinition, *platformclientv2.APIResponse, error)
type deleteLocationFunc func(ctx context.Context, p *locationProxy, id string) (*platformclientv2.APIResponse, error)

type locationProxy struct {
	clientConfig            *platformclientv2.Configuration
	locationsApi            *platformclientv2.LocationsApi
	createLocationAttr      createLocationFunc
	getAllLocationAttr      getAllLocationFunc
	getLocationByIdAttr     getLocationByIdFunc
	getLocationBySearchAttr getLocationBySearchFunc
	updateLocationAttr      updateLocationFunc
	deleteLocationAttr      deleteLocationFunc
	locationCache           rc.CacheInterface[platformclientv2.Locationdefinition]
}

// newLocationProxy initializes the location proxy with all of the data needed to communicate with Genesys Cloud
func newLocationProxy(clientConfig *platformclientv2.Configuration) *locationProxy {
	api := platformclientv2.NewLocationsApiWithConfig(clientConfig)
	locationCache := rc.NewResourceCache[platformclientv2.Locationdefinition]()
	return &locationProxy{
		clientConfig:            clientConfig,
		locationsApi:            api,
		createLocationAttr:      createLocationFn,
		getAllLocationAttr:      getAllLocationFn,
		getLocationByIdAttr:     getLocationByIdFn,
		getLocationBySearchAttr: getLocationBySearchFn,
		updateLocationAttr:      updateLocationFn,
		deleteLocationAttr:      deleteLocationFn,
		locationCache:           locationCache,
	}
}

// getLocationProxy acts as a singleton to for the internalProxy.  It also ensures
// that we can still proxy our tests by directly setting internalProxy package variable
func getLocationProxy(clientConfig *platformclientv2.Configuration) *locationProxy {
	if internalProxy == nil {
		internalProxy = newLocationProxy(clientConfig)
	}

	return internalProxy
}

func (p *locationProxy) getAllLocation(ctx context.Context) (*[]platformclientv2.Locationdefinition, *platformclientv2.APIResponse, error) {
	return p.getAllLocationAttr(ctx, p)
}

func (p *locationProxy) createLocation(ctx context.Context, location *platformclientv2.Locationcreatedefinition) (*platformclientv2.Locationdefinition, *platformclientv2.APIResponse, error) {
	return p.createLocationAttr(ctx, p, location)
}

func (p *locationProxy) getLocationById(ctx context.Context, id string, expand []string) (*platformclientv2.Locationdefinition, *platformclientv2.APIResponse, error) {
	return p.getLocationByIdAttr(ctx, p, id, expand)
}

func (p *locationProxy) getLocationBySearch(ctx context.Context, body *platformclientv2.Locationsearchrequest) (*platformclientv2.Locationdefinition, *platformclientv2.APIResponse, error) {
	return p.getLocationBySearchAttr(ctx, p, body)
}

func (p *locationProxy) updateLocation(ctx context.Context, id string, updateReq *platformclientv2.Locationupdatedefinition) (*platformclientv2.Locationdefinition, *platformclientv2.APIResponse, error) {
	return p.updateLocationAttr(ctx, p, id, updateReq)
}

func (p *locationProxy) deleteLocation(ctx context.Context, id string) (*platformclientv2.APIResponse, error) {
	return p.deleteLocationAttr(ctx, p, id)
}

func getAllLocationFn(ctx context.Context, p *locationProxy) (*[]platformclientv2.Locationdefinition, *platformclientv2.APIResponse, error) {
	var allLocations []platformclientv2.Locationdefinition
	const pageSize = 100

	locations, resp, err := p.locationsApi.GetLocations(pageSize, 1, nil, "")
	if err != nil {
		return nil, resp, fmt.Errorf("failed to get locations %s", err)
	}

	if locations.Entities == nil || len(*locations.Entities) == 0 {
		return &allLocations, resp, nil
	}
	allLocations = append(allLocations, *locations.Entities...)

	for pageNum := 2; pageNum <= *locations.PageCount; pageNum++ {
		locations, resp, err := p.locationsApi.GetLocations(pageSize, pageNum, nil, "")
		if err != nil {
			return nil, resp, fmt.Errorf("failed to get locations %s", err)
		}

		if locations.Entities == nil || len(*locations.Entities) == 0 {
			break
		}
		allLocations = append(allLocations, *locations.Entities...)
	}

	for _, location := range allLocations {
		rc.SetCache(p.locationCache, *location.Id, location)
	}

	return &allLocations, resp, nil
}

func createLocationFn(ctx context.Context, p *locationProxy, location *platformclientv2.Locationcreatedefinition) (*platformclientv2.Locationdefinition, *platformclientv2.APIResponse, error) {
	return p.locationsApi.PostLocations(*location)
}

func getLocationByIdFn(ctx context.Context, p *locationProxy, id string, expand []string) (*platformclientv2.Locationdefinition, *platformclientv2.APIResponse, error) {
	if location := rc.GetCacheItem(p.locationCache, id); location != nil {
		return location, nil, nil
	}
	return p.locationsApi.GetLocation(id, expand)
}

func getLocationBySearchFn(ctx context.Context, p *locationProxy, body *platformclientv2.Locationsearchrequest) (*platformclientv2.Locationdefinition, *platformclientv2.APIResponse, error) {
	locations, resp, err := p.locationsApi.PostLocationsSearch(*body)
	if err != nil {
		return nil, resp, fmt.Errorf("failed to get location %s", err)
	}

	if *locations.Total == 0 {
		return nil, resp, fmt.Errorf("404 - no locations found with search criteria error")
	}

	location := (*locations.Results)[0]

	return &location, resp, nil
}

func updateLocationFn(ctx context.Context, p *locationProxy, id string, updateReq *platformclientv2.Locationupdatedefinition) (*platformclientv2.Locationdefinition, *platformclientv2.APIResponse, error) {
	return p.locationsApi.PatchLocation(id, *updateReq)
}

func deleteLocationFn(ctx context.Context, p *locationProxy, id string) (*platformclientv2.APIResponse, error) {
	return p.locationsApi.DeleteLocation(id)
}
