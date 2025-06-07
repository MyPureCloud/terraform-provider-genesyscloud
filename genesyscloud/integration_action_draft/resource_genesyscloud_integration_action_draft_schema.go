package integration_action_draft

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_register"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
)

const ResourceType = "genesyscloud_integration_action_draft"

// SetRegistrar registers all of the resources, datasources and exporters in the package
func SetRegistrar(l registrar.Registrar) {
	//l.RegisterDataSource(ResourceType, DataSourceIntegrationActionDraft())
	l.RegisterResource(ResourceType, ResourceIntegrationActionDraft())
	l.RegisterExporter(ResourceType, IntegrationActionDraftExporter())
}

// ResourceIntegrationActionDraft registers the genesyscloud_integration_action_draft resource with Terraform
func ResourceIntegrationActionDraft() *schema.Resource {
	actionDraftConfigRequest := &schema.Resource{
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

	actionDraftConfigResponse := &schema.Resource{
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
		Description: "Genesys Cloud Integration Action Drafts. See this page for detailed information on configuring Actions: https://help.mypurecloud.com/articles/add-configuration-custom-actions-integrations/",

		CreateContext: provider.CreateWithPooledClient(createIntegrationActionDraft),
		ReadContext:   provider.ReadWithPooledClient(readIntegrationActionDraft),
		UpdateContext: provider.UpdateWithPooledClient(updateIntegrationActionDraft),
		DeleteContext: provider.DeleteWithPooledClient(deleteIntegrationActionDraft),
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
			"config_request": {
				Description: "Configuration of outbound request.",
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Elem:        actionDraftConfigRequest,
			},
			"config_response": {
				Description: "Configuration of response processing.",
				Type:        schema.TypeList,
				Optional:    true,
				Computed:    true,
				MaxItems:    1,
				Elem:        actionDraftConfigResponse,
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
		},
	}
}

// IntegrationActionDraftExporter returns the resourceExporter object used to hold the genesyscloud_integration_action_draft exporter's config
func IntegrationActionDraftExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllIntegrationActionDrafts),
		RefAttrs: map[string]*resourceExporter.RefAttrSettings{
			"integration_id": {RefType: "genesyscloud_integration"},
		},
		JsonEncodeAttributes: []string{"contract_input", "contract_output"},
	}
}

// DataSourceIntegrationActionDraft registers the genesyscloud_integration_action_draft data source
func DataSourceIntegrationActionDraft() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for Genesys Cloud integration action draft. Select an integration action by name",
		ReadContext: provider.ReadWithPooledClient(dataSourceIntegrationActionDraftRead),
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "The name of the integration action",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}
