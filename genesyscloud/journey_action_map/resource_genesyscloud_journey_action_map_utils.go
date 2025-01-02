package journey_action_map

import (
	"fmt"
	"log"
	"strings"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	lists "terraform-provider-genesyscloud/genesyscloud/util/lists"
	"terraform-provider-genesyscloud/genesyscloud/util/resourcedata"
	"terraform-provider-genesyscloud/genesyscloud/util/stringmap"
	"terraform-provider-genesyscloud/genesyscloud/util/testrunner"
	"terraform-provider-genesyscloud/genesyscloud/util/typeconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v150/platformclientv2"
)

func flattenActionMap(d *schema.ResourceData, actionMap *platformclientv2.Actionmap) {
	d.Set("is_active", *actionMap.IsActive)
	d.Set("display_name", *actionMap.DisplayName)
	d.Set("trigger_with_segments", lists.StringListToSetOrNil(actionMap.TriggerWithSegments))
	resourcedata.SetNillableValue(d, "trigger_with_event_conditions", lists.FlattenList(actionMap.TriggerWithEventConditions, flattenEventCondition))
	resourcedata.SetNillableValue(d, "trigger_with_outcome_probability_conditions", lists.FlattenList(actionMap.TriggerWithOutcomeProbabilityConditions, flattenOutcomeProbabilityCondition))
	resourcedata.SetNillableValue(d, "trigger_with_outcome_quantile_conditions", lists.FlattenList(actionMap.TriggerWithOutcomeQuantileConditions, flattenOutcomeQuantileCondition))
	resourcedata.SetNillableValue(d, "page_url_conditions", lists.FlattenList(actionMap.PageUrlConditions, flattenUrlCondition))
	d.Set("activation", lists.FlattenAsList(actionMap.Activation, flattenActivation))
	d.Set("weight", *actionMap.Weight)
	resourcedata.SetNillableValue(d, "action", lists.FlattenAsList(actionMap.Action, flattenActionMapAction))
	resourcedata.SetNillableValue(d, "action_map_schedule_groups", lists.FlattenAsList(actionMap.ActionMapScheduleGroups, flattenActionMapScheduleGroups))
	d.Set("ignore_frequency_cap", *actionMap.IgnoreFrequencyCap)
	resourcedata.SetNillableTime(d, "start_date", actionMap.StartDate)
	resourcedata.SetNillableTime(d, "end_date", actionMap.EndDate)
}

func buildSdkActionMap(actionMap *schema.ResourceData) *platformclientv2.Actionmap {
	isActive := actionMap.Get("is_active").(bool)
	displayName := actionMap.Get("display_name").(string)
	triggerWithSegments := lists.BuildSdkStringList(actionMap, "trigger_with_segments")
	triggerWithEventConditions := resourcedata.BuildSdkList(actionMap, "trigger_with_event_conditions", buildSdkEventCondition)
	triggerWithOutcomeProbabilityConditions := resourcedata.BuildSdkList(actionMap, "trigger_with_outcome_probability_conditions", buildSdkOutcomeProbabilityCondition)
	triggerWithOutcomeQuantileConditions := resourcedata.BuildSdkList(actionMap, "trigger_with_outcome_quantile_conditions", buildSdkOutcomeQuantileCondition)
	pageUrlConditions := resourcedata.BuildSdkList(actionMap, "page_url_conditions", buildSdkUrlCondition)
	activation := resourcedata.BuildSdkListFirstElement(actionMap, "activation", buildSdkActivation, true)
	weight := actionMap.Get("weight").(int)
	action := resourcedata.BuildSdkListFirstElement(actionMap, "action", buildSdkActionMapAction, true)
	actionMapScheduleGroups := resourcedata.BuildSdkListFirstElement(actionMap, "action_map_schedule_groups", buildSdkActionMapScheduleGroups, true)
	ignoreFrequencyCap := actionMap.Get("ignore_frequency_cap").(bool)
	startDate := resourcedata.GetNillableTime(actionMap, "start_date")
	endDate := resourcedata.GetNillableTime(actionMap, "end_date")

	return &platformclientv2.Actionmap{
		IsActive:                                &isActive,
		DisplayName:                             &displayName,
		TriggerWithSegments:                     triggerWithSegments,
		TriggerWithEventConditions:              triggerWithEventConditions,
		TriggerWithOutcomeProbabilityConditions: triggerWithOutcomeProbabilityConditions,
		TriggerWithOutcomeQuantileConditions:    triggerWithOutcomeQuantileConditions,
		PageUrlConditions:                       pageUrlConditions,
		Activation:                              activation,
		Weight:                                  &weight,
		Action:                                  action,
		ActionMapScheduleGroups:                 actionMapScheduleGroups,
		IgnoreFrequencyCap:                      &ignoreFrequencyCap,
		StartDate:                               startDate,
		EndDate:                                 endDate,
	}
}

