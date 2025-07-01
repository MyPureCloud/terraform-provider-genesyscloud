package guide_version

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"log"
	"net/http"
	"strings"
)

func parseId(id string) (string, string, error) {
	ids := strings.Split(id, "/")
	if len(ids) != 2 {
		return "", "", fmt.Errorf("invalid resource ID format: %s", id)
	}

	if ids[0] == "" || ids[1] == "" {
		return "", "", fmt.Errorf("invalid resource ID format: %s", id)
	}

	guideId := ids[0]
	versionId := ids[1]

	return guideId, versionId, nil
}

func buildRequestHeader(r *http.Request, p *guideVersionProxy) *http.Request {
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("Authorization", "Bearer "+p.clientConfig.AccessToken)
	return r
}

func buildGuideVersionFromResourceData(d *schema.ResourceData) *CreateGuideVersionRequest {
	log.Printf("Building Guide Version from Resource Data")

	guideVersion := &CreateGuideVersionRequest{
		GuideID:     d.Get("guide_id").(string),
		Instruction: d.Get("instruction").(string),
	}

	variables := d.Get("variables").([]interface{})
	if variables != nil && len(variables) != 0 {
		guideVersion.Variables = buildGuideVersionVariables(variables)
	}

	resources := d.Get("resources").([]interface{})
	if resources != nil && len(resources) != 0 {
		guideVersion.Resources = buildGuideVersionResources(resources)
	}

	log.Printf("Succesfully Built Guide Version from Resource Data")
	return guideVersion
}

func buildGuideVersionResources(resourcesList []interface{}) GuideVersionResources {
	log.Printf("Building Resource Attributes for Guide")
	resourcesMap := resourcesList[0].(map[string]interface{})

	var dataActions []DataAction
	if dataActionsList, ok := resourcesMap["data_action"].([]interface{}); ok {
		for _, v := range dataActionsList {
			dataActionMap := v.(map[string]interface{})
			dataAction := DataAction{
				ID:    dataActionMap["data_action_id"].(string),
				Label: dataActionMap["label"].(string),
			}

			if description, ok := dataActionMap["description"].(string); ok && description != "" {
				dataAction.Description = description
			}

			dataActions = append(dataActions, dataAction)
		}
	}

	log.Printf("Succesfully Built Resource Attributes for Guide")
	return GuideVersionResources{
		DataActions: dataActions,
	}
}

func buildGuideVersionVariables(vars []interface{}) []Variable {
	variables := make([]Variable, len(vars))

	for i, v := range vars {
		variables[i] = Variable{
			Name:  v.(map[string]interface{})["name"].(string),
			Type:  v.(map[string]interface{})["type"].(string),
			Scope: v.(map[string]interface{})["scope"].(string),
		}

		if description := v.(map[string]interface{})["description"].(string); description != "" {
			variables[i].Description = description
		}
	}

	return variables
}

func buildGuideVersionForUpdate(d *schema.ResourceData) *UpdateGuideVersion {
	log.Printf("Building Guide Version from Resource Data")

	guideVersion := &UpdateGuideVersion{
		GuideID:     d.Get("guide_id").(string),
		Instruction: d.Get("instruction").(string),
	}

	if vars := d.Get("variables").([]interface{}); vars != nil {
		guideVersion.Variables = buildGuideVersionVariables(vars)
	}

	if resource := d.Get("resources").([]interface{}); resource != nil {
		guideVersion.Resources = buildGuideVersionResources(resource)
	}

	log.Printf("Succesfully Built Guide Version from Resource Data")
	return guideVersion
}

func flattenGuideVersionResources(resources GuideVersionResources) []interface{} {
	if len(resources.DataActions) == 0 {
		return nil
	}

	resourceMap := map[string]interface{}{}

	// Convert data actions
	if len(resources.DataActions) > 0 {
		dataActions := make([]interface{}, len(resources.DataActions))
		for i, action := range resources.DataActions {
			actionMap := map[string]interface{}{
				"data_action_id": action.ID,
				"label":          action.Label,
			}
			if action.Description != "" {
				actionMap["description"] = action.Description
			}
			dataActions[i] = actionMap
		}
		resourceMap["data_action"] = dataActions
	}

	return []interface{}{resourceMap}
}

func flattenGuideVersionVariables(variables []Variable) []interface{} {
	if len(variables) == 0 {
		return nil
	}

	result := make([]interface{}, len(variables))
	for i, v := range variables {
		varMap := map[string]interface{}{
			"name":  v.Name,
			"type":  v.Type,
			"scope": v.Scope,
		}
		if v.Description != "" {
			varMap["description"] = v.Description
		}
		result[i] = varMap
	}

	return result
}

// Structs

type Variable struct {
	Name        string `json:"name,omitempty"`
	Type        string `json:"type,omitempty"`
	Scope       string `json:"scope,omitempty"`
	Description string `json:"description,omitempty"`
}

type GuideVersionPublishJobRequest struct {
	GuideId      string              `json:"guideId,omitempty"`
	VersionId    string              `json:"versionId,omitempty"`
	GuideVersion GuideVersionPublish `json:"guideVersion,omitempty"`
}

type GuideVersionPublish struct {
	State string `json:"state,omitempty"`
}

type DataAction struct {
	ID          string `json:"id,omitempty"`
	Label       string `json:"label,omitempty"`
	Description string `json:"description,omitempty"`
}

type GuideVersionResources struct {
	DataActions []DataAction `json:"dataActions,omitempty"`
}

type CreateGuideVersionRequest struct {
	GuideID     string                `json:"guideId,omitempty"`
	Instruction string                `json:"instruction,omitempty"`
	Variables   []Variable            `json:"variables,omitempty"`
	Resources   GuideVersionResources `json:"resources,omitempty"`
}

type UpdateGuideVersion struct {
	GuideID     string                `json:"guideId,omitempty"`
	Instruction string                `json:"instruction,omitempty"`
	Variables   []Variable            `json:"variables,omitempty"`
	Resources   GuideVersionResources `json:"resources,omitempty"`
}

type Guide struct {
	Id      string `json:"id,omitempty"`
	SelfUri string `json:"selfUri,omitempty"`
}

type VersionResponse struct {
	Id          *string               `json:"id,omitempty"`
	Guide       Guide                 `json:"guide,omitempty"`
	Instruction string                `json:"instruction,omitempty"`
	Variables   []Variable            `json:"variables,omitempty"`
	Resources   GuideVersionResources `json:"resources,omitempty"`
	Version     string                `json:"version,omitempty"`
	State       string                `json:"state,omitempty"`
}

type VersionJobResponse struct {
	Id           *string          `json:"id,omitempty"`
	GuideId      *string          `json:"guideId,omitempty"`
	Status       *string          `json:"status,omitempty"`
	GuideVersion *VersionResponse `json:"guideVersion,omitempty"`
	Errors       []ErrorBody      `json:"errors,omitempty"`
}

type ErrorBody struct {
	Message           string `json:"message,omitempty"`
	Code              string `json:"code,omitempty"`
	Status            int    `json:"status,omitempty"`
	EntityId          string `json:"entityId,omitempty"`
	EntityName        string `json:"entityName,omitempty"`
	MessageWithParams string `json:"messageWithParams,omitempty"`
}
