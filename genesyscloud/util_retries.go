package genesyscloud

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/mypurecloud/platform-client-sdk-go/v56/platformclientv2"
)

func withRetries(ctx context.Context, timeout time.Duration, method func() *resource.RetryError) diag.Diagnostics {
	return diag.FromErr(resource.RetryContext(ctx, timeout, method))
}

func withRetriesForRead(ctx context.Context, timeout time.Duration, d *schema.ResourceData, method func() *resource.RetryError) diag.Diagnostics {
	err := diag.FromErr(resource.RetryContext(ctx, timeout, method))
	if err != nil && strings.Contains(fmt.Sprintf("%v", err), "API Error: 404") {
		// Set ID empty if the object isn't found after the specified timeout
		d.SetId("")
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
		if resp.StatusCode == 409 ||
			isAdditionalCode(resp.StatusCode, additionalCodes...) ||
			(resp.StatusCode == 400 && resp.Error != nil && strings.Contains((*resp.Error).Message, "does not match the current version")) {
			return true
		}
	}
	return false
}

func isStatus404(resp *platformclientv2.APIResponse, additionalCodes ...int) bool {
	if resp != nil {
		if resp.StatusCode == 404 || isAdditionalCode(resp.StatusCode, additionalCodes...) {
			return true
		}
	}
	return false
}

func isStatus400(resp *platformclientv2.APIResponse, additionalCodes ...int) bool {
	if resp != nil {
		if resp.StatusCode == 400 || isAdditionalCode(resp.StatusCode, additionalCodes...) {
			return true
		}
	}
	return false
}
