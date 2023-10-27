package genesyscloud

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceRoutingSettingsBasic(t *testing.T) {
	var (
		settingsResource = "settings-basic"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { TestAccPreCheck(t) },
		ProviderFactories: GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create
				Config: generateRoutingSettingsResource(settingsResource, FalseValue),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_routing_settings."+settingsResource, "reset_agent_on_presence_change", FalseValue),
				),
			},
			{
				// Update
				Config: generateRoutingSettingsResource(settingsResource, TrueValue),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_routing_settings."+settingsResource, "reset_agent_on_presence_change", TrueValue),
				),
			},
			{
				// Update
				Config: generateRoutingSettingsResource(settingsResource, FalseValue),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_routing_settings."+settingsResource, "reset_agent_on_presence_change", FalseValue),
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_routing_settings." + settingsResource,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccResourceRoutingSettingsContactCenter(t *testing.T) {
	var (
		settingsResource = "settings-contactCenter"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { TestAccPreCheck(t) },
		ProviderFactories: GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create with contact center
				Config: generateRoutingSettingsWithCustomAttrs(
					settingsResource,
					TrueValue,
					generateSettingsContactCenter(TrueValue),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_routing_settings."+settingsResource, "contactcenter.0.remove_skills_from_blind_transfer", TrueValue),
				),
			},
			{
				// Update contact center
				Config: generateRoutingSettingsWithCustomAttrs(
					settingsResource,
					TrueValue,
					generateSettingsContactCenter(FalseValue),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_routing_settings."+settingsResource, "contactcenter.0.remove_skills_from_blind_transfer", FalseValue),
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_routing_settings." + settingsResource,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccResourceRoutingSettingsTranscription(t *testing.T) {
	var (
		settingsResource = "settings-transcription"
		transcription1   = "Disabled"
		transcription2   = "EnabledQueueFlow"
		confidence1      = "1"
		confidence2      = "2"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { TestAccPreCheck(t) },
		ProviderFactories: GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create with transcription
				Config: generateRoutingSettingsWithCustomAttrs(
					settingsResource,
					TrueValue,
					generateSettingsTranscription(transcription1, confidence1, TrueValue, TrueValue),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_routing_settings."+settingsResource, "transcription.0.transcription", transcription1),
					resource.TestCheckResourceAttr("genesyscloud_routing_settings."+settingsResource, "transcription.0.transcription_confidence_threshold", confidence1),
					resource.TestCheckResourceAttr("genesyscloud_routing_settings."+settingsResource, "transcription.0.low_latency_transcription_enabled", TrueValue),
					resource.TestCheckResourceAttr("genesyscloud_routing_settings."+settingsResource, "transcription.0.content_search_enabled", TrueValue),
				),
			},
			{
				// Update transcription
				Config: generateRoutingSettingsWithCustomAttrs(
					settingsResource,
					TrueValue,
					generateSettingsTranscription(transcription2, confidence2, FalseValue, FalseValue),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_routing_settings."+settingsResource, "transcription.0.transcription", transcription2),
					resource.TestCheckResourceAttr("genesyscloud_routing_settings."+settingsResource, "transcription.0.transcription_confidence_threshold", confidence2),
					resource.TestCheckResourceAttr("genesyscloud_routing_settings."+settingsResource, "transcription.0.low_latency_transcription_enabled", FalseValue),
					resource.TestCheckResourceAttr("genesyscloud_routing_settings."+settingsResource, "transcription.0.content_search_enabled", FalseValue),
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_routing_settings." + settingsResource,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func generateRoutingSettingsResource(
	resourceId string,
	resetAgentScoreOnPresenceChange string) string {
	return fmt.Sprintf(`resource "genesyscloud_routing_settings" "%s"{
		reset_agent_on_presence_change = %s
	}
	`, resourceId, resetAgentScoreOnPresenceChange)
}

func generateRoutingSettingsWithCustomAttrs(
	resourceId string,
	resetAgentScoreOnPresenceChange string,
	attrs ...string) string {
	return fmt.Sprintf(`resource "genesyscloud_routing_settings" "%s" {
		reset_agent_on_presence_change = %s
		%s
	}
	`, resourceId, resetAgentScoreOnPresenceChange, strings.Join(attrs, "\n"))
}

func generateSettingsContactCenter(removeSkillsFromBlindTransfer string) string {
	return fmt.Sprintf(`contactcenter {
		remove_skills_from_blind_transfer = %s
	}
	`, removeSkillsFromBlindTransfer)
}

func generateSettingsTranscription(
	transcription string,
	transcriptionConfidence string,
	lowLatency string,
	contentSearch string) string {
	return fmt.Sprintf(`transcription {
		transcription = "%s"
		transcription_confidence_threshold = %s
		low_latency_transcription_enabled = %s
		content_search_enabled = %s
	}
	`, transcription, transcriptionConfidence, lowLatency, contentSearch)
}
