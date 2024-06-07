package journey_views

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	registrar "terraform-provider-genesyscloud/genesyscloud/resource_register"
)

const resourceName = "genesyscloud_journey_views"

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
				Description: "A time constraint on this link, which requires a customer must complete the downstream element after this amount of time to be counted..",
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
				Description: "The element's attribute being filtered on",
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
	filtersResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"type": {
				Description:  "Boolean operation to apply to the provided predicates and clauses. Valid values: And.",
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
)

func SetRegistrar(regInstance registrar.Registrar) {
	regInstance.RegisterResource(resourceName, ResourceJourneyViews())
	/***
	TODO: Add DataSource and Exporter once we are done with https://inindca.atlassian.net/browse/JM-109
	regInstance.RegisterDataSource(resourceName, DataSourceGroup())
	regInstance.RegisterExporter(resourceName, JourneyViewExporter())
	***/
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
		},
	}
}
