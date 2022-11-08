package genesyscloud

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v80/platformclientv2"
)

func dataSourceOutboundSettings() *schema.Resource {
	return &schema.Resource{
		Description:   "An organization's outbound settings",
		ReadContext:   readWithPooledClient(dataSourceOutboundSettingsRead),
		SchemaVersion: 1,
		Schema:        resourceOutboundSettings().Schema,
	}
}

func dataSourceOutboundSettingsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sdkConfig := m.(*providerMeta).ClientConfig
	routingAPI := platformclientv2.NewOutboundApiWithConfig(sdkConfig)

	settings, _, getErr := routingAPI.GetOutboundSettings()
	if getErr != nil {
		return diag.Errorf("Error requesting outbound settings: %s", getErr)
	}

	d.SetId("datasource-settings")
	if settings.MaxCallsPerAgent != nil {
		d.Set("max_calls_per_agent", *settings.MaxCallsPerAgent)
	}

	if settings.MaxLineUtilization != nil {
		d.Set("max_line_utilization", *settings.MaxLineUtilization)
	}

	if settings.AbandonSeconds != nil {
		d.Set("abandon_seconds", *settings.AbandonSeconds)
	}

	if settings.ComplianceAbandonRateDenominator != nil {
		d.Set("compliance_abandon_rate_denominator", *settings.ComplianceAbandonRateDenominator)
	}

	if settings.AutomaticTimeZoneMapping != nil {
		automaticTimeZoneMapping := d.Get("automatic_time_zone_mapping").([]interface{})
		d.Set("automatic_time_zone_mapping", flattenOutboundSettingsAutomaticTimeZoneMapping(*settings.AutomaticTimeZoneMapping, automaticTimeZoneMapping))
	}
	
	return nil
}
