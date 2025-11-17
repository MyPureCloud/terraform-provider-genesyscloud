package ai_studio_summary_setting

import (
	"fmt"
	"testing"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

/*
Test Class for the ai studio summary setting Data Source
*/

func TestAccDataSourceAiStudioSummarySetting(t *testing.T) {
	//t.Parallel()
	var (
		aiStudioSummarySettingDataLabel = "data-aiStudioSummarySetting"

		aiStudioSummarySettingResourceLabel = "resource-aiStudioSummarySetting"
		resourceName                        = "test-summary-setting"
		name                                = "summary setting test"
		language                            = "en-au"
		summaryType                         = "Concise"
		settingType                         = "Basic"
		format                              = "BulletPoints"
		maskPii                             = "true"
		participantLabelInternal            = "Advisor"
		participantLabelExternal            = "Member"
		predefinedInsights0                 = "ReasonForCall"
		predefinedInsights1                 = "Resolution"
		prompt                              = "Summaries should be no more then 300 characters and include 3 dot points of key information"
	)
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create Summary Setting
				Config: GenerateFullAiStudioSummarySettingResource(aiStudioSummarySettingResourceLabel, name, language, summaryType, settingType, format, maskPii, participantLabelInternal, participantLabelExternal, predefinedInsights0, predefinedInsights1, prompt) + generateAiStudioSummarySettingDataSource(aiStudioSummarySettingDataLabel, resourceName, "genesyscloud_ai_studio_summary_setting."+aiStudioSummarySettingResourceLabel),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"data.genesyscloud_ai_studio_summary_setting."+aiStudioSummarySettingDataLabel, "id",
						"genesyscloud_ai_studio_summary_setting."+aiStudioSummarySettingResourceLabel, "id",
					),
				),
			},
		},
	})
}

func generateAiStudioSummarySettingDataSource(resourceName string, name string, dependsOn string) string {
	return fmt.Sprintf(`data "genesyscloud_ai_studio_summary_setting" "%s" {
		name = "%s"
		depends_on = [%s]
	}
	`, resourceName, name, dependsOn)
}
