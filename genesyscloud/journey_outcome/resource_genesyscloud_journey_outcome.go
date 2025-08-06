package journey_outcome

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

// getAllJourneyOutcomes retrieves all journey outcomes from the Genesys Cloud platform.
//
// Parameters:
//   - ctx: The context.Context for the request, used for cancellation and timeouts
//   - clientConfig: The platformclientv2.Configuration containing client configuration settings
//
// Returns:
//   - resourceExporter.ResourceIDMetaMap: A map containing journey outcome IDs as keys and ResourceMeta as values
//   - diag.Diagnostics: Any error diagnostics that occurred during the operation
//
// The function performs the following operations:
//  1. Creates a proxy to interact with journey outcomes
//  2. Retrieves all journey outcomes using the proxy
//  3. Validates the response for nil values
//  4. Processes each outcome, skipping entries with nil ID or DisplayName
//  5. Creates a map of resource metadata using outcome IDs and display names
//
// Error Handling:
//   - Returns diagnostic error if the API call fails
//   - Returns diagnostic error if the outcomes response is nil
//   - Logs warnings for outcomes with nil ID or DisplayName and skips them
func getAllJourneyOutcomes(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	proxy := getJourneyOutcomeProxy(clientConfig)
	resources := make(resourceExporter.ResourceIDMetaMap)

	outComes, proxyResponse, getErr := proxy.getAllJourneyOutcomes(ctx)
	if getErr != nil {
		return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get page of journey outcomes: %s", getErr), proxyResponse)
	}

	// Check if outComes is nil
	if outComes == nil {
		return nil, util.BuildAPIDiagnosticError(ResourceType, "Received nil for outComes from API", proxyResponse)
	}

	for _, outCome := range *outComes {
		// Skip invalid entries where Id or DisplayName is nil
		if outCome.Id == nil {
			log.Printf("Warning: Skipping outComes with nil ID")
			continue
		}
		if outCome.DisplayName == nil {
			log.Printf("Warning: Skipping outComes %s with nil DisplayName", *outCome.Id)
			continue
		}
		resources[*outCome.Id] = &resourceExporter.ResourceMeta{BlockLabel: *outCome.DisplayName}
	}

	return resources, nil
}

// createJourneyOutcome creates a new journey outcome in Genesys Cloud.
//
// Parameters:
//   - ctx: The context.Context for managing the request lifecycle
//   - d: *schema.ResourceData containing the resource configuration data
//   - meta: interface{} containing the provider metadata, specifically the ProviderMeta with ClientConfig
//
// Returns:
//   - diag.Diagnostics: Contains any error diagnostics encountered during creation
//
// The function performs the following operations:
//  1. Extracts the SDK configuration from the provider metadata
//  2. Gets a proxy instance for journey outcome operations
//  3. Builds the journey outcome object from the resource data
//  4. Creates the journey outcome via the API
//  5. Handles various error cases and nil checks
//  6. Sets the resource ID upon successful creation
//  7. Performs a final read of the created resource
//
// Error Handling:
//   - Returns diagnostic error if journey outcome creation fails
//   - Performs nil checks on critical response objects
//   - Safely handles potential nil pointers in response data
//   - Includes detailed error messages with input data for debugging
//
// Example Usage:
//
//	This function is called by the Terraform provider when creating a new journey outcome resource:
//	```hcl
//	resource "genesyscloud_journey_outcome" "example" {
//	  display_name = "Example Outcome"
//	  description  = "An example journey outcome"
//	  is_active    = true
//	}
//	```
func createJourneyOutcome(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getJourneyOutcomeProxy(sdkConfig)
	journeyOutcome := buildSdkJourneyOutcome(d)

	log.Printf("Creating journey outcome %s", *journeyOutcome.DisplayName)

	outComeResponse, proxyResponse, err := proxy.createJourneyOutcome(ctx, journeyOutcome)
	if err != nil {
		// First check if journeyOutcome is nil
		var input interface{}
		if journeyOutcome != nil {
			input, _ = util.InterfaceToJson(*journeyOutcome)
		}

		// Build error message safely without dereferencing nil pointers
		errMsg := "failed to create journey outcome"
		if outComeResponse != nil && outComeResponse.DisplayName != nil {
			errMsg = fmt.Sprintf("failed to create journey outcome %s", *outComeResponse.DisplayName)
		}

		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("%s: %s\n(input: %+v)", errMsg, err, input), proxyResponse)
	}

	// Verify outComeResponse is not nil before proceeding
	if outComeResponse == nil {
		return util.BuildAPIDiagnosticError(ResourceType, "received nil response when creating journey outcome", proxyResponse)
	}

	// Verify required fields are not nil
	if outComeResponse.Id == nil {
		return util.BuildAPIDiagnosticError(ResourceType, "received response with nil ID when creating journey outcome", proxyResponse)
	}

	d.SetId(*outComeResponse.Id)

	log.Printf("Created journey outcome %s %s",
		GetStringValue(outComeResponse.DisplayName, "unknown"), // Safely handle potentially nil DisplayName
		*outComeResponse.Id,
	)

	return readJourneyOutcome(ctx, d, meta)
}

