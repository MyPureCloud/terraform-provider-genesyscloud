package genesyscloud

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/testrunner"
)

func TestAccDataJourneyActionMap(t *testing.T) {
	runDataJourneyActionMapTestCase(t, "find_by_name")
}

func runDataJourneyActionMapTestCase(t *testing.T, testCaseName string) {
	const resourceName = "genesyscloud_journey_action_map"
	testObjectName := resourceName + "." + testrunner.TestObjectIdPrefix + testCaseName
	setupJourneyActionMap(t, testrunner.TestObjectIdPrefix, testCaseName)

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
