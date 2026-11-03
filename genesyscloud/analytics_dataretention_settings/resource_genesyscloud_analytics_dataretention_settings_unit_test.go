package analytics_dataretention_settings

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v200/platformclientv2"
	"github.com/stretchr/testify/assert"
)

// Unit Test

func generateAnalyticsDataretentionSettingsData() platformclientv2.Analyticsdataretentionresponse {
	return platformclientv2.Analyticsdataretentionresponse{
		RetentionDays: platformclientv2.Int(3650),
	}
}

func buildAnalyticsDataretentionSettingsDataMap(s platformclientv2.Analyticsdataretentionresponse) map[string]interface{} {
	return map[string]interface{}{
		"retention_days": *s.RetentionDays,
	}
}

// newTestProxy returns a proxy whose getter returns the given settings. The update func is left to each test.
func newTestProxy(testSettings platformclientv2.Analyticsdataretentionresponse) *analyticsDataretentionSettingsProxy {
	proxy := &analyticsDataretentionSettingsProxy{}
	proxy.getAnalyticsDataretentionSettingsAttr = func(ctx context.Context, p *analyticsDataretentionSettingsProxy) (*platformclientv2.Analyticsdataretentionresponse, *platformclientv2.APIResponse, error) {
		settings := testSettings
		return &settings, &platformclientv2.APIResponse{StatusCode: http.StatusOK}, nil
	}
	return proxy
}

func TestUnitResourceAnalyticsDataretentionSettingsCreate(t *testing.T) {
	testSettings := generateAnalyticsDataretentionSettingsData()

	var capturedPutBody *platformclientv2.Updateanalyticsdataretentionrequest
	proxy := newTestProxy(testSettings)
	proxy.updateAnalyticsDataretentionSettingsAttr = func(ctx context.Context, p *analyticsDataretentionSettingsProxy, request *platformclientv2.Updateanalyticsdataretentionrequest) (*platformclientv2.Analyticsdataretentionresponse, *platformclientv2.APIResponse, error) {
		capturedPutBody = request
		settings := testSettings
		return &settings, &platformclientv2.APIResponse{StatusCode: http.StatusOK}, nil
	}
	internalProxy = proxy
	defer func() { internalProxy = nil }()

	ctx := context.Background()
	gcloud := &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}

	resourceSchema := ResourceAnalyticsDataretentionSettings().Schema
	resourceDataMap := buildAnalyticsDataretentionSettingsDataMap(testSettings)
	// A resource being created has no ID yet
	d := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)

	diag := createAnalyticsDataretentionSettings(ctx, d, gcloud)
	assert.Equal(t, false, diag.HasError())
	// Create must assign the fixed singleton ID
	assert.Equal(t, analyticsDataretentionSettingsId, d.Id())
	// Create delegates to update, so the PUT body must have been sent
	if assert.NotNil(t, capturedPutBody) && assert.NotNil(t, capturedPutBody.RetentionDays) {
		assert.Equal(t, *testSettings.RetentionDays, *capturedPutBody.RetentionDays)
	}
	assert.Equal(t, *testSettings.RetentionDays, d.Get("retention_days").(int))
}

func TestUnitResourceAnalyticsDataretentionSettingsRead(t *testing.T) {
	testSettings := generateAnalyticsDataretentionSettingsData()

	internalProxy = newTestProxy(testSettings)
	defer func() { internalProxy = nil }()

	ctx := context.Background()
	gcloud := &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}

	resourceSchema := ResourceAnalyticsDataretentionSettings().Schema
	// The configured value must match the API value, otherwise the consistency checker (correctly) keeps retrying
	resourceDataMap := buildAnalyticsDataretentionSettingsDataMap(testSettings)
	d := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
	d.SetId(analyticsDataretentionSettingsId)

	diag := readAnalyticsDataretentionSettings(ctx, d, gcloud)
	assert.Equal(t, false, diag.HasError())
	assert.Equal(t, analyticsDataretentionSettingsId, d.Id())
	assert.Equal(t, *testSettings.RetentionDays, d.Get("retention_days").(int))
}

