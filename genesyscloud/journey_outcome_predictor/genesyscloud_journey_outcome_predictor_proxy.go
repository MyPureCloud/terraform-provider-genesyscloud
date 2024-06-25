package journey_outcome_predictor

import (
	"context"

	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

/*
The genesyscloud_journey_outcome_predictor_proxy.go file contains the proxy structures and methods that interact
with the Genesys Cloud SDK. We use composition here for each function on the proxy so individual functions can be stubbed
out during testing.
*/

// internalProxy holds a proxy instance that can be used throughout the package
var internalProxy *journeyOutcomePredictorProxy

// Type definitions for each func on our proxy so we can easily mock them out later
type createJourneyOutcomePredictorFunc func(ctx context.Context, p *journeyOutcomePredictorProxy, outcomePredictor *platformclientv2.Outcomepredictorrequest) (*platformclientv2.Outcomepredictor, *platformclientv2.APIResponse, error)
type getAllJourneyOutcomePredictorFunc func(ctx context.Context, p *journeyOutcomePredictorProxy) (*[]platformclientv2.Outcomepredictor, *platformclientv2.APIResponse, error)
type getJourneyOutcomePredictorByIdFunc func(ctx context.Context, p *journeyOutcomePredictorProxy, id string) (outcomePredictor *platformclientv2.Outcomepredictor, response *platformclientv2.APIResponse, err error)
type deleteJourneyOutcomePredictorFunc func(ctx context.Context, p *journeyOutcomePredictorProxy, id string) (response *platformclientv2.APIResponse, err error)

// journeyOutcomePredictorProxy contains all of the methods that call genesys cloud APIs.
type journeyOutcomePredictorProxy struct {
	clientConfig                       *platformclientv2.Configuration
	journeyApi                         *platformclientv2.JourneyApi
	createJourneyOutcomePredictorAttr  createJourneyOutcomePredictorFunc
	getAllJourneyOutcomePredictorAttr  getAllJourneyOutcomePredictorFunc
	getJourneyOutcomePredictorByIdAttr getJourneyOutcomePredictorByIdFunc
	deleteJourneyOutcomePredictorAttr  deleteJourneyOutcomePredictorFunc
}

// newJourneyOutcomePredictorProxy initializes the journey outcome predictor proxy with all of the data needed to communicate with Genesys Cloud
func newJourneyOutcomePredictorProxy(clientConfig *platformclientv2.Configuration) *journeyOutcomePredictorProxy {
	journeyApi := platformclientv2.NewJourneyApiWithConfig(clientConfig)
	return &journeyOutcomePredictorProxy{
		clientConfig:                       clientConfig,
		journeyApi:                         journeyApi,
		createJourneyOutcomePredictorAttr:  createJourneyOutcomePredictorFn,
		getAllJourneyOutcomePredictorAttr:  getAllJourneyOutcomePredictorFn,
		getJourneyOutcomePredictorByIdAttr: getJourneyOutcomePredictorByIdFn,
		deleteJourneyOutcomePredictorAttr:  deleteJourneyOutcomePredictorFn,
	}
}

// getJourneyOutcomePredictorProxy acts as a singleton to for the internalProxy.  It also ensures
// that we can still proxy our tests by directly setting internalProxy package variable
func getJourneyOutcomePredictorProxy(clientConfig *platformclientv2.Configuration) *journeyOutcomePredictorProxy {
	if internalProxy == nil {
		internalProxy = newJourneyOutcomePredictorProxy(clientConfig)
	}

	return internalProxy
}

// createJourneyOutcomePredictor creates a Genesys Cloud journey outcome predictor
func (p *journeyOutcomePredictorProxy) createJourneyOutcomePredictor(ctx context.Context, outcomePredictor *platformclientv2.Outcomepredictorrequest) (*platformclientv2.Outcomepredictor, *platformclientv2.APIResponse, error) {
	return p.createJourneyOutcomePredictorAttr(ctx, p, outcomePredictor)
}

// getJourneyOutcomePredictor retrieves all Genesys Cloud journey outcome predictor
func (p *journeyOutcomePredictorProxy) getAllJourneyOutcomePredictor(ctx context.Context) (*[]platformclientv2.Outcomepredictor, *platformclientv2.APIResponse, error) {
	return p.getAllJourneyOutcomePredictorAttr(ctx, p)
}

// getJourneyOutcomePredictorById returns a single Genesys Cloud journey outcome predictor by Id
func (p *journeyOutcomePredictorProxy) getJourneyOutcomePredictorById(ctx context.Context, predictorId string) (journeyOutcomePredictor *platformclientv2.Outcomepredictor, response *platformclientv2.APIResponse, err error) {
	return p.getJourneyOutcomePredictorByIdAttr(ctx, p, predictorId)
}

// deleteJourneyOutcomePredictor deletes a Genesys Cloud journey outcome predictor by Id
func (p *journeyOutcomePredictorProxy) deleteJourneyOutcomePredictor(ctx context.Context, predictorId string) (response *platformclientv2.APIResponse, err error) {
	return p.deleteJourneyOutcomePredictorAttr(ctx, p, predictorId)
}

// createJourneyOutcomePredictorFn is an implementation function for creating a Genesys Cloud journey outcome predictor
func createJourneyOutcomePredictorFn(ctx context.Context, p *journeyOutcomePredictorProxy, outcomePredictor *platformclientv2.Outcomepredictorrequest) (*platformclientv2.Outcomepredictor, *platformclientv2.APIResponse, error) {
	predictor, resp, err := p.journeyApi.PostJourneyOutcomesPredictors(*outcomePredictor)
	if err != nil {
		return nil, resp, err
	}

	return predictor, resp, nil
}

// getAllJourneyOutcomePredictorFn is the implementation for retrieving all journey outcome predictor in Genesys Cloud
func getAllJourneyOutcomePredictorFn(ctx context.Context, p *journeyOutcomePredictorProxy) (*[]platformclientv2.Outcomepredictor, *platformclientv2.APIResponse, error) {
	var allPredictors []platformclientv2.Outcomepredictor
	predictors, resp, err := p.journeyApi.GetJourneyOutcomesPredictors()

	if err != nil {
		return nil, resp, err
	}

	for _, predictor := range *predictors.Entities {
		allPredictors = append(allPredictors, predictor)
	}

	return &allPredictors, resp, nil
}

// getJourneyOutcomePredictorByIdFn is an implementation of the function to get a Genesys Cloud journey outcome predictor by Id
func getJourneyOutcomePredictorByIdFn(ctx context.Context, p *journeyOutcomePredictorProxy, predictorId string) (journeyOutcomePredictor *platformclientv2.Outcomepredictor, response *platformclientv2.APIResponse, err error) {
	predictor, resp, err := p.journeyApi.GetJourneyOutcomesPredictor(predictorId)
	if err != nil {
		return nil, resp, err
	}

	return predictor, resp, nil
}

// deleteJourneyOutcomePredictorFn is an implementation function for deleting a Genesys Cloud journey outcome predictor
func deleteJourneyOutcomePredictorFn(ctx context.Context, p *journeyOutcomePredictorProxy, predictorId string) (response *platformclientv2.APIResponse, err error) {
	resp, err := p.journeyApi.DeleteJourneyOutcomesPredictor(predictorId)
	if err != nil {
		return resp, err
	}

	return resp, nil
}
