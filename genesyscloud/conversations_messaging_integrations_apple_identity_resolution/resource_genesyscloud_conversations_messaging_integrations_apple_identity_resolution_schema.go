package conversations_messaging_integrations_apple_identity_resolution

// @team: AAPL
// @chat: AAPL-DEV
// @pm: Katie Ritz
// @jira: AAPL
// @description: Messaging Connector for Apple Messages for Business. Allows Apple Messages to send inbound and receive outbound messages through the Messaging Platform.

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	authDivision "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/auth_division"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_register"
)

const ResourceType = "genesyscloud_conversations_messaging_integrations_apple_identity_resolution"

// SetRegistrar registers all the resources and exporters in the package.
func SetRegistrar(regInstance registrar.Registrar) {
	regInstance.RegisterResource(ResourceType, ResourceConversationsMessagingIntegrationsAppleIdentityResolution())
	regInstance.RegisterExporter(ResourceType, ConversationsMessagingIntegrationsAppleIdentityResolutionExporter())
}

func ResourceConversationsMessagingIntegrationsAppleIdentityResolution() *schema.Resource {
	return &schema.Resource{
		Description:   `Genesys Cloud conversations messaging integrations apple identity resolution settings. Destroy restores the apple integration to its default identity resolution configuration (resolve_identities = true, unassigned division).`,
		CreateContext: provider.CreateWithPooledClient(createConversationsMessagingIntegrationsAppleIdentityResolution),
		ReadContext:   provider.ReadWithPooledClient(readConversationsMessagingIntegrationsAppleIdentityResolution),
		UpdateContext: provider.UpdateWithPooledClient(updateConversationsMessagingIntegrationsAppleIdentityResolution),
		DeleteContext: provider.DeleteWithPooledClient(deleteConversationsMessagingIntegrationsAppleIdentityResolution),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 0,
		Schema: map[string]*schema.Schema{
			"apple_integration_id": {
				Description: "ID of the apple integration this identity resolution config belongs to.",
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

func ConversationsMessagingIntegrationsAppleIdentityResolutionExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllConversationsMessagingIntegrationsAppleIdentityResolution),
		RefAttrs: map[string]*resourceExporter.RefAttrSettings{
			"apple_integration_id": {RefType: "genesyscloud_conversations_messaging_integrations_apple"},
			"division_id": {
				RefType:   authDivision.ResourceType,
				AltValues: []string{"*"},
			},
		},
	}
}
