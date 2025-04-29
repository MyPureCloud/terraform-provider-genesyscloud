package journey_views

import (
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_register"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

const ResourceType = "genesyscloud_journey_views"

var (
	constraintResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"unit": {
				Description:  "The unit for the link's time constraint.Valid values: Seconds, Minutes, Hours, Days, Weeks, Months.",
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"Seconds", "Minutes", "Hours", "Days", "Weeks", "Months"}, false),
			},
			"value": {
				Description: "The value for the link's time constraint.",
				Type:        schema.TypeInt,
				Optional:    true,
			},
		},
	}
	followedByResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"id": {
				Description: "The identifier of the element downstream.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"constraint_within": {
				Description: "A time constraint on this link, which requires a customer to complete the downstream element within this amount of time to be counted.",
				Type:        schema.TypeList,
				Elem:        constraintResource,
				MaxItems:    1,
				Optional:    true,
			},
			"constraint_after": {
				Description: "A time constraint on this link, which requires a customer must complete the downstream element after this amount of time to be counted.",
				Type:        schema.TypeList,
				Elem:        constraintResource,
				MaxItems:    1,
				Optional:    true,
			},
			"event_count_type": {
				Description:  "The type of events that will be counted. Note: Concurrent will override any JourneyViewLinkTimeConstraint. Default is Sequential.Valid values: All, Concurrent, Sequential.",
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"Sequential", "All", "Concurrent"}, false),
			},
			"join_attributes": {
				Description: "Other (secondary) attributes on which this link should join the customers being counted.",
				Type:        schema.TypeList,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Optional:    true,
			},
		},
	}
	predicatesResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"dimension": {
				Description: "The element's attribute being filtered on.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"values": {
				Description: "The identifier for the element based on its type.",
				Type:        schema.TypeList,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Required:    true,
			},
			"operator": {
				Description:  "Optional operator, default is Matches. Valid values: Matches.Valid values: Matches, NotMatches.",
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"Matches", "NotMatches"}, false),
				Default:      "Matches",
			},
			"no_value": {
				Description: "set this to true if no specific value to be considered.",
				Type:        schema.TypeBool,
				Optional:    true,
			},
		},
	}
	rangeElementDataResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"number": {
				Description: "Number value. Only one of number or duration must be specified.",
				Type:        schema.TypeFloat,
				Optional:    true,
			},
			"duration": {
				Description: "An ISO 8601 time duration. Only one of number or duration must be specified",
				Type:        schema.TypeString,
				Optional:    true,
			},
		},
	}

	rangeElementResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"lt": {
				Description: "Comparator: less than",
				Type:        schema.TypeSet,
				MaxItems:    1,
				Elem:        rangeElementDataResource,
				Optional:    true,
			},
			"lte": {
				Description: "Comparator: less than or equal",
				Type:        schema.TypeSet,
				MaxItems:    1,
				Elem:        rangeElementDataResource,
				Optional:    true,
			},
			"gt": {
				Description: "Comparator: greater than",
				Type:        schema.TypeSet,
				MaxItems:    1,
				Elem:        rangeElementDataResource,
				Optional:    true,
			},
			"gte": {
				Description: "Comparator: greater than or equal",
				Type:        schema.TypeSet,
				MaxItems:    1,
				Elem:        rangeElementDataResource,
				Optional:    true,
			},
			"eq": {
				Description: "Comparator: equal",
				Type:        schema.TypeSet,
				MaxItems:    1,
				Elem:        rangeElementDataResource,
				//AtLeastOneOf: []string{"lt", "lte", "gt", "gte", "neq"},
				Optional: true,
			},
			"neq": {
				Description: "Comparator: not equal",
				Type:        schema.TypeSet,
				MaxItems:    1,
				Elem:        rangeElementDataResource,
				Optional:    true,
			},
		},
	}

	numberPredicatesResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"dimension": {
				Description: "The element's attribute being filtered on.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"range": {
				Description: "the range of comparators to filter on.",
				Type:        schema.TypeList,
				Elem:        rangeElementResource,
				MaxItems:    1,
				Required:    true,
			},
			"operator": {
				Description:  "Optional operator, default is Matches. Valid values: Matches.Valid values: Matches, NotMatches.",
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"Matches", "NotMatches"}, false),
				Default:      "Matches",
			},
			"no_value": {
				Description: "set this to true if no specific value to be considered.",
				Type:        schema.TypeBool,
				Optional:    true,
			},
		},
	}
	attributesResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"type": {
				Description:  "The type of the element (e.g. Event).Valid values: Event.",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"Event"}, false),
			},
			"id": {
				Description: "The identifier for the element based on its type.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"source": {
				Description: "The source for the element (e.g. IVR, Voice, Chat). Used for informational purposes only.",
				Type:        schema.TypeString,
				Optional:    true,
			},
		},
	}
	elementDisplayAttributesResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"x": {
				Description: "The horizontal position (x-coordinate) of the element on the journey view canvas.",
				Type:        schema.TypeInt,
				Required:    true,
			},
			"y": {
				Description: "The vertical position (y-coordinate) of the element on the journey view canvas.",
				Type:        schema.TypeInt,
				Required:    true,
			},
			"col": {
				Description: "The column position for the element in the journey view canvas.",
				Type:        schema.TypeInt,
				Required:    true,
			},
		},
	}
	filtersResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"type": {
				Description:  "Boolean operation to apply to the provided predicates, numberPredicates and clauses. Valid values: And.",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"And"}, false),
			},
			"predicates": {
				Description: "A filter on an element within a journey view.",
				Type:        schema.TypeList,
				Elem:        predicatesResource,
				Optional:    true,
			},
			"number_predicates": {
				Description: "A number filter on an element within a journey view.",
				Type:        schema.TypeList,
				Elem:        numberPredicatesResource,
				Optional:    true,
			},
		},
	}
	elementsResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"id": {
				Description: "The unique identifier of the element within the elements list.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"name": {
				Description: "The unique name of the element within the view.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"attributes": {
				Description: "Attributes on an element in a journey view.",
				Type:        schema.TypeList,
				Required:    true,
				Elem:        attributesResource,
				MaxItems:    1,
			},
			"display_attributes": {
				Description: "Display Attributes on an element in a journey view.",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        elementDisplayAttributesResource,
				MaxItems:    1,
			},
			"filter": {
				Description: "A set of filters on an element within a journey view.",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        filtersResource,
				MaxItems:    1,
			},
			"followed_by": {
				Description: "A list of JourneyViewLink objects, listing the elements downstream of this element.",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        followedByResource,
			},
		},
	}
	metricsResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"id": {
				Description: "The unique identifier of the metric within the metrics list.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"element_id": {
				Description: "The reference of element.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"aggregate": {
				Description:  "How to aggregate the given element. Valid values: EventCount, CustomerCount.",
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"EventCount", "CustomerCount"}, false),
			},
			"display_label": {
				Description: "Display label of metric.",
				Type:        schema.TypeString,
				Optional:    true,
			},
		},
	}
	displayAttributesResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"var_type": {
				Description:  "The type of chart to display. Valid values: Bar, Column, Line.",
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"Bar", "Column", "Line"}, false),
			},
			"group_by_title": {
				Description: "A title for the grouped by attributes (aka the x axis).",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"metrics_title": {
				Description: "A title for the metrics (aka the y axis).",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"show_legend": {
				Description: "Whether to show a legend",
				Type:        schema.TypeBool,
				Optional:    true,
			},
		},
	}
	groupByAttributesResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"element_id": {
				Description: "The element in the list of elements which is being grouped by.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"attribute": {
				Description: "The attribute of the element being grouped by.",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
	chartsResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"id": {
				Description: "The unique identifier of the chart within the charts list.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"name": {
				Description: "The unique name of the chart within the view.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"version": {
				Description: "The version of chart",
				Type:        schema.TypeInt,
				Required:    true,
			},
			"metrics": {
				Description: "A set of metrics to be displayed on the chart.",
				Type:        schema.TypeList,
				Elem:        metricsResource,
				Required:    true,
				MinItems:    1,
			},
			"group_by_time": {
				Description:  "A time unit to group the metrics by. Valid values: Day, Week, Month, Year.",
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"Day", "Week", "Month", "Year"}, false),
			},
			"group_by_max": {
				Description: "A maximum on the number of values being grouped by",
				Type:        schema.TypeInt,
				Optional:    true,
			},
			"display_attributes": {
				Description: "Optional display attributes for rendering the chart",
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Elem:        displayAttributesResource,
			},
			"group_by_attributes": {
				Description: "A list of attributes to group the metrics by",
				Type:        schema.TypeList,
				Optional:    true,
				MinItems:    1,
				Elem:        groupByAttributesResource,
			},
		},
	}
)

