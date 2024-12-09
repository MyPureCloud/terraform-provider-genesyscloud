package genesyscloud

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"terraform-provider-genesyscloud/genesyscloud/util/constants"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"
	"terraform-provider-genesyscloud/genesyscloud/util/resourcedata"
	"terraform-provider-genesyscloud/genesyscloud/util/stringmap"

	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	lists "terraform-provider-genesyscloud/genesyscloud/util/lists"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/mypurecloud/platform-client-sdk-go/v146/platformclientv2"
)

var (
	journeySegmentSchema = map[string]*schema.Schema{
		"is_active": {
			Description: "Whether or not the segment is active.",
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     true,
		},
		"display_name": {
			Description: "The display name of the segment.",
			Type:        schema.TypeString,
			Required:    true,
		},
		"description": {
			Description: "A description of the segment.",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"color": {
			Description: "The hexadecimal color value of the segment.",
			Type:        schema.TypeString,
			Required:    true,
			ValidateFunc: validation.StringMatch(func() *regexp.Regexp {
				r, _ := regexp.Compile("^#[a-fA-F\\d]{6}$")
				return r
			}(), ""),
		},
		"scope": {
			Description:  "The target entity that a segment applies to. Valid values: Session",
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true, // scope can be only set during creation
			ValidateFunc: validation.StringInSlice([]string{"Session"}, false),
		},
		"should_display_to_agent": {
			Description: "Whether or not the segment should be displayed to agent/supervisor users.",
			Type:        schema.TypeBool,
			Optional:    true,
		},
		"context": {
			Description: "The context of the segment.",
			Type:        schema.TypeSet,
			Optional:    true,
			MaxItems:    1,
			Elem:        contextResource,
		},
		"journey": {
			Description: "The pattern of rules defining the segment.",
			Type:        schema.TypeSet,
			Optional:    true,
			MaxItems:    1,
			Elem:        journeyResource,
		},
	}

	contextResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"patterns": {
				Description: "A list of one or more patterns to match.",
				Type:        schema.TypeSet,
				Required:    true,
				Elem:        contextPatternResource,
			},
		},
	}

	contextPatternResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"criteria": {
				Description: "A list of one or more criteria to satisfy.",
				Type:        schema.TypeSet,
				Required:    true,
				Elem:        entityTypeCriteriaResource,
			},
		},
	}

	entityTypeCriteriaResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"key": {
				Description:  "The criteria key.",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"device.category", "device.type", "device.osFamily", "browser.family", "browser.lang", "browser.version", "mktCampaign.source", "mktCampaign.medium", "mktCampaign.name", "mktCampaign.term", "mktCampaign.content", "mktCampaign.clickId", "mktCampaign.network", "geolocation.countryName", "geolocation.locality", "geolocation.region", "geolocation.postalCode", "geolocation.country", "ipOrganization", "referrer.url", "referrer.medium", "referrer.hostname", "authenticated"}, false),
			},
			"values": {
				Description: "The criteria values.",
				Type:        schema.TypeSet,
				Required:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"should_ignore_case": {
				Description: "Should criteria be case insensitive.",
				Type:        schema.TypeBool,
				Required:    true,
			},
			"operator": {
				Description:  "The comparison operator. Valid values: containsAll, containsAny, notContainsAll, notContainsAny, equal, notEqual, greaterThan, greaterThanOrEqual, lessThan, lessThanOrEqual, startsWith, endsWith.",
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "equal",
				ValidateFunc: validation.StringInSlice([]string{"containsAll", "containsAny", "notContainsAll", "notContainsAny", "equal", "notEqual", "greaterThan", "greaterThanOrEqual", "lessThan", "lessThanOrEqual", "startsWith", "endsWith"}, false),
			},
			"entity_type": {
				Description:  "The entity to match the pattern against.Valid values: visit.",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"visit"}, false),
			},
		},
	}

	journeyResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"patterns": {
				Description: "A list of zero or more patterns to match.",
				Type:        schema.TypeSet,
				Required:    true,
				Elem:        journeyPatternResource,
			},
		},
	}

	journeyPatternResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"criteria": {
				Description: "A list of one or more criteria to satisfy.",
				Type:        schema.TypeSet,
				Required:    true,
				Elem:        criteriaResource,
			},
			"count": {
				Description: "The number of times the pattern must match.",
				Type:        schema.TypeInt,
				Required:    true,
			},
			"stream_type": {
				Description:  "The stream type for which this pattern can be matched on. Valid values: Web, App.",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"Web", "App" /*, "Custom", "Conversation"*/}, false), // Custom and Conversation seem not to be supported by the API despite the documentation (DEVENGSD-607)
			},
			"session_type": {
				Description:  "The session type for which this pattern can be matched on. Valid values: web, app.",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"web", "app"}, false), // custom value seems not to be supported by the API despite the documentation
			},
			"event_name": {
				Description: "The name of the event for which this pattern can be matched on.",
				Type:        schema.TypeString,
				Optional:    true,
				Default:     nil,
			},
		},
	}

	criteriaResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"key": {
				Description: "The criteria key.",
				Type:        schema.TypeString,
				Required:    true,
				ValidateFunc: validation.Any(
					validation.StringInSlice([]string{"eventName", "page.url", "page.title", "page.hostname", "page.domain", "page.fragment", "page.keywords", "page.pathname", "searchQuery", "page.queryString"}, false),
					validation.StringMatch(func() *regexp.Regexp {
						r, _ := regexp.Compile("attributes\\..*\\.value")
						return r
					}(), ""),
				),
			},
			"values": {
				Description: "The criteria values.",
				Type:        schema.TypeSet,
				Required:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"should_ignore_case": {
				Description: "Should criteria be case insensitive.",
				Type:        schema.TypeBool,
				Required:    true,
			},
			"operator": {
				Description:  "The comparison operator.Valid values: containsAll, containsAny, notContainsAll, notContainsAny, equal, notEqual, greaterThan, greaterThanOrEqual, lessThan, lessThanOrEqual, startsWith, endsWith.",
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "equal",
				ValidateFunc: validation.StringInSlice([]string{"containsAll", "containsAny", "notContainsAll", "notContainsAny", "equal", "notEqual", "greaterThan", "greaterThanOrEqual", "lessThan", "lessThanOrEqual", "startsWith", "endsWith"}, false),
			},
		},
	}
)

