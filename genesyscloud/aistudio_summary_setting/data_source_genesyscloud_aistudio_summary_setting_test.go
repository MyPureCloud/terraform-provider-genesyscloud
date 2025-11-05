package aistudio_summary_setting

import (
	"fmt"
	"testing"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

/*
Test Class for the aistudio summary setting Data Source
*/

func TestAccDataSourceAistudioSummarySetting(t *testing.T) {
	//t.Parallel()
	var (
		aistudioSummarySettingDataLabel = "data-aistudioSummarySetting"

		aiStudioSummarySettingResourceLabel = "resource-aistudioSummarySetting"
		resourceName                        = "test-summary-setting"
		name                                = "summary setting test"
		language                            = "en-au"
		summaryType                         = "Concise"
		settingType                         = "Basic"
		format                              = "BulletPoints"
		maskPii                             = "true"
		participantLabelInternal            = "Advisor"
		participantLabelExternal            = "Member"
		predefinedInsightsLabel             = "label1"
		predefinedInsightsDescription       = "description1"
		prompt                              = "Summaries should be no more then 300 characters and include 3 dot points of key information"
		externalSystemUrl                   = "https://externalsystemurl.com"
	)
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create Summary Setting
				Config: GenerateFullAistudioSummarySettingResource(aiStudioSummarySettingResourceLabel, name, language, summaryType, settingType, format, maskPii, participantLabelInternal, participantLabelExternal, predefinedInsightsLabel, predefinedInsightsDescription, prompt, externalSystemUrl) + generateAistudioSummarySettingDataSource(aistudioSummarySettingDataLabel, resourceName, "genesyscloud_aistudio_summary_setting."+aiStudioSummarySettingResourceLabel),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"data.genesyscloud_aistudio_summary_setting."+aistudioSummarySettingDataLabel, "id",
						"genesyscloud_aistudio_summary_setting."+aiStudioSummarySettingResourceLabel, "id",
					),
				),
			},
		},
	})
}

func generateAistudioSummarySettingDataSource(resresourceName string, resourceName string, dependsOn string) string {
	return fmt.Sprintf(`data "genesyscloud_aistudio_summary_setting" "%s" {
		name = "%s"
		depends_on = [%s]
	}
	`, resresourceName, resourceName, dependsOn)
}