func SetRegistrar(regInstance registrar.Registrar) {
	regInstance.RegisterResource(ResourceType, ResourceJourneyViews())
	regInstance.RegisterDataSource(ResourceType, DataSourceJourneyView())
	regInstance.RegisterExporter(ResourceType, JourneyViewExporter())
}

func ResourceJourneyViews() *schema.Resource {
	return &schema.Resource{
		Description: "Genesys Cloud Directory JourneyView",

		CreateContext: provider.CreateWithPooledClient(createJourneyView),
		ReadContext:   provider.ReadWithPooledClient(readJourneyView),
		UpdateContext: provider.UpdateWithPooledClient(updateJourneyView),
		DeleteContext: provider.DeleteWithPooledClient(deleteJourneyView),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "JourneyView name.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"description": {
				Description: "A description of the journey view.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"interval": {
				Description: "An absolute timeframe for the journey view, expressed as an ISO 8601 interval. Only one of interval or duration must be specified. Intervals are represented as an ISO-8601 string. For example: YYYY-MM-DDThh:mm:ss/YYYY-MM-DDThh:mm:ss.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"version": {
				Description: "Version of JourneyView.",
				Type:        schema.TypeInt,
				Computed:    true,
			},
			"duration": {
				Description: "A relative timeframe for the journey view, expressed as an ISO 8601 duration. Only one of interval or duration must be specified. Periods are represented as an ISO-8601 string. For example: P1D or P1DT12H.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"elements": {
				Description: "The elements within the journey view.",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        elementsResource,
			},
			"charts": {
				Description: "The charts within the journey view.",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        chartsResource,
			},
		},
	}
}

func JourneyViewExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllJourneyViews),
		AllowZeroValues:  []string{"bullseye_rings.expansion_timeout_seconds"},
	}
}

func DataSourceJourneyView() *schema.Resource {
	return &schema.Resource{
		Description:        "Data source for Genesys Cloud Journey Views. Select a Journey View by name.",
		ReadWithoutTimeout: provider.ReadWithPooledClient(dataSourceJourneyViewRead),
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "JourneyView name.",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}
