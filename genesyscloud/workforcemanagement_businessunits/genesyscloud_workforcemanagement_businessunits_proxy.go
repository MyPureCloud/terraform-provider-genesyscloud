package workforcemanagement_businessunits

import (
	"context"
	"fmt"
	"log"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"

	"github.com/mypurecloud/platform-client-sdk-go/v176/platformclientv2"
)

/*
The genesyscloud_workforcemanagement_businessunits_proxy.go file contains the proxy structures and methods that interact
with the Genesys Cloud SDK. We use composition here for each function on the proxy so individual functions can be stubbed
out during testing.
*/

// internalProxy is a proxy instance that can be used throughout the package.
var internalProxy *workforceManagementBusinessUnitsProxy

// Type definitions for each func on our proxy so we can easily mock them out later
type createWorkforceManagementBusinessUnitFunc func(ctx context.Context, p *workforceManagementBusinessUnitsProxy, businessUnitResponse *platformclientv2.Createbusinessunitrequest) (*platformclientv2.Businessunitresponse, *platformclientv2.APIResponse, error)
type getAllWorkforceManagementBusinessUnitsFunc func(ctx context.Context, p *workforceManagementBusinessUnitsProxy) (*[]platformclientv2.Businessunitlistitem, *platformclientv2.APIResponse, error)
type getWorkforceManagementBusinessUnitIdByExactNameFunc func(ctx context.Context, p *workforceManagementBusinessUnitsProxy, name string) (string, *platformclientv2.APIResponse, bool, error)
type getWorkforceManagementBusinessUnitByIdFunc func(ctx context.Context, p *workforceManagementBusinessUnitsProxy, id string) (*platformclientv2.Businessunitresponse, *platformclientv2.APIResponse, error)
type updateWorkforceManagementBusinessUnitFunc func(ctx context.Context, p *workforceManagementBusinessUnitsProxy, id string, businessUnitResponse *platformclientv2.Updatebusinessunitrequest) (*platformclientv2.Businessunitresponse, *platformclientv2.APIResponse, error)
type deleteWorkforceManagementBusinessUnitFunc func(ctx context.Context, p *workforceManagementBusinessUnitsProxy, id string) (*platformclientv2.APIResponse, error)

// workforceManagementBusinessUnitsProxy contains all the methods that call genesys cloud APIs.
type workforceManagementBusinessUnitsProxy struct {
	clientConfig                                        *platformclientv2.Configuration
	workforceManagementApi                              *platformclientv2.WorkforceManagementApi
	createWorkforceManagementBusinessUnitAttr           createWorkforceManagementBusinessUnitFunc
	getAllWorkforceManagementBusinessUnitsAttr          getAllWorkforceManagementBusinessUnitsFunc
	getWorkforceManagementBusinessUnitIdByExactNameAttr getWorkforceManagementBusinessUnitIdByExactNameFunc
	getWorkforceManagementBusinessUnitByIdAttr          getWorkforceManagementBusinessUnitByIdFunc
	updateWorkforceManagementBusinessUnitAttr           updateWorkforceManagementBusinessUnitFunc
	deleteWorkforceManagementBusinessUnitAttr           deleteWorkforceManagementBusinessUnitFunc
}

// newWorkforceManagementBusinessUnitsProxy initializes the workforce management business units proxy with all the data needed to communicate with Genesys Cloud
func newWorkforceManagementBusinessUnitsProxy(clientConfig *platformclientv2.Configuration) *workforceManagementBusinessUnitsProxy {
	api := platformclientv2.NewWorkforceManagementApiWithConfig(clientConfig)
	return &workforceManagementBusinessUnitsProxy{
		clientConfig:           clientConfig,
		workforceManagementApi: api,
		createWorkforceManagementBusinessUnitAttr:           createWorkforceManagementBusinessUnitFn,
		getAllWorkforceManagementBusinessUnitsAttr:          getAllWorkforceManagementBusinessUnitsFn,
		getWorkforceManagementBusinessUnitIdByExactNameAttr: getWorkforceManagementBusinessUnitIdByExactNameFn,
		getWorkforceManagementBusinessUnitByIdAttr:          getWorkforceManagementBusinessUnitByIdFn,
		updateWorkforceManagementBusinessUnitAttr:           updateWorkforceManagementBusinessUnitFn,
		deleteWorkforceManagementBusinessUnitAttr:           deleteWorkforceManagementBusinessUnitFn,
	}
}

