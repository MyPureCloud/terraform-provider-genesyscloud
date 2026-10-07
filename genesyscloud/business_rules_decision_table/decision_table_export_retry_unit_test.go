package business_rules_decision_table

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/mypurecloud/platform-client-sdk-go/v199/platformclientv2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// clearExportRetryBudgetEnv unsets the override for the test. t.Setenv cannot express an absent
// variable, so the unset case needs os.Unsetenv with an explicit restore.
func clearExportRetryBudgetEnv(t *testing.T) {
	t.Helper()
	original, had := os.LookupEnv(exportRetryBudgetEnvVar)
	t.Cleanup(func() {
		if had {
			require.NoError(t, os.Setenv(exportRetryBudgetEnvVar, original))
			return
		}
		require.NoError(t, os.Unsetenv(exportRetryBudgetEnvVar))
	})
	require.NoError(t, os.Unsetenv(exportRetryBudgetEnvVar))
}

// Only a strictly positive duration is honoured. Everything else falls back to the default, so a
// typo cannot disable or invert the budget.
func TestUnitResolveExportRetryBudget(t *testing.T) {
	cases := []struct {
		name  string
		value string
		set   bool
		want  time.Duration
	}{
		{name: "valid override", value: "45m", set: true, want: 45 * time.Minute},
		{name: "valid override with mixed units", value: "1h30m", set: true, want: 90 * time.Minute},
		{name: "unset", set: false, want: defaultExportRetryBudget},
		{name: "empty", value: "", set: true, want: defaultExportRetryBudget},
		{name: "unparseable bare number", value: "30", set: true, want: defaultExportRetryBudget},
		{name: "unparseable text", value: "soon", set: true, want: defaultExportRetryBudget},
		{name: "zero", value: "0s", set: true, want: defaultExportRetryBudget},
		{name: "negative", value: "-5m", set: true, want: defaultExportRetryBudget},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.set {
				t.Setenv(exportRetryBudgetEnvVar, tc.value)
			} else {
				clearExportRetryBudgetEnv(t)
			}

			assert.Equal(t, tc.want, resolveExportRetryBudget())
		})
	}
}

// shrinkExportRetryWindow narrows the wait window for the test and restores it afterwards, so
// tests exercise the real timer for microseconds instead of seconds.
func shrinkExportRetryWindow(t *testing.T, min, max time.Duration) {
	t.Helper()
	originalMin, originalMax := exportRetryMinInterval, exportRetryMaxInterval
	t.Cleanup(func() {
		exportRetryMinInterval, exportRetryMaxInterval = originalMin, originalMax
	})
	exportRetryMinInterval, exportRetryMaxInterval = min, max
}

// retryAfterResponse builds a 429 carrying a Retry-After header. The header is parsed as whole
// seconds, so seconds is an integer count.
func retryAfterResponse(seconds string) *platformclientv2.APIResponse {
	return &platformclientv2.APIResponse{
		StatusCode: http.StatusTooManyRequests,
		Response: &http.Response{
			Header: http.Header{"Retry-After": []string{seconds}},
		},
	}
}

// With no header the draw stays inside the window. A longer Retry-After overrides it; a shorter
// one is floored by it.
func TestUnitExportRetryInterval(t *testing.T) {
	const draws = 200

	t.Run("no header stays within the window", func(t *testing.T) {
		shrinkExportRetryWindow(t, 10*time.Millisecond, 20*time.Millisecond)

		for i := 0; i < draws; i++ {
			interval := exportRetryInterval(nil)
			assert.GreaterOrEqual(t, interval, exportRetryMinInterval)
			assert.LessOrEqual(t, interval, exportRetryMaxInterval)
		}
	})

	t.Run("header longer than the window wins", func(t *testing.T) {
		shrinkExportRetryWindow(t, 10*time.Millisecond, 20*time.Millisecond)
		resp := retryAfterResponse("1")

		for i := 0; i < draws; i++ {
			assert.Equal(t, time.Second, exportRetryInterval(resp))
		}
	})

	t.Run("header shorter than the draw is floored by the draw", func(t *testing.T) {
		shrinkExportRetryWindow(t, 10*time.Second, 30*time.Second)
		resp := retryAfterResponse("1")

		for i := 0; i < draws; i++ {
			interval := exportRetryInterval(resp)
			assert.GreaterOrEqual(t, interval, exportRetryMinInterval)
			assert.LessOrEqual(t, interval, exportRetryMaxInterval)
		}
	})

	t.Run("unparseable header falls back to the window", func(t *testing.T) {
		shrinkExportRetryWindow(t, 10*time.Millisecond, 20*time.Millisecond)
		resp := retryAfterResponse("soon")

		for i := 0; i < draws; i++ {
			interval := exportRetryInterval(resp)
			assert.GreaterOrEqual(t, interval, exportRetryMinInterval)
			assert.LessOrEqual(t, interval, exportRetryMaxInterval)
		}
	})
}

