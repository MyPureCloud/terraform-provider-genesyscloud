package integration_config

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_register"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
)

const ResourceType = "genesyscloud_integration_config"

func SetRegistrar(l registrar.Registrar) {
	l.RegisterResource(ResourceType, ResourceIntegrationConfig())
	l.RegisterExporter(ResourceType, IntegrationConfigExporter())
}

func ResourceIntegrationConfig() *schema.Resource {
	return &schema.Resource{
		Description: `Genesys Cloud Integration Config. This resource manages the configuration (credentials, properties, advanced settings) 
for an integration separately from the integration resource itself. 

**Important:** This resource requires the ENABLE_STANDALONE_INTEGRATION_CONFIG environment variable to be set. 
When enabled, the config block on genesyscloud_integration will not be read or managed — this resource takes over. 
The two approaches are mutually exclusive.`,

		CreateContext: provider.CreateWithPooledClient(createIntegrationConfig),
		ReadContext:   provider.ReadWithPooledClient(readIntegrationConfig),
		UpdateContext: provider.UpdateWithPooledClient(updateIntegrationConfig),
		DeleteContext: provider.DeleteWithPooledClient(deleteIntegrationConfig),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"integration_id": {
				Description: "The ID of the integration this config belongs to.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"name": {
				Description: "Integration config name. Used to distinguish this integration from others of the same type.",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},
			"notes": {
				Description: "Notes about the integration.",
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
			},
			"properties": {
				Description:      "Integration config properties (JSON string). Schema varies by integration type.",
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				DiffSuppressFunc: util.SuppressEquivalentJsonDiffs,
			},
			"advanced": {
				Description:      "Integration advanced config (JSON string). Schema varies by integration type.",
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				DiffSuppressFunc: util.SuppressEquivalentJsonDiffs,
			},
			"credentials": {
				Description: "Credentials for the integration. Map of credential type to credential ID. The required keys are indicated in the credentials property of the Integration Type.",
				Type:        schema.TypeMap,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func IntegrationConfigExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllIntegrationConfigs),
		RefAttrs: map[string]*resourceExporter.RefAttrSettings{
			"integration_id": {RefType: "genesyscloud_integration"},
			"credentials.*":  {RefType: "genesyscloud_integration_credential"},
		},
		JsonEncodeAttributes: []string{"properties", "advanced"},
	}
}
