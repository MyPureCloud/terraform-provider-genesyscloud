package outbound_contactlistfilter

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"time"
)

/*
   The data_source_genesyscloud_outbound_contactlistfilter.go contains the data source implementation
   for the resource.
*/

// dataSourceOutboundContactlistfilterRead retrieves by name the id in question
func dataSourceOutboundContactlistfilterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getOutboundContactlistfilterProxy(sdkConfig)
	name := d.Get("name").(string)

	return util.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		contactListFilterId, retryable, resp, err := proxy.getOutboundContactlistfilterIdByName(ctx, name)

		if err != nil && !retryable {
			return retry.NonRetryableError(fmt.Errorf("error requesting contact list filter %s: %s %v", name, err, resp))
		}

		if retryable {
			return retry.RetryableError(fmt.Errorf("no contact list filters found with name %s", name))
		}

		d.SetId(contactListFilterId)
		return nil
	})
}
