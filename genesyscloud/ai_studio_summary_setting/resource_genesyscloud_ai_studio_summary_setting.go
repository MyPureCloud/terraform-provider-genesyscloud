package ai_studio_summary_setting

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v178/platformclientv2"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/consistency_checker"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/constants"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/resourcedata"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

/*
The resource_genesyscloud_ai_studio_summary_setting.go contains all of the methods that perform the core logic for a resource.
*/

// getAllAuthAiStudioSummarySetting retrieves all of the ai studio summary setting via Terraform in the Genesys Cloud and is used for the exporter
func getAllAuthAiStudioSummarySettings(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	proxy := newAiStudioSummarySettingProxy(clientConfig)
	resources := make(resourceExporter.ResourceIDMetaMap)

	summarySettings, resp, err := proxy.getAllAiStudioSummarySetting(ctx, "")
	if err != nil {
		return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get ai studio summary setting: %v", err), resp)
	}

	for _, summarySetting := range *summarySettings {
		resources[*summarySetting.Id] = &resourceExporter.ResourceMeta{BlockLabel: *summarySetting.Name}
	}
	log.Printf("Successfully retrieved all ai studio summary settings")
	return resources, nil
}

// createAiStudioSummarySetting is used by the ai_studio_summary_setting resource to create Genesys cloud ai studio summary setting
func createAiStudioSummarySetting(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getAiStudioSummarySettingProxy(sdkConfig)

	aiStudioSummarySetting := getAiStudioSummarySettingFromResourceData(d)

	log.Printf("Creating ai studio summary setting %s", *aiStudioSummarySetting.Name)
	summarySetting, resp, err := proxy.createAiStudioSummarySetting(ctx, &aiStudioSummarySetting)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to create ai studio summary setting: %s", err), resp)
	}

	d.SetId(*summarySetting.Id)
	log.Printf("Created ai studio summary setting %s", *summarySetting.Id)
	return readAiStudioSummarySetting(ctx, d, meta)
}

// readAiStudioSummarySetting is used by the ai_studio_summary_setting resource to read an ai studio summary setting from genesys cloud
func readAiStudioSummarySetting(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getAiStudioSummarySettingProxy(sdkConfig)
	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceAiStudioSummarySetting(), constants.ConsistencyChecks(), ResourceType)

	log.Printf("Reading ai studio summary setting %s", d.Id())

	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		summarySetting, resp, getErr := proxy.getAiStudioSummarySettingById(ctx, d.Id())
		if getErr != nil {
			if util.IsStatus404(resp) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read ai studio summary setting %s: %s", d.Id(), getErr), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read ai studio summary setting %s: %s", d.Id(), getErr), resp))
		}

		resourcedata.SetNillableValue(d, "name", summarySetting.Name)
		resourcedata.SetNillableValue(d, "language", summarySetting.Language)
		resourcedata.SetNillableValue(d, "summary_type", summarySetting.SummaryType)
		resourcedata.SetNillableValue(d, "format", summarySetting.Format)
		resourcedata.SetNillableValueWithInterfaceArrayWithFunc(d, "mask_p_i_i", summarySetting.MaskPII, flattenSummarySettingPIIs)
		resourcedata.SetNillableValueWithInterfaceArrayWithFunc(d, "participant_labels", summarySetting.ParticipantLabels, flattenSummarySettingParticipantLabelss)
		resourcedata.SetNillableValue(d, "predefined_insights", summarySetting.PredefinedInsights)
		resourcedata.SetNillableValueWithInterfaceArrayWithFunc(d, "custom_entities", summarySetting.CustomEntities, flattenSummarySettingCustomEntitys)
		resourcedata.SetNillableValue(d, "setting_type", summarySetting.SettingType)
		resourcedata.SetNillableValue(d, "prompt", summarySetting.Prompt)

		log.Printf("Read ai studio summary setting %s %s", d.Id(), *summarySetting.Name)
		return cc.CheckState(d)
	})
}

// updateAiStudioSummarySetting is used by the ai_studio_summary_setting resource to update an ai studio summary setting in Genesys Cloud
func updateAiStudioSummarySetting(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getAiStudioSummarySettingProxy(sdkConfig)

	aiStudioSummarySetting := getAiStudioSummarySettingFromResourceData(d)

	log.Printf("Updating ai studio summary setting %s", *aiStudioSummarySetting.Name)
	summarySetting, resp, err := proxy.updateAiStudioSummarySetting(ctx, d.Id(), &aiStudioSummarySetting)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to update ai studio summary setting %s: %s", d.Id(), err), resp)
	}

	log.Printf("Updated ai studio summary setting %s", *summarySetting.Id)
	return readAiStudioSummarySetting(ctx, d, meta)
}

// deleteAiStudioSummarySetting is used by the ai_studio_summary_setting resource to delete an ai studio summary setting from Genesys cloud
func deleteAiStudioSummarySetting(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getAiStudioSummarySettingProxy(sdkConfig)

	resp, err := proxy.deleteAiStudioSummarySetting(ctx, d.Id())
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to delete ai studio summary setting %s: %s", d.Id(), err), resp)
	}

	return util.WithRetries(ctx, 180*time.Second, func() *retry.RetryError {
		_, resp, err := proxy.getAiStudioSummarySettingById(ctx, d.Id())

		if err != nil {
			if util.IsStatus404(resp) {
				log.Printf("Deleted ai studio summary setting %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Error deleting ai studio summary setting %s: %s", d.Id(), err), resp))
		}

		return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("ai studio summary setting %s still exists", d.Id()), resp))
	})
}
