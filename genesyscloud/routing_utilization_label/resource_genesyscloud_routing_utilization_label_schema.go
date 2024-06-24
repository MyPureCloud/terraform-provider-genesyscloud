package routing_utilization_label

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "terraform-provider-genesyscloud/genesyscloud/resource_register"
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

func DataSourceRoutingUtilizationLabel() *schema.Resource {
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
