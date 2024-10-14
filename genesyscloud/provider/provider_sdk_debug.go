package provider

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"net/http"
)

type sdkDebugRequest struct {
	DebugType        string `json:"debug_type,omitempty"`        //Indicates whether it is a request or response debug
	TransactionId    string `json:"transaction_id,omitempty"`    //Unique id to link the request and response
	InvocationCount  int    `json:"invocation_count,omitempty"`  //Number of times the URL has been invoked.  Will be greater then zero when it is a retry
	InvocationMethod string `json:"invocation_method,omitempty"` //HTTP method that will be invoked
	InvocationUrl    string `json:"invocation_url,omitempty"`    //HTTP URL
}

func (s *sdkDebugRequest) ToJSON() (err error, jsonStr string) {
	jsonData, err := json.Marshal(s)
	if err != nil {
		fmt.Println(err)
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
}

func (s *sdkDebugResponse) ToJSON() (err error, jsonStr string) {
	jsonData, err := json.Marshal(s)
	if err != nil {
		fmt.Println(err)
		return err, ""
	}

	// Print the JSON string
	return nil, string(jsonData)
}

func newSDKDebugRequest(request *http.Request, count int) *sdkDebugRequest {
	transactionId := uuid.NewString()
	return &sdkDebugRequest{

		DebugType:        "SDK DEBUG REQUEST",
		TransactionId:    transactionId,
		InvocationCount:  count,
		InvocationMethod: request.Method,
		InvocationUrl:    request.URL.Path,
	}
}

func newSDKDebugResponse(response *http.Response) *sdkDebugResponse {
	transactionId := response.Request.Header.Get("TF-Correlation-Id")
	return &sdkDebugResponse{

		DebugType:            "SDK DEBUG RESPONSE",
		TransactionId:        transactionId,
		InvocationCount:      0,
		InvocationMethod:     response.Request.Method,
		InvocationUrl:        response.Request.URL.Path,
		InvocationStatusCode: response.StatusCode,
		InvocationRetryAfter: response.Request.Header.Get("Retry-After"),
	}
}
