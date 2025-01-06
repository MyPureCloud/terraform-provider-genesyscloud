package routing_queue

import (
	"context"
	"fmt"
	"os"
	"strings"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"terraform-provider-genesyscloud/genesyscloud/util/lists"
	"terraform-provider-genesyscloud/genesyscloud/util/resourcedata"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v150/platformclientv2"
)

// Build Functions

func buildSdkMediaSettings(d *schema.ResourceData) *platformclientv2.Queuemediasettings {
	queueMediaSettings := &platformclientv2.Queuemediasettings{}

	mediaSettingsCall := d.Get("media_settings_call").([]interface{})
	if len(mediaSettingsCall) > 0 {
		queueMediaSettings.Call = buildSdkMediaSetting(mediaSettingsCall)
	}

	mediaSettingsCallback := d.Get("media_settings_callback").([]interface{})
	if len(mediaSettingsCallback) > 0 {
		queueMediaSettings.Callback = buildSdkMediaSettingCallback(mediaSettingsCallback)
	}

	mediaSettingsChat := d.Get("media_settings_chat").([]interface{})
	if len(mediaSettingsChat) > 0 {
		queueMediaSettings.Chat = buildSdkMediaSetting(mediaSettingsChat)
	}

	mediaSettingsEmail := d.Get("media_settings_email").([]interface{})
	if len(mediaSettingsEmail) > 0 {
		queueMediaSettings.Email = buildSdkMediaEmailSetting(mediaSettingsEmail)
	}

	mediaSettingsMessage := d.Get("media_settings_message").([]interface{})
	if len(mediaSettingsMessage) > 0 {
		queueMediaSettings.Message = buildSdkMediaSetting(mediaSettingsMessage)
	}

	return queueMediaSettings
}

func buildSdkAcwSettings(d *schema.ResourceData) *platformclientv2.Acwsettings {
	acwWrapupPrompt := d.Get("acw_wrapup_prompt").(string)

	acwSettings := platformclientv2.Acwsettings{
		WrapupPrompt: &acwWrapupPrompt, // Set or default
	}

	// Only set timeout for certain wrapup prompt types
	if acwWrapupPrompt == "MANDATORY_TIMEOUT" || acwWrapupPrompt == "MANDATORY_FORCED_TIMEOUT" || acwWrapupPrompt == "AGENT_REQUESTED" {
		acwTimeoutMs, hasTimeout := d.GetOk("acw_timeout_ms")
		if hasTimeout {
			timeout := acwTimeoutMs.(int)
			acwSettings.TimeoutMs = &timeout
		}
	}
	return &acwSettings
}

func buildSdkDefaultScriptsMap(d *schema.ResourceData) *map[string]platformclientv2.Script {
	if scriptIds, ok := d.GetOk("default_script_ids"); ok {
		scriptMap := scriptIds.(map[string]interface{})

		results := make(map[string]platformclientv2.Script)
		for k, v := range scriptMap {
			scriptID := v.(string)
			results[k] = platformclientv2.Script{Id: &scriptID}
		}
		return &results
	}
	return nil
}

func buildSdkDirectRouting(d *schema.ResourceData) *platformclientv2.Directrouting {
	directRouting := d.Get("direct_routing").([]interface{})
	if len(directRouting) > 0 {
		settingsMap := directRouting[0].(map[string]interface{})

		agentWaitSeconds := settingsMap["agent_wait_seconds"].(int)
		waitForAgent := settingsMap["wait_for_agent"].(bool)

		callUseAgentAddressOutbound := settingsMap["call_use_agent_address_outbound"].(bool)
		callSettings := &platformclientv2.Directroutingmediasettings{
			UseAgentAddressOutbound: &callUseAgentAddressOutbound,
		}

		emailUseAgentAddressOutbound := settingsMap["email_use_agent_address_outbound"].(bool)
		emailSettings := &platformclientv2.Directroutingmediasettings{
			UseAgentAddressOutbound: &emailUseAgentAddressOutbound,
		}

		messageUseAgentAddressOutbound := settingsMap["message_use_agent_address_outbound"].(bool)
		messageSettings := &platformclientv2.Directroutingmediasettings{
			UseAgentAddressOutbound: &messageUseAgentAddressOutbound,
		}

		sdkDirectRouting := &platformclientv2.Directrouting{
			CallMediaSettings:    callSettings,
			EmailMediaSettings:   emailSettings,
			MessageMediaSettings: messageSettings,
			WaitForAgent:         &waitForAgent,
			AgentWaitSeconds:     &agentWaitSeconds,
		}

		if backUpQueueId, ok := settingsMap["backup_queue_id"].(string); ok && backUpQueueId != "" {
			sdkDirectRouting.BackupQueueId = &backUpQueueId
		}

		return sdkDirectRouting
	}
	return nil
}

func buildAgentOwnedRouting(routing []interface{}) *platformclientv2.Agentownedrouting {
	settingsMap := routing[0].(map[string]interface{})
	return &platformclientv2.Agentownedrouting{
		EnableAgentOwnedCallbacks:  platformclientv2.Bool(settingsMap["enable_agent_owned_callbacks"].(bool)),
		MaxOwnedCallbackDelayHours: platformclientv2.Int(settingsMap["max_owned_callback_delay_hours"].(int)),
		MaxOwnedCallbackHours:      platformclientv2.Int(settingsMap["max_owned_callback_hours"].(int)),
	}
}

