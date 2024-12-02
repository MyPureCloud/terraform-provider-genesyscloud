package flow_outcome

import (
	"context"
	"fmt"
	"log"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"terraform-provider-genesyscloud/genesyscloud/util/constants"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v146/platformclientv2"

	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"

	"terraform-provider-genesyscloud/genesyscloud/util/resourcedata"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

/*
The resource_genesyscloud_flow_outcome.go contains all of the methods that perform the core logic for a resource.
*/

// getAllAuthFlowOutcome retrieves all of the flow outcome via Terraform in the Genesys Cloud and is used for the exporter
func getAllAuthFlowOutcomes(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	proxy := newFlowOutcomeProxy(clientConfig)
	resources := make(resourceExporter.ResourceIDMetaMap)

	flowOutcomes, resp, err := proxy.getAllFlowOutcome(ctx)
	if err != nil {
		return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get flow outcomes error: %s", err), resp)
	}

	for _, flowOutcome := range *flowOutcomes {
		resources[*flowOutcome.Id] = &resourceExporter.ResourceMeta{BlockLabel: *flowOutcome.Name}
	}
	return resources, nil
}

// createFlowOutcome is used by the flow_outcome resource to create Genesys cloud flow outcome
func createFlowOutcome(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getFlowOutcomeProxy(sdkConfig)

	flowOutcome := getFlowOutcomeFromResourceData(d)

	log.Printf("Creating flow outcome %s", *flowOutcome.Name)
	outcome, resp, err := proxy.createFlowOutcome(ctx, &flowOutcome)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to create flow outcome %s error: %s", *flowOutcome.Name, err), resp)
	}

	d.SetId(*outcome.Id)
	log.Printf("Created flow outcome %s", *outcome.Id)
	return readFlowOutcome(ctx, d, meta)
}

// readFlowOutcome is used by the flow_outcome resource to read an flow outcome from genesys cloud
func readFlowOutcome(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getFlowOutcomeProxy(sdkConfig)
	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceFlowOutcome(), constants.ConsistencyChecks(), ResourceType)

	log.Printf("Reading flow outcome %s", d.Id())

	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		flowOutcome, resp, getErr := proxy.getFlowOutcomeById(ctx, d.Id())

		if getErr != nil {
			if util.IsStatus404(resp) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read flow outcome %s | error: %s", d.Id(), getErr), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read flow outcome %s | error: %s", d.Id(), getErr), resp))
		}

		resourcedata.SetNillableValue(d, "name", flowOutcome.Name)
		resourcedata.SetNillableReferenceWritableDivision(d, "division_id", flowOutcome.Division)

		// The api is adding a description with a space to outcomes without a description with is causing a mismatch with the state
		// We need to check if the description is not " " before reading it
		if flowOutcome.Description != nil && *flowOutcome.Description != " " {
			_ = d.Set("description", *flowOutcome.Description)
		} else {
			_ = d.Set("description", nil)
		}

		log.Printf("Read flow outcome %s %s", d.Id(), *flowOutcome.Name)
		return cc.CheckState(d)
	})
}

// updateFlowOutcome is used by the flow_outcome resource to update an flow outcome in Genesys Cloud
func updateFlowOutcome(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getFlowOutcomeProxy(sdkConfig)

	flowOutcome := getFlowOutcomeFromResourceData(d)

	log.Printf("Updating flow outcome %s", *flowOutcome.Name)
	_, resp, err := proxy.updateFlowOutcome(ctx, d.Id(), &flowOutcome)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to update flow outcome %s error: %s", *flowOutcome.Name, err), resp)
	}

	log.Printf("Updated flow outcome %s", d.Id())
	return readFlowOutcome(ctx, d, meta)
}

// deleteFlowOutcome is used by the flow_outcome resource to delete an flow outcome from Genesys cloud
func deleteFlowOutcome(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return nil
}
