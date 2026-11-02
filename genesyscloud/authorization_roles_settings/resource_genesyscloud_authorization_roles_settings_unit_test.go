package authorization_roles_settings

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v200/platformclientv2"
	"github.com/stretchr/testify/assert"
)

// Unit Tests

func generateAuthorizationRolesSettingsResponse(backfillEnabled bool, grantDivisionAware bool, bypass bool) roleSettings {
	return roleSettings{
		Id:                              platformclientv2.String(authorizationRolesSettingsId),
		BackfillEnabled:                 platformclientv2.Bool(backfillEnabled),
		AuthorizationGrantDivisionAware: platformclientv2.Bool(grantDivisionAware),
		GenesysOrgPolicyBypass:          platformclientv2.Bool(bypass),
	}
}

func buildAuthorizationRolesSettingsDataMap(backfillEnabled bool, grantDivisionAware bool, bypass bool) map[string]interface{} {
	return map[string]interface{}{
		"backfill_enabled":                   backfillEnabled,
		"authorization_grant_division_aware": grantDivisionAware,
		"genesys_org_policy_bypass":          bypass,
	}
}

// rawConfigObject builds the cty value the SDK would hand the provider for a config in which
// only the given attributes are set. Attributes left out are null, i.e. unconfigured.
func rawConfigObject(attrs map[string]cty.Value) cty.Value {
	full := map[string]cty.Value{
		"id":                                 cty.NullVal(cty.String),
		"backfill_enabled":                   cty.NullVal(cty.Bool),
		"authorization_grant_division_aware": cty.NullVal(cty.Bool),
		"genesys_org_policy_bypass":          cty.NullVal(cty.Bool),
	}
	for k, v := range attrs {
		full[k] = v
	}
	return cty.ObjectVal(full)
}

func TestUnitMergeAuthorizationRolesSettingsPreservesUnconfiguredAttributes(t *testing.T) {
	// The organization currently has both flags on. The practitioner manages only
	// backfill_enabled and turns it off. authorization_grant_division_aware must survive
	// the PUT rather than being sent as the zero value.
	current := generateAuthorizationRolesSettingsResponse(true, true, true)

	merged := mergeAuthorizationRolesSettings(
		rawConfigObject(map[string]cty.Value{"backfill_enabled": cty.False}),
		&current,
	)

	assert.NotNil(t, merged.BackfillEnabled)
	assert.Equal(t, false, *merged.BackfillEnabled, "explicitly configured false should be sent")
	assert.NotNil(t, merged.AuthorizationGrantDivisionAware)
	assert.Equal(t, true, *merged.AuthorizationGrantDivisionAware, "unconfigured attribute must keep its current value")
	assert.NotNil(t, merged.GenesysOrgPolicyBypass)
	assert.Equal(t, true, *merged.GenesysOrgPolicyBypass, "unconfigured genesys_org_policy_bypass must not be cleared")
}

func TestUnitMergeAuthorizationRolesSettingsExplicitTrueOverridesCurrent(t *testing.T) {
	current := generateAuthorizationRolesSettingsResponse(false, false, true)

	merged := mergeAuthorizationRolesSettings(
		rawConfigObject(map[string]cty.Value{
			"backfill_enabled":                   cty.True,
			"authorization_grant_division_aware": cty.True,
		}),
		&current,
	)

	assert.Equal(t, true, *merged.BackfillEnabled)
	assert.Equal(t, true, *merged.AuthorizationGrantDivisionAware)
	assert.Equal(t, true, *merged.GenesysOrgPolicyBypass, "bypass stays at its current value when not configured")
}

func TestUnitMergeAuthorizationRolesSettingsEmptyConfigSendsCurrent(t *testing.T) {
	// A config that declares the resource but sets nothing must be a no-op write, not a
	// reset of every flag to false.
	current := generateAuthorizationRolesSettingsResponse(true, true, true)

	merged := mergeAuthorizationRolesSettings(rawConfigObject(nil), &current)

	assert.Equal(t, true, *merged.BackfillEnabled)
	assert.Equal(t, true, *merged.AuthorizationGrantDivisionAware)
	assert.Equal(t, true, *merged.GenesysOrgPolicyBypass)
}

