package outbound_contactlistfilter

import (
	"fmt"
	"strings"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"terraform-provider-genesyscloud/genesyscloud/util/resourcedata"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

func getContactlistfilterFromResourceData(d *schema.ResourceData) platformclientv2.Contactlistfilter {
	filter := platformclientv2.Contactlistfilter{
		Name:    platformclientv2.String(d.Get("name").(string)),
		Clauses: buildContactListFilterClauses(d.Get("clauses").([]interface{})),
	}
	contactList := util.BuildSdkDomainEntityRef(d, "contact_list_id")
	contactListTemplate := util.BuildSdkDomainEntityRef(d, "contact_list_template_id")

	if contactList != nil {
		filter.ContactList = contactList
		contactListSource := "ContactList"
		filter.SourceType = &contactListSource
	}
	if contactListTemplate != nil {
		filter.ContactListTemplate = contactListTemplate
		contactListTemplateSource := "ContactListTemplate"
		filter.SourceType = &contactListTemplateSource
	}

	filterType := d.Get("filter_type").(string)
	if filterType != "" {
		filter.FilterType = &filterType
	}

	return filter
}

func buildContactListFilterClauses(clauses []interface{}) *[]platformclientv2.Contactlistfilterclause {
	if clauses == nil || len(clauses) == 0 {
		return nil
	}

	sdkClauses := make([]platformclientv2.Contactlistfilterclause, 0)
	for _, clause := range clauses {
		var sdkClause platformclientv2.Contactlistfilterclause
		contactListFilterMap := clause.(map[string]interface{})

		resourcedata.BuildSDKStringValueIfNotNil(&sdkClause.FilterType, contactListFilterMap, "filter_type")
		if predicates := contactListFilterMap["predicates"]; predicates != nil {
			sdkClause.Predicates = buildContactListFilterPredicate(predicates.([]interface{}))
		}

		sdkClauses = append(sdkClauses, sdkClause)
	}

	return &sdkClauses
}

func buildContactListFilterPredicate(predicates []interface{}) *[]platformclientv2.Contactlistfilterpredicate {
	if predicates == nil || len(predicates) == 0 {
		return nil
	}

	sdkPredicates := make([]platformclientv2.Contactlistfilterpredicate, 0)
	for _, predicate := range predicates {
		if predicateMap, ok := predicate.(map[string]interface{}); ok {
			var sdkPredicate platformclientv2.Contactlistfilterpredicate

			resourcedata.BuildSDKStringValueIfNotNil(&sdkPredicate.Column, predicateMap, "column")
			resourcedata.BuildSDKStringValueIfNotNil(&sdkPredicate.ColumnType, predicateMap, "column_type")
			resourcedata.BuildSDKStringValueIfNotNil(&sdkPredicate.Operator, predicateMap, "operator")
			resourcedata.BuildSDKStringValueIfNotNil(&sdkPredicate.Value, predicateMap, "value")
			if varRangeSet := predicateMap["var_range"].(*schema.Set); varRangeSet != nil && len(varRangeSet.List()) > 0 {
				sdkPredicate.VarRange = buildContactListFilterRange(varRangeSet)
			}
			if inverted, ok := predicateMap["inverted"].(bool); ok {
				sdkPredicate.Inverted = &inverted
			}

			sdkPredicates = append(sdkPredicates, sdkPredicate)
		}
	}
	return &sdkPredicates
}

func buildContactListFilterRange(contactListFilterRange *schema.Set) *platformclientv2.Contactlistfilterrange {
	contactListFilterRangeList := contactListFilterRange.List()
	contactListFilterRangeMap := contactListFilterRangeList[0].(map[string]interface{})

	var sdkContactListFilterRange platformclientv2.Contactlistfilterrange
	resourcedata.BuildSDKStringValueIfNotNil(&sdkContactListFilterRange.Min, contactListFilterRangeMap, "min")
	resourcedata.BuildSDKStringValueIfNotNil(&sdkContactListFilterRange.Max, contactListFilterRangeMap, "max")
	if minInclusive, ok := contactListFilterRangeMap["min_inclusive"].(bool); ok {
		sdkContactListFilterRange.MinInclusive = &minInclusive
	}
	if maxInclusive, ok := contactListFilterRangeMap["max_inclusive"].(bool); ok {
		sdkContactListFilterRange.MaxInclusive = &maxInclusive
	}

	inSet := make([]string, 0)
	for _, v := range contactListFilterRangeMap["in_set"].([]interface{}) {
		inSet = append(inSet, v.(string))
	}
	sdkContactListFilterRange.InSet = &inSet

	return &sdkContactListFilterRange
}

func flattenContactListFilterClauses(contactListFilterClauses *[]platformclientv2.Contactlistfilterclause) []interface{} {
	if len(*contactListFilterClauses) == 0 {
		return nil
	}

	contactListFilterClauseList := make([]interface{}, 0)
	for _, contactListFilterClause := range *contactListFilterClauses {
		contactListFilterClauseMap := make(map[string]interface{})

		resourcedata.SetMapValueIfNotNil(contactListFilterClauseMap, "filter_type", contactListFilterClause.FilterType)
		resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(contactListFilterClauseMap, "predicates", contactListFilterClause.Predicates, flattenContactListFilterPredicates)

		contactListFilterClauseList = append(contactListFilterClauseList, contactListFilterClauseMap)
	}
	return contactListFilterClauseList
}

func flattenContactListFilterPredicates(contactListFilterPredicates *[]platformclientv2.Contactlistfilterpredicate) []interface{} {
	if len(*contactListFilterPredicates) == 0 {
		return nil
	}

	contactListFilterPredicateList := make([]interface{}, 0)
	for _, contactListFilterPredicate := range *contactListFilterPredicates {
		contactListFilterPredicateMap := make(map[string]interface{})

		resourcedata.SetMapValueIfNotNil(contactListFilterPredicateMap, "column", contactListFilterPredicate.Column)
		resourcedata.SetMapValueIfNotNil(contactListFilterPredicateMap, "column_type", contactListFilterPredicate.ColumnType)
		resourcedata.SetMapValueIfNotNil(contactListFilterPredicateMap, "operator", contactListFilterPredicate.Operator)
		resourcedata.SetMapValueIfNotNil(contactListFilterPredicateMap, "value", contactListFilterPredicate.Value)
		resourcedata.SetMapValueIfNotNil(contactListFilterPredicateMap, "inverted", contactListFilterPredicate.Inverted)
		resourcedata.SetMapSchemaSetWithFuncIfNotNil(contactListFilterPredicateMap, "var_range", contactListFilterPredicate.VarRange, flattenContactListFilterRange)

		contactListFilterPredicateList = append(contactListFilterPredicateList, contactListFilterPredicateMap)
	}
	return contactListFilterPredicateList
}

func flattenContactListFilterRange(contactListFilterRange *platformclientv2.Contactlistfilterrange) *schema.Set {
	if contactListFilterRange == nil {
		return nil
	}

	contactListFilterRangeSet := schema.NewSet(schema.HashResource(rangeResource), []interface{}{})
	contactListFilterRangeMap := make(map[string]interface{})

	resourcedata.SetMapValueIfNotNil(contactListFilterRangeMap, "min", contactListFilterRange.Min)
	resourcedata.SetMapValueIfNotNil(contactListFilterRangeMap, "max", contactListFilterRange.Max)
	resourcedata.SetMapValueIfNotNil(contactListFilterRangeMap, "min_inclusive", contactListFilterRange.MinInclusive)
	resourcedata.SetMapValueIfNotNil(contactListFilterRangeMap, "max_inclusive", contactListFilterRange.MaxInclusive)

	if contactListFilterRange.InSet != nil {
		// Changed []string to []interface{} to prevent type conversion panic
		inSet := make([]interface{}, 0)
		for _, v := range *contactListFilterRange.InSet {
			inSet = append(inSet, v)
		}
		contactListFilterRangeMap["in_set"] = inSet
	}
	if len(contactListFilterRangeMap) == 0 {
		return nil
	}
	contactListFilterRangeSet.Add(contactListFilterRangeMap)

	return contactListFilterRangeSet
}

func GenerateOutboundContactListFilter(
	resourceId string,
	name string,
	contactListId string,
	filterType string,
	nestedBlocks ...string,
) string {
	if filterType != "" {
		filterType = fmt.Sprintf(`filter_type = "%s"`, filterType)
	}
	return fmt.Sprintf(`
resource "genesyscloud_outbound_contactlistfilter" "%s" {
	name            = "%s"
	contact_list_id = %s
	%s
	%s
}
`, resourceId, name, contactListId, filterType, strings.Join(nestedBlocks, "\n"))
}

func GenerateOutboundContactListFilterClause(filterType string, nestedBlocks ...string) string {
	if filterType != "" {
		filterType = fmt.Sprintf(`filter_type = "%s"`, filterType)
	}
	return fmt.Sprintf(`
	clauses {
		%s
		%s
	}
`, filterType, strings.Join(nestedBlocks, "\n"))
}

func GenerateOutboundContactListFilterPredicates(
	column string,
	columnType string,
	operator string,
	value string,
	inverted string,
	varRangeBlock string,
) string {
	if column != "" {
		column = fmt.Sprintf(`column = "%s"`, column)
	}
	if columnType != "" {
		columnType = fmt.Sprintf(`column_type = "%s"`, columnType)
	}
	if operator != "" {
		operator = fmt.Sprintf(`operator = "%s"`, operator)
	}
	if value != "" {
		value = fmt.Sprintf(`value = "%s"`, value)
	}
	if inverted != "" {
		inverted = fmt.Sprintf(`inverted = %s`, inverted)
	}
	return fmt.Sprintf(`
		predicates {
			%s
			%s
			%s
			%s
			%s
			%s
		}
`, column, columnType, operator, value, inverted, varRangeBlock)
}
