package journey_outcome_predictor

import (
	"context"
	"fmt"
	"log"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"terraform-provider-genesyscloud/genesyscloud/util/constants"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"

	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v146/platformclientv2"
)

/*
The resource_genesyscloud_journey_outcome_predictor.go contains all of the methods that perform the core logic for a resource.
*/

// getAllAuthJourneyOutcomePredictor retrieves all of the journey outcome predictor via Terraform in the Genesys Cloud and is used for the exporter
func getAllAuthJourneyOutcomePredictors(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	op := getJourneyOutcomePredictorProxy(clientConfig)
	resources := make(resourceExporter.ResourceIDMetaMap)

	predictors, resp, err := op.getAllJourneyOutcomePredictor(ctx)
	if err != nil {
		return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get predictors: %v", err), resp)
	}

	for _, predictor := range *predictors {
		resources[*predictor.Id] = &resourceExporter.ResourceMeta{BlockLabel: *predictor.Id}
	}

	return resources, nil
}

// createJourneyOutcomePredictor is used by the journey_outcome_predictor resource to create Genesys cloud journey outcome predictor
func createJourneyOutcomePredictor(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	op := getJourneyOutcomePredictorProxy(sdkConfig)

	outcomeId := d.Get("outcome_id").(string)

	outcome := &platformclientv2.Outcomerefrequest{
		Id: &outcomeId,
	}

	predictorRequest := platformclientv2.Outcomepredictorrequest{
		Outcome: outcome,
	}

	log.Printf("Creating predictor for outcome %s", outcomeId)

	predictor, resp, err := op.createJourneyOutcomePredictor(ctx, &predictorRequest)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to create predictor: %s | error: %s", outcomeId, err), resp)
	}

	d.SetId(*predictor.Id)
	log.Printf("Created predictor %s", *predictor.Id)
	return readJourneyOutcomePredictor(ctx, d, meta)
}

// readJourneyOutcomePredictor is used by the journey_outcome_predictor resource to read an journey outcome predictor from genesys cloud
func readJourneyOutcomePredictor(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	op := getJourneyOutcomePredictorProxy(sdkConfig)
	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceJourneyOutcomePredictor(), constants.ConsistencyChecks(), ResourceType)

	log.Printf("Reading predictor %s", d.Id())

	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		predictor, resp, getErr := op.getJourneyOutcomePredictorById(ctx, d.Id())
		if getErr != nil {
			if util.IsStatus404(resp) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read predictor %s | error: %s", d.Id(), getErr), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read predictor %s | error: %s", d.Id(), getErr), resp))
		}

		d.Set("outcome_id", *predictor.Outcome.Id)

		return cc.CheckState(d)
	})
}

// deleteJourneyOutcomePredictor is used by the journey_outcome_predictor resource to delete an journey outcome predictor from Genesys cloud
func deleteJourneyOutcomePredictor(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	op := getJourneyOutcomePredictorProxy(sdkConfig)

	resp, err := op.deleteJourneyOutcomePredictor(ctx, d.Id())
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to delete predictor %s | error: %s", d.Id(), err), resp)
	}

	return util.WithRetries(ctx, 180*time.Second, func() *retry.RetryError {
		_, resp, err := op.getJourneyOutcomePredictorById(ctx, d.Id())

		if err == nil {
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Error deleting predictor %s | error: %s", d.Id(), err), resp))
		}
		if util.IsStatus404(resp) {
			// Success  : Predictor deleted
			log.Printf("Deleted predictor %s", d.Id())
			return nil
		}

		return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Predictor %s still exists", d.Id()), resp))
	})
}
