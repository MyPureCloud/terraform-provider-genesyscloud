package quality_forms_evaluation

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/mypurecloud/platform-client-sdk-go/v165/platformclientv2"
)

/*
The file genesyscloud_quality_forms_evaluation_proxy.go manages the interaction between our software and
the Genesys Cloud SDK. Within this file, we define proxy structures and methods.
We employ a technique called composition for each function on the proxy. This means that each function
is built by combining smaller, independent parts. One advantage of this approach is that it allows us
to isolate and test individual functions more easily. For testing purposes, we can replace or
simulate these smaller parts, known as stubs, to ensure that each function behaves correctly in different scenarios.
*/

// internalProxy holds a proxy instance that can be used throughout the package
var internalProxy *qualityFormsEvaluationProxy

// Type definitions for each func on our proxy so we can easily mock them out later
type createQualityFormsEvaluationFunc func(ctx context.Context, p *qualityFormsEvaluationProxy, evaluationForm *platformclientv2.Evaluationform) (*platformclientv2.Evaluationformresponse, *platformclientv2.APIResponse, error)
type getAllQualityFormsEvaluationFunc func(ctx context.Context, p *qualityFormsEvaluationProxy) (*[]platformclientv2.Evaluationformresponse, *platformclientv2.APIResponse, error)
type getQualityFormsEvaluationIdByNameFunc func(ctx context.Context, p *qualityFormsEvaluationProxy, name string) (id string, retryable bool, response *platformclientv2.APIResponse, err error)
type getQualityFormsEvaluationByIdFunc func(ctx context.Context, p *qualityFormsEvaluationProxy, id string) (evaluationForm *platformclientv2.Evaluationformresponse, response *platformclientv2.APIResponse, err error)
type updateQualityFormsEvaluationFunc func(ctx context.Context, p *qualityFormsEvaluationProxy, id string, evaluationForm *platformclientv2.Evaluationform) (*platformclientv2.Evaluationformresponse, *platformclientv2.APIResponse, error)
type deleteQualityFormsEvaluationFunc func(ctx context.Context, p *qualityFormsEvaluationProxy, id string) (*platformclientv2.APIResponse, error)
type publishQualityFormsEvaluationFunc func(ctx context.Context, p *qualityFormsEvaluationProxy, id string) (*platformclientv2.Evaluationformresponse, *platformclientv2.APIResponse, error)
type getQualityFormsEvaluationsBulkContextsFunc func(ctx context.Context, p *qualityFormsEvaluationProxy, contextIds []string) ([]platformclientv2.Evaluationformresponse, *platformclientv2.APIResponse, error)
type getEvaluationFormRecentVerIdFunc func(ctx context.Context, p *qualityFormsEvaluationProxy, formId string) (string, *platformclientv2.APIResponse, error)

/*
The qualityFormsEvaluationProxy struct holds all the methods responsible for making calls to
the Genesys Cloud APIs. This means that within this struct, you'll find all the functions designed
to interact directly with the various features and services offered by Genesys Cloud,
enabling this terraform provider software to perform tasks like retrieving data, updating information,
or triggering actions within the Genesys Cloud environment.
*/
type qualityFormsEvaluationProxy struct {
	clientConfig                               *platformclientv2.Configuration
	qualityApi                                 *platformclientv2.QualityApi
	createQualityFormsEvaluationAttr           createQualityFormsEvaluationFunc
	getAllQualityFormsEvaluationAttr           getAllQualityFormsEvaluationFunc
	getQualityFormsEvaluationIdByNameAttr      getQualityFormsEvaluationIdByNameFunc
	getQualityFormsEvaluationByIdAttr          getQualityFormsEvaluationByIdFunc
	updateQualityFormsEvaluationAttr           updateQualityFormsEvaluationFunc
	deleteQualityFormsEvaluationAttr           deleteQualityFormsEvaluationFunc
	publishQualityFormsEvaluationAttr          publishQualityFormsEvaluationFunc
	getQualityFormsEvaluationsBulkContextsAttr getQualityFormsEvaluationsBulkContextsFunc
	getEvaluationFormRecentVerIdAttr           getEvaluationFormRecentVerIdFunc
}

