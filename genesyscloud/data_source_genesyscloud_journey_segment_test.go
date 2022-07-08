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
	testObjectName := resourceName + "." + testrunner.TestObjectIdPrefix + testCaseName
	setupJourneySegment(t, testrunner.TestObjectIdPrefix, testCaseName)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: testrunner.GenerateDataSourceTestSteps(resourceName, testCaseName, testrunner.TestObjectIdPrefix, []resource.TestCheckFunc{
			resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttrPair("data."+testObjectName, "id", testObjectName, "id"),
				resource.TestCheckResourceAttr(testObjectName, "display_name", testrunner.TestObjectIdPrefix+testCaseName+"_to_find"),
			),
		}),
	})
}
