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
	"terraform-provider-genesyscloud/genesyscloud/util/constants"
	featureToggles "terraform-provider-genesyscloud/genesyscloud/util/feature_toggles"
	"terraform-provider-genesyscloud/genesyscloud/util/resourcedata"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	chunksProcess "terraform-provider-genesyscloud/genesyscloud/util/chunks"
	"terraform-provider-genesyscloud/genesyscloud/util/lists"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

var bullseyeExpansionTypeTimeout = "TIMEOUT_SECONDS"

func getAllRoutingQueues(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(resourceExporter.ResourceIDMetaMap)
	proxy := GetRoutingQueueProxy(clientConfig)

	// Newly created resources often aren't returned unless there's a delay
	time.Sleep(5 * time.Second)

	queues, resp, err := proxy.GetAllRoutingQueues(ctx, "")
	if err != nil {
		return nil, util.BuildAPIDiagnosticError(resourceName, fmt.Sprintf("failed to get routing queues: %s", err), resp)
	}

	for _, queue := range *queues {
		resources[*queue.Id] = &resourceExporter.ResourceMeta{Name: *queue.Name}
	}

	return resources, nil
}

func createQueue(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	routingAPI := platformclientv2.NewRoutingApiWithConfig(sdkConfig)

	divisionID := d.Get("division_id").(string)
	scoringMethod := d.Get("scoring_method").(string)
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
	}

	if exists := featureToggles.CSGToggleExists(); !exists {
		conditionalGroupRouting, diagErr := buildSdkConditionalGroupRouting(d)
		if diagErr != nil {
			return diagErr
		}
		createQueue.ConditionalGroupRouting = conditionalGroupRouting
	} else {
		log.Printf("%s is set, not creating conditional_group_routing_rules attribute in routing_queue %s resource", featureToggles.CSGToggleName(), d.Id())
	}

	if exists := featureToggles.OEAToggleExists(); !exists {
		createQueue.OutboundEmailAddress = buildSdkQueueEmailAddress(d)
	} else {
		log.Printf("%s is set, not creating outbound_email_address attribute in routing_queue %s resource", featureToggles.OEAToggleName(), d.Id())
	}

	if divisionID != "" {
		createQueue.Division = &platformclientv2.Writabledivision{Id: &divisionID}
	}

	if scoringMethod != "" {
		createQueue.ScoringMethod = &scoringMethod
	}
	queue, resp, err := routingAPI.PostRoutingQueues(createQueue)
	if err != nil {
		log.Printf("error while trying to create queue: %s. Err %s", *createQueue.Name, err)
		return util.BuildAPIDiagnosticError(resourceName, fmt.Sprintf("Failed to create queue %s error: %s", *createQueue.Name, err), resp)
	}

	if resp.StatusCode != http.StatusOK {
		return util.BuildAPIDiagnosticError(resourceName, fmt.Sprintf("Failed to create queue %s with error: %s, status code %v", *createQueue.Name, err, resp.StatusCode), resp)
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
	proxy := GetRoutingQueueProxy(sdkConfig)
	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceRoutingQueue(), constants.DefaultConsistencyChecks, resourceName)

	log.Printf("Reading queue %s", d.Id())
	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		currentQueue, resp, getErr := proxy.getRoutingQueueById(ctx, d.Id())
		if getErr != nil {
			if util.IsStatus404(resp) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(resourceName, fmt.Sprintf("Failed to read queue %s | error: %s", d.Id(), getErr), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(resourceName, fmt.Sprintf("Failed to read queue %s | error: %s", d.Id(), getErr), resp))
		}

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
		_ = d.Set("agent_owned_routing", nil)

		if currentQueue.MediaSettings != nil {
			resourcedata.SetNillableValueWithInterfaceArrayWithFunc(d, "media_settings_call", currentQueue.MediaSettings.Call, flattenMediaSetting)
			resourcedata.SetNillableValueWithInterfaceArrayWithFunc(d, "media_settings_callback", currentQueue.MediaSettings.Callback, flattenMediaSettingCallback)
			resourcedata.SetNillableValueWithInterfaceArrayWithFunc(d, "media_settings_chat", currentQueue.MediaSettings.Chat, flattenMediaSetting)
			resourcedata.SetNillableValueWithInterfaceArrayWithFunc(d, "media_settings_email", currentQueue.MediaSettings.Email, flattenMediaSetting)
			resourcedata.SetNillableValueWithInterfaceArrayWithFunc(d, "media_settings_message", currentQueue.MediaSettings.Message, flattenMediaSetting)
		}

		if currentQueue.AgentOwnedRouting != nil {
			resourcedata.SetNillableValueWithInterfaceArrayWithFunc(d, "agent_owned_routing", currentQueue.AgentOwnedRouting, flattenAgentOwnedRouting)
		}

		resourcedata.SetNillableValueWithInterfaceArrayWithFunc(d, "routing_rules", currentQueue.RoutingRules, flattenRoutingRules)

		if currentQueue.Bullseye != nil {
			resourcedata.SetNillableValueWithInterfaceArrayWithFunc(d, "bullseye_rings", currentQueue.Bullseye.Rings, flattenBullseyeRings)
		}

		resourcedata.SetNillableReference(d, "queue_flow_id", currentQueue.QueueFlow)
		resourcedata.SetNillableReference(d, "message_in_queue_flow_id", currentQueue.MessageInQueueFlow)
		resourcedata.SetNillableReference(d, "email_in_queue_flow_id", currentQueue.EmailInQueueFlow)
		resourcedata.SetNillableReference(d, "whisper_prompt_id", currentQueue.WhisperPrompt)
		resourcedata.SetNillableReference(d, "on_hold_prompt_id", currentQueue.OnHoldPrompt)
		resourcedata.SetNillableValue(d, "auto_answer_only", currentQueue.AutoAnswerOnly)
		resourcedata.SetNillableValue(d, "enable_transcription", currentQueue.EnableTranscription)
		resourcedata.SetNillableValue(d, "suppress_in_queue_call_recording", currentQueue.SuppressInQueueCallRecording)
		resourcedata.SetNillableValue(d, "enable_audio_monitoring", currentQueue.EnableAudioMonitoring)
		resourcedata.SetNillableValue(d, "enable_manual_assignment", currentQueue.EnableManualAssignment)
		resourcedata.SetNillableValue(d, "calling_party_name", currentQueue.CallingPartyName)
		resourcedata.SetNillableValue(d, "calling_party_number", currentQueue.CallingPartyNumber)
		resourcedata.SetNillableValue(d, "scoring_method", currentQueue.ScoringMethod)

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

		resourcedata.SetNillableValueWithInterfaceArrayWithFunc(d, "direct_routing", currentQueue.DirectRouting, flattenDirectRouting)

		wrapupCodes, err := flattenQueueWrapupCodes(ctx, d.Id(), proxy)
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

		if exists := featureToggles.CSGToggleExists(); !exists {
			_ = d.Set("conditional_group_routing_rules", flattenConditionalGroupRoutingRules(currentQueue))
		} else {
			log.Printf("%s is set, not reading conditional_group_routing_rules attribute in routing_queue %s resource", featureToggles.CSGToggleName(), d.Id())
		}

		if exists := featureToggles.OEAToggleExists(); !exists {
			if currentQueue.OutboundEmailAddress != nil && *currentQueue.OutboundEmailAddress != nil {
				outboundEmailAddress := *currentQueue.OutboundEmailAddress
				_ = d.Set("outbound_email_address", []interface{}{FlattenQueueEmailAddress(*outboundEmailAddress)})
			} else {
				_ = d.Set("outbound_email_address", nil)
			}
		} else {
			log.Printf("%s is set, not reading outbound_email_address attribute in routing_queue %s resource", featureToggles.OEAToggleName(), d.Id())
		}

		log.Printf("Done reading queue %s %s", d.Id(), *currentQueue.Name)
		return cc.CheckState(d)
	})
}