func buildSdkMediaEmailSetting(settings []interface{}) *platformclientv2.Emailmediasettings {
	settingsMap := settings[0].(map[string]interface{})

	return &platformclientv2.Emailmediasettings{
		AlertingTimeoutSeconds: platformclientv2.Int(settingsMap["alerting_timeout_sec"].(int)),
		EnableAutoAnswer:       platformclientv2.Bool(settingsMap["enable_auto_answer"].(bool)),
		ServiceLevel: &platformclientv2.Servicelevel{
			Percentage: platformclientv2.Float64(settingsMap["service_level_percentage"].(float64)),
			DurationMs: platformclientv2.Int(settingsMap["service_level_duration_ms"].(int)),
		},
	}
}

func buildSdkMediaSetting(settings []interface{}) *platformclientv2.Mediasettings {
	settingsMap := settings[0].(map[string]interface{})

	return &platformclientv2.Mediasettings{
		AlertingTimeoutSeconds: platformclientv2.Int(settingsMap["alerting_timeout_sec"].(int)),
		EnableAutoAnswer:       platformclientv2.Bool(settingsMap["enable_auto_answer"].(bool)),
		ServiceLevel: &platformclientv2.Servicelevel{
			Percentage: platformclientv2.Float64(settingsMap["service_level_percentage"].(float64)),
			DurationMs: platformclientv2.Int(settingsMap["service_level_duration_ms"].(int)),
		},
		SubTypeSettings: buildSubTypeSettings(settingsMap["sub_type_settings"].([]interface{})),
	}
}

func buildSdkMediaSettingCallback(settings []interface{}) *platformclientv2.Callbackmediasettings {
	settingsMap := settings[0].(map[string]interface{})

	return &platformclientv2.Callbackmediasettings{
		AlertingTimeoutSeconds: platformclientv2.Int(settingsMap["alerting_timeout_sec"].(int)),
		ServiceLevel: &platformclientv2.Servicelevel{
			Percentage: platformclientv2.Float64(settingsMap["service_level_percentage"].(float64)),
			DurationMs: platformclientv2.Int(settingsMap["service_level_duration_ms"].(int)),
		},
		EnableAutoAnswer:     platformclientv2.Bool(settingsMap["enable_auto_answer"].(bool)),
		AutoEndDelaySeconds:  platformclientv2.Int(settingsMap["auto_end_delay_seconds"].(int)),
		AutoDialDelaySeconds: platformclientv2.Int(settingsMap["auto_dial_delay_seconds"].(int)),
		EnableAutoDialAndEnd: platformclientv2.Bool(settingsMap["enable_auto_dial_and_end"].(bool)),
	}
}

func buildSubTypeSettings(subTypeList []interface{}) *map[string]platformclientv2.Basemediasettings {

	returnObj := make(map[string]platformclientv2.Basemediasettings)

	for _, subTypeItem := range subTypeList {
		if subTypeItem == nil {
			continue
		}
		subTypeMap := subTypeItem.(map[string]interface{})
		mediaType := subTypeMap["media_type"].(string)
		enableAutoAnswer := subTypeMap["enable_auto_answer"].(bool)
		baseMediaSettings := platformclientv2.Basemediasettings{
			EnableAutoAnswer: &enableAutoAnswer,
		}
		returnObj[mediaType] = baseMediaSettings
	}

	if len(returnObj) > 0 {
		return &returnObj
	}
	return nil

}

func buildCannedResponseLibraries(d *schema.ResourceData) *platformclientv2.Cannedresponselibraries {
	var cannedResponseSdk platformclientv2.Cannedresponselibraries
	cannedResponseList := d.Get("canned_response_libraries").([]interface{})
	if len(cannedResponseList) > 0 {
		cannedResponseMap := cannedResponseList[0].(map[string]interface{})
		resourcedata.BuildSDKStringValueIfNotNil(&cannedResponseSdk.Mode, cannedResponseMap, "mode")
		if libraryIds, exists := cannedResponseMap["library_ids"].([]interface{}); exists {
			libraryIdList := lists.InterfaceListToStrings(libraryIds)
			cannedResponseSdk.LibraryIds = &libraryIdList
		}
		return &cannedResponseSdk

	}
	return nil
}

func buildSdkRoutingRules(d *schema.ResourceData) *[]platformclientv2.Routingrule {
	var routingRules []platformclientv2.Routingrule
	if configRoutingRules, ok := d.GetOk("routing_rules"); ok {
		for _, configRule := range configRoutingRules.([]interface{}) {
			ruleSettings, ok := configRule.(map[string]interface{})
			if !ok {
				continue
			}
			var sdkRule platformclientv2.Routingrule

			resourcedata.BuildSDKStringValueIfNotNil(&sdkRule.Operator, ruleSettings, "operator")
			if threshold, ok := ruleSettings["threshold"]; ok {
				v := threshold.(int)
				sdkRule.Threshold = &v
			}
			if waitSeconds, ok := ruleSettings["wait_seconds"].(float64); ok {
				sdkRule.WaitSeconds = &waitSeconds
			}

			routingRules = append(routingRules, sdkRule)
		}
	}
	return &routingRules
}

