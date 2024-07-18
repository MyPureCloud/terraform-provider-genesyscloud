package routing_skill_group

import (
	"terraform-provider-genesyscloud/genesyscloud/util"
	lists "terraform-provider-genesyscloud/genesyscloud/util/lists"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

func organizeMemberDivisionIdsForUpdate(schemaIds, apiIds []string) ([]string, []string) {
	toAdd := make([]string, 0)
	toRemove := make([]string, 0)
	// items that are in hcl and not in api-returned list - add
	for _, id := range schemaIds {
		if !lists.ItemInSlice(id, apiIds) {
			toAdd = append(toAdd, id)
		}
	}
	// items that are not in hcl and are in api-returned list - remove
	for _, id := range apiIds {
		if !lists.ItemInSlice(id, schemaIds) {
			toRemove = append(toRemove, id)
		}
	}
	return toAdd, toRemove
}

// Prepare member_division_ids list to avoid an unnecessary plan not empty error
func organizeMemberDivisionIdsForRead(schemaList, apiList []string, divisionId string) []string {
	if !lists.ItemInSlice(divisionId, schemaList) {
		apiList = lists.RemoveStringFromSlice(divisionId, apiList)
	}
	if len(schemaList) == 1 && schemaList[0] == "*" {
		return schemaList
	} else {
		// if hcl & api lists are the same but with different ordering - set with original ordering
		if lists.AreEquivalent(schemaList, apiList) {
			return schemaList
		} else {
			return apiList
		}
	}
}

// Remove the value of division_id, or if this field was left blank; the home division ID
func removeSkillGroupDivisionID(d *schema.ResourceData, list []string) ([]string, diag.Diagnostics) {
	if len(list) == 0 || list == nil {
		return list, nil
	}
	divisionId := d.Get("division_id").(string)
	if divisionId == "" {
		id, diagErr := util.GetHomeDivisionID()
		if diagErr != nil {
			return nil, diagErr
		}
		divisionId = id
	}
	if lists.ItemInSlice(divisionId, list) {
		list = lists.RemoveStringFromSlice(divisionId, list)
	}
	return list, nil
}

func allMemberDivisionsSpecified(schemaSkillGroupMemberDivisionIds []string) bool {
	return lists.ItemInSlice("*", schemaSkillGroupMemberDivisionIds)
}

func BuildHeaderParams(routingAPI *platformclientv2.RoutingApi) map[string]string {
	headerParams := make(map[string]string)

	for key := range routingAPI.Configuration.DefaultHeader {
		headerParams[key] = routingAPI.Configuration.DefaultHeader[key]
	}

	headerParams["Authorization"] = "Bearer " + routingAPI.Configuration.AccessToken
	headerParams["Content-Type"] = "application/json"
	headerParams["Accept"] = "application/json"

	return headerParams
}
