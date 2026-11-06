package integration_credential

import (
	"testing"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v200/platformclientv2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Unit tests for the write-only credential fields (fields_wo / fields_wo_version), DEVTOOLING-1899.
// A distinctive marker is used as the "secret" so tests can assert it never appears in error text.
const testSecretMarker = "TOPSECRET-MARKER-1899"

// rawConfigWith builds a raw config object of the resource's implied type with every attribute null
// except the supplied overrides. This is what Terraform sends the provider as the config.
func rawConfigWith(overrides map[string]cty.Value) cty.Value {
	implied := ResourceIntegrationCredential().CoreConfigSchema().ImpliedType()
	attrs := make(map[string]cty.Value, len(implied.AttributeTypes()))
	for name, ty := range implied.AttributeTypes() {
		attrs[name] = cty.NullVal(ty)
	}
	for k, v := range overrides {
		attrs[k] = v
	}
	return cty.ObjectVal(attrs)
}

// dataWithRawConfig returns ResourceData whose raw config is the given value and whose state holds the
// given attributes (used for the existing `fields` map and credential_type_name).
func dataWithRawConfig(raw cty.Value, stateAttrs map[string]string) *schema.ResourceData {
	state := &terraform.InstanceState{ID: "cred-id", Attributes: stateAttrs, RawConfig: raw}
	return ResourceIntegrationCredential().Data(state)
}

func TestUnitParseFieldsWoJSON(t *testing.T) {
	t.Run("valid object", func(t *testing.T) {
		got, err := parseFieldsWoJSON(`{"userName":"u","password":"` + testSecretMarker + `"}`)
		require.NoError(t, err)
		assert.Equal(t, map[string]string{"userName": "u", "password": testSecretMarker}, got)
	})

	errCases := map[string]struct {
		input       string
		wantContain string
	}{
		"syntax error reports offset only": {`{"password":"` + testSecretMarker + `"`, "offset"},
		"top level array":                  {`["` + testSecretMarker + `"]`, "JSON object"},
		"top level string":                 {`"` + testSecretMarker + `"`, "JSON object"},
		"empty object is rejected":         {`{}`, "at least one field"},
		"json null is rejected":            {`null`, "at least one field"},
		"number value":                     {`{"port":443,"password":"` + testSecretMarker + `"}`, `"port"`},
		"boolean value":                    {`{"flag":true}`, `"flag"`},
		"nested object value":              {`{"nested":{"password":"` + testSecretMarker + `"}}`, `"nested"`},
	}
	for name, tc := range errCases {
		t.Run(name, func(t *testing.T) {
			got, err := parseFieldsWoJSON(tc.input)
			require.Error(t, err)
			assert.Nil(t, got)
			assert.Contains(t, err.Error(), tc.wantContain)
			// The value is a secret: errors may name keys/offsets but must never echo any value.
			assert.NotContains(t, err.Error(), testSecretMarker)
		})
	}
}

func TestUnitValidateFieldsWoJSON(t *testing.T) {
	assert.False(t, validateFieldsWoJSON(`{"a":"b"}`, nil).HasError())

	diags := validateFieldsWoJSON(`{"password":"`+testSecretMarker+`"`, nil)
	require.True(t, diags.HasError())
	for _, d := range diags {
		assert.NotContains(t, d.Summary+d.Detail, testSecretMarker)
	}

	assert.True(t, validateFieldsWoJSON(123, nil).HasError(), "non-string input must be an error")
}

func TestUnitGetFieldsWoFromRawConfig(t *testing.T) {
	t.Run("set and valid", func(t *testing.T) {
		raw := rawConfigWith(map[string]cty.Value{
			"fields_wo": cty.StringVal(`{"clientId":"id","clientSecret":"` + testSecretMarker + `"}`),
		})
		got, set, err := getFieldsWoFromRawConfig(dataWithRawConfig(raw, nil))
		require.NoError(t, err)
		assert.True(t, set)
		assert.Equal(t, map[string]string{"clientId": "id", "clientSecret": testSecretMarker}, got)
	})

	t.Run("null attribute means not set", func(t *testing.T) {
		_, set, err := getFieldsWoFromRawConfig(dataWithRawConfig(rawConfigWith(nil), nil))
		require.NoError(t, err)
		assert.False(t, set)
	})

	t.Run("unknown attribute means not set", func(t *testing.T) {
		raw := rawConfigWith(map[string]cty.Value{"fields_wo": cty.UnknownVal(cty.String)})
		_, set, err := getFieldsWoFromRawConfig(dataWithRawConfig(raw, nil))
		require.NoError(t, err)
		assert.False(t, set)
	})

	t.Run("no raw config at all (e.g. import) means not set", func(t *testing.T) {
		d := ResourceIntegrationCredential().Data(&terraform.InstanceState{ID: "cred-id"})
		_, set, err := getFieldsWoFromRawConfig(d)
		require.NoError(t, err)
		assert.False(t, set)
	})

	t.Run("invalid json is an error without the secret", func(t *testing.T) {
		raw := rawConfigWith(map[string]cty.Value{
			"fields_wo": cty.StringVal(`{"password":"` + testSecretMarker + `"`),
		})
		_, set, err := getFieldsWoFromRawConfig(dataWithRawConfig(raw, nil))
		require.Error(t, err)
		assert.False(t, set)
		assert.NotContains(t, err.Error(), testSecretMarker)
	})
}

func TestUnitBuildCredentialFieldsSourceSelection(t *testing.T) {
	sdkConfig := &platformclientv2.Configuration{}
	stateWithFields := map[string]string{
		"credential_type_name": "basicAuth",
		"fields.%":             "2",
		"fields.userName":      "old-user",
		"fields.password":      "old-password",
	}

	t.Run("fields_wo is used when set", func(t *testing.T) {
		raw := rawConfigWith(map[string]cty.Value{
			"fields_wo": cty.StringVal(`{"userName":"new-user","password":"` + testSecretMarker + `"}`),
		})
		state := map[string]string{"credential_type_name": "basicAuth"}
		got, diags := buildCredentialFields(dataWithRawConfig(raw, state), sdkConfig)
		require.False(t, diags.HasError())
		assert.Equal(t, map[string]string{"userName": "new-user", "password": testSecretMarker}, got)
	})

	t.Run("existing fields map still works when fields_wo is not set (backward compatible)", func(t *testing.T) {
		got, diags := buildCredentialFields(dataWithRawConfig(rawConfigWith(nil), stateWithFields), sdkConfig)
		require.False(t, diags.HasError())
		assert.Equal(t, map[string]string{"userName": "old-user", "password": "old-password"}, got)
	})

	t.Run("existing fields map works with no raw config at all", func(t *testing.T) {
		d := ResourceIntegrationCredential().Data(&terraform.InstanceState{ID: "cred-id", Attributes: stateWithFields})
		got, diags := buildCredentialFields(d, sdkConfig)
		require.False(t, diags.HasError())
		assert.Equal(t, map[string]string{"userName": "old-user", "password": "old-password"}, got)
	})

	t.Run("neither set yields empty fields", func(t *testing.T) {
		state := map[string]string{"credential_type_name": "basicAuth"}
		got, diags := buildCredentialFields(dataWithRawConfig(rawConfigWith(nil), state), sdkConfig)
		require.False(t, diags.HasError())
		assert.Empty(t, got)
	})

	t.Run("oauth client secret supplied through fields_wo is kept as given", func(t *testing.T) {
		raw := rawConfigWith(map[string]cty.Value{
			"fields_wo": cty.StringVal(`{"clientId":"id","clientSecret":"` + testSecretMarker + `"}`),
		})
		state := map[string]string{"credential_type_name": "pureCloudOAuthClient"}
		got, diags := buildCredentialFields(dataWithRawConfig(raw, state), sdkConfig)
		require.False(t, diags.HasError())
		assert.Equal(t, map[string]string{"clientId": "id", "clientSecret": testSecretMarker}, got)
	})

	t.Run("invalid fields_wo returns an error diagnostic without the secret", func(t *testing.T) {
		raw := rawConfigWith(map[string]cty.Value{
			"fields_wo": cty.StringVal(`{"password":"` + testSecretMarker + `"`),
		})
		state := map[string]string{"credential_type_name": "basicAuth"}
		got, diags := buildCredentialFields(dataWithRawConfig(raw, state), sdkConfig)
		require.True(t, diags.HasError())
		assert.Nil(t, got)
		for _, d := range diags {
			assert.NotContains(t, d.Summary+d.Detail, testSecretMarker)
		}
	})
}

func TestUnitCredentialChanged(t *testing.T) {
	res := ResourceIntegrationCredential()
	base := map[string]string{
		"name":                 "cred",
		"credential_type_name": "basicAuth",
		"fields_wo_version":    "1",
	}
	withDiff := func(diffAttrs map[string]*terraform.ResourceAttrDiff) *schema.ResourceData {
		state := &terraform.InstanceState{ID: "cred-id", Attributes: base}
		d, err := schema.InternalMap(res.Schema).Data(state, &terraform.InstanceDiff{Attributes: diffAttrs})
		require.NoError(t, err)
		return d
	}

	t.Run("no change does not update", func(t *testing.T) {
		assert.False(t, credentialChanged(withDiff(nil)))
	})
	t.Run("version bump alone triggers update so a rotated secret is re-sent", func(t *testing.T) {
		assert.True(t, credentialChanged(withDiff(map[string]*terraform.ResourceAttrDiff{
			"fields_wo_version": {Old: "1", New: "2"},
		})))
	})
	t.Run("name change triggers update", func(t *testing.T) {
		assert.True(t, credentialChanged(withDiff(map[string]*terraform.ResourceAttrDiff{
			"name": {Old: "cred", New: "renamed"},
		})))
	})
}

func TestUnitFieldsWoSchema(t *testing.T) {
	res := ResourceIntegrationCredential()

	t.Run("schema is valid for the SDK", func(t *testing.T) {
		require.NoError(t, res.InternalValidate(nil, true))
	})

	t.Run("fields_wo is write-only and sensitive", func(t *testing.T) {
		s := res.Schema["fields_wo"]
		require.NotNil(t, s)
		assert.Equal(t, schema.TypeString, s.Type)
		assert.True(t, s.WriteOnly)
		assert.True(t, s.Sensitive)
		assert.True(t, s.Optional)
		assert.False(t, s.Computed, "a write-only attribute cannot be computed")
		assert.Contains(t, s.ConflictsWith, "fields")
		assert.Contains(t, s.RequiredWith, "fields_wo_version")
	})

	t.Run("fields_wo_version is a stored attribute (not write-only)", func(t *testing.T) {
		s := res.Schema["fields_wo_version"]
		require.NotNil(t, s)
		assert.Equal(t, schema.TypeInt, s.Type)
		assert.False(t, s.WriteOnly, "the trigger must be stored in state to be diffed")
		assert.True(t, s.Optional)
		assert.False(t, s.Computed)
		assert.Contains(t, s.RequiredWith, "fields_wo")
	})

	t.Run("existing fields attribute is unchanged", func(t *testing.T) {
		s := res.Schema["fields"]
		require.NotNil(t, s)
		assert.Equal(t, schema.TypeMap, s.Type)
		assert.True(t, s.Optional)
		assert.True(t, s.Computed)
		assert.True(t, s.Sensitive)
		assert.False(t, s.WriteOnly)
	})
}

func TestUnitFieldsWoConfigValidation(t *testing.T) {
	res := ResourceIntegrationCredential()
	validate := func(cfg map[string]interface{}) bool {
		return res.Validate(terraform.NewResourceConfigRaw(cfg)).HasError()
	}
	base := func() map[string]interface{} {
		return map[string]interface{}{"credential_type_name": "basicAuth", "name": "cred"}
	}

	t.Run("existing fields-only config is still valid", func(t *testing.T) {
		cfg := base()
		cfg["fields"] = map[string]interface{}{"userName": "u", "password": "p"}
		assert.False(t, validate(cfg))
	})
	t.Run("fields_wo with version is valid", func(t *testing.T) {
		cfg := base()
		cfg["fields_wo"] = `{"userName":"u","password":"p"}`
		cfg["fields_wo_version"] = 1
		assert.False(t, validate(cfg))
	})
	t.Run("fields and fields_wo together conflict", func(t *testing.T) {
		cfg := base()
		cfg["fields"] = map[string]interface{}{"userName": "u"}
		cfg["fields_wo"] = `{"userName":"u"}`
		cfg["fields_wo_version"] = 1
		assert.True(t, validate(cfg))
	})
	t.Run("fields_wo without version is rejected", func(t *testing.T) {
		cfg := base()
		cfg["fields_wo"] = `{"userName":"u"}`
		assert.True(t, validate(cfg))
	})
	t.Run("version without fields_wo is rejected", func(t *testing.T) {
		cfg := base()
		cfg["fields_wo_version"] = 1
		assert.True(t, validate(cfg))
	})
	t.Run("invalid fields_wo json is rejected at validate time", func(t *testing.T) {
		cfg := base()
		cfg["fields_wo"] = `{"userName":`
		cfg["fields_wo_version"] = 1
		assert.True(t, validate(cfg))
	})
}