func getAllJourneySegments(_ context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(resourceExporter.ResourceIDMetaMap)
	journeyApi := platformclientv2.NewJourneyApiWithConfig(clientConfig)

	pageCount := 1 // Needed because of broken journey common paging
	for pageNum := 1; pageNum <= pageCount; pageNum++ {
		const pageSize = 100
		journeySegments, resp, getErr := journeyApi.GetJourneySegments("", pageSize, pageNum, true, nil, nil, "")
		if getErr != nil {
			return nil, util.BuildAPIDiagnosticError("genesyscloud_journey_segment", fmt.Sprintf("Failed to get page of journey segments error: %s", getErr), resp)
		}

		if journeySegments.Entities == nil || len(*journeySegments.Entities) == 0 {
			break
		}

		for _, journeySegment := range *journeySegments.Entities {
			resources[*journeySegment.Id] = &resourceExporter.ResourceMeta{BlockLabel: *journeySegment.DisplayName}
		}

		pageCount = *journeySegments.PageCount
	}

	return resources, nil
}

func JourneySegmentExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllJourneySegments),
		RefAttrs:         map[string]*resourceExporter.RefAttrSettings{}, // No references
	}
}

func ResourceJourneySegment() *schema.Resource {
	return &schema.Resource{
		Description:   "Genesys Cloud Journey Segment",
		CreateContext: provider.CreateWithPooledClient(createJourneySegment),
		ReadContext:   provider.ReadWithPooledClient(readJourneySegment),
		UpdateContext: provider.UpdateWithPooledClient(updateJourneySegment),
		DeleteContext: provider.DeleteWithPooledClient(deleteJourneySegment),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema:        journeySegmentSchema,
	}
}

