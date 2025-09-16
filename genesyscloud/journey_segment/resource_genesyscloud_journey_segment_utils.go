package journey_segment

import (
	lists "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/lists"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/resourcedata"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/stringmap"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v165/platformclientv2"
)

func flattenJourneySegment(d *schema.ResourceData, journeySegment *platformclientv2.Journeysegment) {
	d.Set("is_active", *journeySegment.IsActive)
	d.Set("display_name", *journeySegment.DisplayName)
	resourcedata.SetNillableValue(d, "description", journeySegment.Description)
	d.Set("color", *journeySegment.Color)
	resourcedata.SetNillableValue(d, "should_display_to_agent", journeySegment.ShouldDisplayToAgent)
	resourcedata.SetNillableValue(d, "context", lists.FlattenAsList(journeySegment.Context, flattenContext))
	resourcedata.SetNillableValue(d, "journey", lists.FlattenAsList(journeySegment.Journey, flattenJourney))
	resourcedata.SetNillableValue(d, "assignment_expiration_days", journeySegment.AssignmentExpirationDays)
}

func buildSdkJourneySegment(journeySegment *schema.ResourceData) *platformclientv2.Journeysegmentrequest {
	isActive := journeySegment.Get("is_active").(bool)
	displayName := journeySegment.Get("display_name").(string)
	description := resourcedata.GetNillableValue[string](journeySegment, "description")
	color := journeySegment.Get("color").(string)
	shouldDisplayToAgent := resourcedata.GetNillableBool(journeySegment, "should_display_to_agent")
	sdkContext := resourcedata.BuildSdkListFirstElement(journeySegment, "context", buildSdkRequestContext, false)
	journey := resourcedata.BuildSdkListFirstElement(journeySegment, "journey", buildSdkRequestJourney, false)
	assignmentExpirationDays := resourcedata.GetNillableValue[int](journeySegment, "assignment_expiration_days")

	return &platformclientv2.Journeysegmentrequest{
		IsActive:                 &isActive,
		DisplayName:              &displayName,
		Description:              description,
		Color:                    &color,
		ShouldDisplayToAgent:     shouldDisplayToAgent,
		Context:                  sdkContext,
		Journey:                  journey,
		AssignmentExpirationDays: assignmentExpirationDays,
	}
}

func buildSdkPatchSegment(journeySegment *schema.ResourceData) *platformclientv2.Patchsegment {
	isActive := journeySegment.Get("is_active").(bool)
	displayName := journeySegment.Get("display_name").(string)
	description := resourcedata.GetNillableValue[string](journeySegment, "description")
	color := journeySegment.Get("color").(string)
	shouldDisplayToAgent := resourcedata.GetNillableBool(journeySegment, "should_display_to_agent")
	sdkContext := resourcedata.BuildSdkListFirstElement(journeySegment, "context", buildSdkPatchContext, false)
	journey := resourcedata.BuildSdkListFirstElement(journeySegment, "journey", buildSdkPatchJourney, false)

	sdkPatchSegment := platformclientv2.Patchsegment{}
	sdkPatchSegment.SetField("IsActive", &isActive)
	sdkPatchSegment.SetField("DisplayName", &displayName)
	sdkPatchSegment.SetField("Description", description)
	sdkPatchSegment.SetField("Color", &color)
	sdkPatchSegment.SetField("ShouldDisplayToAgent", shouldDisplayToAgent)
	sdkPatchSegment.SetField("Context", sdkContext)
	sdkPatchSegment.SetField("Journey", journey)
	sdkPatchSegment.SetField("AssignmentExpirationDays", resourcedata.GetNillableValue[int](journeySegment, "assignment_expiration_days"))

	return &sdkPatchSegment
}

func flattenContext(context *platformclientv2.Context) map[string]interface{} {
	if len(*context.Patterns) == 0 {
		return nil
	}
	contextMap := make(map[string]interface{})
	if context.Patterns != nil {
		stringmap.SetValueIfNotNil(contextMap, "patterns", lists.FlattenList(context.Patterns, flattenContextPattern))
	}
	return contextMap
}

func buildSdkRequestContext(context map[string]interface{}) *platformclientv2.Requestcontext {
	patterns := &[]platformclientv2.Requestcontextpattern{}
	if context != nil {
		patterns = stringmap.BuildSdkList(context, "patterns", buildSdkRequestContextPattern)
	}
	return &platformclientv2.Requestcontext{
		Patterns: patterns,
	}
}