func buildSdkConditionalGroupRouting(d *schema.ResourceData) (*platformclientv2.Conditionalgrouprouting, diag.Diagnostics) {
	cgrRules, ok := d.Get("conditional_group_routing_rules").([]interface{})
	if !ok || len(cgrRules) == 0 {
		return nil, nil
	}

	var sdkCGRRules []platformclientv2.Conditionalgrouproutingrule

	for i, rule := range cgrRules {
		ruleSettings, ok := rule.(map[string]interface{})
		if !ok {
			continue
		}
		var sdkCGRRule platformclientv2.Conditionalgrouproutingrule

		if waitSeconds, ok := ruleSettings["wait_seconds"].(int); ok {
			sdkCGRRule.WaitSeconds = &waitSeconds
		}

		resourcedata.BuildSDKStringValueIfNotNil(&sdkCGRRule.Operator, ruleSettings, "operator")
		resourcedata.BuildSDKStringValueIfNotNil(&sdkCGRRule.Metric, ruleSettings, "metric")

		if conditionValue, ok := ruleSettings["condition_value"].(float64); ok {
			sdkCGRRule.ConditionValue = &conditionValue
		}

		if queueId, ok := ruleSettings["queue_id"].(string); ok && queueId != "" {
			if i == 0 {
				return nil, util.BuildDiagnosticError(ResourceType, "For rule 1, queue_id is always assumed to be the current queue, so queue id should not be specified", fmt.Errorf("queue id is not nil"))
			}
			sdkCGRRule.Queue = &platformclientv2.Domainentityref{Id: &queueId}
		}

		if memberGroupSet, ok := ruleSettings["groups"].(*schema.Set); ok {
			sdkCGRRule.Groups = buildCGRGroups(memberGroupSet)
		}
		sdkCGRRules = append(sdkCGRRules, sdkCGRRule)
	}

	return &platformclientv2.Conditionalgrouprouting{Rules: &sdkCGRRules}, nil
}

func buildCGRGroups(groups *schema.Set) *[]platformclientv2.Membergroup {
	groupList := groups.List()
	if len(groupList) == 0 {
		return nil
	}

	sdkMemberGroups := make([]platformclientv2.Membergroup, 0)
	for _, group := range groupList {
		groupMap, ok := group.(map[string]interface{})
		if !ok {
			continue
		}

		sdkGroup := platformclientv2.Membergroup{
			Id:      platformclientv2.String(groupMap["member_group_id"].(string)),
			VarType: platformclientv2.String(groupMap["member_group_type"].(string)),
		}

		sdkMemberGroups = append(sdkMemberGroups, sdkGroup)
	}

	return &sdkMemberGroups
}

func buildMemberGroupList(d *schema.ResourceData, groupKey string, groupType string) *[]platformclientv2.Membergroup {
	var memberGroups []platformclientv2.Membergroup
	if mg, ok := d.GetOk(groupKey); ok {

		for _, mgId := range mg.(*schema.Set).List() {
			id := mgId.(string)

			memberGroup := &platformclientv2.Membergroup{Id: &id, VarType: &groupType}
			memberGroups = append(memberGroups, *memberGroup)
		}
	}

	return &memberGroups
}

func buildSdkBullseyeSettings(d *schema.ResourceData) *platformclientv2.Bullseye {
	if configRings, ok := d.GetOk("bullseye_rings"); ok {
		var sdkRings []platformclientv2.Ring

		for _, configRing := range configRings.([]interface{}) {
			ringSettings, ok := configRing.(map[string]interface{})
			if !ok {
				continue
			}
			var sdkRing platformclientv2.Ring

			if waitSeconds, ok := ringSettings["expansion_timeout_seconds"].(float64); ok {
				sdkRing.ExpansionCriteria = &[]platformclientv2.Expansioncriterium{
					{
						VarType:   &bullseyeExpansionTypeTimeout,
						Threshold: &waitSeconds,
					},
				}
			}

			if skillsToRemove, ok := ringSettings["skills_to_remove"]; ok {
				skillIds := skillsToRemove.(*schema.Set).List()
				if len(skillIds) > 0 {
					sdkSkillsToRemove := make([]platformclientv2.Skillstoremove, len(skillIds))
					for i, id := range skillIds {
						skillID := id.(string)
						sdkSkillsToRemove[i] = platformclientv2.Skillstoremove{
							Id: &skillID,
						}
					}
					sdkRing.Actions = &platformclientv2.Actions{
						SkillsToRemove: &sdkSkillsToRemove,
					}
				}
			}

			if memberGroups, ok := ringSettings["member_groups"]; ok {
				memberGroupList := memberGroups.(*schema.Set).List()
				if len(memberGroupList) > 0 {

					sdkMemberGroups := make([]platformclientv2.Membergroup, len(memberGroupList))
					for i, memberGroup := range memberGroupList {
						settingsMap := memberGroup.(map[string]interface{})
						memberGroupID := settingsMap["member_group_id"].(string)
						memberGroupType := settingsMap["member_group_type"].(string)

						sdkMemberGroups[i] = platformclientv2.Membergroup{
							Id:      &memberGroupID,
							VarType: &memberGroupType,
						}
					}
					sdkRing.MemberGroups = &sdkMemberGroups
				}
			}
			sdkRings = append(sdkRings, sdkRing)
		}

		/*
			The routing queues API is a little unusual.  You can have up to six bullseye routing rings but the last one is always
			a treated as the default ring.  This means you can actually ony define a maximum of 5.  So, I have changed the behavior of this
			resource to only allow you to add 5 items and then the code always adds a 6 item (see the code below) with a default timeout of 2.
		*/
		var defaultSdkRing platformclientv2.Ring
		defaultTimeoutInt := 2
		defaultTimeoutFloat := float64(defaultTimeoutInt)
		defaultSdkRing.ExpansionCriteria = &[]platformclientv2.Expansioncriterium{
			{
				VarType:   &bullseyeExpansionTypeTimeout,
				Threshold: &defaultTimeoutFloat,
			},
		}

		sdkRings = append(sdkRings, defaultSdkRing)
		return &platformclientv2.Bullseye{Rings: &sdkRings}
	}
	return nil
}

