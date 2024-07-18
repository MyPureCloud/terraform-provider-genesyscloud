package genesyscloud

import (
	"fmt"
	"log"
	"strings"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"

	"terraform-provider-genesyscloud/genesyscloud/util/testrunner"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

func TestAccResourceJourneySegment(t *testing.T) {
	runResourceJourneySegmentTestCase(t, "basic_attributes")
}

func TestAccResourceJourneySegmentContextOnly(t *testing.T) {
	runResourceJourneySegmentTestCase(t, "context_only_to_journey_only")
}

func TestAccResourceJourneySegmentOptionalAttributes(t *testing.T) {
	runResourceJourneySegmentTestCase(t, "optional_attributes")
}

func runResourceJourneySegmentTestCase(t *testing.T, testCaseName string) {
	const resourceName = "genesyscloud_journey_segment"
	setupJourneySegment(t, testCaseName)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps:             testrunner.GenerateResourceTestSteps(resourceName, testCaseName, nil),
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
			util.BuildDiagnosticError("genesyscloud_journey_segment", fmt.Sprintf("failed to delete journey segment %s (%s)", *journeySegment.Id, *journeySegment.DisplayName), delErr)
			return
		}
		log.Printf("Deleted journey segment %s (%s)", *journeySegment.Id, *journeySegment.DisplayName)
	}
}

func testVerifyJourneySegmentsDestroyed(state *terraform.State) error {
	journeyApi := platformclientv2.NewJourneyApiWithConfig(sdkConfig)
	for _, rs := range state.RootModule().Resources {
		if rs.Type != "genesyscloud_journey_segment" {
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
