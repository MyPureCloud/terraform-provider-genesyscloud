package outbound_campaignrule

import (
	"fmt"
	"strconv"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/resourcedata"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v192/platformclientv2"
)

// validateCampaignRuleBeforeAPICall validates that OBR-723 blocks (for_duration, date_time_parameters, etc.)
// are not used in legacy campaign_rule_conditions. This runs at apply-time in Create/Update.
// Uses a separate helper (validateNoTimeBasedConditionsInLegacyFromResourceData) because
// ResourceData returns *schema.Set for TypeSet fields, unlike CustomizeDiff which uses []interface{}.
func validateCampaignRuleBeforeAPICall(d *schema.ResourceData) error {
	processing := d.Get("campaign_rule_processing").(string)
	if processing == "v2" {
		return nil // v2 mode allows all blocks
	}

	// Reuse the same validation as CustomizeDiff for legacy conditions
	conditions := d.Get("campaign_rule_conditions").([]interface{})
	return validateNoTimeBasedConditionsInLegacyFromResourceData(conditions)
}

// validateNoTimeBasedConditionsInLegacyFromResourceData is the apply-time version that handles
// ResourceData types (*schema.Set for parameters) vs CustomizeDiff types ([]interface{}).
func validateNoTimeBasedConditionsInLegacyFromResourceData(conditions []interface{}) error {
	v2OnlyConditionTypes := map[string]bool{
		"timeOfDay":        true,
		"dayOfWeek":        true,
		"dayOfMonth":       true,
		"specificDate":     true,
		"weekDayOfMonth":   true,
		"campaignRunTime":  true,
		"campaignWaitTime": true,
	}

	for i, c := range conditions {
		if c == nil {
			continue
		}
		condMap := c.(map[string]interface{})
		condType, _ := condMap["condition_type"].(string)

		if v2OnlyConditionTypes[condType] {
			return fmt.Errorf("campaign_rule_conditions[%d]: condition_type %q requires campaign_rule_processing = \"v2\" with condition_groups", i, condType)
		}
		if v, ok := condMap["date_time_parameters"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
			return fmt.Errorf("campaign_rule_conditions[%d]: date_time_parameters is only valid with campaign_rule_processing = \"v2\" and condition_groups", i)
		}
		if v, ok := condMap["campaign_run_time_settings"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
			return fmt.Errorf("campaign_rule_conditions[%d]: campaign_run_time_settings is only valid with campaign_rule_processing = \"v2\" and condition_groups", i)
		}
		if v, ok := condMap["campaign_wait_time_settings"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
			return fmt.Errorf("campaign_rule_conditions[%d]: campaign_wait_time_settings is only valid with campaign_rule_processing = \"v2\" and condition_groups", i)
		}

		// Check for_duration in parameters
		params := condMap["parameters"].(*schema.Set)
		if params != nil && params.Len() > 0 {
			paramsMap := params.List()[0].(map[string]interface{})
			if v, ok := paramsMap["for_duration"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
				return fmt.Errorf("campaign_rule_conditions[%d]: for_duration is only valid with campaign_rule_processing = \"v2\" and condition_groups", i)
			}
		}
	}
	return nil
}

