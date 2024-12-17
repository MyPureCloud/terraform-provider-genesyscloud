package outbound_digitalruleset

import (
	"context"
	"fmt"
	"log"

	"github.com/mypurecloud/platform-client-sdk-go/v149/platformclientv2"
)

/*
The genesyscloud_outbound_digitalruleset_proxy.go file contains the proxy structures and methods that interact
with the Genesys Cloud SDK. We use composition here for each function on the proxy so individual functions can be stubbed
out during testing.
*/

// internalProxy holds a proxy instance that can be used throughout the package
var internalProxy *outboundDigitalrulesetProxy

// Type definitions for each func on our proxy so we can easily mock them out later
type createOutboundDigitalrulesetFunc func(ctx context.Context, p *outboundDigitalrulesetProxy, digitalRuleSet *platformclientv2.Digitalruleset) (*platformclientv2.Digitalruleset, *platformclientv2.APIResponse, error)
type getAllOutboundDigitalrulesetFunc func(ctx context.Context, p *outboundDigitalrulesetProxy) (*[]platformclientv2.Digitalruleset, *platformclientv2.APIResponse, error)
type getOutboundDigitalrulesetIdByNameFunc func(ctx context.Context, p *outboundDigitalrulesetProxy, name string) (id string, response *platformclientv2.APIResponse, retryable bool, err error)
type getOutboundDigitalrulesetByIdFunc func(ctx context.Context, p *outboundDigitalrulesetProxy, id string) (digitalRuleSet *platformclientv2.Digitalruleset, response *platformclientv2.APIResponse, err error)
type updateOutboundDigitalrulesetFunc func(ctx context.Context, p *outboundDigitalrulesetProxy, id string, digitalRuleSet *platformclientv2.Digitalruleset) (*platformclientv2.Digitalruleset, *platformclientv2.APIResponse, error)
type deleteOutboundDigitalrulesetFunc func(ctx context.Context, p *outboundDigitalrulesetProxy, id string) (response *platformclientv2.APIResponse, err error)

// outboundDigitalrulesetProxy contains all of the methods that call genesys cloud APIs.
type outboundDigitalrulesetProxy struct {
	clientConfig                          *platformclientv2.Configuration
	outboundApi                           *platformclientv2.OutboundApi
	createOutboundDigitalrulesetAttr      createOutboundDigitalrulesetFunc
	getAllOutboundDigitalrulesetAttr      getAllOutboundDigitalrulesetFunc
	getOutboundDigitalrulesetIdByNameAttr getOutboundDigitalrulesetIdByNameFunc
	getOutboundDigitalrulesetByIdAttr     getOutboundDigitalrulesetByIdFunc
	updateOutboundDigitalrulesetAttr      updateOutboundDigitalrulesetFunc
	deleteOutboundDigitalrulesetAttr      deleteOutboundDigitalrulesetFunc
}

// newOutboundDigitalrulesetProxy initializes the outbound digitalruleset proxy with all of the data needed to communicate with Genesys Cloud
func newOutboundDigitalrulesetProxy(clientConfig *platformclientv2.Configuration) *outboundDigitalrulesetProxy {
	api := platformclientv2.NewOutboundApiWithConfig(clientConfig)
	return &outboundDigitalrulesetProxy{
		clientConfig:                          clientConfig,
		outboundApi:                           api,
		createOutboundDigitalrulesetAttr:      createOutboundDigitalrulesetFn,
		getAllOutboundDigitalrulesetAttr:      getAllOutboundDigitalrulesetFn,
		getOutboundDigitalrulesetIdByNameAttr: getOutboundDigitalrulesetIdByNameFn,
		getOutboundDigitalrulesetByIdAttr:     getOutboundDigitalrulesetByIdFn,
		updateOutboundDigitalrulesetAttr:      updateOutboundDigitalrulesetFn,
		deleteOutboundDigitalrulesetAttr:      deleteOutboundDigitalrulesetFn,
	}
}

