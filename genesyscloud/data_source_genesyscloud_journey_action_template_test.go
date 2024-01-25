package genesyscloud

import (
	"testing"

	"terraform-provider-genesyscloud/genesyscloud/util/testrunner"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceJourneyActionTemplate(t *testing.T) {
	runDataJourneyActionTemplateTestCase(t, "find_by_name")
}

func runDataJourneyActionTemplateTestCase(t *testing.T, testCaseName string) {
	const resourceName = "genesyscloud_journey_action_template"
	testObjectName := testrunner.TestObjectIdPrefix + testCaseName
	testObjectFullName := resourceName + "." + testObjectName
	setupJourneyActionMap(t, testCaseName)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { TestAccPreCheck(t) },
		ProviderFactories: GetProviderFactories(providerResources, providerDataSources),
		Steps: testrunner.GenerateDataSourceTestSteps(resourceName, testCaseName, []resource.TestCheckFunc{
			resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttrPair("data."+testObjectFullName, "id", testObjectFullName, "id"),
				resource.TestCheckResourceAttr(testObjectFullName, "name", testObjectName),
			),
		}),
	})
}
