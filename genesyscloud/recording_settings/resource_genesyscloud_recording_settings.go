package recording_settings

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
The resource_genesyscloud_recording_settings.go contains all the methods that perform the core logic for the resource.
*/

// recordingSettingsId is the fixed ID assigned to this singleton resource
const recordingSettingsId = "recording_settings"

// getAllRecordingSettings fetches the organization recording settings for the exporter.
func getAllRecordingSettings(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	// Although this resource has only a single instance, we attempt to fetch the data from
	// the API in order to verify the user's permission to access this resource's API endpoint.
	proxy := getRecordingSettingsProxy(clientConfig)
	resources := make(resourceExporter.ResourceIDMetaMap)

	_, resp, err := proxy.getRecordingSettings(ctx)
	if err != nil {
		if util.IsStatus404(resp) {
			// Don't export if config doesn't exist
			return resources, nil
		}
		return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get %s resource due to error: %s", ResourceType, err), resp)
	}

	resources[recordingSettingsId] = &resourceExporter.ResourceMeta{BlockLabel: ResourceType}
	return resources, nil
}

// createRecordingSettings is used by the recording_settings resource to create Genesys Cloud recording settings.
// Since the API has no create operation, we assign a fixed ID and delegate to update.
func createRecordingSettings(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Printf("Creating recording settings")
	d.SetId(recordingSettingsId)
	return updateRecordingSettings(ctx, d, meta)
}

// readRecordingSettings is used by the recording_settings resource to read the recording settings from Genesys Cloud
func readRecordingSettings(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getRecordingSettingsProxy(sdkConfig)
	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceRecordingSettings(), constants.ConsistencyChecks(), ResourceType)

	log.Printf("Reading recording settings %s", d.Id())

	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		settings, resp, getErr := proxy.getRecordingSettings(ctx)
		if getErr != nil {
			if util.IsStatus404(resp) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read recording settings %s | error: %s", d.Id(), getErr), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read recording settings %s | error: %s", d.Id(), getErr), resp))
		}

		setRecordingSettingsToResourceData(d, settings)

		log.Printf("Read recording settings %s", d.Id())
		return cc.CheckState(d)
	})
}

// updateRecordingSettings is used by the recording_settings resource to update the recording settings in Genesys Cloud
func updateRecordingSettings(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getRecordingSettingsProxy(sdkConfig)
	settings := getRecordingSettingsFromResourceData(d)

	log.Printf("Updating recording settings %s", d.Id())

	_, resp, err := proxy.updateRecordingSettings(ctx, &settings)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to update recording settings: %s", err), resp)
	}

	log.Printf("Updated recording settings %s", d.Id())
	return readRecordingSettings(ctx, d, meta)
}

// deleteRecordingSettings is a no-op for the recording_settings resource. The organization recording
// settings singleton cannot be deleted; Terraform simply removes it from state.
func deleteRecordingSettings(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Printf("Deleting recording settings %s is a no-op; removing from state only", d.Id())
	return nil
}
