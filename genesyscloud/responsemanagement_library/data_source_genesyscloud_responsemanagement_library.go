package responsemanagement_library

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

/*
   The data_source_genesyscloud_responsemanagement_library.go contains the data source implementation
   for the resource.
*/

// dataSourceResponsemanagementLibraryRead retrieves by name the id in question
func dataSourceResponsemanagementLibraryRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := newResponsemanagementLibraryProxy(sdkConfig)

	name := d.Get("name").(string)
	return util.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		libraryId, retryable, resp, err := proxy.getResponsemanagementLibraryIdByName(ctx, name)

		if err != nil && !retryable {
			return retry.NonRetryableError(fmt.Errorf("Error searching responsemanagement library %s: %s %v", name, err, resp))
		}

		if retryable {
			return retry.RetryableError(fmt.Errorf("No responsemanagement library found with name %s", name))
		}
		d.SetId(libraryId)
		return nil
	})
}
