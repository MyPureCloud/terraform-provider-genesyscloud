package outbound

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"

	gcloud "terraform-provider-genesyscloud/genesyscloud"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	lists "terraform-provider-genesyscloud/genesyscloud/util/lists"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/mypurecloud/platform-client-sdk-go/v115/platformclientv2"
)

var (
	outboundSettingsAutomaticTimeZoneMappingResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			`callable_windows`: {
				Description: "The time intervals to use for automatic time zone mapping.",
				Optional:    true,
				Type:        schema.TypeSet,
				MaxItems:    1,
				Elem:        outboundSettingsCallableWindowsResource,
			},
			`supported_countries`: {
				Description: "The countries that are supported for automatic time zone mapping.",
				Optional:    true,
				Type:        schema.TypeList,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
		},
	}

	outboundSettingsCallableWindowsResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			`mapped`: {
				Description: "The time interval to place outbound calls, for contacts that can be mapped to a time zone.",
				Optional:    true,
				Type:        schema.TypeSet,
				MaxItems:    1,
				Elem:        outboundSettingsMappedResource,
			},
			`unmapped`: {
				Description: "The time interval and time zone to place outbound calls, for contacts that cannot be mapped to a time zone.",
				Optional:    true,
				Type:        schema.TypeSet,
				MaxItems:    1,
				Elem:        outboundSettingsUnmappedResource,
			},
		},
	}
	outboundSettingsMappedResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			`earliest_callable_time`: {
				Description:      "The earliest time to dial a contact. Valid format is HH:mm",
				Optional:         true,
				ValidateDiagFunc: gcloud.ValidateTimeHHMM,
				Type:             schema.TypeString,
			},
			`latest_callable_time`: {
				Description:      "The latest time to dial a contact. Valid format is HH:mm.",
				Optional:         true,
				ValidateDiagFunc: gcloud.ValidateTimeHHMM,
				Type:             schema.TypeString,
			},
		},
	}
	outboundSettingsUnmappedResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			`earliest_callable_time`: {
				Description:      "The earliest time to dial a contact. Valid format is HH:mm.",
				Optional:         true,
				ValidateDiagFunc: gcloud.ValidateTimeHHMM,
				Type:             schema.TypeString,
			},
			`latest_callable_time`: {
				Description:      "The latest time to dial a contact. Valid format is HH:mm.",
				Optional:         true,
				ValidateDiagFunc: gcloud.ValidateTimeHHMM,
				Type:             schema.TypeString,
			},
			`time_zone_id`: {
				Description: "The time zone to use for contacts that cannot be mapped.",
				Optional:    true,
				Type:        schema.TypeString,
			},
		},
	}
)

func ResourceOutboundSettings() *schema.Resource {
	return &schema.Resource{
		Description: "An organization's outbound settings",

		CreateContext: gcloud.CreateWithPooledClient(createOutboundSettings),
		ReadContext:   gcloud.ReadWithPooledClient(readOutboundSettings),
		UpdateContext: gcloud.UpdateWithPooledClient(updateOutboundSettings),
		DeleteContext: gcloud.DeleteWithPooledClient(deleteOutboundSettings),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			`max_calls_per_agent`: {
				Description: "The maximum number of calls that can be placed per agent on any campaign.",
				Optional:    true,
				Type:        schema.TypeInt,
			},
			`max_line_utilization`: {
				Description:  "The maximum percentage of lines that should be used for Outbound, expressed as a decimal in the range [0.0, 1.0].",
				Optional:     true,
				ValidateFunc: validation.FloatBetween(0.0, 1.0),
				Type:         schema.TypeFloat,
			},
			`abandon_seconds`: {
				Description: "The number of seconds used to determine if a call is abandoned.",
				Optional:    true,
				Type:        schema.TypeFloat,
			},
			`compliance_abandon_rate_denominator`: {
				Description:  "The denominator to be used in determining the compliance abandon rate.Valid values: ALL_CALLS, CALLS_THAT_REACHED_QUEUE.",
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"ALL_CALLS", "CALLS_THAT_REACHED_QUEUE", ""}, false),
				Type:         schema.TypeString,
			},
			`automatic_time_zone_mapping`: {
				Description: "The settings for automatic time zone mapping. Note that changing these settings will change them for both voice and messaging campaigns.",
				Optional:    true,
				Type:        schema.TypeList,
				Elem:        outboundSettingsAutomaticTimeZoneMappingResource,
			},
		},
	}
}

func getAllOutboundSettings(_ context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(resourceExporter.ResourceIDMetaMap)
	resources["0"] = &resourceExporter.ResourceMeta{Name: "outbound_settings"}
	return resources, nil
}

func OutboundSettingsExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: gcloud.GetAllWithPooledClient(getAllOutboundSettings),
		RefAttrs:         map[string]*resourceExporter.RefAttrSettings{}, // No references
	}
}

func createOutboundSettings(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Printf("Creating Outbound Setting")
	d.SetId("settings")
	return updateOutboundSettings(ctx, d, meta)
}

