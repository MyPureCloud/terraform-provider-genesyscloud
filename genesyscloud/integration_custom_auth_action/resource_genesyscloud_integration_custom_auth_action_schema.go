package integration_custom_auth_action

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "terraform-provider-genesyscloud/genesyscloud/resource_register"
)

/*
resource_genesyscloud_integration_custom_auth_action_schema.go should hold four types of functions within it:

1.  The registration code that registers the Datasource, Resource and Exporter for the package.
2.  The resource schema definitions for the integration_custom_auth_action resource.
3.  The datasource schema definitions for the integration_custom_auth_action datasource.
4.  The resource exporter configuration for the integration_custom_auth_action exporter.
*/
const resourceName = "genesyscloud_integration_custom_auth_action"

// SetRegistrar registers all of the resources, datasources and exporters in the package
func SetRegistrar(l registrar.Registrar) {
	l.RegisterDataSource(resourceName, DataSourceIntegrationCustomAuthAction())
	l.RegisterResource(resourceName, ResourceIntegrationCustomAuthAction())
	l.RegisterExporter(resourceName, IntegrationCustomAuthActionExporter())
}

// ResourceIntegrationCustomAuthAction registers the genesyscloud_integration_custom_auth_action resource with Terraform
func ResourceIntegrationCustomAuthAction() *schema.Resource {
	actionConfigRequest := &schema.Resource{
		Schema: map[string]*schema.Schema{
			"request_url_template": {
				Description: "URL that may include placeholders for requests to 3rd party service.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"request_type": {
				Description:  "HTTP method to use for request (GET | PUT | POST | PATCH | DELETE).",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"GET", "PUT", "POST", "PATCH", "DELETE"}, false),
			},
			"request_template": {
				Description: "Velocity template to define request body sent to 3rd party service. Any instances of '${' must be properly escaped as '$${'",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},
			"headers": {
				Description: "Map of headers in name, value pairs to include in request.",
				Type:        schema.TypeMap,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
		},
	}

	actionConfigResponse := &schema.Resource{
		Schema: map[string]*schema.Schema{
			"translation_map": {
				Description: "Map 'attribute name' and 'JSON path' pairs used to extract data from REST response.",
				Type:        schema.TypeMap,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"translation_map_defaults": {
				Description: "Map 'attribute name' and 'default value' pairs used as fallback values if JSON path extraction fails for specified key.",
				Type:        schema.TypeMap,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"success_template": {
				Description: "Velocity template to build response to return from Action. Any instances of '${' must be properly escaped as '$${'.",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},
		},
	}

	return &schema.Resource{
		Description: "Genesys Cloud Integration Actions. See this page for detailed information on configuring Actions: https://help.mypurecloud.com/articles/add-configuration-custom-actions-integrations/",

		CreateContext: provider.CreateWithPooledClient(createIntegrationCustomAuthAction),
		ReadContext:   provider.ReadWithPooledClient(readIntegrationCustomAuthAction),
		UpdateContext: provider.UpdateWithPooledClient(updateIntegrationCustomAuthAction),
		DeleteContext: provider.DeleteWithPooledClient(deleteIntegrationCustomAuthAction),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"integration_id": {
				Description: "The ID of the integration this action is associated with. The integration is required to be of type `custom-rest-actions` and its credentials type set as `userDefinedOAuth`.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"name": {
				Description:  "Name of the action to override the default name. Can be up to 256 characters long",
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringLenBetween(1, 256),
			},
			"config_request": {
				Description: "Configuration of outbound request.",
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Elem:        actionConfigRequest,
			},
			"config_response": {
				Description: "Configuration of response processing.",
				Type:        schema.TypeList,
				Optional:    true,
				Computed:    true,
				MaxItems:    1,
				Elem:        actionConfigResponse,
			},
		},
	}
}

// IntegrationCustomAuthActionExporter returns the resourceExporter object used to hold the genesyscloud_integration_custom_auth_action exporter's config
func IntegrationCustomAuthActionExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllModifiedCustomAuthActions),
		RefAttrs: map[string]*resourceExporter.RefAttrSettings{
			"integration_id": {RefType: "genesyscloud_integration"},
		},
	}
}

// DataSourceIntegrationCustomAuthAction registers the genesyscloud_integration_custom_auth_action data source
func DataSourceIntegrationCustomAuthAction() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for Genesys Cloud integration custom auth action. Select the custom auth action by its associated integration's id.",
		ReadContext: provider.ReadWithPooledClient(dataSourceIntegrationCustomAuthActionRead),
		Schema: map[string]*schema.Schema{
			"parent_integration_id": {
				Description: "The id of the integration associated with the custom auth action",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}