func buildSdkQueueMessagingAddresses(d *schema.ResourceData) *platformclientv2.Queuemessagingaddresses {
	var messagingAddresses platformclientv2.Queuemessagingaddresses

	if _, ok := d.GetOk("outbound_messaging_sms_address_id"); ok {
		messagingAddresses.SmsAddress = util.BuildSdkDomainEntityRef(d, "outbound_messaging_sms_address_id")
	}

	if _, ok := d.GetOk("outbound_messaging_whatsapp_recipient_id"); ok {
		messagingAddresses.WhatsAppRecipient = util.BuildSdkDomainEntityRef(d, "outbound_messaging_whatsapp_recipient_id")

	}
	if _, ok := d.GetOk("outbound_messaging_open_messaging_recipient_id"); ok {
		messagingAddresses.OpenMessagingRecipient = util.BuildSdkDomainEntityRef(d, "outbound_messaging_open_messaging_recipient_id")
	}

	return &messagingAddresses
}

func buildSdkQueueEmailAddress(d *schema.ResourceData) *platformclientv2.Queueemailaddress {
	outboundEmailAddress := d.Get("outbound_email_address").([]interface{})
	if len(outboundEmailAddress) > 0 {
		settingsMap := outboundEmailAddress[0].(map[string]interface{})

		inboundRoute := &platformclientv2.Inboundroute{
			Id: platformclientv2.String(settingsMap["route_id"].(string)),
		}
		return &platformclientv2.Queueemailaddress{
			Domain: &platformclientv2.Domainentityref{Id: platformclientv2.String(settingsMap["domain_id"].(string))},
			Route:  &inboundRoute,
		}
	}
	return nil
}

func constructAgentOwnedRouting(d *schema.ResourceData) *platformclientv2.Agentownedrouting {
	if agentOwnedRouting, ok := d.Get("agent_owned_routing").([]interface{}); ok {
		if len(agentOwnedRouting) > 0 {
			return buildAgentOwnedRouting(agentOwnedRouting)
		}
	}
	return nil
}

// Flatten Functions

func flattenMediaEmailSetting(settings *platformclientv2.Emailmediasettings) []interface{} {
	settingsMap := make(map[string]interface{})

	settingsMap["alerting_timeout_sec"] = *settings.AlertingTimeoutSeconds
	resourcedata.SetMapValueIfNotNil(settingsMap, "enable_auto_answer", settings.EnableAutoAnswer)
	settingsMap["service_level_percentage"] = *settings.ServiceLevel.Percentage
	settingsMap["service_level_duration_ms"] = *settings.ServiceLevel.DurationMs
	return []interface{}{settingsMap}
}

func flattenMediaSetting(settings *platformclientv2.Mediasettings) []interface{} {
	settingsMap := make(map[string]interface{})

	settingsMap["alerting_timeout_sec"] = *settings.AlertingTimeoutSeconds
	resourcedata.SetMapValueIfNotNil(settingsMap, "enable_auto_answer", settings.EnableAutoAnswer)
	settingsMap["service_level_percentage"] = *settings.ServiceLevel.Percentage
	settingsMap["service_level_duration_ms"] = *settings.ServiceLevel.DurationMs
	if settings.SubTypeSettings != nil {
		settingsMap["sub_type_settings"] = flattenSubTypeSettings(*settings.SubTypeSettings)
	}
	return []interface{}{settingsMap}
}

func flattenSubTypeSettings(subType map[string]platformclientv2.Basemediasettings) []interface{} {
	if subType == nil {
		return nil
	}
	subTypeList := make([]interface{}, 0)
	for key, value := range subType {
		subTypeMap := make(map[string]interface{})
		resourcedata.SetMapValueIfNotNil(subTypeMap, "media_type", &key)
		resourcedata.SetMapValueIfNotNil(subTypeMap, "enable_auto_answer", value.EnableAutoAnswer)
		subTypeList = append(subTypeList, subTypeMap)
	}
	return subTypeList
}

