package workforcemanagement_businessunits_activitycode

import (
	"context"

	"github.com/mypurecloud/platform-client-sdk-go/v200/platformclientv2"
)

/*
The genesyscloud_workforcemanagement_businessunits_activitycode_proxy.go file contains the proxy structures and methods that interact
with the Genesys Cloud SDK. We use composition here for each function on the proxy so individual functions can be stubbed
out during testing.
*/

// internalProxy holds a proxy instance that can be used throughout the package
var internalProxy *businessUnitActivityCodeProxy

// Type definitions for each func on our proxy so we can easily mock them out later
type createBusinessUnitActivityCodeFunc func(ctx context.Context, p *businessUnitActivityCodeProxy, businessUnitId string, activityCode *platformclientv2.Createactivitycoderequest) (*platformclientv2.Businessunitactivitycode, *platformclientv2.APIResponse, error)
type getAllBusinessUnitActivityCodesFunc func(ctx context.Context, p *businessUnitActivityCodeProxy, businessUnitId string) (*[]platformclientv2.Businessunitactivitycode, *platformclientv2.APIResponse, error)
type getBusinessUnitActivityCodeIdByNameFunc func(ctx context.Context, p *businessUnitActivityCodeProxy, businessUnitId string, name string) (string, *platformclientv2.APIResponse, error)
type getDefaultBusinessUnitActivityCodeIdByCategoryFunc func(ctx context.Context, p *businessUnitActivityCodeProxy, businessUnitId string, category string) (string, *platformclientv2.APIResponse, error)
type getBusinessUnitActivityCodeByIdFunc func(ctx context.Context, p *businessUnitActivityCodeProxy, businessUnitId string, activityCodeId string) (*platformclientv2.Businessunitactivitycode, *platformclientv2.APIResponse, error)
type updateBusinessUnitActivityCodeFunc func(ctx context.Context, p *businessUnitActivityCodeProxy, businessUnitId string, activityCodeId string, activityCode *platformclientv2.Updateactivitycoderequest) (*platformclientv2.Businessunitactivitycode, *platformclientv2.APIResponse, error)
type deleteBusinessUnitActivityCodeFunc func(ctx context.Context, p *businessUnitActivityCodeProxy, businessUnitId string, activityCodeId string) (*platformclientv2.APIResponse, error)
type getAllBusinessUnitsFunc func(ctx context.Context, p *businessUnitActivityCodeProxy) (*[]platformclientv2.Businessunitlistitem, *platformclientv2.APIResponse, error)

// businessUnitActivityCodeProxy contains all the methods that call genesys cloud APIs.
type businessUnitActivityCodeProxy struct {
	clientConfig                                       *platformclientv2.Configuration
	workforceManagementApi                             *platformclientv2.WorkforceManagementApi
	createBusinessUnitActivityCodeAttr                 createBusinessUnitActivityCodeFunc
	getAllBusinessUnitActivityCodesAttr                getAllBusinessUnitActivityCodesFunc
	getBusinessUnitActivityCodeIdByNameAttr            getBusinessUnitActivityCodeIdByNameFunc
	getDefaultBusinessUnitActivityCodeIdByCategoryAttr getDefaultBusinessUnitActivityCodeIdByCategoryFunc
	getBusinessUnitActivityCodeByIdAttr                getBusinessUnitActivityCodeByIdFunc
	updateBusinessUnitActivityCodeAttr                 updateBusinessUnitActivityCodeFunc
	deleteBusinessUnitActivityCodeAttr                 deleteBusinessUnitActivityCodeFunc
	getAllBusinessUnitsAttr                            getAllBusinessUnitsFunc
}

// newBusinessUnitActivityCodeProxy initializes the business unit activity code proxy with all of the data needed to communicate with Genesys Cloud
func newBusinessUnitActivityCodeProxy(clientConfig *platformclientv2.Configuration) *businessUnitActivityCodeProxy {
	api := platformclientv2.NewWorkforceManagementApiWithConfig(clientConfig)
	return &businessUnitActivityCodeProxy{
		clientConfig:                                       clientConfig,
		workforceManagementApi:                             api,
		createBusinessUnitActivityCodeAttr:                 createBusinessUnitActivityCodeFn,
		getAllBusinessUnitActivityCodesAttr:                getAllBusinessUnitActivityCodesFn,
		getBusinessUnitActivityCodeIdByNameAttr:            getBusinessUnitActivityCodeIdByNameFn,
		getDefaultBusinessUnitActivityCodeIdByCategoryAttr: getDefaultBusinessUnitActivityCodeIdByCategoryFn,
		getBusinessUnitActivityCodeByIdAttr:                getBusinessUnitActivityCodeByIdFn,
		updateBusinessUnitActivityCodeAttr:                 updateBusinessUnitActivityCodeFn,
		deleteBusinessUnitActivityCodeAttr:                 deleteBusinessUnitActivityCodeFn,
		getAllBusinessUnitsAttr:                            getAllBusinessUnitsFn,
	}
}

