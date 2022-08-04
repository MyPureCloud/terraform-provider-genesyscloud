package genesyscloud

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"strings"
	"testing"
)

func TestAccDataSourceRoutingSettings(t *testing.T) {
	var (
		settingsResource   = "test-settings"
		settingsDataSource = "test-settings-data"
		transcription      = "EnabledQueueFlow"
		confidence         = "1"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				// Search by transcription
				Config: generateRoutingSettingsWithCustomAttrs(
					settingsResource,
					nullValue,
					generateSettingsTranscription(transcription, confidence, trueValue, trueValue),
				) + generateRoutingSettingsDataSource(
					settingsDataSource,
					nullValue,
					"genesyscloud_routing_settings."+settingsResource,
					"", //generateSettingsTranscription(transcription, confidence, trueValue, trueValue),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.genesys_routing_settings."+settingsDataSource, "id", "genesyscloud_routing_settings."+settingsResource, "id"),
				),
			},
		},
	})
}

func generateRoutingSettingsDataSource(
	resourceID string,
	resetAgentScoreOnPresenceChange string,
// Must explicitly use depends_on in terraform v0.13 when a data source references a resource
// Fixed in v0.14 https://github.com/hashicorp/terraform/pull/26284
	dependsOnResource string,
	attrs ...string) string {
	return fmt.Sprintf(`data "genesyscloud_routing_settings" "%s" {
		reset_agent_on_presence_change = %s
		%s
        depends_on=[%s]
	}
	`, resourceID, resetAgentScoreOnPresenceChange, strings.Join(attrs, "\n"), dependsOnResource)
}
