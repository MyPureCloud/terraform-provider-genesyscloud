package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
)

// sdkDebugHookRequestBodyEnabled mirrors provider sdk_debug: when true, RequestLogHook
// JSON includes full request_body. When false, bodies are only read when needed to
// infer resource id/name (avoids logging large/sensitive payloads on every request).
var sdkDebugHookRequestBodyEnabled atomic.Bool

// sdkDebugErrorMirrorPath is the sdk_debug log file path when sdk_debug is enabled; empty disables mirroring.
var sdkDebugErrorMirrorPath atomic.Value // string

var sdkDebugErrorMirrorMu sync.Mutex

// sdkDebugMirrorRequestBodies maps TF-Correlation-Id (set in RequestLogHook) to the outbound body bytes
var sdkDebugMirrorRequestBodies sync.Map

// maxMirrorBodyBytes caps mirrored request/response bodies written to sdk_debug.log.
const maxMirrorBodyBytes = 1 << 20

func storeSDKDebugMirrorRequestBodyForHook(dr *sdkDebugRequest) {
	if dr == nil || dr.TransactionId == "" || !sdkDebugHookRequestBodyEnabled.Load() {
		return
	}
	body := dr.RequestBody
	if body == "" {
		return
	}
	if len(body) > maxMirrorBodyBytes {
		body = body[:maxMirrorBodyBytes]
	}
	sdkDebugMirrorRequestBodies.Store(dr.TransactionId, body)
}

func popSDKDebugMirrorRequestBody(correlationID string) string {
	if correlationID == "" {
		return ""
	}
	v, ok := sdkDebugMirrorRequestBodies.LoadAndDelete(correlationID)
	if !ok {
		return ""
	}
	s, _ := v.(string)
	return s
}

func sdkDebugStatusNeedsErrorMirror(statusCode int) bool {
	if statusCode == http.StatusTooManyRequests {
		return true
	}
	if statusCode == 0 || (statusCode >= 500 && statusCode != http.StatusNotImplemented) {
		return true
	}
	return false
}

// sdkDebugFileErrorLine matches the Genesys platform SDK logStatement JSON when sdk_debug_format is Json
// (see platformclientv2/logger.go logStatement).
type sdkDebugFileErrorLine struct {
	Date            *time.Time  `json:"date,omitempty"`
	Level           string      `json:"level,omitempty"`
	Method          string      `json:"method,omitempty"`
	URL             string      `json:"url,omitempty"`
	RequestHeaders  http.Header `json:"requestHeaders,omitempty"`
	ResponseHeaders http.Header `json:"responseHeaders,omitempty"`
	CorrelationId   string      `json:"correlationId,omitempty"`
	StatusCode      int         `json:"statusCode,omitempty"`
	RequestBody     string      `json:"requestBody,omitempty"`
	ResponseBody    string      `json:"responseBody,omitempty"`
}

func cloneHTTPHeaderRedactAuth(src http.Header) http.Header {
	if src == nil {
		return nil
	}
	dst := make(http.Header, len(src))
	for k, vals := range src {
		nv := make([]string, len(vals))
		copy(nv, vals)
		dst[k] = nv
	}
	for k := range dst {
		if strings.EqualFold(k, "Authorization") {
			dst.Set(k, "[REDACTED]")
		}
	}
	return dst
}

func sdkDebugMirrorCorrelationID(h http.Header) string {
	for k, vals := range h {
		if strings.EqualFold(k, "inin-correlation-id") && len(vals) > 0 && vals[0] != "" {
			return vals[0]
		}
	}
	return ""
}

func mirrorRequestBodyForSDKFile(req *http.Request) string {
	if req == nil || !sdkDebugHookRequestBodyEnabled.Load() {
		return ""
	}
	if req.GetBody == nil {
		return ""
	}
	rc, err := req.GetBody()
	if err != nil || rc == nil {
		return ""
	}
	defer rc.Close()
	b, err := io.ReadAll(io.LimitReader(rc, maxMirrorBodyBytes))
	if err != nil {
		return ""
	}
	return string(b)
}

func mirrorResponseBodyForSDKFile(resp *http.Response) string {
	if resp == nil || resp.Body == nil || !sdkDebugHookRequestBodyEnabled.Load() {
		return ""
	}
	b, err := io.ReadAll(io.LimitReader(resp.Body, maxMirrorBodyBytes))
	if err != nil {
		return ""
	}
	resp.Body = io.NopCloser(bytes.NewReader(b))
	return string(b)
}

