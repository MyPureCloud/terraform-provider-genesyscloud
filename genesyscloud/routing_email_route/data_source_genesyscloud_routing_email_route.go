package routing_email_route

import (
	"context"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

/*
   The data_source_genesyscloud_routing_email_route.go contains the data source implementation
   for the resource.
*/

// dataSourceRoutingEmailRouteRead retrieves by pattern, domainId the id in question
func dataSourceRoutingEmailRouteRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sdkConfig := m.(*provider.ProviderMeta).ClientConfig
	proxy := getRoutingEmailRouteProxy(sdkConfig)

	pattern := d.Get("pattern").(string)
	domainId := d.Get("domain_id").(string)

	return util.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		responseId, retryable, resp, err := proxy.getRoutingEmailRouteIdByPattern(ctx, pattern, domainId)

		if err != nil {
			if !retryable {
				return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, err.Error(), resp))
			}
			return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, err.Error(), resp))
		}

		d.SetId(responseId)
		return nil
	})
}
