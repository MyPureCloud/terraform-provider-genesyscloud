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
	"github.com/mypurecloud/platform-client-sdk-go/v72/platformclientv2"
)

var (
	externalSegmentResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"id": {
				Description: "Identifier for the external segment in the system where it originates from.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"name": {
				Description: "Name for the external segment in the system where it originates from.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"source": {
				Description:  "The external system where the segment originates from.Valid values: AdobeExperiencePlatform, Custom.",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"AdobeExperiencePlatform", "Custom"}, false),
			},
		},
	}

	contextCriteriaResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"key": {
				Description: "The criteria key.",
				Type:        schema.TypeString,
				Required:    true,
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
				ValidateFunc: validation.StringInSlice([]string{"containsAll", "containsAny", "notContainsAll", "notContainsAny", "equal", "notEqual", "greaterThan", "greaterThanOrEqual", "lessThan", "lessThanOrEqual", "startsWith", "endsWith"}, false),
			},
			"entity_type": {
				Description:  "The entity to match the pattern against.Valid values: visit.",
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"visit"}, false),
			},
		},
	}

	journeyCriteriaResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"key": {
				Description: "The criteria key.",
				Type:        schema.TypeString,
				Required:    true,
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
				ValidateFunc: validation.StringInSlice([]string{"containsAll", "containsAny", "notContainsAll", "notContainsAny", "equal", "notEqual", "greaterThan", "greaterThanOrEqual", "lessThan", "lessThanOrEqual", "startsWith", "endsWith"}, false),
			},
		},
	}

	contextPatternResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"criteria": {
				Description: "A list of one or more criteria to satisfy.",
				Type:        schema.TypeSet,
				Required:    true,
				Elem:        contextCriteriaResource,
			},
		},
	}

	journeyPatternResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"criteria": {
				Description: "A list of one or more criteria to satisfy.",
				Type:        schema.TypeSet,
				Required:    true,
				Elem:        journeyCriteriaResource,
			},
			"count": {
				Description: "The number of times the pattern must match.",
				Type:        schema.TypeInt,
				Optional:    true,
			},
			"stream_type": {
				Description:  "The stream type for which this pattern can be matched on.Valid values: Web, Custom, Conversation.",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"Web", "Custom", "Conversation"}, false),
			},
			"session_type": {
				Description: "The session type for which this pattern can be matched on.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"event_name": {
				Description: "The name of the event for which this pattern can be matched on.",
				Type:        schema.TypeString,
				Optional:    true,
			},
		},
	}
)

func getAllJourneySegments(_ context.Context, clientConfig *platformclientv2.Configuration) (ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(ResourceIDMetaMap)
	journeyAPI := platformclientv2.NewJourneyApiWithConfig(clientConfig)

	for pageNum := 1; ; pageNum++ {
		const pageSize = 100
		journeySegments, _, getErr := journeyAPI.GetJourneySegments("", pageSize, pageNum, true, nil, nil, "")
		if getErr != nil {
			return nil, diag.Errorf("Failed to get page of journey segments: %v", getErr)
		}

		if journeySegments.Entities == nil || len(*journeySegments.Entities) == 0 {
			break
		}

		for _, journeySegment := range *journeySegments.Entities {
			resources[*journeySegment.Id] = &ResourceMeta{Name: *journeySegment.DisplayName}
		}
	}

	return resources, nil
}

func journeySegmentExporter() *ResourceExporter {
	return &ResourceExporter{
		GetResourcesFunc: getAllWithPooledClient(getAllJourneySegments),
		RefAttrs:         map[string]*RefAttrSettings{}, // No references
	}
}

