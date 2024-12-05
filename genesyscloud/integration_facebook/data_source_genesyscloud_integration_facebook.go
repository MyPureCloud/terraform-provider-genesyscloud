package integration_facebook

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
   The data_source_genesyscloud_integration_facebook.go contains the data source implementation
   for the resource.
*/

// dataSourceIntegrationFacebookRead retrieves by name the id in question
func dataSourceIntegrationFacebookRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getIntegrationFacebookProxy(sdkConfig)

	name := d.Get("name").(string)

	return util.WithRetries(ctx, 30*time.Second, func() *retry.RetryError {
		facebookIntegrationRequestId, retryable, resp, err := proxy.getIntegrationFacebookIdByName(ctx, name)

		if err != nil && !retryable {
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Error searching integration facebook %s: %s", name, err), resp))
		}

		if retryable {
			return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("No integration facebook found with name %s", name), resp))
		}

		d.SetId(facebookIntegrationRequestId)
		return nil
	})
}
