package genesyscloud

import (
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"

	"terraform-provider-genesyscloud/genesyscloud/util/testrunner"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceJourneyOutcome(t *testing.T) {
	runDataJourneyOutcomeTestCase(t, "find_by_name")
}

func runDataJourneyOutcomeTestCase(t *testing.T, testCaseName string) {
	const resourceName = "genesyscloud_journey_outcome"
	testObjectName := testrunner.TestObjectIdPrefix + testCaseName
	testObjectFullName := resourceName + "." + testObjectName
	setupJourneyOutcome(t, testCaseName)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: testrunner.GenerateDataSourceTestSteps(resourceName, testCaseName, []resource.TestCheckFunc{
			resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttrPair("data."+testObjectFullName, "id", testObjectFullName, "id"),
				resource.TestCheckResourceAttr(testObjectFullName, "display_name", testObjectName+"_to_find"),
			),
		}),
	})
}
