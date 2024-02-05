package webdeployments_configuration_utils

import (
	"terraform-provider-genesyscloud/genesyscloud/util/lists"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v119/platformclientv2"
)

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
		"enable_attachments":            settings.EnableAttachments,
		"use_supported_content_profile": settings.UseSupportedContentProfile,
		"mode":                          modes,
	}}
}

func flattenAppConversations(conversations *platformclientv2.Conversationappsettings) []interface{} {
	if conversations == nil {
		return nil
	}

	retMap := map[string]interface{}{
		"enabled":                     conversations.Enabled,
		"show_agent_typing_indicator": conversations.ShowAgentTypingIndicator,
		"show_user_typing_indicator":  conversations.ShowUserTypingIndicator,
	}

	if conversations.AutoStart != nil {
		retMap["auto_start_enabled"] = conversations.AutoStart.Enabled
	}

	if conversations.Markdown != nil {
		retMap["markdown_enabled"] = conversations.Markdown.Enabled
	}

	if conversations.ConversationDisconnect != nil {
		retMap["conversation_disconnect"] = map[string]interface{}{
			"enabled": conversations.ConversationDisconnect.Enabled,
			"type":    conversations.ConversationDisconnect.VarType,
		}
	}

	if conversations.ConversationClear != nil {
		retMap["conversation_clear_enabled"] = conversations.ConversationClear.Enabled
	}

	if conversations.Humanize != nil {
		retMap["humanize"] = map[string]interface{}{
			"enabled": conversations.Humanize.Enabled,
			"bot": []interface{}{map[string]interface{}{
				"name":       conversations.Humanize.Bot.Name,
				"avatar_url": conversations.Humanize.Bot.AvatarUrl,
			}},
		}
	}

	return []interface{}{retMap}
}

func flattenAppKnowledge(knowledge *platformclientv2.Knowledge) []interface{} {
	if knowledge == nil {
		return nil
	}

	retMap := map[string]interface{}{
		"enabled": knowledge.Enabled,
	}

	if knowledge.KnowledgeBase != nil {
		retMap["knowledge_base_id"] = knowledge.KnowledgeBase.Id
	}

	return []interface{}{retMap}
}

func flattenMessengerApps(apps *platformclientv2.Messengerapps) []interface{} {
	if apps == nil {
		return nil
	}

	return []interface{}{map[string]interface{}{
		"conversations": flattenAppConversations(apps.Conversations),
		"knowledge":     flattenAppKnowledge(apps.Knowledge),
	}}
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

func FlattenMessengerSettings(messengerSettings *platformclientv2.Messengersettings) []interface{} {
	if messengerSettings == nil {
		return nil
	}

	return []interface{}{map[string]interface{}{
		"enabled":         messengerSettings.Enabled,
		"styles":          flattenStyles(messengerSettings.Styles),
		"launcher_button": flattenLauncherButton(messengerSettings.LauncherButton),
		"home_screen":     flattenHomeScreen(messengerSettings.HomeScreen),
		"file_upload":     flattenFileUpload(messengerSettings.FileUpload),
		"apps":            flattenMessengerApps(messengerSettings.Apps),
	}}
}
