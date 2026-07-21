package provider

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestSdkDebugStatusNeedsErrorMirror covers sdkDebugStatusNeedsErrorMirror status boundaries (429, 5xx, 501 excluded, etc.).
func TestSdkDebugStatusNeedsErrorMirror(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		status int
		want   bool
	}{
		{"ok_200", 200, false},
		{"client_400", 400, false},
		{"client_404", 404, false},
		{"rate_limit_429", 429, true},
		{"status_zero", 0, true},
		{"server_500", 500, true},
		{"server_502", 502, true},
		{"server_599", 599, true},
		{"not_implemented_501_excluded", http.StatusNotImplemented, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := sdkDebugStatusNeedsErrorMirror(tt.status); got != tt.want {
				t.Fatalf("sdkDebugStatusNeedsErrorMirror(%d) = %v, want %v", tt.status, got, tt.want)
			}
		})
	}
}

// TestCloneHTTPHeaderRedactAuth verifies cloneHTTPHeaderRedactAuth redacts Authorization and preserves other headers.
func TestCloneHTTPHeaderRedactAuth(t *testing.T) {
	t.Parallel()

	t.Run("nil", func(t *testing.T) {
		t.Parallel()
		if got := cloneHTTPHeaderRedactAuth(nil); got != nil {
			t.Fatalf("expected nil, got %#v", got)
		}
	})

	t.Run("redacts_authorization_case_insensitive", func(t *testing.T) {
		t.Parallel()
		// Use Header.Set so keys match what net/http expects for Get (canonical MIME keys).
		src := make(http.Header)
		src.Set("X-Custom", "keep-me")
		src.Set("authorization", "Bearer secret-token")
		src.Add("Other", "v1")
		src.Add("Other", "v2")
		wantAuth := src.Get("Authorization")

		dst := cloneHTTPHeaderRedactAuth(src)

		if got := dst.Get("Authorization"); got != "[REDACTED]" {
			t.Fatalf("authorization got %q, want [REDACTED]", got)
		}
		if got := dst.Values("Other"); len(got) != 2 || got[0] != "v1" || got[1] != "v2" {
			t.Fatalf("Other header: %#v", got)
		}
		if got := dst.Get("X-Custom"); got != "keep-me" {
			t.Fatalf("X-Custom: %q", got)
		}
		if src.Get("Authorization") != wantAuth {
			t.Fatalf("source Authorization was mutated")
		}
	})

	t.Run("AUTHORIZATION_uppercase_key", func(t *testing.T) {
		t.Parallel()
		src := http.Header{"AUTHORIZATION": {"Basic abc"}}
		dst := cloneHTTPHeaderRedactAuth(src)
		if dst.Get("AUTHORIZATION") != "[REDACTED]" {
			t.Fatalf("got %q", dst.Get("AUTHORIZATION"))
		}
	})
}

