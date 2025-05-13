package routing_queue_conditional_group_routing

import (
	"fmt"
	routingQueue "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/routing_queue"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/resourcedata"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v157/platformclientv2"
)

func buildConditionalGroupRouting(rules []interface{}) ([]platformclientv2.Conditionalgrouproutingrule, error) {
	var sdkRules []platformclientv2.Conditionalgrouproutingrule
	for i, rule := range rules {
		configRule, ok := rule.(map[string]interface{})
		if !ok {
			continue
		}
		var sdkRule platformclientv2.Conditionalgrouproutingrule

		if operator, ok := configRule["operator"].(string); ok {
			sdkRule.Operator = &operator
		}

		if conditionValue, ok := configRule["condition_value"].(float64); ok {
			sdkRule.ConditionValue = &conditionValue
		}

		if evaluatedQueue, ok := configRule["evaluated_queue_id"].(string); ok && evaluatedQueue != "" {
			if i == 0 {
				return nil, fmt.Errorf("for rule 1, the current queue is used so evaluated_queue_id should not be specified")
			}
			sdkRule.Queue = &platformclientv2.Domainentityref{Id: &evaluatedQueue}
		}

		resourcedata.BuildSDKStringValueIfNotNil(&sdkRule.Metric, configRule, "metric")
		if waitSeconds, ok := configRule["wait_seconds"].(int); ok {
			sdkRule.WaitSeconds = &waitSeconds
		}

		if memberGroupSet, ok := configRule["groups"].(*schema.Set); ok {
			sdkRule.Groups = routingQueue.BuildCGRGroups(memberGroupSet)
		}

		sdkRules = append(sdkRules, sdkRule)
	}

	return sdkRules, nil
}

func flattenConditionalGroupRouting(sdkRules *[]platformclientv2.Conditionalgrouproutingrule) []interface{} {
	if sdkRules == nil {
		return nil
	}

	var rules []interface{}
	for i, sdkRule := range *sdkRules {
		rule := make(map[string]interface{})

		// The first rule is assumed to apply to this queue, so evaluated_queue_id should be omitted
		if i > 0 {
			resourcedata.SetMapReferenceValueIfNotNil(rule, "evaluated_queue_id", sdkRule.Queue)
		}
		resourcedata.SetMapValueIfNotNil(rule, "wait_seconds", sdkRule.WaitSeconds)
		resourcedata.SetMapValueIfNotNil(rule, "operator", sdkRule.Operator)
		resourcedata.SetMapValueIfNotNil(rule, "condition_value", sdkRule.ConditionValue)
		resourcedata.SetMapValueIfNotNil(rule, "metric", sdkRule.Metric)

		if sdkRule.Groups != nil {
			rule["groups"] = routingQueue.FlattenCGRGroups(*sdkRule.Groups)
		}

		rules = append(rules, rule)
	}
	return rules
}
