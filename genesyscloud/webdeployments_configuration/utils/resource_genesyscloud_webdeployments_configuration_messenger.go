package webdeployments_configuration_utils

import (
	"terraform-provider-genesyscloud/genesyscloud/util/lists"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

func buildAppConversations(conversations []interface{}) *platformclientv2.Conversationappsettings {
	if len(conversations) < 1 || (len(conversations) == 1 && conversations[0] == nil) {
		return nil
	}

	conversation := conversations[0].(map[string]interface{})
	ret := &platformclientv2.Conversationappsettings{
		Enabled:                  platformclientv2.Bool(conversation["enabled"].(bool)),
		ShowAgentTypingIndicator: platformclientv2.Bool(conversation["show_agent_typing_indicator"].(bool)),
		ShowUserTypingIndicator:  platformclientv2.Bool(conversation["show_user_typing_indicator"].(bool)),
		AutoStart: &platformclientv2.Autostart{
			Enabled: platformclientv2.Bool(conversation["auto_start_enabled"].(bool)),
		},
		Markdown: &platformclientv2.Markdown{
			Enabled: platformclientv2.Bool(conversation["markdown_enabled"].(bool)),
		},
		ConversationClear: &platformclientv2.Conversationclearsettings{
			Enabled: platformclientv2.Bool(conversation["conversation_clear_enabled"].(bool)),
		},
	}

	if conversationDisconnectArr, ok := conversation["conversation_disconnect"].([]interface{}); ok && len(conversationDisconnectArr) > 0 && conversationDisconnectArr[0] != nil {
		conversationDisconnect := conversationDisconnectArr[0].(map[string]interface{})
		ret.ConversationDisconnect = &platformclientv2.Conversationdisconnectsettings{
			Enabled: platformclientv2.Bool(conversationDisconnect["enabled"].(bool)),
			VarType: platformclientv2.String(conversationDisconnect["type"].(string)),
		}
	}

	if humanizeArr, ok := conversation["humanize"].([]interface{}); ok && len(humanizeArr) > 0 && humanizeArr[0] != nil {
		humanize := humanizeArr[0].(map[string]interface{})
		ret.Humanize = &platformclientv2.Humanize{
			Enabled: platformclientv2.Bool(humanize["enabled"].(bool)),
		}
		if botArr, ok := humanize["bot"].([]interface{}); ok && len(botArr) > 0 && botArr[0] != nil {
			bot := botArr[0].(map[string]interface{})
			ret.Humanize.Bot = &platformclientv2.Botmessengerprofile{
				Name:      platformclientv2.String(bot["name"].(string)),
				AvatarUrl: platformclientv2.String(bot["avatar_url"].(string)),
			}
		}
	}

	return ret
}

func buildAppKnowledge(knowledge []interface{}) *platformclientv2.Knowledge {
	if len(knowledge) < 1 || (len(knowledge) == 1 && knowledge[0] == nil) {
		return nil
	}

	knowledgeCfg := knowledge[0].(map[string]interface{})
	ret := &platformclientv2.Knowledge{
		Enabled: platformclientv2.Bool(knowledgeCfg["enabled"].(bool)),
	}

	if knowledgeBaseId, ok := knowledgeCfg["knowledge_base_id"].(string); ok {
		ret.KnowledgeBase = &platformclientv2.Addressableentityref{
			Id: &knowledgeBaseId,
		}
	}
	return ret
}

func buildMessengerApps(apps []interface{}) *platformclientv2.Messengerapps {
	if len(apps) < 1 || (len(apps) == 1 && apps[0] == nil) {
		return nil
	}

	messengerApps := platformclientv2.Messengerapps{}
	app := apps[0].(map[string]interface{})

	if conversations, ok := app["conversations"].([]interface{}); ok {
		messengerApps.Conversations = buildAppConversations(conversations)
	}

	if knowledge, ok := app["knowledge"].([]interface{}); ok {
		messengerApps.Knowledge = buildAppKnowledge(knowledge)
	}
	return &messengerApps
}

func buildMessengerSettings(d *schema.ResourceData) *platformclientv2.Messengersettings {
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
		Apps:    buildMessengerApps(cfg["apps"].([]interface{})),
	}

	if styles, ok := cfg["styles"].([]interface{}); ok && len(styles) > 0 && styles[0] != nil {
		style := styles[0].(map[string]interface{})
		if primaryColor, ok := style["primary_color"].(string); ok {
			messengerSettings.Styles = &platformclientv2.Messengerstyles{
				PrimaryColor: &primaryColor,
			}
		}
	}

	if launchers, ok := cfg["launcher_button"].([]interface{}); ok && len(launchers) > 0 && launchers[0] != nil {
		launcher := launchers[0].(map[string]interface{})
		if visibility, ok := launcher["visibility"].(string); ok {
			messengerSettings.LauncherButton = &platformclientv2.Launcherbuttonsettings{
				Visibility: &visibility,
			}
		}
	}

	if screens, ok := cfg["home_screen"].([]interface{}); ok && len(screens) > 0 && screens[0] != nil {
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

	if fileUploads, ok := cfg["file_upload"].([]interface{}); ok && len(fileUploads) > 0 && fileUploads[0] != nil {
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

			messengerSettings.FileUpload = &platformclientv2.Fileuploadsettings{}

			if len(modes) > 0 {
				messengerSettings.FileUpload.Modes = &modes
			}
		}

	}

	return messengerSettings
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

	ret := map[string]interface{}{
		"mode": modes,
	}

	return []interface{}{ret}
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
		retMap["conversation_disconnect"] = []interface{}{map[string]interface{}{
			"enabled": conversations.ConversationDisconnect.Enabled,
			"type":    conversations.ConversationDisconnect.VarType,
		}}
	}

	if conversations.ConversationClear != nil {
		retMap["conversation_clear_enabled"] = conversations.ConversationClear.Enabled
	}

	if conversations.Humanize != nil {
		if conversations.Humanize.Bot != nil {
			retMap["humanize"] = []interface{}{map[string]interface{}{
				"enabled": conversations.Humanize.Enabled,
				"bot": []interface{}{map[string]interface{}{
					"name":       conversations.Humanize.Bot.Name,
					"avatar_url": conversations.Humanize.Bot.AvatarUrl,
				}},
			}}
		} else {
			retMap["humanize"] = []interface{}{map[string]interface{}{
				"enabled": conversations.Humanize.Enabled,
			},
			}
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
