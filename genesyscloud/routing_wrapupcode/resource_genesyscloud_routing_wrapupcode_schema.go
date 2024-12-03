package routing_wrapupcode

import (
	"terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "terraform-provider-genesyscloud/genesyscloud/resource_register"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const ResourceType = "genesyscloud_routing_wrapupcode"

// SetRegistrar registers all of the resources, datasources and exporters in the pakage
func SetRegistrar(regInstance registrar.Registrar) {
	regInstance.RegisterResource(ResourceType, ResourceRoutingWrapupCode())
	regInstance.RegisterDataSource(ResourceType, DataSourceRoutingWrapupCode())
	regInstance.RegisterExporter(ResourceType, RoutingWrapupCodeExporter())
}

func RoutingWrapupCodeExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllRoutingWrapupCodes),
		RefAttrs: map[string]*resourceExporter.RefAttrSettings{
			"division_id": {RefType: "genesyscloud_auth_division"},
		},
	}
}

func ResourceRoutingWrapupCode() *schema.Resource {
	return &schema.Resource{
		Description: "Genesys Cloud Routing Wrapup Code",

		CreateContext: provider.CreateWithPooledClient(createRoutingWrapupCode),
		ReadContext:   provider.ReadWithPooledClient(readRoutingWrapupCode),
		UpdateContext: provider.UpdateWithPooledClient(updateRoutingWrapupCode),
		DeleteContext: provider.DeleteWithPooledClient(deleteRoutingWrapupCode),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Wrapup Code name.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"division_id": {
				Description: "The division to which this routing wrapupcode will belong. If not set, * will be used to indicate all divisions.",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},
		},
	}
}

func DataSourceRoutingWrapupCode() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for Genesys Cloud Wrap-up Code. Select a wrap-up code by name",
		ReadContext: provider.ReadWithPooledClient(dataSourceRoutingWrapupcodeRead),
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Wrap-up code name.",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}
