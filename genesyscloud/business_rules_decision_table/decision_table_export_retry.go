package business_rules_decision_table

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/mypurecloud/platform-client-sdk-go/v199/platformclientv2"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
)

// decisionTableConcurrentExportLimitMessage identifies a 403 raised because an export job cap is
// full. The 403 error code is the generic "forbidden", shared with real permission denials, so the
// message is the only usable signal.
//
// Do not shorten this string. Export job creation has three 403s and "concurrent" is the only word
// separating them:
//
//   - "...concurrent export jobs [3] reached for user X" is pending jobs per user, cleared by waiting.
//   - "...concurrent export jobs [5] reached for table X" is pending jobs per table, cleared by waiting.
//   - "Maximum number of export jobs [100] reached for table X" is the lifetime total per table,
//     cleared only by deleting a job. Retrying it would waste the whole budget.
const decisionTableConcurrentExportLimitMessage = "Maximum number of concurrent export jobs"

const (
	// defaultExportRetryBudget caps the total time spent retrying one export job creation call.
	// The cap is 3 concurrent jobs per user, so 30 minutes covers around 180 maximum-size tables
	// draining 3 at a time, and still fails in reasonable time against a cap that never clears.
	defaultExportRetryBudget = 30 * time.Minute

	// exportRetryBudgetEnvVar overrides defaultExportRetryBudget with any time.ParseDuration
	// value, for example "45m".
	exportRetryBudgetEnvVar = "GENESYSCLOUD_DECISION_TABLE_EXPORT_RETRY_BUDGET"
)

// Bounds of the jittered wait between attempts. Vars rather than consts only so tests can shrink
// the window; production never assigns them.
var (
	exportRetryMinInterval = 10 * time.Second
	exportRetryMaxInterval = 30 * time.Second
)

// resolveExportRetryBudget returns the override when it parses as a positive duration, and the
// default otherwise. A bad value warns rather than failing the export.
func resolveExportRetryBudget() time.Duration {
	value, ok := os.LookupEnv(exportRetryBudgetEnvVar)
	if !ok || value == "" {
		return defaultExportRetryBudget
	}
	budget, err := time.ParseDuration(value)
	if err != nil || budget <= 0 {
		log.Printf("[WARN] Invalid %s value %q; using default export retry budget %v",
			exportRetryBudgetEnvVar, value, defaultExportRetryBudget)
		return defaultExportRetryBudget
	}
	return budget
}

// isExportCapacityRetryable reports whether waiting could clear the response: a capacity 403 or a
// 429. A permission 403 must not match, or a misconfigured role would retry for the whole budget.
func isExportCapacityRetryable(resp *platformclientv2.APIResponse) bool {
	if resp == nil {
		return false
	}
	if util.IsStatus429(resp) {
		return true
	}
	if resp.StatusCode != http.StatusForbidden {
		return false
	}
	// Contains, not HasPrefix: the resp.ErrorMessage fallback prefixes the server text with
	// "API Error: 403 - ". The 403 check above keeps the match narrow.
	return strings.Contains(capacityRejectionMessage(resp), decisionTableConcurrentExportLimitMessage)
}

// capacityRejectionMessage returns the server message for a rejected response, falling back to
// resp.ErrorMessage when the body did not parse into resp.Error.
func capacityRejectionMessage(resp *platformclientv2.APIResponse) string {
	if resp == nil {
		return ""
	}
	if resp.Error != nil {
		return resp.Error.Message
	}
	return resp.ErrorMessage
}

// exportRetryInterval draws a uniform wait in [exportRetryMinInterval, exportRetryMaxInterval].
// A longer Retry-After wins; a shorter one is floored by the draw, so a small server-suggested
// delay cannot turn the exporter threads into a hammer.
func exportRetryInterval(resp *platformclientv2.APIResponse) time.Duration {
	interval := exportRetryMinInterval
	if span := int64(exportRetryMaxInterval - exportRetryMinInterval); span > 0 {
		interval += time.Duration(rand.Int63n(span + 1))
	}
	if headerDelay, ok := util.GetRetryAfterDelay(resp); ok && headerDelay > interval {
		return headerDelay
	}
	return interval
}

// waitWithContext sleeps for d unless ctx is cancelled first. The returned error wraps ctx.Err(),
// so callers can match it with errors.Is.
func waitWithContext(ctx context.Context, d time.Duration) error {
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return fmt.Errorf("cancelled while waiting to retry decision table export job creation: %w", ctx.Err())
	case <-timer.C:
		return nil
	}
}
