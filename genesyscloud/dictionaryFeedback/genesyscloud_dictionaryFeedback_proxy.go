package dictionary_feedback

import (
	"context"
	"fmt"
	"github.com/mypurecloud/platform-client-sdk-go/v176/platformclientv2"
	"log"
)

/*
The genesyscloud_dictionary_feedback_proxy.go file contains the proxy structures and methods that interact
with the Genesys Cloud SDK. We use composition here for each function on the proxy so individual functions can be stubbed
out during testing.
*/

// internalProxy holds a proxy instance that can be used throughout the package
var internalProxy *dictionaryFeedbackProxy

// Type definitions for each func on our proxy so we can easily mock them out later
type createDictionaryFeedbackFunc func(ctx context.Context, p *dictionaryFeedbackProxy, dictionaryFeedback *platformclientv2.Dictionaryfeedback) (*platformclientv2.Dictionaryfeedback, *platformclientv2.APIResponse, error)
type getAllDictionaryFeedbackFunc func(ctx context.Context, p *dictionaryFeedbackProxy) (*[]platformclientv2.Dictionaryfeedback, *platformclientv2.APIResponse, error)
type getDictionaryFeedbackIdByNameFunc func(ctx context.Context, p *dictionaryFeedbackProxy, name string) (string, *platformclientv2.APIResponse, bool, error)
type getDictionaryFeedbackByIdFunc func(ctx context.Context, p *dictionaryFeedbackProxy, id string) (*platformclientv2.Dictionaryfeedback, *platformclientv2.APIResponse, error)
type updateDictionaryFeedbackFunc func(ctx context.Context, p *dictionaryFeedbackProxy, id string, dictionaryFeedback *platformclientv2.Dictionaryfeedback) (*platformclientv2.Dictionaryfeedback, *platformclientv2.APIResponse, error)
type deleteDictionaryFeedbackFunc func(ctx context.Context, p *dictionaryFeedbackProxy, id string) (*platformclientv2.APIResponse, error)

// dictionaryFeedbackProxy contains all of the methods that call genesys cloud APIs.
type dictionaryFeedbackProxy struct {
	clientConfig                      *platformclientv2.Configuration
	speechTextAnalyticsApi            *platformclientv2.SpeechTextAnalyticsApi
	createDictionaryFeedbackAttr      createDictionaryFeedbackFunc
	getAllDictionaryFeedbackAttr      getAllDictionaryFeedbackFunc
	getDictionaryFeedbackIdByNameAttr getDictionaryFeedbackIdByNameFunc
	getDictionaryFeedbackByIdAttr     getDictionaryFeedbackByIdFunc
	updateDictionaryFeedbackAttr      updateDictionaryFeedbackFunc
	deleteDictionaryFeedbackAttr      deleteDictionaryFeedbackFunc
}

// newDictionaryFeedbackProxy initializes the dictionary feedback proxy with all of the data needed to communicate with Genesys Cloud
func newDictionaryFeedbackProxy(clientConfig *platformclientv2.Configuration) *dictionaryFeedbackProxy {
	api := platformclientv2.NewSpeechTextAnalyticsApiWithConfig(clientConfig)
	return &dictionaryFeedbackProxy{
		clientConfig:                      clientConfig,
		speechTextAnalyticsApi:            api,
		createDictionaryFeedbackAttr:      createDictionaryFeedbackFn,
		getAllDictionaryFeedbackAttr:      getAllDictionaryFeedbackFn,
		getDictionaryFeedbackIdByNameAttr: getDictionaryFeedbackIdByNameFn,
		getDictionaryFeedbackByIdAttr:     getDictionaryFeedbackByIdFn,
		updateDictionaryFeedbackAttr:      updateDictionaryFeedbackFn,
		deleteDictionaryFeedbackAttr:      deleteDictionaryFeedbackFn,
	}
}

// getDictionaryFeedbackProxy acts as a singleton to for the internalProxy.  It also ensures
// that we can still proxy our tests by directly setting internalProxy package variable
func getDictionaryFeedbackProxy(clientConfig *platformclientv2.Configuration) *dictionaryFeedbackProxy {
	if internalProxy == nil {
		internalProxy = newDictionaryFeedbackProxy(clientConfig)
	}

	return internalProxy
}

// createDictionaryFeedback creates a Genesys Cloud dictionary feedback
func (p *dictionaryFeedbackProxy) createDictionaryFeedback(ctx context.Context, dictionaryFeedback *platformclientv2.Dictionaryfeedback) (*platformclientv2.Dictionaryfeedback, *platformclientv2.APIResponse, error) {
	return p.createDictionaryFeedbackAttr(ctx, p, dictionaryFeedback)
}

// getDictionaryFeedback retrieves all Genesys Cloud dictionary feedback
func (p *dictionaryFeedbackProxy) getAllDictionaryFeedback(ctx context.Context) (*[]platformclientv2.Dictionaryfeedback, *platformclientv2.APIResponse, error) {
	return p.getAllDictionaryFeedbackAttr(ctx, p)
}

// getDictionaryFeedbackIdByName returns a single Genesys Cloud dictionary feedback by a name
func (p *dictionaryFeedbackProxy) getDictionaryFeedbackIdByName(ctx context.Context, name string) (string, *platformclientv2.APIResponse, bool, error) {
	return p.getDictionaryFeedbackIdByNameAttr(ctx, p, name)
}

// getDictionaryFeedbackById returns a single Genesys Cloud dictionary feedback by Id
func (p *dictionaryFeedbackProxy) getDictionaryFeedbackById(ctx context.Context, id string) (*platformclientv2.Dictionaryfeedback, *platformclientv2.APIResponse, error) {
	return p.getDictionaryFeedbackByIdAttr(ctx, p, id)
}

