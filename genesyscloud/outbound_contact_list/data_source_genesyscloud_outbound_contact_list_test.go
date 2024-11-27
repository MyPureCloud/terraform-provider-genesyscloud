package outbound_contact_list

import (
	"fmt"
	"strconv"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceOutboundContactList(t *testing.T) {

	var (
		resourceLabel   = "contact_list"
		dataSourceLabel = "contact_list_data"
		contactListName = "Contact List " + uuid.NewString()
	)
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: GenerateOutboundContactList(
					resourceLabel,
					contactListName,
					util.NullValue, // divisionId
					util.NullValue, // previewModeColumnName
					[]string{},     // previewModeAcceptedValues
					[]string{strconv.Quote("Cell")},
					util.FalseValue, // automaticTimeZoneMapping
					util.NullValue,  // zipCodeColumnName
					util.NullValue,  // attemptLimitId
					GeneratePhoneColumnsBlock(
						"Cell",
						"cell",
						util.NullValue,
					),
				) + generateOutboundContactListDataSource(
					dataSourceLabel,
					contactListName,
					ResourceType+"."+resourceLabel,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data."+ResourceType+"."+dataSourceLabel, "id",
						ResourceType+"."+resourceLabel, "id"),
				),
			},
		},
	})
}

func generateOutboundContactListDataSource(dataSourceLabel string, name string, dependsOn string) string {
	return fmt.Sprintf(`
data "%s" "%s" {
	name = "%s"
	depends_on = [%s]
}
`, ResourceType, dataSourceLabel, name, dependsOn)
}
