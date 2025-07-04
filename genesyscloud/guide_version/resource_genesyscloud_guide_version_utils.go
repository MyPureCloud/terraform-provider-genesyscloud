package guide_version

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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
	if len(variables) != 0 {
		guideVersion.Variables = buildGuideVersionVariables(variables)
	}

	resources := d.Get("resources").([]interface{})
	if len(resources) != 0 {
		builtResources := buildGuideVersionResources(resources)
		// Only set resources if there are valid data actions
		if len(builtResources.DataActions) > 0 {
			guideVersion.Resources = builtResources
		}
	}

	log.Printf("Successfully Built Guide Version from Resource Data")
	return guideVersion
}

func buildGuideVersionResources(resourcesList []interface{}) GuideVersionResources {
	if len(resourcesList) == 0 {
		log.Printf("Warning: Resources list is nil or empty")
		return GuideVersionResources{
			DataActions: []DataAction{},
		}
	}

	log.Printf("Building Resource Attributes for Guide")

	resourcesMap, ok := resourcesList[0].(map[string]interface{})
	if !ok {
		log.Printf("Warning: Invalid resources map structure")
		return GuideVersionResources{
			DataActions: []DataAction{},
		}
	}

	var dataActions []DataAction
	if dataActionsList, ok := resourcesMap["data_action"].([]interface{}); ok {
		for _, v := range dataActionsList {
			dataActionMap, ok := v.(map[string]interface{})
			if !ok {
				log.Printf("Warning: Invalid data action map structure")
				continue
			}

			if dataActionID, ok := dataActionMap["data_action_id"].(string); ok && dataActionID != "" {
				if label, ok := dataActionMap["label"].(string); ok && label != "" {
					dataAction := DataAction{
						ID:    dataActionID,
						Label: label,
					}

					if description, ok := dataActionMap["description"].(string); ok && description != "" {
						dataAction.Description = description
					}

					dataActions = append(dataActions, dataAction)
				}
			}
		}
	}

	log.Printf("Successfully Built Resource Attributes for Guide with %d valid data actions", len(dataActions))
	return GuideVersionResources{
		DataActions: dataActions,
	}
}

func buildGuideVersionVariables(vars []interface{}) []Variable {
	if len(vars) == 0 {
		log.Printf("Warning: Variables list is nil or empty")
		return []Variable{}
	}

	var variables []Variable

	for _, v := range vars {
		variable := Variable{
			Name:  v.(map[string]interface{})["name"].(string),
			Type:  v.(map[string]interface{})["type"].(string),
			Scope: v.(map[string]interface{})["scope"].(string),
		}

		if description := v.(map[string]interface{})["description"].(string); description != "" {
			variable.Description = description
		}

		variables = append(variables, variable)
	}

	log.Printf("Successfully built %d valid variables", len(variables))
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
		builtResources := buildGuideVersionResources(resource)
		if len(builtResources.DataActions) > 0 {
			guideVersion.Resources = builtResources
		}
	}

	log.Printf("Successfully Built Guide Version from Resource Data")
	return guideVersion
}

func flattenGuideVersionResources(resources GuideVersionResources) []interface{} {
	if len(resources.DataActions) == 0 {
		return nil
	}

	var validDataActions []DataAction
	for _, action := range resources.DataActions {
		if action.ID != "" && action.Label != "" {
			validDataActions = append(validDataActions, action)
		}
	}

	if len(validDataActions) == 0 {
		return nil
	}

	resourceMap := map[string]interface{}{}

	dataActions := make([]interface{}, len(validDataActions))
	for i, action := range validDataActions {
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
	Id                           *string          `json:"id,omitempty"`
	Name                         *string          `json:"name,omitempty"`
	Source                       *string          `json:"source,omitempty"`
	Status                       *string          `json:"status,omitempty"`
	LatestSavedVersion           *GuideVersionRef `json:"latestSavedVersion,omitempty"`
	LatestProductionReadyVersion *GuideVersionRef `json:"latestProductionReadyVersion,omitempty"`
	SelfUri                      string           `json:"selfUri,omitempty"`
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

type GuideVersionRef struct {
	Version *string `json:"version,omitempty"`
	SelfUri *string `json:"selfUri,omitempty"`
}

type GuideEntityListing struct {
	SetFieldNames map[string]bool `json:"-"`
	Entities      *[]Guide        `json:"entities,omitempty"`
	PageNumber    *int            `json:"pageNumber,omitempty"`
	PageSize      *int            `json:"pageSize,omitempty"`
	NextUri       *string         `json:"nextUri,omitempty"`
	PreviousUri   *string         `json:"previousUri,omitempty"`
	FirstUri      *string         `json:"firstUri,omitempty"`
	SelfUri       *string         `json:"selfUri,omitempty"`
	PageCount     *int            `json:"pageCount,omitempty"`
}
