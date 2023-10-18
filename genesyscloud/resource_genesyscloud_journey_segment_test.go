package genesyscloud

import (
	"fmt"
	"log"
	"strings"
	"testing"

	"terraform-provider-genesyscloud/genesyscloud/util/testrunner"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v115/platformclientv2"
)

func TestAccResourceJourneySegmentCustomer(t *testing.T) {
	runResourceJourneySegmentTestCase(t, "basic_customer_attributes")
}

func TestAccResourceJourneySegmentSession(t *testing.T) {
	runResourceJourneySegmentTestCase(t, "basic_session_attributes")
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
		PreCheck:          func() { TestAccPreCheck(t) },
		ProviderFactories: GetProviderFactories(providerResources, providerDataSources),
		Steps:             testrunner.GenerateResourceTestSteps(resourceName, testCaseName, nil),
		CheckDestroy:      testVerifyJourneySegmentsDestroyed,
	})
}

func setupJourneySegment(t *testing.T, testCaseName string) {
	_, err := AuthorizeSdk()
	if err != nil {
		t.Fatal(err)
	}

	cleanupJourneySegments(testrunner.TestObjectIdPrefix + testCaseName)
}

func cleanupJourneySegments(idPrefix string) {
	journeyApi := platformclientv2.NewJourneyApiWithConfig(sdkConfig)

	pageCount := 1 // Needed because of broken journey common paging
	for pageNum := 1; pageNum <= pageCount; pageNum++ {
		const pageSize = 100
		journeySegments, _, getErr := journeyApi.GetJourneySegments("", pageSize, pageNum, true, nil, nil, "")
		if getErr != nil {
			return
		}

		if journeySegments.Entities == nil || len(*journeySegments.Entities) == 0 {
			break
		}

		for _, journeySegment := range *journeySegments.Entities {
			if journeySegment.DisplayName != nil && strings.HasPrefix(*journeySegment.DisplayName, idPrefix) {
				_, delErr := journeyApi.DeleteJourneySegment(*journeySegment.Id)
				if delErr != nil {
					diag.Errorf("failed to delete journey segment %s (%s): %s", *journeySegment.Id, *journeySegment.DisplayName, delErr)
					return
				}
				log.Printf("Deleted journey segment %s (%s)", *journeySegment.Id, *journeySegment.DisplayName)
			}
		}

		pageCount = *journeySegments.PageCount
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

		if IsStatus404(resp) {
			// Journey segment not found as expected
			continue
		}

		// Unexpected error
		return fmt.Errorf("unexpected error: %s", err)
	}
	// Success. All Journey segment destroyed
	return nil
}