func buildSdkPatchActionMap(patchActionMap *schema.ResourceData) *platformclientv2.Patchactionmap {
	isActive := patchActionMap.Get("is_active").(bool)
	displayName := patchActionMap.Get("display_name").(string)
	triggerWithSegments := lists.BuildSdkStringList(patchActionMap, "trigger_with_segments")
	triggerWithEventConditions := lists.NilToEmptyList(resourcedata.BuildSdkList(patchActionMap, "trigger_with_event_conditions", buildSdkEventCondition))
	triggerWithOutcomeProbabilityConditions := lists.NilToEmptyList(resourcedata.BuildSdkList(patchActionMap, "trigger_with_outcome_probability_conditions", buildSdkOutcomeProbabilityCondition))
	triggerWithOutcomeQuantileConditions := lists.NilToEmptyList(resourcedata.BuildSdkList(patchActionMap, "trigger_with_outcome_quantile_conditions", buildSdkOutcomeQuantileCondition))
	pageUrlConditions := lists.NilToEmptyList(resourcedata.BuildSdkList(patchActionMap, "page_url_conditions", buildSdkUrlCondition))
	activation := resourcedata.BuildSdkListFirstElement(patchActionMap, "activation", buildSdkActivation, true)
	weight := patchActionMap.Get("weight").(int)
	action := resourcedata.BuildSdkListFirstElement(patchActionMap, "action", buildSdkPatchAction, true)
	actionMapScheduleGroups := resourcedata.BuildSdkListFirstElement(patchActionMap, "action_map_schedule_groups", buildSdkPatchActionMapScheduleGroups, true)
	ignoreFrequencyCap := patchActionMap.Get("ignore_frequency_cap").(bool)
	startDate := resourcedata.GetNillableTime(patchActionMap, "start_date")
	endDate := resourcedata.GetNillableTime(patchActionMap, "end_date")

	sdkPatchActionMap := platformclientv2.Patchactionmap{}
	sdkPatchActionMap.SetField("IsActive", &isActive)
	sdkPatchActionMap.SetField("DisplayName", &displayName)
	sdkPatchActionMap.SetField("TriggerWithSegments", triggerWithSegments)
	sdkPatchActionMap.SetField("TriggerWithEventConditions", triggerWithEventConditions)
	sdkPatchActionMap.SetField("TriggerWithOutcomeProbabilityConditions", triggerWithOutcomeProbabilityConditions)
	sdkPatchActionMap.SetField("TriggerWithOutcomeQuantileConditions", triggerWithOutcomeQuantileConditions)
	sdkPatchActionMap.SetField("PageUrlConditions", pageUrlConditions)
	sdkPatchActionMap.SetField("Activation", activation)
	sdkPatchActionMap.SetField("Weight", &weight)
	sdkPatchActionMap.SetField("Action", action)
	sdkPatchActionMap.SetField("ActionMapScheduleGroups", actionMapScheduleGroups)
	sdkPatchActionMap.SetField("IgnoreFrequencyCap", &ignoreFrequencyCap)
	sdkPatchActionMap.SetField("StartDate", startDate)
	sdkPatchActionMap.SetField("EndDate", endDate)
	return &sdkPatchActionMap
}

func flattenEventCondition(eventCondition *platformclientv2.Eventcondition) map[string]interface{} {
	eventConditionMap := make(map[string]interface{})
	eventConditionMap["key"] = *eventCondition.Key
	eventConditionMap["values"] = lists.StringListToSet(*eventCondition.Values)
	eventConditionMap["operator"] = *eventCondition.Operator
	eventConditionMap["stream_type"] = *eventCondition.StreamType
	eventConditionMap["session_type"] = *eventCondition.SessionType
	stringmap.SetValueIfNotNil(eventConditionMap, "event_name", eventCondition.EventName)
	return eventConditionMap
}

