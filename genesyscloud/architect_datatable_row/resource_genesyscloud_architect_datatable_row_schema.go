package architect_datatable_row

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "terraform-provider-genesyscloud/genesyscloud/resource_register"
	"terraform-provider-genesyscloud/genesyscloud/util"
)

const resourceName = "genesyscloud_architect_datatable_row"

// SetRegistrar registers all of the resources, datasources and exporters in the package
func SetRegistrar(regInstance registrar.Registrar) {
	regInstance.RegisterResource(resourceName, ResourceArchitectDatatableRow())
	//No Datasource defined
	regInstance.RegisterExporter(resourceName, ArchitectDatatableRowExporter())
}

func ArchitectDatatableRowExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllArchitectDatatableRows),
		RefAttrs: map[string]*resourceExporter.RefAttrSettings{
			"datatable_id": {RefType: "genesyscloud_architect_datatable"},
		},
		JsonEncodeAttributes: []string{"properties_json"},
	}
}

func ResourceArchitectDatatableRow() *schema.Resource {
	return &schema.Resource{
		Description: "Genesys Cloud Architect Datatable Row",

		CreateContext: provider.CreateWithPooledClient(createArchitectDatatableRow),
		ReadContext:   provider.ReadWithPooledClient(readArchitectDatatableRow),
		UpdateContext: provider.UpdateWithPooledClient(updateArchitectDatatableRow),
		DeleteContext: provider.DeleteWithPooledClient(deleteArchitectDatatableRow),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"datatable_id": {
				Description: "Datatable ID that contains this row. If this is changed, a new row is created.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"key_value": {
				Description: "Value for this row's key. If this is changed, a new row is created.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"properties_json": {
				Description:      "JSON object containing properties and values for this row. Defaults will be set for missing properties.",
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				DiffSuppressFunc: util.SuppressEquivalentJsonDiffs,
			},
		},
		CustomizeDiff: customizeDatatableRowDiff,
	}
}