func TestUnitMergeAuthorizationRolesSettingsNullConfigAndNoCurrentSendsNothing(t *testing.T) {
	// Import and similar operations run without a configuration. Nothing is known, so no
	// field should be populated and omitempty leaves them out of the request body.
	merged := mergeAuthorizationRolesSettings(cty.NilVal, nil)

	assert.Nil(t, merged.BackfillEnabled)
	assert.Nil(t, merged.AuthorizationGrantDivisionAware)
	assert.Nil(t, merged.GenesysOrgPolicyBypass)
}

func TestUnitConfiguredBoolIgnoresUnknownAndMissingAttributes(t *testing.T) {
	// Unknown values appear when an attribute is derived from something not yet applied.
	_, ok := configuredBool(rawConfigObject(map[string]cty.Value{
		"backfill_enabled": cty.UnknownVal(cty.Bool),
	}), "backfill_enabled")
	assert.False(t, ok, "unknown value should not be treated as configured")

	_, ok = configuredBool(rawConfigObject(nil), "not_an_attribute")
	assert.False(t, ok, "absent attribute should not panic or report configured")

	_, ok = configuredBool(cty.NullVal(cty.Object(map[string]cty.Type{
		"backfill_enabled": cty.Bool,
	})), "backfill_enabled")
	assert.False(t, ok, "null config object should not report configured")
}

func TestUnitResourceAuthorizationRolesSettingsRead(t *testing.T) {
	testResponse := generateAuthorizationRolesSettingsResponse(true, false, true)

	proxy := &authorizationRolesSettingsProxy{}
	proxy.getAuthorizationRolesSettingsAttr = func(ctx context.Context, p *authorizationRolesSettingsProxy) (*roleSettings, *platformclientv2.APIResponse, error) {
		resp := testResponse
		return &resp, &platformclientv2.APIResponse{StatusCode: http.StatusOK}, nil
	}
	internalProxy = proxy
	defer func() { internalProxy = nil }()

	ctx := context.Background()
	gcloud := &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}

	d := schema.TestResourceDataRaw(t, ResourceAuthorizationRolesSettings().Schema, buildAuthorizationRolesSettingsDataMap(true, false, true))
	d.SetId(authorizationRolesSettingsId)

	diag := readAuthorizationRolesSettings(ctx, d, gcloud)
	assert.Equal(t, false, diag.HasError())
	assert.Equal(t, authorizationRolesSettingsId, d.Id())
	assert.Equal(t, true, d.Get("backfill_enabled").(bool))
	assert.Equal(t, false, d.Get("authorization_grant_division_aware").(bool))
	assert.Equal(t, true, d.Get("genesys_org_policy_bypass").(bool))
}

