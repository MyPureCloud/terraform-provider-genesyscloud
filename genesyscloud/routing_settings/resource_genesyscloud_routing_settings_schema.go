package routing_settings

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "terraform-provider-genesyscloud/genesyscloud/resource_register"
)

var resourceName = "genesyscloud_routing_settings"

func SetRegistrar(regInstance registrar.Registrar) {
	regInstance.RegisterResource(resourceName, ResourceRoutingSettings())
	regInstance.RegisterExporter(resourceName, RoutingSettingsExporter())
}

func ResourceRoutingSettings() *schema.Resource {
	return &schema.Resource{
		Description: "An organization's routing settings",

		CreateContext: provider.CreateWithPooledClient(createRoutingSettings),
		ReadContext:   provider.ReadWithPooledClient(readRoutingSettings),
		UpdateContext: provider.UpdateWithPooledClient(updateRoutingSettings),
		DeleteContext: provider.DeleteWithPooledClient(deleteRoutingSettings),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"reset_agent_on_presence_change": {
				Description: "Reset agent score when agent presence changes from off-queue to on-queue",
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
			},
			"contactcenter": {
				Description: "Contact center settings",
				Type:        schema.TypeList,
				Optional:    true,
				Computed:    true,
				MaxItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"remove_skills_from_blind_transfer": {
							Description: "Strip skills from transfer",
							Type:        schema.TypeBool,
							Optional:    true,
						},
					},
				},
			},
			"transcription": {
				Description: "Transcription settings",
				Type:        schema.TypeList,
				Optional:    true,
				Computed:    true,
				MaxItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"transcription": {
							Description: "Setting to enable/disable transcription capability.Valid values: Disabled, EnabledGlobally, EnabledQueueFlow",
							Type:        schema.TypeString,
							Optional:    true,
						},
						"transcription_confidence_threshold": {
							Description: "Configure confidence threshold. The possible values are from 1 to 100",
							Type:        schema.TypeInt,
							Optional:    true,
						},
						"low_latency_transcription_enabled": {
							Description: "Boolean flag indicating whether low latency transcription via Notification API is enabled",
							Type:        schema.TypeBool,
							Optional:    true,
						},
						"content_search_enabled": {
							Description: "Setting to enable/disable content search",
							Type:        schema.TypeBool,
							Optional:    true,
						},
						`pci_dss_redaction_enabled`: {
							Description: `Setting to enable/disable PCI DSS Redaction`,
							Optional:    true,
							Type:        schema.TypeBool,
						},
						`pii_redaction_enabled`: {
							Description: `Setting to enable/disable PII Redaction`,
							Optional:    true,
							Type:        schema.TypeBool,
						},
					},
				},
			},
		},
	}
}

func RoutingSettingsExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllRoutingSettings),
	}
}
