package outbound_contact_list

import (
	"fmt"
	"strconv"
	"testing"

	gcloud "terraform-provider-genesyscloud/genesyscloud"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const nullValue = "null"
const falseValue = "false"
const trueValue = "true"

func TestAccDataSourceOutboundContactList(t *testing.T) {

	var (
		resourceId      = "contact_list"
		dataSourceId    = "contact_list_data"
		contactListName = "Contact List " + uuid.NewString()
	)
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { gcloud.TestAccPreCheck(t) },
		ProviderFactories: gcloud.GetProviderFactories(providerResources, providerDataSources),
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
					GeneratePhoneColumnsBlock(
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