func buildSdkEventCondition(eventCondition map[string]interface{}) *platformclientv2.Eventcondition {
	key := eventCondition["key"].(string)
	values := stringmap.BuildSdkStringList(eventCondition, "values")
	operator := eventCondition["operator"].(string)
	streamType := eventCondition["stream_type"].(string)
	sessionType := eventCondition["session_type"].(string)
	eventName := stringmap.GetNonDefaultValue[string](eventCondition, "event_name")

	return &platformclientv2.Eventcondition{
		Key:         &key,
		Values:      values,
		Operator:    &operator,
		StreamType:  &streamType,
		SessionType: &sessionType,
		EventName:   eventName,
	}
}

func flattenOutcomeProbabilityCondition(outcomeProbabilityCondition *platformclientv2.Outcomeprobabilitycondition) map[string]interface{} {
	outcomeProbabilityConditionMap := make(map[string]interface{})
	outcomeProbabilityConditionMap["outcome_id"] = *outcomeProbabilityCondition.OutcomeId
	outcomeProbabilityConditionMap["maximum_probability"] = *typeconv.Float32to64(outcomeProbabilityCondition.MaximumProbability)
	stringmap.SetValueIfNotNil(outcomeProbabilityConditionMap, "probability", typeconv.Float32to64(outcomeProbabilityCondition.Probability))
	return outcomeProbabilityConditionMap
}

func buildSdkOutcomeProbabilityCondition(outcomeProbabilityCondition map[string]interface{}) *platformclientv2.Outcomeprobabilitycondition {
	outcomeId := outcomeProbabilityCondition["outcome_id"].(string)
	maximumProbability64 := outcomeProbabilityCondition["maximum_probability"].(float64)
	maximumProbability := typeconv.Float64to32(&maximumProbability64)
	probability := typeconv.Float64to32(stringmap.GetNonDefaultValue[float64](outcomeProbabilityCondition, "probability"))

	return &platformclientv2.Outcomeprobabilitycondition{
		OutcomeId:          &outcomeId,
		MaximumProbability: maximumProbability,
		Probability:        probability,
	}
}

func flattenOutcomeQuantileCondition(outcomeQuantileCondition *platformclientv2.Outcomequantilecondition) map[string]interface{} {
	outcomeQuantileConditionMap := make(map[string]interface{})
	outcomeQuantileConditionMap["outcome_id"] = *outcomeQuantileCondition.OutcomeId
	outcomeQuantileConditionMap["max_quantile_threshold"] = *typeconv.Float32to64(outcomeQuantileCondition.MaxQuantileThreshold)
	stringmap.SetValueIfNotNil(outcomeQuantileConditionMap, "fallback_quantile_threshold", typeconv.Float32to64(outcomeQuantileCondition.FallbackQuantileThreshold))
	return outcomeQuantileConditionMap
}

func buildSdkOutcomeQuantileCondition(outcomeQuantileCondition map[string]interface{}) *platformclientv2.Outcomequantilecondition {
	outcomeId := outcomeQuantileCondition["outcome_id"].(string)
	maxQuantileThreshold64 := outcomeQuantileCondition["max_quantile_threshold"].(float64)
	maxQuantileThreshold := typeconv.Float64to32(&maxQuantileThreshold64)
	fallbackQuantileThreshold := typeconv.Float64to32(stringmap.GetNonDefaultValue[float64](outcomeQuantileCondition, "fallback_quantile_threshold"))

	return &platformclientv2.Outcomequantilecondition{
		OutcomeId:                 &outcomeId,
		MaxQuantileThreshold:      maxQuantileThreshold,
		FallbackQuantileThreshold: fallbackQuantileThreshold,
	}
}

func flattenUrlCondition(urlCondition *platformclientv2.Urlcondition) map[string]interface{} {
	urlConditionMap := make(map[string]interface{})
	urlConditionMap["values"] = lists.StringListToSet(*urlCondition.Values)
	urlConditionMap["operator"] = *urlCondition.Operator
	return urlConditionMap
}

