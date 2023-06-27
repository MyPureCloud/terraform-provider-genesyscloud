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
				Config: generateRoutingSettingsResource(settingsResource, falseValue),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_routing_settings."+settingsResource, "reset_agent_on_presence_change", falseValue),
				),
			},
			{
				// Update
				Config: generateRoutingSettingsResource(settingsResource, trueValue),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_routing_settings."+settingsResource, "reset_agent_on_presence_change", trueValue),
				),
			},
			{
				// Update
				Config: generateRoutingSettingsResource(settingsResource, falseValue),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_routing_settings."+settingsResource, "reset_agent_on_presence_change", falseValue),
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
					trueValue,
					generateSettingsContactCenter(trueValue),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_routing_settings."+settingsResource, "contactcenter.0.remove_skills_from_blind_transfer", trueValue),
				),
			},
			{
				// Update contact center
				Config: generateRoutingSettingsWithCustomAttrs(
					settingsResource,
					trueValue,
					generateSettingsContactCenter(falseValue),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_routing_settings."+settingsResource, "contactcenter.0.remove_skills_from_blind_transfer", falseValue),
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
					trueValue,
					generateSettingsTranscription(transcription1, confidence1, trueValue, trueValue),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_routing_settings."+settingsResource, "transcription.0.transcription", transcription1),
					resource.TestCheckResourceAttr("genesyscloud_routing_settings."+settingsResource, "transcription.0.transcription_confidence_threshold", confidence1),
					resource.TestCheckResourceAttr("genesyscloud_routing_settings."+settingsResource, "transcription.0.low_latency_transcription_enabled", trueValue),
					resource.TestCheckResourceAttr("genesyscloud_routing_settings."+settingsResource, "transcription.0.content_search_enabled", trueValue),
				),
			},
			{
				// Update transcription
				Config: generateRoutingSettingsWithCustomAttrs(
					settingsResource,
					trueValue,
					generateSettingsTranscription(transcription2, confidence2, falseValue, falseValue),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_routing_settings."+settingsResource, "transcription.0.transcription", transcription2),
					resource.TestCheckResourceAttr("genesyscloud_routing_settings."+settingsResource, "transcription.0.transcription_confidence_threshold", confidence2),
					resource.TestCheckResourceAttr("genesyscloud_routing_settings."+settingsResource, "transcription.0.low_latency_transcription_enabled", falseValue),
					resource.TestCheckResourceAttr("genesyscloud_routing_settings."+settingsResource, "transcription.0.content_search_enabled", falseValue),
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
