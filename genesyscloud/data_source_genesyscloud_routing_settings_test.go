package genesyscloud

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceRoutingSettings(t *testing.T) {
	var (
		settingsResource   = "test-settings"
		settingsDataSource = "test-settings-data"
		transcription      = "EnabledQueueFlow"
		confidence         = "1"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { TestAccPreCheck(t) },
		ProviderFactories: GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Search by contact center
				Config: generateRoutingSettingsWithCustomAttrs(
					settingsResource,
					trueValue,
					generateSettingsContactCenter(falseValue),
					generateSettingsTranscription(transcription, confidence, trueValue, falseValue),
				) + generateRoutingSettingsDataSource(
					settingsDataSource,
					trueValue,
					"genesyscloud_routing_settings."+settingsResource,
					generateSettingsContactCenter(falseValue),
					generateSettingsTranscription(transcription, confidence, trueValue, falseValue),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.genesyscloud_routing_settings."+settingsDataSource, "reset_agent_on_presence", "genesyscloud_routing_settings."+settingsResource, "reset_agent_on_presence"),
					resource.TestCheckResourceAttrPair("data.genesyscloud_routing_settings."+settingsDataSource, "contactcenter.remove_skills_from_blind_transfer", "genesyscloud_routing_settings."+settingsResource, "contactcenter.remove_skills_from_blind_transfer"),
					resource.TestCheckResourceAttrPair("data.genesyscloud_routing_settings."+settingsDataSource, "transcription.0.transcription", "genesyscloud_routing_settings."+settingsResource, "transcription.0.transcription"),
					resource.TestCheckResourceAttrPair("data.genesyscloud_routing_settings."+settingsDataSource, "transcription.0.transcription_confidence_threshold", "genesyscloud_routing_settings."+settingsResource, "transcription.0.transcription_confidence_threshold"),
					resource.TestCheckResourceAttrPair("data.genesyscloud_routing_settings."+settingsDataSource, "transcription.0.low_latency_transcription_enabled", "genesyscloud_routing_settings."+settingsResource, "transcription.0.low_latency_transcription_enabled"),
					resource.TestCheckResourceAttrPair("data.genesyscloud_routing_settings."+settingsDataSource, "transcription.0.content_search_enabled", "genesyscloud_routing_settings."+settingsResource, "transcription.0.content_search_enabled"),
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
