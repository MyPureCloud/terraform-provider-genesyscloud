package genesyscloud

import (
	"context"
	"fmt"

	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

// Type definitions for each func on our proxy so we can easily mock them out later
type createSmsAddressFunc func(p *routingSmsAddressProxy, body platformclientv2.Smsaddressprovision) (*platformclientv2.Smsaddress, *platformclientv2.APIResponse, error)
type getAllSmsAddressesFunc func(p *routingSmsAddressProxy, ctx context.Context) (*[]platformclientv2.Smsaddress, *platformclientv2.APIResponse, error)
type getSmsAddressByIdFunc func(p *routingSmsAddressProxy, id string) (*platformclientv2.Smsaddress, *platformclientv2.APIResponse, error)
type getSmsAddressIdByNameFunc func(p *routingSmsAddressProxy, name string, ctx context.Context) (id string, retryable bool, resp *platformclientv2.APIResponse, err error)
type deleteSmsAddressByIdFunc func(p *routingSmsAddressProxy, id string) (*platformclientv2.APIResponse, error)

// routingSmsAddressProxy contains all of the methods that call genesys cloud APIs.
type routingSmsAddressProxy struct {
	routingApi                *platformclientv2.RoutingApi
	createSmsAddressAttr      createSmsAddressFunc
	getAllSmsAddressesAttr    getAllSmsAddressesFunc
	getSmsAddressByIdAttr     getSmsAddressByIdFunc
	getSmsAddressIdByNameAttr getSmsAddressIdByNameFunc
	deleteSmsAddressByIdAttr  deleteSmsAddressByIdFunc
}

// internalProxy holds a proxy instance that can be used throughout the package
var internalProxy *routingSmsAddressProxy

// newRoutingSmsAddressProxy initializes the sms address proxy with all of the data needed to communicate with Genesys Cloud
func newRoutingSmsAddressProxy(clientConfig *platformclientv2.Configuration) *routingSmsAddressProxy {
	api := platformclientv2.NewRoutingApiWithConfig(clientConfig)
	return &routingSmsAddressProxy{
		routingApi:                api,
		createSmsAddressAttr:      createSmsAddressFn,
		getAllSmsAddressesAttr:    getAllSmsAddressesFn,
		getSmsAddressByIdAttr:     getSmsAddressByIdFn,
		getSmsAddressIdByNameAttr: getSmsAddressIdByNameFn,
		deleteSmsAddressByIdAttr:  deleteSmsAddressByIdFn,
	}
}

// getRoutingSmsAddressProxy acts as a singleton for the internalProxy.  It also ensures
// that we can still proxy our tests by directly setting internalProxy package variable
func getRoutingSmsAddressProxy(clientConfig *platformclientv2.Configuration) *routingSmsAddressProxy {
	if internalProxy == nil {
		internalProxy = newRoutingSmsAddressProxy(clientConfig)
	}
	return internalProxy
}

// createSmsAddress creates a Genesys Cloud Sms Address
func (p *routingSmsAddressProxy) createSmsAddress(body platformclientv2.Smsaddressprovision) (*platformclientv2.Smsaddress, *platformclientv2.APIResponse, error) {
	return p.createSmsAddressAttr(p, body)
}

// getSmsAddressById gets a Genesys Cloud Sms Address by ID
func (p *routingSmsAddressProxy) getSmsAddressById(id string) (*platformclientv2.Smsaddress, *platformclientv2.APIResponse, error) {
	return p.getSmsAddressByIdAttr(p, id)
}

// getSmsAddressIdByName gets a Genesys Cloud Sms Address ID by name
func (p *routingSmsAddressProxy) getSmsAddressIdByName(name string, ctx context.Context) (id string, retryable bool, resp *platformclientv2.APIResponse, err error) {
	return p.getSmsAddressIdByNameAttr(p, name, ctx)
}

// getAllSmsAddresses gets all Genesys Cloud Sms Addresses
func (p *routingSmsAddressProxy) getAllSmsAddresses(ctx context.Context) (*[]platformclientv2.Smsaddress, *platformclientv2.APIResponse, error) {
	return p.getAllSmsAddressesAttr(p, ctx)
}

// deleteSmsAddress deletes a Genesys Cloud Sms Address by ID
func (p *routingSmsAddressProxy) deleteSmsAddress(id string) (*platformclientv2.APIResponse, error) {
	return p.deleteSmsAddressByIdAttr(p, id)
}

// createSmsAddressFn is an implementation function for creating a Genesys Cloud Sms Address
func createSmsAddressFn(p *routingSmsAddressProxy, body platformclientv2.Smsaddressprovision) (*platformclientv2.Smsaddress, *platformclientv2.APIResponse, error) {
	return p.routingApi.PostRoutingSmsAddresses(body)
}

// getAllSmsAddressesFn is an implementation function for getting all Sms Addresses
func getAllSmsAddressesFn(p *routingSmsAddressProxy, ctx context.Context) (*[]platformclientv2.Smsaddress, *platformclientv2.APIResponse, error) {
	var allSmsAddresses []platformclientv2.Smsaddress
	var response *platformclientv2.APIResponse
	for pageNum := 1; ; pageNum++ {
		const pageSize = 100
		smsAddresses, resp, getErr := p.routingApi.GetRoutingSmsAddresses(pageSize, pageNum)
		if getErr != nil {
			return nil, resp, fmt.Errorf("error requesting page of Routing Sms Addresses: %s", getErr)
		}
		response = resp
		if smsAddresses.Entities == nil || len(*smsAddresses.Entities) == 0 {
			break
		}
		for _, entity := range *smsAddresses.Entities {
			allSmsAddresses = append(allSmsAddresses, entity)
		}
	}
	return &allSmsAddresses, response, nil
}

// getSmsAddressByIdFn is an implementation function for getting a Genesys Cloud Sms Address by ID
func getSmsAddressByIdFn(p *routingSmsAddressProxy, id string) (*platformclientv2.Smsaddress, *platformclientv2.APIResponse, error) {
	return p.routingApi.GetRoutingSmsAddress(id)
}

// getSmsAddressIdByNameFn is an implementation function for getting a sms address ID by name.
func getSmsAddressIdByNameFn(p *routingSmsAddressProxy, name string, ctx context.Context) (id string, retryable bool, resp *platformclientv2.APIResponse, err error) {
	smsAddresses, resp, err := getAllSmsAddressesFn(p, ctx)
	if err != nil {
		return "", false, resp, fmt.Errorf("failed to read sms addresses: %v", err)
	}
	if smsAddresses == nil || len(*smsAddresses) == 0 {
		return "", true, resp, fmt.Errorf("failed to read sms addresses: %v", err)
	}
	for _, address := range *smsAddresses {
		if *address.Name == name {
			return *address.Id, false, resp, nil
		}
	}
	return "", true, resp, fmt.Errorf("failed to find sms address with name '%s'", name)
}

// deleteSmsAddressByIdFn is an implementation function for deleting a Genesys Cloud Sms Address by ID
func deleteSmsAddressByIdFn(p *routingSmsAddressProxy, id string) (*platformclientv2.APIResponse, error) {
	return p.routingApi.DeleteRoutingSmsAddress(id)
}
