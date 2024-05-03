package telephony_providers_edges_did_pool

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

// dataSourceDidPoolRead retrieves the did pool id using the start and end number
func dataSourceDidPoolRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sdkConfig := m.(*provider.ProviderMeta).ClientConfig
	proxy := getTelephonyDidPoolProxy(sdkConfig)

	didPoolStartPhoneNumber := d.Get("start_phone_number").(string)
	didPoolEndPhoneNumber := d.Get("end_phone_number").(string)

	return util.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		id, retryable, resp, err := proxy.getTelephonyDidPoolIdByStartAndEndNumber(ctx, didPoolStartPhoneNumber, didPoolEndPhoneNumber)
		if err != nil && !retryable {
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(resourceName, fmt.Sprintf("Failed to get telephony DID pool %s", err), resp))
		}
		if retryable {
			return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(resourceName, fmt.Sprintf("Failed to get telephony DID pool %s", err), resp))
		}
		d.SetId(id)
		return nil
	})
}
