package conversations_messaging_integrations_open

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "terraform-provider-genesyscloud/genesyscloud/resource_register"
	"terraform-provider-genesyscloud/genesyscloud/util"
)

/*
resource_genesycloud_conversations_messaging_integrations_open_schema.go holds four functions within it:

1.  The registration code that registers the Datasource, Resource and Exporter for the package.
2.  The resource schema definitions for the conversations_messaging_integrations_open resource.
3.  The datasource schema definitions for the conversations_messaging_integrations_open datasource.
4.  The resource exporter configuration for the conversations_messaging_integrations_open exporter.
*/
const ResourceType = "genesyscloud_conversations_messaging_integrations_open"

// SetRegistrar registers all of the resources, datasources and exporters in the package
func SetRegistrar(regInstance registrar.Registrar) {
	regInstance.RegisterResource(ResourceType, ResourceConversationsMessagingIntegrationsOpen())
	regInstance.RegisterDataSource(ResourceType, DataSourceConversationsMessagingIntegrationsOpen())
	regInstance.RegisterExporter(ResourceType, ConversationsMessagingIntegrationsOpenExporter())
}

// ResourceConversationsMessagingIntegrationsOpen registers the genesyscloud_conversations_messaging_integrations_open resource with Terraform
func ResourceConversationsMessagingIntegrationsOpen() *schema.Resource {

	return &schema.Resource{
		Description: `Genesys Cloud conversations messaging integrations open`,

		CreateContext: provider.CreateWithPooledClient(createConversationsMessagingIntegrationsOpen),
		ReadContext:   provider.ReadWithPooledClient(readConversationsMessagingIntegrationsOpen),
		UpdateContext: provider.UpdateWithPooledClient(updateConversationsMessagingIntegrationsOpen),
		DeleteContext: provider.DeleteWithPooledClient(deleteConversationsMessagingIntegrationsOpen),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			`name`: {
				Description: `The name of the Open messaging integration.`,
				Required:    true,
				Type:        schema.TypeString,
			},
			`supported_content_id`: {
				Description: `Reference to supported content profile associated with the integration`,
				Optional:    true,
				Type:        schema.TypeString,
			},
			`messaging_setting_id`: {
				Description: `Messaging Setting for messaging platform integrations`,
				Optional:    true,
				Type:        schema.TypeString,
			},
			`outbound_notification_webhook_url`: {
				Description: `The outbound notification webhook URL for the Open messaging integration`,
				Required:    true,
				Type:        schema.TypeString,
			},
			`outbound_notification_webhook_signature_secret_token`: {
				Description: `The outbound notification webhook signature secret token. This token must be longer than 15 characters.`,
				Required:    true,
				Type:        schema.TypeString,
			},
			`webhook_headers`: {
				Description:      `The user specified headers for the Open messaging integration.`,
				Optional:         true,
				Type:             schema.TypeString,
				Computed:         true,
				DiffSuppressFunc: util.SuppressEquivalentJsonDiffs,
			},
		},
	}
}

// ConversationsMessagingIntegrationsOpenExporter returns the resourceExporter object used to hold the genesyscloud_conversations_messaging_integrations_open exporter's config
func ConversationsMessagingIntegrationsOpenExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllAuthConversationsMessagingIntegrationsOpens),
		RefAttrs: map[string]*resourceExporter.RefAttrSettings{
			`messaging_setting_id`: {
				RefType: `genesyscloud_conversations_messaging_settings`,
			},
			`supported_content_id`: {
				RefType: `genesyscloud_conversations_messaging_supportedcontent`,
			},
		},
	}
}

// DataSourceConversationsMessagingIntegrationsOpen registers the genesyscloud_conversations_messaging_integrations_open data source
func DataSourceConversationsMessagingIntegrationsOpen() *schema.Resource {
	return &schema.Resource{
		Description: `Genesys Cloud conversations messaging integrations open data source. Select an conversations messaging integrations open by name`,
		ReadContext: provider.ReadWithPooledClient(dataSourceConversationsMessagingIntegrationsOpenRead),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Description: `conversations messaging integrations open name`,
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}
