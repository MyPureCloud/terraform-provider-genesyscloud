package journey_action_map

import (
	"terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "terraform-provider-genesyscloud/genesyscloud/resource_register"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"terraform-provider-genesyscloud/genesyscloud/validators"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

const ResourceType = "genesyscloud_journey_action_map"

var (
	journeyActionMapSchema = map[string]*schema.Schema{
		"is_active": {
			Description: "Whether the action map is active.",
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     true,
		},
		"display_name": {
			Description: "Display name of the action map.",
			Type:        schema.TypeString,
			Required:    true,
		},
		"trigger_with_segments": {
			Description: "Trigger action map if any segment in the list is assigned to a given customer.",
			Type:        schema.TypeSet,
			Optional:    true,
			Elem:        &schema.Schema{Type: schema.TypeString},
		},
		"trigger_with_event_conditions": {
			Description: "List of event conditions that must be satisfied to trigger the action map.",
			Type:        schema.TypeSet,
			Optional:    true,
			Elem:        eventConditionResource,
		},
		"trigger_with_outcome_probability_conditions": {
			Description: "Probability conditions for outcomes that must be satisfied to trigger the action map.",
			Type:        schema.TypeSet,
			Optional:    true,
			Elem:        outcomeProbabilityConditionResource,
			Deprecated:  "Use trigger_with_outcome_quantile_conditions attribute instead.",
		},
		"trigger_with_outcome_quantile_conditions": {
			Description: "Quantile conditions for outcomes that must be satisfied to trigger the action map.",
			Type:        schema.TypeSet,
			Optional:    true,
			Elem:        outcomeQuantileConditionResource,
		},
		"page_url_conditions": {
			Description: "URL conditions that a page must match for web actions to be displayable.",
			Type:        schema.TypeSet,
			Optional:    true,
			Elem:        urlConditionResource,
		},
		"activation": {
			Description: "Type of activation.",
			Type:        schema.TypeSet,
			Required:    true,
			MaxItems:    1,
			Elem:        activationResource,
		},
		"weight": {
			Description:  "Weight of the action map with higher number denoting higher weight. Low=1, Medium=2, High=3.",
			Type:         schema.TypeInt,
			Optional:     true,
			Default:      2,
			ValidateFunc: validation.IntBetween(1, 3),
		},
		"action": {
			Description: "The action that will be executed if this action map is triggered.",
			Type:        schema.TypeSet,
			Required:    true,
			MaxItems:    1,
			Elem:        actionMapActionResource,
		},
		"action_map_schedule_groups": {
			Description: "The action map's associated schedule groups.",
			Type:        schema.TypeSet,
			Optional:    true,
			MaxItems:    1,
			Elem:        actionMapScheduleGroupsResource,
		},
		"ignore_frequency_cap": {
			Description: "Override organization-level frequency cap and always offer web engagements from this action map.",
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
		},
		"start_date": {
			Description:      "Timestamp at which the action map is scheduled to start firing. Date time is represented as an ISO-8601 string without a timezone. For example: 2006-01-02T15:04:05.000000.",
			Type:             schema.TypeString,
			Required:         true, // Now is the default value for this field. Better to make it required.
			ValidateDiagFunc: validators.ValidateLocalDateTimes,
		},
		"end_date": {
			Description:      "Timestamp at which the action map is scheduled to stop firing. Date time is represented as an ISO-8601 string without a timezone. For example: 2006-01-02T15:04:05.000000.",
			Type:             schema.TypeString,
			Optional:         true,
			ValidateDiagFunc: validators.ValidateLocalDateTimes,
		},
	}

	eventConditionResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"key": {
				Description: "The event key.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"values": {
				Description: "The event values.",
				Type:        schema.TypeSet,
				Required:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"operator": {
				Description:  "The comparison operator. Valid values: containsAll, containsAny, notContainsAll, notContainsAny, equal, notEqual, greaterThan, greaterThanOrEqual, lessThan, lessThanOrEqual, startsWith, endsWith.",
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "equal",
				ValidateFunc: validation.StringInSlice([]string{"containsAll", "containsAny", "notContainsAll", "notContainsAny", "equal", "notEqual", "greaterThan", "greaterThanOrEqual", "lessThan", "lessThanOrEqual", "startsWith", "endsWith"}, false),
			},
			"stream_type": {
				Description:  "The stream type for which this condition can be satisfied. Valid values: Web, App.",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"Web", "App" /*,"Custom", "Conversation" */}, false), // Custom and Conversation seem not to be supported by the API despite the documentation (DEVENGSD-607)
			},
			"session_type": {
				Description:  "The session type for which this condition can be satisfied. Valid values: web, app.",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"web", "app"}, false), // custom value seems not to be supported by the API despite the documentation
			},
			"event_name": {
				Description: "The name of the event for which this condition can be satisfied.",
				Type:        schema.TypeString,
				Optional:    true,
			},
		},
	}

	outcomeProbabilityConditionResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"outcome_id": {
				Description: "The outcome ID.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"maximum_probability": {
				Description: "Probability value for the selected outcome at or above which the action map will trigger.",
				Type:        schema.TypeFloat,
				Required:    true,
			},
			"probability": {
				Description: "Additional probability condition, where if set, the action map will trigger if the current outcome probability is lower or equal to the value.",
				Type:        schema.TypeFloat,
				Optional:    true,
			},
		},
	}

	outcomeQuantileConditionResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"outcome_id": {
				Description: "The outcome ID.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"max_quantile_threshold": {
				Description: "This Outcome Quantile Condition is met when sessionMaxQuantile of the OutcomeScore is above this value, (unless fallbackQuantile is set). Range 0.00-1.00",
				Type:        schema.TypeFloat,
				Required:    true,
			},
			"fallback_quantile_threshold": {
				Description: "If set, this Condition is met when max_quantile_threshold is met, AND the current quantile of the OutcomeScore is below this fallback_quantile_threshold. Range 0.00-1.00",
				Type:        schema.TypeFloat,
				Optional:    true,
			},
		},
	}

	urlConditionResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"values": {
				Description: "The URL condition value.",
				Type:        schema.TypeSet,
				Required:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"operator": {
				Description:  "The comparison operator. Valid values: containsAll, containsAny, notContainsAll, notContainsAny, equal, notEqual, greaterThan, greaterThanOrEqual, lessThan, lessThanOrEqual, startsWith, endsWith.",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"containsAll", "containsAny", "notContainsAll", "notContainsAny", "equal", "notEqual", "greaterThan", "greaterThanOrEqual", "lessThan", "lessThanOrEqual", "startsWith", "endsWith"}, false),
			},
		},
	}

	activationResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"type": {
				Description:  "Type of activation. Valid values: immediate, on-next-visit, on-next-session, delay.",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"immediate", "on-next-visit", "on-next-session", "delay"}, false),
			},
			"delay_in_seconds": {
				Description: "Activation delay time amount.",
				Type:        schema.TypeInt,
				Optional:    true,
			},
		},
	}

	actionMapActionResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"action_template_id": {
				Description: "Action template associated with the action map. For media type contentOffer.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"media_type": {
				Description:  "Media type of action. Valid values: webchat, webMessagingOffer, contentOffer, architectFlow, openAction.",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"webchat", "webMessagingOffer", "contentOffer", "architectFlow", "openAction"}, false),
			},
			"architect_flow_fields": {
				Description: "Architect Flow Id and input contract. For media type architectFlow.",
				Type:        schema.TypeSet,
				Optional:    true,
				MaxItems:    1,
				Elem:        architectFlowFieldsResource,
			},
			"web_messaging_offer_fields": {
				Description: "Admin-configurable fields of a web messaging offer action. For media type webMessagingOffer.",
				Type:        schema.TypeSet,
				Optional:    true,
				MaxItems:    1,
				Elem:        webMessagingOfferFieldsResource,
			},
			"open_action_fields": {
				Description: "Admin-configurable fields of an open action. For media type openAction.",
				Type:        schema.TypeSet,
				Optional:    true,
				MaxItems:    1,
				Elem:        openActionFieldsResource,
			},
			"is_pacing_enabled": {
				Description: "Whether this action should be throttled.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
			},
		},
	}

	architectFlowFieldsResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"architect_flow_id": {
				Description: "The architect flow.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"flow_request_mappings": {
				Description: "Collection of Architect Flow Request Mappings to use.",
				Type:        schema.TypeSet,
				Optional:    true,
				Elem:        requestMappingResource,
			},
		},
	}

	requestMappingResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Name of the Integration Action Attribute to supply the value for",
				Type:        schema.TypeString,
				Required:    true,
			},
			"attribute_type": {
				Description:  "Type of the value supplied. Valid values: String, Number, Integer, Boolean.",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"String", "Number", "Integer", "Boolean"}, false),
			},
			"mapping_type": {
				Description:  "Method of finding value to use with Attribute. Valid values: Lookup, HardCoded.",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"Lookup", "HardCoded"}, false),
			},
			"value": {
				Description: "Value to supply for the specified Attribute",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}

	webMessagingOfferFieldsResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"offer_text": {
				Description: "Text value to be used when inviting a visitor to engage with a web messaging offer.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"architect_flow_id": {
				Description: "Flow to be invoked, overrides default flow when specified.",
				Type:        schema.TypeString,
				Optional:    true,
			},
		},
	}

	openActionFieldsResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"open_action": {
				Description: "The specific type of the open action.",
				Type:        schema.TypeSet,
				Required:    true,
				MaxItems:    1,
				Elem:        domainEntityRefResource,
			},
			"configuration_fields": {
				Description:      "Custom fields defined in the schema referenced by the open action type selected.",
				Type:             schema.TypeString,
				Optional:         true,
				DiffSuppressFunc: util.SuppressEquivalentJsonDiffs,
			},
		},
	}

	domainEntityRefResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"id": {
				Description: "Id.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"name": {
				Description: "Name.",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}

	actionMapScheduleGroupsResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"action_map_schedule_group_id": {
				Description: "The actions map's associated schedule group.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"emergency_action_map_schedule_group_id": {
				Description: "The action map's associated emergency schedule group.",
				Type:        schema.TypeString,
				Optional:    true,
			},
		},
	}
)

