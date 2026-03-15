package custom_api_client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"testing"

	"github.com/mypurecloud/platform-client-sdk-go/v179/platformclientv2"
	"github.com/stretchr/testify/assert"
)

// helper to build a Client with a mocked callAPIAttr
func newTestClient(mockFn callAPIFunc) *Client {
	config := &platformclientv2.Configuration{
		BasePath:      "https://api.example.com",
		AccessToken:   "test-token",
		DefaultHeader: map[string]string{"X-Custom": "custom-value"},
	}
	return &Client{
		config:       config,
		resourceType: "genesyscloud_test_resource",
		callAPIAttr:  mockFn,
	}
}

func TestNewClient(t *testing.T) {
	config := platformclientv2.GetDefaultConfiguration()
	config.BasePath = "https://api.example.com"
	config.AccessToken = "test-token"

	client := NewClient(config, "genesyscloud_test_resource")

	assert.NotNil(t, client)
	assert.Equal(t, config, client.Config())
	assert.Equal(t, "genesyscloud_test_resource", client.resourceType)
	assert.NotNil(t, client.callAPIAttr)
}

func TestBuildHeaders(t *testing.T) {
	config := &platformclientv2.Configuration{
		AccessToken:   "my-token",
		DefaultHeader: map[string]string{"X-Custom": "value"},
	}
	client := &Client{config: config}

	headers := client.buildHeaders()

	assert.Equal(t, "Bearer my-token", headers["Authorization"])
	assert.Equal(t, "application/json", headers["Content-Type"])
	assert.Equal(t, "application/json", headers["Accept"])
	assert.Equal(t, "value", headers["X-Custom"])
}

func TestBuildHeadersNoToken(t *testing.T) {
	config := &platformclientv2.Configuration{
		DefaultHeader: map[string]string{},
	}
	client := &Client{config: config}

	headers := client.buildHeaders()

	_, hasAuth := headers["Authorization"]
	assert.False(t, hasAuth)
}

func TestBuildPath(t *testing.T) {
	config := &platformclientv2.Configuration{BasePath: "https://api.example.com"}
	client := &Client{config: config}

	// No query params
	assert.Equal(t, "https://api.example.com/api/v2/things", client.buildPath("/api/v2/things", nil))

	// Single value params
	params := url.Values{"pageSize": {"100"}, "pageNumber": {"1"}}
	path := client.buildPath("/api/v2/things", params)
	assert.Contains(t, path, "https://api.example.com/api/v2/things?")
	assert.Contains(t, path, "pageSize=100")
	assert.Contains(t, path, "pageNumber=1")

	// Multi-value params
	multiParams := url.Values{"type": {"inboundcall", "outboundcall"}}
	multiPath := client.buildPath("/api/v2/flows", multiParams)
	assert.Contains(t, multiPath, "type=inboundcall")
	assert.Contains(t, multiPath, "type=outboundcall")
}

func TestDoSuccess(t *testing.T) {
	type TestResponse struct {
		Id   string `json:"id"`
		Name string `json:"name"`
	}

	expectedBody, _ := json.Marshal(TestResponse{Id: "123", Name: "test"})

	mockFn := func(path, method string, postBody interface{}, headerParams, queryParams map[string]string, formParams url.Values, fileName string, fileBytes []byte, pathName string) (*platformclientv2.APIResponse, error) {
		assert.Equal(t, "https://api.example.com/api/v2/things", path)
		assert.Equal(t, MethodPost, method)
		assert.Equal(t, "Bearer test-token", headerParams["Authorization"])
		assert.Equal(t, "application/json", headerParams["Content-Type"])
		assert.Equal(t, "custom-value", headerParams["X-Custom"])
		assert.Nil(t, queryParams) // query params encoded in path
		return &platformclientv2.APIResponse{
			RawBody:    expectedBody,
			StatusCode: 200,
		}, nil
	}

	client := newTestClient(mockFn)
	result, resp, err := Do[TestResponse](context.Background(), client, MethodPost, "/api/v2/things", map[string]string{"key": "val"}, nil)

	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, "123", result.Id)
	assert.Equal(t, "test", result.Name)
}

func TestDoCallAPIError(t *testing.T) {
	type TestResponse struct{}

	mockFn := func(path, method string, postBody interface{}, headerParams, queryParams map[string]string, formParams url.Values, fileName string, fileBytes []byte, pathName string) (*platformclientv2.APIResponse, error) {
		return nil, fmt.Errorf("connection refused")
	}

	client := newTestClient(mockFn)
	result, _, err := Do[TestResponse](context.Background(), client, MethodGet, "/api/v2/things", nil, nil)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "connection refused")
}

