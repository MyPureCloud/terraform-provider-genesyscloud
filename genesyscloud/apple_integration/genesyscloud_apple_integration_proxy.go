package apple_integration

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
var internalProxy *appleIntegrationProxy

// Type definitions for each func on our proxy so we can easily mock them out later
type createAppleIntegrationFunc func(ctx context.Context, p *appleIntegrationProxy, appleIntegration *platformclientv2.Appleintegration) (*platformclientv2.Appleintegration, *platformclientv2.APIResponse, error)
type getAllAppleIntegrationFunc func(ctx context.Context, p *appleIntegrationProxy) (*[]platformclientv2.Appleintegration, *platformclientv2.APIResponse, error)
type getAppleIntegrationIdByNameFunc func(ctx context.Context, p *appleIntegrationProxy, name string) (string, *platformclientv2.APIResponse, bool, error)
type getAppleIntegrationByIdFunc func(ctx context.Context, p *appleIntegrationProxy, id string) (*platformclientv2.Appleintegration, *platformclientv2.APIResponse, error)
type updateAppleIntegrationFunc func(ctx context.Context, p *appleIntegrationProxy, id string, appleIntegration *platformclientv2.Appleintegration) (*platformclientv2.Appleintegration, *platformclientv2.APIResponse, error)
type deleteAppleIntegrationFunc func(ctx context.Context, p *appleIntegrationProxy, id string) (*platformclientv2.APIResponse, error)

// appleIntegrationProxy contains all of the methods that call genesys cloud APIs.
type appleIntegrationProxy struct {
	clientConfig                    *platformclientv2.Configuration
	conversationsApi                *platformclientv2.ConversationsApi
	createAppleIntegrationAttr      createAppleIntegrationFunc
	getAllAppleIntegrationAttr      getAllAppleIntegrationFunc
	getAppleIntegrationIdByNameAttr getAppleIntegrationIdByNameFunc
	getAppleIntegrationByIdAttr     getAppleIntegrationByIdFunc
	updateAppleIntegrationAttr      updateAppleIntegrationFunc
	deleteAppleIntegrationAttr      deleteAppleIntegrationFunc
}

// newAppleIntegrationProxy initializes the apple integration proxy with all of the data needed to communicate with Genesys Cloud
func newAppleIntegrationProxy(clientConfig *platformclientv2.Configuration) *appleIntegrationProxy {
	api := platformclientv2.NewConversationsApiWithConfig(clientConfig)
	return &appleIntegrationProxy{
		clientConfig:                    clientConfig,
		conversationsApi:                api,
		createAppleIntegrationAttr:      createAppleIntegrationFn,
		getAllAppleIntegrationAttr:      getAllAppleIntegrationFn,
		getAppleIntegrationIdByNameAttr: getAppleIntegrationIdByNameFn,
		getAppleIntegrationByIdAttr:     getAppleIntegrationByIdFn,
		updateAppleIntegrationAttr:      updateAppleIntegrationFn,
		deleteAppleIntegrationAttr:      deleteAppleIntegrationFn,
	}
}

// getAppleIntegrationProxy acts as a singleton to for the internalProxy.  It also ensures
// that we can still proxy our tests by directly setting internalProxy package variable
func getAppleIntegrationProxy(clientConfig *platformclientv2.Configuration) *appleIntegrationProxy {
	if internalProxy == nil {
		internalProxy = newAppleIntegrationProxy(clientConfig)
	}

	return internalProxy
}

// createAppleIntegration creates a Genesys Cloud apple integration
func (p *appleIntegrationProxy) createAppleIntegration(ctx context.Context, appleIntegration *platformclientv2.Appleintegration) (*platformclientv2.Appleintegration, *platformclientv2.APIResponse, error) {
	return p.createAppleIntegrationAttr(ctx, p, appleIntegration)
}

// getAppleIntegration retrieves all Genesys Cloud apple integration
func (p *appleIntegrationProxy) getAllAppleIntegration(ctx context.Context) (*[]platformclientv2.Appleintegration, *platformclientv2.APIResponse, error) {
	return p.getAllAppleIntegrationAttr(ctx, p)
}

// getAppleIntegrationIdByName returns a single Genesys Cloud apple integration by a name
func (p *appleIntegrationProxy) getAppleIntegrationIdByName(ctx context.Context, name string) (string, *platformclientv2.APIResponse, bool, error) {
	return p.getAppleIntegrationIdByNameAttr(ctx, p, name)
}

// getAppleIntegrationById returns a single Genesys Cloud apple integration by Id
func (p *appleIntegrationProxy) getAppleIntegrationById(ctx context.Context, id string) (*platformclientv2.Appleintegration, *platformclientv2.APIResponse, error) {
	return p.getAppleIntegrationByIdAttr(ctx, p, id)
}

