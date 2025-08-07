package quality_forms_evaluation

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/tfexporter_state"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/constants"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/resourcedata"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/consistency_checker"

	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v165/platformclientv2"
)

// getAllEvaluationForms retrieves all evaluation forms from Genesys Cloud
// It returns a map of resource IDs to resource metadata for use in the exporter
func getAllEvaluationForms(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	proxy := getQualityFormsEvaluationProxy(clientConfig)
	resources := make(resourceExporter.ResourceIDMetaMap)

	evaluationForms, proxyResponse, getErr := proxy.getAllQualityFormsEvaluation(ctx)
	if getErr != nil {
		return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get page of evaluation forms: %s", getErr), proxyResponse)
	}

	for _, evaluationForm := range *evaluationForms {
		if evaluationForm.Id == nil {
			continue // Skip if Id is nil as it's required for the map key
		}

		blockLabel := *evaluationForm.Id // Default to using Id as BlockLabel
		if evaluationForm.Name != nil {
			blockLabel = *evaluationForm.Name // Use Name if available
		}

		resources[*evaluationForm.Id] = &resourceExporter.ResourceMeta{BlockLabel: blockLabel}
	}

	return resources, nil
}

// createEvaluationForm creates a new evaluation form in Genesys Cloud
// It handles both the creation of the form and optionally publishing it if the published flag is set to true
func createEvaluationForm(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)
	published := d.Get("published").(bool)

	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getQualityFormsEvaluationProxy(sdkConfig)

	evaluationForm := &platformclientv2.Evaluationform{
		Name:           &name,
		QuestionGroups: buildSdkQuestionGroups(d),
	}

	log.Printf("Creating Evaluation Form %s", name)
	formResponse, proxyResponse, err := proxy.createQualityFormsEvaluation(ctx, evaluationForm)
	if err != nil {
		if formResponse != nil && formResponse.Name != nil {
			input, _ := util.InterfaceToJson(*evaluationForm)
			return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("failed to create evaluation form %s: %s\n(input: %+v)", *formResponse.Name, err, input), proxyResponse)
		}
		input, _ := util.InterfaceToJson(*evaluationForm)
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("failed to create evaluation form: %s\n(input: %+v)", err, input), proxyResponse)
	}

	// Make sure form is properly created
	time.Sleep(2 * time.Second)

	// Publishing
	if published {
		newDraftEval, proxyResponse, err := proxy.publishQualityFormsEvaluation(ctx, *formResponse.Id)
		if err != nil {
			return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("failed to publish evaluation form '%s': %s", *formResponse.Id, err), proxyResponse)
		}
		_ = d.Set("published_id", *formResponse.Id)
		d.SetId(*newDraftEval.Id)
	} else {
		d.Set("published_id", "")
		d.SetId(*formResponse.Id)
	}

	d.Set("context_id", *formResponse.ContextId)

	log.Printf("Created evaluation form %s %s", name, *formResponse.Id)
	return readEvaluationForm(ctx, d, meta)
}

// readEvaluationForm retrieves an evaluation form from Genesys Cloud and updates the Terraform state
// It handles special logic for the exporter to determine if a form is published
// and uses consistency checking to ensure the state is properly updated
func readEvaluationForm(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getQualityFormsEvaluationProxy(sdkConfig)
	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceEvaluationForm(), constants.ConsistencyChecks(), "genesyscloud_quality_forms_evaluation")

	log.Printf("Reading evaluation form %s", d.Id())
	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		evaluationForm, resp, getErr := proxy.getQualityFormsEvaluationById(ctx, d.Id())
		if getErr != nil {
			if util.IsStatus404(resp) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read evaluation form %s: %s", d.Id(), getErr), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read evaluation form %s: %s", d.Id(), getErr), resp))
		}

		resourcedata.SetNillableValue(d, "context_id", evaluationForm.ContextId)
		publishedVersions, resp, err := proxy.getQualityFormsEvaluationsBulkContexts(ctx, []string{*evaluationForm.ContextId})
		if err != nil {
			if util.IsStatus404(resp) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to retrieve a list of the latest published evaluation form versions %s", *evaluationForm.ContextId), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to retrieve a list of the latest published evaluation form versions %s", *evaluationForm.ContextId), resp))
		}
		if len(publishedVersions) > 0 {
			if publishedVersions[0].Id != nil {
				resourcedata.SetNillableValue(d, "published_id", publishedVersions[0].Id)
			}
		}

		// During an export, published should be set to true if there are any published versions of an evaluation form
		if tfexporter_state.IsExporterActive() {
			if len(publishedVersions) > 0 {
				_ = d.Set("published", true)
			} else {
				_ = d.Set("published", false)
			}
		}

		resourcedata.SetNillableValue(d, "name", evaluationForm.Name)
		if evaluationForm.QuestionGroups != nil {
			_ = d.Set("question_groups", flattenQuestionGroups(evaluationForm.QuestionGroups))
		}

		return cc.CheckState(d)
	})
}

