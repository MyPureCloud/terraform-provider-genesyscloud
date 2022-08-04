package genesyscloud

import (
	"context"
	"github.com/mypurecloud/platform-client-sdk-go/v75/platformclientv2"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceRoutingSettings() *schema.Resource {
	return &schema.Resource{
		Description:   "An organization's routing settings",
		ReadContext:   readWithPooledClient(dataSourceRoutingSettingsRead),
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"reset_agent_on_presence_change": {
				Description: "Reset agent score when agent presence changes from off-queue to on-queue",
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
			},
			"contactcenter": {
				Description: "Contact center settings",
				Type:        schema.TypeList,
				Optional:    true,
				Computed:    true,
				MaxItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"remove_skills_from_blind_transfer": {
							Description: "Strip skills from transfer",
							Type:        schema.TypeBool,
							Optional:    true,
						},
					},
				},
			},
			"transcription": {
				Description: "Transcription settings",
				Type:        schema.TypeList,
				Optional:    true,
				Computed:    true,
				MaxItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"transcription": {
							Description: "Setting to enable/disable transcription capability.Valid values: Disabled, EnabledGlobally, EnabledQueueFlow",
							Type:        schema.TypeString,
							Optional:    true,
						},
						"transcription_confidence_threshold": {
							Description: "Configure confidence threshold. The possible values are from 1 to 100",
							Type:        schema.TypeInt,
							Optional:    true,
						},
						"low_latency_transcription_enabled": {
							Description: "Boolean flag indicating whether low latency transcription via Notification API is enabled",
							Type:        schema.TypeBool,
							Optional:    true,
						},
						"content_search_enabled": {
							Description: "Setting to enable/disable content search",
							Type:        schema.TypeBool,
							Optional:    true,
						},
					},
				},
			},
		},
	}
}

func dataSourceRoutingSettingsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sdkConfig := m.(*providerMeta).ClientConfig
	routingAPI := platformclientv2.NewRoutingApiWithConfig(sdkConfig)

	settings, _, getErr := routingAPI.GetRoutingSettings()
	if getErr != nil {
		return diag.Errorf("Error requesting routing settings: %s", getErr)
	}
	if settings.ResetAgentScoreOnPresenceChange != nil {
		d.Set("reset_agent_score_on_presence_change", *settings.ResetAgentScoreOnPresenceChange)
	}

	contactcenter, _, getErr := routingAPI.GetRoutingSettingsContactcenter()
	if getErr != nil {
		return diag.Errorf("Error requesting routing settings contact center: %s", getErr)
	}
	contactSettings := make(map[string]interface{})
	if contactcenter.RemoveSkillsFromBlindTransfer != nil {
		contactSettings["remove_skills_from_blind_transfer"] = *contactcenter.RemoveSkillsFromBlindTransfer
	}
	d.Set("contactcenter", []interface{}{contactSettings})

	transcription, _, getErr := routingAPI.GetRoutingSettingsTranscription()
	if getErr != nil {
		return diag.Errorf("Failed to read Contact center for routing setting %s: %s\n", d.Id(), getErr)
	}
	transcriptionSettings := make(map[string]interface{})
	if transcription.Transcription != nil {
		transcriptionSettings["transcription"] = *transcription.Transcription
	}
	if transcription.TranscriptionConfidenceThreshold != nil {
		transcriptionSettings["transcription_confidence_threshold"] = *transcription.TranscriptionConfidenceThreshold
	}
	if transcription.LowLatencyTranscriptionEnabled != nil {
		transcriptionSettings["low_latency_transcription_enabled"] = *transcription.LowLatencyTranscriptionEnabled
	}
	if transcription.ContentSearchEnabled != nil {
		transcriptionSettings["content_search_enabled"] = *transcription.ContentSearchEnabled
	}
	d.Set("transcription", []interface{}{transcriptionSettings})

	return nil
}
