package genesyscloud

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"strconv"
	"testing"
)

func TestAccDataSourceOutboundContactListFilter(t *testing.T) {
	var (
		resourceId            = "clf"
		contactListResourceId = "contact_list"
		dataSourceId          = "clf_data"
		contactListName       = "Contact List " + uuid.NewString()
		contactListFilterName = "Contact List Filter " + uuid.NewString()
		column1               = "Phone"
		column2               = "Zipcode"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: generateOutboundContactList(
					contactListResourceId,
					contactListName,
					"",
					"",
					[]string{},
					[]string{strconv.Quote(column1), strconv.Quote(column2)},
					"",
					"",
					"",
					"",
					generatePhoneColumnsBlock(
						column1,
						"cell",
						"",
					),
				) + generateOutboundContactListFilter(
					resourceId,
					contactListFilterName,
					"genesyscloud_outbound_contact_list."+contactListResourceId+".id",
					"AND",
					generateOutboundContactListFilterClause(
						"",
						generateOutboundContactListFilterPredicates(
							column1,
							"numeric",
							"EQUALS",
							"+12345123456",
							"",
							"",
						),
					),
				) + generateOutboundContactListFilterDataSource(
					dataSourceId,
					contactListFilterName,
					"genesyscloud_outbound_contactlistfilter."+resourceId),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_contactlistfilter."+resourceId, "id",
						"data.genesyscloud_outbound_contactlistfilter."+dataSourceId, "id"),
				),
			},
		},
	})
}

func generateOutboundContactListFilterDataSource(id string, name string, dependsOn string) string {
	return fmt.Sprintf(`
data "genesyscloud_outbound_contactlistfilter" "%s" {
	name       = "%s"
	depends_on = [%s]
}
`, id, name, dependsOn)
}
