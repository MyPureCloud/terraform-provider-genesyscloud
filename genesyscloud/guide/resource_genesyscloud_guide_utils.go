package guide

import (
	"fmt"
	"log"
	"net/http"
	"net/url"

	"github.com/mypurecloud/platform-client-sdk-go/v162/platformclientv2"
)

func setRequestHeader(r *http.Request, p *guideProxy) *http.Request {
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("Accept", "application/json")
	r.Header.Set("Authorization", "Bearer "+p.clientConfig.AccessToken)
	return r
}

type ErrorBody struct {
	Message           string `json:"message,omitempty"`
	Code              string `json:"code,omitempty"`
	Status            int    `json:"status,omitempty"`
	EntityId          string `json:"entityId,omitempty"`
	EntityName        string `json:"entityName,omitempty"`
	MessageWithParams string `json:"messageWithParams,omitempty"`
}

type DeleteObjectJob struct {
	Id      string      `json:"id,omitempty"`
	GuideId string      `json:"guideId,omitempty"`
	Status  string      `json:"status,omitempty"`
	Errors  []ErrorBody `json:"errors,omitempty"`
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

// GenerateGuideResource generates terraform for a guide resource
func GenerateGuideResource(resourceID string, name string, source string) string {
	return fmt.Sprintf(`resource "%s" "%s" {
		name = "%s"
		source = "%s"
	}
	`, ResourceType, resourceID, name, source)
}

func GuideFtIsEnabled() bool {
	clientConfig := platformclientv2.GetDefaultConfiguration()
	client := &http.Client{}
	baseURL := clientConfig.BasePath + "/api/v2/guides"

	u, err := url.Parse(baseURL)
	if err != nil {
		log.Printf("Error parsing URL: %v", err)
	}

	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		log.Printf("Error creating request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+clientConfig.AccessToken)

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error sending request: %v", err)
	}

	defer resp.Body.Close()

	return resp.StatusCode < 500
}
