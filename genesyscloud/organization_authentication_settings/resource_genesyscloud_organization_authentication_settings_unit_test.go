package organization_authentication_settings

import (
	"context"
	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v121/platformclientv2"
	"github.com/stretchr/testify/assert"
	"net/http"
	gcloud2 "terraform-provider-genesyscloud/genesyscloud"
	"testing"
)

// Unit Test

func TestUnitResourceOrganizationAuthenticationSettingsRead(t *testing.T) {
	tMultifactorAuthRequired := true
	tDomainAllowListEnabled := false
	tiDomainAllowList := []string{"Genesys.com", "Google.com"}
	tDomainAllowList := make([]interface{}, len(tiDomainAllowList))
	for i, v := range tiDomainAllowList {
		tDomainAllowList[i] = v
	}
	IpAddressAllowList := []string{"0.0.0.0/32", "1.1.1.1/32"}
	tIpAddressAllowList := make([]interface{}, len(tiDomainAllowList))
	for i, v := range IpAddressAllowList {
		tIpAddressAllowList[i] = v
	}

	tMinimumLength := 8
	tMinimumDigits := 2
	tMinimumLetters := 2
	tMinimumUpper := 1
	tMinimumLower := 1
	tMinimumSpecials := 1
	tMinimumAgeSeconds := 1
	tExpirationDays := 90

	pReq := &platformclientv2.Passwordrequirements{
		MinimumLength:     &tMinimumLength,
		MinimumDigits:     &tMinimumDigits,
		MinimumLetters:    &tMinimumLetters,
		MinimumUpper:      &tMinimumUpper,
		MinimumLower:      &tMinimumLower,
		MinimumSpecials:   &tMinimumSpecials,
		MinimumAgeSeconds: &tMinimumAgeSeconds,
		ExpirationDays:    &tExpirationDays,
	}

	orgAuthProxy := &orgAuthSettingsProxy{}
	orgAuthProxy.getOrgAuthSettingsByIdAttr = func(ctx context.Context, o *orgAuthSettingsProxy, id string) (*platformclientv2.Orgauthsettings, int, error) {
		orgAuthSettings := &platformclientv2.Orgauthsettings{
			MultifactorAuthenticationRequired: &tMultifactorAuthRequired,
			DomainAllowlistEnabled:            &tDomainAllowListEnabled,
			DomainAllowlist:                   &tiDomainAllowList,
			IpAddressAllowlist:                &IpAddressAllowList,
			PasswordRequirements:              pReq,
		}

		apiResponse := &platformclientv2.APIResponse{StatusCode: http.StatusOK}
		return orgAuthSettings, apiResponse.StatusCode, nil
	}
	internalProxy = orgAuthProxy
	defer func() { internalProxy = nil }()

	ctx := context.Background()
	gcloud := &gcloud2.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}

	resourceSchema := ResourceOrganizationAuthenticationSettings().Schema

	resourceDataMap := buildOrgAuthSettingsDataMap(tMultifactorAuthRequired, tDomainAllowListEnabled, tDomainAllowList, tIpAddressAllowList, pReq)

	d := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)

	diag := readOrganizationAuthenticationSettings(ctx, d, gcloud)
	assert.Equal(t, false, diag.HasError())
	assert.Equal(t, tMultifactorAuthRequired, d.Get("multifactor_authentication_required").(bool))
	assert.Equal(t, tDomainAllowListEnabled, d.Get("domain_allowlist_enabled").(bool))
	assert.Equal(t, tDomainAllowList, d.Get("domain_allowlist").([]interface{}))
	assert.Equal(t, tIpAddressAllowList, d.Get("ip_address_allowlist").([]interface{}))
	equal := cmp.Equal(pReq, d.Get("password_requirements").([]interface{}))
	assert.Equal(t, true, equal, "Password Requirements not equal to expected value in read: %s", cmp.Diff(pReq, d.Get("password_requirements").([]interface{})))
}

