package outbound_messagingcampaign

import (
	"context"
	"fmt"
	"log"

	"github.com/mypurecloud/platform-client-sdk-go/v157/platformclientv2"

	rc "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_cache"
)

/*
The genesyscloud_outbound_messagingcampaign_proxy.go file contains the proxy structures and methods that interact
with the Genesys Cloud SDK. We use composition here for each function on the proxy so individual functions can be stubbed
out during testing.
*/

// internalProxy holds a proxy instance that can be used throughout the package
var internalProxy *outboundMessagingcampaignProxy

// Type definitions for each func on our proxy so we can easily mock them out later
type createOutboundMessagingcampaignFunc func(ctx context.Context, p *outboundMessagingcampaignProxy, messagingCampaign *platformclientv2.Messagingcampaign) (*platformclientv2.Messagingcampaign, *platformclientv2.APIResponse, error)
type getAllOutboundMessagingcampaignFunc func(ctx context.Context, p *outboundMessagingcampaignProxy) (*[]platformclientv2.Messagingcampaign, *platformclientv2.APIResponse, error)
type getOutboundMessagingcampaignIdByNameFunc func(ctx context.Context, p *outboundMessagingcampaignProxy, name string) (id string, retryable bool, response *platformclientv2.APIResponse, err error)
type getOutboundMessagingcampaignByIdFunc func(ctx context.Context, p *outboundMessagingcampaignProxy, id string) (messagingCampaign *platformclientv2.Messagingcampaign, response *platformclientv2.APIResponse, err error)
type updateOutboundMessagingcampaignFunc func(ctx context.Context, p *outboundMessagingcampaignProxy, id string, messagingCampaign *platformclientv2.Messagingcampaign) (*platformclientv2.Messagingcampaign, *platformclientv2.APIResponse, error)
type deleteOutboundMessagingcampaignFunc func(ctx context.Context, p *outboundMessagingcampaignProxy, id string) (messagingCampaign *platformclientv2.Messagingcampaign, response *platformclientv2.APIResponse, err error)

// outboundMessagingcampaignProxy contains all of the methods that call genesys cloud APIs.
type outboundMessagingcampaignProxy struct {
	clientConfig                             *platformclientv2.Configuration
	outboundApi                              *platformclientv2.OutboundApi
	createOutboundMessagingcampaignAttr      createOutboundMessagingcampaignFunc
	getAllOutboundMessagingcampaignAttr      getAllOutboundMessagingcampaignFunc
	getOutboundMessagingcampaignIdByNameAttr getOutboundMessagingcampaignIdByNameFunc
	getOutboundMessagingcampaignByIdAttr     getOutboundMessagingcampaignByIdFunc
	updateOutboundMessagingcampaignAttr      updateOutboundMessagingcampaignFunc
	deleteOutboundMessagingcampaignAttr      deleteOutboundMessagingcampaignFunc
	obMessagingCampaignCache                 rc.CacheInterface[platformclientv2.Messagingcampaign]
}

// newOutboundMessagingcampaignProxy initializes the outbound messagingcampaign proxy with all of the data needed to communicate with Genesys Cloud
func newOutboundMessagingcampaignProxy(clientConfig *platformclientv2.Configuration) *outboundMessagingcampaignProxy {
	api := platformclientv2.NewOutboundApiWithConfig(clientConfig)
	obMessagingCampaignCache := rc.NewResourceCache[platformclientv2.Messagingcampaign]()
	return &outboundMessagingcampaignProxy{
		clientConfig:                             clientConfig,
		outboundApi:                              api,
		createOutboundMessagingcampaignAttr:      createOutboundMessagingcampaignFn,
		getAllOutboundMessagingcampaignAttr:      getAllOutboundMessagingcampaignFn,
		getOutboundMessagingcampaignIdByNameAttr: getOutboundMessagingcampaignIdByNameFn,
		getOutboundMessagingcampaignByIdAttr:     getOutboundMessagingcampaignByIdFn,
		updateOutboundMessagingcampaignAttr:      updateOutboundMessagingcampaignFn,
		deleteOutboundMessagingcampaignAttr:      deleteOutboundMessagingcampaignFn,
		obMessagingCampaignCache:                 obMessagingCampaignCache,
	}
}

