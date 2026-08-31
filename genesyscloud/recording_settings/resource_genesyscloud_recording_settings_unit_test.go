package recording_settings

import (
	"context"
	"net/http"
	"testing"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v195/platformclientv2"
	"github.com/stretchr/testify/assert"
)

// Unit Test

func generateRecordingSettingsData() platformclientv2.Recordingsettings {
	return platformclientv2.Recordingsettings{
		MaxSimultaneousStreams:                    platformclientv2.Int(100),
		MaxConfigurableScreenRecordingStreams:     platformclientv2.Int(200),
		RegionalRecordingStorageEnabled:           platformclientv2.Bool(true),
		RecordingPlaybackUrlTtl:                   platformclientv2.Int(60),
		RecordingBatchDownloadUrlTtl:              platformclientv2.Int(30),
		StopRecordingWhenOnlyExternalParticipants: platformclientv2.Bool(true),
	}
}

func buildRecordingSettingsDataMap(s platformclientv2.Recordingsettings) map[string]interface{} {
	return map[string]interface{}{
		"max_simultaneous_streams":                       *s.MaxSimultaneousStreams,
		"max_configurable_screen_recording_streams":      *s.MaxConfigurableScreenRecordingStreams,
		"regional_recording_storage_enabled":             *s.RegionalRecordingStorageEnabled,
		"recording_playback_url_ttl":                     *s.RecordingPlaybackUrlTtl,
		"recording_batch_download_url_ttl":               *s.RecordingBatchDownloadUrlTtl,
		"stop_recording_when_only_external_participants": *s.StopRecordingWhenOnlyExternalParticipants,
	}
}

func TestUnitResourceRecordingSettingsCreate(t *testing.T) {
	testSettings := generateRecordingSettingsData()

	var capturedPutBody *platformclientv2.Recordingsettings
	proxy := &recordingSettingsProxy{}
	proxy.updateRecordingSettingsAttr = func(ctx context.Context, p *recordingSettingsProxy, settings *platformclientv2.Recordingsettings) (*platformclientv2.Recordingsettings, *platformclientv2.APIResponse, error) {
		capturedPutBody = settings
		return settings, &platformclientv2.APIResponse{StatusCode: http.StatusOK}, nil
	}
	proxy.getRecordingSettingsAttr = func(ctx context.Context, p *recordingSettingsProxy) (*platformclientv2.Recordingsettings, *platformclientv2.APIResponse, error) {
		settings := testSettings
		return &settings, &platformclientv2.APIResponse{StatusCode: http.StatusOK}, nil
	}
	internalProxy = proxy
	defer func() { internalProxy = nil }()

	ctx := context.Background()
	gcloud := &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}

	resourceSchema := ResourceRecordingSettings().Schema
	resourceDataMap := buildRecordingSettingsDataMap(testSettings)
	// A resource being created has no ID yet
	d := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)

	diag := createRecordingSettings(ctx, d, gcloud)
	assert.Equal(t, false, diag.HasError())
	// Create must assign the fixed singleton ID
	assert.Equal(t, recordingSettingsId, d.Id())
	// Create delegates to update, so the PUT body must have been sent
	assert.NotNil(t, capturedPutBody)
	assert.Equal(t, *testSettings.MaxSimultaneousStreams, *capturedPutBody.MaxSimultaneousStreams)
	// The read-only field must never be sent in the create/update body
	assert.Nil(t, capturedPutBody.MaxConfigurableScreenRecordingStreams)
}

func TestUnitResourceRecordingSettingsRead(t *testing.T) {
	tId := uuid.NewString()
	testSettings := generateRecordingSettingsData()

	proxy := &recordingSettingsProxy{}
	proxy.getRecordingSettingsAttr = func(ctx context.Context, p *recordingSettingsProxy) (*platformclientv2.Recordingsettings, *platformclientv2.APIResponse, error) {
		settings := testSettings
		return &settings, &platformclientv2.APIResponse{StatusCode: http.StatusOK}, nil
	}
	internalProxy = proxy
	defer func() { internalProxy = nil }()

	ctx := context.Background()
	gcloud := &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}

	resourceSchema := ResourceRecordingSettings().Schema
	resourceDataMap := buildRecordingSettingsDataMap(testSettings)
	d := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
	d.SetId(tId)

	diag := readRecordingSettings(ctx, d, gcloud)
	assert.Equal(t, false, diag.HasError())
	assert.Equal(t, tId, d.Id())
	assert.Equal(t, *testSettings.MaxSimultaneousStreams, d.Get("max_simultaneous_streams").(int))
	assert.Equal(t, *testSettings.MaxConfigurableScreenRecordingStreams, d.Get("max_configurable_screen_recording_streams").(int))
	assert.Equal(t, *testSettings.RegionalRecordingStorageEnabled, d.Get("regional_recording_storage_enabled").(bool))
	assert.Equal(t, *testSettings.RecordingPlaybackUrlTtl, d.Get("recording_playback_url_ttl").(int))
	assert.Equal(t, *testSettings.RecordingBatchDownloadUrlTtl, d.Get("recording_batch_download_url_ttl").(int))
	assert.Equal(t, *testSettings.StopRecordingWhenOnlyExternalParticipants, d.Get("stop_recording_when_only_external_participants").(bool))
}

