package conversations_messaging_integrations_apple

import (
	"context"
	"fmt"
	"log"

	"github.com/mypurecloud/platform-client-sdk-go/v171/platformclientv2"
)

/*
The genesyscloud_apple_integration_proxy.go file contains the proxy structures and methods that interact
with the Genesys Cloud SDK. We use composition here for each function on the proxy so individual functions can be stubbed
out during testing.
*/

// internalProxy holds a proxy instance that can be used throughout the package
var internalProxy *conversationsMessagingIntegrationsAppleProxy

// Type definitions for each func on our proxy so we can easily mock them out later
type createConversationsMessagingIntegrationsAppleFunc func(ctx context.Context, p *conversationsMessagingIntegrationsAppleProxy, request *platformclientv2.Appleintegrationrequest) (*platformclientv2.Appleintegration, *platformclientv2.APIResponse, error)
type getAllConversationsMessagingIntegrationsAppleFunc func(ctx context.Context, p *conversationsMessagingIntegrationsAppleProxy) (*[]platformclientv2.Appleintegration, *platformclientv2.APIResponse, error)
type getConversationsMessagingIntegrationsAppleIdByNameFunc func(ctx context.Context, p *conversationsMessagingIntegrationsAppleProxy, name string) (string, *platformclientv2.APIResponse, bool, error)
type getConversationsMessagingIntegrationsAppleByIdFunc func(ctx context.Context, p *conversationsMessagingIntegrationsAppleProxy, id string) (*platformclientv2.Appleintegration, *platformclientv2.APIResponse, error)
type updateConversationsMessagingIntegrationsAppleFunc func(ctx context.Context, p *conversationsMessagingIntegrationsAppleProxy, id string, request *platformclientv2.Appleintegrationupdaterequest) (*platformclientv2.Appleintegration, *platformclientv2.APIResponse, error)
type deleteConversationsMessagingIntegrationsAppleFunc func(ctx context.Context, p *conversationsMessagingIntegrationsAppleProxy, id string) (*platformclientv2.APIResponse, error)

// appleIntegrationProxy contains all of the methods that call genesys cloud APIs.
type conversationsMessagingIntegrationsAppleProxy struct {
	clientConfig                                           *platformclientv2.Configuration
	conversationsApi                                       *platformclientv2.ConversationsApi
	createConversationsMessagingIntegrationsAppleAttr      createConversationsMessagingIntegrationsAppleFunc
	getAllConversationsMessagingIntegrationsAppleAttr      getAllConversationsMessagingIntegrationsAppleFunc
	getConversationsMessagingIntegrationsAppleIdByNameAttr getConversationsMessagingIntegrationsAppleIdByNameFunc
	getConversationsMessagingIntegrationsAppleByIdAttr     getConversationsMessagingIntegrationsAppleByIdFunc
	updateConversationsMessagingIntegrationsAppleAttr      updateConversationsMessagingIntegrationsAppleFunc
	deleteConversationsMessagingIntegrationsAppleAttr      deleteConversationsMessagingIntegrationsAppleFunc
}

// newAppleIntegrationProxy initializes the apple integration proxy with all of the data needed to communicate with Genesys Cloud
func newConversationsMessagingIntegrationsAppleProxy(clientConfig *platformclientv2.Configuration) *conversationsMessagingIntegrationsAppleProxy {
	api := platformclientv2.NewConversationsApiWithConfig(clientConfig)
	return &conversationsMessagingIntegrationsAppleProxy{
		clientConfig:     clientConfig,
		conversationsApi: api,
		createConversationsMessagingIntegrationsAppleAttr:      createConversationsMessagingIntegrationsAppleFn,
		getAllConversationsMessagingIntegrationsAppleAttr:      getAllConversationsMessagingIntegrationsAppleFn,
		getConversationsMessagingIntegrationsAppleIdByNameAttr: getConversationsMessagingIntegrationsAppleIdByNameFn,
		getConversationsMessagingIntegrationsAppleByIdAttr:     getConversationsMessagingIntegrationsAppleByIdFn,
		updateConversationsMessagingIntegrationsAppleAttr:      updateConversationsMessagingIntegrationsAppleFn,
		deleteConversationsMessagingIntegrationsAppleAttr:      deleteConversationsMessagingIntegrationsAppleFn,
	}
}

