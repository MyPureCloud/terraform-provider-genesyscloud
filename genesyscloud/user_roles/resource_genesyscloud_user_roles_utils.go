package user_roles

import (
	"fmt"
	"strings"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/lists"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v195/platformclientv2"
)

func flattenSubjectRoles(d *schema.ResourceData, p *userRolesProxy) (*schema.Set, *platformclientv2.APIResponse, error) {
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
		// Guard against malformed/partial API responses where Role, Division, or their
		// Ids may be nil. Skipping such grants prevents a nil pointer panic (which would
		// crash the whole provider plugin) and lets the read complete gracefully.
		if grant.Role == nil || grant.Role.Id == nil || grant.Division == nil || grant.Division.Id == nil {
			continue
		}
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
	grants := make([]platformclientv2.Roledivisionpair, 0, len(grantPairs))
	for _, pair := range grantPairs {
		roleDiv := strings.SplitN(pair, ":", 2)
		// A valid pair is "roleID:divisionID". Skip anything malformed to avoid an
		// index-out-of-range panic on roleDiv[1].
		if len(roleDiv) != 2 {
			continue
		}
		roleID := roleDiv[0]
		divID := roleDiv[1]
		grants = append(grants, platformclientv2.Roledivisionpair{
			RoleId:     &roleID,
			DivisionId: &divID,
		})
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
		if !ok {
			continue
		}
		// find the role in question (guard the type assertion to avoid a panic)
		roleIDVal, ok := roleMap["role_id"].(string)
		if !ok || roleIDVal != roleId {
			continue
		}
		divs, ok := roleMap["division_ids"].(*schema.Set)
		if !ok || divs == nil {
			continue
		}
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
		// Guard against nil Role/Division to avoid a panic on malformed API responses.
		if grant.Role == nil || grant.Role.Id == nil || grant.Division == nil || grant.Division.Id == nil {
			continue
		}
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