func TestDoResponseError(t *testing.T) {
	type TestResponse struct{}

	mockFn := func(path, method string, postBody interface{}, headerParams, queryParams map[string]string, formParams url.Values, fileName string, fileBytes []byte, pathName string) (*platformclientv2.APIResponse, error) {
		return &platformclientv2.APIResponse{
			StatusCode:   404,
			Error:        &platformclientv2.APIError{Message: "not found"},
			ErrorMessage: "Resource not found",
		}, nil
	}

	client := newTestClient(mockFn)
	result, resp, err := Do[TestResponse](context.Background(), client, MethodGet, "/api/v2/things/123", nil, nil)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, 404, resp.StatusCode)
	assert.Contains(t, err.Error(), "Resource not found")
}

func TestDoUnmarshalError(t *testing.T) {
	type TestResponse struct {
		Id int `json:"id"`
	}

	mockFn := func(path, method string, postBody interface{}, headerParams, queryParams map[string]string, formParams url.Values, fileName string, fileBytes []byte, pathName string) (*platformclientv2.APIResponse, error) {
		return &platformclientv2.APIResponse{
			RawBody:    []byte(`not valid json`),
			StatusCode: 200,
		}, nil
	}

	client := newTestClient(mockFn)
	result, _, err := Do[TestResponse](context.Background(), client, MethodGet, "/api/v2/things", nil, nil)

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestDoQueryParams(t *testing.T) {
	type TestResponse struct{}

	mockFn := func(path, method string, postBody interface{}, headerParams, queryParams map[string]string, formParams url.Values, fileName string, fileBytes []byte, pathName string) (*platformclientv2.APIResponse, error) {
		// Query params are encoded in the path, not passed separately
		assert.Nil(t, queryParams)
		assert.Contains(t, path, "pageSize=100")
		assert.Contains(t, path, "pageNumber=1")
		return &platformclientv2.APIResponse{
			RawBody:    []byte(`{}`),
			StatusCode: 200,
		}, nil
	}

	client := newTestClient(mockFn)
	params := url.Values{"pageSize": {"100"}, "pageNumber": {"1"}}
	_, _, err := Do[TestResponse](context.Background(), client, MethodGet, "/api/v2/things", nil, params)

	assert.NoError(t, err)
}

func TestDoMultiValueQueryParams(t *testing.T) {
	type TestResponse struct{}

	mockFn := func(path, method string, postBody interface{}, headerParams, queryParams map[string]string, formParams url.Values, fileName string, fileBytes []byte, pathName string) (*platformclientv2.APIResponse, error) {
		assert.Nil(t, queryParams)
		assert.Contains(t, path, "type=inboundcall")
		assert.Contains(t, path, "type=outboundcall")
		return &platformclientv2.APIResponse{
			RawBody:    []byte(`{}`),
			StatusCode: 200,
		}, nil
	}

	client := newTestClient(mockFn)
	params := url.Values{"type": {"inboundcall", "outboundcall"}}
	_, _, err := Do[TestResponse](context.Background(), client, MethodGet, "/api/v2/flows", nil, params)

	assert.NoError(t, err)
}

func TestDoNoResponseSuccess(t *testing.T) {
	mockFn := func(path, method string, postBody interface{}, headerParams, queryParams map[string]string, formParams url.Values, fileName string, fileBytes []byte, pathName string) (*platformclientv2.APIResponse, error) {
		assert.Equal(t, "https://api.example.com/api/v2/things/123", path)
		assert.Equal(t, MethodDelete, method)
		return &platformclientv2.APIResponse{StatusCode: 204}, nil
	}

	client := newTestClient(mockFn)
	resp, err := DoNoResponse(context.Background(), client, MethodDelete, "/api/v2/things/123", nil, nil)

	assert.NoError(t, err)
	assert.Equal(t, 204, resp.StatusCode)
}

func TestDoNoResponseCallAPIError(t *testing.T) {
	mockFn := func(path, method string, postBody interface{}, headerParams, queryParams map[string]string, formParams url.Values, fileName string, fileBytes []byte, pathName string) (*platformclientv2.APIResponse, error) {
		return nil, fmt.Errorf("timeout")
	}

	client := newTestClient(mockFn)
	_, err := DoNoResponse(context.Background(), client, MethodDelete, "/api/v2/things/123", nil, nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "timeout")
}

func TestDoNoResponseResponseError(t *testing.T) {
	mockFn := func(path, method string, postBody interface{}, headerParams, queryParams map[string]string, formParams url.Values, fileName string, fileBytes []byte, pathName string) (*platformclientv2.APIResponse, error) {
		return &platformclientv2.APIResponse{
			StatusCode:   409,
			Error:        &platformclientv2.APIError{Message: "conflict"},
			ErrorMessage: "Version conflict",
		}, nil
	}

	client := newTestClient(mockFn)
	resp, err := DoNoResponse(context.Background(), client, MethodDelete, "/api/v2/things/123", nil, nil)

	assert.Error(t, err)
	assert.Equal(t, 409, resp.StatusCode)
	assert.Contains(t, err.Error(), "Version conflict")
}

func TestDoRawSuccess(t *testing.T) {
	expectedJSON := []byte(`{"id":"123","deleted":true}`)

	mockFn := func(path, method string, postBody interface{}, headerParams, queryParams map[string]string, formParams url.Values, fileName string, fileBytes []byte, pathName string) (*platformclientv2.APIResponse, error) {
		return &platformclientv2.APIResponse{
			RawBody:    expectedJSON,
			StatusCode: 200,
		}, nil
	}

	client := newTestClient(mockFn)
	raw, resp, err := DoRaw(context.Background(), client, MethodGet, "/api/v2/things/123", nil, nil)

	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, expectedJSON, raw)
}

