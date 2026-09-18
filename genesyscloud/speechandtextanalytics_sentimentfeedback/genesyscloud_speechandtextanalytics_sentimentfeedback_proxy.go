package speechandtextanalytics_sentimentfeedback

import (
	"context"
	"fmt"
	"log"
	"net/http"

	rc "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_cache"

	"github.com/mypurecloud/platform-client-sdk-go/v195/platformclientv2"
)

/*
The genesyscloud_speechandtextanalytics_sentimentfeedback_proxy.go file contains the proxy structures and methods that interact
with the Genesys Cloud SDK. We use composition here for each function on the proxy so individual functions can be stubbed
out during testing.

The sentiment feedback API only supports POST (create), GET (list), and DELETE (by id). There is no GET-by-id and no
update endpoint, so reads are performed by listing all sentiment feedback and matching on id.
*/

// internalProxy holds a proxy instance that can be used throughout the package
var internalProxy *sentimentFeedbackProxy

var sentimentFeedbackCache = rc.NewResourceCache[platformclientv2.Sentimentfeedback]()

// Type definitions for each func on our proxy so we can easily mock them out later
type (
	createSentimentFeedbackFunc        func(ctx context.Context, p *sentimentFeedbackProxy, sentimentFeedback *platformclientv2.Sentimentfeedback) (*platformclientv2.Sentimentfeedback, *platformclientv2.APIResponse, error)
	getAllSentimentFeedbackFunc        func(ctx context.Context, p *sentimentFeedbackProxy) (*[]platformclientv2.Sentimentfeedback, *platformclientv2.APIResponse, error)
	getSentimentFeedbackByIdFunc       func(ctx context.Context, p *sentimentFeedbackProxy, id string) (*platformclientv2.Sentimentfeedback, *platformclientv2.APIResponse, error)
	getSentimentFeedbackIdByPhraseFunc func(ctx context.Context, p *sentimentFeedbackProxy, phrase string, dialect string) (string, *platformclientv2.APIResponse, bool, error)
	deleteSentimentFeedbackFunc        func(ctx context.Context, p *sentimentFeedbackProxy, id string) (*platformclientv2.APIResponse, error)
)

// sentimentFeedbackProxy contains all of the methods that call genesys cloud APIs.
type sentimentFeedbackProxy struct {
	clientConfig                       *platformclientv2.Configuration
	speechTextAnalyticsApi             *platformclientv2.SpeechTextAnalyticsApi
	createSentimentFeedbackAttr        createSentimentFeedbackFunc
	getAllSentimentFeedbackAttr        getAllSentimentFeedbackFunc
	getSentimentFeedbackByIdAttr       getSentimentFeedbackByIdFunc
	getSentimentFeedbackIdByPhraseAttr getSentimentFeedbackIdByPhraseFunc
	deleteSentimentFeedbackAttr        deleteSentimentFeedbackFunc
	sentimentFeedbackCache             rc.CacheInterface[platformclientv2.Sentimentfeedback]
}

// newSentimentFeedbackProxy initializes the sentiment feedback proxy with all of the data needed to communicate with Genesys Cloud
func newSentimentFeedbackProxy(clientConfig *platformclientv2.Configuration) *sentimentFeedbackProxy {
	api := platformclientv2.NewSpeechTextAnalyticsApiWithConfig(clientConfig)

	return &sentimentFeedbackProxy{
		clientConfig:                       clientConfig,
		speechTextAnalyticsApi:             api,
		sentimentFeedbackCache:             sentimentFeedbackCache,
		createSentimentFeedbackAttr:        createSentimentFeedbackFn,
		getAllSentimentFeedbackAttr:        getAllSentimentFeedbackFn,
		getSentimentFeedbackByIdAttr:       getSentimentFeedbackByIdFn,
		getSentimentFeedbackIdByPhraseAttr: getSentimentFeedbackIdByPhraseFn,
		deleteSentimentFeedbackAttr:        deleteSentimentFeedbackFn,
	}
}

// getSentimentFeedbackProxy acts as a singleton for the internalProxy. It also ensures
// that we can still proxy our tests by directly setting internalProxy package variable
func getSentimentFeedbackProxy(clientConfig *platformclientv2.Configuration) *sentimentFeedbackProxy {
	if internalProxy == nil {
		internalProxy = newSentimentFeedbackProxy(clientConfig)
	}

	return internalProxy
}

// createSentimentFeedback creates a Genesys Cloud sentiment feedback
func (p *sentimentFeedbackProxy) createSentimentFeedback(ctx context.Context, sentimentFeedback *platformclientv2.Sentimentfeedback) (*platformclientv2.Sentimentfeedback, *platformclientv2.APIResponse, error) {
	return p.createSentimentFeedbackAttr(ctx, p, sentimentFeedback)
}

// getAllSentimentFeedback retrieves all Genesys Cloud sentiment feedback
func (p *sentimentFeedbackProxy) getAllSentimentFeedback(ctx context.Context) (*[]platformclientv2.Sentimentfeedback, *platformclientv2.APIResponse, error) {
	return p.getAllSentimentFeedbackAttr(ctx, p)
}

// getSentimentFeedbackById returns a single Genesys Cloud sentiment feedback by Id
func (p *sentimentFeedbackProxy) getSentimentFeedbackById(ctx context.Context, id string) (*platformclientv2.Sentimentfeedback, *platformclientv2.APIResponse, error) {
	if sentimentFeedback := rc.GetCacheItem(p.sentimentFeedbackCache, id); sentimentFeedback != nil { // GET the sentimentFeedback from the cache, if not found then call the API
		return sentimentFeedback, nil, nil
	}
	return p.getSentimentFeedbackByIdAttr(ctx, p, id)
}

