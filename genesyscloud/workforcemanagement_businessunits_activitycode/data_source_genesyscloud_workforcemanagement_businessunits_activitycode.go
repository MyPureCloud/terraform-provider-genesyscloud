package workforcemanagement_businessunits_activitycode

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
)

/*
   The data_source_genesyscloud_workforcemanagement_businessunits_activitycode.go contains the data source implementation
   for the resource.
*/

// dataSourceBusinessUnitActivityCodeRead retrieves the id of an activity code by name, or the id of the default activity code of a category
func dataSourceBusinessUnitActivityCodeRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getBusinessUnitActivityCodeProxy(sdkConfig)
	businessUnitId := d.Get("business_unit_id").(string)

	if category, ok := d.GetOk("category"); ok {
		return readDefaultActivityCodeByCategory(ctx, d, proxy, businessUnitId, category.(string))
	}
	return readActivityCodeByName(ctx, d, proxy, businessUnitId, d.Get("name").(string))
}

// readActivityCodeByName looks up an activity code by its name
func readActivityCodeByName(ctx context.Context, d *schema.ResourceData, proxy *businessUnitActivityCodeProxy, businessUnitId string, name string) diag.Diagnostics {
	activityCodeId, resp, err := proxy.getBusinessUnitActivityCodeIdByName(ctx, businessUnitId, name)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Error searching business unit %s for the activity code named %s | error: %s", businessUnitId, name, err), resp)
	}

	d.SetId(businessUnitId + "/" + activityCodeId)
	return nil
}

// readDefaultActivityCodeByCategory looks up the default activity code of a category
func readDefaultActivityCodeByCategory(ctx context.Context, d *schema.ResourceData, proxy *businessUnitActivityCodeProxy, businessUnitId string, category string) diag.Diagnostics {
	activityCodeId, resp, err := proxy.getDefaultBusinessUnitActivityCodeIdByCategory(ctx, businessUnitId, category)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Error searching business unit %s for the default activity code of category %s | error: %s", businessUnitId, category, err), resp)
	}

	d.SetId(businessUnitId + "/" + activityCodeId)
	return nil
}
