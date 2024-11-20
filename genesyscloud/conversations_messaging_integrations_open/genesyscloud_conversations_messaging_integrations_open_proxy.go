package conversations_messaging_integrations_open

import (
	"context"
	"fmt"
	"log"

	"github.com/mypurecloud/platform-client-sdk-go/v146/platformclientv2"
)

/*
The genesyscloud_conversations_messaging_integrations_open_proxy.go file contains the proxy structures and methods that interact
with the Genesys Cloud SDK. We use composition here for each function on the proxy so individual functions can be stubbed
out during testing.
*/

// internalProxy holds a proxy instance that can be used throughout the package
var internalProxy *conversationsMessagingIntegrationsOpenProxy

// Type definitions for each func on our proxy so we can easily mock them out later
type createConversationsMessagingIntegrationsOpenFunc func(ctx context.Context, p *conversationsMessagingIntegrationsOpenProxy, openIntegrationRequest *platformclientv2.Openintegrationrequest) (*platformclientv2.Openintegration, *platformclientv2.APIResponse, error)
type getAllConversationsMessagingIntegrationsOpenFunc func(ctx context.Context, p *conversationsMessagingIntegrationsOpenProxy) (*[]platformclientv2.Openintegration, *platformclientv2.APIResponse, error)
type getConversationsMessagingIntegrationsOpenIdByNameFunc func(ctx context.Context, p *conversationsMessagingIntegrationsOpenProxy, name string) (id string, retryable bool, response *platformclientv2.APIResponse, err error)
type getConversationsMessagingIntegrationsOpenByIdFunc func(ctx context.Context, p *conversationsMessagingIntegrationsOpenProxy, id string) (openIntegrationRequest *platformclientv2.Openintegration, response *platformclientv2.APIResponse, err error)
type updateConversationsMessagingIntegrationsOpenFunc func(ctx context.Context, p *conversationsMessagingIntegrationsOpenProxy, id string, openIntegrationRequest *platformclientv2.Openintegrationupdaterequest) (*platformclientv2.Openintegration, *platformclientv2.APIResponse, error)
type deleteConversationsMessagingIntegrationsOpenFunc func(ctx context.Context, p *conversationsMessagingIntegrationsOpenProxy, id string) (response *platformclientv2.APIResponse, err error)

// conversationsMessagingIntegrationsOpenProxy contains all of the methods that call genesys cloud APIs.
type conversationsMessagingIntegrationsOpenProxy struct {
	clientConfig                                          *platformclientv2.Configuration
	conversationsApi                                      *platformclientv2.ConversationsApi
	createConversationsMessagingIntegrationsOpenAttr      createConversationsMessagingIntegrationsOpenFunc
	getAllConversationsMessagingIntegrationsOpenAttr      getAllConversationsMessagingIntegrationsOpenFunc
	getConversationsMessagingIntegrationsOpenIdByNameAttr getConversationsMessagingIntegrationsOpenIdByNameFunc
	getConversationsMessagingIntegrationsOpenByIdAttr     getConversationsMessagingIntegrationsOpenByIdFunc
	updateConversationsMessagingIntegrationsOpenAttr      updateConversationsMessagingIntegrationsOpenFunc
	deleteConversationsMessagingIntegrationsOpenAttr      deleteConversationsMessagingIntegrationsOpenFunc
}

// newConversationsMessagingIntegrationsOpenProxy initializes the conversations messaging integrations open proxy with all of the data needed to communicate with Genesys Cloud
func newConversationsMessagingIntegrationsOpenProxy(clientConfig *platformclientv2.Configuration) *conversationsMessagingIntegrationsOpenProxy {
	api := platformclientv2.NewConversationsApiWithConfig(clientConfig)
	return &conversationsMessagingIntegrationsOpenProxy{
		clientConfig:     clientConfig,
		conversationsApi: api,
		createConversationsMessagingIntegrationsOpenAttr:      createConversationsMessagingIntegrationsOpenFn,
		getAllConversationsMessagingIntegrationsOpenAttr:      getAllConversationsMessagingIntegrationsOpenFn,
		getConversationsMessagingIntegrationsOpenIdByNameAttr: getConversationsMessagingIntegrationsOpenIdByNameFn,
		getConversationsMessagingIntegrationsOpenByIdAttr:     getConversationsMessagingIntegrationsOpenByIdFn,
		updateConversationsMessagingIntegrationsOpenAttr:      updateConversationsMessagingIntegrationsOpenFn,
		deleteConversationsMessagingIntegrationsOpenAttr:      deleteConversationsMessagingIntegrationsOpenFn,
	}
}

