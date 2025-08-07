package journey_views

import (
	"errors"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/resourcedata"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v165/platformclientv2"
)

func buildElements(d *schema.ResourceData) (*[]platformclientv2.Journeyviewelement, error) {
	elementsSlice := d.Get("elements").([]interface{})
	if len(elementsSlice) == 0 {
		emptySlice := make([]platformclientv2.Journeyviewelement, 0)
		return &emptySlice, nil
	}

	var elements []platformclientv2.Journeyviewelement

	for _, elem := range elementsSlice {
		elemMap, ok := elem.(map[string]interface{})
		if !ok {
			return nil, errors.New("element is not a map[string]interface{}")
		}

		var element platformclientv2.Journeyviewelement
		element.Id = getStringPointerFromInterface(elemMap["id"])
		element.Name = getStringPointerFromInterface(elemMap["name"])

		if attributesSlice, ok := elemMap["attributes"].([]interface{}); ok {
			attributes := buildJourneyViewElementAttributes(attributesSlice)
			element.Attributes = &attributes
		}

		if displayAttributesSlice, ok := elemMap["display_attributes"].([]interface{}); ok {
			displayAttributes := buildJourneyViewElementDisplayAttributes(displayAttributesSlice)
			element.DisplayAttributes = displayAttributes
		}

		if filterSlice, ok := elemMap["filter"].([]interface{}); ok {
			filter := buildJourneyViewElementFilter(filterSlice)
			element.Filter = filter
		}

		if followedBySlice, ok := elemMap["followed_by"].([]interface{}); ok {
			followedBy := make([]platformclientv2.Journeyviewlink, len(followedBySlice))
			for i, fb := range followedBySlice {
				followedByMap, ok := fb.(map[string]interface{})
				if !ok {
					return nil, errors.New("followedBy element is not a map[string]interface{}")
				}
				followedBy[i] = buildJourneyViewLink(followedByMap)
			}
			element.FollowedBy = &followedBy
		}
		elements = append(elements, element)
	}

	return &elements, nil
}

func buildJourneyViewElementAttributes(attributesSlice []interface{}) platformclientv2.Journeyviewelementattributes {
	var attributes platformclientv2.Journeyviewelementattributes
	for _, elem := range attributesSlice {
		if attributesMap, ok := elem.(map[string]interface{}); ok {
			attributes.VarType = getStringPointerFromInterface(attributesMap["type"])
			attributes.Id = getStringPointerFromInterface(attributesMap["id"])
			attributes.Source = getStringPointerFromInterface(attributesMap["source"])
		}
	}
	return attributes
}

func buildJourneyViewElementDisplayAttributes(displayAttributesSlice []interface{}) *platformclientv2.Journeyviewelementdisplayattributes {
	var displayAttributes platformclientv2.Journeyviewelementdisplayattributes
	if len(displayAttributesSlice) == 0 {
		return nil
	}

	for _, elem := range displayAttributesSlice {
		if displayAttributesMap, ok := elem.(map[string]interface{}); ok {
			displayAttributes.X = getIntPointerFromInterface(displayAttributesMap["x"])
			displayAttributes.Y = getIntPointerFromInterface(displayAttributesMap["y"])
			displayAttributes.Col = getIntPointerFromInterface(displayAttributesMap["col"])
		}
	}
	return &displayAttributes
}

func buildJourneyViewElementFilter(filterSlice []interface{}) *platformclientv2.Journeyviewelementfilter {
	var filter platformclientv2.Journeyviewelementfilter
	for _, elem := range filterSlice {
		if filterMap, ok := elem.(map[string]interface{}); ok {
			filter.VarType = getStringPointerFromInterface(filterMap["type"])
			if predicatesSlice, ok := filterMap["predicates"].([]interface{}); ok {
				if len(predicatesSlice) > 0 {
					predicates := make([]platformclientv2.Journeyviewelementfilterpredicate, len(predicatesSlice))
					for i, predicate := range predicatesSlice {
						predicateMap, ok := predicate.(map[string]interface{})
						if ok {
							predicates[i] = buildJourneyviewelementfilterpredicate(predicateMap)
						}
					}
					filter.Predicates = &predicates
				}
			}

			if numberPredicatesSlice, ok := filterMap["number_predicates"].([]interface{}); ok {
				if len(numberPredicatesSlice) > 0 {
					numberPredicates := make([]platformclientv2.Journeyviewelementfilternumberpredicate, len(numberPredicatesSlice))
					for i, numberPredicate := range numberPredicatesSlice {
						numberPredicateMap, ok := numberPredicate.(map[string]interface{})
						if ok {
							numberPredicates[i] = buildJourneyviewelementfilternumberpredicate(numberPredicateMap)
						}
					}
					filter.NumberPredicates = &numberPredicates
				}
			}
		}
	}
	if len(filterSlice) == 0 {
		return nil
	}
	return &filter
}

