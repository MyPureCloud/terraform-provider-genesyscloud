package responsemanagement_responseasset

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v121/platformclientv2"
	gcloud "terraform-provider-genesyscloud/genesyscloud"
	"time"
)

func dataSourceResponseManagementResponseAssetRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var (
		name    = d.Get("name").(string)
		field   = "name"
		fields  = []string{field}
		varType = "TERM"
		filter  = platformclientv2.Responseassetfilter{
			Fields:  &fields,
			Value:   &name,
			VarType: &varType,
		}
		body = platformclientv2.Responseassetsearchrequest{
			Query:  &[]platformclientv2.Responseassetfilter{filter},
			SortBy: &field,
		}
	)

	sdkConfig := m.(*gcloud.ProviderMeta).ClientConfig
	proxy := getRespManagementRespAssetProxy(sdkConfig)

	return gcloud.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		responseData, _, getErr := proxy.responseManagementApi.PostResponsemanagementResponseassetsSearch(body, nil)
		if getErr != nil {
			return retry.NonRetryableError(fmt.Errorf("Error requesting response asset %s: %s", name, getErr))
		}
		if responseData.Results == nil || len(*responseData.Results) == 0 {
			return retry.RetryableError(fmt.Errorf("No response asset found with name %s", name))
		}
		asset := (*responseData.Results)[0]
		d.SetId(*asset.Id)
		return nil
	})
}
