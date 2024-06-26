package process_automation_trigger

import (
	"encoding/json"

	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

type ProcessAutomationTrigger struct {
	Id              *string `json:"id,omitempty"`
	TopicName       *string `json:"topicName,omitempty"`
	Name            *string `json:"name,omitempty"`
	Target          *Target `json:"target,omitempty"`
	MatchCriteria   *string `json:"-"`
	Enabled         *bool   `json:"enabled,omitempty"`
	EventTTLSeconds *int    `json:"eventTTLSeconds,omitempty"`
	DelayBySeconds  *int    `json:"delayBySeconds,omitempty"`
	Version         *int    `json:"version,omitempty"`
	Description     *string `json:"description,omitempty"`
}

type WorkflowTargetSettings struct {
	DataFormat *string `json:"dataFormat,omitempty"`
}

type Target struct {
	Type                   *string                 `json:"type,omitempty"`
	Id                     *string                 `json:"id,omitempty"`
	WorkflowTargetSettings *WorkflowTargetSettings `json:"workflowTargetSettings,omitempty"`
}

func (p *ProcessAutomationTrigger) toJSONString() (string, error) {
	//Step #1: Converting the process automation trigger to a JSON byte arrays
	b, err := json.Marshal(p)
	if err != nil {
		return "", err
	}
	patJson := string(b)

	//Step #2: Converting the JSON string to a Golang Map
	var patMap map[string]interface{}
	err = json.Unmarshal([]byte(patJson), &patMap)
	if err != nil {
		return "", err
	}

	//Step #3: Converting the MatchCriteria field from a string to Map
	var data []map[string]interface{}
	err = json.Unmarshal([]byte(*p.MatchCriteria), &data)
	if err != nil {
		return "", err
	}

	matchCriteriaArray := make([]interface{}, len(data))
	for i, obj := range data {
		value := make(map[string]interface{})

		value["jsonPath"] = obj["jsonPath"]
		value["operator"] = obj["operator"]
		value["value"] = obj["value"]
		value["values"] = obj["values"]

		matchCriteriaArray[i] = value
	}

	//Step #4: Merging the match criteria array into the main map
	patMap["matchCriteria"] = matchCriteriaArray

	//Step #5: Converting the merged Map into a JSON string
	finalJsonBytes, err := json.Marshal(patMap)
	if err != nil {
		return "", err
	}

	finalPAT := string(finalJsonBytes)
	return finalPAT, nil
}

// Constructor that will take an platform client response object and build a new ProcessAutomationTrigger from it
func NewProcessAutomationFromPayload(response *platformclientv2.APIResponse) (*ProcessAutomationTrigger, error) {
	httpPayload := response.RawBody
	pat := &ProcessAutomationTrigger{}
	patMap := make(map[string]interface{})
	err := json.Unmarshal(httpPayload, &patMap)
	if err != nil {
		return nil, err
	}

	matchCriteria := patMap["matchCriteria"]
	matchCriteriaBytes, err := json.Marshal(matchCriteria)
	matchCriteriaStr := string(matchCriteriaBytes)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(httpPayload, &pat)
	if err != nil {
		return nil, err
	}
	pat.MatchCriteria = &matchCriteriaStr

	return pat, nil
}
