package journey_views

import (
	"context"
	"fmt"
	rc "terraform-provider-genesyscloud/genesyscloud/resource_cache"

	"github.com/mypurecloud/platform-client-sdk-go/v143/platformclientv2"
)

var internalProxy *journeyViewsProxy

type GetAllJourneyViewsFunc func(ctx context.Context, p *journeyViewsProxy, name string) (*[]platformclientv2.Journeyview, *platformclientv2.APIResponse, error)
type getJourneyViewByNameFunc func(ctx context.Context, p *journeyViewsProxy, name string) (*platformclientv2.Journeyview, *platformclientv2.APIResponse, error)
type getJourneyViewByViewIdFunc func(ctx context.Context, p *journeyViewsProxy, viewId string) (*platformclientv2.Journeyview, *platformclientv2.APIResponse, error)
type createJourneyViewFunc func(ctx context.Context, p *journeyViewsProxy, journeyView *platformclientv2.Journeyview) (*platformclientv2.Journeyview, *platformclientv2.APIResponse, error)
type updateJourneyViewFunc func(ctx context.Context, p *journeyViewsProxy, viewId string, journeyView *platformclientv2.Journeyview) (*platformclientv2.Journeyview, *platformclientv2.APIResponse, error)
type deleteJourneyViewFunc func(ctx context.Context, p *journeyViewsProxy, viewId string) (*platformclientv2.APIResponse, error)

type journeyViewsProxy struct {
	clientConfig           *platformclientv2.Configuration
	journeyViewsApi        *platformclientv2.JourneyApi
	GetAllJourneyViewsAttr GetAllJourneyViewsFunc
	getJourneyViewAttr     getJourneyViewByViewIdFunc
	createJourneyViewAttr  createJourneyViewFunc
	updateJourneyViewAttr  updateJourneyViewFunc
	deleteJourneyViewAttr  deleteJourneyViewFunc
	journeyViewCache       rc.CacheInterface[platformclientv2.Journeyview]
}

func newJourneyViewsProxy(clientConfig *platformclientv2.Configuration) *journeyViewsProxy {
	api := platformclientv2.NewJourneyApiWithConfig(clientConfig)
	journeyViewCache := rc.NewResourceCache[platformclientv2.Journeyview]()
	return &journeyViewsProxy{
		clientConfig:          clientConfig,
		journeyViewsApi:       api,
		getJourneyViewAttr:    getJourneyViewByViewIdFn,
		createJourneyViewAttr: createJourneyViewFn,
		updateJourneyViewAttr: updateJourneyViewFn,
		deleteJourneyViewAttr: deleteJourneyViewFn,
		journeyViewCache:      journeyViewCache,
	}
}

func getJourneyViewProxy(clientConfig *platformclientv2.Configuration) *journeyViewsProxy {
	if internalProxy == nil {
		internalProxy = newJourneyViewsProxy(clientConfig)
	}
	return internalProxy
}

func (p *journeyViewsProxy) GetAllJourneyViews(ctx context.Context, name string) (*[]platformclientv2.Journeyview, *platformclientv2.APIResponse, error) {
	return p.GetAllJourneyViewsAttr(ctx, p, name)
}

func (p *journeyViewsProxy) getJourneyViewById(ctx context.Context, viewId string) (*platformclientv2.Journeyview, *platformclientv2.APIResponse, error) {
	return p.getJourneyViewAttr(ctx, p, viewId)
}

func (p *journeyViewsProxy) createJourneyView(ctx context.Context, journeyView *platformclientv2.Journeyview) (*platformclientv2.Journeyview, *platformclientv2.APIResponse, error) {
	return p.createJourneyViewAttr(ctx, p, journeyView)
}

func (p *journeyViewsProxy) updateJourneyView(ctx context.Context, viewId string, journeyView *platformclientv2.Journeyview) (*platformclientv2.Journeyview, *platformclientv2.APIResponse, error) {
	return p.updateJourneyViewAttr(ctx, p, viewId, journeyView)
}

func (p *journeyViewsProxy) deleteJourneyView(ctx context.Context, viewId string) (*platformclientv2.APIResponse, error) {
	return p.deleteJourneyViewAttr(ctx, p, viewId)
}

func getJourneyViewByViewIdFn(_ context.Context, p *journeyViewsProxy, viewId string) (*platformclientv2.Journeyview, *platformclientv2.APIResponse, error) {
	return p.journeyViewsApi.GetJourneyView(viewId)
}

func createJourneyViewFn(_ context.Context, p *journeyViewsProxy, journeyView *platformclientv2.Journeyview) (*platformclientv2.Journeyview, *platformclientv2.APIResponse, error) {
	return p.journeyViewsApi.PostJourneyViews(*journeyView)
}

func updateJourneyViewFn(_ context.Context, p *journeyViewsProxy, viewId string, journeyView *platformclientv2.Journeyview) (*platformclientv2.Journeyview, *platformclientv2.APIResponse, error) {
	return p.journeyViewsApi.PostJourneyViewVersions(viewId, *journeyView)
}

func deleteJourneyViewFn(_ context.Context, p *journeyViewsProxy, viewId string) (*platformclientv2.APIResponse, error) {
	return p.journeyViewsApi.DeleteJourneyView(viewId)
}

// GetAllJourneyViewsFn is the implementation for retrieving all journey views in Genesys Cloud
func GetAllJourneyViewsFn(ctx context.Context, p *journeyViewsProxy, name string) (*[]platformclientv2.Journeyview, *platformclientv2.APIResponse, error) {
	var allJourneys []platformclientv2.Journeyview
	const pageSize = 100

	journeys, resp, getErr := p.journeyViewsApi.GetJourneyViews(1, pageSize, name, "", "")
	if getErr != nil {
		return nil, resp, fmt.Errorf("failed to get first page of journeys: %v", getErr)
	}

	// Check if the journey view cache is populated with all the data, if it is, return that instead
	// If the size of the cache is the same as the total number of journeys, the cache is up-to-date
	if rc.GetCacheSize(p.journeyViewCache) == *journeys.Total && rc.GetCacheSize(p.journeyViewCache) != 0 {
		return rc.GetCache(p.journeyViewCache), nil, nil
	} else if rc.GetCacheSize(p.journeyViewCache) != *journeys.Total && rc.GetCacheSize(p.journeyViewCache) != 0 {
		// The cache is populated but not with the right data, clear the cache so it can be re populated
		p.journeyViewCache = rc.NewResourceCache[platformclientv2.Journeyview]()
	}

	if journeys.Entities == nil || len(*journeys.Entities) == 0 {
		return &allJourneys, resp, nil
	}

	allJourneys = append(allJourneys, *journeys.Entities...)

	for pageNum := 2; pageNum <= *journeys.PageCount; pageNum++ {
		journeys, resp, getErr := p.journeyViewsApi.GetJourneyViews(pageNum, pageSize, name, "", "")
		if getErr != nil {
			return nil, resp, fmt.Errorf("failed to get page of journeys: %v", getErr)
		}

		if journeys.Entities == nil || len(*journeys.Entities) == 0 {
			break
		}

		allJourneys = append(allJourneys, *journeys.Entities...)
	}

	for _, journeys := range allJourneys {
		rc.SetCache(p.journeyViewCache, *journeys.Id, journeys)
	}

	return &allJourneys, resp, nil
}
