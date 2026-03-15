// Package custom_api_client provides a lightweight wrapper around the Platform SDK's
// APIClient.CallAPI() method, eliminating the boilerplate required when calling Genesys
// Cloud API endpoints that don't have generated SDK methods.
//
// # When to use this package
//
// Use custom_api_client when a proxy function needs to call a Platform API endpoint
// that is not available as a typed method on the SDK's API classes (e.g., RoutingApi,
// UsersApi). This typically happens when:
//   - The SDK hasn't been updated to include a new endpoint yet
//   - The endpoint requires query parameter combinations the SDK doesn't support
//   - You need raw byte access to the response body
//
// Do NOT use this package for:
//   - Endpoints that already have SDK methods — use the SDK directly
//   - S3 uploads or presigned URL flows — use net/http directly
//
// # Available functions
//
//   - Do[T]             — Generic typed response (JSON unmarshaled into T)
//   - DoNoResponse      — No response body expected (DELETE, PUT with no return)
//   - DoRaw             — Returns raw []byte response for custom unmarshaling
//   - DoWithAcceptHeader — Custom Accept header (e.g., "text/csv")
//
// # Usage in a proxy struct
//
// Add a customApiClient field to the proxy struct and initialize it in the constructor:
//
//	import customapi "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/custom_api_client"
//
//	type myProxy struct {
//	    clientConfig    *platformclientv2.Configuration
//	    myApi           *platformclientv2.MyApi
//	    customApiClient *customapi.Client
//	}
//
//	func newMyProxy(clientConfig *platformclientv2.Configuration) *myProxy {
//	    return &myProxy{
//	        clientConfig:    clientConfig,
//	        myApi:           platformclientv2.NewMyApiWithConfig(clientConfig),
//	        customApiClient: customapi.NewClient(clientConfig, ResourceType),
//	    }
//	}
//
// Then call from a proxy function:
//
//	result, resp, err := customapi.Do[platformclientv2.MyEntity](ctx, p.customApiClient, customapi.MethodGet, "/api/v2/my/endpoint", nil, nil)
//
// # Usage in a standalone function (no proxy struct)
//
//	c := customapi.NewClient(api.Configuration, ResourceType)
//	_, err := customapi.DoNoResponse(ctx, c, customapi.MethodDelete, "/api/v2/things/"+id, nil, nil)
//
// # Query parameters
//
// Use NewQueryParams for simple key-value pairs, or QueryParams directly for multi-value keys:
//
//	// Simple
//	qp := customapi.NewQueryParams(map[string]string{"pageSize": "100"})
//
//	// Multi-value (e.g., ?type=inboundcall&type=outboundcall)
//	qp := customapi.QueryParams{}
//	qp.Add("type", "inboundcall")
//	qp.Add("type", "outboundcall")
//
// # Testing
//
// The Client uses a function attribute (callAPIAttr) for the underlying SDK call,
// following the same pattern as proxy files. Unit tests within this package inject
// a mock callAPIAttr directly. See custom_api_client_test.go for examples.
package custom_api_client

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"

	"github.com/mypurecloud/platform-client-sdk-go/v179/platformclientv2"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
)

// callAPIFunc is the function signature matching the SDK's APIClient.CallAPI method.
type callAPIFunc func(path, method string, postBody interface{}, headerParams, queryParams map[string]string, formParams url.Values, fileName string, fileBytes []byte, pathName string) (*platformclientv2.APIResponse, error)

// Client wraps the SDK's APIClient.CallAPI() to eliminate boilerplate
// for custom platform API calls not covered by generated SDK methods.
type Client struct {
	config       *platformclientv2.Configuration
	resourceType string
	callAPIAttr  callAPIFunc
}

// NewClient creates a new custom API client.
func NewClient(config *platformclientv2.Configuration, resourceType string) *Client {
	return &Client{
		config:       config,
		resourceType: resourceType,
		callAPIAttr:  config.APIClient.CallAPI,
	}
}

// Config returns the underlying SDK configuration.
func (c *Client) Config() *platformclientv2.Configuration {
	return c.config
}

// buildHeaders constructs the standard header map from the SDK configuration.
func (c *Client) buildHeaders() map[string]string {
	headers := make(map[string]string)
	for key := range c.config.DefaultHeader {
		headers[key] = c.config.DefaultHeader[key]
	}
	if c.config.AccessToken != "" {
		headers["Authorization"] = "Bearer " + c.config.AccessToken
	}
	headers["Content-Type"] = "application/json"
	headers["Accept"] = "application/json"
	return headers
}

