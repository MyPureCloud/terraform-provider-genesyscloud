package journey_action_map

import (
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"

	"terraform-provider-genesyscloud/genesyscloud/util/testrunner"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceJourneyActionMap(t *testing.T) {
	runDataJourneyActionMapTestCase(t, "find_by_name")
}

func runDataJourneyActionMapTestCase(t *testing.T, testCaseName string) {
	testObjectName := testrunner.TestObjectIdPrefix + testCaseName
	testObjectFullName := ResourceType + "." + testObjectName
	SetupJourneyActionMap(t, testCaseName, sdkConfig)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: testrunner.GenerateDataSourceTestSteps(ResourceType, testCaseName, []resource.TestCheckFunc{
			resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttrPair("data."+testObjectFullName, "id", testObjectFullName, "id"),
				resource.TestCheckResourceAttr(testObjectFullName, "display_name", testObjectName+"_to_find"),
			),
		}),
	})
}
