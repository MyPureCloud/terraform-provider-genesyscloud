package webdeployments_deployment_identity_resolution

// @jira: RELATE-25225
// @description: Identity resolution settings for a web deployment. Controls whether contacts are stitched for the deployment's channels, which division performs the stitching, the external source used, and whether authenticated web messaging and web tracking sessions are automerged.

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	authDivision "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/auth_division"
	externalSource "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/external_contacts_external_source"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_register"
	webDeployment "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/webdeployments_deployment"
)

const ResourceType = "genesyscloud_webdeployments_deployment_identity_resolution"

// SetRegistrar registers all the resources and exporters in the package.
func SetRegistrar(regInstance registrar.Registrar) {
	regInstance.RegisterResource(ResourceType, ResourceWebDeploymentIdentityResolution())
	regInstance.RegisterExporter(ResourceType, WebDeploymentIdentityResolutionExporter())
}

var automergeConfigSchema = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"authenticated_web_messaging": {
			Description: "Whether automerging is enabled for Authenticated Webmessaging conversations in this channel.",
			Type:        schema.TypeBool,
			Required:    true,
		},
		"web_tracking": {
			Description: "Whether automerging is enabled for Web Tracking sessions in this channel.",
			Type:        schema.TypeBool,
			Required:    true,
		},
	},
}

func ResourceWebDeploymentIdentityResolution() *schema.Resource {
	return &schema.Resource{
		Description:   `Genesys Cloud web deployment identity resolution config`,
		CreateContext: provider.CreateWithPooledClient(createWebDeploymentIdentityResolution),
		ReadContext:   provider.ReadWithPooledClient(readWebDeploymentIdentityResolution),
		UpdateContext: provider.UpdateWithPooledClient(updateWebDeploymentIdentityResolution),
		DeleteContext: provider.DeleteWithPooledClient(deleteWebDeploymentIdentityResolution),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 0,
		Schema: map[string]*schema.Schema{
			"deployment_id": {
				Description: "ID of the web deployment this identity resolution config belongs to.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"resolve_identities": {
				Description: "Whether identities should be resolved.",
				Type:        schema.TypeBool,
				Required:    true,
			},
			"division_id": {
				Description:      "The division to use when performing identity resolution. If not set, * means the unassigned (star) division. '*' may also be set explicitly for the same behavior.",
				Type:             schema.TypeString,
				Optional:         true,
				DiffSuppressFunc: suppressUnassignedDivisionIdDiff,
			},
			"external_source_id": {
				Description: "The external source used for stitching this channel.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"automerge_config": {
				Description: "Whether automerging of contacts should be enabled for each channel.",
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Elem:        automergeConfigSchema,
			},
		},
	}
}

func WebDeploymentIdentityResolutionExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllWebDeploymentIdentityResolutions),
		RefAttrs: map[string]*resourceExporter.RefAttrSettings{
			"deployment_id": {RefType: webDeployment.ResourceType},
			"division_id": {
				RefType:   authDivision.ResourceType,
				AltValues: []string{"*"},
			},
			"external_source_id": {RefType: externalSource.ResourceType},
		},
	}
}
