package external_contacts

import (
	"fmt"
	externalContactOrganization "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/external_contacts_organization"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
	"strconv"
	"testing"

	"github.com/google/uuid"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v157/platformclientv2"
)

/*
The resource_genesyscloud_externalcontacts_contact_test.go contains all of the test cases for running the resource
tests for external_contacts.
*/
func TestAccResourceExternalContacts(t *testing.T) {
	var (
		contactResourceLabel1 = "externalcontact-contact1"
		title1                = "integration team"

		firstname2  = "jean"
		middlename2 = "jacques"
		lastname2   = "dupont"
		title2      = "integration team"

		phoneDisplay     = "+33 1 00 00 00 01"
		phoneExtension   = "2"
		phoneAcceptssms  = "false"
		phoneE164        = "+33100000001"
		phoneCountrycode = "FR"

		address1     = "1 rue de la paix"
		address2     = "2 rue de la paix"
		city         = "Paris"
		state        = "île-de-France"
		postal_code  = "75000"
		country_code = "FR"

		twitterId         = "twitterId"
		twitterName       = "twitterName"
		twitterScreenname = "twitterScreenname"

		lineId          = "lineID12345"
		lineDisplayname = "lineDisplayname"

		whatsappPhoneDisplay     = "+33 1 00 00 00 01"
		whatsappPhoneE164        = "+33100000001"
		whatsappPhoneCountryCode = "FR"
		whatsappPhoneDisplayName = "whatsappName"

		facebookScopedid    = "facebookScopedid"
		facebookDisplayname = "facebookDisplayname"
		surveyoptout        = "false"
		externalsystemurl   = "https://externalsystemurl.com"
	)
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create
				Config: GenerateBasicExternalContactResource(
					contactResourceLabel1,
					title1,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_externalcontacts_contact."+contactResourceLabel1, "title", title1),
					resource.TestCheckResourceAttr("genesyscloud_externalcontacts_contact."+contactResourceLabel1, "first_name", ""),
					resource.TestCheckResourceAttr("genesyscloud_externalcontacts_contact."+contactResourceLabel1, "middle_name", ""),
					resource.TestCheckResourceAttr("genesyscloud_externalcontacts_contact."+contactResourceLabel1, "last_name", ""),
				),
			},
			{
				// Update with more attributes
				Config: generateFullExternalContactResource(
					contactResourceLabel1,
					firstname2,
					middlename2,
					lastname2,
					title2,
					phoneDisplay,
					phoneExtension,
					phoneAcceptssms,
					phoneE164,
					phoneCountrycode,
					address1,
					address2,
					city,
					state,
					postal_code,
					country_code,
					twitterId,
					twitterName,
					twitterScreenname,
					lineId,
					lineDisplayname,
					whatsappPhoneDisplay,
					whatsappPhoneE164,
					whatsappPhoneCountryCode,
					whatsappPhoneDisplayName,
					facebookScopedid,
					facebookDisplayname,
					surveyoptout,
					externalsystemurl,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_externalcontacts_contact."+contactResourceLabel1, "title", title2),
					resource.TestCheckResourceAttr("genesyscloud_externalcontacts_contact."+contactResourceLabel1, "first_name", firstname2),
					resource.TestCheckResourceAttr("genesyscloud_externalcontacts_contact."+contactResourceLabel1, "middle_name", middlename2),
					resource.TestCheckResourceAttr("genesyscloud_externalcontacts_contact."+contactResourceLabel1, "last_name", lastname2),
					resource.TestCheckResourceAttr("genesyscloud_externalcontacts_contact."+contactResourceLabel1, "work_phone.0.display", phoneDisplay),
					resource.TestCheckResourceAttr("genesyscloud_externalcontacts_contact."+contactResourceLabel1, "work_phone.0.extension", phoneExtension),
					resource.TestCheckResourceAttr("genesyscloud_externalcontacts_contact."+contactResourceLabel1, "work_phone.0.accepts_sms", phoneAcceptssms),
					resource.TestCheckResourceAttr("genesyscloud_externalcontacts_contact."+contactResourceLabel1, "work_phone.0.e164", phoneE164),
					resource.TestCheckResourceAttr("genesyscloud_externalcontacts_contact."+contactResourceLabel1, "work_phone.0.country_code", phoneCountrycode),
					resource.TestCheckResourceAttr("genesyscloud_externalcontacts_contact."+contactResourceLabel1, "cell_phone.0.display", phoneDisplay),
					resource.TestCheckResourceAttr("genesyscloud_externalcontacts_contact."+contactResourceLabel1, "cell_phone.0.extension", phoneExtension),
					resource.TestCheckResourceAttr("genesyscloud_externalcontacts_contact."+contactResourceLabel1, "cell_phone.0.accepts_sms", phoneAcceptssms),
					resource.TestCheckResourceAttr("genesyscloud_externalcontacts_contact."+contactResourceLabel1, "cell_phone.0.e164", phoneE164),
					resource.TestCheckResourceAttr("genesyscloud_externalcontacts_contact."+contactResourceLabel1, "cell_phone.0.country_code", phoneCountrycode),
					resource.TestCheckResourceAttr("genesyscloud_externalcontacts_contact."+contactResourceLabel1, "home_phone.0.display", phoneDisplay),
					resource.TestCheckResourceAttr("genesyscloud_externalcontacts_contact."+contactResourceLabel1, "home_phone.0.extension", phoneExtension),
					resource.TestCheckResourceAttr("genesyscloud_externalcontacts_contact."+contactResourceLabel1, "home_phone.0.accepts_sms", phoneAcceptssms),
					resource.TestCheckResourceAttr("genesyscloud_externalcontacts_contact."+contactResourceLabel1, "home_phone.0.e164", phoneE164),
					resource.TestCheckResourceAttr("genesyscloud_externalcontacts_contact."+contactResourceLabel1, "home_phone.0.country_code", phoneCountrycode),
					resource.TestCheckResourceAttr("genesyscloud_externalcontacts_contact."+contactResourceLabel1, "other_phone.0.display", phoneDisplay),
					resource.TestCheckResourceAttr("genesyscloud_externalcontacts_contact."+contactResourceLabel1, "other_phone.0.extension", phoneExtension),
					resource.TestCheckResourceAttr("genesyscloud_externalcontacts_contact."+contactResourceLabel1, "other_phone.0.accepts_sms", phoneAcceptssms),
					resource.TestCheckResourceAttr("genesyscloud_externalcontacts_contact."+contactResourceLabel1, "other_phone.0.e164", phoneE164),
					resource.TestCheckResourceAttr("genesyscloud_externalcontacts_contact."+contactResourceLabel1, "other_phone.0.country_code", phoneCountrycode),
					resource.TestCheckResourceAttr("genesyscloud_externalcontacts_contact."+contactResourceLabel1, "address.0.address1", address1),
					resource.TestCheckResourceAttr("genesyscloud_externalcontacts_contact."+contactResourceLabel1, "address.0.address2", address2),
					resource.TestCheckResourceAttr("genesyscloud_externalcontacts_contact."+contactResourceLabel1, "address.0.city", city),
					resource.TestCheckResourceAttr("genesyscloud_externalcontacts_contact."+contactResourceLabel1, "address.0.state", state),
					resource.TestCheckResourceAttr("genesyscloud_externalcontacts_contact."+contactResourceLabel1, "address.0.postal_code", postal_code),
					resource.TestCheckResourceAttr("genesyscloud_externalcontacts_contact."+contactResourceLabel1, "address.0.country_code", country_code),
					resource.TestCheckResourceAttr("genesyscloud_externalcontacts_contact."+contactResourceLabel1, "twitter_id.0.id", twitterId),
					resource.TestCheckResourceAttr("genesyscloud_externalcontacts_contact."+contactResourceLabel1, "twitter_id.0.name", twitterName),
					resource.TestCheckResourceAttr("genesyscloud_externalcontacts_contact."+contactResourceLabel1, "twitter_id.0.screen_name", twitterScreenname),
					resource.TestCheckResourceAttr("genesyscloud_externalcontacts_contact."+contactResourceLabel1, "line_id.0.ids.0.user_id", lineId),
					resource.TestCheckResourceAttr("genesyscloud_externalcontacts_contact."+contactResourceLabel1, "line_id.0.display_name", lineDisplayname),
					resource.TestCheckResourceAttr("genesyscloud_externalcontacts_contact."+contactResourceLabel1, "whatsapp_id.0.phone_number.0.display", whatsappPhoneDisplay),
					resource.TestCheckResourceAttr("genesyscloud_externalcontacts_contact."+contactResourceLabel1, "whatsapp_id.0.phone_number.0.e164", whatsappPhoneE164),
					resource.TestCheckResourceAttr("genesyscloud_externalcontacts_contact."+contactResourceLabel1, "whatsapp_id.0.phone_number.0.country_code", whatsappPhoneCountryCode),
					resource.TestCheckResourceAttr("genesyscloud_externalcontacts_contact."+contactResourceLabel1, "whatsapp_id.0.display_name", whatsappPhoneDisplayName),
					resource.TestCheckResourceAttr("genesyscloud_externalcontacts_contact."+contactResourceLabel1, "facebook_id.0.ids.0.scoped_id", facebookScopedid),
					resource.TestCheckResourceAttr("genesyscloud_externalcontacts_contact."+contactResourceLabel1, "facebook_id.0.display_name", facebookDisplayname),
					resource.TestCheckResourceAttr("genesyscloud_externalcontacts_contact."+contactResourceLabel1, "survey_opt_out", surveyoptout),
					resource.TestCheckResourceAttr("genesyscloud_externalcontacts_contact."+contactResourceLabel1, "external_system_url", externalsystemurl),
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_externalcontacts_contact." + contactResourceLabel1,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifyContactDestroyed,
	})
}

