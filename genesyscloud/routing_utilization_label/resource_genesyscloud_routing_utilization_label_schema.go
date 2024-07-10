package routing_utilization_label

import (
	"fmt"
	"strings"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "terraform-provider-genesyscloud/genesyscloud/resource_register"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

const resourceName = "genesyscloud_routing_utilization_label"

// SetRegistrar registers all the resources and exporters in the package
func SetRegistrar(regInstance registrar.Registrar) {
	regInstance.RegisterResource(resourceName, ResourceRoutingUtilizationLabel())
	regInstance.RegisterDataSource(resourceName, DataSourceRoutingUtilizationLabel())
	regInstance.RegisterExporter(resourceName, RoutingUtilizationLabelExporter())
}

func ResourceRoutingUtilizationLabel() *schema.Resource {
	return &schema.Resource{
		Description: "Genesys Cloud Routing Utilization Label.",

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
				ValidateFunc: validation.All(
					validation.StringIsNotEmpty,
					stringDoesNotStartOrEndWithSpaces,
					validation.StringDoesNotContainAny("*"),
				),
				Required: true,
			},
		},
	}
}

func DataSourceRoutingUtilizationLabel() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for Genesys Cloud Routing Utilization Labels. Select a label by name.",
		ReadContext: provider.ReadWithPooledClient(dataSourceRoutingUtilizationLabelRead),
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Label name.",
				Type:        schema.TypeString,
				ValidateFunc: validation.All(
					validation.StringIsNotEmpty,
					stringDoesNotStartOrEndWithSpaces,
					validation.StringDoesNotContainAny("*"),
				),
				Required: true,
			},
		},
	}
}

func RoutingUtilizationLabelExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllRoutingUtilizationLabels),
	}
}

func stringDoesNotStartOrEndWithSpaces(input interface{}, k string) ([]string, []error) {
	inputAsString, ok := input.(string)
	if !ok {
		return nil, []error{fmt.Errorf("expected type of %q to be string", k)}
	}

	if len(strings.TrimSpace(inputAsString)) != len(inputAsString) {
		return nil, []error{fmt.Errorf("expected %q to not start or end with spaces", k)}
	}

	return nil, nil
}