func TestUnitResourceAuthorizationRolesSettingsUpdateReadsBeforeWriting(t *testing.T) {
	// The update path must GET first and seed the PUT body from that response. With no raw
	// config available in unit tests every attribute reads as unconfigured, so the captured
	// body should be exactly the current settings.
	testResponse := generateAuthorizationRolesSettingsResponse(true, true, true)

	var capturedPutBody *roleSettings
	getCalls := 0

	proxy := &authorizationRolesSettingsProxy{}
	proxy.getAuthorizationRolesSettingsAttr = func(ctx context.Context, p *authorizationRolesSettingsProxy) (*roleSettings, *platformclientv2.APIResponse, error) {
		getCalls++
		resp := testResponse
		return &resp, &platformclientv2.APIResponse{StatusCode: http.StatusOK}, nil
	}
	proxy.updateAuthorizationRolesSettingsAttr = func(ctx context.Context, p *authorizationRolesSettingsProxy, settings *roleSettings) (*roleSettings, *platformclientv2.APIResponse, error) {
		capturedPutBody = settings
		resp := testResponse
		return &resp, &platformclientv2.APIResponse{StatusCode: http.StatusOK}, nil
	}
	internalProxy = proxy
	defer func() { internalProxy = nil }()

	ctx := context.Background()
	gcloud := &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}

	d := schema.TestResourceDataRaw(t, ResourceAuthorizationRolesSettings().Schema, buildAuthorizationRolesSettingsDataMap(false, false, true))
	d.SetId(authorizationRolesSettingsId)

	diag := updateAuthorizationRolesSettings(ctx, d, gcloud)
	assert.Equal(t, false, diag.HasError())
	assert.GreaterOrEqual(t, getCalls, 1, "update must read current settings before writing")

	assert.NotNil(t, capturedPutBody)
	assert.NotNil(t, capturedPutBody.BackfillEnabled)
	assert.Equal(t, true, *capturedPutBody.BackfillEnabled)
	assert.NotNil(t, capturedPutBody.AuthorizationGrantDivisionAware)
	assert.Equal(t, true, *capturedPutBody.AuthorizationGrantDivisionAware)
	assert.NotNil(t, capturedPutBody.GenesysOrgPolicyBypass)
	assert.Equal(t, true, *capturedPutBody.GenesysOrgPolicyBypass, "bypass must be carried into the PUT body, not dropped")
}

func TestUnitResourceAuthorizationRolesSettingsUpdateAbortsWhenReadFails(t *testing.T) {
	// Without the current settings the PUT body cannot be built safely, so the update must
	// fail rather than write a partially populated object.
	updateCalled := false

	proxy := &authorizationRolesSettingsProxy{}
	proxy.getAuthorizationRolesSettingsAttr = func(ctx context.Context, p *authorizationRolesSettingsProxy) (*roleSettings, *platformclientv2.APIResponse, error) {
		return nil, &platformclientv2.APIResponse{StatusCode: http.StatusForbidden}, fmt.Errorf("403 forbidden")
	}
	proxy.updateAuthorizationRolesSettingsAttr = func(ctx context.Context, p *authorizationRolesSettingsProxy, settings *roleSettings) (*roleSettings, *platformclientv2.APIResponse, error) {
		updateCalled = true
		return nil, nil, nil
	}
	internalProxy = proxy
	defer func() { internalProxy = nil }()

	ctx := context.Background()
	gcloud := &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}

	d := schema.TestResourceDataRaw(t, ResourceAuthorizationRolesSettings().Schema, buildAuthorizationRolesSettingsDataMap(true, true, true))
	d.SetId(authorizationRolesSettingsId)

	diag := updateAuthorizationRolesSettings(ctx, d, gcloud)
	assert.Equal(t, true, diag.HasError())
	assert.False(t, updateCalled, "no PUT should be attempted when the preceding read fails")
}

func TestUnitDataSourceAuthorizationRolesSettingsRead(t *testing.T) {
	testResponse := generateAuthorizationRolesSettingsResponse(false, true, true)

	proxy := &authorizationRolesSettingsProxy{}
	proxy.getAuthorizationRolesSettingsAttr = func(ctx context.Context, p *authorizationRolesSettingsProxy) (*roleSettings, *platformclientv2.APIResponse, error) {
		resp := testResponse
		return &resp, &platformclientv2.APIResponse{StatusCode: http.StatusOK}, nil
	}
	internalProxy = proxy
	defer func() { internalProxy = nil }()

	ctx := context.Background()
	gcloud := &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}

	d := schema.TestResourceDataRaw(t, DataSourceAuthorizationRolesSettings().Schema, map[string]interface{}{})

	diag := dataSourceAuthorizationRolesSettingsRead(ctx, d, gcloud)
	assert.Equal(t, false, diag.HasError())
	assert.Equal(t, authorizationRolesSettingsId, d.Id())
	assert.Equal(t, false, d.Get("backfill_enabled").(bool))
	assert.Equal(t, true, d.Get("authorization_grant_division_aware").(bool))
	assert.Equal(t, true, d.Get("genesys_org_policy_bypass").(bool))
}

