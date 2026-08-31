package speechandtextanalytics_settings

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
)

func TestAccDataSourceSpeechAndTextAnalyticsSettings(t *testing.T) {
	var (
		resourceLabel   = "sta_settings"
		dataSourceLabel = "sta_settings_data"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: generateSpeechAndTextAnalyticsSettingsResource(
					resourceLabel,
					[]string{"en-US"},
					util.TrueValue,
					util.FalseValue,
				) + generateSpeechAndTextAnalyticsSettingsDataSource(dataSourceLabel, resourceLabel),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"data."+ResourceType+"."+dataSourceLabel, "text_analytics_enabled",
						ResourceType+"."+resourceLabel, "text_analytics_enabled",
					),
					resource.TestCheckResourceAttrPair(
						"data."+ResourceType+"."+dataSourceLabel, "agent_empathy_enabled",
						ResourceType+"."+resourceLabel, "agent_empathy_enabled",
					),
					resource.TestCheckResourceAttrPair(
						"data."+ResourceType+"."+dataSourceLabel, "expected_dialects.#",
						ResourceType+"."+resourceLabel, "expected_dialects.#",
					),
				),
			},
		},
	})
}

func generateSpeechAndTextAnalyticsSettingsDataSource(dataSourceLabel string, resourceLabel string) string {
	return fmt.Sprintf(`
data "%s" "%s" {
	depends_on = [%s.%s]
}
`, ResourceType, dataSourceLabel, ResourceType, resourceLabel)
}
