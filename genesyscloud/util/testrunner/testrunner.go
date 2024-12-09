package testrunner

import (
	"log"
	"os"
	"path"
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

func GetTestDataPath(elem ...string) string {
	basePath := filepath.Join("..", "test", "data")
	subPath := filepath.Join(elem...)
	return filepath.Join(basePath, subPath)
}

func NormalizePath(path string) (string, error) {
	fullyQualifiedPath, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}

	if runtime.GOOS == "windows" {
		// Convert single backslashes to dobule backslashes if necessary
		fullyQualifiedPath = strings.ReplaceAll(fullyQualifiedPath, "\\", "\\\\")
	}

	return fullyQualifiedPath, nil
}

func NormalizeFileName(filename string) (string, error) {
	fullyQualifiedFineName, err := filepath.Abs(filename)
	if err != nil {
		return "", err
	}

	if runtime.GOOS == "windows" {
		// Convert single backslashes to single forwardslashes if necessary
		fullyQualifiedFineName = strings.ReplaceAll(fullyQualifiedFineName, "\\", "/")
	}

	return fullyQualifiedFineName, nil
}

func NormalizeSlash(fileNameWithSlash string) string {
	fullyQualifiedFileName := fileNameWithSlash

	if runtime.GOOS == "windows" {
		// Convert single backslashes to dobule backslashes if necessary
		fullyQualifiedFileName = strings.ReplaceAll(fullyQualifiedFileName, "\\", "\\\\")
	}

	return fullyQualifiedFileName
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
	if resourceType == "genesyscloud_journey_action_map" || resourceType == "genesyscloud_journey_action_template" {
		testCasePath = path.Join("../", testCasePath)
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