// updateEvaluationForm updates an existing evaluation form in Genesys Cloud
// It handles the complexity of working with published forms by finding the latest unpublished version
// and manages the publishing state based on the 'published' attribute
func updateEvaluationForm(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)
	published := d.Get("published").(bool)

	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getQualityFormsEvaluationProxy(sdkConfig)

	// Get the latest unpublished version using the proxy method directly
	unpublishedVersionId, _, err := proxy.getEvaluationFormRecentVerId(ctx, d.Id())
	if err != nil {
		log.Printf("Failed to get latest unpublished version. Using '%s' instead. Error: %s", d.Id(), err.Error())
		unpublishedVersionId = d.Id()
	}

	updatedResourceDataIdAfterPut := false
	if d.HasChangesExcept("published") {
		idToUse := d.Id()
		if formIsPublishedRemotely(d) {
			idToUse = unpublishedVersionId
		}

		evaluationForm := &platformclientv2.Evaluationform{
			Name:           &name,
			QuestionGroups: buildSdkQuestionGroups(d),
		}

		log.Printf("Updating Evaluation Form %s", name)
		updatedForm, proxyResponse, err := proxy.updateQualityFormsEvaluation(ctx, idToUse, evaluationForm)
		if err != nil {
			input, _ := util.InterfaceToJson(*evaluationForm)
			return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to update evaluation form %s: %s\n(input: %+v)", name, err, input), proxyResponse)
		}
		d.SetId(*updatedForm.Id)
		updatedResourceDataIdAfterPut = true
	}

	// Set published property on evaluation form update.
	if d.HasChange("published") {
		if published {
			formId := d.Id()
			newDraftEval, proxyResponse, err := proxy.publishQualityFormsEvaluation(ctx, formId)
			if err != nil {
				return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("failed to publish evaluation '%s': %s", d.Id(), err), proxyResponse)
			}
			_ = d.Set("published_id", formId)
			d.SetId(*newDraftEval.Id)
		} else if !updatedResourceDataIdAfterPut {
			d.SetId(unpublishedVersionId)
		}
	}

	log.Printf("Updated evaluation form '%s'. ID: '%s'", name, d.Id())
	return readEvaluationForm(ctx, d, meta)
}

// deleteEvaluationForm deletes an evaluation form from Genesys Cloud
// It attempts to find the latest unpublished version of the form before deletion
// and uses retries to confirm the deletion was successful
func deleteEvaluationForm(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)

	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getQualityFormsEvaluationProxy(sdkConfig)

	// Get the latest unpublished version using the proxy method directly
	unpublishedVersionId, _, err := proxy.getEvaluationFormRecentVerId(ctx, d.Id())
	if err != nil {
		log.Printf("Failed to get latest unpublished version for form '%s'. Error: %s", name, err.Error())
	} else {
		d.SetId(unpublishedVersionId)
	}

	log.Printf("Deleting evaluation form %s", name)
	proxyResponse, err := proxy.deleteQualityFormsEvaluation(ctx, d.Id())
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to delete evaluation form '%s': %s", name, err), proxyResponse)
	}

	return util.WithRetries(ctx, 30*time.Second, func() *retry.RetryError {
		_, resp, err := proxy.getQualityFormsEvaluationById(ctx, d.Id())
		if err != nil {
			if util.IsStatus404(resp) {
				// Evaluation form deleted
				log.Printf("Deleted evaluation form %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Error deleting evaluation form %s: %s", d.Id(), err), resp))
		}

		return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Evaluation form %s still exists", d.Id()), resp))
	})
}
