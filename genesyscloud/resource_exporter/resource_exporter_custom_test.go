package resource_exporter

import (
	"testing"

	"encoding/json"

	"github.com/google/uuid"
)

type customMemberGroupTest struct {
	MemberGroupID        string
	MemberGroupType      string
	ExporterResourceType string
	GroupName            string
	ExpectedResult       string
}

type propertyGroupTest struct {
	Skills               string
	SkillName            string
	ExporterResourceType string
	ExpectedResult       string
}

/*
This test is more of a unit test then am acceptance test.  It is using a table based approach to test the three different types of groups that can be resolved.  We currently support SKILLGROUP and GROUP.  Team has not been implemented
yet so the custom resolver should return keep the original id associated the config map.

This is a unit test because it is just testing this single function without any dependency of Terraform actually being run.
*/
func TestAccExporterCustomMemberGroup(t *testing.T) {
	teamID := uuid.NewString()
	testResults := []*customMemberGroupTest{
		{MemberGroupID: uuid.NewString(), MemberGroupType: "SKILLGROUP", GroupName: "test_skill_group_name", ExporterResourceType: "genesyscloud_routing_skill_group", ExpectedResult: "${genesyscloud_routing_skill_group.test_skill_group_name.id}"},
		{MemberGroupID: uuid.NewString(), MemberGroupType: "GROUP", GroupName: "test_group_name", ExporterResourceType: "genesyscloud_group", ExpectedResult: "${genesyscloud_group.test_group_name.id}"},
		{MemberGroupID: teamID, MemberGroupType: "TEAM", GroupName: "test_team_name", ExporterResourceType: "genesyscloud_team_NA", ExpectedResult: teamID},
	}

	for _, testResult := range testResults {
		configMap := make(map[string]interface{})
		exporters := make(map[string]*ResourceExporter)

		//Make the config map object
		configMap["member_group_type"] = testResult.MemberGroupType
		configMap["member_group_id"] = testResult.MemberGroupID

		//Create an exporter
		skillGroupSanitizedResourceMap := make(map[string]*ResourceMeta)
		skillGroupSanitizedResourceMap[testResult.MemberGroupID] = &ResourceMeta{Name: testResult.GroupName}

		firstResourceExport := &ResourceExporter{
			SanitizedResourceMap: skillGroupSanitizedResourceMap,
		}
		exporters[testResult.ExporterResourceType] = firstResourceExport

		//Pre-Check to make sure the member_group_id has been set to the GUID I have at the start of the test
		if configMap["member_group_id"] != testResult.MemberGroupID {
			t.Errorf("The member_group_id set in the config map was %v,but  wanted %v", configMap["member_group_id"], testResult.MemberGroupID)
		}

		//Invoke the resolver
		err := MemberGroupsResolver(configMap, exporters, testResult.ExporterResourceType)

		if err != nil && testResult.MemberGroupType != "TEAM" {
			t.Errorf("Received an unexpected error while calling MemberGroupResolver:  %v", err)
		}

		//The member_group_id should now be replaced by the expected out put with th
		if configMap["member_group_id"].(string) != testResult.ExpectedResult {
			t.Errorf("The member_group_id set in the config map was %v, but wanted %v", configMap["member_group_id"], testResult.ExpectedResult)
		}
	}

}

func TestRuleSetPropertyGroup(t *testing.T) {

	uuid := uuid.NewString()

	jsonData, err := json.Marshal([]string{uuid})
	if err != nil {
		t.Errorf("Received an unexpected error converting json:  %v", err)
	}
	jsonString := string(jsonData)

	testResults := []*propertyGroupTest{
		{Skills: jsonString, SkillName: "test_skill_name", ExporterResourceType: "genesyscloud_routing_skill", ExpectedResult: "[\"${genesyscloud_routing_skill.test_skill_name.id}\"]"},
	}

	for _, testResult := range testResults {
		configMap := make(map[string]interface{})
		exporters := make(map[string]*ResourceExporter)

		//Make the config map object
		configMap["skills"] = testResult.Skills

		//Create an exporter
		skillSanitizedResourceMap := make(map[string]*ResourceMeta)
		skillSanitizedResourceMap[uuid] = &ResourceMeta{Name: testResult.SkillName}

		firstResourceExport := &ResourceExporter{
			SanitizedResourceMap: skillSanitizedResourceMap,
		}
		exporters[testResult.ExporterResourceType] = firstResourceExport

		//Pre-Check to make sure the member_group_id has been set to the GUID I have at the start of the test
		if configMap["skills"] != testResult.Skills {
			t.Errorf("The skills set in the config map was %v,but  wanted %v", configMap["skills"], testResult.Skills)
		}

		//Invoke the resolver
		err := RuleSetSkillPropertyResolver(configMap, exporters, testResult.ExporterResourceType)

		if err != nil {
			t.Errorf("Received an unexpected error while calling RuleSetSkillPropertyResolver:  %v", err)
		}

		if configMap["skills"].(string) != testResult.ExpectedResult {
			t.Errorf("The skills set in the config map was %v,but  wanted %v", configMap["skills"], testResult.ExpectedResult)
		}
	}

}
