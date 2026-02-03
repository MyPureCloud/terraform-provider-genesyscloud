package util

import (
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/mypurecloud/platform-client-sdk-go/v178/platformclientv2"
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
	// Negative seconds should fail to parse as integer, so it should return false
	if ok {
		t.Errorf("Expected GetRetryAfterDelay to return false for negative seconds, got true")
	}
	if delay != 0 {
		t.Errorf("Expected delay 0 for negative seconds, got %v", delay)
	}
}
