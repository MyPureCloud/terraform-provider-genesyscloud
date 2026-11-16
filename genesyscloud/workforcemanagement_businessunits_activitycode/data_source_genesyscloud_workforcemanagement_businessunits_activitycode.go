package workforcemanagement_businessunits_activitycode

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v200/platformclientv2"
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
	name := d.Get("name").(string)
	category := d.Get("category").(string)

	var activityCodeId string
	var resp *platformclientv2.APIResponse
	var err error
	if category != "" {
		activityCodeId, resp, err = proxy.getDefaultBusinessUnitActivityCodeIdByCategory(ctx, businessUnitId, category)
	} else {
		activityCodeId, resp, err = proxy.getBusinessUnitActivityCodeIdByName(ctx, businessUnitId, name)
	}
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Error searching business unit %s activity code | error: %s", businessUnitId, err), resp)
	}

	d.SetId(businessUnitId + "/" + activityCodeId)
	return nil
}
