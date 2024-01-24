package outbound_campaign

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"log"
	gcloud "terraform-provider-genesyscloud/genesyscloud"
	"time"

	"github.com/mypurecloud/platform-client-sdk-go/v119/platformclientv2"
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
type getOutboundCampaignByIdFunc func(ctx context.Context, p *outboundCampaignProxy, id string) (campaign *platformclientv2.Campaign, response *platformclientv2.APIResponse, err error)
type updateOutboundCampaignFunc func(ctx context.Context, p *outboundCampaignProxy, id string, campaign *platformclientv2.Campaign) (*platformclientv2.Campaign, error)
type deleteOutboundCampaignFunc func(ctx context.Context, p *outboundCampaignProxy, id string) (response *platformclientv2.APIResponse, err error)

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
func (p *outboundCampaignProxy) getOutboundCampaignById(ctx context.Context, id string) (outboundCampaign *platformclientv2.Campaign, response *platformclientv2.APIResponse, err error) {
	return p.getOutboundCampaignByIdAttr(ctx, p, id)
}

// updateOutboundCampaign updates a Genesys Cloud outbound campaign
func (p *outboundCampaignProxy) updateOutboundCampaign(ctx context.Context, id string, outboundCampaign *platformclientv2.Campaign) (*platformclientv2.Campaign, error) {
	return p.updateOutboundCampaignAttr(ctx, p, id, outboundCampaign)
}

// deleteOutboundCampaign deletes a Genesys Cloud outbound campaign by Id
func (p *outboundCampaignProxy) deleteOutboundCampaign(ctx context.Context, id string) (response *platformclientv2.APIResponse, err error) {
	return p.deleteOutboundCampaignAttr(ctx, p, id)
}

// turnOffCampaign sets a campaign's campaign_status to 'off' before confirming the update using retry logic and get calls
func (p *outboundCampaignProxy) turnOffCampaign(ctx context.Context, campaignId string) diag.Diagnostics {
	log.Printf("Reading Outbound Campaign %s", campaignId)
	outboundCampaign, _, getErr := p.getOutboundCampaignById(ctx, campaignId)
	if getErr != nil {
		return diag.Errorf("Failed to read Outbound Campaign %s: %s", campaignId, getErr)
	}
	log.Printf("Read Outbound Campaign %s", campaignId)

	log.Printf("Updating campaign '%s' campaign_status to off", *outboundCampaign.Name)
	if diagErr := updateOutboundCampaignStatus(ctx, campaignId, p, *outboundCampaign, "off"); diagErr != nil {
		return diagErr
	}
	log.Printf("Updated campaign '%s'", *outboundCampaign.Name)

	return gcloud.WithRetries(ctx, 30*time.Second, func() *retry.RetryError {
		log.Printf("Reading Outbound Campaign %s to ensure campaign_status is 'off'", campaignId)
		outboundCampaign, _, getErr := p.getOutboundCampaignById(ctx, campaignId)
		if getErr != nil {
			return retry.NonRetryableError(fmt.Errorf("failed to read Outbound Campaign %s: %s", campaignId, getErr))
		}
		log.Printf("Read Outbound Campaign %s", campaignId)
		if *outboundCampaign.CampaignStatus == "on" {
			time.Sleep(5 * time.Second)
			return retry.RetryableError(fmt.Errorf("campaign %s campaign_status is still %s", campaignId, *outboundCampaign.CampaignStatus))
		}
		// Success
		return nil
	})
}

// createOutboundCampaignFn is an implementation function for creating a Genesys Cloud outbound campaign
func createOutboundCampaignFn(ctx context.Context, p *outboundCampaignProxy, outboundCampaign *platformclientv2.Campaign) (*platformclientv2.Campaign, error) {
	campaign, _, err := p.outboundApi.PostOutboundCampaigns(*outboundCampaign)
	if err != nil {
		return nil, fmt.Errorf("Failed to create campaign %s", err)
	}

	return campaign, nil
}

