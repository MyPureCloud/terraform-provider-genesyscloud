package genesyscloud

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"testing"
)

func TestAccDataSourceResponseManagementResponses(t *testing.T) {
	var ()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				// Search by name
				Config: generateResponseManagementResponsesResource(
					"",
					"",
					[]string{},
					"",
					"",
					generateTextsBlock(
						"",
						"",
					),
					generateSubstitutionsBlock(
						"",
						"",
					),
					generateMessagingTemplateBlock(
						generateWhatsappBlock(
							"",
							"",
							"",
						),
					),
				) + generateResponseManagementResponsesDataSource(),
				Check: resource.ComposeTestCheckFunc(
					//resource.TestCheckResourceAttrPair(),
				),
			},
		},
	})
}

func generateResponseManagementResponsesDataSource() string {
	return fmt.Sprintf("")
}
