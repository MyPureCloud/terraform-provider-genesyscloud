package journey_segment

import (
	"fmt"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
	"log"
	"strings"
	"testing"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/testrunner"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v157/platformclientv2"
)

func TestAccResourceJourneySegment(t *testing.T) {
	if !assignmentExpirationFeatureToggleIsEnabled() {
		t.Skip("assignmentExpirationDays is still behind a feature toggle")
	}
	runResourceJourneySegmentTestCase(t, "basic_attributes")
}

func TestAccResourceJourneySegmentContextOnly(t *testing.T) {
	runResourceJourneySegmentTestCase(t, "context_only_to_journey_only")
}

func TestAccResourceJourneySegmentOptionalAttributes(t *testing.T) {
	runResourceJourneySegmentTestCase(t, "optional_attributes")
}

func runResourceJourneySegmentTestCase(t *testing.T, testCaseName string) {
	setupJourneySegment(t, testCaseName)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps:             testrunner.GenerateResourceJourneyTestSteps(ResourceType, testCaseName, nil),
		CheckDestroy:      testVerifyJourneySegmentsDestroyed,
	})
}

func setupJourneySegment(t *testing.T, testCaseName string) {
	_, err := provider.AuthorizeSdk()
	if err != nil {
		t.Fatal(err)
	}

	cleanupJourneySegments(testrunner.TestObjectIdPrefix + testCaseName)
}

func cleanupJourneySegments(idPrefix string) {
	journeyApi := platformclientv2.NewJourneyApiWithConfig(sdkConfig)

	segmentsToDelete := make([]platformclientv2.Journeysegment, 0)

	// go through all segments to find those to delete
	const pageSize = 200
	for pageNum := 1; ; pageNum++ {
		journeySegments, _, getErr := journeyApi.GetJourneySegments("", pageSize, pageNum, true, nil, nil, "")
		if getErr != nil {
			log.Printf("failed to get page %v of journeySegments: %v", pageNum, getErr)
			return
		}

		for _, journeySegment := range *journeySegments.Entities {
			if journeySegment.DisplayName != nil && strings.HasPrefix(*journeySegment.DisplayName, idPrefix) {
				segmentsToDelete = append(segmentsToDelete, journeySegment)
			}
		}

		if *journeySegments.PageNumber >= *journeySegments.PageCount {
			break
		}
	}

	// delete them
	for _, journeySegment := range segmentsToDelete {
		_, delErr := journeyApi.DeleteJourneySegment(*journeySegment.Id)
		if delErr != nil {
			util.BuildDiagnosticError(ResourceType, fmt.Sprintf("failed to delete journey segment %s (%s)", *journeySegment.Id, *journeySegment.DisplayName), delErr)
			return
		}
		log.Printf("Deleted journey segment %s (%s)", *journeySegment.Id, *journeySegment.DisplayName)
	}
}

func testVerifyJourneySegmentsDestroyed(state *terraform.State) error {
	journeyApi := platformclientv2.NewJourneyApiWithConfig(sdkConfig)
	for _, rs := range state.RootModule().Resources {
		if rs.Type != ResourceType {
			continue
		}

		journeySegment, resp, err := journeyApi.GetJourneySegment(rs.Primary.ID)
		if journeySegment != nil {
			return fmt.Errorf("journey segment (%s) still exists", rs.Primary.ID)
		}

		if util.IsStatus404(resp) {
			// Journey segment not found as expected
			continue
		}

		// Unexpected error
		return fmt.Errorf("unexpected error: %s", err)
	}
	// Success. All Journey segment destroyed
	return nil
}

// assignmentExpirationFeatureToggleIsEnabled checks if the assignmentExpirationDays field is still behind
// a feature toggle by attempting to create a journey segment with only this field.
//
// The function works by:
// 1. Making a POST request to create a journey segment with only assignmentExpirationDays field
// 2. If the feature toggle is disabled (field is restricted):
//   - The request will return a 501 status code (Not Implemented)
//   - Function returns false
//
// 3. If the feature toggle is enabled (field is available):
//   - The request will return a different status code (likely 400 for invalid request)
//   - Function returns true
//
// 4. If the request somehow succeeds, the created segment is cleaned up and function returns true
//
// Returns:
//   - bool: true if the feature toggle is enabled (field is not restricted), false otherwise
//
// Note: This function uses an intentionally invalid request body to test the feature toggle.
// When the toggle is disabled, the API returns 501 instead of 400
func assignmentExpirationFeatureToggleIsEnabled() bool {
	journeyApi := platformclientv2.NewJourneyApiWithConfig(sdkConfig)
	body := platformclientv2.Journeysegmentrequest{
		AssignmentExpirationDays: platformclientv2.Int(2),
	}

	log.Printf("Checking if assignmentExpirationDays is still behind a feature toggle")
	data, response, postErr := journeyApi.PostJourneySegments(body)
	if postErr != nil && response != nil {
		log.Printf("Received status code %d. Error: %s", response.StatusCode, postErr.Error())
		return response.StatusCode != 501
	}

	if data != nil && data.Id != nil {
		log.Printf("Deleting journey %s", *data.Id)
		_, deleteErr := journeyApi.DeleteJourneySegment(*data.Id)
		if deleteErr != nil {
			log.Printf("Failed to delete journey %s: %s", *data.Id, deleteErr.Error())
		}
	}

	return true
}