func buildJourneyviewelementfilterpredicate(predicateMap map[string]interface{}) platformclientv2.Journeyviewelementfilterpredicate {
	var predicate platformclientv2.Journeyviewelementfilterpredicate
	predicate.Dimension = getStringPointerFromInterface(predicateMap["dimension"])
	if valuesSlice, ok := predicateMap["values"].([]interface{}); ok {
		values := make([]string, len(valuesSlice))
		for i, value := range valuesSlice {
			if stringValue, ok := value.(string); ok {
				values[i] = stringValue
			}
		}
		predicate.Values = &values
	}
	predicate.Operator = getStringPointerFromInterface(predicateMap["operator"])
	predicate.NoValue = getBoolPointerFromInterface(predicateMap["no_value"])
	return predicate
}

// Nicolas
func buildJourneyviewelementfilternumberpredicate(numberPredicateMap map[string]interface{}) platformclientv2.Journeyviewelementfilternumberpredicate {
	var numberPredicate platformclientv2.Journeyviewelementfilternumberpredicate
	numberPredicate.Dimension = getStringPointerFromInterface(numberPredicateMap["dimension"])

	if rangeSlice, ok := numberPredicateMap["range"].([]interface{}); ok {
		numberPredicate.VarRange = buildJourneyviewelementfilterrange(rangeSlice)
	}
	numberPredicate.Operator = getStringPointerFromInterface(numberPredicateMap["operator"])
	numberPredicate.NoValue = getBoolPointerFromInterface(numberPredicateMap["no_value"])
	return numberPredicate
}

// Nicolas
func buildJourneyviewelementfilterrange(rangeSlice []interface{}) *platformclientv2.Journeyviewelementfilterrange {
	var varRange platformclientv2.Journeyviewelementfilterrange

	for _, rangeObj := range rangeSlice {
		rangeMap, ok := rangeObj.(map[string]interface{})
		if ok {
			if rangeMap["lt"] != nil {
				ltSlice, ok := rangeMap["lt"].(*schema.Set)
				if ok {
					varRange.Lt = buildJourneyviewelementfilterrangedata(ltSlice)
				}
			}
			if rangeMap["lte"] != nil {
				lteSlice, ok := rangeMap["lte"].(*schema.Set)
				if ok {
					varRange.Lte = buildJourneyviewelementfilterrangedata(lteSlice)
				}
			}
			if rangeMap["gt"] != nil {
				gtSlice, ok := rangeMap["gt"].(*schema.Set)
				if ok {
					varRange.Gt = buildJourneyviewelementfilterrangedata(gtSlice)
				}
			}
			if rangeMap["gte"] != nil {
				gteMap, ok := rangeMap["gte"].(*schema.Set)
				if ok {
					varRange.Gte = buildJourneyviewelementfilterrangedata(gteMap)
				}
			}
			if rangeMap["eq"] != nil {
				eqSlice, ok := rangeMap["eq"].(*schema.Set)
				if ok {
					varRange.Eq = buildJourneyviewelementfilterrangedata(eqSlice)
				}
			}
			if rangeMap["neq"] != nil {
				neqSlice, ok := rangeMap["neq"].(*schema.Set)
				if ok {
					varRange.Neq = buildJourneyviewelementfilterrangedata(neqSlice)
				}
			}
		}
	}

	return &varRange
}

func buildJourneyviewelementfilterrangedata(dataSlice *schema.Set) *platformclientv2.Journeyviewelementfilterrangedata {
	var data platformclientv2.Journeyviewelementfilterrangedata

	dataList := dataSlice.List()

	if len(dataList) == 0 {
		return nil
	}

	for _, attr := range dataSlice.List() {
		dataMap, ok := attr.(map[string]interface{})
		if !ok {
			return nil
		}
		data.Duration = getStringPointerFromInterface(dataMap["duration"])
		data.Number = getFloatPointerFromInterface(dataMap["number"])
	}
	return &data
}