func flattenCannedResponse(cannedResponse *platformclientv2.Cannedresponselibraries) []interface{} {
	if cannedResponse == nil {
		return nil
	}
	cannedResponseList := make([]interface{}, 0)
	cannedResponseMap := make(map[string]interface{})
	resourcedata.SetMapValueIfNotNil(cannedResponseMap, "mode", cannedResponse.Mode)
	if cannedResponse.LibraryIds != nil {
		cannedResponseMap["library_ids"] = lists.StringListToInterfaceList(*cannedResponse.LibraryIds)
	}
	cannedResponseList = append(cannedResponseList, cannedResponseMap)

	return cannedResponseList
}

func flattenDefaultScripts(sdkScripts map[string]platformclientv2.Script) map[string]interface{} {
	if len(sdkScripts) == 0 {
		return nil
	}

	results := make(map[string]interface{})
	for k, v := range sdkScripts {
		results[k] = *v.Id
	}
	return results
}

func flattenDirectRouting(settings *platformclientv2.Directrouting) []interface{} {
	settingsMap := make(map[string]interface{})

	if settings.BackupQueueId != nil {
		settingsMap["backup_queue_id"] = *settings.BackupQueueId
	}
	if settings.AgentWaitSeconds != nil {
		settingsMap["agent_wait_seconds"] = *settings.AgentWaitSeconds
	}
	if settings.WaitForAgent != nil {
		settingsMap["wait_for_agent"] = *settings.WaitForAgent
	}

	if settings.CallMediaSettings != nil {
		callSettings := *settings.CallMediaSettings
		settingsMap["call_use_agent_address_outbound"] = *callSettings.UseAgentAddressOutbound
	}
	if settings.EmailMediaSettings != nil {
		emailSettings := *settings.EmailMediaSettings
		settingsMap["email_use_agent_address_outbound"] = *emailSettings.UseAgentAddressOutbound
	}
	if settings.MessageMediaSettings != nil {
		messageSettings := *settings.MessageMediaSettings
		settingsMap["message_use_agent_address_outbound"] = *messageSettings.UseAgentAddressOutbound
	}

	return []interface{}{settingsMap}
}

func flattenAgentOwnedRouting(settings *platformclientv2.Agentownedrouting) []interface{} {
	settingsMap := make(map[string]interface{})

	settingsMap["max_owned_callback_delay_hours"] = *settings.MaxOwnedCallbackDelayHours
	resourcedata.SetMapValueIfNotNil(settingsMap, "enable_agent_owned_callbacks", settings.EnableAgentOwnedCallbacks)
	settingsMap["max_owned_callback_hours"] = *settings.MaxOwnedCallbackHours

	return []interface{}{settingsMap}
}

func flattenMediaSettingCallback(settings *platformclientv2.Callbackmediasettings) []interface{} {
	settingsMap := make(map[string]interface{})

	settingsMap["alerting_timeout_sec"] = *settings.AlertingTimeoutSeconds
	settingsMap["service_level_percentage"] = *settings.ServiceLevel.Percentage
	settingsMap["service_level_duration_ms"] = *settings.ServiceLevel.DurationMs
	resourcedata.SetMapValueIfNotNil(settingsMap, "enable_auto_answer", settings.EnableAutoAnswer)
	resourcedata.SetMapValueIfNotNil(settingsMap, "enable_auto_dial_and_end", settings.EnableAutoDialAndEnd)
	settingsMap["auto_end_delay_seconds"] = *settings.AutoEndDelaySeconds
	settingsMap["auto_dial_delay_seconds"] = *settings.AutoDialDelaySeconds
	return []interface{}{settingsMap}
}

func flattenRoutingRules(sdkRoutingRules *[]platformclientv2.Routingrule) []interface{} {
	rules := make([]interface{}, len(*sdkRoutingRules))
	for i, sdkRule := range *sdkRoutingRules {
		ruleSettings := make(map[string]interface{})

		resourcedata.SetMapValueIfNotNil(ruleSettings, "operator", sdkRule.Operator)
		resourcedata.SetMapValueIfNotNil(ruleSettings, "threshold", sdkRule.Threshold)
		resourcedata.SetMapValueIfNotNil(ruleSettings, "wait_seconds", sdkRule.WaitSeconds)

		rules[i] = ruleSettings
	}
	return rules
}

func flattenConditionalGroupRoutingRules(queue *platformclientv2.Queue) []interface{} {
	if queue.ConditionalGroupRouting == nil || len(*queue.ConditionalGroupRouting.Rules) == 0 {
		return nil
	}

	rules := make([]interface{}, len(*queue.ConditionalGroupRouting.Rules))
	for i, rule := range *queue.ConditionalGroupRouting.Rules {
		ruleSettings := make(map[string]interface{})

		resourcedata.SetMapValueIfNotNil(ruleSettings, "wait_seconds", rule.WaitSeconds)
		resourcedata.SetMapValueIfNotNil(ruleSettings, "operator", rule.Operator)
		resourcedata.SetMapValueIfNotNil(ruleSettings, "condition_value", rule.ConditionValue)
		resourcedata.SetMapValueIfNotNil(ruleSettings, "metric", rule.Metric)

		// The first rule is assumed to apply to this queue, so queue_id should be omitted if the conditional grouping routing rule
		//is the first one being looked at.
		if rule.Queue != nil && i > 0 {
			ruleSettings["queue_id"] = *rule.Queue.Id
		}

		if rule.Groups != nil {
			ruleSettings["groups"] = flattenCGRGroups(*rule.Groups)
		}

		rules[i] = ruleSettings
	}

	return rules
}