func resourceJourneySegment() *schema.Resource {
	return &schema.Resource{
		Description: "Genesys Cloud Journey Segment",

		CreateContext: createWithPooledClient(createJourneySegment),
		//ReadContext:   readWithPooledClient(readJourneySegment),
		UpdateContext: updateWithPooledClient(updateJourneySegment),
		DeleteContext: deleteWithPooledClient(deleteJourneySegment),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"id": {
				Description: "The globally unique identifier for the object.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"isActive": {
				Description: "Whether or not the segment is active.",
				Type:        schema.TypeBool,
				Optional:    true,
			},
			"version": {
				Description: "The version of the segment.",
				Type:        schema.TypeInt,
				Optional:    true,
			},
			"displayName": {
				Description: "The display name of the segment.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"description": {
				Description: "A description of the segment",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"color": {
				Description: "The hexadecimal color value of the segment.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"scope": {
				Description: "The target entity that a segment applies to.Valid values: Session, Customer.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"shouldDisplayToAgent": {
				Description: "Whether or not the segment should be displayed to agent/supervisor users.",
				Type:        schema.TypeBool,
				Optional:    true,
			},
			"context": {
				Description: "The context of the segment.",
				Type:        schema.TypeSet,
				// 				MinItems:    1, // TODO: context and journey min 1
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"patterns": {
							Description: "A list of one or more patterns to match.",
							Type:        schema.TypeSet,
							Required:    true,
							Elem:        contextPatternResource,
						},
					},
				},
			},
			"journey": {
				Description: "The pattern of rules defining the segment.",
				Type:        schema.TypeSet,
				// 				MinItems:    1, // TODO: context and journey min 1
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"patterns": {
							Description: "A list of one or more patterns to match.",
							Type:        schema.TypeSet,
							Required:    true,
							Elem:        journeyPatternResource,
						},
					},
				},
			},
			"externalSegment": {
				Description: "Details of an entity corresponding to this segment in an external system.",
				Type:        schema.TypeSet,
				Optional:    true,
				MaxItems:    1,
				Elem:        externalSegmentResource,
			},
			"assignmentExpirationDays": {
				Description: "Time, in days, from when the segment is assigned until it is automatically unassigned.",
				Type:        schema.TypeInt,
				Optional:    true,
			},
			"selfUri": {
				Description: "The URI for this object.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"createdDate": {
				Description: "Timestamp indicating when the segment was created. Date time is represented as an ISO-8601 string. For example: yyyy-MM-ddTHH:mm:ss[.mmm]Z.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"modifiedDate": {
				Description: "Timestamp indicating when the the segment was last updated. Date time is represented as an ISO-8601 string. For example: yyyy-MM-ddTHH:mm:ss[.mmm]Z.",
				Type:        schema.TypeString,
				Optional:    true,
			},
		},
	}
}

func createJourneySegment(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*providerMeta).ClientConfig
	journeyApi := platformclientv2.NewJourneyApiWithConfig(sdkConfig)
	journeySegment := buildSdkJourneySegment(d)

	log.Printf("Creating journey segment %s", *journeySegment.DisplayName)

	result, _, err := journeyApi.PostJourneySegments(*journeySegment)
	if err != nil {
		return diag.Errorf("Failed to create journey segment %s: %s", *journeySegment.DisplayName, err)
	}

	d.SetId(*result.Id)

	log.Printf("Created journey segment %s %s", *result.DisplayName, *result.Id)
	//return readJourneySegment(ctx, d, meta)
	return nil
}

//func readJourneySegment(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
//	sdkConfig := meta.(*providerMeta).ClientConfig
//	telephonyApi := platformclientv2.NewTelephonyProvidersEdgeApiWithConfig(sdkConfig)
//
//	log.Printf("Reading DID pool %s", d.Id())
//	return withRetriesForRead(ctx, d, func() *resource.RetryError {
//		journeySegment, resp, getErr := telephonyApi.GetTelephonyProvidersEdgesDidpool(d.Id())
//		if getErr != nil {
//			if isStatus404(resp) {
//				return resource.RetryableError(fmt.Errorf("Failed to read DID pool %s: %s", d.Id(), getErr))
//			}
//			return resource.NonRetryableError(fmt.Errorf("Failed to read DID pool %s: %s", d.Id(), getErr))
//		}
//
//		if journeySegment.State != nil && *journeySegment.State == "deleted" {
//			d.SetId("")
//			return nil
//		}
//
//		cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, resourceJourneySegment())
//		d.Set("start_phone_number", *journeySegment.StartPhoneNumber)
//		d.Set("end_phone_number", *journeySegment.EndPhoneNumber)
//
//		if journeySegment.Description != nil {
//			d.Set("description", *journeySegment.Description)
//		} else {
//			d.Set("description", nil)
//		}
//
//		if journeySegment.Comments != nil {
//			d.Set("comments", *journeySegment.Comments)
//		} else {
//			d.Set("comments", nil)
//		}
//
//		if journeySegment.Provider != nil {
//			d.Set("pool_provider", *journeySegment.Provider)
//		} else {
//			d.Set("pool_provider", nil)
//		}
//
//		log.Printf("Read DID pool %s %s", d.Id(), *journeySegment.StartPhoneNumber)
//		return cc.CheckState()
//	})
//}

