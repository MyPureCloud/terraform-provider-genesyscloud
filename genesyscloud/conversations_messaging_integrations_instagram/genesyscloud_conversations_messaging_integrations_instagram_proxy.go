package conversations_messaging_integrations_instagram

import (
	"context"
	"fmt"
	"log"

	"github.com/mypurecloud/platform-client-sdk-go/v146/platformclientv2"
)

/*
The genesyscloud_conversations_messaging_integrations_instagram_proxy.go file contains the proxy structures and methods that interact
with the Genesys Cloud SDK. We use composition here for each function on the proxy so individual functions can be stubbed
out during testing.
*/

// internalProxy holds a proxy instance that can be used throughout the package
var internalProxy *conversationsMessagingIntegrationsInstagramProxy

// Type definitions for each func on our proxy so we can easily mock them out later
type createConversationsMessagingIntegrationsInstagramFunc func(ctx context.Context, p *conversationsMessagingIntegrationsInstagramProxy, instagramIntegrationRequest *platformclientv2.Instagramintegrationrequest) (*platformclientv2.Instagramintegration, *platformclientv2.APIResponse, error)
type getAllConversationsMessagingIntegrationsInstagramFunc func(ctx context.Context, p *conversationsMessagingIntegrationsInstagramProxy) (*[]platformclientv2.Instagramintegration, *platformclientv2.APIResponse, error)
type getConversationsMessagingIntegrationsInstagramIdByNameFunc func(ctx context.Context, p *conversationsMessagingIntegrationsInstagramProxy, name string) (id string, retryable bool, response *platformclientv2.APIResponse, err error)
type getConversationsMessagingIntegrationsInstagramByIdFunc func(ctx context.Context, p *conversationsMessagingIntegrationsInstagramProxy, id string) (instagramIntegrationRequest *platformclientv2.Instagramintegration, response *platformclientv2.APIResponse, err error)
type updateConversationsMessagingIntegrationsInstagramFunc func(ctx context.Context, p *conversationsMessagingIntegrationsInstagramProxy, id string, instagramIntegrationRequest *platformclientv2.Instagramintegrationupdaterequest) (*platformclientv2.Instagramintegration, *platformclientv2.APIResponse, error)
type deleteConversationsMessagingIntegrationsInstagramFunc func(ctx context.Context, p *conversationsMessagingIntegrationsInstagramProxy, id string) (response *platformclientv2.APIResponse, err error)

// conversationsMessagingIntegrationsInstagramProxy contains all of the methods that call genesys cloud APIs.
type conversationsMessagingIntegrationsInstagramProxy struct {
	clientConfig                                               *platformclientv2.Configuration
	conversationsApi                                           *platformclientv2.ConversationsApi
	createConversationsMessagingIntegrationsInstagramAttr      createConversationsMessagingIntegrationsInstagramFunc
	getAllConversationsMessagingIntegrationsInstagramAttr      getAllConversationsMessagingIntegrationsInstagramFunc
	getConversationsMessagingIntegrationsInstagramIdByNameAttr getConversationsMessagingIntegrationsInstagramIdByNameFunc
	getConversationsMessagingIntegrationsInstagramByIdAttr     getConversationsMessagingIntegrationsInstagramByIdFunc
	updateConversationsMessagingIntegrationsInstagramAttr      updateConversationsMessagingIntegrationsInstagramFunc
	deleteConversationsMessagingIntegrationsInstagramAttr      deleteConversationsMessagingIntegrationsInstagramFunc
}

// newConversationsMessagingIntegrationsInstagramProxy initializes the conversations messaging integrations instagram proxy with all of the data needed to communicate with Genesys Cloud
func newConversationsMessagingIntegrationsInstagramProxy(clientConfig *platformclientv2.Configuration) *conversationsMessagingIntegrationsInstagramProxy {
	api := platformclientv2.NewConversationsApiWithConfig(clientConfig)
	return &conversationsMessagingIntegrationsInstagramProxy{
		clientConfig:     clientConfig,
		conversationsApi: api,
		createConversationsMessagingIntegrationsInstagramAttr:      createConversationsMessagingIntegrationsInstagramFn,
		getAllConversationsMessagingIntegrationsInstagramAttr:      getAllConversationsMessagingIntegrationsInstagramFn,
		getConversationsMessagingIntegrationsInstagramIdByNameAttr: getConversationsMessagingIntegrationsInstagramIdByNameFn,
		getConversationsMessagingIntegrationsInstagramByIdAttr:     getConversationsMessagingIntegrationsInstagramByIdFn,
		updateConversationsMessagingIntegrationsInstagramAttr:      updateConversationsMessagingIntegrationsInstagramFn,
		deleteConversationsMessagingIntegrationsInstagramAttr:      deleteConversationsMessagingIntegrationsInstagramFn,
	}
}

