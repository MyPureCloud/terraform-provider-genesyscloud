package responsemanagement_response

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

/*
   The data_source_genesyscloud_responsemanagement_response.go contains the data source implementation
   for the resource.
*/

// dataSourceResponsemanagementResponseRead retrieves by name the id in question
func dataSourceResponsemanagementResponseRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := newResponsemanagementResponseProxy(sdkConfig)

	name := d.Get("name").(string)
	library := d.Get("library_id").(string)

	return util.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		managementResponseId, retryable, resp, err := proxy.getResponsemanagementResponseIdByName(ctx, name, library)

		if err != nil && !retryable {
			return retry.NonRetryableError(fmt.Errorf("error requesting responsemanagement response %s: %s %v", name, err, resp))
		}

		if retryable {
			return retry.RetryableError(fmt.Errorf("no responsemanagement response found with name %s: %s", name, err))
		}

		d.SetId(managementResponseId)
		return nil
	})
}
