package case_management_caseplan

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/resourcedata"
)

func createCaseManagementCaseplanPublish(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getCaseManagementCaseplanProxy(sdkConfig)
	caseplanID := d.Get("caseplan_id").(string)

	log.Printf("Publishing case management caseplan %s (%s)", caseplanID, publishResourceLogName)
	_, resp, err := proxy.publishCaseManagementCaseplan(ctx, caseplanID)
	if err != nil {
		return util.BuildAPIDiagnosticError(PublishResourceType, fmt.Sprintf("Failed to publish caseplan: %s", err), resp)
	}

	d.SetId(caseplanID)
	return readCaseManagementCaseplanPublish(ctx, d, meta)
}

func readCaseManagementCaseplanPublish(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getCaseManagementCaseplanProxy(sdkConfig)

	caseplanID := d.Get("caseplan_id").(string)
	if caseplanID == "" {
		caseplanID = d.Id()
	}
	if caseplanID == "" {
		return diag.Errorf("caseplan_id is required")
	}

	caseplan, resp, err := proxy.getCaseManagementCaseplanById(ctx, caseplanID)
	if err != nil {
		if util.IsStatus404(resp) {
			return diag.Errorf("caseplan %s not found", caseplanID)
		}
		return util.BuildAPIDiagnosticError(PublishResourceType, fmt.Sprintf("Failed to read caseplan: %s", err), resp)
	}

	_ = d.Set("caseplan_id", caseplanID)
	resourcedata.SetNillableValue(d, "published", caseplan.Published)
	resourcedata.SetNillableValue(d, "latest", caseplan.Latest)
	resourcedata.SetNillableValue(d, "version_state", caseplan.VersionState)
	return nil
}

func updateCaseManagementCaseplanPublish(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getCaseManagementCaseplanProxy(sdkConfig)
	caseplanID := d.Get("caseplan_id").(string)

	log.Printf("Re-publishing case management caseplan %s (%s update)", caseplanID, publishResourceLogName)
	_, resp, err := proxy.publishCaseManagementCaseplan(ctx, caseplanID)
	if err != nil {
		return util.BuildAPIDiagnosticError(PublishResourceType, fmt.Sprintf("Failed to publish caseplan: %s", err), resp)
	}
	return readCaseManagementCaseplanPublish(ctx, d, meta)
}

func deleteCaseManagementCaseplanPublish(_ context.Context, d *schema.ResourceData, _ interface{}) diag.Diagnostics {
	log.Printf("Removing %s from state for caseplan %s (no unpublish API)", publishResourceLogName, d.Id())
	return nil
}
