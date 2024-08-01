package outbound_campaignrule

import (
	"terraform-provider-genesyscloud/genesyscloud/util"
	"terraform-provider-genesyscloud/genesyscloud/util/resourcedata"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

func getCampaignruleFromResourceData(d *schema.ResourceData) platformclientv2.Campaignrule {
	matchAnyConditions := d.Get("match_any_conditions").(bool)

	campaignRule := platformclientv2.Campaignrule{
		Name:                   platformclientv2.String(d.Get("name").(string)),
		Enabled:                platformclientv2.Bool(false), // All campaign rules have to be created in an "off" state to start out with
		CampaignRuleEntities:   buildCampaignRuleEntities(d.Get("campaign_rule_entities").(*schema.Set)),
		CampaignRuleConditions: buildCampaignRuleConditions(d.Get("campaign_rule_conditions").([]interface{})),
		CampaignRuleActions:    buildCampaignRuleAction(d.Get("campaign_rule_actions").([]interface{})),
		MatchAnyConditions:     &matchAnyConditions,
	}
	return campaignRule
}

func buildCampaignRuleEntities(entities *schema.Set) *platformclientv2.Campaignruleentities {
	if entities == nil {
		return nil
	}
	var campaignRuleEntities platformclientv2.Campaignruleentities

	campaignRuleEntitiesList := entities.List()

	if len(campaignRuleEntitiesList) <= 0 {
		return &campaignRuleEntities
	}

	campaignRuleEntitiesMap := campaignRuleEntitiesList[0].(map[string]interface{})
	if campaigns := campaignRuleEntitiesMap["campaign_ids"].([]interface{}); campaigns != nil {
		campaignRuleEntities.Campaigns = util.BuildSdkDomainEntityRefArrFromArr(campaigns)
	}
	if sequences := campaignRuleEntitiesMap["sequence_ids"].([]interface{}); sequences != nil {
		campaignRuleEntities.Sequences = util.BuildSdkDomainEntityRefArrFromArr(sequences)
	}
	return &campaignRuleEntities
}

func buildCampaignRuleConditions(campaignRuleConditions []interface{}) *[]platformclientv2.Campaignrulecondition {
	var campaignRuleConditionSlice []platformclientv2.Campaignrulecondition

	for _, campaignRuleCondition := range campaignRuleConditions {
		sdkCondition := platformclientv2.Campaignrulecondition{}
		conditionMap := campaignRuleCondition.(map[string]interface{})

		sdkCondition.Parameters = buildCampaignRuleParameters(conditionMap["parameters"].(*schema.Set))
		resourcedata.BuildSDKStringValueIfNotNil(&sdkCondition.Id, conditionMap, "id")
		resourcedata.BuildSDKStringValueIfNotNil(&sdkCondition.ConditionType, conditionMap, "condition_type")

		campaignRuleConditionSlice = append(campaignRuleConditionSlice, sdkCondition)
	}

	return &campaignRuleConditionSlice
}

func buildCampaignRuleAction(campaignRuleActions []interface{}) *[]platformclientv2.Campaignruleaction {
	var campaignRuleActionSlice []platformclientv2.Campaignruleaction

	for _, campaignRuleAction := range campaignRuleActions {
		var sdkCampaignRuleAction platformclientv2.Campaignruleaction
		actionMap := campaignRuleAction.(map[string]interface{})

		resourcedata.BuildSDKStringValueIfNotNil(&sdkCampaignRuleAction.Id, actionMap, "id")
		resourcedata.BuildSDKStringValueIfNotNil(&sdkCampaignRuleAction.ActionType, actionMap, "action_type")
		sdkCampaignRuleAction.Parameters = buildCampaignRuleParameters(actionMap["parameters"].(*schema.Set))
		sdkCampaignRuleAction.CampaignRuleActionEntities = buildCampaignRuleActionEntities(actionMap["campaign_rule_action_entities"].(*schema.Set))

		campaignRuleActionSlice = append(campaignRuleActionSlice, sdkCampaignRuleAction)
	}

	return &campaignRuleActionSlice
}

func buildCampaignRuleParameters(set *schema.Set) *platformclientv2.Campaignruleparameters {
	var sdkCampaignRuleParameters platformclientv2.Campaignruleparameters

	paramsList := set.List()

	if len(paramsList) <= 0 {
		return &sdkCampaignRuleParameters
	}

	paramsMap := paramsList[0].(map[string]interface{})

	resourcedata.BuildSDKStringValueIfNotNil(&sdkCampaignRuleParameters.Operator, paramsMap, "operator")
	resourcedata.BuildSDKStringValueIfNotNil(&sdkCampaignRuleParameters.Value, paramsMap, "value")
	resourcedata.BuildSDKStringValueIfNotNil(&sdkCampaignRuleParameters.Priority, paramsMap, "priority")
	resourcedata.BuildSDKStringValueIfNotNil(&sdkCampaignRuleParameters.DialingMode, paramsMap, "dialing_mode")

	return &sdkCampaignRuleParameters
}

func buildCampaignRuleActionEntities(set *schema.Set) *platformclientv2.Campaignruleactionentities {
	var (
		sdkCampaignRuleActionEntities platformclientv2.Campaignruleactionentities
		entities                      = set.List()
	)

	if len(entities) <= 0 {
		return &sdkCampaignRuleActionEntities
	}

	entitiesMap := entities[0].(map[string]interface{})

	sdkCampaignRuleActionEntities.UseTriggeringEntity = platformclientv2.Bool(entitiesMap["use_triggering_entity"].(bool))

	if campaignIds := entitiesMap["campaign_ids"].([]interface{}); campaignIds != nil {
		sdkCampaignRuleActionEntities.Campaigns = util.BuildSdkDomainEntityRefArrFromArr(campaignIds)
	}

	if sequenceIds := entitiesMap["sequence_ids"].([]interface{}); sequenceIds != nil {
		sdkCampaignRuleActionEntities.Sequences = util.BuildSdkDomainEntityRefArrFromArr(sequenceIds)
	}

	return &sdkCampaignRuleActionEntities
}

func flattenCampaignRuleEntities(campaignRuleEntities *platformclientv2.Campaignruleentities) *schema.Set {
	var (
		campaignRuleEntitiesSet = schema.NewSet(schema.HashResource(outboundCampaignRuleEntities), []interface{}{})
		campaignRuleEntitiesMap = make(map[string]interface{})

		// had to change from []string to []interface{}
		campaigns []interface{}
		sequences []interface{}
	)

	if campaignRuleEntities == nil {
		return nil
	}

	if campaignRuleEntities.Campaigns != nil {
		for _, v := range *campaignRuleEntities.Campaigns {
			campaigns = append(campaigns, *v.Id)
		}
	}

	if campaignRuleEntities.Sequences != nil {
		for _, v := range *campaignRuleEntities.Sequences {
			sequences = append(sequences, *v.Id)
		}
	}

	campaignRuleEntitiesMap["campaign_ids"] = campaigns
	campaignRuleEntitiesMap["sequence_ids"] = sequences

	campaignRuleEntitiesSet.Add(campaignRuleEntitiesMap)
	return campaignRuleEntitiesSet
}

func flattenCampaignRuleConditions(campaignRuleConditions *[]platformclientv2.Campaignrulecondition) []interface{} {
	if campaignRuleConditions == nil || len(*campaignRuleConditions) == 0 {
		return nil
	}

	var ruleConditionList []interface{}

	for _, currentSdkCondition := range *campaignRuleConditions {
		campaignRuleConditionsMap := make(map[string]interface{})

		resourcedata.SetMapValueIfNotNil(campaignRuleConditionsMap, "id", currentSdkCondition.Id)
		resourcedata.SetMapValueIfNotNil(campaignRuleConditionsMap, "condition_type", currentSdkCondition.ConditionType)
		resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(campaignRuleConditionsMap, "parameters", currentSdkCondition.Parameters, flattenRuleParameters)

		ruleConditionList = append(ruleConditionList, campaignRuleConditionsMap)
	}
	return ruleConditionList
}

func flattenCampaignRuleAction[T any](campaignRuleActions *[]platformclientv2.Campaignruleaction, actionEntitiesFunc func(*platformclientv2.Campaignruleactionentities) T) []interface{} {
	if campaignRuleActions == nil {
		return nil
	}

	var ruleActionsList []interface{}

	for _, currentAction := range *campaignRuleActions {
		actionMap := make(map[string]interface{})

		resourcedata.SetMapValueIfNotNil(actionMap, "id", currentAction.Id)
		resourcedata.SetMapValueIfNotNil(actionMap, "action_type", currentAction.ActionType)
		resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(actionMap, "parameters", currentAction.Parameters, flattenRuleParameters)
		if currentAction.CampaignRuleActionEntities != nil {
			actionMap["campaign_rule_action_entities"] = actionEntitiesFunc(currentAction.CampaignRuleActionEntities)
		}

		ruleActionsList = append(ruleActionsList, actionMap)
	}

	return ruleActionsList
}

func flattenCampaignRuleActionEntities(sdkActionEntity *platformclientv2.Campaignruleactionentities) *schema.Set {
	var (
		campaigns   []interface{}
		sequences   []interface{}
		entitiesSet = schema.NewSet(schema.HashResource(outboundCampaignRuleActionEntities), []interface{}{})
		entitiesMap = make(map[string]interface{})
	)

	if sdkActionEntity == nil {
		return nil
	}

	if sdkActionEntity.Campaigns != nil {
		for _, campaign := range *sdkActionEntity.Campaigns {
			campaigns = append(campaigns, *campaign.Id)
		}
	}

	if sdkActionEntity.Sequences != nil {
		for _, sequence := range *sdkActionEntity.Sequences {
			sequences = append(sequences, *sequence.Id)
		}
	}

	entitiesMap["campaign_ids"] = campaigns
	entitiesMap["sequence_ids"] = sequences
	entitiesMap["use_triggering_entity"] = *sdkActionEntity.UseTriggeringEntity

	entitiesSet.Add(entitiesMap)
	return entitiesSet
}

func flattenRuleParameters(params *platformclientv2.Campaignruleparameters) []interface{} {
	paramsMap := make(map[string]interface{})

	resourcedata.SetMapValueIfNotNil(paramsMap, "operator", params.Operator)
	resourcedata.SetMapValueIfNotNil(paramsMap, "value", params.Value)
	resourcedata.SetMapValueIfNotNil(paramsMap, "priority", params.Priority)
	resourcedata.SetMapValueIfNotNil(paramsMap, "dialing_mode", params.DialingMode)

	return []interface{}{paramsMap}
}