// TestUnitSetAnalyticsDataretentionSettingsToResourceData verifies the API response overwrites the value held in state.
func TestUnitSetAnalyticsDataretentionSettingsToResourceData(t *testing.T) {
	testSettings := generateAnalyticsDataretentionSettingsData()

	resourceSchema := ResourceAnalyticsDataretentionSettings().Schema
	d := schema.TestResourceDataRaw(t, resourceSchema, map[string]interface{}{"retention_days": 90})

	setAnalyticsDataretentionSettingsToResourceData(d, &testSettings)
	assert.Equal(t, *testSettings.RetentionDays, d.Get("retention_days").(int))
}

// TestUnitGetAnalyticsDataretentionSettingsFromResourceData verifies the request body only carries retentionDays.
func TestUnitGetAnalyticsDataretentionSettingsFromResourceData(t *testing.T) {
	resourceSchema := ResourceAnalyticsDataretentionSettings().Schema
	d := schema.TestResourceDataRaw(t, resourceSchema, map[string]interface{}{"retention_days": 120})

	request := getAnalyticsDataretentionSettingsFromResourceData(d)
	if assert.NotNil(t, request.RetentionDays) {
		assert.Equal(t, 120, *request.RetentionDays)
	}
}

func TestUnitResourceAnalyticsDataretentionSettingsUpdate(t *testing.T) {
	testSettings := generateAnalyticsDataretentionSettingsData()

	proxy := newTestProxy(testSettings)
	proxy.updateAnalyticsDataretentionSettingsAttr = func(ctx context.Context, p *analyticsDataretentionSettingsProxy, request *platformclientv2.Updateanalyticsdataretentionrequest) (*platformclientv2.Analyticsdataretentionresponse, *platformclientv2.APIResponse, error) {
		// Verify the writable field is marshalled correctly from resource data
		if assert.NotNil(t, request.RetentionDays) {
			assert.Equal(t, *testSettings.RetentionDays, *request.RetentionDays)
		}
		settings := testSettings
		return &settings, &platformclientv2.APIResponse{StatusCode: http.StatusOK}, nil
	}
	internalProxy = proxy
	defer func() { internalProxy = nil }()

	ctx := context.Background()
	gcloud := &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}

	resourceSchema := ResourceAnalyticsDataretentionSettings().Schema
	resourceDataMap := buildAnalyticsDataretentionSettingsDataMap(testSettings)
	d := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
	d.SetId(analyticsDataretentionSettingsId)

	diag := updateAnalyticsDataretentionSettings(ctx, d, gcloud)
	assert.Equal(t, false, diag.HasError())
	assert.Equal(t, analyticsDataretentionSettingsId, d.Id())
}

func TestUnitResourceAnalyticsDataretentionSettingsDelete(t *testing.T) {
	ctx := context.Background()
	gcloud := &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}

	resourceSchema := ResourceAnalyticsDataretentionSettings().Schema
	d := schema.TestResourceDataRaw(t, resourceSchema, map[string]interface{}{"retention_days": 3650})
	d.SetId(analyticsDataretentionSettingsId)

	// Delete is a no-op for this singleton and should never error. The proxy is intentionally
	// not set: delete must not call the API.
	diag := deleteAnalyticsDataretentionSettings(ctx, d, gcloud)
	assert.Equal(t, false, diag.HasError())
}

func TestUnitDataSourceAnalyticsDataretentionSettingsRead(t *testing.T) {
	testSettings := generateAnalyticsDataretentionSettingsData()

	internalProxy = newTestProxy(testSettings)
	defer func() { internalProxy = nil }()

	ctx := context.Background()
	gcloud := &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}

	dataSourceSchema := DataSourceAnalyticsDataretentionSettings().Schema
	d := schema.TestResourceDataRaw(t, dataSourceSchema, map[string]interface{}{})

	diag := dataSourceAnalyticsDataretentionSettingsRead(ctx, d, gcloud)
	assert.Equal(t, false, diag.HasError())
	// The data source must set the fixed singleton ID and expose the setting value
	assert.Equal(t, analyticsDataretentionSettingsId, d.Id())
	assert.Equal(t, *testSettings.RetentionDays, d.Get("retention_days").(int))
}

