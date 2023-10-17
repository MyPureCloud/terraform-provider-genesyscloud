package outbound_ruleset

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/mypurecloud/platform-client-sdk-go/v115/platformclientv2"
)

/*
The genesyscloud_outbound_ruleset_proxy.go file contains the proxy structures and methods that interact
with the Genesys Cloud SDK. We use composition here for each function on the proxy so individual functions can be stubbed
out during testing.
*/

// internalProxy holds a proxy instance that can be used throughout the package
var internalProxy *outboundRulesetProxy

// Type definitions for each func on our proxy so we can easily mock them out later
type createOutboundRulesetFunc func(ctx context.Context, p *outboundRulesetProxy, ruleset *platformclientv2.Ruleset) (*platformclientv2.Ruleset, error)
type getAllOutboundRulesetFunc func(ctx context.Context, p *outboundRulesetProxy) (*[]platformclientv2.Ruleset, error)
type getOutboundRulesetByIdFunc func(ctx context.Context, p *outboundRulesetProxy, rulesetId string) (ruleset *platformclientv2.Ruleset, responseCode int, err error)
type getOutboundRulesetIdByNameFunc func(ctx context.Context, p *outboundRulesetProxy, search string) (rulesetId string, retryable bool, err error)
type updateOutboundRulesetFunc func(ctx context.Context, p *outboundRulesetProxy, rulesetId string, ruleset *platformclientv2.Ruleset) (*platformclientv2.Ruleset, error)
type deleteOutboundRulesetFunc func(ctx context.Context, p *outboundRulesetProxy, rulesetId string) (responseCode int, err error)

// outboundRulesetProxy contains all of the methods that call genesys cloud APIs.
type outboundRulesetProxy struct {
	clientConfig                   *platformclientv2.Configuration
	outboundApi                    *platformclientv2.OutboundApi
	createOutboundRulesetAttr      createOutboundRulesetFunc
	getAllOutboundRulesetAttr      getAllOutboundRulesetFunc
	getOutboundRulesetByIdAttr     getOutboundRulesetByIdFunc
	getOutboundRulesetIdByNameAttr getOutboundRulesetIdByNameFunc
	updateOutboundRulesetAttr      updateOutboundRulesetFunc
	deleteOutboundRulesetAttr      deleteOutboundRulesetFunc
}

// newOutboundRulesetProxy initializes the ruleset proxy with all of the data needed to communicate with Genesys Cloud
func newOutboundRulesetProxy(clientConfig *platformclientv2.Configuration) *outboundRulesetProxy {
	api := platformclientv2.NewOutboundApiWithConfig(clientConfig)
	return &outboundRulesetProxy{
		clientConfig:                   clientConfig,
		outboundApi:                    api,
		createOutboundRulesetAttr:      createOutboundRulesetFn,
		getAllOutboundRulesetAttr:      getAllOutboundRulesetFn,
		getOutboundRulesetByIdAttr:     getOutboundRulesetByIdFn,
		getOutboundRulesetIdByNameAttr: getOutboundRulesetIdByNameFn,
		updateOutboundRulesetAttr:      updateOutboundRulesetFn,
		deleteOutboundRulesetAttr:      deleteOutboundRulesetFn,
	}
}

// getOutboundRulesetProxy acts as a singleton to for the internalProxy.  It also ensures
// that we can still proxy our tests by directly setting internalProxy package variable
func getOutboundRulesetProxy(clientConfig *platformclientv2.Configuration) *outboundRulesetProxy {
	if internalProxy == nil {
		internalProxy = newOutboundRulesetProxy(clientConfig)
	}

	return internalProxy
}

// createOutboundRuleset creates a Genesys Cloud Outbound Ruleset
func (p *outboundRulesetProxy) createOutboundRuleset(ctx context.Context, ruleset *platformclientv2.Ruleset) (*platformclientv2.Ruleset, error) {
	return p.createOutboundRulesetAttr(ctx, p, ruleset)
}

// getOutboundRuleset retrieves all Genesys Cloud Outbound Ruleset
func (p *outboundRulesetProxy) getAllOutboundRuleset(ctx context.Context) (*[]platformclientv2.Ruleset, error) {
	return p.getAllOutboundRulesetAttr(ctx, p)
}

// getOutboundRulesetById returns a single Genesys Cloud Outbound Ruleset by Id
func (p *outboundRulesetProxy) getOutboundRulesetById(ctx context.Context, rulesetId string) (ruleset *platformclientv2.Ruleset, statusCode int, err error) {
	return p.getOutboundRulesetByIdAttr(ctx, p, rulesetId)
}

// getOutboundRulesetIdByName returns a single Genesys Cloud Outbound Ruleset by a name
func (p *outboundRulesetProxy) getOutboundRulesetIdByName(ctx context.Context, name string) (rulesetId string, retryable bool, err error) {
	return p.getOutboundRulesetIdByNameAttr(ctx, p, name)
}

// updateOutboundRuleset updates a Genesys Cloud Outbound Ruleset
func (p *outboundRulesetProxy) updateOutboundRuleset(ctx context.Context, rulesetId string, ruleset *platformclientv2.Ruleset) (*platformclientv2.Ruleset, error) {
	return p.updateOutboundRulesetAttr(ctx, p, rulesetId, ruleset)
}

