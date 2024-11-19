package task_management_workbin

import (
	"fmt"
)

// GenerateWorkbinResource is a public util method to generate a workbin terraform resource for testing
func GenerateWorkbinResource(resourceId string, name string, description string, divisionIdRef string) string {
	return fmt.Sprintf(`resource "%s" "%s" {
		name = "%s"
		description = "%s"
		division_id = %s
	}
	`, resourceName, resourceId, name, description, divisionIdRef)
}
