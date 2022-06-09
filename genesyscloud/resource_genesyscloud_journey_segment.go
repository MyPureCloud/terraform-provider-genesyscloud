package genesyscloud

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
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
			"shouldIgnoreCase": {
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
			"entityType": {
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
			"shouldIgnoreCase": {
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
			"streamType": {
				Description:  "The stream type for which this pattern can be matched on.Valid values: Web, Custom, Conversation.",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"Web", "Custom", "Conversation"}, false),
			},
			"sessionType": {
				Description: "The session type for which this pattern can be matched on.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"eventName": {
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
		//UpdateContext: updateWithPooledClient(updateJourneySegment),
		//DeleteContext: deleteWithPooledClient(deleteJourneySegment),
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
	isActive := d.Get("isActive").(bool)
	displayName := d.Get("displayName").(string)
	version := d.Get("version").(int)
	description := d.Get("description").(string)
	color := d.Get("color").(string)
	scope := d.Get("scope").(string)
	shouldDisplayToAgent := d.Get("shouldDisplayToAgent").(bool)
	sdkContext := buildSdkContext(d)
	//journey := d.Get("journey").(journey)
	externalSegment := buildSdkExternalSegment(d)

	assignmentExpirationDays := d.Get("assignmentExpirationDays").(int)
	selfUri := d.Get("selfUri").(string)

	sdkConfig := meta.(*providerMeta).ClientConfig
	journeyApi := platformclientv2.NewJourneyApiWithConfig(sdkConfig)

	log.Printf("Creating journey segment %s", displayName)
	journeySegment, _, err := journeyApi.PostJourneySegments(platformclientv2.Journeysegment{
		IsActive:             &isActive,
		DisplayName:          &displayName,
		Version:              &version,
		Description:          &description,
		Color:                &color,
		Scope:                &scope,
		ShouldDisplayToAgent: &shouldDisplayToAgent,
		Context:              sdkContext,
		//Journey:         &journey,
		ExternalSegment:          externalSegment,
		AssignmentExpirationDays: &assignmentExpirationDays,
		SelfUri:                  &selfUri,
	})
	if err != nil {
		return diag.Errorf("Failed to create journey segment %s: %s", displayName, err)
	}

	d.SetId(*journeySegment.Id)

	log.Printf("Created journey segment %s %s", displayName, *journeySegment.Id)
	//return readJourneySegment(ctx, d, meta)
	return nil
}

/*
func readJourneySegment(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*providerMeta).ClientConfig
	telephonyApi := platformclientv2.NewTelephonyProvidersEdgeApiWithConfig(sdkConfig)

	log.Printf("Reading DID pool %s", d.Id())
	return withRetriesForRead(ctx, d, func() *resource.RetryError {
		journeySegment, resp, getErr := telephonyApi.GetTelephonyProvidersEdgesDidpool(d.Id())
		if getErr != nil {
			if isStatus404(resp) {
				return resource.RetryableError(fmt.Errorf("Failed to read DID pool %s: %s", d.Id(), getErr))
			}
			return resource.NonRetryableError(fmt.Errorf("Failed to read DID pool %s: %s", d.Id(), getErr))
		}

		if journeySegment.State != nil && *journeySegment.State == "deleted" {
			d.SetId("")
			return nil
		}

		cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, resourceJourneySegment())
		d.Set("start_phone_number", *journeySegment.StartPhoneNumber)
		d.Set("end_phone_number", *journeySegment.EndPhoneNumber)

		if journeySegment.Description != nil {
			d.Set("description", *journeySegment.Description)
		} else {
			d.Set("description", nil)
		}

		if journeySegment.Comments != nil {
			d.Set("comments", *journeySegment.Comments)
		} else {
			d.Set("comments", nil)
		}

		if journeySegment.Provider != nil {
			d.Set("pool_provider", *journeySegment.Provider)
		} else {
			d.Set("pool_provider", nil)
		}

		log.Printf("Read DID pool %s %s", d.Id(), *journeySegment.StartPhoneNumber)
		return cc.CheckState()
	})
}

func updateJourneySegment(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	startPhoneNumber := d.Get("start_phone_number").(string)
	endPhoneNumber := d.Get("end_phone_number").(string)
	description := d.Get("description").(string)
	comments := d.Get("comments").(string)
	poolProvider := d.Get("pool_provider").(string)

	sdkConfig := meta.(*providerMeta).ClientConfig
	telephonyApi := platformclientv2.NewTelephonyProvidersEdgeApiWithConfig(sdkConfig)

	journeySegmentBody := platformclientv2.JourneySegment{
		StartPhoneNumber: &startPhoneNumber,
		EndPhoneNumber:   &endPhoneNumber,
		Description:      &description,
		Comments:         &comments,
		Provider:         &poolProvider,
	}

	log.Printf("Updating DID pool %s", d.Id())
	if _, _, err := telephonyApi.PutTelephonyProvidersEdgesDidpool(d.Id(), journeySegmentBody); err != nil {
		return diag.Errorf("Error updating DID pool %s: %s", startPhoneNumber, err)
	}

	log.Printf("Updated DID pool %s", d.Id())
	return readJourneySegment(ctx, d, meta)
}

func deleteJourneySegment(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	startPhoneNumber := d.Get("start_phone_number").(string)

	sdkConfig := meta.(*providerMeta).ClientConfig
	telephonyApi := platformclientv2.NewTelephonyProvidersEdgeApiWithConfig(sdkConfig)

	log.Printf("Deleting DID pool with starting number %s", startPhoneNumber)
	if _, err := telephonyApi.DeleteTelephonyProvidersEdgesDidpool(d.Id()); err != nil {
		return diag.Errorf("Failed to delete DID pool with starting number %s: %s", startPhoneNumber, err)
	}

	return withRetries(ctx, 30*time.Second, func() *resource.RetryError {
		journeySegment, resp, err := telephonyApi.GetTelephonyProvidersEdgesDidpool(d.Id())
		if err != nil {
			if isStatus404(resp) {
				// DID pool deleted
				log.Printf("Deleted DID pool %s", d.Id())
				return nil
			}
			return resource.NonRetryableError(fmt.Errorf("Error deleting DID pool %s: %s", d.Id(), err))
		}

		if journeySegment.State != nil && *journeySegment.State == "deleted" {
			// DID pool deleted
			log.Printf("Deleted DID pool %s", d.Id())
			return nil
		}

		return resource.RetryableError(fmt.Errorf("DID pool %s still exists", d.Id()))
	})
}
*/

func buildSdkContext(d *schema.ResourceData) *platformclientv2.Context {
	contextArray := d.Get("context").([]interface{})
	if contextArray != nil {
		sdkContext := platformclientv2.Context{}
		if len(contextArray) > 0 {
			if _, ok := contextArray[0].(map[string]interface{}); !ok {
				return nil
			}

			contextSchema := contextArray[0].(*schema.ResourceData)

			patterns := buildSdkContextPattern(contextSchema)
			//patterns := buildSdkContextPattern(d.Get("contextArray").(*schema.ResourceData))

			sdkContext = platformclientv2.Context{
				Patterns: patterns,
			}
		}
		return &sdkContext
	}
	return nil
}

//func buildSdkContextPattern(d *schema.ResourceData) *[]platformclientv2.Contextpattern {
//	contextPattern := d.Get("patterns").([]interface{})
//	if contextPattern != nil {
//		var sdkContextPattern []platformclientv2.Contextpattern
//		if len(contextPattern) > 0 {
//			if _, ok := contextPattern[0].(map[string]interface{}); !ok {
//				return nil
//			}
//			contextPatternSchema := contextPattern[0].(*schema.ResourceData)
//
//			criteria := buildSdkContextPatternCriteria(contextPatternSchema)
//
//			sdkContextPattern = platformclientv2.Contextpattern{
//				Criteria:     criteria,
//			}
//		}
//		return &sdkContextPattern
//	}
//	return nil
//}

func buildSdkContextPattern(d *schema.ResourceData) *[]platformclientv2.Contextpattern {
	var sdkContextPattern []platformclientv2.Contextpattern
	if patterns, ok := d.GetOk("patterns"); ok {
		patternList := patterns.(*schema.Set).List()
		for _, pattern := range patternList {
			patternMap := pattern.(map[string]interface{})
			criteria := buildSdkContextPatternCriteria(patternMap["criteria"].(*schema.ResourceData))
			contextPattern := platformclientv2.Contextpattern{
				Criteria: criteria,
			}
			sdkContextPattern = append(sdkContextPattern, contextPattern)
		}
	}
	return &sdkContextPattern
}

func buildSdkContextPatternCriteria(d *schema.ResourceData) *[]platformclientv2.Entitytypecriteria {
	var sdkCriteria []platformclientv2.Entitytypecriteria
	if criteria, ok := d.GetOk("criteria"); ok {
		criteriaList := criteria.(*schema.Set).List()
		for _, criterion := range criteriaList {
			criterionMap := criterion.(map[string]interface{})
			key := criterionMap["key"].(string)
			values := buildSdkStringList(criterion.(*schema.ResourceData), "values")
			shouldIgnoreCase := criterionMap["shouldIgnoreCase"].(bool)
			operator := criterionMap["operator"].(string)
			entityType := criterionMap["entityType"].(string)
			entityTypeCriteria := platformclientv2.Entitytypecriteria{
				Key:              &key,
				Values:           values,
				ShouldIgnoreCase: &shouldIgnoreCase,
				Operator:         &operator,
				EntityType:       &entityType,
			}
			sdkCriteria = append(sdkCriteria, entityTypeCriteria)
		}
	}
	return &sdkCriteria
}

//
//func buildSdkContextPatternCriteriaArray(d *schema.ResourceData) *platformclientv2.Entitytypecriteria {
//	contextPatternCriteriaList := d.Get("criteria").([]interface{})
//	if contextPatternCriteria != nil {
//		sdkContextPatternCriteria := []platformclientv2.Entitytypecriteria{}
//		if len(contextPatternCriteriaList) > 0 {
//			if _, ok := contextPatternCriteriaList[0].(map[string]interface{}); !ok {
//				return nil
//			}
//
//			sdkContextPatternCriteria = buildSdkContextPatternCriteria(contextPatternCriteriaList)
//		}
//		return &sdkContextPatternCriteria
//	}
//	return nil
//}

func buildSdkExternalSegment(d *schema.ResourceData) *platformclientv2.Externalsegment {
	externalSegment := d.Get("externalSegment").([]interface{})
	if externalSegment != nil {
		sdkExternalSegment := platformclientv2.Externalsegment{}
		if len(externalSegment) > 0 {
			if _, ok := externalSegment[0].(map[string]interface{}); !ok {
				return nil
			}
			externalSegmentMap := externalSegment[0].(map[string]interface{})

			// Only set non-empty values.
			id := externalSegmentMap["id"].(string)
			name := externalSegmentMap["name"].(string)
			source := externalSegmentMap["source"].(string)

			sdkExternalSegment = platformclientv2.Externalsegment{
				Id:     &id,
				Name:   &name,
				Source: &source,
			}
		}
		return &sdkExternalSegment
	}
	return nil
}
