package conversations_messaging_integrations_open_identity_resolution

// @team: Messaging Platform
// @chat: Messaging Platform
// @jira: PLT
// @description: Messaging Connector Open is the Open Messaging connector for Messaging Platform (Platypus).

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	authDivision "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/auth_division"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_register"
)

const ResourceType = "genesyscloud_conversations_messaging_integrations_open_identity_resolution"

// SetRegistrar registers all the resources and exporters in the package.
func SetRegistrar(regInstance registrar.Registrar) {
	regInstance.RegisterResource(ResourceType, ResourceConversationsMessagingIntegrationsOpenIdentityResolution())
	regInstance.RegisterExporter(ResourceType, ConversationsMessagingIntegrationsOpenIdentityResolutionExporter())
}

func ResourceConversationsMessagingIntegrationsOpenIdentityResolution() *schema.Resource {
	return &schema.Resource{
		Description:   `Genesys Cloud conversations messaging integrations open identity resolution settings. Destroy restores the open integration to its default identity resolution configuration (resolve_identities = true, unassigned division).`,
		CreateContext: provider.CreateWithPooledClient(createConversationsMessagingIntegrationsOpenIdentityResolution),
		ReadContext:   provider.ReadWithPooledClient(readConversationsMessagingIntegrationsOpenIdentityResolution),
		UpdateContext: provider.UpdateWithPooledClient(updateConversationsMessagingIntegrationsOpenIdentityResolution),
		DeleteContext: provider.DeleteWithPooledClient(deleteConversationsMessagingIntegrationsOpenIdentityResolution),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 0,
		Schema: map[string]*schema.Schema{
			"open_integration_id": {
				Description: "ID of the open integration this identity resolution config belongs to.",
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
				Description: "ID of the external source to use when performing identity resolution.",
				Type:        schema.TypeString,
				Optional:    true,
			},
		},
	}
}

func ConversationsMessagingIntegrationsOpenIdentityResolutionExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllConversationsMessagingIntegrationsOpenIdentityResolution),
		RefAttrs: map[string]*resourceExporter.RefAttrSettings{
			"open_integration_id": {RefType: "genesyscloud_conversations_messaging_integrations_open"},
			"division_id": {
				RefType:   authDivision.ResourceType,
				AltValues: []string{"*"},
			},
			"external_source_id": {
				RefType: "genesyscloud_externalcontacts_external_source",
			},
		},
	}
}
