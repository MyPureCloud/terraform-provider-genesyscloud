package genesyscloud

import (
	"context"
	"fmt"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceRoutingSmsAddressRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	smsAddressProxy := getRoutingSmsAddressProxy(sdkConfig)
	name := d.Get("name").(string)

	log.Printf("Searching for routing sms address with name '%s'", name)
	return util.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		smsAddressId, retryable, resp, err := smsAddressProxy.getSmsAddressIdByName(name, ctx)
		if err != nil && !retryable {
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to get SMS Address | error: %s", err), resp))
		}
		if retryable {
			return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to get SMS Address | error: %s", err), resp))
		}
		d.SetId(smsAddressId)
		return nil
	})
}