func TestUnitResourceOrganizationAuthenticationSettingsUpdate(t *testing.T) {
	tMultifactorAuthRequired := true
	tDomainAllowListEnabled := false
	tiDomainAllowList := []string{"Genesys.com", "Google.com"}
	tDomainAllowList := make([]interface{}, len(tiDomainAllowList))
	for i, v := range tiDomainAllowList {
		tDomainAllowList[i] = v
	}
	IpAddressAllowList := []string{"0.0.0.0/32", "1.1.1.1/32"}
	tIpAddressAllowList := make([]interface{}, len(tiDomainAllowList))
	for i, v := range IpAddressAllowList {
		tIpAddressAllowList[i] = v
	}

	tMinimumLength := 6
	tMinimumDigits := 3
	tMinimumLetters := 4
	tMinimumUpper := 2
	tMinimumLower := 2
	tMinimumSpecials := 2
	tMinimumAgeSeconds := 1
	tExpirationDays := 90

	pReq := &platformclientv2.Passwordrequirements{
		MinimumLength:     &tMinimumLength,
		MinimumDigits:     &tMinimumDigits,
		MinimumLetters:    &tMinimumLetters,
		MinimumUpper:      &tMinimumUpper,
		MinimumLower:      &tMinimumLower,
		MinimumSpecials:   &tMinimumSpecials,
		MinimumAgeSeconds: &tMinimumAgeSeconds,
		ExpirationDays:    &tExpirationDays,
	}

	orgAuthProxy := &orgAuthSettingsProxy{}
	orgAuthProxy.getOrgAuthSettingsByIdAttr = func(ctx context.Context, p *orgAuthSettingsProxy, id string) (orgAuthSettings *platformclientv2.Orgauthsettings, responseCode int, err error) {
		oAuthSettings := &platformclientv2.Orgauthsettings{
			MultifactorAuthenticationRequired: &tMultifactorAuthRequired,
			DomainAllowlistEnabled:            &tDomainAllowListEnabled,
			DomainAllowlist:                   &tiDomainAllowList,
			IpAddressAllowlist:                &IpAddressAllowList,
			PasswordRequirements:              pReq,
		}
		apiResponse := &platformclientv2.APIResponse{StatusCode: http.StatusOK}
		return oAuthSettings, apiResponse.StatusCode, nil
	}

	orgAuthProxy.updateOrgAuthSettingsAttr = func(ctx context.Context, p *orgAuthSettingsProxy, orgAuthSettings *platformclientv2.Orgauthsettings) (*platformclientv2.Orgauthsettings, int, error) {
		assert.Equal(t, tMultifactorAuthRequired, *orgAuthSettings.MultifactorAuthenticationRequired, "orgAuthSettings.MultifactorAuthenticationRequired check failed in updateOrgAuthSettingsAttr")
		assert.Equal(t, tDomainAllowListEnabled, *orgAuthSettings.DomainAllowlistEnabled, "orgAuthSettings.DomainAllowlistEnabled check failed in updateOrgAuthSettingsAttr")
		assert.Equal(t, tDomainAllowList, *orgAuthSettings.DomainAllowlist, "orgAuthSettings.DomainAllowlist check failed in updateOrgAuthSettingsAttr")
		assert.Equal(t, tIpAddressAllowList, *orgAuthSettings.IpAddressAllowlist, "orgAuthSettings.IpAddressAllowlist check failed in updateOrgAuthSettingsAttr")
		assert.Equal(t, pReq, *orgAuthSettings.PasswordRequirements, "orgAuthSettings.PasswordRequirements check failed in updateOrgAuthSettingsAttr")

		apiResponse := &platformclientv2.APIResponse{StatusCode: http.StatusOK}
		return orgAuthSettings, apiResponse.StatusCode, nil
	}

	internalProxy = orgAuthProxy
	defer func() { internalProxy = nil }()

	ctx := context.Background()
	gcloud := &gcloud2.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}

	resourceSchema := ResourceOrganizationAuthenticationSettings().Schema

	resourceDataMap := buildOrgAuthSettingsDataMap(tMultifactorAuthRequired, tDomainAllowListEnabled, tDomainAllowList, tIpAddressAllowList, pReq)

	d := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)

	diag := updateOrganizationAuthenticationSettings(ctx, d, gcloud)
	assert.Equal(t, false, diag.HasError())
}

func buildOrgAuthSettingsDataMap(tMultifactorAuthRequired bool, tDomainAllowListEnabled bool, tDomainAllowList []interface{}, tIpAddressAllowList []interface{}, pReq *platformclientv2.Passwordrequirements) map[string]interface{} {
	// Convert pReq to a map representation
	passwordRequirementsMap := map[string]interface{}{
		"minimum_length":      *pReq.MinimumLength,
		"minimum_digits":      *pReq.MinimumDigits,
		"minimum_letters":     *pReq.MinimumLetters,
		"minimum_upper":       *pReq.MinimumUpper,
		"minimum_lower":       *pReq.MinimumLower,
		"minimum_specials":    *pReq.MinimumSpecials,
		"minimum_age_seconds": *pReq.MinimumAgeSeconds,
		"expiration_days":     *pReq.ExpirationDays,
	}

	orgAuthSettingsDataMap := map[string]interface{}{
		"multifactor_authentication_required": tMultifactorAuthRequired,
		"domain_allowlist_enabled":            tDomainAllowListEnabled,
		"domain_allowlist":                    tDomainAllowList,
		"ip_address_allowlist":                tIpAddressAllowList,
		"password_requirements":               passwordRequirementsMap,
	}
	return orgAuthSettingsDataMap
}
