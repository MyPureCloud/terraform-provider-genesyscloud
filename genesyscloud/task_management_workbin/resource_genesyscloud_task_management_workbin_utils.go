package task_management_workbin

import (
	"fmt"
)

// GenerateWorkbinResource is a public util method to generate a workbin terraform resource for testing
func GenerateWorkbinResource(resourceLabel string, name string, description string, divisionIdRef string) string {
	return fmt.Sprintf(`resource "%s" "%s" {
		name = "%s"
		description = "%s"
		division_id = %s
	}
	`, ResourceType, resourceLabel, name, description, divisionIdRef)
}
