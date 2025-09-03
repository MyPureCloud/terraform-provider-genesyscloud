package conversations_messaging_integrations_whatsapp

import (
	"context"
	"fmt"
	"log"

	rc "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_cache"

	"github.com/mypurecloud/platform-client-sdk-go/v165/platformclientv2"
)

/*
The genesyscloud_conversations_messaging_integrations_whatsapp_proxy.go file contains the proxy structures and methods that interact
with the Genesys Cloud SDK. We use composition here for each function on the proxy so individual functions can be stubbed
out during testing.
*/

// internalProxy holds a proxy instance that can be used throughout the package
var internalProxy *conversationsMessagingIntegrationsWhatsappProxy

// Type definitions for each func on our proxy so we can easily mock them out later
type createConversationsMessagingIntegrationsWhatsappFunc func(ctx context.Context, p *conversationsMessagingIntegrationsWhatsappProxy, whatsAppEmbeddedSignupIntegrationRequest *platformclientv2.Whatsappembeddedsignupintegrationrequest) (*platformclientv2.Whatsappintegration, *platformclientv2.APIResponse, error)
type getAllConversationsMessagingIntegrationsWhatsappFunc func(ctx context.Context, p *conversationsMessagingIntegrationsWhatsappProxy) (*[]platformclientv2.Whatsappintegration, *platformclientv2.APIResponse, error)
type getConversationsMessagingIntegrationsWhatsappIdByNameFunc func(ctx context.Context, p *conversationsMessagingIntegrationsWhatsappProxy, name string) (id string, retryable bool, response *platformclientv2.APIResponse, err error)
type getConversationsMessagingIntegrationsWhatsappByIdFunc func(ctx context.Context, p *conversationsMessagingIntegrationsWhatsappProxy, id string) (conversationsMessagingIntegrationsWhatsapp *platformclientv2.Whatsappintegration, response *platformclientv2.APIResponse, err error)
type updateConversationsMessagingIntegrationsWhatsappEmbeddedSignupFunc func(ctx context.Context, p *conversationsMessagingIntegrationsWhatsappProxy, id string, conversationsMessagingIntegrationsWhatsapp *platformclientv2.Whatsappembeddedsignupintegrationactivationrequest) (*platformclientv2.Whatsappintegration, *platformclientv2.APIResponse, error)
type updateConversationsMessagingIntegrationsWhatsappFunc func(ctx context.Context, p *conversationsMessagingIntegrationsWhatsappProxy, id string, whatsAppEmbeddedSignupIntegrationRequest *platformclientv2.Whatsappintegrationupdaterequest) (*platformclientv2.Whatsappintegration, *platformclientv2.APIResponse, error)
type deleteConversationsMessagingIntegrationsWhatsappFunc func(ctx context.Context, p *conversationsMessagingIntegrationsWhatsappProxy, id string) (response *platformclientv2.APIResponse, err error)

// conversationsMessagingIntegrationsWhatsappProxy contains all of the methods that call genesys cloud APIs.
type conversationsMessagingIntegrationsWhatsappProxy struct {
	clientConfig                                                       *platformclientv2.Configuration
	conversationsApi                                                   *platformclientv2.ConversationsApi
	createConversationsMessagingIntegrationsWhatsappAttr               createConversationsMessagingIntegrationsWhatsappFunc
	getAllConversationsMessagingIntegrationsWhatsappAttr               getAllConversationsMessagingIntegrationsWhatsappFunc
	getConversationsMessagingIntegrationsWhatsappIdByNameAttr          getConversationsMessagingIntegrationsWhatsappIdByNameFunc
	getConversationsMessagingIntegrationsWhatsappByIdAttr              getConversationsMessagingIntegrationsWhatsappByIdFunc
	updateConversationsMessagingIntegrationsWhatsappEmbeddedSignupAttr updateConversationsMessagingIntegrationsWhatsappEmbeddedSignupFunc
	updateConversationsMessagingIntegrationsWhatsappAttr               updateConversationsMessagingIntegrationsWhatsappFunc
	deleteConversationsMessagingIntegrationsWhatsappAttr               deleteConversationsMessagingIntegrationsWhatsappFunc
	whatsappCache                                                      rc.CacheInterface[platformclientv2.Whatsappintegration]
}

