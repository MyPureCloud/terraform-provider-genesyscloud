package conversations_messaging_supportedcontent

import (
	"context"
	"fmt"
	"log"
	rc "terraform-provider-genesyscloud/genesyscloud/resource_cache"

	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

/*
The genesyscloud_conversations_messaging_supportedcontent_proxy.go file contains the proxy structures and methods that interact
with the Genesys Cloud SDK. We use composition here for each function on the proxy so individual functions can be stubbed
out during testing.
*/

// internalProxy holds a proxy instance that can be used throughout the package
var internalProxy *supportedContentProxy

// Type definitions for each func on our proxy so we can easily mock them out later
type createSupportedContentFunc func(ctx context.Context, p *supportedContentProxy, supportedContent *platformclientv2.Supportedcontent) (*platformclientv2.Supportedcontent, *platformclientv2.APIResponse, error)
type getAllSupportedContentFunc func(ctx context.Context, p *supportedContentProxy) (*[]platformclientv2.Supportedcontent, *platformclientv2.APIResponse, error)
type getSupportedContentIdByNameFunc func(ctx context.Context, p *supportedContentProxy, name string) (id string, retryable bool, response *platformclientv2.APIResponse, err error)
type getSupportedContentByIdFunc func(ctx context.Context, p *supportedContentProxy, id string) (supportedContent *platformclientv2.Supportedcontent, response *platformclientv2.APIResponse, err error)
type updateSupportedContentFunc func(ctx context.Context, p *supportedContentProxy, id string, supportedContent *platformclientv2.Supportedcontent) (*platformclientv2.Supportedcontent, *platformclientv2.APIResponse, error)
type deleteSupportedContentFunc func(ctx context.Context, p *supportedContentProxy, id string) (response *platformclientv2.APIResponse, err error)

// supportedContentProxy contains all of the methods that call genesys cloud APIs.
type supportedContentProxy struct {
	clientConfig                    *platformclientv2.Configuration
	conversationsApi                *platformclientv2.ConversationsApi
	createSupportedContentAttr      createSupportedContentFunc
	getAllSupportedContentAttr      getAllSupportedContentFunc
	getSupportedContentIdByNameAttr getSupportedContentIdByNameFunc
	getSupportedContentByIdAttr     getSupportedContentByIdFunc
	updateSupportedContentAttr      updateSupportedContentFunc
	deleteSupportedContentAttr      deleteSupportedContentFunc
	supportedContentCache           rc.CacheInterface[platformclientv2.Supportedcontent]
}

// newSupportedContentProxy initializes the supported content proxy with all of the data needed to communicate with Genesys Cloud
func newSupportedContentProxy(clientConfig *platformclientv2.Configuration) *supportedContentProxy {
	api := platformclientv2.NewConversationsApiWithConfig(clientConfig)
	supportedContentCache := rc.NewResourceCache[platformclientv2.Supportedcontent]()
	return &supportedContentProxy{
		clientConfig:                    clientConfig,
		conversationsApi:                api,
		createSupportedContentAttr:      createSupportedContentFn,
		getAllSupportedContentAttr:      getAllSupportedContentFn,
		getSupportedContentIdByNameAttr: getSupportedContentIdByNameFn,
		getSupportedContentByIdAttr:     getSupportedContentByIdFn,
		updateSupportedContentAttr:      updateSupportedContentFn,
		deleteSupportedContentAttr:      deleteSupportedContentFn,
		supportedContentCache:           supportedContentCache,
	}
}

// getSupportedContentProxy acts as a singleton to for the internalProxy.  It also ensures
// that we can still proxy our tests by directly setting internalProxy package variable
func getSupportedContentProxy(clientConfig *platformclientv2.Configuration) *supportedContentProxy {
	if internalProxy == nil {
		internalProxy = newSupportedContentProxy(clientConfig)
	}

	return internalProxy
}

// createSupportedContent creates a Genesys Cloud supported content
func (p *supportedContentProxy) createSupportedContent(ctx context.Context, supportedContent *platformclientv2.Supportedcontent) (*platformclientv2.Supportedcontent, *platformclientv2.APIResponse, error) {
	return p.createSupportedContentAttr(ctx, p, supportedContent)
}

// getSupportedContent retrieves all Genesys Cloud supported content
func (p *supportedContentProxy) getAllSupportedContent(ctx context.Context) (*[]platformclientv2.Supportedcontent, *platformclientv2.APIResponse, error) {
	return p.getAllSupportedContentAttr(ctx, p)
}

// getSupportedContentIdByName returns a single Genesys Cloud supported content by a name
func (p *supportedContentProxy) getSupportedContentIdByName(ctx context.Context, name string) (id string, retryable bool, response *platformclientv2.APIResponse, err error) {
	return p.getSupportedContentIdByNameAttr(ctx, p, name)
}

// getSupportedContentById returns a single Genesys Cloud supported content by Id
func (p *supportedContentProxy) getSupportedContentById(ctx context.Context, id string) (supportedContent *platformclientv2.Supportedcontent, response *platformclientv2.APIResponse, err error) {
	return p.getSupportedContentByIdAttr(ctx, p, id)
}

// updateSupportedContent updates a Genesys Cloud supported content
func (p *supportedContentProxy) updateSupportedContent(ctx context.Context, id string, supportedContent *platformclientv2.Supportedcontent) (*platformclientv2.Supportedcontent, *platformclientv2.APIResponse, error) {
	return p.updateSupportedContentAttr(ctx, p, id, supportedContent)
}

// deleteSupportedContent deletes a Genesys Cloud supported content by Id
func (p *supportedContentProxy) deleteSupportedContent(ctx context.Context, id string) (response *platformclientv2.APIResponse, err error) {
	return p.deleteSupportedContentAttr(ctx, p, id)
}

// createSupportedContentFn is an implementation function for creating a Genesys Cloud supported content
func createSupportedContentFn(ctx context.Context, p *supportedContentProxy, supportedContent *platformclientv2.Supportedcontent) (*platformclientv2.Supportedcontent, *platformclientv2.APIResponse, error) {
	return p.conversationsApi.PostConversationsMessagingSupportedcontent(*supportedContent)
}

// getAllSupportedContentFn is the implementation for retrieving all supported content in Genesys Cloud
func getAllSupportedContentFn(ctx context.Context, p *supportedContentProxy) (*[]platformclientv2.Supportedcontent, *platformclientv2.APIResponse, error) {
	var allSupportedContents []platformclientv2.Supportedcontent
	const pageSize = 100

	supportedContents, resp, err := p.conversationsApi.GetConversationsMessagingSupportedcontent(pageSize, 1)
	if err != nil {
		return nil, resp, fmt.Errorf("Failed to get supported content: %v", err)
	}
	if supportedContents.Entities == nil || len(*supportedContents.Entities) == 0 {
		return &allSupportedContents, resp, nil
	}

	allSupportedContents = append(allSupportedContents, *supportedContents.Entities...)

	for pageNum := 2; pageNum <= *supportedContents.PageCount; pageNum++ {
		supportedContents, _, err := p.conversationsApi.GetConversationsMessagingSupportedcontent(pageSize, pageNum)
		if err != nil {
			return nil, resp, fmt.Errorf("Failed to get supported content: %v", err)
		}

		if supportedContents.Entities == nil || len(*supportedContents.Entities) == 0 {
			break
		}

		allSupportedContents = append(allSupportedContents, *supportedContents.Entities...)
	}

	for _, content := range allSupportedContents {
		rc.SetCache(p.supportedContentCache, *content.Id, content)
	}

	return &allSupportedContents, resp, nil
}

// getSupportedContentIdByNameFn is an implementation of the function to get a Genesys Cloud supported content by name
func getSupportedContentIdByNameFn(ctx context.Context, p *supportedContentProxy, name string) (id string, retryable bool, response *platformclientv2.APIResponse, err error) {
	supportedContents, resp, err := getAllSupportedContentFn(ctx, p)
	if err != nil {
		return "", false, resp, err
	}

	if supportedContents == nil || len(*supportedContents) == 0 {
		return "", true, resp, fmt.Errorf("No supported content found with name %s", name)
	}

	for _, supportedContent := range *supportedContents {
		if *supportedContent.Name == name {
			log.Printf("Retrieved the supported content id %s by name %s", *supportedContent.Id, name)
			return *supportedContent.Id, false, resp, nil
		}
	}

	return "", true, resp, fmt.Errorf("Unable to find supported content with name %s", name)
}

// getSupportedContentByIdFn is an implementation of the function to get a Genesys Cloud supported content by Id
func getSupportedContentByIdFn(ctx context.Context, p *supportedContentProxy, id string) (supportedContent *platformclientv2.Supportedcontent, response *platformclientv2.APIResponse, err error) {
	content := rc.GetCacheItem(p.supportedContentCache, id)
	if content != nil {
		return content, nil, nil
	}
	return p.conversationsApi.GetConversationsMessagingSupportedcontentSupportedContentId(id)
}

// updateSupportedContentFn is an implementation of the function to update a Genesys Cloud supported content
func updateSupportedContentFn(ctx context.Context, p *supportedContentProxy, id string, supportedContent *platformclientv2.Supportedcontent) (*platformclientv2.Supportedcontent, *platformclientv2.APIResponse, error) {
	return p.conversationsApi.PatchConversationsMessagingSupportedcontentSupportedContentId(id, *supportedContent)
}

// deleteSupportedContentFn is an implementation function for deleting a Genesys Cloud supported content
func deleteSupportedContentFn(ctx context.Context, p *supportedContentProxy, id string) (response *platformclientv2.APIResponse, err error) {
	return p.conversationsApi.DeleteConversationsMessagingSupportedcontentSupportedContentId(id)
}
