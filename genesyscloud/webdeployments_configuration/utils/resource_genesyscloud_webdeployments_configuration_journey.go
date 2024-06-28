package webdeployments_configuration_utils

import (
	"terraform-provider-genesyscloud/genesyscloud/util/lists"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

func buildSelectorEventTriggers(triggers []interface{}) *[]platformclientv2.Selectoreventtrigger {
	if triggers == nil || len(triggers) < 1 {
		return nil
	}

	results := make([]platformclientv2.Selectoreventtrigger, len(triggers))
	for i, value := range triggers {
		if trigger, ok := value.(map[string]interface{}); ok {
			selector := trigger["selector"].(string)
			eventName := trigger["event_name"].(string)
			results[i] = platformclientv2.Selectoreventtrigger{
				Selector:  &selector,
				EventName: &eventName,
			}
		}
	}

	return &results
}

func buildFormsTrackTriggers(triggers []interface{}) *[]platformclientv2.Formstracktrigger {
	if triggers == nil || len(triggers) < 1 {
		return nil
	}

	results := make([]platformclientv2.Formstracktrigger, len(triggers))
	for i, value := range triggers {
		if trigger, ok := value.(map[string]interface{}); ok {
			selector := trigger["selector"].(string)
			formName := trigger["form_name"].(string)
			captureDataOnAbandon := trigger["capture_data_on_form_abandon"].(bool)
			captureDataOnSubmit := trigger["capture_data_on_form_submit"].(bool)
			results[i] = platformclientv2.Formstracktrigger{
				Selector:                 &selector,
				FormName:                 &formName,
				CaptureDataOnFormAbandon: &captureDataOnAbandon,
				CaptureDataOnFormSubmit:  &captureDataOnSubmit,
			}
		}
	}

	return &results
}

func buildIdleEventTriggers(triggers []interface{}) *[]platformclientv2.Idleeventtrigger {
	if triggers == nil || len(triggers) < 1 {
		return nil
	}

	results := make([]platformclientv2.Idleeventtrigger, len(triggers))
	for i, value := range triggers {
		if trigger, ok := value.(map[string]interface{}); ok {
			eventName := trigger["event_name"].(string)
			idleAfterSeconds := trigger["idle_after_seconds"].(int)
			results[i] = platformclientv2.Idleeventtrigger{
				EventName:        &eventName,
				IdleAfterSeconds: &idleAfterSeconds,
			}
		}
	}

	return &results
}

func buildScrollPercentageEventTriggers(triggers []interface{}) *[]platformclientv2.Scrollpercentageeventtrigger {
	if triggers == nil || len(triggers) < 1 {
		return nil
	}

	results := make([]platformclientv2.Scrollpercentageeventtrigger, len(triggers))
	for i, value := range triggers {
		if trigger, ok := value.(map[string]interface{}); ok {
			eventName := trigger["event_name"].(string)
			percentage := trigger["percentage"].(int)
			results[i] = platformclientv2.Scrollpercentageeventtrigger{
				EventName:  &eventName,
				Percentage: &percentage,
			}
		}
	}

	return &results
}

func flattenSelectorEventTriggers(triggers *[]platformclientv2.Selectoreventtrigger) []interface{} {
	if triggers == nil || len(*triggers) < 1 {
		return nil
	}

	result := make([]interface{}, len(*triggers))
	for i, trigger := range *triggers {
		result[i] = map[string]interface{}{
			"selector":   trigger.Selector,
			"event_name": trigger.EventName,
		}
	}
	return result
}

func flattenFormsTrackTriggers(triggers *[]platformclientv2.Formstracktrigger) []interface{} {
	if triggers == nil || len(*triggers) < 1 {
		return nil
	}

	result := make([]interface{}, len(*triggers))
	for i, trigger := range *triggers {
		result[i] = map[string]interface{}{
			"selector":                     trigger.Selector,
			"form_name":                    trigger.FormName,
			"capture_data_on_form_abandon": trigger.CaptureDataOnFormAbandon,
			"capture_data_on_form_submit":  trigger.CaptureDataOnFormSubmit,
		}
	}
	return result
}

func flattenIdleEventTriggers(triggers *[]platformclientv2.Idleeventtrigger) []interface{} {
	if triggers == nil || len(*triggers) < 1 {
		return nil
	}

	result := make([]interface{}, len(*triggers))
	for i, trigger := range *triggers {
		result[i] = map[string]interface{}{
			"event_name":         trigger.EventName,
			"idle_after_seconds": trigger.IdleAfterSeconds,
		}
	}
	return result
}

func flattenScrollPercentageEventTriggers(triggers *[]platformclientv2.Scrollpercentageeventtrigger) []interface{} {
	if triggers == nil || len(*triggers) < 1 {
		return nil
	}

	result := make([]interface{}, len(*triggers))
	for i, trigger := range *triggers {
		result[i] = map[string]interface{}{
			"event_name": trigger.EventName,
			"percentage": trigger.Percentage,
		}
	}
	return result
}

func buildJourneySettings(d *schema.ResourceData) *platformclientv2.Journeyeventssettings {
	value, ok := d.GetOk("journey_events")
	if !ok {
		return nil
	}

	cfgs := value.([]interface{})
	if len(cfgs) < 1 {
		return nil
	}

	cfg := cfgs[0].(map[string]interface{})
	enabled, _ := cfg["enabled"].(bool)
	journeySettings := &platformclientv2.Journeyeventssettings{
		Enabled: &enabled,
	}

	excludedQueryParams := lists.InterfaceListToStrings(cfg["excluded_query_parameters"].([]interface{}))
	journeySettings.ExcludedQueryParameters = &excludedQueryParams

	if keepUrlFragment, ok := cfg["should_keep_url_fragment"].(bool); ok && keepUrlFragment {
		journeySettings.ShouldKeepUrlFragment = &keepUrlFragment
	}

	searchQueryParameters := lists.InterfaceListToStrings(cfg["search_query_parameters"].([]interface{}))
	journeySettings.SearchQueryParameters = &searchQueryParameters

	pageviewConfig := cfg["pageview_config"]
	if value, ok := pageviewConfig.(string); ok {
		if value != "" {
			journeySettings.PageviewConfig = &value
		}
	}

	if clickEvents := buildSelectorEventTriggers(cfg["click_event"].([]interface{})); clickEvents != nil {
		journeySettings.ClickEvents = clickEvents
	}

	if formsTrackEvents := buildFormsTrackTriggers(cfg["form_track_event"].([]interface{})); formsTrackEvents != nil {
		journeySettings.FormsTrackEvents = formsTrackEvents
	}

	if idleEvents := buildIdleEventTriggers(cfg["idle_event"].([]interface{})); idleEvents != nil {
		journeySettings.IdleEvents = idleEvents
	}

	if inViewportEvents := buildSelectorEventTriggers(cfg["in_viewport_event"].([]interface{})); inViewportEvents != nil {
		journeySettings.InViewportEvents = inViewportEvents
	}

	if scrollDepthEvents := buildScrollPercentageEventTriggers(cfg["scroll_depth_event"].([]interface{})); scrollDepthEvents != nil {
		journeySettings.ScrollDepthEvents = scrollDepthEvents
	}

	return journeySettings
}

func FlattenJourneyEvents(journeyEvents *platformclientv2.Journeyeventssettings) []interface{} {
	if journeyEvents == nil {
		return nil
	}

	return []interface{}{map[string]interface{}{
		"enabled":                   journeyEvents.Enabled,
		"excluded_query_parameters": journeyEvents.ExcludedQueryParameters,
		"should_keep_url_fragment":  journeyEvents.ShouldKeepUrlFragment,
		"search_query_parameters":   journeyEvents.SearchQueryParameters,
		"pageview_config":           journeyEvents.PageviewConfig,
		"click_event":               flattenSelectorEventTriggers(journeyEvents.ClickEvents),
		"form_track_event":          flattenFormsTrackTriggers(journeyEvents.FormsTrackEvents),
		"idle_event":                flattenIdleEventTriggers(journeyEvents.IdleEvents),
		"in_viewport_event":         flattenSelectorEventTriggers(journeyEvents.InViewportEvents),
		"scroll_depth_event":        flattenScrollPercentageEventTriggers(journeyEvents.ScrollDepthEvents),
	}}
}
