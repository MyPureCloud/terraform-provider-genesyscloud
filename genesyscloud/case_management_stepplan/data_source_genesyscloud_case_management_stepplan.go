package case_management_stepplan

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/case_management_stageplan"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
)

func dataSourceCaseManagementStepplanRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getCaseManagementStepplanProxy(sdkConfig)

	caseplanID := d.Get("caseplan_id").(string)
	stageNumber := d.Get("stage_number").(int)

	stage, _, err := case_management_stageplan.ResolveStageplanForCaseplanOrdinal(ctx, sdkConfig, caseplanID, stageNumber)
	if err != nil {
		return diag.FromErr(err)
	}
	if stage == nil || stage.Id == nil {
		return diag.Errorf("could not resolve stageplan for caseplan %s stage_number %d", caseplanID, stageNumber)
	}

	step, _, err := ResolveSingleStepplanForStage(ctx, proxy, caseplanID, *stage.Id)
	if err != nil {
		return diag.FromErr(err)
	}
	if step == nil || step.Id == nil {
		return diag.Errorf("could not resolve stepplan for caseplan %s stageplan %s", caseplanID, *stage.Id)
	}

	d.SetId(formatStepplanResourceID(caseplanID, stageNumber, *step.Id))
	return nil
}
