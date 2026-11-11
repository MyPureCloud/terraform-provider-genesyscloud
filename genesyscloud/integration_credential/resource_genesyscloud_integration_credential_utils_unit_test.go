package integration_credential

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v200/platformclientv2"
	"github.com/stretchr/testify/assert"
)

/*
The resource_genesyscloud_integration_credential_utils_unit_test.go contains unit tests for the
write-only fields handling (fields_wo / fields_wo_version) in the credential resource.
These tests run with TF_UNIT=1 and do not require a Genesys Cloud org.
*/

func TestUnitCredentialSchemaInternalValidate(t *testing.T) {
	assert.NoError(t, ResourceIntegrationCredential().InternalValidate(nil, true))
}

func TestUnitGetWriteOnlyFieldsPresent(t *testing.T) {
	d := schema.TestResourceDataRaw(t, ResourceIntegrationCredential().Schema, map[string]interface{}{
		"fields_wo":         `{"clientId":"someId","clientSecret":"someSecret"}`,
		"fields_wo_version": 1,
	})

	got, err := getWriteOnlyFieldsFromResource(d)
	assert.NoError(t, err)
	assert.Equal(t, map[string]string{"clientId": "someId", "clientSecret": "someSecret"}, got)
}

func TestUnitGetWriteOnlyFieldsAbsent(t *testing.T) {
	d := schema.TestResourceDataRaw(t, ResourceIntegrationCredential().Schema, map[string]interface{}{
		"fields": map[string]interface{}{"userName": "someUserName"},
	})

	got, err := getWriteOnlyFieldsFromResource(d)
	assert.NoError(t, err)
	assert.Nil(t, got)
}

func TestUnitGetWriteOnlyFieldsInvalidJson(t *testing.T) {
	d := schema.TestResourceDataRaw(t, ResourceIntegrationCredential().Schema, map[string]interface{}{
		"fields_wo":         `{"clientId": 12345}`,
		"fields_wo_version": 1,
	})

	_, err := getWriteOnlyFieldsFromResource(d)
	assert.Error(t, err)
}

func TestUnitParseWriteOnlyFieldsValid(t *testing.T) {
	got, err := parseWriteOnlyFields(`{"userName":"someUserName","password":"somePassword"}`)
	assert.NoError(t, err)
	assert.Equal(t, map[string]string{"userName": "someUserName", "password": "somePassword"}, got)
}

func TestUnitParseWriteOnlyFieldsRejectsNonStringValues(t *testing.T) {
	_, err := parseWriteOnlyFields(`{"retries": 3}`)
	assert.Error(t, err)
}

func TestUnitBuildCredentialFieldsPrefersWriteOnly(t *testing.T) {
	d := schema.TestResourceDataRaw(t, ResourceIntegrationCredential().Schema, map[string]interface{}{
		"credential_type_name": "userDefined",
		"fields_wo":            `{"authToken":"secretToken"}`,
		"fields_wo_version":    1,
	})

	got, err := buildCredentialFields(d, &platformclientv2.Configuration{})
	assert.NoError(t, err)
	assert.Equal(t, map[string]string{"authToken": "secretToken"}, got)
}

func TestUnitBuildCredentialFieldsFallsBackToStateFields(t *testing.T) {
	d := schema.TestResourceDataRaw(t, ResourceIntegrationCredential().Schema, map[string]interface{}{
		"credential_type_name": "basicAuth",
		"fields":               map[string]interface{}{"userName": "someUserName", "password": "somePassword"},
	})

	got, err := buildCredentialFields(d, &platformclientv2.Configuration{})
	assert.NoError(t, err)
	assert.Equal(t, map[string]string{"userName": "someUserName", "password": "somePassword"}, got)
}

// When fields_wo carries the clientSecret for a pureCloudOAuthClient credential, the
// oauth-client cache backfill (DEVTOOLING-448) must not run — no API call, no metadata
// cache read, the supplied write-only secret is used as-is.
func TestUnitBuildCredentialFieldsWriteOnlySecretSkipsCacheBackfill(t *testing.T) {
	d := schema.TestResourceDataRaw(t, ResourceIntegrationCredential().Schema, map[string]interface{}{
		"credential_type_name": "pureCloudOAuthClient",
		"fields_wo":            `{"clientId":"someId","clientSecret":"someSecret"}`,
		"fields_wo_version":    1,
	})

	got, err := buildCredentialFields(d, &platformclientv2.Configuration{})
	assert.NoError(t, err)
	assert.Equal(t, map[string]string{"clientId": "someId", "clientSecret": "someSecret"}, got)
}

func TestUnitBuildCredentialFieldsNilArgs(t *testing.T) {
	_, err := buildCredentialFields(nil, nil)
	assert.Error(t, err)
}