// mirrorSDKDebugHTTPErrorToFile appends one JSON line matching the SDK file format for ERROR logs when
// retryablehttp returns (nil, err) on retry exhaustion so platformclientv2 never calls LoggingConfiguration.error().
// storedRequestBody comes from popSDKDebugMirrorRequestBody(TF-Correlation-Id) in ResponseLogHook.
func mirrorSDKDebugHTTPErrorToFile(resp *http.Response, storedRequestBody string) {
	if resp == nil || resp.Request == nil || !sdkDebugStatusNeedsErrorMirror(resp.StatusCode) {
		return
	}
	v := sdkDebugErrorMirrorPath.Load()
	path, _ := v.(string)
	if path == "" {
		return
	}

	now := time.Now()
	reqURL := ""
	if resp.Request.URL != nil {
		reqURL = resp.Request.URL.String()
	}

	reqBody := storedRequestBody
	if reqBody == "" {
		reqBody = mirrorRequestBodyForSDKFile(resp.Request)
	}

	lineObj := sdkDebugFileErrorLine{
		Date:            &now,
		Level:           "error",
		Method:          resp.Request.Method,
		URL:             reqURL,
		RequestHeaders:  cloneHTTPHeaderRedactAuth(resp.Request.Header),
		ResponseHeaders: cloneHTTPHeaderRedactAuth(resp.Header),
		CorrelationId:   sdkDebugMirrorCorrelationID(resp.Header),
		StatusCode:      resp.StatusCode,
		RequestBody:     reqBody,
		ResponseBody:    mirrorResponseBodyForSDKFile(resp),
	}

	line, err := json.Marshal(lineObj)
	if err != nil {
		return
	}

	sdkDebugErrorMirrorMu.Lock()
	defer sdkDebugErrorMirrorMu.Unlock()

	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o666)
	if err != nil {
		return
	}
	defer f.Close()
	_, _ = f.Write(append(line, '\n'))
}

// maxRequestBodyGrowHint caps ContentLength-based preallocation (avoids huge Grow on bogus headers).
const maxRequestBodyGrowHint = 4 << 20

// readRequestBodyForHook reads r fully; uses a single buffer grow when Content-Length is trustworthy.
func readRequestBodyForHook(r io.Reader, contentLength int64) ([]byte, error) {
	if contentLength > 0 && contentLength <= maxRequestBodyGrowHint {
		var buf bytes.Buffer
		buf.Grow(int(contentLength))
		if _, err := buf.ReadFrom(r); err != nil {
			return nil, err
		}
		return buf.Bytes(), nil
	}
	return io.ReadAll(r)
}