// getOutboundDigitalrulesetProxy acts as a singleton to for the internalProxy.  It also ensures
// that we can still proxy our tests by directly setting internalProxy package variable
func getOutboundDigitalrulesetProxy(clientConfig *platformclientv2.Configuration) *outboundDigitalrulesetProxy {
	if internalProxy == nil {
		internalProxy = newOutboundDigitalrulesetProxy(clientConfig)
	}

	return internalProxy
}

// createOutboundDigitalruleset creates a Genesys Cloud outbound digitalruleset
func (p *outboundDigitalrulesetProxy) createOutboundDigitalruleset(ctx context.Context, outboundDigitalruleset *platformclientv2.Digitalruleset) (*platformclientv2.Digitalruleset, *platformclientv2.APIResponse, error) {
	return p.createOutboundDigitalrulesetAttr(ctx, p, outboundDigitalruleset)
}

// getOutboundDigitalruleset retrieves all Genesys Cloud outbound digitalruleset
func (p *outboundDigitalrulesetProxy) getAllOutboundDigitalruleset(ctx context.Context) (*[]platformclientv2.Digitalruleset, *platformclientv2.APIResponse, error) {
	return p.getAllOutboundDigitalrulesetAttr(ctx, p)
}

// getOutboundDigitalrulesetIdByName returns a single Genesys Cloud outbound digitalruleset by a name
func (p *outboundDigitalrulesetProxy) getOutboundDigitalrulesetIdByName(ctx context.Context, name string) (id string, response *platformclientv2.APIResponse, retryable bool, err error) {
	return p.getOutboundDigitalrulesetIdByNameAttr(ctx, p, name)
}

// getOutboundDigitalrulesetById returns a single Genesys Cloud outbound digitalruleset by Id
func (p *outboundDigitalrulesetProxy) getOutboundDigitalrulesetById(ctx context.Context, id string) (outboundDigitalruleset *platformclientv2.Digitalruleset, response *platformclientv2.APIResponse, err error) {
	return p.getOutboundDigitalrulesetByIdAttr(ctx, p, id)
}

// updateOutboundDigitalruleset updates a Genesys Cloud outbound digitalruleset
func (p *outboundDigitalrulesetProxy) updateOutboundDigitalruleset(ctx context.Context, id string, outboundDigitalruleset *platformclientv2.Digitalruleset) (*platformclientv2.Digitalruleset, *platformclientv2.APIResponse, error) {
	return p.updateOutboundDigitalrulesetAttr(ctx, p, id, outboundDigitalruleset)
}

// deleteOutboundDigitalruleset deletes a Genesys Cloud outbound digitalruleset by Id
func (p *outboundDigitalrulesetProxy) deleteOutboundDigitalruleset(ctx context.Context, id string) (status *platformclientv2.APIResponse, err error) {
	return p.deleteOutboundDigitalrulesetAttr(ctx, p, id)
}

// createOutboundDigitalrulesetFn is an implementation function for creating a Genesys Cloud outbound digitalruleset
func createOutboundDigitalrulesetFn(ctx context.Context, p *outboundDigitalrulesetProxy, outboundDigitalruleset *platformclientv2.Digitalruleset) (*platformclientv2.Digitalruleset, *platformclientv2.APIResponse, error) {
	return p.outboundApi.PostOutboundDigitalrulesets(*outboundDigitalruleset)
}

