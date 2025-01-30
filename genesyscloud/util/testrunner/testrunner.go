package testrunner

import (
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	DataSourceTestType              = "data_source"
	ResourceTestType                = "resource"
	TestObjectIdPrefix              = "terraform_test_"
	testObjectIdTestCasePlaceHolder = "-TEST-CASE-"
)

var RootDir string

func init() {
	RootDir = getRootDir()
}

// Helper function that retrieves the location of the root directory
func getRootDir() string {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		log.Fatal("Could not get caller info")
	}

	// Get the directory containing the current file
	dir := filepath.Dir(filename)

	// Keep going up until we find the directory containing main.go
	for {
		if _, err := os.Stat(filepath.Join(dir, "main.go")); err == nil {
			// Found the directory containing main.go
			return dir
		}

		// Move up one directory
		parent := filepath.Dir(dir)
		if parent == dir {
			// We've reached the root of the filesystem without finding main.go
			log.Fatal("Could not find directory containing main.go")
		}
		dir = parent
	}
}

func GetTestDataPath(elem ...string) string {
	basePath := filepath.Join(RootDir, "test", "data")
	subPath := filepath.Join(elem...)
	return filepath.Join(basePath, subPath)
}

func GetTestTempPath(elem ...string) string {
	basePath := filepath.Join(RootDir, "test", "temp")
	subPath := filepath.Join(elem...)
	return filepath.Join(basePath, subPath)
}

func GenerateDataSourceTestSteps(resourceType string, testCaseName string, checkFuncs []resource.TestCheckFunc) []resource.TestStep {
	return GenerateTestSteps(DataSourceTestType, resourceType, testCaseName, checkFuncs)
}

func GenerateResourceTestSteps(resourceType string, testCaseName string, checkFuncs []resource.TestCheckFunc) []resource.TestStep {
	return GenerateTestSteps(ResourceTestType, resourceType, testCaseName, checkFuncs)
}

func GenerateTestSteps(testType string, resourceType string, testCaseName string, checkFuncs []resource.TestCheckFunc) []resource.TestStep {
	var testSteps []resource.TestStep
	var testCasePath string
	testCasePath = GetTestDataPath(testType, resourceType, testCaseName)
	if resourceType == "genesyscloud_journey_action_map" || resourceType == "genesyscloud_journey_action_template" || resourceType == "genesyscloud_journey_outcome" {
		testCasePath = filepath.Join("../", testCasePath)
	}
	testCaseDirEntries, _ := os.ReadDir(testCasePath)
	checkFuncIndex := 0
	for _, testCaseDirEntry := range testCaseDirEntries {
		if !testCaseDirEntry.IsDir() && strings.HasSuffix(testCaseDirEntry.Name(), ".tf") {
			testCaseStepFilePath := filepath.Join(testCasePath, testCaseDirEntry.Name())
			testCaseResource, _ := os.ReadFile(testCaseStepFilePath)
			config := strings.ReplaceAll(string(testCaseResource), testObjectIdTestCasePlaceHolder, testCaseName)
			var checkFunc resource.TestCheckFunc = nil
			if checkFuncs != nil && checkFuncIndex < len(checkFuncs) {
				checkFunc = checkFuncs[checkFuncIndex]
			}
			testSteps = append(testSteps, resource.TestStep{
				PreConfig: func() { log.Printf("Executing test step config => %s", testCaseStepFilePath) },
				Config:    config,
				Check:     checkFunc})
			checkFuncIndex++
		}
	}
	log.Printf("Generated %d test steps for testcase => %s", len(testSteps), testCasePath)

	testSteps = append(testSteps, resource.TestStep{
		PreConfig:         func() { log.Printf("Executing ImportState test step config => %s", testCaseName) },
		ResourceName:      resourceType + "." + TestObjectIdPrefix + testCaseName,
		ImportState:       true,
		ImportStateVerify: true,
	})

	return testSteps
}

func GenerateFullPathId(resourceType string, resourceLabel string) string {
	return resourceType + "." + resourceLabel + "." + "id"
}