// newConversationsMessagingIntegrationsWhatsappProxy initializes the conversations messaging integrations whatsapp proxy with all of the data needed to communicate with Genesys Cloud
func newConversationsMessagingIntegrationsWhatsappProxy(clientConfig *platformclientv2.Configuration) *conversationsMessagingIntegrationsWhatsappProxy {
	api := platformclientv2.NewConversationsApiWithConfig(clientConfig)
	whatsappCache := rc.NewResourceCache[platformclientv2.Whatsappintegration]()
	return &conversationsMessagingIntegrationsWhatsappProxy{
		clientConfig:     clientConfig,
		conversationsApi: api,
		createConversationsMessagingIntegrationsWhatsappAttr:               createConversationsMessagingIntegrationsWhatsappFn,
		getAllConversationsMessagingIntegrationsWhatsappAttr:               getAllConversationsMessagingIntegrationsWhatsappFn,
		getConversationsMessagingIntegrationsWhatsappIdByNameAttr:          getConversationsMessagingIntegrationsWhatsappIdByNameFn,
		getConversationsMessagingIntegrationsWhatsappByIdAttr:              getConversationsMessagingIntegrationsWhatsappByIdFn,
		updateConversationsMessagingIntegrationsWhatsappEmbeddedSignupAttr: updateConversationsMessagingIntegrationsWhatsappEmbeddedSignupFn,
		updateConversationsMessagingIntegrationsWhatsappAttr:               updateConversationsMessagingIntegrationsWhatsappFn,
		deleteConversationsMessagingIntegrationsWhatsappAttr:               deleteConversationsMessagingIntegrationsWhatsappFn,
		whatsappCache: whatsappCache,
	}
}

// getConversationsMessagingIntegrationsWhatsappProxy acts as a singleton to for the internalProxy.  It also ensures
// that we can still proxy our tests by directly setting internalProxy package variable
func getConversationsMessagingIntegrationsWhatsappProxy(clientConfig *platformclientv2.Configuration) *conversationsMessagingIntegrationsWhatsappProxy {
	if internalProxy == nil {
		internalProxy = newConversationsMessagingIntegrationsWhatsappProxy(clientConfig)
	}

	return internalProxy
}

// createConversationsMessagingIntegrationsWhatsapp creates a Genesys Cloud conversations messaging integrations whatsapp
func (p *conversationsMessagingIntegrationsWhatsappProxy) createConversationsMessagingIntegrationsWhatsapp(ctx context.Context, conversationsMessagingIntegrationsWhatsapp *platformclientv2.Whatsappembeddedsignupintegrationrequest) (*platformclientv2.Whatsappintegration, *platformclientv2.APIResponse, error) {
	return p.createConversationsMessagingIntegrationsWhatsappAttr(ctx, p, conversationsMessagingIntegrationsWhatsapp)
}

// getConversationsMessagingIntegrationsWhatsapp retrieves all Genesys Cloud conversations messaging integrations whatsapp
func (p *conversationsMessagingIntegrationsWhatsappProxy) getAllConversationsMessagingIntegrationsWhatsapp(ctx context.Context) (*[]platformclientv2.Whatsappintegration, *platformclientv2.APIResponse, error) {
	return p.getAllConversationsMessagingIntegrationsWhatsappAttr(ctx, p)
}

// getConversationsMessagingIntegrationsWhatsappIdByName returns a single Genesys Cloud conversations messaging integrations whatsapp by a name
func (p *conversationsMessagingIntegrationsWhatsappProxy) getConversationsMessagingIntegrationsWhatsappIdByName(ctx context.Context, name string) (id string, retryable bool, response *platformclientv2.APIResponse, err error) {
	return p.getConversationsMessagingIntegrationsWhatsappIdByNameAttr(ctx, p, name)
}

// getConversationsMessagingIntegrationsWhatsappById returns a single Genesys Cloud conversations messaging integrations whatsapp by Id
func (p *conversationsMessagingIntegrationsWhatsappProxy) getConversationsMessagingIntegrationsWhatsappById(ctx context.Context, id string) (conversationsMessagingIntegrationsWhatsapp *platformclientv2.Whatsappintegration, response *platformclientv2.APIResponse, err error) {
	return p.getConversationsMessagingIntegrationsWhatsappByIdAttr(ctx, p, id)
}

