package genesyscloud

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"log"
	"regexp"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v74/platformclientv2"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/consistency_checker"
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
			Description: " The field from the event indicating the associated value.",
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
)

func getAllJourneyOutcomes(_ context.Context, clientConfig *platformclientv2.Configuration) (ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(ResourceIDMetaMap)
	journeyApi := platformclientv2.NewJourneyApiWithConfig(clientConfig)

	pageCount := 1 // Needed because of broken journey common paging
	for pageNum := 1; pageNum <= pageCount; pageNum++ {
		const pageSize = 100
		journeyOutcomes, _, getErr := journeyApi.GetJourneyOutcomes(pageNum, pageSize, "", nil, nil, "")
		if getErr != nil {
			return nil, diag.Errorf("Failed to get page of journey outcomes: %v", getErr)
		}

		if journeyOutcomes.Entities == nil || len(*journeyOutcomes.Entities) == 0 {
			break
		}

		for _, journeyOutcome := range *journeyOutcomes.Entities {
			resources[*journeyOutcome.Id] = &ResourceMeta{Name: *journeyOutcome.DisplayName}
		}

		pageCount = *journeyOutcomes.PageCount
	}

	return resources, nil
}

func journeyOutcomeExporter() *ResourceExporter {
	return &ResourceExporter{
		GetResourcesFunc: getAllWithPooledClient(getAllJourneyOutcomes),
		RefAttrs:         map[string]*RefAttrSettings{}, // No references
	}
}
func resourceJourneyOutcome() *schema.Resource {
	return &schema.Resource{
		Description:   "Genesys Cloud Journey Outcome",
		CreateContext: createWithPooledClient(createJourneyOutcome),
		ReadContext:   readWithPooledClient(readJourneyOutcome),
		UpdateContext: updateWithPooledClient(updateJourneyOutcome),
		DeleteContext: deleteWithPooledClient(deleteJourneyOutcome),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema:        journeyOutcomeSchema,
	}
}

func createJourneyOutcome(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*providerMeta).ClientConfig
	journeyApi := platformclientv2.NewJourneyApiWithConfig(sdkConfig)
	journeyOutcome := buildSdkJourneyOutcome(d)

	log.Printf("Creating journey outcome %s", *journeyOutcome.DisplayName)
	result, resp, err := journeyApi.PostJourneyOutcomes(*journeyOutcome)
	if err != nil {
		return diag.Errorf("failed to create journey outcome %s: %s\n(input: %+v)\n(resp: %s)", *journeyOutcome.DisplayName, err, *journeyOutcome, resp.RawBody)
	}

	d.SetId(*result.Id)

	log.Printf("Created journey outcome %s %s", *result.DisplayName, *result.Id)
	return readJourneyOutcome(ctx, d, meta)
}

func readJourneyOutcome(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*providerMeta).ClientConfig
	journeyApi := platformclientv2.NewJourneyApiWithConfig(sdkConfig)

	log.Printf("Reading journey outcome %s", d.Id())
	return withRetriesForRead(ctx, d, func() *resource.RetryError {
		journeyOutcome, resp, getErr := journeyApi.GetJourneyOutcome(d.Id())
		if getErr != nil {
			if isStatus404(resp) {
				return resource.RetryableError(fmt.Errorf("failed to read journey outcome %s: %s", d.Id(), getErr))
			}
			return resource.NonRetryableError(fmt.Errorf("failed to read journey outcome %s: %s", d.Id(), getErr))
		}

		cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, resourceJourneyOutcome())
		flattenJourneyOutcome(d, journeyOutcome)

		log.Printf("Read journey outcome %s %s", d.Id(), *journeyOutcome.DisplayName)
		return cc.CheckState()
	})
}

