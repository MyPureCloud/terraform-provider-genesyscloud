package auth_role

import (
	"context"
	"fmt"
	"log"
	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"terraform-provider-genesyscloud/genesyscloud/util/constants"
	"terraform-provider-genesyscloud/genesyscloud/util/lists"
	"terraform-provider-genesyscloud/genesyscloud/util/resourcedata"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v146/platformclientv2"
)

/*
The resource_genesyscloud_auth_role_utils.go contains all of the methods that perform the core logic for a resource.
*/

// getAllAuthAuthRole retrieves all of the auth role via Terraform in the Genesys Cloud and is used for the exporter
func getAllAuthAuthRoles(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(resourceExporter.ResourceIDMetaMap)
	proxy := getAuthRoleProxy(clientConfig)

	roles, proxyResponse, getErr := proxy.getAllAuthRole(ctx)
	if getErr != nil {
		return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get page of roles %s", getErr), proxyResponse)
	}

	for _, role := range *roles {
		resources[*role.Id] = &resourceExporter.ResourceMeta{BlockLabel: *role.Name}
	}

	return resources, nil
}

// createAuthRole is used by the auth_role resource to create Genesys cloud auth role
func createAuthRole(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getAuthRoleProxy(sdkConfig)

	// Validate each permission policy exists before continuing
	// This is a workaround for a bug in the auth roles APIs
	// Bug reported to auth team in ticket AUTHZ-315
	policies := buildSdkRolePermPolicies(d)
	if policies != nil {
		for _, policy := range *policies {
			resp, err := validatePermissionPolicy(proxy, policy)
			if err != nil {
				return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Permission policy not found: %s, ensure your org has the required product for this permission", err), resp)
			}
		}
	}

	name := d.Get("name").(string)
	description := d.Get("description").(string)
	defaultRoleID := d.Get("default_role_id").(string)

	log.Printf("Creating role %s", name)
	if defaultRoleID != "" {
		// Default roles must already exist, or they cannot be modified
		defaultRole, proxyResponse, err := proxy.getDefaultRoleById(ctx, defaultRoleID)
		if err != nil {
			return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get default role %s error: %s", d.Id(), err), proxyResponse)
		}
		d.SetId(defaultRole)
		return updateAuthRole(ctx, d, meta)
	}

	roleObj := platformclientv2.Domainorganizationrolecreate{
		Name:               &name,
		Description:        &description,
		Permissions:        buildSdkRolePermissions(d),
		PermissionPolicies: policies,
	}

	role, proxyResponse, err := proxy.createAuthRole(ctx, &roleObj)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to create role %s %s errror: %s", d.Id(), *roleObj.Name, err), proxyResponse)
	}

	d.SetId(*role.Id)
	log.Printf("Created role %s %s", name, *role.Id)
	return readAuthRole(ctx, d, meta)
}

// readAuthRole is used by the auth_role resource to read an auth role from genesys cloud
func readAuthRole(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getAuthRoleProxy(sdkConfig)
	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceAuthRole(), constants.ConsistencyChecks(), ResourceType)

	log.Printf("Reading role %s", d.Id())

	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		role, proxyResponse, getErr := proxy.getAuthRoleById(ctx, d.Id())
		if getErr != nil {
			if util.IsStatus404(proxyResponse) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read role %s | error: %s", d.Id(), getErr), proxyResponse))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read role %s | error: %s", d.Id(), getErr), proxyResponse))
		}

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
		return cc.CheckState(d)
	})
}

// updateAuthRole is used by the auth_role resource to update an auth role in Genesys Cloud
func updateAuthRole(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getAuthRoleProxy(sdkConfig)

	// Validate each permission policy exists before continuing
	// This is a workaround for a bug in the auth roles APIs
	// Bug reported to auth team in ticket AUTHZ-315
	policies := buildSdkRolePermPolicies(d)
	if policies != nil {
		for _, policy := range *policies {
			resp, err := validatePermissionPolicy(proxy, policy)
			if err != nil {
				return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Permission policy not found: %s, ensure your org has the required product for this permission", err), resp)
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
		PermissionPolicies: policies,
		DefaultRoleId:      &defaultRoleID,
	}
	_, proxyResponse, err := proxy.updateAuthRole(ctx, d.Id(), &roleObj)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to update role %s %s error: %s", d.Id(), *roleObj.Name, err), proxyResponse)
	}

	log.Printf("Updated role %s", name)
	return readAuthRole(ctx, d, meta)
}

// deleteAuthRole is used by the auth_role resource to delete an auth role from Genesys cloud
func deleteAuthRole(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getAuthRoleProxy(sdkConfig)

	name := d.Get("name").(string)
	defaultRoleID := d.Get("default_role_id").(string)

	if defaultRoleID != "" {
		// Restore default roles to their default state instead of deleting them
		log.Printf("Restoring default role %s", name)
		id := d.Id()
		proxyResponse, err := proxy.restoreDefaultRoles(ctx, &[]platformclientv2.Domainorganizationrole{
			{
				Id: &id,
			},
		})
		if err != nil {
			return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to restore default role %s error: %s", defaultRoleID, err), proxyResponse)
		}
		return nil
	}

	log.Printf("Deleting role %s", name)
	proxyResponse, err := proxy.deleteAuthRole(ctx, d.Id())
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to delete auth role %s %s error: %s", d.Id(), name, err), proxyResponse)
	}

	return util.WithRetries(ctx, 60*time.Second, func() *retry.RetryError {
		_, proxyResponse, err := proxy.getAuthRoleById(ctx, d.Id())
		if err != nil {
			if util.IsStatus404(proxyResponse) {
				// role deleted
				log.Printf("Deleted role %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Error deleting role %s | error: %s", d.Id(), err), proxyResponse))
		}
		return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Role %s still exists", d.Id()), proxyResponse))
	})
}
