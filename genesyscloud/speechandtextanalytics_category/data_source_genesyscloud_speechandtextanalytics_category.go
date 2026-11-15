package speechandtextanalytics_category

import (
	"context"
	"fmt"
	"time"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// dataSourceCategoryRead retrieves a Genesys Cloud category by name
func dataSourceCategoryRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getCategoryProxy(sdkConfig)

	name := d.Get("name").(string)

	return util.WithRetries(ctx, 120*time.Second, func() *retry.RetryError {
		categoryId, retryable, resp, err := proxy.getCategoryIdByName(ctx, name)
		if err != nil {
			if retryable {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("No category found with name %s", name), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Error searching category %s | error: %s", name, err), resp))
		}
		d.SetId(categoryId)
		return nil
	})
}
