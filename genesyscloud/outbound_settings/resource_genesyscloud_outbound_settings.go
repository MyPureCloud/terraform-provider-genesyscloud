package outbound_settings

import (
	"context"
	"fmt"
	"log"
	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	"terraform-provider-genesyscloud/genesyscloud/tfexporter_state"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"terraform-provider-genesyscloud/genesyscloud/util/constants"
	"terraform-provider-genesyscloud/genesyscloud/util/resourcedata"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v146/platformclientv2"
)

/*
The resource_genesyscloud_outbound_settings.go contains all the methods that perform the core logic for a resource.
*/

func getAllOutboundSettings(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	// Although this resource typically has only a single instance,
	// we are attempting to fetch the data from the API in order to
	// verify the user's permission to access this resource's API endpoint(s).

	proxy := getOutboundSettingsProxy(clientConfig)
	resources := make(resourceExporter.ResourceIDMetaMap)
	_, resp, err := proxy.getOutboundSettings(ctx)
	if err != nil {
		if util.IsStatus404(resp) {
			// Don't export if config doesn't exist
			return resources, nil
		}
		return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get %s due to error: %s", ResourceType, err), resp)
	}
	resources["0"] = &resourceExporter.ResourceMeta{BlockLabel: "outbound_settings"}
	return resources, nil
}

// createOutboundSettings is used by the outbound_settings resource to create Genesys cloud outbound settings
func createOutboundSettings(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Printf("Creating Outbound Setting")
	d.SetId("settings")
	return updateOutboundSettings(ctx, d, meta)
}

// readOutboundSettings is used by the outbound_settings resource to read an outbound settings from genesys cloud
func readOutboundSettings(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getOutboundSettingsProxy(sdkConfig)
	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceOutboundSettings(), constants.ConsistencyChecks(), ResourceType)

	maxCallsPerAgent := d.Get("max_calls_per_agent").(int)
	maxLineUtilization := d.Get("max_line_utilization").(float64)
	abandonSeconds := d.Get("abandon_seconds").(float64)
	complianceAbandonRateDenominator := d.Get("compliance_abandon_rate_denominator").(string)
	automaticTimeZoneMapping := d.Get("automatic_time_zone_mapping").([]interface{})
	rescheduleTimeZoneSkippedContacts := d.Get("reschedule_time_zone_skipped_contacts").(bool)

	log.Printf("Reading Outbound Settings %s", d.Id())

	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		settings, resp, getErr := proxy.getOutboundSettings(ctx)
		if getErr != nil {
			if util.IsStatus404(resp) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read Outbound Setting: %s", getErr), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read Outbound Setting: %s", getErr), resp))

		}

		// Only read values if they are part of the terraform plan or during Export
		if maxCallsPerAgent != 0 || tfexporter_state.IsExporterActive() {
			resourcedata.SetNillableValue(d, "max_calls_per_agent", settings.MaxCallsPerAgent)
		}
		if maxLineUtilization != 0 || tfexporter_state.IsExporterActive() {
			resourcedata.SetNillableValue(d, "max_line_utilization", settings.MaxLineUtilization)
		}
		if abandonSeconds != 0 || tfexporter_state.IsExporterActive() {
			resourcedata.SetNillableValue(d, "abandon_seconds", settings.AbandonSeconds)
		}
		if complianceAbandonRateDenominator != "" || tfexporter_state.IsExporterActive() {
			resourcedata.SetNillableValue(d, "compliance_abandon_rate_denominator", settings.ComplianceAbandonRateDenominator)
		}
		if settings.AutomaticTimeZoneMapping != nil && (len(automaticTimeZoneMapping) > 0 || tfexporter_state.IsExporterActive()) {
			_ = d.Set("automatic_time_zone_mapping", flattenOutboundSettingsAutomaticTimeZoneMapping(*settings.AutomaticTimeZoneMapping, automaticTimeZoneMapping))
		}
		resourcedata.SetNillableValue(d, "reschedule_time_zone_skipped_contacts", &rescheduleTimeZoneSkippedContacts)

		log.Printf("Read Outbound Setting")
		return cc.CheckState(d)
	})
}

// updateOutboundSettings is used by the outbound_settings resource to update an outbound settings in Genesys Cloud
func updateOutboundSettings(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getOutboundSettingsProxy(sdkConfig)

	maxCallsPerAgent := d.Get("max_calls_per_agent").(int)
	maxLineUtilization := d.Get("max_line_utilization").(float64)
	abandonSeconds := d.Get("abandon_seconds").(float64)
	complianceAbandonRateDenominator := d.Get("compliance_abandon_rate_denominator").(string)
	automaticTimeZoneMapping := d.Get("automatic_time_zone_mapping").([]interface{})

	log.Printf("Updating Outbound Settings %s", d.Id())

	diagErr := util.RetryWhen(util.IsVersionMismatch, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		// Get current Outbound settings version
		setting, resp, getErr := proxy.getOutboundSettings(ctx)
		if getErr != nil {
			return resp, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to update Outbound Setting %s error: %s", d.Id(), getErr), resp)
		}

		update := platformclientv2.Outboundsettings{
			Name:                              setting.Name,
			Version:                           setting.Version,
			RescheduleTimeZoneSkippedContacts: platformclientv2.Bool(d.Get("reschedule_time_zone_skipped_contacts").(bool)),
		}
		if maxCallsPerAgent != 0 || tfexporter_state.IsExporterActive() {
			update.MaxCallsPerAgent = &maxCallsPerAgent
		}
		if maxLineUtilization != 0 || tfexporter_state.IsExporterActive() {
			update.MaxLineUtilization = &maxLineUtilization
		}
		if abandonSeconds != 0 || tfexporter_state.IsExporterActive() {
			update.AbandonSeconds = &abandonSeconds
		}
		if complianceAbandonRateDenominator != "" || tfexporter_state.IsExporterActive() {
			update.ComplianceAbandonRateDenominator = &complianceAbandonRateDenominator
		}
		if automaticTimeZoneMapping != nil || tfexporter_state.IsExporterActive() {
			update.AutomaticTimeZoneMapping = buildOutboundSettingsAutomaticTimeZoneMapping(d)
		}

		_, resp, err := proxy.updateOutboundSettings(ctx, &update)
		if err != nil {
			return resp, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to update Outbound settings %s error: %s", *setting.Name, err), resp)
		}
		return nil, nil
	})
	if diagErr != nil {
		return diagErr
	}

	log.Printf("Updated Outbound settings %s", d.Id())
	return readOutboundSettings(ctx, d, meta)
}

// deleteOutboundSettings is used by the outbound_settings resource to delete an outbound settings from Genesys cloud
func deleteOutboundSettings(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return nil
}
