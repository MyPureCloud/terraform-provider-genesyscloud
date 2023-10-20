package telephony_providers_edges_phone

import (
	"context"
	"fmt"

	"github.com/mypurecloud/platform-client-sdk-go/v115/platformclientv2"
)

/*
The genesyscloud_telephony_providers_edges_phone_proxy.go file contains the proxy structures and methods that interact
with the Genesys Cloud SDK. We use composition here for each function on the proxy so individual functions can be stubbed
out during testing.

Each proxy implementation:

1.  Should provide a private package level variable that holds a instance of a proxy class.
2.  A New... constructor function  to initialize the proxy object.  This constructor should only be used within
    the proxy.
3.  A get private constructor function that the classes in the package can be used to to retrieve
    the proxy.  This proxy should check to see if the package level proxy instance is nil and
    should initialize it, otherwise it should return the instance
4.  Type definitions for each function that will be used in the proxy.  We use composition here
    so that we can easily provide mocks for testing.
5.  A struct for the proxy that holds an attribute for each function type.
6.  Wrapper methods on each of the elements on the struct.
7.  Function implementations for each function type definition.

*/

// internalProxy holds a proxy instance that can be used throughout the package
var internalProxy *phoneProxy

// Type definitions for each func on our proxy so we can easily mock them out later
type getAllPhonesFunc func(ctx context.Context, p *phoneProxy) (*[]platformclientv2.Phone, error)
type createPhoneFunc func(ctx context.Context, p *phoneProxy, phoneConfig *platformclientv2.Phone) (*platformclientv2.Phone, *platformclientv2.APIResponse, error)
type getPhoneByIdFunc func(ctx context.Context, p *phoneProxy, phoneId string) (*platformclientv2.Phone, *platformclientv2.APIResponse, error)
type getPhoneByNameFunc func(ctx context.Context, p *phoneProxy, phoneName string) (phone *platformclientv2.Phone, retryable bool, err error)
type updatePhoneFunc func(ctx context.Context, p *phoneProxy, phoneId string, phoneConfig *platformclientv2.Phone) (*platformclientv2.Phone, error)
type deletePhoneFunc func(ctx context.Context, p *phoneProxy, phoneId string) (responseCode int, err error)

type getPhoneBaseSettingFunc func(ctx context.Context, p *phoneProxy, phoneBaseSettingsId string) (*platformclientv2.Phonebase, error)
type getStationOfUserFunc func(ctx context.Context, p *phoneProxy, userId string) (station *platformclientv2.Station, retryable bool, err error)
type unassignUserFromStationFunc func(ctx context.Context, p *phoneProxy, stationId string) (*platformclientv2.APIResponse, error)
type assignUserToStationFunc func(ctx context.Context, p *phoneProxy, userId string, stationId string) (*platformclientv2.APIResponse, error)

// phoneProxy contains all of the methods that call genesys cloud APIs.
type phoneProxy struct {
	clientConfig *platformclientv2.Configuration
	edgesApi     *platformclientv2.TelephonyProvidersEdgeApi
	stationsApi  *platformclientv2.StationsApi
	usersApi     *platformclientv2.UsersApi

	getAllPhonesAttr   getAllPhonesFunc
	createPhoneAttr    createPhoneFunc
	getPhoneByIdAttr   getPhoneByIdFunc
	getPhoneByNameAttr getPhoneByNameFunc
	updatePhoneAttr    updatePhoneFunc
	deletePhoneAttr    deletePhoneFunc

	getPhoneBaseSettingAttr     getPhoneBaseSettingFunc
	getStationOfUserAttr        getStationOfUserFunc
	unassignUserFromStationAttr unassignUserFromStationFunc
	assignUserToStationAttr     assignUserToStationFunc
}

// newPhoneProxy initializes the Phone proxy with all of the data needed to communicate with Genesys Cloud
func newPhoneProxy(clientConfig *platformclientv2.Configuration) *phoneProxy {
	edgesApi := platformclientv2.NewTelephonyProvidersEdgeApiWithConfig(clientConfig)
	stationsApi := platformclientv2.NewStationsApiWithConfig(clientConfig)
	usersApi := platformclientv2.NewUsersApiWithConfig(clientConfig)

	return &phoneProxy{
		clientConfig: clientConfig,
		edgesApi:     edgesApi,
		stationsApi:  stationsApi,
		usersApi:     usersApi,

		getAllPhonesAttr:   getAllPhonesFn,
		createPhoneAttr:    createPhoneFn,
		getPhoneByIdAttr:   getPhoneByIdFn,
		getPhoneByNameAttr: getPhoneByNameFn,
		updatePhoneAttr:    updatePhoneFn,
		deletePhoneAttr:    deletePhoneFn,

		getPhoneBaseSettingAttr:     getPhoneBaseSettingFn,
		getStationOfUserAttr:        getStationOfUserFn,
		unassignUserFromStationAttr: unassignUserFromStationFn,
		assignUserToStationAttr:     assignUserToStationFn,
	}
}

