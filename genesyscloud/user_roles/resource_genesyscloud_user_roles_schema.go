package user_roles

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"terraform-provider-genesyscloud/genesyscloud"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "terraform-provider-genesyscloud/genesyscloud/resource_register"
)

const resourceName = "genesyscloud_user_roles"

// SetRegistrar registers all of the resources and exporters in the package
func SetRegistrar(l registrar.Registrar) {
	l.RegisterResource(resourceName, ResourceUserRoles())
	l.RegisterExporter(resourceName, userRolesExporter())
}

// ResourceUserRoles registers the genesyscloud_user_roles resource with terraform
func ResourceUserRoles() *schema.Resource {
	return &schema.Resource{
		Description: `Genesys Cloud User Roles maintains user role assignments.

Terraform expects to manage the resources that are defined in its stack. You can use this resource to assign roles to existing users that are not managed by Terraform. However, one thing you have to remember is that when you use this resource to assign roles to existing users, you must define all roles assigned to those users in this resource. Otherwise, you will inadvertently drop all of the existing roles assigned to the user and replace them with the one defined in this resource. Keep this in mind, as the author of this note inadvertently stripped his Genesys admin account of administrator privileges while using this resource to assign a role to his account. The best lessons in life are often free and self-inflicted.`,

		CreateContext: genesyscloud.CreateWithPooledClient(createUserRoles),
		ReadContext:   genesyscloud.ReadWithPooledClient(readUserRoles),
		UpdateContext: genesyscloud.UpdateWithPooledClient(updateUserRoles),
		DeleteContext: genesyscloud.DeleteWithPooledClient(deleteUserRoles),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"user_id": {
				Description: "User ID that will be managed by this resource. Changing the user_id attribute will cause the roles object to be dropped and recreated with a new ID.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"roles": {
				Description: "Roles and their divisions assigned to this user.",
				Type:        schema.TypeSet,
				Optional:    true,
				Elem:        genesyscloud.RoleAssignmentResource,
			},
		},
	}
}

// userRolesExporter returns the resourceExporter object used to hold the genesyscloud_user_roles exporter's config
func userRolesExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: genesyscloud.GetAllWithPooledClient(genesyscloud.GetAllUsers),
		RefAttrs: map[string]*resourceExporter.RefAttrSettings{
			"user_id":            {RefType: "genesyscloud_user"},
			"roles.role_id":      {RefType: "genesyscloud_auth_role"},
			"roles.division_ids": {RefType: "genesyscloud_auth_division", AltValues: []string{"*"}},
		},
		RemoveIfMissing: map[string][]string{
			"roles": {"role_id"},
		},
	}
}
