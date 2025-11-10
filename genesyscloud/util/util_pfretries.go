package util

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/mypurecloud/platform-client-sdk-go/v165/platformclientv2"
)

// PFCheckResponseFunc is a function type that determines if a response should trigger a retry
// It receives the API response and optional additional status codes to check
type PFCheckResponseFunc func(*platformclientv2.APIResponse, ...int) bool

// PFCallSdkFunc is a function type that makes an SDK call and returns response and diagnostics
// This is the Plugin Framework equivalent of callSdkFunc
type PFCallSdkFunc func() (*platformclientv2.APIResponse, diag.Diagnostics)

// PFRetryWhen retries up to 10 times while the shouldRetry condition returns true
// This is the Plugin Framework equivalent of RetryWhen
// Useful for adding custom retry logic to normally non-retryable error codes
//
// Parameters:
//   - shouldRetry: Function that determines if the response should trigger a retry
//   - callSdk: Function that makes the SDK call
//   - additionalCodes: Optional additional HTTP status codes to check
//
// Returns:
//   - diag.Diagnostics: Framework diagnostics (empty on success, error diagnostics on failure)
//
// Example usage:
//
//	diags := util.PFRetryWhen(util.IsVersionMismatch, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
//	    _, resp, err := proxy.deleteUser(ctx, userId)
//	    if err != nil {
//	        return resp, util.BuildFrameworkAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed: %s", err), resp)
//	    }
//	    return resp, nil
//	})
func PFRetryWhen(shouldRetry PFCheckResponseFunc, callSdk PFCallSdkFunc, additionalCodes ...int) diag.Diagnostics {
	var lastErr diag.Diagnostics

	const maxRetries = 10

	for i := 0; i < maxRetries; i++ {
		resp, sdkErr := callSdk()

		if sdkErr != nil && sdkErr.HasError() {
			if resp != nil && shouldRetry(resp, additionalCodes...) {
				// Wait with exponential backoff and try again
				lastErr = sdkErr
				backoffDuration := time.Duration((i+1)*500) * time.Millisecond
				time.Sleep(backoffDuration) // total 27.5 seconds for 10 retries with exponential backoff
				continue
			} else {
				// Non-retryable error
				return sdkErr
			}
		}

		// Success
		return nil
	}

	// Exhausted retries
	var finalDiags diag.Diagnostics
	finalDiags.AddError(
		"Retry Limit Exceeded",
		fmt.Sprintf("Exhausted %d retries. Last error: %v", maxRetries, lastErr),
	)
	return finalDiags
}

// PFRetryFunc is a function type for retry operations
// Returns true to continue retrying, false to stop (success)
// Returns error to stop with failure
type PFRetryFunc func() (shouldRetry bool, err error)

// PFWithRetries retries a function until timeout or success
// This is the Plugin Framework equivalent of WithRetries
//
// Parameters:
//   - ctx: Context for cancellation and deadlines
//   - timeout: Maximum duration to retry
//   - method: Function to retry - returns (shouldRetry bool, error)
//   - Return (false, nil) for success
//   - Return (true, error) to retry with the error
//   - Return (false, error) to fail immediately
//
// Returns:
//   - diag.Diagnostics: Framework diagnostics (empty on success, error diagnostics on failure)
//
// Example usage:
//
//	diags := util.PFWithRetries(ctx, 3*time.Minute, func() (bool, error) {
//	    id, err := getDeletedUserId(email, proxy)
//	    if err != nil {
//	        return false, fmt.Errorf("error searching: %v", err) // Non-retryable
//	    }
//	    if id == nil {
//	        return true, fmt.Errorf("not yet deleted") // Retryable
//	    }
//	    return false, nil // Success
//	})
func PFWithRetries(ctx context.Context, timeout time.Duration, method PFRetryFunc) diag.Diagnostics {
	deadline := time.Now().Add(timeout)
	attempt := 0
	var lastErr error

	for time.Now().Before(deadline) {
		// Check if context is cancelled
		select {
		case <-ctx.Done():
			var diags diag.Diagnostics
			diags.AddError(
				"Operation Cancelled",
				fmt.Sprintf("Operation cancelled by context: %v", ctx.Err()),
			)
			return diags
		default:
			// Continue with retry
		}

		attempt++
		shouldRetry, err := method()

		if err == nil && !shouldRetry {
			// Success
			return nil
		}

		if err != nil && !shouldRetry {
			// Non-retryable error
			var diags diag.Diagnostics
			diags.AddError("Operation Failed", err.Error())
			return diags
		}

		// Retryable error or shouldRetry is true
		lastErr = err

		// Calculate backoff
		backoff := time.Duration(attempt) * time.Second
		if backoff > 10*time.Second {
			backoff = 10 * time.Second
		}

		// Check if we have time for another attempt
		if time.Now().Add(backoff).After(deadline) {
			break
		}

		// Sleep with context awareness
		select {
		case <-ctx.Done():
			var diags diag.Diagnostics
			diags.AddError(
				"Operation Cancelled",
				fmt.Sprintf("Operation cancelled during backoff: %v", ctx.Err()),
			)
			return diags
		case <-time.After(backoff):
			// Continue to next attempt
		}
	}

	// Timeout reached
	var diags diag.Diagnostics
	if lastErr != nil {
		diags.AddError(
			"Operation Timeout",
			fmt.Sprintf("Operation timed out after %v. Last error: %v", timeout, lastErr),
		)
	} else {
		diags.AddError(
			"Operation Timeout",
			fmt.Sprintf("Operation timed out after %v", timeout),
		)
	}
	return diags
}
