package util

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

func WithRetries(ctx context.Context, timeout time.Duration, method func() *retry.RetryError) diag.Diagnostics {
	err := diag.FromErr(retry.RetryContext(ctx, timeout, method))
	if err != nil && strings.Contains(fmt.Sprintf("%v", err), "timeout while waiting for state to become") {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		return WithRetries(ctx, timeout, method)
	}
	return err
}

func WithRetriesForRead(ctx context.Context, d *schema.ResourceData, method func() *retry.RetryError) diag.Diagnostics {
	return WithRetriesForReadCustomTimeout(ctx, 5*time.Minute, d, method)
}

func WithRetriesForReadCustomTimeout(ctx context.Context, timeout time.Duration, d *schema.ResourceData, method func() *retry.RetryError) diag.Diagnostics {
	err := diag.FromErr(retry.RetryContext(ctx, timeout, method))
	if err != nil {
		if strings.Contains(fmt.Sprintf("%v", err), "API Error: 404") {
			// Set ID empty if the object isn't found after the specified timeout
			d.SetId("")
		}
		errStringLower := strings.ToLower(fmt.Sprintf("%v", err))
		if strings.Contains(errStringLower, "timeout while waiting for state to become") ||
			strings.Contains(errStringLower, "context deadline exceeded") {
			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			defer cancel()
			return WithRetriesForRead(ctx, d, method)
		}
		if d.Id() != "" {
			consistency_checker.DeleteConsistencyCheck(d.Id())
		}
	}
	return err
}

type checkResponseFunc func(resp *platformclientv2.APIResponse, additionalCodes ...int) bool
type callSdkFunc func() (*platformclientv2.APIResponse, diag.Diagnostics)

// Retries up to 10 times while the shouldRetry condition returns true
// Useful for adding custom retry logic to normally non-retryable error codes
func RetryWhen(shouldRetry checkResponseFunc, callSdk callSdkFunc, additionalCodes ...int) diag.Diagnostics {
	var lastErr diag.Diagnostics
	for i := 0; i < 10; i++ {
		resp, sdkErr := callSdk()
		if sdkErr != nil {
			if resp != nil && shouldRetry(resp, additionalCodes...) {
				// Wait a second and try again
				lastErr = sdkErr
				time.Sleep(time.Second)
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

func GetBody(apiResponse *platformclientv2.APIResponse) string {
	if apiResponse != nil {
		return string(apiResponse.RawBody)
	}
	return ""
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

func IsStatus412ByInt(respCode int, additionalCodes ...int) bool {
	if respCode == http.StatusPreconditionFailed ||
		respCode == http.StatusRequestTimeout ||
		IsAdditionalCode(respCode, additionalCodes...) {
		return true
	}

	return false
}
