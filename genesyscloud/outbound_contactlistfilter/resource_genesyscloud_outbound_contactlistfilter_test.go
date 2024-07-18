package outbound_contactlistfilter

import (
	"fmt"
	"strconv"
	"strings"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"

	obContactList "terraform-provider-genesyscloud/genesyscloud/outbound_contact_list"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

func TestAccResourceOutboundContactListFilter(t *testing.T) {

	t.Parallel()
	var (
		resourceId            = "contact_list_filter"
		name                  = "Test CLF " + uuid.NewString()
		contactListResourceId = "contact_list"
		contactListName       = "Test Contact List " + uuid.NewString()
		filterType            = "AND"
		column                = "Phone"
		columnType            = "numeric"
		operator              = "EQUALS"
		predicateValue        = "+12345123456"
		inverted              = util.FalseValue
		rangeMin              = "1"
		rangeMax              = "10"
		minInclusive          = util.TrueValue
		maxInclusive          = util.FalseValue
		rangeInSet            = []string{"a"}

		nameUpdated           = "Test CLF " + uuid.NewString()
		filterTypeUpdated     = "OR"
		columnUpdated         = "Zipcode"
		columnTypeUpdated     = "alphabetic"
		operatorUpdated       = "CONTAINS"
		predicateValueUpdated = "XYZ"
		invertedUpdated       = util.TrueValue
		rangeMinUpdated       = "2"
		rangeMaxUpdated       = "12"
		minInclusiveUpdated   = util.FalseValue
		maxInclusiveUpdated   = util.TrueValue
		rangeInSetUpdated     = []string{"a", "b"}
	)

	contactListResource := obContactList.GenerateOutboundContactList(
		contactListResourceId,
		contactListName,
		util.NullValue,
		util.NullValue,
		[]string{},
		[]string{strconv.Quote(column), strconv.Quote(columnUpdated)},
		util.NullValue,
		util.NullValue,
		util.NullValue,
		obContactList.GeneratePhoneColumnsBlock(
			column,
			"cell",
			util.NullValue,
		),
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: contactListResource + GenerateOutboundContactListFilter(
					resourceId,
					name,
					"genesyscloud_outbound_contact_list."+contactListResourceId+".id",
					"",
					GenerateOutboundContactListFilterClause(
						"",
						GenerateOutboundContactListFilterPredicates(
							column,
							columnType,
							operator,
							predicateValue,
							"",
							"",
						),
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_outbound_contactlistfilter."+resourceId, "name", name),
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_contactlistfilter."+resourceId, "contact_list_id",
						"genesyscloud_outbound_contact_list."+contactListResourceId, "id"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contactlistfilter."+resourceId, "clauses.0.predicates.0.column", column),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contactlistfilter."+resourceId, "clauses.0.predicates.0.column_type", columnType),
				),
			},
			{
				Config: contactListResource + GenerateOutboundContactListFilter(
					resourceId,
					name,
					"genesyscloud_outbound_contact_list."+contactListResourceId+".id",
					filterType,
					GenerateOutboundContactListFilterClause(
						filterType,
						GenerateOutboundContactListFilterPredicates(
							column,
							columnType,
							operator,
							predicateValue,
							inverted,
							generatePredicateVarRangeBlock(
								rangeMin,
								rangeMax,
								minInclusive,
								maxInclusive,
								rangeInSet,
							),
						),
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_outbound_contactlistfilter."+resourceId, "name", name),
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_contactlistfilter."+resourceId, "contact_list_id",
						"genesyscloud_outbound_contact_list."+contactListResourceId, "id"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contactlistfilter."+resourceId, "filter_type", filterType),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contactlistfilter."+resourceId, "clauses.0.filter_type", filterType),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contactlistfilter."+resourceId, "clauses.0.predicates.0.column", column),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contactlistfilter."+resourceId, "clauses.0.predicates.0.column_type", columnType),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contactlistfilter."+resourceId, "clauses.0.predicates.0.operator", operator),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contactlistfilter."+resourceId, "clauses.0.predicates.0.value", predicateValue),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contactlistfilter."+resourceId, "clauses.0.predicates.0.inverted", inverted),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contactlistfilter."+resourceId, "clauses.0.predicates.0.var_range.0.min", rangeMin),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contactlistfilter."+resourceId, "clauses.0.predicates.0.var_range.0.max", rangeMax),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contactlistfilter."+resourceId, "clauses.0.predicates.0.var_range.0.min_inclusive", minInclusive),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contactlistfilter."+resourceId, "clauses.0.predicates.0.var_range.0.max_inclusive", maxInclusive),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contactlistfilter."+resourceId, "clauses.0.predicates.0.var_range.0.in_set.0", rangeInSet[0]),
				),
			},
			{
				Config: contactListResource + GenerateOutboundContactListFilter(
					resourceId,
					nameUpdated,
					"genesyscloud_outbound_contact_list."+contactListResourceId+".id",
					filterTypeUpdated,
					GenerateOutboundContactListFilterClause(
						filterType,
						GenerateOutboundContactListFilterPredicates(
							column,
							columnType,
							operator,
							predicateValue,
							inverted,
							generatePredicateVarRangeBlock(
								rangeMin,
								rangeMax,
								minInclusive,
								maxInclusive,
								rangeInSet,
							),
						),
					),
					GenerateOutboundContactListFilterClause(
						filterTypeUpdated,
						GenerateOutboundContactListFilterPredicates(
							column,
							columnType,
							operator,
							predicateValue,
							inverted,
							"",
						),
						GenerateOutboundContactListFilterPredicates(
							columnUpdated,
							columnTypeUpdated,
							operatorUpdated,
							predicateValueUpdated,
							invertedUpdated,
							generatePredicateVarRangeBlock(
								rangeMinUpdated,
								rangeMaxUpdated,
								minInclusiveUpdated,
								maxInclusiveUpdated,
								rangeInSetUpdated,
							),
						),
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_outbound_contactlistfilter."+resourceId, "name", nameUpdated),
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_contactlistfilter."+resourceId, "contact_list_id",
						"genesyscloud_outbound_contact_list."+contactListResourceId, "id"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contactlistfilter."+resourceId, "filter_type", filterTypeUpdated),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contactlistfilter."+resourceId, "clauses.0.filter_type", filterType),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contactlistfilter."+resourceId, "clauses.0.predicates.0.column", column),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contactlistfilter."+resourceId, "clauses.0.predicates.0.column_type", columnType),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contactlistfilter."+resourceId, "clauses.0.predicates.0.operator", operator),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contactlistfilter."+resourceId, "clauses.0.predicates.0.value", predicateValue),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contactlistfilter."+resourceId, "clauses.0.predicates.0.inverted", inverted),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contactlistfilter."+resourceId, "clauses.0.predicates.0.var_range.0.min", rangeMin),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contactlistfilter."+resourceId, "clauses.0.predicates.0.var_range.0.max", rangeMax),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contactlistfilter."+resourceId, "clauses.0.predicates.0.var_range.0.min_inclusive", minInclusive),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contactlistfilter."+resourceId, "clauses.0.predicates.0.var_range.0.max_inclusive", maxInclusive),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contactlistfilter."+resourceId, "clauses.0.predicates.0.var_range.0.in_set.0", rangeInSet[0]),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contactlistfilter."+resourceId, "clauses.1.filter_type", filterTypeUpdated),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contactlistfilter."+resourceId, "clauses.1.predicates.0.column", column),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contactlistfilter."+resourceId, "clauses.1.predicates.0.column_type", columnType),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contactlistfilter."+resourceId, "clauses.1.predicates.0.operator", operator),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contactlistfilter."+resourceId, "clauses.1.predicates.0.value", predicateValue),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contactlistfilter."+resourceId, "clauses.1.predicates.0.inverted", inverted),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contactlistfilter."+resourceId, "clauses.1.predicates.1.column", columnUpdated),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contactlistfilter."+resourceId, "clauses.1.predicates.1.column_type", columnTypeUpdated),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contactlistfilter."+resourceId, "clauses.1.predicates.1.operator", operatorUpdated),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contactlistfilter."+resourceId, "clauses.1.predicates.1.value", predicateValueUpdated),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contactlistfilter."+resourceId, "clauses.1.predicates.1.inverted", invertedUpdated),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contactlistfilter."+resourceId, "clauses.1.predicates.1.var_range.0.min", rangeMinUpdated),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contactlistfilter."+resourceId, "clauses.1.predicates.1.var_range.0.max", rangeMaxUpdated),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contactlistfilter."+resourceId, "clauses.1.predicates.1.var_range.0.min_inclusive", minInclusiveUpdated),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contactlistfilter."+resourceId, "clauses.1.predicates.1.var_range.0.max_inclusive", maxInclusiveUpdated),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contactlistfilter."+resourceId, "clauses.1.predicates.1.var_range.0.in_set.0", rangeInSetUpdated[0]),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contactlistfilter."+resourceId, "clauses.1.predicates.1.var_range.0.in_set.1", rangeInSetUpdated[1]),
				),
			},
			{
				Config: contactListResource + GenerateOutboundContactListFilter(
					resourceId,
					name,
					"genesyscloud_outbound_contact_list."+contactListResourceId+".id",
					"",
					GenerateOutboundContactListFilterClause(
						"",
						GenerateOutboundContactListFilterPredicates(
							column,
							columnType,
							operator,
							predicateValue,
							"",
							"",
						),
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_outbound_contactlistfilter."+resourceId, "name", name),
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_contactlistfilter."+resourceId, "contact_list_id",
						"genesyscloud_outbound_contact_list."+contactListResourceId, "id"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contactlistfilter."+resourceId, "clauses.0.predicates.0.column", column),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contactlistfilter."+resourceId, "clauses.0.predicates.0.column_type", columnType),
				),
			},
			{
				ResourceName:      "genesyscloud_outbound_contactlistfilter." + resourceId,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifyOutboundContactListFilterDestroyed,
	})
}

