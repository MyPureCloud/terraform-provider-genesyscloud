package outbound_campaignrule

import (
	"context"
	"fmt"
	"log"

	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

/*
The genesyscloud_outbound_campaignrule_proxy.go file contains the proxy structures and methods that interact
with the Genesys Cloud SDK. We use composition here for each function on the proxy so individual functions can be stubbed
out during testing.
*/

// internalProxy holds a proxy instance that can be used throughout the package
var internalProxy *outboundCampaignruleProxy

// Type definitions for each func on our proxy so we can easily mock them out later
type createOutboundCampaignruleFunc func(ctx context.Context, p *outboundCampaignruleProxy, campaignRule *platformclientv2.Campaignrule) (*platformclientv2.Campaignrule, *platformclientv2.APIResponse, error)
type getAllOutboundCampaignruleFunc func(ctx context.Context, p *outboundCampaignruleProxy) (*[]platformclientv2.Campaignrule, *platformclientv2.APIResponse, error)
type getOutboundCampaignruleIdByNameFunc func(ctx context.Context, p *outboundCampaignruleProxy, name string) (id string, retryable bool, response *platformclientv2.APIResponse, err error)
type getOutboundCampaignruleByIdFunc func(ctx context.Context, p *outboundCampaignruleProxy, id string) (campaignRule *platformclientv2.Campaignrule, response *platformclientv2.APIResponse, err error)
type updateOutboundCampaignruleFunc func(ctx context.Context, p *outboundCampaignruleProxy, id string, campaignRule *platformclientv2.Campaignrule) (*platformclientv2.Campaignrule, *platformclientv2.APIResponse, error)
type deleteOutboundCampaignruleFunc func(ctx context.Context, p *outboundCampaignruleProxy, id string) (response *platformclientv2.APIResponse, err error)

// outboundCampaignruleProxy contains all of the methods that call genesys cloud APIs.
type outboundCampaignruleProxy struct {
	clientConfig                        *platformclientv2.Configuration
	outboundApi                         *platformclientv2.OutboundApi
	createOutboundCampaignruleAttr      createOutboundCampaignruleFunc
	getAllOutboundCampaignruleAttr      getAllOutboundCampaignruleFunc
	getOutboundCampaignruleIdByNameAttr getOutboundCampaignruleIdByNameFunc
	getOutboundCampaignruleByIdAttr     getOutboundCampaignruleByIdFunc
	updateOutboundCampaignruleAttr      updateOutboundCampaignruleFunc
	deleteOutboundCampaignruleAttr      deleteOutboundCampaignruleFunc
}

// newOutboundCampaignruleProxy initializes the outbound campaignrule proxy with all of the data needed to communicate with Genesys Cloud
func newOutboundCampaignruleProxy(clientConfig *platformclientv2.Configuration) *outboundCampaignruleProxy {
	api := platformclientv2.NewOutboundApiWithConfig(clientConfig)
	return &outboundCampaignruleProxy{
		clientConfig:                        clientConfig,
		outboundApi:                         api,
		createOutboundCampaignruleAttr:      createOutboundCampaignruleFn,
		getAllOutboundCampaignruleAttr:      getAllOutboundCampaignruleFn,
		getOutboundCampaignruleIdByNameAttr: getOutboundCampaignruleIdByNameFn,
		getOutboundCampaignruleByIdAttr:     getOutboundCampaignruleByIdFn,
		updateOutboundCampaignruleAttr:      updateOutboundCampaignruleFn,
		deleteOutboundCampaignruleAttr:      deleteOutboundCampaignruleFn,
	}
}

// getOutboundCampaignruleProxy acts as a singleton to for the internalProxy.  It also ensures
// that we can still proxy our tests by directly setting internalProxy package variable
func getOutboundCampaignruleProxy(clientConfig *platformclientv2.Configuration) *outboundCampaignruleProxy {
	if internalProxy == nil {
		internalProxy = newOutboundCampaignruleProxy(clientConfig)
	}
	return internalProxy
}

// createOutboundCampaignrule creates a Genesys Cloud outbound campaignrule
func (p *outboundCampaignruleProxy) createOutboundCampaignrule(ctx context.Context, outboundCampaignrule *platformclientv2.Campaignrule) (*platformclientv2.Campaignrule, *platformclientv2.APIResponse, error) {
	return p.createOutboundCampaignruleAttr(ctx, p, outboundCampaignrule)
}

// getOutboundCampaignrule retrieves all Genesys Cloud outbound campaignrule
func (p *outboundCampaignruleProxy) getAllOutboundCampaignrule(ctx context.Context) (*[]platformclientv2.Campaignrule, *platformclientv2.APIResponse, error) {
	return p.getAllOutboundCampaignruleAttr(ctx, p)
}

// getOutboundCampaignruleIdByName returns a single Genesys Cloud outbound campaignrule by a name
func (p *outboundCampaignruleProxy) getOutboundCampaignruleIdByName(ctx context.Context, name string) (id string, retryable bool, response *platformclientv2.APIResponse, err error) {
	return p.getOutboundCampaignruleIdByNameAttr(ctx, p, name)
}

// getOutboundCampaignruleById returns a single Genesys Cloud outbound campaignrule by Id
func (p *outboundCampaignruleProxy) getOutboundCampaignruleById(ctx context.Context, id string) (outboundCampaignrule *platformclientv2.Campaignrule, response *platformclientv2.APIResponse, err error) {
	return p.getOutboundCampaignruleByIdAttr(ctx, p, id)
}

// updateOutboundCampaignrule updates a Genesys Cloud outbound campaignrule
func (p *outboundCampaignruleProxy) updateOutboundCampaignrule(ctx context.Context, id string, outboundCampaignrule *platformclientv2.Campaignrule) (*platformclientv2.Campaignrule, *platformclientv2.APIResponse, error) {
	return p.updateOutboundCampaignruleAttr(ctx, p, id, outboundCampaignrule)
}

// deleteOutboundCampaignrule deletes a Genesys Cloud outbound campaignrule by Id
func (p *outboundCampaignruleProxy) deleteOutboundCampaignrule(ctx context.Context, id string) (response *platformclientv2.APIResponse, err error) {
	return p.deleteOutboundCampaignruleAttr(ctx, p, id)
}

// createOutboundCampaignruleFn is an implementation function for creating a Genesys Cloud outbound campaignrule
func createOutboundCampaignruleFn(ctx context.Context, p *outboundCampaignruleProxy, outboundCampaignrule *platformclientv2.Campaignrule) (*platformclientv2.Campaignrule, *platformclientv2.APIResponse, error) {
	rule, resp, err := p.outboundApi.PostOutboundCampaignrules(*outboundCampaignrule)
	if err != nil {
		return nil, resp, fmt.Errorf("Failed to create campaign rule %s", err)
	}
	return rule, resp, nil
}

// getAllOutboundCampaignruleFn is the implementation for retrieving all outbound campaignrule in Genesys Cloud
func getAllOutboundCampaignruleFn(ctx context.Context, p *outboundCampaignruleProxy) (*[]platformclientv2.Campaignrule, *platformclientv2.APIResponse, error) {
	var allCampaignRules []platformclientv2.Campaignrule
	const pageSize = 100

	campaignRules, resp, err := p.outboundApi.GetOutboundCampaignrules(pageSize, 1, true, "", "", "", "")
	if err != nil {
		return nil, resp, fmt.Errorf("Failed to get campaign rules: %v", err)
	}

	if campaignRules.Entities == nil || len(*campaignRules.Entities) == 0 {
		return &allCampaignRules, resp, nil
	}

	for _, campaignRule := range *campaignRules.Entities {
		allCampaignRules = append(allCampaignRules, campaignRule)
	}

	for pageNum := 2; pageNum <= *campaignRules.PageCount; pageNum++ {
		campaignRules, resp, err := p.outboundApi.GetOutboundCampaignrules(pageSize, pageNum, true, "", "", "", "")
		if err != nil {
			return nil, resp, fmt.Errorf("Failed to get campaign rules: %v", err)
		}

		if campaignRules.Entities == nil || len(*campaignRules.Entities) == 0 {
			break
		}

		for _, campaignRule := range *campaignRules.Entities {
			allCampaignRules = append(allCampaignRules, campaignRule)
		}
	}
	return &allCampaignRules, resp, nil
}

// getOutboundCampaignruleIdByNameFn is an implementation of the function to get a Genesys Cloud outbound campaignrule by name
func getOutboundCampaignruleIdByNameFn(ctx context.Context, p *outboundCampaignruleProxy, name string) (id string, retryable bool, response *platformclientv2.APIResponse, err error) {
	campaignRules, resp, err := p.outboundApi.GetOutboundCampaignrules(100, 1, true, "", name, "", "")
	if err != nil {
		return "", false, resp, err
	}

	if campaignRules.Entities == nil || len(*campaignRules.Entities) == 0 {
		return "", true, resp, fmt.Errorf("No outbound campaignrule with name %s", name)
	}

	for _, campaignRule := range *campaignRules.Entities {
		if *campaignRule.Name == name {
			log.Printf("Retrieved the outbound capaign rule id %s by name %s", *campaignRule.Id, name)
			return *campaignRule.Id, false, resp, nil
		}
	}
	return "", true, resp, fmt.Errorf("Unable to find outbound campaign rule with name %s", name)
}

// getOutboundCampaignruleByIdFn is an implementation of the function to get a Genesys Cloud outbound campaignrule by Id
func getOutboundCampaignruleByIdFn(ctx context.Context, p *outboundCampaignruleProxy, id string) (outboundCampaignrule *platformclientv2.Campaignrule, response *platformclientv2.APIResponse, err error) {
	rule, resp, err := p.outboundApi.GetOutboundCampaignrule(id)
	if err != nil {
		return nil, resp, fmt.Errorf("Failed to retrieve campaign rule by id %s: %s", id, err)
	}
	return rule, resp, nil
}

// updateOutboundCampaignruleFn is an implementation of the function to update a Genesys Cloud outbound campaignrule
func updateOutboundCampaignruleFn(ctx context.Context, p *outboundCampaignruleProxy, id string, outboundCampaignrule *platformclientv2.Campaignrule) (*platformclientv2.Campaignrule, *platformclientv2.APIResponse, error) {
	rule, resp, err := getOutboundCampaignruleByIdFn(ctx, p, id)
	if err != nil {
		return nil, resp, fmt.Errorf("Failed to campaign rule by id %s: %s", id, err)
	}

	outboundCampaignrule.Version = rule.Version
	campaignRule, resp, err := p.outboundApi.PutOutboundCampaignrule(id, *outboundCampaignrule)
	if err != nil {
		return nil, resp, fmt.Errorf("Failed to update campaign rule: %s", err)
	}
	return campaignRule, resp, nil
}

// deleteOutboundCampaignruleFn is an implementation function for deleting a Genesys Cloud outbound campaignrule
func deleteOutboundCampaignruleFn(ctx context.Context, p *outboundCampaignruleProxy, id string) (response *platformclientv2.APIResponse, err error) {
	resp, err := p.outboundApi.DeleteOutboundCampaignrule(id)
	if err != nil {
		return resp, fmt.Errorf("Failed to delete campaign rule: %s", err)
	}
	return resp, nil
}
