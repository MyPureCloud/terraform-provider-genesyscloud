package routing_queue

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"terraform-provider-genesyscloud/genesyscloud/util/resourcedata"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	chunksProcess "terraform-provider-genesyscloud/genesyscloud/util/chunks"
	"terraform-provider-genesyscloud/genesyscloud/util/lists"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v125/platformclientv2"
)

var bullseyeExpansionTypeTimeout = "TIMEOUT_SECONDS"

func getAllRoutingQueues(_ context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(resourceExporter.ResourceIDMetaMap)
	routingAPI := platformclientv2.NewRoutingApiWithConfig(clientConfig)

	// Newly created resources often aren't returned unless there's a delay
	time.Sleep(5 * time.Second)

	queues, _, getErr := routingAPI.GetRoutingQueues(1, 100, "", "", nil, nil, nil, false)
	if getErr != nil {
		return nil, diag.Errorf("Failed to get first page of queues: %v", getErr)
	}
	if queues.Entities == nil || len(*queues.Entities) == 0 {
		return resources, nil
	}
	for _, queue := range *queues.Entities {
		resources[*queue.Id] = &resourceExporter.ResourceMeta{Name: *queue.Name}
	}

	for pageNum := 2; pageNum <= *queues.PageCount; pageNum++ {
		const pageSize = 100
		queues, _, getErr := routingAPI.GetRoutingQueues(pageNum, pageSize, "", "", nil, nil, nil, false)
		if getErr != nil {
			return nil, diag.Errorf("Failed to get page of queues: %v", getErr)
		}

		if queues.Entities == nil || len(*queues.Entities) == 0 {
			break
		}

		for _, queue := range *queues.Entities {
			resources[*queue.Id] = &resourceExporter.ResourceMeta{Name: *queue.Name}
		}
	}

	return resources, nil
}

func createQueue(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	routingAPI := platformclientv2.NewRoutingApiWithConfig(sdkConfig)

	divisionID := d.Get("division_id").(string)
	skillGroups := buildMemberGroupList(d, "skill_groups", "SKILLGROUP")
	groups := buildMemberGroupList(d, "groups", "GROUP")
	teams := buildMemberGroupList(d, "teams", "TEAM")
	memberGroups := append(*skillGroups, *groups...)
	memberGroups = append(memberGroups, *teams...)

	createQueue := platformclientv2.Createqueuerequest{
		Name:                         platformclientv2.String(d.Get("name").(string)),
		Description:                  platformclientv2.String(d.Get("description").(string)),
		MediaSettings:                buildSdkMediaSettings(d),
		RoutingRules:                 buildSdkRoutingRules(d),
		Bullseye:                     buildSdkBullseyeSettings(d),
		AcwSettings:                  buildSdkAcwSettings(d),
		SkillEvaluationMethod:        platformclientv2.String(d.Get("skill_evaluation_method").(string)),
		QueueFlow:                    util.BuildSdkDomainEntityRef(d, "queue_flow_id"),
		EmailInQueueFlow:             util.BuildSdkDomainEntityRef(d, "email_in_queue_flow_id"),
		MessageInQueueFlow:           util.BuildSdkDomainEntityRef(d, "message_in_queue_flow_id"),
		WhisperPrompt:                util.BuildSdkDomainEntityRef(d, "whisper_prompt_id"),
		AutoAnswerOnly:               platformclientv2.Bool(d.Get("auto_answer_only").(bool)),
		CallingPartyName:             platformclientv2.String(d.Get("calling_party_name").(string)),
		CallingPartyNumber:           platformclientv2.String(d.Get("calling_party_number").(string)),
		DefaultScripts:               buildSdkDefaultScriptsMap(d),
		OutboundMessagingAddresses:   buildSdkQueueMessagingAddresses(d),
		OutboundEmailAddress:         buildSdkQueueEmailAddress(d),
		EnableTranscription:          platformclientv2.Bool(d.Get("enable_transcription").(bool)),
		SuppressInQueueCallRecording: platformclientv2.Bool(d.Get("suppress_in_queue_call_recording").(bool)),
		EnableManualAssignment:       platformclientv2.Bool(d.Get("enable_manual_assignment").(bool)),
		DirectRouting:                buildSdkDirectRouting(d),
		MemberGroups:                 &memberGroups,
	}

	if divisionID != "" {
		createQueue.Division = &platformclientv2.Writabledivision{Id: &divisionID}
	}

	log.Printf("creating queue %s using routingAPI.PostRoutingQueues", *createQueue.Name)
	queue, resp, err := routingAPI.PostRoutingQueues(createQueue)
	if err != nil {
		log.Printf("error while trying to create queue: %s. Err %s", *createQueue.Name, err)
		return diag.Errorf("Failed to create queue %s: %s", *createQueue.Name, err)
	}

	if resp.StatusCode != http.StatusOK {
		return diag.Errorf("Failed to create queue %s: with httpStatus code: %d", *createQueue.Name, resp.StatusCode)
	}

	d.SetId(*queue.Id)

	diagErr := updateQueueMembers(d, sdkConfig)
	if diagErr != nil {
		return diagErr
	}

	diagErr = updateQueueWrapupCodes(d, routingAPI)
	if diagErr != nil {
		return diagErr
	}

	return readQueue(ctx, d, meta)
}