// TestUnitResourceAnalyticsDataretentionSettingsReadError verifies that a non-404 API error from the
// proxy surfaces as an error diagnostic from the resource read. The custom retry timeout
// is forced to 0 so the read fails fast instead of retrying for the default 5 minutes.
func TestUnitResourceAnalyticsDataretentionSettingsReadError(t *testing.T) {
	originalEnv := os.Getenv("GENESYSCLOUD_CUSTOM_RETRY_TIMEOUT")
	os.Setenv("GENESYSCLOUD_CUSTOM_RETRY_TIMEOUT", "0")
	defer os.Setenv("GENESYSCLOUD_CUSTOM_RETRY_TIMEOUT", originalEnv)

	proxy := &analyticsDataretentionSettingsProxy{}
	proxy.getAnalyticsDataretentionSettingsAttr = func(ctx context.Context, p *analyticsDataretentionSettingsProxy) (*platformclientv2.Analyticsdataretentionresponse, *platformclientv2.APIResponse, error) {
		return nil, &platformclientv2.APIResponse{StatusCode: http.StatusInternalServerError}, fmt.Errorf("boom")
	}
	internalProxy = proxy
	defer func() { internalProxy = nil }()

	ctx := context.Background()
	gcloud := &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}

	resourceSchema := ResourceAnalyticsDataretentionSettings().Schema
	d := schema.TestResourceDataRaw(t, resourceSchema, map[string]interface{}{"retention_days": 3650})
	d.SetId(analyticsDataretentionSettingsId)

	diag := readAnalyticsDataretentionSettings(ctx, d, gcloud)
	assert.Equal(t, true, diag.HasError())
}

// TestUnitResourceAnalyticsDataretentionSettingsUpdateConflict verifies that a 409 returned by the PUT
// (existing policies or recordings exceed the requested value) surfaces as an error diagnostic and
// short-circuits before the read-back.
func TestUnitResourceAnalyticsDataretentionSettingsUpdateConflict(t *testing.T) {
	proxy := &analyticsDataretentionSettingsProxy{}
	proxy.updateAnalyticsDataretentionSettingsAttr = func(ctx context.Context, p *analyticsDataretentionSettingsProxy, request *platformclientv2.Updateanalyticsdataretentionrequest) (*platformclientv2.Analyticsdataretentionresponse, *platformclientv2.APIResponse, error) {
		return nil, &platformclientv2.APIResponse{StatusCode: http.StatusConflict}, fmt.Errorf("analytics.data.retention.update.conflict")
	}
	// getAnalyticsDataretentionSettingsAttr intentionally left unset: update must fail before any read-back.
	internalProxy = proxy
	defer func() { internalProxy = nil }()

	ctx := context.Background()
	gcloud := &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}

	resourceSchema := ResourceAnalyticsDataretentionSettings().Schema
	d := schema.TestResourceDataRaw(t, resourceSchema, map[string]interface{}{"retention_days": 90})
	d.SetId(analyticsDataretentionSettingsId)

	diag := updateAnalyticsDataretentionSettings(ctx, d, gcloud)
	assert.Equal(t, true, diag.HasError())
}

// TestUnitDataSourceAnalyticsDataretentionSettingsReadError verifies that an API error from the proxy
// surfaces as an error diagnostic from the data source read.
func TestUnitDataSourceAnalyticsDataretentionSettingsReadError(t *testing.T) {
	proxy := &analyticsDataretentionSettingsProxy{}
	proxy.getAnalyticsDataretentionSettingsAttr = func(ctx context.Context, p *analyticsDataretentionSettingsProxy) (*platformclientv2.Analyticsdataretentionresponse, *platformclientv2.APIResponse, error) {
		return nil, &platformclientv2.APIResponse{StatusCode: http.StatusInternalServerError}, fmt.Errorf("boom")
	}
	internalProxy = proxy
	defer func() { internalProxy = nil }()

	ctx := context.Background()
	gcloud := &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}

	dataSourceSchema := DataSourceAnalyticsDataretentionSettings().Schema
	d := schema.TestResourceDataRaw(t, dataSourceSchema, map[string]interface{}{})

	diag := dataSourceAnalyticsDataretentionSettingsRead(ctx, d, gcloud)
	assert.Equal(t, true, diag.HasError())
}

