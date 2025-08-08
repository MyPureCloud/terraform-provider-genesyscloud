package journey_views

import (
	"fmt"
	"strings"
	"testing"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v165/platformclientv2"
)

func TestAccResourceJourneyViewsBasic(t *testing.T) {
	var (
		name                        = "test journey from tf Nicolas"
		nameUpdated                 = "test journey from tf 1 updated"
		duration                    = "P1Y"
		elementId                   = "ac6c61b5-1cd4-4c6e-a8a5-edb74d9117eb"
		elementName                 = "Wrap Up"
		attributeType               = "Event"
		attributeId                 = "a416328b-167c-0365-d0e1-f072cd5d4ded"
		attributeSource             = "Voice"
		elementDisplayAttributesX   = 601
		elementDisplayAttributesY   = 240
		elementDisplayAttributesCol = 1
		filterType                  = "And"
		predicatesDimension         = "mediaType"
		predicatesValues            = "VOICE"
		predicatesOperator          = "Matches"
		predicatesNoValue           = false
		numberPredicatesDimension   = "wrapupDurationMs"
		numberPredicatesOperator    = "Matches"
		numberPredicatesNoValue     = false
		numberPredicatesRange       = []string{"lt:duration:P1D", "lte:duration:P1Y"}
		//Element 2
		element2Id                    = "ac6c61b5-1cd4-4c6e-a8a5-edb74d9aaaaaa"
		element2Name                  = "Bot End"
		attribute2Type                = "Event"
		attribute2Id                  = "b56bd9cc-74a9-1f0e-9ecf-494b4024a3be"
		numberPredicatesDimension2    = "queryCount"
		numberPredicatesRange2        = []string{"eq:number:1000", "neq:number:2000", "gte:number:3000", "gt:number:4000"}
		journeyResourceLabel          = "journey_resource1"
		chartName                     = "Chart 1"
		chartName2                    = "Chart 2"
		chartVersion                  = 1
		metricId                      = "Metric 1"
		metricDisplayLabel            = "Display Label"
		metricAggregate               = "CustomerCount"
		chartGroupByTime              = "Day"
		chartGroupByMax               = 1
		displayAttributesVarType      = "Column"
		displayAttributesGroupByTitle = "Group By Title"
		displayAttributesMetricsTitle = "Metrics Title"
		displayAttributesShowLegend   = false
		groupByAttributesAttribute    = "queueId"
	)
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, nil),
		Steps: []resource.TestStep{
			{
				//Create
				Config: generateJourneyView(journeyResourceLabel, name, duration,
					generateObjectsList([]string{
						generateElements(
							elementId,
							elementName,
							generateAttributes(attributeType, attributeId, attributeSource),
							generateElementDisplayAttributes(elementDisplayAttributesX, elementDisplayAttributesY, elementDisplayAttributesCol),
							generateFilter(filterType,
								generatePredicates(predicatesDimension, predicatesValues, predicatesOperator, predicatesNoValue),
								generateNumberPredicates(numberPredicatesDimension, numberPredicatesRange, numberPredicatesOperator, numberPredicatesNoValue)),
							generateFollowedBy(element2Id)),
						generateElements(
							element2Id,
							element2Name,
							generateAttributes(attribute2Type, attribute2Id, attributeSource),
							"",
							generateFilter(filterType,
								"", //	generatePredicates(predicatesDimension, predicatesValues, predicatesOperator, predicatesNoValue),
								generateNumberPredicates(numberPredicatesDimension2, numberPredicatesRange2, numberPredicatesOperator, numberPredicatesNoValue)),
							""),
					}),
					generateObjectsList([]string{
						generateCharts(chartName, chartVersion,
							generateMetrics(metricId, elementId, metricAggregate, metricDisplayLabel), chartGroupByTime, chartGroupByMax,
							generateDisplayAttributes(displayAttributesVarType, displayAttributesGroupByTitle, displayAttributesMetricsTitle, displayAttributesShowLegend),
							""),
						generateCharts(chartName2, chartVersion,
							generateMetrics(metricId, elementId, metricAggregate, metricDisplayLabel), "", chartGroupByMax,
							generateDisplayAttributes(displayAttributesVarType, displayAttributesGroupByTitle, displayAttributesMetricsTitle, displayAttributesShowLegend),
							generateGroupeByAttributes(elementId, groupByAttributesAttribute)),
					}),
				),

				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_journey_views."+journeyResourceLabel, "name", name),
					resource.TestCheckResourceAttr("genesyscloud_journey_views."+journeyResourceLabel, "duration", duration),
					resource.TestCheckResourceAttr("genesyscloud_journey_views."+journeyResourceLabel, "elements.#", "2"),
					resource.TestCheckResourceAttr("genesyscloud_journey_views."+journeyResourceLabel, "elements.0.id", elementId),
					resource.TestCheckResourceAttr("genesyscloud_journey_views."+journeyResourceLabel, "elements.0.name", elementName),
					resource.TestCheckResourceAttr("genesyscloud_journey_views."+journeyResourceLabel, "elements.0.attributes.#", "1"),
					resource.TestCheckResourceAttr("genesyscloud_journey_views."+journeyResourceLabel, "elements.0.attributes.0.type", attributeType),
					resource.TestCheckResourceAttr("genesyscloud_journey_views."+journeyResourceLabel, "elements.0.attributes.0.id", attributeId),
					resource.TestCheckResourceAttr("genesyscloud_journey_views."+journeyResourceLabel, "elements.0.attributes.0.source", attributeSource),
					resource.TestCheckResourceAttr("genesyscloud_journey_views."+journeyResourceLabel, "elements.0.display_attributes.#", "1"),
					resource.TestCheckResourceAttr("genesyscloud_journey_views."+journeyResourceLabel, "elements.0.display_attributes.0.x", fmt.Sprintf("%d", elementDisplayAttributesX)),
					resource.TestCheckResourceAttr("genesyscloud_journey_views."+journeyResourceLabel, "elements.0.display_attributes.0.y", fmt.Sprintf("%d", elementDisplayAttributesY)),
					resource.TestCheckResourceAttr("genesyscloud_journey_views."+journeyResourceLabel, "elements.0.display_attributes.0.col", fmt.Sprintf("%d", elementDisplayAttributesCol)),
					resource.TestCheckResourceAttr("genesyscloud_journey_views."+journeyResourceLabel, "elements.0.filter.#", "1"),
					resource.TestCheckResourceAttr("genesyscloud_journey_views."+journeyResourceLabel, "elements.0.filter.0.type", "And"),
					resource.TestCheckResourceAttr("genesyscloud_journey_views."+journeyResourceLabel, "elements.0.filter.0.predicates.#", "1"),
					resource.TestCheckResourceAttr("genesyscloud_journey_views."+journeyResourceLabel, "elements.0.filter.0.predicates.0.dimension", predicatesDimension),
					resource.TestCheckResourceAttr("genesyscloud_journey_views."+journeyResourceLabel, "elements.0.filter.0.predicates.0.values.#", "1"),
					resource.TestCheckResourceAttr("genesyscloud_journey_views."+journeyResourceLabel, "elements.0.filter.0.predicates.0.values.0", predicatesValues),
					resource.TestCheckResourceAttr("genesyscloud_journey_views."+journeyResourceLabel, "elements.0.filter.0.predicates.0.operator", predicatesOperator),
					resource.TestCheckResourceAttr("genesyscloud_journey_views."+journeyResourceLabel, "elements.0.filter.0.predicates.0.no_value", fmt.Sprintf("%t", predicatesNoValue)),
					resource.TestCheckResourceAttr("genesyscloud_journey_views."+journeyResourceLabel, "elements.0.filter.0.number_predicates.#", "1"),
					resource.TestCheckResourceAttr("genesyscloud_journey_views."+journeyResourceLabel, "elements.0.filter.0.number_predicates.0.dimension", numberPredicatesDimension),
					resource.TestCheckResourceAttr("genesyscloud_journey_views."+journeyResourceLabel, "elements.0.filter.0.number_predicates.0.range.#", "1"),
					resource.TestCheckResourceAttr("genesyscloud_journey_views."+journeyResourceLabel, "elements.0.filter.0.number_predicates.0.range.0.lt.0.duration", "P1D"),
					resource.TestCheckResourceAttr("genesyscloud_journey_views."+journeyResourceLabel, "elements.0.filter.0.number_predicates.0.range.0.lte.0.duration", "P1Y"),
					resource.TestCheckResourceAttr("genesyscloud_journey_views."+journeyResourceLabel, "elements.0.filter.0.number_predicates.0.operator", numberPredicatesOperator),
					resource.TestCheckResourceAttr("genesyscloud_journey_views."+journeyResourceLabel, "elements.0.filter.0.number_predicates.0.no_value", fmt.Sprintf("%t", numberPredicatesNoValue)),
					resource.TestCheckResourceAttr("genesyscloud_journey_views."+journeyResourceLabel, "elements.1.id", element2Id),
					resource.TestCheckResourceAttr("genesyscloud_journey_views."+journeyResourceLabel, "elements.1.name", element2Name),
					resource.TestCheckResourceAttr("genesyscloud_journey_views."+journeyResourceLabel, "elements.1.attributes.#", "1"),
					resource.TestCheckResourceAttr("genesyscloud_journey_views."+journeyResourceLabel, "elements.1.attributes.0.type", attribute2Type),
					resource.TestCheckResourceAttr("genesyscloud_journey_views."+journeyResourceLabel, "elements.1.attributes.0.id", attribute2Id),
					resource.TestCheckResourceAttr("genesyscloud_journey_views."+journeyResourceLabel, "elements.1.attributes.0.source", attributeSource),
					resource.TestCheckResourceAttr("genesyscloud_journey_views."+journeyResourceLabel, "elements.1.filter.#", "1"),
					resource.TestCheckResourceAttr("genesyscloud_journey_views."+journeyResourceLabel, "elements.1.filter.0.type", "And"),
					resource.TestCheckResourceAttr("genesyscloud_journey_views."+journeyResourceLabel, "elements.1.filter.0.number_predicates.#", "1"),
					resource.TestCheckResourceAttr("genesyscloud_journey_views."+journeyResourceLabel, "elements.1.filter.0.number_predicates.0.dimension", numberPredicatesDimension2),
					resource.TestCheckResourceAttr("genesyscloud_journey_views."+journeyResourceLabel, "elements.1.filter.0.number_predicates.0.range.#", "1"),
					resource.TestCheckResourceAttr("genesyscloud_journey_views."+journeyResourceLabel, "elements.1.filter.0.number_predicates.0.range.0.eq.0.number", "1000"),
					resource.TestCheckResourceAttr("genesyscloud_journey_views."+journeyResourceLabel, "elements.1.filter.0.number_predicates.0.range.0.neq.0.number", "2000"),
					resource.TestCheckResourceAttr("genesyscloud_journey_views."+journeyResourceLabel, "elements.1.filter.0.number_predicates.0.range.0.gte.0.number", "3000"),
					resource.TestCheckResourceAttr("genesyscloud_journey_views."+journeyResourceLabel, "elements.1.filter.0.number_predicates.0.range.0.gt.0.number", "4000"),
					resource.TestCheckResourceAttr("genesyscloud_journey_views."+journeyResourceLabel, "elements.1.filter.0.number_predicates.0.operator", numberPredicatesOperator),
					resource.TestCheckResourceAttr("genesyscloud_journey_views."+journeyResourceLabel, "elements.1.filter.0.number_predicates.0.no_value", fmt.Sprintf("%t", numberPredicatesNoValue)),
					resource.TestCheckResourceAttr("genesyscloud_journey_views."+journeyResourceLabel, "charts.#", "2"),
					resource.TestCheckResourceAttr("genesyscloud_journey_views."+journeyResourceLabel, "charts.0.name", chartName),
					resource.TestCheckResourceAttr("genesyscloud_journey_views."+journeyResourceLabel, "charts.0.version", fmt.Sprintf("%v", chartVersion)),
					resource.TestCheckResourceAttr("genesyscloud_journey_views."+journeyResourceLabel, "charts.0.metrics.#", "1"),
					resource.TestCheckResourceAttr("genesyscloud_journey_views."+journeyResourceLabel, "charts.0.metrics.0.id", metricId),
					resource.TestCheckResourceAttr("genesyscloud_journey_views."+journeyResourceLabel, "charts.0.metrics.0.display_label", metricDisplayLabel),
					resource.TestCheckResourceAttr("genesyscloud_journey_views."+journeyResourceLabel, "charts.0.metrics.0.aggregate", metricAggregate),
					resource.TestCheckResourceAttr("genesyscloud_journey_views."+journeyResourceLabel, "charts.0.metrics.0.element_id", elementId),
					resource.TestCheckResourceAttr("genesyscloud_journey_views."+journeyResourceLabel, "charts.0.group_by_time", chartGroupByTime),
					resource.TestCheckResourceAttr("genesyscloud_journey_views."+journeyResourceLabel, "charts.0.group_by_max", fmt.Sprintf("%d", chartGroupByMax)),
					resource.TestCheckResourceAttr("genesyscloud_journey_views."+journeyResourceLabel, "charts.0.display_attributes.#", "1"),
					resource.TestCheckResourceAttr("genesyscloud_journey_views."+journeyResourceLabel, "charts.0.display_attributes.0.var_type", displayAttributesVarType),
					resource.TestCheckResourceAttr("genesyscloud_journey_views."+journeyResourceLabel, "charts.0.display_attributes.0.group_by_title", displayAttributesGroupByTitle),
					resource.TestCheckResourceAttr("genesyscloud_journey_views."+journeyResourceLabel, "charts.0.display_attributes.0.metrics_title", displayAttributesMetricsTitle),
					resource.TestCheckResourceAttr("genesyscloud_journey_views."+journeyResourceLabel, "charts.0.display_attributes.0.show_legend", fmt.Sprintf("%v", displayAttributesShowLegend)),
					resource.TestCheckResourceAttr("genesyscloud_journey_views."+journeyResourceLabel, "charts.0.group_by_attributes.#", "0"),
					resource.TestCheckResourceAttr("genesyscloud_journey_views."+journeyResourceLabel, "charts.1.name", chartName2),
					resource.TestCheckResourceAttr("genesyscloud_journey_views."+journeyResourceLabel, "charts.1.group_by_attributes.#", "1"),
					resource.TestCheckResourceAttr("genesyscloud_journey_views."+journeyResourceLabel, "charts.1.group_by_attributes.0.element_id", elementId),
					resource.TestCheckResourceAttr("genesyscloud_journey_views."+journeyResourceLabel, "charts.1.group_by_attributes.0.attribute", groupByAttributesAttribute),
				),
			},
			{
				//update
				//Remove Chart2
				Config: generateJourneyView(journeyResourceLabel, nameUpdated, duration,
					generateObjectsList([]string{
						generateElements(
							elementId,
							elementName,
							generateAttributes(attributeType, attributeId, attributeSource),
							generateElementDisplayAttributes(elementDisplayAttributesX, elementDisplayAttributesY, elementDisplayAttributesCol),
							generateFilter(filterType,
								generatePredicates(predicatesDimension, predicatesValues, predicatesOperator, predicatesNoValue),
								generateNumberPredicates(numberPredicatesDimension, numberPredicatesRange, numberPredicatesOperator, numberPredicatesNoValue)),
							generateFollowedBy(element2Id)),
						generateElements(
							element2Id,
							element2Name,
							generateAttributes(attribute2Type, attribute2Id, attributeSource),
							"",
							generateFilter(filterType,
								"", //	generatePredicates(predicatesDimension, predicatesValues, predicatesOperator, predicatesNoValue),
								generateNumberPredicates(numberPredicatesDimension2, numberPredicatesRange2, numberPredicatesOperator, numberPredicatesNoValue)),
							""),
					}),
					generateCharts(chartName, chartVersion,
						generateMetrics(metricId, elementId, metricAggregate, metricDisplayLabel), chartGroupByTime, chartGroupByMax,
						generateDisplayAttributes(displayAttributesVarType, displayAttributesGroupByTitle, displayAttributesMetricsTitle, displayAttributesShowLegend),
						""),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_journey_views."+journeyResourceLabel, "name", nameUpdated),
					resource.TestCheckResourceAttr("genesyscloud_journey_views."+journeyResourceLabel, "duration", duration),
					resource.TestCheckResourceAttr("genesyscloud_journey_views."+journeyResourceLabel, "elements.#", "2"),
					resource.TestCheckResourceAttr("genesyscloud_journey_views."+journeyResourceLabel, "elements.0.id", elementId),
					resource.TestCheckResourceAttr("genesyscloud_journey_views."+journeyResourceLabel, "elements.0.name", elementName),
					resource.TestCheckResourceAttr("genesyscloud_journey_views."+journeyResourceLabel, "elements.0.attributes.#", "1"),
					resource.TestCheckResourceAttr("genesyscloud_journey_views."+journeyResourceLabel, "elements.0.attributes.0.type", attributeType),
					resource.TestCheckResourceAttr("genesyscloud_journey_views."+journeyResourceLabel, "elements.0.attributes.0.id", attributeId),
					resource.TestCheckResourceAttr("genesyscloud_journey_views."+journeyResourceLabel, "elements.0.attributes.0.source", attributeSource),
					resource.TestCheckResourceAttr("genesyscloud_journey_views."+journeyResourceLabel, "elements.0.display_attributes.#", "1"),
					resource.TestCheckResourceAttr("genesyscloud_journey_views."+journeyResourceLabel, "elements.0.display_attributes.0.x", fmt.Sprintf("%d", elementDisplayAttributesX)),
					resource.TestCheckResourceAttr("genesyscloud_journey_views."+journeyResourceLabel, "elements.0.display_attributes.0.y", fmt.Sprintf("%d", elementDisplayAttributesY)),
					resource.TestCheckResourceAttr("genesyscloud_journey_views."+journeyResourceLabel, "elements.0.display_attributes.0.col", fmt.Sprintf("%d", elementDisplayAttributesCol)),
					resource.TestCheckResourceAttr("genesyscloud_journey_views."+journeyResourceLabel, "elements.0.filter.#", "1"),
					resource.TestCheckResourceAttr("genesyscloud_journey_views."+journeyResourceLabel, "elements.0.filter.0.type", "And"),
					resource.TestCheckResourceAttr("genesyscloud_journey_views."+journeyResourceLabel, "elements.0.filter.0.predicates.#", "1"),
					resource.TestCheckResourceAttr("genesyscloud_journey_views."+journeyResourceLabel, "elements.0.filter.0.predicates.0.dimension", predicatesDimension),
					resource.TestCheckResourceAttr("genesyscloud_journey_views."+journeyResourceLabel, "elements.0.filter.0.predicates.0.values.#", "1"),
					resource.TestCheckResourceAttr("genesyscloud_journey_views."+journeyResourceLabel, "elements.0.filter.0.predicates.0.values.0", predicatesValues),
					resource.TestCheckResourceAttr("genesyscloud_journey_views."+journeyResourceLabel, "elements.0.filter.0.predicates.0.operator", predicatesOperator),
					resource.TestCheckResourceAttr("genesyscloud_journey_views."+journeyResourceLabel, "elements.0.filter.0.predicates.0.no_value", fmt.Sprintf("%t", predicatesNoValue)),
					resource.TestCheckResourceAttr("genesyscloud_journey_views."+journeyResourceLabel, "elements.0.filter.0.number_predicates.#", "1"),
					resource.TestCheckResourceAttr("genesyscloud_journey_views."+journeyResourceLabel, "elements.0.filter.0.number_predicates.0.dimension", numberPredicatesDimension),
					resource.TestCheckResourceAttr("genesyscloud_journey_views."+journeyResourceLabel, "elements.0.filter.0.number_predicates.0.range.#", "1"),
					resource.TestCheckResourceAttr("genesyscloud_journey_views."+journeyResourceLabel, "elements.0.filter.0.number_predicates.0.range.0.lt.0.duration", "P1D"),
					resource.TestCheckResourceAttr("genesyscloud_journey_views."+journeyResourceLabel, "elements.0.filter.0.number_predicates.0.range.0.lte.0.duration", "P1Y"),
					resource.TestCheckResourceAttr("genesyscloud_journey_views."+journeyResourceLabel, "elements.0.filter.0.number_predicates.0.operator", numberPredicatesOperator),
					resource.TestCheckResourceAttr("genesyscloud_journey_views."+journeyResourceLabel, "elements.0.filter.0.number_predicates.0.no_value", fmt.Sprintf("%t", numberPredicatesNoValue)),
					resource.TestCheckResourceAttr("genesyscloud_journey_views."+journeyResourceLabel, "elements.1.id", element2Id),
					resource.TestCheckResourceAttr("genesyscloud_journey_views."+journeyResourceLabel, "elements.1.name", element2Name),
					resource.TestCheckResourceAttr("genesyscloud_journey_views."+journeyResourceLabel, "elements.1.attributes.#", "1"),
					resource.TestCheckResourceAttr("genesyscloud_journey_views."+journeyResourceLabel, "elements.1.attributes.0.type", attribute2Type),
					resource.TestCheckResourceAttr("genesyscloud_journey_views."+journeyResourceLabel, "elements.1.attributes.0.id", attribute2Id),
					resource.TestCheckResourceAttr("genesyscloud_journey_views."+journeyResourceLabel, "elements.1.attributes.0.source", attributeSource),
					resource.TestCheckResourceAttr("genesyscloud_journey_views."+journeyResourceLabel, "elements.1.filter.#", "1"),
					resource.TestCheckResourceAttr("genesyscloud_journey_views."+journeyResourceLabel, "elements.1.filter.0.type", "And"),
					resource.TestCheckResourceAttr("genesyscloud_journey_views."+journeyResourceLabel, "elements.1.filter.0.number_predicates.#", "1"),
					resource.TestCheckResourceAttr("genesyscloud_journey_views."+journeyResourceLabel, "elements.1.filter.0.number_predicates.0.dimension", numberPredicatesDimension2),
					resource.TestCheckResourceAttr("genesyscloud_journey_views."+journeyResourceLabel, "elements.1.filter.0.number_predicates.0.range.#", "1"),
					resource.TestCheckResourceAttr("genesyscloud_journey_views."+journeyResourceLabel, "elements.1.filter.0.number_predicates.0.range.0.eq.0.number", "1000"),
					resource.TestCheckResourceAttr("genesyscloud_journey_views."+journeyResourceLabel, "elements.1.filter.0.number_predicates.0.range.0.neq.0.number", "2000"),
					resource.TestCheckResourceAttr("genesyscloud_journey_views."+journeyResourceLabel, "elements.1.filter.0.number_predicates.0.range.0.gte.0.number", "3000"),
					resource.TestCheckResourceAttr("genesyscloud_journey_views."+journeyResourceLabel, "elements.1.filter.0.number_predicates.0.range.0.gt.0.number", "4000"),
					resource.TestCheckResourceAttr("genesyscloud_journey_views."+journeyResourceLabel, "elements.1.filter.0.number_predicates.0.operator", numberPredicatesOperator),
					resource.TestCheckResourceAttr("genesyscloud_journey_views."+journeyResourceLabel, "elements.1.filter.0.number_predicates.0.no_value", fmt.Sprintf("%t", numberPredicatesNoValue)),
					resource.TestCheckResourceAttr("genesyscloud_journey_views."+journeyResourceLabel, "charts.#", "1"),
					resource.TestCheckResourceAttr("genesyscloud_journey_views."+journeyResourceLabel, "charts.0.name", chartName),
					resource.TestCheckResourceAttr("genesyscloud_journey_views."+journeyResourceLabel, "charts.0.version", fmt.Sprintf("%v", chartVersion)),
					resource.TestCheckResourceAttr("genesyscloud_journey_views."+journeyResourceLabel, "charts.0.metrics.#", "1"),
					resource.TestCheckResourceAttr("genesyscloud_journey_views."+journeyResourceLabel, "charts.0.metrics.0.id", metricId),
					resource.TestCheckResourceAttr("genesyscloud_journey_views."+journeyResourceLabel, "charts.0.metrics.0.display_label", metricDisplayLabel),
					resource.TestCheckResourceAttr("genesyscloud_journey_views."+journeyResourceLabel, "charts.0.metrics.0.aggregate", metricAggregate),
					resource.TestCheckResourceAttr("genesyscloud_journey_views."+journeyResourceLabel, "charts.0.metrics.0.element_id", elementId),
					resource.TestCheckResourceAttr("genesyscloud_journey_views."+journeyResourceLabel, "charts.0.group_by_time", chartGroupByTime),
					resource.TestCheckResourceAttr("genesyscloud_journey_views."+journeyResourceLabel, "charts.0.group_by_max", fmt.Sprintf("%d", chartGroupByMax)),
					resource.TestCheckResourceAttr("genesyscloud_journey_views."+journeyResourceLabel, "charts.0.display_attributes.#", "1"),
					resource.TestCheckResourceAttr("genesyscloud_journey_views."+journeyResourceLabel, "charts.0.display_attributes.0.var_type", displayAttributesVarType),
					resource.TestCheckResourceAttr("genesyscloud_journey_views."+journeyResourceLabel, "charts.0.display_attributes.0.group_by_title", displayAttributesGroupByTitle),
					resource.TestCheckResourceAttr("genesyscloud_journey_views."+journeyResourceLabel, "charts.0.display_attributes.0.metrics_title", displayAttributesMetricsTitle),
					resource.TestCheckResourceAttr("genesyscloud_journey_views."+journeyResourceLabel, "charts.0.display_attributes.0.show_legend", fmt.Sprintf("%v", displayAttributesShowLegend)),
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_journey_views." + journeyResourceLabel,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifyJourneyViewsDestroyed,
	})

}