func readQueue(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	routingAPI := platformclientv2.NewRoutingApiWithConfig(sdkConfig)

	log.Printf("Reading queue %s", d.Id())
	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		currentQueue, resp, getErr := routingAPI.GetRoutingQueue(d.Id())
		if getErr != nil {
			if util.IsStatus404(resp) {
				return retry.RetryableError(fmt.Errorf("Failed to read queue %s: %s", d.Id(), getErr))
			}
			return retry.NonRetryableError(fmt.Errorf("Failed to read queue %s: %s", d.Id(), getErr))
		}

		cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceRoutingQueue())

		resourcedata.SetNillableValue(d, "name", currentQueue.Name)
		resourcedata.SetNillableValue(d, "description", currentQueue.Description)
		resourcedata.SetNillableValue(d, "skill_evaluation_method", currentQueue.SkillEvaluationMethod)

		resourcedata.SetNillableReferenceDivision(d, "division_id", currentQueue.Division)

		_ = d.Set("acw_wrapup_prompt", nil)
		_ = d.Set("acw_timeout_ms", nil)

		if currentQueue.AcwSettings != nil {
			resourcedata.SetNillableValue(d, "acw_wrapup_prompt", currentQueue.AcwSettings.WrapupPrompt)
			resourcedata.SetNillableValue(d, "acw_timeout_ms", currentQueue.AcwSettings.TimeoutMs)
		}

		_ = d.Set("media_settings_call", nil)
		_ = d.Set("media_settings_callback", nil)
		_ = d.Set("media_settings_chat", nil)
		_ = d.Set("media_settings_email", nil)
		_ = d.Set("media_settings_message", nil)

		if currentQueue.MediaSettings != nil {
			resourcedata.SetNillableValueWithInterfaceArrayWithFunc(d, "media_settings_call", currentQueue.MediaSettings.Call, flattenMediaSetting)
			resourcedata.SetNillableValueWithInterfaceArrayWithFunc(d, "media_settings_callback", currentQueue.MediaSettings.Callback, flattenMediaSettingCallback)
			resourcedata.SetNillableValueWithInterfaceArrayWithFunc(d, "media_settings_chat", currentQueue.MediaSettings.Chat, flattenMediaSetting)
			resourcedata.SetNillableValueWithInterfaceArrayWithFunc(d, "media_settings_email", currentQueue.MediaSettings.Email, flattenMediaSetting)
			resourcedata.SetNillableValueWithInterfaceArrayWithFunc(d, "media_settings_message", currentQueue.MediaSettings.Message, flattenMediaSetting)
		}

		resourcedata.SetNillableValueWithInterfaceArrayWithFunc(d, "routing_rules", currentQueue.RoutingRules, flattenRoutingRules)

		if currentQueue.Bullseye != nil {
			resourcedata.SetNillableValueWithInterfaceArrayWithFunc(d, "bullseye_rings", currentQueue.Bullseye.Rings, flattenBullseyeRings)
		}

		resourcedata.SetNillableReference(d, "queue_flow_id", currentQueue.QueueFlow)
		resourcedata.SetNillableReference(d, "message_in_queue_flow_id", currentQueue.MessageInQueueFlow)
		resourcedata.SetNillableReference(d, "email_in_queue_flow_id", currentQueue.EmailInQueueFlow)
		resourcedata.SetNillableReference(d, "whisper_prompt_id", currentQueue.WhisperPrompt)
		resourcedata.SetNillableValue(d, "auto_answer_only", currentQueue.AutoAnswerOnly)
		resourcedata.SetNillableValue(d, "enable_transcription", currentQueue.EnableTranscription)
		resourcedata.SetNillableValue(d, "suppress_in_queue_call_recording", currentQueue.SuppressInQueueCallRecording)
		resourcedata.SetNillableValue(d, "enable_manual_assignment", currentQueue.EnableManualAssignment)
		resourcedata.SetNillableValue(d, "calling_party_name", currentQueue.CallingPartyName)
		resourcedata.SetNillableValue(d, "calling_party_number", currentQueue.CallingPartyNumber)

		if currentQueue.DefaultScripts != nil {
			_ = d.Set("default_script_ids", flattenDefaultScripts(*currentQueue.DefaultScripts))
		} else {
			_ = d.Set("default_script_ids", nil)
		}

		if currentQueue.OutboundMessagingAddresses != nil && currentQueue.OutboundMessagingAddresses.SmsAddress != nil {
			_ = d.Set("outbound_messaging_sms_address_id", *currentQueue.OutboundMessagingAddresses.SmsAddress.Id)
		} else {
			_ = d.Set("outbound_messaging_sms_address_id", nil)
		}

		if currentQueue.OutboundEmailAddress != nil && *currentQueue.OutboundEmailAddress != nil {
			outboundEmailAddress := *currentQueue.OutboundEmailAddress
			_ = d.Set("outbound_email_address", []interface{}{FlattenQueueEmailAddress(*outboundEmailAddress)})
		} else {
			_ = d.Set("outbound_email_address", nil)
		}

		resourcedata.SetNillableValueWithInterfaceArrayWithFunc(d, "direct_routing", currentQueue.DirectRouting, flattenDirectRouting)

		wrapupCodes, err := flattenQueueWrapupCodes(d.Id(), routingAPI)
		if err != nil {
			return retry.NonRetryableError(fmt.Errorf("%v", err))
		}
		_ = d.Set("wrapup_codes", wrapupCodes)

		members, err := flattenQueueMembers(d.Id(), "user", sdkConfig)
		if err != nil {
			return retry.NonRetryableError(fmt.Errorf("%v", err))
		}
		_ = d.Set("members", members)

		skillGroup := "SKILLGROUP"
		team := "TEAM"
		group := "GROUP"

		_ = d.Set("skill_groups", flattenQueueMemberGroupsList(currentQueue, &skillGroup))
		_ = d.Set("teams", flattenQueueMemberGroupsList(currentQueue, &team))
		_ = d.Set("groups", flattenQueueMemberGroupsList(currentQueue, &group))

		log.Printf("Done reading queue %s %s", d.Id(), *currentQueue.Name)
		return cc.CheckState()
	})
}

