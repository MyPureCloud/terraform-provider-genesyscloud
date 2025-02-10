package journey_outcome

import (
	"context"
	"fmt"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v150/platformclientv2"
)

// dataSourceJourneyOutcomeRead retrieves a journey outcome by name from the Genesys Cloud Platform
// Parameters:
//   - ctx: The context.Context for managing timeouts and cancellation
//   - d: The schema.ResourceData containing the resource state
//   - m: The provider meta interface containing client configuration
//
// Returns:
//   - diag.Diagnostics: Any error diagnostics that occurred during the operation
//
// The function implements a retry mechanism with a 15-second timeout and performs the following:
//  1. Iterates through pages of journey outcomes (100 items per page)
//  2. Searches for an outcome matching the specified name
//  3. Sets the resource ID when a match is found
//  4. Returns an error if no matching outcome is found after checking all pages
func dataSourceJourneyOutcomeRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sdkConfig := m.(*provider.ProviderMeta).ClientConfig
	journeyApi := platformclientv2.NewJourneyApiWithConfig(sdkConfig)
	var response *platformclientv2.APIResponse

	name := d.Get("name").(string)

	return util.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		pageCount := 1 // Needed because of broken journey common paging
		for pageNum := 1; pageNum <= pageCount; pageNum++ {
			const pageSize = 100
			journeyOutcomes, resp, getErr := journeyApi.GetJourneyOutcomes(pageNum, pageSize, "", nil, nil, "")
			if getErr != nil {
				return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("failed to get page of journey outcomes: %v", getErr), resp))
			}

			response = resp

			if journeyOutcomes.Entities == nil || len(*journeyOutcomes.Entities) == 0 {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("no journey outcome found with name %s", name), resp))
			}

			for _, journeyOutcome := range *journeyOutcomes.Entities {
				if journeyOutcome.DisplayName != nil && *journeyOutcome.DisplayName == name {
					if journeyOutcome.Id == nil {
						return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, "journey outcome ID is nil", resp))
					}
					d.SetId(*journeyOutcome.Id)
					return nil
				}
			}

			if journeyOutcomes.PageCount == nil {
				return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, "journey outcomes page count is nil", resp))
			}
			pageCount = *journeyOutcomes.PageCount
		}
		return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("no journey outcome found with name %s", name), response))
	})
}
