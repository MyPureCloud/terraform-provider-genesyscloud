package routing_utilization_label

import (
<<<<<<< HEAD
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
=======
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/mypurecloud/platform-client-sdk-go/v130/platformclientv2"
	"strings"
>>>>>>> f33044e5 (refactor routing utilization label)
	"terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "terraform-provider-genesyscloud/genesyscloud/resource_register"
)

const resourceName = "genesyscloud_routing_utilization_label"

// SetRegistrar registers all the resources and exporters in the package
func SetRegistrar(regInstance registrar.Registrar) {
	regInstance.RegisterResource(resourceName, ResourceRoutingUtilizationLabel())
<<<<<<< HEAD
	regInstance.RegisterDataSource(resourceName, dataSourceRoutingUtilizationLabel())
=======
>>>>>>> f33044e5 (refactor routing utilization label)
	regInstance.RegisterExporter(resourceName, RoutingUtilizationLabelExporter())
}

func ResourceRoutingUtilizationLabel() *schema.Resource {
	return &schema.Resource{
		Description: "Genesys Cloud Routing Utilization Label. This resource is not yet widely available. Only use it if the feature is enabled.",

		CreateContext: provider.CreateWithPooledClient(createRoutingUtilizationLabel),
		ReadContext:   provider.ReadWithPooledClient(readRoutingUtilizationLabel),
		UpdateContext: provider.UpdateWithPooledClient(updateRoutingUtilizationLabel),
		DeleteContext: provider.DeleteWithPooledClient(deleteRoutingUtilizationLabel),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Label name.",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}

func dataSourceRoutingUtilizationLabel() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for Genesys Cloud Routing Utilization Labels. Select a label by name.",
		ReadContext: provider.ReadWithPooledClient(dataSourceRoutingUtilizationLabelRead),
		Schema: map[string]*schema.Schema{
			"name": {
				Description:  "Label name.",
				Type:         schema.TypeString,
				ValidateFunc: validation.StringDoesNotContainAny("*"),
				Required:     true,
			},
		},
	}
}

func RoutingUtilizationLabelExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllRoutingUtilizationLabels),
	}
}
<<<<<<< HEAD
=======

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
	api := platformclientv2.NewRoutingApiWithConfig(sdkConfig) // the variable sdkConfig exists at a package level in ./genesyscloud and is already authorized
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
>>>>>>> f33044e5 (refactor routing utilization label)
