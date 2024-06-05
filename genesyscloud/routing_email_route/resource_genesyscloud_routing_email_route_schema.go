package routing_email_route

import (
	"terraform-provider-genesyscloud/genesyscloud/provider"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "terraform-provider-genesyscloud/genesyscloud/resource_register"
)

/*
resource_genesycloud_routing_email_route_schema.go holds four functions within it:

1.  The registration code that registers the Datasource, Resource and Exporter for the package.
2.  The resource schema definitions for the routing_email_route resource.
3.  The datasource schema definitions for the routing_email_route datasource.
4.  The resource exporter configuration for the routing_email_route exporter.
*/
const resourceName = "genesyscloud_routing_email_route"

var (
	bccEmailResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"email": {
				Description: "Email address.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"name": {
				Description: "Name associated with the email.",
				Type:        schema.TypeString,
				Optional:    true,
			},
		},
	}
)

// SetRegistrar registers all of the resources, datasources and exporters in the package
func SetRegistrar(regInstance registrar.Registrar) {
	regInstance.RegisterResource(resourceName, ResourceRoutingEmailRoute())
	regInstance.RegisterExporter(resourceName, RoutingEmailRouteExporter())
	regInstance.RegisterDataSource(resourceName, DataSourceRoutingEmailRoute())
}

func ResourceRoutingEmailRoute() *schema.Resource {
	return &schema.Resource{
		Description: "Genesys Cloud Routing Email Domain Route",

		CreateContext: provider.CreateWithPooledClient(createRoutingEmailRoute),
		ReadContext:   provider.ReadWithPooledClient(readRoutingEmailRoute),
		UpdateContext: provider.UpdateWithPooledClient(updateRoutingEmailRoute),
		DeleteContext: provider.DeleteWithPooledClient(deleteRoutingEmailRoute),
		Importer: &schema.ResourceImporter{
			StateContext: importRoutingEmailRoute,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"domain_id": {
				Description: "ID of the routing domain such as: 'example.com'. Changing the domain_id attribute will cause the email_route object to be dropped and recreated with a new ID.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"pattern": {
				Description: "The search pattern that the mailbox name should match.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"from_name": {
				Description: "The sender name to use for outgoing replies.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"from_email": {
				Description:   "The sender email to use for outgoing replies. This should not be set if reply_email_address is specified.",
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"reply_email_address"},
			},
			"queue_id": {
				Description:   "The queue to route the emails to. This should not be set if a flow_id is specified.",
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"flow_id"},
			},
			"priority": {
				Description: "The priority to use for routing.",
				Type:        schema.TypeInt,
				Optional:    true,
			},
			"skill_ids": {
				Description: "The skills to use for routing.",
				Type:        schema.TypeSet,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"language_id": {
				Description: "The language to use for routing.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"flow_id": {
				Description:   "The flow to use for processing the email. This should not be set if a queue_id is specified.",
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"queue_id"},
			},
			"reply_email_address": {
				Description:   "The route to use for email replies. This should not be set if from_email or auto_bcc are specified.",
				Type:          schema.TypeList,
				MaxItems:      1,
				Optional:      true,
				ConflictsWith: []string{"from_email"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"domain_id": {
							Description: "Domain of the route.",
							Type:        schema.TypeString,
							Required:    true,
						},
						"route_id": {
							Description: "ID of the route.",
							Type:        schema.TypeString,
							Required:    false,
							Optional:    true,
						},
						"self_reference_route": {
							Description: `Use this route as the reply email address. If true you will use the route id for this resource as the reply and you
							              can not set a route. If you set this value to false (or leave the attribute off)you must set a route id.`,
							Type:     schema.TypeBool,
							Required: false,
							Optional: true,
							Default:  false,
						},
					},
				},
			},
			"auto_bcc": {
				Description:   "The recipients that should be automatically blind copied on outbound emails associated with this route. This should not be set if reply_email_address is specified.",
				Type:          schema.TypeSet,
				Optional:      true,
				Elem:          bccEmailResource,
				ConflictsWith: []string{"reply_email_address"},
			},
			"spam_flow_id": {
				Description: "The flow to use for processing inbound emails that have been marked as spam.",
				Type:        schema.TypeString,
				Optional:    true,
			},
		},
	}
}

func DataSourceRoutingEmailRoute() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for Genesys Cloud Routing Email Route. Select a routing email route by pattern and domain ID.",
		ReadContext: provider.ReadWithPooledClient(dataSourceRoutingEmailRouteRead),
		Schema: map[string]*schema.Schema{
			"pattern": {
				Description: "Routing pattern.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"domain_id": {
				Description: "Domain of the route.",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}

// RoutingEmailRouteExporter returns the resourceExporter object used to hold the genesyscloud_routing_email_route exporter's config
func RoutingEmailRouteExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllRoutingEmailRoutes),
		RefAttrs: map[string]*resourceExporter.RefAttrSettings{
			"domain_id":                     {RefType: "genesyscloud_routing_email_domain"},
			"queue_id":                      {RefType: "genesyscloud_routing_queue"},
			"skill_ids":                     {RefType: "genesyscloud_routing_skill"},
			"language_id":                   {RefType: "genesyscloud_routing_language"},
			"flow_id":                       {RefType: "genesyscloud_flow"},
			"spam_flow_id":                  {RefType: "genesyscloud_flow"},
			"reply_email_address.domain_id": {RefType: "genesyscloud_routing_email_domain"},
			"reply_email_address.route_id":  {RefType: "genesyscloud_routing_email_route"},
		},
		RemoveIfMissing: map[string][]string{
			"reply_email_address": {"route_id", "self_reference_route"},
		},
		CustomAttributeResolver: map[string]*resourceExporter.RefAttrCustomResolver{
			"reply_email_address.self_reference_route": {ResolverFunc: resourceExporter.ReplyEmailAddressSelfReferenceRouteExporterResolver},
		},
	}
}
