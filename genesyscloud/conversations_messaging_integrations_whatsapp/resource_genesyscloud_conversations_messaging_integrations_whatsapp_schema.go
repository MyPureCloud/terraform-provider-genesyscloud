package conversations_messaging_integrations_whatsapp

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "terraform-provider-genesyscloud/genesyscloud/resource_register"
)

/*
resource_genesycloud_conversations_messaging_integrations_whatsapp_schema.go holds four functions within it:

1.  The registration code that registers the Datasource, Resource and Exporter for the package.
2.  The resource schema definitions for the conversations_messaging_integrations_whatsapp resource.
3.  The datasource schema definitions for the conversations_messaging_integrations_whatsapp datasource.
4.  The resource exporter configuration for the conversations_messaging_integrations_whatsapp exporter.
*/
const resourceName = "genesyscloud_conversations_messaging_integrations_whatsapp"

// SetRegistrar registers all of the resources, datasources and exporters in the package
func SetRegistrar(regInstance registrar.Registrar) {
	regInstance.RegisterResource(resourceName, ResourceConversationsMessagingIntegrationsWhatsapp())
	regInstance.RegisterDataSource(resourceName, DataSourceConversationsMessagingIntegrationsWhatsapp())
	regInstance.RegisterExporter(resourceName, ConversationsMessagingIntegrationsWhatsappExporter())
}

// ResourceConversationsMessagingIntegrationsWhatsapp registers the genesyscloud_conversations_messaging_integrations_whatsapp resource with Terraform
func ResourceConversationsMessagingIntegrationsWhatsapp() *schema.Resource {

	return &schema.Resource{
		Description: `Genesys Cloud conversations messaging integrations whatsapp`,

		CreateContext: provider.CreateWithPooledClient(createConversationsMessagingIntegrationsWhatsapp),
		ReadContext:   provider.ReadWithPooledClient(readConversationsMessagingIntegrationsWhatsapp),
		UpdateContext: provider.UpdateWithPooledClient(updateConversationsMessagingIntegrationsWhatsapp),
		DeleteContext: provider.DeleteWithPooledClient(deleteConversationsMessagingIntegrationsWhatsapp),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			`name`: {
				Description: `The name of the WhatsApp Integration`,
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
			`embedded_signup_access_token`: {
				Description: `The access token returned from the embedded signup flow`,
				Required:    true,
				Type:        schema.TypeString,
			},
		},
	}
}

// ConversationsMessagingIntegrationsWhatsappExporter returns the resourceExporter object used to hold the genesyscloud_conversations_messaging_integrations_whatsapp exporter's config
func ConversationsMessagingIntegrationsWhatsappExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllAuthConversationsMessagingIntegrationsWhatsapps),
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

// DataSourceConversationsMessagingIntegrationsWhatsapp registers the genesyscloud_conversations_messaging_integrations_whatsapp data source
func DataSourceConversationsMessagingIntegrationsWhatsapp() *schema.Resource {
	return &schema.Resource{
		Description: `Genesys Cloud conversations messaging integrations whatsapp data source. Select an conversations messaging integrations whatsapp by name`,
		ReadContext: provider.ReadWithPooledClient(dataSourceConversationsMessagingIntegrationsWhatsappRead),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Description: `conversations messaging integrations whatsapp name`,
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}
