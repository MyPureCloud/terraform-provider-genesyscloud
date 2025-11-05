package aistudio_summary_setting

import (
	"fmt"
	"testing"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

/*
The resource_genesyscloud_aistudio_summary_setting_test.go contains all of the test cases for running the resource
tests for aistudio_summary_setting.
*/

func TestAccResourceAistudioSummarySetting(t *testing.T) {
	//t.Parallel()
	var (
		resourceName                  = "test-summary-setting"
		name                          = "summary setting test"
		language                      = "en-au"
		summaryType                   = "Concise"
		settingType                   = "Basic"
		format                        = "BulletPoints"
		maskPii                       = "true"
		participantLabelInternal      = "Advisor"
		participantLabelExternal      = "Member"
		predefinedInsightsLabel       = "label1"
		predefinedInsightsDescription = "description1"
		prompt                        = "Summaries should be no more then 300 characters and include 3 dot points of key information"
		externalSystemUrl             = "https://externalsystemurl.com"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create
				Config: GenerateBasicSummarySettingResource(resourceName, name, language, ""),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_aistudio_summary_setting."+resourceName, "name", name),
					resource.TestCheckResourceAttr("genesyscloud_aistudio_summary_setting."+resourceName, "language", language),
					resource.TestCheckResourceAttr("genesyscloud_aistudio_summary_setting."+resourceName, "summary_type", summaryType),
					resource.TestCheckResourceAttr("genesyscloud_aistudio_summary_setting."+resourceName, "setting_type", settingType),
					resource.TestCheckResourceAttr("genesyscloud_aistudio_summary_setting."+resourceName, "format", format),
					resource.TestCheckResourceAttr("genesyscloud_aistudio_summary_setting."+resourceName, "mask_p_i_i", maskPii),
					resource.TestCheckResourceAttr("genesyscloud_aistudio_summary_setting."+resourceName, "participant_labels.internal", participantLabelInternal),
					resource.TestCheckResourceAttr("genesyscloud_aistudio_summary_setting."+resourceName, "participant_labels.external", participantLabelExternal),
					resource.TestCheckResourceAttr("genesyscloud_aistudio_summary_setting."+resourceName, "predefined_insights.0.label", predefinedInsightsLabel),
					resource.TestCheckResourceAttr("genesyscloud_aistudio_summary_setting."+resourceName, "predefined_insights.0.description", predefinedInsightsDescription),
					resource.TestCheckResourceAttr("genesyscloud_aistudio_summary_setting."+resourceName, "prompt", prompt),
				),
			},
			{
				// Update
				Config: GenerateFullAistudioSummarySettingResource(resourceName, name, language, summaryType, settingType, format, maskPii, participantLabelInternal, participantLabelExternal, predefinedInsightsLabel, predefinedInsightsDescription, prompt, externalSystemUrl),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_aistudio_summary_setting."+resourceName, "name", name),
					resource.TestCheckResourceAttr("genesyscloud_aistudio_summary_setting."+resourceName, "language", language),
					resource.TestCheckResourceAttr("genesyscloud_aistudio_summary_setting."+resourceName, "summary_type", summaryType),
					resource.TestCheckResourceAttr("genesyscloud_aistudio_summary_setting."+resourceName, "setting_type", settingType),
					resource.TestCheckResourceAttr("genesyscloud_aistudio_summary_setting."+resourceName, "format", format),
					resource.TestCheckResourceAttr("genesyscloud_aistudio_summary_setting."+resourceName, "mask_p_i_i", maskPii),
					resource.TestCheckResourceAttr("genesyscloud_aistudio_summary_setting."+resourceName, "participant_labels.internal", participantLabelInternal),
					resource.TestCheckResourceAttr("genesyscloud_aistudio_summary_setting."+resourceName, "participant_labels.external", participantLabelExternal),
					resource.TestCheckResourceAttr("genesyscloud_aistudio_summary_setting."+resourceName, "predefined_insights.0.label", predefinedInsightsLabel),
					resource.TestCheckResourceAttr("genesyscloud_aistudio_summary_setting."+resourceName, "predefined_insights.0.description", predefinedInsightsDescription),
					resource.TestCheckResourceAttr("genesyscloud_aistudio_summary_setting."+resourceName, "prompt", prompt),
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_aistudio_summary_setting." + resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifyAistudioSummarySettingDestroyed,
	})
}

func GenerateFullAistudioSummarySettingResource(resourceName, name, language, summaryType, settingType, format, maskPii, participantLabelInternal, participantLabelExternal, predefinedInsightsLabel, predefinedInsightsDescription, prompt, externalSystemUrl string) string {
	return fmt.Sprintf(`resource "genesyscloud_aistudio_summary_setting" "%s" {
		name             = "%s"
		language         = "%s"
		summary_type     = "%s"
		setting_type     = "%s"
		format           = "%s"
		mask_p_i_i {
		all = %s
		}
		participant_labels {
			internal = "%s"
			external = "%s"
		}
		predefined_insights {
			label       = "%s"
			description = "%s"
			}
		prompt = "%s"
		external_system_url = "%s"
	}`, resourceName, name, language, summaryType, settingType, format, maskPii, participantLabelInternal, participantLabelExternal, predefinedInsightsLabel, predefinedInsightsDescription, prompt, externalSystemUrl)
}

func testVerifyAistudioSummarySettingDestroyed(state *terraform.State) error {
	return nil
}