// exportJobResponse is one scripted outcome of the inner export job POST.
type exportJobResponse struct {
	job  *platformclientv2.Decisiontableexportjob
	resp *platformclientv2.APIResponse
	err  error
}

// exportJobStub replays scripted outcomes through the postExportJobAttr seam and counts calls.
// The last entry repeats once the script runs out, so an endless stream of rejections is one entry.
type exportJobStub struct {
	responses []exportJobResponse
	calls     int
}

func (s *exportJobStub) postFn(_ context.Context, _ *BusinessRulesDecisionTableProxy, _ string,
	_ *platformclientv2.Decisiontableexportjobrequest) (*platformclientv2.Decisiontableexportjob, *platformclientv2.APIResponse, error) {
	index := s.calls
	s.calls++
	if index >= len(s.responses) {
		index = len(s.responses) - 1
	}
	scripted := s.responses[index]
	return scripted.job, scripted.resp, scripted.err
}

// capacityRejectionOutcome is a full per-user concurrent cap: a 403 carrying the capacity message.
func capacityRejectionOutcome() exportJobResponse {
	return exportJobResponse{
		resp: &platformclientv2.APIResponse{
			StatusCode: http.StatusForbidden,
			Error: &platformclientv2.APIError{
				Message: decisionTableConcurrentExportLimitMessage +
					" [3] reached for user 00000000-0000-0000-0000-000000000000",
			},
		},
		err: errors.New("API Error: 403 - forbidden"),
	}
}

// throttleOutcome is a 429. Pass a nil response for a throttle with no Retry-After header.
func throttleOutcome(response *http.Response) exportJobResponse {
	return exportJobResponse{
		resp: &platformclientv2.APIResponse{
			StatusCode: http.StatusTooManyRequests,
			Response:   response,
		},
		err: errors.New("API Error: 429 - too many requests"),
	}
}

func exportSuccessOutcome() exportJobResponse {
	return exportJobResponse{
		job:  &platformclientv2.Decisiontableexportjob{},
		resp: &platformclientv2.APIResponse{StatusCode: http.StatusOK},
	}
}

// exportRetryProxy populates only the inner POST seam, which keeps these tests off the network.
func exportRetryProxy(stub *exportJobStub) *BusinessRulesDecisionTableProxy {
	return &BusinessRulesDecisionTableProxy{postExportJobAttr: stub.postFn}
}

