package genesyscloud

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/consistency_checker"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/mypurecloud/platform-client-sdk-go/v74/platformclientv2"
)

func withRetries(ctx context.Context, timeout time.Duration, method func() *resource.RetryError) diag.Diagnostics {
	err := diag.FromErr(resource.RetryContext(ctx, timeout, method))
	if err != nil && strings.Contains(fmt.Sprintf("%v", err), "timeout while waiting for state to become") {
		ctx, _ := context.WithTimeout(context.Background(), timeout)
		return withRetries(ctx, timeout, method)
	}
	return err
}

func withRetriesForRead(ctx context.Context, d *schema.ResourceData, method func() *resource.RetryError) diag.Diagnostics {
	return withRetriesForReadCustomTimeout(ctx, 5*time.Minute, d, method)
}

func withRetriesForReadCustomTimeout(ctx context.Context, timeout time.Duration, d *schema.ResourceData, method func() *resource.RetryError) diag.Diagnostics {
	err := diag.FromErr(resource.RetryContext(ctx, timeout, method))
	if err != nil {
		if strings.Contains(fmt.Sprintf("%v", err), "API Error: 404") {
			// Set ID empty if the object isn't found after the specified timeout
			d.SetId("")
		}
		errStringLower := strings.ToLower(fmt.Sprintf("%v", err))
		if strings.Contains(errStringLower, "timeout while waiting for state to become") ||
			strings.Contains(errStringLower, "context deadline exceeded") {
			ctx, _ := context.WithTimeout(context.Background(), timeout)
			return withRetriesForRead(ctx, d, method)
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
func retryWhen(shouldRetry checkResponseFunc, callSdk callSdkFunc, additionalCodes ...int) diag.Diagnostics {
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

func isAdditionalCode(statusCode int, additionalCodes ...int) bool {
	for _, additionalCode := range additionalCodes {
		if statusCode == additionalCode {
			return true
		}
	}

	return false
}

func isVersionMismatch(resp *platformclientv2.APIResponse, additionalCodes ...int) bool {
	// Version mismatch from directory may be a 409 or 400 with specific error message
	if resp != nil {
		if resp.StatusCode == http.StatusConflict ||
			resp.StatusCode == http.StatusRequestTimeout ||
			isAdditionalCode(resp.StatusCode, additionalCodes...) ||
			(resp.StatusCode == http.StatusBadRequest && resp.Error != nil && strings.Contains((*resp.Error).Message, "does not match the current version")) {
			return true
		}
	}
	return false
}

func isStatus404(resp *platformclientv2.APIResponse, additionalCodes ...int) bool {
	if resp != nil {
		if resp.StatusCode == http.StatusNotFound ||
			resp.StatusCode == http.StatusRequestTimeout ||
			resp.StatusCode == http.StatusGone ||
			isAdditionalCode(resp.StatusCode, additionalCodes...) {
			return true
		}
	}
	return false
}

func isStatus400(resp *platformclientv2.APIResponse, additionalCodes ...int) bool {
	if resp != nil {
		if resp.StatusCode == http.StatusBadRequest ||
			resp.StatusCode == http.StatusRequestTimeout ||
			isAdditionalCode(resp.StatusCode, additionalCodes...) {
			return true
		}
	}
	return false
}

func getBody(apiResponse *platformclientv2.APIResponse) string {
	if apiResponse != nil {
		return string(apiResponse.RawBody)
	}
	return ""
}
