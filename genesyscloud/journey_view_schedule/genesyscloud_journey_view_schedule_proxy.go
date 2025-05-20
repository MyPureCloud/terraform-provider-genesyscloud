package journey_view_schedule

import (
	"context"
	"fmt"

	"github.com/mypurecloud/platform-client-sdk-go/v157/platformclientv2"

	rc "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_cache"
)

/*
The genesyscloud_journey_view_schedule_proxy.go file contains the proxy structures and methods that interact
with the Genesys Cloud SDK. We use composition here for each function on the proxy so individual functions can be stubbed
out during testing.

*/

// internalProxy holds a proxy instance that can be used throughout the package
var internalProxy *journeyViewScheduleProxy

// Type definitions for each func on our proxy so we can easily mock them out later
type getJourneyViewScheduleByViewIdFunc func(ctx context.Context, p *journeyViewScheduleProxy, viewId string) (*platformclientv2.Journeyviewschedule, *platformclientv2.APIResponse, error)
type createJourneyViewScheduleFunc func(ctx context.Context, p *journeyViewScheduleProxy, viewId string, journeyViewSchedule *platformclientv2.Journeyviewschedule) (*platformclientv2.Journeyviewschedule, *platformclientv2.APIResponse, error)
type updateJourneyViewScheduleFunc func(ctx context.Context, p *journeyViewScheduleProxy, viewId string, journeyViewSchedule *platformclientv2.Journeyviewschedule) (*platformclientv2.Journeyviewschedule, *platformclientv2.APIResponse, error)
type deleteJourneyViewScheduleFunc func(ctx context.Context, p *journeyViewScheduleProxy, viewId string) (*platformclientv2.APIResponse, error)
type getAllJourneyViewScheduleFunc func(ctx context.Context, p *journeyViewScheduleProxy) (*[]platformclientv2.Journeyviewschedule, *platformclientv2.APIResponse, error)

// journeyViewScheduleProxy contains all the methods that call genesys cloud APIs.
type journeyViewScheduleProxy struct {
	clientConfig                       *platformclientv2.Configuration
	journeyViewsApi                    *platformclientv2.JourneyApi
	getAllJourneyViewScheduleAttr      getAllJourneyViewScheduleFunc
	getJourneyViewScheduleByViewIdAttr getJourneyViewScheduleByViewIdFunc
	createJourneyViewScheduleAttr      createJourneyViewScheduleFunc
	updateJourneyViewScheduleAttr      updateJourneyViewScheduleFunc
	deleteJourneyViewScheduleAttr      deleteJourneyViewScheduleFunc
	journeyViewScheduleCache           rc.CacheInterface[platformclientv2.Journeyviewschedule]
}

// newJourneyViewScheduleProxy initializes the journey view schedule proxy with all the data needed to communicate with Genesys Cloud
func newJourneyViewScheduleProxy(clientConfig *platformclientv2.Configuration) *journeyViewScheduleProxy {
	api := platformclientv2.NewJourneyApiWithConfig(clientConfig)
	journeyViewScheduleCache := rc.NewResourceCache[platformclientv2.Journeyviewschedule]()
	return &journeyViewScheduleProxy{
		clientConfig:                       clientConfig,
		journeyViewsApi:                    api,
		getJourneyViewScheduleByViewIdAttr: getJourneyViewScheduleByViewIdFn,
		getAllJourneyViewScheduleAttr:      getAllJourneyViewScheduleFn,
		createJourneyViewScheduleAttr:      createJourneyViewScheduleFn,
		updateJourneyViewScheduleAttr:      updateJourneyViewScheduleFn,
		deleteJourneyViewScheduleAttr:      deleteJourneyViewScheduleFn,
		journeyViewScheduleCache:           journeyViewScheduleCache,
	}
}

// getJourneyViewScheduleProxy acts as a singleton to for the internalProxy.  It also ensures
// that we can still proxy our tests by directly setting internalProxy package variable
func getJourneyViewScheduleProxy(clientConfig *platformclientv2.Configuration) *journeyViewScheduleProxy {
	if internalProxy == nil {
		internalProxy = newJourneyViewScheduleProxy(clientConfig)
	}
	return internalProxy
}

func (p *journeyViewScheduleProxy) getJourneyViewScheduleByViewId(ctx context.Context, viewId string) (*platformclientv2.Journeyviewschedule, *platformclientv2.APIResponse, error) {
	return p.getJourneyViewScheduleByViewIdAttr(ctx, p, viewId)
}

func (p *journeyViewScheduleProxy) createJourneyViewSchedule(ctx context.Context, viewId string, journeyViewSchedule *platformclientv2.Journeyviewschedule) (*platformclientv2.Journeyviewschedule, *platformclientv2.APIResponse, error) {
	return p.createJourneyViewScheduleAttr(ctx, p, viewId, journeyViewSchedule)
}

