package telephony_providers_edges_did

import (
	"context"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	gcloud "terraform-provider-genesyscloud/genesyscloud/util"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// dataSourceDidRead retrieves by DID ID by DID number
func dataSourceDidRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sdkConfig := m.(*provider.ProviderMeta).ClientConfig
	proxy := getTelephonyProvidersEdgesDidProxy(sdkConfig)

	didPhoneNumber := d.Get("phone_number").(string)

	return gcloud.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		id, retryable, err := proxy.getTelephonyProvidersEdgesDidIdByDid(ctx, didPhoneNumber)
		if err != nil && !retryable {
			return retry.NonRetryableError(err)
		}
		if retryable {
			return retry.RetryableError(err)
		}
		d.SetId(id)
		return nil
	})
}
