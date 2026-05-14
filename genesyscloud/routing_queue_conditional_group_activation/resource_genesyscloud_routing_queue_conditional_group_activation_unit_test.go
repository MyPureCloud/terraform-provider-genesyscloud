package routing_queue_conditional_group_activation

import (
	"testing"

	"github.com/google/uuid"
	"github.com/mypurecloud/platform-client-sdk-go/v179/platformclientv2"
)

func TestUnitBuildConditionalGroupActivation(t *testing.T) {
	var (
		conditionExpression = "(C1 and C2)"
		operator1           = "GreaterThan"
		metric1             = "EstimatedWaitTime"
		value1              = 10.0
		operator2           = "LessThan"
		metric2             = "IdleAgentCount"
		value2              = 5.0
		groupId             = uuid.NewString()
		groupType           = "SKILLGROUP"
	)

	cgaConfig := map[string]interface{}{
		"rules": []interface{}{
			map[string]interface{}{
				"condition_expression": conditionExpression,
				"conditions": []interface{}{
					map[string]interface{}{
						"operator": operator1,
						"value":    value1,
						"simple_metric": []interface{}{
							map[string]interface{}{
								"metric": metric1,
							},
						},
					},
					map[string]interface{}{
						"operator": operator2,
						"value":    value2,
						"simple_metric": []interface{}{
							map[string]interface{}{
								"metric": metric2,
							},
						},
					},
				},
				"groups": []interface{}{
					map[string]interface{}{
						"member_group_id":   groupId,
						"member_group_type": groupType,
					},
				},
			},
		},
	}

	sdkCga := buildConditionalGroupActivation(cgaConfig)

	if sdkCga.Rules == nil {
		t.Fatalf("Expected rules to not be nil")
	}
	if len(*sdkCga.Rules) != 1 {
		t.Fatalf("Expected 1 rule, got %d", len(*sdkCga.Rules))
	}

	rule := (*sdkCga.Rules)[0]
	if rule.ConditionExpression == nil || *rule.ConditionExpression != conditionExpression {
		t.Errorf("Expected condition_expression to be '%s', got '%v'", conditionExpression, rule.ConditionExpression)
	}
	if rule.Conditions == nil || len(*rule.Conditions) != 2 {
		t.Fatalf("Expected 2 conditions, got %v", rule.Conditions)
	}
	if rule.Groups == nil || len(*rule.Groups) != 1 {
		t.Fatalf("Expected 1 group, got %v", rule.Groups)
	}

	condition1 := (*rule.Conditions)[0]
	if condition1.Operator == nil || *condition1.Operator != operator1 {
		t.Errorf("Expected operator '%s', got '%v'", operator1, condition1.Operator)
	}
	if condition1.Value == nil || *condition1.Value != value1 {
		t.Errorf("Expected value %f, got %v", value1, condition1.Value)
	}
	if condition1.SimpleMetric == nil || condition1.SimpleMetric.Metric == nil || *condition1.SimpleMetric.Metric != metric1 {
		t.Errorf("Expected metric '%s', got %v", metric1, condition1.SimpleMetric)
	}

	group := (*rule.Groups)[0]
	if group.Id == nil || *group.Id != groupId {
		t.Errorf("Expected group id '%s', got '%v'", groupId, group.Id)
	}
	if group.VarType == nil || *group.VarType != groupType {
		t.Errorf("Expected group type '%s', got '%v'", groupType, group.VarType)
	}
}

