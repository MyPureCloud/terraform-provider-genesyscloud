package genesyscloud

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"strconv"
	"testing"
)

func TestAccDataSourceOutboundContactList(t *testing.T) {
	var (
		resourceId      = "contact_list"
		dataSourceId    = "contact_list_data"
		contactListName = "Contact List " + uuid.NewString()
	)
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: generateOutboundContactList(
					resourceId,
					contactListName,
					"",
					"",
					[]string{},
					[]string{strconv.Quote("Cell")},
					falseValue,
					"",
					"",
					generatePhoneColumnsBlock("Cell",
						"cell",
						"",
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
