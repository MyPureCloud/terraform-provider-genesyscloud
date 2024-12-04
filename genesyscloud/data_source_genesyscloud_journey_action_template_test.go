package genesyscloud

import (
	"terraform-provider-genesyscloud/genesyscloud/journey_action_map"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"

	"terraform-provider-genesyscloud/genesyscloud/util/testrunner"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceJourneyActionTemplate(t *testing.T) {
	runDataJourneyActionTemplateTestCase(t, "find_by_name")
}

func runDataJourneyActionTemplateTestCase(t *testing.T, testCaseName string) {
	const resourceType = "genesyscloud_journey_action_template"
	testObjectName := testrunner.TestObjectIdPrefix + testCaseName
	testObjectFullName := resourceType + "." + testObjectName
	journey_action_map.SetupJourneyActionMap(t, testCaseName, sdkConfig)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: testrunner.GenerateDataSourceTestSteps(resourceType, testCaseName, []resource.TestCheckFunc{
			resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttrPair("data."+testObjectFullName, "id", testObjectFullName, "id"),
				resource.TestCheckResourceAttr(testObjectFullName, "name", testObjectName),
			),
		}),
	})
}
