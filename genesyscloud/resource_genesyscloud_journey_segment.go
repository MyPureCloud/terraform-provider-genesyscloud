package genesyscloud

import (
	"context"
	"fmt"
	"log"
	"regexp"
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
	"github.com/mypurecloud/platform-client-sdk-go/v115/platformclientv2"
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
			Description:  "The target entity that a segment applies to.Valid values: Session, Customer. Changing the scope attribute will cause the existing journey_segment to be dropped and recreated with new ID.",
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true, // scope can be only set during creation
			ValidateFunc: validation.StringInSlice([]string{"Session", "Customer"}, false),
		},
		"should_display_to_agent": {
			Description: "Whether or not the segment should be displayed to agent/supervisor users.",
			Type:        schema.TypeBool,
			Optional:    true,
			// Customer scope only supports false for this value
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
		"external_segment": {
			Description: "Details of an entity corresponding to this segment in an external system.",
			Type:        schema.TypeSet,
			Optional:    true,
			MaxItems:    1,
			Elem:        externalSegmentResource, // can only be used with Customer scope
		},
		"assignment_expiration_days": {
			Description:  "Time, in days, from when the segment is assigned until it is automatically unassigned.",
			Type:         schema.TypeInt,
			Optional:     true,
			ValidateFunc: validation.IntBetween(2, 60),
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
				Description:  "The stream type for which this pattern can be matched on.Valid values: Web, Custom, Conversation.",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"Web" /*, "Custom", "Conversation"*/}, false), // Custom and Conversation seem not to be supported by the API despite the documentation
			},
			"session_type": {
				Description:  "The session type for which this pattern can be matched on.",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"web"}, false), // custom value seems not to be supported by the API despite the documentation
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

	externalSegmentResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"id": {
				Description: "Identifier for the external segment in the system where it originates from. Changing the id attribute will cause the journey_segment resource to be dropped and recreated with a new ID.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"name": {
				Description: "Name for the external segment in the system where it originates from.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"source": {
				Description:  "The external system where the segment originates from.Valid values: AdobeExperiencePlatform, Custom. Changing the source attribute will cause the journey_segment resource to be dropped and recreated with a new ID.",
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice([]string{"AdobeExperiencePlatform", "Custom"}, false),
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
		journeySegments, _, getErr := journeyApi.GetJourneySegments("", pageSize, pageNum, true, nil, nil, "")
		if getErr != nil {
			return nil, diag.Errorf("Failed to get page of journey segments: %v", getErr)
		}

		if journeySegments.Entities == nil || len(*journeySegments.Entities) == 0 {
			break
		}

		for _, journeySegment := range *journeySegments.Entities {
			resources[*journeySegment.Id] = &resourceExporter.ResourceMeta{Name: *journeySegment.DisplayName}
		}

		pageCount = *journeySegments.PageCount
	}

	return resources, nil
}

func JourneySegmentExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: GetAllWithPooledClient(getAllJourneySegments),
		RefAttrs:         map[string]*resourceExporter.RefAttrSettings{}, // No references
	}
}

func ResourceJourneySegment() *schema.Resource {
	return &schema.Resource{
		Description:   "Genesys Cloud Journey Segment",
		CreateContext: CreateWithPooledClient(createJourneySegment),
		ReadContext:   ReadWithPooledClient(readJourneySegment),
		UpdateContext: UpdateWithPooledClient(updateJourneySegment),
		DeleteContext: DeleteWithPooledClient(deleteJourneySegment),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema:        journeySegmentSchema,
	}
}

func createJourneySegment(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*ProviderMeta).ClientConfig
	journeyApi := platformclientv2.NewJourneyApiWithConfig(sdkConfig)
	journeySegment := buildSdkJourneySegment(d)

	log.Printf("Creating journey segment %s", *journeySegment.DisplayName)
	result, resp, err := journeyApi.PostJourneySegments(*journeySegment)
	if err != nil {
		input, _ := interfaceToJson(*journeySegment)
		return diag.Errorf("failed to create journey segment %s: %s\n(input: %+v)\n(resp: %s)", *journeySegment.DisplayName, err, input, GetBody(resp))
	}

	d.SetId(*result.Id)

	log.Printf("Created journey segment %s %s", *result.DisplayName, *result.Id)
	return readJourneySegment(ctx, d, meta)
}