func flattenCGRGroups(sdkGroups []platformclientv2.Membergroup) *schema.Set {
	groupSet := schema.NewSet(schema.HashResource(memberGroupResource), []interface{}{})
	for _, sdkGroup := range sdkGroups {
		groupMap := make(map[string]interface{})
		resourcedata.SetMapValueIfNotNil(groupMap, "member_group_id", sdkGroup.Id)
		resourcedata.SetMapValueIfNotNil(groupMap, "member_group_type", sdkGroup.VarType)
		groupSet.Add(groupMap)
	}
	return groupSet
}

func flattenQueueMemberGroupsList(queue *platformclientv2.Queue, groupType *string) *schema.Set {
	var groupIds []string

	if queue == nil || queue.MemberGroups == nil {
		return nil
	}

	for _, memberGroup := range *queue.MemberGroups {
		if strings.Compare(*memberGroup.VarType, *groupType) == 0 {
			groupIds = append(groupIds, *memberGroup.Id)
		}
	}

	if groupIds != nil {
		return lists.StringListToSet(groupIds)
	}

	return nil
}

/*
The flattenBullseyeRings function maps the data retrieved from our SDK call over to the bullseye_ring attribute within the provider.
You might notice in the code that we are always mapping all but the last item in the list of rings retrieved by the API.  The reason for this
is that when you submit a list of bullseye_rings to the API, the API will always take the last item in the list and use it to drive default behavior
This is a change from earlier parts of the API where you could define 6 bullseye rings and there would always be six.  Now when you define bullseye rings,
the public API will take the list item in the list and make it the default and it will not show up on the screen.  To get around this you needed
to always add a dumb bullseye ring block.  Now, we automatically add one for you.  We only except a maximum of 5 bullseyes_ring blocks, but we will always
remove the last block returned by the API.
*/
func flattenBullseyeRings(sdkRings *[]platformclientv2.Ring) []interface{} {
	rings := make([]interface{}, len(*sdkRings)-1) //Sizing the target array of Rings to account for us removing the default block
	for i, sdkRing := range *sdkRings {
		if i < len(*sdkRings)-1 { //Checking to make sure we are do nothing with the last item in the list by skipping processing if it is defined
			ringSettings := make(map[string]interface{})
			if sdkRing.ExpansionCriteria != nil {
				for _, criteria := range *sdkRing.ExpansionCriteria {
					if *criteria.VarType == bullseyeExpansionTypeTimeout {
						ringSettings["expansion_timeout_seconds"] = *criteria.Threshold
						break
					}
				}
			}

			if sdkRing.Actions != nil && sdkRing.Actions.SkillsToRemove != nil {
				skillIds := make([]interface{}, len(*sdkRing.Actions.SkillsToRemove))
				for s, skill := range *sdkRing.Actions.SkillsToRemove {
					skillIds[s] = *skill.Id
				}
				ringSettings["skills_to_remove"] = schema.NewSet(schema.HashString, skillIds)
			}

			if sdkRing.MemberGroups != nil {
				memberGroups := schema.NewSet(schema.HashResource(memberGroupResource), []interface{}{})

				for _, memberGroup := range *sdkRing.MemberGroups {
					memberGroupMap := make(map[string]interface{})
					memberGroupMap["member_group_id"] = *memberGroup.Id
					memberGroupMap["member_group_type"] = *memberGroup.VarType

					memberGroups.Add(memberGroupMap)
				}

				ringSettings["member_groups"] = memberGroups
			}
			rings[i] = ringSettings
		}
	}
	return rings
}

func FlattenQueueEmailAddress(settings platformclientv2.Queueemailaddress) map[string]interface{} {
	settingsMap := make(map[string]interface{})

	if settings.Domain != nil {
		settingsMap["domain_id"] = *settings.Domain.Id
	}

	if settings.Route != nil {
		route := *settings.Route
		settingsMap["route_id"] = *route.Id
	}

	return settingsMap
}

func flattenQueueMembers(queueID string, memberBy string, sdkConfig *platformclientv2.Configuration) (*schema.Set, diag.Diagnostics) {
	members, err := getRoutingQueueMembers(queueID, memberBy, sdkConfig)
	if err != nil {
		return nil, err
	}

	memberSet := schema.NewSet(schema.HashResource(queueMemberResource), []interface{}{})
	for _, member := range members {
		memberMap := make(map[string]interface{})
		memberMap["user_id"] = *member.Id
		memberMap["ring_num"] = *member.RingNumber
		memberSet.Add(memberMap)
	}

	return memberSet, nil
}

