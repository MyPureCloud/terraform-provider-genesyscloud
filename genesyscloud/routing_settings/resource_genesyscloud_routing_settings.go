package routing_settings

import (
	"context"
	"fmt"
	"log"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"terraform-provider-genesyscloud/genesyscloud/util/constants"
	"terraform-provider-genesyscloud/genesyscloud/util/resourcedata"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"

	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v146/platformclientv2"
)

func getAllRoutingSettings(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	// Although this resource typically has only a single instance,
	// we are attempting to fetch the data from the API in order to
	// verify the user's permission to access this resource's API endpoint(s).

	proxy := getRoutingSettingsProxy(clientConfig)
	resources := make(resourceExporter.ResourceIDMetaMap)

	_, resp, err := proxy.getRoutingSettings(ctx)
	if err != nil {
		if util.IsStatus404(resp) {
			// Don't export if config doesn't exist
			return resources, nil
		}
		return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get %s due to error: %s", ResourceType, err), resp)
	}

	_, resp, err = proxy.getRoutingSettingsContactCenter(ctx)
	if err != nil {
		if util.IsStatus404(resp) {
			// Don't export if config doesn't exist
			return resources, nil
		}
		return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get %s contact center due to error: %s", ResourceType, err), resp)
	}

	_, resp, err = proxy.getRoutingSettingsTranscription(ctx)
	if err != nil {
		if util.IsStatus404(resp) {
			// Don't export if config doesn't exist
			return resources, nil
		}
		return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get %s transcription due to error: %s", ResourceType, err), resp)
	}

	resources["0"] = &resourceExporter.ResourceMeta{BlockLabel: "routing_settings"}
	return resources, nil
}

func createRoutingSettings(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Printf("Creating Routing Setting")
	d.SetId("settings")
	return updateRoutingSettings(ctx, d, meta)
}

func readRoutingSettings(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getRoutingSettingsProxy(sdkConfig)
	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceRoutingSettings(), constants.ConsistencyChecks(), ResourceType)

	log.Printf("Reading routing settings")

	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		settings, resp, getErr := proxy.getRoutingSettings(ctx)
		if getErr != nil {
			if util.IsStatus404(resp) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read Routing Setting %s | error: %s", d.Id(), getErr), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read Routing Setting %s | error: %s", d.Id(), getErr), resp))
		}

		resourcedata.SetNillableValue(d, "reset_agent_on_presence_change", settings.ResetAgentScoreOnPresenceChange)

		if diagErr := readRoutingSettingsContactCenter(ctx, d, proxy); diagErr != nil {
			return retry.NonRetryableError(fmt.Errorf("%v", diagErr))
		}

		if diagErr := readRoutingSettingsTranscription(ctx, d, proxy); diagErr != nil {
			return retry.NonRetryableError(fmt.Errorf("%v", diagErr))
		}

		log.Printf("Read Routing Setting")
		return cc.CheckState(d)
	})
}

func updateRoutingSettings(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	resetAgentOnPresenceChange := d.Get("reset_agent_on_presence_change").(bool)
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getRoutingSettingsProxy(sdkConfig)

	log.Printf("Updating Routing Settings")
	update := platformclientv2.Routingsettings{
		ResetAgentScoreOnPresenceChange: &resetAgentOnPresenceChange,
	}

	diagErr := updateContactCenter(ctx, d, proxy)
	if diagErr != nil {
		return diagErr
	}

	diagErr = updateTranscription(ctx, d, proxy)
	if diagErr != nil {
		return diagErr
	}

	_, resp, err := proxy.updateRoutingSettings(ctx, &update)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to update routing settings %s error: %s", d.Id(), err), resp)
	}

	time.Sleep(5 * time.Second)

	log.Printf("Updated Routing Settings")
	return readRoutingSettings(ctx, d, meta)
}

func deleteRoutingSettings(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getRoutingSettingsProxy(sdkConfig)

	log.Printf("Resetting Routing Setting")
	resp, err := proxy.deleteRoutingSettings(ctx)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to delete routing settings %s error: %s", d.Id(), err), resp)
	}

	log.Printf("Reset Routing Settings")
	return nil
}

func readRoutingSettingsContactCenter(ctx context.Context, d *schema.ResourceData, proxy *routingSettingsProxy) diag.Diagnostics {
	contactCenter, resp, getErr := proxy.getRoutingSettingsContactCenter(ctx)
	if getErr != nil {
		if util.IsStatus404(resp) {
			return nil
		}
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to read contact center for routing setting %s error: %s", d.Id(), getErr), resp)
	}

	if contactCenter == nil {
		_ = d.Set("contactcenter", nil)
		return nil
	}

	contactSettings := make(map[string]interface{})
	resourcedata.SetMapValueIfNotNil(contactSettings, "remove_skills_from_blind_transfer", contactCenter.RemoveSkillsFromBlindTransfer)

	_ = d.Set("contactcenter", []interface{}{contactSettings})
	return nil
}

