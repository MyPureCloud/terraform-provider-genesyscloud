package genesyscloud

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataJourneyActionMap(t *testing.T) {
	runDataJourneyActionMapTestCase(t, "find_by_name")
}

func runDataJourneyActionMapTestCase(t *testing.T, testCaseName string) {
	const testType = "data_source"
	const testSuitName = "journey_action_map"
	const resourceName = "genesyscloud_journey_segment"
	const idPrefix = "terraform_test_"
	testObjectName := resourceName + "." + idPrefix + testCaseName
	setupJourneySegment(t, idPrefix, testCaseName)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: generateTestSteps(testType, testSuitName, testCaseName, resourceName, idPrefix, []resource.TestCheckFunc{
			resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttrPair("data."+testObjectName, "id", testObjectName, "id"),
				resource.TestCheckResourceAttr(testObjectName, "display_name", idPrefix+testCaseName+"_to_find"),
			),
		}),
	})
}
