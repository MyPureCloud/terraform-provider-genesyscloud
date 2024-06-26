package outbound_ruleset

import (
	"testing"

	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"

	"github.com/stretchr/testify/assert"

	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

func TestUnitDoesRuleConditionsRefDeletedSkill(t *testing.T) {
	rule := platformclientv2.Dialerrule{}
	skillMap := resourceExporter.ResourceIDMetaMap{
		"Skill1": {Name: "Skill1"},
		"Skill2": {Name: "Skill2"},
		"Skill3": {Name: "Skill3"},
	}

	// Test case 1: All skills exist in the map, function should return false
	rule.Conditions = &[]platformclientv2.Condition{
		{
			AttributeName: platformclientv2.String("skill"),
			Value:         platformclientv2.String("Skill1"),
		},
		{
			AttributeName: platformclientv2.String("skill"),
			Value:         platformclientv2.String("Skill2"),
		},
	}
	assert.False(t, doesRuleConditionsRefDeletedSkill(rule, skillMap))

	// Test case 2: One skill does not exist in the map, function should return true
	rule.Conditions = &[]platformclientv2.Condition{
		{
			AttributeName: platformclientv2.String("skill"),
			Value:         platformclientv2.String("Skill2"),
		},
		{
			AttributeName: platformclientv2.String("skill"),
			Value:         platformclientv2.String("NonExistentSkill"),
		},
		{
			AttributeName: platformclientv2.String("skill"),
			Value:         platformclientv2.String("Skill3"),
		},
	}
	assert.True(t, doesRuleConditionsRefDeletedSkill(rule, skillMap))
}

func TestUnitDoesRuleActionsRefDeletedSkill(t *testing.T) {
	rule := platformclientv2.Dialerrule{}
	skillMap := resourceExporter.ResourceIDMetaMap{
		"Skill1": {Name: "Skill1"},
		"Skill2": {Name: "Skill2"},
		"Skill3": {Name: "Skill3"},
	}

	// Test case 1: All skills exist in the map, function should return false
	rule.Actions = &[]platformclientv2.Dialeraction{
		{
			ActionTypeName: platformclientv2.String("set_skills"),
			Properties: &map[string]string{
				"skills": `["Skill1", "Skill2"]`,
			},
		},
	}
	exists := doesRuleActionsRefDeletedSkill(rule, skillMap)
	assert.False(t, exists)

	// Test case 2: One skill does not exist in the map, function should return true
	rule.Actions = &[]platformclientv2.Dialeraction{
		{
			ActionTypeName: platformclientv2.String("set_skills"),
			Properties: &map[string]string{
				"skills": `["Skill2", "NonExistentSkill", "Skill3"]`,
			},
		},
	}
	exists = doesRuleActionsRefDeletedSkill(rule, skillMap)
	assert.True(t, exists)

	// Test case 3: JSON unmarshaling error, function should return true
	rule.Actions = &[]platformclientv2.Dialeraction{
		{
			ActionTypeName: platformclientv2.String("set_skills"),
			Properties: &map[string]string{
				"skills": "invalid-json",
			},
		},
	}
	exists = doesRuleActionsRefDeletedSkill(rule, skillMap)
	assert.True(t, exists)
}
