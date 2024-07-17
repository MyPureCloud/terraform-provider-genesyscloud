package routing_skill_group

import (
	"encoding/json"
	"fmt"
	"strings"
	"terraform-provider-genesyscloud/genesyscloud/util"
	lists "terraform-provider-genesyscloud/genesyscloud/util/lists"
	"terraform-provider-genesyscloud/genesyscloud/util/resourcedata"

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

// getRoutingSkillGroupsFromResourceData maps data from schema ResourceData object to a platformclientv2.Skillgroupwithmemberdivisions
func getRoutingSkillGroupsFromResourceData(d *schema.ResourceData) platformclientv2.Skillgroupwithmemberdivisions {
	return platformclientv2.Skillgroupwithmemberdivisions{
		Name:        platformclientv2.String(d.Get("name").(string)),
		Division:    &platformclientv2.Writabledivision{Id: platformclientv2.String(d.Get("division_id").(string))},
		Description: platformclientv2.String(d.Get("description").(string)),
		//MemberCount:     platformclientv2.Int(d.Get("member_count").(int)),
		//Status:          platformclientv2.String(d.Get("status").(string)),
		// SkillConditions: buildSkillGroupConditions(d.Get("skill_conditions").([]interface{})),
		// TODO: Handle member_divisions property

	}
}

// buildSkillGroupRoutingConditions maps an []interface{} into a Genesys Cloud *[]platformclientv2.Skillgrouproutingcondition
func buildSkillGroupRoutingConditions(skillGroupRoutingConditions []interface{}) *[]platformclientv2.Skillgrouproutingcondition {
	skillGroupRoutingConditionsSlice := make([]platformclientv2.Skillgrouproutingcondition, 0)
	for _, skillGroupRoutingCondition := range skillGroupRoutingConditions {
		var sdkSkillGroupRoutingCondition platformclientv2.Skillgrouproutingcondition
		skillGroupRoutingConditionsMap, ok := skillGroupRoutingCondition.(map[string]interface{})
		if !ok {
			continue
		}

		resourcedata.BuildSDKStringValueIfNotNil(&sdkSkillGroupRoutingCondition.RoutingSkill, skillGroupRoutingConditionsMap, "routing_skill")
		resourcedata.BuildSDKStringValueIfNotNil(&sdkSkillGroupRoutingCondition.Comparator, skillGroupRoutingConditionsMap, "comparator")
		sdkSkillGroupRoutingCondition.Proficiency = platformclientv2.Int(skillGroupRoutingConditionsMap["proficiency"].(int))
		resourcedata.BuildSDKInterfaceArrayValueIfNotNil(&sdkSkillGroupRoutingCondition.ChildConditions, skillGroupRoutingConditionsMap, "child_conditions", buildSkillGroupConditions)

		skillGroupRoutingConditionsSlice = append(skillGroupRoutingConditionsSlice, sdkSkillGroupRoutingCondition)
	}

	return &skillGroupRoutingConditionsSlice
}

// buildSkillGroupLanguageConditions maps an []interface{} into a Genesys Cloud *[]platformclientv2.Skillgrouplanguagecondition
func buildSkillGroupLanguageConditions(skillGroupLanguageConditions []interface{}) *[]platformclientv2.Skillgrouplanguagecondition {
	skillGroupLanguageConditionsSlice := make([]platformclientv2.Skillgrouplanguagecondition, 0)
	for _, skillGroupLanguageCondition := range skillGroupLanguageConditions {
		var sdkSkillGroupLanguageCondition platformclientv2.Skillgrouplanguagecondition
		skillGroupLanguageConditionsMap, ok := skillGroupLanguageCondition.(map[string]interface{})
		if !ok {
			continue
		}

		resourcedata.BuildSDKStringValueIfNotNil(&sdkSkillGroupLanguageCondition.LanguageSkill, skillGroupLanguageConditionsMap, "language_skill")
		resourcedata.BuildSDKStringValueIfNotNil(&sdkSkillGroupLanguageCondition.Comparator, skillGroupLanguageConditionsMap, "comparator")
		sdkSkillGroupLanguageCondition.Proficiency = platformclientv2.Int(skillGroupLanguageConditionsMap["proficiency"].(int))
		resourcedata.BuildSDKInterfaceArrayValueIfNotNil(&sdkSkillGroupLanguageCondition.ChildConditions, skillGroupLanguageConditionsMap, "child_conditions", buildSkillGroupConditions)

		skillGroupLanguageConditionsSlice = append(skillGroupLanguageConditionsSlice, sdkSkillGroupLanguageCondition)
	}

	return &skillGroupLanguageConditionsSlice
}

// buildSkillGroupConditions maps an []interface{} into a Genesys Cloud *[]platformclientv2.Skillgroupcondition
func buildSkillGroupConditions(skillGroupConditions []interface{}) *[]platformclientv2.Skillgroupcondition {
	skillGroupConditionsSlice := make([]platformclientv2.Skillgroupcondition, 0)
	for _, skillGroupCondition := range skillGroupConditions {
		var sdkSkillGroupCondition platformclientv2.Skillgroupcondition
		skillGroupConditionsMap, ok := skillGroupCondition.(map[string]interface{})
		if !ok {
			continue
		}

		resourcedata.BuildSDKInterfaceArrayValueIfNotNil(&sdkSkillGroupCondition.RoutingSkillConditions, skillGroupConditionsMap, "routing_skill_conditions", buildSkillGroupRoutingConditions)
		resourcedata.BuildSDKInterfaceArrayValueIfNotNil(&sdkSkillGroupCondition.LanguageSkillConditions, skillGroupConditionsMap, "language_skill_conditions", buildSkillGroupLanguageConditions)
		resourcedata.BuildSDKStringValueIfNotNil(&sdkSkillGroupCondition.Operation, skillGroupConditionsMap, "operation")

		skillGroupConditionsSlice = append(skillGroupConditionsSlice, sdkSkillGroupCondition)
	}

	return &skillGroupConditionsSlice
}

// flattenSkillGroupRoutingConditions maps a Genesys Cloud *[]platformclientv2.Skillgrouproutingcondition into a []interface{}
func flattenSkillGroupRoutingConditions(skillGroupRoutingConditions *[]platformclientv2.Skillgrouproutingcondition) []interface{} {
	if len(*skillGroupRoutingConditions) == 0 {
		return nil
	}

	var skillGroupRoutingConditionList []interface{}
	for _, skillGroupRoutingCondition := range *skillGroupRoutingConditions {
		skillGroupRoutingConditionMap := make(map[string]interface{})

		resourcedata.SetMapValueIfNotNil(skillGroupRoutingConditionMap, "routing_skill", skillGroupRoutingCondition.RoutingSkill)
		resourcedata.SetMapValueIfNotNil(skillGroupRoutingConditionMap, "comparator", skillGroupRoutingCondition.Comparator)
		resourcedata.SetMapValueIfNotNil(skillGroupRoutingConditionMap, "proficiency", skillGroupRoutingCondition.Proficiency)
		resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(skillGroupRoutingConditionMap, "child_conditions", skillGroupRoutingCondition.ChildConditions, flattenSkillGroupConditions)

		skillGroupRoutingConditionList = append(skillGroupRoutingConditionList, skillGroupRoutingConditionMap)
	}

	return skillGroupRoutingConditionList
}

// flattenSkillGroupLanguageConditions maps a Genesys Cloud *[]platformclientv2.Skillgrouplanguagecondition into a []interface{}
func flattenSkillGroupLanguageConditions(skillGroupLanguageConditions *[]platformclientv2.Skillgrouplanguagecondition) []interface{} {
	if len(*skillGroupLanguageConditions) == 0 {
		return nil
	}

	var skillGroupLanguageConditionList []interface{}
	for _, skillGroupLanguageCondition := range *skillGroupLanguageConditions {
		skillGroupLanguageConditionMap := make(map[string]interface{})

		resourcedata.SetMapValueIfNotNil(skillGroupLanguageConditionMap, "language_skill", skillGroupLanguageCondition.LanguageSkill)
		resourcedata.SetMapValueIfNotNil(skillGroupLanguageConditionMap, "comparator", skillGroupLanguageCondition.Comparator)
		resourcedata.SetMapValueIfNotNil(skillGroupLanguageConditionMap, "proficiency", skillGroupLanguageCondition.Proficiency)
		resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(skillGroupLanguageConditionMap, "child_conditions", skillGroupLanguageCondition.ChildConditions, flattenSkillGroupConditions)

		skillGroupLanguageConditionList = append(skillGroupLanguageConditionList, skillGroupLanguageConditionMap)
	}

	return skillGroupLanguageConditionList
}

// flattenSkillGroupConditions maps a Genesys Cloud *[]platformclientv2.Skillgroupcondition into a []interface{}
func flattenSkillGroupConditions(skillGroupConditions *[]platformclientv2.Skillgroupcondition) []interface{} {
	if len(*skillGroupConditions) == 0 {
		return nil
	}

	var skillGroupConditionList []interface{}
	for _, skillGroupCondition := range *skillGroupConditions {
		skillGroupConditionMap := make(map[string]interface{})

		resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(skillGroupConditionMap, "routing_skill_conditions", skillGroupCondition.RoutingSkillConditions, flattenSkillGroupRoutingConditions)
		resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(skillGroupConditionMap, "language_skill_conditions", skillGroupCondition.LanguageSkillConditions, flattenSkillGroupLanguageConditions)
		resourcedata.SetMapValueIfNotNil(skillGroupConditionMap, "operation", skillGroupCondition.Operation)

		skillGroupConditionList = append(skillGroupConditionList, skillGroupConditionMap)
	}

	return skillGroupConditionList
}

// Todo: refactor

/*
Sometimes you just need to get ugly.  skillConditions has a recursive function that is super ugly to manage to a static Golang
Struct.  So our struct always has a placeholder "skillConditions": {} field. So what I do is convert the struct to JSON and then
check to see if skill_conditions on the Terraform resource data.  I then do a string replace on the skillConditions json attribute
and replace the empty stringConditions string with the contents of skill_conditions.

Not the most eloquent code, but these are uncivilized times.
*/
func mergeSkillConditionsIntoSkillGroups(d *schema.ResourceData, skillGroupsRequest *SkillGroupsRequest) (string, error) {
	skillsConditionsJsonString := fmt.Sprintf(`"skillConditions": %s`, d.Get("skill_conditions").(string))

	//Get the before image of the JSON.  Note this a byte array
	skillGroupsRequestBefore, err := json.Marshal(skillGroupsRequest)
	if err != nil {
		return "", err
	}

	skillGroupsRequestAfter := ""

	//Skill conditions are present, replace skill conditions with the content of the string
	if d.Get("skill_conditions").(string) != "" {
		skillGroupsRequestAfter = strings.Replace(string(skillGroupsRequestBefore), `"skillConditions":{}`, skillsConditionsJsonString, 1)
	} else {
		//Skill conditions are not present, get rid of skill conditions.
		skillGroupsRequestAfter = strings.Replace(string(skillGroupsRequestBefore), `,"skillConditions":{}`, "", 1)
	}

	return skillGroupsRequestAfter, nil
}

func mergeSkillConditionsIntoSkillGroupsCreate(d *schema.ResourceData, skillGroupCreate *platformclientv2.Skillgroupwithmemberdivisions) (string, error) {
	skillsConditionsJsonString := fmt.Sprintf(`"skillConditions": %s`, d.Get("skill_conditions").(string))

	//Get the before image of the JSON.  Note this a byte array
	skillGroupsRequestBefore, err := json.Marshal(skillGroupCreate)
	if err != nil {
		return "", err
	}

	skillGroupsRequestAfter := ""

	//Skill conditions are present, replace skill conditions with the content of the string
	if d.Get("skill_conditions").(string) != "" {
		skillGroupsRequestAfter = strings.Replace(string(skillGroupsRequestBefore), `"skillConditions":{}`, skillsConditionsJsonString, 1)
	} else {
		//Skill conditions are not present, get rid of skill conditions.
		skillGroupsRequestAfter = strings.Replace(string(skillGroupsRequestBefore), `,"skillConditions":{}`, "", 1)
	}

	return skillGroupsRequestAfter, nil
}

func mergeSkillConditionsIntoSkillGroupsUpdate(d *schema.ResourceData, skillGroupUpdate *platformclientv2.Skillgroup) (string, error) {
	skillsConditionsJsonString := fmt.Sprintf(`"skillConditions": %s`, d.Get("skill_conditions").(string))

	//Get the before image of the JSON.  Note this a byte array
	skillGroupsRequestBefore, err := json.Marshal(skillGroupUpdate)
	if err != nil {
		return "", err
	}

	skillGroupsRequestAfter := ""

	//Skill conditions are present, replace skill conditions with the content of the string
	if d.Get("skill_conditions").(string) != "" {
		skillGroupsRequestAfter = strings.Replace(string(skillGroupsRequestBefore), `"skillConditions":{}`, skillsConditionsJsonString, 1)
	} else {
		//Skill conditions are not present, get rid of skill conditions.
		skillGroupsRequestAfter = strings.Replace(string(skillGroupsRequestBefore), `,"skillConditions":{}`, "", 1)
	}

	return skillGroupsRequestAfter, nil
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
