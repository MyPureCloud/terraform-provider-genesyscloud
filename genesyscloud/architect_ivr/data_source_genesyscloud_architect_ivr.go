package architect_ivr

import (
	"context"
	"fmt"
	gcloud "terraform-provider-genesyscloud/genesyscloud"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// dataSourceIvrRead retrieves the Genesys Cloud architect ivr id by name
func dataSourceIvrRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sdkConfig := m.(*gcloud.ProviderMeta).ClientConfig
	ap := getArchitectIvrProxy(sdkConfig)
	name := d.Get("name").(string)
	// Query ivr by name. Retry in case search has not yet indexed the ivr.
	return gcloud.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		id, retryable, err := ap.getArchitectIvrIdByName(ctx, name)
		if err != nil && !retryable {
			return retry.NonRetryableError(fmt.Errorf("error requesting IVR %s: %s", name, err))
		}
		if retryable {
			return retry.RetryableError(err)
		}
		d.SetId(id)
		return nil
	})
}
