package routing_queue_conditional_group_routing

import (
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v165/platformclientv2"
)

func TestUnitBuildConditionalGroupRouting(t *testing.T) {
	var (
		metric                 = "test"
		operator               = "GreaterThan"
		conditionValue float64 = 2345
		waitSeconds            = 5432
	)

	var cgrRulesMap = make(map[string]any)
	cgrRulesMap["metric"] = metric
	cgrRulesMap["operator"] = operator
	cgrRulesMap["condition_value"] = conditionValue
	cgrRulesMap["wait_seconds"] = waitSeconds
	cgrRulesMap["groups"] = schema.NewSet(schema.HashResource(memberGroupResource), []interface{}{
		map[string]any{
			"member_group_id":   "111",
			"member_group_type": "222",
		},
		map[string]any{
			"member_group_id":   "333",
			"member_group_type": "444",
		},
		map[string]any{
			"member_group_id":   "555",
			"member_group_type": "666",
		},
	})

	cgrRulesList := []interface{}{cgrRulesMap}

	// testing buildConditionalGroupRouting function
	cgrRules, err := buildConditionalGroupRouting(cgrRulesList)
	if err != nil {
		t.Errorf("Error building conditional group routing: %v", err)
	}
	if cgrRules == nil {
		t.Fatalf("Expected conditional group routing, got nil")
	}
	if len(cgrRules) != 1 {
		t.Errorf("Expected 1 rule, got %d", len(cgrRules))
	}
	cgrRule := cgrRules[0]
	if cgrRule.Metric == nil {
		t.Fatalf("Expected metric to not be nil")
	}
	if *cgrRule.Metric != metric {
		t.Errorf("Expected metric to be '%s', got %s", metric, *cgrRule.Metric)
	}

	if cgrRule.Operator == nil {
		t.Fatalf("Expected operator to not be nil")
	}
	if *cgrRule.Operator != operator {
		t.Errorf("Expected operator to be '%s', got %s", operator, *cgrRule.Operator)
	}

	if cgrRule.ConditionValue == nil {
		t.Fatalf("Expected condition_value to not be nil")
	}
	if *cgrRule.ConditionValue != conditionValue {
		t.Errorf("Expected condition_value to be %f, got %f", conditionValue, *cgrRule.ConditionValue)
	}

	if cgrRule.WaitSeconds == nil {
		t.Fatalf("Expected WaitSeconds to not be nil")
	}
	if *cgrRule.WaitSeconds != waitSeconds {
		t.Errorf("Expected WaitSeconds to be %d, got %d", waitSeconds, *cgrRule.WaitSeconds)
	}

	if cgrRule.Groups == nil {
		t.Fatalf("Expected groups to not be nil")
	}
	if len(*cgrRule.Groups) != 3 {
		t.Errorf("Expected 3 groups, got %d", len(*cgrRule.Groups))
	}
}

func TestUnitFlattenConditionalGroupRouting(t *testing.T) {
	var (
		metric                 = "test"
		operator               = "GreaterThan"
		conditionValue float64 = 2345
		waitSeconds            = 5432
	)
	cgrRules := generateRuleData(metric, operator, conditionValue, waitSeconds)

	// testing buildConditionalGroupRouting function
	cgrRulesFlattened := flattenConditionalGroupRouting(&cgrRules)
	if len(cgrRulesFlattened) != 1 {
		t.Errorf("Expected 1 rule, got %d", len(cgrRulesFlattened))
	}

	cgrRule, ok := cgrRulesFlattened[0].(map[string]any)
	if !ok {
		t.Fatalf("Expected map[string]any, got %T", cgrRulesFlattened[0])
	}

	metricFlattened, ok := cgrRule["metric"].(string)
	if !ok {
		t.Fatalf("Expected metric to be a string, got %T", cgrRule["metric"])
	}
	if metricFlattened != metric {
		t.Errorf("Expected metric to be 'test', got %s", cgrRule["metric"].(string))
	}

	operatorFlattened, ok := cgrRule["operator"].(string)
	if !ok {
		t.Fatalf("Expected operator to be a string, got %T", cgrRule["operator"])
	}
	if operatorFlattened != "GreaterThan" {
		t.Errorf("Expected operator to be 'GreaterThan', got %s", operatorFlattened)
	}

	conditionValueFlattened, ok := cgrRule["condition_value"].(float64)
	if !ok {
		t.Fatalf("Expected condition_value to be a float64, got %T", cgrRule["condition_value"])
	}
	if conditionValueFlattened != 2345 {
		t.Errorf("Expected condition_value to be 2345, got %f", conditionValueFlattened)
	}

	groupsFlattened, ok := cgrRule["groups"].(*schema.Set)
	if !ok {
		t.Fatalf("Expected *schema.Set, got %T", cgrRule["groups"])
	}
	if len(groupsFlattened.List()) != 3 {
		t.Errorf("Expected 3 groups, got %d", len(groupsFlattened.List()))
	}
}

func generateRuleData(metric, operator string, conditionValue float64, waitSeconds int) []platformclientv2.Conditionalgrouproutingrule {
	groupMember1 := platformclientv2.Membergroup{
		Id:      platformclientv2.String(uuid.NewString()),
		VarType: platformclientv2.String("TEAM"),
	}
	groupMember2 := platformclientv2.Membergroup{
		Id:      platformclientv2.String(uuid.NewString()),
		VarType: platformclientv2.String("SKILLGROUP"),
	}
	groupMember3 := platformclientv2.Membergroup{
		Id:      platformclientv2.String(uuid.NewString()),
		VarType: platformclientv2.String("GROUP"),
	}
	group1 := []platformclientv2.Membergroup{groupMember1, groupMember2, groupMember3}

	rule1 := platformclientv2.Conditionalgrouproutingrule{
		Metric:         platformclientv2.String(metric),
		Operator:       platformclientv2.String(operator),
		ConditionValue: platformclientv2.Float64(conditionValue),
		Groups:         &group1,
		WaitSeconds:    platformclientv2.Int(waitSeconds),
	}

	rules := []platformclientv2.Conditionalgrouproutingrule{rule1}

	return rules
}