// getPhoneProxy acts as a singleton to for the internalProxy.  It also ensures
// that we can still proxy our tests by directly setting internalProxy package variable
func getPhoneProxy(clientConfig *platformclientv2.Configuration) *phoneProxy {
	if internalProxy == nil {
		internalProxy = newPhoneProxy(clientConfig)
	}

	return internalProxy
}

// getAllPhones retrieves all Genesys Cloud Phones
func (p *phoneProxy) getAllPhones(ctx context.Context) (*[]platformclientv2.Phone, error) {
	return p.getAllPhonesAttr(ctx, p)
}

// createPhone creates a Genesys Cloud Phone
func (p *phoneProxy) createPhone(ctx context.Context, phoneConfig *platformclientv2.Phone) (*platformclientv2.Phone, *platformclientv2.APIResponse, error) {
	return p.createPhoneAttr(ctx, p, phoneConfig)
}

// getPhoneById retrieves a Genesys Cloud Phone by id
func (p *phoneProxy) getPhoneById(ctx context.Context, phoneId string) (*platformclientv2.Phone, *platformclientv2.APIResponse, error) {
	return p.getPhoneByIdAttr(ctx, p, phoneId)
}

// getPhoneByName retrieves a Genesys Cloud Phone by name
func (p *phoneProxy) getPhoneByName(ctx context.Context, phoneName string) (phone *platformclientv2.Phone, retryable bool, err error) {
	return p.getPhoneByNameAttr(ctx, p, phoneName)
}

// updatePhone updates a Genesys Cloud Phone
func (p *phoneProxy) updatePhone(ctx context.Context, phoneId string, phoneConfig *platformclientv2.Phone) (*platformclientv2.Phone, error) {
	return p.updatePhoneAttr(ctx, p, phoneId, phoneConfig)
}

// deletePhone deletes a Genesys Cloud Phone
func (p *phoneProxy) deletePhone(ctx context.Context, phoneId string) (responseCode int, err error) {
	return p.deletePhoneAttr(ctx, p, phoneId)
}

// getPhoneBaseSetting retrieves a Genesys Cloud Phone Base Setting
func (p *phoneProxy) getPhoneBaseSetting(ctx context.Context, phoneBaseSettingsId string) (*platformclientv2.Phonebase, error) {
	return p.getPhoneBaseSettingAttr(ctx, p, phoneBaseSettingsId)
}

// getStationOfUser retrieves the station of a user
func (p *phoneProxy) getStationOfUser(ctx context.Context, userId string) (*platformclientv2.Station, bool, error) {
	return p.getStationOfUserAttr(ctx, p, userId)
}

// unassignUserFromStation unassigns a user from the station
func (p *phoneProxy) unassignUserFromStation(ctx context.Context, stationId string) (*platformclientv2.APIResponse, error) {
	return p.unassignUserFromStationAttr(ctx, p, stationId)
}

// assignUserToStation assigns a user to the station
func (p *phoneProxy) assignUserToStation(ctx context.Context, userId string, stationId string) (*platformclientv2.APIResponse, error) {
	return p.assignUserToStationAttr(ctx, p, userId, stationId)
}

// getAllPhonesFn is an implementation function for retrieving all Genesys Cloud Phones
func getAllPhonesFn(ctx context.Context, p *phoneProxy) (*[]platformclientv2.Phone, error) {
	var allPhones []platformclientv2.Phone

	phones, _, _ := p.edgesApi.GetTelephonyProvidersEdgesPhones(1, 100, "", "", "", "", "", "", "", "", "", "", "", "", "", nil, nil)
	for _, phone := range *phones.Entities {
		if phone.State != nil && *phone.State != "deleted" {
			allPhones = append(allPhones, phone)
		}
	}

	for pageNum := 2; pageNum <= *phones.PageCount; pageNum++ {
		const pageSize = 100
		const sortBy = "id"
		phones, _, err := p.edgesApi.GetTelephonyProvidersEdgesPhones(pageNum, pageSize, sortBy, "", "", "", "", "", "", "", "", "", "", "", "", nil, nil)
		if err != nil {
			return nil, err
		}

		if phones.Entities == nil || len(*phones.Entities) == 0 {
			break
		}

		for _, phone := range *phones.Entities {
			if phone.State != nil && *phone.State != "deleted" {
				allPhones = append(allPhones, phone)
			}
		}
	}

	return &allPhones, nil
}

