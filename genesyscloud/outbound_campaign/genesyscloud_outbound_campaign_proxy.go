package outbound_campaign

import (
	"context"
	"github.com/mypurecloud/platform-client-sdk-go/v115/platformclientv2"
)

/*
The genesyscloud_outbound_campaign_proxy.go file contains the proxy structures and methods that interact
with the Genesys Cloud SDK. We use composition here for each function on the proxy so individual functions can be stubbed
out during testing.
*/

// internalProxy holds a proxy instance that can be used throughout the package
var internalProxy *outboundCampaignProxy

// Type definitions for each func on our proxy so we can easily mock them out later
type createOutboundCampaignFunc func(ctx context.Context, p *outboundCampaignProxy, campaign *platformclientv2.Campaign) (*platformclientv2.Campaign, error)
type getAllOutboundCampaignFunc func(ctx context.Context, p *outboundCampaignProxy) (*[]platformclientv2.Campaign, error)
type getOutboundCampaignIdByNameFunc func(ctx context.Context, p *outboundCampaignProxy, name string) (id string, retryable bool, err error)
type getOutboundCampaignByIdFunc func(ctx context.Context, p *outboundCampaignProxy, id string) (campaign *platformclientv2.Campaign, responseCode int, err error)
type updateOutboundCampaignFunc func(ctx context.Context, p *outboundCampaignProxy, id string, campaign *platformclientv2.Campaign) (*platformclientv2.Campaign, error)
type deleteOutboundCampaignFunc func(ctx context.Context, p *outboundCampaignProxy, id string) (responseCode int, err error)

// outboundCampaignProxy contains all of the methods that call genesys cloud APIs.
type outboundCampaignProxy struct {
	clientConfig                    *platformclientv2.Configuration
	outboundApi                     *platformclientv2.OutboundApi
	createOutboundCampaignAttr      createOutboundCampaignFunc
	getAllOutboundCampaignAttr      getAllOutboundCampaignFunc
	getOutboundCampaignIdByNameAttr getOutboundCampaignIdByNameFunc
	getOutboundCampaignByIdAttr     getOutboundCampaignByIdFunc
	updateOutboundCampaignAttr      updateOutboundCampaignFunc
	deleteOutboundCampaignAttr      deleteOutboundCampaignFunc
}

// newOutboundCampaignProxy initializes the outbound campaign proxy with all of the data needed to communicate with Genesys Cloud
func newOutboundCampaignProxy(clientConfig *platformclientv2.Configuration) *outboundCampaignProxy {
	api := platformclientv2.NewOutboundApiWithConfig(clientConfig)
	return &outboundCampaignProxy{
		clientConfig:                    clientConfig,
		outboundApi:                     api,
		createOutboundCampaignAttr:      createOutboundCampaignFn,
		getAllOutboundCampaignAttr:      getAllOutboundCampaignFn,
		getOutboundCampaignIdByNameAttr: getOutboundCampaignIdByNameFn,
		getOutboundCampaignByIdAttr:     getOutboundCampaignByIdFn,
		updateOutboundCampaignAttr:      updateOutboundCampaignFn,
		deleteOutboundCampaignAttr:      deleteOutboundCampaignFn,
	}
}

// getOutboundCampaignProxy acts as a singleton to for the internalProxy.  It also ensures
// that we can still proxy our tests by directly setting internalProxy package variable
func getOutboundCampaignProxy(clientConfig *platformclientv2.Configuration) *outboundCampaignProxy {
	if internalProxy == nil {
		internalProxy = newOutboundCampaignProxy(clientConfig)
	}

	return internalProxy
}

// createOutboundCampaign creates a Genesys Cloud outbound campaign
func (p *outboundCampaignProxy) createOutboundCampaign(ctx context.Context, outboundCampaign *platformclientv2.Campaign) (*platformclientv2.Campaign, error) {
	return p.createOutboundCampaignAttr(ctx, p, outboundCampaign)
}

// getOutboundCampaign retrieves all Genesys Cloud outbound campaign
func (p *outboundCampaignProxy) getAllOutboundCampaign(ctx context.Context) (*[]platformclientv2.Campaign, error) {
	return p.getAllOutboundCampaignAttr(ctx, p)
}

// getOutboundCampaignIdByName returns a single Genesys Cloud outbound campaign by a name
func (p *outboundCampaignProxy) getOutboundCampaignIdByName(ctx context.Context, name string) (id string, retryable bool, err error) {
	return p.getOutboundCampaignIdByNameAttr(ctx, p, name)
}

// getOutboundCampaignById returns a single Genesys Cloud outbound campaign by Id
func (p *outboundCampaignProxy) getOutboundCampaignById(ctx context.Context, id string) (outboundCampaign *platformclientv2.Campaign, statusCode int, err error) {
	return p.getOutboundCampaignByIdAttr(ctx, p, id)
}

// updateOutboundCampaign updates a Genesys Cloud outbound campaign
func (p *outboundCampaignProxy) updateOutboundCampaign(ctx context.Context, id string, outboundCampaign *platformclientv2.Campaign) (*platformclientv2.Campaign, error) {
	return p.updateOutboundCampaignAttr(ctx, p, id, outboundCampaign)
}

// deleteOutboundCampaign deletes a Genesys Cloud outbound campaign by Id
func (p *outboundCampaignProxy) deleteOutboundCampaign(ctx context.Context, id string) (statusCode int, err error) {
	return p.deleteOutboundCampaignAttr(ctx, p, id)
}

// createOutboundCampaignFn is an implementation function for creating a Genesys Cloud outbound campaign
func createOutboundCampaignFn(ctx context.Context, p *outboundCampaignProxy, outboundCampaign *platformclientv2.Campaign) (*platformclientv2.Campaign, error) {
	return nil, nil
}

// getAllOutboundCampaignFn is the implementation for retrieving all outbound campaign in Genesys Cloud
func getAllOutboundCampaignFn(ctx context.Context, p *outboundCampaignProxy) (*[]platformclientv2.Campaign, error) {
	return nil, nil
}

// getOutboundCampaignIdByNameFn is an implementation of the function to get a Genesys Cloud outbound campaign by name
func getOutboundCampaignIdByNameFn(ctx context.Context, p *outboundCampaignProxy, name string) (id string, retryable bool, err error) {
	return "", false, nil
}

// getOutboundCampaignByIdFn is an implementation of the function to get a Genesys Cloud outbound campaign by Id
func getOutboundCampaignByIdFn(ctx context.Context, p *outboundCampaignProxy, id string) (outboundCampaign *platformclientv2.Campaign, statusCode int, err error) {
	return nil, 0, nil
}

// updateOutboundCampaignFn is an implementation of the function to update a Genesys Cloud outbound campaign
func updateOutboundCampaignFn(ctx context.Context, p *outboundCampaignProxy, id string, outboundCampaign *platformclientv2.Campaign) (*platformclientv2.Campaign, error) {
	return nil, nil
}

// deleteOutboundCampaignFn is an implementation function for deleting a Genesys Cloud outbound campaign
func deleteOutboundCampaignFn(ctx context.Context, p *outboundCampaignProxy, id string) (statusCode int, err error) {
	return 0, nil
}