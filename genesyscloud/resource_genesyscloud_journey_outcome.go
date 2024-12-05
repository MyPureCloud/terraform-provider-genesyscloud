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

	"terraform-provider-genesyscloud/genesyscloud/util/resourcedata"

	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"

	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	lists "terraform-provider-genesyscloud/genesyscloud/util/lists"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/mypurecloud/platform-client-sdk-go/v146/platformclientv2"
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

func buildSdkAssociatedValueField(associatedValueField map[string]interface{}) *platformclientv2.Associatedvaluefield {
	dataType := associatedValueField["data_type"].(string)
	name := associatedValueField["name"].(string)

	return &platformclientv2.Associatedvaluefield{
		DataType: &dataType,
		Name:     &name,
	}
}
