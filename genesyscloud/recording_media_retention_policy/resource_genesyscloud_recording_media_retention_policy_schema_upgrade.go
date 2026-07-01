package recording_media_retention_policy

import (
	"context"
	"sort"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/lists"
)

func resourceMediaRetentionPolicySchemaV1() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"conditions": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"team_ids": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
			"media_policies": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"call_policy":    mediaPolicyBlockSchemaV1(),
						"chat_policy":    mediaPolicyBlockSchemaV1(),
						"email_policy":   mediaPolicyBlockSchemaV1(),
						"message_policy": mediaPolicyBlockSchemaV1(),
					},
				},
			},
		},
	}
}

func mediaPolicyBlockSchemaV1() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"conditions": {
					Type:     schema.TypeList,
					Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"team_ids": {
								Type:     schema.TypeSet,
								Optional: true,
								Elem:     &schema.Schema{Type: schema.TypeString},
							},
						},
					},
				},
			},
		},
	}
}

func upgradeMediaRetentionPolicyStateV1(_ context.Context, rawState map[string]interface{}, _ interface{}) (map[string]interface{}, error) {
	upgradeTeamIdsInConditionsList(rawState["conditions"])
	upgradeTeamIdsInMediaPolicies(rawState["media_policies"])
	return rawState, nil
}

func upgradeTeamIdsInMediaPolicies(raw interface{}) {
	mediaPolicies, ok := raw.([]interface{})
	if !ok {
		return
	}

	for _, mediaPolicy := range mediaPolicies {
		mediaPolicyMap, ok := mediaPolicy.(map[string]interface{})
		if !ok {
			continue
		}

		for _, policyType := range []string{"call_policy", "chat_policy", "email_policy", "message_policy"} {
			policies, ok := mediaPolicyMap[policyType].([]interface{})
			if !ok {
				continue
			}
			for _, policy := range policies {
				policyMap, ok := policy.(map[string]interface{})
				if !ok {
					continue
				}
				upgradeTeamIdsInConditionsList(policyMap["conditions"])
			}
		}
	}
}

func upgradeTeamIdsInConditionsList(raw interface{}) {
	conditions, ok := raw.([]interface{})
	if !ok {
		return
	}

	for _, condition := range conditions {
		conditionMap, ok := condition.(map[string]interface{})
		if !ok {
			continue
		}
		conditionMap["team_ids"] = teamIdsValueToSortedList(conditionMap["team_ids"])
	}
}

func teamIdsValueToSortedList(value interface{}) []interface{} {
	switch teamIds := value.(type) {
	case *schema.Set:
		return sortedStringInterfaceList(lists.InterfaceListToStrings(teamIds.List()))
	case []interface{}:
		return sortedStringInterfaceList(lists.InterfaceListToStrings(teamIds))
	case []string:
		return sortedStringInterfaceList(teamIds)
	default:
		return []interface{}{}
	}
}

func sortedStringInterfaceList(values []string) []interface{} {
	sort.Strings(values)
	return lists.StringListToInterfaceList(values)
}
