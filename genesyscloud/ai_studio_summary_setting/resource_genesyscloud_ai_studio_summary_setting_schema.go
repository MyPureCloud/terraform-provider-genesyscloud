package ai_studio_summary_setting

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_register"
)

/*
resource_genesycloud_ai_studio_summary_setting_schema.go holds four functions within it:

1.  The registration code that registers the Datasource, Resource and Exporter for the package.
2.  The resource schema definitions for the ai_studio_summary_setting resource.
3.  The datasource schema definitions for the ai_studio_summary_setting datasource.
4.  The resource exporter configuration for the ai_studio_summary_setting exporter.
*/
const ResourceType = "genesyscloud_ai_studio_summary_setting"

// SetRegistrar registers all of the resources, datasources and exporters in the package
func SetRegistrar(regInstance registrar.Registrar) {
	regInstance.RegisterResource(ResourceType, ResourceAiStudioSummarySetting())
	regInstance.RegisterDataSource(ResourceType, DataSourceAiStudioSummarySetting())
	regInstance.RegisterExporter(ResourceType, AiStudioSummarySettingExporter())
}

// ResourceAiStudioSummarySetting registers the genesyscloud_ai_studio_summary_setting resource with Terraform
func ResourceAiStudioSummarySetting() *schema.Resource {
	summarySettingPIIResource := &schema.Resource{
		Schema: map[string]*schema.Schema{
			`all`: {
				Description: `Toggle PII visibility in summary.`,
				Required:    true,
				Type:        schema.TypeBool,
			},
		},
	}

	summarySettingParticipantLabelsResource := &schema.Resource{
		Schema: map[string]*schema.Schema{
			`internal`: {
				Description: `Specify how to refer the internal participant of the interaction.`,
				Optional:    true,
				Type:        schema.TypeString,
			},
			`external`: {
				Description: `Specify how to refer the external participant of the interaction.`,
				Optional:    true,
				Type:        schema.TypeString,
			},
		},
	}

	summarySettingCustomEntityResource := &schema.Resource{
		Schema: map[string]*schema.Schema{
			`label`: {
				Description: `Label how the entity should be called.`,
				Optional:    true,
				Type:        schema.TypeString,
			},
			`description`: {
				Description: `Describe the information the entity captures.`,
				Optional:    true,
				Type:        schema.TypeString,
			},
		},
	}

	return &schema.Resource{
		Description: `Genesys Cloud ai studio summary setting`,

		CreateContext: provider.CreateWithPooledClient(createAiStudioSummarySetting),
		ReadContext:   provider.ReadWithPooledClient(readAiStudioSummarySetting),
		UpdateContext: provider.UpdateWithPooledClient(updateAiStudioSummarySetting),
		DeleteContext: provider.DeleteWithPooledClient(deleteAiStudioSummarySetting),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			`name`: {
				Description: `Name of the summary setting.`,
				Required:    true,
				Type:        schema.TypeString,
			},
			`language`: {
				Description: `Language of the generated summary, e.g. en-US, it-IT.`,
				Required:    true,
				Type:        schema.TypeString,
			},
			`summary_type`: {
				Description:  `Level of detail of the generated summary.`,
				Optional:     true,
				Type:         schema.TypeString,
				ValidateFunc: validation.StringInSlice([]string{"Concise", "Detailed"}, false),
			},
			`format`: {
				Description:  `Format of the generated summary.`,
				Optional:     true,
				Type:         schema.TypeString,
				ValidateFunc: validation.StringInSlice([]string{"TextBlock", "BulletPoints", "GroupedTextBlocks", "GroupedBulletPoints"}, false),
			},
			`mask_p_i_i`: {
				Description: `Displaying PII in the generated summary.`,
				Optional:    true,
				Type:        schema.TypeList,
				MaxItems:    1,
				Elem:        summarySettingPIIResource,
			},
			`participant_labels`: {
				Description: `How to refer to interaction participants in the generated summary.`,
				Optional:    true,
				Type:        schema.TypeList,
				MaxItems:    1,
				Elem:        summarySettingParticipantLabelsResource,
			},
			`predefined_insights`: {
				Description: `Set which insights to include in the generated summary by default.`,
				Optional:    true,
				Type:        schema.TypeList,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			`custom_entities`: {
				Description: `Custom entity definition.`,
				Optional:    true,
				Type:        schema.TypeList,
				Elem:        summarySettingCustomEntityResource,
			},
			`setting_type`: {
				Description:  `Type of the summary setting.`,
				Optional:     true,
				Type:         schema.TypeString,
				ValidateFunc: validation.StringInSlice([]string{"Basic", "Prompt"}, false),
			},
			`prompt`: {
				Description: `Custom prompt of summary setting.`,
				Optional:    true,
				Type:        schema.TypeString,
			},
		},
	}
}

// AiStudioSummarySettingExporter returns the resourceExporter object used to hold the genesyscloud_ai_studio_summary_setting exporter's config
func AiStudioSummarySettingExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllAuthAiStudioSummarySettings),
		RefAttrs:         map[string]*resourceExporter.RefAttrSettings{
			// TODO: Add any reference attributes here
		},
	}
}

// DataSourceAiStudioSummarySetting registers the genesyscloud_ai_studio_summary_setting data source
func DataSourceAiStudioSummarySetting() *schema.Resource {
	return &schema.Resource{
		Description: `Genesys Cloud ai studio summary setting data source. Select an ai studio summary setting by name`,
		ReadContext: provider.ReadWithPooledClient(dataSourceAiStudioSummarySettingRead),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Description: `ai studio summary setting name`,
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}
