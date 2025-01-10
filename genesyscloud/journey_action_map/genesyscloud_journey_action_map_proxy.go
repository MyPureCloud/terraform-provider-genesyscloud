package journey_action_map

import (
	"context"
	"fmt"
	"log"
	rc "terraform-provider-genesyscloud/genesyscloud/resource_cache"

	"github.com/mypurecloud/platform-client-sdk-go/v150/platformclientv2"
)

/*
The file genesyscloud_journey_action_map_proxy.go manages the interaction between our software and
the Genesys Cloud SDK. Within this file, we define proxy structures and methods.
We employ a technique called composition for each function on the proxy. This means that each function
is built by combining smaller, independent parts. One advantage of this approach is that it allows us
to isolate and test individual functions more easily. For testing purposes, we can replace or
simulate these smaller parts, known as stubs, to ensure that each function behaves correctly in different scenarios.
*/

// internalProxy holds a proxy instance that can be used throughout the package
var internalProxy *journeyActionMapProxy

// Type definitions for each func on our proxy so we can easily mock them out later
type createJourneyActionMapFunc func(ctx context.Context, p *journeyActionMapProxy, actionMap *platformclientv2.Actionmap) (*platformclientv2.Actionmap, *platformclientv2.APIResponse, error)
type getAllJourneyActionMapsFunc func(ctx context.Context, p *journeyActionMapProxy) (*[]platformclientv2.Actionmap, *platformclientv2.APIResponse, error)
type getJourneyActionMapIdByNameFunc func(ctx context.Context, p *journeyActionMapProxy, name string) (id string, retryable bool, response *platformclientv2.APIResponse, err error)
type getJourneyActionMapByIdFunc func(ctx context.Context, p *journeyActionMapProxy, id string) (actionMap *platformclientv2.Actionmap, response *platformclientv2.APIResponse, err error)
type updateJourneyActionMapFunc func(ctx context.Context, p *journeyActionMapProxy, id string, actionMap *platformclientv2.Patchactionmap) (*platformclientv2.Actionmap, *platformclientv2.APIResponse, error)
type deleteJourneyActionMapFunc func(ctx context.Context, p *journeyActionMapProxy, id string) (*platformclientv2.APIResponse, error)

/*
The journeyActionMapProxy struct holds all the methods responsible for making calls to
the Genesys Cloud APIs. This means that within this struct, you'll find all the functions designed
to interact directly with the various features and services offered by Genesys Cloud,
enabling this terraform provider software to perform tasks like retrieving data, updating information,
or triggering actions within the Genesys Cloud environment.
*/
type journeyActionMapProxy struct {
	clientConfig                    *platformclientv2.Configuration
	journeyApi                      *platformclientv2.JourneyApi
	createJourneyActionMapAttr      createJourneyActionMapFunc
	getAllJourneyActionMapsAttr     getAllJourneyActionMapsFunc
	getJourneyActionMapIdByNameAttr getJourneyActionMapIdByNameFunc
	getJourneyActionMapByIdAttr     getJourneyActionMapByIdFunc
	updateJourneyActionMapAttr      updateJourneyActionMapFunc
	deleteJourneyActionMapAttr      deleteJourneyActionMapFunc
	actionMapCache                  rc.CacheInterface[platformclientv2.Actionmap]
}

/*
The function newJourneyActionMapProxy sets up the journey action map proxy by providing it
with all the necessary information to communicate effectively with Genesys Cloud.
This includes configuring the proxy with the required data and settings so that it can interact
seamlessly with the Genesys Cloud platform.
*/
func newJourneyActionMapProxy(clientConfig *platformclientv2.Configuration) *journeyActionMapProxy {
	api := platformclientv2.NewJourneyApiWithConfig(clientConfig)
	actionMapCache := rc.NewResourceCache[platformclientv2.Actionmap]()

	return &journeyActionMapProxy{
		clientConfig:                    clientConfig,
		journeyApi:                      api,
		actionMapCache:                  actionMapCache,
		createJourneyActionMapAttr:      createJourneyActionMapFn,
		getAllJourneyActionMapsAttr:     getAllJourneyActionMapsFn,
		getJourneyActionMapIdByNameAttr: getJourneyActionMapIdByNameFn,
		getJourneyActionMapByIdAttr:     getJourneyActionMapByIdFn,
		updateJourneyActionMapAttr:      updateJourneyActionMapFn,
		deleteJourneyActionMapAttr:      deleteJourneyActionMapFn,
	}
}

