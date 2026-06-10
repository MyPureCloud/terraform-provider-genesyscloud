package case_management_caseplan

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
)

func createCaseManagementCaseplanCreateVersion(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getCaseManagementCaseplanProxy(sdkConfig)
	caseplanID := d.Get("caseplan_id").(string)

	log.Printf("Creating new caseplan draft version for %s (%s)", caseplanID, createVersionResourceLogName)
	_, resp, err := proxy.postCaseManagementCaseplanVersions(ctx, caseplanID)
	if err != nil {
		return util.BuildAPIDiagnosticError(CreateVersionResourceType, fmt.Sprintf("Failed to POST caseplan versions: %s", err), resp)
	}

	d.SetId(caseplanID)
	return readCaseManagementCaseplanCreateVersion(ctx, d, meta)
}

func readCaseManagementCaseplanCreateVersion(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getCaseManagementCaseplanProxy(sdkConfig)

	caseplanID := d.Get("caseplan_id").(string)
	if caseplanID == "" {
		caseplanID = d.Id()
	}
	if caseplanID == "" {
		return diag.Errorf("caseplan_id is required")
	}

	_, resp, err := proxy.getCaseManagementCaseplanById(ctx, caseplanID)
	if err != nil {
		if util.IsStatus404(resp) {
			return diag.Errorf("caseplan %s not found", caseplanID)
		}
		return util.BuildAPIDiagnosticError(CreateVersionResourceType, fmt.Sprintf("Failed to read caseplan: %s", err), resp)
	}

	_ = d.Set("caseplan_id", caseplanID)
	return nil
}

func updateCaseManagementCaseplanCreateVersion(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getCaseManagementCaseplanProxy(sdkConfig)
	caseplanID := d.Get("caseplan_id").(string)

	log.Printf("POST caseplan versions again for %s (%s update)", caseplanID, createVersionResourceLogName)
	_, resp, err := proxy.postCaseManagementCaseplanVersions(ctx, caseplanID)
	if err != nil {
		return util.BuildAPIDiagnosticError(CreateVersionResourceType, fmt.Sprintf("Failed to POST caseplan versions: %s", err), resp)
	}
	return readCaseManagementCaseplanCreateVersion(ctx, d, meta)
}

func deleteCaseManagementCaseplanCreateVersion(_ context.Context, d *schema.ResourceData, _ interface{}) diag.Diagnostics {
	log.Printf("Removing %s from state for caseplan %s (no delete-version API)", createVersionResourceLogName, d.Id())
	return nil
}