func generateJourneyView(journeyResourceLabel string, name string, duration string, elementsBlock string, chartsBlock string) string {
	return fmt.Sprintf(`resource "genesyscloud_journey_views" "%s" {
    duration = "%s"
    name = "%s"
    %s
	%s
	}
	`, journeyResourceLabel, duration, name, func() string {
		if elementsBlock != "" {
			return elementsBlock
		}
		return ""
	}(),
		func() string {
			if chartsBlock != "" {
				return chartsBlock
			}
			return ""
		}())
}

func generateElements(id string, name string, attributesBlock string, displayAttributesBlock, filter string, followedBy string) string {
	return fmt.Sprintf(`
    elements {
        id = "%s"
        name = "%s"
        %s
		%s
        %s
		%s
    }
    `, id, name, attributesBlock, displayAttributesBlock, filter, followedBy)
}

func generateFilter(filterType string, nestedBlocks ...string) string {
	return fmt.Sprintf(`
        filter {
            type       = "%s"
            %s
        }
        `, filterType, strings.Join(nestedBlocks, "\n"))
}

func generateElementDisplayAttributes(attributeX int, attributeY int, attributeCol int) string {
	return fmt.Sprintf(`
        display_attributes {
            x   = %d
            y   = %d
            col = %d
        }
        `, attributeX, attributeY, attributeCol)
}

