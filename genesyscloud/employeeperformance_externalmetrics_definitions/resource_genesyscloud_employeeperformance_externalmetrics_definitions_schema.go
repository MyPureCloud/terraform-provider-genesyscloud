package employeeperformance_externalmetrics_definitions

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"terraform-provider-genesyscloud/genesyscloud/provider"
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
const resourceName = "genesyscloud_employeeperformance_externalmetrics_definitions"

// SetRegistrar registers all of the resources, datasources and exporters in the package
func SetRegistrar(regInstance registrar.Registrar) {
	regInstance.RegisterResource(resourceName, ResourceEmployeeperformanceExternalmetricsDefinition())
	regInstance.RegisterDataSource(resourceName, DataSourceEmployeeperformanceExternalmetricsDefinition())
	regInstance.RegisterExporter(resourceName, EmployeeperformanceExternalmetricsDefinitionExporter())
}

// ResourceEmployeeperformanceExternalmetricsDefinition registers the genesyscloud_employeeperformance_externalmetrics_definitions resource with Terraform
func ResourceEmployeeperformanceExternalmetricsDefinition() *schema.Resource {
	return &schema.Resource{
		Description: `Genesys Cloud employeeperformance externalmetrics definition`,

		CreateContext: provider.CreateWithPooledClient(createEmployeeperformanceExternalmetricsDefinition),
		ReadContext:   provider.ReadWithPooledClient(readEmployeeperformanceExternalmetricsDefinition),
		UpdateContext: provider.UpdateWithPooledClient(updateEmployeeperformanceExternalmetricsDefinition),
		DeleteContext: provider.DeleteWithPooledClient(deleteEmployeeperformanceExternalmetricsDefinition),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			`name`: {
				Description: `The name of the External Metric Definition`,
				Required:    true,
				Type:        schema.TypeString,
			},
			`precision`: {
				Description:  `The decimal precision of the External Metric Definition. Must be at least 0 and at most 5`,
				Required:     true,
				Type:         schema.TypeInt,
				ValidateFunc: validation.IntBetween(0, 5),
			},
			`default_objective_type`: {
				Description:  `The default objective type of the External Metric Definition`,
				Required:     true,
				Type:         schema.TypeString,
				ValidateFunc: validation.StringInSlice([]string{`HigherIsBetter`, `LowerIsBetter`, `TargetArea`}, false),
			},
			`enabled`: {
				Description: `True if the External Metric Definition is enabled`,
				Required:    true,
				Type:        schema.TypeBool,
			},
			`unit`: {
				Description:  `The unit of the External Metric Definition. Note: Changing the unit property will cause the external metric object to be dropped and recreated with a new ID.`,
				Required:     true,
				ForceNew:     true,
				Type:         schema.TypeString,
				ValidateFunc: validation.StringInSlice([]string{`Seconds`, `Percent`, `Number`, `Currency`}, false),
			},
			`unit_definition`: {
				Description: `The unit definition of the External Metric Definition. Note: Changing the unit definition property will cause the external metric object to be dropped and recreated with a new ID.`,
				Optional:    true,
				ForceNew:    true,
				Type:        schema.TypeString,
			},
		},
	}
}

// EmployeeperformanceExternalmetricsDefinitionExporter returns the resourceExporter object used to hold the genesyscloud_employeeperformance_externalmetrics_definition exporter's config
func EmployeeperformanceExternalmetricsDefinitionExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllAuthEmployeeperformanceExternalmetricsDefinitions),
		AllowZeroValues:  []string{"precision"},
	}
}

// DataSourceEmployeeperformanceExternalmetricsDefinition registers the genesyscloud_employeeperformance_externalmetrics_definition data source
func DataSourceEmployeeperformanceExternalmetricsDefinition() *schema.Resource {
	return &schema.Resource{
		Description: `Data source for Genesys Cloud Employeeperformance Externalmetrics Definition. Select a Employeeperformance Externalmetrics Definition by name.`,
		ReadContext: provider.ReadWithPooledClient(dataSourceEmployeeperformanceExternalmetricsDefinitionRead),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Description: `Employeeperformance Externalmetrics Definition name.`,
				Type:        schema.TypeString,
				Optional:    true,
			},
		},
	}
}
