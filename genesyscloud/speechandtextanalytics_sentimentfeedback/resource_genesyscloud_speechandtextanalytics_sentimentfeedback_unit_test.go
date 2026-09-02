package speechandtextanalytics_sentimentfeedback

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v195/platformclientv2"
	"github.com/stretchr/testify/assert"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
)

/*
The resource_genesyscloud_speechandtextanalytics_sentimentfeedback_unit_test.go file contains unit tests that stub the
proxy func pointers so no live API is required.
*/

const (
	testSentimentFeedbackId = "test-sentiment-feedback-id"
	testPhrase              = "the service was great"
	testDialect             = "en-US"
	testFeedbackValue       = FeedbackValuePositive
	testCreatedById         = "test-user-id"
)

func buildTestSentimentFeedback() platformclientv2.Sentimentfeedback {
	dateCreated := time.Date(2025, 1, 2, 3, 4, 5, 0, time.UTC)
	return platformclientv2.Sentimentfeedback{
		Id:            platformclientv2.String(testSentimentFeedbackId),
		Phrase:        platformclientv2.String(testPhrase),
		Dialect:       platformclientv2.String(testDialect),
		FeedbackValue: platformclientv2.String(testFeedbackValue),
		DateCreated:   &dateCreated,
		CreatedBy:     &platformclientv2.Addressableentityref{Id: platformclientv2.String(testCreatedById)},
	}
}

func buildTestSentimentFeedbackDataMap() map[string]interface{} {
	return map[string]interface{}{
		"phrase":         testPhrase,
		"dialect":        testDialect,
		"feedback_value": testFeedbackValue,
	}
}

func TestUnitResourceSentimentFeedbackCreate(t *testing.T) {
	var capturedCreateBody *platformclientv2.Sentimentfeedback
	testFeedback := buildTestSentimentFeedback()

	proxy := &sentimentFeedbackProxy{}
	proxy.createSentimentFeedbackAttr = func(ctx context.Context, p *sentimentFeedbackProxy, sentimentFeedback *platformclientv2.Sentimentfeedback) (*platformclientv2.Sentimentfeedback, *platformclientv2.APIResponse, error) {
		capturedCreateBody = sentimentFeedback
		created := testFeedback
		return &created, &platformclientv2.APIResponse{StatusCode: http.StatusOK}, nil
	}
	proxy.getSentimentFeedbackByIdAttr = func(ctx context.Context, p *sentimentFeedbackProxy, id string) (*platformclientv2.Sentimentfeedback, *platformclientv2.APIResponse, error) {
		found := testFeedback
		return &found, &platformclientv2.APIResponse{StatusCode: http.StatusOK}, nil
	}

	internalProxy = proxy
	defer func() { internalProxy = nil }()

	d := schema.TestResourceDataRaw(t, ResourceSentimentFeedback().Schema, buildTestSentimentFeedbackDataMap())
	gcloud := &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}

	diagErr := createSentimentFeedback(context.Background(), d, gcloud)
	assert.Equal(t, false, diagErr.HasError())
	assert.Equal(t, testSentimentFeedbackId, d.Id())

	// Read-only fields must never be sent on the create body
	assert.NotNil(t, capturedCreateBody)
	assert.Nil(t, capturedCreateBody.Id)
	assert.Nil(t, capturedCreateBody.DateCreated)
	assert.Nil(t, capturedCreateBody.CreatedBy)
	// Writable fields must be sent
	assert.Equal(t, testPhrase, *capturedCreateBody.Phrase)
	assert.Equal(t, testDialect, *capturedCreateBody.Dialect)
	assert.Equal(t, testFeedbackValue, *capturedCreateBody.FeedbackValue)
}

func TestUnitResourceSentimentFeedbackRead(t *testing.T) {
	testFeedback := buildTestSentimentFeedback()

	proxy := &sentimentFeedbackProxy{}
	proxy.getSentimentFeedbackByIdAttr = func(ctx context.Context, p *sentimentFeedbackProxy, id string) (*platformclientv2.Sentimentfeedback, *platformclientv2.APIResponse, error) {
		assert.Equal(t, testSentimentFeedbackId, id)
		found := testFeedback
		return &found, &platformclientv2.APIResponse{StatusCode: http.StatusOK}, nil
	}

	internalProxy = proxy
	defer func() { internalProxy = nil }()

	d := schema.TestResourceDataRaw(t, ResourceSentimentFeedback().Schema, buildTestSentimentFeedbackDataMap())
	d.SetId(testSentimentFeedbackId)
	gcloud := &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}

	diagErr := readSentimentFeedback(context.Background(), d, gcloud)
	assert.Equal(t, false, diagErr.HasError())
	assert.Equal(t, testPhrase, d.Get("phrase").(string))
	assert.Equal(t, testDialect, d.Get("dialect").(string))
	assert.Equal(t, testFeedbackValue, d.Get("feedback_value").(string))
}

