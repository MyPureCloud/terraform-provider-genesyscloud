package telephony_providers_edges_edge_group

import (
	"context"
	"fmt"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceEdgeGroupRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sdkConfig := m.(*provider.ProviderMeta).ClientConfig
	edgeGroupProxy := getEdgeGroupProxy(sdkConfig)

	name := d.Get("name").(string)
	managed := d.Get("managed").(bool)

	return util.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {

		edgeGroup, retryable, getErr := edgeGroupProxy.getEdgeGroupByName(ctx, name, managed)

		if getErr != nil && !retryable {
			return retry.NonRetryableError(fmt.Errorf("Error requesting edge group %s: %s", name, getErr))
		}

		if retryable {
			return retry.RetryableError(fmt.Errorf("No edge group found with name %s", name))
		}

		d.SetId(edgeGroup)
		return nil
	})
}
