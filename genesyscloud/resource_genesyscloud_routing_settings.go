package genesyscloud

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"

	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v115/platformclientv2"
)

func ResourceRoutingSettings() *schema.Resource {
	return &schema.Resource{
		Description: "An organization's routing settings",

		CreateContext: CreateWithPooledClient(createRoutingSettings),
		ReadContext:   ReadWithPooledClient(readRoutingSettings),
		UpdateContext: UpdateWithPooledClient(updateRoutingSettings),
		DeleteContext: DeleteWithPooledClient(deleteRoutingSettings),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
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

func getAllRoutingSettings(_ context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(resourceExporter.ResourceIDMetaMap)
	resources["0"] = &resourceExporter.ResourceMeta{Name: "routing_settings"}
	return resources, nil
}

func RoutingSettingsExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: GetAllWithPooledClient(getAllRoutingSettings),
		RefAttrs:         map[string]*resourceExporter.RefAttrSettings{}, // No references
	}
}

func createRoutingSettings(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Printf("Creating Routing Setting")
	d.SetId("settings")
	return updateRoutingSettings(ctx, d, meta)
}

func readRoutingSettings(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*ProviderMeta).ClientConfig
	routingAPI := platformclientv2.NewRoutingApiWithConfig(sdkConfig)

	log.Printf("Reading setting: %s", d.Id())
	return WithRetriesForRead(ctx, d, func() *retry.RetryError {
		settings, resp, getErr := routingAPI.GetRoutingSettings()

		if getErr != nil {
			if IsStatus404(resp) {
				//createRoutingSettings(ctx, d, meta)
				return retry.RetryableError(fmt.Errorf("Failed to read Routing Setting: %s", getErr))
			}
			return retry.NonRetryableError(fmt.Errorf("Failed to read Routing Setting: %s", getErr))
		}

		cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceRoutingSettings())
		if settings.ResetAgentScoreOnPresenceChange != nil {
			d.Set("reset_agent_on_presence_change", *settings.ResetAgentScoreOnPresenceChange)
		} else {
			d.Set("reset_agent_on_presence_change", nil)
		}

		if diagErr := readRoutingSettingsContactCenter(d, routingAPI); diagErr != nil {
			return retry.NonRetryableError(fmt.Errorf("%v", diagErr))
		}

		if diagErr := readRoutingSettingsTranscription(d, routingAPI); diagErr != nil {
			return retry.NonRetryableError(fmt.Errorf("%v", diagErr))
		}

		log.Printf("Read Routing Setting")
		return cc.CheckState()
	})
}

func updateRoutingSettings(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	resetAgentOnPresenceChange := d.Get("reset_agent_on_presence_change").(bool)

	sdkConfig := meta.(*ProviderMeta).ClientConfig
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
	sdkConfig := meta.(*ProviderMeta).ClientConfig
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
		if IsStatus404(resp) {
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
		if IsStatus404(resp) {
			return nil
		}
		return diag.Errorf("Failed to read Contact center for routing setting %s: %s\n", d.Id(), getErr)
	}

	if transcription == nil {
		d.Set("transcription", nil)
		return nil
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
	var removeSkillsFromBlindTransfer bool

	if contactCenterConfig := d.Get("contactcenter"); contactCenterConfig != nil {
		if contactCenterList := contactCenterConfig.([]interface{}); len(contactCenterList) > 0 {
			contactCenterMap := contactCenterList[0].(map[string]interface{})

			if contactCenterMap["remove_skills_from_blind_transfer"] != nil {
				removeSkillsFromBlindTransfer = contactCenterMap["remove_skills_from_blind_transfer"].(bool)
			}
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

			var transcription string
			var transcriptionConfidenceThreshold int
			var lowLatencyTranscriptionEnabled bool
			var contentSearchEnabled bool

			if transcriptionMap["transcription"] != nil {
				transcription = transcriptionMap["transcription"].(string)
			}
			if transcriptionMap["transcription_confidence_threshold"] != nil {
				transcriptionConfidenceThreshold = transcriptionMap["transcription_confidence_threshold"].(int)
			}
			if transcriptionMap["low_latency_transcription_enabled"] != nil {
				lowLatencyTranscriptionEnabled = transcriptionMap["low_latency_transcription_enabled"].(bool)
			}
			if transcriptionMap["content_search_enabled"] != nil {
				contentSearchEnabled = transcriptionMap["content_search_enabled"].(bool)
			}

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