func TestUnitFlattenConditionalGroupActivation(t *testing.T) {
	var (
		conditionExpression = "(C1 and C2)"
		operator1           = "GreaterThan"
		metric1             = "EstimatedWaitTime"
		value1              = 60.0
		groupId             = uuid.NewString()
		groupType           = "SKILLGROUP"
	)

	sdkCga := &platformclientv2.Conditionalgroupactivation{
		Rules: &[]platformclientv2.Conditionalgroupactivationrule{
			{
				ConditionExpression: &conditionExpression,
				Conditions: &[]platformclientv2.Conditionalgroupactivationcondition{
					{
						Operator: &operator1,
						Value:    &value1,
						SimpleMetric: &platformclientv2.Conditionalgroupactivationsimplemetric{
							Metric: &metric1,
						},
					},
				},
				Groups: &[]platformclientv2.Membergroup{
					{
						Id:      &groupId,
						VarType: &groupType,
					},
				},
			},
		},
	}

	flattened := flattenConditionalGroupActivation(sdkCga)
	if flattened == nil {
		t.Fatalf("Expected non-nil result")
	}

	rules, ok := flattened["rules"].([]interface{})
	if !ok {
		t.Fatalf("Expected rules to be []interface{}, got %T", flattened["rules"])
	}
	if len(rules) != 1 {
		t.Fatalf("Expected 1 rule, got %d", len(rules))
	}

	rule, ok := rules[0].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected rule to be map[string]interface{}, got %T", rules[0])
	}

	if ce, ok := rule["condition_expression"].(string); !ok || ce != conditionExpression {
		t.Errorf("Expected condition_expression '%s', got '%v'", conditionExpression, rule["condition_expression"])
	}

	conditions, ok := rule["conditions"].([]interface{})
	if !ok || len(conditions) != 1 {
		t.Fatalf("Expected 1 condition, got %v", rule["conditions"])
	}

	groups, ok := rule["groups"].([]interface{})
	if !ok || len(groups) != 1 {
		t.Fatalf("Expected 1 group, got %v", rule["groups"])
	}
}

func TestUnitFlattenConditionalGroupActivationNil(t *testing.T) {
	flattened := flattenConditionalGroupActivation(nil)
	if flattened != nil {
		t.Errorf("Expected nil for nil input, got %v", flattened)
	}
}

func TestUnitBuildAndFlattenWithPilotRule(t *testing.T) {
	var (
		pilotExpression = "C1"
		pilotOperator   = "GreaterThan"
		pilotMetric     = "EstimatedWaitTime"
		pilotValue      = 30.0
		ruleExpression  = "C1"
		ruleOperator    = "LessThan"
		ruleMetric      = "ServiceLevel"
		ruleValue       = 0.8
		groupId         = uuid.NewString()
		groupType       = "GROUP"
	)

	cgaConfig := map[string]interface{}{
		"pilot_rule": []interface{}{
			map[string]interface{}{
				"condition_expression": pilotExpression,
				"conditions": []interface{}{
					map[string]interface{}{
						"operator": pilotOperator,
						"value":    pilotValue,
						"simple_metric": []interface{}{
							map[string]interface{}{
								"metric": pilotMetric,
							},
						},
					},
				},
			},
		},
		"rules": []interface{}{
			map[string]interface{}{
				"condition_expression": ruleExpression,
				"conditions": []interface{}{
					map[string]interface{}{
						"operator": ruleOperator,
						"value":    ruleValue,
						"simple_metric": []interface{}{
							map[string]interface{}{
								"metric": ruleMetric,
							},
						},
					},
				},
				"groups": []interface{}{
					map[string]interface{}{
						"member_group_id":   groupId,
						"member_group_type": groupType,
					},
				},
			},
		},
	}

	sdkCga := buildConditionalGroupActivation(cgaConfig)

	if sdkCga.PilotRule == nil {
		t.Fatalf("Expected pilot_rule to not be nil")
	}
	if sdkCga.PilotRule.ConditionExpression == nil || *sdkCga.PilotRule.ConditionExpression != pilotExpression {
		t.Errorf("Expected pilot condition_expression '%s', got '%v'", pilotExpression, sdkCga.PilotRule.ConditionExpression)
	}
	if sdkCga.PilotRule.Conditions == nil || len(*sdkCga.PilotRule.Conditions) != 1 {
		t.Fatalf("Expected 1 pilot condition, got %v", sdkCga.PilotRule.Conditions)
	}

	flattened := flattenConditionalGroupActivation(&sdkCga)
	if flattened == nil {
		t.Fatalf("Expected non-nil flattened result")
	}

	pilotRules, ok := flattened["pilot_rule"].([]interface{})
	if !ok || len(pilotRules) != 1 {
		t.Fatalf("Expected 1 pilot_rule in flattened, got %v", flattened["pilot_rule"])
	}

	rules, ok := flattened["rules"].([]interface{})
	if !ok || len(rules) != 1 {
		t.Fatalf("Expected 1 rule in flattened, got %v", flattened["rules"])
	}
}
