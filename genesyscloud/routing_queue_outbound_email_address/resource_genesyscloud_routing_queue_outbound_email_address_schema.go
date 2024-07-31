package routing_queue_outbound_email_address

import (
	"terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "terraform-provider-genesyscloud/genesyscloud/resource_register"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const resourceName = "genesyscloud_routing_queue_outbound_email_address"

// SetRegistrar registers all the resources, datasources and exporters in the package
func SetRegistrar(regInstance registrar.Registrar) {
	regInstance.RegisterResource(resourceName, ResourceRoutingQueueOutboundEmailAddress())
	regInstance.RegisterExporter(resourceName, OutboundRoutingQueueOutboundEmailAddressExporter())
}

func ResourceRoutingQueueOutboundEmailAddress() *schema.Resource {
	return &schema.Resource{
		Description:   `Genesys Cloud Routing Queue Outbound Email Address`,
		CreateContext: provider.CreateWithPooledClient(createRoutingQueueOutboundEmailAddress),
		ReadContext:   provider.ReadWithPooledClient(readRoutingQueueOutboundEmailAddress),
		UpdateContext: provider.UpdateWithPooledClient(updateRoutingQueueOutboundEmailAddress),
		DeleteContext: provider.DeleteWithPooledClient(deleteRoutingQueueOutboundEmailAddress),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"queue_id": {
				Description: "The routing queue to which the outbound email address is for.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"domain_id": {
				Description: "Unique ID of the email domain. e.g. \"test.example.com\"",
				Type:        schema.TypeString,
				Required:    true,
			},
			"route_id": {
				Description: "Unique ID of the email route.",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}

func OutboundRoutingQueueOutboundEmailAddressExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllAuthRoutingQueueOutboundEmailAddress),
		RefAttrs: map[string]*resourceExporter.RefAttrSettings{
			"queue_id":  {RefType: "genesyscloud_routing_queue"},
			"route_id":  {RefType: "genesyscloud_routing_email_route"},
			"domain_id": {RefType: "genesyscloud_routing_email_domain"},
		},
	}
}
