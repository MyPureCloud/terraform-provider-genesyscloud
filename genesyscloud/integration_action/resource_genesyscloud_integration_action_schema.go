package integration_action

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "terraform-provider-genesyscloud/genesyscloud/resource_register"
	"terraform-provider-genesyscloud/genesyscloud/util"
)

/*
resource_genesyscloud_integration_action_schema.go should hold four types of functions within it:

1.  The registration code that registers the Datasource, Resource and Exporter for the package.
2.  The resource schema definitions for the integration_action resource.
3.  The datasource schema definitions for the integration_action datasource.
4.  The resource exporter configuration for the integration_action exporter.
*/
const resourceName = "genesyscloud_integration_action"

// SetRegistrar registers all of the resources, datasources and exporters in the package
func SetRegistrar(l registrar.Registrar) {
	l.RegisterDataSource(resourceName, DataSourceIntegrationAction())
	l.RegisterResource(resourceName, ResourceIntegrationAction())
	l.RegisterExporter(resourceName, IntegrationActionExporter())
}

// ResourceIntegrationAction registers the genesyscloud_integration_action resource with Terraform
func ResourceIntegrationAction() *schema.Resource {
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

		CreateContext: provider.CreateWithPooledClient(createIntegrationAction),
		ReadContext:   provider.ReadWithPooledClient(readIntegrationAction),
		UpdateContext: provider.UpdateWithPooledClient(updateIntegrationAction),
		DeleteContext: provider.DeleteWithPooledClient(deleteIntegrationAction),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"name": {
				Description:  "Name of the action. Can be up to 256 characters long",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 256),
			},
			"category": {
				Description:  "Category of action. Can be up to 256 characters long.",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 256),
			},
			"integration_id": {
				Description: "The ID of the integration this action is associated with. Changing the integration_id attribute will cause the existing integration_action to be dropped and recreated with a new ID.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"secure": {
				Description: "Indication of whether or not the action is designed to accept sensitive data. Changing the secure attribute will cause the existing integration_action to be dropped and recreated with a new ID.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				ForceNew:    true,
			},
			"config_timeout_seconds": {
				Description:  "Optional 1-60 second timeout enforced on the execution or test of this action. This setting is invalid for Custom Authentication Actions.",
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(1, 60),
			},
			"contract_input": {
				Description:      "JSON Schema that defines the body of the request that the client (edge/architect/postman) is sending to the service, on the /execute path. Changing the contract_input attribute will cause the existing integration_action to be dropped and recreated with a new ID.",
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				DiffSuppressFunc: util.SuppressEquivalentJsonDiffs,
			},
			"contract_output": {
				Description:      "JSON schema that defines the transformed, successful result that will be sent back to the caller. Changing the contract_output attribute will cause the existing integration_action to be dropped and recreated with a new ID.",
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				DiffSuppressFunc: util.SuppressEquivalentJsonDiffs,
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

// IntegrationActionExporter returns the resourceExporter object used to hold the genesyscloud_integration_action exporter's config
func IntegrationActionExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllIntegrationActions),
		RefAttrs: map[string]*resourceExporter.RefAttrSettings{
			"integration_id": {RefType: "genesyscloud_integration"},
		},
		JsonEncodeAttributes: []string{"contract_input", "contract_output"},
	}
}

// DataSourceIntegrationAction registers the genesyscloud_integration_action data source
func DataSourceIntegrationAction() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for Genesys Cloud integration action. Select an integration action by name",
		ReadContext: provider.ReadWithPooledClient(dataSourceIntegrationActionRead),
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "The name of the integration action",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}