func buildSdkUrlCondition(eventCondition map[string]interface{}) *platformclientv2.Urlcondition {
	values := stringmap.BuildSdkStringList(eventCondition, "values")
	operator := eventCondition["operator"].(string)

	return &platformclientv2.Urlcondition{
		Values:   values,
		Operator: &operator,
	}
}

func flattenActivation(activation *platformclientv2.Activation) map[string]interface{} {
	activationMap := make(map[string]interface{})
	activationMap["type"] = *activation.VarType
	stringmap.SetValueIfNotNil(activationMap, "delay_in_seconds", activation.DelayInSeconds)
	return activationMap
}

func buildSdkActivation(activation map[string]interface{}) *platformclientv2.Activation {
	varType := activation["type"].(string)
	delayInSeconds := stringmap.GetNonDefaultValue[int](activation, "delay_in_seconds")

	return &platformclientv2.Activation{
		VarType:        &varType,
		DelayInSeconds: delayInSeconds,
	}
}

func flattenActionMapAction(actionMapAction *platformclientv2.Actionmapaction) map[string]interface{} {
	actionMapActionMap := make(map[string]interface{})
	actionMapActionMap["media_type"] = *actionMapAction.MediaType
	actionMapActionMap["is_pacing_enabled"] = *actionMapAction.IsPacingEnabled
	if actionMapAction.ActionTemplate != nil {
		stringmap.SetValueIfNotNil(actionMapActionMap, "action_template_id", actionMapAction.ActionTemplate.Id)
	}
	stringmap.SetValueIfNotNil(actionMapActionMap, "architect_flow_fields", lists.FlattenAsList(actionMapAction.ArchitectFlowFields, flattenArchitectFlowFields))
	stringmap.SetValueIfNotNil(actionMapActionMap, "web_messaging_offer_fields", lists.FlattenAsList(actionMapAction.WebMessagingOfferFields, flattenWebMessagingOfferFields))
	stringmap.SetValueIfNotNil(actionMapActionMap, "open_action_fields", lists.FlattenAsList(actionMapAction.OpenActionFields, flattenOpenActionFields))
	return actionMapActionMap
}

func buildSdkActionMapAction(actionMapAction map[string]interface{}) *platformclientv2.Actionmapaction {
	mediaType := actionMapAction["media_type"].(string)
	isPacingEnabled := actionMapAction["is_pacing_enabled"].(bool)
	actionMapActionTemplate := getActionMapActionTemplate(actionMapAction)
	architectFlowFields := stringmap.BuildSdkListFirstElement(actionMapAction, "architect_flow_fields", buildSdkArchitectFlowFields, true)
	webMessagingOfferFields := stringmap.BuildSdkListFirstElement(actionMapAction, "web_messaging_offer_fields", buildSdkWebMessagingOfferFields, true)
	openActionFields := stringmap.BuildSdkListFirstElement(actionMapAction, "open_action_fields", buildSdkOpenActionFields, true)

	return &platformclientv2.Actionmapaction{
		MediaType:               &mediaType,
		IsPacingEnabled:         &isPacingEnabled,
		ActionTemplate:          actionMapActionTemplate,
		ArchitectFlowFields:     architectFlowFields,
		WebMessagingOfferFields: webMessagingOfferFields,
		OpenActionFields:        openActionFields,
	}
}

func buildSdkPatchAction(patchAction map[string]interface{}) *platformclientv2.Patchaction {
	mediaType := patchAction["media_type"].(string)
	isPacingEnabled := patchAction["is_pacing_enabled"].(bool)
	actionMapActionTemplate := getActionMapActionTemplate(patchAction)
	architectFlowFields := stringmap.BuildSdkListFirstElement(patchAction, "architect_flow_fields", buildSdkArchitectFlowFields, true)
	webMessagingOfferFields := stringmap.BuildSdkListFirstElement(patchAction, "web_messaging_offer_fields", buildSdkPatchWebMessagingOfferFields, true)
	openActionFields := stringmap.BuildSdkListFirstElement(patchAction, "open_action_fields", buildSdkOpenActionFields, true)

	sdkPatchAction := platformclientv2.Patchaction{}
	sdkPatchAction.SetField("MediaType", &mediaType)
	sdkPatchAction.SetField("IsPacingEnabled", &isPacingEnabled)
	sdkPatchAction.SetField("ActionTemplate", actionMapActionTemplate)
	sdkPatchAction.SetField("ArchitectFlowFields", architectFlowFields)
	sdkPatchAction.SetField("WebMessagingOfferFields", webMessagingOfferFields)
	sdkPatchAction.SetField("OpenActionFields", openActionFields)
	return &sdkPatchAction
}

