package external_user

import (
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_register"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const ResourceType = "genesyscloud_externalusers_identity"

// SetRegistrar registers all of the resources and exporters in the package
func SetRegistrar(l registrar.Registrar) {
	l.RegisterResource(ResourceType, ResourceExternalUserIdentity())
	l.RegisterExporter(ResourceType, ExternalUserIdentityExporter())
}

// ResourceExternalContact registers the resource with Terraform
func ResourceExternalUserIdentity() *schema.Resource {

	return &schema.Resource{
		Description: "Genesys Cloud External Contact",

		CreateContext: provider.CreateWithPooledClient(createExternalUser),
		ReadContext:   provider.ReadWithPooledClient(readExternalUser),
		UpdateContext: provider.UpdateWithPooledClient(updateExternalUser),
		DeleteContext: provider.DeleteWithPooledClient(deleteExternalUser),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"authority_name": {
				Description: "Authority or System of Record which owns the External Identifier",
				Type:        schema.TypeString,
				Required:    true,
			},
			"external_key": {
				Description: "The identifier for the user within the Authority that owns the external identifier",
				Type:        schema.TypeString,
				Required:    true,
			},
			"user_id": {
				Description: "The user identifier inside the genesys org",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}

func ExternalUserIdentityExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllExternalUserIdentity),
		RefAttrs: map[string]*resourceExporter.RefAttrSettings{
			"user_id": {RefType: "genesyscloud_user"},
		},
	}
}
