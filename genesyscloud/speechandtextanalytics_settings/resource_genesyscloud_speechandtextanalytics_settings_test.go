package speechandtextanalytics_settings

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
)

/*
NOTE: genesyscloud_speechandtextanalytics_settings is a singleton, organization-wide resource.
These acceptance tests mutate real org-level Speech & Text Analytics settings. There is no
CheckDestroy because the settings object always exists and cannot be deleted.
*/

func TestAccResourceSpeechAndTextAnalyticsSettings(t *testing.T) {
	var (
		resourceLabel = "sta_settings"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create
				Config: generateSpeechAndTextAnalyticsSettingsResource(
					resourceLabel,
					[]string{"en-US"},
					util.TrueValue,
					util.FalseValue,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "expected_dialects.#", "1"),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "expected_dialects.0", "en-US"),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "text_analytics_enabled", util.TrueValue),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "agent_empathy_enabled", util.FalseValue),
				),
			},
			{
				// Update
				Config: generateSpeechAndTextAnalyticsSettingsResource(
					resourceLabel,
					[]string{"en-US", "es-US"},
					util.FalseValue,
					util.TrueValue,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "expected_dialects.#", "2"),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "text_analytics_enabled", util.FalseValue),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "agent_empathy_enabled", util.TrueValue),
				),
			},
			{
				// Import/Read
				ResourceName:      ResourceType + "." + resourceLabel,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func generateSpeechAndTextAnalyticsSettingsResource(resourceLabel string, dialects []string, textAnalyticsEnabled string, agentEmpathyEnabled string) string {
	dialectsStr := ""
	for _, d := range dialects {
		dialectsStr += fmt.Sprintf("%q, ", d)
	}
	return fmt.Sprintf(`
resource "%s" "%s" {
	expected_dialects      = [%s]
	text_analytics_enabled = %s
	agent_empathy_enabled  = %s
}
`, ResourceType, resourceLabel, dialectsStr, textAnalyticsEnabled, agentEmpathyEnabled)
}