func TestAccResourceExternalContactsOrganizationRef(t *testing.T) {
	var (
		organizationResourceLabel1 = "external_organization_test"
		organizationResourcePath1  = externalContactOrganization.ResourceType + "." + organizationResourceLabel1
		name1                      = "ABCCorp-" + uuid.NewString()
		phoneDisplay1              = "+1 321-700-1243"
		countryCode1               = "US"
		address1                   = "1011 New Hope St"
		city1                      = "Norristown"
		state1                     = "PA"
		postalCode1                = "19401"
		twitterId1                 = "twitterId"
		twitterName1               = "twitterName"
		twitterScreenName1         = "twitterScreenname"
		symbol1                    = "ABC"
		exchange1                  = "NYSE"
		tags1                      = []string{
			strconv.Quote("news"),
			strconv.Quote("channel"),
		}

		organizationResourceLabel2 = "new_external_organization"
		organizationResourcePath2  = externalContactOrganization.ResourceType + "." + organizationResourceLabel2
		name2                      = "NewCorp-" + uuid.NewString()
		phoneDisplay2              = "+1 321-990-9876"
		countryCode2               = "US"
		address2                   = "65 Upper Street"
		city2                      = "Springfield"
		state2                     = "MO"
		postalCode2                = "67890"
		twitterId2                 = "twitterId2"
		twitterName2               = "twitterName2"
		twitterScreenName2         = "twitterScreenname2"
		symbol2                    = "NEW"
		exchange2                  = "NYSE"
		tags2                      = []string{
			strconv.Quote("SWE"),
			strconv.Quote("development"),
		}
	)
	var (
		contactResourceLabel                 = "testing_contact"
		contactResourcePath                  = ResourceType + "." + contactResourceLabel
		title                                = "testing team"
		title2                               = "dev team"
		externalContactOrganizationResource1 = externalContactOrganization.GenerateBasicExternalOrganizationResource(
			organizationResourceLabel1,
			name1,
			phoneDisplay1, countryCode1,
			address1, city1, state1, postalCode1, countryCode1,
			twitterId1, twitterName1, twitterScreenName1,
			symbol1, exchange1,
			tags1,
			"",
		)
		externalContactOrganizationResource2 = externalContactOrganization.GenerateBasicExternalOrganizationResource(
			organizationResourceLabel2,
			name2,
			phoneDisplay2, countryCode2,
			address2, city2, state2, postalCode2, countryCode2,
			twitterId2, twitterName2, twitterScreenName2,
			symbol2, exchange2,
			tags2,
			"",
		)
	)
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				//Create
				Config: externalContactOrganizationResource1 +
					GenerateBasicExternalContactResource(
						contactResourceLabel,
						title,
						"external_organization_id = "+organizationResourcePath1+".id",
					),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(contactResourcePath, "title", title),
					resource.TestCheckResourceAttrPair(
						organizationResourcePath1, "id",
						contactResourcePath, "external_organization_id",
					),
				),
			},
			{
				//Update
				Config: externalContactOrganizationResource2 + GenerateBasicExternalContactResource(
					contactResourceLabel,
					title2,
					"external_organization_id = "+organizationResourcePath2+".id",
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(contactResourcePath, "title", title2),
					resource.TestCheckResourceAttrPair(
						organizationResourcePath2, "id",
						contactResourcePath, "external_organization_id",
					),
				),
			},
			{
				// Import/Read
				ResourceName:      contactResourcePath,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifyContactDestroyed,
	})
}

