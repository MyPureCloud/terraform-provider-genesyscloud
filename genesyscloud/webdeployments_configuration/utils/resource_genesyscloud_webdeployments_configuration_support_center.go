package webdeployments_configuration_utils

import (
	"terraform-provider-genesyscloud/genesyscloud/util/resourcedata"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

func buildSupportCenterHeroStyle(styles []interface{}) *platformclientv2.Supportcenterherostyle {
	if styles == nil || len(styles) < 1 {
		return nil
	}

	style := styles[0].(map[string]interface{})
	heroStyle := &platformclientv2.Supportcenterherostyle{
		BackgroundColor: platformclientv2.String(style["background_color"].(string)),
		TextColor:       platformclientv2.String(style["text_color"].(string)),
	}

	if imageUri, ok := style["image_uri"].(string); ok {
		heroStyle.Image = &platformclientv2.Supportcenterimage{
			Source: &platformclientv2.Supportcenterimagesource{
				DefaultUrl: &imageUri,
			},
		}
	}

	return heroStyle
}

func buildSupportCenterGlobalStyle(styles []interface{}) *platformclientv2.Supportcenterglobalstyle {
	if styles == nil || len(styles) < 1 {
		return nil
	}

	style := styles[0].(map[string]interface{})
	globalStyle := &platformclientv2.Supportcenterglobalstyle{
		BackgroundColor:   platformclientv2.String(style["background_color"].(string)),
		PrimaryColor:      platformclientv2.String(style["primary_color"].(string)),
		PrimaryColorDark:  platformclientv2.String(style["primary_color_dark"].(string)),
		PrimaryColorLight: platformclientv2.String(style["primary_color_light"].(string)),
		TextColor:         platformclientv2.String(style["text_color"].(string)),
		FontFamily:        platformclientv2.String(style["font_family"].(string)),
	}

	return globalStyle
}

func buildStyleSettings(styles []interface{}) *platformclientv2.Supportcenterstylesetting {
	if styles == nil || len(styles) < 1 {
		return nil
	}

	style := styles[0].(map[string]interface{})
	styleSetting := &platformclientv2.Supportcenterstylesetting{
		HeroStyle:   buildSupportCenterHeroStyle(style["hero_style_setting"].([]interface{})),
		GlobalStyle: buildSupportCenterGlobalStyle(style["global_style_setting"].([]interface{})),
	}

	return styleSetting
}

func buildEnabledCategories(categories []interface{}) *[]platformclientv2.Supportcentercategory {
	if categories == nil || len(categories) < 1 {
		return nil
	}

	results := make([]platformclientv2.Supportcentercategory, len(categories))
	for i, value := range categories {
		category, ok := value.(map[string]interface{})
		if !ok {
			continue
		}

		scCategory := platformclientv2.Supportcentercategory{
			Id: platformclientv2.String(category["category_id"].(string)),
		}

		if imageUri, ok := category["image_uri"].(string); ok {
			scCategory.Image = &platformclientv2.Supportcenterimage{
				Source: &platformclientv2.Supportcenterimagesource{
					DefaultUrl: &imageUri,
				},
			}
		}

		results[i] = scCategory
	}

	return &results
}

func buildCustomMessages(messages []interface{}) *[]platformclientv2.Supportcentercustommessage {
	if messages == nil || len(messages) < 1 {
		return nil
	}

	results := make([]platformclientv2.Supportcentercustommessage, len(messages))
	for i, value := range messages {
		if message, ok := value.(map[string]interface{}); ok {
			results[i] = platformclientv2.Supportcentercustommessage{
				DefaultValue: platformclientv2.String(message["default_value"].(string)),
				VarType:      platformclientv2.String(message["type"].(string)),
			}
		}
	}

	return &results
}

func buildModuleSettings(settings []interface{}) *[]platformclientv2.Supportcentermodulesetting {
	if settings == nil || len(settings) < 1 {
		return nil
	}

	ret := make([]platformclientv2.Supportcentermodulesetting, len(settings))
	for i, setting := range settings {
		settingMap, ok := setting.(map[string]interface{})
		if !ok {
			continue
		}

		moduleSetting := platformclientv2.Supportcentermodulesetting{
			VarType: resourcedata.GetNillableValueFromMap[string](settingMap, "type"),
			Enabled: resourcedata.GetNillableValueFromMap[bool](settingMap, "enabled"),
		}

		if compactModActive, ok := settingMap["compact_category_module_template_active"].(bool); ok {
			moduleSetting.CompactCategoryModuleTemplate = &platformclientv2.Supportcentercompactcategorymoduletemplate{
				Active: &compactModActive,
			}
		}

		if detailedModTempArr, ok := settingMap["detailed_category_module_template"].([]interface{}); ok && len(detailedModTempArr) > 0 {
			detailedModTemp := detailedModTempArr[0].(map[string]interface{})
			moduleSetting.DetailedCategoryModuleTemplate = &platformclientv2.Supportcenterdetailedcategorymoduletemplate{
				Active: platformclientv2.Bool(detailedModTemp["active"].(bool)),
				Sidebar: &platformclientv2.Supportcenterdetailedcategorymodulesidebar{
					Enabled: platformclientv2.Bool(detailedModTemp["sidebar_enabled"].(bool)),
				},
			}
		}

		ret[i] = moduleSetting
	}

	return &ret
}

func buildScreens(screens []interface{}) *[]platformclientv2.Supportcenterscreen {
	if screens == nil || len(screens) < 1 {
		return nil
	}

	results := make([]platformclientv2.Supportcenterscreen, len(screens))
	for i, value := range screens {
		if screen, ok := value.(map[string]interface{}); ok {
			results[i] = platformclientv2.Supportcenterscreen{
				VarType:        platformclientv2.String(screen["type"].(string)),
				ModuleSettings: buildModuleSettings(screen["module_settings"].([]interface{})),
			}
		}
	}

	return &results
}

func flattenCustomMessages(messages *[]platformclientv2.Supportcentercustommessage) []interface{} {
	if messages == nil || len(*messages) < 1 {
		return nil
	}

	result := make([]interface{}, len(*messages))
	for i, message := range *messages {
		result[i] = map[string]interface{}{
			"default_value": message.DefaultValue,
			"type":          message.VarType,
		}
	}
	return result
}

func flattenScreens(screens *[]platformclientv2.Supportcenterscreen) []interface{} {
	if screens == nil || len(*screens) < 1 {
		return nil
	}

	result := make([]interface{}, len(*screens))
	for i, screen := range *screens {
		result[i] = map[string]interface{}{
			"type":            screen.VarType,
			"module_settings": flattenModuleSettings(screen.ModuleSettings),
		}
	}
	return result
}

func flattenModuleSettings(settings *[]platformclientv2.Supportcentermodulesetting) []interface{} {
	if settings == nil || len(*settings) < 1 {
		return nil
	}

	result := make([]interface{}, len(*settings))
	for i, setting := range *settings {
		settingMap := map[string]interface{}{
			"type":    setting.VarType,
			"enabled": setting.Enabled,
		}

		if setting.CompactCategoryModuleTemplate != nil {
			settingMap["compact_category_module_template_active"] = setting.CompactCategoryModuleTemplate.Active
		}

		if setting.DetailedCategoryModuleTemplate != nil {
			settingMap["detailed_category_module_template"] = []interface{}{
				map[string]interface{}{
					"active":          setting.DetailedCategoryModuleTemplate.Active,
					"sidebar_enabled": setting.DetailedCategoryModuleTemplate.Sidebar.Enabled,
				},
			}
		}

		result[i] = settingMap
	}
	return result
}

func flattenEnabledCategories(categories *[]platformclientv2.Supportcentercategory) []interface{} {
	if categories == nil || len(*categories) < 1 {
		return nil
	}

	result := make([]interface{}, len(*categories))
	for i, category := range *categories {
		categoryMap := map[string]interface{}{
			"category_id": category.Id,
		}

		if category.Image != nil && category.Image.Source != nil && category.Image.Source.DefaultUrl != nil {
			categoryMap["image_uri"] = *category.Image.Source.DefaultUrl
		}

		result[i] = categoryMap
	}
	return result
}

func flattenStyleSettings(style *platformclientv2.Supportcenterstylesetting) []interface{} {
	if style == nil {
		return nil
	}

	return []interface{}{map[string]interface{}{
		"hero_style_setting":   flattenHeroStyle(style.HeroStyle),
		"global_style_setting": flattenGlobalStyle(style.GlobalStyle),
	}}
}

func flattenHeroStyle(style *platformclientv2.Supportcenterherostyle) []interface{} {
	if style == nil {
		return nil
	}

	styleMap := map[string]interface{}{
		"background_color": style.BackgroundColor,
		"text_color":       style.TextColor,
	}

	if style.Image != nil && style.Image.Source != nil && style.Image.Source.DefaultUrl != nil {
		styleMap["image_uri"] = *style.Image.Source.DefaultUrl
	}

	return []interface{}{styleMap}
}

func flattenGlobalStyle(style *platformclientv2.Supportcenterglobalstyle) []interface{} {
	if style == nil {
		return nil
	}

	return []interface{}{map[string]interface{}{
		"background_color":    style.BackgroundColor,
		"primary_color":       style.PrimaryColor,
		"primary_color_dark":  style.PrimaryColorDark,
		"primary_color_light": style.PrimaryColorLight,
		"text_color":          style.TextColor,
		"font_family":         style.FontFamily,
	}}
}

func buildSupportCenterSettings(d *schema.ResourceData) *platformclientv2.Supportcentersettings {
	value, ok := d.GetOk("support_center")
	if !ok {
		return nil
	}

	cfgs := value.([]interface{})
	if len(cfgs) < 1 {
		return nil
	}

	cfg := cfgs[0].(map[string]interface{})
	supportCenterSettings := &platformclientv2.Supportcentersettings{
		Enabled:           platformclientv2.Bool(cfg["enabled"].(bool)),
		EnabledCategories: buildEnabledCategories(cfg["enabled_categories"].([]interface{})),
		StyleSetting:      buildStyleSettings(cfg["style_setting"].([]interface{})),
		Feedback: &platformclientv2.Supportcenterfeedbacksettings{
			Enabled: platformclientv2.Bool(cfg["feedback_enabled"].(bool)),
		},
		RouterType: resourcedata.GetNillableNonZeroValueFromMap[string](cfg, "router_type"),
	}

	if kbId, ok := cfg["knowledge_base_id"].(string); ok && kbId != "" {
		supportCenterSettings.KnowledgeBase = &platformclientv2.Addressableentityref{
			Id: &kbId,
		}
	}

	if customMessages := buildCustomMessages(cfg["custom_messages"].([]interface{})); customMessages != nil {
		supportCenterSettings.CustomMessages = customMessages
	}

	if screens := buildScreens(cfg["screens"].([]interface{})); screens != nil {
		supportCenterSettings.Screens = screens
	}

	return supportCenterSettings
}

func FlattenSupportCenterSettings(supportCenterSettings *platformclientv2.Supportcentersettings) []interface{} {
	if supportCenterSettings == nil {
		return nil
	}

	settingsMap := map[string]interface{}{
		"enabled":            supportCenterSettings.Enabled,
		"custom_messages":    flattenCustomMessages(supportCenterSettings.CustomMessages),
		"router_type":        supportCenterSettings.RouterType,
		"screens":            flattenScreens(supportCenterSettings.Screens),
		"enabled_categories": flattenEnabledCategories(supportCenterSettings.EnabledCategories),
	}

	if supportCenterSettings.KnowledgeBase != nil {
		settingsMap["knowledge_base_id"] = supportCenterSettings.KnowledgeBase.Id
	}

	if supportCenterSettings.Feedback != nil {
		settingsMap["feedback_enabled"] = supportCenterSettings.Feedback.Enabled
	}

	if supportCenterSettings.StyleSetting != nil {
		settingsMap["style_setting"] = flattenStyleSettings(supportCenterSettings.StyleSetting)
	}

	return []interface{}{settingsMap}
}
