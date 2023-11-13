package genesyscloud

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	lists "terraform-provider-genesyscloud/genesyscloud/util/lists"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v115/platformclientv2"
)

var (
	roleAssignmentResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"role_id": {
				Description: "Role ID.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"division_ids": {
				Description: "Division IDs applied to this resource. If not set, the home division will be used. '*' may be set for all divisions.",
				Type:        schema.TypeSet,
				Optional:    true,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
		},
	}
)

// Get subject grants and filters out inherited grants
func getAssignedGrants(subjectID string, authAPI *platformclientv2.AuthorizationApi) ([]platformclientv2.Authzgrant, *platformclientv2.APIResponse, diag.Diagnostics) {
	var grants []platformclientv2.Authzgrant

	subject, resp, err := authAPI.GetAuthorizationSubject(subjectID)
	if err != nil {
		return nil, resp, diag.Errorf("Failed to get current grants for subject %s: %s", subjectID, err)
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

func readSubjectRoles(d *schema.ResourceData, subjectID string, authAPI *platformclientv2.AuthorizationApi) ([]interface{}, *platformclientv2.APIResponse, diag.Diagnostics) {
	grants, resp, err := getAssignedGrants(subjectID, authAPI)
	if err != nil {
		return nil, resp, err
	}

	roleDivsMap := make(map[string][]interface{})
	for _, grant := range grants {
		if currentDivs, ok := roleDivsMap[*grant.Role.Id]; ok {
			currentDivs = append(currentDivs, *grant.Division.Id)
		} else {
			roleDivsMap[*grant.Role.Id] = []interface{}{*grant.Division.Id}
		}
	}

	var roleList []interface{}
	for roleID, divs := range roleDivsMap {
		role := make(map[string]interface{})
		role["role_id"] = roleID
		role["division_ids"] = divs
		roleList = append(roleList, role)
	}

	// If the role IDs are the same in the schema state and in the response from the GET,
	// re-organize the items to match the ordering in the schema
	rolesFromSchema, ok := d.Get("roles").([]interface{})
	if !ok {
		return roleList, resp, nil
	}

	roleIdsFromSchema := getRoleIdsFromRolesList(rolesFromSchema)
	var roleIdsFromApi []string
	for roleId, _ := range roleDivsMap {
		roleIdsFromApi = append(roleIdsFromApi, roleId)
	}

	if lists.AreEquivalent(roleIdsFromSchema, roleIdsFromApi) {
		// re-organise roleList so that order of items is the same as in the schema
		roleListReordered := make([]interface{}, 0)
		for _, roleId := range roleIdsFromSchema {
			currentRole := make(map[string]interface{}, 0)
			currentRole["role_id"] = roleId
			currentRole["division_ids"] = roleDivsMap[roleId]
			roleListReordered = append(roleListReordered, currentRole)
		}
		roleList = roleListReordered
	}

	return roleList, resp, nil
}

func getRoleIdsFromRolesList(roles []interface{}) []string {
	var roleIds []string
	for _, r := range roles {
		if rMap, ok := r.(map[string]interface{}); ok {
			roleIds = append(roleIds, rMap["role_id"].(string))
		}
	}
	return roleIds
}

func updateSubjectRoles(_ context.Context, d *schema.ResourceData, authAPI *platformclientv2.AuthorizationApi, subjectType string) diag.Diagnostics {
	if !d.HasChange("roles") {
		return nil
	}
	rolesList, ok := d.Get("roles").([]interface{})
	if !ok {
		return nil
	}
	// Get existing roles/divisions
	grants, _, err := getAssignedGrants(d.Id(), authAPI)
	if err != nil {
		return err
	}

	var existingGrants []string
	for _, grant := range grants {
		existingGrants = append(existingGrants, createRoleDivisionPair(*grant.Role.Id, *grant.Division.Id))
	}

	homeDiv, diagErr := getHomeDivisionID()
	if diagErr != nil {
		return diagErr
	}

	var configGrants []string
	for _, configRole := range rolesList {
		roleMap, ok := configRole.(map[string]interface{})
		if !ok {
			continue
		}
		roleID := roleMap["role_id"].(string)

		var divisionIDs []string
		if configDivs, ok := roleMap["division_ids"]; ok {
			divisionIDs = *lists.SetToStringList(configDivs.(*schema.Set))
		}

		if len(divisionIDs) == 0 {
			// No division set. Use the home division
			divisionIDs = []string{homeDiv}
		}

		for _, divID := range divisionIDs {
			configGrants = append(configGrants, createRoleDivisionPair(roleID, divID))
		}
	}

	grantsToRemove := lists.SliceDifference(existingGrants, configGrants)
	if len(grantsToRemove) > 0 {
		// It's possible for a role or division to be removed before this update is processed,
		// and the bulk remove API returns failure if any roles/divisions no longer exist.
		// Work around by removing all grants individually and ignore 404s.
		sdkGrantsToRemove := roleDivPairsToGrants(grantsToRemove)
		for _, grant := range *sdkGrantsToRemove.Grants {
			resp, err := authAPI.DeleteAuthorizationSubjectDivisionRole(d.Id(), *grant.DivisionId, *grant.RoleId)
			if err != nil {
				if resp == nil || resp.StatusCode != 404 {
					return diag.Errorf("Failed to remove role grants for subject %s: %s", d.Id(), err)
				}
			}
		}
	}

	grantsToAdd := lists.SliceDifference(configGrants, existingGrants)
	if len(grantsToAdd) > 0 {
		// In some cases new roles or divisions have not yet been added to the auth service cache causing 404s that should be retried.
		diagErr = RetryWhen(IsStatus404, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
			resp, err := authAPI.PostAuthorizationSubjectBulkadd(d.Id(), roleDivPairsToGrants(grantsToAdd), subjectType)
			if err != nil {
				return resp, diag.Errorf("Failed to add role grants for subject %s: %s", d.Id(), err)
			}
			return nil, nil
		})
		if diagErr != nil {
			return diagErr
		}
	}
	return nil
}

func createRoleDivisionPair(roleID string, divisionID string) string {
	return roleID + ":" + divisionID
}

func roleDivPairsToGrants(grantPairs []string) platformclientv2.Roledivisiongrants {
	grants := make([]platformclientv2.Roledivisionpair, len(grantPairs))
	for i, pair := range grantPairs {
		roleDiv := strings.Split(pair, ":")
		grants[i] = platformclientv2.Roledivisionpair{
			RoleId:     &roleDiv[0],
			DivisionId: &roleDiv[1],
		}
	}
	return platformclientv2.Roledivisiongrants{
		Grants: &grants,
	}
}

// Testing common
func GenerateResourceRoles(skillID string, divisionIds ...string) string {
	var divAttr string
	if len(divisionIds) > 0 {
		divAttr = "division_ids = [" + strings.Join(divisionIds, ",") + "]"
	}
	return fmt.Sprintf(`roles {
		role_id = %s
		%s
	}
	`, skillID, divAttr)
}

func validateResourceRole(resourceName string, roleResourceName string, divisions ...string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		resourceState, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Failed to find %s in state", resourceName)
		}
		resourceID := resourceState.Primary.ID

		roleResource, ok := state.RootModule().Resources[roleResourceName]
		if !ok {
			return fmt.Errorf("Failed to find role %s in state", roleResourceName)
		}
		roleID := roleResource.Primary.ID

		if len(divisions) == 0 {
			// If no division specified, role should be in the home division
			homeDiv, err := getHomeDivisionID()
			if err != nil {
				return fmt.Errorf("Failed to query home div: %v", err)
			}
			divisions = []string{homeDiv}
		} else if divisions[0] != "*" {
			// Get the division IDs from state
			divisionIDs := make([]string, len(divisions))
			for i, divResourceName := range divisions {
				divResource, ok := state.RootModule().Resources[divResourceName]
				if !ok {
					return fmt.Errorf("Failed to find %s in state", divResourceName)
				}
				divisionIDs[i] = divResource.Primary.ID
			}
			divisions = divisionIDs
		}

		resourceAttrs := resourceState.Primary.Attributes
		numRolesAttr, _ := resourceAttrs["roles.#"]
		numRoles, _ := strconv.Atoi(numRolesAttr)
		for i := 0; i < numRoles; i++ {
			if resourceAttrs["roles."+strconv.Itoa(i)+".role_id"] == roleID {
				numDivsAttr, _ := resourceAttrs["roles."+strconv.Itoa(i)+".division_ids.#"]
				numDivs, _ := strconv.Atoi(numDivsAttr)
				stateDivs := make([]string, numDivs)
				for j := 0; j < numDivs; j++ {
					stateDivs[j] = resourceAttrs["roles."+strconv.Itoa(i)+".division_ids."+strconv.Itoa(j)]
				}

				extraDivs := lists.SliceDifference(stateDivs, divisions)
				if len(extraDivs) > 0 {
					return fmt.Errorf("Unexpected divisions found for role %s in state: %v", roleID, extraDivs)
				}

				missingDivs := lists.SliceDifference(divisions, stateDivs)
				if len(missingDivs) > 0 {
					return fmt.Errorf("Missing expected divisions for role %s in state: %v", roleID, missingDivs)
				}

				// Found expected role and divisions
				return nil
			}
		}
		return fmt.Errorf("Missing expected role for resource %s in state: %s", resourceID, roleID)
	}
}
