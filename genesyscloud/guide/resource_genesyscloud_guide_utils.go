package guide

import "net/http"

func setRequestHeader(r *http.Request, p *guideProxy) *http.Request {
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("Accept", "application/json")
	r.Header.Set("Authorization", "Bearer "+p.clientConfig.AccessToken)
	return r
}

type DeleteObject struct {
	Id      string `json:"id,omitempty"`
	GuideId string `json:"guideId,omitempty"`
	Status  string `json:"status,omitempty"`
}

type Guide struct {
	Id                           *string          `json:"id,omitempty"`
	Name                         *string          `json:"name,omitempty"`
	Source                       *string          `json:"source,omitempty"`
	Status                       *string          `json:"status,omitempty"`
	LatestSavedVersion           *GuideVersionRef `json:"latestSavedVersion,omitempty"`
	LatestProductionReadyVersion *GuideVersionRef `json:"latestProductionReadyVersion,omitempty"`
}

type CreateGuide struct {
	SetFieldNames map[string]bool `json:"setFieldNames,omitempty"`
	Name          *string         `json:"name,omitempty"`
	Source        *string         `json:"source,omitempty"`
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

type GuideVersionRef struct {
	Version *string `json:"version,omitempty"`
	SelfUri *string `json:"selfUri,omitempty"`
}
