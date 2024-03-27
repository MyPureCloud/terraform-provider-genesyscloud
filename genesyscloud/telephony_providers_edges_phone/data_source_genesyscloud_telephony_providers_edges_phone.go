package telephony_providers_edges_phone

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

func dataSourcePhoneRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sdkConfig := m.(*provider.ProviderMeta).ClientConfig
	pp := getPhoneProxy(sdkConfig)

	name := d.Get("name").(string)

	return util.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		phone, retryable, resp, err := pp.getPhoneByName(ctx, name)
		if err != nil && !retryable {
			return retry.NonRetryableError(fmt.Errorf("error requesting phone %s: %s %v", name, err, resp))
		}

		if retryable {
			return retry.RetryableError(fmt.Errorf("no phone found with name %s", name))
		}

		d.SetId(*phone.Id)
		return nil
	})
}