func readJourneySegment(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*ProviderMeta).ClientConfig
	journeyApi := platformclientv2.NewJourneyApiWithConfig(sdkConfig)

	log.Printf("Reading journey segment %s", d.Id())
	return WithRetriesForRead(ctx, d, func() *retry.RetryError {
		journeySegment, resp, getErr := journeyApi.GetJourneySegment(d.Id())
		if getErr != nil {
			if IsStatus404(resp) {
				return retry.RetryableError(fmt.Errorf("failed to read journey segment %s: %s", d.Id(), getErr))
			}
			return retry.NonRetryableError(fmt.Errorf("failed to read journey segment %s: %s", d.Id(), getErr))
		}

		cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceJourneySegment())
		flattenJourneySegment(d, journeySegment)

		log.Printf("Read journey segment %s %s", d.Id(), *journeySegment.DisplayName)
		return cc.CheckState()
	})
}

func updateJourneySegment(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*ProviderMeta).ClientConfig
	journeyApi := platformclientv2.NewJourneyApiWithConfig(sdkConfig)
	patchSegment := buildSdkPatchSegment(d)

	log.Printf("Updating journey segment %s", d.Id())
	diagErr := RetryWhen(IsVersionMismatch, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		// Get current journey segment version
		journeySegment, resp, getErr := journeyApi.GetJourneySegment(d.Id())
		if getErr != nil {
			return resp, diag.Errorf("Failed to read current journey segment %s: %s", d.Id(), getErr)
		}

		patchSegment.Version = journeySegment.Version
		_, resp, patchErr := journeyApi.PatchJourneySegment(d.Id(), *patchSegment)
		if patchErr != nil {
			input, _ := interfaceToJson(*patchSegment)
			return resp, diag.Errorf("Error updating journey segment %s: %s\n(input: %+v)\n(resp: %s)", *patchSegment.DisplayName, patchErr, input, GetBody(resp))
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

	sdkConfig := meta.(*ProviderMeta).ClientConfig
	journeyApi := platformclientv2.NewJourneyApiWithConfig(sdkConfig)

	log.Printf("Deleting journey segment with display name %s", displayName)
	if _, err := journeyApi.DeleteJourneySegment(d.Id()); err != nil {
		return diag.Errorf("Failed to delete journey segment with display name %s: %s", displayName, err)
	}

	return WithRetries(ctx, 30*time.Second, func() *retry.RetryError {
		_, resp, err := journeyApi.GetJourneySegment(d.Id())
		if err != nil {
			if IsStatus404(resp) {
				// journey segment deleted
				log.Printf("Deleted journey segment %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(fmt.Errorf("error deleting journey segment %s: %s", d.Id(), err))
		}

		return retry.RetryableError(fmt.Errorf("journey segment %s still exists", d.Id()))
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
	resourcedata.SetNillableValue(d, "external_segment", lists.FlattenAsList(journeySegment.ExternalSegment, flattenExternalSegment))
	resourcedata.SetNillableValue(d, "assignment_expiration_days", journeySegment.AssignmentExpirationDays)
}

func buildSdkJourneySegment(journeySegment *schema.ResourceData) *platformclientv2.Journeysegment {
	isActive := journeySegment.Get("is_active").(bool)
	displayName := journeySegment.Get("display_name").(string)
	description := resourcedata.GetNillableValue[string](journeySegment, "description")
	color := journeySegment.Get("color").(string)
	scope := journeySegment.Get("scope").(string)
	shouldDisplayToAgent := resourcedata.GetNillableBool(journeySegment, "should_display_to_agent")
	sdkContext := resourcedata.BuildSdkListFirstElement(journeySegment, "context", buildSdkContext, false)
	journey := resourcedata.BuildSdkListFirstElement(journeySegment, "journey", buildSdkJourney, false)
	externalSegment := resourcedata.BuildSdkListFirstElement(journeySegment, "external_segment", buildSdkExternalSegment, true)
	assignmentExpirationDays := resourcedata.GetNillableValue[int](journeySegment, "assignment_expiration_days")

	return &platformclientv2.Journeysegment{
		IsActive:                 &isActive,
		DisplayName:              &displayName,
		Description:              description,
		Color:                    &color,
		Scope:                    &scope,
		ShouldDisplayToAgent:     shouldDisplayToAgent,
		Context:                  sdkContext,
		Journey:                  journey,
		ExternalSegment:          externalSegment,
		AssignmentExpirationDays: assignmentExpirationDays,
	}
}

func buildSdkPatchSegment(journeySegment *schema.ResourceData) *platformclientv2.Patchsegment {
	isActive := journeySegment.Get("is_active").(bool)
	displayName := journeySegment.Get("display_name").(string)
	description := resourcedata.GetNillableValue[string](journeySegment, "description")
	color := journeySegment.Get("color").(string)
	shouldDisplayToAgent := resourcedata.GetNillableBool(journeySegment, "should_display_to_agent")
	sdkContext := resourcedata.BuildSdkListFirstElement(journeySegment, "context", buildSdkContext, false)
	journey := resourcedata.BuildSdkListFirstElement(journeySegment, "journey", buildSdkJourney, false)
	externalSegment := resourcedata.BuildSdkListFirstElement(journeySegment, "external_segment", buildSdkPatchExternalSegment, true)
	assignmentExpirationDays := resourcedata.GetNillableValue[int](journeySegment, "assignment_expiration_days")

	sdkPatchSegment := platformclientv2.Patchsegment{}
	sdkPatchSegment.SetField("IsActive", &isActive)
	sdkPatchSegment.SetField("DisplayName", &displayName)
	sdkPatchSegment.SetField("Description", description)
	sdkPatchSegment.SetField("Color", &color)
	sdkPatchSegment.SetField("ShouldDisplayToAgent", shouldDisplayToAgent)
	sdkPatchSegment.SetField("Context", sdkContext)
	sdkPatchSegment.SetField("Journey", journey)
	sdkPatchSegment.SetField("ExternalSegment", externalSegment)
	sdkPatchSegment.SetField("AssignmentExpirationDays", assignmentExpirationDays)
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

func buildSdkContext(context map[string]interface{}) *platformclientv2.Context {
	patterns := &[]platformclientv2.Contextpattern{}
	if context != nil {
		patterns = stringmap.BuildSdkList(context, "patterns", buildSdkContextPattern)
	}
	return &platformclientv2.Context{
		Patterns: patterns,
	}
}

func flattenContextPattern(contextPattern *platformclientv2.Contextpattern) map[string]interface{} {
	contextPatternMap := make(map[string]interface{})
	contextPatternMap["criteria"] = *lists.FlattenList(contextPattern.Criteria, flattenEntityTypeCriteria)
	return contextPatternMap
}

func buildSdkContextPattern(contextPattern map[string]interface{}) *platformclientv2.Contextpattern {
	return &platformclientv2.Contextpattern{
		Criteria: stringmap.BuildSdkList(contextPattern, "criteria", buildSdkEntityTypeCriteria),
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

func buildSdkEntityTypeCriteria(entityTypeCriteria map[string]interface{}) *platformclientv2.Entitytypecriteria {
	key := entityTypeCriteria["key"].(string)
	values := stringmap.BuildSdkStringList(entityTypeCriteria, "values")
	shouldIgnoreCase := entityTypeCriteria["should_ignore_case"].(bool)
	operator := entityTypeCriteria["operator"].(string)
	entityType := entityTypeCriteria["entity_type"].(string)

	return &platformclientv2.Entitytypecriteria{
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

func buildSdkJourney(journey map[string]interface{}) *platformclientv2.Journey {
	patterns := &[]platformclientv2.Journeypattern{}
	if journey != nil {
		patterns = stringmap.BuildSdkList(journey, "patterns", buildSdkJourneyPattern)
	}
	return &platformclientv2.Journey{
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

func buildSdkJourneyPattern(journeyPattern map[string]interface{}) *platformclientv2.Journeypattern {
	criteria := stringmap.BuildSdkList(journeyPattern, "criteria", buildSdkCriteria)
	count := journeyPattern["count"].(int)
	streamType := journeyPattern["stream_type"].(string)
	sessionType := journeyPattern["session_type"].(string)
	eventName := stringmap.GetNonDefaultValue[string](journeyPattern, "event_name")

	return &platformclientv2.Journeypattern{
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

func buildSdkCriteria(criteria map[string]interface{}) *platformclientv2.Criteria {
	key := criteria["key"].(string)
	values := stringmap.BuildSdkStringList(criteria, "values")
	shouldIgnoreCase := criteria["should_ignore_case"].(bool)
	operator := criteria["operator"].(string)

	return &platformclientv2.Criteria{
		Key:              &key,
		Values:           values,
		ShouldIgnoreCase: &shouldIgnoreCase,
		Operator:         &operator,
	}
}

func flattenExternalSegment(externalSegment *platformclientv2.Externalsegment) map[string]interface{} {
	externalSegmentMap := make(map[string]interface{})
	externalSegmentMap["id"] = *externalSegment.Id
	externalSegmentMap["name"] = *externalSegment.Name
	externalSegmentMap["source"] = *externalSegment.Source
	return externalSegmentMap
}

func buildSdkExternalSegment(externalSegment map[string]interface{}) *platformclientv2.Externalsegment {
	id := externalSegment["id"].(string)
	name := externalSegment["name"].(string)
	source := externalSegment["source"].(string)

	return &platformclientv2.Externalsegment{
		Id:     &id,
		Name:   &name,
		Source: &source,
	}
}

func buildSdkPatchExternalSegment(externalSegment map[string]interface{}) *platformclientv2.Patchexternalsegment {
	name := externalSegment["name"].(string)

	return &platformclientv2.Patchexternalsegment{
		Name: &name,
	}
}
