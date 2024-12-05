package routing_utilization_label

import (
	"fmt"
	"strings"
)

func GenerateRoutingUtilizationLabelResource(resourceLabel string, name string, dependsOnResource string) string {
	dependsOn := ""

	if dependsOnResource != "" {
		dependsOn = fmt.Sprintf("depends_on=[%s.%s]", ResourceType, dependsOnResource)
	}

	return fmt.Sprintf(`resource "%s" "%s" {
		name = "%s"
		%s
	}
	`, ResourceType, resourceLabel, name, dependsOn)
}

func GenerateLabelUtilization(
	labelResource string,
	maxCapacity string,
	interruptingLabelResourceLabels ...string) string {

	interruptingLabelResources := make([]string, 0)
	for _, resourceLabel := range interruptingLabelResourceLabels {
		interruptingLabelResources = append(interruptingLabelResources, ResourceType+"."+resourceLabel+".id")
	}

	return fmt.Sprintf(`label_utilizations {
		label_id = %s.%s.id
		maximum_capacity = %s
		interrupting_label_ids = [%s]
	}
	`, ResourceType, labelResource, maxCapacity, strings.Join(interruptingLabelResources, ","))
}
