package journey_outcome

import (
	"context"
	"fmt"

	rc "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_cache"

	"github.com/mypurecloud/platform-client-sdk-go/v165/platformclientv2"
)

/*
The file genesyscloud_journey_outcome_proxy.go manages the interaction between our software and
the Genesys Cloud SDK. Within this file, we define proxy structures and methods.
We employ a technique called composition for each function on the proxy. This means that each function
is built by combining smaller, independent parts. One advantage of this approach is that it allows us
to isolate and test individual functions more easily. For testing purposes, we can replace or
simulate these smaller parts, known as stubs, to ensure that each function behaves correctly in different scenarios.
*/

// internalProxy holds a proxy instance that can be used throughout the package
var internalProxy *journeyOutcomeProxy

// Type definitions for each func on our proxy so we can easily mock them out later
type createJourneyOutcomeFunc func(ctx context.Context, p *journeyOutcomeProxy, outcome *platformclientv2.Outcomerequest) (*platformclientv2.Outcome, *platformclientv2.APIResponse, error)
type getAllJourneyOutcomesFunc func(ctx context.Context, p *journeyOutcomeProxy) (*[]platformclientv2.Outcome, *platformclientv2.APIResponse, error)
type getJourneyOutcomeIdByNameFunc func(ctx context.Context, p *journeyOutcomeProxy, name string) (id string, retryable bool, response *platformclientv2.APIResponse, err error)
type getJourneyOutcomeByIdFunc func(ctx context.Context, p *journeyOutcomeProxy, id string) (outcome *platformclientv2.Outcome, response *platformclientv2.APIResponse, err error)
type updateJourneyOutcomeFunc func(ctx context.Context, p *journeyOutcomeProxy, id string, outcome *platformclientv2.Patchoutcome) (*platformclientv2.Outcome, *platformclientv2.APIResponse, error)
type deleteJourneyOutcomeFunc func(ctx context.Context, p *journeyOutcomeProxy, id string) (*platformclientv2.APIResponse, error)

/*
The journeyOutcomeProxy struct holds all the methods responsible for making calls to
the Genesys Cloud APIs. This means that within this struct, you'll find all the functions designed
to interact directly with the various features and services offered by Genesys Cloud,
enabling this terraform provider software to perform tasks like retrieving data, updating information,
or triggering actions within the Genesys Cloud environment.
*/
type journeyOutcomeProxy struct {
	clientConfig                  *platformclientv2.Configuration
	journeyApi                    *platformclientv2.JourneyApi
	createJourneyOutcomeAttr      createJourneyOutcomeFunc
	getAllJourneyOutcomesAttr     getAllJourneyOutcomesFunc
	getJourneyOutcomeIdByNameAttr getJourneyOutcomeIdByNameFunc
	getJourneyOutcomeByIdAttr     getJourneyOutcomeByIdFunc
	updateJourneyOutcomeAttr      updateJourneyOutcomeFunc
	deleteJourneyOutcomeAttr      deleteJourneyOutcomeFunc
	outcomeCache                  rc.CacheInterface[platformclientv2.Outcome]
}

/*
The function newJourneyOutcomeProxy sets up the journey outcome proxy by providing it
with all the necessary information to communicate effectively with Genesys Cloud.
This includes configuring the proxy with the required data and settings so that it can interact
seamlessly with the Genesys Cloud platform.
*/
func newJourneyOutcomeProxy(clientConfig *platformclientv2.Configuration) *journeyOutcomeProxy {
	api := platformclientv2.NewJourneyApiWithConfig(clientConfig)
	outcomeCache := rc.NewResourceCache[platformclientv2.Outcome]()

	return &journeyOutcomeProxy{
		clientConfig:                  clientConfig,
		journeyApi:                    api,
		outcomeCache:                  outcomeCache,
		createJourneyOutcomeAttr:      createJourneyOutcomeFn,
		getAllJourneyOutcomesAttr:     getAllJourneyOutcomesFn,
		getJourneyOutcomeIdByNameAttr: getJourneyOutcomeIdByNameFn,
		getJourneyOutcomeByIdAttr:     getJourneyOutcomeByIdFn,
		updateJourneyOutcomeAttr:      updateJourneyOutcomeFn,
		deleteJourneyOutcomeAttr:      deleteJourneyOutcomeFn,
	}
}

/*
The function getJourneyOutcomeProxy serves a dual purpose: first, it functions as a singleton for
the internalProxy, meaning it ensures that only one instance of the internalProxy exists. Second,
it enables us to proxy our tests by allowing us to directly set the internalProxy package variable.
This ensures consistency and control in managing the internalProxy across our codebase, while also
facilitating efficient testing by providing a straightforward way to substitute the proxy for testing purposes.
*/
func getJourneyOutcomeProxy(clientConfig *platformclientv2.Configuration) *journeyOutcomeProxy {
	if internalProxy == nil {
		internalProxy = newJourneyOutcomeProxy(clientConfig)
	}
	return internalProxy
}

// createJourneyOutcome creates a Genesys Cloud journey outcome
func (p *journeyOutcomeProxy) createJourneyOutcome(ctx context.Context, outcome *platformclientv2.Outcomerequest) (*platformclientv2.Outcome, *platformclientv2.APIResponse, error) {
	return p.createJourneyOutcomeAttr(ctx, p, outcome)
}

// getAllJourneyOutcomes retrieves all Genesys Cloud journey outcomes
func (p *journeyOutcomeProxy) getAllJourneyOutcomes(ctx context.Context) (*[]platformclientv2.Outcome, *platformclientv2.APIResponse, error) {
	return p.getAllJourneyOutcomesAttr(ctx, p)
}