func getCampaignruleFromResourceData(d *schema.ResourceData) platformclientv2.Campaignrule {
	matchAnyConditions := d.Get("match_any_conditions").(bool)

	campaignRule := platformclientv2.Campaignrule{
		Name:                 platformclientv2.String(d.Get("name").(string)),
		Enabled:              platformclientv2.Bool(false), // All campaign rules have to be created in an "off" state to start out with
		CampaignRuleEntities: buildCampaignRuleEntities(d.Get("campaign_rule_entities").(*schema.Set)),
		CampaignRuleActions:  buildCampaignRuleAction(d.Get("campaign_rule_actions").([]interface{})),
		MatchAnyConditions:   &matchAnyConditions,
	}

	if v, ok := d.GetOk("campaign_rule_processing"); ok && v.(string) != "" {
		campaignRule.CampaignRuleProcessing = platformclientv2.String(v.(string))
	}
	if v, ok := d.GetOk("condition_groups"); ok && len(v.([]interface{})) > 0 {
		campaignRule.ConditionGroups = buildCampaignRuleConditionGroups(v.([]interface{}))
	}
	if v, ok := d.GetOk("campaign_rule_conditions"); ok && len(v.([]interface{})) > 0 {
		campaignRule.CampaignRuleConditions = buildCampaignRuleConditions(v.([]interface{}))
	}
	if v, ok := d.GetOk("execution_settings"); ok && len(v.([]interface{})) > 0 {
		if firstElement := v.([]interface{})[0]; firstElement != nil {
			campaignRule.ExecutionSettings = buildExecutionSettings(firstElement.(map[string]interface{}))
		}
	}
	if v, ok := d.GetOk("time_zone_id"); ok && v.(string) != "" {
		campaignRule.TimeZoneId = platformclientv2.String(v.(string))
	}

	return campaignRule
}

func buildExecutionSettings(m map[string]interface{}) *platformclientv2.Campaignruleexecutionsettings {
	sdk := &platformclientv2.Campaignruleexecutionsettings{}
	if v, ok := m["frequency"].(string); ok && v != "" {
		sdk.Frequency = platformclientv2.String(v)
	}
	resourcedata.BuildSDKStringValueIfNotNil(&sdk.TimeZoneId, m, "time_zone_id")
	return sdk
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

		if v, ok := conditionMap["date_time_parameters"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
			sdkCondition.DateTimeParameters = buildDateTimeParameters(v[0].(map[string]interface{}))
		}
		if v, ok := conditionMap["campaign_run_time_settings"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
			m := v[0].(map[string]interface{})
			includeWaiting := m["include_waiting_time"].(bool)
			sdkCondition.CampaignRunTimeSettings = &platformclientv2.Campaignrulecampaignruntimesettings{
				IncludeWaitingTime: &includeWaiting,
			}
		}
		if v, ok := conditionMap["campaign_wait_time_settings"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
			m := v[0].(map[string]interface{})
			waitType := m["wait_type"].(string)
			sdkCondition.CampaignWaitTimeSettings = &platformclientv2.Campaignrulecampaignwaittimesettings{
				WaitType: &waitType,
			}
		}

		campaignRuleConditionSlice = append(campaignRuleConditionSlice, sdkCondition)
	}

	return &campaignRuleConditionSlice
}

func buildCampaignRuleConditionGroups(conditionGroups []interface{}) *[]platformclientv2.Campaignruleconditiongroup {
	if len(conditionGroups) == 0 {
		return nil
	}
	var sdkGroups []platformclientv2.Campaignruleconditiongroup
	for _, g := range conditionGroups {
		groupMap := g.(map[string]interface{})
		matchAny := groupMap["match_any_conditions"].(bool)
		conditionsRaw := groupMap["conditions"].([]interface{})
		sdkConditions := buildCampaignRuleConditions(conditionsRaw)
		sdkGroups = append(sdkGroups, platformclientv2.Campaignruleconditiongroup{
			MatchAnyConditions: &matchAny,
			Conditions:         sdkConditions,
		})
	}
	return &sdkGroups
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

	if v, ok := paramsMap["for_duration"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		durationMap := v[0].(map[string]interface{})
		seconds := durationMap["seconds"].(int)
		sdkCampaignRuleParameters.ForDuration = &platformclientv2.Duration{
			Seconds: &seconds,
		}
	}

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

		if currentSdkCondition.DateTimeParameters != nil {
			campaignRuleConditionsMap["date_time_parameters"] = flattenDateTimeParameters(currentSdkCondition.DateTimeParameters)
		}
		if currentSdkCondition.CampaignRunTimeSettings != nil {
			m := make(map[string]interface{})
			resourcedata.SetMapValueIfNotNil(m, "include_waiting_time", currentSdkCondition.CampaignRunTimeSettings.IncludeWaitingTime)
			campaignRuleConditionsMap["campaign_run_time_settings"] = []interface{}{m}
		}
		if currentSdkCondition.CampaignWaitTimeSettings != nil {
			m := make(map[string]interface{})
			resourcedata.SetMapValueIfNotNil(m, "wait_type", currentSdkCondition.CampaignWaitTimeSettings.WaitType)
			campaignRuleConditionsMap["campaign_wait_time_settings"] = []interface{}{m}
		}

		ruleConditionList = append(ruleConditionList, campaignRuleConditionsMap)
	}
	return ruleConditionList
}

