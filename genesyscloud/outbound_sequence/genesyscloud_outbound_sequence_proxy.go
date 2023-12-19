package outbound_sequence

import (
	"context"
	"fmt"
	"log"

	"github.com/mypurecloud/platform-client-sdk-go/v119/platformclientv2"
)

/*
The genesyscloud_outbound_sequence_proxy.go file contains the proxy structures and methods that interact
with the Genesys Cloud SDK. We use composition here for each function on the proxy so individual functions can be stubbed
out during testing.
*/

// internalProxy holds a proxy instance that can be used throughout the package
var internalProxy *outboundSequenceProxy

// Type definitions for each func on our proxy so we can easily mock them out later
type createOutboundSequenceFunc func(ctx context.Context, p *outboundSequenceProxy, campaignSequence *platformclientv2.Campaignsequence) (*platformclientv2.Campaignsequence, error)
type getAllOutboundSequenceFunc func(ctx context.Context, p *outboundSequenceProxy) (*[]platformclientv2.Campaignsequence, error)
type getOutboundSequenceIdByNameFunc func(ctx context.Context, p *outboundSequenceProxy, name string) (id string, retryable bool, err error)
type getOutboundSequenceByIdFunc func(ctx context.Context, p *outboundSequenceProxy, id string) (campaignSequence *platformclientv2.Campaignsequence, responseCode int, err error)
type updateOutboundSequenceFunc func(ctx context.Context, p *outboundSequenceProxy, id string, campaignSequence *platformclientv2.Campaignsequence) (*platformclientv2.Campaignsequence, error)
type deleteOutboundSequenceFunc func(ctx context.Context, p *outboundSequenceProxy, id string) (responseCode int, err error)

// outboundSequenceProxy contains all of the methods that call genesys cloud APIs.
type outboundSequenceProxy struct {
	clientConfig                    *platformclientv2.Configuration
	outboundApi                     *platformclientv2.OutboundApi
	createOutboundSequenceAttr      createOutboundSequenceFunc
	getAllOutboundSequenceAttr      getAllOutboundSequenceFunc
	getOutboundSequenceIdByNameAttr getOutboundSequenceIdByNameFunc
	getOutboundSequenceByIdAttr     getOutboundSequenceByIdFunc
	updateOutboundSequenceAttr      updateOutboundSequenceFunc
	deleteOutboundSequenceAttr      deleteOutboundSequenceFunc
}

// newOutboundSequenceProxy initializes the outbound sequence proxy with all of the data needed to communicate with Genesys Cloud
func newOutboundSequenceProxy(clientConfig *platformclientv2.Configuration) *outboundSequenceProxy {
	api := platformclientv2.NewOutboundApiWithConfig(clientConfig)
	return &outboundSequenceProxy{
		clientConfig:                    clientConfig,
		outboundApi:                     api,
		createOutboundSequenceAttr:      createOutboundSequenceFn,
		getAllOutboundSequenceAttr:      getAllOutboundSequenceFn,
		getOutboundSequenceIdByNameAttr: getOutboundSequenceIdByNameFn,
		getOutboundSequenceByIdAttr:     getOutboundSequenceByIdFn,
		updateOutboundSequenceAttr:      updateOutboundSequenceFn,
		deleteOutboundSequenceAttr:      deleteOutboundSequenceFn,
	}
}

// getOutboundSequenceProxy acts as a singleton to for the internalProxy.  It also ensures
// that we can still proxy our tests by directly setting internalProxy package variable
func getOutboundSequenceProxy(clientConfig *platformclientv2.Configuration) *outboundSequenceProxy {
	if internalProxy == nil {
		internalProxy = newOutboundSequenceProxy(clientConfig)
	}

	return internalProxy
}

// createOutboundSequence creates a Genesys Cloud outbound sequence
func (p *outboundSequenceProxy) createOutboundSequence(ctx context.Context, outboundSequence *platformclientv2.Campaignsequence) (*platformclientv2.Campaignsequence, error) {
	return p.createOutboundSequenceAttr(ctx, p, outboundSequence)
}

// getOutboundSequence retrieves all Genesys Cloud outbound sequence
func (p *outboundSequenceProxy) getAllOutboundSequence(ctx context.Context) (*[]platformclientv2.Campaignsequence, error) {
	return p.getAllOutboundSequenceAttr(ctx, p)
}

// getOutboundSequenceIdByName returns a single Genesys Cloud outbound sequence by a name
func (p *outboundSequenceProxy) getOutboundSequenceIdByName(ctx context.Context, name string) (id string, retryable bool, err error) {
	return p.getOutboundSequenceIdByNameAttr(ctx, p, name)
}

// getOutboundSequenceById returns a single Genesys Cloud outbound sequence by Id
func (p *outboundSequenceProxy) getOutboundSequenceById(ctx context.Context, id string) (outboundSequence *platformclientv2.Campaignsequence, statusCode int, err error) {
	return p.getOutboundSequenceByIdAttr(ctx, p, id)
}

// updateOutboundSequence updates a Genesys Cloud outbound sequence
func (p *outboundSequenceProxy) updateOutboundSequence(ctx context.Context, id string, outboundSequence *platformclientv2.Campaignsequence) (*platformclientv2.Campaignsequence, error) {
	return p.updateOutboundSequenceAttr(ctx, p, id, outboundSequence)
}

