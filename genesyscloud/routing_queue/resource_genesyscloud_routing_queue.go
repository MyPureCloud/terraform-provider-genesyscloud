package routing_queue

import (
	"context"
	"fmt"
	"log"
	"net/http"
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
	"github.com/mypurecloud/platform-client-sdk-go/v146/platformclientv2"
)

var bullseyeExpansionTypeTimeout = "TIMEOUT_SECONDS"

func getAllRoutingQueues(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(resourceExporter.ResourceIDMetaMap)
	proxy := GetRoutingQueueProxy(clientConfig)

	// Newly created resources often aren't returned unless there's a delay
	time.Sleep(5 * time.Second)

	queues, resp, err := proxy.GetAllRoutingQueues(ctx, "")
	if err != nil {
		return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("failed to get routing queues: %s", err), resp)
	}

	if queues == nil || len(*queues) == 0 {
		return resources, nil
	}

	for _, queue := range *queues {
		resources[*queue.Id] = &resourceExporter.ResourceMeta{BlockLabel: *queue.Name}
	}

	return resources, nil
}

func createRoutingQueue(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := GetRoutingQueueProxy(sdkConfig)

	divisionID := d.Get("division_id").(string)
	scoringMethod := d.Get("scoring_method").(string)
	peerId := d.Get("peer_id").(string)
	sourceQueueId := d.Get("source_queue_id").(string)
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
	if peerId != "" {
		createQueue.PeerId = &peerId
	}
	if sourceQueueId != "" {
		createQueue.SourceQueueId = &sourceQueueId
	}

	log.Printf("Creating Routing Queue %s", *createQueue.Name)

	queue, resp, err := proxy.createRoutingQueue(ctx, &createQueue)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to create queue %s | error: %s", *createQueue.Name, err), resp)
	}
	if resp.StatusCode != http.StatusOK {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to create queue %s with error: %s, status code %v", *createQueue.Name, err, resp.StatusCode), resp)
	}

	d.SetId(*queue.Id)

	diagErr := updateQueueMembers(d, sdkConfig)
	if diagErr != nil {
		return diagErr
	}

	diagErr = updateQueueWrapupCodes(d, sdkConfig)
	if diagErr != nil {
		return diagErr
	}

	log.Printf("Created Routing Queue %s", d.Id())
	return readRoutingQueue(ctx, d, meta)
}

func readRoutingQueue(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := GetRoutingQueueProxy(sdkConfig)
	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceRoutingQueue(), constants.ConsistencyChecks(), ResourceType)

	log.Printf("Reading queue %s", d.Id())

	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		currentQueue, resp, getErr := proxy.getRoutingQueueById(ctx, d.Id(), true)
		if getErr != nil {
			if util.IsStatus404(resp) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read queue %s | error: %s", d.Id(), getErr), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read queue %s | error: %s", d.Id(), getErr), resp))
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
			resourcedata.SetNillableValueWithInterfaceArrayWithFunc(d, "media_settings_email", currentQueue.MediaSettings.Email, flattenMediaEmailSetting)
			resourcedata.SetNillableValueWithInterfaceArrayWithFunc(d, "media_settings_message", currentQueue.MediaSettings.Message, flattenMediaSetting)
		}

		_ = d.Set("outbound_messaging_sms_address_id", nil)
		_ = d.Set("outbound_messaging_whatsapp_recipient_id", nil)
		_ = d.Set("outbound_messaging_open_messaging_recipient_id", nil)

		if currentQueue.OutboundMessagingAddresses != nil {
			if currentQueue.OutboundMessagingAddresses.SmsAddress != nil {
				_ = d.Set("outbound_messaging_sms_address_id", *currentQueue.OutboundMessagingAddresses.SmsAddress.Id)
			}
			if currentQueue.OutboundMessagingAddresses.WhatsAppRecipient != nil {
				_ = d.Set("outbound_messaging_whatsapp_recipient_id", *currentQueue.OutboundMessagingAddresses.WhatsAppRecipient.Id)
			}
			if currentQueue.OutboundMessagingAddresses.OpenMessagingRecipient != nil {
				_ = d.Set("outbound_messaging_open_messaging_recipient_id", *currentQueue.OutboundMessagingAddresses.OpenMessagingRecipient.Id)
			}
		}

		if currentQueue.AgentOwnedRouting != nil {
			resourcedata.SetNillableValueWithInterfaceArrayWithFunc(d, "agent_owned_routing", currentQueue.AgentOwnedRouting, flattenAgentOwnedRouting)
		}

		if currentQueue.Bullseye != nil {
			resourcedata.SetNillableValueWithInterfaceArrayWithFunc(d, "bullseye_rings", currentQueue.Bullseye.Rings, flattenBullseyeRings)
		}

		resourcedata.SetNillableValueWithInterfaceArrayWithFunc(d, "routing_rules", currentQueue.RoutingRules, flattenRoutingRules)
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
		resourcedata.SetNillableValue(d, "peer_id", currentQueue.PeerId)
		resourcedata.SetNillableValueWithInterfaceArrayWithFunc(d, "direct_routing", currentQueue.DirectRouting, flattenDirectRouting)

		if currentQueue.DefaultScripts != nil {
			_ = d.Set("default_script_ids", flattenDefaultScripts(*currentQueue.DefaultScripts))
		} else {
			_ = d.Set("default_script_ids", nil)
		}

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

		log.Printf("Read queue %s %s", d.Id(), *currentQueue.Name)
		return cc.CheckState(d)
	})
}

