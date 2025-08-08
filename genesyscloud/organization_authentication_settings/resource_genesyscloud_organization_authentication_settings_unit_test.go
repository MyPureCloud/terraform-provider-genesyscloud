package organization_authentication_settings

import (
	"context"
	"net/http"
	"reflect"
	"testing"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v165/platformclientv2"
	"github.com/stretchr/testify/assert"
)

// Unit Test

func generateAuthSettingsData(domainAllowList []string, ipAllowList []string) platformclientv2.Orgauthsettings {
	passwordRequirements := &platformclientv2.Passwordrequirements{
		MinimumLength:     platformclientv2.Int(8),
		MinimumDigits:     platformclientv2.Int(2),
		MinimumLetters:    platformclientv2.Int(2),
		MinimumUpper:      platformclientv2.Int(1),
		MinimumLower:      platformclientv2.Int(1),
		MinimumSpecials:   platformclientv2.Int(1),
		MinimumAgeSeconds: platformclientv2.Int(1),
		ExpirationDays:    platformclientv2.Int(90),
	}

	return platformclientv2.Orgauthsettings{
		MultifactorAuthenticationRequired: platformclientv2.Bool(true),
		DomainAllowlistEnabled:            platformclientv2.Bool(false),
		DomainAllowlist:                   &domainAllowList,
		IpAddressAllowlist:                &ipAllowList,
		PasswordRequirements:              passwordRequirements,
	}
}

func generateTimeOutSettings() *platformclientv2.Idletokentimeout {
	return &platformclientv2.Idletokentimeout{
		EnableIdleTokenTimeout:  platformclientv2.Bool(true),
		IdleTokenTimeoutSeconds: platformclientv2.Int(3000),
	}

}

func TestUnitResourceOrganizationAuthenticationSettingsRead(t *testing.T) {
	tId := uuid.NewString()
	domainAllowList := []string{"Genesys.com", "Google.com"}
	allowList := make([]interface{}, len(domainAllowList))
	for i, v := range domainAllowList {
		allowList[i] = v
	}
	ipAllowList := []string{"0.0.0.0/32", "1.1.1.1/32"}
	ipAddressAllowList := make([]interface{}, len(ipAllowList))
	for i, v := range ipAllowList {
		ipAddressAllowList[i] = v
	}
	testOrgAuthSettings := generateAuthSettingsData(domainAllowList, ipAllowList)
	orgAuthProxy := &orgAuthSettingsProxy{}
	orgAuthProxy.getOrgAuthSettingsAttr = func(ctx context.Context, o *orgAuthSettingsProxy) (*platformclientv2.Orgauthsettings, *platformclientv2.APIResponse, error) {
		orgAuthSettings := &testOrgAuthSettings

		apiResponse := &platformclientv2.APIResponse{StatusCode: http.StatusOK}
		return orgAuthSettings, apiResponse, nil
	}
	timeOutSettings := generateTimeOutSettings()
	orgAuthProxy.getTokensTimeOutSettingsAttr = func(ctx context.Context, p *orgAuthSettingsProxy) (*platformclientv2.Idletokentimeout, *platformclientv2.APIResponse, error) {
		apiResponse := &platformclientv2.APIResponse{StatusCode: http.StatusOK}
		return timeOutSettings, apiResponse, nil
	}
	internalProxy = orgAuthProxy
	defer func() { internalProxy = nil }()

	ctx := context.Background()
	gcloud := &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}
	// Get defined Schema
	resourceSchema := ResourceOrganizationAuthenticationSettings().Schema
	//Setup map of values
	resourceDataMap := buildOrgAuthSettingsDataMap(*testOrgAuthSettings.MultifactorAuthenticationRequired, *testOrgAuthSettings.DomainAllowlistEnabled, allowList, ipAddressAllowList, *testOrgAuthSettings.PasswordRequirements)
	d := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
	d.SetId(tId)

	diag := readOrganizationAuthenticationSettings(ctx, d, gcloud)
	assert.Equal(t, tId, d.Id())
	assert.Equal(t, false, diag.HasError())

	authSettings := getOrganizationAuthenticationSettingsFromResourceData(d)
	equal := cmp.Equal(testOrgAuthSettings, authSettings)
	assert.Equal(t, true, equal, "Password Requirements not equal to expected value in read: %s", cmp.Diff(testOrgAuthSettings, authSettings))
}

