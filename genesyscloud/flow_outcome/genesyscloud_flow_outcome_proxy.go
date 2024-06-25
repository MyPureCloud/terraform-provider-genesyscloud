package flow_outcome

import (
	"context"
	"fmt"
	"log"

	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

/*
The genesyscloud_flow_outcome_proxy.go file contains the proxy structures and methods that interact
with the Genesys Cloud SDK. We use composition here for each function on the proxy so individual functions can be stubbed
out during testing.
*/

// internalProxy holds a proxy instance that can be used throughout the package
var internalProxy *flowOutcomeProxy

// Type definitions for each func on our proxy so we can easily mock them out later
type createFlowOutcomeFunc func(ctx context.Context, p *flowOutcomeProxy, flowOutcome *platformclientv2.Flowoutcome) (*platformclientv2.Flowoutcome, *platformclientv2.APIResponse, error)
type getAllFlowOutcomeFunc func(ctx context.Context, p *flowOutcomeProxy) (*[]platformclientv2.Flowoutcome, *platformclientv2.APIResponse, error)
type getFlowOutcomeIdByNameFunc func(ctx context.Context, p *flowOutcomeProxy, name string) (id string, retryable bool, response *platformclientv2.APIResponse, err error)
type getFlowOutcomeByIdFunc func(ctx context.Context, p *flowOutcomeProxy, id string) (flowOutcome *platformclientv2.Flowoutcome, response *platformclientv2.APIResponse, err error)
type updateFlowOutcomeFunc func(ctx context.Context, p *flowOutcomeProxy, id string, flowOutcome *platformclientv2.Flowoutcome) (*platformclientv2.Flowoutcome, *platformclientv2.APIResponse, error)

// flowOutcomeProxy contains all of the methods that call genesys cloud APIs.
type flowOutcomeProxy struct {
	clientConfig               *platformclientv2.Configuration
	architectApi               *platformclientv2.ArchitectApi
	createFlowOutcomeAttr      createFlowOutcomeFunc
	getAllFlowOutcomeAttr      getAllFlowOutcomeFunc
	getFlowOutcomeIdByNameAttr getFlowOutcomeIdByNameFunc
	getFlowOutcomeByIdAttr     getFlowOutcomeByIdFunc
	updateFlowOutcomeAttr      updateFlowOutcomeFunc
}

// newFlowOutcomeProxy initializes the flow outcome proxy with all of the data needed to communicate with Genesys Cloud
func newFlowOutcomeProxy(clientConfig *platformclientv2.Configuration) *flowOutcomeProxy {
	api := platformclientv2.NewArchitectApiWithConfig(clientConfig)
	return &flowOutcomeProxy{
		clientConfig:               clientConfig,
		architectApi:               api,
		createFlowOutcomeAttr:      createFlowOutcomeFn,
		getAllFlowOutcomeAttr:      getAllFlowOutcomeFn,
		getFlowOutcomeIdByNameAttr: getFlowOutcomeIdByNameFn,
		getFlowOutcomeByIdAttr:     getFlowOutcomeByIdFn,
		updateFlowOutcomeAttr:      updateFlowOutcomeFn,
	}
}

// getFlowOutcomeProxy acts as a singleton to for the internalProxy.  It also ensures
// that we can still proxy our tests by directly setting internalProxy package variable
func getFlowOutcomeProxy(clientConfig *platformclientv2.Configuration) *flowOutcomeProxy {
	if internalProxy == nil {
		internalProxy = newFlowOutcomeProxy(clientConfig)
	}
	return internalProxy
}

// createFlowOutcome creates a Genesys Cloud flow outcome
func (p *flowOutcomeProxy) createFlowOutcome(ctx context.Context, flowOutcome *platformclientv2.Flowoutcome) (*platformclientv2.Flowoutcome, *platformclientv2.APIResponse, error) {
	return p.createFlowOutcomeAttr(ctx, p, flowOutcome)
}

// getFlowOutcome retrieves all Genesys Cloud flow outcome
func (p *flowOutcomeProxy) getAllFlowOutcome(ctx context.Context) (*[]platformclientv2.Flowoutcome, *platformclientv2.APIResponse, error) {
	return p.getAllFlowOutcomeAttr(ctx, p)
}

// getFlowOutcomeIdByName returns a single Genesys Cloud flow outcome by a name
func (p *flowOutcomeProxy) getFlowOutcomeIdByName(ctx context.Context, name string) (id string, retryable bool, response *platformclientv2.APIResponse, err error) {
	return p.getFlowOutcomeIdByNameAttr(ctx, p, name)
}

// getFlowOutcomeById returns a single Genesys Cloud flow outcome by Id
func (p *flowOutcomeProxy) getFlowOutcomeById(ctx context.Context, id string) (flowOutcome *platformclientv2.Flowoutcome, response *platformclientv2.APIResponse, err error) {
	return p.getFlowOutcomeByIdAttr(ctx, p, id)
}

// updateFlowOutcome updates a Genesys Cloud flow outcome
func (p *flowOutcomeProxy) updateFlowOutcome(ctx context.Context, id string, flowOutcome *platformclientv2.Flowoutcome) (*platformclientv2.Flowoutcome, *platformclientv2.APIResponse, error) {
	return p.updateFlowOutcomeAttr(ctx, p, id, flowOutcome)
}

// createFlowOutcomeFn is an implementation function for creating a Genesys Cloud flow outcome
func createFlowOutcomeFn(ctx context.Context, p *flowOutcomeProxy, flowOutcome *platformclientv2.Flowoutcome) (*platformclientv2.Flowoutcome, *platformclientv2.APIResponse, error) {
	flowOutcome, resp, err := p.architectApi.PostFlowsOutcomes(*flowOutcome)
	if err != nil {
		return nil, resp, fmt.Errorf("Failed to create flow outcome: %s", err)
	}
	return flowOutcome, resp, nil
}

// getAllFlowOutcomeFn is the implementation for retrieving all flow outcome in Genesys Cloud
func getAllFlowOutcomeFn(ctx context.Context, p *flowOutcomeProxy) (*[]platformclientv2.Flowoutcome, *platformclientv2.APIResponse, error) {
	var allFlowOutcomes []platformclientv2.Flowoutcome
	const pageSize = 100

	flowOutcomes, resp, err := p.architectApi.GetFlowsOutcomes(1, pageSize, "", "", nil, "", "", "", nil)
	if err != nil {
		return nil, resp, fmt.Errorf("Failed to get flow outcome: %v", err)
	}
	if flowOutcomes.Entities == nil || len(*flowOutcomes.Entities) == 0 {
		return &allFlowOutcomes, resp, nil
	}
	for _, flowOutcome := range *flowOutcomes.Entities {
		allFlowOutcomes = append(allFlowOutcomes, flowOutcome)
	}

	for pageNum := 2; pageNum <= *flowOutcomes.PageCount; pageNum++ {
		flowOutcomes, resp, err := p.architectApi.GetFlowsOutcomes(pageNum, pageSize, "", "", nil, "", "", "", nil)
		if err != nil {
			return nil, resp, fmt.Errorf("Failed to get flow outcome: %v", err)
		}

		if flowOutcomes.Entities == nil || len(*flowOutcomes.Entities) == 0 {
			break
		}

		for _, flowOutcome := range *flowOutcomes.Entities {
			allFlowOutcomes = append(allFlowOutcomes, flowOutcome)
		}
	}
	return &allFlowOutcomes, resp, nil
}

// getFlowOutcomeIdByNameFn is an implementation of the function to get a Genesys Cloud flow outcome by name
func getFlowOutcomeIdByNameFn(ctx context.Context, p *flowOutcomeProxy, name string) (id string, retryable bool, response *platformclientv2.APIResponse, err error) {
	flowOutcomes, resp, err := p.architectApi.GetFlowsOutcomes(1, 100, "", "", nil, name, "", "", nil)
	if err != nil {
		return "", false, resp, err
	}

	if flowOutcomes.Entities == nil || len(*flowOutcomes.Entities) == 0 {
		return "", true, resp, fmt.Errorf("No flow outcome found with name %s", name)
	}

	for _, flowOutcomeSdk := range *flowOutcomes.Entities {
		if *flowOutcomeSdk.Name == name {
			log.Printf("Retrieved the flow outcome id %s by name %s", *flowOutcomeSdk.Id, name)
			return *flowOutcomeSdk.Id, false, resp, nil
		}
	}
	return "", true, resp, fmt.Errorf("Unable to find flow outcome with name %s", name)
}

// getFlowOutcomeByIdFn is an implementation of the function to get a Genesys Cloud flow outcome by Id
func getFlowOutcomeByIdFn(ctx context.Context, p *flowOutcomeProxy, id string) (flowOutcome *platformclientv2.Flowoutcome, response *platformclientv2.APIResponse, err error) {
	flowOutcome, resp, err := p.architectApi.GetFlowsOutcome(id)
	if err != nil {
		return nil, resp, fmt.Errorf("Failed to retrieve flow outcome by id %s: %s", id, err)
	}
	return flowOutcome, resp, nil
}

// updateFlowOutcomeFn is an implementation of the function to update a Genesys Cloud flow outcome
func updateFlowOutcomeFn(ctx context.Context, p *flowOutcomeProxy, id string, flowOutcome *platformclientv2.Flowoutcome) (*platformclientv2.Flowoutcome, *platformclientv2.APIResponse, error) {
	_, resp, err := p.architectApi.PutFlowsOutcome(id, *flowOutcome)
	if err != nil {
		return nil, resp, fmt.Errorf("Failed to update flow outcome: %s", err)
	}
	return flowOutcome, resp, nil
}
