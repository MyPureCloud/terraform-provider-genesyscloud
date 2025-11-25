package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"runtime"
	"sync"

	"github.com/google/uuid"
)

type sdkDebugRequest struct {
	DebugType        string `json:"debug_type,omitempty"`        //Indicates whether it is a request or response debug
	TransactionId    string `json:"transaction_id,omitempty"`    //Unique id to link the request and response
	InvocationCount  int    `json:"invocation_count,omitempty"`  //Number of times the URL has been invoked.  Will be greater then zero when it is a retry
	InvocationMethod string `json:"invocation_method,omitempty"` //HTTP method that will be invoked
	InvocationUrl    string `json:"invocation_url,omitempty"`    //HTTP URL
	ResourceType     string `json:"resource_type,omitempty"`     //Terraform resource type (e.g., "genesyscloud_routing_queue")
	ResourceId       string `json:"resource_id,omitempty"`       //Terraform resource ID
	ResourceName     string `json:"resource_name,omitempty"`     //Terraform resource name
}

func (s *sdkDebugRequest) ToJSON() (err error, jsonStr string) {
	jsonData, err := json.Marshal(s)
	if err != nil {
		log.Println(err)
		return err, ""
	}

	// Print the JSON string
	return nil, string(jsonData)
}

type sdkDebugResponse struct {
	DebugType            string `json:"debug_type,omitempty"`             //Indicates whether it is a request or response debug
	TransactionId        string `json:"transaction_id,omitempty"`         //Unique id to link the request and response
	InvocationCount      int    `json:"invocation_count,omitempty"`       //Number of times the URL has been invoked.  Will be greater then zero when it is a retry
	InvocationMethod     string `json:"invocation_method,omitempty"`      //HTTP method that will be invoked
	InvocationUrl        string `json:"invocation_url,omitempty"`         //HTTP URL
	InvocationStatusCode int    `json:"invocation_status_code,omitempty"` //HTTP status code that has been returned
	InvocationRetryAfter string `json:"invocation_retry_after,omitempty"` //Retry-After header value
	ResourceType         string `json:"resource_type,omitempty"`          //Terraform resource type (e.g., "genesyscloud_routing_queue")
	ResourceId           string `json:"resource_id,omitempty"`            //Terraform resource ID
	ResourceName         string `json:"resource_name,omitempty"`          //Terraform resource name
}

func (s *sdkDebugResponse) ToJSON() (err error, jsonStr string) {
	jsonData, err := json.Marshal(s)
	if err != nil {
		log.Println(err)
		return err, ""
	}

	// Print the JSON string
	return nil, string(jsonData)
}

// resourceContextKey is a type for context keys to avoid collisions
type resourceContextKey struct{}

// ResourceContextKey returns the context key for resource metadata
// This is exported so utility functions can check if context is already set
func ResourceContextKey() resourceContextKey {
	return resourceContextKey{}
}

// contextStorage provides thread-local storage for context values using goroutine IDs
// This allows us to pass context from proxy functions to the RequestLogHook
// even though the SDK doesn't accept context parameters.
// Note: This uses runtime.Stack to get goroutine IDs, which works but is not officially supported.
// The storage is cleaned up automatically when contexts are no longer referenced.
var contextStorage = &sync.Map{} // map[uint64]context.Context

// bufferPool reduces allocations for the stack trace buffer
var bufferPool = sync.Pool{
	New: func() interface{} {
		return make([]byte, 64)
	},
}

// getGoroutineID extracts the goroutine ID from the stack trace
// This is a workaround since Go doesn't provide direct access to goroutine IDs.
// Optimized to use a buffer pool to reduce allocations.
func getGoroutineID() uint64 {
	// Get buffer from pool
	buf := bufferPool.Get().([]byte)
	defer bufferPool.Put(buf)

	n := runtime.Stack(buf, false)
	if n == 0 {
		return 0
	}

	// Parse the goroutine ID from the stack trace
	// Format: "goroutine 123 [running]:"
	// Skip "goroutine " (10 bytes) - optimized parsing
	var id uint64
	for i := 10; i < n && i < len(buf); i++ {
		c := buf[i]
		if c >= '0' && c <= '9' {
			id = id*10 + uint64(c-'0')
		} else if id > 0 {
			break
		}
	}

	return id
}

