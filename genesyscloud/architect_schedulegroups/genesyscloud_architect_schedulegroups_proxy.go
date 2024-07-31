package architect_schedulegroups

import (
	"context"
	"fmt"
	"log"

	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

/*
The genesyscloud_architect_schedulegroups_proxy.go file contains the proxy structures and methods that interact
with the Genesys Cloud SDK. We use composition here for each function on the proxy so individual functions can be stubbed
out during testing.
*/

// internalProxy holds a proxy instance that can be used throughout the package
var internalProxy *architectSchedulegroupsProxy

// Type definitions for each func on our proxy so we can easily mock them out later
type createArchitectSchedulegroupsFunc func(ctx context.Context, p *architectSchedulegroupsProxy, scheduleGroup *platformclientv2.Schedulegroup) (*platformclientv2.Schedulegroup, *platformclientv2.APIResponse, error)
type getAllArchitectSchedulegroupsFunc func(ctx context.Context, p *architectSchedulegroupsProxy) (*[]platformclientv2.Schedulegroup, *platformclientv2.APIResponse, error)
type getArchitectSchedulegroupsIdByNameFunc func(ctx context.Context, p *architectSchedulegroupsProxy, name string) (id string, retryable bool, response *platformclientv2.APIResponse, err error)
type getArchitectSchedulegroupsByIdFunc func(ctx context.Context, p *architectSchedulegroupsProxy, id string) (scheduleGroup *platformclientv2.Schedulegroup, response *platformclientv2.APIResponse, err error)
type updateArchitectSchedulegroupsFunc func(ctx context.Context, p *architectSchedulegroupsProxy, id string, scheduleGroup *platformclientv2.Schedulegroup) (*platformclientv2.Schedulegroup, *platformclientv2.APIResponse, error)
type deleteArchitectSchedulegroupsFunc func(ctx context.Context, p *architectSchedulegroupsProxy, id string) (*platformclientv2.APIResponse, error)

// architectSchedulegroupsProxy contains all of the methods that call genesys cloud APIs.
type architectSchedulegroupsProxy struct {
	clientConfig                           *platformclientv2.Configuration
	architectApi                           *platformclientv2.ArchitectApi
	createArchitectSchedulegroupsAttr      createArchitectSchedulegroupsFunc
	getAllArchitectSchedulegroupsAttr      getAllArchitectSchedulegroupsFunc
	getArchitectSchedulegroupsIdByNameAttr getArchitectSchedulegroupsIdByNameFunc
	getArchitectSchedulegroupsByIdAttr     getArchitectSchedulegroupsByIdFunc
	updateArchitectSchedulegroupsAttr      updateArchitectSchedulegroupsFunc
	deleteArchitectSchedulegroupsAttr      deleteArchitectSchedulegroupsFunc
}

// newArchitectSchedulegroupsProxy initializes the architect schedulegroups proxy with all of the data needed to communicate with Genesys Cloud
func newArchitectSchedulegroupsProxy(clientConfig *platformclientv2.Configuration) *architectSchedulegroupsProxy {
	api := platformclientv2.NewArchitectApiWithConfig(clientConfig)
	return &architectSchedulegroupsProxy{
		clientConfig:                           clientConfig,
		architectApi:                           api,
		createArchitectSchedulegroupsAttr:      createArchitectSchedulegroupsFn,
		getAllArchitectSchedulegroupsAttr:      getAllArchitectSchedulegroupsFn,
		getArchitectSchedulegroupsIdByNameAttr: getArchitectSchedulegroupsIdByNameFn,
		getArchitectSchedulegroupsByIdAttr:     getArchitectSchedulegroupsByIdFn,
		updateArchitectSchedulegroupsAttr:      updateArchitectSchedulegroupsFn,
		deleteArchitectSchedulegroupsAttr:      deleteArchitectSchedulegroupsFn,
	}
}

// getArchitectSchedulegroupsProxy acts as a singleton to for the internalProxy.  It also ensures
// that we can still proxy our tests by directly setting internalProxy package variable
func getArchitectSchedulegroupsProxy(clientConfig *platformclientv2.Configuration) *architectSchedulegroupsProxy {
	if internalProxy == nil {
		internalProxy = newArchitectSchedulegroupsProxy(clientConfig)
	}
	return internalProxy
}

// createArchitectSchedulegroups creates a Genesys Cloud architect schedulegroups
func (p *architectSchedulegroupsProxy) createArchitectSchedulegroups(ctx context.Context, architectSchedulegroups *platformclientv2.Schedulegroup) (*platformclientv2.Schedulegroup, *platformclientv2.APIResponse, error) {
	return p.createArchitectSchedulegroupsAttr(ctx, p, architectSchedulegroups)
}

// getArchitectSchedulegroups retrieves all Genesys Cloud architect schedulegroups
func (p *architectSchedulegroupsProxy) getAllArchitectSchedulegroups(ctx context.Context) (*[]platformclientv2.Schedulegroup, *platformclientv2.APIResponse, error) {
	return p.getAllArchitectSchedulegroupsAttr(ctx, p)
}

// getArchitectSchedulegroupsIdByName returns a single Genesys Cloud architect schedulegroups by a name
func (p *architectSchedulegroupsProxy) getArchitectSchedulegroupsIdByName(ctx context.Context, name string) (id string, retryable bool, response *platformclientv2.APIResponse, err error) {
	return p.getArchitectSchedulegroupsIdByNameAttr(ctx, p, name)
}

// getArchitectSchedulegroupsById returns a single Genesys Cloud architect schedulegroups by Id
func (p *architectSchedulegroupsProxy) getArchitectSchedulegroupsById(ctx context.Context, id string) (architectSchedulegroups *platformclientv2.Schedulegroup, response *platformclientv2.APIResponse, err error) {
	return p.getArchitectSchedulegroupsByIdAttr(ctx, p, id)
}

// updateArchitectSchedulegroups updates a Genesys Cloud architect schedulegroups
func (p *architectSchedulegroupsProxy) updateArchitectSchedulegroups(ctx context.Context, id string, architectSchedulegroups *platformclientv2.Schedulegroup) (*platformclientv2.Schedulegroup, *platformclientv2.APIResponse, error) {
	return p.updateArchitectSchedulegroupsAttr(ctx, p, id, architectSchedulegroups)
}

// deleteArchitectSchedulegroups deletes a Genesys Cloud architect schedulegroups by Id
func (p *architectSchedulegroupsProxy) deleteArchitectSchedulegroups(ctx context.Context, id string) (*platformclientv2.APIResponse, error) {
	return p.deleteArchitectSchedulegroupsAttr(ctx, p, id)
}

// createArchitectSchedulegroupsFn is an implementation function for creating a Genesys Cloud architect schedulegroups
func createArchitectSchedulegroupsFn(ctx context.Context, p *architectSchedulegroupsProxy, architectSchedulegroups *platformclientv2.Schedulegroup) (*platformclientv2.Schedulegroup, *platformclientv2.APIResponse, error) {
	scheduleGroup, apiResponse, err := p.architectApi.PostArchitectSchedulegroups(*architectSchedulegroups)
	if err != nil {
		return nil, apiResponse, fmt.Errorf("Failed to create architect schedulegroups: %s", err)
	}
	return scheduleGroup, apiResponse, nil
}

// getAllArchitectSchedulegroupsFn is the implementation for retrieving all architect schedulegroups in Genesys Cloud
func getAllArchitectSchedulegroupsFn(ctx context.Context, p *architectSchedulegroupsProxy) (*[]platformclientv2.Schedulegroup, *platformclientv2.APIResponse, error) {
	var allScheduleGroups []platformclientv2.Schedulegroup
	const pageSize = 100

	scheduleGroups, apiResponse, err := p.architectApi.GetArchitectSchedulegroups(1, pageSize, "", "", "", "", nil)
	if err != nil {
		return nil, apiResponse, fmt.Errorf("Failed to get schedule group: %v", err)
	}
	if scheduleGroups.Entities == nil || len(*scheduleGroups.Entities) == 0 {
		return &allScheduleGroups, apiResponse, nil
	}
	for _, scheduleGroup := range *scheduleGroups.Entities {
		allScheduleGroups = append(allScheduleGroups, scheduleGroup)
	}

	for pageNum := 2; pageNum <= *scheduleGroups.PageCount; pageNum++ {
		scheduleGroups, apiResponse, err := p.architectApi.GetArchitectSchedulegroups(pageNum, pageSize, "", "", "", "", nil)
		if err != nil {
			return nil, apiResponse, fmt.Errorf("Failed to get schedule group: %v", err)
		}

		if scheduleGroups.Entities == nil || len(*scheduleGroups.Entities) == 0 {
			break
		}

		for _, scheduleGroup := range *scheduleGroups.Entities {
			allScheduleGroups = append(allScheduleGroups, scheduleGroup)
		}
	}
	return &allScheduleGroups, apiResponse, nil
}

// getArchitectSchedulegroupsIdByNameFn is an implementation of the function to get a Genesys Cloud architect schedulegroups by name
func getArchitectSchedulegroupsIdByNameFn(ctx context.Context, p *architectSchedulegroupsProxy, name string) (id string, retryable bool, response *platformclientv2.APIResponse, err error) {
	scheduleGroups, apiResponse, err := getAllArchitectSchedulegroupsFn(ctx, p)
	if err != nil {
		return "", false, apiResponse, err
	}

	if scheduleGroups == nil || len(*scheduleGroups) == 0 {
		return "", true, apiResponse, fmt.Errorf("No architect schedulegroups found with name %s", name)
	}

	for _, scheduleGroup := range *scheduleGroups {
		if *scheduleGroup.Name == name {
			log.Printf("Retrieved the architect schedulegroups id %s by name %s", *scheduleGroup.Id, name)
			return *scheduleGroup.Id, false, apiResponse, nil
		}
	}
	return "", true, apiResponse, fmt.Errorf("Unable to find architect schedulegroups with name %s", name)
}

// getArchitectSchedulegroupsByIdFn is an implementation of the function to get a Genesys Cloud architect schedulegroups by Id
func getArchitectSchedulegroupsByIdFn(ctx context.Context, p *architectSchedulegroupsProxy, id string) (architectSchedulegroups *platformclientv2.Schedulegroup, response *platformclientv2.APIResponse, err error) {
	scheduleGroup, apiResponse, err := p.architectApi.GetArchitectSchedulegroup(id)
	if err != nil {
		return nil, apiResponse, fmt.Errorf("Failed to retrieve architect schedulegroups by id %s: %s", id, err)
	}
	return scheduleGroup, apiResponse, nil
}

// updateArchitectSchedulegroupsFn is an implementation of the function to update a Genesys Cloud architect schedulegroups
func updateArchitectSchedulegroupsFn(ctx context.Context, p *architectSchedulegroupsProxy, id string, architectSchedulegroups *platformclientv2.Schedulegroup) (*platformclientv2.Schedulegroup, *platformclientv2.APIResponse, error) {
	group, apiResponse, err := getArchitectSchedulegroupsByIdFn(ctx, p, id)
	if err != nil {
		return nil, apiResponse, fmt.Errorf("failed to get schedule group %s by id: %s", id, err)
	}
	architectSchedulegroups.Version = group.Version
	scheduleGroup, apiResponse, err := p.architectApi.PutArchitectSchedulegroup(id, *architectSchedulegroups)
	if err != nil {
		return nil, apiResponse, fmt.Errorf("Failed to update architect schedulegroups: %s", err)
	}
	return scheduleGroup, apiResponse, nil
}

// deleteArchitectSchedulegroupsFn is an implementation function for deleting a Genesys Cloud architect schedulegroups
func deleteArchitectSchedulegroupsFn(ctx context.Context, p *architectSchedulegroupsProxy, id string) (*platformclientv2.APIResponse, error) {
	resp, err := p.architectApi.DeleteArchitectSchedulegroup(id)
	if err != nil {
		return resp, fmt.Errorf("Failed to delete architect schedulegroups: %s", err)
	}
	return resp, nil
}