func createJourneySegment(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	journeyApi := platformclientv2.NewJourneyApiWithConfig(sdkConfig)
	journeySegment := buildSdkJourneySegment(d)

	log.Printf("Creating journey segment %s", *journeySegment.DisplayName)
	result, resp, err := journeyApi.PostJourneySegments(*journeySegment)
	if err != nil {
		input, _ := util.InterfaceToJson(*journeySegment)
		return util.BuildAPIDiagnosticError("genesyscloud_journey_segment", fmt.Sprintf("failed to create journey segment %s: %s\n(input: %+v)", *journeySegment.DisplayName, err, input), resp)
	}

	d.SetId(*result.Id)

	log.Printf("Created journey segment %s %s", *result.DisplayName, *result.Id)
	return readJourneySegment(ctx, d, meta)
}

func readJourneySegment(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	journeyApi := platformclientv2.NewJourneyApiWithConfig(sdkConfig)
	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceJourneySegment(), constants.ConsistencyChecks(), "genesyscloud_journey_segment")

	log.Printf("Reading journey segment %s", d.Id())
	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		journeySegment, resp, getErr := journeyApi.GetJourneySegment(d.Id())
		if getErr != nil {
			if util.IsStatus404(resp) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError("genesyscloud_journey_segment", fmt.Sprintf("failed to read journey segment %s | error: %s", d.Id(), getErr), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError("genesyscloud_journey_segment", fmt.Sprintf("failed to read journey segment %s | error: %s", d.Id(), getErr), resp))
		}

		flattenJourneySegment(d, journeySegment)

		log.Printf("Read journey segment %s %s", d.Id(), *journeySegment.DisplayName)
		return cc.CheckState(d)
	})
}

func updateJourneySegment(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	journeyApi := platformclientv2.NewJourneyApiWithConfig(sdkConfig)
	patchSegment := buildSdkPatchSegment(d)

	log.Printf("Updating journey segment %s", d.Id())
	diagErr := util.RetryWhen(util.IsVersionMismatch, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		// Get current journey segment version
		journeySegment, resp, getErr := journeyApi.GetJourneySegment(d.Id())
		if getErr != nil {
			return resp, util.BuildAPIDiagnosticError("genesyscloud_journey_segment", fmt.Sprintf("Failed to read current journey segment %s error: %s", d.Id(), getErr), resp)
		}

		patchSegment.Version = journeySegment.Version
		_, resp, patchErr := journeyApi.PatchJourneySegment(d.Id(), *patchSegment)
		if patchErr != nil {
			input, _ := util.InterfaceToJson(*patchSegment)
			return resp, util.BuildAPIDiagnosticError("genesyscloud_journey_segment", fmt.Sprintf("Failed to update journey segment %s (input: %+v) error: %s", *patchSegment.DisplayName, input, patchErr), resp)
		}
		return resp, nil
	})
	if diagErr != nil {
		return diagErr
	}

	log.Printf("Updated journey segment %s", d.Id())
	return readJourneySegment(ctx, d, meta)
}

