package util

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/consistency_checker"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/constants"
	prl "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/panic_recovery_logger"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v179/platformclientv2"
)

func WithRetries(ctx context.Context, timeout time.Duration, method func() *retry.RetryError) diag.Diagnostics {
	method = wrapReadMethodWithRecover(method)
	err := diag.FromErr(retry.RetryContext(ctx, timeout, method))
	if err != nil && strings.Contains(fmt.Sprintf("%v", err), "timeout while waiting for state to become") {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		return WithRetries(ctx, timeout, method)
	}
	return err
}

// WithRetriesForRead retries a read operation with the configured custom retry timeout.
// The timeout is configurable via the provider's custom_retry_timeout attribute or the
// GENESYSCLOUD_CUSTOM_RETRY_TIMEOUT environment variable. Default is 5 minutes.
// Setting timeout to 0 means no retries (immediate fail-fast behavior).
func WithRetriesForRead(ctx context.Context, d *schema.ResourceData, method func() *retry.RetryError) diag.Diagnostics {
	timeout := provider.GetCustomRetryTimeout()
	return WithRetriesForReadCustomTimeout(ctx, timeout, d, method)
}

// WithRetriesForReadCustomTimeout retries a read operation with the specified timeout.
// A timeout of 0 means no retries - the method is called once and if it returns a 404,
// the resource is immediately removed from state (fail-fast behavior).
func WithRetriesForReadCustomTimeout(ctx context.Context, timeout time.Duration, d *schema.ResourceData, method func() *retry.RetryError) diag.Diagnostics {
	method = wrapReadMethodWithRecover(method)

	// Special handling for zero timeout: execute once without retry
	if timeout <= 0 {
		retryErr := method()
		if retryErr == nil {
			return nil
		}
		err := retryErr.Err
		errStr := fmt.Sprintf("%v", err)
		if strings.Contains(errStr, "API Error: 404") {
			// Resource not found - remove from state immediately
			d.SetId("")
			return nil
		}
		if d.Id() != "" {
			consistency_checker.DeleteConsistencyCheck(d.Id())
		}
		return diag.FromErr(err)
	}

	err := retry.RetryContext(ctx, timeout, method)
	if err != nil {
		if strings.Contains(fmt.Sprintf("%v", err), "API Error: 404") {
			// Set ID empty if the object isn't found after the specified timeout
			d.SetId("")
			return nil
		}

		if IsTimeoutError(err) {
			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			defer cancel()
			return WithRetriesForRead(ctx, d, method)
		}
		if d.Id() != "" {
			consistency_checker.DeleteConsistencyCheck(d.Id())
		}
	}
	return diag.FromErr(err)
}

// wrapReadMethodWithRecover will wrap the method with a recover if the panic recovery logger is enabled
func wrapReadMethodWithRecover(method func() *retry.RetryError) func() *retry.RetryError {
	return func() (retryErr *retry.RetryError) {
		defer func() {
			panicRecoveryLogger := prl.GetPanicRecoveryLoggerInstance()
			if !panicRecoveryLogger.LoggerEnabled {
				return
			}
			if r := recover(); r != nil {
				err := panicRecoveryLogger.HandleRecovery(r, constants.Read)
				if err != nil {
					retryErr = retry.NonRetryableError(err)
				}
			}
		}()
		return method()
	}
}

type checkResponseFunc func(resp *platformclientv2.APIResponse, additionalCodes ...int) bool
type callSdkFunc func() (*platformclientv2.APIResponse, diag.Diagnostics)

var maxRetries = 10

func SetMaxRetriesForTests(retries int) (previous int) {
	previous = maxRetries
	if retries > 0 {
		maxRetries = retries
	}
	return previous
}

// RetryWhen Retries up to 10 times while the shouldRetry condition returns true
// Useful for adding custom retry logic to normally non-retryable error codes
// Respects Retry-After header if present, otherwise uses exponential backoff
func RetryWhen(shouldRetry checkResponseFunc, callSdk callSdkFunc, additionalCodes ...int) diag.Diagnostics {
	var lastErr diag.Diagnostics
	for i := 0; i < maxRetries; i++ {
		resp, sdkErr := callSdk()
		if sdkErr != nil {
			if resp != nil && shouldRetry(resp, additionalCodes...) {
				lastErr = sdkErr
				// Check for Retry-After header first
				if delay, ok := GetRetryAfterDelay(resp); ok {
					time.Sleep(delay)
				} else {
					// Fall back to exponential backoff if no Retry-After header
					time.Sleep(time.Duration((i+1)*500) * time.Millisecond) // total 27.5 seconds for the 10 retries with exponential backoff on each retry
				}
				continue
			} else {
				return sdkErr
			}
		}
		// Success
		return nil
	}
	return diag.Errorf("Exhausted retries. Last error: %v", lastErr)
}