/*
The function newQualityFormsEvaluationProxy sets up the quality forms evaluation proxy by providing it
with all the necessary information to communicate effectively with Genesys Cloud.
This includes configuring the proxy with the required data and settings so that it can interact
seamlessly with the Genesys Cloud platform.
*/
func newQualityFormsEvaluationProxy(clientConfig *platformclientv2.Configuration) *qualityFormsEvaluationProxy {
	api := platformclientv2.NewQualityApiWithConfig(clientConfig) // NewQualityApiWithConfig creates a Genesys Cloud API instance using the provided configuration
	return &qualityFormsEvaluationProxy{
		clientConfig:                               clientConfig,
		qualityApi:                                 api,
		createQualityFormsEvaluationAttr:           createQualityFormsEvaluationFn,
		getAllQualityFormsEvaluationAttr:           getAllQualityFormsEvaluationFn,
		getQualityFormsEvaluationIdByNameAttr:      getQualityFormsEvaluationIdByNameFn,
		getQualityFormsEvaluationByIdAttr:          getQualityFormsEvaluationByIdFn,
		updateQualityFormsEvaluationAttr:           updateQualityFormsEvaluationFn,
		deleteQualityFormsEvaluationAttr:           deleteQualityFormsEvaluationFn,
		publishQualityFormsEvaluationAttr:          publishQualityFormsEvaluationFn,
		getQualityFormsEvaluationsBulkContextsAttr: getQualityFormsEvaluationsBulkContextsFn,
		getEvaluationFormRecentVerIdAttr:           getEvaluationFormRecentVerIdFn,
	}
}

/*
The function getQualityFormsEvaluationProxy serves a dual purpose: first, it functions as a singleton for
the internalProxy, meaning it ensures that only one instance of the internalProxy exists. Second,
it enables us to proxy our tests by allowing us to directly set the internalProxy package variable.
This ensures consistency and control in managing the internalProxy across our codebase, while also
facilitating efficient testing by providing a straightforward way to substitute the proxy for testing purposes.
*/
func getQualityFormsEvaluationProxy(clientConfig *platformclientv2.Configuration) *qualityFormsEvaluationProxy {
	if internalProxy == nil {
		internalProxy = newQualityFormsEvaluationProxy(clientConfig)
	}
	return internalProxy
}

// createQualityFormsEvaluation creates a Genesys Cloud quality forms evaluation
func (p *qualityFormsEvaluationProxy) createQualityFormsEvaluation(ctx context.Context, evaluationForm *platformclientv2.Evaluationform) (*platformclientv2.Evaluationformresponse, *platformclientv2.APIResponse, error) {
	return p.createQualityFormsEvaluationAttr(ctx, p, evaluationForm)
}

// getAllQualityFormsEvaluation retrieves all Genesys Cloud quality forms evaluations
func (p *qualityFormsEvaluationProxy) getAllQualityFormsEvaluation(ctx context.Context) (*[]platformclientv2.Evaluationformresponse, *platformclientv2.APIResponse, error) {
	return p.getAllQualityFormsEvaluationAttr(ctx, p)
}

// getQualityFormsEvaluationIdByName returns a single Genesys Cloud quality forms evaluation by a name
func (p *qualityFormsEvaluationProxy) getQualityFormsEvaluationIdByName(ctx context.Context, name string) (id string, retryable bool, response *platformclientv2.APIResponse, err error) {
	return p.getQualityFormsEvaluationIdByNameAttr(ctx, p, name)
}

// getQualityFormsEvaluationById returns a single Genesys Cloud quality forms evaluation by Id
func (p *qualityFormsEvaluationProxy) getQualityFormsEvaluationById(ctx context.Context, id string) (evaluationForm *platformclientv2.Evaluationformresponse, response *platformclientv2.APIResponse, err error) {
	return p.getQualityFormsEvaluationByIdAttr(ctx, p, id)
}

// updateQualityFormsEvaluation updates a Genesys Cloud quality forms evaluation
func (p *qualityFormsEvaluationProxy) updateQualityFormsEvaluation(ctx context.Context, id string, evaluationForm *platformclientv2.Evaluationform) (*platformclientv2.Evaluationformresponse, *platformclientv2.APIResponse, error) {
	return p.updateQualityFormsEvaluationAttr(ctx, p, id, evaluationForm)
}

// deleteQualityFormsEvaluation deletes a Genesys Cloud quality forms evaluation by Id
func (p *qualityFormsEvaluationProxy) deleteQualityFormsEvaluation(ctx context.Context, id string) (*platformclientv2.APIResponse, error) {
	return p.deleteQualityFormsEvaluationAttr(ctx, p, id)
}

