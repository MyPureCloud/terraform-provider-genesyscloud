package genesyscloud

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/testrunner"
)

func TestAccDataJourneySegment(t *testing.T) {
	runDataJourneySegmentTestCase(t, "find_by_name")
}

func runDataJourneySegmentTestCase(t *testing.T, testCaseName string) {
	const resourceName = "genesyscloud_journey_segment"
	testObjectName := testrunner.TestObjectIdPrefix + testCaseName
	testObjectFullName := resourceName + "." + testObjectName
	setupJourneySegment(t, testCaseName)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: testrunner.GenerateDataSourceTestSteps(resourceName, testCaseName, []resource.TestCheckFunc{
			resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttrPair("data."+testObjectFullName, "id", testObjectFullName, "id"),
				resource.TestCheckResourceAttr(testObjectFullName, "display_name", testObjectName+"_to_find"),
			),
		}),
	})
}
