package architect_ivr_identity_resolution

// @team: Architect Team
// @chat: #Genesys Cloud Architect support
// @pm: Amelie Wisniak
// @jira: RELATE-25226
// @description: manages user and system data for Genesys Cloud Architect. This includes Architect flows, user and system prompts as well as flow outcomes.

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	authDivision "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/auth_division"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_register"
)

const ResourceType = "genesyscloud_architect_ivr_identity_resolution"

// SetRegistrar registers all the resources and exporters in the package.
func SetRegistrar(regInstance registrar.Registrar) {
	regInstance.RegisterResource(ResourceType, ResourceArchitectIvrIdentityResolution())
	regInstance.RegisterExporter(ResourceType, ArchitectIvrIdentityResolutionExporter())
}

func ResourceArchitectIvrIdentityResolution() *schema.Resource {
	return &schema.Resource{
		Description:   `Genesys Cloud architect IVR identity resolution settings. Destroy restores the IVR to its default identity resolution configuration (resolve_identities = true, unassigned division).`,
		CreateContext: provider.CreateWithPooledClient(createArchitectIvrIdentityResolution),
		ReadContext:   provider.ReadWithPooledClient(readArchitectIvrIdentityResolution),
		UpdateContext: provider.UpdateWithPooledClient(updateArchitectIvrIdentityResolution),
		DeleteContext: provider.DeleteWithPooledClient(deleteArchitectIvrIdentityResolution),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 0,
		Schema: map[string]*schema.Schema{
			"ivr_id": {
				Description: "ID of the architect IVR this identity resolution config belongs to.",
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
		},
	}
}

func ArchitectIvrIdentityResolutionExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllArchitectIvrIdentityResolution),
		RefAttrs: map[string]*resourceExporter.RefAttrSettings{
			"ivr_id": {RefType: "genesyscloud_architect_ivr"},
			"division_id": {
				RefType:   authDivision.ResourceType,
				AltValues: []string{"*"},
			},
		},
	}
}