func flattenQueueWrapupCodes(ctx context.Context, queueID string, proxy *RoutingQueueProxy) (*schema.Set, diag.Diagnostics) {
	codes, resp, err := proxy.getAllRoutingQueueWrapupCodes(ctx, queueID)
	codeIds := getWrapupCodeIds(codes)

	if err != nil {
		return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("failed to query wrapup codes for queue %s", queueID), resp)
	}

	if codeIds != nil {
		return lists.StringListToSet(codeIds), nil
	}

	return nil, nil
}

// Generate Functions

func GenerateRoutingQueueResourceBasic(resourceLabel string, name string, nestedBlocks ...string) string {
	return fmt.Sprintf(`resource "genesyscloud_routing_queue" "%s" {
		name = "%s"
		%s
	}
	`, resourceLabel, name, strings.Join(nestedBlocks, "\n"))
}

func GenerateRoutingQueueResource(
	resourceLabel string,
	name string,
	desc string,
	acwWrapupPrompt string,
	acwTimeout string,
	skillEvalMethod string,
	autoAnswerOnly string,
	callingPartyName string,
	callingPartyNumber string,
	enableTranscription string,
	suppressInQueueCallRecording string,
	enableAudioMonitoring string,
	enableManualAssignment string,
	scoringMethod string,
	peerId string,
	sourceQueueId string,
	nestedBlocks ...string) string {
	return fmt.Sprintf(`resource "genesyscloud_routing_queue" "%s" {
		name = "%s"
		description = "%s"
		acw_wrapup_prompt = %s
		acw_timeout_ms = %s
		skill_evaluation_method = %s
		auto_answer_only = %s
		calling_party_name = %s
		calling_party_number = %s
		enable_transcription = %s
		scoring_method = %s
		peer_id = %s
		source_queue_id = %s
        suppress_in_queue_call_recording = %s
		enable_audio_monitoring = %s
  		enable_manual_assignment = %s
		%s
	}
	`, resourceLabel,
		name,
		desc,
		acwWrapupPrompt,
		acwTimeout,
		skillEvalMethod,
		autoAnswerOnly,
		callingPartyName,
		callingPartyNumber,
		enableTranscription,
		scoringMethod,
		peerId,
		sourceQueueId,
		suppressInQueueCallRecording,
		enableAudioMonitoring,
		enableManualAssignment,
		strings.Join(nestedBlocks, "\n"))
}

// GenerateRoutingQueueResourceBasicWithDepends Used when testing skills group dependencies.
func GenerateRoutingQueueResourceBasicWithDepends(resourceLabel string, dependsOn string, name string, nestedBlocks ...string) string {
	return fmt.Sprintf(`resource "genesyscloud_routing_queue" "%s" {
		depends_on = [%s]
		name = "%s"
		%s
	}
	`, resourceLabel, dependsOn, name, strings.Join(nestedBlocks, "\n"))
}

func GenerateAgentOwnedRouting(attrName string, enableAgentOwnedCallBacks string, maxOwnedCallBackHours string, maxOwnedCallBackDelayHours string) string {
	return fmt.Sprintf(`%s {
		enable_agent_owned_callbacks = %s
		max_owned_callback_hours = %s
		max_owned_callback_delay_hours = %s
	}
	`, attrName, enableAgentOwnedCallBacks, maxOwnedCallBackHours, maxOwnedCallBackDelayHours)
}

func GenerateMediaSettings(attrName string, alertingTimeout string, enableAutoAnswer string, slPercent string, slDurationMs string) string {
	return fmt.Sprintf(`%s {
		alerting_timeout_sec = %s
		enable_auto_answer = %s
		service_level_percentage = %s
		service_level_duration_ms = %s
	}
	`, attrName, alertingTimeout, enableAutoAnswer, slPercent, slDurationMs)
}

func GenerateMediaSettingsCallBack(attrName string, alertingTimeout string, enableAutoAnswer string, slPercent string, slDurationMs string, enableAutoDial string, autoEndDelay string, autoDailDelay string) string {
	return fmt.Sprintf(`%s {
		alerting_timeout_sec = %s
		enable_auto_answer = %s
		service_level_percentage = %s
		service_level_duration_ms = %s
		enable_auto_dial_and_end = %s
		auto_end_delay_seconds = %s
		auto_dial_delay_seconds = %s
	}
	`, attrName, alertingTimeout, enableAutoAnswer, slPercent, slDurationMs, enableAutoDial, autoEndDelay, autoDailDelay)
}

func GenerateRoutingRules(operator string, threshold string, waitSeconds string) string {
	return fmt.Sprintf(`routing_rules {
		operator = "%s"
		threshold = %s
		wait_seconds = %s
	}
	`, operator, threshold, waitSeconds)
}

func GenerateDefaultScriptIDs(chat, email string) string {
	return fmt.Sprintf(`default_script_ids = {
		CHAT  = "%s"
		EMAIL = "%s"
	}`, chat, email)
}

func GenerateBullseyeSettings(expTimeout string, skillsToRemove ...string) string {
	return fmt.Sprintf(`bullseye_rings {
		expansion_timeout_seconds = %s
		skills_to_remove = [%s]
	}
	`, expTimeout, strings.Join(skillsToRemove, ", "))
}

