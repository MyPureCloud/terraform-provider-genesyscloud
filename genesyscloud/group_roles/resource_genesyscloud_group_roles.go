package group_roles

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"log"
	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"terraform-provider-genesyscloud/genesyscloud/util/constants"

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
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getGroupRolesProxy(sdkConfig)
	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceGroupRoles(), constants.ConsistencyChecks(), ResourceType)

	log.Printf("Reading roles for group %s", d.Id())

	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		d.Set("group_id", d.Id())

		roles, resp, err := flattenSubjectRoles(d, proxy)
		if err != nil {
			if util.IsStatus404ByInt(resp.StatusCode) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read roles for group %s | error: %v", d.Id(), err), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read roles for group %s | error: %v", d.Id(), err), resp))
		}
		d.Set("roles", roles)

		log.Printf("Read roles for group %s", d.Id())
		return cc.CheckState(d)
	})
}

func updateGroupRoles(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getGroupRolesProxy(sdkConfig)

	if !d.HasChange("roles") {
		return nil
	}
	rolesConfig := d.Get("roles").(*schema.Set)
	if rolesConfig == nil {
		return nil
	}

	log.Printf("Updating roles for group %s", d.Id())
	resp, diagErr := proxy.updateGroupRoles(ctx, d.Id(), rolesConfig, "PC_GROUP")

	if diagErr != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to update group role %s error: %s", d.Id(), diagErr), resp)
	}
	log.Printf("Updated group roles %v", d.Id())
	return readGroupRoles(ctx, d, meta)
}
