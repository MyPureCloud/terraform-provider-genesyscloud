package routing_queue

import (
	"fmt"
	"strings"
)

func GenerateRoutingQueueResource(
	resourceID string,
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
	enableManualAssignment string,
	scoringMethod string,
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
        suppress_in_queue_call_recording = %s
  		enable_manual_assignment = %s
		%s
	}
	`, resourceID,
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
		suppressInQueueCallRecording,
		enableManualAssignment,
		strings.Join(nestedBlocks, "\n"))
}

func GenerateRoutingQueueResourceBasic(resourceID string, name string, nestedBlocks ...string) string {
	return fmt.Sprintf(`resource "genesyscloud_routing_queue" "%s" {
		name = "%s"
		%s
	}
	`, resourceID, name, strings.Join(nestedBlocks, "\n"))
}

// GenerateRoutingQueueResourceBasicWithDepends Used when testing skills group dependencies.
func GenerateRoutingQueueResourceBasicWithDepends(resourceID string, dependsOn string, name string, nestedBlocks ...string) string {
	return fmt.Sprintf(`resource "genesyscloud_routing_queue" "%s" {
		depends_on = [%s]
		name = "%s"
		%s
	}
	`, resourceID, dependsOn, name, strings.Join(nestedBlocks, "\n"))
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