func updateQueue(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	routingAPI := platformclientv2.NewRoutingApiWithConfig(sdkConfig)
	scoringMethod := d.Get("scoring_method").(string)
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
	}

	diagErr := addCGRAndOEA(routingAPI, d, &updateQueue)
	if diagErr != nil {
		return diagErr
	}

	log.Printf("Updating queue %s", *updateQueue.Name)

	if scoringMethod != "" {
		updateQueue.ScoringMethod = &scoringMethod
	}

	_, resp, err := routingAPI.PutRoutingQueue(d.Id(), updateQueue)
	if err != nil {
		return util.BuildAPIDiagnosticError(resourceName, fmt.Sprintf("Failed to update queue %s error: %s", *updateQueue.Name, err), resp)
	}

	diagErr = util.UpdateObjectDivision(d, "QUEUE", sdkConfig)
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

/*
DEVTOOLING-751: If conditional group routing rules and outbound email address are managed by their independent resource
they are being removed when the parent queue is updated since the update body does not contain them.
If the independent resources are enabled, pass in the current OEA and/or CGR to the update queue so they are not removed
*/
func addCGRAndOEA(routingAPI *platformclientv2.RoutingApi, d *schema.ResourceData, queue *platformclientv2.Queuerequest) diag.Diagnostics {
	currentQueue, resp, err := routingAPI.GetRoutingQueue(d.Id())
	if err != nil {
		return util.BuildAPIDiagnosticError(resourceName, fmt.Sprintf("Failed to get queue %s for update, error: %s", *queue.Name, err), resp)
	}

	if exists := featureToggles.CSGToggleExists(); !exists {
		conditionalGroupRouting, diagErr := buildSdkConditionalGroupRouting(d)
		if diagErr != nil {
			return diagErr
		}
		queue.ConditionalGroupRouting = conditionalGroupRouting
	} else {
		log.Printf("%s is set, not updating conditional_group_routing_rules attribute in routing_queue %s resource", featureToggles.CSGToggleName(), d.Id())
		queue.ConditionalGroupRouting = currentQueue.ConditionalGroupRouting

		// remove queue_id from first CGR rule to avoid api error
		if len(*queue.ConditionalGroupRouting.Rules) > 0 {
			(*queue.ConditionalGroupRouting.Rules)[0].Queue = nil
		}
	}

	if exists := featureToggles.OEAToggleExists(); !exists {
		queue.OutboundEmailAddress = buildSdkQueueEmailAddress(d)
	} else {
		log.Printf("%s is set, not updating outbound_email_address attribute in routing_queue %s resource", featureToggles.OEAToggleName(), d.Id())
		queue.OutboundEmailAddress = *currentQueue.OutboundEmailAddress
	}

	return nil
}

