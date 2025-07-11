package routing_wrapupcode

import (
	"context"
	"fmt"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceRoutingWrapupcodeRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	proxy := getRoutingWrapupcodeProxy(m.(*provider.ProviderMeta).ClientConfig)
	name := d.Get("name").(string)

	return util.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		wrapupcodeId, retryable, proxyResponse, err := proxy.getRoutingWrapupcodeIdByName(ctx, name)

		if err != nil {
			diagErr := util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Error requesting wrap-up code %s | error: %s", name, err), proxyResponse)
			if !retryable {
				return retry.NonRetryableError(diagErr)
			}
			return retry.RetryableError(diagErr)
		}

		d.SetId(wrapupcodeId)
		return nil
	})
}
