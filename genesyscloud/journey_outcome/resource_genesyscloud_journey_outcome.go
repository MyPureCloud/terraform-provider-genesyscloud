package journey_outcome

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
	"github.com/mypurecloud/platform-client-sdk-go/v150/platformclientv2"
)

func getAllJourneyOutcomes(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	proxy := getJourneyOutcomeProxy(clientConfig)
	resources := make(resourceExporter.ResourceIDMetaMap)

	outComes, proxyResponse, getErr := proxy.getAllJourneyOutcomes(ctx)
	if getErr != nil {
		return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get page of journey outcomes: %s", getErr), proxyResponse)
	}

	for _, outCome := range *outComes {
		resources[*outCome.Id] = &resourceExporter.ResourceMeta{BlockLabel: *outCome.DisplayName}
	}

	return resources, nil
}

func createJourneyOutcome(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getJourneyOutcomeProxy(sdkConfig)
	journeyOutcome := buildSdkJourneyOutcome(d)

	log.Printf("Creating journey outcome %s", *journeyOutcome.DisplayName)
	outComeResponse, proxyResponse, err := proxy.createJourneyOutcome(ctx, journeyOutcome)
	if err != nil {
		input, _ := util.InterfaceToJson(*journeyOutcome)
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("failed to create journey action map %s: %s\n(input: %+v)", *outComeResponse.DisplayName, err, input), proxyResponse)
	}

	d.SetId(*outComeResponse.Id)

	log.Printf("Created journey outcome %s %s", *outComeResponse.DisplayName, *outComeResponse.Id)
	return readJourneyOutcome(ctx, d, meta)
}

func readJourneyOutcome(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getJourneyOutcomeProxy(sdkConfig)
	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceJourneyOutcome(), constants.ConsistencyChecks(), ResourceType)

	log.Printf("Reading journey outcome %s", d.Id())
	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		outComeResponse, proxyResponse, getErr := proxy.getJourneyOutcomeById(ctx, d.Id())
		if getErr != nil {
			if util.IsStatus404(proxyResponse) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("failed to read journey outcome %s | error: %s", d.Id(), getErr), proxyResponse))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("failed to read journey outcome %s | error: %s", d.Id(), getErr), proxyResponse))
		}

		flattenJourneyOutcome(d, outComeResponse)

		log.Printf("Read journey outcome %s %s", d.Id(), *outComeResponse.DisplayName)
		return cc.CheckState(d)
	})
}

func updateJourneyOutcome(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getJourneyOutcomeProxy(sdkConfig)
	patchOutcome := buildSdkPatchOutcome(d)

	log.Printf("Updating journey outcome %s", d.Id())
	diagErr := util.RetryWhen(util.IsVersionMismatch, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		// Get current journey outcome version
		outComeResponse, proxyResponse, getErr := proxy.getJourneyOutcomeById(ctx, d.Id())
		if getErr != nil {
			return proxyResponse, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("failed to read journey outcome %s error: %s", d.Id(), getErr), proxyResponse)
		}

		patchOutcome.Version = outComeResponse.Version
		_, proxyResponse, patchErr := proxy.updateJourneyOutcome(ctx, d.Id(), patchOutcome)
		if patchErr != nil {
			input, _ := util.InterfaceToJson(*patchOutcome)
			return proxyResponse, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Error updating journey outcome %s: %s\n(input: %+v)", *patchOutcome.DisplayName, patchErr, input), proxyResponse)
		}
		return proxyResponse, nil
	})
	if diagErr != nil {
		return diagErr
	}

	log.Printf("Updated journey outcome %s", d.Id())
	return readJourneyOutcome(ctx, d, meta)
}

func deleteJourneyOutcome(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getJourneyOutcomeProxy(sdkConfig)

	displayName := d.Get("display_name").(string)
	log.Printf("Deleting journey outcome with display name %s", displayName)
	if proxyResponse, err := proxy.deleteJourneyOutcome(ctx, d.Id()); err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("failed to delete journey outcome with display name %s error: %s", displayName, err), proxyResponse)
	}

	return util.WithRetries(ctx, 30*time.Second, func() *retry.RetryError {
		_, proxyResponse, err := proxy.getJourneyOutcomeById(ctx, d.Id())
		if err != nil {
			if util.IsStatus404(proxyResponse) {
				// journey action map deleted
				log.Printf("Deleted journey outcome %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("error deleting journey outcome %s | error: %s", d.Id(), err), proxyResponse))
		}
		return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("journey outcome %s still exists", d.Id()), proxyResponse))
	})
}