func getActionMapActionTemplate(actionMapAction map[string]interface{}) *platformclientv2.Actionmapactiontemplate {
	actionMapActionTemplateId := stringmap.GetNonDefaultValue[string](actionMapAction, "action_template_id")
	var actionMapActionTemplate *platformclientv2.Actionmapactiontemplate = nil
	if actionMapActionTemplateId != nil {
		actionMapActionTemplate = &platformclientv2.Actionmapactiontemplate{
			Id: actionMapActionTemplateId,
		}
	}
	return actionMapActionTemplate
}

func flattenArchitectFlowFields(architectFlowFields *platformclientv2.Architectflowfields) map[string]interface{} {
	architectFlowFieldsMap := make(map[string]interface{})
	architectFlowFieldsMap["architect_flow_id"] = *architectFlowFields.ArchitectFlow.Id
	stringmap.SetValueIfNotNil(architectFlowFieldsMap, "flow_request_mappings", lists.FlattenList(architectFlowFields.FlowRequestMappings, flattenRequestMapping))
	return architectFlowFieldsMap
}

func buildSdkArchitectFlowFields(architectFlowFields map[string]interface{}) *platformclientv2.Architectflowfields {
	architectFlow := getArchitectFlow(architectFlowFields)
	flowRequestMappings := stringmap.BuildSdkList(architectFlowFields, "flow_request_mappings", buildSdkRequestMapping)

	return &platformclientv2.Architectflowfields{
		ArchitectFlow:       architectFlow,
		FlowRequestMappings: flowRequestMappings,
	}
}

func flattenRequestMapping(requestMapping *platformclientv2.Requestmapping) map[string]interface{} {
	requestMappingMap := make(map[string]interface{})
	requestMappingMap["name"] = *requestMapping.Name
	requestMappingMap["attribute_type"] = *requestMapping.AttributeType
	requestMappingMap["mapping_type"] = *requestMapping.MappingType
	requestMappingMap["value"] = *requestMapping.Value
	return requestMappingMap
}

func buildSdkRequestMapping(RequestMapping map[string]interface{}) *platformclientv2.Requestmapping {
	name := RequestMapping["name"].(string)
	attributeType := RequestMapping["attribute_type"].(string)
	mappingType := RequestMapping["mapping_type"].(string)
	value := RequestMapping["value"].(string)

	return &platformclientv2.Requestmapping{
		Name:          &name,
		AttributeType: &attributeType,
		MappingType:   &mappingType,
		Value:         &value,
	}
}

func flattenWebMessagingOfferFields(webMessagingOfferFields *platformclientv2.Webmessagingofferfields) map[string]interface{} {
	webMessagingOfferFieldsMap := make(map[string]interface{})
	if webMessagingOfferFields.OfferText == nil && (webMessagingOfferFields.ArchitectFlow == nil || webMessagingOfferFields.ArchitectFlow.Id == nil) {
		return nil
	}
	stringmap.SetValueIfNotNil(webMessagingOfferFieldsMap, "offer_text", webMessagingOfferFields.OfferText)
	if webMessagingOfferFields.ArchitectFlow != nil {
		stringmap.SetValueIfNotNil(webMessagingOfferFieldsMap, "architect_flow_id", webMessagingOfferFields.ArchitectFlow.Id)
	}
	return webMessagingOfferFieldsMap
}

func buildSdkWebMessagingOfferFields(webMessagingOfferFields map[string]interface{}) *platformclientv2.Webmessagingofferfields {
	offerText := stringmap.GetNonDefaultValue[string](webMessagingOfferFields, "offer_text")
	architectFlow := getArchitectFlow(webMessagingOfferFields)

	return &platformclientv2.Webmessagingofferfields{
		OfferText:     offerText,
		ArchitectFlow: architectFlow,
	}
}

