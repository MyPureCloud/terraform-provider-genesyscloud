package util

import (
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/mypurecloud/platform-client-sdk-go/v171/platformclientv2"
)

func TestUnitGetRetryAfterDelay_IntegerFormat(t *testing.T) {
	// Test with a static integer value between 5 and 60 seconds
	seconds := 25
	expectedDelay := time.Duration(seconds) * time.Second

	// Create mock APIResponse with integer Retry-After header
	apiResponse := &platformclientv2.APIResponse{
		Response: &http.Response{
			Header: make(http.Header),
		},
	}
	apiResponse.Response.Header.Set("Retry-After", strconv.Itoa(seconds))

	delay, ok := GetRetryAfterDelay(apiResponse)
	if !ok {
		t.Errorf("Expected GetRetryAfterDelay to return true for integer value %d, got false", seconds)
	}
	if delay != expectedDelay {
		t.Errorf("Expected delay %v for integer value %d, got %v", expectedDelay, seconds, delay)
	}
}

func TestUnitGetRetryAfterDelay_RFC1123Format(t *testing.T) {
	// Test with a static delay between 5 and 60 seconds using RFC1123 format
	seconds := 42
	futureTime := time.Now().Add(time.Duration(seconds) * time.Second)
	expectedDelay := time.Duration(seconds) * time.Second

	// Create mock APIResponse with RFC1123 Retry-After header
	apiResponse := &platformclientv2.APIResponse{
		Response: &http.Response{
			Header: make(http.Header),
		},
	}
	apiResponse.Response.Header.Set("Retry-After", futureTime.Format(time.RFC1123))

	delay, ok := GetRetryAfterDelay(apiResponse)
	if !ok {
		t.Errorf("Expected GetRetryAfterDelay to return true for RFC1123 format, got false")
	}
	// Allow small tolerance for timing differences (within 1 second)
	diff := delay - expectedDelay
	if diff < 0 {
		diff = -diff
	}
	if diff > time.Second {
		t.Errorf("Expected delay approximately %v for RFC1123 format, got %v (difference: %v)", expectedDelay, delay, diff)
	}
}

func TestUnitGetRetryAfterDelay_RFC1123ZFormat(t *testing.T) {
	// Test with a static delay between 5 and 60 seconds using RFC1123Z format
	seconds := 18
	futureTime := time.Now().Add(time.Duration(seconds) * time.Second)
	expectedDelay := time.Duration(seconds) * time.Second

	// Create mock APIResponse with RFC1123Z Retry-After header
	apiResponse := &platformclientv2.APIResponse{
		Response: &http.Response{
			Header: make(http.Header),
		},
	}
	apiResponse.Response.Header.Set("Retry-After", futureTime.Format(time.RFC1123Z))

	delay, ok := GetRetryAfterDelay(apiResponse)
	if !ok {
		t.Errorf("Expected GetRetryAfterDelay to return true for RFC1123Z format, got false")
	}
	// Allow small tolerance for timing differences (within 1 second)
	diff := delay - expectedDelay
	if diff < 0 {
		diff = -diff
	}
	if diff > time.Second {
		t.Errorf("Expected delay approximately %v for RFC1123Z format, got %v (difference: %v)", expectedDelay, delay, diff)
	}
}

func TestUnitGetRetryAfterDelay_RFC822Format(t *testing.T) {
	// Test with a static delay between 5 and 60 seconds using RFC822 format
	// Note: RFC822 uses abbreviated timezone names which can be ambiguous,
	// so we test that it at least parses and returns a positive delay
	seconds := 33
	futureTime := time.Now().Add(time.Duration(seconds) * time.Second)

	// Create mock APIResponse with RFC822 Retry-After header
	apiResponse := &platformclientv2.APIResponse{
		Response: &http.Response{
			Header: make(http.Header),
		},
	}
	apiResponse.Response.Header.Set("Retry-After", futureTime.Format(time.RFC822))

	delay, ok := GetRetryAfterDelay(apiResponse)
	// RFC822 parsing can be unreliable due to timezone abbreviation ambiguity,
	// so we just verify it either parses correctly or fails gracefully
	if ok {
		// If it parsed, verify it's a positive delay (at least 1 second due to timing)
		if delay <= 0 {
			t.Errorf("Expected positive delay for RFC822 format, got %v", delay)
		}
	}
	// If it didn't parse (ok == false), that's acceptable due to RFC822 timezone ambiguity
}

