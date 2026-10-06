package recording_settings

import (
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_register"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

/*
resource_genesyscloud_recording_settings_schema.go holds four functions within it:

1.  The registration code that registers the Datasource, Resource and Exporter for the package.
2.  The resource schema definitions for the recording_settings resource.
3.  The datasource schema definitions for the recording_settings datasource.
4.  The resource exporter configuration for the recording_settings exporter.

This is an organization-level singleton resource. The /api/v2/recording/settings endpoint
only supports GET and PUT, so there is no create or delete operation. Create is implemented
by assigning a fixed ID and applying an update, and delete is a no-op.
*/
const ResourceType = "genesyscloud_recording_settings"

// SetRegistrar registers all of the resources, datasources and exporters in the package
func SetRegistrar(l registrar.Registrar) {
	l.RegisterResource(ResourceType, ResourceRecordingSettings())
	l.RegisterDataSource(ResourceType, DataSourceRecordingSettings())
	l.RegisterExporter(ResourceType, RecordingSettingsExporter())
}

func ResourceRecordingSettings() *schema.Resource {
	return &schema.Resource{
		Description: "Genesys Cloud Organization Recording Settings. This is a singleton resource; only one instance exists per organization.",

		CreateContext: provider.CreateWithPooledClient(createRecordingSettings),
		ReadContext:   provider.ReadWithPooledClient(readRecordingSettings),
		UpdateContext: provider.UpdateWithPooledClient(updateRecordingSettings),
		DeleteContext: provider.DeleteWithPooledClient(deleteRecordingSettings),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"max_simultaneous_streams": {
				Description: "Maximum number of simultaneous screen recording streams.",
				Type:        schema.TypeInt,
				Optional:    true,
				Computed:    true,
			},
			"max_configurable_screen_recording_streams": {
				Description: "Upper limit that max_simultaneous_streams can be configured to. This value is read-only.",
				Type:        schema.TypeInt,
				Computed:    true,
			},
			"regional_recording_storage_enabled": {
				Description: "Store call recordings in the region where they are intended to be recorded, otherwise in the organization's home region.",
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
			},
			"recording_playback_url_ttl": {
				Description:  "The duration in minutes for which the generated URL for recording playback remains valid. The default duration is 60 minutes, with a minimum of 2 minutes and a maximum of 60 minutes.",
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.IntBetween(2, 60),
			},
			"recording_batch_download_url_ttl": {
				Description:  "The duration in minutes for which the generated URL for recording batch download remains valid. The default duration is 60 minutes, with a minimum of 2 minutes and a maximum of 60 minutes.",
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.IntBetween(2, 60),
			},
			"stop_recording_when_only_external_participants": {
				Description: "Whether to stop recording in conference when only external participants remain.",
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
			},
		},
	}
}

func DataSourceRecordingSettings() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for Genesys Cloud Organization Recording Settings. Retrieves the organization's recording settings. This is a singleton, so no arguments are required to identify it.",
		ReadContext: provider.ReadWithPooledClient(dataSourceRecordingSettingsRead),
		Schema: map[string]*schema.Schema{
			"max_simultaneous_streams": {
				Description: "Maximum number of simultaneous screen recording streams.",
				Type:        schema.TypeInt,
				Computed:    true,
			},
			"max_configurable_screen_recording_streams": {
				Description: "Upper limit that max_simultaneous_streams can be configured to.",
				Type:        schema.TypeInt,
				Computed:    true,
			},
			"regional_recording_storage_enabled": {
				Description: "Whether call recordings are stored in the region where they are intended to be recorded, otherwise in the organization's home region.",
				Type:        schema.TypeBool,
				Computed:    true,
			},
			"recording_playback_url_ttl": {
				Description: "The duration in minutes for which the generated URL for recording playback remains valid.",
				Type:        schema.TypeInt,
				Computed:    true,
			},
			"recording_batch_download_url_ttl": {
				Description: "The duration in minutes for which the generated URL for recording batch download remains valid.",
				Type:        schema.TypeInt,
				Computed:    true,
			},
			"stop_recording_when_only_external_participants": {
				Description: "Whether to stop recording in conference when only external participants remain.",
				Type:        schema.TypeBool,
				Computed:    true,
			},
		},
	}
}

func RecordingSettingsExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllRecordingSettings),
		IsSingleton:      true,
		ExportId:         ResourceType,
		RefAttrs:         map[string]*resourceExporter.RefAttrSettings{},
	}
}
