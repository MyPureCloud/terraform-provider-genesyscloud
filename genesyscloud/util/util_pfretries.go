package util

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/mypurecloud/platform-client-sdk-go/v176/platformclientv2"
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

// PFWithRetriesForRead retries a read operation with 404 handling
// Returns a special diagnostic if resource is not found (404)
// The caller should check for this and remove the resource from state
//
// Parameters:
//   - ctx: Context for cancellation and deadlines
//   - method: Function to retry - returns (shouldRetry bool, error)
//
// Returns:
//   - diag.Diagnostics: Framework diagnostics
//   - Returns a Warning diagnostic if resource not found (404)
//   - Returns empty diagnostics on success
//   - Returns Error diagnostic on other failures
//
// Example usage:
//
//	diags := util.PFWithRetriesForRead(ctx, func() (bool, error) {
//	    user, resp, err := proxy.getUserById(ctx, userId)
//	    if err != nil {
//	        if util.IsStatus404(resp) {
//	            return true, fmt.Errorf("API Error: 404") // Will be caught as not found
//	        }
//	        return false, fmt.Errorf("failed to read: %v", err)
//	    }
//	    return false, nil // Success
//	})
//	if diags.HasError() {
//	    return diags
//	}
//	// Check if resource was not found
//	if len(diags) > 0 && diags[0].Severity() == diag.SeverityWarning {
//	    resp.State.RemoveResource(ctx)
//	    return nil
//	}
func PFWithRetriesForRead(ctx context.Context, method PFRetryFunc) diag.Diagnostics {
	return PFWithRetriesForReadCustomTimeout(ctx, 5*time.Minute, method)
}

// PFWithRetriesForReadCustomTimeout retries a read operation with custom timeout and 404 handling
// Returns a special diagnostic if resource is not found (404)
// The caller should check for this and remove the resource from state
//
// Parameters:
//   - ctx: Context for cancellation and deadlines
//   - timeout: Maximum duration to retry
//   - method: Function to retry - returns (shouldRetry bool, error)
//
// Returns:
//   - diag.Diagnostics: Framework diagnostics
//   - Returns a Warning diagnostic if resource not found (404)
//   - Returns empty diagnostics on success
//   - Returns Error diagnostic on other failures
//
// Example usage:
//
//	diags := util.PFWithRetriesForReadCustomTimeout(ctx, 3*time.Minute, func() (bool, error) {
//	    user, resp, err := proxy.getUserById(ctx, userId)
//	    if err != nil {
//	        if util.IsStatus404(resp) {
//	            return true, fmt.Errorf("API Error: 404") // Will be caught as not found
//	        }
//	        return false, fmt.Errorf("failed to read: %v", err)
//	    }
//	    return false, nil // Success
//	})
//	if diags.HasError() {
//	    return diags
//	}
//	// Check if resource was not found
//	if len(diags) > 0 && diags[0].Severity() == diag.SeverityWarning {
//	    resp.State.RemoveResource(ctx)
//	    return nil
//	}
func PFWithRetriesForReadCustomTimeout(ctx context.Context, timeout time.Duration, method PFRetryFunc) diag.Diagnostics {
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
			// Non-retryable error - check if it's a 404
			if is404Error(err) {
				var diags diag.Diagnostics
				diags.AddWarning(
					"Resource Not Found",
					"The resource was not found (404). It may have been deleted outside of Terraform.",
				)
				return diags
			}

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

	// Timeout reached - check if last error was 404
	if lastErr != nil && is404Error(lastErr) {
		var diags diag.Diagnostics
		diags.AddWarning(
			"Resource Not Found",
			"The resource was not found (404) after timeout. It may have been deleted outside of Terraform.",
		)
		return diags
	}

	// Timeout with other error
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

// is404Error checks if an error message contains 404 indicators
func is404Error(err error) bool {
	if err == nil {
		return false
	}
	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "404") ||
		strings.Contains(errStr, "not found") ||
		strings.Contains(errStr, "api error: 404")
}