func TestUnitGetRetryAfterDelay_NoHeader(t *testing.T) {
	// Test with no Retry-After header
	apiResponse := &platformclientv2.APIResponse{
		Response: &http.Response{
			Header: make(http.Header),
		},
	}

	delay, ok := GetRetryAfterDelay(apiResponse)
	if ok {
		t.Errorf("Expected GetRetryAfterDelay to return false when header is missing, got true")
	}
	if delay != 0 {
		t.Errorf("Expected delay 0 when header is missing, got %v", delay)
	}
}

func TestUnitGetRetryAfterDelay_EmptyHeader(t *testing.T) {
	// Test with empty Retry-After header
	apiResponse := &platformclientv2.APIResponse{
		Response: &http.Response{
			Header: make(http.Header),
		},
	}
	apiResponse.Response.Header.Set("Retry-After", "")

	delay, ok := GetRetryAfterDelay(apiResponse)
	if ok {
		t.Errorf("Expected GetRetryAfterDelay to return false when header is empty, got true")
	}
	if delay != 0 {
		t.Errorf("Expected delay 0 when header is empty, got %v", delay)
	}
}

func TestUnitGetRetryAfterDelay_NilResponse(t *testing.T) {
	// Test with nil APIResponse
	delay, ok := GetRetryAfterDelay(nil)
	if ok {
		t.Errorf("Expected GetRetryAfterDelay to return false when APIResponse is nil, got true")
	}
	if delay != 0 {
		t.Errorf("Expected delay 0 when APIResponse is nil, got %v", delay)
	}
}

func TestUnitGetRetryAfterDelay_NilResponseResponse(t *testing.T) {
	// Test with nil Response field
	apiResponse := &platformclientv2.APIResponse{
		Response: nil,
	}

	delay, ok := GetRetryAfterDelay(apiResponse)
	if ok {
		t.Errorf("Expected GetRetryAfterDelay to return false when Response is nil, got true")
	}
	if delay != 0 {
		t.Errorf("Expected delay 0 when Response is nil, got %v", delay)
	}
}

func TestUnitGetRetryAfterDelay_InvalidFormat(t *testing.T) {
	// Test with invalid format
	apiResponse := &platformclientv2.APIResponse{
		Response: &http.Response{
			Header: make(http.Header),
		},
	}
	apiResponse.Response.Header.Set("Retry-After", "invalid-format")

	delay, ok := GetRetryAfterDelay(apiResponse)
	if ok {
		t.Errorf("Expected GetRetryAfterDelay to return false for invalid format, got true")
	}
	if delay != 0 {
		t.Errorf("Expected delay 0 for invalid format, got %v", delay)
	}
}

func TestUnitGetRetryAfterDelay_ZeroSeconds(t *testing.T) {
	// Test with zero seconds (should return false)
	apiResponse := &platformclientv2.APIResponse{
		Response: &http.Response{
			Header: make(http.Header),
		},
	}
	apiResponse.Response.Header.Set("Retry-After", "0")

	delay, ok := GetRetryAfterDelay(apiResponse)
	if ok {
		t.Errorf("Expected GetRetryAfterDelay to return false for zero seconds, got true")
	}
	if delay != 0 {
		t.Errorf("Expected delay 0 for zero seconds, got %v", delay)
	}
}

func TestUnitGetRetryAfterDelay_NegativeSeconds(t *testing.T) {
	// Test with negative seconds (should return false)
	apiResponse := &platformclientv2.APIResponse{
		Response: &http.Response{
			Header: make(http.Header),
		},
	}
	apiResponse.Response.Header.Set("Retry-After", "-5")

	delay, ok := GetRetryAfterDelay(apiResponse)
	// Negative seconds should fail to parse as integer, so it will try date formats
	// Since it's not a valid date either, it should return false
	if ok {
		t.Errorf("Expected GetRetryAfterDelay to return false for negative seconds, got true")
	}
	if delay != 0 {
		t.Errorf("Expected delay 0 for negative seconds, got %v", delay)
	}
}

func TestUnitGetRetryAfterDelay_PastDate(t *testing.T) {
	// Test with past date (should return false)
	pastTime := time.Now().Add(-10 * time.Second)
	apiResponse := &platformclientv2.APIResponse{
		Response: &http.Response{
			Header: make(http.Header),
		},
	}
	apiResponse.Response.Header.Set("Retry-After", pastTime.Format(time.RFC1123))

	delay, ok := GetRetryAfterDelay(apiResponse)
	if ok {
		t.Errorf("Expected GetRetryAfterDelay to return false for past date, got true")
	}
	if delay != 0 {
		t.Errorf("Expected delay 0 for past date, got %v", delay)
	}
}
