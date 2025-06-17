package guide_version

import "net/http"

func buildRequestHeader(r *http.Request, p *guideVersionProxy) *http.Request {
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("Authorization", "Bearer "+p.clientConfig.AccessToken)
	return r
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

type VersionResponse struct {
	Id          *string               `json:"id,omitempty"`
	GuideID     string                `json:"guideId,omitempty"`
	Instruction string                `json:"instruction,omitempty"`
	Variables   []Variable            `json:"variables,omitempty"`
	Resources   GuideVersionResources `json:"resources,omitempty"`
	Version     string                `json:"version,omitempty"`
	State       string                `json:"state,omitempty"`
}
