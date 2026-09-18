package speechandtextanalytics_settings

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v195/platformclientv2"
	"github.com/stretchr/testify/assert"
)

// Unit Tests

func generateSpeechAndTextAnalyticsSettingsResponse(programId string, dialects []string, textAnalytics bool, agentEmpathy bool) platformclientv2.Speechtextanalyticssettingsresponse {
	return platformclientv2.Speechtextanalyticssettingsresponse{
		DefaultProgram:       &platformclientv2.Addressableentityref{Id: platformclientv2.String(programId)},
		ExpectedDialects:     &dialects,
		TextAnalyticsEnabled: platformclientv2.Bool(textAnalytics),
		AgentEmpathyEnabled:  platformclientv2.Bool(agentEmpathy),
	}
}

func buildSpeechAndTextAnalyticsSettingsDataMap(programId string, dialects []interface{}, textAnalytics bool, agentEmpathy bool) map[string]interface{} {
	return map[string]interface{}{
		"default_program_id":     programId,
		"expected_dialects":      dialects,
		"text_analytics_enabled": textAnalytics,
		"agent_empathy_enabled":  agentEmpathy,
	}
}

func TestUnitResourceSpeechAndTextAnalyticsSettingsRead(t *testing.T) {
	programId := "program-123"
	dialects := []string{"en-US", "es-US"}
	dialectsIface := make([]interface{}, len(dialects))
	for i, v := range dialects {
		dialectsIface[i] = v
	}
	testResponse := generateSpeechAndTextAnalyticsSettingsResponse(programId, dialects, true, false)

	proxy := &speechAndTextAnalyticsSettingsProxy{}
	proxy.getSpeechAndTextAnalyticsSettingsAttr = func(ctx context.Context, p *speechAndTextAnalyticsSettingsProxy) (*platformclientv2.Speechtextanalyticssettingsresponse, *platformclientv2.APIResponse, error) {
		resp := testResponse
		return &resp, &platformclientv2.APIResponse{StatusCode: http.StatusOK}, nil
	}
	internalProxy = proxy
	defer func() { internalProxy = nil }()

	ctx := context.Background()
	gcloud := &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}

	d := schema.TestResourceDataRaw(t, ResourceSpeechAndTextAnalyticsSettings().Schema, buildSpeechAndTextAnalyticsSettingsDataMap(programId, dialectsIface, true, false))
	d.SetId(speechAndTextAnalyticsSettingsId)

	diag := readSpeechAndTextAnalyticsSettings(ctx, d, gcloud)
	assert.Equal(t, false, diag.HasError())
	assert.Equal(t, speechAndTextAnalyticsSettingsId, d.Id())
	assert.Equal(t, programId, d.Get("default_program_id").(string))
	assert.Equal(t, true, d.Get("text_analytics_enabled").(bool))
	assert.Equal(t, false, d.Get("agent_empathy_enabled").(bool))
	assert.Equal(t, 2, len(d.Get("expected_dialects").([]interface{})))
}

func TestUnitResourceSpeechAndTextAnalyticsSettingsUpdate(t *testing.T) {
	programId := "program-456"
	dialects := []string{"en-GB"}
	dialectsIface := make([]interface{}, len(dialects))
	for i, v := range dialects {
		dialectsIface[i] = v
	}
	testResponse := generateSpeechAndTextAnalyticsSettingsResponse(programId, dialects, false, true)

	var capturedPutBody *platformclientv2.Speechtextanalyticssettingsrequest
	proxy := &speechAndTextAnalyticsSettingsProxy{}
	proxy.getSpeechAndTextAnalyticsSettingsAttr = func(ctx context.Context, p *speechAndTextAnalyticsSettingsProxy) (*platformclientv2.Speechtextanalyticssettingsresponse, *platformclientv2.APIResponse, error) {
		resp := testResponse
		return &resp, &platformclientv2.APIResponse{StatusCode: http.StatusOK}, nil
	}
	proxy.updateSpeechAndTextAnalyticsSettingsAttr = func(ctx context.Context, p *speechAndTextAnalyticsSettingsProxy, settings *platformclientv2.Speechtextanalyticssettingsrequest) (*platformclientv2.Speechtextanalyticssettingsresponse, *platformclientv2.APIResponse, error) {
		capturedPutBody = settings
		return &testResponse, &platformclientv2.APIResponse{StatusCode: http.StatusOK}, nil
	}
	internalProxy = proxy
	defer func() { internalProxy = nil }()

	ctx := context.Background()
	gcloud := &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}

	d := schema.TestResourceDataRaw(t, ResourceSpeechAndTextAnalyticsSettings().Schema, buildSpeechAndTextAnalyticsSettingsDataMap(programId, dialectsIface, false, true))
	d.SetId(speechAndTextAnalyticsSettingsId)

	diag := updateSpeechAndTextAnalyticsSettings(ctx, d, gcloud)
	assert.Equal(t, false, diag.HasError())
	assert.Equal(t, speechAndTextAnalyticsSettingsId, d.Id())

	// Verify the PUT body sent the schema values through the request model
	assert.NotNil(t, capturedPutBody)
	assert.Equal(t, programId, *capturedPutBody.DefaultProgramId)
	assert.Equal(t, false, *capturedPutBody.TextAnalyticsEnabled)
	assert.Equal(t, true, *capturedPutBody.AgentEmpathyEnabled)
	assert.Equal(t, dialects, *capturedPutBody.ExpectedDialects)
}

