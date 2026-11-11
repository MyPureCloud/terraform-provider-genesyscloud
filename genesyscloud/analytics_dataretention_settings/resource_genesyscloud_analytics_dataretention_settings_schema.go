package analytics_dataretention_settings

// @team: — Recording
// @chat: — Genesys Cloud Recording
// @description: — analytics maximum interaction data retention time organization setting
// @pm: — Daniel Ho
// @jira: — REC

import (
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_register"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

/*
resource_genesyscloud_analytics_dataretention_settings_schema.go holds four functions within it:

1.  The registration code that registers the Datasource, Resource and Exporter for the package.
2.  The resource schema definitions for the analytics_dataretention_settings resource.
3.  The datasource schema definitions for the analytics_dataretention_settings datasource.
4.  The resource exporter configuration for the analytics_dataretention_settings exporter.

This is an organization-level singleton resource. The /api/v2/analytics/dataretention/settings endpoint
only supports GET and PUT, so there is no create or delete operation. Create is implemented
by assigning a fixed ID and applying an update, and delete is a no-op.
*/
const ResourceType = "genesyscloud_analytics_dataretention_settings"

// SetRegistrar registers all of the resources, datasources and exporters in the package
func SetRegistrar(l registrar.Registrar) {
	l.RegisterResource(ResourceType, ResourceAnalyticsDataretentionSettings())
	l.RegisterDataSource(ResourceType, DataSourceAnalyticsDataretentionSettings())
	l.RegisterExporter(ResourceType, AnalyticsDataretentionSettingsExporter())
}

func ResourceAnalyticsDataretentionSettings() *schema.Resource {
	return &schema.Resource{
		Description: "Genesys Cloud Analytics Maximum Interaction Data Retention Time. This is a singleton resource; only one instance exists per organization. " +
			"Once the setting has been configured it cannot be cleared, only changed to another value. Destroying this resource only removes it from the Terraform state and does not change the setting in Genesys Cloud.",

		CreateContext: provider.CreateWithPooledClient(createAnalyticsDataretentionSettings),
		ReadContext:   provider.ReadWithPooledClient(readAnalyticsDataretentionSettings),
		UpdateContext: provider.UpdateWithPooledClient(updateAnalyticsDataretentionSettings),
		DeleteContext: provider.DeleteWithPooledClient(deleteAnalyticsDataretentionSettings),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"retention_days": {
				Description: "The maximum number of days that interaction data and recordings are retained for the organization. " +
					"All interactions and recordings, including protected recordings, are deleted after this many days from their initial recording. " +
					"The minimum value is determined by the organization's limits; the Genesys Cloud UI requires at least 90 days. " +
					"The update is rejected if existing recording media retention policies, recordings or the snippet recording settings exceed the new value.",
				Type:         schema.TypeInt,
				Required:     true,
				ValidateFunc: validation.IntAtLeast(1),
			},
		},
	}
}

func DataSourceAnalyticsDataretentionSettings() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for the Genesys Cloud Analytics Maximum Interaction Data Retention Time. Retrieves the organization's analytics data retention setting. This is a singleton, so no arguments are required to identify it.",
		ReadContext: provider.ReadWithPooledClient(dataSourceAnalyticsDataretentionSettingsRead),
		Schema: map[string]*schema.Schema{
			"retention_days": {
				Description: "The maximum number of days that interaction data and recordings are retained for the organization.",
				Type:        schema.TypeInt,
				Computed:    true,
			},
		},
	}
}

func AnalyticsDataretentionSettingsExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllAnalyticsDataretentionSettings),
		IsSingleton:      true,
		ExportId:         ResourceType,
		RefAttrs:         map[string]*resourceExporter.RefAttrSettings{},
	}
}