func TestUnitResourceOrganizationAuthenticationSettingsUpdate(t *testing.T) {
	tId := uuid.NewString()
	domainAllowList := []string{"Genesys.ie", "Updated.com"}
	allowList := make([]interface{}, len(domainAllowList))
	for i, v := range domainAllowList {
		allowList[i] = v
	}
	ipAllowList := []string{"2.2.2.2/32", "3.3.3.3/32"}
	ipAddressAllowList := make([]interface{}, len(ipAllowList))
	for i, v := range ipAllowList {
		ipAddressAllowList[i] = v
	}
	testOrgAuthSettings := generateAuthSettingsData(domainAllowList, ipAllowList)

	orgAuthProxy := &orgAuthSettingsProxy{}
	orgAuthProxy.getOrgAuthSettingsAttr = func(ctx context.Context, p *orgAuthSettingsProxy) (*platformclientv2.Orgauthsettings, *platformclientv2.APIResponse, error) {
		orgAuthSettings := &testOrgAuthSettings

		apiResponse := &platformclientv2.APIResponse{StatusCode: http.StatusOK}
		return orgAuthSettings, apiResponse, nil
	}

	timeOutSettings := generateTimeOutSettings()
	orgAuthProxy.getTokensTimeOutSettingsAttr = func(ctx context.Context, p *orgAuthSettingsProxy) (*platformclientv2.Idletokentimeout, *platformclientv2.APIResponse, error) {
		apiResponse := &platformclientv2.APIResponse{StatusCode: http.StatusOK}
		return timeOutSettings, apiResponse, nil
	}
	orgAuthProxy.updateOrgAuthSettingsAttr = func(ctx context.Context, p *orgAuthSettingsProxy, orgAuthSettings *platformclientv2.Orgauthsettings) (*platformclientv2.Orgauthsettings, *platformclientv2.APIResponse, error) {

		equal := cmp.Equal(testOrgAuthSettings, *orgAuthSettings)
		assert.Equal(t, true, equal, "orgAuthSettings not equal to expected value in update: %s", cmp.Diff(testOrgAuthSettings, *orgAuthSettings))

		apiResponse := &platformclientv2.APIResponse{StatusCode: http.StatusOK}
		return orgAuthSettings, apiResponse, nil
	}
	orgAuthProxy.updateTokensTimeOutSettingsAttr = func(ctx context.Context, p *orgAuthSettingsProxy, idletimeout *platformclientv2.Idletokentimeout) (*platformclientv2.Idletokentimeout, *platformclientv2.APIResponse, error) {

		equal := cmp.Equal(timeOutSettings, *idletimeout)
		assert.Equal(t, false, equal, "timeout settings not equal to expected value in update: %s", cmp.Diff(timeOutSettings, *idletimeout))

		apiResponse := &platformclientv2.APIResponse{StatusCode: http.StatusOK}
		return timeOutSettings, apiResponse, nil
	}
	internalProxy = orgAuthProxy
	defer func() { internalProxy = nil }()

	ctx := context.Background()
	gcloud := &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}

	resourceSchema := ResourceOrganizationAuthenticationSettings().Schema

	resourceDataMap := buildOrgAuthSettingsDataMap(*testOrgAuthSettings.MultifactorAuthenticationRequired, *testOrgAuthSettings.DomainAllowlistEnabled, allowList, ipAddressAllowList, *testOrgAuthSettings.PasswordRequirements)

	d := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
	d.SetId(tId)

	diag := updateOrganizationAuthenticationSettings(ctx, d, gcloud)
	assert.Equal(t, false, diag.HasError())
	assert.Equal(t, tId, d.Id())
	assert.Equal(t, *testOrgAuthSettings.DomainAllowlist, domainAllowList)
}

func buildOrgAuthSettingsDataMap(tMultifactorAuthRequired bool, tDomainAllowListEnabled bool, tDomainAllowList []interface{}, tIpAddressAllowList []interface{}, tPasswordRequirements platformclientv2.Passwordrequirements) map[string]interface{} {
	resourceDataMap := map[string]interface{}{
		"multifactor_authentication_required": tMultifactorAuthRequired,
		"domain_allowlist_enabled":            tDomainAllowListEnabled,
		"domain_allowlist":                    tDomainAllowList,
		"ip_address_allowlist":                tIpAddressAllowList,
		"password_requirements":               flattenPasswordRequirements(&tPasswordRequirements),
		"timeout_settings":                    flattenTimeOutSettings(generateTimeOutSettings()),
	}
	return resourceDataMap
}

