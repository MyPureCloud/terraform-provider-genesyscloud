package genesyscloud

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataJourneyOutcome(t *testing.T) {
	runDataJourneyOutcomeTestCase(t, "find_by_name")
}

func runDataJourneyOutcomeTestCase(t *testing.T, testCaseName string) {
	const testType = "data_source"
	const testSuitName = "journey_outcome"
	const resourceName = "genesyscloud_journey_outcome"
	const idPrefix = "terraform_test_"
	testObjectName := resourceName + "." + idPrefix + testCaseName
	setupJourneyOutcome(t, idPrefix, testCaseName)

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