func buildJourneyViewLink(linkMap map[string]interface{}) platformclientv2.Journeyviewlink {
	var link platformclientv2.Journeyviewlink
	link.Id = getStringPointerFromInterface(linkMap["id"])
	if constraintWithinSlice, ok := linkMap["constraint_within"].([]interface{}); ok {
		constraintWithin := buildJourneyViewLinkTimeConstraint(constraintWithinSlice)
		if constraintWithin != nil {
			link.ConstraintWithin = constraintWithin
		}
	}
	if constraintAfterSlice, ok := linkMap["constraint_after"].([]interface{}); ok {
		constraintAfter := buildJourneyViewLinkTimeConstraint(constraintAfterSlice)
		if constraintAfter != nil {
			link.ConstraintAfter = constraintAfter
		}
	}
	link.EventCountType = getStringPointerFromInterface(linkMap["event_count_type"])
	if joinAttributesSlice, ok := linkMap["join_attributes"].([]interface{}); ok {
		joinAttributes := make([]string, len(joinAttributesSlice))
		for i, attr := range joinAttributesSlice {
			if stringValue, ok := attr.(string); ok {
				joinAttributes[i] = stringValue
			}
		}
		if len(joinAttributes) > 0 {
			link.JoinAttributes = &joinAttributes
		}
	}
	return link
}

func buildJourneyViewLinkTimeConstraint(timeConstraintSlice []interface{}) *platformclientv2.Journeyviewlinktimeconstraint {
	if timeConstraintSlice == nil || len(timeConstraintSlice) == 0 {
		return nil
	}
	var timeConstraint platformclientv2.Journeyviewlinktimeconstraint
	for _, elem := range timeConstraintSlice {
		timeConstraintMap, ok := elem.(map[string]interface{})
		if ok {
			timeConstraint.Unit = getStringPointerFromInterface(timeConstraintMap["unit"])
			timeConstraint.Value = getIntPointerFromInterface(timeConstraintMap["value"])
		}
	}
	return &timeConstraint
}

func buildCharts(d *schema.ResourceData) *[]platformclientv2.Journeyviewchart {
	chartsSlice := d.Get("charts").([]interface{})
	if len(chartsSlice) == 0 {
		emptySlice := make([]platformclientv2.Journeyviewchart, 0)
		return &emptySlice
	}

	var charts []platformclientv2.Journeyviewchart

	for _, obj := range chartsSlice {
		chartMap, ok := obj.(map[string]interface{})
		if !ok {
			return nil //"chart is not a map[string]interface{}")
		}

		var chart platformclientv2.Journeyviewchart
		//element.Id = getStringPointerFromInterface(elemMap["id"])
		chart.Name = getStringPointerFromInterface(chartMap["name"])
		chart.Version = getIntPointerFromInterface(chartMap["version"])
		if metricsSlice, ok := chartMap["metrics"].([]interface{}); ok {
			chart.Metrics = buildMetrics(metricsSlice)
		}
		chart.GroupByTime = getStringPointerFromInterface(chartMap["group_by_time"])
		chart.GroupByMax = getIntPointerFromInterface(chartMap["group_by_max"])
		if displayAttributesSlice, ok := chartMap["display_attributes"].([]interface{}); ok {
			chart.DisplayAttributes = buildDisplayAttributes(displayAttributesSlice)
		}
		if groupByAttributesSlice, ok := chartMap["group_by_attributes"].([]interface{}); ok {
			chart.GroupByAttributes = buildGroupByAttributes(groupByAttributesSlice)
		}
		charts = append(charts, chart)
	}
	return &charts
}

func buildMetrics(objsSlice []interface{}) *[]platformclientv2.Journeyviewchartmetric {
	if len(objsSlice) == 0 {
		emptySlice := make([]platformclientv2.Journeyviewchartmetric, 0)
		return &emptySlice
	}

	var objs []platformclientv2.Journeyviewchartmetric

	for _, obj := range objsSlice {
		objMap, ok := obj.(map[string]interface{})
		if !ok {
			return nil //"metric is not a map[string]interface{}")
		}

		var metric platformclientv2.Journeyviewchartmetric
		metric.Id = getStringPointerFromInterface(objMap["id"])
		metric.Aggregate = getStringPointerFromInterface(objMap["aggregate"])
		metric.DisplayLabel = getStringPointerFromInterface(objMap["display_label"])
		metric.ElementId = getStringPointerFromInterface(objMap["element_id"])
		objs = append(objs, metric)
	}

	return &objs
}