// getConversationsMessagingIntegrationsOpenProxy acts as a singleton to for the internalProxy.  It also ensures
// that we can still proxy our tests by directly setting internalProxy package variable
func getConversationsMessagingIntegrationsOpenProxy(clientConfig *platformclientv2.Configuration) *conversationsMessagingIntegrationsOpenProxy {
	if internalProxy == nil {
		internalProxy = newConversationsMessagingIntegrationsOpenProxy(clientConfig)
	}

	return internalProxy
}

// createConversationsMessagingIntegrationsOpen creates a Genesys Cloud conversations messaging integrations open
func (p *conversationsMessagingIntegrationsOpenProxy) createConversationsMessagingIntegrationsOpen(ctx context.Context, conversationsMessagingIntegrationsOpen *platformclientv2.Openintegrationrequest) (*platformclientv2.Openintegration, *platformclientv2.APIResponse, error) {
	return p.createConversationsMessagingIntegrationsOpenAttr(ctx, p, conversationsMessagingIntegrationsOpen)
}

// getConversationsMessagingIntegrationsOpen retrieves all Genesys Cloud conversations messaging integrations open
func (p *conversationsMessagingIntegrationsOpenProxy) getAllConversationsMessagingIntegrationsOpen(ctx context.Context) (*[]platformclientv2.Openintegration, *platformclientv2.APIResponse, error) {
	return p.getAllConversationsMessagingIntegrationsOpenAttr(ctx, p)
}

// getConversationsMessagingIntegrationsOpenIdByName returns a single Genesys Cloud conversations messaging integrations open by a name
func (p *conversationsMessagingIntegrationsOpenProxy) getConversationsMessagingIntegrationsOpenIdByName(ctx context.Context, name string) (id string, retryable bool, response *platformclientv2.APIResponse, err error) {
	return p.getConversationsMessagingIntegrationsOpenIdByNameAttr(ctx, p, name)
}

// getConversationsMessagingIntegrationsOpenById returns a single Genesys Cloud conversations messaging integrations open by Id
func (p *conversationsMessagingIntegrationsOpenProxy) getConversationsMessagingIntegrationsOpenById(ctx context.Context, id string) (conversationsMessagingIntegrationsOpen *platformclientv2.Openintegration, response *platformclientv2.APIResponse, err error) {
	return p.getConversationsMessagingIntegrationsOpenByIdAttr(ctx, p, id)
}

// updateConversationsMessagingIntegrationsOpen updates a Genesys Cloud conversations messaging integrations open
func (p *conversationsMessagingIntegrationsOpenProxy) updateConversationsMessagingIntegrationsOpen(ctx context.Context, id string, conversationsMessagingIntegrationsOpen *platformclientv2.Openintegrationupdaterequest) (*platformclientv2.Openintegration, *platformclientv2.APIResponse, error) {
	return p.updateConversationsMessagingIntegrationsOpenAttr(ctx, p, id, conversationsMessagingIntegrationsOpen)
}

// deleteConversationsMessagingIntegrationsOpen deletes a Genesys Cloud conversations messaging integrations open by Id
func (p *conversationsMessagingIntegrationsOpenProxy) deleteConversationsMessagingIntegrationsOpen(ctx context.Context, id string) (response *platformclientv2.APIResponse, err error) {
	return p.deleteConversationsMessagingIntegrationsOpenAttr(ctx, p, id)
}

// createConversationsMessagingIntegrationsOpenFn is an implementation function for creating a Genesys Cloud conversations messaging integrations open
func createConversationsMessagingIntegrationsOpenFn(ctx context.Context, p *conversationsMessagingIntegrationsOpenProxy, conversationsMessagingIntegrationsOpen *platformclientv2.Openintegrationrequest) (*platformclientv2.Openintegration, *platformclientv2.APIResponse, error) {
	return p.conversationsApi.PostConversationsMessagingIntegrationsOpen(*conversationsMessagingIntegrationsOpen)
}

