package quality_forms_survey

import (
	"context"
	"fmt"

	rc "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_cache"

	"github.com/mypurecloud/platform-client-sdk-go/v157/platformclientv2"
)

/*
The file genesyscloud_quality_forms_survey_proxy.go manages the interaction between our software and
the Genesys Cloud SDK. Within this file, we define proxy structures and methods.
We employ a technique called composition for each function on the proxy. This means that each function
is built by combining smaller, independent parts. One advantage of this approach is that it allows us
to isolate and test individual functions more easily. For testing purposes, we can replace or
simulate these smaller parts, known as stubs, to ensure that each function behaves correctly in different scenarios.
*/

// internalProxy holds a proxy instance that can be used throughout the package
var internalProxy *qualityFormsSurveyProxy

// Type definitions for each func on our proxy so we can easily mock them out later
type createQualityFormsSurveyFunc func(ctx context.Context, p *qualityFormsSurveyProxy, form *platformclientv2.Surveyform) (*platformclientv2.Surveyform, *platformclientv2.APIResponse, error)
type getAllQualityFormsSurveyFunc func(ctx context.Context, p *qualityFormsSurveyProxy) (*[]platformclientv2.Surveyform, *platformclientv2.APIResponse, error)
type getQualityFormsSurveyByNameFunc func(ctx context.Context, p *qualityFormsSurveyProxy, name string) (id string, retryable bool, response *platformclientv2.APIResponse, err error)
type getQualityFormsSurveyByIdFunc func(ctx context.Context, p *qualityFormsSurveyProxy, id string) (form *platformclientv2.Surveyform, response *platformclientv2.APIResponse, err error)
type updateQualityFormsSurveyFunc func(ctx context.Context, p *qualityFormsSurveyProxy, id string, form *platformclientv2.Surveyform) (*platformclientv2.Surveyform, *platformclientv2.APIResponse, error)
type deleteQualityFormsSurveyFunc func(ctx context.Context, p *qualityFormsSurveyProxy, id string) (*platformclientv2.APIResponse, error)
type publishQualityFormsSurveyFunc func(ctx context.Context, p *qualityFormsSurveyProxy, id string, published bool) (*platformclientv2.APIResponse, error)
type getQualityFormsSurveyVersionsFunc func(ctx context.Context, p *qualityFormsSurveyProxy, formId string, pageSize int, pageNumber int) (*platformclientv2.Surveyformentitylisting, *platformclientv2.APIResponse, error)
type patchQualityFormsSurveyFunc func(ctx context.Context, p *qualityFormsSurveyProxy, formId string, body platformclientv2.Surveyform) (*platformclientv2.Surveyform, *platformclientv2.APIResponse, error)

type qualityFormsSurveyProxy struct {
	clientConfig                      *platformclientv2.Configuration
	qualityApi                        *platformclientv2.QualityApi
	createQualityFormsSurveyAttr      createQualityFormsSurveyFunc
	getAllQualityFormsSurveyAttr      getAllQualityFormsSurveyFunc
	getQualityFormsSurveyByIdAttr     getQualityFormsSurveyByIdFunc
	getQualityFormsSurveyByNameAttr   getQualityFormsSurveyByNameFunc
	updateQualityFormsSurveyAttr      updateQualityFormsSurveyFunc
	deleteQualityFormsSurveyAttr      deleteQualityFormsSurveyFunc
	publishQualityFormsSurveyAttr     publishQualityFormsSurveyFunc
	getQualityFormsSurveyVersionsAttr getQualityFormsSurveyVersionsFunc
	patchQualityFormsSurveyAttr       patchQualityFormsSurveyFunc
	formsCache                        rc.CacheInterface[platformclientv2.Surveyform]
}

// newQualityFormsSurveyProxy initializes the quality forms survey proxy with all the data needed to communicate with Genesys Cloud
func newQualityFormsSurveyProxy(clientConfig *platformclientv2.Configuration) *qualityFormsSurveyProxy {
	api := platformclientv2.NewQualityApiWithConfig(clientConfig)
	formsCache := rc.NewResourceCache[platformclientv2.Surveyform]()
	return &qualityFormsSurveyProxy{
		clientConfig:                      clientConfig,
		qualityApi:                        api,
		formsCache:                        formsCache,
		createQualityFormsSurveyAttr:      createQualityFormsSurveyFn,
		getAllQualityFormsSurveyAttr:      getAllQualityFormsSurveyFn,
		getQualityFormsSurveyByIdAttr:     getQualityFormsSurveyByIdFn,
		getQualityFormsSurveyByNameAttr:   getQualityFormsSurveyByNameFn,
		updateQualityFormsSurveyAttr:      updateQualityFormsSurveyFn,
		deleteQualityFormsSurveyAttr:      deleteQualityFormsSurveyFn,
		publishQualityFormsSurveyAttr:     publishQualityFormsSurveyFn,
		getQualityFormsSurveyVersionsAttr: getQualityFormsSurveyVersionsFn,
		patchQualityFormsSurveyAttr:       patchQualityFormsSurveyFn,
	}
}