func buildSdkPatchWebMessagingOfferFields(webMessagingOfferFields map[string]interface{}) *platformclientv2.Patchwebmessagingofferfields {
	offerText := stringmap.GetNonDefaultValue[string](webMessagingOfferFields, "offer_text")
	architectFlow := getArchitectFlow(webMessagingOfferFields)

	return &platformclientv2.Patchwebmessagingofferfields{
		OfferText:     offerText,
		ArchitectFlow: architectFlow,
	}
}

func getArchitectFlow(actionMapAction map[string]interface{}) *platformclientv2.Addressableentityref {
	architectFlowId := stringmap.GetNonDefaultValue[string](actionMapAction, "architect_flow_id")
	var architectFlow *platformclientv2.Addressableentityref = nil
	if architectFlowId != nil {
		architectFlow = &platformclientv2.Addressableentityref{
			Id: architectFlowId,
		}
	}
	return architectFlow
}

func flattenOpenActionFields(openActionFields *platformclientv2.Openactionfields) map[string]interface{} {
	architectFlowFieldsMap := make(map[string]interface{})
	architectFlowFieldsMap["open_action"] = lists.FlattenAsList(openActionFields.OpenAction, flattenOpenActionDomainEntityRef)
	if openActionFields.ConfigurationFields != nil {
		jsonString, err := util.InterfaceToJson(openActionFields.ConfigurationFields)
		if err != nil {
			log.Printf("Error marshalling '%s': %v", "configuration_fields", err)
		}
		architectFlowFieldsMap["configuration_fields"] = jsonString
	}
	return architectFlowFieldsMap
}

func buildSdkOpenActionFields(openActionFieldsMap map[string]interface{}) *platformclientv2.Openactionfields {
	openAction := stringmap.BuildSdkListFirstElement(openActionFieldsMap, "open_action", buildSdkOpenActionDomainEntityRef, true)
	openActionFields := platformclientv2.Openactionfields{
		OpenAction: openAction,
	}

	configurationFieldsString := stringmap.GetNonDefaultValue[string](openActionFieldsMap, "configuration_fields")
	if configurationFieldsString != nil {
		configurationFieldsInterface, err := util.JsonStringToInterface(*configurationFieldsString)
		if err != nil {
			log.Printf("Error unmarshalling '%s': %v", "configuration_fields", err)
		}
		configurationFieldsMap := configurationFieldsInterface.(map[string]interface{})
		openActionFields.ConfigurationFields = &configurationFieldsMap
	}

	return &openActionFields
}

func flattenOpenActionDomainEntityRef(domainEntityRef *platformclientv2.Domainentityref) map[string]interface{} {
	domainEntityRefMap := make(map[string]interface{})
	domainEntityRefMap["id"] = *domainEntityRef.Id
	domainEntityRefMap["name"] = *domainEntityRef.Name
	return domainEntityRefMap
}

func buildSdkOpenActionDomainEntityRef(domainEntityRef map[string]interface{}) *platformclientv2.Domainentityref {
	id := domainEntityRef["id"].(string)
	name := domainEntityRef["name"].(string)

	return &platformclientv2.Domainentityref{
		Id:   &id,
		Name: &name,
	}
}

func flattenActionMapScheduleGroups(actionMapScheduleGroups *platformclientv2.Actionmapschedulegroups) map[string]interface{} {
	actionMapScheduleGroupsMap := make(map[string]interface{})
	actionMapScheduleGroupsMap["action_map_schedule_group_id"] = *actionMapScheduleGroups.ActionMapScheduleGroup.Id
	if actionMapScheduleGroups.EmergencyActionMapScheduleGroup != nil {
		stringmap.SetValueIfNotNil(actionMapScheduleGroupsMap, "emergency_action_map_schedule_group_id", actionMapScheduleGroups.EmergencyActionMapScheduleGroup.Id)
	}
	return actionMapScheduleGroupsMap
}

func buildSdkActionMapScheduleGroups(actionMapScheduleGroups map[string]interface{}) *platformclientv2.Actionmapschedulegroups {
	actionMapScheduleGroup, emergencyActionMapScheduleGroup := getActionMapScheduleGroupPair(actionMapScheduleGroups)

	return &platformclientv2.Actionmapschedulegroups{
		ActionMapScheduleGroup:          actionMapScheduleGroup,
		EmergencyActionMapScheduleGroup: emergencyActionMapScheduleGroup,
	}
}