// publishQualityFormsEvaluation publishes a Genesys Cloud quality forms evaluation by Id
func (p *qualityFormsEvaluationProxy) publishQualityFormsEvaluation(ctx context.Context, id string) (*platformclientv2.Evaluationformresponse, *platformclientv2.APIResponse, error) {
	return p.publishQualityFormsEvaluationAttr(ctx, p, id)
}

// getQualityFormsEvaluationsBulkContexts retrieves published evaluation forms by context IDs
func (p *qualityFormsEvaluationProxy) getQualityFormsEvaluationsBulkContexts(ctx context.Context, contextIds []string) ([]platformclientv2.Evaluationformresponse, *platformclientv2.APIResponse, error) {
	return p.getQualityFormsEvaluationsBulkContextsAttr(ctx, p, contextIds)
}

// getEvaluationFormRecentVerId retrieves the latest unpublished version ID of a form
func (p *qualityFormsEvaluationProxy) getEvaluationFormRecentVerId(ctx context.Context, formId string) (string, *platformclientv2.APIResponse, error) {
	return p.getEvaluationFormRecentVerIdAttr(ctx, p, formId)
}

// publishQualityFormsEvaluationFn is an implementation function for publishing a Genesys Cloud quality forms evaluation
func publishQualityFormsEvaluationFn(ctx context.Context, p *qualityFormsEvaluationProxy, id string) (*platformclientv2.Evaluationformresponse, *platformclientv2.APIResponse, error) {
	// Check if the form is already published
	form, apiResponse, err := p.qualityApi.GetQualityFormsEvaluation(id)
	if err != nil {
		return nil, apiResponse, fmt.Errorf("Failed to check existing state of quality forms evaluation: %s", err)
	}

	if *form.Published {
		log.Printf("No need to publish form '%s' because it's already published", id)
		return nil, nil, nil
	}

	// Publish the form
	newDraftEval, apiResponse, err := p.qualityApi.PostQualityPublishedformsEvaluations(platformclientv2.Publishform{
		Id:        &id,
		Published: platformclientv2.Bool(true),
	})

	if err != nil {
		return nil, apiResponse, fmt.Errorf("Failed to publish quality forms evaluation: %s", err)
	}

	return newDraftEval, apiResponse, nil
}

// createQualityFormsEvaluationFn is an implementation function for creating a Genesys Cloud quality forms evaluation
func createQualityFormsEvaluationFn(ctx context.Context, p *qualityFormsEvaluationProxy, evaluationForm *platformclientv2.Evaluationform) (*platformclientv2.Evaluationformresponse, *platformclientv2.APIResponse, error) {
	form, apiResponse, err := p.qualityApi.PostQualityFormsEvaluations(*evaluationForm)
	if err != nil {
		return nil, apiResponse, fmt.Errorf("Failed to create quality forms evaluation: %s", err)
	}
	return form, apiResponse, nil
}

// getAllQualityFormsEvaluationFn is the implementation for retrieving all quality forms evaluations in Genesys Cloud
func getAllQualityFormsEvaluationFn(ctx context.Context, p *qualityFormsEvaluationProxy) (*[]platformclientv2.Evaluationformresponse, *platformclientv2.APIResponse, error) {
	var allForms []platformclientv2.Evaluationformresponse
	const pageSize = 100

	forms, apiResponse, err := p.qualityApi.GetQualityFormsEvaluations(pageSize, 1, "", "", "", "", "", "")
	if err != nil {
		return nil, apiResponse, fmt.Errorf("Failed to get evaluation forms: %v", err)
	}

	if forms == nil || forms.Entities == nil || len(*forms.Entities) == 0 {
		return &allForms, apiResponse, nil
	}

	allForms = append(allForms, *forms.Entities...)

	for pageNum := 2; pageNum <= *forms.PageCount; pageNum++ {
		forms, apiResponse, err := p.qualityApi.GetQualityFormsEvaluations(pageSize, pageNum, "", "", "", "", "", "")
		if err != nil {
			return nil, apiResponse, fmt.Errorf("Failed to get evaluation forms: %v", err)
		}

		if forms == nil || forms.Entities == nil || len(*forms.Entities) == 0 {
			break
		}

		allForms = append(allForms, *forms.Entities...)
	}

	return &allForms, apiResponse, nil
}