func generateAttributes(attributeType string, attributeId string, attributeSource string) string {
	return fmt.Sprintf(`
        attributes {
            type   = "%s"
            id     = "%s"
            source = "%s"
        }
        `, attributeType, attributeId, attributeSource)
}

func generateFollowedBy(id string) string {
	return fmt.Sprintf(`
        followed_by {
            id   = "%s"
        }
        `, id)
}

func generatePredicates(dimension string, values string, operator string, noValue bool) string {
	return fmt.Sprintf(`
            predicates  {
                dimension = "%s"
                values    = ["%s"]
                operator  = "%s"
                no_value  = %v
            }
            `, dimension, values, operator, noValue)
}

//Nicolas

func generateRange(properties []string) string {
	var result strings.Builder
	result.WriteString("{")

	for _, condition := range properties {
		parts := strings.Split(condition, ":")
		if len(parts) == 3 {
			// Remove quotes if present and trim spaces
			operator := strings.Trim(parts[0], "\"")
			property := strings.Trim(parts[1], "\"")
			value := strings.Trim(parts[2], "\"")
			var valueObj = fmt.Sprintf(`
            %s  {
                  %s = %s
            }
            `, operator, property, func() string {
				if property == "number" {
					return value
				}
				return fmt.Sprintf(`"%s"`, value)
			}())

			result.WriteString(fmt.Sprintf("    %s", valueObj))
		}
	}

	result.WriteString("}")
	return result.String()
}