// getAppleIntegrationProxy acts as a singleton to for the internalProxy.  It also ensures
// that we can still proxy our tests by directly setting internalProxy package variable
func getConversationsMessagingIntegrationsAppleProxy(clientConfig *platformclientv2.Configuration) *conversationsMessagingIntegrationsAppleProxy {
	if internalProxy == nil {
		internalProxy = newConversationsMessagingIntegrationsAppleProxy(clientConfig)
	}

	return internalProxy
}

// createAppleIntegration creates a Genesys Cloud apple integration
func (p *conversationsMessagingIntegrationsAppleProxy) createConversationsMessagingIntegrationsApple(ctx context.Context, request *platformclientv2.Appleintegrationrequest) (*platformclientv2.Appleintegration, *platformclientv2.APIResponse, error) {
	return p.createConversationsMessagingIntegrationsAppleAttr(ctx, p, request)
}

// getAppleIntegration retrieves all Genesys Cloud apple integration
func (p *conversationsMessagingIntegrationsAppleProxy) getAllConversationsMessagingIntegrationsApple(ctx context.Context) (*[]platformclientv2.Appleintegration, *platformclientv2.APIResponse, error) {
	return p.getAllConversationsMessagingIntegrationsAppleAttr(ctx, p)
}

// getAppleIntegrationIdByName returns a single Genesys Cloud apple integration by a name
func (p *conversationsMessagingIntegrationsAppleProxy) getConversationsMessagingIntegrationsAppleIdByName(ctx context.Context, name string) (string, *platformclientv2.APIResponse, bool, error) {
	return p.getConversationsMessagingIntegrationsAppleIdByNameAttr(ctx, p, name)
}

// getAppleIntegrationById returns a single Genesys Cloud apple integration by Id
func (p *conversationsMessagingIntegrationsAppleProxy) getConversationsMessagingIntegrationsAppleById(ctx context.Context, id string) (*platformclientv2.Appleintegration, *platformclientv2.APIResponse, error) {
	return p.getConversationsMessagingIntegrationsAppleByIdAttr(ctx, p, id)
}

// updateAppleIntegration updates a Genesys Cloud apple integration
func (p *conversationsMessagingIntegrationsAppleProxy) updateConversationsMessagingIntegrationsApple(ctx context.Context, id string, request *platformclientv2.Appleintegrationupdaterequest) (*platformclientv2.Appleintegration, *platformclientv2.APIResponse, error) {
	return p.updateConversationsMessagingIntegrationsAppleAttr(ctx, p, id, request)
}

// deleteAppleIntegration deletes a Genesys Cloud apple integration by Id
func (p *conversationsMessagingIntegrationsAppleProxy) deleteConversationsMessagingIntegrationsApple(ctx context.Context, id string) (*platformclientv2.APIResponse, error) {
	return p.deleteConversationsMessagingIntegrationsAppleAttr(ctx, p, id)
}

// createConversationsMessagingIntegrationsAppleFn is an implementation function for creating a Genesys Cloud apple integration
func createConversationsMessagingIntegrationsAppleFn(ctx context.Context, p *conversationsMessagingIntegrationsAppleProxy, request *platformclientv2.Appleintegrationrequest) (*platformclientv2.Appleintegration, *platformclientv2.APIResponse, error) {
	return p.conversationsApi.PostConversationsMessagingIntegrationsApple(*request)
}