func updateJourneySegment(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*providerMeta).ClientConfig
	journeyApi := platformclientv2.NewJourneyApiWithConfig(sdkConfig)
	journeySegment := buildSdkPatchSegment(d)

	log.Printf("Updating journey segment %s", d.Id())
	if _, _, err := journeyApi.PatchJourneySegment(d.Id(), *journeySegment); err != nil {
		return diag.Errorf("Error updating journey segment %s: %s", journeySegment.DisplayName, err)
	}

	log.Printf("Updated journey segment %s", d.Id())
	return nil
	//return readJourneySegment(ctx, d, meta)
}

func deleteJourneySegment(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	displayName := d.Get("displayName").(string)

	sdkConfig := meta.(*providerMeta).ClientConfig
	journeyApi := platformclientv2.NewJourneyApiWithConfig(sdkConfig)

	log.Printf("Deleting jounrey segment with display name %s", displayName)
	if _, err := journeyApi.DeleteJourneySegment(d.Id()); err != nil {
		return diag.Errorf("Failed to delete journey segment with display name %s: %s", displayName, err)
	}

	return withRetries(ctx, 30*time.Second, func() *resource.RetryError {
		journeySegment, resp, err := journeyApi.GetJourneySegment(d.Id())
		if err != nil {
			if isStatus404(resp) {
				// journey segment deleted
				log.Printf("Deleted journey segment %s", d.Id())
				return nil
			}
			return resource.NonRetryableError(fmt.Errorf("error deleting journey segment %s: %s", d.Id(), err))
		}

		if journeySegment.IsActive != nil && !*journeySegment.IsActive {
			// journey segment inactive
			log.Printf("Inactive journey segment %s", d.Id())
			return nil
		}

		return resource.RetryableError(fmt.Errorf("journey segment %s still exists", d.Id()))
	})
}

func buildSdkJourneySegment(journeySegment *schema.ResourceData) *platformclientv2.Journeysegment {
	isActive := journeySegment.Get("isActive").(bool)
	displayName := journeySegment.Get("displayName").(string)
	version := journeySegment.Get("version").(int)
	description := journeySegment.Get("description").(string)
	color := journeySegment.Get("color").(string)
	scope := journeySegment.Get("scope").(string)
	shouldDisplayToAgent := journeySegment.Get("shouldDisplayToAgent").(bool)
	sdkContext := buildSdkGenericListFirstElement(journeySegment, "context", buildSdkContext)
	journey := buildSdkGenericListFirstElement(journeySegment, "journey", buildSdkJourney)
	externalSegment := buildSdkGenericListFirstElement(journeySegment, "externalSegment", buildSdkExternalSegment)

	assignmentExpirationDays := journeySegment.Get("assignmentExpirationDays").(int)
	selfUri := journeySegment.Get("selfUri").(string)

	return &platformclientv2.Journeysegment{
		IsActive:                 &isActive,
		DisplayName:              &displayName,
		Version:                  &version,
		Description:              &description,
		Color:                    &color,
		Scope:                    &scope,
		ShouldDisplayToAgent:     &shouldDisplayToAgent,
		Context:                  sdkContext,
		Journey:                  journey,
		ExternalSegment:          externalSegment,
		AssignmentExpirationDays: &assignmentExpirationDays,
		SelfUri:                  &selfUri,
	}
}

