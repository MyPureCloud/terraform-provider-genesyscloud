package group_roles

import (
	"context"
	"fmt"
	"log"
	"terraform-provider-genesyscloud/genesyscloud"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v119/platformclientv2"
)

/*
The resource_genesyscloud_group_roles.go contains all of the methods that perform the core logic for a resource
*/

func createGroupRoles(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	groupID := d.Get("group_id").(string)
	d.SetId(groupID)
	return updateGroupRoles(ctx, d, meta)
}

// readGroupRoles is used by the group_roles resource to read Group Roles from the genesys cloud
func readGroupRoles(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*genesyscloud.ProviderMeta).ClientConfig
	authAPI := platformclientv2.NewAuthorizationApiWithConfig(sdkConfig)

	log.Printf("Reading roles for group %s", d.Id())

	return genesyscloud.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceGroupRoles())
		_ = d.Set("group_id", d.Id())

		roles, resp, err := genesyscloud.ReadSubjectRoles(d, authAPI)
		if err != nil {
			if genesyscloud.IsStatus404(resp) {
				return retry.RetryableError(fmt.Errorf("Failed to read roles for group %s: %v", d.Id(), err))
			}
			return retry.NonRetryableError(fmt.Errorf("Failed to read roles for group %s: %v", d.Id(), err))
		}

		_ = d.Set("roles", roles)

		log.Printf("Read roles for group %s", d.Id())
		return cc.CheckState()
	})
}

func updateGroupRoles(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*genesyscloud.ProviderMeta).ClientConfig
	authAPI := platformclientv2.NewAuthorizationApiWithConfig(sdkConfig)

	log.Printf("Updating roles for group %s", d.Id())
	diagErr := genesyscloud.UpdateSubjectRoles(ctx, d, authAPI, "PC_GROUP")
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