// updateAppleIntegration updates a Genesys Cloud apple integration
func (p *appleIntegrationProxy) updateAppleIntegration(ctx context.Context, id string, appleIntegration *platformclientv2.Appleintegration) (*platformclientv2.Appleintegration, *platformclientv2.APIResponse, error) {
	return p.updateAppleIntegrationAttr(ctx, p, id, appleIntegration)
}

// deleteAppleIntegration deletes a Genesys Cloud apple integration by Id
func (p *appleIntegrationProxy) deleteAppleIntegration(ctx context.Context, id string) (*platformclientv2.APIResponse, error) {
	return p.deleteAppleIntegrationAttr(ctx, p, id)
}

// createAppleIntegrationFn is an implementation function for creating a Genesys Cloud apple integration
func createAppleIntegrationFn(ctx context.Context, p *appleIntegrationProxy, appleIntegration *platformclientv2.Appleintegration) (*platformclientv2.Appleintegration, *platformclientv2.APIResponse, error) {
	// Convert Appleintegration to Appleintegrationrequest
	request := platformclientv2.Appleintegrationrequest{
		Name:                  appleIntegration.Name,
		SupportedContent:      appleIntegration.SupportedContent,
		MessagesForBusinessId: appleIntegration.MessagesForBusinessId,
		BusinessName:          appleIntegration.BusinessName,
		LogoUrl:               appleIntegration.LogoUrl,
		AppleIMessageApp:      appleIntegration.AppleIMessageApp,
		AppleAuthentication:   appleIntegration.AppleAuthentication,
		ApplePay:              appleIntegration.ApplePay,
	}
	return p.conversationsApi.PostConversationsMessagingIntegrationsApple(request)
}

// getAllAppleIntegrationFn is the implementation for retrieving all apple integration in Genesys Cloud
func getAllAppleIntegrationFn(ctx context.Context, p *appleIntegrationProxy) (*[]platformclientv2.Appleintegration, *platformclientv2.APIResponse, error) {
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

// getAppleIntegrationIdByNameFn is an implementation of the function to get a Genesys Cloud apple integration by name
func getAppleIntegrationIdByNameFn(ctx context.Context, p *appleIntegrationProxy, name string) (string, *platformclientv2.APIResponse, bool, error) {
	appleIntegrations, resp, err := p.conversationsApi.GetConversationsMessagingIntegrationsApple(100, 1, "", "", "")
	if err != nil {
		return "", resp, false, err
	}

	if appleIntegrations.Entities == nil || len(*appleIntegrations.Entities) == 0 {
		return "", resp, true, err
	}

	for _, appleIntegration := range *appleIntegrations.Entities {
		if *appleIntegration.Name == name {
			log.Printf("Retrieved the apple integration id %s by name %s", *appleIntegration.Id, name)
			return *appleIntegration.Id, resp, false, nil
		}
	}

	return "", resp, true, fmt.Errorf("Unable to find apple integration with name %s", name)
}

// getAppleIntegrationByIdFn is an implementation of the function to get a Genesys Cloud apple integration by Id
func getAppleIntegrationByIdFn(ctx context.Context, p *appleIntegrationProxy, id string) (*platformclientv2.Appleintegration, *platformclientv2.APIResponse, error) {
	return p.conversationsApi.GetConversationsMessagingIntegrationsAppleIntegrationId(id, "")
}

// updateAppleIntegrationFn is an implementation of the function to update a Genesys Cloud apple integration
func updateAppleIntegrationFn(ctx context.Context, p *appleIntegrationProxy, id string, appleIntegration *platformclientv2.Appleintegration) (*platformclientv2.Appleintegration, *platformclientv2.APIResponse, error) {
	// Convert Appleintegration to Appleintegrationupdaterequest
	request := platformclientv2.Appleintegrationupdaterequest{
		Name:                appleIntegration.Name,
		SupportedContent:    appleIntegration.SupportedContent,
		BusinessName:        appleIntegration.BusinessName,
		LogoUrl:             appleIntegration.LogoUrl,
		AppleIMessageApp:    appleIntegration.AppleIMessageApp,
		AppleAuthentication: appleIntegration.AppleAuthentication,
		ApplePay:            appleIntegration.ApplePay,
	}
	return p.conversationsApi.PatchConversationsMessagingIntegrationsAppleIntegrationId(id, request)
}

// deleteAppleIntegrationFn is an implementation function for deleting a Genesys Cloud apple integration
func deleteAppleIntegrationFn(ctx context.Context, p *appleIntegrationProxy, id string) (*platformclientv2.APIResponse, error) {
	return p.conversationsApi.DeleteConversationsMessagingIntegrationsAppleIntegrationId(id)
}