// Rejections are waited out, not surfaced: the caller sees only the eventual success.
func TestUnitCreateExportJobRetriesCapacityRejection(t *testing.T) {
	// A generous budget pinned explicitly, so an inherited override cannot shorten the run below
	// the waits these three attempts need.
	t.Setenv(exportRetryBudgetEnvVar, "1m")
	shrinkExportRetryWindow(t, time.Millisecond, 2*time.Millisecond)

	stub := &exportJobStub{responses: []exportJobResponse{
		capacityRejectionOutcome(),
		capacityRejectionOutcome(),
		exportSuccessOutcome(),
	}}

	job, resp, err := createDecisionTableExportJobFn(context.Background(), exportRetryProxy(stub),
		"table-id", &platformclientv2.Decisiontableexportjobrequest{})

	assert.NoError(t, err)
	assert.Equal(t, 3, stub.calls, "both rejections should have been retried")
	require.NotNil(t, job)
	require.NotNil(t, resp)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// A Retry-After longer than the window is honoured in full, so the server's backpressure wins.
func TestUnitCreateExportJobRetriesThrottleWithRetryAfter(t *testing.T) {
	t.Setenv(exportRetryBudgetEnvVar, "1m")
	// Both bounds sit well below the header's 1 second, so a wait of at least a second can only
	// have come from the header.
	shrinkExportRetryWindow(t, time.Millisecond, 2*time.Millisecond)

	stub := &exportJobStub{responses: []exportJobResponse{
		throttleOutcome(&http.Response{Header: http.Header{"Retry-After": []string{"1"}}}),
		exportSuccessOutcome(),
	}}

	start := time.Now()
	job, resp, err := createDecisionTableExportJobFn(context.Background(), exportRetryProxy(stub),
		"table-id", &platformclientv2.Decisiontableexportjobrequest{})
	elapsed := time.Since(start)

	assert.NoError(t, err)
	assert.Equal(t, 2, stub.calls)
	assert.GreaterOrEqual(t, elapsed, time.Second, "the Retry-After header should have set the wait")
	require.NotNil(t, job)
	require.NotNil(t, resp)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// A throttle with no Retry-After falls back to the window, not to a fixed default or no wait.
func TestUnitCreateExportJobRetriesThrottleWithoutRetryAfter(t *testing.T) {
	t.Setenv(exportRetryBudgetEnvVar, "1m")
	shrinkExportRetryWindow(t, time.Millisecond, 2*time.Millisecond)

	stub := &exportJobStub{responses: []exportJobResponse{
		throttleOutcome(nil),
		exportSuccessOutcome(),
	}}

	start := time.Now()
	job, resp, err := createDecisionTableExportJobFn(context.Background(), exportRetryProxy(stub),
		"table-id", &platformclientv2.Decisiontableexportjobrequest{})
	elapsed := time.Since(start)

	assert.NoError(t, err)
	assert.Equal(t, 2, stub.calls)
	assert.Less(t, elapsed, time.Second, "the shrunk jitter window should have been used, not a header delay")
	require.NotNil(t, job)
	require.NotNil(t, resp)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// A permission 403 returns on the first attempt, so a misconfigured role fails fast instead of
// retrying for the whole budget.
func TestUnitCreateExportJobDoesNotRetryPermissionDenied(t *testing.T) {
	// A window far larger than any plausible test runtime, so any wait at all would be obvious.
	shrinkExportRetryWindow(t, 10*time.Second, 10*time.Second)

	permissionDenied := errors.New("API Error: 403 - You are not authorized to perform the requested action.")
	stub := &exportJobStub{responses: []exportJobResponse{{
		resp: &platformclientv2.APIResponse{
			StatusCode: http.StatusForbidden,
			Error: &platformclientv2.APIError{
				Message: "You are not authorized to perform the requested action.",
			},
		},
		err: permissionDenied,
	}}}

	start := time.Now()
	_, resp, err := createDecisionTableExportJobFn(context.Background(), exportRetryProxy(stub),
		"table-id", &platformclientv2.Decisiontableexportjobrequest{})
	elapsed := time.Since(start)

	assert.Equal(t, permissionDenied, err, "the permission error should be handed straight back")
	assert.Equal(t, 1, stub.calls, "a permission 403 must not be retried")
	assert.Less(t, elapsed, 100*time.Millisecond, "no wait should have happened")
	require.NotNil(t, resp)
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}

// Anything other than a capacity 403 or a 429 returns on the first attempt.
func TestUnitCreateExportJobDoesNotRetryNonCapacityResponses(t *testing.T) {
	cases := []struct {
		name    string
		outcome exportJobResponse
	}{
		{
			name: "transport failure with no response",
			outcome: exportJobResponse{
				resp: nil,
				err:  errors.New("Post \"https://api.example.com\": dial tcp: connection refused"),
			},
		},
		{
			name: "403 with an unparsed body",
			outcome: exportJobResponse{
				resp: &platformclientv2.APIResponse{
					StatusCode:   http.StatusForbidden,
					ErrorMessage: "API Error: 403 - forbidden (abc-123)",
				},
				err: errors.New("API Error: 403 - forbidden"),
			},
		},
		{
			// The per-table total limit: a 403 that waiting can never clear, excluded from the
			// retry predicate only because its message omits "concurrent". Retrying it would burn
			// the whole budget on a limit that needs an export job deleted by hand.
			name: "403 for the non-transient per-table total export job limit",
			outcome: exportJobResponse{
				resp: &platformclientv2.APIResponse{
					StatusCode: http.StatusForbidden,
					Error: &platformclientv2.APIError{
						Message: "Maximum number of export jobs [100] reached for table table-id. " +
							"Delete an existing export job before creating a new one.",
					},
				},
				err: errors.New("API Error: 403 - forbidden"),
			},
		},
		{
			name: "400 bad request",
			outcome: exportJobResponse{
				resp: &platformclientv2.APIResponse{StatusCode: http.StatusBadRequest},
				err:  errors.New("API Error: 400 - bad request"),
			},
		},
		{
			name: "404 not found",
			outcome: exportJobResponse{
				resp: &platformclientv2.APIResponse{StatusCode: http.StatusNotFound},
				err:  errors.New("API Error: 404 - not found"),
			},
		},
		{
			name: "500 internal server error",
			outcome: exportJobResponse{
				resp: &platformclientv2.APIResponse{StatusCode: http.StatusInternalServerError},
				err:  errors.New("API Error: 500 - internal server error"),
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			shrinkExportRetryWindow(t, 10*time.Second, 10*time.Second)

			stub := &exportJobStub{responses: []exportJobResponse{tc.outcome}}

			start := time.Now()
			_, resp, err := createDecisionTableExportJobFn(context.Background(), exportRetryProxy(stub),
				"table-id", &platformclientv2.Decisiontableexportjobrequest{})
			elapsed := time.Since(start)

			assert.Equal(t, tc.outcome.err, err, "the underlying error should be handed straight back")
			assert.Equal(t, 1, stub.calls, "a non-retryable response must not be retried")
			assert.Less(t, elapsed, 100*time.Millisecond, "no wait should have happened")
			assert.Equal(t, tc.outcome.resp, resp)
		})
	}
}

// silenceExportRetryLogs discards log output for the test. The budget tests retry hundreds of
// times, and one WARN line each would bury the rest of the suite.
func silenceExportRetryLogs(t *testing.T) {
	t.Helper()
	original := log.Writer()
	t.Cleanup(func() { log.SetOutput(original) })
	log.SetOutput(io.Discard)
}

// A cap that never clears is not retried forever. Once the budget is spent the caller gets the
// last attempt as the API returned it, not a synthesised timeout.
func TestUnitCreateExportJobStopsAtRetryBudget(t *testing.T) {
	const budget = 200 * time.Millisecond

	t.Setenv(exportRetryBudgetEnvVar, budget.String())
	// Equal bounds pin every wait at 1ms, so the whole budget is spent in a fifth of a second.
	shrinkExportRetryWindow(t, time.Millisecond, time.Millisecond)
	silenceExportRetryLogs(t)

	// One scripted outcome: exportJobStub repeats its last entry forever, which is the unbounded
	// stream of rejections this test needs. Held in a variable so the returned pointers can be
	// compared by identity.
	rejection := capacityRejectionOutcome()
	stub := &exportJobStub{responses: []exportJobResponse{rejection}}

	start := time.Now()
	job, resp, err := createDecisionTableExportJobFn(context.Background(), exportRetryProxy(stub),
		"table-id", &platformclientv2.Decisiontableexportjobrequest{})
	elapsed := time.Since(start)

	assert.Equal(t, rejection.err, err, "the last attempt's error should be handed back untouched")
	assert.Same(t, rejection.resp, resp, "the last attempt's response should be handed back untouched")
	assert.Nil(t, job)
	assert.Greater(t, stub.calls, 1, "the rejection should have been retried before the budget ran out")
	// Generous bounds either side: the loop must have spent most of the budget rather than
	// bailing out early, without failing on a slow CI box that overshoots the deadline.
	assert.GreaterOrEqual(t, elapsed, budget/2, "retrying should have consumed most of the budget")
	assert.Less(t, elapsed, budget+2*time.Second, "retrying should have stopped once the budget was spent")
}

// The bound is the deadline, not an attempt count: ten times the budget buys many more attempts
// at the same interval. Asserting the ratio is what catches a reintroduced fixed cap.
func TestUnitCreateExportJobAttemptsScaleWithBudget(t *testing.T) {
	// Any plausible fixed cap a refactor might reintroduce; the generic exporter wrapper uses 3.
	const plausibleFixedAttemptCap = 3

	callsForBudget := func(t *testing.T, budget time.Duration) int {
		t.Helper()
		t.Setenv(exportRetryBudgetEnvVar, budget.String())
		shrinkExportRetryWindow(t, time.Millisecond, time.Millisecond)
		silenceExportRetryLogs(t)

		stub := &exportJobStub{responses: []exportJobResponse{capacityRejectionOutcome()}}
		createDecisionTableExportJobFn(context.Background(), exportRetryProxy(stub),
			"table-id", &platformclientv2.Decisiontableexportjobrequest{})
		return stub.calls
	}

	var smallBudgetCalls, largeBudgetCalls int
	t.Run("100ms budget", func(t *testing.T) {
		smallBudgetCalls = callsForBudget(t, 100*time.Millisecond)
	})
	t.Run("1s budget", func(t *testing.T) {
		largeBudgetCalls = callsForBudget(t, time.Second)
	})

	assert.Greater(t, largeBudgetCalls, plausibleFixedAttemptCap*4,
		"a one second budget at a 1ms interval must buy far more attempts than any fixed cap")
	// A factor of two rather than the nominal ten: per-attempt timer and scheduler overhead eats
	// into the shorter budget proportionally more, and CI jitter widens that further.
	assert.Greater(t, largeBudgetCalls, 2*smallBudgetCalls,
		"ten times the budget should buy substantially more attempts at the same interval")
}

// A cancelled run does not sit out the rest of a wait, and the error wraps context.Canceled so
// errors.Is sees the cancellation rather than the rejection.
func TestUnitCreateExportJobHonoursContextCancellation(t *testing.T) {
	// A budget comfortably longer than the 10 second window, so the loop is willing to start the
	// wait: it refuses any wait that would cross the deadline, and a short budget would return the
	// rejection instead of ever reaching the cancellable wait this test is about.
	t.Setenv(exportRetryBudgetEnvVar, "1m")
	// Equal bounds pin the wait at 10 seconds, so a return inside 1 second can only have come from
	// the cancellation and not from the timer firing.
	shrinkExportRetryWindow(t, 10*time.Second, 10*time.Second)
	silenceExportRetryLogs(t)

	stub := &exportJobStub{responses: []exportJobResponse{capacityRejectionOutcome()}}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		time.Sleep(20 * time.Millisecond)
		cancel()
	}()

	start := time.Now()
	_, _, err := createDecisionTableExportJobFn(ctx, exportRetryProxy(stub),
		"table-id", &platformclientv2.Decisiontableexportjobrequest{})
	elapsed := time.Since(start)

	require.Error(t, err)
	assert.ErrorIs(t, err, context.Canceled, "the cancellation should be visible through errors.Is")
	assert.Less(t, elapsed, time.Second, "cancellation should have cut the 10 second wait short")
}

// Each retry logs a WARN naming the table, the status and the wait, so a noisy export is
// greppable per table.
func TestUnitCreateExportJobLogsRetryWarning(t *testing.T) {
	const tableId = "table-warn-log-7f3a"
	const interval = 5 * time.Millisecond

	t.Setenv(exportRetryBudgetEnvVar, "1m")
	// Equal bounds pin the wait at exactly one value, so the interval can be asserted literally
	// instead of as a range.
	shrinkExportRetryWindow(t, interval, interval)

	// A capture buffer rather than silenceExportRetryLogs: this test is about the log content.
	var logs bytes.Buffer
	original := log.Writer()
	t.Cleanup(func() { log.SetOutput(original) })
	log.SetOutput(&logs)

	// One rejection then a success, so exactly one retry decision is made and the buffer holds
	// exactly one WARN line.
	stub := &exportJobStub{responses: []exportJobResponse{
		capacityRejectionOutcome(),
		exportSuccessOutcome(),
	}}

	_, _, err := createDecisionTableExportJobFn(context.Background(), exportRetryProxy(stub),
		tableId, &platformclientv2.Decisiontableexportjobrequest{})

	assert.NoError(t, err)
	require.Equal(t, 2, stub.calls)

	captured := logs.String()
	assert.Equal(t, 1, strings.Count(captured, "[WARN]"), "one retry decision should log one WARN line")
	assert.Contains(t, captured, "[WARN]")
	assert.Contains(t, captured, tableId)
	assert.Contains(t, captured, "403")
	assert.Contains(t, captured, interval.String())
}
