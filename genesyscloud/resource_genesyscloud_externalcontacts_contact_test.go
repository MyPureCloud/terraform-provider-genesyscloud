package genesyscloud

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v92/platformclientv2"
)

func TestAccExternalConctactContactBasic(t *testing.T) {
	var (
		contactresource1 = "externalcontact-contact1"
		title1           = "integration team"

		firstname2  = "jean"
		middlename2 = "jacques"
		lastname2   = "dupont"
		title2      = "integration team"
	)
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { TestAccPreCheck(t) },
		ProviderFactories: ProviderFactories,
		Steps: []resource.TestStep{
			{
				// Create
				Config: generateExternalContactBasic(
					contactresource1,
					title1,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_externalcontacts_contact."+contactresource1, "title", title1),
					resource.TestCheckResourceAttr("genesyscloud_externalcontacts_contact."+contactresource1, "firstname", ""),
					resource.TestCheckResourceAttr("genesyscloud_externalcontacts_contact."+contactresource1, "middlename", ""),
					resource.TestCheckResourceAttr("genesyscloud_externalcontacts_contact."+contactresource1, "lastname", ""),
				),
			},
			{
				// Update with a new name and description
				Config: generateExternalContactResource(
					contactresource1,
					firstname2,
					middlename2,
					lastname2,
					title2,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_externalcontacts_contact."+contactresource1, "title", title2),
					resource.TestCheckResourceAttr("genesyscloud_externalcontacts_contact."+contactresource1, "firstName", firstname2),
					resource.TestCheckResourceAttr("genesyscloud_externalcontacts_contact."+contactresource1, "middleName", middlename2),
					resource.TestCheckResourceAttr("genesyscloud_externalcontacts_contact."+contactresource1, "lastName", lastname2),
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

func generateExternalContactBasic(resourceID string, title string) string {
	return generateExternalContactResource(resourceID, nullValue, nullValue, nullValue, title)
}

func generateExternalContactResource(
	resourceID string,
	firstname string,
	middlename string,
	lastname string,
	title string) string {
	return fmt.Sprintf(`resource "genesyscloud_externalcontacts_contact" "%s" {
		firstname = "%s"
		middlename = "%s"
		lastname = "%s"
		title = "%s"
	}
	`, resourceID, firstname, middlename, lastname, title)
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
	// Success. All divisions destroyed
	return nil
}