// deleteOutboundSequence deletes a Genesys Cloud outbound sequence by Id
func (p *outboundSequenceProxy) deleteOutboundSequence(ctx context.Context, id string) (statusCode int, err error) {
	return p.deleteOutboundSequenceAttr(ctx, p, id)
}

// createOutboundSequenceFn is an implementation function for creating a Genesys Cloud outbound sequence
func createOutboundSequenceFn(ctx context.Context, p *outboundSequenceProxy, outboundSequence *platformclientv2.Campaignsequence) (*platformclientv2.Campaignsequence, error) {
	campaignSequence, _, err := p.outboundApi.PostOutboundSequences(*outboundSequence)
	if err != nil {
		return nil, fmt.Errorf("Failed to create outbound sequence: %s", err)
	}

	return campaignSequence, nil
}

// getAllOutboundSequenceFn is the implementation for retrieving all outbound sequence in Genesys Cloud
func getAllOutboundSequenceFn(ctx context.Context, p *outboundSequenceProxy) (*[]platformclientv2.Campaignsequence, error) {
	var allCampaignSequences []platformclientv2.Campaignsequence
	const pageSize = 100

	campaignSequences, _, err := p.outboundApi.GetOutboundSequences(pageSize, 1, true, "", "", "", "")
	if err != nil {
		return nil, fmt.Errorf("Failed to get campaign sequence: %v", err)
	}
	if campaignSequences.Entities == nil || len(*campaignSequences.Entities) == 0 {
		return &allCampaignSequences, nil
	}
	for _, campaignSequence := range *campaignSequences.Entities {
		allCampaignSequences = append(allCampaignSequences, campaignSequence)
	}

	for pageNum := 2; pageNum <= *campaignSequences.PageCount; pageNum++ {
		campaignSequences, _, err := p.outboundApi.GetOutboundSequences(pageSize, pageNum, true, "", "", "", "")
		if err != nil {
			return nil, fmt.Errorf("Failed to get campaign sequence: %v", err)
		}

		if campaignSequences.Entities == nil || len(*campaignSequences.Entities) == 0 {
			break
		}

		for _, campaignSequence := range *campaignSequences.Entities {
			allCampaignSequences = append(allCampaignSequences, campaignSequence)
		}
	}

	return &allCampaignSequences, nil
}

// getOutboundSequenceIdByNameFn is an implementation of the function to get a Genesys Cloud outbound sequence by name
func getOutboundSequenceIdByNameFn(ctx context.Context, p *outboundSequenceProxy, name string) (id string, retryable bool, err error) {
	campaignSequences, _, err := p.outboundApi.GetOutboundSequences(100, 1, true, "", name, "", "")
	if err != nil {
		return "", false, err
	}

	if campaignSequences.Entities == nil || len(*campaignSequences.Entities) == 0 {
		return "", true, fmt.Errorf("No outbound sequence found with name %s", name)
	}

	for _, campaignSequence := range *campaignSequences.Entities {
		if *campaignSequence.Name == name {
			log.Printf("Retrieved the outbound sequence id %s by name %s", *campaignSequence.Id, name)
			return *campaignSequence.Id, false, nil
		}
	}

	return "", true, fmt.Errorf("Unable to find outbound sequence with name %s", name)
}

// getOutboundSequenceByIdFn is an implementation of the function to get a Genesys Cloud outbound sequence by Id
func getOutboundSequenceByIdFn(ctx context.Context, p *outboundSequenceProxy, id string) (outboundSequence *platformclientv2.Campaignsequence, statusCode int, err error) {
	campaignSequence, resp, err := p.outboundApi.GetOutboundSequence(id)
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("Failed to retrieve outbound sequence by id %s: %s", id, err)
	}

	return campaignSequence, resp.StatusCode, nil
}

// updateOutboundSequenceFn is an implementation of the function to update a Genesys Cloud outbound sequence
func updateOutboundSequenceFn(ctx context.Context, p *outboundSequenceProxy, id string, outboundSequence *platformclientv2.Campaignsequence) (*platformclientv2.Campaignsequence, error) {
	sequence, _, err := getOutboundSequenceByIdFn(ctx, p, id)
	if err != nil {
		return nil, fmt.Errorf("Failed to sequence by id %s: %s", id, err)
	}

	outboundSequence.Version = sequence.Version
	campaignSequence, _, err := p.outboundApi.PutOutboundSequence(id, *outboundSequence)
	if err != nil {
		return nil, fmt.Errorf("Failed to update outbound sequence: %s", err)
	}
	return campaignSequence, nil
}

// deleteOutboundSequenceFn is an implementation function for deleting a Genesys Cloud outbound sequence
func deleteOutboundSequenceFn(ctx context.Context, p *outboundSequenceProxy, id string) (statusCode int, err error) {
	resp, err := p.outboundApi.DeleteOutboundSequence(id)
	if err != nil {
		return resp.StatusCode, fmt.Errorf("Failed to delete outbound sequence: %s", err)
	}

	return resp.StatusCode, nil
}
