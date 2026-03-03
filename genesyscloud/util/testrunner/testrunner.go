package testrunner

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

const (
	DataSourceTestType              = "data_source"
	ResourceTestType                = "resource"
	TestObjectIdPrefix              = "terraform_test_"
	testObjectIdTestCasePlaceHolder = "-TEST-CASE-"
)

var RootDir string

func init() {
	if isRunningTests() {
		RootDir = getRootDir()
	}
}

func isRunningTests() bool {
	if os.Getenv("TF_ACC") != "" {
		return true
	}

	return false
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
	if !isRunningTests() {
		return ""
	}
	basePath := filepath.Join(RootDir, "test", "data")
	subPath := filepath.Join(elem...)
	return NormalizePath(filepath.Join(basePath, subPath))
}

func GetTestTempPath(elem ...string) string {
	if !isRunningTests() {
		return ""
	}
	basePath := filepath.Join(RootDir, "test", "temp")
	subPath := filepath.Join(elem...)
	return NormalizePath(filepath.Join(basePath, subPath))
}

func NormalizePath(path string) string {
	fullyQualifiedPath := path

	if runtime.GOOS == "windows" {
		// Convert single backslashes to dobule backslashes if necessary
		fullyQualifiedPath = strings.ReplaceAll(path, "\\", "\\\\")
	}
	return fullyQualifiedPath
}

func GenerateDataJourneySourceTestSteps(resourceType string, testCaseName string, checkFuncs []resource.TestCheckFunc) []resource.TestStep {
	return GenerateJourneyTestSteps(DataSourceTestType, resourceType, testCaseName, checkFuncs)
}

func GenerateResourceJourneyTestSteps(resourceType string, testCaseName string, checkFuncs []resource.TestCheckFunc) []resource.TestStep {
	return GenerateJourneyTestSteps(ResourceTestType, resourceType, testCaseName, checkFuncs)
}

func GenerateJourneyTestSteps(testType string, resourceType string, testCaseName string, checkFuncs []resource.TestCheckFunc) []resource.TestStep {
	var testSteps []resource.TestStep
	var testCasePath string
	testCasePath = GetTestDataPath(testType, resourceType, testCaseName)
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

// Helper function to create test provider
func GenerateTestProvider(resourceName string, schemas map[string]*schema.Schema, diff schema.CustomizeDiffFunc) *schema.Provider {
	return &schema.Provider{
		Schema: schemas,
		ResourcesMap: map[string]*schema.Resource{
			resourceName: {
				Schema:        schemas,
				CustomizeDiff: diff,
			},
		},
	}
}
func GenerateTestDiff(provider *schema.Provider, resourceName string, oldValue, newValue map[string]string) (*terraform.InstanceDiff, error) {
	// Convert newValue map[string]string to map[string]interface{} and handle list attributes
	newI := make(map[string]interface{})
	for k, v := range newValue {
		if strings.Contains(k, ".#") {
			// This is a list length indicator - skip it as we'll handle the list elements
			continue
		}

		// Check if this is a list element (e.g., "list_attr.0", "list_attr.1")
		if idx := strings.LastIndex(k, "."); idx != -1 {
			listName := k[:idx]
			if _, err := strconv.Atoi(k[idx+1:]); err == nil {
				// This is a list element
				// Initialize the list if it doesn't exist
				if _, exists := newI[listName]; !exists {
					// Find the length of the list from the ".#" attribute
					if lenStr, ok := newValue[listName+".#"]; ok {
						length, _ := strconv.Atoi(lenStr)
						newI[listName] = make([]interface{}, length)
					}
				}

				// Get the list and ensure it's the correct type
				if list, ok := newI[listName].([]interface{}); ok {
					index, _ := strconv.Atoi(k[idx+1:])
					if index < len(list) {
						list[index] = v
					}
				}
				continue
			}
		}

		// Regular (non-list) attribute
		newI[k] = v
	}

	if resource, ok := provider.ResourcesMap[resourceName]; ok {
		return resource.Diff(
			context.Background(),
			&terraform.InstanceState{
				Attributes: oldValue,
			},
			&terraform.ResourceConfig{
				Config: newI,
			},
			provider.Meta(),
		)
	} else {
		return nil, fmt.Errorf("Resource %s not found in provider", resourceName)
	}
}