func TestUnitResourceSentimentFeedbackDelete(t *testing.T) {
	var deletedId string
	callCount := 0

	proxy := &sentimentFeedbackProxy{}
	proxy.deleteSentimentFeedbackAttr = func(ctx context.Context, p *sentimentFeedbackProxy, id string) (*platformclientv2.APIResponse, error) {
		deletedId = id
		return &platformclientv2.APIResponse{StatusCode: http.StatusOK}, nil
	}
	proxy.getSentimentFeedbackByIdAttr = func(ctx context.Context, p *sentimentFeedbackProxy, id string) (*platformclientv2.Sentimentfeedback, *platformclientv2.APIResponse, error) {
		callCount++
		return nil, &platformclientv2.APIResponse{StatusCode: http.StatusNotFound}, &notFoundError{}
	}

	internalProxy = proxy
	defer func() { internalProxy = nil }()

	d := schema.TestResourceDataRaw(t, ResourceSentimentFeedback().Schema, buildTestSentimentFeedbackDataMap())
	d.SetId(testSentimentFeedbackId)
	gcloud := &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}

	diagErr := deleteSentimentFeedback(context.Background(), d, gcloud)
	assert.Equal(t, false, diagErr.HasError())
	assert.Equal(t, testSentimentFeedbackId, deletedId)
	assert.GreaterOrEqual(t, callCount, 1)
}

func TestUnitDataSourceSentimentFeedbackRead(t *testing.T) {
	testFeedback := buildTestSentimentFeedback()

	proxy := &sentimentFeedbackProxy{}
	proxy.getSentimentFeedbackIdByPhraseAttr = func(ctx context.Context, p *sentimentFeedbackProxy, phrase string, dialect string) (string, *platformclientv2.APIResponse, bool, error) {
		assert.Equal(t, testPhrase, phrase)
		assert.Equal(t, testDialect, dialect)
		if testFeedback.Id != nil {
			return *testFeedback.Id, &platformclientv2.APIResponse{StatusCode: http.StatusOK}, false, nil
		}
		return "", &platformclientv2.APIResponse{StatusCode: http.StatusNotFound}, true, nil
	}

	internalProxy = proxy
	defer func() { internalProxy = nil }()

	d := schema.TestResourceDataRaw(t, DataSourceSentimentFeedback().Schema, map[string]interface{}{
		"phrase":  testPhrase,
		"dialect": testDialect,
	})
	gcloud := &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}

	diagErr := dataSourceSentimentFeedbackRead(context.Background(), d, gcloud)
	assert.Equal(t, false, diagErr.HasError())
	assert.Equal(t, testSentimentFeedbackId, d.Id())
}

func TestUnitResourceSentimentFeedbackCreateError(t *testing.T) {
	proxy := &sentimentFeedbackProxy{}
	proxy.createSentimentFeedbackAttr = func(ctx context.Context, p *sentimentFeedbackProxy, sentimentFeedback *platformclientv2.Sentimentfeedback) (*platformclientv2.Sentimentfeedback, *platformclientv2.APIResponse, error) {
		return nil, &platformclientv2.APIResponse{StatusCode: http.StatusBadRequest}, fmt.Errorf("bad request")
	}

	internalProxy = proxy
	defer func() { internalProxy = nil }()

	d := schema.TestResourceDataRaw(t, ResourceSentimentFeedback().Schema, buildTestSentimentFeedbackDataMap())
	gcloud := &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}

	diagErr := createSentimentFeedback(context.Background(), d, gcloud)
	assert.Equal(t, true, diagErr.HasError())
	// A failed create must not leave an id on the resource
	assert.Equal(t, "", d.Id())
}

