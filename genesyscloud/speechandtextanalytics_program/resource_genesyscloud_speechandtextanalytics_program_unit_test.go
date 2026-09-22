package speechandtextanalytics_program

import (
	"context"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v199/platformclientv2"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/stretchr/testify/assert"
)

func buildTestProgram(id, name, description string, topicIds, tags []string, published bool) *platformclientv2.Program {
	topics := make([]platformclientv2.Basetopicentitiy, 0, len(topicIds))
	for _, tid := range topicIds {
		idCopy := tid
		topics = append(topics, platformclientv2.Basetopicentitiy{Id: &idCopy})
	}
	return &platformclientv2.Program{
		Id:          &id,
		Name:        &name,
		Description: &description,
		Topics:      &topics,
		Tags:        &tags,
		Published:   &published,
	}
}

func TestUnitResourceSpeechAndTextAnalyticsProgramCreate(t *testing.T) {
	tId := uuid.NewString()
	tName := "unit test program"
	tDescription := "unit test description"
	tTopicIds := []string{uuid.NewString(), uuid.NewString()}
	tTags := []string{"alpha", "beta"}

	proxy := &sttProgramProxy{}

	proxy.createProgramAttr = func(ctx context.Context, p *sttProgramProxy, body *platformclientv2.Programrequest) (*platformclientv2.Program, *platformclientv2.APIResponse, error) {
		// Only writable fields should be present in the request.
		assert.Equal(t, tName, *body.Name, "Name not equal")
		assert.Equal(t, tDescription, *body.Description, "Description not equal")
		assert.ElementsMatch(t, tTopicIds, *body.TopicIds, "TopicIds not equal")
		assert.ElementsMatch(t, tTags, *body.Tags, "Tags not equal")
		return buildTestProgram(tId, tName, tDescription, tTopicIds, tTags, false), &platformclientv2.APIResponse{StatusCode: http.StatusOK}, nil
	}
	proxy.getProgramAttr = func(ctx context.Context, p *sttProgramProxy, id string) (*platformclientv2.Program, *platformclientv2.APIResponse, error) {
		assert.Equal(t, tId, id)
		return buildTestProgram(tId, tName, tDescription, tTopicIds, tTags, false), &platformclientv2.APIResponse{StatusCode: http.StatusOK}, nil
	}

	internalProxy = proxy
	defer func() { internalProxy = nil }()

	ctx := context.Background()
	gcloud := &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}

	resourceSchema := ResourceSpeechAndTextAnalyticsProgram().Schema
	resourceDataMap := map[string]interface{}{
		"name":        tName,
		"description": tDescription,
		"topic_ids":   []interface{}{tTopicIds[0], tTopicIds[1]},
		"tags":        []interface{}{tTags[0], tTags[1]},
	}
	d := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)

	diag := createProgram(ctx, d, gcloud)
	assert.False(t, diag.HasError())
	assert.Equal(t, tId, d.Id())
	assert.Equal(t, tName, d.Get("name"))
	assert.Equal(t, tDescription, d.Get("description"))
	assert.Equal(t, false, d.Get("published"))
}

func TestUnitResourceSpeechAndTextAnalyticsProgramRead(t *testing.T) {
	tId := uuid.NewString()
	tName := "unit test program"
	tDescription := "unit test description"
	tTopicIds := []string{uuid.NewString()}
	tTags := []string{"alpha"}

	proxy := &sttProgramProxy{}
	proxy.getProgramAttr = func(ctx context.Context, p *sttProgramProxy, id string) (*platformclientv2.Program, *platformclientv2.APIResponse, error) {
		assert.Equal(t, tId, id)
		return buildTestProgram(tId, tName, tDescription, tTopicIds, tTags, true), &platformclientv2.APIResponse{StatusCode: http.StatusOK}, nil
	}

	internalProxy = proxy
	defer func() { internalProxy = nil }()

	ctx := context.Background()
	gcloud := &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}

	resourceSchema := ResourceSpeechAndTextAnalyticsProgram().Schema
	d := schema.TestResourceDataRaw(t, resourceSchema, map[string]interface{}{})
	d.SetId(tId)

	diag := readProgram(ctx, d, gcloud)
	assert.False(t, diag.HasError())
	assert.Equal(t, tId, d.Id())
	assert.Equal(t, tName, d.Get("name"))
	assert.Equal(t, tDescription, d.Get("description"))
	assert.Equal(t, true, d.Get("published"))

	gotTopicIds := d.Get("topic_ids").(*schema.Set).List()
	assert.Len(t, gotTopicIds, 1)
	assert.Equal(t, tTopicIds[0], gotTopicIds[0])
}

