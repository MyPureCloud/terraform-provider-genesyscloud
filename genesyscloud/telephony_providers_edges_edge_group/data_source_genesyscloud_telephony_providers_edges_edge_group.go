package telephony_providers_edges_edge_group

import (
	"context"
	"fmt"
	"time"

	gcloud "terraform-provider-genesyscloud/genesyscloud"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v119/platformclientv2"
)

func dataSourceEdgeGroupRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sdkConfig := m.(*gcloud.ProviderMeta).ClientConfig
	edgesAPI := platformclientv2.NewTelephonyProvidersEdgeApiWithConfig(sdkConfig)

	name := d.Get("name").(string)
	managed := d.Get("managed").(bool)

	return gcloud.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		for pageNum := 1; ; pageNum++ {
			const pageSize = 100
			edgeGroup, _, getErr := edgesAPI.GetTelephonyProvidersEdgesEdgegroups(pageSize, pageNum, name, "", managed)

			if getErr != nil {
				return retry.NonRetryableError(fmt.Errorf("Error requesting edge group %s: %s", name, getErr))
			}

			if edgeGroup.Entities == nil || len(*edgeGroup.Entities) == 0 {
				return retry.RetryableError(fmt.Errorf("No edge group found with name %s", name))
			}

			d.SetId(*(*edgeGroup.Entities)[0].Id)
			return nil
		}
	})
}