func buildSdkPatchContext(context map[string]interface{}) *platformclientv2.Patchcontext {
	patterns := &[]platformclientv2.Patchcontextpattern{}
	if context != nil {
		patterns = stringmap.BuildSdkList(context, "patterns", buildSdkPatchContextPattern)
	}
	return &platformclientv2.Patchcontext{
		Patterns: patterns,
	}
}

func flattenContextPattern(contextPattern *platformclientv2.Contextpattern) map[string]interface{} {
	contextPatternMap := make(map[string]interface{})
	if contextPattern.Criteria != nil {
		stringmap.SetValueIfNotNil(contextPatternMap, "criteria", lists.FlattenList(contextPattern.Criteria, flattenEntityTypeCriteria))
	}
	return contextPatternMap
}

func buildSdkRequestContextPattern(contextPattern map[string]interface{}) *platformclientv2.Requestcontextpattern {
	return &platformclientv2.Requestcontextpattern{
		Criteria: stringmap.BuildSdkList(contextPattern, "criteria", buildSdkRequestEntityTypeCriteria),
	}
}

func buildSdkPatchContextPattern(contextPattern map[string]interface{}) *platformclientv2.Patchcontextpattern {
	return &platformclientv2.Patchcontextpattern{
		Criteria: stringmap.BuildSdkList(contextPattern, "criteria", buildSdkPatchEntityTypeCriteria),
	}
}

func flattenEntityTypeCriteria(entityTypeCriteria *platformclientv2.Entitytypecriteria) map[string]interface{} {
	entityTypeCriteriaMap := make(map[string]interface{})
	resourcedata.SetMapValueIfNotNil(entityTypeCriteriaMap, "key", entityTypeCriteria.Key)
	resourcedata.SetMapValueIfNotNil(entityTypeCriteriaMap, "should_ignore_case", entityTypeCriteria.ShouldIgnoreCase)
	resourcedata.SetMapValueIfNotNil(entityTypeCriteriaMap, "operator", entityTypeCriteria.Operator)
	resourcedata.SetMapValueIfNotNil(entityTypeCriteriaMap, "entity_type", entityTypeCriteria.EntityType)
	if entityTypeCriteria.Values != nil {
		entityTypeCriteriaMap["values"] = lists.StringListToSet(*entityTypeCriteria.Values)
	}
	return entityTypeCriteriaMap
}

func buildSdkRequestEntityTypeCriteria(entityTypeCriteria map[string]interface{}) *platformclientv2.Requestentitytypecriteria {
	key := entityTypeCriteria["key"].(string)
	values := stringmap.BuildSdkStringList(entityTypeCriteria, "values")
	shouldIgnoreCase := entityTypeCriteria["should_ignore_case"].(bool)
	operator := entityTypeCriteria["operator"].(string)
	entityType := entityTypeCriteria["entity_type"].(string)

	return &platformclientv2.Requestentitytypecriteria{
		Key:              &key,
		Values:           values,
		ShouldIgnoreCase: &shouldIgnoreCase,
		Operator:         &operator,
		EntityType:       &entityType,
	}
}

func buildSdkPatchEntityTypeCriteria(entityTypeCriteria map[string]interface{}) *platformclientv2.Patchentitytypecriteria {
	key := entityTypeCriteria["key"].(string)
	values := stringmap.BuildSdkStringList(entityTypeCriteria, "values")
	shouldIgnoreCase := entityTypeCriteria["should_ignore_case"].(bool)
	operator := entityTypeCriteria["operator"].(string)
	entityType := entityTypeCriteria["entity_type"].(string)

	return &platformclientv2.Patchentitytypecriteria{
		Key:              &key,
		Values:           values,
		ShouldIgnoreCase: &shouldIgnoreCase,
		Operator:         &operator,
		EntityType:       &entityType,
	}
}

func flattenJourney(journey *platformclientv2.Journey) map[string]interface{} {
	if len(*journey.Patterns) == 0 {
		return nil
	}
	journeyMap := make(map[string]interface{})
	if journey.Patterns != nil {
		stringmap.SetValueIfNotNil(journeyMap, "patterns", lists.FlattenList(journey.Patterns, flattenJourneyPattern))
	}
	return journeyMap
}

func buildSdkRequestJourney(journey map[string]interface{}) *platformclientv2.Requestjourney {
	patterns := &[]platformclientv2.Requestjourneypattern{}
	if journey != nil {
		patterns = stringmap.BuildSdkList(journey, "patterns", buildSdkRequestJourneyPattern)
	}
	return &platformclientv2.Requestjourney{
		Patterns: patterns,
	}
}