// buildPath constructs the full URL path with query parameters encoded.
// We encode query params into the path ourselves and pass nil to CallAPI's queryParams
// so that url.Values (which supports multi-value keys) works correctly.
func (c *Client) buildPath(path string, queryParams url.Values) string {
	fullPath := c.config.BasePath + path
	if len(queryParams) > 0 {
		fullPath += "?" + queryParams.Encode()
	}
	return fullPath
}

// Do makes a platform API call and unmarshals the response into T.
// path is relative, e.g. "/api/v2/processAutomation/triggers".
// body is passed directly to CallAPI as postBody (can be nil for GET/DELETE).
// queryParams can be nil. Supports multi-value keys via url.Values.
func Do[T any](ctx context.Context, c *Client, method, path string, body interface{}, queryParams url.Values) (*T, *platformclientv2.APIResponse, error) {
	ctx = provider.EnsureResourceContext(ctx, c.resourceType)

	fullPath := c.buildPath(path, queryParams)
	headers := c.buildHeaders()

	response, err := c.callAPIAttr(fullPath, method, body, headers, nil, nil, "", nil, "")
	if err != nil {
		return nil, response, err
	}
	if response.Error != nil {
		return nil, response, errors.New(response.ErrorMessage)
	}

	var result T
	if err := json.Unmarshal(response.RawBody, &result); err != nil {
		return nil, response, err
	}
	return &result, response, nil
}

// DoNoResponse makes a platform API call that returns no body (e.g. DELETE).
func DoNoResponse(ctx context.Context, c *Client, method, path string, body interface{}, queryParams url.Values) (*platformclientv2.APIResponse, error) {
	ctx = provider.EnsureResourceContext(ctx, c.resourceType)

	fullPath := c.buildPath(path, queryParams)
	headers := c.buildHeaders()

	response, err := c.callAPIAttr(fullPath, method, body, headers, nil, nil, "", nil, "")
	if err != nil {
		return response, err
	}
	if response.Error != nil {
		return response, errors.New(response.ErrorMessage)
	}
	return response, nil
}

// DoRaw makes a platform API call and returns the raw response body as bytes.
// Useful when the caller needs to inspect the raw JSON (e.g. checking for a "deleted" field
// that gets stripped by SDK type unmarshaling).
func DoRaw(ctx context.Context, c *Client, method, path string, body interface{}, queryParams url.Values) ([]byte, *platformclientv2.APIResponse, error) {
	ctx = provider.EnsureResourceContext(ctx, c.resourceType)

	fullPath := c.buildPath(path, queryParams)
	headers := c.buildHeaders()

	response, err := c.callAPIAttr(fullPath, method, body, headers, nil, nil, "", nil, "")
	if err != nil {
		return nil, response, err
	}
	if response.Error != nil {
		return nil, response, errors.New(response.ErrorMessage)
	}
	return response.RawBody, response, nil
}

// DoWithAcceptHeader makes a platform API call with a custom Accept header.
// Useful for endpoints that return non-JSON content (e.g. templates returning text/*).
func DoWithAcceptHeader(ctx context.Context, c *Client, method, path string, body interface{}, queryParams url.Values, accept string) ([]byte, *platformclientv2.APIResponse, error) {
	ctx = provider.EnsureResourceContext(ctx, c.resourceType)

	fullPath := c.buildPath(path, queryParams)
	headers := c.buildHeaders()
	headers["Accept"] = accept

	response, err := c.callAPIAttr(fullPath, method, body, headers, nil, nil, "", nil, "")
	if err != nil {
		return nil, response, err
	}
	if response.Error != nil {
		return nil, response, errors.New(response.ErrorMessage)
	}
	return response.RawBody, response, nil
}

// MethodGet and friends are convenience constants so callers don't need to import net/http.
const (
	MethodGet    = http.MethodGet
	MethodPost   = http.MethodPost
	MethodPut    = http.MethodPut
	MethodPatch  = http.MethodPatch
	MethodDelete = http.MethodDelete
)

// QueryParams is an alias for url.Values so callers don't need to import net/url.
type QueryParams = url.Values

// NewQueryParams builds a QueryParams from simple key-value pairs.
// For multi-value keys, use the returned QueryParams.Add() method.
func NewQueryParams(params map[string]string) QueryParams {
	qp := make(url.Values, len(params))
	for k, v := range params {
		qp.Set(k, v)
	}
	return qp
}
