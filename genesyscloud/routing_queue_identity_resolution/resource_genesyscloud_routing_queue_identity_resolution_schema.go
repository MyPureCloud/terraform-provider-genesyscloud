package routing_queue_identity_resolution

// @team: Assignment
// @chat: #genesys-cloud-acd-routing
// @pm: Rob Blane
// @jira: RELATE-25224
// @description: Routing configuration service for queues, skills, wrapup codes, and utilization settings. Manages how contacts are distributed to agents based on skills, capacity, and routing rules across all interaction channels.

import (
	authDivision "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/auth_division"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_register"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const ResourceType = "genesyscloud_routing_queue_identity_resolution"

// SetRegistrar registers all the resources and exporters in the package.
func SetRegistrar(regInstance registrar.Registrar) {
	regInstance.RegisterResource(ResourceType, ResourceRoutingQueueIdentityResolution())
	regInstance.RegisterExporter(ResourceType, RoutingQueueIdentityResolutionExporter())
}

var callOnBehalfOfQueueSchema = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"resolve_identities": {
			Description: "Whether the channel should resolve identities.",
			Type:        schema.TypeBool,
			Required:    true,
		},
		"division_id": {
			Description:      "Division ID used during identity resolution. If not set, * will be used for all divisions. '*' may also be set explicitly for all divisions.",
			Type:             schema.TypeString,
			Optional:         true,
			DiffSuppressFunc: suppressAllDivisionsDivisionIdDiff,
		},
	},
}

func ResourceRoutingQueueIdentityResolution() *schema.Resource {
	return &schema.Resource{
		Description:   `Genesys Cloud routing queue identity resolution settings. Destroy restores the queue to its default identity resolution configuration (resolve_identities = true, all divisions).`,
		CreateContext: provider.CreateWithPooledClient(createRoutingQueueIdentityResolution),
		ReadContext:   provider.ReadWithPooledClient(readRoutingQueueIdentityResolution),
		UpdateContext: provider.UpdateWithPooledClient(updateRoutingQueueIdentityResolution),
		DeleteContext: provider.DeleteWithPooledClient(deleteRoutingQueueIdentityResolution),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"queue_id": {
				Description: "ID of the routing queue.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"call_on_behalf_of_queue": {
				Description: "Identity resolution settings for outbound calls placed on behalf of the queue.",
				Type:        schema.TypeList,
				MaxItems:    1,
				Required:    true,
				Elem:        callOnBehalfOfQueueSchema,
			},
		},
	}
}

func RoutingQueueIdentityResolutionExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllRoutingQueueIdentityResolution),
		RefAttrs: map[string]*resourceExporter.RefAttrSettings{
			"queue_id": {RefType: "genesyscloud_routing_queue"},
			"call_on_behalf_of_queue.division_id": {
				RefType:   authDivision.ResourceType,
				AltValues: []string{"*"},
			},
		},
	}
}
