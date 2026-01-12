package dictionary_feedback

import (
	"context"
	"fmt"
	"log"

	rc "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_cache"

	"github.com/mypurecloud/platform-client-sdk-go/v176/platformclientv2"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
)

/*
The genesyscloud_dictionary_feedback_proxy.go file contains the proxy structures and methods that interact
with the Genesys Cloud SDK. We use composition here for each function on the proxy so individual functions can be stubbed
out during testing.
*/

// internalProxy holds a proxy instance that can be used throughout the package
var internalProxy *dictionaryFeedbackProxy

// Type definitions for each func on our proxy so we can easily mock them out later
type (
	createDictionaryFeedbackFunc      func(ctx context.Context, p *dictionaryFeedbackProxy, dictionaryFeedback *platformclientv2.Dictionaryfeedback) (*platformclientv2.Dictionaryfeedback, *platformclientv2.APIResponse, error)
	getAllDictionaryFeedbackFunc      func(ctx context.Context, p *dictionaryFeedbackProxy) (*[]platformclientv2.Dictionaryfeedback, *platformclientv2.APIResponse, error)
	getDictionaryFeedbackIdByNameFunc func(ctx context.Context, p *dictionaryFeedbackProxy, name string) (string, *platformclientv2.APIResponse, bool, error)
	getDictionaryFeedbackByIdFunc     func(ctx context.Context, p *dictionaryFeedbackProxy, id string) (*platformclientv2.Dictionaryfeedback, *platformclientv2.APIResponse, error)
	updateDictionaryFeedbackFunc      func(ctx context.Context, p *dictionaryFeedbackProxy, id string, dictionaryFeedback *platformclientv2.Dictionaryfeedback) (*platformclientv2.Dictionaryfeedback, *platformclientv2.APIResponse, error)
	deleteDictionaryFeedbackFunc      func(ctx context.Context, p *dictionaryFeedbackProxy, id string) (*platformclientv2.APIResponse, error)
)

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
	dictionaryFeedbackCache           rc.CacheInterface[platformclientv2.Dictionaryfeedback]
}

// newDictionaryFeedbackProxy initializes the dictionary feedback proxy with all of the data needed to communicate with Genesys Cloud
func newDictionaryFeedbackProxy(clientConfig *platformclientv2.Configuration) *dictionaryFeedbackProxy {
	api := platformclientv2.NewSpeechTextAnalyticsApiWithConfig(clientConfig)
	dictionaryFeedbackCache := rc.NewResourceCache[platformclientv2.Dictionaryfeedback]()
	return &dictionaryFeedbackProxy{
		clientConfig:                      clientConfig,
		speechTextAnalyticsApi:            api,
		dictionaryFeedbackCache:           dictionaryFeedbackCache,
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
	if dictionaryFeedback := rc.GetCacheItem(p.dictionaryFeedbackCache, id); dictionaryFeedback != nil { // GET the dictionaryFeedback from the cache, if not found then call the API
		return dictionaryFeedback, nil, nil
	}
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
	var (
		nextPage               string
		err                    error
		allDictionaryFeedbacks []platformclientv2.Dictionaryfeedback
	)
	const pageSize = 100

	for {
		dictionaryFeedbacks, resp, getErr := p.speechTextAnalyticsApi.GetSpeechandtextanalyticsDictionaryfeedback("", nextPage, pageSize)

		if getErr != nil {
			return nil, resp, getErr
		}

		if dictionaryFeedbacks.Entities == nil || len(*dictionaryFeedbacks.Entities) == 0 {
			break
		}

		// Need to GET each dictionary item as full response from GET API doesn't include full object
		for _, dictionaryFeedback := range *dictionaryFeedbacks.Entities {
			feedback, resp, err := p.speechTextAnalyticsApi.GetSpeechandtextanalyticsDictionaryfeedbackDictionaryFeedbackId(*dictionaryFeedback.Id)
			if err != nil {
				return nil, resp, err
			}
			allDictionaryFeedbacks = append(allDictionaryFeedbacks, *feedback)
			rc.SetCache(p.dictionaryFeedbackCache, *dictionaryFeedback.Id, *feedback)
		}

		if dictionaryFeedbacks.NextUri == nil || *dictionaryFeedbacks.NextUri == "" {
			break
		}

		previousNextPage := nextPage
		nextPage, err = util.GetQueryParamValueFromUri(*dictionaryFeedbacks.NextUri, "nextPage")
		if err != nil {
			return nil, resp, err
		}
		if nextPage == "" || nextPage == previousNextPage {
			break
		}
	}

	return &allDictionaryFeedbacks, nil, nil
}

// getDictionaryFeedbackIdByNameFn is an implementation of the function to get a Genesys Cloud dictionary feedback by name
func getDictionaryFeedbackIdByNameFn(ctx context.Context, p *dictionaryFeedbackProxy, name string) (string, *platformclientv2.APIResponse, bool, error) {
	// As there is no API to GET based on "name" used cache to get term and if not in cache then getAll
	dictionaryFeedbacks := rc.GetCache(p.dictionaryFeedbackCache)
	if dictionaryFeedbacks != nil {
		for _, dictionaryFeedback := range *dictionaryFeedbacks {
			if *dictionaryFeedback.Term == name {
				log.Printf("Retrieved the dictionary feedback id %s by name %s from cache", *dictionaryFeedback.Id, name)
				return *dictionaryFeedback.Id, nil, false, nil
			}
		}
	}

	dictionaryFeedbacksReq, resp, err := p.getAllDictionaryFeedback(ctx)
	if err != nil {
		return "", resp, false, err
	}

	for _, dictionaryFeedbackGet := range *dictionaryFeedbacksReq {
		if *dictionaryFeedbackGet.Term == name {
			log.Printf("Retrieved the dictionary feedback id %s by name %s", *dictionaryFeedbackGet.Id, name)
			return *dictionaryFeedbackGet.Id, nil, false, nil
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