func TestUnitResourceRecordingSettingsUpdate(t *testing.T) {
	tId := uuid.NewString()
	testSettings := generateRecordingSettingsData()

	proxy := &recordingSettingsProxy{}
	proxy.getRecordingSettingsAttr = func(ctx context.Context, p *recordingSettingsProxy) (*platformclientv2.Recordingsettings, *platformclientv2.APIResponse, error) {
		settings := testSettings
		return &settings, &platformclientv2.APIResponse{StatusCode: http.StatusOK}, nil
	}
	proxy.updateRecordingSettingsAttr = func(ctx context.Context, p *recordingSettingsProxy, settings *platformclientv2.Recordingsettings) (*platformclientv2.Recordingsettings, *platformclientv2.APIResponse, error) {
		// Verify the writable fields are marshalled correctly from resource data
		assert.Equal(t, *testSettings.MaxSimultaneousStreams, *settings.MaxSimultaneousStreams)
		assert.Equal(t, *testSettings.RegionalRecordingStorageEnabled, *settings.RegionalRecordingStorageEnabled)
		assert.Equal(t, *testSettings.RecordingPlaybackUrlTtl, *settings.RecordingPlaybackUrlTtl)
		assert.Equal(t, *testSettings.RecordingBatchDownloadUrlTtl, *settings.RecordingBatchDownloadUrlTtl)
		assert.Equal(t, *testSettings.StopRecordingWhenOnlyExternalParticipants, *settings.StopRecordingWhenOnlyExternalParticipants)
		// The read-only field must NOT be sent in the update body
		assert.Nil(t, settings.MaxConfigurableScreenRecordingStreams)
		return settings, &platformclientv2.APIResponse{StatusCode: http.StatusOK}, nil
	}
	internalProxy = proxy
	defer func() { internalProxy = nil }()

	ctx := context.Background()
	gcloud := &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}

	resourceSchema := ResourceRecordingSettings().Schema
	resourceDataMap := buildRecordingSettingsDataMap(testSettings)
	d := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
	d.SetId(tId)

	diag := updateRecordingSettings(ctx, d, gcloud)
	assert.Equal(t, false, diag.HasError())
	assert.Equal(t, tId, d.Id())
}

func TestUnitResourceRecordingSettingsDelete(t *testing.T) {
	ctx := context.Background()
	gcloud := &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}

	resourceSchema := ResourceRecordingSettings().Schema
	d := schema.TestResourceDataRaw(t, resourceSchema, map[string]interface{}{})
	d.SetId(recordingSettingsId)

	// Delete is a no-op for this singleton and should never error
	diag := deleteRecordingSettings(ctx, d, gcloud)
	assert.Equal(t, false, diag.HasError())
}

func TestUnitDataSourceRecordingSettingsRead(t *testing.T) {
	testSettings := generateRecordingSettingsData()

	proxy := &recordingSettingsProxy{}
	proxy.getRecordingSettingsAttr = func(ctx context.Context, p *recordingSettingsProxy) (*platformclientv2.Recordingsettings, *platformclientv2.APIResponse, error) {
		settings := testSettings
		return &settings, &platformclientv2.APIResponse{StatusCode: http.StatusOK}, nil
	}
	internalProxy = proxy
	defer func() { internalProxy = nil }()

	ctx := context.Background()
	gcloud := &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}

	dataSourceSchema := DataSourceRecordingSettings().Schema
	d := schema.TestResourceDataRaw(t, dataSourceSchema, map[string]interface{}{})

	diag := dataSourceRecordingSettingsRead(ctx, d, gcloud)
	assert.Equal(t, false, diag.HasError())
	// The data source must set the fixed singleton ID and expose the settings values
	assert.Equal(t, recordingSettingsId, d.Id())
	assert.Equal(t, *testSettings.MaxSimultaneousStreams, d.Get("max_simultaneous_streams").(int))
	assert.Equal(t, *testSettings.MaxConfigurableScreenRecordingStreams, d.Get("max_configurable_screen_recording_streams").(int))
	assert.Equal(t, *testSettings.RegionalRecordingStorageEnabled, d.Get("regional_recording_storage_enabled").(bool))
	assert.Equal(t, *testSettings.RecordingPlaybackUrlTtl, d.Get("recording_playback_url_ttl").(int))
	assert.Equal(t, *testSettings.RecordingBatchDownloadUrlTtl, d.Get("recording_batch_download_url_ttl").(int))
	assert.Equal(t, *testSettings.StopRecordingWhenOnlyExternalParticipants, d.Get("stop_recording_when_only_external_participants").(bool))
}
