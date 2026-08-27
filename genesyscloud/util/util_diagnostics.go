package util

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/mypurecloud/platform-client-sdk-go/v195/platformclientv2"
)

// Stable diagnostic codes for machine classification (e.g. by MRMO replicator).
const (
	DiagnosticCodeFlowPublishFailed  = "flow_publish_failed"
	DiagnosticCodeFlowPublishTimeout = "flow_publish_timeout"
)

type detailedDiagnosticInfo struct {
	ResourceType        string        `json:"resourceType,omitempty"`
	Method              string        `json:"method,omitempty"`
	Path                string        `json:"path,omitempty"`
	StatusCode          int           `json:"statusCode,omitempty"`
	ErrorMessage        string        `json:"errorMessage,omitempty"`
	ErrorMessageContext interface{}   `json:"errorMessageContext,omitempty"`
	ErrorMessageDetails []interface{} `json:"errorMessageDetails,omitempty"`
	CorrelationID       string        `json:"correlationId,omitempty"`
	Code                string        `json:"code,omitempty"`
	JobID               string        `json:"jobId,omitempty"`
}

func convertResponseToWrapper(resourceType string, apiResponse *platformclientv2.APIResponse) *detailedDiagnosticInfo {

	detailedDiagnosticInfo := &detailedDiagnosticInfo{
		ResourceType:  resourceType,
		StatusCode:    apiResponse.StatusCode,
		ErrorMessage:  apiResponse.ErrorMessage,
		CorrelationID: apiResponse.CorrelationID,
	}
	if apiResponse.Error != nil && len(apiResponse.Error.MessageParams) > 0 {
		detailedDiagnosticInfo.ErrorMessageContext = apiResponse.Error.MessageParams
	}
	if apiResponse.Error != nil && len(apiResponse.Error.Details) > 0 {
		for _, detail := range apiResponse.Error.Details {
			detailedDiagnosticInfo.ErrorMessageDetails = append(detailedDiagnosticInfo.ErrorMessageDetails, detail)
		}
	}
	if apiResponse.Response != nil && apiResponse.Response.Request != nil {
		detailedDiagnosticInfo.Method = apiResponse.Response.Request.Method
		if apiResponse.Response.Request.URL != nil {
			detailedDiagnosticInfo.Path = apiResponse.Response.Request.URL.Path
		}
	}
	return detailedDiagnosticInfo
}

func BuildAPIDiagnosticError(resourceType string, summary string, apiResponse *platformclientv2.APIResponse) diag.Diagnostics {
	//Checking to make sure we have properly formed response
	if apiResponse == nil {
		err := fmt.Errorf("unable to build a message from the response because the APIResponse does not contain the appropriate data.%s", "")
		return BuildDiagnosticError(resourceType, summary, err)
	}
	diagInfo := convertResponseToWrapper(resourceType, apiResponse)
	diagInfoByte, err := json.Marshal(diagInfo)

	//Checking to see if we can Marshall the data
	if err != nil {
		err = fmt.Errorf("unable to unmarshal diagnostic info while building diagnostic error. Error: %w", err)
		return BuildDiagnosticError(resourceType, summary, err)
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

// BuildCodedAPIDiagnosticError builds an API diagnostic with a stable machine-readable code and optional job ID.
func BuildCodedAPIDiagnosticError(resourceType, code, summary, jobId string, apiResponse *platformclientv2.APIResponse) diag.Diagnostics {
	if apiResponse == nil {
		err := fmt.Errorf("unable to build a message from the response because the APIResponse does not contain the appropriate data.%s", "")
		diags := BuildDiagnosticError(resourceType, summary, err)
		return withDiagnosticCodeAndJobID(diags, code, jobId)
	}
	diagInfo := convertResponseToWrapper(resourceType, apiResponse)
	diagInfo.Code = code
	diagInfo.JobID = jobId
	diagInfoByte, err := json.Marshal(diagInfo)
	if err != nil {
		err = fmt.Errorf("unable to unmarshal diagnostic info while building diagnostic error. Error: %w", err)
		return withDiagnosticCodeAndJobID(BuildDiagnosticError(resourceType, summary, err), code, jobId)
	}

	dg := diag.Diagnostic{Severity: diag.Error, Summary: summary, Detail: string(diagInfoByte)}
	return diag.Diagnostics{dg}
}

// BuildCodedWithRetriesApiDiagnosticError is like BuildWithRetriesApiDiagnosticError but includes
// a stable diagnostic code and job ID in the JSON Detail for downstream classification.
func BuildCodedWithRetriesApiDiagnosticError(resourceType, code, summary, jobId string, apiResponse *platformclientv2.APIResponse) error {
	var errorMsg string
	diagnostic := BuildCodedAPIDiagnosticError(resourceType, code, summary, jobId, apiResponse)
	for _, d := range diagnostic {
		errorMsg += fmt.Sprintf("%s\n%s\n", d.Summary, d.Detail)
	}
	return errors.New(errorMsg)
}

func withDiagnosticCodeAndJobID(diags diag.Diagnostics, code, jobId string) diag.Diagnostics {
	if len(diags) == 0 || (code == "" && jobId == "") {
		return diags
	}
	info := &detailedDiagnosticInfo{}
	if err := json.Unmarshal([]byte(diags[0].Detail), info); err != nil {
		info = &detailedDiagnosticInfo{}
	}
	if code != "" {
		info.Code = code
	}
	if jobId != "" {
		info.JobID = jobId
	}
	diagInfoByte, err := json.Marshal(info)
	if err != nil {
		return diags
	}
	diags[0].Detail = string(diagInfoByte)
	return diags
}