// createPhoneFn is an implementation function for creating a Genesys Cloud Phone
func createPhoneFn(ctx context.Context, p *phoneProxy, phoneConfig *platformclientv2.Phone) (*platformclientv2.Phone, *platformclientv2.APIResponse, error) {
	phone, resp, err := p.edgesApi.PostTelephonyProvidersEdgesPhones(*phoneConfig)
	if err != nil {
		return nil, resp, err
	}

	return phone, resp, nil
}

// getPhoneByIdFn is an implementation function for retrieving a Genesys Cloud Phone by id
func getPhoneByIdFn(ctx context.Context, p *phoneProxy, phoneId string) (*platformclientv2.Phone, *platformclientv2.APIResponse, error) {
	phone, resp, err := p.edgesApi.GetTelephonyProvidersEdgesPhone(phoneId)
	if err != nil {
		return nil, resp, err
	}

	return phone, resp, nil
}

// getPhoneByNameFn is an implementation function for retrieving a Genesys Cloud Phone by name
func getPhoneByNameFn(ctx context.Context, p *phoneProxy, phoneName string) (phone *platformclientv2.Phone, retryable bool, err error) {
	const pageSize = 100
	phones, _, err := p.edgesApi.GetTelephonyProvidersEdgesPhones(1, pageSize, "", "", "", "", "", "", "", "", "", "", phoneName, "", "", nil, nil)
	if err != nil {
		return nil, false, err
	}
	if phones.Entities == nil || len(*phones.Entities) == 0 {
		return nil, true, fmt.Errorf("failed to find ID of phone '%s'", phoneName)
	}

	for _, phone := range *phones.Entities {
		if *phone.Name == phoneName {
			return &phone, false, nil
		}
	}

	for pageNum := 2; pageNum <= *phones.PageCount; pageNum++ {
		phones, _, err := p.edgesApi.GetTelephonyProvidersEdgesPhones(pageNum, pageSize, "", "", "", "", "", "", "", "", "", "", phoneName, "", "", nil, nil)
		if err != nil {
			return nil, false, err
		}
		if phones.Entities == nil {
			return nil, true, fmt.Errorf("failed to find ID of phone '%s'", phoneName)
		}

		for _, phone := range *phones.Entities {
			if *phone.Name == phoneName {
				return &phone, false, nil
			}
		}
	}

	return nil, true, fmt.Errorf("failed to find ID of phone '%s'", phoneName)
}

// updatePhoneFn is an implementation function for updating a Genesys Cloud Phone
func updatePhoneFn(ctx context.Context, p *phoneProxy, phoneId string, phoneConfig *platformclientv2.Phone) (*platformclientv2.Phone, error) {
	phone, _, err := p.edgesApi.PutTelephonyProvidersEdgesPhone(phoneId, *phoneConfig)
	if err != nil {
		return nil, err
	}

	return phone, err
}

// deletePhoneFn is an implementation function for deleting a Genesys Cloud Phone
func deletePhoneFn(ctx context.Context, p *phoneProxy, phoneId string) (responseCode int, err error) {
	resp, err := p.edgesApi.DeleteTelephonyProvidersEdgesPhone(phoneId)
	return resp.StatusCode, err
}

// getPhoneBaseSettingFn is an implementation function for retrieving a Genesys Cloud Phone Base Setting
func getPhoneBaseSettingFn(ctx context.Context, p *phoneProxy, phoneBaseSettingsId string) (*platformclientv2.Phonebase, error) {
	phoneBase, _, err := p.edgesApi.GetTelephonyProvidersEdgesPhonebasesetting(phoneBaseSettingsId)
	if err != nil {
		return nil, err
	}

	return phoneBase, nil
}

// getStationOfUserFn is an implementation function for retrieving a Genesys Cloud User Station
func getStationOfUserFn(ctx context.Context, p *phoneProxy, userId string) (station *platformclientv2.Station, retryable bool, err error) {
	const pageSize = 100
	const pageNum = 1
	stations, _, err := p.stationsApi.GetStations(pageSize, pageNum, "", "", "", userId, "", "")
	if err != nil {
		return nil, false, err
	}
	if stations.Entities == nil || len(*stations.Entities) == 0 {
		return nil, true, nil
	}

	return &(*stations.Entities)[0], false, err
}

// unassignUserFromStationFn is an implementation function for unassigning a Genesys Cloud User from a Station
func unassignUserFromStationFn(ctx context.Context, p *phoneProxy, stationId string) (*platformclientv2.APIResponse, error) {
	return p.stationsApi.DeleteStationAssociateduser(stationId)
}

// assignUserToStationFn is an implementation function for assigning a Genesys Cloud User to a Station
func assignUserToStationFn(ctx context.Context, p *phoneProxy, userId string, stationId string) (*platformclientv2.APIResponse, error) {
	return p.usersApi.PutUserStationAssociatedstationStationId(userId, stationId)
}