func flattenCampaignRuleConditionGroups(conditionGroups *[]platformclientv2.Campaignruleconditiongroup) []interface{} {
	if conditionGroups == nil || len(*conditionGroups) == 0 {
		return nil
	}
	var result []interface{}
	for _, g := range *conditionGroups {
		groupMap := make(map[string]interface{})
		resourcedata.SetMapValueIfNotNil(groupMap, "match_any_conditions", g.MatchAnyConditions)
		groupMap["conditions"] = flattenCampaignRuleConditions(g.Conditions)
		result = append(result, groupMap)
	}
	return result
}

func flattenExecutionSettings(sdk *platformclientv2.Campaignruleexecutionsettings) []interface{} {
	if sdk == nil {
		return nil
	}
	m := make(map[string]interface{})
	resourcedata.SetMapValueIfNotNil(m, "frequency", sdk.Frequency)
	resourcedata.SetMapValueIfNotNil(m, "time_zone_id", sdk.TimeZoneId)
	return []interface{}{m}
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

	if params.ForDuration != nil {
		paramsMap["for_duration"] = flattenForDuration(params.ForDuration)
	}

	return []interface{}{paramsMap}
}

func buildDateTimeParameters(m map[string]interface{}) *platformclientv2.Campaignruledatetimeconditionparameters {
	sdk := &platformclientv2.Campaignruledatetimeconditionparameters{}

	if v, ok := m["inverted"].(bool); ok {
		sdk.Inverted = &v
	}

	if v, ok := m["time_of_day"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		todMap := v[0].(map[string]interface{})
		tod := &platformclientv2.Campaignruletimeofdayparameters{}
		resourcedata.BuildSDKStringValueIfNotNil(&tod.ThresholdValue, todMap, "threshold_value")
		if interval, ok := todMap["interval"].([]interface{}); ok && len(interval) > 0 && interval[0] != nil {
			iMap := interval[0].(map[string]interface{})
			min := iMap["min"].(string)
			max := iMap["max"].(string)
			tod.Interval = &platformclientv2.Campaignruletimeofdayinterval{Min: &min, Max: &max}
		}
		sdk.TimeOfDay = tod
	}

	if v, ok := m["day_of_week"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		dowMap := v[0].(map[string]interface{})
		dow := &platformclientv2.Campaignruledayofweekparameters{}
		if inSet, ok := dowMap["in_set"].([]interface{}); ok && len(inSet) > 0 {
			ints := make([]int, len(inSet))
			for i, val := range inSet {
				ints[i] = val.(int)
			}
			dow.InSet = &ints
		}
		if interval, ok := dowMap["interval"].([]interface{}); ok && len(interval) > 0 && interval[0] != nil {
			iMap := interval[0].(map[string]interface{})
			min := iMap["min"].(int)
			max := iMap["max"].(int)
			dow.Interval = &platformclientv2.Campaignruledayofweekinterval{Min: &min, Max: &max}
		}
		sdk.DayOfWeek = dow
	}

	if v, ok := m["day_of_month"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		domMap := v[0].(map[string]interface{})
		dom := &platformclientv2.Campaignruledayofmonthparameters{}
		resourcedata.BuildSDKStringValueIfNotNil(&dom.ThresholdValue, domMap, "threshold_value")
		if inSet, ok := domMap["in_set"].([]interface{}); ok && len(inSet) > 0 {
			strs := make([]string, len(inSet))
			for i, val := range inSet {
				strs[i] = val.(string)
			}
			dom.InSet = &strs
		}
		if interval, ok := domMap["interval"].([]interface{}); ok && len(interval) > 0 && interval[0] != nil {
			iMap := interval[0].(map[string]interface{})
			min := iMap["min"].(string)
			max := iMap["max"].(string)
			dom.Interval = &platformclientv2.Campaignruledayofmonthinterval{Min: &min, Max: &max}
		}
		sdk.DayOfMonth = dom
	}

	if v, ok := m["specific_date"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		sdMap := v[0].(map[string]interface{})
		sd := &platformclientv2.Campaignrulespecificdateparameters{}
		if includeYear, ok := sdMap["include_year"].(bool); ok {
			sd.IncludeYear = &includeYear
		}
		resourcedata.BuildSDKStringValueIfNotNil(&sd.ThresholdValue, sdMap, "threshold_value")
		if interval, ok := sdMap["interval"].([]interface{}); ok && len(interval) > 0 && interval[0] != nil {
			iMap := interval[0].(map[string]interface{})
			min := iMap["min"].(string)
			max := iMap["max"].(string)
			sd.Interval = &platformclientv2.Campaignrulespecificdateinterval{Min: &min, Max: &max}
		}
		sdk.SpecificDate = sd
	}

	if v, ok := m["week_day_of_month"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		wdomMap := v[0].(map[string]interface{})
		wdom := &platformclientv2.Campaignruleweekdayofmonthparameters{}
		if tv, ok := wdomMap["threshold_value"].([]interface{}); ok && len(tv) > 0 && tv[0] != nil {
			wdom.ThresholdValue = buildWeekDayOfMonth(tv[0].(map[string]interface{}))
		}
		if interval, ok := wdomMap["interval"].([]interface{}); ok && len(interval) > 0 && interval[0] != nil {
			iMap := interval[0].(map[string]interface{})
			sdkInterval := &platformclientv2.Campaignruleweekdayofmonthinterval{}
			if minList, ok := iMap["min"].([]interface{}); ok && len(minList) > 0 && minList[0] != nil {
				sdkInterval.Min = buildWeekDayOfMonth(minList[0].(map[string]interface{}))
			}
			if maxList, ok := iMap["max"].([]interface{}); ok && len(maxList) > 0 && maxList[0] != nil {
				sdkInterval.Max = buildWeekDayOfMonth(maxList[0].(map[string]interface{}))
			}
			wdom.Interval = sdkInterval
		}
		sdk.WeekDayOfMonth = wdom
	}

	return sdk
}

