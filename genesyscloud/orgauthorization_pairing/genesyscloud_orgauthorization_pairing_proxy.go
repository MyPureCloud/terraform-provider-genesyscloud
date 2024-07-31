package orgauthorization_pairing

import (
	"context"
	"fmt"

	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

var internalProxy *orgauthorizationPairingProxy

// Type definitions for each func on our proxy so we can easily mock them out later
type createOrgauthorizationPairingFunc func(ctx context.Context, p *orgauthorizationPairingProxy, trustRequestCreate *platformclientv2.Trustrequestcreate) (*platformclientv2.Trustrequest, *platformclientv2.APIResponse, error)
type getOrgauthorizationPairingByIdFunc func(ctx context.Context, p *orgauthorizationPairingProxy, id string) (trustRequest *platformclientv2.Trustrequest, response *platformclientv2.APIResponse, err error)

// orgauthorizationPairingProxy contains all the methods that call genesys cloud APIs.
type orgauthorizationPairingProxy struct {
	clientConfig                       *platformclientv2.Configuration
	organizationAuthorizationApi       *platformclientv2.OrganizationAuthorizationApi
	createOrgauthorizationPairingAttr  createOrgauthorizationPairingFunc
	getOrgauthorizationPairingByIdAttr getOrgauthorizationPairingByIdFunc
}

// newOrgauthorizationPairingProxy initializes the orgauthorization pairing proxy with all of the data needed to communicate with Genesys Cloud
func newOrgauthorizationPairingProxy(clientConfig *platformclientv2.Configuration) *orgauthorizationPairingProxy {
	api := platformclientv2.NewOrganizationAuthorizationApiWithConfig(clientConfig)
	return &orgauthorizationPairingProxy{
		clientConfig:                       clientConfig,
		organizationAuthorizationApi:       api,
		createOrgauthorizationPairingAttr:  createOrgauthorizationPairingFn,
		getOrgauthorizationPairingByIdAttr: getOrgauthorizationPairingByIdFn,
	}
}

// getOrgauthorizationPairingProxy acts as a singleton to for the internalProxy.  It also ensures
// that we can still proxy our tests by directly setting internalProxy package variable
func getOrgauthorizationPairingProxy(clientConfig *platformclientv2.Configuration) *orgauthorizationPairingProxy {
	if internalProxy == nil {
		internalProxy = newOrgauthorizationPairingProxy(clientConfig)
	}
	return internalProxy
}

// createOrgauthorizationPairing creates a Genesys Cloud orgauthorization pairing
func (p *orgauthorizationPairingProxy) createOrgauthorizationPairing(ctx context.Context, orgauthorizationPairing *platformclientv2.Trustrequestcreate) (*platformclientv2.Trustrequest, *platformclientv2.APIResponse, error) {
	return p.createOrgauthorizationPairingAttr(ctx, p, orgauthorizationPairing)
}

// getOrgauthorizationPairingById returns a single Genesys Cloud orgauthorization pairing by Id
func (p *orgauthorizationPairingProxy) getOrgauthorizationPairingById(ctx context.Context, id string) (orgauthorizationPairing *platformclientv2.Trustrequest, response *platformclientv2.APIResponse, err error) {
	return p.getOrgauthorizationPairingByIdAttr(ctx, p, id)
}

// createOrgauthorizationPairingFn is an implementation function for creating a Genesys Cloud orgauthorization pairing
func createOrgauthorizationPairingFn(ctx context.Context, p *orgauthorizationPairingProxy, orgauthorizationPairing *platformclientv2.Trustrequestcreate) (*platformclientv2.Trustrequest, *platformclientv2.APIResponse, error) {
	trustRequestCreate, resp, err := p.organizationAuthorizationApi.PostOrgauthorizationPairings(*orgauthorizationPairing)
	if err != nil {
		return nil, resp, fmt.Errorf("failed to create orgauthorization pairing: %s", err)
	}
	return trustRequestCreate, resp, nil
}

// getOrgauthorizationPairingByIdFn is an implementation of the function to get a Genesys Cloud orgauthorization pairing by Id
func getOrgauthorizationPairingByIdFn(ctx context.Context, p *orgauthorizationPairingProxy, id string) (orgauthorizationPairing *platformclientv2.Trustrequest, response *platformclientv2.APIResponse, err error) {
	trustRequestCreate, resp, err := p.organizationAuthorizationApi.GetOrgauthorizationPairing(id)
	if err != nil {
		return nil, resp, fmt.Errorf("failed to retrieve orgauthorization pairing by id %s: %s", id, err)
	}
	return trustRequestCreate, resp, nil
}
