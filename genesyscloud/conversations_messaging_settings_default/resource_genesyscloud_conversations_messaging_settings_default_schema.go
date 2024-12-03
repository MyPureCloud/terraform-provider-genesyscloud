package conversations_messaging_settings_default

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"terraform-provider-genesyscloud/genesyscloud/provider"
	registrar "terraform-provider-genesyscloud/genesyscloud/resource_register"
)

const ResourceType = "genesyscloud_conversations_messaging_settings_default"

// SetRegistrar registers all of the resources, datasources and exporters in the package
func SetRegistrar(regInstance registrar.Registrar) {
	regInstance.RegisterResource(ResourceType, ResourceConversationsMessagingSettingsDefault())
}

// ResourceConversationsMessagingSettingsDefault registers the genesyscloud_conversations_messaging_settings_default resource with Terraform
func ResourceConversationsMessagingSettingsDefault() *schema.Resource {
	return &schema.Resource{
		Description: `Genesys Cloud conversations messaging settings default`,

		CreateContext: provider.CreateWithPooledClient(createConversationsMessagingSettingsDefault),
		ReadContext:   provider.ReadWithPooledClient(readConversationsMessagingSettingsDefault),
		UpdateContext: provider.UpdateWithPooledClient(updateConversationsMessagingSettingsDefault),
		DeleteContext: provider.DeleteWithPooledClient(deleteConversationsMessagingSettingsDefault),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			`setting_id`: {
				Description: `Messaging Setting ID to be used as the default for this Organization.`,
				Required:    true,
				Type:        schema.TypeString,
			},
		},
	}
}
