package organization_authentication_settings

import (
	"terraform-provider-genesyscloud/genesyscloud/util/lists"
	"terraform-provider-genesyscloud/genesyscloud/util/resourcedata"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

/*
The resource_genesyscloud_organization_authentication_settings_utils.go file contains various helper methods to marshal
and unmarshal data into formats consumable by Terraform and/or Genesys Cloud.
*/

// getOrganizationAuthenticationSettingsFromResourceData maps data from schema ResourceData object to a platformclientv2.Orgauthsettings
func getOrganizationAuthenticationSettingsFromResourceData(d *schema.ResourceData) platformclientv2.Orgauthsettings {
	return platformclientv2.Orgauthsettings{
		PasswordRequirements:              buildPasswordRequirements(d, "password_requirements"),
		MultifactorAuthenticationRequired: platformclientv2.Bool(d.Get("multifactor_authentication_required").(bool)),
		DomainAllowlistEnabled:            platformclientv2.Bool(d.Get("domain_allowlist_enabled").(bool)),
		DomainAllowlist:                   lists.BuildSdkStringListFromInterfaceArray(d, "domain_allowlist"),
		IpAddressAllowlist:                lists.BuildSdkStringListFromInterfaceArray(d, "ip_address_allowlist"),
	}
}

// buildPasswordRequirements maps an []interface{} into a Genesys Cloud *[]platformclientv2.Passwordrequirements
func buildPasswordRequirements(d *schema.ResourceData, key string) *platformclientv2.Passwordrequirements {
	if d.Get(key) != nil {
		passwordData := d.Get(key).([]interface{})
		if len(passwordData) > 0 {
			pReqMap := passwordData[0].(map[string]interface{})
			minimumLength := pReqMap["minimum_length"].(int)
			minimumDigits := pReqMap["minimum_digits"].(int)
			minimumLetters := pReqMap["minimum_letters"].(int)
			minimumUpper := pReqMap["minimum_upper"].(int)
			minimumLower := pReqMap["minimum_lower"].(int)
			minimumSpecials := pReqMap["minimum_specials"].(int)
			minimumAgeSeconds := pReqMap["minimum_age_seconds"].(int)
			expirationDays := pReqMap["expiration_days"].(int)

			return &platformclientv2.Passwordrequirements{
				MinimumLength:     &minimumLength,
				MinimumDigits:     &minimumDigits,
				MinimumLetters:    &minimumLetters,
				MinimumUpper:      &minimumUpper,
				MinimumLower:      &minimumLower,
				MinimumSpecials:   &minimumSpecials,
				MinimumAgeSeconds: &minimumAgeSeconds,
				ExpirationDays:    &expirationDays,
			}
		}
	}
	return nil
}

// flattenPasswordRequirements maps a Genesys Cloud []platformclientv2.Passwordrequirements into a []interface{}
func flattenPasswordRequirements(passwordRequirements *platformclientv2.Passwordrequirements) []interface{} {
	pReqInterface := make(map[string]interface{})

	resourcedata.SetMapValueIfNotNil(pReqInterface, "minimum_length", passwordRequirements.MinimumLength)
	resourcedata.SetMapValueIfNotNil(pReqInterface, "minimum_digits", passwordRequirements.MinimumDigits)
	resourcedata.SetMapValueIfNotNil(pReqInterface, "minimum_letters", passwordRequirements.MinimumLetters)
	resourcedata.SetMapValueIfNotNil(pReqInterface, "minimum_upper", passwordRequirements.MinimumUpper)
	resourcedata.SetMapValueIfNotNil(pReqInterface, "minimum_lower", passwordRequirements.MinimumLower)
	resourcedata.SetMapValueIfNotNil(pReqInterface, "minimum_specials", passwordRequirements.MinimumSpecials)
	resourcedata.SetMapValueIfNotNil(pReqInterface, "minimum_age_seconds", passwordRequirements.MinimumAgeSeconds)
	resourcedata.SetMapValueIfNotNil(pReqInterface, "expiration_days", passwordRequirements.ExpirationDays)

	return []interface{}{pReqInterface}
}