func buildSdkPatchActionMapScheduleGroups(actionMapScheduleGroups map[string]interface{}) *platformclientv2.Patchactionmapschedulegroups {
	if actionMapScheduleGroups == nil {
		return nil
	}

	actionMapScheduleGroup, emergencyActionMapScheduleGroup := getActionMapScheduleGroupPair(actionMapScheduleGroups)

	sdkPatchActionMapScheduleGroups := platformclientv2.Patchactionmapschedulegroups{}
	sdkPatchActionMapScheduleGroups.SetField("ActionMapScheduleGroup", actionMapScheduleGroup)
	sdkPatchActionMapScheduleGroups.SetField("EmergencyActionMapScheduleGroup", emergencyActionMapScheduleGroup)
	return &sdkPatchActionMapScheduleGroups
}

func getActionMapScheduleGroupPair(actionMapScheduleGroups map[string]interface{}) (*platformclientv2.Actionmapschedulegroup, *platformclientv2.Actionmapschedulegroup) {
	actionMapScheduleGroupId := actionMapScheduleGroups["action_map_schedule_group_id"].(string)
	actionMapScheduleGroup := &platformclientv2.Actionmapschedulegroup{
		Id: &actionMapScheduleGroupId,
	}
	emergencyActionMapScheduleGroupId := stringmap.GetNonDefaultValue[string](actionMapScheduleGroups, "emergency_action_map_schedule_group_id")
	var emergencyActionMapScheduleGroup *platformclientv2.Actionmapschedulegroup = nil
	if emergencyActionMapScheduleGroupId != nil {
		emergencyActionMapScheduleGroup = &platformclientv2.Actionmapschedulegroup{
			Id: emergencyActionMapScheduleGroupId,
		}
	}
	return actionMapScheduleGroup, emergencyActionMapScheduleGroup
}

func SetupJourneyActionMap(t *testing.T, testCaseName string, sdkConfig *platformclientv2.Configuration) {
	_, err := provider.AuthorizeSdk()
	if err != nil {
		t.Fatal(err)
	}

	testCasePrefix := testrunner.TestObjectIdPrefix + testCaseName
	cleanupJourneySegments(testCasePrefix, sdkConfig)
	cleanupArchitectScheduleGroups(testCasePrefix)
	if err := cleanupArchitectSchedules(testCasePrefix); err != nil {
		t.Log(err)
	}
	cleanupFlows(testCasePrefix, sdkConfig)
	cleanupJourneyActionMaps(testCasePrefix, sdkConfig)
}

func cleanupFlows(idPrefix string, sdkConfig *platformclientv2.Configuration) {
	architectApi := platformclientv2.NewArchitectApiWithConfig(sdkConfig)

	for pageNum := 1; ; pageNum++ {
		const pageSize = 50
		flows, _, getErr := architectApi.GetFlows(nil, pageNum, pageSize, "", "", nil, "", "", "", "", "", "", "", "", false, true, "", "", nil)
		if getErr != nil {
			return
		}

		if flows.Entities == nil || len(*flows.Entities) == 0 {
			break
		}

		for _, flow := range *flows.Entities {
			if flow.Name != nil && strings.HasPrefix(*flow.Name, idPrefix) {
				resp, delErr := architectApi.DeleteFlow(*flow.Id)
				if delErr != nil {
					util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("failed to delete flow %s (%s): %s", *flow.Id, *flow.Name, delErr), resp)
					return
				}
				log.Printf("Deleted flow %s (%s)", *flow.Id, *flow.Name)
			}
		}
	}
}

func cleanupArchitectScheduleGroups(idPrefix string) {
	architectApi := platformclientv2.NewArchitectApi()

	for pageNum := 1; ; pageNum++ {
		const pageSize = 100
		architectScheduleGroups, _, getErr := architectApi.GetArchitectSchedulegroups(pageNum, pageSize, "", "", "", "", nil)
		if getErr != nil {
			return
		}

		if architectScheduleGroups.Entities == nil || len(*architectScheduleGroups.Entities) == 0 {
			break
		}

		for _, scheduleGroup := range *architectScheduleGroups.Entities {
			if scheduleGroup.Name != nil && strings.HasPrefix(*scheduleGroup.Name, idPrefix) {
				resp, delErr := architectApi.DeleteArchitectSchedulegroup(*scheduleGroup.Id)
				if delErr != nil {
					util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("failed to delete architect schedule group %s (%s): %s", *scheduleGroup.Id, *scheduleGroup.Name, delErr), resp)
					return
				}
				log.Printf("Deleted architect schedule group %s (%s)", *scheduleGroup.Id, *scheduleGroup.Name)
			}
		}
	}
}

