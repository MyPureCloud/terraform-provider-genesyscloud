package guide

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"

	"github.com/mypurecloud/platform-client-sdk-go/v165/platformclientv2"
)

// GenerateGuideResource generates terraform for a guide resource
func GenerateGuideResource(resourceID string, name string) string {
	return fmt.Sprintf(`resource "%s" "%s" {
		name = "%s"
	}
	`, ResourceType, resourceID, name)
}

// Helper function to check if the guide feature toggle is enabled
// Achieved by a GET request to the guides endpoint, checking if the status code is 5xx
func GuideFtIsEnabled() bool {
	clientConfig := platformclientv2.GetDefaultConfiguration()
	client := &http.Client{}
	baseURL := clientConfig.BasePath + "/api/v2/guides"

	u, err := url.Parse(baseURL)
	if err != nil {
		log.Printf("Error parsing URL: %v", err)
		return false
	}

	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		log.Printf("Error creating request: %v", err)
		return false
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+clientConfig.AccessToken)

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error sending request: %v", err)
		return false
	}

	defer resp.Body.Close()

	return resp.StatusCode < 500
}

// setRequestHeader sets the request header for the guide proxy
func setRequestHeader(r *http.Request, p *guideProxy) *http.Request {
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("Accept", "application/json")
	r.Header.Set("Authorization", "Bearer "+p.clientConfig.AccessToken)
	return r
}

// createHTTPRequest creates a new HTTP request with proper headers
func createHTTPRequest(method, url string, body io.Reader, p *guideProxy) (*http.Request, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}
	req = setRequestHeader(req, p)
	return req, nil
}

// marshalAndCreateRequest marshals a body to JSON and creates an HTTP request
func marshalAndCreateRequest(method, url string, body interface{}, p *guideProxy) (*http.Request, error) {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request body: %v", err)
	}
	return createHTTPRequest(method, url, bytes.NewBuffer(jsonBody), p)
}

// unmarshalResponse unmarshals a JSON response into the target struct
func unmarshalResponse(respBody []byte, target interface{}) error {
	if err := json.Unmarshal(respBody, target); err != nil {
		return fmt.Errorf("error unmarshaling response: %v", err)
	}
	return nil
}

// Structs

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
	Id   *string `json:"id,omitempty"`
	Name *string `json:"name,omitempty"`
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