func generateNumberPredicates(dimension string, rangeValues []string, operator string, noValue bool) string {

	rangeObj := generateRange(rangeValues)
	//rangeObj = strings.ReplaceAll(rangeObj, "\n", "")

	return fmt.Sprintf(`
            number_predicates  {
                dimension = "%s"
                range  %s
                operator  = "%s"
                no_value  = %v
            }
            `, dimension, rangeObj, operator, noValue)
}

func generateObjectsList(objs []string) string {
	return strings.Join(objs, "")
}

func generateCharts(name string, version int, metricsBlock string, groupByTime string, groupByMax int,
	displayAttributesBlock string, groupByAttributesblock string) string {
	return fmt.Sprintf(`
    charts {
        name = "%s"
        version = %d
        %s
		%s
		%s
		%s
		%s
    }
    `, name, version, metricsBlock,
		func() string {
			if groupByTime != "" {
				return fmt.Sprintf(`group_by_time = "%s"`, groupByTime)
			}
			return ""
		}(),
		func() string {
			if groupByMax != 0 {
				return fmt.Sprintf(`group_by_max = %d`, groupByMax)
			}
			return ""
		}(),
		displayAttributesBlock, groupByAttributesblock)
}

func generateMetrics(id string, elementId string, aggregate string, displayLabel string) string {
	return fmt.Sprintf(`
        metrics {
            id = "%s"
            element_id = "%s"
            aggregate = "%s"
            display_label = "%s"
        }
        `, id, elementId, aggregate, displayLabel)
}

