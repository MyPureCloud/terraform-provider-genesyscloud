package journey_views

import (
	"context"
	"fmt"
	"log"
	"strconv"
	rc "terraform-provider-genesyscloud/genesyscloud/resource_cache"

	"github.com/mypurecloud/platform-client-sdk-go/v150/platformclientv2"
)

var internalProxy *journeyViewsProxy

type getAllJourneyViewsFunc func(ctx context.Context, p *journeyViewsProxy, name string) (*[]platformclientv2.Journeyview, *platformclientv2.APIResponse, error)
type getJourneyViewByNameFunc func(ctx context.Context, p *journeyViewsProxy, name string) (string, *platformclientv2.APIResponse, error, bool)
type getJourneyViewByViewIdFunc func(ctx context.Context, p *journeyViewsProxy, viewId string) (*platformclientv2.Journeyview, *platformclientv2.APIResponse, error)
type createJourneyViewFunc func(ctx context.Context, p *journeyViewsProxy, journeyView *platformclientv2.Journeyview) (*platformclientv2.Journeyview, *platformclientv2.APIResponse, error)
type updateJourneyViewFunc func(ctx context.Context, p *journeyViewsProxy, viewId string, versionId int, journeyView *platformclientv2.Journeyview) (*platformclientv2.Journeyview, *platformclientv2.APIResponse, error)
type deleteJourneyViewFunc func(ctx context.Context, p *journeyViewsProxy, viewId string) (*platformclientv2.APIResponse, error)

type journeyViewsProxy struct {
	clientConfig             *platformclientv2.Configuration
	journeyViewsApi          *platformclientv2.JourneyApi
	getAllJourneyViewsAttr   getAllJourneyViewsFunc
	getJourneyViewAttr       getJourneyViewByViewIdFunc
	getJourneyViewByNameAttr getJourneyViewByNameFunc
	createJourneyViewAttr    createJourneyViewFunc
	updateJourneyViewAttr    updateJourneyViewFunc
	deleteJourneyViewAttr    deleteJourneyViewFunc
	journeyViewCache         rc.CacheInterface[platformclientv2.Journeyview]
}

func newJourneyViewsProxy(clientConfig *platformclientv2.Configuration) *journeyViewsProxy {
	api := platformclientv2.NewJourneyApiWithConfig(clientConfig)
	journeyViewCache := rc.NewResourceCache[platformclientv2.Journeyview]()
	return &journeyViewsProxy{
		clientConfig:             clientConfig,
		journeyViewsApi:          api,
		getAllJourneyViewsAttr:   getAllJourneyViewsFn,
		getJourneyViewAttr:       getJourneyViewByViewIdFn,
		getJourneyViewByNameAttr: getJourneyViewByNameFn,
		createJourneyViewAttr:    createJourneyViewFn,
		updateJourneyViewAttr:    updateJourneyViewFn,
		deleteJourneyViewAttr:    deleteJourneyViewFn,
		journeyViewCache:         journeyViewCache,
	}
}

func getJourneyViewProxy(clientConfig *platformclientv2.Configuration) *journeyViewsProxy {
	if internalProxy == nil {
		internalProxy = newJourneyViewsProxy(clientConfig)
	}
	return internalProxy
}

func (p *journeyViewsProxy) getAllJourneyViews(ctx context.Context, name string) (*[]platformclientv2.Journeyview, *platformclientv2.APIResponse, error) {
	return p.getAllJourneyViewsAttr(ctx, p, name)
}

func (p *journeyViewsProxy) getJourneyViewById(ctx context.Context, viewId string) (*platformclientv2.Journeyview, *platformclientv2.APIResponse, error) {
	return p.getJourneyViewAttr(ctx, p, viewId)
}

func (p *journeyViewsProxy) getJourneyViewByName(ctx context.Context, viewName string) (string, *platformclientv2.APIResponse, error, bool) {
	return p.getJourneyViewByNameAttr(ctx, p, viewName)
}

