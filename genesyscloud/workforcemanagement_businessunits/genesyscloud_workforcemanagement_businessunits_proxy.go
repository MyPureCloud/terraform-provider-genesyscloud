package workforcemanagement_businessunits

import (
	"context"
	"fmt"
	"github.com/mypurecloud/platform-client-sdk-go/v165/platformclientv2"
	"log"
)

/*
The genesyscloud_workforcemanagement_businessunits_proxy.go file contains the proxy structures and methods that interact
with the Genesys Cloud SDK. We use composition here for each function on the proxy so individual functions can be stubbed
out during testing.
*/

// internalProxy holds a proxy instance that can be used throughout the package
var internalProxy *workforcemanagementBusinessunitsProxy

// Type definitions for each func on our proxy so we can easily mock them out later
type createWorkforcemanagementBusinessunitsFunc func(ctx context.Context, p *workforcemanagementBusinessunitsProxy, businessUnitResponse *platformclientv2.Createbusinessunitrequest) (*platformclientv2.Businessunitresponse, *platformclientv2.APIResponse, error)
type getAllWorkforcemanagementBusinessunitsFunc func(ctx context.Context, p *workforcemanagementBusinessunitsProxy) (*[]platformclientv2.Businessunitlistitem, *platformclientv2.APIResponse, error)
type getWorkforcemanagementBusinessunitsIdByNameFunc func(ctx context.Context, p *workforcemanagementBusinessunitsProxy, name string) (string, *platformclientv2.APIResponse, bool, error)
type getWorkforcemanagementBusinessunitsByIdFunc func(ctx context.Context, p *workforcemanagementBusinessunitsProxy, id string) (*platformclientv2.Businessunitresponse, *platformclientv2.APIResponse, error)
type updateWorkforcemanagementBusinessunitsFunc func(ctx context.Context, p *workforcemanagementBusinessunitsProxy, id string, businessUnitResponse *platformclientv2.Updatebusinessunitrequest) (*platformclientv2.Businessunitresponse, *platformclientv2.APIResponse, error)
type deleteWorkforcemanagementBusinessunitsFunc func(ctx context.Context, p *workforcemanagementBusinessunitsProxy, id string) (*platformclientv2.APIResponse, error)

// workforcemanagementBusinessunitsProxy contains all the methods that call genesys cloud APIs.
type workforcemanagementBusinessunitsProxy struct {
	clientConfig                                    *platformclientv2.Configuration
	workforceManagementApi                          *platformclientv2.WorkforceManagementApi
	createWorkforcemanagementBusinessunitsAttr      createWorkforcemanagementBusinessunitsFunc
	getAllWorkforcemanagementBusinessunitsAttr      getAllWorkforcemanagementBusinessunitsFunc
	getWorkforcemanagementBusinessunitsIdByNameAttr getWorkforcemanagementBusinessunitsIdByNameFunc
	getWorkforcemanagementBusinessunitsByIdAttr     getWorkforcemanagementBusinessunitsByIdFunc
	updateWorkforcemanagementBusinessunitsAttr      updateWorkforcemanagementBusinessunitsFunc
	deleteWorkforcemanagementBusinessunitsAttr      deleteWorkforcemanagementBusinessunitsFunc
}

// newWorkforcemanagementBusinessunitsProxy initializes the workforcemanagement businessunits proxy with all the data needed to communicate with Genesys Cloud
func newWorkforcemanagementBusinessunitsProxy(clientConfig *platformclientv2.Configuration) *workforcemanagementBusinessunitsProxy {
	api := platformclientv2.NewWorkforceManagementApiWithConfig(clientConfig)
	return &workforcemanagementBusinessunitsProxy{
		clientConfig:           clientConfig,
		workforceManagementApi: api,
		createWorkforcemanagementBusinessunitsAttr:      createWorkforcemanagementBusinessunitsFn,
		getAllWorkforcemanagementBusinessunitsAttr:      getAllWorkforcemanagementBusinessunitsFn,
		getWorkforcemanagementBusinessunitsIdByNameAttr: getWorkforcemanagementBusinessunitsIdByNameFn,
		getWorkforcemanagementBusinessunitsByIdAttr:     getWorkforcemanagementBusinessunitsByIdFn,
		updateWorkforcemanagementBusinessunitsAttr:      updateWorkforcemanagementBusinessunitsFn,
		deleteWorkforcemanagementBusinessunitsAttr:      deleteWorkforcemanagementBusinessunitsFn,
	}
}

// getWorkforcemanagementBusinessunitsProxy acts as a singleton to for the internalProxy.  It also ensures
// that we can still proxy our tests by directly setting internalProxy package variable
func getWorkforcemanagementBusinessunitsProxy(clientConfig *platformclientv2.Configuration) *workforcemanagementBusinessunitsProxy {
	if internalProxy == nil {
		internalProxy = newWorkforcemanagementBusinessunitsProxy(clientConfig)
	}

	return internalProxy
}

// createWorkforcemanagementBusinessUnit creates a Genesys Cloud workforcemanagement businessunits
func (p *workforcemanagementBusinessunitsProxy) createWorkforcemanagementBusinessunits(ctx context.Context, createBuRequest *platformclientv2.Createbusinessunitrequest) (*platformclientv2.Businessunitresponse, *platformclientv2.APIResponse, error) {
	return p.createWorkforcemanagementBusinessunitsAttr(ctx, p, createBuRequest)
}

