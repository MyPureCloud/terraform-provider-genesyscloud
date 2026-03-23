package guide

import (
	"context"
	"fmt"
	"log"

	customapi "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/custom_api_client"

	"github.com/mypurecloud/platform-client-sdk-go/v179/platformclientv2"
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
	c := customapi.NewClient(clientConfig, ResourceType)
	_, resp, err := customapi.DoRaw(context.Background(), c, customapi.MethodGet, "/api/v2/guides", nil, nil)
	if err != nil {
		if resp != nil && resp.StatusCode < 500 {
			return true
		}
		log.Printf("Error checking guide feature toggle: %v", err)
		return false
	}
	return true
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
