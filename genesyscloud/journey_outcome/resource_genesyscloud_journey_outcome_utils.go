package journey_outcome

import (
	lists "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/lists"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/resourcedata"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/stringmap"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v157/platformclientv2"
)

func flattenContext(context *platformclientv2.Context) map[string]interface{} {
	if len(*context.Patterns) == 0 {
		return nil
	}
	contextMap := make(map[string]interface{})
	contextMap["patterns"] = *lists.FlattenList(context.Patterns, flattenContextPattern)
	return contextMap
}

func flattenContextPattern(contextPattern *platformclientv2.Contextpattern) map[string]interface{} {
	contextPatternMap := make(map[string]interface{})
	contextPatternMap["criteria"] = *lists.FlattenList(contextPattern.Criteria, flattenEntityTypeCriteria)
	return contextPatternMap
}

func flattenEntityTypeCriteria(entityTypeCriteria *platformclientv2.Entitytypecriteria) map[string]interface{} {
	entityTypeCriteriaMap := make(map[string]interface{})
	entityTypeCriteriaMap["key"] = *entityTypeCriteria.Key
	entityTypeCriteriaMap["values"] = lists.StringListToSet(*entityTypeCriteria.Values)
	entityTypeCriteriaMap["should_ignore_case"] = *entityTypeCriteria.ShouldIgnoreCase
	entityTypeCriteriaMap["operator"] = *entityTypeCriteria.Operator
	entityTypeCriteriaMap["entity_type"] = *entityTypeCriteria.EntityType
	return entityTypeCriteriaMap
}

func flattenJourney(journey *platformclientv2.Journey) map[string]interface{} {
	if len(*journey.Patterns) == 0 {
		return nil
	}
	journeyMap := make(map[string]interface{})
	journeyMap["patterns"] = *lists.FlattenList(journey.Patterns, flattenJourneyPattern)
	return journeyMap
}

func flattenJourneyPattern(journeyPattern *platformclientv2.Journeypattern) map[string]interface{} {
	journeyPatternMap := make(map[string]interface{})
	journeyPatternMap["criteria"] = *lists.FlattenList(journeyPattern.Criteria, flattenCriteria)
	journeyPatternMap["count"] = *journeyPattern.Count
	journeyPatternMap["stream_type"] = *journeyPattern.StreamType
	journeyPatternMap["session_type"] = *journeyPattern.SessionType
	stringmap.SetValueIfNotNil(journeyPatternMap, "event_name", journeyPattern.EventName)
	return journeyPatternMap
}