func updateContactCenter(ctx context.Context, d *schema.ResourceData, proxy *routingSettingsProxy) diag.Diagnostics {
	var removeSkillsFromBlindTransfer bool

	if contactCenterConfig := d.Get("contactcenter"); contactCenterConfig != nil {
		if contactCenterList := contactCenterConfig.([]interface{}); len(contactCenterList) > 0 {
			contactCenterMap := contactCenterList[0].(map[string]interface{})

			if contactCenterMap["remove_skills_from_blind_transfer"] != nil {
				removeSkillsFromBlindTransfer = contactCenterMap["remove_skills_from_blind_transfer"].(bool)
			}

			contactCenterSettings := platformclientv2.Contactcentersettings{
				RemoveSkillsFromBlindTransfer: &removeSkillsFromBlindTransfer,
			}

			resp, err := proxy.updateRoutingSettingsContactCenter(ctx, contactCenterSettings)
			if err != nil {
				return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to update contact center for routing settings %s error: %s", d.Id(), err), resp)
			}
		}
	}
	return nil
}

func readRoutingSettingsTranscription(ctx context.Context, d *schema.ResourceData, proxy *routingSettingsProxy) diag.Diagnostics {
	transcription, resp, getErr := proxy.getRoutingSettingsTranscription(ctx)
	if getErr != nil {
		if util.IsStatus404(resp) {
			return nil
		}
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to read contact center for routing settings %s error: %s", d.Id(), getErr), resp)
	}

	if transcription == nil {
		_ = d.Set("transcription", nil)
		return nil
	}

	transcriptionSettings := make(map[string]interface{})
	resourcedata.SetMapValueIfNotNil(transcriptionSettings, "transcription", transcription.Transcription)
	resourcedata.SetMapValueIfNotNil(transcriptionSettings, "transcription_confidence_threshold", transcription.TranscriptionConfidenceThreshold)
	resourcedata.SetMapValueIfNotNil(transcriptionSettings, "low_latency_transcription_enabled", transcription.LowLatencyTranscriptionEnabled)
	resourcedata.SetMapValueIfNotNil(transcriptionSettings, "content_search_enabled", transcription.ContentSearchEnabled)
	resourcedata.SetMapValueIfNotNil(transcriptionSettings, "pci_dss_redaction_enabled", transcription.PciDssRedactionEnabled)
	resourcedata.SetMapValueIfNotNil(transcriptionSettings, "pii_redaction_enabled", transcription.PiiRedactionEnabled)

	_ = d.Set("transcription", []interface{}{transcriptionSettings})
	return nil
}

func updateTranscription(ctx context.Context, d *schema.ResourceData, proxy *routingSettingsProxy) diag.Diagnostics {
	transcriptionRequest := platformclientv2.Transcriptionsettings{}

	if transcriptionConfigList, ok := d.Get("transcription").([]interface{}); ok && len(transcriptionConfigList) > 0 {
		transcriptionMap, ok := transcriptionConfigList[0].(map[string]interface{})
		if !ok {
			return nil
		}
		if transcription, ok := transcriptionMap["transcription"].(string); ok && transcription != "" {
			transcriptionRequest.Transcription = &transcription
		}
		if transcriptionConfidenceThreshold, ok := transcriptionMap["transcription_confidence_threshold"].(int); ok {
			transcriptionRequest.TranscriptionConfidenceThreshold = &transcriptionConfidenceThreshold
		}
		if lowLatencyTranscriptionEnabled, ok := transcriptionMap["low_latency_transcription_enabled"].(bool); ok {
			transcriptionRequest.LowLatencyTranscriptionEnabled = &lowLatencyTranscriptionEnabled
		}
		if contentSearchEnabled, ok := transcriptionMap["content_search_enabled"].(bool); ok {
			transcriptionRequest.ContentSearchEnabled = &contentSearchEnabled
		}
		if pciEnabled, ok := transcriptionMap["pci_dss_redaction_enabled"].(bool); ok {
			transcriptionRequest.PciDssRedactionEnabled = &pciEnabled
		}
		if piiEnabled, ok := transcriptionMap["pii_redaction_enabled"].(bool); ok {
			transcriptionRequest.PiiRedactionEnabled = &piiEnabled
		}

		_, resp, err := proxy.updateRoutingSettingsTranscription(ctx, transcriptionRequest)
		if err != nil {
			return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to update Transcription for routing settings %s error: %s", d.Id(), err), resp)
		}
	}
	return nil
}