func (p *journeyViewsProxy) createJourneyView(ctx context.Context, journeyView *platformclientv2.Journeyview) (*platformclientv2.Journeyview, *platformclientv2.APIResponse, error) {
	return p.createJourneyViewAttr(ctx, p, journeyView)
}

func (p *journeyViewsProxy) updateJourneyView(ctx context.Context, viewId string, versionId int, journeyView *platformclientv2.Journeyview) (*platformclientv2.Journeyview, *platformclientv2.APIResponse, error) {
	return p.updateJourneyViewAttr(ctx, p, viewId, versionId, journeyView)
}

func (p *journeyViewsProxy) deleteJourneyView(ctx context.Context, viewId string) (*platformclientv2.APIResponse, error) {
	resp, err := p.deleteJourneyViewAttr(ctx, p, viewId)
	if err != nil {
		return resp, err
	}
	rc.DeleteCacheItem(p.journeyViewCache, viewId)
	return resp, nil
}

func getJourneyViewByViewIdFn(_ context.Context, p *journeyViewsProxy, viewId string) (*platformclientv2.Journeyview, *platformclientv2.APIResponse, error) {
	// Check the cache first
	journeyView := rc.GetCacheItem(p.journeyViewCache, viewId)
	if journeyView != nil {
		return journeyView, nil, nil
	}
	return p.journeyViewsApi.GetJourneyView(viewId)
}

func getJourneyViewByNameFn(ctx context.Context, p *journeyViewsProxy, viewName string) (string, *platformclientv2.APIResponse, error, bool) {
	journeys, resp, err := p.getAllJourneyViews(ctx, viewName)
	if err != nil {
		return "", resp, err, false
	}

	if journeys == nil || len(*journeys) == 0 {
		return "", resp, fmt.Errorf("no journey view found with name %s", viewName), true
	}

	for _, journey := range *journeys {
		if *journey.Name == viewName {
			log.Printf("Retrieved the journey view id %s by name %s", *journey.Id, viewName)
			return *journey.Id, resp, nil, false
		}
	}
	return "", resp, fmt.Errorf("unable to find journey view with name %s", viewName), true
}

func createJourneyViewFn(_ context.Context, p *journeyViewsProxy, journeyView *platformclientv2.Journeyview) (*platformclientv2.Journeyview, *platformclientv2.APIResponse, error) {
	return p.journeyViewsApi.PostJourneyViews(*journeyView)
}

func updateJourneyViewFn(_ context.Context, p *journeyViewsProxy, viewId string, versionId int, journeyView *platformclientv2.Journeyview) (*platformclientv2.Journeyview, *platformclientv2.APIResponse, error) {
	versionIdToString := strconv.Itoa(versionId)
	return p.journeyViewsApi.PutJourneyViewVersion(viewId, versionIdToString, *journeyView)
}

func deleteJourneyViewFn(_ context.Context, p *journeyViewsProxy, viewId string) (*platformclientv2.APIResponse, error) {
	return p.journeyViewsApi.DeleteJourneyView(viewId)
}

// GetAllJourneyViewsFn is the implementation for retrieving all journey views in Genesys Cloud
func getAllJourneyViewsFn(ctx context.Context, p *journeyViewsProxy, name string) (*[]platformclientv2.Journeyview, *platformclientv2.APIResponse, error) {
	var allJourneys []platformclientv2.Journeyview
	const pageSize = 100

	journeys, resp, getErr := p.journeyViewsApi.GetJourneyViews(1, pageSize, name, "", "")
	if getErr != nil {
		return nil, resp, fmt.Errorf("failed to get first page of journeys: %v", getErr)
	}

	// Check if the journey view cache is populated, if it is, return that instead
	if rc.GetCacheSize(p.journeyViewCache) != 0 {
		return rc.GetCache(p.journeyViewCache), nil, nil
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
