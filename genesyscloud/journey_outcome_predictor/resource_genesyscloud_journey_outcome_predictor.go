package journey_outcome_predictor

import (
	"context"
	"fmt"
	"log"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"

	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v123/platformclientv2"
)

/*
The resource_genesyscloud_journey_outcome_predictor.go contains all of the methods that perform the core logic for a resource.
*/

// getAllAuthJourneyOutcomePredictor retrieves all of the journey outcome predictor via Terraform in the Genesys Cloud and is used for the exporter
func getAllAuthJourneyOutcomePredictors(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	op := getJourneyOutcomePredictorProxy(clientConfig)
	resources := make(resourceExporter.ResourceIDMetaMap)

	predictors, err := op.getAllJourneyOutcomePredictor(ctx)
	if err != nil {
		return nil, diag.Errorf("Failed to get predictors: %v", err)
	}

	for _, predictor := range *predictors {
		resources[*predictor.Id] = &resourceExporter.ResourceMeta{Name: *predictor.Id}
	}

	return resources, nil
}

// createJourneyOutcomePredictor is used by the journey_outcome_predictor resource to create Genesys cloud journey outcome predictor
func createJourneyOutcomePredictor(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	op := getJourneyOutcomePredictorProxy(sdkConfig)

	outcomeId := d.Get("outcome.0.id").(string)

	outcome := &platformclientv2.Outcomerefrequest{
		Id: &outcomeId,
	}

	predictorRequest := platformclientv2.Outcomepredictorrequest{
		Outcome: outcome,
	}

	log.Printf("Creating predictor for outcome %s", outcomeId)

	predictor, err := op.createJourneyOutcomePredictor(ctx, &predictorRequest)
	if err != nil {
		return diag.Errorf("Failed to create predictor: %s", err)
	}

	d.SetId(*predictor.Id)
	log.Printf("Created predictor %s", *predictor.Id)
	return readJourneyOutcomePredictor(ctx, d, meta)
}

// readJourneyOutcomePredictor is used by the journey_outcome_predictor resource to read an journey outcome predictor from genesys cloud
func readJourneyOutcomePredictor(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	op := getJourneyOutcomePredictorProxy(sdkConfig)

	log.Printf("Reading predictor %s", d.Id())

	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		predictor, respCode, getErr := op.getJourneyOutcomePredictorById(ctx, d.Id())
		if getErr != nil {
			if util.IsStatus404ByInt(respCode) {
				return retry.RetryableError(fmt.Errorf("Failed to read predictor %s: %s", d.Id(), getErr))
			}
			return retry.NonRetryableError(fmt.Errorf("Failed to read predictor %s: %s", d.Id(), getErr))
		}

		cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceJourneyOutcomePredictor())
		d.Set("outcome_id", *predictor.Outcome.Id)

		return cc.CheckState()
	})
}

// deleteJourneyOutcomePredictor is used by the journey_outcome_predictor resource to delete an journey outcome predictor from Genesys cloud
func deleteJourneyOutcomePredictor(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	op := getJourneyOutcomePredictorProxy(sdkConfig)

	_, err := op.deleteJourneyOutcomePredictor(ctx, d.Id())
	if err != nil {
		return diag.Errorf("Failed to delete predictor %s: %s", d.Id(), err)
	}

	return util.WithRetries(ctx, 180*time.Second, func() *retry.RetryError {
		_, respCode, err := op.getJourneyOutcomePredictorById(ctx, d.Id())

		if err == nil {
			return retry.NonRetryableError(fmt.Errorf("Error deleting predictor %s: %s", d.Id(), err))
		}
		if util.IsStatus404ByInt(respCode) {
			// Success  : External contact deleted
			log.Printf("Deleted predictor %s", d.Id())
			return nil
		}

		return retry.RetryableError(fmt.Errorf("Predictor %s still exists", d.Id()))
	})
}