package genesyscloud

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/mypurecloud/platform-client-sdk-go/v74/platformclientv2"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/consistency_checker"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resourcedata"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/stringmap"
)

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
			Required:    true,
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
		},
		// TODO
		//"page_url_conditions": {
		//	Description: "URL conditions that a page must match for web actions to be displayable.",
		//	Type: schema.TypeSet,
		//	Required: true,
		//	Elem: journeyactionmapurlconditionResource,
		//},
		//"activation": {
		//	Description: "Type of activation.",
		//	Type: schema.TypeSet,
		//	Optional: true,
		//	MaxItems: 1,
		//	Elem: journeyactionmapactivationResource,
		//},
		"weight": {
			Description: "Weight of the action map with higher number denoting higher weight.",
			Type:        schema.TypeInt,
			Optional:    true,
		},
		// TODO
		//"action": {
		//	Description: "The action that will be executed if this action map is triggered.",
		//	Type: schema.TypeSet,
		//	Optional: true,
		//	MaxItems: 1,
		//	Elem: journeyactionmapactionmapactionResource,
		//},
		//"action_map_schedule_groups": {
		//	Description: "The action map's associated schedule groups.",
		//	Type: schema.TypeSet,
		//	Optional: true,
		//	MaxItems: 1,
		//	Elem: journeyactionmapactionmapschedulegroupsResource,
		//},
		"ignore_frequency_cap": {
			Description: "Override organization-level frequency cap and always offer web engagements from this action map.",
			Type:        schema.TypeBool,
			Optional:    true,
		},
		"start_date": {
			Description: "Timestamp at which the action map is scheduled to start firing. Date time is represented as an ISO-8601 string. For example: yyyy-MM-ddTHH:mm:ss[.mmm]Z",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"end_date": {
			Description: "Timestamp at which the action map is scheduled to stop firing. Date time is represented as an ISO-8601 string. For example: yyyy-MM-ddTHH:mm:ss[.mmm]Z",
			Type:        schema.TypeString,
			Optional:    true,
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
				Description:  "The stream type for which this condition can be satisfied. Valid values: Web, Custom, Conversation.",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"Web", "Custom", "Conversation"}, false),
			},
			"session_type": {
				Description: "The session type for which this condition can be satisfied.",
				Type:        schema.TypeString,
				Required:    true,
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
)

func getAllJourneyActionMaps(_ context.Context, clientConfig *platformclientv2.Configuration) (ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(ResourceIDMetaMap)
	journeyApi := platformclientv2.NewJourneyApiWithConfig(clientConfig)

	pageCount := 1 // Needed because of broken journey common paging
	for pageNum := 1; pageNum <= pageCount; pageNum++ {
		const pageSize = 100
		actionMaps, _, getErr := journeyApi.GetJourneyActionmaps(pageNum, pageSize, "", "", "", nil, nil, "")
		if getErr != nil {
			return nil, diag.Errorf("Failed to get page of journey action maps: %v", getErr)
		}

		if actionMaps.Entities == nil || len(*actionMaps.Entities) == 0 {
			break
		}

		for _, actionMap := range *actionMaps.Entities {
			resources[*actionMap.Id] = &ResourceMeta{Name: *actionMap.DisplayName}
		}

		pageCount = *actionMaps.PageCount
	}

	return resources, nil
}

func journeyActionMapExporter() *ResourceExporter {
	return &ResourceExporter{
		GetResourcesFunc: getAllWithPooledClient(getAllJourneyActionMaps),
		RefAttrs:         map[string]*RefAttrSettings{}, // No references
	}
}

func resourceJourneyActionMap() *schema.Resource {
	return &schema.Resource{
		Description:   "Genesys Cloud Journey Action Map",
		CreateContext: createWithPooledClient(createJourneyActionMap),
		ReadContext:   readWithPooledClient(readJourneyActionMap),
		UpdateContext: updateWithPooledClient(updateJourneyActionMap),
		DeleteContext: deleteWithPooledClient(deleteJourneyActionMap),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema:        journeyActionMapSchema,
	}
}

func createJourneyActionMap(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*providerMeta).ClientConfig
	journeyApi := platformclientv2.NewJourneyApiWithConfig(sdkConfig)
	actionMap := buildSdkActionMap(d)

	log.Printf("Creating journey action map %s", *actionMap.DisplayName)
	result, resp, err := journeyApi.PostJourneyActionmaps(*actionMap)
	if err != nil {
		return diag.Errorf("failed to create journey action map %s: %s\n(input: %+v)\n(resp: %s)", *actionMap.DisplayName, err, *actionMap, getBody(resp))
	}

	d.SetId(*result.Id)

	log.Printf("Created journey action map %s %s", *result.DisplayName, *result.Id)
	return readJourneyActionMap(ctx, d, meta)
}