func TestUnitMergeAuthorizationRolesSettingsHonoursExplicitBypassFalse(t *testing.T) {
	// Turning the bypass off is a legitimate, if consequential, choice. An explicit false in
	// configuration must reach the request body rather than being mistaken for "unset".
	current := generateAuthorizationRolesSettingsResponse(true, true, true)

	merged := mergeAuthorizationRolesSettings(
		rawConfigObject(map[string]cty.Value{"genesys_org_policy_bypass": cty.False}),
		&current,
	)

	assert.NotNil(t, merged.GenesysOrgPolicyBypass)
	assert.Equal(t, false, *merged.GenesysOrgPolicyBypass)
	assert.Equal(t, true, *merged.BackfillEnabled, "other attributes stay as they were")
	assert.Equal(t, true, *merged.AuthorizationGrantDivisionAware, "other attributes stay as they were")
}

func TestUnitRoleSettingsJsonWireFormat(t *testing.T) {
	// The model is declared in this package rather than taken from the generated SDK, so the
	// field names on the wire are worth pinning down explicitly.
	body, err := json.Marshal(roleSettings{
		BackfillEnabled:                 platformclientv2.Bool(true),
		AuthorizationGrantDivisionAware: platformclientv2.Bool(false),
		GenesysOrgPolicyBypass:          platformclientv2.Bool(false),
	})
	assert.NoError(t, err)

	var decoded map[string]interface{}
	assert.NoError(t, json.Unmarshal(body, &decoded))

	assert.Equal(t, true, decoded["backfillEnabled"])
	assert.Equal(t, false, decoded["authorizationGrantDivisionAware"])
	assert.Equal(t, false, decoded["genesysOrgPolicyBypass"], "explicit false must serialize, not be dropped by omitempty")
	assert.NotContains(t, decoded, "id", "unset fields are omitted")
	assert.NotContains(t, decoded, "selfUri", "read-only fields are omitted when unset")

	// And the response shape the API returns must decode back into the model.
	var parsed roleSettings
	assert.NoError(t, json.Unmarshal([]byte(`{
		"id": "settings",
		"backfillEnabled": false,
		"authorizationGrantDivisionAware": true,
		"genesysOrgPolicyBypass": true,
		"selfUri": "/api/v2/authorization/roles/settings"
	}`), &parsed))
	assert.Equal(t, "settings", *parsed.Id)
	assert.Equal(t, false, *parsed.BackfillEnabled)
	assert.Equal(t, true, *parsed.AuthorizationGrantDivisionAware)
	assert.Equal(t, true, *parsed.GenesysOrgPolicyBypass)
}

func TestUnitAuthorizationRolesSettingsSchemasAreValid(t *testing.T) {
	// Catches schema mistakes the compiler cannot, such as an attribute that is Required and
	// Computed at once, or a data source attribute left writable.
	if err := ResourceAuthorizationRolesSettings().InternalValidate(nil, true); err != nil {
		t.Fatalf("resource schema failed InternalValidate: %s", err)
	}
	if err := DataSourceAuthorizationRolesSettings().InternalValidate(nil, false); err != nil {
		t.Fatalf("data source schema failed InternalValidate: %s", err)
	}

	// All three attributes must be Optional+Computed. Computed is what lets Terraform carry
	// the organization's current value for anything the practitioner has not configured,
	// which is half of the protection against the full-replace PUT.
	for _, attr := range []string{"backfill_enabled", "authorization_grant_division_aware", "genesys_org_policy_bypass"} {
		s, ok := ResourceAuthorizationRolesSettings().Schema[attr]
		if !ok {
			t.Fatalf("resource schema is missing %s", attr)
		}
		assert.True(t, s.Optional, "%s should be Optional", attr)
		assert.True(t, s.Computed, "%s should be Computed", attr)
	}
}

