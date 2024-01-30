package auth_role

import (
	"context"
	"fmt"
	"github.com/mypurecloud/platform-client-sdk-go/v119/platformclientv2"
)

/*
The genesyscloud_auth_role_proxy.go file contains the proxy structures and methods that interact
with the Genesys Cloud SDK. We use composition here for each function on the proxy so individual functions can be stubbed
out during testing.
*/

// internalProxy holds a proxy instance that can be used throughout the package
var internalProxy *authRoleProxy

// Type definitions for each func on our proxy so we can easily mock them out later
type createAuthRoleFunc func(ctx context.Context, p *authRoleProxy, domainOrganizationRole *platformclientv2.Domainorganizationrolecreate) (*platformclientv2.Domainorganizationrole, error)
type getAllAuthRoleFunc func(ctx context.Context, p *authRoleProxy) (*[]platformclientv2.Domainorganizationrole, error)
type getAuthRoleIdByNameFunc func(ctx context.Context, p *authRoleProxy, name string) (id string, retryable bool, err error)
type getAuthRoleByIdFunc func(ctx context.Context, p *authRoleProxy, id string) (domainOrganizationRole *platformclientv2.Domainorganizationrole, responseCode int, err error)
type updateAuthRoleFunc func(ctx context.Context, p *authRoleProxy, id string, domainOrganizationRole *platformclientv2.Domainorganizationroleupdate) (*platformclientv2.Domainorganizationrole, error)
type deleteAuthRoleFunc func(ctx context.Context, p *authRoleProxy, id string) (responseCode int, err error)
type restoreDefaultRolesFunc func(ctx context.Context, p *authRoleProxy, roles *[]platformclientv2.Domainorganizationrole) error

// authRoleProxy contains all of the methods that call genesys cloud APIs.
type authRoleProxy struct {
	clientConfig            *platformclientv2.Configuration
	authorizationApi        *platformclientv2.AuthorizationApi
	createAuthRoleAttr      createAuthRoleFunc
	getAllAuthRoleAttr      getAllAuthRoleFunc
	getAuthRoleIdByNameAttr getAuthRoleIdByNameFunc
	getAuthRoleByIdAttr     getAuthRoleByIdFunc
	updateAuthRoleAttr      updateAuthRoleFunc
	deleteAuthRoleAttr      deleteAuthRoleFunc
	restoreDefaultRolesAttr restoreDefaultRolesFunc
}

// newAuthRoleProxy initializes the auth role proxy with all of the data needed to communicate with Genesys Cloud
func newAuthRoleProxy(clientConfig *platformclientv2.Configuration) *authRoleProxy {
	api := platformclientv2.NewAuthorizationApiWithConfig(clientConfig)
	return &authRoleProxy{
		clientConfig:            clientConfig,
		authorizationApi:        api,
		createAuthRoleAttr:      createAuthRoleFn,
		getAllAuthRoleAttr:      getAllAuthRoleFn,
		getAuthRoleIdByNameAttr: getAuthRoleIdByNameFn,
		getAuthRoleByIdAttr:     getAuthRoleByIdFn,
		updateAuthRoleAttr:      updateAuthRoleFn,
		deleteAuthRoleAttr:      deleteAuthRoleFn,
		restoreDefaultRolesAttr: restoreDefaultRolesFn,
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
func (p *authRoleProxy) createAuthRole(ctx context.Context, authRole *platformclientv2.Domainorganizationrolecreate) (*platformclientv2.Domainorganizationrole, error) {
	return p.createAuthRoleAttr(ctx, p, authRole)
}

// getAuthRole retrieves all Genesys Cloud auth role
func (p *authRoleProxy) getAllAuthRole(ctx context.Context) (*[]platformclientv2.Domainorganizationrole, error) {
	return p.getAllAuthRoleAttr(ctx, p)
}

// getAuthRoleIdByName returns a single Genesys Cloud auth role by a name
func (p *authRoleProxy) getAuthRoleIdByName(ctx context.Context, name string) (id string, retryable bool, err error) {
	return p.getAuthRoleIdByNameAttr(ctx, p, name)
}

// getAuthRoleById returns a single Genesys Cloud auth role by Id
func (p *authRoleProxy) getAuthRoleById(ctx context.Context, id string) (authRole *platformclientv2.Domainorganizationrole, statusCode int, err error) {
	return p.getAuthRoleByIdAttr(ctx, p, id)
}

// updateAuthRole updates a Genesys Cloud auth role
func (p *authRoleProxy) updateAuthRole(ctx context.Context, id string, authRole *platformclientv2.Domainorganizationroleupdate) (*platformclientv2.Domainorganizationrole, error) {
	return p.updateAuthRoleAttr(ctx, p, id, authRole)
}

// deleteAuthRole deletes a Genesys Cloud auth role by Id
func (p *authRoleProxy) deleteAuthRole(ctx context.Context, id string) (statusCode int, err error) {
	return p.deleteAuthRoleAttr(ctx, p, id)
}

func (p *authRoleProxy) restoreDefaultRoles(ctx context.Context, roles *[]platformclientv2.Domainorganizationrole) error {
	return p.restoreDefaultRolesAttr(ctx, p, roles)
}

// createAuthRoleFn is an implementation function for creating a Genesys Cloud auth role
func createAuthRoleFn(ctx context.Context, p *authRoleProxy, authRole *platformclientv2.Domainorganizationrolecreate) (*platformclientv2.Domainorganizationrole, error) {
	role, _, err := p.authorizationApi.PostAuthorizationRoles(*authRole)
	if err != nil {
		return nil, fmt.Errorf("Failed to create role %s: %s", *authRole.Name, err)
	}

	return role, nil
}

// getAllAuthRoleFn is the implementation for retrieving all auth role in Genesys Cloud
func getAllAuthRoleFn(ctx context.Context, p *authRoleProxy) (*[]platformclientv2.Domainorganizationrole, error) {
	return nil, nil
}

// getAuthRoleIdByNameFn is an implementation of the function to get a Genesys Cloud auth role by name
func getAuthRoleIdByNameFn(ctx context.Context, p *authRoleProxy, name string) (id string, retryable bool, err error) {
	return "", false, nil
}

// getAuthRoleByIdFn is an implementation of the function to get a Genesys Cloud auth role by Id
func getAuthRoleByIdFn(ctx context.Context, p *authRoleProxy, id string) (authRole *platformclientv2.Domainorganizationrole, statusCode int, err error) {
	role, resp, err := p.authorizationApi.GetAuthorizationRole(id, false, nil)
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("Failed to retrieve role %s by id: %s", id, err)
	}
	return role, resp.StatusCode, nil
}

// updateAuthRoleFn is an implementation of the function to update a Genesys Cloud auth role
func updateAuthRoleFn(ctx context.Context, p *authRoleProxy, id string, authRole *platformclientv2.Domainorganizationroleupdate) (*platformclientv2.Domainorganizationrole, error) {
	role, _, err := p.authorizationApi.PutAuthorizationRole(id, *authRole)
	if err != nil {
		return nil, fmt.Errorf("Failed to update role %s: %s", id, err)
	}

	return role, nil
}

// deleteAuthRoleFn is an implementation function for deleting a Genesys Cloud auth role
func deleteAuthRoleFn(ctx context.Context, p *authRoleProxy, id string) (statusCode int, err error) {
	return 0, nil
}

func restoreDefaultRolesFn(ctx context.Context, p *authRoleProxy, roles *[]platformclientv2.Domainorganizationrole) error {
	_, _, err := p.authorizationApi.PutAuthorizationRolesDefault(*roles)
	if err != nil {
		return err
	}

	return nil
}
