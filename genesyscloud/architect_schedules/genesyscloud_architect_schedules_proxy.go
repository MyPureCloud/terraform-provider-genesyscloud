package architect_schedules

import (
	"context"
	"fmt"
	"log"
	rc "terraform-provider-genesyscloud/genesyscloud/resource_cache"

	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

/*
The file genesyscloud_architect_schedules_proxy.go manages the interaction between our software and
the Genesys Cloud SDK. Within this file, we define proxy structures and methods.
We employ a technique called composition for each function on the proxy. This means that each function
is built by combining smaller, independent parts. One advantage of this approach is that it allows us
to isolate and test individual functions more easily. For testing purposes, we can replace or
simulate these smaller parts, known as stubs, to ensure that each function behaves correctly in different scenarios.
*/

// internalProxy holds a proxy instance that can be used throughout the package
var internalProxy *architectSchedulesProxy

// Type definitions for each func on our proxy so we can easily mock them out later
type createArchitectSchedulesFunc func(ctx context.Context, p *architectSchedulesProxy, schedules *platformclientv2.Schedule) (*platformclientv2.Schedule, *platformclientv2.APIResponse, error)
type getAllArchitectSchedulesFunc func(ctx context.Context, p *architectSchedulesProxy) (*[]platformclientv2.Schedule, *platformclientv2.APIResponse, error)
type getArchitectSchedulesIdByNameFunc func(ctx context.Context, p *architectSchedulesProxy, name string) (id string, retryable bool, response *platformclientv2.APIResponse, err error)
type getArchitectSchedulesByIdFunc func(ctx context.Context, p *architectSchedulesProxy, id string) (schedules *platformclientv2.Schedule, response *platformclientv2.APIResponse, err error)
type updateArchitectSchedulesFunc func(ctx context.Context, p *architectSchedulesProxy, id string, schedules *platformclientv2.Schedule) (*platformclientv2.Schedule, *platformclientv2.APIResponse, error)
type deleteArchitectSchedulesFunc func(ctx context.Context, p *architectSchedulesProxy, id string) (*platformclientv2.APIResponse, error)

/*
The architectSchedulesProxy struct holds all the methods responsible for making calls to
the Genesys Cloud APIs. This means that within this struct, you'll find all the functions designed
to interact directly with the various features and services offered by Genesys Cloud,
enabling this terraform provider software to perform tasks like retrieving data, updating information,
or triggering actions within the Genesys Cloud environment.
*/
type architectSchedulesProxy struct {
	clientConfig                      *platformclientv2.Configuration
	architectApi                      *platformclientv2.ArchitectApi
	createArchitectSchedulesAttr      createArchitectSchedulesFunc
	getAllArchitectSchedulesAttr      getAllArchitectSchedulesFunc
	getArchitectSchedulesIdByNameAttr getArchitectSchedulesIdByNameFunc
	getArchitectSchedulesByIdAttr     getArchitectSchedulesByIdFunc
	updateArchitectSchedulesAttr      updateArchitectSchedulesFunc
	deleteArchitectSchedulesAttr      deleteArchitectSchedulesFunc
	schedulesCache                    rc.CacheInterface[platformclientv2.Schedule] //Define the cache for architect schedules resource
}

/*
The function newArchitectSchedulesProxy sets up the architect schedules proxy by providing it
with all the necessary information to communicate effectively with Genesys Cloud.
This includes configuring the proxy with the required data and settings so that it can interact
seamlessly with the Genesys Cloud platform.
*/
func newArchitectSchedulesProxy(clientConfig *platformclientv2.Configuration) *architectSchedulesProxy {
	api := platformclientv2.NewArchitectApiWithConfig(clientConfig)    // NewArchitectApiWithConfig creates an Genesyc Cloud API instance using the provided configuration
	schedulesCache := rc.NewResourceCache[platformclientv2.Schedule]() // Create Cache for architect schedules resource
	return &architectSchedulesProxy{
		clientConfig:                      clientConfig,
		architectApi:                      api,
		schedulesCache:                    schedulesCache,
		createArchitectSchedulesAttr:      createArchitectSchedulesFn,
		getAllArchitectSchedulesAttr:      getAllArchitectSchedulesFn,
		getArchitectSchedulesIdByNameAttr: getArchitectSchedulesIdByNameFn,
		getArchitectSchedulesByIdAttr:     getArchitectSchedulesByIdFn,
		updateArchitectSchedulesAttr:      updateArchitectSchedulesFn,
		deleteArchitectSchedulesAttr:      deleteArchitectSchedulesFn,
	}
}

/*
The function getArchitectSchedulesProxy serves a dual purpose: first, it functions as a singleton for
the internalProxy, meaning it ensures that only one instance of the internalProxy exists. Second,
it enables us to proxy our tests by allowing us to directly set the internalProxy package variable.
This ensures consistency and control in managing the internalProxy across our codebase, while also
facilitating efficient testing by providing a straightforward way to substitute the proxy for testing purposes.
*/
func getArchitectSchedulesProxy(clientConfig *platformclientv2.Configuration) *architectSchedulesProxy {
	if internalProxy == nil {
		internalProxy = newArchitectSchedulesProxy(clientConfig)
	}
	return internalProxy
}

// createArchitectSchedules creates a Genesys Cloud architect schedules
func (p *architectSchedulesProxy) createArchitectSchedules(ctx context.Context, architectSchedules *platformclientv2.Schedule) (*platformclientv2.Schedule, *platformclientv2.APIResponse, error) {
	return p.createArchitectSchedulesAttr(ctx, p, architectSchedules)
}

// getArchitectSchedules retrieves all Genesys Cloud architect schedules
func (p *architectSchedulesProxy) getAllArchitectSchedules(ctx context.Context) (*[]platformclientv2.Schedule, *platformclientv2.APIResponse, error) {
	return p.getAllArchitectSchedulesAttr(ctx, p)
}

// getArchitectSchedulesIdByName returns a single Genesys Cloud architect schedules by a name
func (p *architectSchedulesProxy) getArchitectSchedulesIdByName(ctx context.Context, name string) (id string, retryable bool, response *platformclientv2.APIResponse, err error) {
	return p.getArchitectSchedulesIdByNameAttr(ctx, p, name)
}

// getArchitectSchedulesById returns a single Genesys Cloud architect schedules by Id
func (p *architectSchedulesProxy) getArchitectSchedulesById(ctx context.Context, id string) (architectSchedules *platformclientv2.Schedule, response *platformclientv2.APIResponse, err error) {
	if schedule := rc.GetCacheItem(p.schedulesCache, id); schedule != nil { // Get the schedule from the cache, if not there in the cache then call p.getArchitectSchedulesByIdAttr()
		return schedule, nil, nil
	}
	return p.getArchitectSchedulesByIdAttr(ctx, p, id)
}

// updateArchitectSchedules updates a Genesys Cloud architect schedules
func (p *architectSchedulesProxy) updateArchitectSchedules(ctx context.Context, id string, architectSchedules *platformclientv2.Schedule) (*platformclientv2.Schedule, *platformclientv2.APIResponse, error) {
	return p.updateArchitectSchedulesAttr(ctx, p, id, architectSchedules)
}

// deleteArchitectSchedules deletes a Genesys Cloud architect schedules by Id
func (p *architectSchedulesProxy) deleteArchitectSchedules(ctx context.Context, id string) (*platformclientv2.APIResponse, error) {
	return p.deleteArchitectSchedulesAttr(ctx, p, id)
}

// createArchitectSchedulesFn is an implementation function for creating a Genesys Cloud architect schedules
func createArchitectSchedulesFn(ctx context.Context, p *architectSchedulesProxy, architectSchedules *platformclientv2.Schedule) (*platformclientv2.Schedule, *platformclientv2.APIResponse, error) {
	schedules, apiResponse, err := p.architectApi.PostArchitectSchedules(*architectSchedules)
	if err != nil {
		return nil, apiResponse, fmt.Errorf("Failed to create architect schedules: %s", err)
	}
	return schedules, apiResponse, nil
}

// getAllArchitectSchedulesFn is the implementation for retrieving all architect schedules in Genesys Cloud
func getAllArchitectSchedulesFn(ctx context.Context, p *architectSchedulesProxy) (*[]platformclientv2.Schedule, *platformclientv2.APIResponse, error) {
	var allSchedules []platformclientv2.Schedule
	const pageSize = 100

	schedules, apiResponse, err := p.architectApi.GetArchitectSchedules(1, pageSize, "", "", "", nil)
	if err != nil {
		return nil, apiResponse, fmt.Errorf("Failed to get schedule : %v", err)
	}

	if schedules == nil || schedules.Entities == nil || len(*schedules.Entities) == 0 {
		return &allSchedules, apiResponse, nil
	}

	allSchedules = append(allSchedules, *schedules.Entities...)

	for pageNum := 2; pageNum <= *schedules.PageCount; pageNum++ {
		schedules, apiResponse, err := p.architectApi.GetArchitectSchedules(pageNum, pageSize, "", "", "", nil)
		if err != nil {
			return nil, apiResponse, fmt.Errorf("Failed to get schedule : %v", err)
		}

		if schedules == nil || schedules.Entities == nil || len(*schedules.Entities) == 0 {
			break
		}

		allSchedules = append(allSchedules, *schedules.Entities...)
	}

	// Cache the architect schedules resource into the p.schedulesCache for later use
	for _, schedule := range allSchedules {
		rc.SetCache(p.schedulesCache, *schedule.Id, schedule)
	}

	return &allSchedules, apiResponse, nil
}

// getArchitectSchedulesIdByNameFn is an implementation of the function to get a Genesys Cloud architect schedules by name
func getArchitectSchedulesIdByNameFn(ctx context.Context, p *architectSchedulesProxy, name string) (id string, retryable bool, response *platformclientv2.APIResponse, err error) {
	schedules, apiResponse, err := getAllArchitectSchedulesFn(ctx, p)
	if err != nil {
		return "", false, apiResponse, err
	}

	if schedules == nil || len(*schedules) == 0 {
		return "", true, apiResponse, fmt.Errorf("No architect schedules found with name %s", name)
	}

	for _, schedules := range *schedules {
		if *schedules.Name == name {
			log.Printf("Retrieved the architect schedules id %s by name %s", *schedules.Id, name)
			return *schedules.Id, false, apiResponse, nil
		}
	}

	return "", true, apiResponse, fmt.Errorf("Unable to find architect schedules with name %s", name)
}

// getArchitectSchedulesByIdFn is an implementation of the function to get a Genesys Cloud architect schedules by Id
func getArchitectSchedulesByIdFn(ctx context.Context, p *architectSchedulesProxy, id string) (architectSchedules *platformclientv2.Schedule, response *platformclientv2.APIResponse, err error) {
	schedule, apiResponse, err := p.architectApi.GetArchitectSchedule(id)
	if err != nil {
		return nil, apiResponse, fmt.Errorf("Failed to retrieve architect schedule by id %s: %s", id, err)
	}
	return schedule, apiResponse, nil
}

// updateArchitectSchedulesFn is an implementation of the function to update a Genesys Cloud architect schedules
func updateArchitectSchedulesFn(ctx context.Context, p *architectSchedulesProxy, id string, architectSchedules *platformclientv2.Schedule) (*platformclientv2.Schedule, *platformclientv2.APIResponse, error) {
	schedule, apiResponse, err := getArchitectSchedulesByIdFn(ctx, p, id)
	if err != nil {
		return nil, apiResponse, fmt.Errorf("Failed to get schedule  %s by id: %s", id, err)
	}
	architectSchedules.Version = schedule.Version
	scheduleResponse, apiResponse, err := p.architectApi.PutArchitectSchedule(id, *architectSchedules)
	if err != nil {
		return nil, apiResponse, fmt.Errorf("Failed to update architect schedules: %s", err)
	}
	return scheduleResponse, apiResponse, nil
}

// deleteArchitectSchedulesFn is an implementation function for deleting a Genesys Cloud architect schedules
func deleteArchitectSchedulesFn(ctx context.Context, p *architectSchedulesProxy, id string) (*platformclientv2.APIResponse, error) {
	resp, err := p.architectApi.DeleteArchitectSchedule(id)
	if err != nil {
		return resp, fmt.Errorf("Failed to delete architect schedules: %s", err)
	}
	return resp, nil
}