func readJourneyActionMap(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*providerMeta).ClientConfig
	journeyApi := platformclientv2.NewJourneyApiWithConfig(sdkConfig)

	log.Printf("Reading journey action map %s", d.Id())
	return withRetriesForRead(ctx, d, func() *resource.RetryError {
		actionMap, resp, getErr := journeyApi.GetJourneyActionmap(d.Id())
		if getErr != nil {
			if isStatus404(resp) {
				return resource.RetryableError(fmt.Errorf("failed to read journey action map %s: %s", d.Id(), getErr))
			}
			return resource.NonRetryableError(fmt.Errorf("failed to read journey action map %s: %s", d.Id(), getErr))
		}

		cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, resourceJourneyActionMap())
		flattenActionMap(d, actionMap)

		log.Printf("Read journey action map  %s %s", d.Id(), *actionMap.DisplayName)
		return cc.CheckState()
	})
}

func updateJourneyActionMap(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*providerMeta).ClientConfig
	journeyApi := platformclientv2.NewJourneyApiWithConfig(sdkConfig)
	patchActionMap := buildSdkPatchActionMap(d)

	log.Printf("Updating journey action map %s", d.Id())
	diagErr := retryWhen(isVersionMismatch, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		// Get current journey action map version
		actionMap, resp, getErr := journeyApi.GetJourneyActionmap(d.Id())
		if getErr != nil {
			return resp, diag.Errorf("Failed to read current journey action map %s: %s", d.Id(), getErr)
		}

		patchActionMap.Version = actionMap.Version
		_, resp, patchErr := journeyApi.PatchJourneyActionmap(d.Id(), *patchActionMap)
		if patchErr != nil {
			return resp, diag.Errorf("Error updating journey action map %s: %s\n(input: %+v)\n(resp: %s)", *patchActionMap.DisplayName, patchErr, *patchActionMap, getBody(resp))
		}
		return resp, nil
	})
	if diagErr != nil {
		return diagErr
	}

	log.Printf("Updated journey action map %s", d.Id())
	return readJourneyActionMap(ctx, d, meta)
}

func deleteJourneyActionMap(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	displayName := d.Get("display_name").(string)

	sdkConfig := meta.(*providerMeta).ClientConfig
	journeyApi := platformclientv2.NewJourneyApiWithConfig(sdkConfig)

	log.Printf("Deleting journey action map with display name %s", displayName)
	if _, err := journeyApi.DeleteJourneyActionmap(d.Id()); err != nil {
		return diag.Errorf("Failed to delete journey action map with display name %s: %s", displayName, err)
	}

	return withRetries(ctx, 30*time.Second, func() *resource.RetryError {
		_, resp, err := journeyApi.GetJourneyActionmap(d.Id())
		if err != nil {
			if isStatus404(resp) {
				// journey action map deleted
				log.Printf("Deleted journey action map %s", d.Id())
				return nil
			}
			return resource.NonRetryableError(fmt.Errorf("error deleting journey action map %s: %s", d.Id(), err))
		}

		return resource.RetryableError(fmt.Errorf("journey action map %s still exists", d.Id()))
	})
}

func flattenActionMap(d *schema.ResourceData, actionMap *platformclientv2.Actionmap) {
	d.Set("is_active", *actionMap.IsActive)
	d.Set("display_name", *actionMap.DisplayName)
	d.Set("trigger_with_segments", stringListToSet(*actionMap.TriggerWithSegments))
	resourcedata.SetNillableValue(d, "trigger_with_event_conditions", flattenList(actionMap.TriggerWithEventConditions, flattenEventCondition))
	resourcedata.SetNillableValue(d, "trigger_with_outcome_probability_conditions", flattenList(actionMap.TriggerWithOutcomeProbabilityConditions, flattenOutcomeProbabilityCondition))
	// TODO
	resourcedata.SetNillableValue[int](d, "weight", actionMap.Weight)
	// TODO
	resourcedata.SetNillableValue[bool](d, "ignore_frequency_cap", actionMap.IgnoreFrequencyCap)
	resourcedata.SetNillableTime(d, "start_date", actionMap.StartDate)
	resourcedata.SetNillableTime(d, "end_date", actionMap.EndDate)
}