func TestUnitDataSourceSpeechAndTextAnalyticsSettingsRead(t *testing.T) {
	programId := "program-789"
	dialects := []string{"fr-FR"}
	testResponse := generateSpeechAndTextAnalyticsSettingsResponse(programId, dialects, true, true)

	proxy := &speechAndTextAnalyticsSettingsProxy{}
	proxy.getSpeechAndTextAnalyticsSettingsAttr = func(ctx context.Context, p *speechAndTextAnalyticsSettingsProxy) (*platformclientv2.Speechtextanalyticssettingsresponse, *platformclientv2.APIResponse, error) {
		resp := testResponse
		return &resp, &platformclientv2.APIResponse{StatusCode: http.StatusOK}, nil
	}
	internalProxy = proxy
	defer func() { internalProxy = nil }()

	ctx := context.Background()
	gcloud := &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}

	d := schema.TestResourceDataRaw(t, DataSourceSpeechAndTextAnalyticsSettings().Schema, map[string]interface{}{})

	diag := dataSourceSpeechAndTextAnalyticsSettingsRead(ctx, d, gcloud)
	assert.Equal(t, false, diag.HasError())
	assert.Equal(t, speechAndTextAnalyticsSettingsId, d.Id())
	assert.Equal(t, programId, d.Get("default_program_id").(string))
	assert.Equal(t, true, d.Get("text_analytics_enabled").(bool))
	assert.Equal(t, true, d.Get("agent_empathy_enabled").(bool))
}

// TestUnitResourceSpeechAndTextAnalyticsSettingsReadNilDefaultProgram verifies that when the API
// returns no default program, the flatten leaves default_program_id unset (null) rather than
// coercing it to an empty string. It also covers an empty expected_dialects list.
func TestUnitResourceSpeechAndTextAnalyticsSettingsReadNilDefaultProgram(t *testing.T) {
	emptyDialects := []string{}
	testResponse := platformclientv2.Speechtextanalyticssettingsresponse{
		DefaultProgram:       nil,
		ExpectedDialects:     &emptyDialects,
		TextAnalyticsEnabled: platformclientv2.Bool(false),
		AgentEmpathyEnabled:  platformclientv2.Bool(false),
	}

	proxy := &speechAndTextAnalyticsSettingsProxy{}
	proxy.getSpeechAndTextAnalyticsSettingsAttr = func(ctx context.Context, p *speechAndTextAnalyticsSettingsProxy) (*platformclientv2.Speechtextanalyticssettingsresponse, *platformclientv2.APIResponse, error) {
		resp := testResponse
		return &resp, &platformclientv2.APIResponse{StatusCode: http.StatusOK}, nil
	}
	internalProxy = proxy
	defer func() { internalProxy = nil }()

	ctx := context.Background()
	gcloud := &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}

	d := schema.TestResourceDataRaw(t, ResourceSpeechAndTextAnalyticsSettings().Schema, map[string]interface{}{})
	d.SetId(speechAndTextAnalyticsSettingsId)

	diag := readSpeechAndTextAnalyticsSettings(ctx, d, gcloud)
	assert.Equal(t, false, diag.HasError())
	// default_program_id must remain empty (unset) when the API returns no default program
	assert.Equal(t, "", d.Get("default_program_id").(string))
	assert.Equal(t, false, d.Get("text_analytics_enabled").(bool))
	assert.Equal(t, false, d.Get("agent_empathy_enabled").(bool))
	assert.Equal(t, 0, len(d.Get("expected_dialects").([]interface{})))
}

// TestUnitResourceSpeechAndTextAnalyticsSettingsReadError verifies that a non-404 API error
// surfaces as a diagnostic error from read.
func TestUnitResourceSpeechAndTextAnalyticsSettingsReadError(t *testing.T) {
	proxy := &speechAndTextAnalyticsSettingsProxy{}
	proxy.getSpeechAndTextAnalyticsSettingsAttr = func(ctx context.Context, p *speechAndTextAnalyticsSettingsProxy) (*platformclientv2.Speechtextanalyticssettingsresponse, *platformclientv2.APIResponse, error) {
		return nil, &platformclientv2.APIResponse{StatusCode: http.StatusInternalServerError}, fmt.Errorf("internal server error")
	}
	internalProxy = proxy
	defer func() { internalProxy = nil }()

	ctx := context.Background()
	gcloud := &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}

	d := schema.TestResourceDataRaw(t, ResourceSpeechAndTextAnalyticsSettings().Schema, map[string]interface{}{})
	d.SetId(speechAndTextAnalyticsSettingsId)

	diag := readSpeechAndTextAnalyticsSettings(ctx, d, gcloud)
	assert.Equal(t, true, diag.HasError())
}

