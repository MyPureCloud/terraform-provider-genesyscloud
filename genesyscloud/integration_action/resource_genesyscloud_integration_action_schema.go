package integration_action

// @team: Integration Services Indy
// @chat: #genesys-cloud-integrations
// @pm: Richard Schott
// @jira: INTINDY
// @description: Manages integrations with third-party services and systems. Provides the foundation for connecting Genesys Cloud to external APIs, enabling data exchange and workflow automation across platforms.

import (
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_register"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/validators"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

/*
resource_genesyscloud_integration_action_schema.go should hold four types of functions within it:

1.  The registration code that registers the Datasource, Resource and Exporter for the package.
2.  The resource schema definitions for the integration_action resource.
3.  The datasource schema definitions for the integration_action datasource.
4.  The resource exporter configuration for the integration_action exporter.
*/
const (
	ResourceType = "genesyscloud_integration_action"
	// S3Enabled indicates function zip paths may use local or S3-backed file paths for hashing.
	S3Enabled = true
)

// SetRegistrar registers all of the resources, datasources and exporters in the package
func SetRegistrar(l registrar.Registrar) {
	l.RegisterDataSource(ResourceType, DataSourceIntegrationAction())
	l.RegisterResource(ResourceType, ResourceIntegrationAction())
	l.RegisterExporter(ResourceType, IntegrationActionExporter())
}

// ResourceIntegrationAction registers the genesyscloud_integration_action resource with Terraform
func ResourceIntegrationAction() *schema.Resource {
	actionConfigRequest := &schema.Resource{
		Schema: map[string]*schema.Schema{
			"request_url_template": {
				Description: "URL that may include placeholders for requests to 3rd party service.",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
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

	functionConfig := &schema.Resource{
		Schema: map[string]*schema.Schema{
			"description": {
				Description: "Description of the function.",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},
			"handler": {
				Description: "The handler function name.",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},
			"runtime": {
				Description: "The runtime environment for the function.",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},
			"timeout_seconds": {
				Description: "Timeout in seconds for the function execution.",
				Type:        schema.TypeInt,
				Optional:    true,
				Computed:    true,
			},
			"zip_id": {
				Description: "The ID of the uploaded zip file containing the function code.",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},
			"file_path": {
				Description: "Local path to the zip file containing the function data action code. " +
					"Genesys Cloud does not allow downloading function zip files, so exports cannot retrieve the binary. " +
					"Export sets this attribute to a Terraform variable; assign the variable to your zip path before apply. " +
					"See https://help.genesys.cloud/articles/limitations-of-the-genesys-cloud-function-data-actions-integration/",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validators.ValidatePath,
			},
			"file_content_hash": {
				Description: "Hash value of the function zip file content. Used to detect changes. " +
					"Typically set with filesha256(file_path). Computed by the provider when omitted.",
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
		},
	}

	return &schema.Resource{
		Description: `Genesys Cloud Integration Actions. See this page for detailed information on configuring Actions: https://help.mypurecloud.com/articles/add-configuration-custom-actions-integrations/

Function data action zip files cannot be exported. Genesys Cloud does not allow downloading uploaded function code. See https://help.genesys.cloud/articles/limitations-of-the-genesys-cloud-function-data-actions-integration/. Export emits a Terraform variable for ` + "`function_config.file_path`" + `; set it to a local or S3 zip path before plan or apply.`,

		CreateContext: provider.CreateWithPooledClient(createIntegrationAction),
		ReadContext:   provider.ReadWithPooledClient(readIntegrationAction),
		UpdateContext: provider.UpdateWithPooledClient(updateIntegrationAction),
		DeleteContext: provider.DeleteWithPooledClient(deleteIntegrationAction),
		CustomizeDiff: customdiff.All(
			customdiff.ComputedIf("function_config.0.file_content_hash", validators.ValidateFileContentHashChanged("function_config.0.file_path", "function_config.0.file_content_hash", S3Enabled)),
		),
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
				Description: "Category of action. Can be up to 256 characters long. " +
					"Function data actions are detected when the associated integration type is 'function-data-actions', " +
					"or when the category contains 'function data action' (case-insensitive; underscores and hyphens treated as spaces). " +
					"Function data actions require function_config to be set.",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 256),
			},
			"integration_id": {
				Description: "The ID of the integration this action is associated with. " +
					"When the integration type is 'function-data-actions', this action is created as a draft, the zip is uploaded, then published. " +
					"Changing the integration_id attribute will cause the existing integration_action to be dropped and recreated with a new ID.",
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"secure": {
				Description: "Indication of whether or not the action is designed to accept sensitive data. Changing the secure attribute will cause the existing integration_action to be dropped and recreated with a new ID.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				ForceNew:    true,
			},
			"action_type": {
				Description: "The type of the integration action. Computed based on the action ID prefix. " +
					"Values: `static` (built-in actions shipped by Genesys Cloud) or `custom` (user-created actions).",
				Type:     schema.TypeString,
				Computed: true,
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
			"function_config": {
				Description: "Configuration of the function settings.",
				Type:        schema.TypeList,
				Optional:    true,
				Computed:    true,
				MaxItems:    1,
				Elem:        functionConfig,
			},
		},
	}
}

// IntegrationActionExporter returns the resourceExporter object used to hold the genesyscloud_integration_action exporter's config
func IntegrationActionExporter() *resourceExporter.ResourceExporter {
	functionConfigSchema := ResourceIntegrationAction().Schema["function_config"].Elem.(*schema.Resource).Schema
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllIntegrationActions),
		RefAttrs: map[string]*resourceExporter.RefAttrSettings{
			"integration_id": {RefType: "genesyscloud_integration"},
		},
		JsonEncodeAttributes: []string{"contract_input", "contract_output"},
		AllowZeroValuesInMap: []string{"config_response.translation_map_defaults"},
		// Function zip binaries cannot be downloaded from Genesys Cloud. Treat file_path like
		// the legacy architect flow exporter treats filepath: emit a Terraform variable so
		// apply in another org does not fail looking for a missing local zip.
		UnResolvableAttributes: map[string]*schema.Schema{
			"function_config.file_path": functionConfigSchema["file_path"],
		},
		CustomAttributeResolver: map[string]*resourceExporter.RefAttrCustomResolver{
			"function_config.file_content_hash": {
				ResolverFunc: stripFunctionConfigFileContentHash,
			},
		},
		ThirdPartyRefAttrs: []string{
			"function_config.file_path",
			"function_config.file_content_hash",
		},
		// Static (built-in) data actions are owned by Genesys Cloud and cannot be created,
		// updated, or deleted via the public API. Export them as data sources so that other
		// resources (e.g. Architect flows) can reference them by name + integration_id.
		ExportAsDataFunc: shouldExportIntegrationActionAsDataSource,
	}
}

// stripFunctionConfigFileContentHash drops the hash on export. Without the zip binary the
// hash is meaningless; the provider recomputes it from file_path on apply.
func stripFunctionConfigFileContentHash(configMap map[string]interface{}, _ map[string]*resourceExporter.ResourceExporter, _ string) error {
	delete(configMap, "file_content_hash")
	return nil
}

// DataSourceIntegrationAction registers the genesyscloud_integration_action data source
func DataSourceIntegrationAction() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for Genesys Cloud integration action. Select an integration action by name. " +
			"For static (built-in) data actions whose names may collide across integration instances, " +
			"integration_id and/or action_type can be provided to disambiguate the lookup.",
		ReadContext: provider.ReadWithPooledClient(dataSourceIntegrationActionRead),
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "The name of the integration action",
				Type:        schema.TypeString,
				Required:    true,
			},
			"integration_id": {
				Description: "The ID of the integration that owns the action. Optional, used to disambiguate " +
					"data actions whose names may not be unique across integration instances.",
				Type:     schema.TypeString,
				Optional: true,
			},
			"action_type": {
				Description: "The type of the integration action. Optional, used to disambiguate when a " +
					"static (built-in) action and a custom action share the same name under the same " +
					"integration. Valid values: `static`, `custom`.",
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"static", "custom"}, false),
			},
		},
	}
}