func buildSdkActionMap(actionMap *schema.ResourceData) *platformclientv2.Actionmap {
	isActive := actionMap.Get("is_active").(bool)
	displayName := actionMap.Get("display_name").(string)
	triggerWithSegments := buildSdkStringList(actionMap, "trigger_with_segments")
	triggerWithEventConditions := resourcedata.BuildSdkList(actionMap, "trigger_with_event_conditions", buildSdkEventCondition)
	triggerWithOutcomeProbabilityConditions := nilToEmptyList(resourcedata.BuildSdkList(actionMap, "trigger_with_outcome_probability_conditions", buildSdkOutcomeProbabilityCondition))
	// TODO
	weight := resourcedata.GetNillableValue[int](actionMap, "weight")
	// TODO
	ignoreFrequencyCap := resourcedata.GetNillableBool(actionMap, "ignore_frequency_cap")
	startDate := resourcedata.GetNillableTime(actionMap, "start_date")
	endDate := resourcedata.GetNillableTime(actionMap, "end_date")

	return &platformclientv2.Actionmap{
		IsActive:                                &isActive,
		DisplayName:                             &displayName,
		TriggerWithSegments:                     triggerWithSegments,
		TriggerWithEventConditions:              triggerWithEventConditions,
		TriggerWithOutcomeProbabilityConditions: triggerWithOutcomeProbabilityConditions,
		// TODO
		Weight: weight,
		// TODO
		IgnoreFrequencyCap: ignoreFrequencyCap,
		StartDate:          startDate,
		EndDate:            endDate,
	}
}

func buildSdkPatchActionMap(actionMap *schema.ResourceData) *platformclientv2.Patchactionmap {
	isActive := actionMap.Get("is_active").(bool)
	displayName := actionMap.Get("display_name").(string)
	triggerWithSegments := buildSdkStringList(actionMap, "trigger_with_segments")
	triggerWithEventConditions := resourcedata.BuildSdkList(actionMap, "trigger_with_event_conditions", buildSdkEventCondition)
	triggerWithOutcomeProbabilityConditions := resourcedata.BuildSdkList(actionMap, "trigger_with_outcome_probability_conditions", buildSdkOutcomeProbabilityCondition)
	// TODO
	weight := resourcedata.GetNillableValue[int](actionMap, "weight")
	// TODO
	ignoreFrequencyCap := resourcedata.GetNillableBool(actionMap, "ignore_frequency_cap")
	startDate := resourcedata.GetNillableTime(actionMap, "start_date")
	endDate := resourcedata.GetNillableTime(actionMap, "end_date")

	return &platformclientv2.Patchactionmap{
		IsActive:                                &isActive,
		DisplayName:                             &displayName,
		TriggerWithSegments:                     triggerWithSegments,
		TriggerWithEventConditions:              triggerWithEventConditions,
		TriggerWithOutcomeProbabilityConditions: triggerWithOutcomeProbabilityConditions,
		// TODO
		Weight: weight,
		// TODO
		IgnoreFrequencyCap: ignoreFrequencyCap,
		StartDate:          startDate,
		EndDate:            endDate,
	}
}

func flattenEventCondition(eventCondition *platformclientv2.Eventcondition) map[string]interface{} {
	eventConditionMap := make(map[string]interface{})
	eventConditionMap["key"] = *eventCondition.Key
	eventConditionMap["values"] = stringListToSet(*eventCondition.Values)
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
	outcomeProbabilityConditionMap["maximum_probability"] = *outcomeProbabilityCondition.MaximumProbability
	stringmap.SetValueIfNotNil(outcomeProbabilityConditionMap, "probability", outcomeProbabilityCondition.Probability)
	return outcomeProbabilityConditionMap
}

func buildSdkOutcomeProbabilityCondition(outcomeProbabilityCondition map[string]interface{}) *platformclientv2.Outcomeprobabilitycondition {
	outcomeId := outcomeProbabilityCondition["outcome_id"].(string)
	maximumProbability := float32(outcomeProbabilityCondition["maximum_probability"].(float64))
	var probability *float32 = nil
	probabilityFloat64 := stringmap.GetNillableValue[float64](outcomeProbabilityCondition, "probability")
	if probabilityFloat64 != nil {
		probabilityFloat32 := float32(*probabilityFloat64)
		probability = &probabilityFloat32
	}

	return &platformclientv2.Outcomeprobabilitycondition{
		OutcomeId:          &outcomeId,
		MaximumProbability: &maximumProbability,
		Probability:        probability,
	}
}
