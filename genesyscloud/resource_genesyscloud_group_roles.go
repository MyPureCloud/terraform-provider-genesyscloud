package genesyscloud

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v80/platformclientv2"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/consistency_checker"
)

func groupRolesExporter() *ResourceExporter {
	return &ResourceExporter{
		GetResourcesFunc: getAllWithPooledClient(getAllGroups),
		RefAttrs: map[string]*RefAttrSettings{
			"group_id":           {RefType: "genesyscloud_group"},
			"roles.role_id":      {RefType: "genesyscloud_auth_role"},
			"roles.division_ids": {RefType: "genesyscloud_auth_division", AltValues: []string{"*"}},
		},
		RemoveIfMissing: map[string][]string{
			"roles": {"role_id"},
		},
	}
}

func resourceGroupRoles() *schema.Resource {
	return &schema.Resource{
		Description: `Genesys Cloud Group Roles maintains group role assignments.`,

		CreateContext: createWithPooledClient(createGroupRoles),
		ReadContext:   readWithPooledClient(readGroupRoles),
		UpdateContext: updateWithPooledClient(updateGroupRoles),
		DeleteContext: deleteWithPooledClient(deleteGroupRoles),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"group_id": {
				Description: "Group ID that will be managed by this resource.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"roles": {
				Description: "Roles and their divisions assigned to this group.",
				Type:        schema.TypeSet,
				Optional:    true,
				Elem:        roleAssignmentResource,
			},
		},
	}
}

func createGroupRoles(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	groupID := d.Get("group_id").(string)
	d.SetId(groupID)
	return updateGroupRoles(ctx, d, meta)
}

func readGroupRoles(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*providerMeta).ClientConfig
	authAPI := platformclientv2.NewAuthorizationApiWithConfig(sdkConfig)

	log.Printf("Reading roles for group %s", d.Id())

	return withRetriesForRead(ctx, d, func() *resource.RetryError {
		cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, resourceGroupRoles())
		d.Set("group_id", d.Id())

		roles, resp, err := readSubjectRoles(d.Id(), authAPI)
		if err != nil {
			if isStatus404(resp) {
				return resource.RetryableError(fmt.Errorf("Failed to read roles for group %s: %v", d.Id(), err))
			}
			return resource.NonRetryableError(fmt.Errorf("Failed to read roles for group %s: %v", d.Id(), err))
		}

		d.Set("roles", roles)

		log.Printf("Read roles for group %s", d.Id())
		return cc.CheckState()
	})
}

func updateGroupRoles(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*providerMeta).ClientConfig
	authAPI := platformclientv2.NewAuthorizationApiWithConfig(sdkConfig)

	log.Printf("Updating roles for group %s", d.Id())
	diagErr := updateSubjectRoles(ctx, d, authAPI, "PC_GROUP")
	if diagErr != nil {
		return diagErr
	}

	log.Printf("Updated group roles for %s", d.Id())
	return readGroupRoles(ctx, d, meta)
}

func deleteGroupRoles(_ context.Context, _ *schema.ResourceData, _ interface{}) diag.Diagnostics {
	// Does not delete groups or roles. This resource will just no longer manage roles.
	return nil
}