func TestDoRawCallAPIError(t *testing.T) {
	mockFn := func(path, method string, postBody interface{}, headerParams, queryParams map[string]string, formParams url.Values, fileName string, fileBytes []byte, pathName string) (*platformclientv2.APIResponse, error) {
		return nil, fmt.Errorf("network error")
	}

	client := newTestClient(mockFn)
	raw, _, err := DoRaw(context.Background(), client, MethodGet, "/api/v2/things/123", nil, nil)

	assert.Error(t, err)
	assert.Nil(t, raw)
}

func TestDoRawResponseError(t *testing.T) {
	mockFn := func(path, method string, postBody interface{}, headerParams, queryParams map[string]string, formParams url.Values, fileName string, fileBytes []byte, pathName string) (*platformclientv2.APIResponse, error) {
		return &platformclientv2.APIResponse{
			StatusCode:   500,
			Error:        &platformclientv2.APIError{Message: "server error"},
			ErrorMessage: "Internal server error",
		}, nil
	}

	client := newTestClient(mockFn)
	raw, resp, err := DoRaw(context.Background(), client, MethodGet, "/api/v2/things/123", nil, nil)

	assert.Error(t, err)
	assert.Nil(t, raw)
	assert.Equal(t, 500, resp.StatusCode)
}

func TestDoWithAcceptHeaderSuccess(t *testing.T) {
	expectedBody := []byte(`template content here`)

	mockFn := func(path, method string, postBody interface{}, headerParams, queryParams map[string]string, formParams url.Values, fileName string, fileBytes []byte, pathName string) (*platformclientv2.APIResponse, error) {
		assert.Equal(t, "*/*", headerParams["Accept"])
		assert.Equal(t, "application/json", headerParams["Content-Type"])
		return &platformclientv2.APIResponse{
			RawBody:    expectedBody,
			StatusCode: 200,
		}, nil
	}

	client := newTestClient(mockFn)
	raw, resp, err := DoWithAcceptHeader(context.Background(), client, MethodGet, "/api/v2/things/123/template", nil, nil, "*/*")

	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, expectedBody, raw)
}

func TestDoWithAcceptHeaderCallAPIError(t *testing.T) {
	mockFn := func(path, method string, postBody interface{}, headerParams, queryParams map[string]string, formParams url.Values, fileName string, fileBytes []byte, pathName string) (*platformclientv2.APIResponse, error) {
		return nil, fmt.Errorf("connection error")
	}

	client := newTestClient(mockFn)
	raw, _, err := DoWithAcceptHeader(context.Background(), client, MethodGet, "/api/v2/things/123/template", nil, nil, "*/*")

	assert.Error(t, err)
	assert.Nil(t, raw)
}

func TestConstants(t *testing.T) {
	assert.Equal(t, "GET", MethodGet)
	assert.Equal(t, "POST", MethodPost)
	assert.Equal(t, "PUT", MethodPut)
	assert.Equal(t, "PATCH", MethodPatch)
	assert.Equal(t, "DELETE", MethodDelete)
}
