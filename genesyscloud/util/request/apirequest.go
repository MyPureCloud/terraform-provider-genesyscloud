package request

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/mypurecloud/platform-client-sdk-go/v171/platformclientv2"
)

type APIRequest[T any, U any] struct {
	setRequestHeader func(r *http.Request, p *U) *http.Request
}

func NewAPIRequest[T, U any](setRequestHeader func(r *http.Request, p *U) *http.Request) *APIRequest[T, U] {
	return &APIRequest[T, U]{
		setRequestHeader,
	}
}

// makeAPIRequest performs a complete API request for any of the guide endpoints
func (a *APIRequest[T, U]) MakeAPIRequest(ctx context.Context, method, url string, requestBody interface{}, p *U) (*T, *platformclientv2.APIResponse, error) {
	var req *http.Request
	var err error

	if requestBody != nil {
		req, err = a.MarshalAndCreateRequest(method, url, requestBody, p)
	} else {
		req, err = a.CreateHTTPRequest(method, url, nil, p)
	}

	if err != nil {
		return nil, nil, err
	}

	client := &http.Client{}
	respBody, resp, err := a.CallAPI(ctx, client, req)
	if err != nil {
		return nil, resp, err
	}

	var result T

	if err := a.UnmarshalResponse(respBody, &result); err != nil {
		return nil, resp, err
	}

	return &result, resp, nil
}

// CreateHTTPRequest creates a new HTTP request with proper headers
func (a *APIRequest[T, U]) CreateHTTPRequest(method, url string, body io.Reader, p *U) (*http.Request, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}
	req = a.setRequestHeader(req, p)
	return req, nil
}

// MarshalAndCreateRequest marshals a body to JSON and creates an HTTP request
func (a *APIRequest[T, U]) MarshalAndCreateRequest(method, url string, body interface{}, p *U) (*http.Request, error) {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request body: %v", err)
	}
	return a.CreateHTTPRequest(method, url, bytes.NewBuffer(jsonBody), p)
}

// UnmarshalResponse unmarshals a JSON response into the target struct
func (a *APIRequest[T, U]) UnmarshalResponse(respBody []byte, target interface{}) error {
	if err := json.Unmarshal(respBody, target); err != nil {
		return fmt.Errorf("error unmarshaling response: %v", err)
	}
	return nil
}

// CallAPI is a helper function which will be removed when the endpoints are public
func (a *APIRequest[T, U]) CallAPI(ctx context.Context, client *http.Client, req *http.Request) ([]byte, *platformclientv2.APIResponse, error) {
	req = req.WithContext(ctx)

	resp, err := client.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("error making request: %v", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("error reading response: %v", err)
	}

	response := &platformclientv2.APIResponse{
		StatusCode: resp.StatusCode,
		Response:   resp,
	}

	if resp.StatusCode >= 400 {
		return nil, response, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	return respBody, response, nil
}