func (p *conversationsMessagingIntegrationsWhatsappProxy) updateConversationsMessagingIntegrationsWhatsappEmbeddedSignup(ctx context.Context, id string, conversationsMessagingIntegrationsWhatsapp *platformclientv2.Whatsappembeddedsignupintegrationactivationrequest) (*platformclientv2.Whatsappintegration, *platformclientv2.APIResponse, error) {
	return p.updateConversationsMessagingIntegrationsWhatsappEmbeddedSignupAttr(ctx, p, id, conversationsMessagingIntegrationsWhatsapp)
}

// updateConversationsMessagingIntegrationsWhatsapp updates a Genesys Cloud conversations messaging integrations whatsapp
func (p *conversationsMessagingIntegrationsWhatsappProxy) updateConversationsMessagingIntegrationsWhatsapp(ctx context.Context, id string, conversationsMessagingIntegrationsWhatsapp *platformclientv2.Whatsappintegrationupdaterequest) (*platformclientv2.Whatsappintegration, *platformclientv2.APIResponse, error) {
	return p.updateConversationsMessagingIntegrationsWhatsappAttr(ctx, p, id, conversationsMessagingIntegrationsWhatsapp)
}

// deleteConversationsMessagingIntegrationsWhatsapp deletes a Genesys Cloud conversations messaging integrations whatsapp by Id
func (p *conversationsMessagingIntegrationsWhatsappProxy) deleteConversationsMessagingIntegrationsWhatsapp(ctx context.Context, id string) (response *platformclientv2.APIResponse, err error) {
	return p.deleteConversationsMessagingIntegrationsWhatsappAttr(ctx, p, id)
}

// createConversationsMessagingIntegrationsWhatsappFn is an implementation function for creating a Genesys Cloud conversations messaging integrations whatsapp
func createConversationsMessagingIntegrationsWhatsappFn(ctx context.Context, p *conversationsMessagingIntegrationsWhatsappProxy, conversationsMessagingIntegrationsWhatsapp *platformclientv2.Whatsappembeddedsignupintegrationrequest) (*platformclientv2.Whatsappintegration, *platformclientv2.APIResponse, error) {
	return p.conversationsApi.PostConversationsMessagingIntegrationsWhatsappEmbeddedsignup(*conversationsMessagingIntegrationsWhatsapp)
}

// getAllConversationsMessagingIntegrationsWhatsappFn is the implementation for retrieving all conversations messaging integrations whatsapp in Genesys Cloud
func getAllConversationsMessagingIntegrationsWhatsappFn(ctx context.Context, p *conversationsMessagingIntegrationsWhatsappProxy) (*[]platformclientv2.Whatsappintegration, *platformclientv2.APIResponse, error) {
	var allWhatsAppEmbeddedSignupIntegrationRequests []platformclientv2.Whatsappintegration
	const pageSize = 100

	whatsAppEmbeddedSignupIntegrationRequests, resp, err := p.conversationsApi.GetConversationsMessagingIntegrationsWhatsapp(pageSize, 1, "", "", "")
	if err != nil {
		return nil, resp, fmt.Errorf("Failed to get whats app embedded signup integration request: %v", err)
	}
	if whatsAppEmbeddedSignupIntegrationRequests.Entities == nil || len(*whatsAppEmbeddedSignupIntegrationRequests.Entities) == 0 {
		return &allWhatsAppEmbeddedSignupIntegrationRequests, resp, nil
	}

	allWhatsAppEmbeddedSignupIntegrationRequests = append(allWhatsAppEmbeddedSignupIntegrationRequests, *whatsAppEmbeddedSignupIntegrationRequests.Entities...)

	for pageNum := 2; pageNum <= *whatsAppEmbeddedSignupIntegrationRequests.PageCount; pageNum++ {
		whatsAppEmbeddedSignupIntegrationRequests, resp, err := p.conversationsApi.GetConversationsMessagingIntegrationsWhatsapp(pageSize, pageNum, "", "", "")
		if err != nil {
			return nil, resp, fmt.Errorf("Failed to get whats app embedded signup integration request: %v", err)
		}

		if whatsAppEmbeddedSignupIntegrationRequests.Entities == nil || len(*whatsAppEmbeddedSignupIntegrationRequests.Entities) == 0 {
			break
		}

		allWhatsAppEmbeddedSignupIntegrationRequests = append(allWhatsAppEmbeddedSignupIntegrationRequests, *whatsAppEmbeddedSignupIntegrationRequests.Entities...)
	}

	for _, whatsappRequest := range allWhatsAppEmbeddedSignupIntegrationRequests {
		rc.SetCache(p.whatsappCache, *whatsappRequest.Id, whatsappRequest)
	}

	return &allWhatsAppEmbeddedSignupIntegrationRequests, resp, nil
}