func updateOutboundSettings(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	maxCallsPerAgent := d.Get("max_calls_per_agent").(int)
	maxLineUtilization := d.Get("max_line_utilization").(float64)
	abandonSeconds := d.Get("abandon_seconds").(float64)
	complianceAbandonRateDenominator := d.Get("compliance_abandon_rate_denominator").(string)

	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	outboundApi := platformclientv2.NewOutboundApiWithConfig(sdkConfig)

	log.Printf("Updating Outbound Settings %s", d.Id())

	diagErr := gcloud.RetryWhen(gcloud.IsVersionMismatch, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		// Get current Outbound settings version
		setting, resp, getErr := outboundApi.GetOutboundSettings()
		if getErr != nil {
			return resp, diag.Errorf("Failed to read Outbound Setting %s: %s", d.Id(), getErr)
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

		_, err := outboundApi.PatchOutboundSettings(update)
		if err != nil {
			return resp, diag.Errorf("Failed to update Outbound settings %s: %s", *setting.Name, err)
		}
		return nil, nil
	})

	if diagErr != nil {
		return diagErr
	}

	log.Printf("Updated Outbound settings %s", d.Id())
	return readOutboundSettings(ctx, d, meta)
}

func buildOutboundSettingsAutomaticTimeZoneMapping(d *schema.ResourceData) *platformclientv2.Automatictimezonemappingsettings {
	if mappingRequest := d.Get("automatic_time_zone_mapping"); mappingRequest != nil {
		if mappingList := mappingRequest.([]interface{}); len(mappingList) > 0 {
			mappingMap := mappingList[0].(map[string]interface{})

			return &platformclientv2.Automatictimezonemappingsettings{
				CallableWindows:    buildCallableWindows(mappingMap["callable_windows"].(*schema.Set)),
				SupportedCountries: buildSupportedCountries(d),
			}
		}
	}
	return &platformclientv2.Automatictimezonemappingsettings{}
}

func buildSupportedCountries(d *schema.ResourceData) *[]string {
	supportedCountries := []string{}
	if countries, ok := d.GetOk("automatic_time_zone_mapping.0.supported_countries"); ok {
		supportedCountries = lists.InterfaceListToStrings(countries.([]interface{}))
	}
	return &supportedCountries
}

func buildCallableWindows(windows *schema.Set) *[]platformclientv2.Callablewindow {
	if windows == nil {
		return nil
	}
	windowsSlice := make([]platformclientv2.Callablewindow, 0)
	windowsList := windows.List()
	for _, callableWindow := range windowsList {
		var sdkCallableWindow platformclientv2.Callablewindow

		callableWindowsMap := callableWindow.(map[string]interface{})

		sdkCallableWindow.Mapped = buildCallableWindowsMapped(callableWindowsMap["mapped"].(*schema.Set))
		sdkCallableWindow.Unmapped = buildCallableWindowsUnmapped(callableWindowsMap["unmapped"].(*schema.Set))

		windowsSlice = append(windowsSlice, sdkCallableWindow)
	}
	return &windowsSlice
}

func buildCallableWindowsMapped(mappedWindows *schema.Set) *platformclientv2.Atzmtimeslot {
	if mappedWindows != nil {
		if mappedWindowsList := mappedWindows.List(); len(mappedWindowsList) > 0 {
			mappedWindowsMap := mappedWindowsList[0].(map[string]interface{})

			earliestCallableTime := mappedWindowsMap["earliest_callable_time"].(string)
			latestCallableTime := mappedWindowsMap["latest_callable_time"].(string)

			update := &platformclientv2.Atzmtimeslot{}

			if earliestCallableTime != "" {
				update.EarliestCallableTime = &earliestCallableTime
			}
			if latestCallableTime != "" {
				update.LatestCallableTime = &latestCallableTime
			}

			return update
		}
	}
	return &platformclientv2.Atzmtimeslot{}
}

func buildCallableWindowsUnmapped(unmappedWindows *schema.Set) *platformclientv2.Atzmtimeslotwithtimezone {
	if unmappedWindows != nil {
		if unmappedWindowsList := unmappedWindows.List(); len(unmappedWindowsList) > 0 {
			unmappedWindowsMap := unmappedWindowsList[0].(map[string]interface{})

			earliestCallableTime := unmappedWindowsMap["earliest_callable_time"].(string)
			latestCallableTime := unmappedWindowsMap["latest_callable_time"].(string)
			timeZoneId := unmappedWindowsMap["time_zone_id"].(string)

			update := &platformclientv2.Atzmtimeslotwithtimezone{}

			if earliestCallableTime != "" {
				update.EarliestCallableTime = &earliestCallableTime
			}
			if latestCallableTime != "" {
				update.LatestCallableTime = &latestCallableTime
			}
			if timeZoneId != "" {
				update.TimeZoneId = &timeZoneId
			}

			return update
		}
	}
	return &platformclientv2.Atzmtimeslotwithtimezone{}
}