// getAllConversationsMessagingIntegrationsAppleFn is the implementation for retrieving all apple integration in Genesys Cloud
func getAllConversationsMessagingIntegrationsAppleFn(ctx context.Context, p *conversationsMessagingIntegrationsAppleProxy) (*[]platformclientv2.Appleintegration, *platformclientv2.APIResponse, error) {
	var allAppleIntegrations []platformclientv2.Appleintegration
	const pageSize = 100

	appleIntegrations, resp, err := p.conversationsApi.GetConversationsMessagingIntegrationsApple(pageSize, 1, "", "", "")
	if err != nil {
		return nil, resp, err
	}
	if appleIntegrations.Entities == nil || len(*appleIntegrations.Entities) == 0 {
		return &allAppleIntegrations, resp, nil
	}
	for _, appleIntegration := range *appleIntegrations.Entities {
		allAppleIntegrations = append(allAppleIntegrations, appleIntegration)
	}

	for pageNum := 2; pageNum <= *appleIntegrations.PageCount; pageNum++ {
		appleIntegrations, _, err := p.conversationsApi.GetConversationsMessagingIntegrationsApple(pageSize, pageNum, "", "", "")
		if err != nil {
			return nil, resp, err
		}

		if appleIntegrations.Entities == nil || len(*appleIntegrations.Entities) == 0 {
			break
		}

		for _, appleIntegration := range *appleIntegrations.Entities {
			allAppleIntegrations = append(allAppleIntegrations, appleIntegration)
		}
	}

	return &allAppleIntegrations, resp, nil
}

// getConversationsMessagingIntegrationsAppleIdByNameFn is an implementation of the function to get a Genesys Cloud apple integration by name
func getConversationsMessagingIntegrationsAppleIdByNameFn(ctx context.Context, p *conversationsMessagingIntegrationsAppleProxy, name string) (string, *platformclientv2.APIResponse, bool, error) {
	appleIntegrations, resp, err := p.getAllConversationsMessagingIntegrationsApple(ctx)
	if err != nil {
		return "", resp, false, err
	}

	if appleIntegrations == nil || len(*appleIntegrations) == 0 {
		return "", resp, true, fmt.Errorf("No apple integrations found")
	}

	for _, appleIntegration := range *appleIntegrations {
		if *appleIntegration.Name == name {
			log.Printf("Retrieved the apple integration id %s by name %s", *appleIntegration.Id, name)
			return *appleIntegration.Id, resp, false, nil
		}
	}

	return "", resp, true, fmt.Errorf("Unable to find apple integration with name %s", name)
}

// getConversationsMessagingIntegrationsAppleByIdFn is an implementation of the function to get a Genesys Cloud apple integration by Id
func getConversationsMessagingIntegrationsAppleByIdFn(ctx context.Context, p *conversationsMessagingIntegrationsAppleProxy, id string) (*platformclientv2.Appleintegration, *platformclientv2.APIResponse, error) {
	return p.conversationsApi.GetConversationsMessagingIntegrationsAppleIntegrationId(id, "")
}

// updateConversationsMessagingIntegrationsAppleFn is an implementation of the function to update a Genesys Cloud apple integration
func updateConversationsMessagingIntegrationsAppleFn(ctx context.Context, p *conversationsMessagingIntegrationsAppleProxy, id string, request *platformclientv2.Appleintegrationupdaterequest) (*platformclientv2.Appleintegration, *platformclientv2.APIResponse, error) {
	return p.conversationsApi.PatchConversationsMessagingIntegrationsAppleIntegrationId(id, *request)
}

// deleteConversationsMessagingIntegrationsAppleFn is an implementation function for deleting a Genesys Cloud apple integration
func deleteConversationsMessagingIntegrationsAppleFn(ctx context.Context, p *conversationsMessagingIntegrationsAppleProxy, id string) (*platformclientv2.APIResponse, error) {
	return p.conversationsApi.DeleteConversationsMessagingIntegrationsAppleIntegrationId(id)
}
