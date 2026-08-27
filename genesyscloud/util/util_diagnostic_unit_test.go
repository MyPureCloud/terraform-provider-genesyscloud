package util

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/mypurecloud/platform-client-sdk-go/v195/platformclientv2"
	"github.com/stretchr/testify/assert"
)

func TestUnitTestAPIResponseDiagWithGoodApiResponse(t *testing.T) {
	resourceType := "genesyscloud_tf_exporter"
	sumErrMsg := "This is a dummy error message"
	apiErrorMsg := "DummyError"
	path := "/api/v2/tfexporter?test=123"
	url := &url.URL{
		Path: path,
	}
	request := &http.Request{
		Method: "POST",
		URL:    url,
	}

	response := &http.Response{
		Request: request,
	}

	apiResponse := &platformclientv2.APIResponse{
		Response:      response,
		StatusCode:    http.StatusInternalServerError,
		ErrorMessage:  apiErrorMsg,
		CorrelationID: "e03b48a1-7063-4ae2-921a-f64c8e02702b",
	}

	targetDiag := &detailedDiagnosticInfo{}
	targetResponse := "{\"resourceType\":\"genesyscloud_tf_exporter\",\"method\":\"POST\",\"path\":\"/api/v2/tfexporter?test=123\",\"statusCode\":500,\"errorMessage\":\"DummyError\",\"correlationId\":\"e03b48a1-7063-4ae2-921a-f64c8e02702b\"}"
	_ = json.Unmarshal([]byte(targetResponse), targetDiag)
	diag := BuildAPIDiagnosticError(resourceType, sumErrMsg, apiResponse)

	actualDiag := &detailedDiagnosticInfo{}
	_ = json.Unmarshal([]byte(diag[0].Detail), actualDiag)

	assert.Equal(t, targetDiag.CorrelationID, actualDiag.CorrelationID)
	assert.Equal(t, targetDiag.Method, actualDiag.Method)
	assert.Equal(t, targetDiag.StatusCode, actualDiag.StatusCode)
	assert.Equal(t, targetDiag.ErrorMessage, actualDiag.ErrorMessage)
	assert.Equal(t, targetDiag.ResourceType, actualDiag.ResourceType)
}

func TestUnitTestAPIResponseDiagWithBadApiResponse(t *testing.T) {
	resourceType := "genesyscloud_tf_exporter"
	sumErrMsg := "This is a dummy error message"
	apiErrorMsg := "DummyError"

	apiResponse := &platformclientv2.APIResponse{
		Response:      nil,
		StatusCode:    http.StatusInternalServerError,
		ErrorMessage:  apiErrorMsg,
		CorrelationID: "e03b48a1-7063-4ae2-921a-f64c8e02702b",
	}

	targetDiag := &detailedDiagnosticInfo{}
	targetResponse := "{\"resourceType\":\"genesyscloud_tf_exporter\",\"statusCode\":500,\"errorMessage\":\"DummyError\",\"correlationId\":\"e03b48a1-7063-4ae2-921a-f64c8e02702b\"}"
	json.Unmarshal([]byte(targetResponse), targetDiag)

	diag := BuildAPIDiagnosticError(resourceType, sumErrMsg, apiResponse)
	actualDiag := &detailedDiagnosticInfo{}
	_ = json.Unmarshal([]byte(diag[0].Detail), actualDiag)

	assert.Equal(t, targetDiag.ResourceType, actualDiag.ResourceType)
	assert.Equal(t, diag[0].Summary, sumErrMsg)
	assert.Equal(t, targetResponse, diag[0].Detail)
}

func TestUnitTestAPIResponseWithRetriesDiagWithGoodAPIResponse(t *testing.T) {
	resourceType := "genesyscloud_tf_exporter"
	sumErrMsg := "This is a dummy error message"
	apiErrorMsg := "DummyError"
	path := "/api/v2/tfexporter?test=123"
	url := &url.URL{
		Path: path,
	}
	request := &http.Request{
		Method: "POST",
		URL:    url,
	}

	response := &http.Response{
		Request: request,
	}

	apiResponse := &platformclientv2.APIResponse{
		Response:      response,
		StatusCode:    http.StatusInternalServerError,
		ErrorMessage:  apiErrorMsg,
		CorrelationID: "e03b48a1-7063-4ae2-921a-f64c8e02702b",
	}

	targetDiag := &detailedDiagnosticInfo{}
	targetResponse := "{\"resourceType\":\"genesyscloud_tf_exporter\",\"method\":\"POST\",\"path\":\"/api/v2/tfexporter?test=123\",\"statusCode\":500,\"errorMessage\":\"DummyError\",\"correlationId\":\"e03b48a1-7063-4ae2-921a-f64c8e02702b\"}"
	_ = json.Unmarshal([]byte(targetResponse), targetDiag)

	diag := BuildWithRetriesApiDiagnosticError(resourceType, sumErrMsg, apiResponse)
	actualDiag := &detailedDiagnosticInfo{}

	lines := strings.Split(diag.Error(), "\n")[1]
	_ = json.Unmarshal([]byte(lines), actualDiag)

	assert.Equal(t, targetDiag.CorrelationID, actualDiag.CorrelationID)
	assert.Equal(t, targetDiag.Method, actualDiag.Method)
	assert.Equal(t, targetDiag.StatusCode, actualDiag.StatusCode)
	assert.Equal(t, targetDiag.ErrorMessage, actualDiag.ErrorMessage)
	assert.Equal(t, targetDiag.ResourceType, actualDiag.ResourceType)
}