// SetRegistrar registers all of the resources, datasources and exporters in the package
func SetRegistrar(regInstance registrar.Registrar) {
	regInstance.RegisterResource(ResourceType, ResourceJourneyActionMap())
	regInstance.RegisterDataSource(ResourceType, DataSourceJourneyActionMap())
	regInstance.RegisterExporter(ResourceType, JourneyActionMapExporter())
}

func JourneyActionMapExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllJourneyActionMaps),
		RefAttrs: map[string]*resourceExporter.RefAttrSettings{
			"trigger_with_segments":                                             {RefType: "genesyscloud_journey_segment"},
			"trigger_with_outcome_probability_conditions.outcome_id":            {RefType: "genesyscloud_journey_outcome"},
			"trigger_with_outcome_quantile_conditions.outcome_id":               {RefType: "genesyscloud_journey_outcome"},
			"action.architect_flow_fields.architect_flow_id":                    {RefType: "genesyscloud_flow"},
			"action_map_schedule_groups.action_map_schedule_group_id":           {RefType: "genesyscloud_architect_schedulegroups"},
			"action_map_schedule_groups.emergency_action_map_schedule_group_id": {RefType: "genesyscloud_architect_schedulegroups"},
			"action.action_template_id":                                         {RefType: "genesyscloud_journey_action_template"},
		},
	}
}

func ResourceJourneyActionMap() *schema.Resource {
	return &schema.Resource{
		Description:   "Genesys Cloud Journey Action Map",
		CreateContext: provider.CreateWithPooledClient(createJourneyActionMap),
		ReadContext:   provider.ReadWithPooledClient(readJourneyActionMap),
		UpdateContext: provider.UpdateWithPooledClient(updateJourneyActionMap),
		DeleteContext: provider.DeleteWithPooledClient(deleteJourneyActionMap),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema:        journeyActionMapSchema,
	}
}

func DataSourceJourneyActionMap() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for Genesys Cloud Action Map. Select a journey action map by name",
		ReadContext: provider.ReadWithPooledClient(dataSourceJourneyActionMapRead),
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Journey Action Map name.",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}