// readJourneyOutcome retrieves and reads a journey outcome from Genesys Cloud.
//
// Parameters:
//   - ctx: context.Context for managing the request lifecycle and cancellation
//   - d: *schema.ResourceData containing the current state of the resource
//   - meta: interface{} containing the provider metadata, specifically the ProviderMeta with ClientConfig
//
// Returns:
//   - diag.Diagnostics: Contains any error diagnostics encountered during the read operation
//
// The function performs the following operations:
//  1. Retrieves SDK configuration from provider metadata
//  2. Creates a journey outcome proxy for API operations
//  3. Initializes consistency checker for state validation
//  4. Attempts to read the journey outcome by ID with retries
//  5. Performs nil checks on the response and required fields
//  6. Flattens the response data into the resource schema
//  7. Validates final state with consistency checker
//
// Error Handling:
//   - Returns retryable error for 404 (Not Found) responses
//   - Returns non-retryable error for other API failures
//   - Validates response object and required fields for nil values
//   - Includes detailed error messages for debugging
//
// Example Usage in Terraform:
//
//	```hcl
//	data "genesyscloud_journey_outcome" "example" {
//	  id = "existing-outcome-id"
//	}
//	```
func readJourneyOutcome(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getJourneyOutcomeProxy(sdkConfig)
	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceJourneyOutcome(), constants.ConsistencyChecks(), ResourceType)

	log.Printf("Reading journey outcome %s", d.Id())
	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		outComeResponse, proxyResponse, getErr := proxy.getJourneyOutcomeById(ctx, d.Id())
		if getErr != nil {
			if util.IsStatus404(proxyResponse) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("failed to read journey outcome %s | error: %s", d.Id(), getErr), proxyResponse))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("failed to read journey outcome %s | error: %s", d.Id(), getErr), proxyResponse))
		}

		// Add nil check for outComeResponse
		if outComeResponse == nil {
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("received nil response for journey outcome %s", d.Id()), proxyResponse))
		}

		// Add nil checks for required fields
		if outComeResponse.DisplayName == nil {
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("journey outcome %s has nil DisplayName", d.Id()), proxyResponse))
		}

		if outComeResponse.IsActive == nil {
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("journey outcome %s has nil IsActive", d.Id()), proxyResponse))
		}

		// Now safe to flatten the response - we know outComeResponse and required fields are not nil
		flattenJourneyOutcome(d, outComeResponse)

		log.Printf("Read journey outcome %s %s", d.Id(), *outComeResponse.DisplayName)
		return cc.CheckState(d)
	})
}

