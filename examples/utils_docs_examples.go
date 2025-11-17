package examples

import (
	"os"
	"strings"
)

// getTestResourceTypes returns the list of resource types to test.
// It checks for TEST_RESOURCE_TYPES environment variable first,
// then falls back to testing all resources if not specified.
func getTestResourceTypes() []string {
	if envResources := os.Getenv("TEST_RESOURCE_TYPES"); envResources != "" {
		return strings.Split(envResources, ",")
	}
	return []string{}
}
