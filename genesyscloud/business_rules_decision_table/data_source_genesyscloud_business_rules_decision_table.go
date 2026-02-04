package business_rules_decision_table

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
)

// dataSourceBusinessRulesDecisionTableRead reads a Genesys Cloud business rules decision table by name
func dataSourceBusinessRulesDecisionTableRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getBusinessRulesDecisionTableProxy(sdkConfig)

	name := d.Get("name").(string)

	// Query for business rules decision tables by name. Retry in case new table is not yet indexed by search.
	// As table names are non-unique, fail in case of multiple results.
	return util.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		tables, retryable, resp, err := proxy.getBusinessRulesDecisionTablesByName(ctx, name)
		if err != nil && !retryable {
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("error getting business rules decision table %s | error: %v", name, err), resp))
		}

		if retryable {
			return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("no business rules decision table found with name %s", name), resp))
		}

		if len(*tables) > 1 {
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("ambiguous business rules decision table name: %s", name), resp))
		}

		table := (*tables)[0]
		d.SetId(*table.Id)
		// Set the published version if available (search API includes this when withPublishedVersion=true)
		if table.Published != nil && table.Published.Version != nil {
			if err := d.Set("version", *table.Published.Version); err != nil {
				return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("error setting version for decision table %s | error: %v", name, resp), resp))
			}
		}

		log.Printf("Successfully read Business Rules Decision Table by name: %s", name)
		return nil
	})
}
