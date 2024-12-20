package util

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/mypurecloud/platform-client-sdk-go/v149/platformclientv2"
)

type detailedDiagnosticInfo struct {
	ResourceType  string `json:"resourceType,omitempty"`
	Method        string `json:"method,omitempty"`
	Path          string `json:"path,omitempty"`
	StatusCode    int    `json:"statusCode,omitempty"`
	ErrorMessage  string `json:"errorMessage,omitempty"`
	CorrelationID string `json:"correlationId,omitempty"`
}

func convertResponseToWrapper(resourceType string, apiResponse *platformclientv2.APIResponse) *detailedDiagnosticInfo {
	return &detailedDiagnosticInfo{
		ResourceType:  resourceType,
		Method:        apiResponse.Response.Request.Method,
		Path:          apiResponse.Response.Request.URL.Path,
		StatusCode:    apiResponse.StatusCode,
		ErrorMessage:  apiResponse.ErrorMessage,
		CorrelationID: apiResponse.CorrelationID,
	}
}

func BuildAPIDiagnosticError(resourceType string, summary string, apiResponse *platformclientv2.APIResponse) diag.Diagnostics {
	//Checking to make sure we have properly formed response
	if apiResponse == nil || apiResponse.Response == nil || apiResponse.Response.Request == nil || apiResponse.Response.Request.URL == nil {
		error := fmt.Errorf("Unable to build a message from the response because the APIResponse does not contain the appropriate data.%s", "")
		return BuildDiagnosticError(resourceType, summary, error)
	}
	diagInfo := convertResponseToWrapper(resourceType, apiResponse)
	diagInfoByte, err := json.Marshal(diagInfo)

	//Checking to see if we can Marshall the data
	if err != nil {
		error := fmt.Errorf("Unable to unmarshal diagnostic info while building diagnostic error. Error: %s", err)
		return BuildDiagnosticError(resourceType, summary, error)
	}

	dg := diag.Diagnostic{Severity: diag.Error, Summary: summary, Detail: string(diagInfoByte)}
	var dgs diag.Diagnostics
	dgs = append(dgs, dg)
	return dgs
}

func BuildDiagnosticError(resourceType string, summary string, err error) diag.Diagnostics {
	var msg string
	diagInfo := &detailedDiagnosticInfo{
		ResourceType: resourceType,
		ErrorMessage: fmt.Sprintf("%s", err),
	}
	diagInfoByte, err := json.Marshal(diagInfo)

	if err != nil {
		msg = fmt.Sprintf("{'resourceType': '%s', 'details': 'Unable to unmarshal diagnostic info while building diagnostic error'}", resourceType)
	} else {
		msg = string(diagInfoByte)
	}

	dg := diag.Diagnostic{Severity: diag.Error, Summary: summary, Detail: msg}

	var dgs diag.Diagnostics
	dgs = append(dgs, dg)
	return dgs
}

// BuildWithRetriesApiDiagnosticError converts the diag.Diagnostic error from API responses into an error to be used in withRetries functions for more clear error information
func BuildWithRetriesApiDiagnosticError(resourceType string, summary string, apiResponse *platformclientv2.APIResponse) error {
	var errorMsg string

	diagnostic := BuildAPIDiagnosticError(resourceType, summary, apiResponse)
	for _, diags := range diagnostic {
		errorMsg += fmt.Sprintf("%s\n%s\n", diags.Summary, diags.Detail)
	}
	return errors.New(errorMsg)
}
