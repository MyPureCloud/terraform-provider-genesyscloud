package outbound_campaignrule

import (
	"strconv"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/resourcedata"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v178/platformclientv2"
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
	if campaigns, ok := campaignRuleEntitiesMap["campaign_ids"].([]interface{}); ok && campaigns != nil {
		campaignRuleEntities.Campaigns = util.BuildSdkDomainEntityRefArrFromArr(campaigns)
	}
	if sequences, ok := campaignRuleEntitiesMap["sequence_ids"].([]interface{}); ok && sequences != nil {
		campaignRuleEntities.Sequences = util.BuildSdkDomainEntityRefArrFromArr(sequences)
	}
	if smsCampaigns, ok := campaignRuleEntitiesMap["sms_campaign_ids"].([]interface{}); ok && smsCampaigns != nil {
		campaignRuleEntities.SmsCampaigns = util.BuildSdkDomainEntityRefArrFromArr(smsCampaigns)
	}
	if emailCampaigns, ok := campaignRuleEntitiesMap["email_campaign_ids"].([]interface{}); ok && emailCampaigns != nil {
		campaignRuleEntities.EmailCampaigns = util.BuildSdkDomainEntityRefArrFromArr(emailCampaigns)
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

	if abandonRate, ok := paramsMap["abandon_rate"].(string); ok {
		num, err := strconv.ParseFloat(abandonRate, 32)
		if err == nil {
			sdkCampaignRuleParameters.AbandonRate = platformclientv2.Float32(float32(num))
		}
	}
	if lineCount, ok := paramsMap["outbound_line_count"].(string); ok {
		num, err := strconv.Atoi(lineCount)
		if err == nil {
			sdkCampaignRuleParameters.OutboundLineCount = platformclientv2.Int(num)
		}
	}
	if weight, ok := paramsMap["relative_weight"].(string); ok && weight != "" {
		num, err := strconv.Atoi(weight)
		if err == nil {
			sdkCampaignRuleParameters.RelativeWeight = platformclientv2.Int(num)
		}
	}
	if maxCpa, ok := paramsMap["max_calls_per_agent"].(string); ok {
		num, err := strconv.ParseFloat(maxCpa, 32)
		if err == nil {
			sdkCampaignRuleParameters.MaxCallsPerAgent = platformclientv2.Float32(float32(num))
		}
	}
	if messagesPerMinute, ok := paramsMap["messages_per_minute"].(string); ok {
		num, err := strconv.Atoi(messagesPerMinute)
		if err == nil {
			sdkCampaignRuleParameters.MessagesPerMinute = platformclientv2.Int(num)
		}
	}
	if smsMessagesPerMinute, ok := paramsMap["sms_messages_per_minute"].(string); ok {
		num, err := strconv.Atoi(smsMessagesPerMinute)
		if err == nil {
			sdkCampaignRuleParameters.SmsMessagesPerMinute = platformclientv2.Int(num)
		}
	}
	if emailMessagesPerMinute, ok := paramsMap["email_messages_per_minute"].(string); ok {
		num, err := strconv.Atoi(emailMessagesPerMinute)
		if err == nil {
			sdkCampaignRuleParameters.EmailMessagesPerMinute = platformclientv2.Int(num)
		}
	}
	sdkCampaignRuleParameters.Queue = util.GetNillableDomainEntityRefFromMap(paramsMap, "queue_id")
	sdkCampaignRuleParameters.EmailContentTemplate = util.GetNillableDomainEntityRefFromMap(paramsMap, "email_content_template_id")
	sdkCampaignRuleParameters.SmsContentTemplate = util.GetNillableDomainEntityRefFromMap(paramsMap, "sms_content_template_id")

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

	if campaignIds, ok := entitiesMap["campaign_ids"].([]interface{}); ok && campaignIds != nil {
		sdkCampaignRuleActionEntities.Campaigns = util.BuildSdkDomainEntityRefArrFromArr(campaignIds)
	}

	if sequenceIds, ok := entitiesMap["sequence_ids"].([]interface{}); ok && sequenceIds != nil {
		sdkCampaignRuleActionEntities.Sequences = util.BuildSdkDomainEntityRefArrFromArr(sequenceIds)
	}

	if smsCampaignIds, ok := entitiesMap["sms_campaign_ids"].([]interface{}); ok && smsCampaignIds != nil {
		sdkCampaignRuleActionEntities.SmsCampaigns = util.BuildSdkDomainEntityRefArrFromArr(smsCampaignIds)
	}

	if emailCampaignIds, ok := entitiesMap["email_campaign_ids"].([]interface{}); ok && emailCampaignIds != nil {
		sdkCampaignRuleActionEntities.EmailCampaigns = util.BuildSdkDomainEntityRefArrFromArr(emailCampaignIds)
	}

	return &sdkCampaignRuleActionEntities
}

func flattenCampaignRuleEntities(campaignRuleEntities *platformclientv2.Campaignruleentities) *schema.Set {
	var (
		campaignRuleEntitiesSet = schema.NewSet(schema.HashResource(outboundCampaignRuleEntities), []interface{}{})
		campaignRuleEntitiesMap = make(map[string]interface{})

		// had to change from []string to []interface{}
		campaigns      []interface{}
		sequences      []interface{}
		smsCampaigns   []interface{}
		emailCampaigns []interface{}
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

	if campaignRuleEntities.SmsCampaigns != nil {
		for _, v := range *campaignRuleEntities.SmsCampaigns {
			smsCampaigns = append(smsCampaigns, *v.Id)
		}
	}

	if campaignRuleEntities.EmailCampaigns != nil {
		for _, v := range *campaignRuleEntities.EmailCampaigns {
			emailCampaigns = append(emailCampaigns, *v.Id)
		}
	}

	campaignRuleEntitiesMap["campaign_ids"] = campaigns
	campaignRuleEntitiesMap["sequence_ids"] = sequences
	campaignRuleEntitiesMap["sms_campaign_ids"] = smsCampaigns
	campaignRuleEntitiesMap["email_campaign_ids"] = emailCampaigns

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
		campaigns      []interface{}
		sequences      []interface{}
		smsCampaigns   []interface{}
		emailCampaigns []interface{}
		entitiesSet    = schema.NewSet(schema.HashResource(outboundCampaignRuleActionEntities), []interface{}{})
		entitiesMap    = make(map[string]interface{})
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

	if sdkActionEntity.SmsCampaigns != nil {
		for _, campaign := range *sdkActionEntity.SmsCampaigns {
			smsCampaigns = append(smsCampaigns, *campaign.Id)
		}
	}

	if sdkActionEntity.EmailCampaigns != nil {
		for _, campaign := range *sdkActionEntity.EmailCampaigns {
			emailCampaigns = append(emailCampaigns, *campaign.Id)
		}
	}

	entitiesMap["campaign_ids"] = campaigns
	entitiesMap["sequence_ids"] = sequences
	entitiesMap["sms_campaign_ids"] = smsCampaigns
	entitiesMap["email_campaign_ids"] = emailCampaigns
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
	resourcedata.SetMapReferenceValueIfNotNil(paramsMap, "queue_id", params.Queue)
	resourcedata.SetMapReferenceValueIfNotNil(paramsMap, "sms_content_template_id", params.SmsContentTemplate)
	resourcedata.SetMapReferenceValueIfNotNil(paramsMap, "email_content_template_id", params.EmailContentTemplate)

	if params.AbandonRate != nil {
		paramsMap["abandon_rate"] = strconv.FormatFloat(float64(*params.AbandonRate), 'f', -1, 32)
	}
	if params.OutboundLineCount != nil {
		paramsMap["outbound_line_count"] = strconv.Itoa(*params.OutboundLineCount)
	}
	if params.RelativeWeight != nil {
		paramsMap["relative_weight"] = strconv.Itoa(*params.RelativeWeight)
	}
	if params.MaxCallsPerAgent != nil {
		paramsMap["max_calls_per_agent"] = strconv.FormatFloat(float64(*params.MaxCallsPerAgent), 'f', -1, 32)
	}
	if params.MessagesPerMinute != nil {
		paramsMap["messages_per_minute"] = strconv.Itoa(*params.MessagesPerMinute)
	}
	if params.SmsMessagesPerMinute != nil {
		paramsMap["sms_messages_per_minute"] = strconv.Itoa(*params.SmsMessagesPerMinute)
	}
	if params.EmailMessagesPerMinute != nil {
		paramsMap["email_messages_per_minute"] = strconv.Itoa(*params.EmailMessagesPerMinute)
	}

	return []interface{}{paramsMap}
}