func deleteJourneySegment(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	displayName := d.Get("display_name").(string)

	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	journeyApi := platformclientv2.NewJourneyApiWithConfig(sdkConfig)

	log.Printf("Deleting journey segment with display name %s", displayName)
	if resp, err := journeyApi.DeleteJourneySegment(d.Id()); err != nil {
		return util.BuildAPIDiagnosticError("genesyscloud_journey_segment", fmt.Sprintf("Failed to delete journey segment with display name %s: %s", displayName, err), resp)
	}

	return util.WithRetries(ctx, 30*time.Second, func() *retry.RetryError {
		_, resp, err := journeyApi.GetJourneySegment(d.Id())
		if err != nil {
			if util.IsStatus404(resp) {
				// journey segment deleted
				log.Printf("Deleted journey segment %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError("genesyscloud_journey_segment", fmt.Sprintf("error deleting journey segment %s | error: %s", d.Id(), err), resp))
		}

		return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError("genesyscloud_journey_segment", fmt.Sprintf("journey segment %s still exists", d.Id()), resp))
	})
}

func flattenJourneySegment(d *schema.ResourceData, journeySegment *platformclientv2.Journeysegment) {
	d.Set("is_active", *journeySegment.IsActive)
	d.Set("display_name", *journeySegment.DisplayName)
	resourcedata.SetNillableValue(d, "description", journeySegment.Description)
	d.Set("color", *journeySegment.Color)
	d.Set("scope", *journeySegment.Scope)
	resourcedata.SetNillableValue(d, "should_display_to_agent", journeySegment.ShouldDisplayToAgent)
	resourcedata.SetNillableValue(d, "context", lists.FlattenAsList(journeySegment.Context, flattenContext))
	resourcedata.SetNillableValue(d, "journey", lists.FlattenAsList(journeySegment.Journey, flattenJourney))
}

func buildSdkJourneySegment(journeySegment *schema.ResourceData) *platformclientv2.Journeysegmentrequest {
	isActive := journeySegment.Get("is_active").(bool)
	displayName := journeySegment.Get("display_name").(string)
	description := resourcedata.GetNillableValue[string](journeySegment, "description")
	color := journeySegment.Get("color").(string)
	scope := journeySegment.Get("scope").(string)
	shouldDisplayToAgent := resourcedata.GetNillableBool(journeySegment, "should_display_to_agent")
	sdkContext := resourcedata.BuildSdkListFirstElement(journeySegment, "context", buildSdkRequestContext, false)
	journey := resourcedata.BuildSdkListFirstElement(journeySegment, "journey", buildSdkRequestJourney, false)

	return &platformclientv2.Journeysegmentrequest{
		IsActive:             &isActive,
		DisplayName:          &displayName,
		Description:          description,
		Color:                &color,
		Scope:                &scope,
		ShouldDisplayToAgent: shouldDisplayToAgent,
		Context:              sdkContext,
		Journey:              journey,
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
	return &sdkPatchSegment
}

func flattenContext(context *platformclientv2.Context) map[string]interface{} {
	if len(*context.Patterns) == 0 {
		return nil
	}
	contextMap := make(map[string]interface{})
	contextMap["patterns"] = *lists.FlattenList(context.Patterns, flattenContextPattern)
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
	contextPatternMap["criteria"] = *lists.FlattenList(contextPattern.Criteria, flattenEntityTypeCriteria)
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
	entityTypeCriteriaMap["key"] = *entityTypeCriteria.Key
	entityTypeCriteriaMap["values"] = lists.StringListToSet(*entityTypeCriteria.Values)
	entityTypeCriteriaMap["should_ignore_case"] = *entityTypeCriteria.ShouldIgnoreCase
	entityTypeCriteriaMap["operator"] = *entityTypeCriteria.Operator
	entityTypeCriteriaMap["entity_type"] = *entityTypeCriteria.EntityType
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
	journeyMap["patterns"] = *lists.FlattenList(journey.Patterns, flattenJourneyPattern)
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
	journeyPatternMap["criteria"] = *lists.FlattenList(journeyPattern.Criteria, flattenCriteria)
	journeyPatternMap["count"] = *journeyPattern.Count
	journeyPatternMap["stream_type"] = *journeyPattern.StreamType
	journeyPatternMap["session_type"] = *journeyPattern.SessionType
	stringmap.SetValueIfNotNil(journeyPatternMap, "event_name", journeyPattern.EventName)
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
	criteriaMap["key"] = *criteria.Key
	criteriaMap["values"] = lists.StringListToSet(*criteria.Values)
	criteriaMap["should_ignore_case"] = *criteria.ShouldIgnoreCase
	criteriaMap["operator"] = *criteria.Operator
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