// getOutboundMessagingcampaignProxy acts as a singleton to for the internalProxy.  It also ensures
// that we can still proxy our tests by directly setting internalProxy package variable
func getOutboundMessagingcampaignProxy(clientConfig *platformclientv2.Configuration) *outboundMessagingcampaignProxy {
	if internalProxy == nil {
		internalProxy = newOutboundMessagingcampaignProxy(clientConfig)
	}

	return internalProxy
}

// createOutboundMessagingcampaign creates a Genesys Cloud outbound messagingcampaign
func (p *outboundMessagingcampaignProxy) createOutboundMessagingcampaign(ctx context.Context, outboundMessagingcampaign *platformclientv2.Messagingcampaign) (*platformclientv2.Messagingcampaign, *platformclientv2.APIResponse, error) {
	return p.createOutboundMessagingcampaignAttr(ctx, p, outboundMessagingcampaign)
}

// getOutboundMessagingcampaign retrieves all Genesys Cloud outbound messagingcampaign
func (p *outboundMessagingcampaignProxy) getAllOutboundMessagingcampaign(ctx context.Context) (*[]platformclientv2.Messagingcampaign, *platformclientv2.APIResponse, error) {
	return p.getAllOutboundMessagingcampaignAttr(ctx, p)
}

// getOutboundMessagingcampaignIdByName returns a single Genesys Cloud outbound messagingcampaign by a name
func (p *outboundMessagingcampaignProxy) getOutboundMessagingcampaignIdByName(ctx context.Context, name string) (id string, retryable bool, response *platformclientv2.APIResponse, err error) {
	return p.getOutboundMessagingcampaignIdByNameAttr(ctx, p, name)
}

// getOutboundMessagingcampaignById returns a single Genesys Cloud outbound messagingcampaign by Id
func (p *outboundMessagingcampaignProxy) getOutboundMessagingcampaignById(ctx context.Context, id string) (outboundMessagingcampaign *platformclientv2.Messagingcampaign, response *platformclientv2.APIResponse, err error) {
	return p.getOutboundMessagingcampaignByIdAttr(ctx, p, id)
}

// updateOutboundMessagingcampaign updates a Genesys Cloud outbound messagingcampaign
func (p *outboundMessagingcampaignProxy) updateOutboundMessagingcampaign(ctx context.Context, id string, outboundMessagingcampaign *platformclientv2.Messagingcampaign) (*platformclientv2.Messagingcampaign, *platformclientv2.APIResponse, error) {
	return p.updateOutboundMessagingcampaignAttr(ctx, p, id, outboundMessagingcampaign)
}

// deleteOutboundMessagingcampaign deletes a Genesys Cloud outbound messagingcampaign by Id
func (p *outboundMessagingcampaignProxy) deleteOutboundMessagingcampaign(ctx context.Context, id string) (messagingCampaign *platformclientv2.Messagingcampaign, response *platformclientv2.APIResponse, err error) {
	return p.deleteOutboundMessagingcampaignAttr(ctx, p, id)
}

// createOutboundMessagingcampaignFn is an implementation function for creating a Genesys Cloud outbound messagingcampaign
func createOutboundMessagingcampaignFn(ctx context.Context, p *outboundMessagingcampaignProxy, outboundMessagingcampaign *platformclientv2.Messagingcampaign) (*platformclientv2.Messagingcampaign, *platformclientv2.APIResponse, error) {
	return p.outboundApi.PostOutboundMessagingcampaigns(*outboundMessagingcampaign)
}