func deleteQueue(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)

	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	routingAPI := platformclientv2.NewRoutingApiWithConfig(sdkConfig)

	log.Printf("Deleting queue %s", name)
	resp, err := routingAPI.DeleteRoutingQueue(d.Id(), true)
	if err != nil {
		return util.BuildAPIDiagnosticError(resourceName, fmt.Sprintf("Failed to delete queue %s error: %s", name, err), resp)
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
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(resourceName, fmt.Sprintf("Error deleting queue %s | error: %s", d.Id(), err), resp))
		}
		return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(resourceName, fmt.Sprintf("Queue %s still exists", d.Id()), resp))
	})
}

func constructAgentOwnedRouting(d *schema.ResourceData) *platformclientv2.Agentownedrouting {
	if agentOwnedRouting, ok := d.Get("agent_owned_routing").([]interface{}); ok {
		if agentOwnedRouting != nil && len(agentOwnedRouting) > 0 {
			return buildAgentOwnedRouting(agentOwnedRouting)
		}
	}
	return nil
}

func validateMapCommTypes(val interface{}, _ cty.Path) diag.Diagnostics {
	if val == nil {
		return nil
	}

	commTypes := []string{"CALL", "CALLBACK", "CHAT", "COBROWSE", "EMAIL", "MESSAGE", "SOCIAL_EXPRESSION", "VIDEO", "SCREENSHARE"}
	m := val.(map[string]interface{})
	for k := range m {
		if !lists.ItemInSlice(k, commTypes) {
			return util.BuildDiagnosticError(resourceName, fmt.Sprintf("%s is an invalid communication type key.", k), fmt.Errorf("invalid communication type key"))
		}
	}
	return nil
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
						return util.BuildAPIDiagnosticError(resourceName, fmt.Sprintf("Failed to remove wrapup codes for queue %s error: %s", d.Id(), err), resp)
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
			_, resp, err := api.PostRoutingQueueWrapupcodes(queueID, updateChunk)
			if err != nil {
				return util.BuildAPIDiagnosticError(resourceName, fmt.Sprintf("Failed to update wrapup codes for queue %s error: %s", queueID, err), resp)
			}
		}
	}
	return nil
}