func TestGetTimeOutSettingsFromResourceData(t *testing.T) {
	// Test cases
	tests := []struct {
		name     string
		input    map[string]interface{}
		expected *platformclientv2.Idletokentimeout
	}{
		{
			name: "Valid timeout settings",
			input: map[string]interface{}{
				"timeout_settings": []interface{}{
					map[string]interface{}{
						"enable_idle_token_timeout":  true,
						"idle_token_timeout_seconds": 3600,
					},
				},
			},
			expected: &platformclientv2.Idletokentimeout{
				EnableIdleTokenTimeout:  platformclientv2.Bool(true),
				IdleTokenTimeoutSeconds: platformclientv2.Int(3600),
			},
		},
		{
			name:     "Nil timeout settings",
			input:    map[string]interface{}{"timeout_settings": nil},
			expected: nil,
		},
		{
			name:     "Empty timeout settings",
			input:    map[string]interface{}{"timeout_settings": []interface{}{}},
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new resource data object
			d := schema.TestResourceDataRaw(t, map[string]*schema.Schema{
				"timeout_settings": {
					Type:     schema.TypeList,
					Optional: true,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"enable_idle_token_timeout": {
								Type:     schema.TypeBool,
								Required: true,
							},
							"idle_token_timeout_seconds": {
								Type:     schema.TypeInt,
								Required: true,
							},
						},
					},
				},
			}, tt.input)

			// Call the function
			result := getTimeOutSettingsFromResourceData(d)

			// Check the results
			if tt.expected == nil {
				if result != nil {
					t.Errorf("expected nil, got %v", result)
				}
				return
			}

			if result == nil {
				t.Fatal("expected non-nil result")
			}

			if *result.EnableIdleTokenTimeout != *tt.expected.EnableIdleTokenTimeout {
				t.Errorf("EnableIdleTokenTimeout: expected %v, got %v",
					*tt.expected.EnableIdleTokenTimeout,
					*result.EnableIdleTokenTimeout)
			}

			if *result.IdleTokenTimeoutSeconds != *tt.expected.IdleTokenTimeoutSeconds {
				t.Errorf("IdleTokenTimeoutSeconds: expected %v, got %v",
					*tt.expected.IdleTokenTimeoutSeconds,
					*result.IdleTokenTimeoutSeconds)
			}
		})
	}
}

func TestFlattenTimeOutSettings(t *testing.T) {
	// Test cases
	tests := []struct {
		name     string
		input    *platformclientv2.Idletokentimeout
		expected []interface{}
	}{
		{
			name: "Valid timeout settings",
			input: &platformclientv2.Idletokentimeout{
				EnableIdleTokenTimeout:  platformclientv2.Bool(true),
				IdleTokenTimeoutSeconds: platformclientv2.Int(3600),
			},
			expected: []interface{}{
				map[string]interface{}{
					"enable_idle_token_timeout":  true,
					"idle_token_timeout_seconds": 3600,
				},
			},
		},
		{
			name:     "Nil input",
			input:    nil,
			expected: nil,
		},
		{
			name: "Empty values",
			input: &platformclientv2.Idletokentimeout{
				EnableIdleTokenTimeout:  nil,
				IdleTokenTimeoutSeconds: nil,
			},
			expected: []interface{}{
				map[string]interface{}{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Call the function
			result := flattenTimeOutSettings(tt.input)

			// Check if result is nil
			if tt.expected == nil {
				if result != nil {
					t.Errorf("expected nil, got %v", result)
				}
				return
			}

			// Check if result is empty
			if len(result) == 0 {
				if len(tt.expected) != 0 {
					t.Errorf("expected non-empty result, got empty")
				}
				return
			}

			// Get the first map from the result
			resultMap := result[0].(map[string]interface{})
			expectedMap := tt.expected[0].(map[string]interface{})

			// Compare the maps
			for key, expectedValue := range expectedMap {
				if resultValue, ok := resultMap[key]; !ok {
					t.Errorf("missing key %s in result", key)
				} else if !reflect.DeepEqual(resultValue, expectedValue) {
					t.Errorf("for key %s: expected %v, got %v", key, expectedValue, resultValue)
				}
			}
		})
	}
}