// getAllOutboundMessagingcampaignFn is the implementation for retrieving all outbound messagingcampaign in Genesys Cloud
func getAllOutboundMessagingcampaignFn(ctx context.Context, p *outboundMessagingcampaignProxy) (*[]platformclientv2.Messagingcampaign, *platformclientv2.APIResponse, error) {
	var allMessagingCampaigns []platformclientv2.Messagingcampaign
	const pageSize = 100

	messagingCampaigns, resp, err := p.outboundApi.GetOutboundMessagingcampaigns(pageSize, 1, "", "", "", "", []string{}, "", "", []string{}, "", "")
	if err != nil {
		return nil, resp, fmt.Errorf("Failed to get messaging campaign: %v", err)
	}
	if messagingCampaigns.Entities == nil || len(*messagingCampaigns.Entities) == 0 {
		return &allMessagingCampaigns, resp, nil
	}

	allMessagingCampaigns = append(allMessagingCampaigns, *messagingCampaigns.Entities...)

	for pageNum := 2; pageNum <= *messagingCampaigns.PageCount; pageNum++ {
		messagingCampaigns, resp, err := p.outboundApi.GetOutboundMessagingcampaigns(pageSize, pageNum, "", "", "", "", []string{}, "", "", []string{}, "", "")
		if err != nil {
			return nil, resp, fmt.Errorf("Failed to get messaging campaign: %v", err)
		}

		if messagingCampaigns.Entities == nil || len(*messagingCampaigns.Entities) == 0 {
			break
		}

		allMessagingCampaigns = append(allMessagingCampaigns, *messagingCampaigns.Entities...)
	}

	for _, messagingCampaign := range allMessagingCampaigns {
		rc.SetCache(p.obMessagingCampaignCache, *messagingCampaign.Id, messagingCampaign)
	}

	return &allMessagingCampaigns, resp, nil
}

// getOutboundMessagingcampaignIdByNameFn is an implementation of the function to get a Genesys Cloud outbound messagingcampaign by name
func getOutboundMessagingcampaignIdByNameFn(ctx context.Context, p *outboundMessagingcampaignProxy, name string) (string, bool, *platformclientv2.APIResponse, error) {
	messagingCampaigns, resp, err := getAllOutboundMessagingcampaignFn(ctx, p)
	if err != nil {
		return "", false, resp, err
	}

	if messagingCampaigns == nil || len(*messagingCampaigns) == 0 {
		return "", true, resp, fmt.Errorf("No outbound messagingcampaign found with name %s", name)
	}

	for _, messagingCampaign := range *messagingCampaigns {
		if *messagingCampaign.Name == name {
			log.Printf("Retrieved the outbound messagingcampaign id %s by name %s", *messagingCampaign.Id, name)
			return *messagingCampaign.Id, false, resp, nil
		}
	}

	return "", true, resp, fmt.Errorf("Unable to find outbound messagingcampaign with name %s", name)
}

// getOutboundMessagingcampaignByIdFn is an implementation of the function to get a Genesys Cloud outbound messagingcampaign by Id
func getOutboundMessagingcampaignByIdFn(ctx context.Context, p *outboundMessagingcampaignProxy, id string) (outboundMessagingcampaign *platformclientv2.Messagingcampaign, response *platformclientv2.APIResponse, err error) {
	if outboundMessagingcampaign := rc.GetCacheItem(p.obMessagingCampaignCache, id); outboundMessagingcampaign != nil {
		log.Printf("Retrieved outbound messagingcampaign %s by id from cache", id)
		return outboundMessagingcampaign, nil, nil
	}
	return p.outboundApi.GetOutboundMessagingcampaign(id)
}

// updateOutboundMessagingcampaignFn is an implementation of the function to update a Genesys Cloud outbound messagingcampaign
func updateOutboundMessagingcampaignFn(ctx context.Context, p *outboundMessagingcampaignProxy, id string, outboundMessagingcampaign *platformclientv2.Messagingcampaign) (*platformclientv2.Messagingcampaign, *platformclientv2.APIResponse, error) {
	return p.outboundApi.PutOutboundMessagingcampaign(id, *outboundMessagingcampaign)
}

// deleteOutboundMessagingcampaignFn is an implementation function for deleting a Genesys Cloud outbound messagingcampaign
func deleteOutboundMessagingcampaignFn(ctx context.Context, p *outboundMessagingcampaignProxy, id string) (*platformclientv2.Messagingcampaign, *platformclientv2.APIResponse, error) {
	campaign, resp, err := p.outboundApi.DeleteOutboundMessagingcampaign(id)
	if err != nil {
		return nil, resp, err
	}
	// remove from cache
	rc.DeleteCacheItem(p.obMessagingCampaignCache, id)
	return campaign, resp, nil
}
