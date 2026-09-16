package routing_queue_conditional_group_activation

import (
	"strings"

	routingQueue "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/routing_queue"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/resourcedata"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v195/platformclientv2"
)

// conditionalGroupActivationIdSuffix disambiguates this resource's ID from the routing queue's
// own ID, since a conditional group activation resource is 1:1 with (and keyed off of) its queue.
const conditionalGroupActivationIdSuffix = "/cga"

// BuildConditionalGroupActivationId builds the routing_queue_conditional_group_activation
// composite resource ID, which is always in the format <queue-id>/cga.
func BuildConditionalGroupActivationId(queueId string) (id string) {
	return queueId + conditionalGroupActivationIdSuffix
}

// SplitConditionalGroupActivationId splits the routing_queue_conditional_group_activation
// composite resource ID, which is always in the format <queue-id>/cga, back into the queue ID.
func SplitConditionalGroupActivationId(id string) (queueId string) {
	return strings.Split(id, "/")[0]
}

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
