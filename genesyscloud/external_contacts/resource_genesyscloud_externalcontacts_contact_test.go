package external_contacts

import (
	"fmt"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

/*
The resource_genesyscloud_externalcontacts_contact_test.go contains all of the test cases for running the resource
tests for external_contacts.
*/
func TestAccResourceExternalContacts(t *testing.T) {
	var (
		contactresource1 = "externalcontact-contact1"
		title1           = "integration team"

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
					contactresource1,
					title1,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_externalcontacts_contact."+contactresource1, "title", title1),
					resource.TestCheckResourceAttr("genesyscloud_externalcontacts_contact."+contactresource1, "first_name", ""),
					resource.TestCheckResourceAttr("genesyscloud_externalcontacts_contact."+contactresource1, "middle_name", ""),
					resource.TestCheckResourceAttr("genesyscloud_externalcontacts_contact."+contactresource1, "last_name", ""),
				),
			},
			{
				// Update with more attributes
				Config: generateFullExternalContactResource(
					contactresource1,
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
					resource.TestCheckResourceAttr("genesyscloud_externalcontacts_contact."+contactresource1, "title", title2),
					resource.TestCheckResourceAttr("genesyscloud_externalcontacts_contact."+contactresource1, "first_name", firstname2),
					resource.TestCheckResourceAttr("genesyscloud_externalcontacts_contact."+contactresource1, "middle_name", middlename2),
					resource.TestCheckResourceAttr("genesyscloud_externalcontacts_contact."+contactresource1, "last_name", lastname2),
					resource.TestCheckResourceAttr("genesyscloud_externalcontacts_contact."+contactresource1, "work_phone.0.display", phoneDisplay),
					resource.TestCheckResourceAttr("genesyscloud_externalcontacts_contact."+contactresource1, "work_phone.0.extension", phoneExtension),
					resource.TestCheckResourceAttr("genesyscloud_externalcontacts_contact."+contactresource1, "work_phone.0.accepts_sms", phoneAcceptssms),
					resource.TestCheckResourceAttr("genesyscloud_externalcontacts_contact."+contactresource1, "work_phone.0.e164", phoneE164),
					resource.TestCheckResourceAttr("genesyscloud_externalcontacts_contact."+contactresource1, "work_phone.0.country_code", phoneCountrycode),
					resource.TestCheckResourceAttr("genesyscloud_externalcontacts_contact."+contactresource1, "cell_phone.0.display", phoneDisplay),
					resource.TestCheckResourceAttr("genesyscloud_externalcontacts_contact."+contactresource1, "cell_phone.0.extension", phoneExtension),
					resource.TestCheckResourceAttr("genesyscloud_externalcontacts_contact."+contactresource1, "cell_phone.0.accepts_sms", phoneAcceptssms),
					resource.TestCheckResourceAttr("genesyscloud_externalcontacts_contact."+contactresource1, "cell_phone.0.e164", phoneE164),
					resource.TestCheckResourceAttr("genesyscloud_externalcontacts_contact."+contactresource1, "cell_phone.0.country_code", phoneCountrycode),
					resource.TestCheckResourceAttr("genesyscloud_externalcontacts_contact."+contactresource1, "home_phone.0.display", phoneDisplay),
					resource.TestCheckResourceAttr("genesyscloud_externalcontacts_contact."+contactresource1, "home_phone.0.extension", phoneExtension),
					resource.TestCheckResourceAttr("genesyscloud_externalcontacts_contact."+contactresource1, "home_phone.0.accepts_sms", phoneAcceptssms),
					resource.TestCheckResourceAttr("genesyscloud_externalcontacts_contact."+contactresource1, "home_phone.0.e164", phoneE164),
					resource.TestCheckResourceAttr("genesyscloud_externalcontacts_contact."+contactresource1, "home_phone.0.country_code", phoneCountrycode),
					resource.TestCheckResourceAttr("genesyscloud_externalcontacts_contact."+contactresource1, "other_phone.0.display", phoneDisplay),
					resource.TestCheckResourceAttr("genesyscloud_externalcontacts_contact."+contactresource1, "other_phone.0.extension", phoneExtension),
					resource.TestCheckResourceAttr("genesyscloud_externalcontacts_contact."+contactresource1, "other_phone.0.accepts_sms", phoneAcceptssms),
					resource.TestCheckResourceAttr("genesyscloud_externalcontacts_contact."+contactresource1, "other_phone.0.e164", phoneE164),
					resource.TestCheckResourceAttr("genesyscloud_externalcontacts_contact."+contactresource1, "other_phone.0.country_code", phoneCountrycode),
					resource.TestCheckResourceAttr("genesyscloud_externalcontacts_contact."+contactresource1, "address.0.address1", address1),
					resource.TestCheckResourceAttr("genesyscloud_externalcontacts_contact."+contactresource1, "address.0.address2", address2),
					resource.TestCheckResourceAttr("genesyscloud_externalcontacts_contact."+contactresource1, "address.0.city", city),
					resource.TestCheckResourceAttr("genesyscloud_externalcontacts_contact."+contactresource1, "address.0.state", state),
					resource.TestCheckResourceAttr("genesyscloud_externalcontacts_contact."+contactresource1, "address.0.postal_code", postal_code),
					resource.TestCheckResourceAttr("genesyscloud_externalcontacts_contact."+contactresource1, "address.0.country_code", country_code),
					resource.TestCheckResourceAttr("genesyscloud_externalcontacts_contact."+contactresource1, "twitter_id.0.id", twitterId),
					resource.TestCheckResourceAttr("genesyscloud_externalcontacts_contact."+contactresource1, "twitter_id.0.name", twitterName),
					resource.TestCheckResourceAttr("genesyscloud_externalcontacts_contact."+contactresource1, "twitter_id.0.screen_name", twitterScreenname),
					resource.TestCheckResourceAttr("genesyscloud_externalcontacts_contact."+contactresource1, "line_id.0.ids.0.user_id", lineId),
					resource.TestCheckResourceAttr("genesyscloud_externalcontacts_contact."+contactresource1, "line_id.0.display_name", lineDisplayname),
					resource.TestCheckResourceAttr("genesyscloud_externalcontacts_contact."+contactresource1, "whatsapp_id.0.phone_number.0.display", whatsappPhoneDisplay),
					resource.TestCheckResourceAttr("genesyscloud_externalcontacts_contact."+contactresource1, "whatsapp_id.0.phone_number.0.e164", whatsappPhoneE164),
					resource.TestCheckResourceAttr("genesyscloud_externalcontacts_contact."+contactresource1, "whatsapp_id.0.phone_number.0.country_code", whatsappPhoneCountryCode),
					resource.TestCheckResourceAttr("genesyscloud_externalcontacts_contact."+contactresource1, "whatsapp_id.0.display_name", whatsappPhoneDisplayName),
					resource.TestCheckResourceAttr("genesyscloud_externalcontacts_contact."+contactresource1, "facebook_id.0.ids.0.scoped_id", facebookScopedid),
					resource.TestCheckResourceAttr("genesyscloud_externalcontacts_contact."+contactresource1, "facebook_id.0.display_name", facebookDisplayname),
					resource.TestCheckResourceAttr("genesyscloud_externalcontacts_contact."+contactresource1, "survey_opt_out", surveyoptout),
					resource.TestCheckResourceAttr("genesyscloud_externalcontacts_contact."+contactresource1, "external_system_url", externalsystemurl),
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_externalcontacts_contact." + contactresource1,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifyContactDestroyed,
	})
}

func generateFullExternalContactResource(
	resourceID string,
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
	`, resourceID, firstname, middlename, lastname, title,
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