func updateRoutingQueue(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := GetRoutingQueueProxy(sdkConfig)

	scoringMethod := d.Get("scoring_method").(string)
	skillGroups := buildMemberGroupList(d, "skill_groups", "SKILLGROUP")
	groups := buildMemberGroupList(d, "groups", "GROUP")
	teams := buildMemberGroupList(d, "teams", "TEAM")
	memberGroups := append(*skillGroups, *groups...)
	memberGroups = append(memberGroups, *teams...)
	peerId := d.Get("peer_id").(string)

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

	diagErr := addCGRAndOEA(proxy, d, &updateQueue)
	if diagErr != nil {
		return diagErr
	}

	if scoringMethod != "" {
		updateQueue.ScoringMethod = &scoringMethod
	}
	if peerId != "" {
		updateQueue.PeerId = &peerId
	}

	log.Printf("Updating queue %s", *updateQueue.Name)

	_, resp, err := proxy.updateRoutingQueue(ctx, d.Id(), &updateQueue)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to update queue %s error: %s", *updateQueue.Name, err), resp)
	}

	diagErr = util.UpdateObjectDivision(d, "QUEUE", sdkConfig)
	if diagErr != nil {
		return diagErr
	}

	diagErr = updateQueueMembers(d, sdkConfig)
	if diagErr != nil {
		return diagErr
	}

	diagErr = updateQueueWrapupCodes(d, sdkConfig)
	if diagErr != nil {
		return diagErr
	}

	log.Printf("Updated queue %s", *updateQueue.Name)
	return readRoutingQueue(ctx, d, meta)
}

/*
DEVTOOLING-751: If conditional group routing rules and outbound email address are managed by their independent resource
they are being removed when the parent queue is updated since the update body does not contain them.
If the independent resources are enabled, pass in the current OEA and/or CGR to the update queue so they are not removed
*/
func addCGRAndOEA(proxy *RoutingQueueProxy, d *schema.ResourceData, queue *platformclientv2.Queuerequest) diag.Diagnostics {
	currentQueue, resp, err := proxy.getRoutingQueueById(ctx, d.Id(), true)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get queue %s for update, error: %s", *queue.Name, err), resp)
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
		if queue.ConditionalGroupRouting != nil && len(*queue.ConditionalGroupRouting.Rules) > 0 {
			(*queue.ConditionalGroupRouting.Rules)[0].Queue = nil
		}
	}

	if exists := featureToggles.OEAToggleExists(); !exists {
		queue.OutboundEmailAddress = buildSdkQueueEmailAddress(d)
	} else {
		log.Printf("%s is set, not updating outbound_email_address attribute in routing_queue %s resource", featureToggles.OEAToggleName(), d.Id())

		if currentQueue.OutboundEmailAddress != nil {
			queue.OutboundEmailAddress = *currentQueue.OutboundEmailAddress
		}
	}

	return nil
}

