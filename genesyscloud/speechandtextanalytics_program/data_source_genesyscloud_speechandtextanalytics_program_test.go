package speechandtextanalytics_program

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
)

func TestAccDataSourceSpeechAndTextAnalyticsProgram(t *testing.T) {
	t.Parallel()

	var (
		resourceLabel   = "program_" + uuid.NewString()
		dataSourceLabel = "program_data_" + uuid.NewString()
		name            = "tfacc-program-" + uuid.NewString()
		description     = "Terraform acceptance test program for data source"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: generateSpeechAndTextAnalyticsProgramResource(
					resourceLabel,
					name,
					description,
					[]string{"tfacc"},
				) + generateSpeechAndTextAnalyticsProgramDataSource(
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

func generateSpeechAndTextAnalyticsProgramDataSource(dataSourceLabel, name, dependsOn string) string {
	return fmt.Sprintf(`
data "%s" "%s" {
  name       = %q
  depends_on = [%s]
}
`, ResourceType, dataSourceLabel, name, dependsOn)
}