// getQualityFormsSurveyProxy acts as a singleton for the internalProxy and ensures
// only one instance of the proxy exists
func getQualityFormsSurveyProxy(clientConfig *platformclientv2.Configuration) *qualityFormsSurveyProxy {
	if internalProxy == nil {
		internalProxy = newQualityFormsSurveyProxy(clientConfig)
	}
	return internalProxy
}

// createQualityFormsSurvey creates a Genesys Cloud quality forms survey
func (p *qualityFormsSurveyProxy) createQualityFormsSurvey(ctx context.Context, form *platformclientv2.Surveyform) (*platformclientv2.Surveyform, *platformclientv2.APIResponse, error) {
	return p.createQualityFormsSurveyAttr(ctx, p, form)
}

// getAllQualityFormsSurvey retrieves all Genesys Cloud quality forms surveys
func (p *qualityFormsSurveyProxy) getAllQualityFormsSurvey(ctx context.Context) (*[]platformclientv2.Surveyform, *platformclientv2.APIResponse, error) {
	return p.getAllQualityFormsSurveyAttr(ctx, p)
}

// getQualityFormsSurveyById returns a single Genesys Cloud quality forms survey by Id
func (p *qualityFormsSurveyProxy) getQualityFormsSurveyById(ctx context.Context, id string) (form *platformclientv2.Surveyform, response *platformclientv2.APIResponse, err error) {
	if form := rc.GetCacheItem(p.formsCache, id); form != nil {
		return form, nil, nil
	}
	return p.getQualityFormsSurveyByIdAttr(ctx, p, id)
}

// getQualityFormsSurveyByName returns a single Genesys Cloud quality forms survey by Name
func (p *qualityFormsSurveyProxy) getQualityFormsSurveyByName(ctx context.Context, name string) (id string, retryable bool, response *platformclientv2.APIResponse, err error) {
	return p.getQualityFormsSurveyByNameAttr(ctx, p, id)
}

// updateQualityFormsSurvey updates a Genesys Cloud quality forms survey
func (p *qualityFormsSurveyProxy) updateQualityFormsSurvey(ctx context.Context, id string, form *platformclientv2.Surveyform) (*platformclientv2.Surveyform, *platformclientv2.APIResponse, error) {
	return p.updateQualityFormsSurveyAttr(ctx, p, id, form)
}

// deleteQualityFormsSurvey deletes a Genesys Cloud quality forms survey by Id
func (p *qualityFormsSurveyProxy) deleteQualityFormsSurvey(ctx context.Context, id string) (*platformclientv2.APIResponse, error) {
	return p.deleteQualityFormsSurveyAttr(ctx, p, id)
}

// publishQualityFormsSurvey publishes or unpublishes a Genesys Cloud quality forms survey
func (p *qualityFormsSurveyProxy) publishQualityFormsSurvey(ctx context.Context, id string, published bool) (*platformclientv2.APIResponse, error) {
	return p.publishQualityFormsSurveyAttr(ctx, p, id, published)
}

func (p *qualityFormsSurveyProxy) getQualityFormsSurveyVersions(ctx context.Context, formId string, pageSize int, pageNumber int) (*platformclientv2.Surveyformentitylisting, *platformclientv2.APIResponse, error) {
	return p.getQualityFormsSurveyVersionsAttr(ctx, p, formId, pageSize, pageNumber)
}

func (p *qualityFormsSurveyProxy) patchQualityFormsSurvey(ctx context.Context, formId string, body platformclientv2.Surveyform) (*platformclientv2.Surveyform, *platformclientv2.APIResponse, error) {
	return p.patchQualityFormsSurveyAttr(ctx, p, formId, body)
}

// Implementation functions

func createQualityFormsSurveyFn(ctx context.Context, p *qualityFormsSurveyProxy, form *platformclientv2.Surveyform) (*platformclientv2.Surveyform, *platformclientv2.APIResponse, error) {
	surveyForm, resp, err := p.qualityApi.PostQualityFormsSurveys(*form)
	if err != nil {
		return nil, resp, fmt.Errorf("Failed to create quality forms survey: %s", err)
	}
	return surveyForm, resp, nil
}

