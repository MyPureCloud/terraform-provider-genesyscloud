package routing_queue_conditional_group_activation

import (
	routingQueue "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/routing_queue"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/resourcedata"

	"github.com/mypurecloud/platform-client-sdk-go/v179/platformclientv2"
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
		result["rules"] = routingQueue.FlattenCgaRules(sdkCga.Rules)
	}

	return result
}