func generatePredicateVarRangeBlock(
	min string,
	max string,
	minInclusive string,
	maxInclusive string,
	inSet []string) string {
	var inSetQuoted []string
	for _, v := range inSet {
		inSetQuoted = append(inSetQuoted, strconv.Quote(v))
	}
	if min != "" {
		min = fmt.Sprintf(`min = "%s"`, min)
	}
	if max != "" {
		max = fmt.Sprintf(`max = "%s"`, max)
	}
	if minInclusive != "" {
		minInclusive = fmt.Sprintf(`min_inclusive = %s`, minInclusive)
	}
	if maxInclusive != "" {
		maxInclusive = fmt.Sprintf(`max_inclusive = %s`, maxInclusive)
	}
	return fmt.Sprintf(`
			var_range {
				%s
				%s
				%s
				%s
				in_set = [%s]
			}
`, min, max, minInclusive, maxInclusive, strings.Join(inSetQuoted, ", "))
}

func testVerifyOutboundContactListFilterDestroyed(state *terraform.State) error {
	outboundAPI := platformclientv2.NewOutboundApi()
	for _, rs := range state.RootModule().Resources {
		if rs.Type != "genesyscloud_outbound_contactlistfilter" {
			continue
		}

		contactListFilter, resp, err := outboundAPI.GetOutboundContactlistfilter(rs.Primary.ID)
		if contactListFilter != nil {
			return fmt.Errorf("contact list filter (%s) still exists", rs.Primary.ID)
		} else if util.IsStatus404(resp) {
			// Contact list filter not found as expected
			continue
		} else {
			// Unexpected error
			return fmt.Errorf("unexpected error: %s", err)
		}
	}
	// Success. All contact list filters destroyed
	return nil
}