// updateDictionaryFeedback updates a Genesys Cloud dictionary feedback
func (p *dictionaryFeedbackProxy) updateDictionaryFeedback(ctx context.Context, id string, dictionaryFeedback *platformclientv2.Dictionaryfeedback) (*platformclientv2.Dictionaryfeedback, *platformclientv2.APIResponse, error) {
	return p.updateDictionaryFeedbackAttr(ctx, p, id, dictionaryFeedback)
}

// deleteDictionaryFeedback deletes a Genesys Cloud dictionary feedback by Id
func (p *dictionaryFeedbackProxy) deleteDictionaryFeedback(ctx context.Context, id string) (*platformclientv2.APIResponse, error) {
	return p.deleteDictionaryFeedbackAttr(ctx, p, id)
}

// createDictionaryFeedbackFn is an implementation function for creating a Genesys Cloud dictionary feedback
func createDictionaryFeedbackFn(ctx context.Context, p *dictionaryFeedbackProxy, dictionaryFeedback *platformclientv2.Dictionaryfeedback) (*platformclientv2.Dictionaryfeedback, *platformclientv2.APIResponse, error) {
	return p.speechTextAnalyticsApi.PostSpeechandtextanalyticsDictionaryfeedback(*dictionaryFeedback)
}

// getAllDictionaryFeedbackFn is the implementation for retrieving all dictionary feedback in Genesys Cloud
func getAllDictionaryFeedbackFn(ctx context.Context, p *dictionaryFeedbackProxy) (*[]platformclientv2.Dictionaryfeedback, *platformclientv2.APIResponse, error) {
	var allDictionaryFeedbacks []platformclientv2.Dictionaryfeedback
	const pageSize = 100

	dictionaryFeedbacks, resp, err := p.speechTextAnalyticsApi.GetSpeechandtextanalyticsDictionaryfeedback()
	if err != nil {
		return nil, resp, err
	}
	if dictionaryFeedbacks.Entities == nil || len(*dictionaryFeedbacks.Entities) == 0 {
		return &allDictionaryFeedbacks, resp, nil
	}
	for _, dictionaryFeedback := range *dictionaryFeedbacks.Entities {
		allDictionaryFeedbacks = append(allDictionaryFeedbacks, dictionaryFeedback)
	}

	for pageNum := 2; pageNum <= *dictionaryFeedbacks.PageCount; pageNum++ {
		dictionaryFeedbacks, _, err := p.speechTextAnalyticsApi.GetSpeechandtextanalyticsDictionaryfeedback()
		if err != nil {
			return nil, resp, err
		}

		if dictionaryFeedbacks.Entities == nil || len(*dictionaryFeedbacks.Entities) == 0 {
			break
		}

		for _, dictionaryFeedback := range *dictionaryFeedbacks.Entities {
			allDictionaryFeedbacks = append(allDictionaryFeedbacks, dictionaryFeedback)
		}
	}

	return &allDictionaryFeedbacks, resp, nil
}

// getDictionaryFeedbackIdByNameFn is an implementation of the function to get a Genesys Cloud dictionary feedback by name
func getDictionaryFeedbackIdByNameFn(ctx context.Context, p *dictionaryFeedbackProxy, name string) (string, *platformclientv2.APIResponse, bool, error) {
	dictionaryFeedbacks, resp, err := p.speechTextAnalyticsApi.GetSpeechandtextanalyticsDictionaryfeedback()
	if err != nil {
		return "", resp, false, err
	}

	if dictionaryFeedbacks.Entities == nil || len(*dictionaryFeedbacks.Entities) == 0 {
		return "", resp, true, err
	}

	for _, dictionaryFeedback := range *dictionaryFeedbacks.Entities {
		if *dictionaryFeedback.Name == name {
			log.Printf("Retrieved the dictionary feedback id %s by name %s", *dictionaryFeedback.Id, name)
			return *dictionaryFeedback.Id, resp, false, nil
		}
	}

	return "", resp, true, fmt.Errorf("Unable to find dictionary feedback with name %s", name)
}

// getDictionaryFeedbackByIdFn is an implementation of the function to get a Genesys Cloud dictionary feedback by Id
func getDictionaryFeedbackByIdFn(ctx context.Context, p *dictionaryFeedbackProxy, id string) (*platformclientv2.Dictionaryfeedback, *platformclientv2.APIResponse, error) {
	return p.speechTextAnalyticsApi.GetSpeechandtextanalyticsDictionaryfeedbackDictionaryFeedbackId(id)
}

// updateDictionaryFeedbackFn is an implementation of the function to update a Genesys Cloud dictionary feedback
func updateDictionaryFeedbackFn(ctx context.Context, p *dictionaryFeedbackProxy, id string, dictionaryFeedback *platformclientv2.Dictionaryfeedback) (*platformclientv2.Dictionaryfeedback, *platformclientv2.APIResponse, error) {
	return p.speechTextAnalyticsApi.PutSpeechandtextanalyticsDictionaryfeedbackDictionaryFeedbackId(id, *dictionaryFeedback)
}

// deleteDictionaryFeedbackFn is an implementation function for deleting a Genesys Cloud dictionary feedback
func deleteDictionaryFeedbackFn(ctx context.Context, p *dictionaryFeedbackProxy, id string) (*platformclientv2.APIResponse, error) {
	return p.speechTextAnalyticsApi.DeleteSpeechandtextanalyticsDictionaryfeedbackDictionaryFeedbackId(id)
}
