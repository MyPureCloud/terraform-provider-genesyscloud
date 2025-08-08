package journey_segment

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/constants"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/consistency_checker"

	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/mypurecloud/platform-client-sdk-go/v165/platformclientv2"
)

// getAllJourneySegments retrieves all journey segments from the Genesys Cloud platform.
//
// Parameters:
//   - ctx: The context.Context for the request
//   - clientConfig: The Genesys Cloud platform client configuration
//
// Returns:
//   - resourceExporter.ResourceIDMetaMap: A map containing journey segment IDs as keys and ResourceMeta as values
//   - diag.Diagnostics: Any error diagnostics that occurred during the operation
//
// The function performs the following operations:
//  1. Initializes a journey segment proxy with the provided client configuration
//  2. Retrieves all journey segments using the proxy
//  3. For each journey segment:
//     - Uses DisplayName as BlockLabel if available
//     - Falls back to ID as BlockLabel if DisplayName is nil
//     - Skips segments with nil IDs
//  4. Returns the compiled resource map and any diagnostics
func getAllJourneySegments(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	proxy := getJourneySegmentProxy(clientConfig)
	resources := make(resourceExporter.ResourceIDMetaMap)

	segments, proxyResponse, getErr := proxy.getAllJourneySegments(ctx)
	if getErr != nil {
		return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get page of journey segments: %s", getErr), proxyResponse)
	}

	for _, segment := range *segments {
		if segment.Id == nil {
			continue // Skip if Id is nil as it's required for the map key
		}

		blockLabel := *segment.Id // Default to using Id as BlockLabel
		if segment.DisplayName != nil {
			blockLabel = *segment.DisplayName // Use DisplayName if available
		}

		resources[*segment.Id] = &resourceExporter.ResourceMeta{BlockLabel: blockLabel}
	}

	return resources, nil
}

// createJourneySegment creates a new journey segment in the Genesys Cloud platform.
//
// Parameters:
//   - ctx: The context.Context for the request
//   - d: The schema.ResourceData containing the journey segment configuration
//   - meta: The provider meta data containing client configuration
//
// Returns:
//   - diag.Diagnostics: Any error diagnostics that occurred during the operation
//
// The function performs the following operations:
//  1. Extracts client configuration from provider metadata
//  2. Builds journey segment object from schema data
//  3. Creates the journey segment via API proxy
//  4. Handles error cases with detailed error messages:
//     - Includes segment name in error if available
//     - Includes full input payload in error messages
//  5. Sets the resource ID with the created segment ID
//  6. Performs a final read to ensure state consistency
//
// Note: After successful creation, the function calls readJourneySegment to sync the Terraform state
func createJourneySegment(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getJourneySegmentProxy(sdkConfig)
	segment := buildSdkJourneySegment(d)

	log.Printf("Creating journey segment %s", *segment.DisplayName)
	segmentResponse, proxyResponse, err := proxy.createJourneySegment(ctx, segment)
	if err != nil {
		if segmentResponse != nil && segmentResponse.DisplayName != nil {
			input, _ := util.InterfaceToJson(*segment)
			return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("failed to create journey segment %s: %s\n(input: %+v)", *segmentResponse.DisplayName, err, input), proxyResponse)
		}
		input, _ := util.InterfaceToJson(*segment)
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("failed to create journey segment: %s\n(input: %+v)", err, input), proxyResponse)
	}

	d.SetId(*segmentResponse.Id)

	log.Printf("Created journey segment %s %s", *segmentResponse.DisplayName, *segmentResponse.Id)
	return readJourneySegment(ctx, d, meta)
}

// readJourneySegment retrieves an existing journey segment from the Genesys Cloud platform.
//
// Parameters:
//   - ctx: The context.Context for the request
//   - d: The schema.ResourceData containing the journey segment ID and state
//   - meta: The provider meta data containing client configuration
//
// Returns:
//   - diag.Diagnostics: Any error diagnostics that occurred during the operation
//
// The function performs the following operations:
//  1. Initializes client configuration and journey segment proxy
//  2. Creates a consistency checker for state validation
//  3. Attempts to read the journey segment with retries:
//     - Handles 404 errors as retryable for eventual consistency
//     - Handles other errors as non-retryable
//     - Validates segment response is not nil
//  4. Flattens the API response into schema data
//  5. Performs consistency check on the final state
//
// Note: Uses WithRetriesForRead for handling eventual consistency scenarios
func readJourneySegment(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getJourneySegmentProxy(sdkConfig)
	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceJourneySegment(), constants.ConsistencyChecks(), ResourceType)

	log.Printf("Reading journey segment %s", d.Id())
	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		segmentResponse, proxyResponse, getErr := proxy.getJourneySegmentById(ctx, d.Id())
		if getErr != nil {
			if util.IsStatus404(proxyResponse) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("failed to read journey segment %s | error: %s", d.Id(), getErr), proxyResponse))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("failed to read journey segment %s | error: %s", d.Id(), getErr), proxyResponse))
		}

		if segmentResponse == nil {
			return retry.NonRetryableError(fmt.Errorf("journey segment response is nil"))
		}

		flattenJourneySegment(d, segmentResponse)

		log.Printf("Read journey segment %s", d.Id())
		return cc.CheckState(d)
	})
}

