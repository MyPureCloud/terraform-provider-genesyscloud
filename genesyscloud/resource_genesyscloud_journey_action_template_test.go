package genesyscloud

import (
	"fmt"
	"log"
	"strings"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"

	"terraform-provider-genesyscloud/genesyscloud/util/testrunner"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

const ActionTemplateResourceName = "genesyscloud_journey_action_template"

func TestAccResourceJourneyActionTemplate(t *testing.T) {
	runJourneyActionTemplateTestCase(t, "action_template")
}

func runJourneyActionTemplateTestCase(t *testing.T, testCaseName string) {
	setupJourneyActionTemplate(t, testCaseName)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps:             testrunner.GenerateResourceTestSteps(ActionTemplateResourceName, testCaseName, nil),
		CheckDestroy:      testVerifyJourneyActionTemplatesDestroyed,
	})
}

func setupJourneyActionTemplate(t *testing.T, testCaseName string) {
	_, err := provider.AuthorizeSdk()
	if err != nil {
		t.Fatal(err)
	}

	testCasePrefix := testrunner.TestObjectIdPrefix + testCaseName
	cleanupJourneyActionTemplate(testCasePrefix)
}

func cleanupJourneyActionTemplate(idPrefix string) {
	journeyApi := platformclientv2.NewJourneyApiWithConfig(sdkConfig)

	pageCount := 1 // Needed because of broken journey common paging
	for pageNum := 1; pageNum <= pageCount; pageNum++ {
		const pageSize = 100
		actionTemplate, _, getErr := journeyApi.GetJourneyActiontemplates(pageNum, pageSize, "", "", "", nil, "")
		if getErr != nil {
			return
		}

		if actionTemplate.Entities == nil || len(*actionTemplate.Entities) == 0 {
			break
		}

		for _, actionTemp := range *actionTemplate.Entities {
			if actionTemp.Name != nil && strings.HasPrefix(*actionTemp.Name, idPrefix) {
				resp, delErr := journeyApi.DeleteJourneyActiontemplate(*actionTemp.Id, true)
				if delErr != nil {
					util.BuildAPIDiagnosticError("genesyscloud_journey_action_template", fmt.Sprintf("failed to delete journey action template %s (%s): %s", *actionTemp.Id, *actionTemp.Name, delErr), resp)
					return
				}
				log.Printf("Deleted Journey Action Template %s (%s)", *actionTemp.Id, *actionTemp.Name)
			}
		}

		pageCount = *actionTemplate.PageCount
	}
}

func testVerifyJourneyActionTemplatesDestroyed(state *terraform.State) error {
	journeyApi := platformclientv2.NewJourneyApiWithConfig(sdkConfig)
	for _, rs := range state.RootModule().Resources {
		if rs.Type != ActionTemplateResourceName {
			continue
		}

		actionMap, resp, err := journeyApi.GetJourneyActiontemplate(rs.Primary.ID)
		if actionMap != nil {
			return fmt.Errorf("journey action template (%s) still exists", rs.Primary.ID)
		}

		if util.IsStatus404(resp) {
			// Journey action map not found as expected
			continue
		}

		// Unexpected error
		return fmt.Errorf("unexpected error: %s", err)
	}
	// Success. All Journey action map destroyed
	return nil
}