func generateDisplayAttributes(varType string, groupByTitle string, metricsTitle string, showLegend bool) string {
	return fmt.Sprintf(`
        display_attributes {
            var_type = "%s"
            group_by_title = "%s"
            metrics_title = "%s"
            show_legend = %v
        }
        `, varType, groupByTitle, metricsTitle, showLegend)
}

func generateGroupeByAttributes(elementId string, attribute string) string {
	return fmt.Sprintf(`
        group_by_attributes {
            attribute = "%s"
            element_id = "%s"
        }
        `, attribute, elementId)
}

func testVerifyJourneyViewsDestroyed(state *terraform.State) error {
	journeyViewApi := platformclientv2.NewJourneyApi()
	for _, rs := range state.RootModule().Resources {
		if rs.Type != "genesyscloud_journey_views" {
			continue
		}

		journeyView, resp, err := journeyViewApi.GetJourneyView(rs.Primary.ID)
		if journeyView != nil {
			return fmt.Errorf("journeyView (%s) still exists", rs.Primary.ID)
		} else if util.IsStatus404(resp) {
			// JourneyView not found as expected
			continue
		} else {
			// Unexpected error
			return fmt.Errorf("Unexpected error: %s", err)
		}
	}
	// Success. All journeyView destroyed
	return nil
}
