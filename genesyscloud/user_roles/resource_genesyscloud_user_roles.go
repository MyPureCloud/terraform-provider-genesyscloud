package user_roles

import (
	"context"
	"fmt"
	"log"
	"strings"
	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

/*
The resource_genesyscloud_user_roles.go contains all the methods that perform the core logic for a resource
*/

func createUserRoles(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	userID := d.Get("user_id").(string)
	d.SetId(userID)
	log.Printf("Creating roles for user %s", d.Id())
	return updateUserRoles(ctx, d, meta)
}

func readUserRoles(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getUserRolesProxy(sdkConfig)

	log.Printf("Reading roles for user %s", d.Id())
	d.Set("user_id", d.Id())

	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		roles, resp, err := flattenSubjectRoles(d, proxy)
		if err != nil {
			if util.IsStatus404ByInt(resp.StatusCode) {
				return retry.RetryableError(fmt.Errorf("Failed to read roles for user %s: %v", d.Id(), err))
			}
			return retry.NonRetryableError(fmt.Errorf("Failed to read roles for user %s: %v", d.Id(), err))
		}
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
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getUserRolesProxy(sdkConfig)

	if !d.HasChange("roles") {
		return nil
	}
	rolesConfig := d.Get("roles").(*schema.Set)
	if rolesConfig == nil {
		return nil
	}

	log.Printf("Updating roles for user %s", d.Id())
	resp, diagErr := proxy.updateUserRoles(ctx, d.Id(), rolesConfig, "PC_USER")
	if diagErr != nil {
		return util.BuildAPIDiagnosticError(resourceName, fmt.Sprintf("Failed to update user roles %s error: %s", d.Id(), diagErr), resp)
	}

	log.Printf("Updated user roles for %s", d.Id())
	time.Sleep(4 * time.Second)
	return readUserRoles(ctx, d, meta)
}

func deleteUserRoles(_ context.Context, _ *schema.ResourceData, _ interface{}) diag.Diagnostics {
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
