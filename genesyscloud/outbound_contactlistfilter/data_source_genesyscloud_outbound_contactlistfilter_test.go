package outbound_contactlistfilter

import (
	"fmt"
	"strconv"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"

	obContactList "terraform-provider-genesyscloud/genesyscloud/outbound_contact_list"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceOutboundContactListFilter(t *testing.T) {

	var (
		resourceLabel            = "clf"
		contactListResourceLabel = "contact_list"
		dataSourceLabel          = "clf_data"
		contactListName          = "Contact List " + uuid.NewString()
		contactListFilterName    = "Contact List Filter " + uuid.NewString()
		column1                  = "Phone"
		column2                  = "Zipcode"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: obContactList.GenerateOutboundContactList(
					contactListResourceLabel,
					contactListName,
					util.NullValue,
					util.NullValue,
					[]string{},
					[]string{strconv.Quote(column1), strconv.Quote(column2)},
					util.NullValue,
					util.NullValue,
					util.NullValue,
					"",
					obContactList.GeneratePhoneColumnsBlock(
						column1,
						"cell",
						util.NullValue,
					),
				) + GenerateOutboundContactListFilter(
					resourceLabel,
					contactListFilterName,
					"genesyscloud_outbound_contact_list."+contactListResourceLabel+".id",
					"AND",
					GenerateOutboundContactListFilterClause(
						"",
						GenerateOutboundContactListFilterPredicates(
							column1,
							"numeric",
							"EQUALS",
							"+12345123456",
							"",
							"",
						),
					),
				) + generateOutboundContactListFilterDataSource(
					dataSourceLabel,
					contactListFilterName,
					"genesyscloud_outbound_contactlistfilter."+resourceLabel),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_contactlistfilter."+resourceLabel, "id",
						"data.genesyscloud_outbound_contactlistfilter."+dataSourceLabel, "id"),
				),
			},
		},
	})
}

func generateOutboundContactListFilterDataSource(dataSourceLabel string, name string, dependsOn string) string {
	return fmt.Sprintf(`
data "genesyscloud_outbound_contactlistfilter" "%s" {
	name       = "%s"
	depends_on = [%s]
}
`, dataSourceLabel, name, dependsOn)
}