func readOutboundSettings(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	outboundApi := platformclientv2.NewOutboundApiWithConfig(sdkConfig)

	maxCallsPerAgent := d.Get("max_calls_per_agent").(int)
	maxLineUtilization := d.Get("max_line_utilization").(float64)
	abandonSeconds := d.Get("abandon_seconds").(float64)
	complianceAbandonRateDenominator := d.Get("compliance_abandon_rate_denominator").(string)
	automaticTimeZoneMapping := d.Get("automatic_time_zone_mapping").([]interface{})

	log.Printf("Reading Outbound setting %s", d.Id())

	return gcloud.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		settings, resp, getErr := outboundApi.GetOutboundSettings()

		if getErr != nil {
			if gcloud.IsStatus404(resp) {
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

func flattenOutboundSettingsAutomaticTimeZoneMapping(timeZoneMappings platformclientv2.Automatictimezonemappingsettings, automaticTimeZoneMapping []interface{}) []interface{} {
	callableWindows := automaticTimeZoneMapping[0].(map[string]interface{})["callable_windows"].(*schema.Set)
	requestMap := make(map[string]interface{})
	if timeZoneMappings.CallableWindows != nil {
		requestMap["callable_windows"] = flattenCallableWindows(*timeZoneMappings.CallableWindows, callableWindows)
	}
	if timeZoneMappings.SupportedCountries != nil {
		requestMap["supported_countries"] = *timeZoneMappings.SupportedCountries
	}
	return []interface{}{requestMap}
}

func flattenCallableWindows(windows []platformclientv2.Callablewindow, windowsSchema *schema.Set) *schema.Set {
	if len(windows) == 0 {
		return nil
	}

	var mappedSchema *schema.Set
	var unmappedSchema *schema.Set
	for _, callableWindowsSchema := range windowsSchema.List() {
		mappedSchema = callableWindowsSchema.(map[string]interface{})["mapped"].(*schema.Set)
		unmappedSchema = callableWindowsSchema.(map[string]interface{})["unmapped"].(*schema.Set)
	}

	callableWindowsSet := schema.NewSet(schema.HashResource(outboundSettingsCallableWindowsResource), []interface{}{})
	for _, callableWindow := range windows {
		callableWindowMap := make(map[string]interface{})

		if callableWindow.Mapped != nil {
			callableWindowMap["mapped"] = flattenOutboundSettingsMapped(callableWindow.Mapped, mappedSchema)
		}
		if callableWindow.Unmapped != nil {
			callableWindowMap["unmapped"] = flattenOutboundSettingsUnmapped(callableWindow.Unmapped, unmappedSchema)
		}

		callableWindowsSet.Add(callableWindowMap)
	}
	return callableWindowsSet
}

func flattenOutboundSettingsMapped(mapped *platformclientv2.Atzmtimeslot, mappedSchema *schema.Set) *schema.Set {
	requestSet := schema.NewSet(schema.HashResource(outboundSettingsMappedResource), []interface{}{})
	requestMap := make(map[string]interface{})

	mappedSchemaMap := mappedSchema.List()[0].(map[string]interface{})
	earliestTimeSchema := mappedSchemaMap["earliest_callable_time"].(string)
	latestTimeSchema := mappedSchemaMap["latest_callable_time"].(string)

	if earliestTimeSchema != "" {
		if mapped.EarliestCallableTime != nil {
			requestMap["earliest_callable_time"] = *mapped.EarliestCallableTime
		}
	}
	if latestTimeSchema != "" {
		if mapped.LatestCallableTime != nil {
			requestMap["latest_callable_time"] = *mapped.LatestCallableTime
		}
	}
	requestSet.Add(requestMap)

	return requestSet
}

func flattenOutboundSettingsUnmapped(unmapped *platformclientv2.Atzmtimeslotwithtimezone, unmappedSchema *schema.Set) *schema.Set {
	requestSet := schema.NewSet(schema.HashResource(outboundSettingsMappedResource), []interface{}{})
	requestMap := make(map[string]interface{})

	mappedSchemaMap := unmappedSchema.List()[0].(map[string]interface{})
	earliestTimeSchema := mappedSchemaMap["earliest_callable_time"].(string)
	latestTimeSchema := mappedSchemaMap["latest_callable_time"].(string)
	timeZone := mappedSchemaMap["time_zone_id"].(string)

	if earliestTimeSchema != "" {
		if unmapped.EarliestCallableTime != nil {
			requestMap["earliest_callable_time"] = *unmapped.EarliestCallableTime
		}
	}
	if latestTimeSchema != "" {
		if unmapped.LatestCallableTime != nil {
			requestMap["latest_callable_time"] = *unmapped.LatestCallableTime
		}
	}
	if timeZone != "" {
		if unmapped.TimeZoneId != nil {
			requestMap["time_zone_id"] = *unmapped.TimeZoneId
		}
	}
	requestSet.Add(requestMap)

	return requestSet
}

func deleteOutboundSettings(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return nil
}