/*
The function getJourneyActionMapProxy serves a dual purpose: first, it functions as a singleton for
the internalProxy, meaning it ensures that only one instance of the internalProxy exists. Second,
it enables us to proxy our tests by allowing us to directly set the internalProxy package variable.
This ensures consistency and control in managing the internalProxy across our codebase, while also
facilitating efficient testing by providing a straightforward way to substitute the proxy for testing purposes.
*/
func getJourneyActionMapProxy(clientConfig *platformclientv2.Configuration) *journeyActionMapProxy {
	if internalProxy == nil {
		internalProxy = newJourneyActionMapProxy(clientConfig)
	}
	return internalProxy
}

// createJourneyActionMap creates a Genesys Cloud journey action map
func (p *journeyActionMapProxy) createJourneyActionMap(ctx context.Context, actionMap *platformclientv2.Actionmap) (*platformclientv2.Actionmap, *platformclientv2.APIResponse, error) {
	return p.createJourneyActionMapAttr(ctx, p, actionMap)
}

// getAllJourneyActionMaps retrieves all Genesys Cloud journey action maps
func (p *journeyActionMapProxy) getAllJourneyActionMaps(ctx context.Context) (*[]platformclientv2.Actionmap, *platformclientv2.APIResponse, error) {
	return p.getAllJourneyActionMapsAttr(ctx, p)
}

// getJourneyActionMapIdByName returns a single Genesys Cloud journey action map by name
func (p *journeyActionMapProxy) getJourneyActionMapIdByName(ctx context.Context, name string) (id string, retryable bool, response *platformclientv2.APIResponse, err error) {
	return p.getJourneyActionMapIdByNameAttr(ctx, p, name)
}

// getJourneyActionMapById returns a single Genesys Cloud journey action map by Id
func (p *journeyActionMapProxy) getJourneyActionMapById(ctx context.Context, id string) (actionMap *platformclientv2.Actionmap, response *platformclientv2.APIResponse, err error) {
	if actionMap := rc.GetCacheItem(p.actionMapCache, id); actionMap != nil {
		return actionMap, nil, nil
	}
	return p.getJourneyActionMapByIdAttr(ctx, p, id)
}

// updateJourneyActionMap updates a Genesys Cloud journey action map
func (p *journeyActionMapProxy) updateJourneyActionMap(ctx context.Context, id string, actionMap *platformclientv2.Patchactionmap) (*platformclientv2.Actionmap, *platformclientv2.APIResponse, error) {
	return p.updateJourneyActionMapAttr(ctx, p, id, actionMap)
}

// deleteJourneyActionMap deletes a Genesys Cloud journey action map by Id
func (p *journeyActionMapProxy) deleteJourneyActionMap(ctx context.Context, id string) (*platformclientv2.APIResponse, error) {
	return p.deleteJourneyActionMapAttr(ctx, p, id)
}

// createJourneyActionMapFn is an implementation function for creating a Genesys Cloud journey action map
func createJourneyActionMapFn(ctx context.Context, p *journeyActionMapProxy, actionMap *platformclientv2.Actionmap) (*platformclientv2.Actionmap, *platformclientv2.APIResponse, error) {
	actionLMap, resp, err := p.journeyApi.PostJourneyActionmaps(*actionMap)
	if err != nil {
		return nil, resp, fmt.Errorf("Failed to create journey action map: %s", err)
	}
	return actionLMap, resp, nil
}

