package routing_skill_group

import (
	"context"
	"fmt"
	"log"
	"terraform-provider-genesyscloud/genesyscloud/util"
	lists "terraform-provider-genesyscloud/genesyscloud/util/lists"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Prepare member_division_ids list to avoid an unnecessary plan not empty error
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

// Assign the member division ids to the skill group  
func assignMemberDivisionIds(ctx context.Context, d *schema.ResourceData, meta interface{}, create bool) diag.Diagnostics {
	if create {
		log.Printf("Creating Member Divisions for skill group %s", d.Id())
	} else {
		log.Printf("Updating Member Divisions for skill group %s", d.Id())
	}

	divIds, diagErr := readSkillGroupMemberDivisions(ctx, d, meta)
	if diagErr != nil {
		return diagErr
	}

	diagErr = createRoutingSkillGroupsMemberDivisions(ctx, d, meta, divIds, create)
	if diagErr != nil {
		return diagErr
	}

	return nil
}

func createListsForSkillgroupsMembersDivisions(schemaMemberDivisionIds []string, skillGroupDivisionIds []string, create bool, meta interface{}) ([]string, []string, diag.Diagnostics) {
	toAdd := make([]string, 0)
	toRemove := make([]string, 0)

	if allMemberDivisionsSpecified(schemaMemberDivisionIds) {
		if len(schemaMemberDivisionIds) > 1 {
			return nil, nil, util.BuildDiagnosticError(resourceName, fmt.Sprintf(`member_division_ids should not contain more than one item when the value of an item is "*"`), fmt.Errorf(`member_division_ids should not contain more than one item when the value of an item is "*"`))
		}
		toAdd, err := getAllAuthDivisionIds(meta)
		return toAdd, nil, err
	}

	if len(schemaMemberDivisionIds) > 0 {
		if create {
			return schemaMemberDivisionIds, nil, nil
		}
		toAdd, toRemove = organizeMemberDivisionIdsForUpdate(schemaMemberDivisionIds, skillGroupDivisionIds)
		return toAdd, toRemove, nil
	}

	// Empty array - remove all
	toRemove = append(toRemove, skillGroupDivisionIds...)

	return nil, toRemove, nil
}

func allMemberDivisionsSpecified(schemaSkillGroupMemberDivisionIds []string) bool {
	return lists.ItemInSlice("*", schemaSkillGroupMemberDivisionIds)
}
