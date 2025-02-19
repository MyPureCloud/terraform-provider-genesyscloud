package conversations_messaging_integrations_instagram

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "terraform-provider-genesyscloud/genesyscloud/resource_register"
)

/*
resource_genesycloud_conversations_messaging_integrations_instagram_schema.go holds four functions within it:

1.  The registration code that registers the Datasource, Resource and Exporter for the package.
2.  The resource schema definitions for the conversations_messaging_integrations_instagram resource.
3.  The datasource schema definitions for the conversations_messaging_integrations_instagram datasource.
4.  The resource exporter configuration for the conversations_messaging_integrations_instagram exporter.
*/
const ResourceType = "genesyscloud_conversations_messaging_integrations_instagram"

// SetRegistrar registers all of the resources, datasources and exporters in the package
func SetRegistrar(regInstance registrar.Registrar) {
	regInstance.RegisterResource(ResourceType, ResourceConversationsMessagingIntegrationsInstagram())
	regInstance.RegisterDataSource(ResourceType, DataSourceConversationsMessagingIntegrationsInstagram())
	regInstance.RegisterExporter(ResourceType, ConversationsMessagingIntegrationsInstagramExporter())
}

// ResourceConversationsMessagingIntegrationsInstagram registers the genesyscloud_conversations_messaging_integrations_instagram resource with Terraform
func ResourceConversationsMessagingIntegrationsInstagram() *schema.Resource {
	return &schema.Resource{
		Description: `Genesys Cloud conversations messaging integrations instagram`,

		CreateContext: provider.CreateWithPooledClient(createConversationsMessagingIntegrationsInstagram),
		ReadContext:   provider.ReadWithPooledClient(readConversationsMessagingIntegrationsInstagram),
		UpdateContext: provider.UpdateWithPooledClient(updateConversationsMessagingIntegrationsInstagram),
		DeleteContext: provider.DeleteWithPooledClient(deleteConversationsMessagingIntegrationsInstagram),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			`name`: {
				Description: `The name of the Instagram Integration`,
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
			`page_access_token`: {
				Description: `The long-lived Page Access Token of Instagram page. See https://developers.facebook.com/docs/facebook-login/access-tokens. When a pageAccessToken is provided, pageId and userAccessToken are not required.`,
				Optional:    true,
				Type:        schema.TypeString,
			},
			`user_access_token`: {
				Description: `The short-lived User Access Token of Instagram user logged into Facebook app. See https://developers.facebook.com/docs/facebook-login/access-tokens. When userAccessToken is provided, pageId is mandatory. When userAccessToken/pageId combination is provided, pageAccessToken is not required.`,
				Optional:    true,
				Type:        schema.TypeString,
			},
			`page_id`: {
				Description: `The page ID of Instagram page. The pageId is required when userAccessToken is provided.`,
				Optional:    true,
				Type:        schema.TypeString,
			},
			`app_id`: {
				Description: `The app ID of Facebook app. The appId is required when a customer wants to use their own approved Facebook app.`,
				Optional:    true,
				Type:        schema.TypeString,
			},
			`app_secret`: {
				Description: `The app Secret of Facebook app. The appSecret is required when appId is provided.`,
				Optional:    true,
				Type:        schema.TypeString,
			},
		},
	}
}

// ConversationsMessagingIntegrationsInstagramExporter returns the resourceExporter object used to hold the genesyscloud_conversations_messaging_integrations_instagram exporter's config
func ConversationsMessagingIntegrationsInstagramExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllAuthConversationsMessagingIntegrationsInstagrams),
		RefAttrs: map[string]*resourceExporter.RefAttrSettings{
			"messaging_setting_id": {
				RefType: "genesyscloud_conversations_messaging_settings",
			},
			"supported_content_id": {
				RefType: "genesyscloud_conversations_messaging_supportedcontent",
			},
		},
	}
}

// DataSourceConversationsMessagingIntegrationsInstagram registers the genesyscloud_conversations_messaging_integrations_instagram data source
func DataSourceConversationsMessagingIntegrationsInstagram() *schema.Resource {
	return &schema.Resource{
		Description: `Genesys Cloud conversations messaging integrations instagram data source. Select an conversations messaging integrations instagram by name`,
		ReadContext: provider.ReadWithPooledClient(dataSourceConversationsMessagingIntegrationsInstagramRead),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Description: `conversations messaging integrations instagram name`,
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}