// getAllOutboundCampaignFn is the implementation for retrieving all outbound campaign in Genesys Cloud
func getAllOutboundCampaignFn(ctx context.Context, p *outboundCampaignProxy) (*[]platformclientv2.Campaign, error) {
	var allCampaigns []platformclientv2.Campaign
	const pageSize = 100

	campaigns, _, err := p.outboundApi.GetOutboundCampaigns(pageSize, 1, "", "", nil, "", "", "", "", "", nil, "", "")
	if err != nil {
		return nil, fmt.Errorf("Failed to get campaign: %s", err)
	}

	if campaigns.Entities == nil || len(*campaigns.Entities) == 0 {
		return &allCampaigns, nil
	}

	for _, campaign := range *campaigns.Entities {
		allCampaigns = append(allCampaigns, campaign)
	}

	for pageNum := 2; pageNum <= *campaigns.PageCount; pageNum++ {
		campaigns, _, err := p.outboundApi.GetOutboundCampaigns(pageSize, pageNum, "", "", nil, "", "", "", "", "", nil, "", "")
		if err != nil {
			return nil, fmt.Errorf("Failed to get campaign: %s", err)
		}

		if campaigns.Entities == nil || len(*campaigns.Entities) == 0 {
			break
		}

		for _, campaign := range *campaigns.Entities {
			allCampaigns = append(allCampaigns, campaign)
		}
	}

	return &allCampaigns, nil
}

// getOutboundCampaignIdByNameFn is an implementation of the function to get a Genesys Cloud outbound campaign by name
func getOutboundCampaignIdByNameFn(ctx context.Context, p *outboundCampaignProxy, name string) (id string, retryable bool, err error) {
	campaigns, err := getAllOutboundCampaignFn(ctx, p)
	if err != nil {
		return "", false, err
	}
	if campaigns == nil || len(*campaigns) == 0 {
		return "", true, fmt.Errorf("No campaigns found with name %s", name)
	}

	for _, campaign := range *campaigns {
		if *campaign.Name == name {
			log.Printf("Retrieved the campaign id %s by name %s", *campaign.Id, name)
			return *campaign.Id, false, nil
		}
	}

	return "", true, fmt.Errorf("Unable to find campaign with name %s", name)
}

// getOutboundCampaignByIdFn is an implementation of the function to get a Genesys Cloud outbound campaign by Id
func getOutboundCampaignByIdFn(ctx context.Context, p *outboundCampaignProxy, id string) (outboundCampaign *platformclientv2.Campaign, response *platformclientv2.APIResponse, err error) {
	campaign, resp, err := p.outboundApi.GetOutboundCampaign(id)
	if err != nil {
		return nil, resp, fmt.Errorf("Failed to retrieve campaign by id %s: %s", id, err)
	}
	return campaign, resp, nil
}

// updateOutboundCampaignFn is an implementation of the function to update a Genesys Cloud outbound campaign
func updateOutboundCampaignFn(ctx context.Context, p *outboundCampaignProxy, id string, outboundCampaign *platformclientv2.Campaign) (*platformclientv2.Campaign, error) {
	campaign, _, err := getOutboundCampaignByIdFn(ctx, p, id)
	if err != nil {
		return nil, fmt.Errorf("Failed to campaign by id %s: %s", id, err)
	}

	outboundCampaign.Version = campaign.Version
	outboundCampaign, _, err = p.outboundApi.PutOutboundCampaign(id, *outboundCampaign)
	if err != nil {
		return nil, fmt.Errorf("Failed to update campaign: %s", err)
	}
	return outboundCampaign, nil
}

// deleteOutboundCampaignFn is an implementation function for deleting a Genesys Cloud outbound campaign
func deleteOutboundCampaignFn(ctx context.Context, p *outboundCampaignProxy, id string) (response *platformclientv2.APIResponse, err error) {
	_, resp, err := p.outboundApi.DeleteOutboundCampaign(id)
	if err != nil {
		return resp, fmt.Errorf("Failed to delete campaign: %s", err)
	}

	return resp, nil
}