// getBusinessUnitActivityCodeProxy acts as a singleton for the internalProxy. It also ensures
// that we can still proxy our tests by directly setting internalProxy package variable
func getBusinessUnitActivityCodeProxy(clientConfig *platformclientv2.Configuration) *businessUnitActivityCodeProxy {
	if internalProxy == nil {
		internalProxy = newBusinessUnitActivityCodeProxy(clientConfig)
	}

	return internalProxy
}

// createBusinessUnitActivityCode creates a Genesys Cloud business unit activity code
func (p *businessUnitActivityCodeProxy) createBusinessUnitActivityCode(ctx context.Context, businessUnitId string, activityCode *platformclientv2.Createactivitycoderequest) (*platformclientv2.Businessunitactivitycode, *platformclientv2.APIResponse, error) {
	return p.createBusinessUnitActivityCodeAttr(ctx, p, businessUnitId, activityCode)
}

// getAllBusinessUnitActivityCodes retrieves all Genesys Cloud activity codes of a business unit
func (p *businessUnitActivityCodeProxy) getAllBusinessUnitActivityCodes(ctx context.Context, businessUnitId string) (*[]platformclientv2.Businessunitactivitycode, *platformclientv2.APIResponse, error) {
	return p.getAllBusinessUnitActivityCodesAttr(ctx, p, businessUnitId)
}

// getBusinessUnitActivityCodeIdByName returns the id of a Genesys Cloud business unit activity code by name
func (p *businessUnitActivityCodeProxy) getBusinessUnitActivityCodeIdByName(ctx context.Context, businessUnitId string, name string) (string, *platformclientv2.APIResponse, error) {
	return p.getBusinessUnitActivityCodeIdByNameAttr(ctx, p, businessUnitId, name)
}

// getDefaultBusinessUnitActivityCodeIdByCategory returns the id of the default Genesys Cloud business unit activity code of a category
func (p *businessUnitActivityCodeProxy) getDefaultBusinessUnitActivityCodeIdByCategory(ctx context.Context, businessUnitId string, category string) (string, *platformclientv2.APIResponse, error) {
	return p.getDefaultBusinessUnitActivityCodeIdByCategoryAttr(ctx, p, businessUnitId, category)
}

// getBusinessUnitActivityCodeById returns a single Genesys Cloud business unit activity code by id
func (p *businessUnitActivityCodeProxy) getBusinessUnitActivityCodeById(ctx context.Context, businessUnitId string, activityCodeId string) (*platformclientv2.Businessunitactivitycode, *platformclientv2.APIResponse, error) {
	return p.getBusinessUnitActivityCodeByIdAttr(ctx, p, businessUnitId, activityCodeId)
}

// updateBusinessUnitActivityCode updates a Genesys Cloud business unit activity code
func (p *businessUnitActivityCodeProxy) updateBusinessUnitActivityCode(ctx context.Context, businessUnitId string, activityCodeId string, activityCode *platformclientv2.Updateactivitycoderequest) (*platformclientv2.Businessunitactivitycode, *platformclientv2.APIResponse, error) {
	return p.updateBusinessUnitActivityCodeAttr(ctx, p, businessUnitId, activityCodeId, activityCode)
}

// deleteBusinessUnitActivityCode deletes a Genesys Cloud business unit activity code by id
func (p *businessUnitActivityCodeProxy) deleteBusinessUnitActivityCode(ctx context.Context, businessUnitId string, activityCodeId string) (*platformclientv2.APIResponse, error) {
	return p.deleteBusinessUnitActivityCodeAttr(ctx, p, businessUnitId, activityCodeId)
}

// getAllBusinessUnits retrieves all Genesys Cloud business units. Used by the exporter to walk every business unit.
func (p *businessUnitActivityCodeProxy) getAllBusinessUnits(ctx context.Context) (*[]platformclientv2.Businessunitlistitem, *platformclientv2.APIResponse, error) {
	return p.getAllBusinessUnitsAttr(ctx, p)
}