// getQualityFormsEvaluationIdByNameFn is an implementation of the function to get a Genesys Cloud quality forms evaluation by name
func getQualityFormsEvaluationIdByNameFn(ctx context.Context, p *qualityFormsEvaluationProxy, name string) (id string, retryable bool, response *platformclientv2.APIResponse, err error) {
	forms, apiResponse, err := getAllQualityFormsEvaluationFn(ctx, p)
	if err != nil {
		return "", false, apiResponse, err
	}

	if forms == nil || len(*forms) == 0 {
		return "", true, apiResponse, fmt.Errorf("No quality forms evaluation found with name %s", name)
	}

	for _, form := range *forms {
		if *form.Name == name {
			log.Printf("Retrieved the quality forms evaluation id %s by name %s", *form.Id, name)
			return *form.Id, false, apiResponse, nil
		}
	}

	return "", true, apiResponse, fmt.Errorf("Unable to find quality forms evaluation with name %s", name)
}

// getQualityFormsEvaluationByIdFn is an implementation of the function to get a Genesys Cloud quality forms evaluation by Id
func getQualityFormsEvaluationByIdFn(ctx context.Context, p *qualityFormsEvaluationProxy, id string) (evaluationForm *platformclientv2.Evaluationformresponse, response *platformclientv2.APIResponse, err error) {
	form, apiResponse, err := p.qualityApi.GetQualityFormsEvaluation(id)
	if err != nil {
		return nil, apiResponse, fmt.Errorf("Failed to retrieve quality forms evaluation by id %s: %s", id, err)
	}
	return form, apiResponse, nil
}

// updateQualityFormsEvaluationFn is an implementation of the function to update a Genesys Cloud quality forms evaluation
func updateQualityFormsEvaluationFn(ctx context.Context, p *qualityFormsEvaluationProxy, id string, evaluationForm *platformclientv2.Evaluationform) (*platformclientv2.Evaluationformresponse, *platformclientv2.APIResponse, error) {
	_, apiResponse, err := getQualityFormsEvaluationByIdFn(ctx, p, id)
	if err != nil {
		return nil, apiResponse, fmt.Errorf("Failed to get quality forms evaluation %s by id: %s", id, err)
	}
	// TODO evaluationForm.Version = form.Version
	formResponse, apiResponse, err := p.qualityApi.PutQualityFormsEvaluation(id, *evaluationForm)
	if err != nil {
		return nil, apiResponse, fmt.Errorf("Failed to update quality forms evaluation: %s", err)
	}
	return formResponse, apiResponse, nil
}

// deleteQualityFormsEvaluationFn is an implementation function for deleting a Genesys Cloud quality forms evaluation
func deleteQualityFormsEvaluationFn(ctx context.Context, p *qualityFormsEvaluationProxy, id string) (*platformclientv2.APIResponse, error) {
	resp, err := p.qualityApi.DeleteQualityFormsEvaluation(id)
	if err != nil {
		return resp, fmt.Errorf("Failed to delete quality forms evaluation: %s", err)
	}
	return resp, nil
}

// getQualityFormsEvaluationsBulkContextsFn is an implementation function for retrieving published evaluation forms by context IDs
func getQualityFormsEvaluationsBulkContextsFn(ctx context.Context, p *qualityFormsEvaluationProxy, contextIds []string) ([]platformclientv2.Evaluationformresponse, *platformclientv2.APIResponse, error) {
	publishedVersions, apiResponse, err := p.qualityApi.GetQualityFormsEvaluationsBulkContexts(contextIds)
	if err != nil {
		return nil, apiResponse, fmt.Errorf("Failed to retrieve published evaluation forms by context IDs: %s", err)
	}
	return publishedVersions, apiResponse, nil
}

// getEvaluationFormRecentVerIdFn is an implementation function for retrieving the latest unpublished version ID of a form
func getEvaluationFormRecentVerIdFn(ctx context.Context, p *qualityFormsEvaluationProxy, formId string) (string, *platformclientv2.APIResponse, error) {
	const maxRetries = 3
	var wait = 1 * time.Second

	for i := 0; i < maxRetries; i++ {
		// Get the latest unpublished version of the form
		formVersions, apiResponse, err := p.qualityApi.GetQualityFormsEvaluationVersions(formId, 25, 1, "desc")
		if err != nil {
			return "", apiResponse, fmt.Errorf("Failed to get evaluation form versions for form %s: %s", formId, err)
		}
		if formVersions == nil || formVersions.Entities == nil || len(*formVersions.Entities) == 0 {
			time.Sleep(wait)
			continue
		}

		for _, form := range *formVersions.Entities {
			if !*form.Published {
				return *form.Id, apiResponse, nil
			}
		}
	}

	return "", nil, fmt.Errorf("Could not find any unpublished versions of the form '%s'", formId)
}
