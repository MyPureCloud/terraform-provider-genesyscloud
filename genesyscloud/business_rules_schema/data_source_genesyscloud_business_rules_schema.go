package business_rules_schema

import (
	"context"
	"fmt"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

/*
   The data_source_genesyscloud_business_rules_schema.go contains the data source implementation
   for the resource.
*/

// dataSourceBusinessRulesSchemaRead retrieves by name the id in question
func dataSourceBusinessRulesSchemaRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getBusinessRulesSchemaProxy(sdkConfig)

	name := d.Get("name").(string)

	// Query for business rules schemas by name. Retry in case new schema is not yet indexed by search.
	// As schema names are non-unique, fail in case of multiple results.
	return util.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		schemas, retryable, resp, err := proxy.getBusinessRulesSchemasByName(ctx, name)
		if err != nil && !retryable {
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("error getting business rules schema %s | error: %v", name, err), resp))
		}

		if retryable {
			return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("no business rules schema found with name %s", name), resp))
		}

		if len(*schemas) > 1 {
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("ambiguous business rules schema name: %s", name), resp))
		}

		schema := (*schemas)[0]
		d.SetId(*schema.Id)
		return nil
	})
}