func deleteRoutingQueue(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := GetRoutingQueueProxy(sdkConfig)

	log.Printf("Deleting queue %s", name)
	resp, err := proxy.deleteRoutingQueue(ctx, d.Id(), true)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to delete queue %s error: %s", name, err), resp)
	}

	// Queue deletes are not immediate. Query until queue is no longer found
	// Add a delay before the first request to reduce the likelihood of public API's cache
	// re-populating the queue after the delete. Otherwise it may not expire for a minute.
	time.Sleep(5 * time.Second)

	//DEVTOOLING-238- Increasing this to a 120 seconds to see if we can temporarily mitigate a problem for a customer
	return util.WithRetries(ctx, 120*time.Second, func() *retry.RetryError {
		_, resp, err := proxy.getRoutingQueueById(ctx, d.Id(), false)
		if err != nil {
			if util.IsStatus404(resp) {
				// Queue deleted
				log.Printf("Queue %s deleted", name)
				return nil
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Error deleting queue %s | error: %s", d.Id(), err), resp))
		}
		return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Queue %s still exists", d.Id()), resp))
	})
}

func createRoutingQueueWrapupCodes(queueID string, codesToAdd []string, sdkConfig *platformclientv2.Configuration) diag.Diagnostics {
	proxy := GetRoutingQueueProxy(sdkConfig)
	// API restricts wrapup code adds to 100 per call
	if len(codesToAdd) > 0 {
		chunks := chunksProcess.ChunkItems(codesToAdd, platformWrapupCodeReferenceFunc, 100)

		chunkProcessor := func(chunk []platformclientv2.Wrapupcodereference) diag.Diagnostics {
			_, resp, err := proxy.createRoutingQueueWrapupCode(ctx, queueID, chunk)
			if err != nil {
				return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to update wrapup codes for queue %s error: %s", queueID, err), resp)
			}
			return nil
		}
		return chunksProcess.ProcessChunks(chunks, chunkProcessor)
	}

	return nil
}

func updateQueueWrapupCodes(d *schema.ResourceData, sdkConfig *platformclientv2.Configuration) diag.Diagnostics {
	proxy := GetRoutingQueueProxy(sdkConfig)

	if d.HasChange("wrapup_codes") {
		log.Printf("Updating Routing Queue WrapupCodes")

		if codesConfig := d.Get("wrapup_codes"); codesConfig != nil {
			// Get existing codes
			codes, resp, err := proxy.getAllRoutingQueueWrapupCodes(ctx, d.Id())
			if err != nil {
				return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to query wrapup codes for queue %s error: %s", d.Id(), err), resp)
			}

			existingCodes := getWrapupCodeIds(codes)
			configCodes := *lists.SetToStringList(codesConfig.(*schema.Set))
			codesToRemove := lists.SliceDifference(existingCodes, configCodes)

			// Remove Wrapup Codes
			if len(codesToRemove) > 0 {
				for _, codeId := range codesToRemove {
					resp, err := proxy.deleteRoutingQueueWrapupCode(ctx, d.Id(), codeId)
					if err != nil {
						if util.IsStatus404(resp) {
							// Ignore missing queue or wrapup code
							continue
						}
						return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to remove wrapup codes for queue %s error: %s", d.Id(), err), resp)
					}
				}
			}

			// Add Wrapup Codes
			codesToAdd := lists.SliceDifference(configCodes, existingCodes)
			if len(codesToAdd) > 0 {
				err := createRoutingQueueWrapupCodes(d.Id(), codesToAdd, sdkConfig)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func getWrapupCodeIds(codes *[]platformclientv2.Wrapupcode) []string {
	var wrapupCodes []string
	if codes != nil {
		for _, code := range *codes {
			wrapupCodes = append(wrapupCodes, *code.Id)
		}
	}
	return wrapupCodes
}

func validateMapCommTypes(val interface{}, _ cty.Path) diag.Diagnostics {
	if val == nil {
		return nil
	}

	commTypes := []string{"CALL", "CALLBACK", "CHAT", "COBROWSE", "EMAIL", "MESSAGE", "SOCIAL_EXPRESSION", "VIDEO", "SCREENSHARE"}
	m := val.(map[string]interface{})
	for k := range m {
		if !lists.ItemInSlice(k, commTypes) {
			return util.BuildDiagnosticError(ResourceType, fmt.Sprintf("%s is an invalid communication type key.", k), fmt.Errorf("invalid communication type key"))
		}
	}
	return nil
}

func platformWritableEntityFunc(val string) platformclientv2.Writableentity {
	return platformclientv2.Writableentity{Id: &val}
}

func platformWrapupCodeReferenceFunc(val string) platformclientv2.Wrapupcodereference {
	return platformclientv2.Wrapupcodereference{Id: &val}
}
