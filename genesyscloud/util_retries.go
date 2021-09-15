package genesyscloud

import (
	"context"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/mypurecloud/platform-client-sdk-go/v53/platformclientv2"
)

func withRetries(ctx context.Context, timeout time.Duration, method func() *resource.RetryError) diag.Diagnostics {
	return diag.FromErr(resource.RetryContext(ctx, timeout, method))
}

type checkResponseFunc func(resp *platformclientv2.APIResponse) bool
type callSdkFunc func() (*platformclientv2.APIResponse, diag.Diagnostics)

// Retries up to 10 times while the shouldRetry condition returns true
// Useful for adding custom retry logic to normally non-retryable error codes
func retryWhen(shouldRetry checkResponseFunc, callSdk callSdkFunc) diag.Diagnostics {
	var lastErr diag.Diagnostics
	for i := 0; i < 10; i++ {
		resp, sdkErr := callSdk()
		if sdkErr != nil {
			if resp != nil && shouldRetry(resp) {
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

func isVersionMismatch(resp *platformclientv2.APIResponse) bool {
	// Version mismatch from directory may be a 409 or 400 with specific error message
	if resp != nil {
		if resp.StatusCode == 409 || (resp.StatusCode == 400 &&
			resp.Error != nil && strings.Contains((*resp.Error).Message, "does not match the current version")) {
			return true
		}
	}
	return false
}

func isStatus404(resp *platformclientv2.APIResponse) bool {
	if resp != nil && resp.StatusCode == 404 {
		return true
	}
	return false
}
