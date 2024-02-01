package auth_role

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v121/platformclientv2"
	"log"
	gcloud "terraform-provider-genesyscloud/genesyscloud"
	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	"terraform-provider-genesyscloud/genesyscloud/util/lists"
	"terraform-provider-genesyscloud/genesyscloud/util/resourcedata"
	"time"
)

/*
The resource_genesyscloud_auth_role_utils.go contains all of the methods that perform the core logic for a resource.
*/

// getAllAuthAuthRole retrieves all of the auth role via Terraform in the Genesys Cloud and is used for the exporter
func getAllAuthAuthRoles(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(resourceExporter.ResourceIDMetaMap)
	authAPI := platformclientv2.NewAuthorizationApiWithConfig(clientConfig)

	for pageNum := 1; ; pageNum++ {
		const pageSize = 100
		roles, _, getErr := authAPI.GetAuthorizationRoles(pageSize, pageNum, "", nil, "", "", "", nil, nil, false, nil)
		if getErr != nil {
			return nil, diag.Errorf("Failed to get page of roles: %v", getErr)
		}

		if roles.Entities == nil || len(*roles.Entities) == 0 {
			break
		}

		for _, role := range *roles.Entities {
			resources[*role.Id] = &resourceExporter.ResourceMeta{Name: *role.Name}
		}
	}

	return resources, nil
}

// createAuthRole is used by the auth_role resource to create Genesys cloud auth role
func createAuthRole(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	proxy := getAuthRoleProxy(sdkConfig)

	// Validate each permission policy exists before continuing
	// This is a workaround for a bug in the auth roles APIs
	// Bug reported to auth team in ticket AUTHZ-315
	policies := buildSdkRolePermPolicies(d)
	if policies != nil {
		for _, policy := range *policies {
			err := validatePermissionPolicy(authAPI, &policy)
			if err != nil {
				return diag.Errorf("Permission policy not found: %s, ensure your org has the required product for this permission", err)
			}
		}
	}

	name := d.Get("name").(string)
	description := d.Get("description").(string)
	defaultRoleID := d.Get("default_role_id").(string)

	log.Printf("Creating role %s", name)
	if defaultRoleID != "" {
		// Default roles must already exist, or they cannot be modified
		defaultRole, _, err := proxy.getAuthRoleById(ctx, defaultRoleID)
		if err != nil {
			return diag.Errorf("Error requesting default role %s: %s", defaultRoleID, err)
		}
		d.SetId(*defaultRole.Id)
		return updateAuthRole(ctx, d, meta)
	}

	roleObj := platformclientv2.Domainorganizationrolecreate{
		Name:               &name,
		Description:        &description,
		Permissions:        buildSdkRolePermissions(d),
		PermissionPolicies: buildSdkRolePermPolicies(d),
	}

	// Validate each permission policy exists before continuing
	// This is a workaround for a bug in the auth roles APIs
	// Bug reported to auth team in ticket AUTHZ-315
	if roleObj.PermissionPolicies != nil {
		for _, policy := range *roleObj.PermissionPolicies {
			err := validatePermissionPolicy(proxy, policy)
			if err != nil {
				return diag.Errorf("%s", err)
			}
		}
	}

	role, err := proxy.createAuthRole(ctx, &roleObj)
	if err != nil {
		return diag.Errorf("Failed to create role %s: %s", name, err)
	}

	d.SetId(*role.Id)
	log.Printf("Created role %s %s", name, *role.Id)
	return readAuthRole(ctx, d, meta)
}

// readAuthRole is used by the auth_role resource to read an auth role from genesys cloud
func readAuthRole(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	proxy := getAuthRoleProxy(sdkConfig)

	log.Printf("Reading role %s", d.Id())

	return gcloud.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		role, respCode, getErr := proxy.getAuthRoleById(ctx, d.Id())
		if getErr != nil {
			if gcloud.IsStatus404ByInt(respCode) {
				return retry.RetryableError(fmt.Errorf("Failed to read role %s: %s", d.Id(), getErr))
			}
			return retry.NonRetryableError(fmt.Errorf("Failed to read role %s: %s", d.Id(), getErr))
		}

		cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceAuthRole())

		d.Set("name", *role.Name)
		resourcedata.SetNillableValue(d, "description", role.Description)
		resourcedata.SetNillableValue(d, "default_role_id", role.DefaultRoleId)

		if role.Permissions != nil {
			d.Set("permissions", lists.StringListToSet(*role.Permissions))
		} else {
			d.Set("permissions", nil)
		}

		if role.PermissionPolicies != nil {
			d.Set("permission_policies", flattenRolePermissionPolicies(*role.PermissionPolicies))
		} else {
			d.Set("permission_policies", nil)
		}

		log.Printf("Read role %s %s", d.Id(), *role.Name)
		return cc.CheckState()
	})
}

