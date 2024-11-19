package outbound_contact_list_template

import (
	"fmt"
	"strconv"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceOutboundContactListTemplate(t *testing.T) {

	var (
		resourceId      = "contact_list_template"
		dataSourceId    = "contact_list_template_data"
		contactListName = "Contact List Template" + uuid.NewString()
	)
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: GenerateOutboundContactListTemplate(
					resourceId,
					contactListName,
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
				) + generateOutboundContactListTemplateDataSource(
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

func generateOutboundContactListTemplateDataSource(id string, name string, dependsOn string) string {
	return fmt.Sprintf(`
data "%s" "%s" {
	name = "%s"
	depends_on = [%s]
}
`, resourceName, id, name, dependsOn)
}