// getConversationsMessagingIntegrationsInstagramProxy acts as a singleton to for the internalProxy.  It also ensures
// that we can still proxy our tests by directly setting internalProxy package variable
func getConversationsMessagingIntegrationsInstagramProxy(clientConfig *platformclientv2.Configuration) *conversationsMessagingIntegrationsInstagramProxy {
	if internalProxy == nil {
		internalProxy = newConversationsMessagingIntegrationsInstagramProxy(clientConfig)
	}

	return internalProxy
}

// createConversationsMessagingIntegrationsInstagram creates a Genesys Cloud conversations messaging integrations instagram
func (p *conversationsMessagingIntegrationsInstagramProxy) createConversationsMessagingIntegrationsInstagram(ctx context.Context, conversationsMessagingIntegrationsInstagram *platformclientv2.Instagramintegrationrequest) (*platformclientv2.Instagramintegration, *platformclientv2.APIResponse, error) {
	return p.createConversationsMessagingIntegrationsInstagramAttr(ctx, p, conversationsMessagingIntegrationsInstagram)
}

// getConversationsMessagingIntegrationsInstagram retrieves all Genesys Cloud conversations messaging integrations instagram
func (p *conversationsMessagingIntegrationsInstagramProxy) getAllConversationsMessagingIntegrationsInstagram(ctx context.Context) (*[]platformclientv2.Instagramintegration, *platformclientv2.APIResponse, error) {
	return p.getAllConversationsMessagingIntegrationsInstagramAttr(ctx, p)
}

// getConversationsMessagingIntegrationsInstagramIdByName returns a single Genesys Cloud conversations messaging integrations instagram by a name
func (p *conversationsMessagingIntegrationsInstagramProxy) getConversationsMessagingIntegrationsInstagramIdByName(ctx context.Context, name string) (id string, retryable bool, response *platformclientv2.APIResponse, err error) {
	return p.getConversationsMessagingIntegrationsInstagramIdByNameAttr(ctx, p, name)
}

// getConversationsMessagingIntegrationsInstagramById returns a single Genesys Cloud conversations messaging integrations instagram by Id
func (p *conversationsMessagingIntegrationsInstagramProxy) getConversationsMessagingIntegrationsInstagramById(ctx context.Context, id string) (conversationsMessagingIntegrationsInstagram *platformclientv2.Instagramintegration, response *platformclientv2.APIResponse, err error) {
	return p.getConversationsMessagingIntegrationsInstagramByIdAttr(ctx, p, id)
}

// updateConversationsMessagingIntegrationsInstagram updates a Genesys Cloud conversations messaging integrations instagram
func (p *conversationsMessagingIntegrationsInstagramProxy) updateConversationsMessagingIntegrationsInstagram(ctx context.Context, id string, conversationsMessagingIntegrationsInstagram *platformclientv2.Instagramintegrationupdaterequest) (*platformclientv2.Instagramintegration, *platformclientv2.APIResponse, error) {
	return p.updateConversationsMessagingIntegrationsInstagramAttr(ctx, p, id, conversationsMessagingIntegrationsInstagram)
}

// deleteConversationsMessagingIntegrationsInstagram deletes a Genesys Cloud conversations messaging integrations instagram by Id
func (p *conversationsMessagingIntegrationsInstagramProxy) deleteConversationsMessagingIntegrationsInstagram(ctx context.Context, id string) (response *platformclientv2.APIResponse, err error) {
	return p.deleteConversationsMessagingIntegrationsInstagramAttr(ctx, p, id)
}

// createConversationsMessagingIntegrationsInstagramFn is an implementation function for creating a Genesys Cloud conversations messaging integrations instagram
func createConversationsMessagingIntegrationsInstagramFn(ctx context.Context, p *conversationsMessagingIntegrationsInstagramProxy, conversationsMessagingIntegrationsInstagram *platformclientv2.Instagramintegrationrequest) (*platformclientv2.Instagramintegration, *platformclientv2.APIResponse, error) {
	return p.conversationsApi.PostConversationsMessagingIntegrationsInstagram(*conversationsMessagingIntegrationsInstagram)
}