func updateQueue(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	routingAPI := platformclientv2.NewRoutingApiWithConfig(sdkConfig)

	skillGroups := buildMemberGroupList(d, "skill_groups", "SKILLGROUP")
	groups := buildMemberGroupList(d, "groups", "GROUP")
	teams := buildMemberGroupList(d, "teams", "TEAM")
	memberGroups := append(*skillGroups, *groups...)
	memberGroups = append(memberGroups, *teams...)

	updateQueue := platformclientv2.Queuerequest{
		Name:                         platformclientv2.String(d.Get("name").(string)),
		Description:                  platformclientv2.String(d.Get("description").(string)),
		MediaSettings:                buildSdkMediaSettings(d),
		RoutingRules:                 buildSdkRoutingRules(d),
		Bullseye:                     buildSdkBullseyeSettings(d),
		AcwSettings:                  buildSdkAcwSettings(d),
		SkillEvaluationMethod:        platformclientv2.String(d.Get("skill_evaluation_method").(string)),
		QueueFlow:                    util.BuildSdkDomainEntityRef(d, "queue_flow_id"),
		EmailInQueueFlow:             util.BuildSdkDomainEntityRef(d, "email_in_queue_flow_id"),
		MessageInQueueFlow:           util.BuildSdkDomainEntityRef(d, "message_in_queue_flow_id"),
		WhisperPrompt:                util.BuildSdkDomainEntityRef(d, "whisper_prompt_id"),
		AutoAnswerOnly:               platformclientv2.Bool(d.Get("auto_answer_only").(bool)),
		CallingPartyName:             platformclientv2.String(d.Get("calling_party_name").(string)),
		CallingPartyNumber:           platformclientv2.String(d.Get("calling_party_number").(string)),
		DefaultScripts:               buildSdkDefaultScriptsMap(d),
		OutboundMessagingAddresses:   buildSdkQueueMessagingAddresses(d),
		OutboundEmailAddress:         buildSdkQueueEmailAddress(d),
		EnableTranscription:          platformclientv2.Bool(d.Get("enable_transcription").(bool)),
		SuppressInQueueCallRecording: platformclientv2.Bool(d.Get("suppress_in_queue_call_recording").(bool)),
		EnableManualAssignment:       platformclientv2.Bool(d.Get("enable_manual_assignment").(bool)),
		DirectRouting:                buildSdkDirectRouting(d),
		MemberGroups:                 &memberGroups,
	}

	log.Printf("Updating queue %s", *updateQueue.Name)
	_, _, err := routingAPI.PutRoutingQueue(d.Id(), updateQueue)

	if err != nil {
		return diag.Errorf("Error updating queue %s: %s", *updateQueue.Name, err)
	}

	diagErr := util.UpdateObjectDivision(d, "QUEUE", sdkConfig)
	if diagErr != nil {
		return diagErr
	}

	diagErr = updateQueueMembers(d, sdkConfig)
	if diagErr != nil {
		return diagErr
	}

	diagErr = updateQueueWrapupCodes(d, routingAPI)
	if diagErr != nil {
		return diagErr
	}

	log.Printf("Finished updating queue %s", *updateQueue.Name)
	return readQueue(ctx, d, meta)
}

