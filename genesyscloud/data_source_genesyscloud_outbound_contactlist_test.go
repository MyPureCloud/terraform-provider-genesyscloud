package genesyscloud

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceOutboundContactList(t *testing.T) {
	var (
		resourceId      = "contact_list"
		dataSourceId    = "contact_list_data"
		contactListName = "Contact List " + uuid.NewString()
	)
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { TestAccPreCheck(t) },
		ProviderFactories: ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: GenerateOutboundContactList(
					resourceId,
					contactListName,
					nullValue,  // divisionId
					nullValue,  // previewModeColumnName
					[]string{}, // previewModeAcceptedValues
					[]string{strconv.Quote("Cell")},
					falseValue, // automaticTimeZoneMapping
					nullValue,  // zipCodeColumnName
					nullValue,  // attemptLimitId
					generatePhoneColumnsBlock(
						"Cell",
						"cell",
						nullValue,
					),
				) + generateOutboundContactListDataSource(
					dataSourceId,
					contactListName,
					"genesyscloud_outbound_contact_list."+resourceId,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.genesyscloud_outbound_contact_list."+dataSourceId, "id",
						"genesyscloud_outbound_contact_list."+resourceId, "id"),
				),
			},
		},
	})
}

func generateOutboundContactListDataSource(id string, name string, dependsOn string) string {
	return fmt.Sprintf(`
data "genesyscloud_outbound_contact_list" "%s" {
	name = "%s"
	depends_on = [%s]
}
`, id, name, dependsOn)
}
