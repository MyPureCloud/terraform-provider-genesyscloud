package routing_queue_conditional_group_activation

import (
	routingQueue "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/routing_queue"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/resourcedata"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v192/platformclientv2"
)

func buildConditionalGroupActivation(d map[string]interface{}) platformclientv2.Conditionalgroupactivation {
	var sdkCga platformclientv2.Conditionalgroupactivation

	resourcedata.BuildSDKInterfaceArrayValueIfNotNil(&sdkCga.PilotRule, d, "pilot_rule", routingQueue.BuildCgaPilotRule)
	resourcedata.BuildSDKInterfaceArrayValueIfNotNil(&sdkCga.Rules, d, "rules", routingQueue.BuildCgaNumberedRules)

	return sdkCga
}

func flattenConditionalGroupActivation(sdkCga *platformclientv2.Conditionalgroupactivation) map[string]interface{} {
	if sdkCga == nil {
		return nil
	}

	result := make(map[string]interface{})

	if sdkCga.PilotRule != nil {
		pilotRuleMap := make(map[string]interface{})
		resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(pilotRuleMap, "conditions", sdkCga.PilotRule.Conditions, routingQueue.FlattenCgaRuleConditions)
		resourcedata.SetMapValueIfNotNil(pilotRuleMap, "condition_expression", sdkCga.PilotRule.ConditionExpression)
		result["pilot_rule"] = []interface{}{pilotRuleMap}
	}

	if sdkCga.Rules != nil {
		// FlattenCgaRules returns groups as *schema.Set for genesyscloud_routing_queue nested CGA (TypeSet).
		// This resource uses TypeList for rules.*.groups; Terraform requires a slice here, not a Set.
		result["rules"] = cgaRuleGroupsSetToList(routingQueue.FlattenCgaRules(sdkCga.Rules))
	}

	return result
}

func cgaRuleGroupsSetToList(rules []interface{}) []interface{} {
	for _, rule := range rules {
		ruleMap, ok := rule.(map[string]interface{})
		if !ok {
			continue
		}
		if groupsVal, ok := ruleMap["groups"]; ok {
			if set, ok := groupsVal.(*schema.Set); ok {
				ruleMap["groups"] = set.List()
			}
		}
	}
	return rules
}
