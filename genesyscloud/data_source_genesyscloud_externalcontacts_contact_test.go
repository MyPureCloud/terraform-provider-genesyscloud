package genesyscloud

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceExternalContacts(t *testing.T) {
	var (
		externalContactData = "data-externalContact"
		search              = "jean"

		externalContactResource = "resource-externalContact"
		title                   = "integrator staff"
		firstname               = "jean"
		middlename              = "jacques"
		lastname                = "dupont"

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
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { TestAccPreCheck(t) },
		ProviderFactories: ProviderFactories,
		Steps: []resource.TestStep{
			{
				// Create external contact with an lastname and others property
				Config: generateFullExternalContactResource(
					externalContactResource,
					firstname,
					middlename,
					lastname,
					title,
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
				) + generateExternalContactDataSource(
					externalContactData,
					search,
					"genesyscloud_externalcontacts_contact."+externalContactResource,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"data.genesyscloud_externalcontacts_contact."+externalContactData, "id",
						"genesyscloud_externalcontacts_contact."+externalContactResource, "id",
					),
				),
			},
		},
	})
}

func generateExternalContactDataSource(resourceID string, search string, dependsOn string) string {
	return fmt.Sprintf(`data "genesyscloud_externalcontacts_contact" "%s" {
		search = "%s"
		depends_on = [%s]
	}
	`, resourceID, search, dependsOn)
}