func generateFullExternalContactResource(
	resourceLabel string,
	firstname string, middlename string, lastname string, title string,
	phoneDisplay string, phoneExtension string, phoneAcceptssms string, phoneE164 string, phoneCountrycode string,
	address1 string, address2 string, city string, state string, postal_code string, country_code string,
	twitterId string, twitterName string, twitterScreenname string,
	lineId string, lineDisplayname string,
	whatssappDisplay string, whatssappE164 string, whatssappCountrycode string, whatssappDisplayname string,
	facebookScopeid string, facebookDisplayname string,
	surveyoptout string, externalsystemurl string) string {
	return fmt.Sprintf(`resource "genesyscloud_externalcontacts_contact" "%s" {
		first_name = "%s"
		middle_name = "%s"
		last_name = "%s"
		title = "%s"
		work_phone {
		  display = "%s"
		  extension = %s
		  accepts_sms = %s
		  e164 = "%s"
		  country_code = "%s"
		}
		cell_phone {
		  display = "%s"
		  extension = %s
		  accepts_sms = %s
		  e164 = "%s"
		  country_code = "%s"
		}
		home_phone {
		  display = "%s"
		  extension = %s
		  accepts_sms = %s
		  e164 = "%s"
		  country_code = "%s"
		}
		other_phone {
		  display = "%s"
		  extension = %s
		  accepts_sms = %s
		  e164 = "%s"
		  country_code = "%s"
		}
		address {
		  address1 = "%s"
		  address2 = "%s"
		  city = "%s"
		  state = "%s"
		  postal_code = "%s"
		  country_code = "%s"
		}
		twitter_id {
		  id = "%s"
		  name = "%s"
		  screen_name = "%s"
		}
		line_id {
		  ids {
			user_id = "%s"
		  }
		  display_name = "%s"
		}
		whatsapp_id {
			phone_number {
				display = "%s"
				e164 = "%s"
				country_code = "%s"
			}
			display_name = "%s"
		}
		facebook_id {
			ids {
			  scoped_id = "%s"
			}
			display_name = "%s"
		  }
		survey_opt_out = %s
		external_system_url = "%s"
	}
	`, resourceLabel, firstname, middlename, lastname, title,
		phoneDisplay, phoneExtension, phoneAcceptssms, phoneE164, phoneCountrycode,
		phoneDisplay, phoneExtension, phoneAcceptssms, phoneE164, phoneCountrycode,
		phoneDisplay, phoneExtension, phoneAcceptssms, phoneE164, phoneCountrycode,
		phoneDisplay, phoneExtension, phoneAcceptssms, phoneE164, phoneCountrycode,
		address1, address2, city, state, postal_code, country_code,
		twitterId, twitterName, twitterScreenname,
		lineId, lineDisplayname,
		whatssappDisplay, whatssappE164, whatssappCountrycode, whatssappDisplayname,
		facebookScopeid, facebookDisplayname,
		surveyoptout, externalsystemurl)
}

func testVerifyContactDestroyed(state *terraform.State) error {
	externalAPI := platformclientv2.NewExternalContactsApi()
	for _, rs := range state.RootModule().Resources {
		if rs.Type != "genesyscloud_externalcontacts_contact" {
			continue
		}

		externalContact, resp, err := externalAPI.GetExternalcontactsContact(rs.Primary.ID, nil)
		if externalContact != nil {
			return fmt.Errorf("External contact (%s) still exists", rs.Primary.ID)
		} else if util.IsStatus404(resp) {
			// External Contact not found as expected
			continue
		} else {
			// Unexpected error
			return fmt.Errorf("Unexpected error: %s", err)
		}
	}
	// Success. All contacts destroyed
	return nil
}