func deleteQueue(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)

	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	routingAPI := platformclientv2.NewRoutingApiWithConfig(sdkConfig)

	log.Printf("Deleting queue %s", name)
	_, err := routingAPI.DeleteRoutingQueue(d.Id(), true)
	if err != nil {
		return diag.Errorf("Failed to delete queue %s: %s", name, err)
	}

	// Queue deletes are not immediate. Query until queue is no longer found
	// Add a delay before the first request to reduce the likelihood of public API's cache
	// re-populating the queue after the delete. Otherwise it may not expire for a minute.
	time.Sleep(5 * time.Second)

	//DEVTOOLING-238- Increasing this to a 120 seconds to see if we can temporarily mitigate a problem for a customer
	return util.WithRetries(ctx, 120*time.Second, func() *retry.RetryError {
		_, resp, err := routingAPI.GetRoutingQueue(d.Id())
		if err != nil {
			if util.IsStatus404(resp) {
				// Queue deleted
				log.Printf("Queue %s deleted", name)
				return nil
			}
			return retry.NonRetryableError(fmt.Errorf("Error deleting queue %s: %s", d.Id(), err))
		}
		return retry.RetryableError(fmt.Errorf("Queue %s still exists", d.Id()))
	})
}

func buildSdkMediaSettings(d *schema.ResourceData) *platformclientv2.Queuemediasettings {
	queueMediaSettings := &platformclientv2.Queuemediasettings{}

	mediaSettingsCall := d.Get("media_settings_call").([]interface{})
	if mediaSettingsCall != nil && len(mediaSettingsCall) > 0 {
		queueMediaSettings.Call = buildSdkMediaSetting(mediaSettingsCall)
	}

	mediaSettingsCallback := d.Get("media_settings_callback").([]interface{})
	if mediaSettingsCallback != nil && len(mediaSettingsCallback) > 0 {
		queueMediaSettings.Callback = buildSdkMediaSettingCallback(mediaSettingsCallback)
	}

	mediaSettingsChat := d.Get("media_settings_chat").([]interface{})
	if mediaSettingsChat != nil && len(mediaSettingsChat) > 0 {
		queueMediaSettings.Chat = buildSdkMediaSetting(mediaSettingsChat)
	}

	mediaSettingsEmail := d.Get("media_settings_email").([]interface{})
	log.Printf("The media settings email #%v", mediaSettingsEmail)
	if mediaSettingsEmail != nil && len(mediaSettingsEmail) > 0 {
		queueMediaSettings.Email = buildSdkMediaSetting(mediaSettingsEmail)
	}

	mediaSettingsMessage := d.Get("media_settings_message").([]interface{})
	if mediaSettingsMessage != nil && len(mediaSettingsMessage) > 0 {
		queueMediaSettings.Message = buildSdkMediaSetting(mediaSettingsMessage)
	}

	return queueMediaSettings
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
		EnableAutoAnswer: platformclientv2.Bool(settingsMap["enable_auto_answer"].(bool)),
	}
}

