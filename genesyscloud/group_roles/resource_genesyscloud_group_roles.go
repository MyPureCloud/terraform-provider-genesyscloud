package group_roles

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"log"
	"terraform-provider-genesyscloud/genesyscloud"
	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

/*
The resource_genesyscloud_group_roles.go contains all the methods that perform the core logic for a resource
*/

func createGroupRoles(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	groupID := d.Get("group_id").(string)
	d.SetId(groupID)
	return updateGroupRoles(ctx, d, meta)
}

func deleteGroupRoles(_ context.Context, _ *schema.ResourceData, _ interface{}) diag.Diagnostics {
	// Does not delete groups or roles. This resource will just no longer manage roles.
	return nil
}

// readGroupRoles is used by the group_roles resource to read Group Roles from the genesys cloud
func readGroupRoles(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*genesyscloud.ProviderMeta).ClientConfig
	proxy := getGroupRolesProxy(sdkConfig)

	log.Printf("Reading roles for group %s", d.Id())

	return genesyscloud.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceGroupRoles())
		d.Set("group_id", d.Id())

		roles, resp, err := flattenSubjectRoles(d, proxy)
		if err != nil {
			if genesyscloud.IsStatus404ByInt(resp.StatusCode) {
				return retry.RetryableError(fmt.Errorf("Failed to read roles for group %s: %v", d.Id(), err))
			}
			return retry.NonRetryableError(fmt.Errorf("Failed to read roles for group %s: %v", d.Id(), err))
		}
		d.Set("roles", roles)

		log.Printf("Read roles for group %s", d.Id())
		return cc.CheckState()
	})
}

func updateGroupRoles(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*genesyscloud.ProviderMeta).ClientConfig
	proxy := getGroupRolesProxy(sdkConfig)

	if !d.HasChange("roles") {
		return nil
	}
	rolesConfig := d.Get("roles").(*schema.Set)
	if rolesConfig == nil {
		return nil
	}

	log.Printf("Updating roles for group %s", d.Id())
	_, diagErr := proxy.updateGroupRoles(ctx, d.Id(), rolesConfig, "PC_GROUP")

	if diagErr != nil {

		return diag.Errorf("error %v", diagErr)
	}

	log.Printf("Updated group roles %v", d.Id())
	return readGroupRoles(ctx, d, meta)
}
