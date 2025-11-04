package aistudio_summary_setting

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v171/platformclientv2"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/consistency_checker"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/constants"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/resourcedata"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

/*
The resource_genesyscloud_aistudio_summary_setting.go contains all of the methods that perform the core logic for a resource.
*/

// getAllAuthAistudioSummarySetting retrieves all of the aistudio summary setting via Terraform in the Genesys Cloud and is used for the exporter
func getAllAuthAistudioSummarySettings(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	proxy := newAistudioSummarySettingProxy(clientConfig)
	resources := make(resourceExporter.ResourceIDMetaMap)

	summarySettings, resp, err := proxy.getAllAistudioSummarySetting(ctx)
	if err != nil {
		return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get aistudio summary setting: %v", err), resp)
	}

	for _, summarySetting := range *summarySettings {
		resources[*summarySetting.Id] = &resourceExporter.ResourceMeta{BlockLabel: *summarySetting.Name}
	}

	return resources, nil
}

// createAistudioSummarySetting is used by the aistudio_summary_setting resource to create Genesys cloud aistudio summary setting
func createAistudioSummarySetting(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getAistudioSummarySettingProxy(sdkConfig)

	aistudioSummarySetting := getAistudioSummarySettingFromResourceData(d)

	log.Printf("Creating aistudio summary setting %s", *aistudioSummarySetting.Name)
	summarySetting, resp, err := proxy.createAistudioSummarySetting(ctx, &aistudioSummarySetting)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to create aistudio summary setting: %s", err), resp)
	}

	d.SetId(*summarySetting.Id)
	log.Printf("Created aistudio summary setting %s", *summarySetting.Id)
	return readAistudioSummarySetting(ctx, d, meta)
}

// readAistudioSummarySetting is used by the aistudio_summary_setting resource to read an aistudio summary setting from genesys cloud
func readAistudioSummarySetting(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getAistudioSummarySettingProxy(sdkConfig)
	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceAistudioSummarySetting(), constants.ConsistencyChecks(), ResourceType)

	log.Printf("Reading aistudio summary setting %s", d.Id())

	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		summarySetting, resp, getErr := proxy.getAistudioSummarySettingById(ctx, d.Id())
		if getErr != nil {
			if util.IsStatus404(resp) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read aistudio summary setting %s: %s", d.Id(), getErr), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read aistudio summary setting %s: %s", d.Id(), getErr), resp))
		}

		resourcedata.SetNillableValue(d, "name", summarySetting.Name)
		resourcedata.SetNillableValue(d, "language", summarySetting.Language)
		resourcedata.SetNillableValue(d, "summary_type", summarySetting.SummaryType)
		resourcedata.SetNillableValue(d, "format", summarySetting.Format)
		resourcedata.SetNillableValueWithInterfaceArrayWithFunc(d, "mask_p_i_i", summarySetting.MaskPII, func(item *platformclientv2.Summarysettingpii) []interface{} {
			if item == nil {
				return nil
			}
			tmp := []platformclientv2.Summarysettingpii{*item}
			return flattenSummarySettingPIIs(&tmp)
		})
		resourcedata.SetNillableValueWithInterfaceArrayWithFunc(d, "participant_labels", summarySetting.ParticipantLabels, func(item *platformclientv2.Summarysettingparticipantlabels) []interface{} {
			if item == nil {
				return nil
			}
			tmp := []platformclientv2.Summarysettingparticipantlabels{*item}
			return flattenSummarySettingParticipantLabelss(&tmp)
		})
		resourcedata.SetNillableValue(d, "predefined_insights", summarySetting.PredefinedInsights)
		resourcedata.SetNillableValueWithInterfaceArrayWithFunc(d, "custom_entities", summarySetting.CustomEntities, flattenSummarySettingCustomEntitys)
		resourcedata.SetNillableValue(d, "setting_type", summarySetting.SettingType)
		resourcedata.SetNillableValue(d, "prompt", summarySetting.Prompt)

		log.Printf("Read aistudio summary setting %s %s", d.Id(), *summarySetting.Name)
		return cc.CheckState(d)
	})
}

// updateAistudioSummarySetting is used by the aistudio_summary_setting resource to update an aistudio summary setting in Genesys Cloud
func updateAistudioSummarySetting(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getAistudioSummarySettingProxy(sdkConfig)

	aistudioSummarySetting := getAistudioSummarySettingFromResourceData(d)

	log.Printf("Updating aistudio summary setting %s", *aistudioSummarySetting.Name)
	summarySetting, resp, err := proxy.updateAistudioSummarySetting(ctx, d.Id(), &aistudioSummarySetting)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to update aistudio summary setting %s: %s", d.Id(), err), resp)
	}

	log.Printf("Updated aistudio summary setting %s", *summarySetting.Id)
	return readAistudioSummarySetting(ctx, d, meta)
}

// deleteAistudioSummarySetting is used by the aistudio_summary_setting resource to delete an aistudio summary setting from Genesys cloud
func deleteAistudioSummarySetting(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getAistudioSummarySettingProxy(sdkConfig)

	resp, err := proxy.deleteAistudioSummarySetting(ctx, d.Id())
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to delete aistudio summary setting %s: %s", d.Id(), err), resp)
	}

	return util.WithRetries(ctx, 180*time.Second, func() *retry.RetryError {
		_, resp, err := proxy.getAistudioSummarySettingById(ctx, d.Id())

		if err != nil {
			if util.IsStatus404(resp) {
				log.Printf("Deleted aistudio summary setting %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Error deleting aistudio summary setting %s: %s", d.Id(), err), resp))
		}

		return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("aistudio summary setting %s still exists", d.Id()), resp))
	})
}
