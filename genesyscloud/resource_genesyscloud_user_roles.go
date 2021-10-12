package genesyscloud

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v56/platformclientv2"
)

func userRolesExporter() *ResourceExporter {
	return &ResourceExporter{
		GetResourcesFunc: getAllWithPooledClient(getAllUsers),
		RefAttrs: map[string]*RefAttrSettings{
			"user_id":            {RefType: "genesyscloud_user"},
			"roles.role_id":      {RefType: "genesyscloud_auth_role"},
			"roles.division_ids": {RefType: "genesyscloud_auth_division", AltValues: []string{"*"}},
		},
		RemoveIfMissing: map[string][]string{
			"roles": {"role_id"},
		},
	}
}

func resourceUserRoles() *schema.Resource {
	return &schema.Resource{
		Description: `Genesys Cloud User Roles maintains user role assignments.`,

		CreateContext: createWithPooledClient(createUserRoles),
		ReadContext:   readWithPooledClient(readUserRoles),
		UpdateContext: updateWithPooledClient(updateUserRoles),
		DeleteContext: deleteWithPooledClient(deleteUserRoles),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"user_id": {
				Description: "User ID that will be managed by this resource.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"roles": {
				Description: "Roles and their divisions assigned to this user.",
				Type:        schema.TypeSet,
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
	sdkConfig := meta.(*providerMeta).ClientConfig
	authAPI := platformclientv2.NewAuthorizationApiWithConfig(sdkConfig)

	log.Printf("Reading roles for user %s", d.Id())

	d.Set("user_id", d.Id())

	roles, err := readSubjectRoles(d.Id(), authAPI)
	if err != nil {
		return err
	}
	d.Set("roles", roles)

	log.Printf("Read roles for user %s", d.Id())
	return nil
}

func updateUserRoles(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*providerMeta).ClientConfig
	authAPI := platformclientv2.NewAuthorizationApiWithConfig(sdkConfig)

	log.Printf("Updating roles for user %s", d.Id())
	diagErr := updateSubjectRoles(ctx, d, authAPI, "PC_USER")
	if diagErr != nil {
		return diagErr
	}

	log.Printf("Updated user roles for %s", d.Id())
	return readUserRoles(ctx, d, meta)
}

func deleteUserRoles(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Does not delete users or roles. This resource will just no longer manage roles.
	return nil
}
