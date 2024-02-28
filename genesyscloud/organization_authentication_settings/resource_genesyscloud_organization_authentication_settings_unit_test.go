package organization_authentication_settings

import (
	"context"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v121/platformclientv2"
	"github.com/stretchr/testify/assert"
	"net/http"
	gcloud2 "terraform-provider-genesyscloud/genesyscloud"
	"testing"
)

// Unit Test

func TestUnitResourceOrganizationAuthenticationSettingsRead(t *testing.T) {
	tId := uuid.NewString()
	tMultifactorAuthRequired := true
	tDomainAllowListEnabled := false
	tDomainAllowList := []string{"Genesys.com"}
	tIpAddressAllowList := []string{"0.0.0.0/32", "1.1.1.1/32"}

	tMinimumLength := 8
	tMinimumDigits := 2
	tMinimumLetters := 2
	tMinimumUpper := 1
	tMinimumLower := 1
	tMinimumSpecials := 1
	tMinimumAgeSeconds := 1
	tExpirationDays := 90

	orgAuthProxy := &orgAuthSettingsProxy{}
	orgAuthProxy.getOrgAuthSettingsByIdAttr = func(ctx context.Context, o *orgAuthSettingsProxy, id string) (*platformclientv2.Orgauthsettings, *platformclientv2.APIResponse, error) {
		assert.Equal(t, tId, id)
		orgAuthSettings := &platformclientv2.Orgauthsettings{
			MultifactorAuthenticationRequired: &tMultifactorAuthRequired,
			DomainAllowlistEnabled:            &tDomainAllowListEnabled,
			DomainAllowlist:                   &tDomainAllowList,
			IpAddressAllowlist:                &tIpAddressAllowList,
			PasswordRequirements: &platformclientv2.Passwordrequirements{
				MinimumLength:     &tMinimumLength,
				MinimumDigits:     &tMinimumDigits,
				MinimumLetters:    &tMinimumLetters,
				MinimumUpper:      &tMinimumUpper,
				MinimumLower:      &tMinimumLower,
				MinimumSpecials:   &tMinimumSpecials,
				MinimumAgeSeconds: &tMinimumAgeSeconds,
				ExpirationDays:    &tExpirationDays,
			},
		}

		apiResponse := &platformclientv2.APIResponse{StatusCode: http.StatusOK}
		return orgAuthSettings, apiResponse, nil
	}
	internalProxy = orgAuthProxy
	defer func() { internalProxy = nil }()

	ctx := context.Background()
	gcloud := &gcloud2.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}

	resourceSchema := ResourceOrganizationAuthenticationSettings().Schema

	resourceDataMap := buildOrgAuthSettingsDataMap(tId, tMultifactorAuthRequired, tDomainAllowListEnabled, tDomainAllowList, tIpAddressAllowList, tMinimumLength, tMinimumDigits, tMinimumLetters, tMinimumUpper, tMinimumLower, tMinimumSpecials, tMinimumAgeSeconds, tExpirationDays)

	d := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
	d.SetId(tId)

	diag := readOrganizationAuthenticationSettings(ctx, d, gcloud)
	assert.Equal(t, false, diag.HasError())
	assert.Equal(t, tId, d.Id())
	assert.Equal(t, tMultifactorAuthRequired, d.Get("multifactor_authentication_required").(bool))
	assert.Equal(t, tMultifactorAuthRequired, d.Get("multifactor_authentication_required").(bool))
	assert.Equal(t, tMultifactorAuthRequired, d.Get("multifactor_authentication_required").(bool))
	assert.Equal(t, tMultifactorAuthRequired, d.Get("multifactor_authentication_required").(bool))
}