// setContextForRequest stores the context for the current goroutine
// This allows the RequestLogHook to access the context even though
// the SDK doesn't pass context to HTTP requests
func setContextForRequest(ctx context.Context) {
	if ctx == nil {
		return
	}
	goroutineID := getGoroutineID()
	if goroutineID > 0 {
		contextStorage.Store(goroutineID, ctx)
	}
}

// getContextForRequest retrieves the context for the current goroutine
func getContextForRequest() context.Context {
	goroutineID := getGoroutineID()
	if goroutineID > 0 {
		if ctx, ok := contextStorage.Load(goroutineID); ok {
			return ctx.(context.Context)
		}
	}
	return nil
}

// ResourceContext holds Terraform resource metadata
type ResourceContext struct {
	ResourceType string
	ResourceId   string
	ResourceName string
}

// extractIdOrNameFromJSON safely parses JSON and extracts id and name fields.
// Returns empty strings if the fields cannot be found or if any error occurs.
// This function never panics and handles all errors gracefully.
func extractIdOrNameFromJSON(jsonData []byte) (id string, name string) {
	if len(jsonData) == 0 {
		return "", ""
	}

	// Parse JSON into a generic map
	var jsonMap map[string]interface{}
	if err := json.Unmarshal(jsonData, &jsonMap); err != nil {
		// Silently fail - this is expected for non-JSON bodies
		return "", ""
	}

	// Extract id field if present
	if idVal, ok := jsonMap["id"]; ok && idVal != nil {
		switch v := idVal.(type) {
		case string:
			if v != "" {
				id = v
			}
		case float64:
			// Handle numeric IDs - convert to string
			id = fmt.Sprintf("%.0f", v)
		case int:
			id = fmt.Sprintf("%d", v)
		case int64:
			id = fmt.Sprintf("%d", v)
		}
	}

	// Extract name field if present
	if nameVal, ok := jsonMap["name"]; ok && nameVal != nil {
		if nameStr, ok := nameVal.(string); ok && nameStr != "" {
			name = nameStr
		}
	}

	return id, name
}

// newSDKDebugRequest creates a new SDK debug request with optional resource context from request context
func newSDKDebugRequest(request *http.Request, count int) *sdkDebugRequest {
	transactionId := uuid.NewString()
	debugReq := &sdkDebugRequest{
		DebugType:        "SDK DEBUG REQUEST",
		TransactionId:    transactionId,
		InvocationCount:  count,
		InvocationMethod: request.Method,
		InvocationUrl:    request.URL.Path,
	}

	// Try to extract resource context from request context first
	var resourceCtx *ResourceContext
	if ctx := request.Context(); ctx != nil {
		if rc, ok := ctx.Value(resourceContextKey{}).(*ResourceContext); ok && rc != nil {
			resourceCtx = rc
		}
	}

	// If not found in request context, try to get it from goroutine-local storage
	// This handles the case where the SDK doesn't pass context to HTTP requests
	if resourceCtx == nil {
		if storedCtx := getContextForRequest(); storedCtx != nil {
			if rc, ok := storedCtx.Value(resourceContextKey{}).(*ResourceContext); ok && rc != nil {
				resourceCtx = rc
				// Inject the context into the request so it's available for the response hook
				*request = *request.WithContext(storedCtx)
			}
		}
	}

	// Set resource context fields if found
	if resourceCtx != nil {
		debugReq.ResourceType = resourceCtx.ResourceType
		debugReq.ResourceId = resourceCtx.ResourceId
		debugReq.ResourceName = resourceCtx.ResourceName
	}

	// If ResourceId or ResourceName are empty, nil, or unavailable, try to extract from request body
	if debugReq.ResourceId == "" || debugReq.ResourceId == "unavailable" ||
		debugReq.ResourceName == "" || debugReq.ResourceName == "unavailable" {
		// Safely attempt to read and parse request body
		if request.Body != nil {
			bodyBytes, err := io.ReadAll(request.Body)
			if err == nil && len(bodyBytes) > 0 {
				// Restore the body for the actual request
				request.Body = io.NopCloser(bytes.NewReader(bodyBytes))

				// Extract id and name from JSON body
				extractedId, extractedName := extractIdOrNameFromJSON(bodyBytes)

				// Update ResourceId if it was empty/unavailable and we found an id
				if (debugReq.ResourceId == "" || debugReq.ResourceId == "unavailable") && extractedId != "" {
					debugReq.ResourceId = extractedId
				}

				// Update ResourceName if it was empty/unavailable and we found a name
				if (debugReq.ResourceName == "" || debugReq.ResourceName == "unavailable") && extractedName != "" {
					debugReq.ResourceName = extractedName
				}
			}
		}
	}

	return debugReq
}

