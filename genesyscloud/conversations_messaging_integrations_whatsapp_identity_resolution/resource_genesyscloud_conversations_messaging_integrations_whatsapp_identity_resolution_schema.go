package conversations_messaging_integrations_whatsapp_identity_resolution

// @team: Messaging Platform
// @chat: Messaging Platform
// @jira: PLT
// @description: Messaging Service is the core of Messaging Platform. It provides the APIs for normalized digital private messaging and public social interactions across a range of external providers

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	authDivision "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/auth_division"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_register"
)

const ResourceType = "genesyscloud_conversations_messaging_integrations_whatsapp_identity_resolution"

// SetRegistrar registers all the resources and exporters in the package.
func SetRegistrar(regInstance registrar.Registrar) {
	regInstance.RegisterResource(ResourceType, ResourceConversationsMessagingIntegrationsWhatsappIdentityResolution())
	regInstance.RegisterExporter(ResourceType, ConversationsMessagingIntegrationsWhatsappIdentityResolutionExporter())
}

func ResourceConversationsMessagingIntegrationsWhatsappIdentityResolution() *schema.Resource {
	return &schema.Resource{
		Description:   `Genesys Cloud conversations messaging integrations whatsapp identity resolution settings. Destroy restores the whatsapp integration to its default identity resolution configuration (resolve_identities = true, unassigned division).`,
		CreateContext: provider.CreateWithPooledClient(createConversationsMessagingIntegrationsWhatsappIdentityResolution),
		ReadContext:   provider.ReadWithPooledClient(readConversationsMessagingIntegrationsWhatsappIdentityResolution),
		UpdateContext: provider.UpdateWithPooledClient(updateConversationsMessagingIntegrationsWhatsappIdentityResolution),
		DeleteContext: provider.DeleteWithPooledClient(deleteConversationsMessagingIntegrationsWhatsappIdentityResolution),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 0,
		Schema: map[string]*schema.Schema{
			"whatsapp_integration_id": {
				Description: "ID of the whatsapp integration this identity resolution config belongs to.",
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

func ConversationsMessagingIntegrationsWhatsappIdentityResolutionExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllConversationsMessagingIntegrationsWhatsappIdentityResolution),
		RefAttrs: map[string]*resourceExporter.RefAttrSettings{
			"whatsapp_integration_id": {RefType: "genesyscloud_conversations_messaging_integrations_whatsapp"},
			"division_id": {
				RefType:   authDivision.ResourceType,
				AltValues: []string{"*"},
			},
		},
	}
}
