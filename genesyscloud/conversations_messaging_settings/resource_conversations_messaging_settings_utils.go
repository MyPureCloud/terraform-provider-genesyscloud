package conversations_messaging_settings

import (
	"fmt"
	"strings"
	"terraform-provider-genesyscloud/genesyscloud/util/resourcedata"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

func getConversationsMessagingSettingsFromResourceData(d *schema.ResourceData) platformclientv2.Messagingsettingrequest {
	return platformclientv2.Messagingsettingrequest{
		Name:    platformclientv2.String(d.Get("name").(string)),
		Content: buildContentSettings(d.Get("content").([]interface{})),
		Event:   buildEventSetting(d.Get("event").([]interface{})),
	}
}

func buildContentSettings(contentSettings []interface{}) *platformclientv2.Contentsetting {
	var sdkContentSetting platformclientv2.Contentsetting

	for _, contentSetting := range contentSettings {
		contentSettingsMap, ok := contentSetting.(map[string]interface{})
		if !ok {
			continue
		}

		resourcedata.BuildSDKInterfaceArrayValueIfNotNil(&sdkContentSetting.Story, contentSettingsMap, "story", buildStorySettings)
	}
	return &sdkContentSetting
}

func buildStorySettings(storySettings []interface{}) *platformclientv2.Storysetting {
	var sdkStorySetting platformclientv2.Storysetting

	for _, storySetting := range storySettings {
		storySettingsMap, ok := storySetting.(map[string]interface{})
		if !ok {
			continue
		}

		resourcedata.BuildSDKInterfaceArrayValueIfNotNil(&sdkStorySetting.Mention, storySettingsMap, "mention", buildInboundOnlySettings)
		resourcedata.BuildSDKInterfaceArrayValueIfNotNil(&sdkStorySetting.Reply, storySettingsMap, "reply", buildInboundOnlySettings)
	}

	return &sdkStorySetting
}

func buildInboundOnlySettings(inboundOnlySettings []interface{}) *platformclientv2.Inboundonlysetting {
	var sdkInboundOnlySetting platformclientv2.Inboundonlysetting

	for _, inboundOnlySetting := range inboundOnlySettings {
		inboundOnlySettingsMap, ok := inboundOnlySetting.(map[string]interface{})
		if !ok {
			continue
		}
		resourcedata.BuildSDKStringValueIfNotNil(&sdkInboundOnlySetting.Inbound, inboundOnlySettingsMap, "inbound")
	}

	return &sdkInboundOnlySetting
}

func buildEventSetting(eventSettings []interface{}) *platformclientv2.Eventsetting {
	var sdkEventSetting platformclientv2.Eventsetting

	for _, eventSetting := range eventSettings {
		eventSettingsMap, ok := eventSetting.(map[string]interface{})
		if !ok {
			continue
		}
		resourcedata.BuildSDKInterfaceArrayValueIfNotNil(&sdkEventSetting.Typing, eventSettingsMap, "typing", buildTypingSettings)
	}

	return &sdkEventSetting
}

func buildTypingSettings(typingSettings []interface{}) *platformclientv2.Typingsetting {
	var sdkTypingSetting platformclientv2.Typingsetting

	for _, typingSetting := range typingSettings {
		typingSettingsMap, ok := typingSetting.(map[string]interface{})
		if !ok {
			continue
		}

		resourcedata.BuildSDKInterfaceArrayValueIfNotNil(&sdkTypingSetting.On, typingSettingsMap, "on", buildSettingDirections)
	}

	return &sdkTypingSetting
}

func buildSettingDirections(settingDirections []interface{}) *platformclientv2.Settingdirection {
	var sdkSettingDirection platformclientv2.Settingdirection

	for _, settingDirection := range settingDirections {
		settingDirectionsMap, ok := settingDirection.(map[string]interface{})
		if !ok {
			continue
		}

		resourcedata.BuildSDKStringValueIfNotNil(&sdkSettingDirection.Inbound, settingDirectionsMap, "inbound")
		resourcedata.BuildSDKStringValueIfNotNil(&sdkSettingDirection.Outbound, settingDirectionsMap, "outbound")
	}

	return &sdkSettingDirection
}

// flattenInboundOnlySettings maps a Genesys Cloud *[]platformclientv2.Inboundonlysetting into a []interface{}
func flattenInboundOnlySettings(inboundOnlySettings *platformclientv2.Inboundonlysetting) []interface{} {
	var inboundOnlySettingList []interface{}
	inboundOnlySettingMap := make(map[string]interface{})

	resourcedata.SetMapValueIfNotNil(inboundOnlySettingMap, "inbound", inboundOnlySettings.Inbound)

	inboundOnlySettingList = append(inboundOnlySettingList, inboundOnlySettingMap)

	return inboundOnlySettingList
}

// flattenStorySettings maps a Genesys Cloud *[]platformclientv2.Storysetting into a []interface{}
func flattenStorySettings(storySettings *platformclientv2.Storysetting) []interface{} {
	var storySettingList []interface{}
	storySettingMap := make(map[string]interface{})

	resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(storySettingMap, "mention", storySettings.Mention, flattenInboundOnlySettings)
	resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(storySettingMap, "reply", storySettings.Reply, flattenInboundOnlySettings)

	storySettingList = append(storySettingList, storySettingMap)

	return storySettingList
}

// flattenContentSettings maps a Genesys Cloud *[]platformclientv2.Contentsetting into a []interface{}
func flattenContentSettings(contentSettings *platformclientv2.Contentsetting) []interface{} {
	var contentSettingList []interface{}
	contentSettingMap := make(map[string]interface{})

	resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(contentSettingMap, "story", contentSettings.Story, flattenStorySettings)

	contentSettingList = append(contentSettingList, contentSettingMap)

	return contentSettingList
}

// flattenSettingDirections maps a Genesys Cloud *[]platformclientv2.Settingdirection into a []interface{}
func flattenSettingDirections(settingDirections *platformclientv2.Settingdirection) []interface{} {
	var settingDirectionList []interface{}
	settingDirectionMap := make(map[string]interface{})

	resourcedata.SetMapValueIfNotNil(settingDirectionMap, "inbound", settingDirections.Inbound)
	resourcedata.SetMapValueIfNotNil(settingDirectionMap, "outbound", settingDirections.Outbound)

	settingDirectionList = append(settingDirectionList, settingDirectionMap)

	return settingDirectionList
}

// flattenTypingSettings maps a Genesys Cloud *[]platformclientv2.Typingsetting into a []interface{}
func flattenTypingSettings(typingSettings *platformclientv2.Typingsetting) []interface{} {
	var typingSettingList []interface{}
	typingSettingMap := make(map[string]interface{})

	resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(typingSettingMap, "on", typingSettings.On, flattenSettingDirections)

	typingSettingList = append(typingSettingList, typingSettingMap)

	return typingSettingList
}

// flattenEventSettings maps a Genesys Cloud *[]platformclientv2.Eventsetting into a []interface{}
func flattenEventSettings(eventSettings *platformclientv2.Eventsetting) []interface{} {
	var eventSettingList []interface{}
	eventSettingMap := make(map[string]interface{})

	resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(eventSettingMap, "typing", eventSettings.Typing, flattenTypingSettings)

	eventSettingList = append(eventSettingList, eventSettingMap)

	return eventSettingList
}

func generateConversationsMessagingSettingsResource(resourceID string, name string, nestedBlocks ...string) string {
	return fmt.Sprintf(`resource "genesyscloud_conversations_messaging_settings" "%s" {
		name = "%s"
		%s
	}
	`, resourceID, name, strings.Join(nestedBlocks, "\n"))
}

func generateTypingOnSetting(inbound, outbound string) string {
	return fmt.Sprintf(`
	event {
		typing {
			on {
				inbound = "%s"
				outbound = "%s"
			}
		}
	}`, inbound, outbound)
}

func generateContentStoryBlock(nestedBlocks ...string) string {
	return fmt.Sprintf(`
	content {
		story {
			%s
		}
	}`, strings.Join(nestedBlocks, "\n"))
}

func generateMentionInboundOnlySetting(value string) string {
	return fmt.Sprintf(`
		mention {
			inbound = "%s"
		}
	`, value)
}

func generateReplyInboundOnlySetting(value string) string {
	return fmt.Sprintf(`
	reply {
		inbound = "%s"
	}`, value)
}