// deleteOutboundRuleset deletes a Genesys Cloud Outbound Ruleset by Id
func (p *outboundRulesetProxy) deleteOutboundRuleset(ctx context.Context, rulesetId string) (statusCode int, err error) {
	return p.deleteOutboundRulesetAttr(ctx, p, rulesetId)
}

// createOutboundRulesetFn is an implementation function for creating a Genesys Cloud Outbound Ruleset
func createOutboundRulesetFn(ctx context.Context, p *outboundRulesetProxy, ruleset *platformclientv2.Ruleset) (*platformclientv2.Ruleset, error) {
	ruleset, _, err := p.outboundApi.PostOutboundRulesets(*ruleset)
	if err != nil {
		return nil, fmt.Errorf("Failed to create ruleset: %s", err)
	}

	return ruleset, nil
}

// getAllOutboundRulesetFn is the implementation for retrieving all outbound ruleset in Genesys Cloud
func getAllOutboundRulesetFn(ctx context.Context, p *outboundRulesetProxy) (*[]platformclientv2.Ruleset, error) {
	var allRulesets []platformclientv2.Ruleset

	for pageNum := 1; ; pageNum++ {
		const pageSize = 100

		rulesets, _, err := p.outboundApi.GetOutboundRulesets(pageSize, pageNum, true, "", "", "", "")
		if err != nil {
			return nil, fmt.Errorf("Failed to get outbound rulesets: %v", err)
		}

		if rulesets.Entities == nil || len(*rulesets.Entities) == 0 {
			break
		}

		for _, ruleset := range *rulesets.Entities {
			log.Printf("Dealing with ruleset id : %s", *ruleset.Id)
			allRulesets = append(allRulesets, ruleset)
		}
	}

	return &allRulesets, nil
}

// getOutboundRulesetByIdFn is an implementation of the function to get a Genesys Cloud Outbound Ruleset by Id
func getOutboundRulesetByIdFn(ctx context.Context, p *outboundRulesetProxy, rulesetId string) (ruleset *platformclientv2.Ruleset, statusCode int, err error) {
	ruleset, resp, err := p.outboundApi.GetOutboundRuleset(rulesetId)
	if err != nil {
		//This is an API that throws an error on a 404 instead of just returning a 404.
		if strings.Contains(fmt.Sprintf("%s", err), "API Error: 404") {
			return nil, http.StatusNotFound, nil

		}
		return nil, 0, fmt.Errorf("Failed to retrieve ruleset by id %s: %s", rulesetId, err)
	}

	return ruleset, resp.StatusCode, nil
}

// getOutboundRulesetIdBySearchFn is an implementation of the function to get a Genesys Cloud Outbound Ruleset by name
func getOutboundRulesetIdByNameFn(ctx context.Context, p *outboundRulesetProxy, name string) (rulesetId string, retryable bool, err error) {
	const pageNum = 1
	const pageSize = 100
	rulesets, _, err := p.outboundApi.GetOutboundRulesets(pageSize, pageNum, true, "", name, "", "")
	if err != nil {
		return "", false, fmt.Errorf("Error searching outbound ruleset %s: %s", name, err)
	}

	if rulesets.Entities == nil || len(*rulesets.Entities) == 0 {
		return "", true, fmt.Errorf("No outbound ruleset found with name %s", name)
	}

	var ruleset platformclientv2.Ruleset
	entities := *rulesets.Entities

	for _, rulesetSdk := range entities {
		if *rulesetSdk.Name == name {
			log.Printf("Retrieved the ruleset id %s by name %s", *rulesetSdk.Id, name)
			ruleset = rulesetSdk
			return *ruleset.Id, false, nil
		}
	}

	return "", false, fmt.Errorf("Unable to find ruleset with name %s", name)
}

// updateOutboundRulesetFn is an implementation of the function to update a Genesys Cloud Outbound Rulesets
func updateOutboundRulesetFn(ctx context.Context, p *outboundRulesetProxy, rulesetId string, ruleset *platformclientv2.Ruleset) (*platformclientv2.Ruleset, error) {
	outboundRuleset, _, err := getOutboundRulesetByIdFn(ctx, p, rulesetId)
	if err != nil {
		return nil, fmt.Errorf("Failed to ruleset by id %s: %s", rulesetId, err)
	}

	ruleset.Version = outboundRuleset.Version
	ruleset, _, err = p.outboundApi.PutOutboundRuleset(rulesetId, *ruleset)
	if err != nil {
		return nil, fmt.Errorf("Failed to update ruleset: %s", err)
	}
	return ruleset, nil
}

// deleteOutboundRulesetFn is an implementation function for deleting a Genesys Cloud Outbound Rulesets
func deleteOutboundRulesetFn(ctx context.Context, p *outboundRulesetProxy, rulesetId string) (statusCode int, err error) {
	resp, err := p.outboundApi.DeleteOutboundRuleset(rulesetId)
	if err != nil {
		return resp.StatusCode, fmt.Errorf("Failed to delete ruleset: %s", err)
	}

	return resp.StatusCode, nil
}