// getAllJourneyActionMapsFn is the implementation for retrieving all journey action maps in Genesys Cloud
func getAllJourneyActionMapsFn(ctx context.Context, p *journeyActionMapProxy) (*[]platformclientv2.Actionmap, *platformclientv2.APIResponse, error) {
	var allActionMaps []platformclientv2.Actionmap
	const pageSize = 100

	actionMaps, resp, err := p.journeyApi.GetJourneyActionmaps(pageSize, 1, "", "", "", nil, nil, "")
	if err != nil {
		return nil, resp, fmt.Errorf("Failed to get journey action maps: %s", err)
	}

	if actionMaps == nil || actionMaps.Entities == nil || len(*actionMaps.Entities) == 0 {
		return &allActionMaps, resp, nil
	}

	allActionMaps = append(allActionMaps, *actionMaps.Entities...)

	for pageNum := 2; pageNum <= *actionMaps.PageCount; pageNum++ {
		actionMaps, resp, err := p.journeyApi.GetJourneyActionmaps(pageSize, pageNum, "", "", "", nil, nil, "")
		if err != nil {
			return nil, resp, fmt.Errorf("Failed to get journey action maps page %d: %s", pageNum, err)
		}
		if actionMaps == nil || actionMaps.Entities == nil || len(*actionMaps.Entities) == 0 {
			break
		}

		allActionMaps = append(allActionMaps, *actionMaps.Entities...)
	}

	// Cache the architect schedules resource into the p.schedulesCache for later use
	for _, actionMap := range allActionMaps {
		rc.SetCache(p.actionMapCache, *actionMap.Id, actionMap)
	}

	return &allActionMaps, resp, nil
}

// getJourneyActionMapIdByNameFn is an implementation function for getting a Genesys Cloud journey action map by name
func getJourneyActionMapIdByNameFn(ctx context.Context, p *journeyActionMapProxy, name string) (id string, retryable bool, response *platformclientv2.APIResponse, err error) {
	actionMaps, apiResponse, err := getAllJourneyActionMapsFn(ctx, p)
	if err != nil {
		return "", false, apiResponse, err
	}

	if actionMaps == nil || len(*actionMaps) == 0 {
		return "", true, apiResponse, fmt.Errorf("No journey action map found with name %s", name)
	}

	for _, actionMap := range *actionMaps {
		if *actionMap.DisplayName == name {
			log.Printf("Retrieved the Journey action map id %s by name %s", *actionMap.Id, name)
			return *actionMap.Id, false, apiResponse, nil
		}
	}

	return "", true, apiResponse, fmt.Errorf("Unable to find Journey action map with name %s", name)
}

// getJourneyActionMapByIdFn is an implementation function for getting a Genesys Cloud journey action map by ID
func getJourneyActionMapByIdFn(ctx context.Context, p *journeyActionMapProxy, id string) (actionMap *platformclientv2.Actionmap, response *platformclientv2.APIResponse, err error) {
	actionMap, resp, err := p.journeyApi.GetJourneyActionmap(id)
	if err != nil {
		return nil, resp, fmt.Errorf("Failed to get journey action map by id %s: %s", id, err)
	}
	return actionMap, resp, nil
}

// updateJourneyActionMapFn is an implementation function for updating a Genesys Cloud journey action map
func updateJourneyActionMapFn(ctx context.Context, p *journeyActionMapProxy, id string, journeyActionMap *platformclientv2.Patchactionmap) (*platformclientv2.Actionmap, *platformclientv2.APIResponse, error) {
	actionMap, apiResponse, err := getJourneyActionMapByIdFn(ctx, p, id)
	if err != nil {
		return nil, apiResponse, fmt.Errorf("Failed to get Journy action map  %s by id: %s", id, err)
	}
	journeyActionMap.Version = actionMap.Version

	actionMap, resp, err := p.journeyApi.PatchJourneyActionmap(id, *journeyActionMap)
	if err != nil {
		return nil, resp, fmt.Errorf("Failed to update journey action map %s: %s", id, err)
	}
	return actionMap, resp, nil
}

// deleteJourneyActionMapFn is an implementation function for deleting a Genesys Cloud journey action map
func deleteJourneyActionMapFn(ctx context.Context, p *journeyActionMapProxy, id string) (*platformclientv2.APIResponse, error) {
	resp, err := p.journeyApi.DeleteJourneyActionmap(id)
	if err != nil {
		return resp, fmt.Errorf("Failed to delete journey action map %s: %s", id, err)
	}
	return resp, nil
}
