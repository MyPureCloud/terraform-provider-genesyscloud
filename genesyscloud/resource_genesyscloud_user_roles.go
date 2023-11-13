package genesyscloud

import (
	"context"
	"fmt"
	"log"
	"strings"
	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v115/platformclientv2"
)

func UserRolesExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: GetAllWithPooledClient(getAllUsers),
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

func ResourceUserRoles() *schema.Resource {
	return &schema.Resource{
		Description: `Genesys Cloud User Roles maintains user role assignments.

Terraform expects to manage the resources that are defined in its stack. You can use this resource to assign roles to existing users that are not managed by Terraform. However, one thing you have to remember is that when you use this resource to assign roles to existing users, you must define all roles assigned to those users in this resource. Otherwise, you will inadvertently drop all of the existing roles assigned to the user and replace them with the one defined in this resource. Keep this in mind, as the author of this note inadvertently stripped his Genesys admin account of administrator privileges while using this resource to assign a role to his account. The best lessons in life are often free and self-inflicted.`,

		CreateContext: CreateWithPooledClient(createUserRoles),
		ReadContext:   ReadWithPooledClient(readUserRoles),
		UpdateContext: UpdateWithPooledClient(updateUserRoles),
		DeleteContext: DeleteWithPooledClient(deleteUserRoles),
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
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        roleAssignmentResource,
			},
		},
	}
}

func createUserRoles(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	userID := d.Get("user_id").(string)
	d.SetId(userID)
	return updateUserRoles(ctx, d, meta)
}

func readUserRoles(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*ProviderMeta).ClientConfig
	authAPI := platformclientv2.NewAuthorizationApiWithConfig(sdkConfig)

	log.Printf("Reading roles for user %s", d.Id())
	d.Set("user_id", d.Id())
	return WithRetriesForRead(ctx, d, func() *retry.RetryError {
		roles, _, err := readSubjectRoles(d, d.Id(), authAPI)
		if err != nil {
			return retry.NonRetryableError(fmt.Errorf("%v", err))
		}
		cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceUserRoles())
		_ = d.Set("roles", roles)

		log.Printf("Read roles for user %s", d.Id())
		return cc.CheckState()
	})
}

func updateUserRoles(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*ProviderMeta).ClientConfig
	authAPI := platformclientv2.NewAuthorizationApiWithConfig(sdkConfig)

	log.Printf("Updating roles for user %s", d.Id())
	diagErr := updateSubjectRoles(ctx, d, authAPI, "PC_USER")
	if diagErr != nil {
		return diagErr
	}

	log.Printf("Updated user roles for %s", d.Id())
	time.Sleep(4 * time.Second)
	return readUserRoles(ctx, d, meta)
}

func deleteUserRoles(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Does not delete users or roles. This resource will just no longer manage roles.
	return nil
}

func GenerateUserRoles(resourceID string, userResource string, roles ...string) string {
	return fmt.Sprintf(`resource "genesyscloud_user_roles" "%s" {
		user_id = genesyscloud_user.%s.id
		%s
	}
	`, resourceID, userResource, strings.Join(roles, "\n"))
}
