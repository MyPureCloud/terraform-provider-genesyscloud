package organization_authentication_settings

import (
	"context"
	"net/http"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
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
	orgAuthProxy.getOrgAuthSettingsByIdAttr = func(ctx context.Context, o *orgAuthSettingsProxy, id string) (*platformclientv2.Orgauthsettings, *platformclientv2.APIResponse, error) {
		orgAuthSettings := &testOrgAuthSettings

		apiResponse := &platformclientv2.APIResponse{StatusCode: http.StatusOK}
		return orgAuthSettings, apiResponse, nil
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
	orgAuthProxy.getOrgAuthSettingsByIdAttr = func(ctx context.Context, p *orgAuthSettingsProxy, id string) (*platformclientv2.Orgauthsettings, *platformclientv2.APIResponse, error) {
		orgAuthSettings := &testOrgAuthSettings

		apiResponse := &platformclientv2.APIResponse{StatusCode: http.StatusOK}
		return orgAuthSettings, apiResponse, nil
	}

	orgAuthProxy.updateOrgAuthSettingsAttr = func(ctx context.Context, p *orgAuthSettingsProxy, orgAuthSettings *platformclientv2.Orgauthsettings) (*platformclientv2.Orgauthsettings, *platformclientv2.APIResponse, error) {

		equal := cmp.Equal(testOrgAuthSettings, *orgAuthSettings)
		assert.Equal(t, true, equal, "orgAuthSettings not equal to expected value in update: %s", cmp.Diff(testOrgAuthSettings, *orgAuthSettings))

		apiResponse := &platformclientv2.APIResponse{StatusCode: http.StatusOK}
		return orgAuthSettings, apiResponse, nil
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
	}
	return resourceDataMap
}