func TestUnitResourceSpeechAndTextAnalyticsProgramUpdate(t *testing.T) {
	tId := uuid.NewString()
	tName := "updated program"
	tDescription := "updated description"
	tTopicIds := []string{uuid.NewString()}
	tTags := []string{"gamma"}

	proxy := &sttProgramProxy{}
	proxy.updateProgramAttr = func(ctx context.Context, p *sttProgramProxy, id string, body *platformclientv2.Programrequest) (*platformclientv2.Program, *platformclientv2.APIResponse, error) {
		assert.Equal(t, tId, id)
		assert.Equal(t, tName, *body.Name)
		assert.Equal(t, tDescription, *body.Description)
		return buildTestProgram(tId, tName, tDescription, tTopicIds, tTags, false), &platformclientv2.APIResponse{StatusCode: http.StatusOK}, nil
	}
	proxy.getProgramAttr = func(ctx context.Context, p *sttProgramProxy, id string) (*platformclientv2.Program, *platformclientv2.APIResponse, error) {
		return buildTestProgram(tId, tName, tDescription, tTopicIds, tTags, false), &platformclientv2.APIResponse{StatusCode: http.StatusOK}, nil
	}

	internalProxy = proxy
	defer func() { internalProxy = nil }()

	ctx := context.Background()
	gcloud := &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}

	resourceSchema := ResourceSpeechAndTextAnalyticsProgram().Schema
	resourceDataMap := map[string]interface{}{
		"name":        tName,
		"description": tDescription,
		"topic_ids":   []interface{}{tTopicIds[0]},
		"tags":        []interface{}{tTags[0]},
	}
	d := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
	d.SetId(tId)

	diag := updateProgram(ctx, d, gcloud)
	assert.False(t, diag.HasError())
	assert.Equal(t, tName, d.Get("name"))
}

func TestUnitResourceSpeechAndTextAnalyticsProgramDelete(t *testing.T) {
	tId := uuid.NewString()

	proxy := &sttProgramProxy{}
	deleteCalled := false
	proxy.deleteProgramAttr = func(ctx context.Context, p *sttProgramProxy, id string, forceDelete bool) (*platformclientv2.APIResponse, error) {
		assert.Equal(t, tId, id)
		assert.True(t, forceDelete, "forceDelete should be true")
		deleteCalled = true
		return &platformclientv2.APIResponse{StatusCode: http.StatusOK}, nil
	}
	proxy.getProgramAttr = func(ctx context.Context, p *sttProgramProxy, id string) (*platformclientv2.Program, *platformclientv2.APIResponse, error) {
		return nil, &platformclientv2.APIResponse{StatusCode: http.StatusNotFound}, assert.AnError
	}

	internalProxy = proxy
	defer func() { internalProxy = nil }()

	ctx := context.Background()
	gcloud := &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}

	resourceSchema := ResourceSpeechAndTextAnalyticsProgram().Schema
	d := schema.TestResourceDataRaw(t, resourceSchema, map[string]interface{}{})
	d.SetId(tId)

	diag := deleteProgram(ctx, d, gcloud)
	assert.False(t, diag.HasError())
	assert.True(t, deleteCalled, "delete proxy method should have been called")
}

// TestUnitResourceSpeechAndTextAnalyticsProgramReadNotFound tests that Read clears state when program is 404.
func TestUnitResourceSpeechAndTextAnalyticsProgramReadNotFound(t *testing.T) {
	tId := uuid.NewString()

	proxy := &sttProgramProxy{}
	proxy.getProgramAttr = func(ctx context.Context, p *sttProgramProxy, id string) (*platformclientv2.Program, *platformclientv2.APIResponse, error) {
		return nil, &platformclientv2.APIResponse{StatusCode: http.StatusNotFound}, assert.AnError
	}

	internalProxy = proxy
	defer func() { internalProxy = nil }()

	ctx := context.Background()
	gcloud := &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}

	resourceSchema := ResourceSpeechAndTextAnalyticsProgram().Schema
	d := schema.TestResourceDataRaw(t, resourceSchema, map[string]interface{}{})
	d.SetId(tId)

	diag := readProgram(ctx, d, gcloud)
	assert.False(t, diag.HasError())
	assert.Equal(t, "", d.Id(), "resource ID should be cleared on 404")
}

