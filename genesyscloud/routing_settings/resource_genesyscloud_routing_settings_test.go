package routing_settings

import (
	"fmt"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceRoutingSettingsBasic(t *testing.T) {
	var (
		settingsResourceLabel = "settings-basic"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, nil),
		Steps: []resource.TestStep{
			{
				// Create
				Config: generateRoutingSettingsResource(settingsResourceLabel, util.FalseValue),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_routing_settings."+settingsResourceLabel, "reset_agent_on_presence_change", util.FalseValue),
				),
			},
			{
				// Update
				Config: generateRoutingSettingsResource(settingsResourceLabel, util.TrueValue),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_routing_settings."+settingsResourceLabel, "reset_agent_on_presence_change", util.TrueValue),
				),
			},
			{
				// Update
				Config: generateRoutingSettingsResource(settingsResourceLabel, util.FalseValue),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_routing_settings."+settingsResourceLabel, "reset_agent_on_presence_change", util.FalseValue),
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_routing_settings." + settingsResourceLabel,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccResourceRoutingSettingsContactCenter(t *testing.T) {
	var (
		settingsResourceLabel = "settings-contactCenter"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, nil),
		Steps: []resource.TestStep{
			{
				// Create with contact center
				Config: generateRoutingSettingsWithCustomAttrs(
					settingsResourceLabel,
					util.TrueValue,
					generateSettingsContactCenter(util.TrueValue),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_routing_settings."+settingsResourceLabel, "contactcenter.0.remove_skills_from_blind_transfer", util.TrueValue),
				),
			},
			{
				// Update contact center
				Config: generateRoutingSettingsWithCustomAttrs(
					settingsResourceLabel,
					util.TrueValue,
					generateSettingsContactCenter(util.FalseValue),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_routing_settings."+settingsResourceLabel, "contactcenter.0.remove_skills_from_blind_transfer", util.FalseValue),
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_routing_settings." + settingsResourceLabel,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccResourceRoutingSettingsTranscription(t *testing.T) {
	var (
		settingsResourceLabel = "settings-transcription"
		transcription1        = "Disabled"
		transcription2        = "EnabledQueueFlow"
		confidence1           = "1"
		confidence2           = "2"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, nil),
		Steps: []resource.TestStep{
			{
				// Create with transcription
				Config: generateRoutingSettingsWithCustomAttrs(
					settingsResourceLabel,
					util.TrueValue,
					generateSettingsTranscription(transcription1, confidence1, util.TrueValue, util.TrueValue),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_routing_settings."+settingsResourceLabel, "transcription.0.transcription", transcription1),
					resource.TestCheckResourceAttr("genesyscloud_routing_settings."+settingsResourceLabel, "transcription.0.transcription_confidence_threshold", confidence1),
					resource.TestCheckResourceAttr("genesyscloud_routing_settings."+settingsResourceLabel, "transcription.0.low_latency_transcription_enabled", util.TrueValue),
					resource.TestCheckResourceAttr("genesyscloud_routing_settings."+settingsResourceLabel, "transcription.0.content_search_enabled", util.TrueValue),
				),
			},
			{
				// Update transcription
				Config: generateRoutingSettingsWithCustomAttrs(
					settingsResourceLabel,
					util.TrueValue,
					generateSettingsTranscription(transcription2, confidence2, util.FalseValue, util.FalseValue),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_routing_settings."+settingsResourceLabel, "transcription.0.transcription", transcription2),
					resource.TestCheckResourceAttr("genesyscloud_routing_settings."+settingsResourceLabel, "transcription.0.transcription_confidence_threshold", confidence2),
					resource.TestCheckResourceAttr("genesyscloud_routing_settings."+settingsResourceLabel, "transcription.0.low_latency_transcription_enabled", util.FalseValue),
					resource.TestCheckResourceAttr("genesyscloud_routing_settings."+settingsResourceLabel, "transcription.0.content_search_enabled", util.FalseValue),
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_routing_settings." + settingsResourceLabel,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func generateRoutingSettingsResource(
	resourceLabel string,
	resetAgentScoreOnPresenceChange string) string {
	return fmt.Sprintf(`resource "genesyscloud_routing_settings" "%s"{
		reset_agent_on_presence_change = %s
	}
	`, resourceLabel, resetAgentScoreOnPresenceChange)
}

func generateRoutingSettingsWithCustomAttrs(
	resourceLabel string,
	resetAgentScoreOnPresenceChange string,
	attrs ...string) string {
	return fmt.Sprintf(`resource "genesyscloud_routing_settings" "%s" {
		reset_agent_on_presence_change = %s
		%s
	}
	`, resourceLabel, resetAgentScoreOnPresenceChange, strings.Join(attrs, "\n"))
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
