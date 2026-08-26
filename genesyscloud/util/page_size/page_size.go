package page_size

import (
	"log"
	"os"
	"strconv"
)

const envPrefix = "GENESYSCLOUD_PAGE_SIZE_"

// EnvVarName returns the environment variable used to override list page size for a resource type.
// Example: GENESYSCLOUD_PAGE_SIZE_genesyscloud_routing_wrapupcode
func EnvVarName(resourceType string) string {
	return envPrefix + resourceType
}

// ForResource returns the configured list page size for the given Terraform resource type.
// When GENESYSCLOUD_PAGE_SIZE_<resourceType> is set to a positive integer, that value is used.
// Otherwise defaultPageSize is returned.
func ForResource(resourceType string, defaultPageSize int) int {
	envVar := EnvVarName(resourceType)
	value, exists := os.LookupEnv(envVar)
	if !exists {
		return defaultPageSize
	}

	if value == "" {
		log.Printf("[WARN] %s is set but empty; using default page size %d for %s", envVar, defaultPageSize, resourceType)
		return defaultPageSize
	}

	pageSize, err := strconv.Atoi(value)
	if err != nil || pageSize <= 0 {
		log.Printf("[WARN] invalid %s value %q; using default page size %d for %s", envVar, value, defaultPageSize, resourceType)
		return defaultPageSize
	}

	return pageSize
}