// getJourneyOutcomeIdByName returns a single Genesys Cloud journey outcome by name
func (p *journeyOutcomeProxy) getJourneyOutcomeIdByName(ctx context.Context, name string) (id string, retryable bool, response *platformclientv2.APIResponse, err error) {
	return p.getJourneyOutcomeIdByNameAttr(ctx, p, name)
}

// getJourneyOutcomeById returns a single Genesys Cloud journey outcome by Id
func (p *journeyOutcomeProxy) getJourneyOutcomeById(ctx context.Context, id string) (outcome *platformclientv2.Outcome, response *platformclientv2.APIResponse, err error) {
	if outcome := rc.GetCacheItem(p.outcomeCache, id); outcome != nil {
		return outcome, nil, nil
	}
	return p.getJourneyOutcomeByIdAttr(ctx, p, id)
}

// updateJourneyOutcome updates a Genesys Cloud journey outcome
func (p *journeyOutcomeProxy) updateJourneyOutcome(ctx context.Context, id string, outcome *platformclientv2.Patchoutcome) (*platformclientv2.Outcome, *platformclientv2.APIResponse, error) {
	return p.updateJourneyOutcomeAttr(ctx, p, id, outcome)
}

// deleteJourneyOutcome deletes a Genesys Cloud journey outcome by Id
func (p *journeyOutcomeProxy) deleteJourneyOutcome(ctx context.Context, id string) (*platformclientv2.APIResponse, error) {
	return p.deleteJourneyOutcomeAttr(ctx, p, id)
}

// createJourneyOutcomeFn is an implementation function for creating a Genesys Cloud journey outcome
func createJourneyOutcomeFn(ctx context.Context, p *journeyOutcomeProxy, outcome *platformclientv2.Outcomerequest) (*platformclientv2.Outcome, *platformclientv2.APIResponse, error) {
	outcomeRes, resp, err := p.journeyApi.PostJourneyOutcomes(*outcome)
	if err != nil {
		return nil, resp, err
	}
	return outcomeRes, resp, nil
}

// getAllJourneyOutcomesFn is the implementation for retrieving all journey outcomes in Genesys Cloud
func getAllJourneyOutcomesFn(ctx context.Context, p *journeyOutcomeProxy) (*[]platformclientv2.Outcome, *platformclientv2.APIResponse, error) {
	var allOutcomes []platformclientv2.Outcome
	const pageSize = 100

	outcomes, resp, err := p.journeyApi.GetJourneyOutcomes(1, pageSize, "", nil, nil, "")
	if err != nil {
		return nil, resp, err
	}

	if outcomes == nil || outcomes.Entities == nil || len(*outcomes.Entities) == 0 {
		return &allOutcomes, resp, nil
	}

	allOutcomes = append(allOutcomes, *outcomes.Entities...)

	for pageNum := 2; pageNum <= *outcomes.PageCount; pageNum++ {
		outcomes, resp, err := p.journeyApi.GetJourneyOutcomes(pageNum, pageSize, "", nil, nil, "")
		if err != nil {
			return nil, resp, err
		}
		if outcomes == nil || outcomes.Entities == nil || len(*outcomes.Entities) == 0 {
			break
		}

		allOutcomes = append(allOutcomes, *outcomes.Entities...)
	}

	// Cache the outcomes
	for _, outcome := range allOutcomes {
		rc.SetCache(p.outcomeCache, *outcome.Id, outcome)
	}

	return &allOutcomes, resp, nil
}

// getJourneyOutcomeIdByNameFn is an implementation function for getting a journey outcome by name
func getJourneyOutcomeIdByNameFn(ctx context.Context, p *journeyOutcomeProxy, name string) (id string, retryable bool, response *platformclientv2.APIResponse, err error) {
	outcomes, resp, err := p.getAllJourneyOutcomes(ctx)
	if err != nil {
		return "", false, resp, err
	}

	if outcomes == nil || len(*outcomes) == 0 {
		return "", true, resp, fmt.Errorf("No journey outcomes found with name %s", name)
	}

	for _, outcome := range *outcomes {
		if *outcome.DisplayName == name {
			return *outcome.Id, false, resp, nil
		}
	}

	return "", true, resp, fmt.Errorf("No journey outcomes found with name %s", name)
}

// getJourneyOutcomeByIdFn is an implementation function for getting a journey outcome by ID
func getJourneyOutcomeByIdFn(ctx context.Context, p *journeyOutcomeProxy, id string) (outcome *platformclientv2.Outcome, response *platformclientv2.APIResponse, err error) {
	outcome, resp, err := p.journeyApi.GetJourneyOutcome(id)
	if err != nil {
		return nil, resp, err
	}
	return outcome, resp, nil
}

// updateJourneyOutcomeFn is an implementation function for updating a journey outcome
func updateJourneyOutcomeFn(ctx context.Context, p *journeyOutcomeProxy, id string, outcome *platformclientv2.Patchoutcome) (*platformclientv2.Outcome, *platformclientv2.APIResponse, error) {
	outcomeRes, resp, err := p.journeyApi.PatchJourneyOutcome(id, *outcome)
	if err != nil {
		return nil, resp, err
	}
	return outcomeRes, resp, nil
}

// deleteJourneyOutcomeFn is an implementation function for deleting a journey outcome
func deleteJourneyOutcomeFn(ctx context.Context, p *journeyOutcomeProxy, id string) (*platformclientv2.APIResponse, error) {
	resp, err := p.journeyApi.DeleteJourneyOutcome(id)
	if err != nil {
		return resp, err
	}
	rc.DeleteCacheItem(p.outcomeCache, id)
	return resp, nil
}