func IsAdditionalCode(statusCode int, additionalCodes ...int) bool {
	for _, additionalCode := range additionalCodes {
		if statusCode == additionalCode {
			return true
		}
	}

	return false
}

func IsVersionMismatch(resp *platformclientv2.APIResponse, additionalCodes ...int) bool {
	// Version mismatch from directory may be a 409 or 400 with specific error message
	if resp != nil {
		if resp.StatusCode == http.StatusConflict ||
			resp.StatusCode == http.StatusRequestTimeout ||
			IsAdditionalCode(resp.StatusCode, additionalCodes...) ||
			(resp.StatusCode == http.StatusBadRequest && resp.Error != nil && strings.Contains((*resp.Error).Message, "does not match the current version")) {
			return true
		}
	}
	return false
}

func IsStatus404(resp *platformclientv2.APIResponse, additionalCodes ...int) bool {
	if resp != nil {
		if resp.StatusCode == http.StatusNotFound ||
			resp.StatusCode == http.StatusRequestTimeout ||
			resp.StatusCode == http.StatusGone ||
			IsAdditionalCode(resp.StatusCode, additionalCodes...) {
			return true
		}
	}
	return false
}

func IsStatus404ByInt(respCode int, additionalCodes ...int) bool {

	if respCode == http.StatusNotFound ||
		respCode == http.StatusRequestTimeout ||
		respCode == http.StatusGone ||
		IsAdditionalCode(respCode, additionalCodes...) {
		return true
	}

	return false
}

func IsStatus400(resp *platformclientv2.APIResponse, additionalCodes ...int) bool {
	if resp != nil {
		if resp.StatusCode == http.StatusBadRequest ||
			resp.StatusCode == http.StatusRequestTimeout ||
			IsAdditionalCode(resp.StatusCode, additionalCodes...) {
			return true
		}
	}
	return false
}

func IsStatus409(resp *platformclientv2.APIResponse, additionalCodes ...int) bool {
	if resp != nil {
		if resp.StatusCode == http.StatusConflict ||
			resp.StatusCode == http.StatusRequestTimeout ||
			IsAdditionalCode(resp.StatusCode, additionalCodes...) {
			return true
		}
	}
	return false
}

func IsStatus412(resp *platformclientv2.APIResponse, additionalCodes ...int) bool {
	if resp != nil {
		if resp.StatusCode == http.StatusPreconditionFailed ||
			resp.StatusCode == http.StatusRequestTimeout ||
			IsAdditionalCode(resp.StatusCode, additionalCodes...) {
			return true
		}
	}
	return false
}

func IsTimeoutError(errDiag error) bool {
	errStringLower := strings.ToLower(fmt.Sprintf("%v", errDiag))
	return strings.Contains(errStringLower, "timeout") ||
		strings.Contains(errStringLower, "context deadline exceeded")
}

// GetRetryAfterDelay extracts and parses the Retry-After header from an APIResponse.
// It returns the delay duration and true if the header was present and valid, false otherwise.
// Retry-After is always specified as a number of seconds (integer).
func GetRetryAfterDelay(apiResponse *platformclientv2.APIResponse) (time.Duration, bool) {
	if apiResponse == nil || apiResponse.Response == nil {
		return 0, false
	}

	retryAfterHeader := apiResponse.Response.Header.Get("Retry-After")
	if retryAfterHeader != "" {
		if seconds, err := strconv.Atoi(retryAfterHeader); err == nil && seconds > 0 {
			return time.Duration(seconds) * time.Second, true
		}
	}
	return 0, false
}

// RetryableErrorWithRetryAfter creates a retry.RetryableError that respects the Retry-After header
// from the APIResponse if present. If the Retry-After header is present and valid, it will sleep
// for the specified duration before returning the retryable error. If the Retry-After header is
// not present or invalid, it falls back to the default retry behavior.
// Note: This function should be called within a retry function that has access to context for
// proper cancellation handling.
func RetryableErrorWithRetryAfter(ctx context.Context, err error, apiResponse *platformclientv2.APIResponse) *retry.RetryError {
	if delay, ok := GetRetryAfterDelay(apiResponse); ok {
		// Sleep for the Retry-After duration, respecting context cancellation
		select {
		case <-ctx.Done():
			// Context cancelled, return non-retryable error
			return retry.NonRetryableError(fmt.Errorf("context cancelled while waiting for Retry-After: %w", ctx.Err()))
		case <-time.After(delay):
			// Delay completed, return retryable error
		}
	}
	return retry.RetryableError(err)
}