// getAllConversationsMessagingIntegrationsInstagramFn is the implementation for retrieving all conversations messaging integrations instagram in Genesys Cloud
func getAllConversationsMessagingIntegrationsInstagramFn(ctx context.Context, p *conversationsMessagingIntegrationsInstagramProxy) (*[]platformclientv2.Instagramintegration, *platformclientv2.APIResponse, error) {
	var allInstagramIntegrationRequests []platformclientv2.Instagramintegration
	const pageSize = 100

	instagramIntegrationRequests, resp, err := p.conversationsApi.GetConversationsMessagingIntegrationsInstagram(pageSize, 1, "", "", "")
	if err != nil {
		return nil, resp, fmt.Errorf("Failed to get instagram integration request: %v", err)
	}
	if instagramIntegrationRequests.Entities == nil || len(*instagramIntegrationRequests.Entities) == 0 {
		return &allInstagramIntegrationRequests, resp, nil
	}

	allInstagramIntegrationRequests = append(allInstagramIntegrationRequests, *instagramIntegrationRequests.Entities...)

	for pageNum := 2; pageNum <= *instagramIntegrationRequests.PageCount; pageNum++ {
		instagramIntegrationRequests, resp, err := p.conversationsApi.GetConversationsMessagingIntegrationsInstagram(pageSize, pageNum, "", "", "")
		if err != nil {
			return nil, resp, fmt.Errorf("Failed to get instagram integration request: %v", err)
		}

		if instagramIntegrationRequests.Entities == nil || len(*instagramIntegrationRequests.Entities) == 0 {
			break
		}

		allInstagramIntegrationRequests = append(allInstagramIntegrationRequests, *instagramIntegrationRequests.Entities...)

	}

	return &allInstagramIntegrationRequests, resp, nil
}

// getConversationsMessagingIntegrationsInstagramIdByNameFn is an implementation of the function to get a Genesys Cloud conversations messaging integrations instagram by name
func getConversationsMessagingIntegrationsInstagramIdByNameFn(ctx context.Context, p *conversationsMessagingIntegrationsInstagramProxy, name string) (id string, retryable bool, response *platformclientv2.APIResponse, err error) {
	instagramIntegrationRequests, resp, err := p.getAllConversationsMessagingIntegrationsInstagram(ctx)
	if err != nil {
		return "", false, resp, err
	}

	if instagramIntegrationRequests == nil || len(*instagramIntegrationRequests) == 0 {
		return "", true, resp, fmt.Errorf("No conversations messaging integrations instagram found with name %s", name)
	}

	for _, instagramIntegrationRequest := range *instagramIntegrationRequests {
		if *instagramIntegrationRequest.Name == name {
			log.Printf("Retrieved the conversations messaging integrations instagram id %s by name %s", *instagramIntegrationRequest.Id, name)
			return *instagramIntegrationRequest.Id, false, resp, nil
		}
	}

	return "", true, resp, fmt.Errorf("Unable to find conversations messaging integrations instagram with name %s", name)
}

// getConversationsMessagingIntegrationsInstagramByIdFn is an implementation of the function to get a Genesys Cloud conversations messaging integrations instagram by Id
func getConversationsMessagingIntegrationsInstagramByIdFn(ctx context.Context, p *conversationsMessagingIntegrationsInstagramProxy, id string) (conversationsMessagingIntegrationsInstagram *platformclientv2.Instagramintegration, response *platformclientv2.APIResponse, err error) {
	return p.conversationsApi.GetConversationsMessagingIntegrationsInstagramIntegrationId(id, "")
}

// updateConversationsMessagingIntegrationsInstagramFn is an implementation of the function to update a Genesys Cloud conversations messaging integrations instagram
func updateConversationsMessagingIntegrationsInstagramFn(ctx context.Context, p *conversationsMessagingIntegrationsInstagramProxy, id string, conversationsMessagingIntegrationsInstagram *platformclientv2.Instagramintegrationupdaterequest) (*platformclientv2.Instagramintegration, *platformclientv2.APIResponse, error) {
	return p.conversationsApi.PatchConversationsMessagingIntegrationsInstagramIntegrationId(id, *conversationsMessagingIntegrationsInstagram)
}

// deleteConversationsMessagingIntegrationsInstagramFn is an implementation function for deleting a Genesys Cloud conversations messaging integrations instagram
func deleteConversationsMessagingIntegrationsInstagramFn(ctx context.Context, p *conversationsMessagingIntegrationsInstagramProxy, id string) (response *platformclientv2.APIResponse, err error) {
	return p.conversationsApi.DeleteConversationsMessagingIntegrationsInstagramIntegrationId(id)
}
