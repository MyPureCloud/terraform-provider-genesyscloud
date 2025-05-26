package journey_segment

import (
	"context"
	"fmt"
	rc "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_cache"

	"github.com/mypurecloud/platform-client-sdk-go/v157/platformclientv2"
)

/*
The file genesyscloud_journey_segment_proxy.go manages the interaction between our software and
the Genesys Cloud SDK. Within this file, we define proxy structures and methods.
We employ a technique called composition for each function on the proxy. This means that each function
is built by combining smaller, independent parts. One advantage of this approach is that it allows us
to isolate and test individual functions more easily. For testing purposes, we can replace or
simulate these smaller parts, known as stubs, to ensure that each function behaves correctly in different scenarios.
*/

// internalProxy holds a proxy instance that can be used throughout the package
var internalProxy *journeySegmentProxy

// Type definitions for each func on our proxy so we can easily mock them out later
type createJourneySegmentFunc func(ctx context.Context, p *journeySegmentProxy, segment *platformclientv2.Journeysegmentrequest) (*platformclientv2.Journeysegment, *platformclientv2.APIResponse, error)
type getAllJourneySegmentsFunc func(ctx context.Context, p *journeySegmentProxy) (*[]platformclientv2.Journeysegment, *platformclientv2.APIResponse, error)
type getJourneySegmentIdByNameFunc func(ctx context.Context, p *journeySegmentProxy, name string) (id string, retryable bool, response *platformclientv2.APIResponse, err error)
type getJourneySegmentByIdFunc func(ctx context.Context, p *journeySegmentProxy, id string) (segment *platformclientv2.Journeysegment, response *platformclientv2.APIResponse, err error)
type updateJourneySegmentFunc func(ctx context.Context, p *journeySegmentProxy, id string, segment *platformclientv2.Patchsegment) (*platformclientv2.Journeysegment, *platformclientv2.APIResponse, error)
type deleteJourneySegmentFunc func(ctx context.Context, p *journeySegmentProxy, id string) (*platformclientv2.APIResponse, error)

/*
The journeySegmentProxy struct holds all the methods responsible for making calls to
the Genesys Cloud APIs. This means that within this struct, you'll find all the functions designed
to interact directly with the various features and services offered by Genesys Cloud,
enabling this terraform provider software to perform tasks like retrieving data, updating information,
or triggering actions within the Genesys Cloud environment.
*/
type journeySegmentProxy struct {
	clientConfig                  *platformclientv2.Configuration
	journeyApi                    *platformclientv2.JourneyApi
	createJourneySegmentAttr      createJourneySegmentFunc
	getAllJourneySegmentsAttr     getAllJourneySegmentsFunc
	getJourneySegmentIdByNameAttr getJourneySegmentIdByNameFunc
	getJourneySegmentByIdAttr     getJourneySegmentByIdFunc
	updateJourneySegmentAttr      updateJourneySegmentFunc
	deleteJourneySegmentAttr      deleteJourneySegmentFunc
	segmentCache                  rc.CacheInterface[platformclientv2.Journeysegment]
}

/*
The function newJourneySegmentProxy sets up the journey segment proxy by providing it
with all the necessary information to communicate effectively with Genesys Cloud.
This includes configuring the proxy with the required data and settings so that it can interact
seamlessly with the Genesys Cloud platform.
*/
func newJourneySegmentProxy(clientConfig *platformclientv2.Configuration) *journeySegmentProxy {
	api := platformclientv2.NewJourneyApiWithConfig(clientConfig)
	segmentCache := rc.NewResourceCache[platformclientv2.Journeysegment]()

	return &journeySegmentProxy{
		clientConfig:                  clientConfig,
		journeyApi:                    api,
		segmentCache:                  segmentCache,
		createJourneySegmentAttr:      createJourneySegmentFn,
		getAllJourneySegmentsAttr:     getAllJourneySegmentsFn,
		getJourneySegmentIdByNameAttr: getJourneySegmentIdByNameFn,
		getJourneySegmentByIdAttr:     getJourneySegmentByIdFn,
		updateJourneySegmentAttr:      updateJourneySegmentFn,
		deleteJourneySegmentAttr:      deleteJourneySegmentFn,
	}
}

