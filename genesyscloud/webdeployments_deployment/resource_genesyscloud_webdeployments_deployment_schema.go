package webdeployments_deployment

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "terraform-provider-genesyscloud/genesyscloud/resource_register"
)

const resourceName = "genesyscloud_webdeployments_deployment"

// SetRegistrar registers all the resources, datasources and exporters in the package
func SetRegistrar(l registrar.Registrar) {
	l.RegisterDataSource(resourceName, DataSourceWebDeploymentsDeployment())
	l.RegisterResource(resourceName, ResourceWebDeployment())
	l.RegisterExporter(resourceName, WebDeploymentExporter())
}
func ResourceWebDeployment() *schema.Resource {
	return &schema.Resource{
		Description: "Genesys Cloud Web Deployment",

		CreateContext: provider.CreateWithPooledClient(createWebDeployment),
		ReadContext:   provider.ReadWithPooledClient(readWebDeployment),
		UpdateContext: provider.UpdateWithPooledClient(updateWebDeployment),
		DeleteContext: provider.DeleteWithPooledClient(deleteWebDeployment),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Deployment name",
				Type:        schema.TypeString,
				Required:    true,
			},
			"description": {
				Description: "Deployment description",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"allow_all_domains": {
				Description: "Whether all domains are allowed or not. allowedDomains must be empty when this is true.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
			"allowed_domains": {
				Description: "The list of domains that are approved to use this deployment; the list will be added to CORS headers for ease of web use.",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"flow_id": {
				Description: "A reference to the inboundshortmessage flow used by this deployment.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"status": {
				Description: "The current status of the deployment. Valid values: Pending, Active, Inactive, Error, Deleting.",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				ValidateFunc: validation.StringInSlice([]string{
					"Pending",
					"Active",
					"Inactive",
					"Error",
					"Deleting",
				}, false),
				DiffSuppressFunc: validateDeploymentStatusChange,
			},
			"configuration": {
				Description: "The published configuration version used by this deployment",
				Type:        schema.TypeList,
				Required:    true,
				MaxItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Required: true,
						},
						"version": {
							Type:             schema.TypeString,
							Optional:         true,
							Computed:         true,
							DiffSuppressFunc: alwaysDifferent, // The newly-computed configuration version is not available when computing the diff so we assume it will be different
						},
					},
				},
			},
		},
	}
}

func WebDeploymentExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllWebDeployments),
		RefAttrs: map[string]*resourceExporter.RefAttrSettings{
			"flow_id":          {RefType: "genesyscloud_flow"},
			"configuration.id": {RefType: "genesyscloud_webdeployments_configuration"},
		},
		ExcludedAttributes: []string{"configuration.version"},
	}
}

func DataSourceWebDeploymentsDeployment() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for Genesys Cloud Web Deployments. Select a deployment by name.",
		ReadContext: provider.ReadWithPooledClient(dataSourceDeploymentRead),
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "The name of the deployment",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}
