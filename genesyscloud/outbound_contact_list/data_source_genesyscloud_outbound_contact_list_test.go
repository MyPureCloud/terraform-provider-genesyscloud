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
		resourceId      = "contact_list"
		dataSourceId    = "contact_list_data"
		contactListName = "Contact List " + uuid.NewString()
	)
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: GenerateOutboundContactList(
					resourceId,
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
					dataSourceId,
					contactListName,
					resourceName+"."+resourceId,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data."+resourceName+"."+dataSourceId, "id",
						resourceName+"."+resourceId, "id"),
				),
			},
		},
	})
}

func generateOutboundContactListDataSource(id string, name string, dependsOn string) string {
	return fmt.Sprintf(`
data "%s" "%s" {
	name = "%s"
	depends_on = [%s]
}
`, resourceName, id, name, dependsOn)
}