// getSentimentFeedbackIdByPhrase returns a single Genesys Cloud sentiment feedback id by phrase and dialect
func (p *sentimentFeedbackProxy) getSentimentFeedbackIdByPhrase(ctx context.Context, phrase string, dialect string) (string, *platformclientv2.APIResponse, bool, error) {
	return p.getSentimentFeedbackIdByPhraseAttr(ctx, p, phrase, dialect)
}

// deleteSentimentFeedback deletes a Genesys Cloud sentiment feedback by Id
func (p *sentimentFeedbackProxy) deleteSentimentFeedback(ctx context.Context, id string) (*platformclientv2.APIResponse, error) {
	return p.deleteSentimentFeedbackAttr(ctx, p, id)
}

// createSentimentFeedbackFn is an implementation function for creating a Genesys Cloud sentiment feedback
func createSentimentFeedbackFn(ctx context.Context, p *sentimentFeedbackProxy, sentimentFeedback *platformclientv2.Sentimentfeedback) (*platformclientv2.Sentimentfeedback, *platformclientv2.APIResponse, error) {
	return p.speechTextAnalyticsApi.PostSpeechandtextanalyticsSentimentfeedback(*sentimentFeedback)
}

// getAllSentimentFeedbackFn is the implementation for retrieving all sentiment feedback in Genesys Cloud
func getAllSentimentFeedbackFn(ctx context.Context, p *sentimentFeedbackProxy) (*[]platformclientv2.Sentimentfeedback, *platformclientv2.APIResponse, error) {
	// An empty dialect filter returns the full list of sentiment feedback for the org.
	entityListing, resp, err := p.speechTextAnalyticsApi.GetSpeechandtextanalyticsSentimentfeedback("")
	if err != nil {
		return nil, resp, err
	}

	allSentimentFeedbacks := make([]platformclientv2.Sentimentfeedback, 0)
	if entityListing.Entities == nil {
		return &allSentimentFeedbacks, resp, nil
	}

	for _, sentimentFeedback := range *entityListing.Entities {
		if sentimentFeedback.Id == nil {
			continue
		}
		allSentimentFeedbacks = append(allSentimentFeedbacks, sentimentFeedback)
		rc.SetCache(p.sentimentFeedbackCache, *sentimentFeedback.Id, sentimentFeedback)
	}

	return &allSentimentFeedbacks, resp, nil
}

// getSentimentFeedbackByIdFn retrieves a single sentiment feedback by Id. As the API does not expose a
// GET-by-id endpoint, the full list is retrieved and matched on id.
func getSentimentFeedbackByIdFn(ctx context.Context, p *sentimentFeedbackProxy, id string) (*platformclientv2.Sentimentfeedback, *platformclientv2.APIResponse, error) {
	sentimentFeedbacks, resp, err := p.getAllSentimentFeedback(ctx)
	if err != nil {
		return nil, resp, err
	}

	for _, sentimentFeedback := range *sentimentFeedbacks {
		if sentimentFeedback.Id != nil && *sentimentFeedback.Id == id {
			feedback := sentimentFeedback
			return &feedback, resp, nil
		}
	}

	// The API has no GET-by-id endpoint, so a missing entry in the list means the resource no longer exists.
	// Return a synthetic 404 response (without mutating the real list response) so callers can treat it as removed.
	// The "API Error: 404" marker in the message matches the SDK's own 404 format, which the shared read-retry
	// helper (util.WithRetriesForRead) relies on to remove the resource from state once the consistency window ends.
	notFoundResp := &platformclientv2.APIResponse{StatusCode: http.StatusNotFound}
	return nil, notFoundResp, fmt.Errorf("API Error: 404 - unable to find sentiment feedback with id %s", id)
}

// getSentimentFeedbackIdByPhraseFn retrieves a single sentiment feedback id by phrase and dialect
func getSentimentFeedbackIdByPhraseFn(ctx context.Context, p *sentimentFeedbackProxy, phrase string, dialect string) (string, *platformclientv2.APIResponse, bool, error) {
	sentimentFeedbacks, resp, err := p.getAllSentimentFeedback(ctx)
	if err != nil {
		return "", resp, false, err
	}

	for _, sentimentFeedback := range *sentimentFeedbacks {
		if sentimentFeedback.Id != nil && sentimentFeedbackMatchesFilters(sentimentFeedback, phrase, dialect) {
			log.Printf("Retrieved the sentiment feedback id %s by phrase %s", *sentimentFeedback.Id, phrase)
			return *sentimentFeedback.Id, resp, false, nil
		}
	}

	return "", resp, true, fmt.Errorf("unable to find sentiment feedback with phrase %s", phrase)
}

// deleteSentimentFeedbackFn is an implementation function for deleting a Genesys Cloud sentiment feedback
func deleteSentimentFeedbackFn(ctx context.Context, p *sentimentFeedbackProxy, id string) (*platformclientv2.APIResponse, error) {
	resp, err := p.speechTextAnalyticsApi.DeleteSpeechandtextanalyticsSentimentfeedbackSentimentFeedbackId(id)
	if err != nil {
		return resp, err
	}
	rc.DeleteCacheItem(p.sentimentFeedbackCache, id)
	return resp, nil
}