// updateJourneySegment updates an existing journey segment in the Genesys Cloud platform.
//
// Parameters:
//   - ctx: The context.Context for the request
//   - d: The schema.ResourceData containing the journey segment configuration
//   - meta: The provider meta data containing client configuration
//
// Returns:
//   - diag.Diagnostics: Any error diagnostics that occurred during the operation
//
// The function performs the following operations:
//  1. Initializes client configuration and journey segment proxy
//  2. Builds patch segment object from schema data
//  3. Implements retry logic for version mismatch scenarios:
//     - Fetches current segment version
//     - Validates DisplayName is not nil
//     - Updates segment with current version
//  4. Handles error cases with detailed diagnostics:
//     - Includes segment details in error messages
//     - Captures full input payload for troubleshooting
//  5. Performs final read to ensure state consistency
//
// Note: Uses RetryWhen for handling version mismatch scenarios during concurrent updates
func updateJourneySegment(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getJourneySegmentProxy(sdkConfig)
	patchSegment := buildSdkPatchSegment(d)

	log.Printf("Updating journey segment %s", d.Id())
	diagErr := util.RetryWhen(util.IsVersionMismatch, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		// Get current journey segment version
		segmentResponse, proxyResponse, getErr := proxy.getJourneySegmentById(ctx, d.Id())
		if getErr != nil {
			return proxyResponse, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("failed to read journey segment %s error: %s", d.Id(), getErr), proxyResponse)
		}

		if patchSegment.DisplayName == nil {
			return proxyResponse, util.BuildAPIDiagnosticError(ResourceType, "DisplayName cannot be nil", proxyResponse)
		}

		patchSegment.Version = segmentResponse.Version
		_, proxyResponse, patchErr := proxy.updateJourneySegment(ctx, d.Id(), patchSegment)
		if patchErr != nil {
			input, _ := util.InterfaceToJson(*patchSegment)
			return proxyResponse, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Error updating journey segment %s: %s\n(input: %+v)", *patchSegment.DisplayName, patchErr, input), proxyResponse)
		}
		return proxyResponse, nil
	})
	if diagErr != nil {
		return diagErr
	}

	log.Printf("Updated journey segment %s", d.Id())
	return readJourneySegment(ctx, d, meta)
}

// deleteJourneySegment removes a journey segment from the Genesys Cloud platform.
//
// Parameters:
//   - ctx: The context.Context for the request
//   - d: The schema.ResourceData containing the journey segment ID and configuration
//   - meta: The provider meta data containing client configuration
//
// Returns:
//   - diag.Diagnostics: Any error diagnostics that occurred during the operation
//
// The function performs the following operations:
//  1. Initializes client configuration and journey segment proxy
//  2. Attempts to delete the journey segment using display name and ID
//  3. Implements verification retry logic with 30 second timeout:
//     - Polls for segment existence
//     - Considers 404 response as successful deletion
//     - Handles non-404 errors as non-retryable
//     - Continues retry if segment still exists
//  4. Logs deletion status for tracking
//
// Note: Uses WithRetries to ensure complete deletion and handle eventual consistency
func deleteJourneySegment(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getJourneySegmentProxy(sdkConfig)

	displayName := d.Get("display_name").(string)
	log.Printf("Deleting journey segment with display name %s", displayName)
	if proxyResponse, err := proxy.deleteJourneySegment(ctx, d.Id()); err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("failed to delete journey segment with display name %s error: %s", displayName, err), proxyResponse)
	}

	return util.WithRetries(ctx, 30*time.Second, func() *retry.RetryError {
		_, proxyResponse, err := proxy.getJourneySegmentById(ctx, d.Id())
		if err != nil {
			if util.IsStatus404(proxyResponse) {
				// journey segment deleted
				log.Printf("Deleted journey segment %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("error deleting journey segment %s | error: %s", d.Id(), err), proxyResponse))
		}
		return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("journey journey %s still exists", d.Id()), proxyResponse))
	})
}