func TestUnitResourceSentimentFeedbackReadNotFound(t *testing.T) {
	// Fail-fast: a retry timeout of 0 makes a 404 on the single read attempt remove the resource
	// from state immediately instead of retrying for the full consistency window.
	t.Setenv("GENESYSCLOUD_CUSTOM_RETRY_TIMEOUT", "0")

	proxy := &sentimentFeedbackProxy{}
	// Simulate the resource having been deleted out-of-band: the list-then-find returns a synthetic 404
	// carrying the same "API Error: 404" marker the real proxy produces.
	proxy.getSentimentFeedbackByIdAttr = func(ctx context.Context, p *sentimentFeedbackProxy, id string) (*platformclientv2.Sentimentfeedback, *platformclientv2.APIResponse, error) {
		return nil, &platformclientv2.APIResponse{StatusCode: http.StatusNotFound}, fmt.Errorf("API Error: 404 - unable to find sentiment feedback with id %s", id)
	}

	internalProxy = proxy
	defer func() { internalProxy = nil }()

	d := schema.TestResourceDataRaw(t, ResourceSentimentFeedback().Schema, buildTestSentimentFeedbackDataMap())
	d.SetId(testSentimentFeedbackId)
	gcloud := &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}

	diagErr := readSentimentFeedback(context.Background(), d, gcloud)
	// A 404 on read should clear the resource from state and return no error
	assert.Equal(t, false, diagErr.HasError())
	assert.Equal(t, "", d.Id())
}

func TestUnitGetSentimentFeedbackByIdFn_Found(t *testing.T) {
	testFeedback := buildTestSentimentFeedback()
	other := platformclientv2.Sentimentfeedback{Id: platformclientv2.String("some-other-id")}

	proxy := &sentimentFeedbackProxy{}
	proxy.getAllSentimentFeedbackAttr = func(ctx context.Context, p *sentimentFeedbackProxy) (*[]platformclientv2.Sentimentfeedback, *platformclientv2.APIResponse, error) {
		return &[]platformclientv2.Sentimentfeedback{other, testFeedback}, &platformclientv2.APIResponse{StatusCode: http.StatusOK}, nil
	}

	found, resp, err := getSentimentFeedbackByIdFn(context.Background(), proxy, testSentimentFeedbackId)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.NotNil(t, found)
	assert.Equal(t, testSentimentFeedbackId, *found.Id)
	assert.Equal(t, testPhrase, *found.Phrase)
}

func TestUnitGetSentimentFeedbackByIdFn_NotFound(t *testing.T) {
	proxy := &sentimentFeedbackProxy{}
	proxy.getAllSentimentFeedbackAttr = func(ctx context.Context, p *sentimentFeedbackProxy) (*[]platformclientv2.Sentimentfeedback, *platformclientv2.APIResponse, error) {
		return &[]platformclientv2.Sentimentfeedback{{Id: platformclientv2.String("some-other-id")}}, &platformclientv2.APIResponse{StatusCode: http.StatusOK}, nil
	}

	found, resp, err := getSentimentFeedbackByIdFn(context.Background(), proxy, testSentimentFeedbackId)
	assert.Error(t, err)
	assert.Nil(t, found)
	// A missing entry in the list must surface as a synthetic 404 so callers treat it as removed
	assert.NotNil(t, resp)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	// The error must carry the "API Error: 404" marker so the shared read-retry helper removes it from state
	assert.Contains(t, err.Error(), "API Error: 404")
}

func TestUnitGetSentimentFeedbackByIdFn_ListError(t *testing.T) {
	proxy := &sentimentFeedbackProxy{}
	proxy.getAllSentimentFeedbackAttr = func(ctx context.Context, p *sentimentFeedbackProxy) (*[]platformclientv2.Sentimentfeedback, *platformclientv2.APIResponse, error) {
		return nil, &platformclientv2.APIResponse{StatusCode: http.StatusInternalServerError}, fmt.Errorf("boom")
	}

	found, resp, err := getSentimentFeedbackByIdFn(context.Background(), proxy, testSentimentFeedbackId)
	assert.Error(t, err)
	assert.Nil(t, found)
	// The underlying list error (not a synthetic 404) must be propagated
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
}

