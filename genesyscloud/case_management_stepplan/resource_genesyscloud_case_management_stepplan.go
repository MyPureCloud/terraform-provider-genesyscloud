package case_management_stepplan

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v186/platformclientv2"

	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/case_management_stageplan"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/consistency_checker"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/constants"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/resourcedata"
)

func getAllAuthCaseManagementStepplans(_ context.Context, _ *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	return make(resourceExporter.ResourceIDMetaMap), nil
}

func createCaseManagementStepplan(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getCaseManagementStepplanProxy(sdkConfig)

	caseplanID := d.Get("caseplan_id").(string)
	stageNumber := d.Get("stage_number").(int)

	stage, resp, err := case_management_stageplan.ResolveStageplanForCaseplanOrdinal(ctx, sdkConfig, caseplanID, stageNumber)
	if err != nil {
		return diag.FromErr(err)
	}
	_ = resp
	if stage == nil || stage.Id == nil {
		return diag.Errorf("could not resolve parent stageplan for caseplan %s stage_number %d", caseplanID, stageNumber)
	}

	step, resp2, err := ResolveSingleStepplanForStage(ctx, proxy, caseplanID, *stage.Id)
	if err != nil {
		return diag.FromErr(err)
	}
	_ = resp2
	if step == nil || step.Id == nil {
		return diag.Errorf("could not resolve stepplan for caseplan %s stageplan %s", caseplanID, *stage.Id)
	}

	d.SetId(formatStepplanResourceID(caseplanID, stageNumber, *step.Id))
	_ = d.Set("stepplan_id", *step.Id)
	_ = d.Set("stageplan_id", *stage.Id)

	if diagErr := applyStepplanPatchIfConfigured(ctx, d, meta, caseplanID, *stage.Id, *step.Id); diagErr != nil {
		return diagErr
	}
	return readCaseManagementStepplan(ctx, d, meta)
}

func applyStepplanPatchIfConfigured(ctx context.Context, d *schema.ResourceData, meta interface{}, caseplanID, stageplanID, stepplanID string) diag.Diagnostics {
	body := buildStepplanUpdate(d)
	if body.Name == nil && body.Description == nil && body.ActivityType == nil && body.WorkitemSettings == nil {
		return nil
	}
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getCaseManagementStepplanProxy(sdkConfig)
	_, resp, err := proxy.patchCaseManagementStepplan(ctx, caseplanID, stageplanID, stepplanID, body)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to patch case management stepplan: %s", err), resp)
	}
	return nil
}

func readCaseManagementStepplan(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getCaseManagementStepplanProxy(sdkConfig)
	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceCaseManagementStepplan(), constants.ConsistencyChecks(), resourceName)

	caseplanID, stageNumber, stepplanID, err := parseStepplanResourceID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	stage, _, err := case_management_stageplan.ResolveStageplanForCaseplanOrdinal(ctx, sdkConfig, caseplanID, stageNumber)
	if err != nil {
		return diag.FromErr(err)
	}
	if stage == nil || stage.Id == nil {
		return diag.Errorf("could not resolve parent stageplan for read")
	}
	stageplanID := *stage.Id

	log.Printf("Reading case management stepplan %s (caseplan=%s stage=%d)", stepplanID, caseplanID, stageNumber)

	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		stepplan, resp, getErr := proxy.getCaseManagementStepplan(ctx, caseplanID, stageplanID, stepplanID)
		if getErr != nil {
			if util.IsStatus404(resp) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read case management stepplan %s: %s", stepplanID, getErr), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read case management stepplan %s: %s", stepplanID, getErr), resp))
		}

		_ = d.Set("caseplan_id", caseplanID)
		_ = d.Set("stage_number", stageNumber)
		_ = d.Set("stepplan_id", stepplanID)
		_ = d.Set("stageplan_id", stageplanID)

		resourcedata.SetNillableValue(d, "name", stepplan.Name)
		resourcedata.SetNillableValue(d, "description", stepplan.Description)
		resourcedata.SetNillableValueWithInterfaceArrayWithFunc(d, "caseplan", stepplan.Caseplan, flattenCaseplanReference)
		resourcedata.SetNillableValueWithInterfaceArrayWithFunc(d, "stageplan", stepplan.Stageplan, flattenStageplanReference)
		resourcedata.SetNillableValue(d, "activity_type", stepplan.ActivityType)
		resourcedata.SetNillableValueWithInterfaceArrayWithFunc(d, "workitem_settings", stepplan.WorkitemSettings, flattenWorkitemSettingsResponse)

		log.Printf("Read case management stepplan %s", stepplanID)
		return cc.CheckState(d)
	})
}

func updateCaseManagementStepplan(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	caseplanID, stageNumber, stepplanID, err := parseStepplanResourceID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	stage, _, err := case_management_stageplan.ResolveStageplanForCaseplanOrdinal(ctx, sdkConfig, caseplanID, stageNumber)
	if err != nil {
		return diag.FromErr(err)
	}
	if stage == nil || stage.Id == nil {
		return diag.Errorf("could not resolve parent stageplan for update")
	}
	if diagErr := applyStepplanPatchIfConfigured(ctx, d, meta, caseplanID, *stage.Id, stepplanID); diagErr != nil {
		return diagErr
	}
	return readCaseManagementStepplan(ctx, d, meta)
}

func deleteCaseManagementStepplan(_ context.Context, _ *schema.ResourceData, _ interface{}) diag.Diagnostics {
	return nil
}
