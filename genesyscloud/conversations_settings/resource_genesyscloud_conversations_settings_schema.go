package conversations_settings

import (
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_register"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const ResourceType = "genesyscloud_conversations_settings"

// SetRegistrar registers all of the resources and exporters in the package
func SetRegistrar(regInstance registrar.Registrar) {
	regInstance.RegisterResource(ResourceType, ResourceConversationsSettings())
	regInstance.RegisterExporter(ResourceType, ConversationsSettingsExporter())
}

// ResourceConversationsSettings registers the genesyscloud_conversations_settings resource
func ResourceConversationsSettings() *schema.Resource {
	return &schema.Resource{
		Description: "Genesys Cloud organization conversations settings",

		CreateContext: provider.CreateWithPooledClient(createConversationsSettings),
		ReadContext:   provider.ReadWithPooledClient(readConversationsSettings),
		UpdateContext: provider.UpdateWithPooledClient(updateConversationsSettings),
		DeleteContext: provider.DeleteWithPooledClient(deleteConversationsSettings),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"communication_based_acw": {
				Description: "Communication Based ACW",
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
			},
			"include_non_agent_conversation_summary": {
				Description: "Display communication summary",
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
			},
			"allow_callback_queue_selection": {
				Description: "Allow Callback Queue Selection",
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
			},
			"callbacks_inherit_routing_from_inbound_call": {
				Description: "Inherit callback routing data from inbound calls",
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
			},
			"complete_acw_when_agent_transitions_offline": {
				Description: "Complete ACW When Agent Transitions Offline",
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
			},
			"total_active_callback": {
				Description: "Exclude the 'interacting' duration from the handle calculations of callbacks",
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
			},
		},
	}
}

// ConversationsSettingsExporter returns the resourceExporter object used to hold the genesyscloud_conversations_settings exporter's config
func ConversationsSettingsExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllConversationsSettings),
	}
}
