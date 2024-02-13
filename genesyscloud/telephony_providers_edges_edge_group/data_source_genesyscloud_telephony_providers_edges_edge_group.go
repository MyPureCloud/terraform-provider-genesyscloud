package telephony_providers_edges_edge_group

import (
	"context"
	"fmt"
	"time"

	gcloud "terraform-provider-genesyscloud/genesyscloud"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceEdgeGroupRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sdkConfig := m.(*gcloud.ProviderMeta).ClientConfig
	edgeGroupProxy := getEdgeGroupProxy(sdkConfig)

	name := d.Get("name").(string)
	managed := d.Get("managed").(bool)

	return gcloud.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {

		edgeGroups, getErr := edgeGroupProxy.getEdgeGroupByName(ctx, name, managed)

		if getErr != nil {
			return retry.NonRetryableError(fmt.Errorf("Error requesting edge group %s: %s", name, getErr))
		}

		if edgeGroups == nil || len(*edgeGroups) == 0 {
			return retry.RetryableError(fmt.Errorf("No edge group found with name %s", name))
		}

		d.SetId(*(*edgeGroups)[0].Id)
		return nil
	})
}