func buildSdkPatchSegment(journeySegment *schema.ResourceData) *platformclientv2.Patchsegment {
	isActive := journeySegment.Get("isActive").(bool)
	displayName := journeySegment.Get("displayName").(string)
	version := journeySegment.Get("version").(int)
	description := journeySegment.Get("description").(string)
	color := journeySegment.Get("color").(string)
	shouldDisplayToAgent := journeySegment.Get("shouldDisplayToAgent").(bool)
	sdkContext := buildSdkGenericListFirstElement(journeySegment, "context", buildSdkContext)
	journey := buildSdkGenericListFirstElement(journeySegment, "journey", buildSdkJourney)
	externalSegment := buildSdkGenericListFirstElement(journeySegment, "externalSegment", buildSdkPatchExternalSegment)

	assignmentExpirationDays := journeySegment.Get("assignmentExpirationDays").(int)
	selfUri := journeySegment.Get("selfUri").(string)

	return &platformclientv2.Patchsegment{
		IsActive:                 &isActive,
		DisplayName:              &displayName,
		Version:                  &version,
		Description:              &description,
		Color:                    &color,
		ShouldDisplayToAgent:     &shouldDisplayToAgent,
		Context:                  sdkContext,
		Journey:                  journey,
		ExternalSegment:          externalSegment,
		AssignmentExpirationDays: &assignmentExpirationDays,
		SelfUri:                  &selfUri,
	}
}

func buildSdkContext(context *schema.ResourceData) *platformclientv2.Context {
	return &platformclientv2.Context{
		Patterns: buildSdkGenericList(context, "patterns", buildSdkContextPattern),
	}
}

func buildSdkContextPattern(contextPattern *schema.ResourceData) *platformclientv2.Contextpattern {
	return &platformclientv2.Contextpattern{
		Criteria: buildSdkGenericList(contextPattern, "criteria", buildSdkEntityTypeCriteria),
	}
}

func buildSdkEntityTypeCriteria(entityTypeCriteria *schema.ResourceData) *platformclientv2.Entitytypecriteria {
	key := entityTypeCriteria.Get("key").(string)
	values := buildSdkStringList(entityTypeCriteria, "values")
	shouldIgnoreCase := entityTypeCriteria.Get("should_ignore_case").(bool)
	operator := entityTypeCriteria.Get("operator").(string)
	entityType := entityTypeCriteria.Get("entity_type").(string)

	return &platformclientv2.Entitytypecriteria{
		Key:              &key,
		Values:           values,
		ShouldIgnoreCase: &shouldIgnoreCase,
		Operator:         &operator,
		EntityType:       &entityType,
	}
}

func buildSdkJourney(journey *schema.ResourceData) *platformclientv2.Journey {
	return &platformclientv2.Journey{
		Patterns: buildSdkGenericList(journey, "patterns", buildSdkJourneyPattern),
	}
}

func buildSdkJourneyPattern(journeyPattern *schema.ResourceData) *platformclientv2.Journeypattern {
	criteria := buildSdkGenericList(journeyPattern, "criteria", buildSdkCriteria)
	count := journeyPattern.Get("count").(int)
	streamType := journeyPattern.Get("stream_type").(string)
	sessionType := journeyPattern.Get("session_type").(string)
	eventName := journeyPattern.Get("event_name").(string)

	return &platformclientv2.Journeypattern{
		Criteria:    criteria,
		Count:       &count,
		StreamType:  &streamType,
		SessionType: &sessionType,
		EventName:   &eventName,
	}
}

func buildSdkCriteria(criteria *schema.ResourceData) *platformclientv2.Criteria {
	key := criteria.Get("key").(string)
	values := buildSdkStringList(criteria, "values")
	shouldIgnoreCase := criteria.Get("should_ignore_case").(bool)
	operator := criteria.Get("operator").(string)

	return &platformclientv2.Criteria{
		Key:              &key,
		Values:           values,
		ShouldIgnoreCase: &shouldIgnoreCase,
		Operator:         &operator,
	}
}

func buildSdkExternalSegment(externalSegment *schema.ResourceData) *platformclientv2.Externalsegment {
	name := externalSegment.Get("name").(string)
	source := externalSegment.Get("source").(string)

	return &platformclientv2.Externalsegment{
		Name:   &name,
		Source: &source,
	}
}

func buildSdkPatchExternalSegment(externalSegment *schema.ResourceData) *platformclientv2.Patchexternalsegment {
	name := externalSegment.Get("name").(string)

	return &platformclientv2.Patchexternalsegment{
		Name: &name,
	}
}