// TestUnitResourceSpeechAndTextAnalyticsSettingsCreate verifies that create sets the fixed
// singleton ID and delegates to update (sending the PUT body).
func TestUnitResourceSpeechAndTextAnalyticsSettingsCreate(t *testing.T) {
	programId := "program-create"
	dialects := []string{"en-US"}
	dialectsIface := make([]interface{}, len(dialects))
	for i, v := range dialects {
		dialectsIface[i] = v
	}
	testResponse := generateSpeechAndTextAnalyticsSettingsResponse(programId, dialects, true, false)

	var putCalled bool
	proxy := &speechAndTextAnalyticsSettingsProxy{}
	proxy.getSpeechAndTextAnalyticsSettingsAttr = func(ctx context.Context, p *speechAndTextAnalyticsSettingsProxy) (*platformclientv2.Speechtextanalyticssettingsresponse, *platformclientv2.APIResponse, error) {
		resp := testResponse
		return &resp, &platformclientv2.APIResponse{StatusCode: http.StatusOK}, nil
	}
	proxy.updateSpeechAndTextAnalyticsSettingsAttr = func(ctx context.Context, p *speechAndTextAnalyticsSettingsProxy, settings *platformclientv2.Speechtextanalyticssettingsrequest) (*platformclientv2.Speechtextanalyticssettingsresponse, *platformclientv2.APIResponse, error) {
		putCalled = true
		return &testResponse, &platformclientv2.APIResponse{StatusCode: http.StatusOK}, nil
	}
	internalProxy = proxy
	defer func() { internalProxy = nil }()

	ctx := context.Background()
	gcloud := &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}

	d := schema.TestResourceDataRaw(t, ResourceSpeechAndTextAnalyticsSettings().Schema, buildSpeechAndTextAnalyticsSettingsDataMap(programId, dialectsIface, true, false))

	diag := createSpeechAndTextAnalyticsSettings(ctx, d, gcloud)
	assert.Equal(t, false, diag.HasError())
	assert.Equal(t, true, putCalled)
	assert.Equal(t, speechAndTextAnalyticsSettingsId, d.Id())
}

// TestUnitGetAllSpeechAndTextAnalyticsSettings verifies the exporter getAll returns the single
// singleton entry when the API call succeeds.
func TestUnitGetAllSpeechAndTextAnalyticsSettings(t *testing.T) {
	testResponse := generateSpeechAndTextAnalyticsSettingsResponse("program-1", []string{"en-US"}, true, true)

	proxy := &speechAndTextAnalyticsSettingsProxy{}
	proxy.getSpeechAndTextAnalyticsSettingsAttr = func(ctx context.Context, p *speechAndTextAnalyticsSettingsProxy) (*platformclientv2.Speechtextanalyticssettingsresponse, *platformclientv2.APIResponse, error) {
		resp := testResponse
		return &resp, &platformclientv2.APIResponse{StatusCode: http.StatusOK}, nil
	}
	internalProxy = proxy
	defer func() { internalProxy = nil }()

	ctx := context.Background()
	resources, diag := getAllSpeechAndTextAnalyticsSettings(ctx, &platformclientv2.Configuration{})
	assert.Equal(t, false, diag.HasError())
	assert.Equal(t, 1, len(resources))
	_, ok := resources[speechAndTextAnalyticsSettingsId]
	assert.Equal(t, true, ok)
}

// TestUnitGetAllSpeechAndTextAnalyticsSettings404 verifies the exporter getAll returns an empty
// map (no error) when the API returns 404, so the resource is simply not exported.
func TestUnitGetAllSpeechAndTextAnalyticsSettings404(t *testing.T) {
	proxy := &speechAndTextAnalyticsSettingsProxy{}
	proxy.getSpeechAndTextAnalyticsSettingsAttr = func(ctx context.Context, p *speechAndTextAnalyticsSettingsProxy) (*platformclientv2.Speechtextanalyticssettingsresponse, *platformclientv2.APIResponse, error) {
		return nil, &platformclientv2.APIResponse{StatusCode: http.StatusNotFound}, fmt.Errorf("not found")
	}
	internalProxy = proxy
	defer func() { internalProxy = nil }()

	ctx := context.Background()
	resources, diag := getAllSpeechAndTextAnalyticsSettings(ctx, &platformclientv2.Configuration{})
	assert.Equal(t, false, diag.HasError())
	assert.Equal(t, 0, len(resources))
}
