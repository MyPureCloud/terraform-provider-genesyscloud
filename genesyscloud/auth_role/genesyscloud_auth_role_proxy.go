package auth_role

import (
	"context"
	"fmt"

	rc "terraform-provider-genesyscloud/genesyscloud/resource_cache"

	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

/*
The genesyscloud_auth_role_proxy.go file contains the proxy structures and methods that interact
with the Genesys Cloud SDK. We use composition here for each function on the proxy so individual functions can be stubbed
out during testing.
*/

// internalProxy holds a proxy instance that can be used throughout the package
var internalProxy *authRoleProxy

// Type definitions for each func on our proxy so we can easily mock them out later
type createAuthRoleFunc func(ctx context.Context, p *authRoleProxy, domainOrganizationRole *platformclientv2.Domainorganizationrolecreate) (*platformclientv2.Domainorganizationrole, *platformclientv2.APIResponse, error)
type getAllAuthRoleFunc func(ctx context.Context, p *authRoleProxy) (*[]platformclientv2.Domainorganizationrole, *platformclientv2.APIResponse, error)
type getAuthRoleIdByNameFunc func(ctx context.Context, p *authRoleProxy, name string) (id string, retryable bool, response *platformclientv2.APIResponse, err error)
type getAuthRoleByIdFunc func(ctx context.Context, p *authRoleProxy, id string) (domainOrganizationRole *platformclientv2.Domainorganizationrole, response *platformclientv2.APIResponse, err error)
type getDefaultRoleIdFunc func(ctx context.Context, p *authRoleProxy, defaultRoleID string) (roleId string, response *platformclientv2.APIResponse, err error)
type updateAuthRoleFunc func(ctx context.Context, p *authRoleProxy, id string, domainOrganizationRole *platformclientv2.Domainorganizationroleupdate) (*platformclientv2.Domainorganizationrole, *platformclientv2.APIResponse, error)
type deleteAuthRoleFunc func(ctx context.Context, p *authRoleProxy, id string) (response *platformclientv2.APIResponse, err error)
type restoreDefaultRolesFunc func(ctx context.Context, p *authRoleProxy, roles *[]platformclientv2.Domainorganizationrole) (*platformclientv2.APIResponse, error)
type getAllowedPermissionsFunc func(p *authRoleProxy, domain string) (*map[string][]platformclientv2.Domainpermission, *platformclientv2.APIResponse, error)

// authRoleProxy contains all of the methods that call genesys cloud APIs.
type authRoleProxy struct {
	clientConfig              *platformclientv2.Configuration
	authorizationApi          *platformclientv2.AuthorizationApi
	createAuthRoleAttr        createAuthRoleFunc
	getAllAuthRoleAttr        getAllAuthRoleFunc
	getAuthRoleIdByNameAttr   getAuthRoleIdByNameFunc
	getAuthRoleByIdAttr       getAuthRoleByIdFunc
	getDefaultRoleIdAttr      getDefaultRoleIdFunc
	updateAuthRoleAttr        updateAuthRoleFunc
	deleteAuthRoleAttr        deleteAuthRoleFunc
	restoreDefaultRolesAttr   restoreDefaultRolesFunc
	getAllowedPermissionsAttr getAllowedPermissionsFunc
	authRoleCache             rc.CacheInterface[platformclientv2.Domainorganizationrole]
}

// newAuthRoleProxy initializes the auth role proxy with all of the data needed to communicate with Genesys Cloud
func newAuthRoleProxy(clientConfig *platformclientv2.Configuration) *authRoleProxy {
	api := platformclientv2.NewAuthorizationApiWithConfig(clientConfig)
	authRoleCache := rc.NewResourceCache[platformclientv2.Domainorganizationrole]() // Create Cache for authRole resource
	return &authRoleProxy{
		clientConfig:              clientConfig,
		authorizationApi:          api,
		authRoleCache:             authRoleCache,
		createAuthRoleAttr:        createAuthRoleFn,
		getAllAuthRoleAttr:        getAllAuthRoleFn,
		getAuthRoleIdByNameAttr:   getAuthRoleIdByNameFn,
		getAuthRoleByIdAttr:       getAuthRoleByIdFn,
		getDefaultRoleIdAttr:      getDefaultRoleIdFn,
		updateAuthRoleAttr:        updateAuthRoleFn,
		deleteAuthRoleAttr:        deleteAuthRoleFn,
		restoreDefaultRolesAttr:   restoreDefaultRolesFn,
		getAllowedPermissionsAttr: getAllowedPermissionsFn,
	}
}

// getAuthRoleProxy acts as a singleton to for the internalProxy.  It also ensures
// that we can still proxy our tests by directly setting internalProxy package variable
func getAuthRoleProxy(clientConfig *platformclientv2.Configuration) *authRoleProxy {
	if internalProxy == nil {
		internalProxy = newAuthRoleProxy(clientConfig)
	}
	return internalProxy
}

// createAuthRole creates a Genesys Cloud auth role
func (p *authRoleProxy) createAuthRole(ctx context.Context, authRole *platformclientv2.Domainorganizationrolecreate) (*platformclientv2.Domainorganizationrole, *platformclientv2.APIResponse, error) {
	return p.createAuthRoleAttr(ctx, p, authRole)
}

// getAuthRole retrieves all Genesys Cloud auth role
func (p *authRoleProxy) getAllAuthRole(ctx context.Context) (*[]platformclientv2.Domainorganizationrole, *platformclientv2.APIResponse, error) {
	return p.getAllAuthRoleAttr(ctx, p)
}

// getAuthRoleIdByName returns a single Genesys Cloud auth role by a name
func (p *authRoleProxy) getAuthRoleIdByName(ctx context.Context, name string) (id string, retryable bool, response *platformclientv2.APIResponse, err error) {
	return p.getAuthRoleIdByNameAttr(ctx, p, name)
}

// getAuthRoleById returns a single Genesys Cloud auth role by Id
func (p *authRoleProxy) getAuthRoleById(ctx context.Context, id string) (authRole *platformclientv2.Domainorganizationrole, response *platformclientv2.APIResponse, err error) {
	if authRole := rc.GetCacheItem(p.authRoleCache, id); authRole != nil {
		return authRole, nil, nil
	}
	return p.getAuthRoleByIdAttr(ctx, p, id)
}

// getAuthRoleById returns a single Genesys Cloud auth role by Id
func (p *authRoleProxy) getDefaultRoleById(ctx context.Context, defaultRoleId string) (roleId string, response *platformclientv2.APIResponse, err error) {
	if authRole := rc.GetCacheItem(p.authRoleCache, defaultRoleId); authRole != nil {
		return *authRole.Id, nil, nil
	}
	return p.getDefaultRoleIdAttr(ctx, p, defaultRoleId)
}

// updateAuthRole updates a Genesys Cloud auth role
func (p *authRoleProxy) updateAuthRole(ctx context.Context, id string, authRole *platformclientv2.Domainorganizationroleupdate) (*platformclientv2.Domainorganizationrole, *platformclientv2.APIResponse, error) {
	return p.updateAuthRoleAttr(ctx, p, id, authRole)
}

// deleteAuthRole deletes a Genesys Cloud auth role by Id
func (p *authRoleProxy) deleteAuthRole(ctx context.Context, id string) (response *platformclientv2.APIResponse, err error) {
	return p.deleteAuthRoleAttr(ctx, p, id)
}

func (p *authRoleProxy) restoreDefaultRoles(ctx context.Context, roles *[]platformclientv2.Domainorganizationrole) (*platformclientv2.APIResponse, error) {
	return p.restoreDefaultRolesAttr(ctx, p, roles)
}

// getAllowedPermissions returns an array of available permissions for a given domain e.g. outbound
func (p *authRoleProxy) getAllowedPermissions(domain string) (*map[string][]platformclientv2.Domainpermission, *platformclientv2.APIResponse, error) {
	return p.getAllowedPermissionsAttr(p, domain)
}

// createAuthRoleFn is an implementation function for creating a Genesys Cloud auth role
func createAuthRoleFn(ctx context.Context, p *authRoleProxy, authRole *platformclientv2.Domainorganizationrolecreate) (*platformclientv2.Domainorganizationrole, *platformclientv2.APIResponse, error) {
	role, apiResponse, err := p.authorizationApi.PostAuthorizationRoles(*authRole)
	if err != nil {
		return nil, apiResponse, fmt.Errorf("Failed to create role %s: %s", *authRole.Name, err)
	}
	return role, apiResponse, nil
}

// getAllAuthRoleFn is the implementation for retrieving all auth role in Genesys Cloud
func getAllAuthRoleFn(ctx context.Context, p *authRoleProxy) (*[]platformclientv2.Domainorganizationrole, *platformclientv2.APIResponse, error) {
	const pageSize = 100
	var allAuthRoles []platformclientv2.Domainorganizationrole

	roles, resp, getErr := p.authorizationApi.GetAuthorizationRoles(pageSize, 1, "", nil, "", "", "", nil, nil, false, nil)
	if getErr != nil {
		return nil, resp, fmt.Errorf("failed to get page of auth roles: %s", getErr)
	}

	if roles.Entities != nil && len(*roles.Entities) > 0 {
		allAuthRoles = append(allAuthRoles, *roles.Entities...)
	}

	for pageNum := 2; pageNum <= *roles.PageCount; pageNum++ {
		roles, resp, getErr := p.authorizationApi.GetAuthorizationRoles(pageSize, pageNum, "", nil, "", "", "", nil, nil, false, nil)
		if getErr != nil {
			return nil, resp, fmt.Errorf("failed to get page of auth roles: %s", getErr)
		}

		if roles.Entities == nil || len(*roles.Entities) == 0 {
			break
		}

		allAuthRoles = append(allAuthRoles, *roles.Entities...)
	}

	//Cache the Auth Role resource into the p.authRoleCache for later use
	for _, authRole := range allAuthRoles {
		rc.SetCache(p.authRoleCache, *authRole.Id, authRole)
	}

	return &allAuthRoles, resp, nil
}

// getAuthRoleIdByNameFn is an implementation of the function to get a Genesys Cloud auth role by name
func getAuthRoleIdByNameFn(ctx context.Context, p *authRoleProxy, name string) (id string, retryable bool, response *platformclientv2.APIResponse, err error) {
	return "", false, nil, nil
}

// getAuthRoleByIdFn is an implementation of the function to get a Genesys Cloud auth role by Id
func getAuthRoleByIdFn(ctx context.Context, p *authRoleProxy, id string) (authRole *platformclientv2.Domainorganizationrole, response *platformclientv2.APIResponse, err error) {
	role, apiResponse, err := p.authorizationApi.GetAuthorizationRole(id, false, nil)
	if err != nil {
		return nil, apiResponse, fmt.Errorf("Failed to retrieve role %s by id: %s", id, err)
	}
	return role, apiResponse, nil
}

func getDefaultRoleIdFn(ctx context.Context, p *authRoleProxy, defaultRoleID string) (roleId string, response *platformclientv2.APIResponse, err error) {
	const pageSize = 1
	const pageNum = 1
	roles, apiResponse, getErr := p.authorizationApi.GetAuthorizationRoles(pageSize, pageNum, "", nil, "", "", "", nil, []string{defaultRoleID}, false, nil)
	if getErr != nil {
		return "", apiResponse, fmt.Errorf("Error requesting default role %s: %s", defaultRoleID, getErr)
	}
	if roles.Entities == nil || len(*roles.Entities) == 0 {
		return "", apiResponse, fmt.Errorf("Default role not found: %s", defaultRoleID)
	}
	return *(*roles.Entities)[0].Id, apiResponse, nil
}

// updateAuthRoleFn is an implementation of the function to update a Genesys Cloud auth role
func updateAuthRoleFn(ctx context.Context, p *authRoleProxy, id string, authRole *platformclientv2.Domainorganizationroleupdate) (*platformclientv2.Domainorganizationrole, *platformclientv2.APIResponse, error) {
	role, apiResponse, err := p.authorizationApi.PutAuthorizationRole(id, *authRole)
	if err != nil {
		return nil, apiResponse, fmt.Errorf("Failed to update role %s: %s", id, err)
	}
	return role, apiResponse, nil
}

// deleteAuthRoleFn is an implementation function for deleting a Genesys Cloud auth role
func deleteAuthRoleFn(ctx context.Context, p *authRoleProxy, id string) (response *platformclientv2.APIResponse, err error) {
	apiResponse, err := p.authorizationApi.DeleteAuthorizationRole(id)
	if err != nil {
		return apiResponse, err
	}
	return apiResponse, nil
}

func restoreDefaultRolesFn(ctx context.Context, p *authRoleProxy, roles *[]platformclientv2.Domainorganizationrole) (*platformclientv2.APIResponse, error) {
	_, apiResponse, err := p.authorizationApi.PutAuthorizationRolesDefault(*roles)
	if err != nil {
		return apiResponse, err
	}
	return apiResponse, nil
}

// getAllowedPermissionsFn is an implementation function for getting all allowed permissions for a domain
func getAllowedPermissionsFn(p *authRoleProxy, domain string) (*map[string][]platformclientv2.Domainpermission, *platformclientv2.APIResponse, error) {
	const pageSize = 100
	allowedPermissions := make(map[string][]platformclientv2.Domainpermission)

	permissions, apiResponse, err := p.authorizationApi.GetAuthorizationPermissions(pageSize, 1, "domain", domain)
	if err != nil {
		return nil, apiResponse, err
	}

	if permissions.Entities == nil || len(*permissions.Entities) == 0 {
		return &allowedPermissions, apiResponse, nil
	}

	for _, permission := range *permissions.Entities {
		for entityType, entityPermissions := range *permission.PermissionMap {
			allowedPermissions[entityType] = entityPermissions
		}
	}

	for pageNum := 2; pageNum <= *permissions.PageCount; pageNum++ {
		permissions, apiResponse, err := p.authorizationApi.GetAuthorizationPermissions(pageSize, pageNum, "domain", domain)
		if err != nil {
			return nil, apiResponse, err
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
	return &allowedPermissions, apiResponse, nil
}