// getAllConversationsMessagingIntegrationsOpenFn is the implementation for retrieving all conversations messaging integrations open in Genesys Cloud
func getAllConversationsMessagingIntegrationsOpenFn(ctx context.Context, p *conversationsMessagingIntegrationsOpenProxy) (*[]platformclientv2.Openintegration, *platformclientv2.APIResponse, error) {
	var allOpenIntegrationRequests []platformclientv2.Openintegration
	const pageSize = 100

	openIntegrationRequests, resp, err := p.conversationsApi.GetConversationsMessagingIntegrationsOpen(pageSize, 1, "", "", "")
	if err != nil {
		return nil, resp, fmt.Errorf("Failed to get open integration request: %v", err)
	}
	if openIntegrationRequests.Entities == nil || len(*openIntegrationRequests.Entities) == 0 {
		return &allOpenIntegrationRequests, resp, nil
	}
	for _, openIntegrationRequest := range *openIntegrationRequests.Entities {
		allOpenIntegrationRequests = append(allOpenIntegrationRequests, openIntegrationRequest)
	}

	for pageNum := 2; pageNum <= *openIntegrationRequests.PageCount; pageNum++ {
		openIntegrationRequests, resp, err := p.conversationsApi.GetConversationsMessagingIntegrationsOpen(pageSize, pageNum, "", "", "")
		if err != nil {
			return nil, resp, fmt.Errorf("Failed to get open integration request: %v", err)
		}

		if openIntegrationRequests.Entities == nil || len(*openIntegrationRequests.Entities) == 0 {
			break
		}

		for _, openIntegrationRequest := range *openIntegrationRequests.Entities {
			allOpenIntegrationRequests = append(allOpenIntegrationRequests, openIntegrationRequest)
		}
	}

	return &allOpenIntegrationRequests, resp, nil
}

// getConversationsMessagingIntegrationsOpenIdByNameFn is an implementation of the function to get a Genesys Cloud conversations messaging integrations open by name
func getConversationsMessagingIntegrationsOpenIdByNameFn(ctx context.Context, p *conversationsMessagingIntegrationsOpenProxy, name string) (id string, retryable bool, response *platformclientv2.APIResponse, err error) {
	openIntegrationRequests, resp, err := getAllConversationsMessagingIntegrationsOpenFn(ctx, p)
	if err != nil {
		return "", false, resp, err
	}

	if openIntegrationRequests == nil || len(*openIntegrationRequests) == 0 {
		return "", true, resp, fmt.Errorf("No conversations messaging integrations open found with name %s", name)
	}

	for _, openIntegrationRequest := range *openIntegrationRequests {
		if *openIntegrationRequest.Name == name {
			log.Printf("Retrieved the conversations messaging integrations open id %s by name %s", *openIntegrationRequest.Id, name)
			return *openIntegrationRequest.Id, false, resp, nil
		}
	}

	return "", true, resp, fmt.Errorf("Unable to find conversations messaging integrations open with name %s", name)
}

// getConversationsMessagingIntegrationsOpenByIdFn is an implementation of the function to get a Genesys Cloud conversations messaging integrations open by Id
func getConversationsMessagingIntegrationsOpenByIdFn(ctx context.Context, p *conversationsMessagingIntegrationsOpenProxy, id string) (conversationsMessagingIntegrationsOpen *platformclientv2.Openintegration, response *platformclientv2.APIResponse, err error) {
	openIntegration, resp, err := p.conversationsApi.GetConversationsMessagingIntegrationsOpenIntegrationId(id, "")
	if err != nil {
		return nil, resp, fmt.Errorf("Failed to retrieve conversations messaging integrations open by id %s: %s", id, err)
	}

	return openIntegration, resp, nil
}

// updateConversationsMessagingIntegrationsOpenFn is an implementation of the function to update a Genesys Cloud conversations messaging integrations open
func updateConversationsMessagingIntegrationsOpenFn(ctx context.Context, p *conversationsMessagingIntegrationsOpenProxy, id string, conversationsMessagingIntegrationsOpen *platformclientv2.Openintegrationupdaterequest) (*platformclientv2.Openintegration, *platformclientv2.APIResponse, error) {
	return p.conversationsApi.PatchConversationsMessagingIntegrationsOpenIntegrationId(id, *conversationsMessagingIntegrationsOpen)
}

// deleteConversationsMessagingIntegrationsOpenFn is an implementation function for deleting a Genesys Cloud conversations messaging integrations open
func deleteConversationsMessagingIntegrationsOpenFn(ctx context.Context, p *conversationsMessagingIntegrationsOpenProxy, id string) (response *platformclientv2.APIResponse, err error) {
	return p.conversationsApi.DeleteConversationsMessagingIntegrationsOpenIntegrationId(id)
}
