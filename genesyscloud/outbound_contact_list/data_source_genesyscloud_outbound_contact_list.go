package outbound_contact_list

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

func dataSourceOutboundContactListRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sdkConfig := m.(*provider.ProviderMeta).ClientConfig
	proxy := getOutboundContactlistProxy(sdkConfig)
	name := d.Get("name").(string)

	return util.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		contactListId, retryable, resp, err := proxy.getOutboundContactlistIdByName(ctx, name)
		if err != nil {
			diagErr := util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("error requesting contact list %s | error: %s", name, err), resp)
			if !retryable {
				return retry.NonRetryableError(diagErr)
			}
			return retry.RetryableError(diagErr)
		}

		d.SetId(contactListId)
		return nil
	})
}
