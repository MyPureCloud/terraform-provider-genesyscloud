package guide

import "net/http"

func setRequestHeader(r *http.Request, p *guideProxy) *http.Request {
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("Accept", "application/json")
	r.Header.Set("Authorization", "Bearer "+p.clientConfig.AccessToken)
	return r
}

type Guide struct {
	Id                           *string          `json:"id,omitempty"`
	Name                         *string          `json:"name,omitempty"`
	Source                       *string          `json:"source,omitempty"`
	Status                       *string          `json:"status,omitempty"`
	LatestSavedVersion           *GuideVersionRef `json:"latestSavedVersion,omitempty"`
	LatestProductionReadyVersion *GuideVersionRef `json:"latestProductionReadyVersion,omitempty"`
}

type Createguide struct {
	SetFieldNames map[string]bool `json:"setFieldNames,omitempty"`
	Name          *string         `json:"name,omitempty"`
	Source        *string         `json:"source,omitempty"`
}

type GuideEntityListing struct {
	// SetFieldNames defines the list of fields to use for controlled JSON serialization
	SetFieldNames map[string]bool `json:"-"`
	// Entities
	Entities *[]Guide `json:"entities,omitempty"`

	// PageNumber
	PageNumber *int `json:"pageNumber,omitempty"`

	// PageSize
	PageSize *int `json:"pageSize,omitempty"`

	// NextUri
	NextUri *string `json:"nextUri,omitempty"`

	// PreviousUri
	PreviousUri *string `json:"previousUri,omitempty"`

	// FirstUri
	FirstUri *string `json:"firstUri,omitempty"`

	// SelfUri
	SelfUri *string `json:"selfUri,omitempty"`
}

type GuideVersionRef struct {
	version *string `json:"version,omitempty"`
	selfUri *string `json:"selfUri,omitempty"`
}