// updateJourneyOutcome updates an existing journey outcome in Genesys Cloud.
//
// Parameters:
//   - ctx: context.Context for managing the request lifecycle and cancellation
//   - d: *schema.ResourceData containing the resource configuration and state
//   - meta: interface{} containing the provider metadata with ClientConfig
//
// Returns:
//   - diag.Diagnostics: Contains any error diagnostics encountered during the update operation
//
// The function performs the following operations:
//  1. Gets SDK configuration from provider metadata
//  2. Creates journey outcome proxy for API operations
//  3. Builds patch outcome object from resource data
//  4. Retrieves current journey outcome version
//  5. Updates the journey outcome with version handling
//  6. Performs final read of updated resource
//
// Error Handling:
//   - Implements version mismatch retry logic
//   - Validates nil responses and objects
//   - Checks for nil Version field
//   - Validates patch outcome object
//   - Includes detailed error messages with input data
//
// Version Control:
//   - Fetches current version before update
//   - Handles version conflicts with RetryWhen mechanism
//   - Ensures atomic updates
//
// Nil Checks:
//   - Validates outComeResponse is not nil
//   - Ensures Version field is not nil
//   - Validates patchOutcome object
//   - Checks DisplayName for logging purposes
//
// Example Usage in Terraform:
//
//	```hcl
//	resource "genesyscloud_journey_outcome" "example" {
//	  display_name = "Updated Outcome"
//	  description  = "Updated journey outcome description"
//	  is_active    = true
//	}
//	```
func updateJourneyOutcome(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getJourneyOutcomeProxy(sdkConfig)
	patchOutcome := buildSdkPatchOutcome(d)

	log.Printf("Updating journey outcome %s", d.Id())
	diagErr := util.RetryWhen(util.IsVersionMismatch, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		// Get current journey outcome version
		outComeResponse, proxyResponse, getErr := proxy.getJourneyOutcomeById(ctx, d.Id())
		if getErr != nil {
			return proxyResponse, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("failed to read journey outcome %s error: %s", d.Id(), getErr), proxyResponse)
		}

		// Check if outComeResponse is nil
		if outComeResponse == nil {
			return proxyResponse, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("received nil response when reading journey outcome %s", d.Id()), proxyResponse)
		}

		// Check if Version is nil
		if outComeResponse.Version == nil {
			return proxyResponse, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("journey outcome %s has nil Version", d.Id()), proxyResponse)
		}

		// Check if patchOutcome is valid before any operations
		if patchOutcome == nil {
			return proxyResponse, util.BuildAPIDiagnosticError(ResourceType, "patchOutcome is nil", proxyResponse)
		}

		// Assign version after all nil checks
		patchOutcome.Version = outComeResponse.Version

		_, proxyResponse, patchErr := proxy.updateJourneyOutcome(ctx, d.Id(), patchOutcome)
		if patchErr != nil {
			var errMsg string
			if patchOutcome.DisplayName != nil {
				errMsg = fmt.Sprintf("Error updating journey outcome %s: %s", *patchOutcome.DisplayName, patchErr)
			} else {
				errMsg = fmt.Sprintf("Error updating journey outcome DisplayName is nil: %s", patchErr)
			}

			input, _ := util.InterfaceToJson(*patchOutcome)
			return proxyResponse, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("%s\n(input: %+v)", errMsg, input), proxyResponse)
		}
		return proxyResponse, nil
	})

	if diagErr != nil {
		return diagErr
	}

	log.Printf("Updated journey outcome %s", d.Id())
	return readJourneyOutcome(ctx, d, meta)
}

// deleteJourneyOutcome deletes a journey outcome from Genesys Cloud.
//
// Parameters:
//   - ctx: context.Context for managing the request lifecycle and cancellation
//   - d: *schema.ResourceData containing the resource configuration and state
//   - meta: interface{} containing the provider metadata with ClientConfig
//
// Returns:
//   - diag.Diagnostics: Contains any error diagnostics encountered during deletion
//
// The function performs the following operations:
//  1. Validates input parameters for nil values
//  2. Gets SDK configuration from provider metadata
//  3. Creates journey outcome proxy for API operations
//  4. Retrieves and validates display name
//  5. Executes deletion operation
//  6. Implements retry mechanism to confirm deletion
//
// Error Handling:
//   - Returns error if ResourceData is nil
//   - Returns error if meta interface is nil
//   - Returns error if ClientConfig is nil
//   - Returns error if proxy creation fails
//   - Returns error if display_name is nil
//   - Handles API errors during deletion
//
// Retry Logic:
//   - Implements 30-second retry window
//   - Checks for 404 status to confirm deletion
//   - Returns non-retryable error for API failures
//   - Returns retryable error if resource still exists
//
// Example Usage in Terraform:
//
//	```hcl
//	resource "genesyscloud_journey_outcome" "example" {
//	  display_name = "Example Outcome"
//	  # ... other configuration ...
//	}
//	```
func deleteJourneyOutcome(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getJourneyOutcomeProxy(sdkConfig)

	displayNameRaw := d.Get("display_name")
	if displayNameRaw == nil {
		return diag.Errorf("display_name is nil")
	}
	displayName := displayNameRaw.(string)
	log.Printf("Deleting journey outcome with display name %s", displayName)
	if proxyResponse, err := proxy.deleteJourneyOutcome(ctx, d.Id()); err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("failed to delete journey outcome with display name %s error: %s", displayName, err), proxyResponse)
	}

	return util.WithRetries(ctx, 30*time.Second, func() *retry.RetryError {
		_, proxyResponse, err := proxy.getJourneyOutcomeById(ctx, d.Id())
		if err != nil {
			if util.IsStatus404(proxyResponse) {
				// journey action map deleted
				log.Printf("Deleted journey outcome %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("error deleting journey outcome %s | error: %s", d.Id(), err), proxyResponse))
		}
		return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("journey outcome %s still exists", d.Id()), proxyResponse))
	})
}
