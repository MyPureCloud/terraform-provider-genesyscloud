package genesyscloud

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v75/platformclientv2"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/consistency_checker"
	"log"
)

func resourceRoutingSettings() *schema.Resource {
	return &schema.Resource{
		Description: "This is a random description",

		CreateContext: createWithPooledClient(createRoutingSettings),
		ReadContext:   readWithPooledClient(readRoutingSettings),
		UpdateContext: updateWithPooledClient(updateRoutingSettings),
		DeleteContext: deleteWithPooledClient(deleteRoutingSettings),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"reset_agent_on_presence_change": {
				Description: "True if.",
				Type:        schema.TypeBool,
				Optional:    true,
			},
			"contactcenter": {
				Description: "Description",
				Type:        schema.TypeList,
				Optional:    true,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"remove_skills_from_blind_transfer": {
							Description: "Description",
							Type:        schema.TypeBool,
							Optional:    true,
						},
					},
				},
			},
			"transcription": {
				Description: "Description",
				Type:        schema.TypeList,
				Optional:    true,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"transcription": {
							Description: "Description",
							Type:        schema.TypeString,
							Optional:    true,
						},
						"transcription_confidence_threshold": {
							Description: "Description",
							Type:        schema.TypeInt,
							Optional:    true,
						},
						"low_latency_transcription_enabled": {
							Description: "Description",
							Type:        schema.TypeBool,
							Optional:    true,
						},
						"content_search_enabled": {
							Description: "Description",
							Type:        schema.TypeBool,
							Optional:    true,
						},
					},
				},
			},
		},
	}
}

func createRoutingSettings(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Printf("Creating Routing Setting")
	d.SetId("settings")
	return updateRoutingSettings(ctx, d, meta)
}

func readRoutingSettings(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*providerMeta).ClientConfig
	routingAPI := platformclientv2.NewRoutingApiWithConfig(sdkConfig)

	log.Printf("Reading setting: %s", d.Id())
	return withRetriesForRead(ctx, d, func() *resource.RetryError {
		settings, resp, getErr := routingAPI.GetRoutingSettings()

		if getErr != nil {
			if isStatus404(resp) {
				//createRoutingSettings(ctx, d, meta)
				return resource.RetryableError(fmt.Errorf("Failed to read Routing Setting: %s", getErr))
			}
			return resource.NonRetryableError(fmt.Errorf("Failed to read Routing Setting: %s", getErr))
		}

		cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, resourceRoutingSettings())
		if settings.ResetAgentScoreOnPresenceChange != nil {
			d.Set("reset_agent_on_presence_change", *settings.ResetAgentScoreOnPresenceChange)
		} else {
			d.Set("reset_agent_on_presence_change", nil)
		}

		if diagErr := readRoutingSettingsContactCenter(d, routingAPI); diagErr != nil {
			return resource.NonRetryableError(fmt.Errorf("%v", diagErr))
		}

		if diagErr := readRoutingSettingsTranscription(d, routingAPI); diagErr != nil {
			return resource.NonRetryableError(fmt.Errorf("%v", diagErr))
		}

		log.Printf("Read Routing Setting")
		return cc.CheckState()
	})
}

func updateRoutingSettings(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	resetAgentOnPresenceChange := d.Get("reset_agent_on_presence_change").(bool)

	sdkConfig := meta.(*providerMeta).ClientConfig
	routingAPI := platformclientv2.NewRoutingApiWithConfig(sdkConfig)

	log.Printf("Updating Routing Settings")
	update := platformclientv2.Routingsettings{
		ResetAgentScoreOnPresenceChange: &resetAgentOnPresenceChange,
	}

	diagErr := updateContactCenter(d, routingAPI)
	if diagErr != nil {
		return diagErr
	}

	diagErr = updateTranscription(d, routingAPI)
	if diagErr != nil {
		return diagErr
	}

	_, _, err := routingAPI.PutRoutingSettings(update)
	if err != nil {
		return diag.Errorf("Failed to update routing settings: %s", err)
	}

	log.Printf("Updated Routing Settings")
	return readRoutingSettings(ctx, d, meta)
}

