package speechandtextanalytics_settings

import (
	"context"
	"fmt"
	"log"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/consistency_checker"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/constants"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v195/platformclientv2"
)

/*
The resource_genesyscloud_speechandtextanalytics_settings.go contains all the methods that perform the core logic for the resource.
*/

// speechAndTextAnalyticsSettingsId is the fixed ID for this singleton resource
const speechAndTextAnalyticsSettingsId = "settings"

func getAllSpeechAndTextAnalyticsSettings(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	// Although this resource has only a single instance, we attempt to fetch the data
	// from the API in order to verify the user's permission to access the endpoint.
	proxy := getSpeechAndTextAnalyticsSettingsProxy(clientConfig)
	resources := make(resourceExporter.ResourceIDMetaMap)

	_, resp, err := proxy.getSpeechAndTextAnalyticsSettings(ctx)
	if err != nil {
		if util.IsStatus404(resp) {
			// Don't export if config doesn't exist
			return resources, nil
		}
		return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get %s resource due to error: %s", ResourceType, err), resp)
	}

	resources[speechAndTextAnalyticsSettingsId] = &resourceExporter.ResourceMeta{BlockLabel: ResourceType}
	return resources, nil
}

// createSpeechAndTextAnalyticsSettings is used by the resource to create the settings.
// This is a singleton resource, so create sets a fixed ID and delegates to update.
func createSpeechAndTextAnalyticsSettings(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Printf("Creating speech and text analytics settings")
	d.SetId(speechAndTextAnalyticsSettingsId)
	return updateSpeechAndTextAnalyticsSettings(ctx, d, meta)
}

// readSpeechAndTextAnalyticsSettings is used by the resource to read the settings from Genesys Cloud
func readSpeechAndTextAnalyticsSettings(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getSpeechAndTextAnalyticsSettingsProxy(sdkConfig)
	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceSpeechAndTextAnalyticsSettings(), constants.ConsistencyChecks(), ResourceType)

	log.Printf("Reading speech and text analytics settings %s", d.Id())

	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		settings, resp, getErr := proxy.getSpeechAndTextAnalyticsSettings(ctx)
		if getErr != nil {
			if util.IsStatus404(resp) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read speech and text analytics settings %s | error: %s", d.Id(), getErr), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read speech and text analytics settings %s | error: %s", d.Id(), getErr), resp))
		}

		setSpeechAndTextAnalyticsSettingsToResourceData(d, settings)

		log.Printf("Read speech and text analytics settings %s", d.Id())
		return cc.CheckState(d)
	})
}

// updateSpeechAndTextAnalyticsSettings is used by the resource to update the settings in Genesys Cloud
func updateSpeechAndTextAnalyticsSettings(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getSpeechAndTextAnalyticsSettingsProxy(sdkConfig)
	settings := getSpeechAndTextAnalyticsSettingsFromResourceData(d)

	log.Printf("Updating speech and text analytics settings %s", d.Id())

	_, resp, err := proxy.updateSpeechAndTextAnalyticsSettings(ctx, &settings)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to update speech and text analytics settings: %s", err), resp)
	}

	log.Printf("Updated speech and text analytics settings %s", d.Id())
	return readSpeechAndTextAnalyticsSettings(ctx, d, meta)
}

// deleteSpeechAndTextAnalyticsSettings is a no-op. This is a singleton org-level setting that
// cannot be deleted; Terraform simply drops it from state.
func deleteSpeechAndTextAnalyticsSettings(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return nil
}