// TestStorePopSDKDebugMirrorRequestBodyForHook covers store/pop lifecycle, empty id, disabled flag, truncation, and nil input.
func TestStorePopSDKDebugMirrorRequestBodyForHook(t *testing.T) {
	prev := sdkDebugHookRequestBodyEnabled.Load()
	t.Cleanup(func() {
		sdkDebugHookRequestBodyEnabled.Store(prev)
	})

	t.Run("disabled_noop", func(t *testing.T) {
		sdkDebugHookRequestBodyEnabled.Store(false)
		cid := "cid-disabled-" + t.Name()
		storeSDKDebugMirrorRequestBodyForHook(&sdkDebugRequest{
			TransactionId: cid,
			RequestBody:   "payload",
		})
		if got := popSDKDebugMirrorRequestBody(cid); got != "" {
			t.Fatalf("expected empty pop when disabled, got %q", got)
		}
	})

	t.Run("empty_correlation_id", func(t *testing.T) {
		sdkDebugHookRequestBodyEnabled.Store(true)
		t.Cleanup(func() { sdkDebugHookRequestBodyEnabled.Store(false) })

		storeSDKDebugMirrorRequestBodyForHook(&sdkDebugRequest{
			TransactionId: "",
			RequestBody:   "should-not-store",
		})
		if got := popSDKDebugMirrorRequestBody(""); got != "" {
			t.Fatalf("pop with empty correlation id: got %q, want empty", got)
		}
	})

	t.Run("nil_request", func(t *testing.T) {
		sdkDebugHookRequestBodyEnabled.Store(true)
		t.Cleanup(func() { sdkDebugHookRequestBodyEnabled.Store(false) })
		storeSDKDebugMirrorRequestBodyForHook(nil)
	})

	t.Run("lifecycle_store_then_pop", func(t *testing.T) {
		sdkDebugHookRequestBodyEnabled.Store(true)
		t.Cleanup(func() { sdkDebugHookRequestBodyEnabled.Store(false) })

		cid := "cid-lifecycle-" + t.Name()
		body := "request-json-body"
		storeSDKDebugMirrorRequestBodyForHook(&sdkDebugRequest{
			TransactionId: cid,
			RequestBody:   body,
		})
		if got := popSDKDebugMirrorRequestBody(cid); got != body {
			t.Fatalf("first pop: got %q, want %q", got, body)
		}
		if got := popSDKDebugMirrorRequestBody(cid); got != "" {
			t.Fatalf("second pop should be empty, got %q", got)
		}
	})

	t.Run("empty_body_not_stored", func(t *testing.T) {
		sdkDebugHookRequestBodyEnabled.Store(true)
		t.Cleanup(func() { sdkDebugHookRequestBodyEnabled.Store(false) })

		cid := "cid-empty-body-" + t.Name()
		storeSDKDebugMirrorRequestBodyForHook(&sdkDebugRequest{
			TransactionId: cid,
			RequestBody:   "",
		})
		if got := popSDKDebugMirrorRequestBody(cid); got != "" {
			t.Fatalf("pop: got %q, want empty", got)
		}
	})

	t.Run("truncates_to_maxMirrorBodyBytes", func(t *testing.T) {
		sdkDebugHookRequestBodyEnabled.Store(true)
		t.Cleanup(func() { sdkDebugHookRequestBodyEnabled.Store(false) })

		cid := "cid-truncate-" + t.Name()
		large := strings.Repeat("a", maxMirrorBodyBytes+100)
		storeSDKDebugMirrorRequestBodyForHook(&sdkDebugRequest{
			TransactionId: cid,
			RequestBody:   large,
		})
		got := popSDKDebugMirrorRequestBody(cid)
		if len(got) != maxMirrorBodyBytes {
			t.Fatalf("stored body len %d, want %d", len(got), maxMirrorBodyBytes)
		}
		if want := strings.Repeat("a", maxMirrorBodyBytes); got != want {
			t.Fatalf("truncation prefix mismatch")
		}
	})
}

// TestReadRequestBodyForHook covers readRequestBodyForHook prealloc vs io.ReadAll paths and read error propagation.
func TestReadRequestBodyForHook(t *testing.T) {
	t.Parallel()

	payload := []byte("hello-hook-body")

	t.Run("prealloc_path_positive_content_length", func(t *testing.T) {
		t.Parallel()
		r := bytes.NewReader(payload)
		got, err := readRequestBodyForHook(r, int64(len(payload)))
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(got, payload) {
			t.Fatalf("got %q, want %q", got, payload)
		}
	})

	t.Run("fallback_zero_content_length", func(t *testing.T) {
		t.Parallel()
		r := bytes.NewReader(payload)
		got, err := readRequestBodyForHook(r, 0)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(got, payload) {
			t.Fatalf("got %q, want %q", got, payload)
		}
	})

	t.Run("fallback_negative_content_length", func(t *testing.T) {
		t.Parallel()
		r := bytes.NewReader(payload)
		got, err := readRequestBodyForHook(r, -1)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(got, payload) {
			t.Fatalf("got %q, want %q", got, payload)
		}
	})

	t.Run("fallback_content_length_exceeds_max_hint", func(t *testing.T) {
		t.Parallel()
		r := bytes.NewReader(payload)
		got, err := readRequestBodyForHook(r, int64(maxRequestBodyGrowHint)+1)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(got, payload) {
			t.Fatalf("got %q, want %q", got, payload)
		}
	})

	t.Run("prealloc_path_propagates_read_error", func(t *testing.T) {
		t.Parallel()
		wantErr := errors.New("read fail")
		r := io.LimitReader(errReader{err: wantErr}, 1<<20)
		_, err := readRequestBodyForHook(r, 100)
		if !errors.Is(err, wantErr) {
			t.Fatalf("err = %v, want %v", err, wantErr)
		}
	})

	t.Run("fallback_propagates_read_error", func(t *testing.T) {
		t.Parallel()
		wantErr := errors.New("read fail fallback")
		r := errReader{err: wantErr}
		_, err := readRequestBodyForHook(r, 0)
		if !errors.Is(err, wantErr) {
			t.Fatalf("err = %v, want %v", err, wantErr)
		}
	})
}

// errReader is a stub io.Reader used in tests to force Read errors without mocking net/http.
type errReader struct {
	err error
}

// Read always returns (0, err) for the configured errReader.err.
func (e errReader) Read(_ []byte) (int, error) {
	return 0, e.err
}

