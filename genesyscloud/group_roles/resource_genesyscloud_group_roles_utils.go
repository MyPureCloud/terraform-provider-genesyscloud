package group_roles

import (
	"fmt"
	"strings"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"terraform-provider-genesyscloud/genesyscloud/util/lists"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

func flattenSubjectRoles(d *schema.ResourceData, p *groupRolesProxy) (*schema.Set, *platformclientv2.APIResponse, error) {
	grants, resp, diagErr := getAssignedGrants(d.Id(), p)
	if diagErr != nil {
		return nil, resp, fmt.Errorf("error getting assigned grants %s", diagErr)
	}

	homeDivId, err := util.GetHomeDivisionID()
	if err != nil {
		return nil, nil, fmt.Errorf("error getting home division id %v", err)
	}

	roleDivsMap := make(map[string]*schema.Set)
	for _, grant := range grants {
		if currentDivs, ok := roleDivsMap[*grant.Role.Id]; ok {
			currentDivs.Add(*grant.Division.Id)
		} else {
			roleDivsMap[*grant.Role.Id] = schema.NewSet(schema.HashString, []interface{}{*grant.Division.Id})
		}
	}

	roleSet := schema.NewSet(schema.HashResource(RoleAssignmentResource), []interface{}{})
	for roleID, divs := range roleDivsMap {
		role := make(map[string]interface{})
		role["role_id"] = roleID
		role["division_ids"] = addDivisionIdsSetToRole(d, divs, roleID, homeDivId)
		roleSet.Add(role)
	}
	return roleSet, resp, nil
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

// getExistingAndConfigGrants is used to generate the existing and config grants for the resource
func getExistingAndConfigGrants(grants []platformclientv2.Authzgrant, rolesConfig *schema.Set) ([]string, []string, error) {
	rolesList := rolesConfig.List()
	var existingGrants []string

	for _, grant := range grants {
		existingGrants = append(existingGrants, createRoleDivisionPair(*grant.Role.Id, *grant.Division.Id))
	}

	var configGrants []string
	homeDiv, err := util.GetHomeDivisionID()

	if err != nil {
		return nil, nil, fmt.Errorf("failed to get home division ID %v", err)
	}

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
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load grants: %v", err)
	}

	return existingGrants, configGrants, nil
}

func getGrantsToAddAndRemove(existingGrants []string, configGrants []string) ([]string, []string) {
	grantsToRemove := lists.SliceDifference(existingGrants, configGrants)
	grantsToAdd := lists.SliceDifference(configGrants, existingGrants)
	return grantsToRemove, grantsToAdd
}

func createRoleDivisionPair(roleID string, divisionID string) string {
	return roleID + ":" + divisionID
}
