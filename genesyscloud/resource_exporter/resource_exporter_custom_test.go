package resource_exporter

import (
	"fmt"
	"testing"

	"encoding/json"

	"github.com/google/uuid"
)

type customMemberGroupTest struct {
	MemberGroupID   string
	MemberGroupType string
	GroupName       string
	ExpectedRefType string
}

type propertyGroupTest struct {
	Skills               string
	SkillName            string
	ExporterResourceType string
	ExpectedResult       string
}

func TestUnitOmitUnresolvedGuidFromConfigMap(t *testing.T) {
	guid := uuid.NewString()
	configMap := map[string]interface{}{
		"contact_list_id": guid,
	}

	OmitUnresolvedGuidFromConfigMap(configMap, "contact_list_id")
	if _, ok := configMap["contact_list_id"]; ok {
		t.Fatal("expected unresolved GUID to be omitted from config map")
	}

	configMap = map[string]interface{}{
		"contact_list_id": "${genesyscloud_outbound_contact_list.example.id}",
	}
	OmitUnresolvedGuidFromConfigMap(configMap, "contact_list_id")
	if configMap["contact_list_id"] == nil {
		t.Fatal("expected resolved reference to be kept in config map")
	}
}

// TestUnitExporterCustomMemberGroup uses a table based approach to test the three different types of groups that can be resolved.
// We currently support SKILLGROUP and GROUP.  Team has not been implemented yet so the custom resolver should return keep the original id associated the config map.
func TestUnitExporterCustomMemberGroup(t *testing.T) {
	teamID := uuid.NewString()
	testResults := []*customMemberGroupTest{
		{MemberGroupID: uuid.NewString(), MemberGroupType: "SKILLGROUP", GroupName: "test_skill_group_name", ExpectedRefType: "genesyscloud_routing_skill_group"},
		{MemberGroupID: uuid.NewString(), MemberGroupType: "GROUP", GroupName: "test_group_name", ExpectedRefType: "genesyscloud_group"},
		{MemberGroupID: teamID, MemberGroupType: "TEAM", GroupName: "test_team_name", ExpectedRefType: "genesyscloud_team"},
	}

	for _, testResult := range testResults {
		configMap := make(map[string]interface{})

		//Make the config map object
		configMap["member_group_type"] = testResult.MemberGroupType
		configMap["member_group_id"] = testResult.MemberGroupID

		//Pre-Check to make sure the member_group_id has been set to the GUID I have at the start of the test
		if configMap["member_group_id"] != testResult.MemberGroupID {
			t.Errorf("The member_group_id set in the config map was %v,but  wanted %v", configMap["member_group_id"], testResult.MemberGroupID)
		}

		refType, err := MemberGroupsResolver(configMap)
		if err != nil {
			t.Errorf("Received an unexpected error while calling MemberGroupsResolver: %v", err)
		}

		if refType != testResult.ExpectedRefType {
			t.Errorf("Expected ref type %v but got %v", testResult.ExpectedRefType, refType)
		}
	}

}

func TestUnitExporterCustomMemberGroupMissingOrInvalidType(t *testing.T) {
	_, err := MemberGroupsResolver(map[string]interface{}{})
	if err == nil {
		t.Error("expected error for missing member_group_type")
	}

	_, err = MemberGroupsResolver(map[string]interface{}{"member_group_type": 123})
	if err == nil {
		t.Error("expected error for non-string member_group_type")
	}

	_, err = MemberGroupsResolver(map[string]interface{}{"member_group_type": "NOPE"})
	if err == nil {
		t.Error("expected error for unknown member_group_type")
	}
}

func TestUnitRuleSetPropertyGroup(t *testing.T) {

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
		skillSanitizedResourceMap[uuid] = &ResourceMeta{BlockLabel: testResult.SkillName}

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

// TestUnitReplyEmailAddressSelfReferenceRouteExporterResolver verifies DEVTOOLING-1565: when a route
// self-references as its reply address, the export resolver must set self_reference_route=true and
// clear route_id and domain_id. Leaving domain_id in place conflicts with self_reference_route
// and causes terraform plan to fail.
func TestUnitReplyEmailAddressSelfReferenceRouteExporterResolver(t *testing.T) {
	resourceLabel := "example_email_route"
	configMap := map[string]interface{}{
		"domain_id": "${data.genesyscloud_routing_email_domain.example_email_domain.id}",
		"route_id":  fmt.Sprintf("${genesyscloud_routing_email_route.%s.id}", resourceLabel),
	}

	err := ReplyEmailAddressSelfReferenceRouteExporterResolver(configMap, nil, resourceLabel)
	if err != nil {
		t.Fatalf("unexpected error from resolver: %v", err)
	}

	selfReferenceRoute, ok := configMap["self_reference_route"].(bool)
	if !ok || !selfReferenceRoute {
		t.Fatalf("expected self_reference_route=true, got %#v", configMap["self_reference_route"])
	}

	if configMap["route_id"] != nil {
		t.Fatalf("expected route_id to be cleared, got %#v", configMap["route_id"])
	}

	if configMap["domain_id"] != nil {
		t.Fatalf("expected domain_id to be cleared, got %#v", configMap["domain_id"])
	}
}
