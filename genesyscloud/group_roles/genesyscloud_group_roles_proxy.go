package group_roles

import (
	"context"
	"fmt"
	"terraform-provider-genesyscloud/genesyscloud/util"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

var internalProxy *groupRolesProxy

type getGroupRolesByIdFunc func(ctx context.Context, p *groupRolesProxy, roleId string) (*[]platformclientv2.Authzgrant, *platformclientv2.APIResponse, error)
type updateGroupRolesFunc func(ctx context.Context, p *groupRolesProxy, roleId string, rolesConfig *schema.Set, subjectType string) (*platformclientv2.APIResponse, error)

type groupRolesProxy struct {
	clientConfig          *platformclientv2.Configuration
	authorizationApi      *platformclientv2.AuthorizationApi
	getGroupRolesByIdAttr getGroupRolesByIdFunc
	updateGroupRolesAttr  updateGroupRolesFunc
}

func newGroupRolesProxy(clientConfig *platformclientv2.Configuration) *groupRolesProxy {
	api := platformclientv2.NewAuthorizationApiWithConfig(clientConfig)
	return &groupRolesProxy{
		clientConfig:          clientConfig,
		authorizationApi:      api,
		getGroupRolesByIdAttr: getGroupRolesByIdFn,
		updateGroupRolesAttr:  updateGroupRolesFn,
	}
}

func getGroupRolesProxy(clientConfig *platformclientv2.Configuration) *groupRolesProxy {
	if internalProxy == nil {
		internalProxy = newGroupRolesProxy(clientConfig)
	}
	return internalProxy
}

func (p *groupRolesProxy) getGroupRolesById(ctx context.Context, roleId string) (*[]platformclientv2.Authzgrant, *platformclientv2.APIResponse, error) {
	return p.getGroupRolesByIdAttr(ctx, p, roleId)
}
func (p *groupRolesProxy) updateGroupRoles(ctx context.Context, roleID string, rolesConfig *schema.Set, subjectType string) (*platformclientv2.APIResponse, error) {
	return p.updateGroupRolesAttr(ctx, p, roleID, rolesConfig, subjectType)
}

func getGroupRolesByIdFn(_ context.Context, p *groupRolesProxy, roleId string) (*[]platformclientv2.Authzgrant, *platformclientv2.APIResponse, error) {
	var grants []platformclientv2.Authzgrant
	subject, resp, err := p.authorizationApi.GetAuthorizationSubject(roleId, true)
	if err != nil {
		return nil, resp, fmt.Errorf("failed to get current grants for subject %s: %s", roleId, err)
	}

	if subject != nil && subject.Grants != nil {
		for _, grant := range *subject.Grants {
			if grant.SubjectId != nil && *grant.SubjectId == roleId {
				grants = append(grants, grant)
			}
		}
	}
	if err != nil {
		return nil, resp, err
	}
	return &grants, resp, nil
}

func updateGroupRolesFn(_ context.Context, p *groupRolesProxy, roleId string, rolesConfig *schema.Set, subjectType string) (*platformclientv2.APIResponse, error) {
	// Get existing roles/divisions
	subject, resp, err := p.authorizationApi.GetAuthorizationSubject(roleId, true)
	grants, resp, err := getAssignedGrants(*subject.Id, p)

	existingGrants, configGrants, _ := getExistingAndConfigGrants(grants, rolesConfig)
	if err != nil {
		return resp, fmt.Errorf("failed to get current grants for subject %s: %s", roleId, err)
	}

	if subject != nil && subject.Grants != nil {
		for _, grant := range *subject.Grants {
			if grant.SubjectId != nil && *grant.SubjectId == roleId {
				grants = append(grants, grant)
			}
		}
	}

	grantsToRemove, grantsToAdd := getGrantsToAddAndRemove(existingGrants, configGrants)

	if len(grantsToRemove) > 0 {
		// It's possible for a role or division to be removed before this update is processed,
		// and the bulk remove API returns failure if any roles/divisions no longer exist.
		// Work around by removing all grants individually and ignore 404s.
		sdkGrantsToRemove := roleDivPairsToGrants(grantsToRemove)
		for _, grant := range *sdkGrantsToRemove.Grants {
			resp, err := p.authorizationApi.DeleteAuthorizationSubjectDivisionRole(roleId, *grant.DivisionId, *grant.RoleId)
			if err != nil {
				if resp == nil || resp.StatusCode != 404 {
					return resp, fmt.Errorf("failed to remove role grants for subject %s: %s", roleId, err)
				}
			}
		}
	}
	if len(grantsToAdd) > 0 {
		// In some cases new roles or divisions have not yet been added to the auth service cache causing 404s that should be retried.
		diagErr := util.RetryWhen(util.IsStatus404, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
			resp, err := p.authorizationApi.PostAuthorizationSubjectBulkadd(roleId, roleDivPairsToGrants(grantsToAdd), subjectType)
			if err != nil {
				return resp, util.BuildAPIDiagnosticError(resourceName, fmt.Sprintf("failed to add role grants for subject %s: %s", roleId, err), resp)
			}
			return nil, nil
		})
		if diagErr != nil {
			return resp, fmt.Errorf("error in adding grants: %v", diagErr)
		}
	}
	return resp, nil
}

func getAssignedGrants(subjectID string, p *groupRolesProxy) ([]platformclientv2.Authzgrant, *platformclientv2.APIResponse, error) {
	var grants []platformclientv2.Authzgrant
	subject, resp, err := p.authorizationApi.GetAuthorizationSubject(subjectID, true)

	if err != nil {
		return nil, resp, fmt.Errorf("failed to get current grants for subject %s: %s", subjectID, err)
	}
	if subject != nil && subject.Grants != nil {
		for _, grant := range *subject.Grants {
			if grant.SubjectId != nil && *grant.SubjectId == subjectID {
				grants = append(grants, grant)
			}
		}
	}
	return grants, resp, nil
}