func TestUnitResourceAuthorizationRolesSettingsCreateReadsBeforeWriting(t *testing.T) {
	// Create delegates to update so that a first apply performs the same GET-merge-PUT as any
	// other apply. This is the riskiest moment for genesys_org_policy_bypass: there is no prior
	// state, so an unconfigured bool would otherwise be sent as false. Assert the organization's
	// current value is read and carried into the PUT, and that the singleton ID is assigned.
	testResponse := generateAuthorizationRolesSettingsResponse(true, true, true)

	var capturedPutBody *roleSettings
	var callOrder []string

	proxy := &authorizationRolesSettingsProxy{}
	proxy.getAuthorizationRolesSettingsAttr = func(ctx context.Context, p *authorizationRolesSettingsProxy) (*roleSettings, *platformclientv2.APIResponse, error) {
		callOrder = append(callOrder, "GET")
		resp := testResponse
		return &resp, &platformclientv2.APIResponse{StatusCode: http.StatusOK}, nil
	}
	proxy.updateAuthorizationRolesSettingsAttr = func(ctx context.Context, p *authorizationRolesSettingsProxy, settings *roleSettings) (*roleSettings, *platformclientv2.APIResponse, error) {
		callOrder = append(callOrder, "PUT")
		capturedPutBody = settings
		resp := testResponse
		return &resp, &platformclientv2.APIResponse{StatusCode: http.StatusOK}, nil
	}
	internalProxy = proxy
	defer func() { internalProxy = nil }()

	ctx := context.Background()
	gcloud := &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}

	// No ID yet: this is a create, not an update of something already in state.
	d := schema.TestResourceDataRaw(t, ResourceAuthorizationRolesSettings().Schema, buildAuthorizationRolesSettingsDataMap(false, false, false))

	diag := createAuthorizationRolesSettings(ctx, d, gcloud)
	assert.Equal(t, false, diag.HasError())
	assert.Equal(t, authorizationRolesSettingsId, d.Id(), "create must assign the singleton ID")

	assert.GreaterOrEqual(t, len(callOrder), 2)
	assert.Equal(t, "GET", callOrder[0], "create must read before it writes")
	assert.Equal(t, "PUT", callOrder[1])

	assert.NotNil(t, capturedPutBody)
	assert.NotNil(t, capturedPutBody.GenesysOrgPolicyBypass)
	assert.Equal(t, true, *capturedPutBody.GenesysOrgPolicyBypass,
		"an unconfigured bypass must be carried from the GET, not sent as false on first apply")
}

func TestUnitResourceAuthorizationRolesSettingsCreateAbortsWhenReadFails(t *testing.T) {
	updateCalled := false

	proxy := &authorizationRolesSettingsProxy{}
	proxy.getAuthorizationRolesSettingsAttr = func(ctx context.Context, p *authorizationRolesSettingsProxy) (*roleSettings, *platformclientv2.APIResponse, error) {
		return nil, &platformclientv2.APIResponse{StatusCode: http.StatusForbidden}, fmt.Errorf("403 forbidden")
	}
	proxy.updateAuthorizationRolesSettingsAttr = func(ctx context.Context, p *authorizationRolesSettingsProxy, settings *roleSettings) (*roleSettings, *platformclientv2.APIResponse, error) {
		updateCalled = true
		return nil, nil, nil
	}
	internalProxy = proxy
	defer func() { internalProxy = nil }()

	ctx := context.Background()
	gcloud := &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}

	d := schema.TestResourceDataRaw(t, ResourceAuthorizationRolesSettings().Schema, buildAuthorizationRolesSettingsDataMap(true, true, true))

	diag := createAuthorizationRolesSettings(ctx, d, gcloud)
	assert.Equal(t, true, diag.HasError())
	assert.False(t, updateCalled, "no PUT should be attempted when the preceding read fails")
}