// createBusinessUnitActivityCodeFn is an implementation function for creating a Genesys Cloud business unit activity code
func createBusinessUnitActivityCodeFn(ctx context.Context, p *businessUnitActivityCodeProxy, businessUnitId string, activityCode *platformclientv2.Createactivitycoderequest) (*platformclientv2.Businessunitactivitycode, *platformclientv2.APIResponse, error) {
	return p.workforceManagementApi.PostWorkforcemanagementBusinessunitActivitycodes(businessUnitId, *activityCode)
}

// getAllBusinessUnitActivityCodesFn is the implementation for retrieving all activity codes of a business unit in Genesys Cloud
func getAllBusinessUnitActivityCodesFn(ctx context.Context, p *businessUnitActivityCodeProxy, businessUnitId string) (*[]platformclientv2.Businessunitactivitycode, *platformclientv2.APIResponse, error) {
	listing, resp, err := p.workforceManagementApi.GetWorkforcemanagementBusinessunitActivitycodes(businessUnitId, false)
	if err != nil {
		return nil, resp, err
	}

	return listing.Entities, resp, nil
}

// getBusinessUnitActivityCodeIdByNameFn is an implementation of the function to get the id of an active business unit activity code by name
func getBusinessUnitActivityCodeIdByNameFn(ctx context.Context, p *businessUnitActivityCodeProxy, businessUnitId string, name string) (string, *platformclientv2.APIResponse, error) {
	activityCodes, resp, err := getAllBusinessUnitActivityCodesFn(ctx, p, businessUnitId)
	if err != nil {
		return "", resp, err
	}

	id, err := findActivityCodeIdByName(*activityCodes, name)
	return id, resp, err
}

// getDefaultBusinessUnitActivityCodeIdByCategoryFn is an implementation of the function to get the id of the default business unit activity code of a category
func getDefaultBusinessUnitActivityCodeIdByCategoryFn(ctx context.Context, p *businessUnitActivityCodeProxy, businessUnitId string, category string) (string, *platformclientv2.APIResponse, error) {
	activityCodes, resp, err := getAllBusinessUnitActivityCodesFn(ctx, p, businessUnitId)
	if err != nil {
		return "", resp, err
	}

	id, err := findDefaultActivityCodeIdByCategory(*activityCodes, category)
	return id, resp, err
}

// getBusinessUnitActivityCodeByIdFn is an implementation of the function to get a Genesys Cloud business unit activity code by id
func getBusinessUnitActivityCodeByIdFn(ctx context.Context, p *businessUnitActivityCodeProxy, businessUnitId string, activityCodeId string) (*platformclientv2.Businessunitactivitycode, *platformclientv2.APIResponse, error) {
	return p.workforceManagementApi.GetWorkforcemanagementBusinessunitActivitycode(businessUnitId, activityCodeId)
}

// updateBusinessUnitActivityCodeFn is an implementation of the function to update a Genesys Cloud business unit activity code
func updateBusinessUnitActivityCodeFn(ctx context.Context, p *businessUnitActivityCodeProxy, businessUnitId string, activityCodeId string, activityCode *platformclientv2.Updateactivitycoderequest) (*platformclientv2.Businessunitactivitycode, *platformclientv2.APIResponse, error) {
	return p.workforceManagementApi.PatchWorkforcemanagementBusinessunitActivitycode(businessUnitId, activityCodeId, *activityCode)
}

// deleteBusinessUnitActivityCodeFn is an implementation function for deleting a Genesys Cloud business unit activity code
func deleteBusinessUnitActivityCodeFn(ctx context.Context, p *businessUnitActivityCodeProxy, businessUnitId string, activityCodeId string) (*platformclientv2.APIResponse, error) {
	return p.workforceManagementApi.DeleteWorkforcemanagementBusinessunitActivitycode(businessUnitId, activityCodeId)
}

// getAllBusinessUnitsFn is an implementation of the function to list Genesys Cloud business units
func getAllBusinessUnitsFn(ctx context.Context, p *businessUnitActivityCodeProxy) (*[]platformclientv2.Businessunitlistitem, *platformclientv2.APIResponse, error) {
	listing, resp, err := p.workforceManagementApi.GetWorkforcemanagementBusinessunits("", "")
	if err != nil {
		return nil, resp, err
	}

	return listing.Entities, resp, nil
}
