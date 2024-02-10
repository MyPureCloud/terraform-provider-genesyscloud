package employeeperformance_externalmetrics_definition

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	gcloud "terraform-provider-genesyscloud/genesyscloud"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "terraform-provider-genesyscloud/genesyscloud/resource_register"
)

/*
resource_genesycloud_employeeperformance_externalmetrics_definition_schema.go holds four functions within it:

1.  The registration code that registers the Datasource, Resource and Exporter for the package.
2.  The resource schema definitions for the employeeperformance_externalmetrics_definition resource.
3.  The datasource schema definitions for the employeeperformance_externalmetrics_definition datasource.
4.  The resource exporter configuration for the employeeperformance_externalmetrics_definition exporter.
*/
const resourceName = "genesyscloud_employeeperformance_externalmetrics_definition"

// SetRegistrar registers all of the resources, datasources and exporters in the package
func SetRegistrar(regInstance registrar.Registrar) {
	regInstance.RegisterResource(resourceName, ResourceEmployeeperformanceExternalmetricsDefinition())
	regInstance.RegisterDataSource(resourceName, DataSourceEmployeeperformanceExternalmetricsDefinition())
	regInstance.RegisterExporter(resourceName, EmployeeperformanceExternalmetricsDefinitionExporter())
}

// ResourceEmployeeperformanceExternalmetricsDefinition registers the genesyscloud_employeeperformance_externalmetrics_definition resource with Terraform
func ResourceEmployeeperformanceExternalmetricsDefinition() *schema.Resource {
	return &schema.Resource{
		Description: `Genesys Cloud employeeperformance externalmetrics definition`,

		CreateContext: gcloud.CreateWithPooledClient(createEmployeeperformanceExternalmetricsDefinition),
		ReadContext:   gcloud.ReadWithPooledClient(readEmployeeperformanceExternalmetricsDefinition),
		UpdateContext: gcloud.UpdateWithPooledClient(updateEmployeeperformanceExternalmetricsDefinition),
		DeleteContext: gcloud.DeleteWithPooledClient(deleteEmployeeperformanceExternalmetricsDefinition),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema:        map[string]*schema.Schema{},
	}
}

// EmployeeperformanceExternalmetricsDefinitionExporter returns the resourceExporter object used to hold the genesyscloud_employeeperformance_externalmetrics_definition exporter's config
func EmployeeperformanceExternalmetricsDefinitionExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: gcloud.GetAllWithPooledClient(getAllAuthEmployeeperformanceExternalmetricsDefinitions),
		RefAttrs:         map[string]*resourceExporter.RefAttrSettings{
			// TODO: Add any reference attributes here
		},
	}
}

// DataSourceEmployeeperformanceExternalmetricsDefinition registers the genesyscloud_employeeperformance_externalmetrics_definition data source
func DataSourceEmployeeperformanceExternalmetricsDefinition() *schema.Resource {
	return &schema.Resource{
		Description: `Genesys Cloud employeeperformance externalmetrics definition data source. Select an employeeperformance externalmetrics definition by name`,
		ReadContext: gcloud.ReadWithPooledClient(dataSourceEmployeeperformanceExternalmetricsDefinitionRead),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Description: `employeeperformance externalmetrics definition name`,
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}