// getWorkforceManagementBusinessUnitsProxy acts as a singleton to for the internalProxy.  It also ensures
// that we can still proxy our tests by directly setting internalProxy package variable
func getWorkforceManagementBusinessUnitsProxy(clientConfig *platformclientv2.Configuration) *workforceManagementBusinessUnitsProxy {
	if internalProxy == nil {
		internalProxy = newWorkforceManagementBusinessUnitsProxy(clientConfig)
	}

	return internalProxy
}

// createWorkforceManagementBusinessUnit creates a Genesys Cloud workforce management business unit
func (p *workforceManagementBusinessUnitsProxy) createWorkforceManagementBusinessUnit(ctx context.Context, createBuRequest *platformclientv2.Createbusinessunitrequest) (*platformclientv2.Businessunitresponse, *platformclientv2.APIResponse, error) {
	return p.createWorkforceManagementBusinessUnitAttr(ctx, p, createBuRequest)
}

// getAllWorkforceManagementBusinessUnits retrieves all Genesys Cloud workforce management business units
func (p *workforceManagementBusinessUnitsProxy) getAllWorkforceManagementBusinessUnits(ctx context.Context) (*[]platformclientv2.Businessunitlistitem, *platformclientv2.APIResponse, error) {
	return p.getAllWorkforceManagementBusinessUnitsAttr(ctx, p)
}

// getWorkforceManagementBusinessUnitIdByExactName returns a single Genesys Cloud workforce management business unit ID by exact name match
func (p *workforceManagementBusinessUnitsProxy) getWorkforceManagementBusinessUnitIdByExactName(ctx context.Context, name string) (string, *platformclientv2.APIResponse, bool, error) {
	return p.getWorkforceManagementBusinessUnitIdByExactNameAttr(ctx, p, name)
}

// getWorkforceManagementBusinessUnitById returns a single Genesys Cloud workforce management business unit by ID
func (p *workforceManagementBusinessUnitsProxy) getWorkforceManagementBusinessUnitById(ctx context.Context, id string) (*platformclientv2.Businessunitresponse, *platformclientv2.APIResponse, error) {
	return p.getWorkforceManagementBusinessUnitByIdAttr(ctx, p, id)
}

// updateWorkforceManagementBusinessUnit updates a Genesys Cloud workforce management business unit
func (p *workforceManagementBusinessUnitsProxy) updateWorkforceManagementBusinessUnit(ctx context.Context, id string, updateRequest *platformclientv2.Updatebusinessunitrequest) (*platformclientv2.Businessunitresponse, *platformclientv2.APIResponse, error) {
	return p.updateWorkforceManagementBusinessUnitAttr(ctx, p, id, updateRequest)
}

// deleteWorkforceManagementBusinessUnit deletes a Genesys Cloud workforce management business unit by Id
func (p *workforceManagementBusinessUnitsProxy) deleteWorkforceManagementBusinessUnit(ctx context.Context, id string) (*platformclientv2.APIResponse, error) {
	return p.deleteWorkforceManagementBusinessUnitAttr(ctx, p, id)
}

// createWorkforceManagementBusinessUnitFn is an implementation function for creating a Genesys Cloud workforce management business unit
func createWorkforceManagementBusinessUnitFn(ctx context.Context, p *workforceManagementBusinessUnitsProxy, workforceManagementBusinessUnits *platformclientv2.Createbusinessunitrequest) (*platformclientv2.Businessunitresponse, *platformclientv2.APIResponse, error) {
	ctx = provider.EnsureResourceContext(ctx, ResourceType)
	return p.workforceManagementApi.PostWorkforcemanagementBusinessunits(*workforceManagementBusinessUnits, false)
}

