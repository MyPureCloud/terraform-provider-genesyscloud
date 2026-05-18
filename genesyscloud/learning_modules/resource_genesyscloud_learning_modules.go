package learning_modules

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v179/platformclientv2"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/consistency_checker"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/constants"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/resourcedata"
)

/*
The resource_genesyscloud_learning_modules.go contains all of the methods that perform the core logic for a resource.
*/

// getAllAuthLearningModules retrieves all of the learning modules via Terraform in the Genesys Cloud and is used for the exporter
func getAllAuthLearningModules(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	proxy := newLearningModulesProxy(clientConfig)
	resources := make(resourceExporter.ResourceIDMetaMap)

	modules, resp, err := proxy.getAllLearningModules(ctx, "")
	if err != nil {
		return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get learning modules error: %s", err), resp)
	}

	for _, module := range *modules {
		resources[*module.Id] = &resourceExporter.ResourceMeta{BlockLabel: *module.Name}
	}

	return resources, nil
}

// createLearningModule is used by the learning_modules resource to create Genesys cloud learning module
func createLearningModule(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getLearningModulesProxy(sdkConfig)

	var learningModule platformclientv2.Learningmodulerequest
	learningModule.Name = platformclientv2.String(d.Get("name").(string))
	description := d.Get("description").(string)
	if description != "" {
		learningModule.Description = &description
	}
	learningModule.CompletionTimeInDays = platformclientv2.Int(d.Get("completion_time_in_days").(int))
	learningModule.InformSteps = buildSdkInformSteps(d.Get("inform_steps").([]interface{}))
	moduleType := d.Get("type").(string)
	if moduleType != "" {
		learningModule.VarType = &moduleType
	}
	learningModule.AssessmentForm = buildSdkAssessmentForm(d.Get("assessment_form").([]interface{}))
	learningModule.CoverArt = buildSdkCoverArt(d.Get("cover_art_id").(string))
	lengthInMinutes := d.Get("length_in_minutes").(int)
	if lengthInMinutes > 0 {
		learningModule.LengthInMinutes = &lengthInMinutes
	}
	learningModule.ExcludedFromCatalog = platformclientv2.Bool(d.Get("excluded_from_catalog").(bool))
	externalId := d.Get("external_id").(string)
	if externalId != "" {
		learningModule.ExternalId = &externalId
	}
	learningModule.EnforceContentOrder = platformclientv2.Bool(d.Get("enforce_content_order").(bool))
	learningModule.ReviewAssessmentResults = buildSdkReviewAssessmentResults(d.Get("review_assessment_results").([]interface{}))
	learningModule.AutoAssign = buildSdkAutoAssign(d.Get("auto_assign").([]interface{}))

	log.Printf("Creating learning module %s", *learningModule.Name)
	module, resp, err := proxy.createLearningModule(ctx, &learningModule)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to create learning module %s error: %s", *learningModule.Name, err), resp)
	}

	d.SetId(*module.Id)
	log.Printf("Created learning module %s: %s", *module.Name, *module.Id)
	time.Sleep(2 * time.Second)

	published := d.Get("is_published").(bool)
	if published {
		publishedModule, resp, err := proxy.publishLearningModule(ctx, *module.Id)
		if err != nil {
			return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to publish learning module %s error: %s", *module.Name, err), resp)
		}
		log.Printf("Published learning module %s with version %d", *publishedModule.Id, *publishedModule.Version)
	}

	return readLearningModule(ctx, d, meta)
}

