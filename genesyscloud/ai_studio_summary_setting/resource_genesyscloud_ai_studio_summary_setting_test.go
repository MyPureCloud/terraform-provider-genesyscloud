package ai_studio_summary_setting

import (
	"fmt"
	"testing"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

/*
The resource_genesyscloud_ai_studio_summary_setting_test.go contains all of the test cases for running the resource
tests for ai_studio_summary_setting.
*/

func TestAccResourceAiStudioSummarySetting(t *testing.T) {
	//t.Parallel()
	var (
		resourceName             = "test-summary-setting"
		name                     = "summary setting test"
		language                 = "en-au"
		summaryType              = "Concise"
		settingType              = "Basic"
		format                   = "BulletPoints"
		maskPii                  = "true"
		participantLabelInternal = "Advisor"
		participantLabelExternal = "Member"
		predefinedInsights       = "ReasonForCall"
		prompt                   = "Summaries should be no more then 300 characters and include 3 dot points of key information"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create
				Config: GenerateBasicSummarySettingResource(resourceName, name, language, ""),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_ai_studio_summary_setting."+resourceName, "name", name),
					resource.TestCheckResourceAttr("genesyscloud_ai_studio_summary_setting."+resourceName, "language", language),
					resource.TestCheckResourceAttr("genesyscloud_ai_studio_summary_setting."+resourceName, "summary_type", summaryType),
					resource.TestCheckResourceAttr("genesyscloud_ai_studio_summary_setting."+resourceName, "setting_type", settingType),
					resource.TestCheckResourceAttr("genesyscloud_ai_studio_summary_setting."+resourceName, "format", format),
					resource.TestCheckResourceAttr("genesyscloud_ai_studio_summary_setting."+resourceName, "mask_p_i_i", maskPii),
					resource.TestCheckResourceAttr("genesyscloud_ai_studio_summary_setting."+resourceName, "participant_labels.internal", participantLabelInternal),
					resource.TestCheckResourceAttr("genesyscloud_ai_studio_summary_setting."+resourceName, "participant_labels.external", participantLabelExternal),
					resource.TestCheckResourceAttr("genesyscloud_ai_studio_summary_setting."+resourceName, "predefined_insights.0", predefinedInsights),
					resource.TestCheckResourceAttr("genesyscloud_ai_studio_summary_setting."+resourceName, "prompt", prompt),
				),
			},
			{
				// Update
				Config: GenerateFullAiStudioSummarySettingResource(resourceName, name, language, summaryType, settingType, format, maskPii, participantLabelInternal, participantLabelExternal, predefinedInsights, prompt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_ai_studio_summary_setting."+resourceName, "name", name),
					resource.TestCheckResourceAttr("genesyscloud_ai_studio_summary_setting."+resourceName, "language", language),
					resource.TestCheckResourceAttr("genesyscloud_ai_studio_summary_setting."+resourceName, "summary_type", summaryType),
					resource.TestCheckResourceAttr("genesyscloud_ai_studio_summary_setting."+resourceName, "setting_type", settingType),
					resource.TestCheckResourceAttr("genesyscloud_ai_studio_summary_setting."+resourceName, "format", format),
					resource.TestCheckResourceAttr("genesyscloud_ai_studio_summary_setting."+resourceName, "mask_p_i_i", maskPii),
					resource.TestCheckResourceAttr("genesyscloud_ai_studio_summary_setting."+resourceName, "participant_labels.internal", participantLabelInternal),
					resource.TestCheckResourceAttr("genesyscloud_ai_studio_summary_setting."+resourceName, "participant_labels.external", participantLabelExternal),
					resource.TestCheckResourceAttr("genesyscloud_ai_studio_summary_setting."+resourceName, "predefined_insights.0", predefinedInsights),
					resource.TestCheckResourceAttr("genesyscloud_ai_studio_summary_setting."+resourceName, "prompt", prompt),
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_ai_studio_summary_setting." + resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifyAiStudioSummarySettingDestroyed,
	})
}

func GenerateFullAiStudioSummarySettingResource(resourceName, name, language, summaryType, settingType, format, maskPii, participantLabelInternal, participantLabelExternal, predefinedInsights, prompt string) string {
	return fmt.Sprintf(`resource "genesyscloud_ai_studio_summary_setting" "%s" {
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
		predefined_insights = [ "%s", "%s" ]
		prompt = "%s"
	}
	`, resourceName, name, language, summaryType, settingType, format, maskPii, participantLabelInternal, participantLabelExternal, predefinedInsights, prompt)
}

func testVerifyAiStudioSummarySettingDestroyed(state *terraform.State) error {
	return nil
}