// getAllOutboundDigitalrulesetFn is the implementation for retrieving all outbound digitalruleset in Genesys Cloud
func getAllOutboundDigitalrulesetFn(ctx context.Context, p *outboundDigitalrulesetProxy) (*[]platformclientv2.Digitalruleset, *platformclientv2.APIResponse, error) {
	var allDigitalRuleSets []platformclientv2.Digitalruleset
	const pageSize = 100

	digitalRuleSets, resp, err := p.outboundApi.GetOutboundDigitalrulesets(pageSize, 1, "", "", "", []string{})
	if err != nil {
		return nil, resp, fmt.Errorf("Failed to get digital rule set: %v", err)
	}
	if digitalRuleSets.Entities == nil || len(*digitalRuleSets.Entities) == 0 {
		return &allDigitalRuleSets, resp, nil
	}

	allDigitalRuleSets = append(allDigitalRuleSets, *digitalRuleSets.Entities...)

	for pageNum := 2; pageNum <= *digitalRuleSets.PageCount; pageNum++ {
		digitalRuleSets, resp, err := p.outboundApi.GetOutboundDigitalrulesets(pageSize, pageNum, "", "", "", []string{})
		if err != nil {
			return nil, resp, fmt.Errorf("Failed to get digital rule set: %v", err)
		}

		if digitalRuleSets.Entities == nil || len(*digitalRuleSets.Entities) == 0 {
			break
		}

		allDigitalRuleSets = append(allDigitalRuleSets, *digitalRuleSets.Entities...)
	}

	return &allDigitalRuleSets, resp, nil
}

// getOutboundDigitalrulesetIdByNameFn is an implementation of the function to get a Genesys Cloud outbound digitalruleset by name
func getOutboundDigitalrulesetIdByNameFn(ctx context.Context, p *outboundDigitalrulesetProxy, name string) (id string, response *platformclientv2.APIResponse, retryable bool, err error) {
	const pageSize = 100
	digitalRuleSets, resp, err := p.outboundApi.GetOutboundDigitalrulesets(pageSize, 1, "", "", "", []string{})
	if err != nil {
		return "", resp, false, err
	}

	if digitalRuleSets.Entities == nil || len(*digitalRuleSets.Entities) == 0 {
		return "", resp, true, fmt.Errorf("No outbound digitalruleset found with name %s", name)
	}

	for _, digitalRuleSet := range *digitalRuleSets.Entities {
		if *digitalRuleSet.Name == name {
			log.Printf("Retrieved the outbound digitalruleset id %s by name %s", *digitalRuleSet.Id, name)
			return *digitalRuleSet.Id, resp, false, nil
		}
	}

	return "", resp, true, fmt.Errorf("Unable to find outbound digitalruleset with name %s", name)
}

// getOutboundDigitalrulesetByIdFn is an implementation of the function to get a Genesys Cloud outbound digitalruleset by Id
func getOutboundDigitalrulesetByIdFn(ctx context.Context, p *outboundDigitalrulesetProxy, id string) (outboundDigitalruleset *platformclientv2.Digitalruleset, response *platformclientv2.APIResponse, err error) {
	return p.outboundApi.GetOutboundDigitalruleset(id)
}

// updateOutboundDigitalrulesetFn is an implementation of the function to update a Genesys Cloud outbound digitalruleset
func updateOutboundDigitalrulesetFn(ctx context.Context, p *outboundDigitalrulesetProxy, id string, outboundDigitalruleset *platformclientv2.Digitalruleset) (*platformclientv2.Digitalruleset, *platformclientv2.APIResponse, error) {
	digitalRuleSet, resp, err := getOutboundDigitalrulesetByIdFn(ctx, p, id)
	if err != nil {
		return nil, resp, fmt.Errorf("Failed to fetch ruleset by id %s: %s", id, err)
	}

	outboundDigitalruleset.Version = digitalRuleSet.Version
	outboundDigitalruleset, resp, err = p.outboundApi.PutOutboundDigitalruleset(id, *outboundDigitalruleset)
	if err != nil {
		return nil, resp, fmt.Errorf("Failed to update ruleset %s", err)
	}
	return outboundDigitalruleset, resp, nil
}

// deleteOutboundDigitalrulesetFn is an implementation function for deleting a Genesys Cloud outbound digitalruleset
func deleteOutboundDigitalrulesetFn(ctx context.Context, p *outboundDigitalrulesetProxy, id string) (response *platformclientv2.APIResponse, err error) {
	return p.outboundApi.DeleteOutboundDigitalruleset(id)
}
