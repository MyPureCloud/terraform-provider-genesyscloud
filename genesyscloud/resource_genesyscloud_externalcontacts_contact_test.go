package genesyscloud

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v92/platformclientv2"
)

func TestAccResourceExternalContacts(t *testing.T) {
	var (
		contactresource1 = "externalcontact-contact1"
		title1           = "integration team"

		firstname2  = "jean"
		middlename2 = "jacques"
		lastname2   = "dupont"
		title2      = "integration team"

		whatsappPhoneDisplay     = "+33 1 00 00 00 01"
		whatsappPhoneE164        = "+33100000001"
		whatsappPhoneCountryCode = "FR"
		whatsappPhoneDisplayName = "whatsappName"

		facebookScopedid    = "facebookScopedid"
		facebookDisplayname = "facebookDisplayname"
	)
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { TestAccPreCheck(t) },
		ProviderFactories: ProviderFactories,
		Steps: []resource.TestStep{
			{
				// Create
				Config: generateBasicExternalContactResource(
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
					whatsappPhoneDisplay,
					whatsappPhoneE164,
					whatsappPhoneCountryCode,
					whatsappPhoneDisplayName,
					facebookScopedid,
					facebookDisplayname,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_externalcontacts_contact."+contactresource1, "title", title2),
					resource.TestCheckResourceAttr("genesyscloud_externalcontacts_contact."+contactresource1, "first_name", firstname2),
					resource.TestCheckResourceAttr("genesyscloud_externalcontacts_contact."+contactresource1, "middle_name", middlename2),
					resource.TestCheckResourceAttr("genesyscloud_externalcontacts_contact."+contactresource1, "last_name", lastname2),
					resource.TestCheckResourceAttr("genesyscloud_externalcontacts_contact."+contactresource1, "whatsapp_id.0.phone_number.0.display", whatsappPhoneDisplay),
					resource.TestCheckResourceAttr("genesyscloud_externalcontacts_contact."+contactresource1, "whatsapp_id.0.phone_number.0.e164", whatsappPhoneE164),
					resource.TestCheckResourceAttr("genesyscloud_externalcontacts_contact."+contactresource1, "whatsapp_id.0.phone_number.0.country_code", whatsappPhoneCountryCode),
					resource.TestCheckResourceAttr("genesyscloud_externalcontacts_contact."+contactresource1, "whatsapp_id.0.display_name", whatsappPhoneDisplayName),
					resource.TestCheckResourceAttr("genesyscloud_externalcontacts_contact."+contactresource1, "facebook_id.0.ids.0.scoped_id", facebookScopedid),
					resource.TestCheckResourceAttr("genesyscloud_externalcontacts_contact."+contactresource1, "facebook_id.0.display_name", facebookDisplayname),
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_externalcontacts_contact." + contactresource1,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifyDivisionsDestroyed,
	})
}

func generateBasicExternalContactResource(resourceID string, title string) string {
	return fmt.Sprintf(`resource "genesyscloud_externalcontacts_contact" "%s" {
		title = "%s"
	}`, resourceID, title)
}

func generateFullExternalContactResource(
	resourceID string,
	firstname string,
	middlename string,
	lastname string,
	title string,
	whatssappDisplay string,
	whatssappE164 string,
	whatssappCountrycode string,
	whatssappDisplayname string,
	facebookScopeid string,
	facebookDisplayname string) string {
	return fmt.Sprintf(`resource "genesyscloud_externalcontacts_contact" "%s" {
		first_name = "%s"
		middle_name = "%s"
		last_name = "%s"
		title = "%s"
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
	}
	`, resourceID, firstname, middlename, lastname, title, whatssappDisplay, whatssappE164, whatssappCountrycode, whatssappDisplayname,
		facebookScopeid, facebookDisplayname)
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
		} else if isStatus404(resp) {
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
