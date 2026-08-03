package journey_action_map

import (
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
	"testing"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/testrunner"

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
		Steps:             generateDataJourneyActionMapTestSteps(testCaseName, testObjectFullName),
	})
}

// generateDataJourneyActionMapTestSteps generates test steps from the .tf files and appends
// an import step with a pre-sleep to allow API eventual consistency in Jenkins environments.
func generateDataJourneyActionMapTestSteps(testCaseName string, testObjectFullName string) []resource.TestStep {
	testObjectName := testrunner.TestObjectIdPrefix + testCaseName
	testCasePath := testrunner.GetTestDataPath(testrunner.DataSourceTestType, ResourceType, testCaseName)
	testCaseDirEntries, _ := os.ReadDir(testCasePath)

	var testSteps []resource.TestStep
	checkFuncs := []resource.TestCheckFunc{
		resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttrPair("data."+testObjectFullName, "id", testObjectFullName, "id"),
			resource.TestCheckResourceAttr(testObjectFullName, "display_name", testObjectName+"_to_find"),
		),
	}
	checkFuncIndex := 0
	for _, entry := range testCaseDirEntries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".tf") {
			stepFilePath := filepath.Join(testCasePath, entry.Name())
			fileContent, _ := os.ReadFile(stepFilePath)
			config := strings.ReplaceAll(string(fileContent), "-TEST-CASE-", testCaseName)
			var checkFunc resource.TestCheckFunc
			if checkFuncIndex < len(checkFuncs) {
				checkFunc = checkFuncs[checkFuncIndex]
			}
			testSteps = append(testSteps, resource.TestStep{
				PreConfig: func() { log.Printf("Executing test step config => %s", stepFilePath) },
				Config:    config,
				Check:     checkFunc,
			})
			checkFuncIndex++
		}
	}
	log.Printf("Generated %d test steps for testcase => %s", len(testSteps), testCasePath)

	// Import step with a sleep to allow API eventual consistency.
	// Without this sleep, the resource may not be consistently visible across API nodes in Jenkins.
	testSteps = append(testSteps, resource.TestStep{
		PreConfig: func() {
			log.Printf("Waiting for API consistency before import step for %s", testCaseName)
			time.Sleep(20 * time.Second)
		},
		ResourceName:      ResourceType + "." + testrunner.TestObjectIdPrefix + testCaseName,
		ImportState:       true,
		ImportStateVerify: true,
	})
	return testSteps
}
