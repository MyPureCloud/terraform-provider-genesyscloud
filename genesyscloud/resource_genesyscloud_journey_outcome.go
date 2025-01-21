package genesyscloud

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"terraform-provider-genesyscloud/genesyscloud/util/constants"
	"terraform-provider-genesyscloud/genesyscloud/util/stringmap"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"terraform-provider-genesyscloud/genesyscloud/util/resourcedata"

	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"

	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	lists "terraform-provider-genesyscloud/genesyscloud/util/lists"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/mypurecloud/platform-client-sdk-go/v150/platformclientv2"
)

var (
	journeyOutcomeSchema = map[string]*schema.Schema{
		"is_active": {
			Description: "Whether or not the outcome is active.",
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     true,
		},
		"display_name": {
			Description: "The display name of the outcome.",
			Type:        schema.TypeString,
			Required:    true,
		},
		"description": {
			Description: "A description of the outcome.",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"is_positive": {
			Description: "Whether or not the outcome is positive.",
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     true,
		},
		"context": {
			Description: "The context of the outcome.",
			Type:        schema.TypeSet,
			Optional:    true,
			MaxItems:    1,
			Elem:        contextResource,
		},
		"journey": {
			Description: "The pattern of rules defining the outcome.",
			Type:        schema.TypeSet,
			Optional:    true,
			MaxItems:    1,
			Elem:        journeyResource,
		},
		"associated_value_field": {
			Description: " The field from the event indicating the associated value. Associated_value_field needs `eventtypes` to be created, which is a feature coming soon. More details available here:  https://developer.genesys.cloud/commdigital/digital/webmessaging/journey/eventtypes  https://all.docs.genesys.com/ATC/Current/AdminGuide/Custom_sessions",
			Type:        schema.TypeSet,
			Optional:    true,
			MaxItems:    1,
			Elem:        associatedValueFieldResource,
		},
	}

	associatedValueFieldResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"data_type": {
				Description:  "The data type of the value field.Valid values: Number, Integer.",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"Number", "Integer"}, false),
			},
			"name": {
				Description: "The field name for extracting value from event.",
				Type:        schema.TypeString,
				Required:    true,
				ValidateFunc: validation.StringMatch(func() *regexp.Regexp {
					r, _ := regexp.Compile("^attributes\\..+\\.value$")
					return r
				}(), ""),
			},
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

func getAllJourneyOutcomes(_ context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(resourceExporter.ResourceIDMetaMap)
	journeyApi := platformclientv2.NewJourneyApiWithConfig(clientConfig)

	pageCount := 1 // Needed because of broken journey common paging
	for pageNum := 1; pageNum <= pageCount; pageNum++ {
		const pageSize = 100
		journeyOutcomes, resp, getErr := journeyApi.GetJourneyOutcomes(pageNum, pageSize, "", nil, nil, "")
		if getErr != nil {
			return nil, util.BuildAPIDiagnosticError("genesyscloud_journey_outcome", fmt.Sprintf("Failed to get page of journey outcomes error: %s", getErr), resp)
		}

		if journeyOutcomes.Entities == nil || len(*journeyOutcomes.Entities) == 0 {
			break
		}

		for _, journeyOutcome := range *journeyOutcomes.Entities {
			resources[*journeyOutcome.Id] = &resourceExporter.ResourceMeta{BlockLabel: *journeyOutcome.DisplayName}
		}

		pageCount = *journeyOutcomes.PageCount
	}

	return resources, nil
}

func JourneyOutcomeExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllJourneyOutcomes),
		RefAttrs:         map[string]*resourceExporter.RefAttrSettings{}, // No references
	}
}
func ResourceJourneyOutcome() *schema.Resource {
	return &schema.Resource{
		Description:   "Genesys Cloud Journey Outcome",
		CreateContext: provider.CreateWithPooledClient(createJourneyOutcome),
		ReadContext:   provider.ReadWithPooledClient(readJourneyOutcome),
		UpdateContext: provider.UpdateWithPooledClient(updateJourneyOutcome),
		DeleteContext: provider.DeleteWithPooledClient(deleteJourneyOutcome),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema:        journeyOutcomeSchema,
	}
}

func createJourneyOutcome(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	journeyApi := platformclientv2.NewJourneyApiWithConfig(sdkConfig)
	journeyOutcome := buildSdkJourneyOutcome(d)

	log.Printf("Creating journey outcome %s", *journeyOutcome.DisplayName)
	result, resp, err := journeyApi.PostJourneyOutcomes(*journeyOutcome)
	if err != nil {
		return util.BuildAPIDiagnosticError("genesyscloud_journey_outcome", fmt.Sprintf("Failed to create journey outcome %s error: %s", *journeyOutcome.DisplayName, err), resp)
	}

	d.SetId(*result.Id)

	log.Printf("Created journey outcome %s %s", *result.DisplayName, *result.Id)
	return readJourneyOutcome(ctx, d, meta)
}