func buildDisplayAttributes(objsSlice []interface{}) *platformclientv2.Journeyviewchartdisplayattributes {
	if len(objsSlice) == 0 {
		return nil
	}

	var displayAttribute platformclientv2.Journeyviewchartdisplayattributes
	for _, obj := range objsSlice {
		objMap, ok := obj.(map[string]interface{})
		if !ok {
			return nil //"metric is not a map[string]interface{}")
		}
		displayAttribute.VarType = getStringPointerFromInterface(objMap["var_type"])
		displayAttribute.GroupByTitle = getStringPointerFromInterface(objMap["group_by_title"])
		displayAttribute.MetricsTitle = getStringPointerFromInterface(objMap["metrics_title"])
		displayAttribute.ShowLegend = getBoolPointerFromInterface(objMap["show_legend"])
		break
	}

	return &displayAttribute
}

func buildGroupByAttributes(objsSlice []interface{}) *[]platformclientv2.Journeyviewchartgroupbyattribute {
	if len(objsSlice) == 0 {
		emptySlice := make([]platformclientv2.Journeyviewchartgroupbyattribute, 0)
		return &emptySlice
	}

	var objs []platformclientv2.Journeyviewchartgroupbyattribute

	for _, obj := range objsSlice {
		objMap, ok := obj.(map[string]interface{})
		if !ok {
			return nil //"groupbyattribute is not a map[string]interface{}")
		}

		var groupbyattribute platformclientv2.Journeyviewchartgroupbyattribute
		groupbyattribute.Attribute = getStringPointerFromInterface(objMap["attribute"])
		groupbyattribute.ElementId = getStringPointerFromInterface(objMap["element_id"])
		objs = append(objs, groupbyattribute)
	}

	return &objs
}

func getStringPointerFromInterface(val interface{}) *string {
	if valString, ok := val.(string); ok {
		if valString == "" {
			return nil
		}
		return &valString
	}
	return nil
}

func getBoolPointerFromInterface(val interface{}) *bool {
	if valBool, ok := val.(bool); ok {
		return &valBool
	}
	return nil
}

func getIntPointerFromInterface(val interface{}) *int {
	if valInt, ok := val.(int); ok {
		return &valInt
	}
	return nil
}

func getFloatPointerFromInterface(val interface{}) *float64 {
	if valFloat, ok := val.(float64); ok {
		return &valFloat
	}
	return nil
}

func flattenElements(elements *[]platformclientv2.Journeyviewelement) []interface{} {
	if len(*elements) == 0 {
		return nil
	}
	var elementsList []interface{}
	for _, element := range *elements {
		elementsMap := make(map[string]interface{})
		resourcedata.SetMapValueIfNotNil(elementsMap, "id", element.Id)
		resourcedata.SetMapValueIfNotNil(elementsMap, "name", element.Name)
		resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(elementsMap, "attributes", element.Attributes, flattenAttributes)
		resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(elementsMap, "display_attributes", element.DisplayAttributes, flattenElementDisplayAttributes)
		resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(elementsMap, "filter", element.Filter, flattenFilters)
		resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(elementsMap, "followed_by", element.FollowedBy, flattenJourneyViewLink)
		elementsList = append(elementsList, elementsMap)
	}
	return elementsList
}

func flattenAttributes(attribute *platformclientv2.Journeyviewelementattributes) []interface{} {
	if attribute == nil {
		return nil
	}
	var attributesList []interface{}
	attributesMap := make(map[string]interface{})
	resourcedata.SetMapValueIfNotNil(attributesMap, "id", attribute.Id)
	resourcedata.SetMapValueIfNotNil(attributesMap, "type", attribute.VarType)
	resourcedata.SetMapValueIfNotNil(attributesMap, "source", attribute.Source)
	attributesList = append(attributesList, attributesMap)
	return attributesList
}

func flattenElementDisplayAttributes(displayAttributes *platformclientv2.Journeyviewelementdisplayattributes) []interface{} {
	if displayAttributes == nil {
		return nil
	}
	var displayAttributesList []interface{}
	displayAttributesMap := make(map[string]interface{})
	resourcedata.SetMapValueIfNotNil(displayAttributesMap, "x", displayAttributes.X)
	resourcedata.SetMapValueIfNotNil(displayAttributesMap, "y", displayAttributes.Y)
	resourcedata.SetMapValueIfNotNil(displayAttributesMap, "col", displayAttributes.Col)
	displayAttributesList = append(displayAttributesList, displayAttributesMap)
	return displayAttributesList
}