func updateJourneyOutcome(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*providerMeta).ClientConfig
	journeyApi := platformclientv2.NewJourneyApiWithConfig(sdkConfig)
	patchOutcome := buildSdkPatchOutcome(d)

	log.Printf("Updating journey outcome %s", d.Id())
	diagErr := retryWhen(isVersionMismatch, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		// Get current journey outcome version
		journeyOutcome, resp, getErr := journeyApi.GetJourneyOutcome(d.Id())
		if getErr != nil {
			return resp, diag.Errorf("Failed to read current journey outcome %s: %s", d.Id(), getErr)
		}

		patchOutcome.Version = journeyOutcome.Version
		_, resp, patchErr := journeyApi.PatchJourneyOutcome(d.Id(), *patchOutcome)
		if patchErr != nil {
			return resp, diag.Errorf("Error updating journey outcome %s: %s\n(input: %+v)\n(resp: %s)", *patchOutcome.DisplayName, patchErr, *patchOutcome, resp.RawBody)
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

	sdkConfig := meta.(*providerMeta).ClientConfig
	journeyApi := platformclientv2.NewJourneyApiWithConfig(sdkConfig)

	log.Printf("Deleting journey outcome with display name %s", displayName)
	if _, err := journeyApi.DeleteJourneyOutcome(d.Id()); err != nil {
		return diag.Errorf("Failed to delete journey outcome with display name %s: %s", displayName, err)
	}

	return withRetries(ctx, 30*time.Second, func() *resource.RetryError {
		_, resp, err := journeyApi.GetJourneyOutcome(d.Id())
		if err != nil {
			if isStatus404(resp) {
				// journey outcome deleted
				log.Printf("Deleted journey outcome %s", d.Id())
				return nil
			}
			return resource.NonRetryableError(fmt.Errorf("error deleting journey outcome %s: %s", d.Id(), err))
		}

		return resource.RetryableError(fmt.Errorf("journey outcome %s still exists", d.Id()))
	})
}

func flattenJourneyOutcome(d *schema.ResourceData, journeyOutcome *platformclientv2.Outcome) {
	d.Set("is_active", *journeyOutcome.IsActive)
	d.Set("display_name", *journeyOutcome.DisplayName)
	setNillableValue(d, "description", journeyOutcome.Description)
	setNillableValue(d, "is_positive", journeyOutcome.IsPositive)
	setNillableValue(d, "context", flattenGenericAsList(journeyOutcome.Context, flattenContext))
	setNillableValue(d, "journey", flattenGenericAsList(journeyOutcome.Journey, flattenJourney))
	setNillableValue(d, "associated_value_field", flattenGenericAsList(journeyOutcome.AssociatedValueField, flattenAssociatedValueField))
}

func flattenAssociatedValueField(associatedValueField *platformclientv2.Associatedvaluefield) map[string]interface{} {
	associatedValueFieldMap := make(map[string]interface{})
	associatedValueFieldMap["dataType"] = associatedValueField.DataType
	associatedValueFieldMap["name"] = associatedValueField.Name
	return associatedValueFieldMap
}

func buildSdkJourneyOutcome(journeyOutcome *schema.ResourceData) *platformclientv2.Outcome {
	isActive := getNillableBool(journeyOutcome, "is_active")
	displayName := getNillableValue[string](journeyOutcome, "display_name")
	description := getNillableValue[string](journeyOutcome, "description")
	isPositive := getNillableBool(journeyOutcome, "is_positive")
	sdkContext := buildSdkGenericListFirstElement(journeyOutcome, "context", buildSdkContext)
	journey := buildSdkGenericListFirstElement(journeyOutcome, "journey", buildSdkJourney)
	associatedValueField := buildSdkGenericListFirstElement(journeyOutcome, "associated_value_field", buildSdkAssociatedValueField)

	return &platformclientv2.Outcome{
		IsActive:             isActive,
		DisplayName:          displayName,
		Description:          description,
		IsPositive:           isPositive,
		Context:              sdkContext,
		Journey:              journey,
		AssociatedValueField: associatedValueField,
	}
}

func buildSdkAssociatedValueField(associatedValueField map[string]interface{}) *platformclientv2.Associatedvaluefield {
	if associatedValueField == nil {
		return nil
	}

	dataType := associatedValueField["data_type"].(string)
	name := associatedValueField["name"].(string)

	return &platformclientv2.Associatedvaluefield{
		DataType: &dataType,
		Name:     &name,
	}
}

func buildSdkPatchOutcome(journeyOutcome *schema.ResourceData) *platformclientv2.Patchoutcome {
	isActive := getNillableBool(journeyOutcome, "is_active")
	displayName := getNillableValue[string](journeyOutcome, "display_name")
	description := getNillableValue[string](journeyOutcome, "description")
	isPositive := getNillableBool(journeyOutcome, "is_positive")
	sdkContext := buildSdkGenericListFirstElement(journeyOutcome, "context", buildSdkContext)
	journey := buildSdkGenericListFirstElement(journeyOutcome, "journey", buildSdkJourney)

	return &platformclientv2.Patchoutcome{
		IsActive:    isActive,
		DisplayName: displayName,
		Description: description,
		IsPositive:  isPositive,
		Context:     sdkContext,
		Journey:     journey,
	}
}
