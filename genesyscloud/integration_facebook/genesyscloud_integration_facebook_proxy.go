package integration_facebook

import (
	"context"
	"fmt"
	"log"
	rc "terraform-provider-genesyscloud/genesyscloud/resource_cache"

	"github.com/mypurecloud/platform-client-sdk-go/v146/platformclientv2"
)

/*
The genesyscloud_integration_facebook_proxy.go file contains the proxy structures and methods that interact
with the Genesys Cloud SDK. We use composition here for each function on the proxy so individual functions can be stubbed
out during testing.
*/

// internalProxy holds a proxy instance that can be used throughout the package
var internalProxy *integrationFacebookProxy

// Type definitions for each func on our proxy so we can easily mock them out later
type createIntegrationFacebookFunc func(ctx context.Context, p *integrationFacebookProxy, facebookIntegrationRequest *platformclientv2.Facebookintegrationrequest) (*platformclientv2.Facebookintegration, *platformclientv2.APIResponse, error)
type getAllIntegrationFacebookFunc func(ctx context.Context, p *integrationFacebookProxy) (*[]platformclientv2.Facebookintegration, *platformclientv2.APIResponse, error)
type getIntegrationFacebookIdByNameFunc func(ctx context.Context, p *integrationFacebookProxy, name string) (id string, retryable bool, resp *platformclientv2.APIResponse, err error)
type getIntegrationFacebookByIdFunc func(ctx context.Context, p *integrationFacebookProxy, id string) (facebookIntegrationRequest *platformclientv2.Facebookintegration, response *platformclientv2.APIResponse, err error)
type updateIntegrationFacebookFunc func(ctx context.Context, p *integrationFacebookProxy, id string, facebookIntegrationRequest *platformclientv2.Facebookintegrationupdaterequest) (*platformclientv2.Facebookintegration, *platformclientv2.APIResponse, error)
type deleteIntegrationFacebookFunc func(ctx context.Context, p *integrationFacebookProxy, id string) (response *platformclientv2.APIResponse, err error)

// integrationFacebookProxy contains all of the methods that call genesys cloud APIs.
type integrationFacebookProxy struct {
	clientConfig                       *platformclientv2.Configuration
	conversationsApi                   *platformclientv2.ConversationsApi
	createIntegrationFacebookAttr      createIntegrationFacebookFunc
	getAllIntegrationFacebookAttr      getAllIntegrationFacebookFunc
	getIntegrationFacebookIdByNameAttr getIntegrationFacebookIdByNameFunc
	getIntegrationFacebookByIdAttr     getIntegrationFacebookByIdFunc
	updateIntegrationFacebookAttr      updateIntegrationFacebookFunc
	deleteIntegrationFacebookAttr      deleteIntegrationFacebookFunc
	facebookCache                      rc.CacheInterface[platformclientv2.Facebookintegration]
}

// newIntegrationFacebookProxy initializes the integration facebook proxy with all of the data needed to communicate with Genesys Cloud
func newIntegrationFacebookProxy(clientConfig *platformclientv2.Configuration) *integrationFacebookProxy {
	api := platformclientv2.NewConversationsApiWithConfig(clientConfig)
	facebookCache := rc.NewResourceCache[platformclientv2.Facebookintegration]()
	return &integrationFacebookProxy{
		clientConfig:                       clientConfig,
		conversationsApi:                   api,
		createIntegrationFacebookAttr:      createIntegrationFacebookFn,
		getAllIntegrationFacebookAttr:      getAllIntegrationFacebookFn,
		getIntegrationFacebookIdByNameAttr: getIntegrationFacebookIdByNameFn,
		getIntegrationFacebookByIdAttr:     getIntegrationFacebookByIdFn,
		updateIntegrationFacebookAttr:      updateIntegrationFacebookFn,
		deleteIntegrationFacebookAttr:      deleteIntegrationFacebookFn,
		facebookCache:                      facebookCache,
	}
}

// getIntegrationFacebookProxy acts as a singleton to for the internalProxy.  It also ensures
// that we can still proxy our tests by directly setting internalProxy package variable
func getIntegrationFacebookProxy(clientConfig *platformclientv2.Configuration) *integrationFacebookProxy {
	if internalProxy == nil {
		internalProxy = newIntegrationFacebookProxy(clientConfig)
	}

	return internalProxy
}

// createIntegrationFacebook creates a Genesys Cloud integration facebook
func (p *integrationFacebookProxy) createIntegrationFacebook(ctx context.Context, integrationFacebook *platformclientv2.Facebookintegrationrequest) (*platformclientv2.Facebookintegration, *platformclientv2.APIResponse, error) {
	return p.createIntegrationFacebookAttr(ctx, p, integrationFacebook)
}

// getIntegrationFacebook retrieves all Genesys Cloud integration facebook
func (p *integrationFacebookProxy) getAllIntegrationFacebook(ctx context.Context) (*[]platformclientv2.Facebookintegration, *platformclientv2.APIResponse, error) {
	return p.getAllIntegrationFacebookAttr(ctx, p)
}

// getIntegrationFacebookIdByName returns a single Genesys Cloud integration facebook by a name
func (p *integrationFacebookProxy) getIntegrationFacebookIdByName(ctx context.Context, name string) (id string, retryable bool, resp *platformclientv2.APIResponse, err error) {
	return p.getIntegrationFacebookIdByNameAttr(ctx, p, name)
}

// getIntegrationFacebookById returns a single Genesys Cloud integration facebook by Id
func (p *integrationFacebookProxy) getIntegrationFacebookById(ctx context.Context, id string) (integrationFacebook *platformclientv2.Facebookintegration, response *platformclientv2.APIResponse, err error) {
	return p.getIntegrationFacebookByIdAttr(ctx, p, id)
}