func flattenFilters(filter *platformclientv2.Journeyviewelementfilter) []interface{} {
	if filter == nil {
		return nil
	}
	var filtersList []interface{}
	filtersMap := make(map[string]interface{})
	resourcedata.SetMapValueIfNotNil(filtersMap, "type", filter.VarType)
	resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(filtersMap, "predicates", filter.Predicates, flattenPredicates)
	resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(filtersMap, "number_predicates", filter.NumberPredicates, flattenNumberPredicates)
	filtersList = append(filtersList, filtersMap)
	return filtersList
}

func flattenPredicates(predicates *[]platformclientv2.Journeyviewelementfilterpredicate) []interface{} {
	if len(*predicates) == 0 {
		return nil
	}
	var predicatesList []interface{}
	for _, predicate := range *predicates {
		predicatesMap := make(map[string]interface{})
		resourcedata.SetMapValueIfNotNil(predicatesMap, "dimension", predicate.Dimension)
		resourcedata.SetMapValueIfNotNil(predicatesMap, "values", predicate.Values)
		resourcedata.SetMapValueIfNotNil(predicatesMap, "operator", predicate.Operator)
		resourcedata.SetMapValueIfNotNil(predicatesMap, "no_value", predicate.NoValue)
		predicatesList = append(predicatesList, predicatesMap)
	}
	return predicatesList
}

func flattenNumberPredicates(numberPredicates *[]platformclientv2.Journeyviewelementfilternumberpredicate) []interface{} {
	if len(*numberPredicates) == 0 {
		return nil
	}
	var numberPredicatesList []interface{}
	for _, numberPredicate := range *numberPredicates {
		numberPredicatesMap := make(map[string]interface{})
		resourcedata.SetMapValueIfNotNil(numberPredicatesMap, "dimension", numberPredicate.Dimension)
		resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(numberPredicatesMap, "range", numberPredicate.VarRange, flattenNumberPredicateRange)
		resourcedata.SetMapValueIfNotNil(numberPredicatesMap, "operator", numberPredicate.Operator)
		resourcedata.SetMapValueIfNotNil(numberPredicatesMap, "no_value", numberPredicate.NoValue)
		numberPredicatesList = append(numberPredicatesList, numberPredicatesMap)
	}
	return numberPredicatesList
}

func flattenNumberPredicateRange(varRange *platformclientv2.Journeyviewelementfilterrange) []interface{} {

	var numberPredicatesRangesList []interface{}
	numberPredicatesRangesMap := make(map[string]interface{})
	resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(numberPredicatesRangesMap, "lt", varRange.Lt, flattenNumberPredicateRangeData)
	resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(numberPredicatesRangesMap, "lte", varRange.Lte, flattenNumberPredicateRangeData)
	resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(numberPredicatesRangesMap, "gt", varRange.Gt, flattenNumberPredicateRangeData)
	resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(numberPredicatesRangesMap, "gte", varRange.Gte, flattenNumberPredicateRangeData)
	resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(numberPredicatesRangesMap, "eq", varRange.Eq, flattenNumberPredicateRangeData)
	resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(numberPredicatesRangesMap, "neq", varRange.Neq, flattenNumberPredicateRangeData)
	numberPredicatesRangesList = append(numberPredicatesRangesList, numberPredicatesRangesMap)

	return numberPredicatesRangesList

}

func flattenNumberPredicateRangeData(varRangeData *platformclientv2.Journeyviewelementfilterrangedata) []interface{} {
	var numberPredicatesRangeDatasList []interface{}
	numberPredicatesDataMap := make(map[string]interface{})
	resourcedata.SetMapValueIfNotNil(numberPredicatesDataMap, "number", varRangeData.Number)
	resourcedata.SetMapValueIfNotNil(numberPredicatesDataMap, "duration", varRangeData.Duration)
	numberPredicatesRangeDatasList = append(numberPredicatesRangeDatasList, numberPredicatesDataMap)

	return numberPredicatesRangeDatasList
}

func flattenJourneyViewLink(journeyViewLinks *[]platformclientv2.Journeyviewlink) []interface{} {
	if len(*journeyViewLinks) == 0 {
		return nil
	}
	var journeyViewLinksList []interface{}
	for _, journeyViewLink := range *journeyViewLinks {
		journeyViewLinksMap := make(map[string]interface{})
		resourcedata.SetMapValueIfNotNil(journeyViewLinksMap, "id", journeyViewLink.Id)
		resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(journeyViewLinksMap, "constraint_within", journeyViewLink.ConstraintWithin, flattenConstraints)
		resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(journeyViewLinksMap, "constraint_after", journeyViewLink.ConstraintAfter, flattenConstraints)
		resourcedata.SetMapValueIfNotNil(journeyViewLinksMap, "event_count_type", journeyViewLink.EventCountType)
		resourcedata.SetMapValueIfNotNil(journeyViewLinksMap, "join_attributes", journeyViewLink.JoinAttributes)
		journeyViewLinksList = append(journeyViewLinksList, journeyViewLinksMap)
	}
	return journeyViewLinksList
}

