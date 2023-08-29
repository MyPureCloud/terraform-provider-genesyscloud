package outbound

import (
	"fmt"
	"strconv"
	"testing"

	gcloud "terraform-provider-genesyscloud/genesyscloud"
	obContactList "terraform-provider-genesyscloud/genesyscloud/outbound_contact_list"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
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
		PreCheck:          func() { gcloud.TestAccPreCheck(t) },
		ProviderFactories: gcloud.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: obContactList.GenerateOutboundContactList(
					contactListResourceId,
					contactListName,
					nullValue,
					nullValue,
					[]string{},
					[]string{strconv.Quote(column1), strconv.Quote(column2)},
					nullValue,
					nullValue,
					nullValue,
					"",
					obContactList.GeneratePhoneColumnsBlock(
						column1,
						"cell",
						nullValue,
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