func readJourneyOutcome(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	journeyApi := platformclientv2.NewJourneyApiWithConfig(sdkConfig)
	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceJourneyOutcome(), constants.ConsistencyChecks(), "genesyscloud_journey_outcome")

	log.Printf("Reading journey outcome %s", d.Id())
	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		journeyOutcome, resp, getErr := journeyApi.GetJourneyOutcome(d.Id())
		if getErr != nil {
			if util.IsStatus404(resp) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError("genesyscloud_journey_outcome", fmt.Sprintf("failed to read journey outcome %s | error: %s", d.Id(), getErr), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError("genesyscloud_journey_outcome", fmt.Sprintf("failed to read journey outcome %s | error: %s", d.Id(), getErr), resp))
		}

		flattenJourneyOutcome(d, journeyOutcome)

		log.Printf("Read journey outcome %s %s", d.Id(), *journeyOutcome.DisplayName)
		return cc.CheckState(d)
	})
}

func updateJourneyOutcome(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	journeyApi := platformclientv2.NewJourneyApiWithConfig(sdkConfig)
	patchOutcome := buildSdkPatchOutcome(d)

	log.Printf("Updating journey outcome %s", d.Id())
	diagErr := util.RetryWhen(util.IsVersionMismatch, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		// Get current journey outcome version
		journeyOutcome, resp, getErr := journeyApi.GetJourneyOutcome(d.Id())
		if getErr != nil {
			return nil, util.BuildAPIDiagnosticError("genesyscloud_journey_outcome", fmt.Sprintf("Failed to read journey outcome %s error: %s", d.Id(), getErr), resp)
		}

		patchOutcome.Version = journeyOutcome.Version
		_, resp, patchErr := journeyApi.PatchJourneyOutcome(d.Id(), *patchOutcome)
		if patchErr != nil {
			return resp, util.BuildAPIDiagnosticError("genesyscloud_journey_outcome", fmt.Sprintf("Failed to update journey outcome %s error: %s", *patchOutcome.DisplayName, patchErr), resp)
		}
		return resp, nil
	})
	if diagErr != nil {
		return diagErr
	}

	log.Printf("Updated journey outcome %s", d.Id())
	return readJourneyOutcome(ctx, d, meta)
}

