package guide_jobs

import (
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func buildGuideJobFromResourceData(d *schema.ResourceData) GenerateGuideContentRequest {
	guideJobReq := GenerateGuideContentRequest{}

	description := d.Get("description").(string)
	if description != "" {
		guideJobReq.Description = &description
	}

	url := d.Get("url").(string)
	if url != "" {
		guideJobReq.Url = &url
	}

	return guideJobReq
}

func setRequestHeader(r *http.Request, p *guideJobsProxy) *http.Request {
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("Accept", "application/json")
	r.Header.Set("Authorization", "Bearer "+p.clientConfig.AccessToken)
	return r
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
