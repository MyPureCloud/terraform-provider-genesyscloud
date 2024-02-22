package organization_authentication_settings

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"strings"
	gcloud "terraform-provider-genesyscloud/genesyscloud"
	"testing"
)

func TestAccResourceOrgAuthSettings(t *testing.T) {

	t.Parallel()
	var (
		resourceId          = "OrgAuthSettings"
		domainAllowList1    = []string{"Google.com"}
		ipAddressAllowList1 = []string{"0.0.0.0/32"}

		minimumLength1     = "8"
		minimumDigits1     = "2"
		minimumLetters1    = "2"
		minimumUpper1      = "1"
		minimumLower1      = "1"
		minimumSpecials1   = "1"
		minimumAgeSeconds1 = "1"
		expirationDays1    = "90"

		domainAllowList2    = []string{"Google.ie"}
		ipAddressAllowList2 = []string{"0.0.0.0/32"}

		minimumLength2     = "10"
		minimumDigits2     = "4"
		minimumLetters2    = "4"
		minimumUpper2      = "2"
		minimumLower2      = "2"
		minimumSpecials2   = "2"
		minimumAgeSeconds2 = "2"
		expirationDays2    = "91"
	)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			gcloud.TestAccPreCheck(t)
		},
		ProviderFactories: gcloud.GetProviderFactories(providerResources, nil),
		Steps: []resource.TestStep{
			{
				// Create
				Config: generateOrganizationAuthenticationSettings(
					resourceId,
					"true",
					"true",
					domainAllowList1,
					ipAddressAllowList1,
					generatePasswordRequirements(
						minimumLength1,
						minimumDigits1,
						minimumLetters1,
						minimumUpper1,
						minimumLower1,
						minimumSpecials1,
						minimumAgeSeconds1,
						expirationDays1,
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_organization_authentication_settings."+resourceId, "multifactor_authentication_required", "true"),
					resource.TestCheckResourceAttr("genesyscloud_organization_authentication_settings."+resourceId, "domain_allowlist_enabled", "true"),
					resource.TestCheckResourceAttr("genesyscloud_organization_authentication_settings."+resourceId, "domain_allow_list.0", domainAllowList1[0]),
					resource.TestCheckResourceAttr("genesyscloud_organization_authentication_settings."+resourceId, "ip_address_allow_list.0", ipAddressAllowList1[0]),
					resource.TestCheckResourceAttr("genesyscloud_organization_authentication_settings."+resourceId, "password_requirements.0.minimum_length", minimumLength1),
					resource.TestCheckResourceAttr("genesyscloud_organization_authentication_settings."+resourceId, "password_requirements.0.minimum_digits", minimumDigits1),
					resource.TestCheckResourceAttr("genesyscloud_organization_authentication_settings."+resourceId, "password_requirements.0.minimum_letters", minimumLetters1),
					resource.TestCheckResourceAttr("genesyscloud_organization_authentication_settings."+resourceId, "password_requirements.0.minimum_upper", minimumUpper1),
					resource.TestCheckResourceAttr("genesyscloud_organization_authentication_settings."+resourceId, "password_requirements.0.minimum_lower", minimumLower1),
					resource.TestCheckResourceAttr("genesyscloud_organization_authentication_settings."+resourceId, "password_requirements.0.minimum_specials", minimumSpecials1),
					resource.TestCheckResourceAttr("genesyscloud_organization_authentication_settings."+resourceId, "password_requirements.0.minimum_age_seconds", minimumAgeSeconds1),
					resource.TestCheckResourceAttr("genesyscloud_organization_authentication_settings."+resourceId, "password_requirements.0.expiration_days", expirationDays1),
				),
			},
			{
				// Update with new organization authentication settings and password requirements
				Config: generateOrganizationAuthenticationSettings(
					resourceId,
					"false",
					"true",
					domainAllowList2,
					ipAddressAllowList2,
					generatePasswordRequirements(
						minimumLength2,
						minimumDigits2,
						minimumLetters2,
						minimumUpper2,
						minimumLower2,
						minimumSpecials2,
						minimumAgeSeconds2,
						expirationDays2,
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_organization_authentication_settings."+resourceId, "multifactor_authentication_required", "true"),
					resource.TestCheckResourceAttr("genesyscloud_organization_authentication_settings."+resourceId, "domain_allowlist_enabled", "true"),
					resource.TestCheckResourceAttr("genesyscloud_organization_authentication_settings."+resourceId, "domain_allow_list.0", domainAllowList2[0]),
					resource.TestCheckResourceAttr("genesyscloud_organization_authentication_settings."+resourceId, "ip_address_allow_list.0", ipAddressAllowList2[0]),
					resource.TestCheckResourceAttr("genesyscloud_organization_authentication_settings."+resourceId, "password_requirements.0.minimum_length", minimumLength2),
					resource.TestCheckResourceAttr("genesyscloud_organization_authentication_settings."+resourceId, "password_requirements.0.minimum_digits", minimumDigits2),
					resource.TestCheckResourceAttr("genesyscloud_organization_authentication_settings."+resourceId, "password_requirements.0.minimum_letters", minimumLetters2),
					resource.TestCheckResourceAttr("genesyscloud_organization_authentication_settings."+resourceId, "password_requirements.0.minimum_upper", minimumUpper2),
					resource.TestCheckResourceAttr("genesyscloud_organization_authentication_settings."+resourceId, "password_requirements.0.minimum_lower", minimumLower2),
					resource.TestCheckResourceAttr("genesyscloud_organization_authentication_settings."+resourceId, "password_requirements.0.minimum_specials", minimumSpecials2),
					resource.TestCheckResourceAttr("genesyscloud_organization_authentication_settings."+resourceId, "password_requirements.0.minimum_age_seconds", minimumAgeSeconds2),
					resource.TestCheckResourceAttr("genesyscloud_organization_authentication_settings."+resourceId, "password_requirements.0.expiration_days", expirationDays2),
				),
			},
			{
				// Read
				ResourceName:      "genesyscloud_organization_authentication_settings" + resourceId,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func generateOrganizationAuthenticationSettings(
	resourceId string,
	multifactorAuthenticationRequired string,
	domainAllowlistEnabled string,
	domainAllowlist []string,
	ipAddressAllowlist []string,
	nestedBlocks ...string) string {

	return fmt.Sprintf(`
		resource "genesyscloud_organization_authentication_settings" "%s"{
			multifactor_authentication_required = %s
			domain_allowlist_enabled = %s
			domain_allowlist = ["%s"]
			ip_address_allowlist = ["%s"]
			%s
		}
		`, resourceId, multifactorAuthenticationRequired, domainAllowlistEnabled, domainAllowlist, ipAddressAllowlist, strings.Join(nestedBlocks, "\n"),
	)
}

func generatePasswordRequirements(
	minimumLength string,
	minimumDigits string,
	minimumLetters string,
	minimumUpper string,
	minimumLower string,
	minimumSpecials string,
	minimumAgeSeconds string,
	expirationDays string) string {

	return fmt.Sprintf(`
		password_requirements {
			minimum_length = %s
			minimum_digits = %s
			minimum_letters = %s
			minimum_upper = %s
			minimum_lower = %s
			minimum_specials = %s
			minimum_age_seconds = %s
			expiration_days = %s
		}
	`, minimumLength, minimumDigits, minimumLetters, minimumUpper, minimumLower, minimumSpecials, minimumAgeSeconds, expirationDays)
}