func buildWeekDayOfMonth(m map[string]interface{}) *platformclientv2.Campaignruleweekdayofmonth {
	sdk := &platformclientv2.Campaignruleweekdayofmonth{}
	if v, ok := m["day_of_week"].(int); ok && v != 0 {
		sdk.DayOfWeek = &v
	}
	if v, ok := m["month"].(int); ok && v != 0 {
		sdk.Month = &v
	}
	if v, ok := m["occurrence"].(int); ok && v != 0 {
		sdk.Occurrence = &v
	}
	return sdk
}

func flattenDateTimeParameters(sdk *platformclientv2.Campaignruledatetimeconditionparameters) []interface{} {
	if sdk == nil {
		return nil
	}
	m := make(map[string]interface{})
	resourcedata.SetMapValueIfNotNil(m, "inverted", sdk.Inverted)

	if sdk.TimeOfDay != nil {
		tod := make(map[string]interface{})
		resourcedata.SetMapValueIfNotNil(tod, "threshold_value", sdk.TimeOfDay.ThresholdValue)
		if sdk.TimeOfDay.Interval != nil {
			tod["interval"] = []interface{}{map[string]interface{}{
				"min": nilToEmpty(sdk.TimeOfDay.Interval.Min),
				"max": nilToEmpty(sdk.TimeOfDay.Interval.Max),
			}}
		}
		m["time_of_day"] = []interface{}{tod}
	}

	if sdk.DayOfWeek != nil {
		dow := make(map[string]interface{})
		if sdk.DayOfWeek.InSet != nil {
			dow["in_set"] = *sdk.DayOfWeek.InSet
		}
		if sdk.DayOfWeek.Interval != nil {
			dow["interval"] = []interface{}{map[string]interface{}{
				"min": derefInt(sdk.DayOfWeek.Interval.Min),
				"max": derefInt(sdk.DayOfWeek.Interval.Max),
			}}
		}
		m["day_of_week"] = []interface{}{dow}
	}

	if sdk.DayOfMonth != nil {
		dom := make(map[string]interface{})
		resourcedata.SetMapValueIfNotNil(dom, "threshold_value", sdk.DayOfMonth.ThresholdValue)
		if sdk.DayOfMonth.InSet != nil {
			dom["in_set"] = *sdk.DayOfMonth.InSet
		}
		if sdk.DayOfMonth.Interval != nil {
			dom["interval"] = []interface{}{map[string]interface{}{
				"min": nilToEmpty(sdk.DayOfMonth.Interval.Min),
				"max": nilToEmpty(sdk.DayOfMonth.Interval.Max),
			}}
		}
		m["day_of_month"] = []interface{}{dom}
	}

	if sdk.SpecificDate != nil {
		sd := make(map[string]interface{})
		resourcedata.SetMapValueIfNotNil(sd, "include_year", sdk.SpecificDate.IncludeYear)
		resourcedata.SetMapValueIfNotNil(sd, "threshold_value", sdk.SpecificDate.ThresholdValue)
		if sdk.SpecificDate.Interval != nil {
			sd["interval"] = []interface{}{map[string]interface{}{
				"min": nilToEmpty(sdk.SpecificDate.Interval.Min),
				"max": nilToEmpty(sdk.SpecificDate.Interval.Max),
			}}
		}
		m["specific_date"] = []interface{}{sd}
	}

	if sdk.WeekDayOfMonth != nil {
		wdom := make(map[string]interface{})
		if sdk.WeekDayOfMonth.ThresholdValue != nil {
			wdom["threshold_value"] = []interface{}{flattenWeekDayOfMonth(sdk.WeekDayOfMonth.ThresholdValue)}
		}
		if sdk.WeekDayOfMonth.Interval != nil {
			intervalMap := make(map[string]interface{})
			if sdk.WeekDayOfMonth.Interval.Min != nil {
				intervalMap["min"] = []interface{}{flattenWeekDayOfMonth(sdk.WeekDayOfMonth.Interval.Min)}
			}
			if sdk.WeekDayOfMonth.Interval.Max != nil {
				intervalMap["max"] = []interface{}{flattenWeekDayOfMonth(sdk.WeekDayOfMonth.Interval.Max)}
			}
			wdom["interval"] = []interface{}{intervalMap}
		}
		m["week_day_of_month"] = []interface{}{wdom}
	}

	return []interface{}{m}
}

func flattenWeekDayOfMonth(sdk *platformclientv2.Campaignruleweekdayofmonth) map[string]interface{} {
	m := make(map[string]interface{})
	if sdk.DayOfWeek != nil {
		m["day_of_week"] = *sdk.DayOfWeek
	}
	if sdk.Month != nil {
		m["month"] = *sdk.Month
	}
	if sdk.Occurrence != nil {
		m["occurrence"] = *sdk.Occurrence
	}
	return m
}

func flattenForDuration(sdk *platformclientv2.Duration) []interface{} {
	if sdk == nil {
		return nil
	}
	m := make(map[string]interface{})
	if sdk.Seconds != nil {
		m["seconds"] = *sdk.Seconds
	}
	return []interface{}{m}
}

func nilToEmpty(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func derefInt(i *int) int {
	if i == nil {
		return 0
	}
	return *i
}