/*
The function getJourneySegmentProxy serves a dual purpose: first, it functions as a singleton for
the internalProxy, meaning it ensures that only one instance of the internalProxy exists. Second,
it enables us to proxy our tests by allowing us to directly set the internalProxy package variable.
This ensures consistency and control in managing the internalProxy across our codebase, while also
facilitating efficient testing by providing a straightforward way to substitute the proxy for testing purposes.
*/
func getJourneySegmentProxy(clientConfig *platformclientv2.Configuration) *journeySegmentProxy {
	if internalProxy == nil {
		internalProxy = newJourneySegmentProxy(clientConfig)
	}
	return internalProxy
}

// createJourneySegment creates a Genesys Cloud journey segment
func (p *journeySegmentProxy) createJourneySegment(ctx context.Context, segment *platformclientv2.Journeysegmentrequest) (*platformclientv2.Journeysegment, *platformclientv2.APIResponse, error) {
	return p.createJourneySegmentAttr(ctx, p, segment)
}

// getAllJourneySegments retrieves all Genesys Cloud journey segments
func (p *journeySegmentProxy) getAllJourneySegments(ctx context.Context) (*[]platformclientv2.Journeysegment, *platformclientv2.APIResponse, error) {
	return p.getAllJourneySegmentsAttr(ctx, p)
}

// getJourneySegmentIdByName returns a single Genesys Cloud journey segment by name
func (p *journeySegmentProxy) getJourneySegmentIdByName(ctx context.Context, name string) (id string, retryable bool, response *platformclientv2.APIResponse, err error) {
	return p.getJourneySegmentIdByNameAttr(ctx, p, name)
}

// getJourneySegmentById returns a single Genesys Cloud journey segment by Id
func (p *journeySegmentProxy) getJourneySegmentById(ctx context.Context, id string) (segment *platformclientv2.Journeysegment, response *platformclientv2.APIResponse, err error) {
	if segment := rc.GetCacheItem(p.segmentCache, id); segment != nil {
		return segment, nil, nil
	}
	return p.getJourneySegmentByIdAttr(ctx, p, id)
}

// updateJourneySegment updates a Genesys Cloud journey segment
func (p *journeySegmentProxy) updateJourneySegment(ctx context.Context, id string, segment *platformclientv2.Patchsegment) (*platformclientv2.Journeysegment, *platformclientv2.APIResponse, error) {
	return p.updateJourneySegmentAttr(ctx, p, id, segment)
}

// deleteJourneySegment deletes a Genesys Cloud journey segment by Id
func (p *journeySegmentProxy) deleteJourneySegment(ctx context.Context, id string) (*platformclientv2.APIResponse, error) {
	return p.deleteJourneySegmentAttr(ctx, p, id)
}

// getAllJourneySegmentsFn is the implementation for retrieving all journey segments in Genesys Cloud
func getAllJourneySegmentsFn(ctx context.Context, p *journeySegmentProxy) (*[]platformclientv2.Journeysegment, *platformclientv2.APIResponse, error) {
	if p == nil || p.journeyApi == nil {
		return nil, nil, fmt.Errorf("invalid journey segment proxy or API client")
	}

	var allSegments []platformclientv2.Journeysegment
	const pageSize = 100

	// Get first page
	segments, resp, err := p.journeyApi.GetJourneySegments("", pageSize, 1, true, nil, nil, "")
	if err != nil {
		return nil, resp, err
	}

	if segments == nil {
		return &allSegments, resp, nil
	}

	if segments.Entities == nil || len(*segments.Entities) == 0 {
		return &allSegments, resp, nil
	}

	allSegments = append(allSegments, *segments.Entities...)

	// Check if pageCount is nil before dereferencing
	if segments.PageCount == nil {
		return &allSegments, resp, nil
	}

	// Get remaining pages
	for pageNum := 2; pageNum <= *segments.PageCount; pageNum++ {
		segments, resp, err := p.journeyApi.GetJourneySegments("", pageSize, pageNum, true, nil, nil, "")
		if err != nil {
			return nil, resp, err
		}

		if segments == nil || segments.Entities == nil || len(*segments.Entities) == 0 {
			break
		}

		allSegments = append(allSegments, *segments.Entities...)
	}

	// Cache the segments only if they have valid IDs
	for _, segment := range allSegments {
		if segment.Id != nil {
			rc.SetCache(p.segmentCache, *segment.Id, segment)
		}
	}

	return &allSegments, resp, nil
}