// getWorkforcemanagementBusinessunits retrieves all Genesys Cloud workforcemanagement businessunits
func (p *workforcemanagementBusinessunitsProxy) getAllWorkforcemanagementBusinessunits(ctx context.Context) (*[]platformclientv2.Businessunitlistitem, *platformclientv2.APIResponse, error) {
	return p.getAllWorkforcemanagementBusinessunitsAttr(ctx, p)
}

// getWorkforcemanagementBusinessunitsIdByName returns a single Genesys Cloud workforcemanagement businessunits by a name
func (p *workforcemanagementBusinessunitsProxy) getWorkforcemanagementBusinessunitsIdByName(ctx context.Context, name string) (string, *platformclientv2.APIResponse, bool, error) {
	return p.getWorkforcemanagementBusinessunitsIdByNameAttr(ctx, p, name)
}

// getWorkforcemanagementBusinessunitsById returns a single Genesys Cloud workforcemanagement businessunits by ID
func (p *workforcemanagementBusinessunitsProxy) getWorkforcemanagementBusinessunitsById(ctx context.Context, id string) (*platformclientv2.Businessunitresponse, *platformclientv2.APIResponse, error) {
	return p.getWorkforcemanagementBusinessunitsByIdAttr(ctx, p, id)
}

// updateWorkforcemanagementBusinessunits updates a Genesys Cloud workforcemanagement businessunits
func (p *workforcemanagementBusinessunitsProxy) updateWorkforcemanagementBusinessunits(ctx context.Context, id string, updateRequest *platformclientv2.Updatebusinessunitrequest) (*platformclientv2.Businessunitresponse, *platformclientv2.APIResponse, error) {
	return p.updateWorkforcemanagementBusinessunitsAttr(ctx, p, id, updateRequest)
}

// deleteWorkforcemanagementBusinessunits deletes a Genesys Cloud workforcemanagement businessunits by Id
func (p *workforcemanagementBusinessunitsProxy) deleteWorkforcemanagementBusinessunits(ctx context.Context, id string) (*platformclientv2.APIResponse, error) {
	return p.deleteWorkforcemanagementBusinessunitsAttr(ctx, p, id)
}

// createWorkforcemanagementBusinessunitsFn is an implementation function for creating a Genesys Cloud workforcemanagement businessunits
func createWorkforcemanagementBusinessunitsFn(ctx context.Context, p *workforcemanagementBusinessunitsProxy, workforcemanagementBusinessunits *platformclientv2.Createbusinessunitrequest) (*platformclientv2.Businessunitresponse, *platformclientv2.APIResponse, error) {
	return p.workforceManagementApi.PostWorkforcemanagementBusinessunits(*workforcemanagementBusinessunits, false)
}

// getAllWorkforcemanagementBusinessunitsFn is the implementation for retrieving all workforcemanagement businessunits in Genesys Cloud
func getAllWorkforcemanagementBusinessunitsFn(ctx context.Context, p *workforcemanagementBusinessunitsProxy) (*[]platformclientv2.Businessunitlistitem, *platformclientv2.APIResponse, error) {
	businessUnitResponses, resp, err := p.workforceManagementApi.GetWorkforcemanagementBusinessunits("", "")
	if err != nil {
		return nil, resp, err
	}

	return businessUnitResponses.Entities, resp, nil
}

// getWorkforcemanagementBusinessunitsIdByNameFn is an implementation of the function to get a Genesys Cloud workforcemanagement businessunits by name
func getWorkforcemanagementBusinessunitsIdByNameFn(ctx context.Context, p *workforcemanagementBusinessunitsProxy, name string) (string, *platformclientv2.APIResponse, bool, error) {
	businessUnitResponses, resp, err := p.workforceManagementApi.GetWorkforcemanagementBusinessunits("", "")
	if err != nil {
		return "", resp, false, err
	}

	if businessUnitResponses.Entities == nil || len(*businessUnitResponses.Entities) == 0 {
		return "", resp, true, err
	}

	for _, businessUnitResponse := range *businessUnitResponses.Entities {
		if *businessUnitResponse.Name == name {
			log.Printf("Retrieved the workforcemanagement businessunits id %s by name %s", *businessUnitResponse.Id, name)
			return *businessUnitResponse.Id, resp, false, nil
		}
	}

	return "", resp, true, fmt.Errorf("unable to find workforcemanagement businessunits with name %s", name)
}

// getWorkforcemanagementBusinessunitsByIdFn is an implementation of the function to get a Genesys Cloud workforcemanagement businessunits by ID
func getWorkforcemanagementBusinessunitsByIdFn(ctx context.Context, p *workforcemanagementBusinessunitsProxy, id string) (*platformclientv2.Businessunitresponse, *platformclientv2.APIResponse, error) {
	return p.workforceManagementApi.GetWorkforcemanagementBusinessunit(id, []string{"settings"}, false)
}

// updateWorkforcemanagementBusinessunitsFn is an implementation of the function to update a Genesys Cloud workforcemanagement businessunits
func updateWorkforcemanagementBusinessunitsFn(ctx context.Context, p *workforcemanagementBusinessunitsProxy, id string, updateRequest *platformclientv2.Updatebusinessunitrequest) (*platformclientv2.Businessunitresponse, *platformclientv2.APIResponse, error) {
	return p.workforceManagementApi.PatchWorkforcemanagementBusinessunit(id, *updateRequest, false)
}

// deleteWorkforcemanagementBusinessunitsFn is an implementation function for deleting a Genesys Cloud workforcemanagement businessunits
func deleteWorkforcemanagementBusinessunitsFn(ctx context.Context, p *workforcemanagementBusinessunitsProxy, id string) (*platformclientv2.APIResponse, error) {
	return p.workforceManagementApi.DeleteWorkforcemanagementBusinessunit(id)
}