func getAllQualityFormsSurveyFn(ctx context.Context, p *qualityFormsSurveyProxy) (*[]platformclientv2.Surveyform, *platformclientv2.APIResponse, error) {
	var allForms []platformclientv2.Surveyform
	const pageSize = 100

	forms, resp, err := p.qualityApi.GetQualityFormsSurveys(pageSize, 1, "", "", "", "publishHistory", "", "")
	if err != nil {
		return nil, resp, fmt.Errorf("Failed to get quality forms surveys: %s", err)
	}

	if forms.Entities == nil || len(*forms.Entities) == 0 {
		return &allForms, resp, nil
	}

	allForms = append(allForms, *forms.Entities...)

	for pageNum := 2; ; pageNum++ {
		forms, resp, err := p.qualityApi.GetQualityFormsSurveys(pageSize, pageNum, "", "", "", "publishHistory", "", "")
		if err != nil {
			return nil, resp, fmt.Errorf("Failed to get quality forms surveys: %s", err)
		}

		if forms.Entities == nil || len(*forms.Entities) == 0 {
			break
		}

		allForms = append(allForms, *forms.Entities...)
	}

	// Cache the forms for later use
	for _, form := range allForms {
		rc.SetCache(p.formsCache, *form.Id, form)
	}

	return &allForms, resp, nil
}

func getQualityFormsSurveyByIdFn(ctx context.Context, p *qualityFormsSurveyProxy, id string) (form *platformclientv2.Surveyform, response *platformclientv2.APIResponse, err error) {
	form, resp, err := p.qualityApi.GetQualityFormsSurvey(id)
	if err != nil {
		return nil, resp, fmt.Errorf("Failed to retrieve quality forms survey by id %s: %s", id, err)
	}
	return form, resp, nil
}

func getQualityFormsSurveyByNameFn(ctx context.Context, p *qualityFormsSurveyProxy, name string) (id string, retryable bool, response *platformclientv2.APIResponse, err error) {
	forms, resp, err := p.getAllQualityFormsSurvey(ctx)
	if err != nil {
		return "", false, resp, err
	}

	for _, form := range *forms {
		if *form.Name == name {
			return *form.Id, false, resp, nil
		}
	}
	return "", true, resp, fmt.Errorf("Unable to locate quality forms survey with name %s", name)
}

func updateQualityFormsSurveyFn(ctx context.Context, p *qualityFormsSurveyProxy, id string, form *platformclientv2.Surveyform) (*platformclientv2.Surveyform, *platformclientv2.APIResponse, error) {
	formResponse, resp, err := p.qualityApi.PutQualityFormsSurvey(id, *form)
	if err != nil {
		return nil, resp, fmt.Errorf("Failed to update quality forms survey: %s", err)
	}
	return formResponse, resp, nil
}

func deleteQualityFormsSurveyFn(ctx context.Context, p *qualityFormsSurveyProxy, id string) (*platformclientv2.APIResponse, error) {
	resp, err := p.qualityApi.DeleteQualityFormsSurvey(id)
	if err != nil {
		return resp, fmt.Errorf("Failed to delete quality forms survey: %s", err)
	}
	return resp, nil
}

func publishQualityFormsSurveyFn(ctx context.Context, p *qualityFormsSurveyProxy, id string, published bool) (*platformclientv2.APIResponse, error) {
	_, resp, err := p.qualityApi.PostQualityPublishedformsSurveys(platformclientv2.Publishform{
		Id:        &id,
		Published: &published,
	})
	if err != nil {
		return resp, fmt.Errorf("Failed to publish quality forms survey: %s", err)
	}
	return resp, nil
}

func getQualityFormsSurveyVersionsFn(ctx context.Context, p *qualityFormsSurveyProxy, formId string, pageSize int, pageNumber int) (*platformclientv2.Surveyformentitylisting, *platformclientv2.APIResponse, error) {
	formVersions, resp, err := p.qualityApi.GetQualityFormsSurveyVersions(formId, 25, 1)
	if err != nil {
		return nil, resp, fmt.Errorf("Failed to get quality forms survey versions: %s", err)
	}
	return formVersions, resp, nil
}

func patchQualityFormsSurveyFn(ctx context.Context, p *qualityFormsSurveyProxy, formId string, body platformclientv2.Surveyform) (*platformclientv2.Surveyform, *platformclientv2.APIResponse, error) {
	formResponse, resp, err := p.qualityApi.PatchQualityFormsSurvey(formId, body)
	if err != nil {
		return nil, resp, fmt.Errorf("Failed to patch quality forms survey: %s", err)
	}
	return formResponse, resp, nil
}
