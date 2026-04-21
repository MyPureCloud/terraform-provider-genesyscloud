package util

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v186/platformclientv2"
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

func TestUnitWithRetriesForReadCustomTimeout_ZeroTimeout_404Error(t *testing.T) {
	// Test that zero timeout with 404 error immediately removes resource from state
	resourceSchema := map[string]*schema.Schema{
		"name": {
			Type:     schema.TypeString,
			Optional: true,
		},
	}
	d := schema.TestResourceDataRaw(t, resourceSchema, map[string]interface{}{})
	d.SetId("test-resource-id")

	callCount := 0
	method := func() *retry.RetryError {
		callCount++
		return retry.RetryableError(fmt.Errorf("API Error: 404 - Resource not found"))
	}

	ctx := context.Background()
	diags := WithRetriesForReadCustomTimeout(ctx, 0, d, method)

	// Should have called the method exactly once
	if callCount != 1 {
		t.Errorf("Expected method to be called once, got %d calls", callCount)
	}

	// Should have cleared the ID (removed from state)
	if d.Id() != "" {
		t.Errorf("Expected resource ID to be cleared, got %s", d.Id())
	}

	// Should not return an error for 404
	if diags != nil {
		t.Errorf("Expected no diagnostics for 404 error, got %v", diags)
	}
}

func TestUnitWithRetriesForReadCustomTimeout_ZeroTimeout_NonRetryableError(t *testing.T) {
	// Test that zero timeout with non-404 error returns the error immediately
	resourceSchema := map[string]*schema.Schema{
		"name": {
			Type:     schema.TypeString,
			Optional: true,
		},
	}
	d := schema.TestResourceDataRaw(t, resourceSchema, map[string]interface{}{})
	d.SetId("test-resource-id")

	callCount := 0
	method := func() *retry.RetryError {
		callCount++
		return retry.NonRetryableError(fmt.Errorf("API Error: 403 - Permission denied"))
	}

	ctx := context.Background()
	diags := WithRetriesForReadCustomTimeout(ctx, 0, d, method)

	// Should have called the method exactly once
	if callCount != 1 {
		t.Errorf("Expected method to be called once, got %d calls", callCount)
	}

	// Should NOT have cleared the ID (not a 404)
	if d.Id() == "" {
		t.Errorf("Expected resource ID to remain set for non-404 error")
	}

	// Should return an error
	if diags == nil {
		t.Errorf("Expected diagnostics for non-404 error, got nil")
	}
}

func TestUnitWithRetriesForReadCustomTimeout_ZeroTimeout_Success(t *testing.T) {
	// Test that zero timeout with success returns nil
	resourceSchema := map[string]*schema.Schema{
		"name": {
			Type:     schema.TypeString,
			Optional: true,
		},
	}
	d := schema.TestResourceDataRaw(t, resourceSchema, map[string]interface{}{})
	d.SetId("test-resource-id")

	callCount := 0
	method := func() *retry.RetryError {
		callCount++
		return nil // Success
	}

	ctx := context.Background()
	diags := WithRetriesForReadCustomTimeout(ctx, 0, d, method)

	// Should have called the method exactly once
	if callCount != 1 {
		t.Errorf("Expected method to be called once, got %d calls", callCount)
	}

	// Should NOT have cleared the ID
	if d.Id() != "test-resource-id" {
		t.Errorf("Expected resource ID to remain set, got %s", d.Id())
	}

	// Should not return an error
	if diags != nil {
		t.Errorf("Expected no diagnostics for success, got %v", diags)
	}
}

func TestUnitWithRetriesForReadCustomTimeout_ZeroTimeout_RetryableNon404Error(t *testing.T) {
	// Test that zero timeout with retryable non-404 error returns error immediately (no retry)
	resourceSchema := map[string]*schema.Schema{
		"name": {
			Type:     schema.TypeString,
			Optional: true,
		},
	}
	d := schema.TestResourceDataRaw(t, resourceSchema, map[string]interface{}{})
	d.SetId("test-resource-id")

	callCount := 0
	method := func() *retry.RetryError {
		callCount++
		return retry.RetryableError(fmt.Errorf("API Error: 500 - Internal server error"))
	}

	ctx := context.Background()
	diags := WithRetriesForReadCustomTimeout(ctx, 0, d, method)

	// Should have called the method exactly once (no retries with zero timeout)
	if callCount != 1 {
		t.Errorf("Expected method to be called once with zero timeout, got %d calls", callCount)
	}

	// Should return an error
	if diags == nil {
		t.Errorf("Expected diagnostics for 500 error, got nil")
	}
}

func TestUnitWithRetriesForReadCustomTimeout_NormalTimeout_404Error(t *testing.T) {
	// Test that normal timeout with consistent 404 errors eventually removes resource from state
	resourceSchema := map[string]*schema.Schema{
		"name": {
			Type:     schema.TypeString,
			Optional: true,
		},
	}
	d := schema.TestResourceDataRaw(t, resourceSchema, map[string]interface{}{})
	d.SetId("test-resource-id")

	callCount := 0
	method := func() *retry.RetryError {
		callCount++
		return retry.RetryableError(fmt.Errorf("API Error: 404 - Resource not found"))
	}

	ctx := context.Background()
	// Use a short timeout for testing
	diags := WithRetriesForReadCustomTimeout(ctx, 2*time.Second, d, method)

	// Should have called the method multiple times due to retries
	if callCount < 2 {
		t.Errorf("Expected method to be called multiple times with retries, got %d calls", callCount)
	}

	// Should have cleared the ID (removed from state after timeout)
	if d.Id() != "" {
		t.Errorf("Expected resource ID to be cleared after timeout, got %s", d.Id())
	}

	// Should not return an error for 404 (handled gracefully)
	if diags != nil {
		t.Errorf("Expected no diagnostics for 404 error after timeout, got %v", diags)
	}
}