func TestUnitTestAPIResponseWithRetriesDiagWithBadApiResponse(t *testing.T) {
	resourceType := "genesyscloud_tf_exporter"
	sumErrMsg := "This is a dummy error message"
	apiErrorMsg := "DummyError"

	apiResponse := &platformclientv2.APIResponse{
		Response:      nil,
		StatusCode:    http.StatusInternalServerError,
		ErrorMessage:  apiErrorMsg,
		CorrelationID: "e03b48a1-7063-4ae2-921a-f64c8e02702b",
	}

	targetDiag := &detailedDiagnosticInfo{}
	targetResponse := "{\"resourceType\":\"genesyscloud_tf_exporter\",\"statusCode\":500,\"errorMessage\":\"DummyError\",\"correlationId\":\"e03b48a1-7063-4ae2-921a-f64c8e02702b\"}"
	_ = json.Unmarshal([]byte(targetResponse), targetDiag)

	diag := BuildWithRetriesApiDiagnosticError(resourceType, sumErrMsg, apiResponse)
	actualDiag := &detailedDiagnosticInfo{}

	lines := strings.Split(diag.Error(), "\n")
	_ = json.Unmarshal([]byte(lines[1]), actualDiag)

	assert.Equal(t, targetDiag.ResourceType, actualDiag.ResourceType)
	assert.Equal(t, sumErrMsg, lines[0])
	assert.Equal(t, targetResponse, lines[1])
}

func TestUnitTestCodedAPIDiagnosticErrorIncludesCodeAndJobID(t *testing.T) {
	resourceType := "genesyscloud_flow"
	sumErrMsg := "flow publish failed. JobID: job-123, flowName: test"
	jobId := "job-123"
	apiResponse := &platformclientv2.APIResponse{
		StatusCode:    http.StatusOK,
		ErrorMessage:  "",
		CorrelationID: "corr-1",
	}

	diags := BuildCodedAPIDiagnosticError(resourceType, DiagnosticCodeFlowPublishFailed, sumErrMsg, jobId, apiResponse)
	assert.Len(t, diags, 1)
	assert.Equal(t, sumErrMsg, diags[0].Summary)

	actual := &detailedDiagnosticInfo{}
	err := json.Unmarshal([]byte(diags[0].Detail), actual)
	assert.NoError(t, err)
	assert.Equal(t, DiagnosticCodeFlowPublishFailed, actual.Code)
	assert.Equal(t, jobId, actual.JobID)
	assert.Equal(t, resourceType, actual.ResourceType)
}

func TestUnitTestCodedWithRetriesApiDiagnosticErrorIncludesCodeAndJobID(t *testing.T) {
	resourceType := "genesyscloud_flow"
	sumErrMsg := "Job (job-456) could not finish in 16 minutes and timed out "
	jobId := "job-456"
	apiResponse := &platformclientv2.APIResponse{
		StatusCode:   http.StatusOK,
		ErrorMessage: "",
	}

	err := BuildCodedWithRetriesApiDiagnosticError(resourceType, DiagnosticCodeFlowPublishTimeout, sumErrMsg, jobId, apiResponse)
	assert.Error(t, err)

	lines := strings.Split(err.Error(), "\n")
	assert.Equal(t, sumErrMsg, lines[0])

	actual := &detailedDiagnosticInfo{}
	assert.NoError(t, json.Unmarshal([]byte(lines[1]), actual))
	assert.Equal(t, DiagnosticCodeFlowPublishTimeout, actual.Code)
	assert.Equal(t, jobId, actual.JobID)
	assert.Equal(t, resourceType, actual.ResourceType)
}
