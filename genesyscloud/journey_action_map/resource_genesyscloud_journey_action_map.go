package journey_action_map

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
	"github.com/mypurecloud/platform-client-sdk-go/v149/platformclientv2"
)

func getAllJourneyActionMaps(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	proxy := getJourneyActionMapProxy(clientConfig)
	resources := make(resourceExporter.ResourceIDMetaMap)

	actionMaps, proxyResponse, getErr := proxy.getAllJourneyActionMaps(ctx)
	if getErr != nil {
		return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get page of journey action maps: %s", getErr), proxyResponse)
	}

	for _, actionMap := range *actionMaps {
		resources[*actionMap.Id] = &resourceExporter.ResourceMeta{BlockLabel: *actionMap.DisplayName}
	}

	return resources, nil
}

func createJourneyActionMap(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getJourneyActionMapProxy(sdkConfig)
	actionMap := buildSdkActionMap(d)

	log.Printf("Creating journey action map %s", *actionMap.DisplayName)
	actionMapResponse, proxyResponse, err := proxy.createJourneyActionMap(ctx, actionMap)
	if err != nil {
		input, _ := util.InterfaceToJson(*actionMap)
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("failed to create journey action map %s: %s\n(input: %+v)", *actionMapResponse.DisplayName, err, input), proxyResponse)
	}

	d.SetId(*actionMapResponse.Id)

	log.Printf("Created journey action map %s %s", *actionMapResponse.DisplayName, *actionMapResponse.Id)
	return readJourneyActionMap(ctx, d, meta)
}

func readJourneyActionMap(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getJourneyActionMapProxy(sdkConfig)
	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceJourneyActionMap(), constants.ConsistencyChecks(), ResourceType)

	log.Printf("Reading journey action map %s", d.Id())
	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		actionMapResponse, proxyResponse, getErr := proxy.getJourneyActionMapById(ctx, d.Id())
		if getErr != nil {
			if util.IsStatus404(proxyResponse) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("failed to read journey action map %s | error: %s", d.Id(), getErr), proxyResponse))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("failed to read journey action map %s | error: %s", d.Id(), getErr), proxyResponse))
		}

		flattenActionMap(d, actionMapResponse)

		log.Printf("Read journey action map %s %s", d.Id(), *actionMapResponse.DisplayName)
		return cc.CheckState(d)
	})
}

func updateJourneyActionMap(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getJourneyActionMapProxy(sdkConfig)
	patchActionMap := buildSdkPatchActionMap(d)

	log.Printf("Updating journey action map %s", d.Id())
	diagErr := util.RetryWhen(util.IsVersionMismatch, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		// Get current journey action map version
		actionMapResponse, proxyResponse, getErr := proxy.getJourneyActionMapById(ctx, d.Id())
		if getErr != nil {
			return proxyResponse, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("failed to read journey action map %s error: %s", d.Id(), getErr), proxyResponse)
		}

		patchActionMap.Version = actionMapResponse.Version
		_, proxyResponse, patchErr := proxy.updateJourneyActionMap(ctx, d.Id(), patchActionMap)
		if patchErr != nil {
			input, _ := util.InterfaceToJson(*patchActionMap)
			return proxyResponse, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Error updating journey action map %s: %s\n(input: %+v)", *patchActionMap.DisplayName, patchErr, input), proxyResponse)
		}
		return proxyResponse, nil
	})
	if diagErr != nil {
		return diagErr
	}

	log.Printf("Updated journey action map %s", d.Id())
	return readJourneyActionMap(ctx, d, meta)
}

func deleteJourneyActionMap(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getJourneyActionMapProxy(sdkConfig)

	displayName := d.Get("display_name").(string)
	log.Printf("Deleting journey action map with display name %s", displayName)
	if proxyResponse, err := proxy.deleteJourneyActionMap(ctx, d.Id()); err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("failed to delete journey action map with display name %s error: %s", displayName, err), proxyResponse)
	}

	return util.WithRetries(ctx, 30*time.Second, func() *retry.RetryError {
		_, proxyResponse, err := proxy.getJourneyActionMapById(ctx, d.Id())
		if err != nil {
			if util.IsStatus404(proxyResponse) {
				// journey action map deleted
				log.Printf("Deleted journey action map %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("error deleting journey action map %s | error: %s", d.Id(), err), proxyResponse))
		}
		return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("journey action map %s still exists", d.Id()), proxyResponse))
	})
}