func newSDKDebugResponse(response *http.Response) *sdkDebugResponse {
	transactionId := response.Request.Header.Get("TF-Correlation-Id")
	debugResp := &sdkDebugResponse{
		DebugType:            "SDK DEBUG RESPONSE",
		TransactionId:        transactionId,
		InvocationCount:      0,
		InvocationMethod:     response.Request.Method,
		InvocationUrl:        response.Request.URL.Path,
		InvocationStatusCode: response.StatusCode,
		InvocationRetryAfter: response.Request.Header.Get("Retry-After"),
	}

	// Extract resource context from request context if available
	if ctx := response.Request.Context(); ctx != nil {
		if resourceCtx, ok := ctx.Value(resourceContextKey{}).(*ResourceContext); ok && resourceCtx != nil {
			debugResp.ResourceType = resourceCtx.ResourceType
			debugResp.ResourceId = resourceCtx.ResourceId
			debugResp.ResourceName = resourceCtx.ResourceName
		}
	}

	// If ResourceId or ResourceName are empty, nil, or unavailable, try to extract from response body
	if debugResp.ResourceId == "" || debugResp.ResourceId == "unavailable" ||
		debugResp.ResourceName == "" || debugResp.ResourceName == "unavailable" {
		// Safely attempt to read and parse response body
		if response.Body != nil {
			bodyBytes, err := io.ReadAll(response.Body)
			if err == nil && len(bodyBytes) > 0 {
				// Restore the body in case it's needed later
				response.Body = io.NopCloser(bytes.NewReader(bodyBytes))

				// Extract id and name from JSON body
				extractedId, extractedName := extractIdOrNameFromJSON(bodyBytes)

				// Update ResourceId if it was empty/unavailable and we found an id
				if (debugResp.ResourceId == "" || debugResp.ResourceId == "unavailable") && extractedId != "" {
					debugResp.ResourceId = extractedId
				}

				// Update ResourceName if it was empty/unavailable and we found a name
				if (debugResp.ResourceName == "" || debugResp.ResourceName == "unavailable") && extractedName != "" {
					debugResp.ResourceName = extractedName
				}
			}
		}
	}

	return debugResp
}

// WithResourceContext adds Terraform resource metadata to the context
// This function also stores the context in goroutine-local storage so it can be
// accessed by the SDK's RequestLogHook even though the SDK doesn't pass context to HTTP requests.
func WithResourceContext(ctx context.Context, resourceType, resourceId, resourceName string) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	ctx = context.WithValue(ctx, resourceContextKey{}, &ResourceContext{
		ResourceType: resourceType,
		ResourceId:   resourceId,
		ResourceName: resourceName,
	})

	// Store the context for the RequestLogHook to access
	setContextForRequest(ctx)

	return ctx
}

// EnsureResourceContext ensures that the context has resource metadata set.
// If the context already has resource metadata, it returns the context unchanged.
// Otherwise, it sets the ResourceType with unavailable id/name.
// This is useful in proxy methods where ResourceData is not available.
// This function also stores the context in goroutine-local storage so it can be
// accessed by the SDK's RequestLogHook even though the SDK doesn't pass context to HTTP requests.
func EnsureResourceContext(ctx context.Context, resourceType string) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}

	// Check if context already has resource metadata (don't overwrite)
	if _, ok := ctx.Value(resourceContextKey{}).(*ResourceContext); ok {
		// Store the context for the RequestLogHook to access
		setContextForRequest(ctx)
		return ctx
	}

	// Set resource type with unavailable id/name since we don't have ResourceData in proxy methods
	ctx = WithResourceContext(ctx, resourceType, "unavailable", "unavailable")

	// Store the context for the RequestLogHook to access
	setContextForRequest(ctx)

	return ctx
}