func (p *journeyViewScheduleProxy) updateJourneyViewSchedule(ctx context.Context, viewId string, journeyViewSchedule *platformclientv2.Journeyviewschedule) (*platformclientv2.Journeyviewschedule, *platformclientv2.APIResponse, error) {
	return p.updateJourneyViewScheduleAttr(ctx, p, viewId, journeyViewSchedule)
}

func (p *journeyViewScheduleProxy) deleteJourneyViewSchedule(ctx context.Context, viewId string) (*platformclientv2.APIResponse, error) {
	return p.deleteJourneyViewScheduleAttr(ctx, p, viewId)
}

func (p *journeyViewScheduleProxy) getAllJourneyViewSchedule(ctx context.Context) (*[]platformclientv2.Journeyviewschedule, *platformclientv2.APIResponse, error) {
	return p.getAllJourneyViewScheduleAttr(ctx, p)
}

func getJourneyViewScheduleByViewIdFn(ctx context.Context, p *journeyViewScheduleProxy, viewId string) (*platformclientv2.Journeyviewschedule, *platformclientv2.APIResponse, error) {
	// Check the cache first
	journeyViewSchedule := rc.GetCacheItem(p.journeyViewScheduleCache, viewId)
	if journeyViewSchedule != nil {
		return journeyViewSchedule, nil, nil
	}
	return p.journeyViewsApi.GetJourneyViewSchedules(viewId)
}

func createJourneyViewScheduleFn(ctx context.Context, p *journeyViewScheduleProxy, viewId string, journeyViewSchedule *platformclientv2.Journeyviewschedule) (*platformclientv2.Journeyviewschedule, *platformclientv2.APIResponse, error) {
	return p.journeyViewsApi.PostJourneyViewSchedules(viewId, *journeyViewSchedule)
}

func updateJourneyViewScheduleFn(ctx context.Context, p *journeyViewScheduleProxy, viewId string, journeyViewSchedule *platformclientv2.Journeyviewschedule) (*platformclientv2.Journeyviewschedule, *platformclientv2.APIResponse, error) {
	return p.journeyViewsApi.PutJourneyViewSchedules(viewId, *journeyViewSchedule)
}

func deleteJourneyViewScheduleFn(ctx context.Context, p *journeyViewScheduleProxy, viewId string) (*platformclientv2.APIResponse, error) {
	_, resp, err := p.journeyViewsApi.DeleteJourneyViewSchedules(viewId)
	if err != nil {
		return resp, err
	}
	rc.DeleteCacheItem(p.journeyViewScheduleCache, viewId)
	return resp, nil
}

// getAllJourneyViewScheduleFn is the implementation for retrieving all journey view schedule in Genesys Cloud
func getAllJourneyViewScheduleFn(ctx context.Context, p *journeyViewScheduleProxy) (*[]platformclientv2.Journeyviewschedule, *platformclientv2.APIResponse, error) {
	var allJourneyViewSchedules []platformclientv2.Journeyviewschedule
	const pageSize = 100

	journeyViewSchedules, resp, err := p.journeyViewsApi.GetJourneyViewsSchedules(1, pageSize)
	if err != nil {
		return nil, resp, fmt.Errorf("failed to get first page of journey view schedules: %v", err)
	}
	if journeyViewSchedules.Entities == nil || len(*journeyViewSchedules.Entities) == 0 {
		return &allJourneyViewSchedules, resp, nil
	}
	for _, journeySchedule := range *journeyViewSchedules.Entities {
		allJourneyViewSchedules = append(allJourneyViewSchedules, journeySchedule)
	}

	for pageNum := 2; pageNum <= *journeyViewSchedules.PageCount; pageNum++ {
		journeyViewSchedules, resp, err := p.journeyViewsApi.GetJourneyViewsSchedules(pageNum, pageSize)
		if err != nil {
			return nil, resp, fmt.Errorf("failed to get page of journey view schedules: %v", err)
		}

		if journeyViewSchedules.Entities == nil || len(*journeyViewSchedules.Entities) == 0 {
			break
		}

		for _, journeySchedule := range *journeyViewSchedules.Entities {
			allJourneyViewSchedules = append(allJourneyViewSchedules, journeySchedule)
		}
	}

	// Cache the journey view schedule resources into the cache
	for _, schedules := range allJourneyViewSchedules {
		if schedules.Id == nil {
			continue
		}
		rc.SetCache(p.journeyViewScheduleCache, *schedules.Id, schedules)
	}

	return &allJourneyViewSchedules, resp, nil
}
