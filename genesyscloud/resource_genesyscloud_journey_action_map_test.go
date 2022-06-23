package genesyscloud

import (
	"fmt"
	"log"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v72/platformclientv2"
)

func TestAccResourceJourneyActionMap(t *testing.T) {
	runJourneyActionMapTestCase(t, "basic")
}

func runJourneyActionMapTestCase(t *testing.T, testCaseName string) {
	const testSuitName = "journey_action_map"
	const resourceName = "genesyscloud_journey_action_map"
	idPrefix := setupJourneyActionMap(t)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps:             generateTestSteps(testSuitName, testCaseName, resourceName, idPrefix),
		CheckDestroy:      testVerifyJourneyActionMapsDestroyed,
	})
}

func setupJourneyActionMap(t *testing.T) string {
	const idPrefix = "terraform_test_"

	err := authorizeSdk()
	if err != nil {
		t.Fatal(err)
	}

	cleanupJourneyActionMaps(idPrefix)
	return idPrefix
}

func cleanupJourneyActionMaps(idPrefix string) {
	journeyApi := platformclientv2.NewJourneyApiWithConfig(sdkConfig)

	pageCount := 1 // Needed because of broken journey common paging
	for pageNum := 1; pageNum <= pageCount; pageNum++ {
		const pageSize = 100
		journeyActionMaps, _, getErr := journeyApi.GetJourneyActionmaps(pageNum, pageSize, "", "", "", nil, nil, "")
		if getErr != nil {
			return
		}

		if journeyActionMaps.Entities == nil || len(*journeyActionMaps.Entities) == 0 {
			break
		}

		for _, journeyActionMap := range *journeyActionMaps.Entities {
			if journeyActionMap.DisplayName != nil && strings.HasPrefix(*journeyActionMap.DisplayName, idPrefix) {
				_, delErr := journeyApi.DeleteJourneySegment(*journeyActionMap.Id)
				if delErr != nil {
					diag.Errorf("failed to delete journey action map %s (%s): %s", *journeyActionMap.Id, *journeyActionMap.DisplayName, delErr)
					return
				}
				log.Printf("Deleted journey action map %s (%s)", *journeyActionMap.Id, *journeyActionMap.DisplayName)
			}
		}

		pageCount = *journeyActionMaps.PageCount
	}
}

func testVerifyJourneyActionMapsDestroyed(state *terraform.State) error {
	journeyApi := platformclientv2.NewJourneyApiWithConfig(sdkConfig)
	for _, rs := range state.RootModule().Resources {
		if rs.Type != "genesyscloud_journey_action_map" {
			continue
		}

		journeySegment, resp, err := journeyApi.GetJourneyActionmap(rs.Primary.ID)
		if journeySegment != nil {
			return fmt.Errorf("journey action map (%s) still exists", rs.Primary.ID)
		}

		if isStatus404(resp) {
			// Journey action map not found as expected
			continue
		}

		// Unexpected error
		return fmt.Errorf("unexpected error: %s", err)
	}
	// Success. All Journey action map destroyed
	return nil
}
