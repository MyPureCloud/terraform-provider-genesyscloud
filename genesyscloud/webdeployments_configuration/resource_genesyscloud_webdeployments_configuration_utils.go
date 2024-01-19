package webdeployments_configuration

import (
	"context"
	"strings"
	"terraform-provider-genesyscloud/genesyscloud/util/lists"
	"terraform-provider-genesyscloud/genesyscloud/util/stringmap"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v116/platformclientv2"
)

func customizeConfigurationDiff(ctx context.Context, diff *schema.ResourceDiff, meta interface{}) error {
	if len(diff.GetChangedKeysPrefix("")) > 0 {
		// When any change is made to the configuration we automatically publish a new version, so mark the version as updated
		// so dependent deployments will update appropriately to reference the newest version
		diff.SetNewComputed("version")
	}
	return nil
}

func readSelectorEventTriggers(triggers []interface{}) *[]platformclientv2.Selectoreventtrigger {
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

func readFormsTrackTriggers(triggers []interface{}) *[]platformclientv2.Formstracktrigger {
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

func readIdleEventTriggers(triggers []interface{}) *[]platformclientv2.Idleeventtrigger {
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

func readScrollPercentageEventTriggers(triggers []interface{}) *[]platformclientv2.Scrollpercentageeventtrigger {
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

func readMessengerSettings(d *schema.ResourceData) *platformclientv2.Messengersettings {
	value, ok := d.GetOk("messenger")
	if !ok {
		return nil
	}

	cfgs := value.([]interface{})
	if len(cfgs) < 1 {
		return nil
	}

	cfg := cfgs[0].(map[string]interface{})
	enabled, _ := cfg["enabled"].(bool)
	messengerSettings := &platformclientv2.Messengersettings{
		Enabled: &enabled,
	}

	if styles, ok := cfg["styles"].([]interface{}); ok && len(styles) > 0 {
		style := styles[0].(map[string]interface{})
		if primaryColor, ok := style["primary_color"].(string); ok {
			messengerSettings.Styles = &platformclientv2.Messengerstyles{
				PrimaryColor: &primaryColor,
			}
		}
	}

	if launchers, ok := cfg["launcher_button"].([]interface{}); ok && len(launchers) > 0 {
		launcher := launchers[0].(map[string]interface{})
		if visibility, ok := launcher["visibility"].(string); ok {
			messengerSettings.LauncherButton = &platformclientv2.Launcherbuttonsettings{
				Visibility: &visibility,
			}
		}
	}

	if screens, ok := cfg["home_screen"].([]interface{}); ok && len(screens) > 0 {
		if screen, ok := screens[0].(map[string]interface{}); ok {
			enabled, enabledOk := screen["enabled"].(bool)
			logoUrl, logoUrlOk := screen["logo_url"].(string)

			if enabledOk && logoUrlOk {
				messengerSettings.HomeScreen = &platformclientv2.Messengerhomescreen{
					Enabled: &enabled,
					LogoUrl: &logoUrl,
				}
			}
		}
	}

	if fileUploads, ok := cfg["file_upload"].([]interface{}); ok && len(fileUploads) > 0 {
		fileUpload := fileUploads[0].(map[string]interface{})
		if modesCfg, ok := fileUpload["mode"].([]interface{}); ok && len(modesCfg) > 0 {
			modes := make([]platformclientv2.Fileuploadmode, len(modesCfg))
			for i, modeCfg := range modesCfg {
				if mode, ok := modeCfg.(map[string]interface{}); ok {
					maxFileSize := mode["max_file_size_kb"].(int)
					fileTypes := lists.InterfaceListToStrings(mode["file_types"].([]interface{}))
					modes[i] = platformclientv2.Fileuploadmode{
						FileTypes:     &fileTypes,
						MaxFileSizeKB: &maxFileSize,
					}
				}
			}

			if len(modes) > 0 {
				messengerSettings.FileUpload = &platformclientv2.Fileuploadsettings{
					Modes: &modes,
				}
			}
		}
	}

	return messengerSettings
}

func readCobrowseSettings(d *schema.ResourceData) *platformclientv2.Cobrowsesettings {
	value, ok := d.GetOk("cobrowse")
	if !ok {
		return nil
	}

	cfgs := value.([]interface{})
	if len(cfgs) < 1 {
		return nil
	}

	cfg := cfgs[0].(map[string]interface{})

	enabled, _ := cfg["enabled"].(bool)
	allowAgentControl, _ := cfg["allow_agent_control"].(bool)
	channels := lists.InterfaceListToStrings(cfg["channels"].([]interface{}))
	maskSelectors := lists.InterfaceListToStrings(cfg["mask_selectors"].([]interface{}))
	readonlySelectors := lists.InterfaceListToStrings(cfg["readonly_selectors"].([]interface{}))

	return &platformclientv2.Cobrowsesettings{
		Enabled:           &enabled,
		AllowAgentControl: &allowAgentControl,
		Channels:          &channels,
		MaskSelectors:     &maskSelectors,
		ReadonlySelectors: &readonlySelectors,
	}
}

func readSupportCenterCategory(triggers []interface{}) *[]platformclientv2.Supportcentercategory {
	if triggers == nil || len(triggers) < 1 {
		return nil
	}

	results := make([]platformclientv2.Supportcentercategory, len(triggers))
	for i, value := range triggers {
		if trigger, ok := value.(map[string]interface{}); ok {
			id := trigger["enabled_categories_id"].(string)
			selfuri := trigger["self_uri"].(string)
			imageSource := trigger["image_source"].(string)

			image := &platformclientv2.Supportcenterimage{
				Source: &platformclientv2.Supportcenterimagesource{
					DefaultUrl: &imageSource,
				},
			}

			results[i] = platformclientv2.Supportcentercategory{
				Id:      &id,
				SelfUri: &selfuri,
				Image:   image,
			}
		}
	}

	return &results
}

func readSupportCenterCustomMessage(triggers []interface{}) *[]platformclientv2.Supportcentercustommessage {
	if triggers == nil || len(triggers) < 1 {
		return nil
	}

	results := make([]platformclientv2.Supportcentercustommessage, len(triggers))
	for i, value := range triggers {
		if trigger, ok := value.(map[string]interface{}); ok {
			defaultValue := trigger["default_value"].(string)
			varType := trigger["type"].(string)

			results[i] = platformclientv2.Supportcentercustommessage{
				DefaultValue: &defaultValue,
				VarType:      &varType,
			}
		}
	}

	return &results
}

func readSupportCenterStyleSetting(triggers []interface{}) *[]platformclientv2.Supportcenterstylesetting {
	if triggers == nil || len(triggers) < 1 {
		return nil
	}

	results := make([]platformclientv2.Supportcenterstylesetting, len(triggers))
	for i, value := range triggers {
		if trigger, ok := value.(map[string]interface{}); ok {
			herobackground := trigger["hero_style_background_color"].(string)
			herotextcolor := trigger["hero_style_text_color"].(string)
			heroimage := trigger["hero_style_image"].(string)

			herostyle := platformclientv2.Supportcenterstylesetting{
				HeroStyle: &platformclientv2.Supportcenterherostyle{
					BackgroundColor: &herobackground,
					TextColor:       &herotextcolor,
					Image: &platformclientv2.Supportcenterimage{
						Source: &platformclientv2.Supportcenterimagesource{
							DefaultUrl: &heroimage,
						},
					},
				},
			}

			globalbackground := trigger["global_style_background_color"].(string)
			globalprimarycolor := trigger["global_style_primary_color"].(string)
			globalprimarycolordark := trigger["global_style_primary_color_dark"].(string)
			globalprimarycolorlight := trigger["global_style_primary_color_light"].(string)
			globaltextcolor := trigger["global_style_text_color"].(string)
			globalfontfamily := trigger["global_style_font_family"].(string)

			globalstyle := platformclientv2.Supportcenterstylesetting{
				GlobalStyle: &platformclientv2.Supportcenterglobalstyle{
					BackgroundColor:   &globalbackground,
					PrimaryColor:      &globalprimarycolor,
					PrimaryColorDark:  &globalprimarycolordark,
					PrimaryColorLight: &globalprimarycolorlight,
					TextColor:         &globaltextcolor,
					FontFamily:        &globalfontfamily,
				},
			}

			// Assuming Supportcenterstylesetting has both HeroStyle and GlobalStyle
			results[i] = platformclientv2.Supportcenterstylesetting{
				HeroStyle:   herostyle.HeroStyle,
				GlobalStyle: globalstyle.GlobalStyle,
			}
		}
	}

	return &results
}

func readSupportCenterScreens(triggers []interface{}) *[]platformclientv2.Supportcenterscreen {
	if triggers == nil || len(triggers) < 1 {
		return nil
	}

	results := make([]platformclientv2.Supportcenterscreen, len(triggers))
	for i, value := range triggers {
		if trigger, ok := value.(map[string]interface{}); ok {
			varType := trigger["type"].(string)
			moduleSettingsType := trigger["module_settings_type"].(string)
			moduleSettingsEnabled := trigger["module_settings_enabled"].(bool)
			moduleSettingsCompactCategoryModuleTemplate := trigger["module_settings_compact_category_module_template"].(bool)
			moduleSettingsDetailedCategoryModuleTemplate := trigger["module_settings_detailed_category_module_template"].(bool)

			moduleSettings := []platformclientv2.Supportcentermodulesetting{
				{
					VarType: &moduleSettingsType,
					Enabled: &moduleSettingsEnabled,
					CompactCategoryModuleTemplate: &platformclientv2.Supportcentercompactcategorymoduletemplate{
						Active: &moduleSettingsCompactCategoryModuleTemplate,
					},
					DetailedCategoryModuleTemplate: &platformclientv2.Supportcenterdetailedcategorymoduletemplate{
						Active: &moduleSettingsDetailedCategoryModuleTemplate,
					},
				},
			}

			results[i] = platformclientv2.Supportcenterscreen{
				VarType:        &varType,
				ModuleSettings: &moduleSettings,
			}
		}
	}

	return &results
}

// featureNotImplemented checks the response object to find out if the request failed because a feature is not yet
// implemented in the org that it was ran against. If true, we can pass back the field name and give more context
// in the final error message.
func featureNotImplemented(response *platformclientv2.APIResponse) (bool, string) {
	if response.Error == nil || response.Error.Details == nil || len(response.Error.Details) == 0 {
		return false, ""
	}
	for _, err := range response.Error.Details {
		if err.FieldName == nil {
			continue
		}
		if strings.Contains(*err.ErrorCode, "feature is not yet implemented") {
			return true, *err.FieldName
		}
	}
	return false, ""
}

func validateConfigurationStatusChange(k, old, new string, d *schema.ResourceData) bool {
	// Configs start in a pending status and may not transition to active or error before we retrieve the state, so allow
	// the status to change from pending to something less ephemeral
	return old == "Pending"
}

func flattenMessengerSettings(messengerSettings *platformclientv2.Messengersettings) []interface{} {
	if messengerSettings == nil {
		return nil
	}

	return []interface{}{map[string]interface{}{
		"enabled":         messengerSettings.Enabled,
		"styles":          flattenStyles(messengerSettings.Styles),
		"launcher_button": flattenLauncherButton(messengerSettings.LauncherButton),
		"home_screen":     flattenHomeScreen(messengerSettings.HomeScreen),
		"file_upload":     flattenFileUpload(messengerSettings.FileUpload),
	}}
}

func flattenCobrowseSettings(cobrowseSettings *platformclientv2.Cobrowsesettings) []interface{} {
	if cobrowseSettings == nil {
		return nil
	}

	return []interface{}{map[string]interface{}{
		"enabled":             cobrowseSettings.Enabled,
		"allow_agent_control": cobrowseSettings.AllowAgentControl,
		"channels":            cobrowseSettings.Channels,
		"mask_selectors":      cobrowseSettings.MaskSelectors,
		"readonly_selectors":  cobrowseSettings.ReadonlySelectors,
	}}
}

func flattenStyles(styles *platformclientv2.Messengerstyles) []interface{} {
	if styles == nil {
		return nil
	}

	return []interface{}{map[string]interface{}{
		"primary_color": styles.PrimaryColor,
	}}
}

func flattenLauncherButton(settings *platformclientv2.Launcherbuttonsettings) []interface{} {
	if settings == nil {
		return nil
	}

	return []interface{}{map[string]interface{}{
		"visibility": settings.Visibility,
	}}
}

func flattenHomeScreen(settings *platformclientv2.Messengerhomescreen) []interface{} {
	if settings == nil {
		return nil
	}

	return []interface{}{map[string]interface{}{
		"enabled":  settings.Enabled,
		"logo_url": settings.LogoUrl,
	}}
}

func flattenFileUpload(settings *platformclientv2.Fileuploadsettings) []interface{} {
	if settings == nil || settings.Modes == nil || len(*settings.Modes) < 1 {
		return nil
	}

	modes := make([]map[string]interface{}, len(*settings.Modes))
	for i, mode := range *settings.Modes {
		modes[i] = map[string]interface{}{
			"file_types":       *mode.FileTypes,
			"max_file_size_kb": *mode.MaxFileSizeKB,
		}
	}

	return []interface{}{map[string]interface{}{
		"mode": modes,
	}}
}

func flattenJourneyEvents(journeyEvents *platformclientv2.Journeyeventssettings) []interface{} {
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

func flattenSupportCenterSettings(supportCenterSettings *platformclientv2.Supportcentersettings) []interface{} {
	if supportCenterSettings == nil {
		return nil
	}

	screens := &supportCenterSettings.Screens

	return []interface{}{map[string]interface{}{
		"enabled":                          supportCenterSettings.Enabled,
		"knowledge_base_id":                flattenKnowledgeBaseId(supportCenterSettings.KnowledgeBase),
		"router_type":                      supportCenterSettings.RouterType,
		"custom_messages":                  flattenSupportCenterCustomMessage(supportCenterSettings.CustomMessages),
		"feedback_enabled":                 supportCenterSettings.Feedback.Enabled,
		"enabled_categories":               flattenSupportCenterCategory(supportCenterSettings.EnabledCategories),
		"flattenSupportCenterStyleSetting": flattenSupportCenterStyleSetting(supportCenterSettings.StyleSetting),
		"flattenSupportCenterScreens":      flattenSupportCenterScreens(screens),
	}}
}

func flattenSupportCenterCategory(categories *[]platformclientv2.Supportcentercategory) []interface{} {
	if categories == nil || len(*categories) < 1 {
		return nil
	}

	result := make([]interface{}, len(*categories))
	for i, category := range *categories {

		imgSrc := ""
		if category.Image != nil && category.Image.Source != nil && category.Image.Source.DefaultUrl != nil {
			imgSrc = *category.Image.Source.DefaultUrl
		}

		result[i] = map[string]interface{}{
			"enabled_categories_id": category.Id,
			"self_uri":              category.SelfUri,
			"image_source":          imgSrc,
		}
	}
	return result
}

func flattenSupportCenterStyleSetting(styleSetting *platformclientv2.Supportcenterstylesetting) []interface{} {

	if styleSetting == nil {
		return nil

	}

	heroStyles := styleSetting.HeroStyle
	globalStyles := styleSetting.GlobalStyle

	heroResult := flattenSupportCenterHeroStyle(heroStyles)
	globalResult := flattenSupportCenterGlobalStyle(globalStyles)

	result := stringmap.MergeSingularMaps(heroResult, globalResult)

	result = append(result, heroResult...)
	result = append(result, globalResult...)

	return result
}

func flattenSupportCenterHeroStyle(heroStyles *platformclientv2.Supportcenterherostyle) map[string]interface{} {
	if heroStyles == nil {
		return nil
	}

	result := map[string]interface{}{
		"hero_style_background_color": heroStyles.BackgroundColor,
		"hero_style_text_color":       heroStyles.TextColor,
		"hero_style_image":            heroStyles.Image,
	}

	return result
}

func flattenSupportCenterGlobalStyle(globalStyles *platformclientv2.Supportcenterglobalstyle) map[string]interface{} {
	if globalStyles == nil {
		return nil
	}

	result := map[string]interface{}{
		"global_style_background_color":    globalStyles.BackgroundColor,
		"global_style_primary_color":       globalStyles.PrimaryColor,
		"global_style_primary_color_dark":  globalStyles.PrimaryColorDark,
		"global_style_primary_color_light": globalStyles.PrimaryColorLight,
		"global_style_text_color":          globalStyles.TextColor,
		"global_style_font_family":         globalStyles.FontFamily,
	}

	return result
}

func flattenSupportCenterScreens(supportcentermodulesettings *[]platformclientv2.Supportcentermodulesetting) []interface{} {
	if supportcentermodulesettings == nil {
		return nil
	}

	// flattened := make([]interface{}, len(*flattened))
	var flattened []interface{}
	for _, supportcentermodulesetting := range *supportcentermodulesettings {
		flattened = append(flattened, map[string]interface{}{
			"enabled":              *supportcentermodulesetting.Enabled,
			"module_settings_type": *supportcentermodulesetting.VarType,
			"module_settings_compact_category_module_template":  *supportcentermodulesetting.CompactCategoryModuleTemplate,
			"module_settings_detailed_category_module_template": *supportcentermodulesetting.DetailedCategoryModuleTemplate,
		})
	}
	return flattened
}

func flattenSupportCenterCustomMessage(customMessage *[]platformclientv2.Supportcentercustommessage) []interface{} {
	if customMessage == nil || len(*customMessage) < 1 {
		return nil
	}

	result := make([]interface{}, len(*customMessage))
	for i, customMessage := range *customMessage {
		result[i] = map[string]interface{}{
			"default_value": customMessage.DefaultValue,
			"type":          customMessage.VarType,
		}
	}
	return result
}

func flattenKnowledgeBaseId(knowledgebase *platformclientv2.Addressableentityref) string {
	if knowledgebase == nil {
		return ""
	}

	return *knowledgebase.Id
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

func readWebDeploymentConfigurationFromResourceData(d *schema.ResourceData) (string, *platformclientv2.Webdeploymentconfigurationversion) {
	name := d.Get("name").(string)
	languages := lists.InterfaceListToStrings(d.Get("languages").([]interface{}))
	defaultLanguage := d.Get("default_language").(string)

	inputCfg := &platformclientv2.Webdeploymentconfigurationversion{
		Name:            &name,
		Languages:       &languages,
		DefaultLanguage: &defaultLanguage,
	}

	description, ok := d.Get("description").(string)
	if ok {
		inputCfg.Description = &description
	}

	messengerSettings := readMessengerSettings(d)
	if messengerSettings != nil {
		inputCfg.Messenger = messengerSettings
	}

	cobrowseSettings := readCobrowseSettings(d)
	if cobrowseSettings != nil {
		inputCfg.Cobrowse = cobrowseSettings
	}

	journeySettings := readJourneySettings(d)
	if journeySettings != nil {
		inputCfg.JourneyEvents = journeySettings
	}

	supportCenterSettings := readSupportCenterSettings(d)
	if supportCenterSettings != nil {
		inputCfg.SupportCenter = supportCenterSettings
	}

	return name, inputCfg
}

func readJourneySettings(d *schema.ResourceData) *platformclientv2.Journeyeventssettings {
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

	if clickEvents := readSelectorEventTriggers(cfg["click_event"].([]interface{})); clickEvents != nil {
		journeySettings.ClickEvents = clickEvents
	}

	if formsTrackEvents := readFormsTrackTriggers(cfg["form_track_event"].([]interface{})); formsTrackEvents != nil {
		journeySettings.FormsTrackEvents = formsTrackEvents
	}

	if idleEvents := readIdleEventTriggers(cfg["idle_event"].([]interface{})); idleEvents != nil {
		journeySettings.IdleEvents = idleEvents
	}

	if inViewportEvents := readSelectorEventTriggers(cfg["in_viewport_event"].([]interface{})); inViewportEvents != nil {
		journeySettings.InViewportEvents = inViewportEvents
	}

	if scrollDepthEvents := readScrollPercentageEventTriggers(cfg["scroll_depth_event"].([]interface{})); scrollDepthEvents != nil {
		journeySettings.ScrollDepthEvents = scrollDepthEvents
	}

	return journeySettings
}

func readSupportCenterSettings(d *schema.ResourceData) *platformclientv2.Supportcentersettings {
	value, ok := d.GetOk("support_center")
	if !ok {
		return nil
	}

	cfgs := value.([]interface{})
	if len(cfgs) < 1 {
		return nil
	}

	cfg := cfgs[0].(map[string]interface{})
	enabled, _ := cfg["enabled"].(bool)
	supportCenterSettings := &platformclientv2.Supportcentersettings{
		Enabled: &enabled,
	}

	if id, ok := cfg["knowledge_base_id"].(string); ok {
		supportCenterSettings.KnowledgeBase = &platformclientv2.Addressableentityref{
			Id: &id,
		}
	}

	if routertype, ok := cfg["router_type"].(string); ok {
		supportCenterSettings.RouterType = &routertype
	}

	if customMessages := readSupportCenterCustomMessage(cfg["custom_messages"].([]interface{})); supportCenterCustomMessage != nil {
		supportCenterSettings.CustomMessages = customMessages
	}

	if supportCenterCategory := readSupportCenterCategory(cfg["enabled_categories"].([]interface{})); supportCenterCategory != nil {
		supportCenterSettings.EnabledCategories = supportCenterCategory
	}

	if feedbackEnabled, ok := cfg["feedback_enabled"].(bool); ok {
		supportCenterSettings.Feedback = &platformclientv2.Supportcenterfeedbacksettings{
			Enabled: &feedbackEnabled,
		}
	}

	return supportCenterSettings
}
