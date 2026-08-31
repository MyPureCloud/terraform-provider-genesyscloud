package speechandtextanalytics_settings

import (
	"context"
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
