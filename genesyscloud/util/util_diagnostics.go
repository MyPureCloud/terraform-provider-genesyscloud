package util

import (
	"encoding/json"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/mypurecloud/platform-client-sdk-go/v123/platformclientv2"
)

type detailedDiagnosticInfo struct {
	ResourceName  string `json:"resourceName,omitempty"`
	Method        string `json:"method,omitempty"`
	Path          string `json:"path:omitempty"`
	StatusCode    int    `json:"statusCode:omitempty"`
	ErrorMessage  string `json:"errorMessage,omitempty"`
	CorrelationID string `json:"correlationId,omitempty"`
}

func convertResponseToWrapper(resourceName string, apiResponse *platformclientv2.APIResponse) *detailedDiagnosticInfo {
	return &detailedDiagnosticInfo{
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

	diagInfo := convertResponseToWrapper(resourceName, apiResponse)
	diagInfoByte, err := json.Marshal(diagInfo)

	if err != nil {
		msg = "Unable to unmarshal diagnostic info while building diagnostic error."
	} else {
		msg = string(diagInfoByte)
	}

	summaryMsg := fmt.Sprintf("%s: %s", resourceName, summary)
	dg := diag.Diagnostic{Severity: diag.Error, Summary: summaryMsg, Detail: msg}
	var dgs diag.Diagnostics
	dgs = append(dgs, dg)
	return dgs
}

func BuildDiagnosticError(resourceName string, summary string, err error) diag.Diagnostics {
	var msg string

	diagInfo := &detailedDiagnosticInfo{
		ResourceName: resourceName,
		ErrorMessage: fmt.Sprint("%s", err),
	}
	diagInfoByte, err := json.Marshal(diagInfo)

	if err != nil {
		msg = "Unable to unmarshal diagnostic info while building diagnostic error"
	} else {
		msg = string(diagInfoByte)
	}

	summaryMsg := fmt.Sprintf("%s: %s", resourceName, summary)
	dg := diag.Diagnostic{Severity: diag.Error, Summary: summaryMsg, Detail: msg}
	var dgs diag.Diagnostics
	dgs = append(dgs, dg)
	return dgs
}