func flattenConstraints(constraint *platformclientv2.Journeyviewlinktimeconstraint) []interface{} {
	if constraint == nil {
		return nil
	}
	var constraintsList []interface{}
	constraintMap := make(map[string]interface{})
	resourcedata.SetMapValueIfNotNil(constraintMap, "unit", constraint.Unit)
	resourcedata.SetMapValueIfNotNil(constraintMap, "value", constraint.Value)
	constraintsList = append(constraintsList, constraintMap)
	return constraintsList
}

func flattenCharts(charts *[]platformclientv2.Journeyviewchart) []interface{} {
	if len(*charts) == 0 {
		return nil
	}
	var chartsList []interface{}
	for _, chart := range *charts {
		chartsMap := make(map[string]interface{})
		resourcedata.SetMapValueIfNotNil(chartsMap, "id", chart.Id)
		resourcedata.SetMapValueIfNotNil(chartsMap, "name", chart.Name)
		resourcedata.SetMapValueIfNotNil(chartsMap, "version", chart.Version)
		resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(chartsMap, "metrics", chart.Metrics, flattenMetrics)
		resourcedata.SetMapValueIfNotNil(chartsMap, "group_by_time", chart.GroupByTime)
		resourcedata.SetMapValueIfNotNil(chartsMap, "group_by_max", chart.GroupByMax)
		resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(chartsMap, "display_attributes", chart.DisplayAttributes, flattenDisplayAttributes)
		resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(chartsMap, "group_by_attributes", chart.GroupByAttributes, flattenGroupbyAttributes)
		chartsList = append(chartsList, chartsMap)
	}
	return chartsList
}

func flattenMetrics(metrics *[]platformclientv2.Journeyviewchartmetric) []interface{} {
	if len(*metrics) == 0 {
		return nil
	}
	var metricsList []interface{}
	for _, metric := range *metrics {
		metricsMap := make(map[string]interface{})
		resourcedata.SetMapValueIfNotNil(metricsMap, "id", metric.Id)
		resourcedata.SetMapValueIfNotNil(metricsMap, "element_id", metric.ElementId)
		resourcedata.SetMapValueIfNotNil(metricsMap, "aggregate", metric.Aggregate)
		resourcedata.SetMapValueIfNotNil(metricsMap, "display_label", metric.DisplayLabel)
		metricsList = append(metricsList, metricsMap)
	}
	return metricsList
}

func flattenDisplayAttributes(displayAttributes *platformclientv2.Journeyviewchartdisplayattributes) []interface{} {

	var displayAttributesList []interface{}
	displayAttributesMap := make(map[string]interface{})
	resourcedata.SetMapValueIfNotNil(displayAttributesMap, "metrics_title", displayAttributes.MetricsTitle)
	resourcedata.SetMapValueIfNotNil(displayAttributesMap, "group_by_title", displayAttributes.GroupByTitle)
	resourcedata.SetMapValueIfNotNil(displayAttributesMap, "var_type", displayAttributes.VarType)
	resourcedata.SetMapValueIfNotNil(displayAttributesMap, "show_legend", displayAttributes.ShowLegend)
	displayAttributesList = append(displayAttributesList, displayAttributesMap)

	return displayAttributesList
}

func flattenGroupbyAttributes(groupByAttributes *[]platformclientv2.Journeyviewchartgroupbyattribute) []interface{} {
	if len(*groupByAttributes) == 0 {
		return nil
	}
	var groupByAttributesList []interface{}
	for _, groupByAttribute := range *groupByAttributes {
		groupByAttributeMap := make(map[string]interface{})
		resourcedata.SetMapValueIfNotNil(groupByAttributeMap, "attribute", groupByAttribute.Attribute)
		resourcedata.SetMapValueIfNotNil(groupByAttributeMap, "element_id", groupByAttribute.ElementId)

		groupByAttributesList = append(groupByAttributesList, groupByAttributeMap)
	}
	return groupByAttributesList
}
