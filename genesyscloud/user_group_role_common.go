package genesyscloud

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"terraform-provider-genesyscloud/genesyscloud/util/lists"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v119/platformclientv2"
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

func readSubjectRoles(d *schema.ResourceData, authAPI *platformclientv2.AuthorizationApi) (*schema.Set, *platformclientv2.APIResponse, diag.Diagnostics) {
	grants, resp, err := getAssignedGrants(d.Id(), authAPI)
	if err != nil {
		return nil, resp, err
	}

	homeDivId, err := getHomeDivisionID()
	if err != nil {
		return nil, nil, err
	}

	roleDivsMap := make(map[string]*schema.Set)
	for _, grant := range grants {
		if currentDivs, ok := roleDivsMap[*grant.Role.Id]; ok {
			currentDivs.Add(*grant.Division.Id)
		} else {
			roleDivsMap[*grant.Role.Id] = schema.NewSet(schema.HashString, []interface{}{*grant.Division.Id})
		}
	}

	roleSet := schema.NewSet(schema.HashResource(roleAssignmentResource), []interface{}{})
	for roleID, divs := range roleDivsMap {
		role := make(map[string]interface{})
		role["role_id"] = roleID
		role["division_ids"] = addDivisionIdsSetToRole(d, divs, roleID, homeDivId)
		roleSet.Add(role)
	}
	return roleSet, resp, nil
}

func updateSubjectRoles(_ context.Context, d *schema.ResourceData, authAPI *platformclientv2.AuthorizationApi, subjectType string) diag.Diagnostics {
	if !d.HasChange("roles") {
		return nil
	}
	rolesConfig := d.Get("roles")
	if rolesConfig == nil {
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
	rolesList := rolesConfig.(*schema.Set).List()
	for _, configRole := range rolesList {
		roleMap := configRole.(map[string]interface{})
		roleID := roleMap["role_id"].(string)

		var divisionIDs []string
		if configDivs, ok := roleMap["division_ids"].(*schema.Set); ok {
			divisionIDs = *lists.SetToStringList(configDivs)
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

// If the user provides no division ids, we add the home division to that set for them. Previously, we had division_ids: Computed
// to avoid errors. This only caused more problems with testing because division_ids would always cause a plan not empty error,
// going from ["<home division ID>"] to [(known after apply)]
// Solution: Remove the computed attribute from schema and use the function below to set the division_ids field on read.
// addDivisionIdsSetToRole - checks if the home division was already included in the division_ids set in the local config resource schema
// If yes, set it as such on the read. If not, do not set it back in on the read.
func addDivisionIdsSetToRole(d *schema.ResourceData, divIdsFromApi *schema.Set, roleId, homeDivId string) *schema.Set {
	rolesSet, ok := d.Get("roles").(*schema.Set)
	if !ok {
		return divIdsFromApi
	}
	rolesMaps := rolesSet.List()
	for _, role := range rolesMaps {
		roleMap, ok := role.(map[string]interface{})
		// find the role in question
		if !ok || roleMap["role_id"].(string) != roleId {
			continue
		}
		divs := roleMap["division_ids"].(*schema.Set)
		for _, div := range divs.List() {
			// home division id was included in original config -> use division_ids read from API
			if div.(string) == homeDivId {
				return divIdsFromApi
			}
		}
		// home division ID was not included in original config for this role -> keep it out
		divIdsFromApi.Remove(homeDivId)
		break
	}
	return divIdsFromApi
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

		homeDivID, err := getHomeDivisionID()
		if err != nil {
			return fmt.Errorf("failed to retrieve home division ID: %v", err)
		}

		if len(divisions) > 0 && divisions[0] != "*" {
			// Get the division IDs from state
			divisionIDs := make([]string, len(divisions))
			for i, divResourceName := range divisions {
				divResource, ok := state.RootModule().Resources[divResourceName]
				if !ok {
					return fmt.Errorf("failed to find %s in state", divResourceName)
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
					if len(extraDivs) > 1 || extraDivs[0] != homeDivID {
						return fmt.Errorf("unexpected divisions found for role %s in state: %v", roleID, extraDivs)
					}
				}

				missingDivs := lists.SliceDifference(divisions, stateDivs)
				if len(missingDivs) > 0 {
					return fmt.Errorf("missing expected divisions for role %s in state: %v", roleID, missingDivs)
				}

				// Found expected role and divisions
				return nil
			}
		}
		return fmt.Errorf("Missing expected role for resource %s in state: %s", resourceID, roleID)
	}
}