func flattenMediaSetting(settings *platformclientv2.Mediasettings) []interface{} {
	settingsMap := make(map[string]interface{})

	settingsMap["alerting_timeout_sec"] = *settings.AlertingTimeoutSeconds
	resourcedata.SetMapValueIfNotNil(settingsMap, "enable_auto_answer", settings.EnableAutoAnswer)
	settingsMap["service_level_percentage"] = *settings.ServiceLevel.Percentage
	settingsMap["service_level_duration_ms"] = *settings.ServiceLevel.DurationMs

	return []interface{}{settingsMap}
}

func flattenMediaSettingCallback(settings *platformclientv2.Callbackmediasettings) []interface{} {
	settingsMap := make(map[string]interface{})

	settingsMap["alerting_timeout_sec"] = *settings.AlertingTimeoutSeconds
	settingsMap["service_level_percentage"] = *settings.ServiceLevel.Percentage
	settingsMap["service_level_duration_ms"] = *settings.ServiceLevel.DurationMs
	resourcedata.SetMapValueIfNotNil(settingsMap, "enable_auto_answer", settings.EnableAutoAnswer)

	return []interface{}{settingsMap}
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

func validateMapCommTypes(val interface{}, _ cty.Path) diag.Diagnostics {
	if val == nil {
		return nil
	}

	commTypes := []string{"CALL", "CALLBACK", "CHAT", "COBROWSE", "EMAIL", "MESSAGE", "SOCIAL_EXPRESSION", "VIDEO", "SCREENSHARE"}
	m := val.(map[string]interface{})
	for k := range m {
		if !lists.ItemInSlice(k, commTypes) {
			return diag.Errorf("%s is an invalid communication type key.", k)
		}
	}
	return nil
}

func buildSdkQueueMessagingAddresses(d *schema.ResourceData) *platformclientv2.Queuemessagingaddresses {
	if _, ok := d.GetOk("outbound_messaging_sms_address_id"); ok {
		return &platformclientv2.Queuemessagingaddresses{
			SmsAddress: util.BuildSdkDomainEntityRef(d, "outbound_messaging_sms_address_id"),
		}
	}
	return nil
}

func buildSdkQueueEmailAddress(d *schema.ResourceData) *platformclientv2.Queueemailaddress {
	outboundEmailAddress := d.Get("outbound_email_address").([]interface{})
	if outboundEmailAddress != nil && len(outboundEmailAddress) > 0 {
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

func FlattenQueueEmailAddress(settings platformclientv2.Queueemailaddress) map[string]interface{} {
	settingsMap := make(map[string]interface{})
	resourcedata.SetMapReferenceValueIfNotNil(settingsMap, "domain_id", settings.Domain)

	if settings.Route != nil {
		route := *settings.Route
		settingsMap["route_id"] = *route.Id
	}

	return settingsMap
}

func buildSdkDirectRouting(d *schema.ResourceData) *platformclientv2.Directrouting {
	directRouting := d.Get("direct_routing").([]interface{})
	if directRouting != nil && len(directRouting) > 0 {
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

func updateQueueWrapupCodes(d *schema.ResourceData, routingAPI *platformclientv2.RoutingApi) diag.Diagnostics {
	if d.HasChange("wrapup_codes") {
		if codesConfig := d.Get("wrapup_codes"); codesConfig != nil {
			// Get existing codes
			codes, err := getRoutingQueueWrapupCodes(d.Id(), routingAPI)
			if err != nil {
				return err
			}

			var existingCodes []string
			if codes != nil {
				for _, code := range codes {
					existingCodes = append(existingCodes, *code.Id)
				}
			}
			configCodes := *lists.SetToStringList(codesConfig.(*schema.Set))

			codesToRemove := lists.SliceDifference(existingCodes, configCodes)
			if len(codesToRemove) > 0 {
				for _, codeId := range codesToRemove {
					resp, err := routingAPI.DeleteRoutingQueueWrapupcode(d.Id(), codeId)
					if err != nil {
						if util.IsStatus404(resp) {
							// Ignore missing queue or wrapup code
							continue
						}
						return diag.Errorf("Failed to remove wrapup code from queue %s: %s", d.Id(), err)
					}
				}
			}

			codesToAdd := lists.SliceDifference(configCodes, existingCodes)
			if len(codesToAdd) > 0 {
				err := addWrapupCodesInChunks(d.Id(), codesToAdd, routingAPI)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func addWrapupCodesInChunks(queueID string, codesToAdd []string, api *platformclientv2.RoutingApi) diag.Diagnostics {
	// API restricts wrapup code adds to 100 per call
	const maxBatchSize = 100
	for i := 0; i < len(codesToAdd); i += maxBatchSize {
		end := i + maxBatchSize
		if end > len(codesToAdd) {
			end = len(codesToAdd)
		}
		var updateChunk []platformclientv2.Wrapupcodereference
		for j := i; j < end; j++ {
			updateChunk = append(updateChunk, platformclientv2.Wrapupcodereference{Id: &codesToAdd[j]})
		}

		if len(updateChunk) > 0 {
			_, _, err := api.PostRoutingQueueWrapupcodes(queueID, updateChunk)
			if err != nil {
				return diag.Errorf("Failed to update wrapup codes in queue %s: %s", queueID, err)
			}
		}
	}
	return nil
}

func getRoutingQueueWrapupCodes(queueID string, api *platformclientv2.RoutingApi) ([]platformclientv2.Wrapupcode, diag.Diagnostics) {
	const maxPageSize = 100

	var codes []platformclientv2.Wrapupcode
	for pageNum := 1; ; pageNum++ {
		codeResult, _, err := api.GetRoutingQueueWrapupcodes(queueID, maxPageSize, pageNum)
		if err != nil {
			return nil, diag.Errorf("Failed to query wrapup codes for queue %s: %s", queueID, err)
		}
		if codeResult == nil || codeResult.Entities == nil || len(*codeResult.Entities) == 0 {
			return codes, nil
		}
		for _, code := range *codeResult.Entities {
			codes = append(codes, code)
		}
	}
}

func updateQueueMembers(d *schema.ResourceData, sdkConfig *platformclientv2.Configuration) diag.Diagnostics {
	if !d.HasChange("members") {
		return nil
	}
	membersSet, ok := d.Get("members").(*schema.Set)
	if !ok || membersSet.Len() == 0 {
		if err := removeAllExistingUserMembersFromQueue(d.Id(), sdkConfig); err != nil {
			return diag.FromErr(err)
		}
		return nil
	}

	log.Printf("Updating members for Queue %s", d.Get("name"))
	newUserRingNums := make(map[string]int)
	memberList := membersSet.List()
	newUserIds := make([]string, len(memberList))
	for i, member := range memberList {
		memberMap := member.(map[string]interface{})
		newUserIds[i] = memberMap["user_id"].(string)
		newUserRingNums[newUserIds[i]] = memberMap["ring_num"].(int)
	}

	if len(newUserIds) > 0 {
		log.Printf("Sleeping for 10 seconds")
		time.Sleep(10 * time.Second)
		for _, userId := range newUserIds {
			if err := verifyUserIsNotGroupMemberOfQueue(d.Id(), userId, sdkConfig); err != nil {
				return diag.Errorf("%v", err)
			}
		}
	}

	oldSdkUsers, err := getRoutingQueueMembers(d.Id(), "user", sdkConfig)
	if err != nil {
		return err
	}

	oldUserIds := make([]string, len(oldSdkUsers))
	oldUserRingNums := make(map[string]int)
	for i, user := range oldSdkUsers {
		oldUserIds[i] = *user.Id
		oldUserRingNums[oldUserIds[i]] = *user.RingNumber
	}

	if len(oldUserIds) > 0 {
		usersToRemove := lists.SliceDifference(oldUserIds, newUserIds)
		err := updateMembersInChunks(d.Id(), usersToRemove, true, sdkConfig)
		if err != nil {
			return err
		}
	}

	if len(newUserIds) > 0 {
		usersToAdd := lists.SliceDifference(newUserIds, oldUserIds)
		err := updateMembersInChunks(d.Id(), usersToAdd, false, sdkConfig)
		if err != nil {
			return err
		}
	}

	// Check for ring numbers to update
	for userID, newNum := range newUserRingNums {
		if oldNum, found := oldUserRingNums[userID]; found {
			if newNum != oldNum {
				log.Printf("updating ring_num for user %s because it has updated. New: %v, Old: %v", userID, newNum, oldNum)
				// Number changed. Update ring number
				err := updateQueueUserRingNum(d.Id(), userID, newNum, sdkConfig)
				if err != nil {
					return err
				}
			}
		} else if newNum != 1 {
			// New queue member. Update ring num if not set to the default of 1
			log.Printf("updating user %s ring_num because it is not the default 1", userID)
			err := updateQueueUserRingNum(d.Id(), userID, newNum, sdkConfig)
			if err != nil {
				return err
			}
		}
	}
	log.Printf("Members updated for Queue %s", d.Get("name"))

	return nil
}

// removeAllExistingUserMembersFromQueue get all existing user members of a given queue and remove them from the queue
func removeAllExistingUserMembersFromQueue(queueId string, sdkConfig *platformclientv2.Configuration) error {
	log.Printf("Reading user members of queue %s", queueId)
	oldSdkUsers, err := getRoutingQueueMembers(queueId, "user", sdkConfig)
	if err != nil {
		return fmt.Errorf("%v", err)
	}
	log.Printf("Read user members of queue %s", queueId)

	var oldUserIds []string
	for _, user := range oldSdkUsers {
		oldUserIds = append(oldUserIds, *user.Id)
	}

	if len(oldUserIds) > 0 {
		log.Printf("Removing queue %s user members", queueId)
		if err := updateMembersInChunks(queueId, oldUserIds, true, sdkConfig); err != nil {
			return fmt.Errorf("%v", err)
		}
		log.Printf("Removing queue %s user members", queueId)
	}
	return nil
}

// verifyUserIsNotGroupMemberOfQueue Search through queue group members to verify that a given user is not a group member
func verifyUserIsNotGroupMemberOfQueue(queueId, userId string, sdkConfig *platformclientv2.Configuration) error {
	var (
		userName   string
		routingApi = platformclientv2.NewRoutingApiWithConfig(sdkConfig)
		usersApi   = platformclientv2.NewUsersApiWithConfig(sdkConfig)
	)

	log.Printf("verifying that member '%s' is not assinged to the queue '%s' via a group", userId, queueId)

	// Read name of user to filter results when listing members of queue
	log.Printf("reading user %s to fetch name", userId)
	user, _, err := usersApi.GetUser(userId, nil, "", "")
	if err != nil {
		log.Printf("Failed to read name of user '%s' inside verifyUserIsNotGroupMemberOfQueue: %s. Queue ID: %s", userId, err, queueId)
	} else {
		userName = *user.Name
		log.Printf("read user %s %s", userId, userName)
	}

	const pageSize = 100
	for pageNum := 1; ; pageNum++ {
		users, resp, err := sdkGetRoutingQueueMembers(queueId, "group", userName, pageNum, pageSize, routingApi)
		if err != nil || resp.StatusCode != http.StatusOK {
			log.Printf("Error requesting group members of queue '%s': %v. Cannot validate that user '%s' is not already assigned via a group", queueId, err, userId)
			break
		}
		if users == nil || users.Entities == nil || len(*users.Entities) == 0 {
			break
		}
		for _, member := range *users.Entities {
			if userId == *member.Id {
				return fmt.Errorf("member %s '%s' is already assigned to queue '%s' via a group, and cannot be assigned using the members set", userName, userId, queueId)
			}
		}
	}

	log.Printf("User %s not found as group member in queue %s", userId, queueId)
	return nil
}

func updateMembersInChunks(queueID string, membersToUpdate []string, remove bool, sdkConfig *platformclientv2.Configuration) diag.Diagnostics {
	api := platformclientv2.NewRoutingApiWithConfig(sdkConfig)
	// API restricts member adds/removes to 100 per call
	// Generic call to prepare chunks for the Update. Takes in three args
	// 1. MemberstoUpdate 2. The Entity prepare func for the update 3. Chunk Size
	if len(membersToUpdate) > 0 {
		chunks := chunksProcess.ChunkItems(membersToUpdate, platformWritableEntityFunc, 100)
		// Closure to process the chunks
		chunkProcessor := func(chunk []platformclientv2.Writableentity) diag.Diagnostics {
			_, err := api.PostRoutingQueueMembers(queueID, chunk, remove)
			if err != nil {
				return diag.Errorf("Failed to update members in queue %s: %s", queueID, err)
			}
			return nil
		}
		// Generic Function call which takes in the chunks and the processing function
		return chunksProcess.ProcessChunks(chunks, chunkProcessor)
	}
	return nil

}

func platformWritableEntityFunc(val string) platformclientv2.Writableentity {
	return platformclientv2.Writableentity{Id: &val}
}

func updateQueueUserRingNum(queueID string, userID string, ringNum int, sdkConfig *platformclientv2.Configuration) diag.Diagnostics {
	api := platformclientv2.NewRoutingApiWithConfig(sdkConfig)
	_, err := api.PatchRoutingQueueMember(queueID, userID, platformclientv2.Queuemember{
		Id:         &userID,
		RingNumber: &ringNum,
	})
	if err != nil {
		return diag.Errorf("Failed to update ring number for queue %s user %s: %s", queueID, userID, err)
	}
	return nil
}

func getRoutingQueueMembers(queueID string, memberBy string, sdkConfig *platformclientv2.Configuration) ([]platformclientv2.Queuemember, diag.Diagnostics) {
	var members []platformclientv2.Queuemember
	const pageSize = 100
	api := platformclientv2.NewRoutingApiWithConfig(sdkConfig)

	// Need to call this method to find the member count for a queue. GetRoutingQueueMembers does not return a `total` property for us to use.
	queue, _, err := api.GetRoutingQueue(queueID)
	if err != nil {
		return nil, diag.Errorf("Can't find queue %s", queueID)
	}
	queueMembers := *queue.MemberCount
	log.Printf("%d members belong to queue %s", queueMembers, queueID)

	for pageNum := 1; ; pageNum++ {
		users, resp, err := sdkGetRoutingQueueMembers(queueID, memberBy, "", pageNum, pageSize, api)
		if err != nil || resp.StatusCode != http.StatusOK {
			return nil, diag.Errorf("Failed to query users for queue %s: %s", queueID, err)
		}
		if users == nil || users.Entities == nil || len(*users.Entities) == 0 {
			membersFound := len(members)
			log.Printf("%d queue members found for queue %s", membersFound, queueID)
			if membersFound != queueMembers {
				log.Printf("Member count is not equal to queue member found for queue %s, Correlation Id: %s", queueID, resp.CorrelationID)
			}
			return members, nil
		}
		for _, user := range *users.Entities {
			members = append(members, user)
		}
	}
}

func sdkGetRoutingQueueMembers(queueID, memberBy, name string, pageNumber, pageSize int, api *platformclientv2.RoutingApi) (*platformclientv2.Queuememberentitylisting, *platformclientv2.APIResponse, error) {
	// SDK does not support nil values for boolean query params yet, so we must manually construct this HTTP request for now
	apiClient := &api.Configuration.APIClient

	// create path and map variables
	path := api.Configuration.BasePath + "/api/v2/routing/queues/{queueId}/members"
	path = strings.Replace(path, "{queueId}", queueID, -1)

	headerParams := make(map[string]string)
	queryParams := make(map[string]string)
	formParams := url.Values{}
	var postBody interface{}
	var postFileName string
	var fileBytes []byte

	// oauth required
	if api.Configuration.AccessToken != "" {
		headerParams["Authorization"] = "Bearer " + api.Configuration.AccessToken
	}
	// add default headers if any
	for key := range api.Configuration.DefaultHeader {
		headerParams[key] = api.Configuration.DefaultHeader[key]
	}

	queryParams["pageSize"] = apiClient.ParameterToString(pageSize, "")
	queryParams["pageNumber"] = apiClient.ParameterToString(pageNumber, "")
	if memberBy != "" {
		queryParams["memberBy"] = memberBy
	}
	if name != "" {
		queryParams["name"] = name
	}

	headerParams["Content-Type"] = "application/json"
	headerParams["Accept"] = "application/json"

	var successPayload *platformclientv2.Queuememberentitylisting
	response, err := apiClient.CallAPI(path, http.MethodGet, postBody, headerParams, queryParams, formParams, postFileName, fileBytes)
	if err != nil {
		// Nothing special to do here, but do avoid processing the response
	} else if response.Error != nil {
		err = fmt.Errorf(response.ErrorMessage)
	} else {
		err = json.Unmarshal([]byte(response.RawBody), &successPayload)
	}
	return successPayload, response, err
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

func flattenQueueWrapupCodes(queueID string, api *platformclientv2.RoutingApi) (*schema.Set, diag.Diagnostics) {
	const maxPageSize = 100
	var codeIds []string
	for pageNum := 1; ; pageNum++ {
		codes, _, err := api.GetRoutingQueueWrapupcodes(queueID, maxPageSize, pageNum)
		if err != nil {
			return nil, diag.Errorf("Failed to query wrapup codes for queue %s: %s", queueID, err)
		}
		if codes == nil || codes.Entities == nil || len(*codes.Entities) == 0 {
			break
		}
		for _, code := range *codes.Entities {
			codeIds = append(codeIds, *code.Id)
		}
	}

	if codeIds != nil {
		return lists.StringListToSet(codeIds), nil
	}
	return nil, nil
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