func cleanupArchitectSchedules(idPrefix string) error {
	architectApi := platformclientv2.NewArchitectApi()

	for pageNum := 1; ; pageNum++ {
		const pageSize = 100
		architectSchedules, _, getErr := architectApi.GetArchitectSchedules(pageNum, pageSize, "", "", "", nil)
		if getErr != nil {
			return getErr
		}

		if architectSchedules.Entities == nil || len(*architectSchedules.Entities) == 0 {
			break
		}

		for _, schedule := range *architectSchedules.Entities {
			if schedule.Name != nil && strings.HasPrefix(*schedule.Name, idPrefix) {
				_, delErr := architectApi.DeleteArchitectSchedule(*schedule.Id)
				if delErr != nil {
					return fmt.Errorf("failed to delete architect schedule %s (%s): %s", *schedule.Id, *schedule.Name, delErr)
				}
				log.Printf("Deleted architect schedule %s (%s)", *schedule.Id, *schedule.Name)
			}
		}
	}
	return nil
}

func cleanupJourneySegments(idPrefix string, sdkConfig *platformclientv2.Configuration) {
	journeyApi := platformclientv2.NewJourneyApiWithConfig(sdkConfig)

	segmentsToDelete := make([]platformclientv2.Journeysegment, 0)

	// go through all segments to find those to delete
	const pageSize = 200
	for pageNum := 1; ; pageNum++ {
		journeySegments, _, getErr := journeyApi.GetJourneySegments("", pageSize, pageNum, true, nil, nil, "")
		if getErr != nil {
			log.Printf("failed to get page %v of journeySegments: %v", pageNum, getErr)
			return
		}

		for _, journeySegment := range *journeySegments.Entities {
			if journeySegment.DisplayName != nil && strings.HasPrefix(*journeySegment.DisplayName, idPrefix) {
				segmentsToDelete = append(segmentsToDelete, journeySegment)
			}
		}

		if *journeySegments.PageNumber >= *journeySegments.PageCount {
			break
		}
	}

	// delete them
	for _, journeySegment := range segmentsToDelete {
		_, delErr := journeyApi.DeleteJourneySegment(*journeySegment.Id)
		if delErr != nil {
			util.BuildDiagnosticError("genesyscloud_journey_segment", fmt.Sprintf("failed to delete journey segment %s (%s)", *journeySegment.Id, *journeySegment.DisplayName), delErr)
			return
		}
		log.Printf("Deleted journey segment %s (%s)", *journeySegment.Id, *journeySegment.DisplayName)
	}
}

func cleanupJourneyActionMaps(idPrefix string, sdkConfig *platformclientv2.Configuration) {
	journeyApi := platformclientv2.NewJourneyApiWithConfig(sdkConfig)

	pageCount := 1 // Needed because of broken journey common paging
	for pageNum := 1; pageNum <= pageCount; pageNum++ {
		const pageSize = 100
		actionMaps, _, getErr := journeyApi.GetJourneyActionmaps(pageNum, pageSize, "", "", "", nil, nil, "")
		if getErr != nil {
			return
		}

		if actionMaps.Entities == nil || len(*actionMaps.Entities) == 0 {
			break
		}

		for _, actionMap := range *actionMaps.Entities {
			if actionMap.DisplayName != nil && strings.HasPrefix(*actionMap.DisplayName, idPrefix) {
				resp, delErr := journeyApi.DeleteJourneyActionmap(*actionMap.Id)
				if delErr != nil {
					util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("failed to delete journey action map %s (%s): %s", *actionMap.Id, *actionMap.DisplayName, delErr), resp)
					return
				}
				log.Printf("Deleted journey action map %s (%s)", *actionMap.Id, *actionMap.DisplayName)
			}
		}

		pageCount = *actionMaps.PageCount
	}
}