// getAllWorkforceManagementBusinessUnitsFn is the implementation for retrieving all workforce management business units in Genesys Cloud
func getAllWorkforceManagementBusinessUnitsFn(ctx context.Context, p *workforceManagementBusinessUnitsProxy) (*[]platformclientv2.Businessunitlistitem, *platformclientv2.APIResponse, error) {
	ctx = provider.EnsureResourceContext(ctx, ResourceType)
	// Empty strings are for feature (so no special permission checking overrides) and divisionId (so no filtering by divisionId)
	businessUnitResponses, resp, err := p.workforceManagementApi.GetWorkforcemanagementBusinessunits("", "")
	if err != nil {
		return nil, resp, err
	}

	return businessUnitResponses.Entities, resp, nil
}

// getWorkforceManagementBusinessUnitIdByExactNameFn is an implementation of the function to get a Genesys Cloud workforce management business unit ID by exact name match
func getWorkforceManagementBusinessUnitIdByExactNameFn(ctx context.Context, p *workforceManagementBusinessUnitsProxy, name string) (string, *platformclientv2.APIResponse, bool, error) {
	ctx = provider.EnsureResourceContext(ctx, ResourceType)
	// Empty strings for feature and divisionId parameters mean no filtering - retrieve all business units to search by name
	businessUnitResponses, resp, err := p.workforceManagementApi.GetWorkforcemanagementBusinessunits("", "")
	if err != nil {
		return "", resp, false, err
	}

	if businessUnitResponses.Entities == nil || len(*businessUnitResponses.Entities) == 0 {
		return "", resp, true, err
	}

	for _, businessUnitResponse := range *businessUnitResponses.Entities {
		if *businessUnitResponse.Name == name {
			log.Printf("Retrieved the workforce management business unit id %s by name %s", *businessUnitResponse.Id, name)
			return *businessUnitResponse.Id, resp, false, nil
		}
	}

	return "", resp, true, fmt.Errorf("unable to find workforce management business unit with name %s", name)
}

// getWorkforceManagementBusinessUnitByIdFn is an implementation of the function to get a Genesys Cloud workforce management business unit by ID
func getWorkforceManagementBusinessUnitByIdFn(ctx context.Context, p *workforceManagementBusinessUnitsProxy, id string) (*platformclientv2.Businessunitresponse, *platformclientv2.APIResponse, error) {
	ctx = provider.EnsureResourceContext(ctx, ResourceType)
	return p.workforceManagementApi.GetWorkforcemanagementBusinessunit(id, []string{"settings"}, false)
}

// updateWorkforceManagementBusinessUnitFn is an implementation of the function to update a Genesys Cloud workforce management business unit
func updateWorkforceManagementBusinessUnitFn(ctx context.Context, p *workforceManagementBusinessUnitsProxy, id string, updateRequest *platformclientv2.Updatebusinessunitrequest) (*platformclientv2.Businessunitresponse, *platformclientv2.APIResponse, error) {
	ctx = provider.EnsureResourceContext(ctx, ResourceType)
	return p.workforceManagementApi.PatchWorkforcemanagementBusinessunit(id, *updateRequest, false)
}

// deleteWorkforceManagementBusinessUnitFn is an implementation function for deleting a Genesys Cloud workforce management business unit
func deleteWorkforceManagementBusinessUnitFn(ctx context.Context, p *workforceManagementBusinessUnitsProxy, id string) (*platformclientv2.APIResponse, error) {
	ctx = provider.EnsureResourceContext(ctx, ResourceType)
	return p.workforceManagementApi.DeleteWorkforcemanagementBusinessunit(id)
}
