package testrunner

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	dataSourceTestType = "data_source"
	resourceTestType   = "resource"
	TestObjectIdPrefix = "terraform_test_"
)

func GetTestDataPath(elem ...string) string {
	basePath := filepath.Join("..", "test", "data")
	subPath := filepath.Join(elem...)
	return filepath.Join(basePath, subPath)
}

func GenerateDataSourceTestSteps(resourceName string, testCaseName string, idPrefix string, checkFuncs []resource.TestCheckFunc) []resource.TestStep {
	return GenerateTestSteps(dataSourceTestType, resourceName, testCaseName, idPrefix, checkFuncs)
}

func GenerateResourceTestSteps(resourceName string, testCaseName string, idPrefix string, checkFuncs []resource.TestCheckFunc) []resource.TestStep {
	return GenerateTestSteps(resourceTestType, resourceName, testCaseName, idPrefix, checkFuncs)
}

func GenerateTestSteps(testType string, resourceName string, testCaseName string, idPrefix string, checkFuncs []resource.TestCheckFunc) []resource.TestStep {
	var testSteps []resource.TestStep

	testCasePath := GetTestDataPath(testType, resourceName, testCaseName)
	testCaseDirEntries, _ := os.ReadDir(testCasePath)
	checkFuncIndex := 0
	for _, testCaseDirEntry := range testCaseDirEntries {
		if !testCaseDirEntry.IsDir() && strings.HasSuffix(testCaseDirEntry.Name(), ".tf") {
			testCaseStepFilePath := filepath.Join(testCasePath, testCaseDirEntry.Name())
			testCaseResource, _ := os.ReadFile(testCaseStepFilePath)
			config := strings.ReplaceAll(string(testCaseResource), "-TEST-CASE-", testCaseName)
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
		ResourceName:      resourceName + "." + idPrefix + testCaseName,
		ImportState:       true,
		ImportStateVerify: true,
	})

	return testSteps
}