// updateIntegrationFacebook updates a Genesys Cloud integration facebook
func (p *integrationFacebookProxy) updateIntegrationFacebook(ctx context.Context, id string, integrationFacebook *platformclientv2.Facebookintegrationupdaterequest) (*platformclientv2.Facebookintegration, *platformclientv2.APIResponse, error) {
	return p.updateIntegrationFacebookAttr(ctx, p, id, integrationFacebook)
}

// deleteIntegrationFacebook deletes a Genesys Cloud integration facebook by Id
func (p *integrationFacebookProxy) deleteIntegrationFacebook(ctx context.Context, id string) (response *platformclientv2.APIResponse, err error) {
	return p.deleteIntegrationFacebookAttr(ctx, p, id)
}

// createIntegrationFacebookFn is an implementation function for creating a Genesys Cloud integration facebook
func createIntegrationFacebookFn(ctx context.Context, p *integrationFacebookProxy, integrationFacebook *platformclientv2.Facebookintegrationrequest) (*platformclientv2.Facebookintegration, *platformclientv2.APIResponse, error) {
	return p.conversationsApi.PostConversationsMessagingIntegrationsFacebook(*integrationFacebook)
}

// getAllIntegrationFacebookFn is the implementation for retrieving all integration facebook in Genesys Cloud
func getAllIntegrationFacebookFn(ctx context.Context, p *integrationFacebookProxy) (*[]platformclientv2.Facebookintegration, *platformclientv2.APIResponse, error) {
	var allFacebookIntegrationRequests []platformclientv2.Facebookintegration
	const pageSize = 100

	facebookIntegrationRequests, resp, err := p.conversationsApi.GetConversationsMessagingIntegrationsFacebook(pageSize, 1, "", "", "")
	if err != nil {
		return nil, resp, fmt.Errorf("Failed to get facebook integration request: %v %v", err, resp)
	}
	if facebookIntegrationRequests.Entities == nil || len(*facebookIntegrationRequests.Entities) == 0 {
		return &allFacebookIntegrationRequests, resp, nil
	}

	allFacebookIntegrationRequests = append(allFacebookIntegrationRequests, *facebookIntegrationRequests.Entities...)

	for pageNum := 2; pageNum <= *facebookIntegrationRequests.PageCount; pageNum++ {

		facebookIntegrationRequests, resp, err := p.conversationsApi.GetConversationsMessagingIntegrationsFacebook(pageSize, pageNum, "", "", "")
		if err != nil {
			return nil, resp, fmt.Errorf("Failed to get facebook integration request: %v %v", err, resp)
		}

		if facebookIntegrationRequests.Entities == nil || len(*facebookIntegrationRequests.Entities) == 0 {
			break
		}

		allFacebookIntegrationRequests = append(allFacebookIntegrationRequests, *facebookIntegrationRequests.Entities...)
	}

	for _, facebookReq := range allFacebookIntegrationRequests {
		rc.SetCache(p.facebookCache, *facebookReq.Id, facebookReq)
	}

	return &allFacebookIntegrationRequests, resp, err
}

// getIntegrationFacebookIdByNameFn is an implementation of the function to get a Genesys Cloud integration facebook by name
func getIntegrationFacebookIdByNameFn(ctx context.Context, p *integrationFacebookProxy, name string) (string, bool, *platformclientv2.APIResponse, error) {
	facebookIntegrationRequests, resp, err := getAllIntegrationFacebookFn(ctx, p)
	if err != nil {
		return "", false, resp, err
	}

	if facebookIntegrationRequests == nil || len(*facebookIntegrationRequests) == 0 {
		return "", true, resp, fmt.Errorf("No integration facebook found with name %s", name)
	}

	var facebookIntegration platformclientv2.Facebookintegration
	for _, facebookIntegrationRequest := range *facebookIntegrationRequests {
		if *facebookIntegrationRequest.Name == name {
			log.Printf("Retrieved the integration facebook id %s by name %s", *facebookIntegrationRequest.Id, name)
			facebookIntegration = facebookIntegrationRequest
			return *facebookIntegration.Id, false, resp, nil
		}
	}

	return "", true, resp, fmt.Errorf("No integration facebook found with name %s", name)
}

// getIntegrationFacebookByIdFn is an implementation of the function to get a Genesys Cloud integration facebook by Id
func getIntegrationFacebookByIdFn(ctx context.Context, p *integrationFacebookProxy, id string) (integrationFacebook *platformclientv2.Facebookintegration, response *platformclientv2.APIResponse, err error) {
	facebookReq := rc.GetCacheItem(p.facebookCache, id)
	if facebookReq != nil {
		return facebookReq, nil, nil
	}

	return p.conversationsApi.GetConversationsMessagingIntegrationsFacebookIntegrationId(id, "")
}

// updateIntegrationFacebookFn is an implementation of the function to update a Genesys Cloud integration facebook
func updateIntegrationFacebookFn(ctx context.Context, p *integrationFacebookProxy, id string, integrationFacebook *platformclientv2.Facebookintegrationupdaterequest) (*platformclientv2.Facebookintegration, *platformclientv2.APIResponse, error) {
	return p.conversationsApi.PatchConversationsMessagingIntegrationsFacebookIntegrationId(id, *integrationFacebook)
}

// deleteIntegrationFacebookFn is an implementation function for deleting a Genesys Cloud integration facebook
func deleteIntegrationFacebookFn(ctx context.Context, p *integrationFacebookProxy, id string) (response *platformclientv2.APIResponse, err error) {
	return p.conversationsApi.DeleteConversationsMessagingIntegrationsFacebookIntegrationId(id)
}
