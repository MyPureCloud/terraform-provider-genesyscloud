package integration

import (
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_register"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

/*
resource_genesyscloud_integration_schema.go should hold four types of functions within it:

1.  The registration code that registers the Datasource, Resource and Exporter for the package.
2.  The resource schema definitions for the integration resource.
3.  The datasource schema definitions for the integration datasource.
4.  The resource exporter configuration for the integration exporter.
*/
const ResourceType = "genesyscloud_integration"
const WebhookResourceType = "genesyscloud_integration_webhook"

// SetRegistrar registers all of the resources, datasources and exporters in the package
func SetRegistrar(l registrar.Registrar) {
	l.RegisterDataSource(ResourceType, DataSourceIntegration())
	l.RegisterDataSource(WebhookResourceType, DataSourceIntegrationWebhook())
	l.RegisterResource(ResourceType, ResourceIntegration())
	l.RegisterExporter(ResourceType, IntegrationExporter())
}

// ResourceIntegration registers the genesyscloud_integration resource with Terraform
func ResourceIntegration() *schema.Resource {
	integrationConfigResource := &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Integration name.",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},
			"notes": {
				Description: "Integration notes.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"properties": {
				Description:      "Integration config properties (JSON string).",
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				DiffSuppressFunc: util.SuppressEquivalentJsonDiffs,
			},
			"advanced": {
				Description:      "Integration advanced config (JSON string).",
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				DiffSuppressFunc: util.SuppressEquivalentJsonDiffs,
			},
			"credentials": {
				Description: "Credentials required for the integration. The required keys are indicated in the credentials property of the Integration Type.",
				Type:        schema.TypeMap,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
		},
	}

	return &schema.Resource{
		Description: "Genesys Cloud Integration",

		CreateContext: provider.CreateWithPooledClient(createIntegration),
		ReadContext:   provider.ReadWithPooledClient(readIntegration),
		UpdateContext: provider.UpdateWithPooledClient(updateIntegration),
		DeleteContext: provider.DeleteWithPooledClient(deleteIntegration),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"intended_state": {
				Description:  "Integration state (ENABLED | DISABLED).",
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "DISABLED",
				ValidateFunc: validation.StringInSlice([]string{"ENABLED", "DISABLED"}, false),
			},
			"integration_type": {
				Description: "Integration type.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"config": {
				Description: "Integration config. Each integration type has different schema, use [GET /api/v2/integrations/types/{typeId}/configschemas/{configType}](https://developer.mypurecloud.com/api/rest/v2/integrations/#get-api-v2-integrations-types--typeId--configschemas--configType-) to check schema, then use the correct attribute names for properties.",
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Computed:    true,
				Elem:        integrationConfigResource,
			},
		},
	}
}

// IntegrationExporter returns the resourceExporter object used to hold the genesyscloud_integration exporter's config
func IntegrationExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllIntegrations),
		RefAttrs: map[string]*resourceExporter.RefAttrSettings{
			"config.credentials.*": {RefType: "genesyscloud_integration_credential"},
		},
		JsonEncodeAttributes: []string{"config.properties", "config.advanced"},
		EncodedRefAttrs: map[*resourceExporter.JsonEncodeRefAttr]*resourceExporter.RefAttrSettings{
			{Attr: "config.properties", NestedAttr: "groups"}:                {RefType: "genesyscloud_group"},
			{Attr: "config.properties", NestedAttr: "createTimeOffRequests"}: {RefType: "genesyscloud_flow"},
			{Attr: "config.properties", NestedAttr: "timeOffBalances"}:       {RefType: "genesyscloud_flow"},
			{Attr: "config.properties", NestedAttr: "timeOffTypes"}:          {RefType: "genesyscloud_flow"},
			{Attr: "config.properties", NestedAttr: "updateTimeOffRequests"}: {RefType: "genesyscloud_flow"},
			{Attr: "config.properties", NestedAttr: "userAccountIds"}:        {RefType: "genesyscloud_flow"},
		},
		DataSourceResolver: map[*resourceExporter.DataAttr]*resourceExporter.ResourceAttr{
			{Attr: "name"}: {Attr: "^config\\.\\d+\\.name$"},
		},
	}
}

// DataSourceIntegration registers the genesyscloud_integration data source
func DataSourceIntegration() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for Genesys Cloud integration. Select an integration by name",
		ReadContext: provider.ReadWithPooledClient(dataSourceIntegrationRead),
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "The name of the integration",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}

// DataSourceIntegrationWebhook registers the genesyscloud_integration_webhook data source
func DataSourceIntegrationWebhook() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for Genesys Cloud webhook integration. Select a webhook integration by name",
		ReadContext: provider.ReadWithPooledClient(dataSourceIntegrationWebhookRead),
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "The name of the webhook integration",
				Type:        schema.TypeString,
				Required:    true,
			},
			"web_hook_id": {
				Description: "The webhook ID from the integration attributes",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"invocation_url": {
				Description: "The invocation URL from the integration attributes",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}