// TestNewSDKDebugRequest_CanceledStoredContext verifies that a canceled context
// in goroutine-local storage does not get injected into the HTTP request.
// This reproduces the bug where Terraform's r.read() defers cancel() on its
// timeout context, leaving a canceled context in contextStorage that would
// poison subsequent HTTP requests on the same goroutine.
func TestNewSDKDebugRequest_CanceledStoredContext(t *testing.T) {
	// Create a context with resource metadata and cancel it (simulating defer cancel() in r.read())
	ctx, cancel := context.WithCancel(context.Background())
	ctx = context.WithValue(ctx, resourceContextKey{}, &ResourceContext{
		ResourceType: "genesyscloud_outbound_contact_list",
		ResourceId:   "test-id-123",
		ResourceName: "test-contact-list",
	})
	cancel() // Simulate Terraform's defer cancel()

	// Store the canceled context in goroutine-local storage (simulates what autoInjectResourceContext does)
	setContextForRequest(ctx)

	// Create a bare HTTP request (simulates what the SDK does — no context attached)
	request := &http.Request{
		Method: "GET",
		URL:    &url.URL{Path: "/api/v2/outbound/attemptlimits"},
		Header: make(http.Header),
	}

	// Call the request hook
	debugReq := newSDKDebugRequest(request, 0)

	// The hook should still extract resource metadata for logging
	assert.Equal(t, "genesyscloud_outbound_contact_list", debugReq.ResourceType)
	assert.Equal(t, "test-id-123", debugReq.ResourceId)
	assert.Equal(t, "test-contact-list", debugReq.ResourceName)

	// The request's context must NOT be canceled — this is the critical assertion
	assert.NoError(t, request.Context().Err(), "request context should not be canceled; the logging hook must not inject a canceled context into the HTTP request")

	// Clean up goroutine-local storage
	goroutineID := getGoroutineID()
	contextStorage.Delete(goroutineID)
}

// TestNewSDKDebugRequest_ValidStoredContext verifies that when the stored context
// is still valid, resource metadata is still extracted for logging.
func TestNewSDKDebugRequest_ValidStoredContext(t *testing.T) {
	// Create a valid (non-canceled) context with resource metadata
	ctx := context.WithValue(context.Background(), resourceContextKey{}, &ResourceContext{
		ResourceType: "genesyscloud_auth_division",
		ResourceId:   "div-456",
		ResourceName: "Marketing",
	})

	// Store in goroutine-local storage
	setContextForRequest(ctx)

	// Create a bare HTTP request
	request := &http.Request{
		Method: "GET",
		URL:    &url.URL{Path: "/api/v2/authorization/divisions"},
		Header: make(http.Header),
	}

	// Call the request hook
	debugReq := newSDKDebugRequest(request, 0)

	// Metadata should be extracted
	assert.Equal(t, "genesyscloud_auth_division", debugReq.ResourceType)
	assert.Equal(t, "div-456", debugReq.ResourceId)
	assert.Equal(t, "Marketing", debugReq.ResourceName)

	// Request context should not be canceled
	assert.NoError(t, request.Context().Err())

	// Clean up
	goroutineID := getGoroutineID()
	contextStorage.Delete(goroutineID)
}

// TestNewSDKDebugResponse_FallsBackToGoroutineLocalStorage verifies that the
// response hook can find resource metadata from goroutine-local storage when
// the request context doesn't have it (which is now the case since we no longer
// inject context into requests).
func TestNewSDKDebugResponse_FallsBackToGoroutineLocalStorage(t *testing.T) {
	// Store resource context in goroutine-local storage
	ctx := context.WithValue(context.Background(), resourceContextKey{}, &ResourceContext{
		ResourceType: "genesyscloud_outbound_attempt_limit",
		ResourceId:   "attempt-789",
		ResourceName: "MyAttemptLimit",
	})
	setContextForRequest(ctx)

	// Create a response whose request has NO resource context (bare context)
	response := &http.Response{
		StatusCode: 200,
		Request: &http.Request{
			Method: "GET",
			URL:    &url.URL{Path: "/api/v2/outbound/attemptlimits"},
			Header: make(http.Header),
		},
		Body: io.NopCloser(strings.NewReader(`{"id":"attempt-789","name":"MyAttemptLimit"}`)),
	}

	// Call the response hook
	debugResp := newSDKDebugResponse(response)

	// The response hook should find metadata via goroutine-local storage fallback
	assert.Equal(t, "genesyscloud_outbound_attempt_limit", debugResp.ResourceType)
	assert.Equal(t, "attempt-789", debugResp.ResourceId)
	assert.Equal(t, "MyAttemptLimit", debugResp.ResourceName)

	// Clean up
	goroutineID := getGoroutineID()
	contextStorage.Delete(goroutineID)
}
