package external_contacts

import (
	"fmt"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

/*
Test Class for the External Contacts Data Source
*/
func TestAccDataSourceExternalContacts(t *testing.T) {
	var (
		uniqueStr           = uuid.NewString()
		externalContactData = "data-externalContact"
		search              = "john-" + uniqueStr

		externalContactResource = "resource-externalContact"
		title                   = "integrator staff"
		firstname               = "john-" + uniqueStr
		middlename              = "jacques"
		lastname                = "dupont"

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

		surveyoptout      = "false"
		externalsystemurl = "https://externalsystemurl.com"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create external contact with an lastname and others property
				Config: generateFullExternalContactResource(
					externalContactResource,
					firstname,
					middlename,
					lastname,
					title,
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
