package genesyscloud

import (
	"context"
	"log"
	gcloud "terraform-provider-genesyscloud/genesyscloud"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceRoutingSmsAddressRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	smsAddressProxy := getRoutingSmsAddressProxy(sdkConfig)
	name := d.Get("name").(string)

	log.Printf("Searching for routing sms address with name '%s'", name)
	return gcloud.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		smsAddressId, retryable, err := smsAddressProxy.getSmsAddressIdByName(name, ctx)
		if err != nil && !retryable {
			return retry.NonRetryableError(err)
		}
		if retryable {
			return retry.RetryableError(err)
		}
		d.SetId(smsAddressId)
		return nil
	})
}
