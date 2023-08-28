package simple_routing_queue

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	gcloud "terraform-provider-genesyscloud/genesyscloud"
	registrar "terraform-provider-genesyscloud/genesyscloud/resource_register"
)

const resourceName = "genesyscloud_simple_routing_queue"

func SetRegistrar(l registrar.Registrar) {
	l.RegisterResource(resourceName, ResourceSimpleRoutingQueue())
	l.RegisterDataSource(resourceName, DataSourceSimpleRoutingQueue())
}

func ResourceSimpleRoutingQueue() *schema.Resource {
	return &schema.Resource{
		Description: "Genesys Cloud Simple Routing Queue",

		CreateContext: gcloud.CreateWithPooledClient(createSimpleRoutingQueue),
		ReadContext:   gcloud.ReadWithPooledClient(readSimpleRoutingQueue),
		UpdateContext: gcloud.UpdateWithPooledClient(updateSimpleRoutingQueue),
		DeleteContext: gcloud.DeleteWithPooledClient(deleteSimpleRoutingQueue),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "The name for our routing queue.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"calling_party_name": {
				Description: "The name to use for caller identification for outbound calls from this queue.\n",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"enable_transcription": {
				Description: "Indicates whether voice transcription is enabled for this queue.",
				Type:        schema.TypeBool,
				Optional:    true,
			},
		},
	}
}

func DataSourceSimpleRoutingQueue() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for Genesys Cloud Simple Routing Queues.",
		ReadContext: gcloud.ReadWithPooledClient(dataSourceSimpleRoutingQueueRead),
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "The queue name.",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}