// TestUnitResourceSpeechAndTextAnalyticsProgramCreateMinimal tests creating a program with only required fields.
func TestUnitResourceSpeechAndTextAnalyticsProgramCreateMinimal(t *testing.T) {
	tId := uuid.NewString()
	tName := "minimal program"

	proxy := &sttProgramProxy{}

	proxy.createProgramAttr = func(ctx context.Context, p *sttProgramProxy, body *platformclientv2.Programrequest) (*platformclientv2.Program, *platformclientv2.APIResponse, error) {
		// Only name should be required; description/topics/tags should be nil or empty.
		assert.Equal(t, tName, *body.Name)
		assert.Nil(t, body.Description, "Description should not be set if empty")
		assert.Nil(t, body.TopicIds, "TopicIds should not be set if empty")
		assert.Nil(t, body.Tags, "Tags should not be set if empty")
		return buildTestProgram(tId, tName, "", []string{}, []string{}, false), &platformclientv2.APIResponse{StatusCode: http.StatusOK}, nil
	}
	proxy.getProgramAttr = func(ctx context.Context, p *sttProgramProxy, id string) (*platformclientv2.Program, *platformclientv2.APIResponse, error) {
		return buildTestProgram(tId, tName, "", []string{}, []string{}, false), &platformclientv2.APIResponse{StatusCode: http.StatusOK}, nil
	}

	internalProxy = proxy
	defer func() { internalProxy = nil }()

	ctx := context.Background()
	gcloud := &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}

	resourceSchema := ResourceSpeechAndTextAnalyticsProgram().Schema
	resourceDataMap := map[string]interface{}{
		"name": tName,
	}
	d := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)

	diag := createProgram(ctx, d, gcloud)
	assert.False(t, diag.HasError())
	assert.Equal(t, tId, d.Id())
	assert.Equal(t, tName, d.Get("name"))
}

// TestUnitGetAllPrograms tests the getAllPrograms pagination logic with multiple pages.
func TestUnitGetAllPrograms(t *testing.T) {
	tProgramId1 := uuid.NewString()
	tProgramId2 := uuid.NewString()
	tProgramName1 := "Program 1"
	tProgramName2 := "Program 2"

	// Mock proxy that returns two pages of programs.
	proxy := &sttProgramProxy{}
	callCount := 0
	proxy.listProgramsAttr = func(ctx context.Context, p *sttProgramProxy, nextPage string, pageSize int) (*platformclientv2.Programsentitylisting, *platformclientv2.APIResponse, error) {
		callCount++
		if callCount == 1 {
			// First page: return 1 program + NextUri for page 2
			nextUri := "?nextPage=page2&pageSize=100"
			return &platformclientv2.Programsentitylisting{
				Entities: &[]platformclientv2.Listedprogram{
					{
						Id:   &tProgramId1,
						Name: &tProgramName1,
					},
				},
				NextUri: &nextUri,
			}, &platformclientv2.APIResponse{StatusCode: http.StatusOK}, nil
		}
		// Second page: return 1 program + no NextUri (end of results)
		return &platformclientv2.Programsentitylisting{
			Entities: &[]platformclientv2.Listedprogram{
				{
					Id:   &tProgramId2,
					Name: &tProgramName2,
				},
			},
			NextUri: nil,
		}, &platformclientv2.APIResponse{StatusCode: http.StatusOK}, nil
	}

	internalProxy = proxy
	defer func() { internalProxy = nil }()

	ctx := context.Background()
	config := &platformclientv2.Configuration{}

	resources, diag := getAllPrograms(ctx, config)
	assert.False(t, diag.HasError())
	assert.Len(t, resources, 2, "should have fetched 2 programs across 2 pages")
	assert.Contains(t, resources, tProgramId1)
	assert.Contains(t, resources, tProgramId2)
	assert.Equal(t, 2, callCount, "should have made 2 API calls for pagination")
}