type sdkDebugRequest struct {
	DebugType        string `json:"debug_type,omitempty"`        //Indicates whether it is a request or response debug
	TransactionId    string `json:"transaction_id,omitempty"`    //Unique id to link the request and response
	InvocationCount  int    `json:"invocation_count,omitempty"`  //Number of times the URL has been invoked.  Will be greater then zero when it is a retry
	InvocationMethod string `json:"invocation_method,omitempty"` //HTTP method that will be invoked
	InvocationUrl    string `json:"invocation_url,omitempty"`    //HTTP URL
	RequestBody      string `json:"request_body,omitempty"`      //HTTP request body
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

// setContextForRequest stores the context for the current goroutine.
// This allows the RequestLogHook to access the context even though
// the SDK doesn't pass context to HTTP requests.
func setContextForRequest(ctx context.Context) {
	if ctx == nil {
		return
	}

	// Store by goroutine ID
	goroutineID := getGoroutineID()
	if goroutineID > 0 {
		contextStorage.Store(goroutineID, ctx)
	}
}

// getContextForRequest retrieves the context for the current goroutine.
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

	// Extract name field if present (check top-level first)
	if nameVal, ok := jsonMap["name"]; ok && nameVal != nil {
		if nameStr, ok := nameVal.(string); ok && nameStr != "" {
			name = nameStr
		}
	}

	// If id or name not found at top level, check common nested structures
	if id == "" || name == "" {
		// Check for flow.id and flow.name (used in architect flow job responses)
		if flowVal, ok := jsonMap["flow"]; ok && flowVal != nil {
			if flowMap, ok := flowVal.(map[string]interface{}); ok {
				// Extract flow.id if we don't have an id yet
				if id == "" {
					if flowIdVal, ok := flowMap["id"]; ok && flowIdVal != nil {
						switch v := flowIdVal.(type) {
						case string:
							if v != "" {
								id = v
							}
						case float64:
							id = fmt.Sprintf("%.0f", v)
						case int:
							id = fmt.Sprintf("%d", v)
						case int64:
							id = fmt.Sprintf("%d", v)
						}
					}
				}
				// Extract flow.name if we don't have a name yet
				if name == "" {
					if flowNameVal, ok := flowMap["name"]; ok && flowNameVal != nil {
						if flowNameStr, ok := flowNameVal.(string); ok && flowNameStr != "" {
							name = flowNameStr
						}
					}
				}
			}
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

	captureFull := sdkDebugHookRequestBodyEnabled.Load()
	needsBodyPeek := debugReq.ResourceId == "" || debugReq.ResourceId == "unavailable" ||
		debugReq.ResourceName == "" || debugReq.ResourceName == "unavailable"

	// When sdk_debug is true, read the body for RequestBody in hook JSON (fixes missing body when SDK logging fails on retries).
	// When sdk_debug is false, only read if id/name are missing (legacy), and do not populate RequestBody (avoids huge/sensitive logs).
	var bodyBytes []byte
	if request.Body != nil && (captureFull || needsBodyPeek) {
		body, err := readRequestBodyForHook(request.Body, request.ContentLength)
		if err != nil {
			log.Printf("[WARN] sdk RequestLogHook: read request body: %v", err)
			request.Body = io.NopCloser(bytes.NewReader(nil))
			return debugReq
		}
		bodyBytes = body
		if captureFull && len(bodyBytes) > 0 {
			debugReq.RequestBody = string(bodyBytes)
		}
		request.Body = io.NopCloser(bytes.NewReader(bodyBytes))
	}

	if needsBodyPeek && len(bodyBytes) > 0 {
		extractedId, extractedName := extractIdOrNameFromJSON(bodyBytes)
		if (debugReq.ResourceId == "" || debugReq.ResourceId == "unavailable") && extractedId != "" {
			debugReq.ResourceId = extractedId
		}
		if (debugReq.ResourceName == "" || debugReq.ResourceName == "unavailable") && extractedName != "" {
			debugReq.ResourceName = extractedName
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
		InvocationRetryAfter: response.Header.Get("Retry-After"),
	}

	// Extract resource context from request context if available
	var resourceCtx *ResourceContext
	if ctx := response.Request.Context(); ctx != nil {
		if rc, ok := ctx.Value(resourceContextKey{}).(*ResourceContext); ok && rc != nil {
			resourceCtx = rc
		}
	}

	// Set resource context fields if found
	if resourceCtx != nil {
		debugResp.ResourceType = resourceCtx.ResourceType
		debugResp.ResourceId = resourceCtx.ResourceId
		debugResp.ResourceName = resourceCtx.ResourceName
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

	log.Printf("[DEBUG] WithResourceContext: resourceType=%q, resourceId=%q, resourceName=%q", resourceType, resourceId, resourceName)

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
// If the context already has resource metadata, it merges/updates values:
// - Updates ResourceType if provided and different
// - Preserves existing ResourceId and ResourceName if they are not "unavailable"
// - Sets ResourceId and ResourceName to "unavailable" only if they don't exist
// This is useful in proxy methods where ResourceData is not available.
// This function also stores the context in goroutine-local storage so it can be
// accessed by the SDK's RequestLogHook even though the SDK doesn't pass context to HTTP requests.
func EnsureResourceContext(ctx context.Context, resourceType string) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}

	// Get existing resource context if it exists
	var existingCtx *ResourceContext
	if rc, ok := ctx.Value(resourceContextKey{}).(*ResourceContext); ok && rc != nil {
		existingCtx = rc
	}

	// Determine the values to use
	var finalResourceType, finalResourceId, finalResourceName string

	if existingCtx != nil {
		// Use existing values, but update ResourceType if provided
		finalResourceType = resourceType
		if finalResourceType == "" {
			finalResourceType = existingCtx.ResourceType
		}

		// Preserve existing ResourceId and ResourceName if they're not "unavailable"
		if existingCtx.ResourceId != "" && existingCtx.ResourceId != "unavailable" {
			finalResourceId = existingCtx.ResourceId
		} else {
			finalResourceId = "unavailable"
		}

		if existingCtx.ResourceName != "" && existingCtx.ResourceName != "unavailable" {
			finalResourceName = existingCtx.ResourceName
		} else {
			finalResourceName = "unavailable"
		}
	} else {
		// No existing context, set defaults
		finalResourceType = resourceType
		finalResourceId = "unavailable"
		finalResourceName = "unavailable"
	}

	// Create or update the context with merged values
	ctx = WithResourceContext(ctx, finalResourceType, finalResourceId, finalResourceName)

	// Store the context for the RequestLogHook to access
	setContextForRequest(ctx)

	return ctx
}
