package case_management_stageplan

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
)

func dataSourceCaseManagementStageplanRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getCaseManagementStageplanProxy(sdkConfig)

	caseplanID := d.Get("caseplan_id").(string)
	stageNumber := d.Get("stage_number").(int)

	stage, resp, err := resolveStageplanByOrdinal(ctx, proxy, caseplanID, stageNumber)
	if err != nil {
		return diag.FromErr(err)
	}
	if resp != nil {
		// non-fatal: resolveStageplanByOrdinal only sets resp from list calls
	}
	if stage == nil || stage.Id == nil {
		return diag.Errorf("could not resolve stageplan for caseplan %s stage_number %d", caseplanID, stageNumber)
	}

	d.SetId(formatStageplanResourceID(caseplanID, stageNumber, *stage.Id))
	return nil
}
