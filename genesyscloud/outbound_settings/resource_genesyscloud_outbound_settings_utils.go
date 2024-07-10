package outbound_settings

import (
	"terraform-provider-genesyscloud/genesyscloud/tfexporter_state"
	"terraform-provider-genesyscloud/genesyscloud/util/lists"
	"terraform-provider-genesyscloud/genesyscloud/util/resourcedata"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

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
	return nil
}

func buildSupportedCountries(d *schema.ResourceData) *[]string {
	var supportedCountries []string
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

func flattenOutboundSettingsAutomaticTimeZoneMapping(timeZoneMappings platformclientv2.Automatictimezonemappingsettings, automaticTimeZoneMapping []interface{}) []interface{} {
	requestMap := make(map[string]interface{})

	if tfexporter_state.IsExporterActive() {
		if timeZoneMappings.CallableWindows != nil {
			requestMap["callable_windows"] = flattenCallableWindows(*timeZoneMappings.CallableWindows, nil)
		}
	} else {
		if len(automaticTimeZoneMapping) > 0 {
			if callableWindows, ok := automaticTimeZoneMapping[0].(map[string]interface{})["callable_windows"].(*schema.Set); ok {
				if timeZoneMappings.CallableWindows != nil {
					requestMap["callable_windows"] = flattenCallableWindows(*timeZoneMappings.CallableWindows, callableWindows)
				}
			}
		}
	}

	resourcedata.SetMapValueIfNotNil(requestMap, "supported_countries", timeZoneMappings.SupportedCountries)
	return []interface{}{requestMap}
}

func flattenCallableWindows(windows []platformclientv2.Callablewindow, windowsSchema *schema.Set) *schema.Set {
	if len(windows) == 0 {
		return nil
	}

	callableWindowMap := make(map[string]interface{})
	callableWindowsSet := schema.NewSet(schema.HashResource(callableWindowsResource), []interface{}{})

	if tfexporter_state.IsExporterActive() {
		for _, callableWindow := range windows {
			if callableWindow.Mapped != nil {
				callableWindowMap["mapped"] = flattenOutboundSettingsMapped(callableWindow.Mapped, nil)
			}
			if callableWindow.Unmapped != nil {
				callableWindowMap["unmapped"] = flattenOutboundSettingsUnmapped(callableWindow.Unmapped, nil)
			}
		}
	} else {
		var mappedSchema *schema.Set
		var unmappedSchema *schema.Set

		for _, callableWindowsSchema := range windowsSchema.List() {
			mappedSchema = callableWindowsSchema.(map[string]interface{})["mapped"].(*schema.Set)
			unmappedSchema = callableWindowsSchema.(map[string]interface{})["unmapped"].(*schema.Set)
		}

		for _, callableWindow := range windows {
			if callableWindow.Mapped != nil {
				callableWindowMap["mapped"] = flattenOutboundSettingsMapped(callableWindow.Mapped, mappedSchema)
			}
			if callableWindow.Unmapped != nil {
				callableWindowMap["unmapped"] = flattenOutboundSettingsUnmapped(callableWindow.Unmapped, unmappedSchema)
			}
		}
	}

	callableWindowsSet.Add(callableWindowMap)
	return callableWindowsSet
}

func flattenOutboundSettingsMapped(mapped *platformclientv2.Atzmtimeslot, mappedSchema *schema.Set) *schema.Set {
	requestSet := schema.NewSet(schema.HashResource(mappedResource), []interface{}{})
	requestMap := make(map[string]interface{})

	if tfexporter_state.IsExporterActive() {
		resourcedata.SetMapValueIfNotNil(requestMap, "earliest_callable_time", mapped.EarliestCallableTime)
		resourcedata.SetMapValueIfNotNil(requestMap, "latest_callable_time", mapped.LatestCallableTime)
	} else {
		mappedSchemaMap := mappedSchema.List()[0].(map[string]interface{})
		earliestTimeSchema := mappedSchemaMap["earliest_callable_time"].(string)
		latestTimeSchema := mappedSchemaMap["latest_callable_time"].(string)

		if earliestTimeSchema != "" {
			resourcedata.SetMapValueIfNotNil(requestMap, "earliest_callable_time", mapped.EarliestCallableTime)
		}
		if latestTimeSchema != "" {
			resourcedata.SetMapValueIfNotNil(requestMap, "latest_callable_time", mapped.LatestCallableTime)
		}
	}

	requestSet.Add(requestMap)
	return requestSet
}

func flattenOutboundSettingsUnmapped(unmapped *platformclientv2.Atzmtimeslotwithtimezone, unmappedSchema *schema.Set) *schema.Set {
	requestSet := schema.NewSet(schema.HashResource(UnmappedResource), []interface{}{})
	requestMap := make(map[string]interface{})

	if tfexporter_state.IsExporterActive() {
		resourcedata.SetMapValueIfNotNil(requestMap, "earliest_callable_time", unmapped.EarliestCallableTime)
		resourcedata.SetMapValueIfNotNil(requestMap, "latest_callable_time", unmapped.LatestCallableTime)
		resourcedata.SetMapValueIfNotNil(requestMap, "time_zone_id", unmapped.TimeZoneId)
	} else {
		mappedSchemaMap := unmappedSchema.List()[0].(map[string]interface{})
		earliestTimeSchema := mappedSchemaMap["earliest_callable_time"].(string)
		latestTimeSchema := mappedSchemaMap["latest_callable_time"].(string)
		timeZone := mappedSchemaMap["time_zone_id"].(string)

		if earliestTimeSchema != "" {
			resourcedata.SetMapValueIfNotNil(requestMap, "earliest_callable_time", unmapped.EarliestCallableTime)
		}
		if latestTimeSchema != "" {
			resourcedata.SetMapValueIfNotNil(requestMap, "latest_callable_time", unmapped.LatestCallableTime)
		}
		if timeZone != "" {
			resourcedata.SetMapValueIfNotNil(requestMap, "time_zone_id", unmapped.TimeZoneId)
		}
	}

	requestSet.Add(requestMap)
	return requestSet
}
