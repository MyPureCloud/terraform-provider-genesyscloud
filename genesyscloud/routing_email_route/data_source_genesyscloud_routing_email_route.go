package routing_email_route

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

/*
   The data_source_genesyscloud_routing_email_route.go contains the data source implementation
   for the resource.
*/

// dataSourceRoutingEmailRouteRead retrieves by name the id in question
func dataSourceRoutingEmailRouteRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := newRoutingEmailRouteProxy(sdkConfig)

	name := d.Get("name").(string)

	return util.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		inboundRouteId, retryable, err := proxy.getRoutingEmailRouteIdByName(ctx, name)

		if err != nil && !retryable {
			return retry.NonRetryableError(fmt.Errorf("Error searching routing email route %s: %s", name, err))
		}

		if retryable {
			return retry.RetryableError(fmt.Errorf("No routing email route found with name %s", name))
		}

		d.SetId(inboundRouteId)
		return nil
	})
}
