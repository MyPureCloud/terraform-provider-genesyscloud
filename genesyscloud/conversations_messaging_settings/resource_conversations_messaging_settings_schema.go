package conversations_messaging_settings

import (
	"terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "terraform-provider-genesyscloud/genesyscloud/resource_register"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

const resourceName = "genesyscloud_conversations_messaging_settings"

var (
	eventSettingResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"typing": {
				Description: "Settings regarding typing events",
				Optional:    true,
				Type:        schema.TypeList,
				MaxItems:    1,
				Elem:        typingSettingResource,
			},
		},
	}
	typingSettingResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"on": {
				Description: "Should typing indication Events be sent",
				Optional:    true,
				Type:        schema.TypeList,
				MaxItems:    1,
				Elem:        settingDirectionResource,
			},
		},
	}
	settingDirectionResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"inbound": {
				Description:  "Status for the Inbound Direction. Valid values: Enabled, Disabled.",
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"Enabled", "Disabled"}, false),
			},
			"outbound": {
				Description:  "Status for the outbound Direction. Valid values: Enabled, Disabled.",
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"Enabled", "Disabled"}, false),
			},
		},
	}

	contentSettingResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"story": {
				Description: "Settings relating to facebook and instagram stories feature",
				Optional:    true,
				Type:        schema.TypeList,
				MaxItems:    1,
				Elem:        storySettingResource,
			},
		},
	}
	storySettingResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"mention": {
				Description: "Setting relating to Story Mentions",
				Optional:    true,
				Type:        schema.TypeList,
				MaxItems:    1,
				Elem:        inboundOnlySettingResource,
			},
			"reply": {
				Description: "Setting relating to Story Replies",
				Optional:    true,
				Type:        schema.TypeList,
				MaxItems:    1,
				Elem:        inboundOnlySettingResource,
			},
		},
	}
	inboundOnlySettingResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"inbound": {
				Description:  "Valid values: Enabled, Disabled.",
				Optional:     true,
				Type:         schema.TypeString,
				ValidateFunc: validation.StringInSlice([]string{"Enabled", "Disabled"}, false),
			},
		},
	}
)

// SetRegistrar registers all of the resources, datasources and exporters in the package
func SetRegistrar(regInstance registrar.Registrar) {
	regInstance.RegisterResource(resourceName, ResourceConversationsMessagingSettings())
	regInstance.RegisterDataSource(resourceName, DataSourceConversationsMessagingSettings())
	regInstance.RegisterExporter(resourceName, ConversationsMessagingSettingsExporter())
}

func ResourceConversationsMessagingSettings() *schema.Resource {
	return &schema.Resource{
		Description: "Genesys Cloud conversations messaging settings",

		CreateContext: provider.CreateWithPooledClient(createConversationsMessagingSettings),
		ReadContext:   provider.ReadWithPooledClient(readConversationsMessagingSettings),
		UpdateContext: provider.UpdateWithPooledClient(updateConversationsMessagingSettings),
		DeleteContext: provider.DeleteWithPooledClient(deleteConversationsMessagingSettings),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "The messaging Setting profile name",
				Required:    true,
				Type:        schema.TypeString,
			},
			"content": {
				Description:  "Settings relating to message contents",
				Optional:     true,
				Type:         schema.TypeList,
				MaxItems:     1,
				Elem:         contentSettingResource,
				AtLeastOneOf: []string{"content", "event"},
			},
			"event": {
				Description:  "Settings relating to events which may occur",
				Optional:     true,
				Type:         schema.TypeList,
				MaxItems:     1,
				Elem:         eventSettingResource,
				AtLeastOneOf: []string{"content", "event"},
			},
		},
	}
}

// DataSourceConversationsMessagingSettings registers the genesyscloud_conversations_messaging_settings data source
func DataSourceConversationsMessagingSettings() *schema.Resource {
	return &schema.Resource{
		Description: "Genesys Cloud conversations messaging settings data source. Select an conversations messaging settings by name",
		ReadContext: provider.ReadWithPooledClient(dataSourceConversationsMessagingSettingsRead),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "conversations messaging settings name",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}

// ConversationsMessagingSettingsExporter returns the resourceExporter object used to hold the genesyscloud_conversations_messaging_settings exporter's config
func ConversationsMessagingSettingsExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllAuthConversationsMessagingSettings),
	}
}
