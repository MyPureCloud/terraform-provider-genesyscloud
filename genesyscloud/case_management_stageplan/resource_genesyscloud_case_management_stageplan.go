package case_management_stageplan

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v186/platformclientv2"

	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/consistency_checker"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/constants"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/resourcedata"
)

// getAllAuthCaseManagementStageplans — no org-wide list API; exporter returns nothing for this type.
func getAllAuthCaseManagementStageplans(_ context.Context, _ *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	return make(resourceExporter.ResourceIDMetaMap), nil
}

func createCaseManagementStageplan(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getCaseManagementStageplanProxy(sdkConfig)

	caseplanID := d.Get("caseplan_id").(string)
	stageNumber := d.Get("stage_number").(int)

	stage, _, err := resolveStageplanByOrdinal(ctx, proxy, caseplanID, stageNumber)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(formatStageplanResourceID(caseplanID, stageNumber, *stage.Id))
	if diagErr := d.Set("stageplan_id", *stage.Id); diagErr != nil {
		return diag.FromErr(diagErr)
	}

	if diagErr := applyStageplanPatchIfConfigured(ctx, d, meta, caseplanID, *stage.Id); diagErr != nil {
		return diagErr
	}
	return readCaseManagementStageplan(ctx, d, meta)
}

func applyStageplanPatchIfConfigured(ctx context.Context, d *schema.ResourceData, meta interface{}, caseplanID, stageplanID string) diag.Diagnostics {
	body := buildStageplanUpdate(d)
	if body.Name == nil && body.Description == nil {
		return nil
	}
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getCaseManagementStageplanProxy(sdkConfig)
	_, resp, err := proxy.patchCaseManagementStageplan(ctx, caseplanID, stageplanID, body)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to patch case management stageplan: %s", err), resp)
	}
	return nil
}

func readCaseManagementStageplan(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getCaseManagementStageplanProxy(sdkConfig)
	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceCaseManagementStageplan(), constants.ConsistencyChecks(), resourceName)

	caseplanID, stageNumber, stageplanID, err := parseStageplanResourceID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("Reading case management stageplan %s (caseplan=%s stage=%d)", stageplanID, caseplanID, stageNumber)

	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		stageplan, resp, getErr := proxy.getCaseManagementStageplan(ctx, caseplanID, stageplanID)
		if getErr != nil {
			if util.IsStatus404(resp) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read case management stageplan %s: %s", stageplanID, getErr), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read case management stageplan %s: %s", stageplanID, getErr), resp))
		}

		_ = d.Set("caseplan_id", caseplanID)
		_ = d.Set("stage_number", stageNumber)
		_ = d.Set("stageplan_id", stageplanID)

		resourcedata.SetNillableValue(d, "name", stageplan.Name)
		resourcedata.SetNillableValue(d, "description", stageplan.Description)
		resourcedata.SetNillableValueWithInterfaceArrayWithFunc(d, "caseplan", stageplan.Caseplan, flattenCaseplanReference)

		log.Printf("Read case management stageplan %s", stageplanID)
		return cc.CheckState(d)
	})
}

func updateCaseManagementStageplan(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	caseplanID, _, stageplanID, err := parseStageplanResourceID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	if diagErr := applyStageplanPatchIfConfigured(ctx, d, meta, caseplanID, stageplanID); diagErr != nil {
		return diagErr
	}
	return readCaseManagementStageplan(ctx, d, meta)
}

func deleteCaseManagementStageplan(_ context.Context, _ *schema.ResourceData, _ interface{}) diag.Diagnostics {
	return nil
}