// getConversationsMessagingIntegrationsWhatsappIdByNameFn is an implementation of the function to get a Genesys Cloud conversations messaging integrations whatsapp by name
func getConversationsMessagingIntegrationsWhatsappIdByNameFn(ctx context.Context, p *conversationsMessagingIntegrationsWhatsappProxy, name string) (id string, retryable bool, response *platformclientv2.APIResponse, err error) {
	whatsAppEmbeddedSignupIntegrationRequests, resp, err := getAllConversationsMessagingIntegrationsWhatsappFn(ctx, p)
	if err != nil {
		return "", false, resp, err
	}

	if whatsAppEmbeddedSignupIntegrationRequests == nil || len(*whatsAppEmbeddedSignupIntegrationRequests) == 0 {
		return "", true, resp, fmt.Errorf("No conversations messaging integrations whatsapp found with name %s", name)
	}

	for _, whatsAppEmbeddedSignupIntegrationRequest := range *whatsAppEmbeddedSignupIntegrationRequests {
		if *whatsAppEmbeddedSignupIntegrationRequest.Name == name {
			log.Printf("Retrieved the conversations messaging integrations whatsapp id %s by name %s", *whatsAppEmbeddedSignupIntegrationRequest.Id, name)
			return *whatsAppEmbeddedSignupIntegrationRequest.Id, false, resp, nil
		}
	}

	return "", true, resp, fmt.Errorf("Unable to find conversations messaging integrations whatsapp with name %s", name)
}

// getConversationsMessagingIntegrationsWhatsappByIdFn is an implementation of the function to get a Genesys Cloud conversations messaging integrations whatsapp by Id
func getConversationsMessagingIntegrationsWhatsappByIdFn(ctx context.Context, p *conversationsMessagingIntegrationsWhatsappProxy, id string) (conversationsMessagingIntegrationsWhatsapp *platformclientv2.Whatsappintegration, response *platformclientv2.APIResponse, err error) {
	whatsapp := rc.GetCacheItem(p.whatsappCache, id)
	if whatsapp != nil {
		return whatsapp, nil, nil
	}
	return p.conversationsApi.GetConversationsMessagingIntegrationsWhatsappIntegrationId(id, "")
}

func updateConversationsMessagingIntegrationsWhatsappEmbeddedSignupFn(ctx context.Context, p *conversationsMessagingIntegrationsWhatsappProxy, id string, conversationsMessagingIntegrationsWhatsapp *platformclientv2.Whatsappembeddedsignupintegrationactivationrequest) (*platformclientv2.Whatsappintegration, *platformclientv2.APIResponse, error) {
	return p.conversationsApi.PatchConversationsMessagingIntegrationsWhatsappEmbeddedsignupIntegrationId(id, *conversationsMessagingIntegrationsWhatsapp)
}

// updateConversationsMessagingIntegrationsWhatsappFn is an implementation of the function to update a Genesys Cloud conversations messaging integrations whatsapp
func updateConversationsMessagingIntegrationsWhatsappFn(ctx context.Context, p *conversationsMessagingIntegrationsWhatsappProxy, id string, conversationsMessagingIntegrationsWhatsapp *platformclientv2.Whatsappintegrationupdaterequest) (*platformclientv2.Whatsappintegration, *platformclientv2.APIResponse, error) {
	return p.conversationsApi.PatchConversationsMessagingIntegrationsWhatsappIntegrationId(id, *conversationsMessagingIntegrationsWhatsapp)
}

// deleteConversationsMessagingIntegrationsWhatsappFn is an implementation function for deleting a Genesys Cloud conversations messaging integrations whatsapp
func deleteConversationsMessagingIntegrationsWhatsappFn(ctx context.Context, p *conversationsMessagingIntegrationsWhatsappProxy, id string) (response *platformclientv2.APIResponse, err error) {
	_, resp, err := p.conversationsApi.DeleteConversationsMessagingIntegrationsWhatsappIntegrationId(id)
	if err != nil {
		return resp, fmt.Errorf("Failed to delete conversations messaging integrations whatsapp: %s", err)
	}
	rc.DeleteCacheItem(p.whatsappCache, id)
	return resp, nil
}