func TestUnitGetSentimentFeedbackIdByPhraseFn(t *testing.T) {
	enUS := platformclientv2.Sentimentfeedback{
		Id:      platformclientv2.String("id-en-us"),
		Phrase:  platformclientv2.String(testPhrase),
		Dialect: platformclientv2.String("en-US"),
	}
	enGB := platformclientv2.Sentimentfeedback{
		Id:      platformclientv2.String("id-en-gb"),
		Phrase:  platformclientv2.String(testPhrase),
		Dialect: platformclientv2.String("en-GB"),
	}

	proxy := &sentimentFeedbackProxy{}
	proxy.getAllSentimentFeedbackAttr = func(ctx context.Context, p *sentimentFeedbackProxy) (*[]platformclientv2.Sentimentfeedback, *platformclientv2.APIResponse, error) {
		return &[]platformclientv2.Sentimentfeedback{enUS, enGB}, &platformclientv2.APIResponse{StatusCode: http.StatusOK}, nil
	}

	// Dialect provided -> disambiguates to the matching entry
	id, _, retryable, err := getSentimentFeedbackIdByPhraseFn(context.Background(), proxy, testPhrase, "en-GB")
	assert.NoError(t, err)
	assert.Equal(t, false, retryable)
	assert.Equal(t, "id-en-gb", id)

	// No dialect -> matches the first entry with the phrase
	id, _, retryable, err = getSentimentFeedbackIdByPhraseFn(context.Background(), proxy, testPhrase, "")
	assert.NoError(t, err)
	assert.Equal(t, false, retryable)
	assert.Equal(t, "id-en-us", id)

	// Phrase not present -> retryable not-found
	id, _, retryable, err = getSentimentFeedbackIdByPhraseFn(context.Background(), proxy, "does not exist", "")
	assert.Error(t, err)
	assert.Equal(t, true, retryable)
	assert.Equal(t, "", id)
}

func TestUnitSentimentFeedbackMatchesFilters(t *testing.T) {
	feedback := platformclientv2.Sentimentfeedback{
		Phrase:  platformclientv2.String(testPhrase),
		Dialect: platformclientv2.String("en-US"),
	}

	// Phrase match, dialect ignored (empty)
	assert.True(t, sentimentFeedbackMatchesFilters(feedback, testPhrase, ""))
	// Phrase + dialect both match
	assert.True(t, sentimentFeedbackMatchesFilters(feedback, testPhrase, "en-US"))
	// Phrase matches but dialect differs
	assert.False(t, sentimentFeedbackMatchesFilters(feedback, testPhrase, "en-GB"))
	// Phrase differs
	assert.False(t, sentimentFeedbackMatchesFilters(feedback, "other phrase", ""))

	// Nil phrase never matches
	nilPhrase := platformclientv2.Sentimentfeedback{Dialect: platformclientv2.String("en-US")}
	assert.False(t, sentimentFeedbackMatchesFilters(nilPhrase, testPhrase, ""))

	// Dialect filter set but entry has nil dialect -> no match
	nilDialect := platformclientv2.Sentimentfeedback{Phrase: platformclientv2.String(testPhrase)}
	assert.False(t, sentimentFeedbackMatchesFilters(nilDialect, testPhrase, "en-US"))
}

func TestUnitSentimentFeedbackResourceDataMapping(t *testing.T) {
	// getSentimentFeedbackFromResourceData: only writable fields, no read-only fields
	d := schema.TestResourceDataRaw(t, ResourceSentimentFeedback().Schema, buildTestSentimentFeedbackDataMap())
	body := getSentimentFeedbackFromResourceData(d)

	assert.Equal(t, testPhrase, *body.Phrase)
	assert.Equal(t, testDialect, *body.Dialect)
	assert.Equal(t, testFeedbackValue, *body.FeedbackValue)
	assert.Nil(t, body.Id)
	assert.Nil(t, body.DateCreated)
	assert.Nil(t, body.CreatedBy)

	// setSentimentFeedbackToResourceData: API object flows back into state (read-only fields ignored)
	empty := schema.TestResourceDataRaw(t, ResourceSentimentFeedback().Schema, map[string]interface{}{})
	apiObj := buildTestSentimentFeedback() // includes DateCreated + CreatedBy which must be safely ignored
	setSentimentFeedbackToResourceData(empty, &apiObj)

	assert.Equal(t, testPhrase, empty.Get("phrase").(string))
	assert.Equal(t, testDialect, empty.Get("dialect").(string))
	assert.Equal(t, testFeedbackValue, empty.Get("feedback_value").(string))
}

// notFoundError is a simple error type used to simulate a not-found API error in unit tests
type notFoundError struct{}

func (e *notFoundError) Error() string { return "not found" }
