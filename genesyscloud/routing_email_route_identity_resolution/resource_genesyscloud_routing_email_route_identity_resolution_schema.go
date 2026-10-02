package routing_email_route_identity_resolution

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	authDivision "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/auth_division"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_register"
)

const ResourceType = "genesyscloud_routing_email_route_identity_resolution"

// SetRegistrar registers all the resources and exporters in the package.
func SetRegistrar(regInstance registrar.Registrar) {
	regInstance.RegisterResource(ResourceType, ResourceRoutingEmailRouteIdentityResolution())
	regInstance.RegisterExporter(ResourceType, RoutingEmailRouteIdentityResolutionExporter())
}

func ResourceRoutingEmailRouteIdentityResolution() *schema.Resource {
	return &schema.Resource{
		Description:   `Genesys Cloud routing email route identity resolution settings. Destroy restores the routing email route to its default identity resolution configuration (resolve_identities = true, unassigned division).`,
		CreateContext: provider.CreateWithPooledClient(createRoutingEmailRouteIdentityResolution),
		ReadContext:   provider.ReadWithPooledClient(readRoutingEmailRouteIdentityResolution),
		UpdateContext: provider.UpdateWithPooledClient(updateRoutingEmailRouteIdentityResolution),
		DeleteContext: provider.DeleteWithPooledClient(deleteRoutingEmailRouteIdentityResolution),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 0,
		Schema: map[string]*schema.Schema{
			"domain_name": {
				Description: "Name of the email domain this identity resolution config belongs to.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"route_id": {
				Description: "ID of the routing email route this identity resolution config belongs to.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"resolve_identities": {
				Description: "Whether identities should be resolved.",
				Type:        schema.TypeBool,
				Required:    true,
			},
			"division_id": {
				Description:      "The division to use when performing identity resolution. If not set, * means the unassigned (star) division. '*' may also be set explicitly for the same behavior.",
				Type:             schema.TypeString,
				Optional:         true,
				DiffSuppressFunc: suppressUnassignedDivisionIdDiff,
			},
		},
	}
}

func RoutingEmailRouteIdentityResolutionExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllRoutingEmailRouteIdentityResolution),
		RefAttrs: map[string]*resourceExporter.RefAttrSettings{
			"domain_name": {RefType: "genesyscloud_routing_email_domain"},
			"route_id":    {RefType: "genesyscloud_routing_email_route"},
			"division_id": {
				RefType:   authDivision.ResourceType,
				AltValues: []string{"*"},
			},
		},
	}
}