// TestUnitGetAllAnalyticsDataretentionSettings verifies the exporter lookup: one fixed-ID entry when the
// setting exists, nothing (and no error) on a 404 because the setting was never configured, and an error
// for any other failure.
func TestUnitGetAllAnalyticsDataretentionSettings(t *testing.T) {
	ctx := context.Background()
	clientConfig := &platformclientv2.Configuration{}

	t.Run("setting exists", func(t *testing.T) {
		internalProxy = newTestProxy(generateAnalyticsDataretentionSettingsData())
		defer func() { internalProxy = nil }()

		resources, diag := getAllAnalyticsDataretentionSettings(ctx, clientConfig)
		assert.Equal(t, false, diag.HasError())
		assert.Len(t, resources, 1)
		if meta, ok := resources[analyticsDataretentionSettingsId]; assert.True(t, ok) {
			assert.Equal(t, ResourceType, meta.BlockLabel)
		}
	})

	t.Run("setting never configured (404)", func(t *testing.T) {
		proxy := &analyticsDataretentionSettingsProxy{}
		proxy.getAnalyticsDataretentionSettingsAttr = func(ctx context.Context, p *analyticsDataretentionSettingsProxy) (*platformclientv2.Analyticsdataretentionresponse, *platformclientv2.APIResponse, error) {
			return nil, &platformclientv2.APIResponse{StatusCode: http.StatusNotFound}, fmt.Errorf("not found")
		}
		internalProxy = proxy
		defer func() { internalProxy = nil }()

		resources, diag := getAllAnalyticsDataretentionSettings(ctx, clientConfig)
		assert.Equal(t, false, diag.HasError())
		assert.Empty(t, resources)
	})

	t.Run("other API error", func(t *testing.T) {
		proxy := &analyticsDataretentionSettingsProxy{}
		proxy.getAnalyticsDataretentionSettingsAttr = func(ctx context.Context, p *analyticsDataretentionSettingsProxy) (*platformclientv2.Analyticsdataretentionresponse, *platformclientv2.APIResponse, error) {
			return nil, &platformclientv2.APIResponse{StatusCode: http.StatusForbidden}, fmt.Errorf("forbidden")
		}
		internalProxy = proxy
		defer func() { internalProxy = nil }()

		_, diag := getAllAnalyticsDataretentionSettings(ctx, clientConfig)
		assert.Equal(t, true, diag.HasError())
	})
}

// TestUnitAnalyticsDataretentionSettingsRetentionDaysValidation verifies the plan-time validation of retention_days.
func TestUnitAnalyticsDataretentionSettingsRetentionDaysValidation(t *testing.T) {
	validate := ResourceAnalyticsDataretentionSettings().Schema["retention_days"].ValidateFunc

	_, errs := validate(0, "retention_days")
	assert.NotEmpty(t, errs, "0 days must be rejected")

	_, errs = validate(-1, "retention_days")
	assert.NotEmpty(t, errs, "negative days must be rejected")

	_, errs = validate(90, "retention_days")
	assert.Empty(t, errs, "90 days must be accepted")
}

// TestUnitAnalyticsDataretentionSettingsSchema verifies the schema is valid and retention_days is the only,
// required, attribute of the resource.
func TestUnitAnalyticsDataretentionSettingsSchema(t *testing.T) {
	r := ResourceAnalyticsDataretentionSettings()
	assert.NoError(t, r.InternalValidate(nil, true))
	assert.Len(t, r.Schema, 1)
	assert.True(t, r.Schema["retention_days"].Required)

	ds := DataSourceAnalyticsDataretentionSettings()
	assert.NoError(t, ds.InternalValidate(nil, false))
	assert.True(t, ds.Schema["retention_days"].Computed)
}
