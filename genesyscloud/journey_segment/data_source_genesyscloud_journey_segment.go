package journey_segment

import (
	"context"
	"fmt"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// dataSourceJourneySegmentRead retrieves a Journey Segment by name from Genesys Cloud
//
// Parameters:
//   - ctx: The context.Context for the request
//   - d: The schema.ResourceData containing the resource configuration
//   - m: The provider meta interface containing client configuration
//
// Returns:
//   - diag.Diagnostics: Contains any error diagnostics encountered during the operation
//
// The function performs the following:
//  1. Extracts the client configuration from the provider meta
//  2. Creates a Journey Segment proxy
//  3. Retrieves the segment name from the resource data
//  4. Attempts to find the Journey Segment by name with retries
//  5. Sets the resource ID upon successful retrieval
func dataSourceJourneySegmentRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sdkConfig := m.(*provider.ProviderMeta).ClientConfig
	proxy := getJourneySegmentProxy(sdkConfig)

	name := d.Get("name").(string)

	return util.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		journeySegmentId, retryable, proxyResponse, err := proxy.getJourneySegmentIdByName(ctx, name)

		if err != nil && !retryable {
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Error searching Journey Segment %s | error: %s", name, err), proxyResponse))
		}

		if retryable {
			return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("No Journey Segment found with name %s", name), proxyResponse))
		}

		d.SetId(journeySegmentId)
		return nil
	})
}