func deleteRoutingSettings(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*providerMeta).ClientConfig
	routingAPI := platformclientv2.NewRoutingApiWithConfig(sdkConfig)

	log.Printf("Resetting Routing Setting")
	_, err := routingAPI.DeleteRoutingSettings()
	if err != nil {
		return diag.Errorf("Failed to reset Routing Setting %s", err)
	}
	log.Printf("Reset Routing Settings")
	return nil
}

func readRoutingSettingsContactCenter(d *schema.ResourceData, routingAPI *platformclientv2.RoutingApi) diag.Diagnostics {
	contactcenter, resp, getErr := routingAPI.GetRoutingSettingsContactcenter()
	if getErr != nil {
		if isStatus404(resp) {
			d.SetId("") // Contact center doesn't exist
			return nil
		}
		return diag.Errorf("Failed to read Contact center for routing setting %s: %s\n", d.Id(), getErr)
	}

	if contactcenter == nil {
		d.Set("contactcenter", nil)
		return nil
	}

	contactSettings := make(map[string]interface{})

	if contactcenter.RemoveSkillsFromBlindTransfer != nil {
		contactSettings["remove_skills_from_blind_transfer"] = *contactcenter.RemoveSkillsFromBlindTransfer
	}

	d.Set("contactcenter", []interface{}{contactSettings})
	return nil
}

func readRoutingSettingsTranscription(d *schema.ResourceData, routingAPI *platformclientv2.RoutingApi) diag.Diagnostics {
	transcription, resp, getErr := routingAPI.GetRoutingSettingsTranscription()
	if getErr != nil {
		if isStatus404(resp) {
			d.SetId("") // Transcription doesn't exist
			return nil
		}
		return diag.Errorf("Failed to read Contact center for routing setting %s: %s\n", d.Id(), getErr)
	}

	if transcription == nil {
		d.Set("transcription", nil)
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

func updateContactCenter(d *schema.ResourceData, routingAPI *platformclientv2.RoutingApi) diag.Diagnostics {
	if contactCenterConfig := d.Get("contactcenter"); contactCenterConfig != nil {
		if contactCenterList := contactCenterConfig.([]interface{}); len(contactCenterList) > 0 {
			contactCenterMap := contactCenterList[0].(map[string]interface{})

			removeSkillsFromBlindTransfer := contactCenterMap["remove_skills_from_blind_transfer"].(bool)
			_, err := routingAPI.PatchRoutingSettingsContactcenter(platformclientv2.Contactcentersettings{
				RemoveSkillsFromBlindTransfer: &removeSkillsFromBlindTransfer,
			})

			if err != nil {
				return diag.Errorf("Failed to update Contact center for routing setting %s: %s\n", d.Id(), err)
			}
		}
	}
	return nil
}

func updateTranscription(d *schema.ResourceData, routingAPI *platformclientv2.RoutingApi) diag.Diagnostics {
	if transcriptionConfig := d.Get("transcription"); transcriptionConfig != nil {
		if transcriptionList := transcriptionConfig.([]interface{}); len(transcriptionList) > 0 {
			transcriptionMap := transcriptionList[0].(map[string]interface{})

			transcription := transcriptionMap["transcription"].(string)
			transcriptionConfidenceThreshold := transcriptionMap["transcription_confidence_threshold"].(int)
			lowLatencyTranscriptionEnabled := transcriptionMap["low_latency_transcription_enabled"].(bool)
			contentSearchEnabled := transcriptionMap["content_search_enabled"].(bool)
			_, _, err := routingAPI.PutRoutingSettingsTranscription(platformclientv2.Transcriptionsettings{
				Transcription:                    &transcription,
				TranscriptionConfidenceThreshold: &transcriptionConfidenceThreshold,
				LowLatencyTranscriptionEnabled:   &lowLatencyTranscriptionEnabled,
				ContentSearchEnabled:             &contentSearchEnabled,
			})

			if err != nil {
				return diag.Errorf("Failed to update Transcription for routing setting %s: %s\n", d.Id(), err)
			}
		}
	}
	return nil
}
