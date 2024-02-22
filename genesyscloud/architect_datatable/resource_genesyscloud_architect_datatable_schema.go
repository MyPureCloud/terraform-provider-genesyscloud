package architect_datatable

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "terraform-provider-genesyscloud/genesyscloud/resource_register"
)

const resourceName = "genesyscloud_architect_datatable"

// SetRegistrar registers all of the resources, datasources and exporters in the package
func SetRegistrar(regInstance registrar.Registrar) {
	regInstance.RegisterResource(resourceName, ResourceArchitectDatatable())
	regInstance.RegisterDataSource(resourceName, DataSourceArchitectDatatable())
	regInstance.RegisterExporter(resourceName, ArchitectDatatableExporter())
}

var (
	datatableProperty = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Name of the property.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"type": {
				Description:  "Type of the property (boolean | string | integer | number).",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"boolean", "string", "integer", "number"}, false),
			},
			"title": {
				Description: "Display title of the property.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"default": {
				Description: "Default value of the property. This is converted to the proper type for non-strings (e.g. set 'true' or 'false' for booleans).",
				Type:        schema.TypeString,
				Optional:    true,
			},
		},
	}
)

func ResourceArchitectDatatable() *schema.Resource {
	return &schema.Resource{
		Description: "Genesys Cloud Architect Datatables",

		CreateContext: provider.CreateWithPooledClient(createArchitectDatatable),
		ReadContext:   provider.ReadWithPooledClient(readArchitectDatatable),
		UpdateContext: provider.UpdateWithPooledClient(updateArchitectDatatable),
		DeleteContext: provider.DeleteWithPooledClient(deleteArchitectDatatable),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Name of the architect_datatable.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"division_id": {
				Description: "The division to which this architect_datatable will belong. If not set, the home division will be used.",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},
			"description": {
				Description: "Description of the architect_datatable.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"properties": {
				Description: "Schema properties of the architect_datatable. This must at a minimum contain a string property 'key' that will serve as the row key. Properties cannot be removed from a schema once they have been added",
				Type:        schema.TypeList,
				Required:    true,
				MinItems:    1,
				Elem:        datatableProperty,
			},
		},
	}
}

func DataSourceArchitectDatatable() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for Genesys Cloud Architect Datatables. Select an architect architect_datatable by name.",
		ReadContext: provider.ReadWithPooledClient(DataSourceArchitectDatatableRead),
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Datatable name.",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}

func ArchitectDatatableExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllArchitectDatatables),
		RefAttrs: map[string]*resourceExporter.RefAttrSettings{
			"division_id": {RefType: "genesyscloud_auth_division"},
		},
	}
}
