package outbound_settings

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v125/platformclientv2"
	"log"
	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	"terraform-provider-genesyscloud/genesyscloud/util"
)

/*
The resource_genesyscloud_outbound_settings.go contains all of the methods that perform the core logic for a resource.
*/

func getAllOutboundSettings(_ context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(resourceExporter.ResourceIDMetaMap)
	resources["0"] = &resourceExporter.ResourceMeta{Name: "outbound_settings"}
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

	maxCallsPerAgent := d.Get("max_calls_per_agent").(int)
	maxLineUtilization := d.Get("max_line_utilization").(float64)
	abandonSeconds := d.Get("abandon_seconds").(float64)
	complianceAbandonRateDenominator := d.Get("compliance_abandon_rate_denominator").(string)
	automaticTimeZoneMapping := d.Get("automatic_time_zone_mapping").([]interface{})

	log.Printf("Reading Outbound setting %s", d.Id())

	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		settings, resp, getErr := proxy.getOutboundSettingsById(ctx, d.Id())
		if getErr != nil {
			if util.IsStatus404(resp) {
				return retry.RetryableError(fmt.Errorf("Failed to read Outbound Setting: %s", getErr))
			}
			return retry.NonRetryableError(fmt.Errorf("Failed to read Outbound Setting: %s", getErr))
		}

		cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceOutboundSettings())

		// Only read values if they are part of the terraform plan
		if maxCallsPerAgent != 0 {
			if settings.MaxCallsPerAgent != nil {
				d.Set("max_calls_per_agent", *settings.MaxCallsPerAgent)
			} else {
				d.Set("max_calls_per_agent", nil)
			}
		}

		if maxLineUtilization != 0 {
			if settings.MaxLineUtilization != nil {
				d.Set("max_line_utilization", *settings.MaxLineUtilization)
			} else {
				d.Set("max_line_utilization", nil)
			}
		}

		if abandonSeconds != 0 {
			if settings.AbandonSeconds != nil {
				d.Set("abandon_seconds", *settings.AbandonSeconds)
			} else {
				d.Set("abandon_seconds", nil)
			}
		}

		if complianceAbandonRateDenominator != "" {
			if settings.ComplianceAbandonRateDenominator != nil {
				d.Set("compliance_abandon_rate_denominator", *settings.ComplianceAbandonRateDenominator)
			} else {
				d.Set("compliance_abandon_rate_denominator", nil)
			}
		}

		if len(automaticTimeZoneMapping) > 0 {
			d.Set("automatic_time_zone_mapping", flattenOutboundSettingsAutomaticTimeZoneMapping(*settings.AutomaticTimeZoneMapping, automaticTimeZoneMapping))
		}
		log.Printf("Read Outbound Setting")

		return cc.CheckState()
	})
}

// updateOutboundSettings is used by the outbound_settings resource to update an outbound settings in Genesys Cloud
func updateOutboundSettings(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	maxCallsPerAgent := d.Get("max_calls_per_agent").(int)
	maxLineUtilization := d.Get("max_line_utilization").(float64)
	abandonSeconds := d.Get("abandon_seconds").(float64)
	complianceAbandonRateDenominator := d.Get("compliance_abandon_rate_denominator").(string)

	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getOutboundSettingsProxy(sdkConfig)

	log.Printf("Updating Outbound Settings %s", d.Id())

	diagErr := util.RetryWhen(util.IsVersionMismatch, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		// Get current Outbound settings version
		setting, resp, getErr := proxy.getOutboundSettingsById(ctx, d.Id())
		if getErr != nil {
			return resp, util.BuildAPIDiagnosticError(resourceName, fmt.Sprintf("Failed to update Outbound Setting %s", d.Id()), resp)
		}

		update := platformclientv2.Outboundsettings{
			Name:                     setting.Name,
			Version:                  setting.Version,
			AutomaticTimeZoneMapping: buildOutboundSettingsAutomaticTimeZoneMapping(d),
		}

		if maxCallsPerAgent != 0 {
			update.MaxCallsPerAgent = &maxCallsPerAgent
		}
		if maxLineUtilization != 0 {
			update.MaxLineUtilization = &maxLineUtilization
		}
		if abandonSeconds != 0 {
			update.AbandonSeconds = &abandonSeconds
		}
		if complianceAbandonRateDenominator != "" {
			update.ComplianceAbandonRateDenominator = &complianceAbandonRateDenominator
		}

		_, resp, err := proxy.updateOutboundSettings(ctx, d.Id(), &update)
		if err != nil {
			return resp, util.BuildAPIDiagnosticError(resourceName, fmt.Sprintf("Failed to update Outbound settings %s", *setting.Name), resp)
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
