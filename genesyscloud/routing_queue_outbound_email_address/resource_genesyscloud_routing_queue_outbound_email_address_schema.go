package routing_queue_outbound_email_address

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	registrar "terraform-provider-genesyscloud/genesyscloud/resource_register"
)

const resourceName = "genesyscloud_routing_queue_outbound_email_address"

// SetRegistrar registers all of the resources, datasources and exporters in the package
func SetRegistrar(regInstance registrar.Registrar) {
	regInstance.RegisterResource(resourceName, ResourceRoutingQueueOutboundEmailAddress())
	//regInstance.RegisterExporter(resourceName, OutboundCampaignruleExporter())
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
			"outbound_email_address": {
				Description: "The outbound email address settings for the queue.",
				Type:        schema.TypeList,
				MinItems:    1,
				MaxItems:    1,
				Required:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
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
				},
			},
		},
	}
}
