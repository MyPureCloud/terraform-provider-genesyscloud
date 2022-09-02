package genesyscloud

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"strconv"
	"testing"
)

func TestAccDataSourceOutboundMessagingCampaign(t *testing.T) {
	var (
		resourceId          = "campaign"
		dataSourceId        = "campaign_data"
		digitalCampaignName = "Test Digital Campaign " + uuid.NewString()

		clfResourceId         = "clf"
		clfName               = "Test CLF " + uuid.NewString()
		contactListResourceId = "contact_list"
		contactListName       = "Test Contact List " + uuid.NewString()
		column1               = "phone"
		column2               = "zipcode"

		callableTimeSetId = "5654dc1a-874d-439f-83af-f3c1271dcf7c"

		contactListResource = generateOutboundContactList(
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
				column1,
			),
		)

		contactListFilterResource = generateOutboundContactListFilter(
			clfResourceId,
			clfName,
			"genesyscloud_outbound_contact_list."+contactListResourceId+".id",
			"",
			generateOutboundContactListFilterClause(
				"",
				generateOutboundContactListFilterPredicates(
					column1,
					"alphabetic",
					"EQUALS",
					"XYZ",
					"",
					"",
				),
			),
		)
	)
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: contactListResource + contactListFilterResource + generateOutboundMessagingCampaignResource(
					resourceId,
					digitalCampaignName,
					"genesyscloud_outbound_contact_list."+contactListResourceId+".id",
					"",
					"10",
					falseValue,
					callableTimeSetId,
					[]string{},
					[]string{"genesyscloud_outbound_contactlistfilter." + clfResourceId + ".id"},
					generateOutboundMessagingCampaignSmsConfig(
						column1,
						column1,
						"+19197050640",
					),
					generateOutboundMessagingCampaignContactSort(
						column1,
						"",
						"",
					),
					generateOutboundMessagingCampaignContactSort(
						column2,
						"DESC",
						trueValue,
					),
				) + generateOutboundMessagingCampaignDataSource(
					dataSourceId,
					digitalCampaignName,
					"genesyscloud_outbound_messagingcampaign."+resourceId,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.genesyscloud_outbound_messagingcampaign."+dataSourceId, "id",
						"genesyscloud_outbound_messagingcampaign."+resourceId, "id"),
				),
			},
		},
	})
}

func generateOutboundMessagingCampaignDataSource(id string, name string, dependsOn string) string {
	return fmt.Sprintf(`
data "genesyscloud_outbound_messagingcampaign" "%s" {
	name = "%s"
	depends_on = [%s]
}
`, id, name, dependsOn)
}
