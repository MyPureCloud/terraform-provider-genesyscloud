package journey_views

import (
	"context"
	rc "terraform-provider-genesyscloud/genesyscloud/resource_cache"

	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

var internalProxy *journeyViewsProxy

type getJourneyViewByViewIdFunc func(ctx context.Context, p *journeyViewsProxy, viewId string) (*platformclientv2.Journeyview, *platformclientv2.APIResponse, error)
type createJourneyViewFunc func(ctx context.Context, p *journeyViewsProxy, journeyView *platformclientv2.Journeyview) (*platformclientv2.Journeyview, *platformclientv2.APIResponse, error)
type updateJourneyViewFunc func(ctx context.Context, p *journeyViewsProxy, viewId string, journeyView *platformclientv2.Journeyview) (*platformclientv2.Journeyview, *platformclientv2.APIResponse, error)
type deleteJourneyViewFunc func(ctx context.Context, p *journeyViewsProxy, viewId string) (*platformclientv2.APIResponse, error)

type journeyViewsProxy struct {
	clientConfig          *platformclientv2.Configuration
	journeyViewsApi       *platformclientv2.JourneyApi
	getJourneyViewAttr    getJourneyViewByViewIdFunc
	createJourneyViewAttr createJourneyViewFunc
	updateJourneyViewAttr updateJourneyViewFunc
	deleteJourneyViewAttr deleteJourneyViewFunc
	journeyViewCache      rc.CacheInterface[platformclientv2.Journeyview]
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
