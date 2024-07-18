package supported_content

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
)

/*
   The data_source_genesyscloud_supported_content.go contains the data source implementation
   for the resource.
*/

// dataSourceSupportedContentRead retrieves by name the id in question
func dataSourceSupportedContentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getSupportedContentProxy(sdkConfig)

	name := d.Get("name").(string)

	return util.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		supportedContentId, retryable, resp, err := proxy.getSupportedContentIdByName(ctx, name)

		if err != nil && !retryable {
			retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(resourceName, fmt.Sprintf("Error searching supported content %s: %s", name, err), resp))
		}

		if retryable {
			retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(resourceName, fmt.Sprintf("No supported content found with name %s", name), resp))
		}

		d.SetId(supportedContentId)
		return nil
	})
}
