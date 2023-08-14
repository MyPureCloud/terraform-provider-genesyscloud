//go:build unit
// +build unit

package process_automation_trigger

import (
	"testing"
)

func TestProcessAutomationJSONParsing(t *testing.T) {
	matchCriteriaStr := `
		[
			{
			"jsonPath" : "mediaType",
			"operator" : "Equal",
			"value" : "CHAT"
			}
		] 
  `
	///This was the expected target JSON
	targetJson := "{\"delayBySeconds\":1000,\"description\":\"My Sample Topic\",\"enabled\":false,\"eventTTLSeconds\":1000,\"id\":\"888923de-8d4c-1000-2001-48f526dc3333\",\"matchCriteria\":[{\"jsonPath\":\"mediaType\",\"operator\":\"Equal\",\"value\":\"CHAT\"}],\"name\":\"Terraform trigger1-0156463b-4b13-434e-a15c-41492e60ad47\",\"target\":{\"id\":\"173223de-8d4c-0000-0001-48f526dc3333\",\"type\":\"Workflow\",\"workflowTargetSettings:{\"dataFormat\":\"Json\"}},\"topicName\":\"v2.detail.events.conversation.{id}.customer.end\",\"version\":1}"

	id := "888923de-8d4c-1000-2001-48f526dc3333"
	topicName := "v2.detail.events.conversation.{id}.customer.end"
	name := "Terraform trigger1-0156463b-4b13-434e-a15c-41492e60ad47"
	targetId := "173223de-8d4c-0000-0001-48f526dc3333"
	targetWorkflow := "Workflow"
	workflowTargetSettingsDataFormat := "Json"
	enabled := false
	eventTTLSeconds := 1000
	delayBySeconds := 1000
	version := 1
	description := "My Sample Topic"

	pat := &ProcessAutomationTrigger{
		Id:        &id,
		TopicName: &topicName,
		Name:      &name,
		Target: &Target{
			Id:   &targetId,
			Type: &targetWorkflow,
			WorkflowTargetSettings: &WorkflowTargetSettings{
				DataFormat: &workflowTargetSettingsDataFormat,
			},
		},
		MatchCriteria:   &matchCriteriaStr,
		Enabled:         &enabled,
		EventTTLSeconds: &eventTTLSeconds,
		DelayBySeconds:  &delayBySeconds,
		Version:         &version,
		Description:     &description,
	}

	x, err := pat.toJSONString()

	if x != targetJson {
		t.Errorf("The produced JSON %s does not match expected JSON '%s'", x, targetJson)
	}

	if err != nil {
		t.Errorf("Expected error to be nil, got '%v'", err)
	}

}
