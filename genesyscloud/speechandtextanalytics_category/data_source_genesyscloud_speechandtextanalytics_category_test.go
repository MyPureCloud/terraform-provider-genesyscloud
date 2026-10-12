package speechandtextanalytics_category

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
)

func TestAccDataSourceSpeechAndTextAnalyticsCategory(t *testing.T) {
	t.Parallel()

	var (
		resourceLabel   = "category_" + uuid.NewString()
		dataSourceLabel = "category_data_" + uuid.NewString()

		name        = "tfacc-category-ds-" + uuid.NewString()
		description = "Terraform acceptance test category data source"

		interactionType = "Voice"

		// The API requires "inverted" and "occurrence" on every operand.
		criteria = `jsonencode({
    type       = "OperandGroup"
    inverted   = false
    occurrence = 1
    operands = [
      {
        type       = "OperandGroup"
        inverted   = false
        occurrence = 1
        operands = [
          {
            type       = "Term"
            inverted   = false
            occurrence = 1
            term = {
              word            = "refund"
              participantType = "External"
            }
          }
        ]
      }
    ]
  })`
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: generateCategoryResource(
					resourceLabel,
					name,
					description,
					interactionType,
					criteria,
				) + generateCategoryDataSource(
					dataSourceLabel,
					name,
					ResourceType+"."+resourceLabel,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"data."+ResourceType+"."+dataSourceLabel, "id",
						ResourceType+"."+resourceLabel, "id",
					),
				),
			},
		},
	})
}

func generateCategoryDataSource(dataSourceLabel, name, dependsOnResource string) string {
	return fmt.Sprintf(`
data "%s" "%s" {
  name       = %q
  depends_on = [%s]
}
`, ResourceType, dataSourceLabel, name, dependsOnResource)
}
