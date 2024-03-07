package util

import (
	"encoding/json"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/mypurecloud/platform-client-sdk-go/v123/platformclientv2"
)

type apiResponseWrapper struct {
	ResourceName  string `json:"resourceName,omitempty"`
	Method        string `json:"method,omitempty"`
	Path          string `json:"path:omitempty"`
	StatusCode    int    `json:"statusCode:omitempty"`
	ErrorMessage  string `json:"errorMessage,omitempty"`
	CorrelationID string `json:"correlationId,omitempty"`
}

func convertResponseToWrapper(resourceName string, apiResponse *platformclientv2.APIResponse) *apiResponseWrapper {
	return &apiResponseWrapper{
		ResourceName:  resourceName,
		Method:        apiResponse.Response.Request.Method,
		Path:          apiResponse.Response.Request.URL.Path,
		StatusCode:    apiResponse.StatusCode,
		ErrorMessage:  apiResponse.ErrorMessage,
		CorrelationID: apiResponse.CorrelationID,
	}
}

func BuildAPIDiagnosticError(resourceName string, summary string, apiResponse *platformclientv2.APIResponse) diag.Diagnostics {
	var msg string

	apiResponseWrapper := convertResponseToWrapper(resourceName, apiResponse)
	apiResponseWrapperByte, err := json.Marshal(apiResponseWrapper)

	if err != nil {
		msg = "Unable to unmarshal API response wrrapper hile building diagnostic error."
	} else {
		msg = string(apiResponseWrapperByte)
	}

	summaryMsg := fmt.Sprintf("%s: %s", resourceName, summary)
	dg := diag.Diagnostic{Severity: diag.Error, Summary: summaryMsg, Detail: msg}
	var dgs diag.Diagnostics
	dgs = append(dgs, dg)
	return dgs
}