// updateAuthRole is used by the auth_role resource to update an auth role in Genesys Cloud
func updateAuthRole(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	proxy := getAuthRoleProxy(sdkConfig)

	// Validate each permission policy exists before continuing
	// This is a workaround for a bug in the auth roles APIs
	// Bug reported to auth team in ticket AUTHZ-315
	policies := buildSdkRolePermPolicies(d)
	if policies != nil {
		for _, policy := range *policies {
			err := validatePermissionPolicy(authAPI, &policy)
			if err != nil {
				return diag.Errorf("Permission policy not found: %s, ensure your org has the required product for this permission", err)
			}
		}
	}

	name := d.Get("name").(string)
	description := d.Get("description").(string)
	defaultRoleID := d.Get("default_role_id").(string)

	log.Printf("Updating role %s", name)
	roleObj := platformclientv2.Domainorganizationroleupdate{
		Name:               &name,
		Description:        &description,
		Permissions:        buildSdkRolePermissions(d),
		PermissionPolicies: buildSdkRolePermPolicies(d),
		DefaultRoleId:      &defaultRoleID,
	}
	_, err := proxy.updateAuthRole(ctx, d.Id(), &roleObj)
	if err != nil {
		return diag.Errorf("Failed to update role %s: %s", name, err)
	}

	log.Printf("Updated role %s", name)
	return readAuthRole(ctx, d, meta)
}

// deleteAuthRole is used by the auth_role resource to delete an auth role from Genesys cloud
func deleteAuthRole(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	proxy := getAuthRoleProxy(sdkConfig)

	name := d.Get("name").(string)
	defaultRoleID := d.Get("default_role_id").(string)

	if defaultRoleID != "" {
		// Restore default roles to their default state instead of deleting them
		log.Printf("Restoring default role %s", name)
		id := d.Id()
		err := proxy.restoreDefaultRoles(ctx, &[]platformclientv2.Domainorganizationrole{
			{
				Id: &id,
			},
		})
		if err != nil {
			return diag.Errorf("Failed to restore default role %s: %s", defaultRoleID, err)
		}
		return nil
	}

	log.Printf("Deleting role %s", name)
	_, err := proxy.deleteAuthRole(ctx, d.Id())
	if err != nil {
		return diag.Errorf("Failed to delete role %s: %s", name, err)
	}

	return gcloud.WithRetries(ctx, 60*time.Second, func() *retry.RetryError {
		_, resp, err := proxy.getAuthRoleById(ctx, d.Id())
		if err != nil {
			if gcloud.IsStatus404ByInt(resp) {
				// role deleted
				log.Printf("Deleted role %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(fmt.Errorf("Error deleting role %s: %s", d.Id(), err))
		}
		return retry.RetryableError(fmt.Errorf("Role %s still exists", d.Id()))
	})
}

func validatePermissionPolicy(authApi *platformclientv2.AuthorizationApi, policy *platformclientv2.Domainpermissionpolicy) error {
	allowedPermissions, err := getAllowedPermissions(authApi, *policy.Domain)
	if err != nil {
		return fmt.Errorf("error requesting org permissions: %s", err)
	}
	if len(*allowedPermissions) == 0 {
		return fmt.Errorf("domain %s not found", *policy.Domain)
	}

	if *policy.EntityName == "*" {
		return nil
	}

	// Check entity type (e.g. callableTimeSet) exists in the map of allowed permissions
	if entityPermissions, ok := (*allowedPermissions)[*policy.EntityName]; ok {
		// Check if the policy actions exist for the given domain permission e.g. callableTimeSet: add
		for _, action := range *policy.ActionSet {
			if action == "*" && len(entityPermissions) >= 1 {
				break
			}

			var found bool
			for _, entityPermission := range entityPermissions {
				if action == *entityPermission.Action {
					// action found, move to next action
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("action %s not found for domain %s, entity name %s", action, *policy.Domain, *policy.EntityName)
			}
		}
		// All actions have been found, permission exists
		return nil
	}

	return fmt.Errorf("entity_name %s not found for domain %s", *policy.EntityName, *policy.Domain)
}

// getAllowedPermissions will get all allowed permissions for a domain
func getAllowedPermissions(authApi *platformclientv2.AuthorizationApi, domain string) (*map[string][]platformclientv2.Domainpermission, error) {
	const pageSize = 100
	allowedPermissions := make(map[string][]platformclientv2.Domainpermission)

	permissions, _, err := authApi.GetAuthorizationPermissions(pageSize, 1, "domain", domain)
	if err != nil {
		return nil, err
	}

	if permissions.Entities == nil || len(*permissions.Entities) == 0 {
		return &allowedPermissions, nil
	}

	for _, permission := range *permissions.Entities {
		for entityType, entityPermissions := range *permission.PermissionMap {
			allowedPermissions[entityType] = entityPermissions
		}
	}

	for pageNum := 2; pageNum <= *permissions.PageCount; pageNum++ {
		permissions, _, err := authApi.GetAuthorizationPermissions(pageSize, pageNum, "domain", domain)
		if err != nil {
			return nil, err
		}
		if permissions.Entities == nil || len(*permissions.Entities) == 0 {
			break
		}

		for _, permission := range *permissions.Entities {
			for entityType, entityPermissions := range *permission.PermissionMap {
				allowedPermissions[entityType] = entityPermissions
			}
		}
	}

	return &allowedPermissions, nil
}