// getJourneySegmentIdByNameFn retrieves a journey segment ID by its name
func getJourneySegmentIdByNameFn(ctx context.Context, p *journeySegmentProxy, name string) (id string, retryable bool, response *platformclientv2.APIResponse, err error) {
	if p == nil {
		return "", false, nil, fmt.Errorf("invalid journey segment proxy")
	}

	if name == "" {
		return "", false, nil, fmt.Errorf("name cannot be empty")
	}

	segments, resp, err := p.getAllJourneySegmentsAttr(ctx, p)
	if err != nil {
		return "", false, resp, err
	}

	if segments == nil {
		return "", true, resp, fmt.Errorf("no journey segments found")
	}

	for _, segment := range *segments {
		if segment.DisplayName == nil {
			continue
		}
		if *segment.DisplayName == name {
			if segment.Id == nil {
				return "", false, resp, fmt.Errorf("journey segment found but has no ID")
			}
			return *segment.Id, false, resp, nil
		}
	}

	return "", true, resp, fmt.Errorf("no journey segment found with name %s", name)
}

// getJourneySegmentByIdFn retrieves a journey segment by its ID
func getJourneySegmentByIdFn(ctx context.Context, p *journeySegmentProxy, id string) (*platformclientv2.Journeysegment, *platformclientv2.APIResponse, error) {
	if p == nil {
		return nil, nil, fmt.Errorf("invalid journey segment proxy")
	}

	if id == "" {
		return nil, nil, fmt.Errorf("id cannot be empty")
	}

	// Make API call if not in cache
	if p.journeyApi == nil {
		return nil, nil, fmt.Errorf("journey API client is nil")
	}

	segment, resp, err := p.journeyApi.GetJourneySegment(id)
	if err != nil {
		return nil, resp, err
	}

	if segment == nil {
		return nil, resp, fmt.Errorf("retrieved journey segment is nil")
	}

	return segment, resp, nil
}

// createJourneySegmentFn is an implementation function for creating a Genesys Cloud journey segment
func createJourneySegmentFn(ctx context.Context, p *journeySegmentProxy, segment *platformclientv2.Journeysegmentrequest) (*platformclientv2.Journeysegment, *platformclientv2.APIResponse, error) {
	respSegment, resp, err := p.journeyApi.PostJourneySegments(*segment)
	if err != nil {
		return nil, resp, err
	}
	return respSegment, resp, nil
}

// updateJourneySegmentFn updates an existing journey segment
func updateJourneySegmentFn(ctx context.Context, p *journeySegmentProxy, id string, segment *platformclientv2.Patchsegment) (*platformclientv2.Journeysegment, *platformclientv2.APIResponse, error) {
	updatedSegment, resp, err := p.journeyApi.PatchJourneySegment(id, *segment)
	if err != nil {
		return nil, resp, err
	}

	// Update cache
	p.segmentCache.Set(id, *updatedSegment)
	return updatedSegment, resp, nil
}

// deleteJourneySegmentFn deletes a journey segment by its ID
func deleteJourneySegmentFn(ctx context.Context, p *journeySegmentProxy, id string) (*platformclientv2.APIResponse, error) {
	resp, err := p.journeyApi.DeleteJourneySegment(id)
	if err != nil {
		return resp, err
	}

	// Remove from cache
	rc.DeleteCacheItem(p.segmentCache, id)
	return resp, nil
}
