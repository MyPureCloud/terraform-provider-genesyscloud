package apple_integration

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
)

/*
   The data_source_genesyscloud_apple_integration.go contains the data source implementation
   for the resource.
*/

// dataSourceAppleIntegrationRead retrieves by name the id in question
func dataSourceAppleIntegrationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getAppleIntegrationProxy(sdkConfig)

	name := d.Get("name").(string)

	return util.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		appleIntegrationId, _, retryable, err := proxy.getAppleIntegrationIdByName(ctx, name)

		if err != nil && !retryable {
			return retry.NonRetryableError(fmt.Errorf("Error searching apple integration %s | error: %s", name, err))
		}

		if retryable {
			return retry.RetryableError(fmt.Errorf("No apple integration found with name %s", name))
		}

		d.SetId(appleIntegrationId)
		return nil
	})
}