func getRoutingQueueWrapupCodes(queueID string, api *platformclientv2.RoutingApi) ([]platformclientv2.Wrapupcode, diag.Diagnostics) {
	const maxPageSize = 100

	var codes []platformclientv2.Wrapupcode
	for pageNum := 1; ; pageNum++ {
		codeResult, resp, err := api.GetRoutingQueueWrapupcodes(queueID, maxPageSize, pageNum)
		if err != nil {
			return nil, util.BuildAPIDiagnosticError(resourceName, fmt.Sprintf("Failed to query wrapup codes for queue %s error: %s", queueID, err), resp)
		}
		if codeResult == nil || codeResult.Entities == nil || len(*codeResult.Entities) == 0 {
			return codes, nil
		}
		codes = append(codes, *codeResult.Entities...)

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

		members, diagErr := getRoutingQueueMembers(d.Id(), "group", sdkConfig)
		if diagErr != nil {
			return diagErr
		}

		for _, userId := range newUserIds {
			if err := verifyUserIsNotGroupMemberOfQueue(d.Id(), userId, members); err != nil {
				return util.BuildDiagnosticError(resourceName, "failed to update queue member: ", err)
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
func verifyUserIsNotGroupMemberOfQueue(queueId, userId string, members []platformclientv2.Queuemember) error {
	log.Printf("verifying that member '%s' is not assinged to the queue '%s' via a group", userId, queueId)

	for _, member := range members {
		if *member.Id == userId {
			return fmt.Errorf("member %s  is already assigned to queue %s via a group, and therefore should not be assigned as a member", userId, queueId)
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
			resp, err := api.PostRoutingQueueMembers(queueID, chunk, remove)
			if err != nil {
				return util.BuildAPIDiagnosticError(resourceName, fmt.Sprintf("Failed to update members in queue %s error: %s", queueID, err), resp)
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
	resp, err := api.PatchRoutingQueueMember(queueID, userID, platformclientv2.Queuemember{
		Id:         &userID,
		RingNumber: &ringNum,
	})
	if err != nil {
		return util.BuildAPIDiagnosticError(resourceName, fmt.Sprintf("Failed to update ring number for queue %s user %s error: %s", queueID, userID, err), resp)
	}
	return nil
}

func getRoutingQueueMembers(queueID string, memberBy string, sdkConfig *platformclientv2.Configuration) ([]platformclientv2.Queuemember, diag.Diagnostics) {
	var members []platformclientv2.Queuemember
	const pageSize = 100
	api := platformclientv2.NewRoutingApiWithConfig(sdkConfig)

	// Need to call this method to find the member count for a queue. GetRoutingQueueMembers does not return a `total` property for us to use.
	queue, resp, err := api.GetRoutingQueue(queueID)
	if err != nil {
		return nil, util.BuildAPIDiagnosticError(resourceName, fmt.Sprintf("Failed to find queue %s error: %s", queueID, err), resp)
	}
	queueMembers := *queue.MemberCount
	log.Printf("%d members belong to queue %s", queueMembers, queueID)

	for pageNum := 1; ; pageNum++ {
		users, resp, err := sdkGetRoutingQueueMembers(queueID, memberBy, pageNum, pageSize, api)
		if err != nil || resp.StatusCode != http.StatusOK {
			return nil, util.BuildAPIDiagnosticError(resourceName, fmt.Sprintf("Failed to query users for queue %s error: %s", queueID, err), resp)
		}
		if users == nil || users.Entities == nil || len(*users.Entities) == 0 {
			membersFound := len(members)
			log.Printf("%d queue members found for queue %s", membersFound, queueID)

			if membersFound != queueMembers {
				log.Printf("Member count is not equal to queue member found for queue %s, Correlation Id: %s", queueID, resp.CorrelationID)
			}
			return members, nil
		}

		members = append(members, *users.Entities...)
	}
}

func sdkGetRoutingQueueMembers(queueID, memberBy string, pageNumber, pageSize int, api *platformclientv2.RoutingApi) (*platformclientv2.Queuememberentitylisting, *platformclientv2.APIResponse, error) {
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
