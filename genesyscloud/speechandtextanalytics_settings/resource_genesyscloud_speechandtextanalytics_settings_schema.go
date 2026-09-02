package speechandtextanalytics_settings

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_register"
)

/*
resource_genesyscloud_speechandtextanalytics_settings_schema.go holds four functions within it:

1.  The registration code that registers the Datasource, Resource and Exporter for the package.
2.  The resource schema definitions for the speechandtextanalytics_settings resource.
3.  The datasource schema definitions for the speechandtextanalytics_settings datasource.
4.  The resource exporter configuration for the speechandtextanalytics_settings exporter.
*/
const ResourceType = "genesyscloud_speechandtextanalytics_settings"

// SetRegistrar registers all of the resources, datasources and exporters in the package
func SetRegistrar(l registrar.Registrar) {
	l.RegisterResource(ResourceType, ResourceSpeechAndTextAnalyticsSettings())
	l.RegisterDataSource(ResourceType, DataSourceSpeechAndTextAnalyticsSettings())
	l.RegisterExporter(ResourceType, SpeechAndTextAnalyticsSettingsExporter())
}

// ResourceSpeechAndTextAnalyticsSettings registers the genesyscloud_speechandtextanalytics_settings resource with Terraform
func ResourceSpeechAndTextAnalyticsSettings() *schema.Resource {
	return &schema.Resource{
		Description: "Genesys Cloud organization Speech & Text Analytics settings. This is a singleton resource; only one instance exists per organization.",

		CreateContext: provider.CreateWithPooledClient(createSpeechAndTextAnalyticsSettings),
		ReadContext:   provider.ReadWithPooledClient(readSpeechAndTextAnalyticsSettings),
		UpdateContext: provider.UpdateWithPooledClient(updateSpeechAndTextAnalyticsSettings),
		DeleteContext: provider.DeleteWithPooledClient(deleteSpeechAndTextAnalyticsSettings),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"default_program_id": {
				Description: "The ID of the default program used for topic detection.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"expected_dialects": {
				Description: "The list of expected dialects, e.g. en-US.",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"text_analytics_enabled": {
				Description: "Indicates whether text analytics is enabled.",
				Type:        schema.TypeBool,
				Optional:    true,
			},
			"agent_empathy_enabled": {
				Description: "Indicates whether the Agent Empathy setting is enabled.",
				Type:        schema.TypeBool,
				Optional:    true,
			},
		},
	}
}

// DataSourceSpeechAndTextAnalyticsSettings registers the genesyscloud_speechandtextanalytics_settings data source with Terraform
func DataSourceSpeechAndTextAnalyticsSettings() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for Genesys Cloud organization Speech & Text Analytics settings. This is a singleton, so no arguments are required.",
		ReadContext: provider.ReadWithPooledClient(dataSourceSpeechAndTextAnalyticsSettingsRead),
		Schema: map[string]*schema.Schema{
			"default_program_id": {
				Description: "The ID of the default program used for topic detection.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"expected_dialects": {
				Description: "The list of expected dialects, e.g. en-US.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"text_analytics_enabled": {
				Description: "Indicates whether text analytics is enabled.",
				Type:        schema.TypeBool,
				Computed:    true,
			},
			"agent_empathy_enabled": {
				Description: "Indicates whether the Agent Empathy setting is enabled.",
				Type:        schema.TypeBool,
				Computed:    true,
			},
		},
	}
}

func SpeechAndTextAnalyticsSettingsExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllSpeechAndTextAnalyticsSettings),
		IsSingleton:      true,
		ExportId:         ResourceType,
		RefAttrs:         map[string]*resourceExporter.RefAttrSettings{},
	}
}
