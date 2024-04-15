package orgauthorization_pairing

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	registrar "terraform-provider-genesyscloud/genesyscloud/resource_register"
)

const resourceName = "genesyscloud_orgauthorization_pairing"

func SetRegistrar(regInstance registrar.Registrar) {
	regInstance.RegisterResource(resourceName, ResourceOrgauthorizationPairing())
}

func ResourceOrgauthorizationPairing() *schema.Resource {
	return &schema.Resource{
		Description: `Genesys Cloud orgauthorization pairing`,

		CreateContext: provider.CreateWithPooledClient(createOrgauthorizationPairing),
		ReadContext:   provider.ReadWithPooledClient(readOrgauthorizationPairing),
		DeleteContext: provider.DeleteWithPooledClient(deleteOrgauthorizationPairing),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			`user_ids`: {
				Description: `The list of trustee users that are requesting access. If no users are specified, at least one group is required.  Changing the user_ids attribute will cause the orgauthorization_pairing resource to be dropped and recreated with a new ID.`,
				Optional:    true,
				ForceNew:    true,
				Type:        schema.TypeList,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			`group_ids`: {
				Description: `The list of trustee groups that are requesting access. If no groups are specified, at least one user is required. Changing the group_ids attribute will cause the orgauthorization_pairing resource to be dropped and recreated with a new ID.`,
				Optional:    true,
				ForceNew:    true,
				Type:        schema.TypeList,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}
