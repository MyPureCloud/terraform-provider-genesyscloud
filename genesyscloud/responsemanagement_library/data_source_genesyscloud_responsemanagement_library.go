package responsemanagement_library

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	gcloud "terraform-provider-genesyscloud/genesyscloud"
)

/*
   The data_source_genesyscloud_responsemanagement_library.go contains the data source implementation
   for the resource.
*/

// dataSourceResponsemanagementLibraryRead retrieves by name the id in question
func dataSourceResponsemanagementLibraryRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	proxy := newResponsemanagementLibraryProxy(sdkConfig)

	name := d.Get("name").(string)
	return gcloud.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		libraryId, retryable, err := proxy.getResponsemanagementLibraryIdByName(ctx, name)

		if err != nil && !retryable {
			return retry.NonRetryableError(fmt.Errorf("Error searching responsemanagement library %s: %s", name, err))
		}

		if retryable {
			return retry.RetryableError(fmt.Errorf("No responsemanagement library found with name %s", name))
		}
		d.SetId(libraryId)
		return nil
	})
}
