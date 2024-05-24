package authorization_product

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
)

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

func dataSourceAuthorizationProductRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getauthProductProxy(sdkConfig)
	name := d.Get("name").(string)

	return util.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		// Get the list of enabled products
		authProductId, retryable, resp, err := proxy.getAuthorizationProduct(ctx, name)

		if err != nil {
			if retryable {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(resourceName, fmt.Sprintf("Failed to get Authorization product %s | error: %s", authProductId, err), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(resourceName, fmt.Sprintf("Failed to get Authorization product %s | error: %s", authProductId, err), resp))
		}

		d.SetId(authProductId)
		return nil
	})
}

func GenerateAuthorizationProductDataSource(id, productName, dependsOn string) string {
	return fmt.Sprintf(`
data "genesyscloud_authorization_product" "%s" {
	name = "%s"
	depends_on=[%s]
}
`, id, productName, dependsOn)
}
