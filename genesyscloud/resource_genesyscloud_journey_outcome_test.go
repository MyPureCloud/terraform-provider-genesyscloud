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

func TestAccResourceJourneyOutcome(t *testing.T) {
	runResourceJourneyOutcomeTestCase(t, "basic")
}

func runResourceJourneyOutcomeTestCase(t *testing.T, testCaseName string) {
	const resourceName = "genesyscloud_journey_outcome"
	setupJourneyOutcome(t, testCaseName)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { TestAccPreCheck(t) },
		ProviderFactories: GetProviderFactories(providerResources, providerDataSources),
		Steps:             testrunner.GenerateResourceTestSteps(resourceName, testCaseName, nil),
		CheckDestroy:      testVerifyJourneyOutcomesDestroyed,
	})
}

func setupJourneyOutcome(t *testing.T, testCaseName string) {
	_, err := AuthorizeSdk()
	if err != nil {
		t.Fatal(err)
	}

	testCasePrefix := testrunner.TestObjectIdPrefix + testCaseName
	cleanupJourneyOutcomes(testCasePrefix)
}

func cleanupJourneyOutcomes(idPrefix string) {
	journeyApi := platformclientv2.NewJourneyApiWithConfig(sdkConfig)

	pageCount := 1 // Needed because of broken journey common paging
	for pageNum := 1; pageNum <= pageCount; pageNum++ {
		const pageSize = 100
		journeyOutcomes, _, getErr := journeyApi.GetJourneyOutcomes(pageNum, pageSize, "", nil, nil, "")
		if getErr != nil {
			return
		}

		if journeyOutcomes.Entities == nil || len(*journeyOutcomes.Entities) == 0 {
			break
		}

		for _, journeyOutcome := range *journeyOutcomes.Entities {
			if journeyOutcome.DisplayName != nil && strings.HasPrefix(*journeyOutcome.DisplayName, idPrefix) {
				_, delErr := journeyApi.DeleteJourneyOutcome(*journeyOutcome.Id)
				if delErr != nil {
					diag.Errorf("failed to delete journey outcome %s (%s): %s", *journeyOutcome.Id, *journeyOutcome.DisplayName, delErr)
					return
				}
				log.Printf("Deleted journey outcome %s (%s)", *journeyOutcome.Id, *journeyOutcome.DisplayName)
			}
		}

		pageCount = *journeyOutcomes.PageCount
	}
}

func testVerifyJourneyOutcomesDestroyed(state *terraform.State) error {
	journeyApi := platformclientv2.NewJourneyApiWithConfig(sdkConfig)
	for _, rs := range state.RootModule().Resources {
		if rs.Type != "genesyscloud_journey_outcome" {
			continue
		}

		journeyOutcome, resp, err := journeyApi.GetJourneyOutcome(rs.Primary.ID)
		if journeyOutcome != nil {
			return fmt.Errorf("journey outcome (%s) still exists", rs.Primary.ID)
		}

		if IsStatus404(resp) {
			// Journey outcome not found as expected
			continue
		}

		// Unexpected error
		return fmt.Errorf("unexpected error: %s", err)
	}
	// Success. All Journey outcome destroyed
	return nil
}