func buildSdkPatchJourney(journey map[string]interface{}) *platformclientv2.Patchjourney {
	patterns := &[]platformclientv2.Patchjourneypattern{}
	if journey != nil {
		patterns = stringmap.BuildSdkList(journey, "patterns", buildSdkPatchJourneyPattern)
	}
	return &platformclientv2.Patchjourney{
		Patterns: patterns,
	}
}

func flattenJourneyPattern(journeyPattern *platformclientv2.Journeypattern) map[string]interface{} {
	journeyPatternMap := make(map[string]interface{})
	stringmap.SetValueIfNotNil(journeyPatternMap, "criteria", lists.FlattenList(journeyPattern.Criteria, flattenCriteria))
	stringmap.SetValueIfNotNil(journeyPatternMap, "count", journeyPattern.Count)
	stringmap.SetValueIfNotNil(journeyPatternMap, "stream_type", journeyPattern.StreamType)
	stringmap.SetValueIfNotNil(journeyPatternMap, "session_type", journeyPattern.SessionType)
	stringmap.SetValueIfNotNil(journeyPatternMap, "event_name", journeyPattern.EventName)
	if journeyPattern.Criteria != nil {
		stringmap.SetValueIfNotNil(journeyPatternMap, "criteria", lists.FlattenList(journeyPattern.Criteria, flattenCriteria))
	}
	return journeyPatternMap
}

func buildSdkRequestJourneyPattern(journeyPattern map[string]interface{}) *platformclientv2.Requestjourneypattern {
	criteria := stringmap.BuildSdkList(journeyPattern, "criteria", buildSdkRequestCriteria)
	count := journeyPattern["count"].(int)
	streamType := journeyPattern["stream_type"].(string)
	sessionType := journeyPattern["session_type"].(string)
	eventName := stringmap.GetNonDefaultValue[string](journeyPattern, "event_name")

	return &platformclientv2.Requestjourneypattern{
		Criteria:    criteria,
		Count:       &count,
		StreamType:  &streamType,
		SessionType: &sessionType,
		EventName:   eventName,
	}
}

func buildSdkPatchJourneyPattern(journeyPattern map[string]interface{}) *platformclientv2.Patchjourneypattern {
	criteria := stringmap.BuildSdkList(journeyPattern, "criteria", buildSdkPatchCriteria)
	count := journeyPattern["count"].(int)
	streamType := journeyPattern["stream_type"].(string)
	sessionType := journeyPattern["session_type"].(string)
	eventName := stringmap.GetNonDefaultValue[string](journeyPattern, "event_name")

	return &platformclientv2.Patchjourneypattern{
		Criteria:    criteria,
		Count:       &count,
		StreamType:  &streamType,
		SessionType: &sessionType,
		EventName:   eventName,
	}
}

func flattenCriteria(criteria *platformclientv2.Criteria) map[string]interface{} {
	criteriaMap := make(map[string]interface{})
	stringmap.SetValueIfNotNil(criteriaMap, "key", criteria.Key)
	stringmap.SetValueIfNotNil(criteriaMap, "should_ignore_case", criteria.ShouldIgnoreCase)
	stringmap.SetValueIfNotNil(criteriaMap, "operator", criteria.Operator)
	if criteria.Values != nil {
		criteriaMap["values"] = lists.StringListToSet(*criteria.Values)
	}
	return criteriaMap
}

func buildSdkRequestCriteria(criteria map[string]interface{}) *platformclientv2.Requestcriteria {
	key := criteria["key"].(string)
	values := stringmap.BuildSdkStringList(criteria, "values")
	shouldIgnoreCase := criteria["should_ignore_case"].(bool)
	operator := criteria["operator"].(string)

	return &platformclientv2.Requestcriteria{
		Key:              &key,
		Values:           values,
		ShouldIgnoreCase: &shouldIgnoreCase,
		Operator:         &operator,
	}
}

func buildSdkPatchCriteria(criteria map[string]interface{}) *platformclientv2.Patchcriteria {
	key := criteria["key"].(string)
	values := stringmap.BuildSdkStringList(criteria, "values")
	shouldIgnoreCase := criteria["should_ignore_case"].(bool)
	operator := criteria["operator"].(string)

	return &platformclientv2.Patchcriteria{
		Key:              &key,
		Values:           values,
		ShouldIgnoreCase: &shouldIgnoreCase,
		Operator:         &operator,
	}
}
