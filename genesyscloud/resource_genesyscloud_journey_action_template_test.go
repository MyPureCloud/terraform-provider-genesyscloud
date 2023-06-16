package genesyscloud

import (
	"fmt"
	"log"
	"strings"
	"testing"

	"terraform-provider-genesyscloud/genesyscloud/util/testrunner"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v102/platformclientv2"
)

const ActionTemplateResourceName = "genesyscloud_journey_action_template"

func TestAccResourceJourneyActionTemplate(t *testing.T) {
	runJourneyActionTemplateTestCase(t, "action_template")
}

func runJourneyActionTemplateTestCase(t *testing.T, testCaseName string) {
	setupJourneyActionTemplate(t, testCaseName)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { TestAccPreCheck(t) },
		ProviderFactories: ProviderFactories,
		Steps:             testrunner.GenerateResourceTestSteps(ActionTemplateResourceName, testCaseName, nil),
		CheckDestroy:      testVerifyJourneyActionTemplatesDestroyed,
	})
}

func setupJourneyActionTemplate(t *testing.T, testCaseName string) {
	err := authorizeSdk()
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
				_, delErr := journeyApi.DeleteJourneyActiontemplate(*actionTemp.Id, true)
				if delErr != nil {
					diag.Errorf("failed to delete journey action template %s (%s): %s", *actionTemp.Id, *actionTemp.Name, delErr)
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

		if isStatus404(resp) {
			// Journey action map not found as expected
			continue
		}

		// Unexpected error
		return fmt.Errorf("unexpected error: %s", err)
	}
	// Success. All Journey action map destroyed
	return nil
}