func deleteJourneyOutcome(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	displayName := d.Get("display_name").(string)

	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	journeyApi := platformclientv2.NewJourneyApiWithConfig(sdkConfig)

	log.Printf("Deleting journey outcome with display name %s", displayName)
	if resp, err := journeyApi.DeleteJourneyOutcome(d.Id()); err != nil {
		return util.BuildAPIDiagnosticError("genesyscloud_journey_outcome", fmt.Sprintf("Failed to delete journey outcome %s error: %s", displayName, err), resp)
	}

	return util.WithRetries(ctx, 30*time.Second, func() *retry.RetryError {
		_, resp, err := journeyApi.GetJourneyOutcome(d.Id())
		if err != nil {
			if util.IsStatus404(resp) {
				// journey outcome deleted
				log.Printf("Deleted journey outcome %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError("genesyscloud_journey_outcome", fmt.Sprintf("error deleting journey outcome %s | error: %s", d.Id(), err), resp))
		}

		return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError("genesyscloud_journey_outcome", fmt.Sprintf("journey outcome %s still exists", d.Id()), resp))
	})
}

func flattenJourneyOutcome(d *schema.ResourceData, journeyOutcome *platformclientv2.Outcome) {
	d.Set("is_active", *journeyOutcome.IsActive)
	d.Set("display_name", *journeyOutcome.DisplayName)
	resourcedata.SetNillableValue(d, "description", journeyOutcome.Description)
	resourcedata.SetNillableValue(d, "is_positive", journeyOutcome.IsPositive)
	resourcedata.SetNillableValue(d, "context", lists.FlattenAsList(journeyOutcome.Context, flattenContext))
	resourcedata.SetNillableValue(d, "journey", lists.FlattenAsList(journeyOutcome.Journey, flattenJourney))

	resourcedata.SetNillableValue(d, "associated_value_field", lists.FlattenAsList(journeyOutcome.AssociatedValueField, flattenAssociatedValueField))
}

func flattenContext(context *platformclientv2.Context) map[string]interface{} {
	if len(*context.Patterns) == 0 {
		return nil
	}
	contextMap := make(map[string]interface{})
	contextMap["patterns"] = *lists.FlattenList(context.Patterns, flattenContextPattern)
	return contextMap
}

func flattenJourney(journey *platformclientv2.Journey) map[string]interface{} {
	if len(*journey.Patterns) == 0 {
		return nil
	}
	journeyMap := make(map[string]interface{})
	journeyMap["patterns"] = *lists.FlattenList(journey.Patterns, flattenJourneyPattern)
	return journeyMap
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

func flattenCriteria(criteria *platformclientv2.Criteria) map[string]interface{} {
	criteriaMap := make(map[string]interface{})
	criteriaMap["key"] = *criteria.Key
	criteriaMap["values"] = lists.StringListToSet(*criteria.Values)
	criteriaMap["should_ignore_case"] = *criteria.ShouldIgnoreCase
	criteriaMap["operator"] = *criteria.Operator
	return criteriaMap
}

func flattenContextPattern(contextPattern *platformclientv2.Contextpattern) map[string]interface{} {
	contextPatternMap := make(map[string]interface{})
	contextPatternMap["criteria"] = *lists.FlattenList(contextPattern.Criteria, flattenEntityTypeCriteria)
	return contextPatternMap
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

func flattenAssociatedValueField(associatedValueField *platformclientv2.Associatedvaluefield) map[string]interface{} {
	associatedValueFieldMap := make(map[string]interface{})
	associatedValueFieldMap["data_type"] = associatedValueField.DataType
	associatedValueFieldMap["name"] = associatedValueField.Name
	return associatedValueFieldMap
}

func buildSdkJourneyOutcome(journeyOutcome *schema.ResourceData) *platformclientv2.Outcomerequest {
	isActive := journeyOutcome.Get("is_active").(bool)
	displayName := journeyOutcome.Get("display_name").(string)
	description := resourcedata.GetNillableValue[string](journeyOutcome, "description")
	isPositive := resourcedata.GetNillableBool(journeyOutcome, "is_positive")
	sdkContext := resourcedata.BuildSdkListFirstElement(journeyOutcome, "context", buildSdkRequestContext, false)
	journey := resourcedata.BuildSdkListFirstElement(journeyOutcome, "journey", buildSdkRequestJourney, false)
	associatedValueField := resourcedata.BuildSdkListFirstElement(journeyOutcome, "associated_value_field", buildSdkAssociatedValueField, true)

	return &platformclientv2.Outcomerequest{
		IsActive:             &isActive,
		DisplayName:          &displayName,
		Description:          description,
		IsPositive:           isPositive,
		Context:              sdkContext,
		Journey:              journey,
		AssociatedValueField: associatedValueField,
	}
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

func buildSdkRequestContextPattern(contextPattern map[string]interface{}) *platformclientv2.Requestcontextpattern {
	return &platformclientv2.Requestcontextpattern{
		Criteria: stringmap.BuildSdkList(contextPattern, "criteria", buildSdkRequestEntityTypeCriteria),
	}
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

func buildSdkRequestJourney(journey map[string]interface{}) *platformclientv2.Requestjourney {
	patterns := &[]platformclientv2.Requestjourneypattern{}
	if journey != nil {
		patterns = stringmap.BuildSdkList(journey, "patterns", buildSdkRequestJourneyPattern)
	}
	return &platformclientv2.Requestjourney{
		Patterns: patterns,
	}
}

func buildSdkPatchOutcome(journeyOutcome *schema.ResourceData) *platformclientv2.Patchoutcome {
	isActive := journeyOutcome.Get("is_active").(bool)
	displayName := journeyOutcome.Get("display_name").(string)
	description := resourcedata.GetNillableValue[string](journeyOutcome, "description")
	isPositive := resourcedata.GetNillableBool(journeyOutcome, "is_positive")
	sdkContext := resourcedata.BuildSdkListFirstElement(journeyOutcome, "context", buildSdkPatchContext, false)
	journey := resourcedata.BuildSdkListFirstElement(journeyOutcome, "journey", buildSdkPatchJourney, false)

	return &platformclientv2.Patchoutcome{
		IsActive:    &isActive,
		DisplayName: &displayName,
		Description: description,
		IsPositive:  isPositive,
		Context:     sdkContext,
		Journey:     journey,
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

func buildSdkPatchContextPattern(contextPattern map[string]interface{}) *platformclientv2.Patchcontextpattern {
	return &platformclientv2.Patchcontextpattern{
		Criteria: stringmap.BuildSdkList(contextPattern, "criteria", buildSdkPatchEntityTypeCriteria),
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

func buildSdkPatchJourney(journey map[string]interface{}) *platformclientv2.Patchjourney {
	patterns := &[]platformclientv2.Patchjourneypattern{}
	if journey != nil {
		patterns = stringmap.BuildSdkList(journey, "patterns", buildSdkPatchJourneyPattern)
	}
	return &platformclientv2.Patchjourney{
		Patterns: patterns,
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

func buildSdkAssociatedValueField(associatedValueField map[string]interface{}) *platformclientv2.Associatedvaluefield {
	dataType := associatedValueField["data_type"].(string)
	name := associatedValueField["name"].(string)

	return &platformclientv2.Associatedvaluefield{
		DataType: &dataType,
		Name:     &name,
	}
}
