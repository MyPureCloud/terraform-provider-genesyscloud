package util

import (
	"encoding/json"
	"github.com/mypurecloud/platform-client-sdk-go/v123/platformclientv2"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/url"
	"testing"
)

func TestUnitTestAPIResponseDiagWithGoodApiResponse(t *testing.T) {
	resource := "genesyscloud_tf_exporter"
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
	targetResponse := "{\"resourceName\":\"genesyscloud_tf_exporter\",\"method\":\"POST\",\"path:omitempty\":\"/api/v2/tfexporter?test=123\",\"statusCode:omitempty\":500,\"errorMessage\":\"DummyError\",\"correlationId\":\"e03b48a1-7063-4ae2-921a-f64c8e02702b\"}"
	json.Unmarshal([]byte(targetResponse), targetDiag)

	diag := BuildAPIDiagnosticError(resource, sumErrMsg, apiResponse)
	actualDiag := &detailedDiagnosticInfo{}
	json.Unmarshal([]byte(diag[0].Detail), actualDiag)

	assert.Equal(t, targetDiag.CorrelationID, actualDiag.CorrelationID)
	assert.Equal(t, targetDiag.Method, actualDiag.Method)
	assert.Equal(t, targetDiag.StatusCode, actualDiag.StatusCode)
	assert.Equal(t, targetDiag.ErrorMessage, actualDiag.ErrorMessage)
	assert.Equal(t, targetDiag.ResourceName, actualDiag.ResourceName)
}

func TestUnitTestAPIResponseDiagWithBadApiResponse(t *testing.T) {
	resource := "genesyscloud_tf_exporter"
	sumErrMsg := "This is a dummy error message"
	apiErrorMsg := "DummyError"

	apiResponse := &platformclientv2.APIResponse{
		Response:      nil,
		StatusCode:    http.StatusInternalServerError,
		ErrorMessage:  apiErrorMsg,
		CorrelationID: "e03b48a1-7063-4ae2-921a-f64c8e02702b",
	}

	targetDiag := &detailedDiagnosticInfo{}
	targetResponse := "{\"resourceName\":\"genesyscloud_tf_exporter\",\"path:omitempty\":\"\",\"statusCode:omitempty\":0,\"errorMessage\":\"Unable to build a message from the response because the APIResponse does not contain the appropriate data.\"}"
	json.Unmarshal([]byte(targetResponse), targetDiag)

	diag := BuildAPIDiagnosticError(resource, sumErrMsg, apiResponse)
	actualDiag := &detailedDiagnosticInfo{}
	json.Unmarshal([]byte(diag[0].Detail), actualDiag)

	assert.Equal(t, targetDiag.ResourceName, actualDiag.ResourceName)
	assert.Equal(t, diag[0].Summary, sumErrMsg)
	assert.Equal(t, diag[0].Detail, targetResponse)
}
