package routing_utilization_label

import (
	"fmt"
	"github.com/mypurecloud/platform-client-sdk-go/v131/platformclientv2"
	"strings"
	"terraform-provider-genesyscloud/genesyscloud/provider"
)

func GenerateRoutingUtilizationLabelResource(resourceID string, name string, dependsOnResource string) string {
	dependsOn := ""

	if dependsOnResource != "" {
		dependsOn = fmt.Sprintf("depends_on=[genesyscloud_routing_utilization_label.%s]", dependsOnResource)
	}

	return fmt.Sprintf(`resource "genesyscloud_routing_utilization_label" "%s" {
		name = "%s"
		%s
	}
	`, resourceID, name, dependsOn)
}

func CheckIfLabelsAreEnabled() error { // remove once the feature is globally enabled
	sdkConfig, err := provider.AuthorizeSdk()
	if err != nil {
		return err
	}

	api := platformclientv2.NewRoutingApiWithConfig(sdkConfig)
	_, resp, _ := api.GetRoutingUtilizationLabels(100, 1, "", "")
	if resp.StatusCode == 501 {
		return fmt.Errorf("feature is not yet implemented in this org.")
	}
	return nil
}

func GenerateLabelUtilization(
	labelResource string,
	maxCapacity string,
	interruptingLabelResourceNames ...string) string {

	interruptingLabelResources := make([]string, 0)
	for _, resourceName := range interruptingLabelResourceNames {
		interruptingLabelResources = append(interruptingLabelResources, "genesyscloud_routing_utilization_label."+resourceName+".id")
	}

	return fmt.Sprintf(`label_utilizations {
		label_id = genesyscloud_routing_utilization_label.%s.id
		maximum_capacity = %s
		interrupting_label_ids = [%s]
	}
	`, labelResource, maxCapacity, strings.Join(interruptingLabelResources, ","))
}
