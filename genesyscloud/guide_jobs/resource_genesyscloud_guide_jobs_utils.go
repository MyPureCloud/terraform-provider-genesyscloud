package guide_jobs

import (
	"log"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func buildGuideJobFromResourceData(d *schema.ResourceData) GenerateGuideContentRequest {
	log.Printf("Building guide job from resource data")
	guideJobReq := GenerateGuideContentRequest{}

	description := d.Get("description").(string)
	if description != "" {
		guideJobReq.Description = &description
	}

	url := d.Get("url").(string)
	if url != "" {
		guideJobReq.Url = &url
	}

	log.Printf("Successfully built guide job from resource data")
	return guideJobReq
}

func setRequestHeader(r *http.Request, p *guideJobsProxy) *http.Request {
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("Accept", "application/json")
	r.Header.Set("Authorization", "Bearer "+p.clientConfig.AccessToken)
	return r
}

func flattenGuideContent(guideContent *GeneratedGuideContent) *[]map[string]interface{} {
	if guideContent == nil {
		return nil
	}

	guideContentMap := make(map[string]interface{})

	if guideContent.Instruction != "" {
		guideContentMap["instruction"] = guideContent.Instruction
	}
	if guideContent.Variables != nil {
		guideContentMap["variables"] = flattenVariables(guideContent.Variables)
	}
	if guideContent.Resources.DataActions != nil {
		guideContentMap["resources"] = flattenResources(guideContent.Resources)
	}

	result := []map[string]interface{}{guideContentMap}
	return &result
}

func flattenVariables(variables []Variable) []map[string]interface{} {
	if variables == nil {
		return nil
	}

	var variablesList []map[string]interface{}
	for _, variable := range variables {
		variableMap := make(map[string]interface{})
		if variable.Name != "" {
			variableMap["name"] = variable.Name
		}
		if variable.Type != "" {
			variableMap["type"] = variable.Type
		}
		if variable.Scope != "" {
			variableMap["scope"] = variable.Scope
		}
		if variable.Description != "" {
			variableMap["description"] = variable.Description
		}
		variablesList = append(variablesList, variableMap)
	}
	return variablesList
}

func flattenResources(resources GuideVersionResources) []map[string]interface{} {
	if resources.DataActions == nil {
		return nil
	}

	var dataActionsList []map[string]interface{}
	for _, dataAction := range resources.DataActions {
		dataActionMap := make(map[string]interface{})
		if dataAction.ID != "" {
			dataActionMap["data_action_id"] = dataAction.ID
		}
		if dataAction.Label != "" {
			dataActionMap["label"] = dataAction.Label
		}
		if dataAction.Description != "" {
			dataActionMap["description"] = dataAction.Description
		}
		dataActionsList = append(dataActionsList, dataActionMap)
	}

	return []map[string]interface{}{
		{
			"data_action": dataActionsList,
		},
	}
}

// structs

type GuideJob struct {
	Id           string       `json:"id,omitempty"`
	Guide        Guide        `json:"guide,omitempty"`
	Status       string       `json:"status,omitempty"`
	Errors       []ErrorBody  `json:"errors,omitempty"`
	GuideContent GuideContent `json:"guideContent,omitempty"`
	SelfUri      string       `json:"selfUri,omitempty"`
}

type ErrorBody struct {
	Message           string `json:"message,omitempty"`
	Code              string `json:"code,omitempty"`
	Status            int    `json:"status,omitempty"`
	EntityId          string `json:"entityId,omitempty"`
	EntityName        string `json:"entityName,omitempty"`
	MessageWithParams string `json:"messageWithParams,omitempty"`
}

type GuideContent struct {
	Instruction string                `json:"instruction,omitempty"`
	Variables   []Variable            `json:"variables,omitempty"`
	Resources   GuideVersionResources `json:"resources,omitempty"`
}

type Variable struct {
	Name        string `json:"name,omitempty"`
	Type        string `json:"type,omitempty"`
	Scope       string `json:"scope,omitempty"`
	Description string `json:"description,omitempty"`
}

type DataAction struct {
	ID          string `json:"id,omitempty"`
	Label       string `json:"label,omitempty"`
	Description string `json:"description,omitempty"`
}

type GuideVersionResources struct {
	DataActions []DataAction `json:"dataActions,omitempty"`
}

type Guide struct {
	Id      string `json:"id,omitempty"`
	SelfUri string `json:"selfUri,omitempty"`
}

type GeneratedGuideContent struct {
	Instruction string                `json:"instruction,omitempty"`
	Variables   []Variable            `json:"variables,omitempty"`
	Resources   GuideVersionResources `json:"resources,omitempty"`
}

type JobResponse struct {
	Id           *string                `json:"id,omitempty"`
	GuideId      *string                `json:"guideId,omitempty"`
	Status       *string                `json:"status,omitempty"`
	GuideContent *GeneratedGuideContent `json:"guideContent,omitempty"`
	Errors       []ErrorBody            `json:"errors,omitempty"`
}