func GenerateConditionalGroupRoutingRules(queueId, operator, metric, conditionValue, waitSeconds string, nestedBlocks ...string) string {
	return fmt.Sprintf(`conditional_group_routing_rules {
		queue_id        = %s
		operator        = "%s"
		metric          = "%s"
		condition_value = %s
		wait_seconds    = %s
		%s
	}
	`, queueId, operator, metric, conditionValue, waitSeconds, strings.Join(nestedBlocks, "\n"))
}

func GenerateConditionalGroupRoutingRuleGroup(groupId, groupType string) string {
	return fmt.Sprintf(`groups {
		member_group_id   = %s
		member_group_type = "%s"
	}
	`, groupId, groupType)
}

func GenerateBullseyeSettingsWithMemberGroup(expTimeout, memberGroupId, memberGroupType string, skillsToRemove ...string) string {
	return fmt.Sprintf(`bullseye_rings {
		expansion_timeout_seconds = %s
		skills_to_remove = [%s]
		member_groups {
			member_group_id = %s
			member_group_type = "%s"
		}
	}
	`, expTimeout, strings.Join(skillsToRemove, ", "), memberGroupId, memberGroupType)
}

func GenerateMemberBlock(userID, ringNum string) string {
	return fmt.Sprintf(`members {
		user_id = %s
		ring_num = %s
	}
	`, userID, ringNum)
}

func GenerateQueueWrapupCodes(wrapupCodes ...string) string {
	return fmt.Sprintf(`
		wrapup_codes = [%s]
	`, strings.Join(wrapupCodes, ", "))
}

func getRoutingQueueFromResourceData(d *schema.ResourceData) platformclientv2.Queue {
	skillGroups := buildMemberGroupList(d, "skill_groups", "SKILLGROUP")
	groups := buildMemberGroupList(d, "groups", "GROUP")
	teams := buildMemberGroupList(d, "teams", "TEAM")
	memberGroups := append(*skillGroups, *groups...)
	memberGroups = append(memberGroups, *teams...)
	division := d.Get("division_id").(string)

	return platformclientv2.Queue{
		Name:        platformclientv2.String(d.Get("name").(string)),
		Description: platformclientv2.String(d.Get("description").(string)),
		Division: &platformclientv2.Division{
			Id: &division,
		},
		MediaSettings:                buildSdkMediaSettings(d),
		RoutingRules:                 buildSdkRoutingRules(d),
		Bullseye:                     buildSdkBullseyeSettings(d),
		AcwSettings:                  buildSdkAcwSettings(d),
		AgentOwnedRouting:            constructAgentOwnedRouting(d),
		SkillEvaluationMethod:        platformclientv2.String(d.Get("skill_evaluation_method").(string)),
		QueueFlow:                    util.BuildSdkDomainEntityRef(d, "queue_flow_id"),
		EmailInQueueFlow:             util.BuildSdkDomainEntityRef(d, "email_in_queue_flow_id"),
		MessageInQueueFlow:           util.BuildSdkDomainEntityRef(d, "message_in_queue_flow_id"),
		WhisperPrompt:                util.BuildSdkDomainEntityRef(d, "whisper_prompt_id"),
		OnHoldPrompt:                 util.BuildSdkDomainEntityRef(d, "on_hold_prompt_id"),
		AutoAnswerOnly:               platformclientv2.Bool(d.Get("auto_answer_only").(bool)),
		CallingPartyName:             platformclientv2.String(d.Get("calling_party_name").(string)),
		CallingPartyNumber:           platformclientv2.String(d.Get("calling_party_number").(string)),
		DefaultScripts:               buildSdkDefaultScriptsMap(d),
		OutboundMessagingAddresses:   buildSdkQueueMessagingAddresses(d),
		EnableTranscription:          platformclientv2.Bool(d.Get("enable_transcription").(bool)),
		SuppressInQueueCallRecording: platformclientv2.Bool(d.Get("suppress_in_queue_call_recording").(bool)),
		EnableAudioMonitoring:        platformclientv2.Bool(d.Get("enable_audio_monitoring").(bool)),
		EnableManualAssignment:       platformclientv2.Bool(d.Get("enable_manual_assignment").(bool)),
		DirectRouting:                buildSdkDirectRouting(d),
		MemberGroups:                 &memberGroups,
		PeerId:                       platformclientv2.String(d.Get("peer_id").(string)),
		ScoringMethod:                platformclientv2.String(d.Get("scoring_method").(string)),
	}
}

/*
The below code is used during unit tests to go back to the singleton proxy approach
so that we can continue to mock proxy methods
*/

const unitTestsAreActiveEnv string = "TF_UNIT_ROUTING_QUEUE_TESTS"

func setRoutingQueueUnitTestsEnvVar() error {
	return os.Setenv(unitTestsAreActiveEnv, "true")
}

func unsetRoutingQueueUnitTestsEnvVar() error {
	return os.Unsetenv(unitTestsAreActiveEnv)
}

func isRoutingQueueUnitTestsActive() bool {
	_, isSet := os.LookupEnv(unitTestsAreActiveEnv)
	return isSet
}