// readLearningModule is used by the learning_modules resource to read a learning module from genesys cloud
func readLearningModule(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getLearningModulesProxy(sdkConfig)
	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceLearningModules(), constants.ConsistencyChecks(), ResourceType)

	log.Printf("Reading learning module %s", d.Id())

	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		module, resp, getErr := proxy.getLearningModuleById(ctx, d.Id())
		if getErr != nil {
			if util.IsStatus404(resp) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read learning module %s | error: %s", d.Id(), getErr), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read learning module %s | error: %s", d.Id(), getErr), resp))
		}

		resourcedata.SetNillableValue(d, "name", module.Name)
		resourcedata.SetNillableValue(d, "description", module.Description)
		resourcedata.SetNillableValue(d, "completion_time_in_days", module.CompletionTimeInDays)
		if module.InformSteps != nil {
			d.Set("inform_steps", flattenInformSteps(module.InformSteps))
		}
		resourcedata.SetNillableValue(d, "type", module.VarType)
		if module.AssessmentForm != nil {
			d.Set("assessment_form", flattenAssessmentForm(module.AssessmentForm))
		}
		resourcedata.SetNillableValue(d, "cover_art_id", flattenCoverArt(module.CoverArt))
		resourcedata.SetNillableValue(d, "length_in_minutes", module.LengthInMinutes)
		resourcedata.SetNillableValue(d, "excluded_from_catalog", module.ExcludedFromCatalog)
		resourcedata.SetNillableValue(d, "external_id", module.ExternalId)
		resourcedata.SetNillableValue(d, "enforce_content_order", module.EnforceContentOrder)
		if module.ReviewAssessmentResults != nil {
			d.Set("review_assessment_results", flattenReviewAssessmentResults(module.ReviewAssessmentResults))
		}
		if module.AutoAssign != nil {
			d.Set("auto_assign", flattenAutoAssign(module.AutoAssign))
		}
		resourcedata.SetNillableValue(d, "is_published", module.IsPublished)

		log.Printf("Read learning module %s %s", d.Id(), *module.Name)
		return cc.CheckState(d)
	})
}

// updateLearningModule is used by the learning_modules resource to update a learning module in Genesys Cloud
func updateLearningModule(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getLearningModulesProxy(sdkConfig)

	var learningModule platformclientv2.Learningmodulerequest
	learningModule.Name = platformclientv2.String(d.Get("name").(string))
	description := d.Get("description").(string)
	if description != "" {
		learningModule.Description = &description
	}
	learningModule.CompletionTimeInDays = platformclientv2.Int(d.Get("completion_time_in_days").(int))
	learningModule.InformSteps = buildSdkInformSteps(d.Get("inform_steps").([]interface{}))
	moduleType := d.Get("type").(string)
	if moduleType != "" {
		learningModule.VarType = &moduleType
	}
	learningModule.AssessmentForm = buildSdkAssessmentForm(d.Get("assessment_form").([]interface{}))
	learningModule.CoverArt = buildSdkCoverArt(d.Get("cover_art_id").(string))
	lengthInMinutes := d.Get("length_in_minutes").(int)
	if lengthInMinutes > 0 {
		learningModule.LengthInMinutes = &lengthInMinutes
	}
	learningModule.ExcludedFromCatalog = platformclientv2.Bool(d.Get("excluded_from_catalog").(bool))
	externalId := d.Get("external_id").(string)
	if externalId != "" {
		learningModule.ExternalId = &externalId
	}
	learningModule.EnforceContentOrder = platformclientv2.Bool(d.Get("enforce_content_order").(bool))
	learningModule.ReviewAssessmentResults = buildSdkReviewAssessmentResults(d.Get("review_assessment_results").([]interface{}))
	learningModule.AutoAssign = buildSdkAutoAssign(d.Get("auto_assign").([]interface{}))

	log.Printf("Updating learning module %s: %s", *learningModule.Name, d.Id())

	module, resp, err := proxy.updateLearningModule(ctx, d.Id(), &learningModule)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to update learning module %s error: %s", *learningModule.Name, err), resp)
	}

	log.Printf("Updated learning module %s", *module.Id)

	time.Sleep(2 * time.Second)

	published := d.Get("is_published").(bool)
	if published {
		publishedModule, resp, err := proxy.publishLearningModule(ctx, d.Id())
		if err != nil {
			return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to publish learning module %s error: %s", d.Id(), err), resp)
		}
		log.Printf("Published learning module %s with version %d", *publishedModule.Id, *publishedModule.Version)
	}

	return readLearningModule(ctx, d, meta)
}

// deleteLearningModule is used by the learning_modules resource to delete a learning module from Genesys cloud
func deleteLearningModule(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getLearningModulesProxy(sdkConfig)

	resp, err := proxy.deleteLearningModule(ctx, d.Id())
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to delete learning module %s error: %s", d.Id(), err), resp)
	}

	return util.WithRetries(ctx, 180*time.Second, func() *retry.RetryError {
		_, resp, err := proxy.getLearningModuleById(ctx, d.Id())

		if err != nil {
			if util.IsStatus404(resp) {
				log.Printf("Deleted learning module %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Error deleting learning module %s | error: %s", d.Id(), err), resp))
		}
		return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("learning module %s still exists", d.Id()), resp))
	})
}