func flattenCriteria(criteria *platformclientv2.Criteria) map[string]interface{} {
	criteriaMap := make(map[string]interface{})
	criteriaMap["key"] = *criteria.Key
	criteriaMap["values"] = lists.StringListToSet(*criteria.Values)
	criteriaMap["should_ignore_case"] = *criteria.ShouldIgnoreCase
	criteriaMap["operator"] = *criteria.Operator
	return criteriaMap
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

func buildSdkRequestContextPattern(contextPattern map[string]interface{}) *platformclientv2.Requestcontextpattern {
	return &platformclientv2.Requestcontextpattern{
		Criteria: stringmap.BuildSdkList(contextPattern, "criteria", buildSdkRequestEntityTypeCriteria),
	}
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

func buildSdkRequestJourney(journey map[string]interface{}) *platformclientv2.Requestjourney {
	patterns := &[]platformclientv2.Requestjourneypattern{}
	if journey != nil {
		patterns = stringmap.BuildSdkList(journey, "patterns", buildSdkRequestJourneyPattern)
	}
	return &platformclientv2.Requestjourney{
		Patterns: patterns,
	}
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

func buildSdkPatchContext(context map[string]interface{}) *platformclientv2.Patchcontext {
	patterns := &[]platformclientv2.Patchcontextpattern{}
	if context != nil {
		patterns = stringmap.BuildSdkList(context, "patterns", buildSdkPatchContextPattern)
	}
	return &platformclientv2.Patchcontext{
		Patterns: patterns,
	}
}

func buildSdkPatchContextPattern(contextPattern map[string]interface{}) *platformclientv2.Patchcontextpattern {
	return &platformclientv2.Patchcontextpattern{
		Criteria: stringmap.BuildSdkList(contextPattern, "criteria", buildSdkPatchEntityTypeCriteria),
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

func buildSdkPatchJourney(journey map[string]interface{}) *platformclientv2.Patchjourney {
	patterns := &[]platformclientv2.Patchjourneypattern{}
	if journey != nil {
		patterns = stringmap.BuildSdkList(journey, "patterns", buildSdkPatchJourneyPattern)
	}
	return &platformclientv2.Patchjourney{
		Patterns: patterns,
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

func flattenJourneyOutcome(d *schema.ResourceData, journeyOutcome *platformclientv2.Outcome) {
	d.Set("is_active", *journeyOutcome.IsActive)
	d.Set("display_name", *journeyOutcome.DisplayName)
	resourcedata.SetNillableValue(d, "description", journeyOutcome.Description)
	resourcedata.SetNillableValue(d, "is_positive", journeyOutcome.IsPositive)
	resourcedata.SetNillableValue(d, "context", lists.FlattenAsList(journeyOutcome.Context, flattenContext))
	resourcedata.SetNillableValue(d, "journey", lists.FlattenAsList(journeyOutcome.Journey, flattenJourney))

	resourcedata.SetNillableValue(d, "associated_value_field", lists.FlattenAsList(journeyOutcome.AssociatedValueField, flattenAssociatedValueField))
}

func flattenAssociatedValueField(associatedValueField *platformclientv2.Associatedvaluefield) map[string]interface{} {
	associatedValueFieldMap := make(map[string]interface{})
	associatedValueFieldMap["data_type"] = associatedValueField.DataType
	associatedValueFieldMap["name"] = associatedValueField.Name
	return associatedValueFieldMap
}

func buildSdkJourneyOutcome(journeyOutcome *schema.ResourceData) *platformclientv2.Outcomerequest {
	isActive := journeyOutcome.Get("is_active").(bool)
	displayName := journeyOutcome.Get("display_name").(string)
	description := resourcedata.GetNillableValue[string](journeyOutcome, "description")
	isPositive := resourcedata.GetNillableBool(journeyOutcome, "is_positive")
	sdkContext := resourcedata.BuildSdkListFirstElement(journeyOutcome, "context", buildSdkRequestContext, false)
	journey := resourcedata.BuildSdkListFirstElement(journeyOutcome, "journey", buildSdkRequestJourney, false)
	associatedValueField := resourcedata.BuildSdkListFirstElement(journeyOutcome, "associated_value_field", buildSdkAssociatedValueField, true)

	return &platformclientv2.Outcomerequest{
		IsActive:             &isActive,
		DisplayName:          &displayName,
		Description:          description,
		IsPositive:           isPositive,
		Context:              sdkContext,
		Journey:              journey,
		AssociatedValueField: associatedValueField,
	}
}

func buildSdkPatchOutcome(journeyOutcome *schema.ResourceData) *platformclientv2.Patchoutcome {
	isActive := journeyOutcome.Get("is_active").(bool)
	displayName := journeyOutcome.Get("display_name").(string)
	description := resourcedata.GetNillableValue[string](journeyOutcome, "description")
	isPositive := resourcedata.GetNillableBool(journeyOutcome, "is_positive")
	sdkContext := resourcedata.BuildSdkListFirstElement(journeyOutcome, "context", buildSdkPatchContext, false)
	journey := resourcedata.BuildSdkListFirstElement(journeyOutcome, "journey", buildSdkPatchJourney, false)

	return &platformclientv2.Patchoutcome{
		IsActive:    &isActive,
		DisplayName: &displayName,
		Description: description,
		IsPositive:  isPositive,
		Context:     sdkContext,
		Journey:     journey,
	}
}

func buildSdkAssociatedValueField(associatedValueField map[string]interface{}) *platformclientv2.Associatedvaluefield {
	dataType := associatedValueField["data_type"].(string)
	name := associatedValueField["name"].(string)

	return &platformclientv2.Associatedvaluefield{
		DataType: &dataType,
		Name:     &name,
	}
}

func GetStringValue(s *string, defaultValue string) string {
	if s != nil {
		return *s
	}
	return defaultValue
}
