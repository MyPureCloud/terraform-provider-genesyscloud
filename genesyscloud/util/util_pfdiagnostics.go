package util

import (
	"encoding/json"
	"fmt"

	frameworkdiag "github.com/hashicorp/terraform-plugin-framework/diag"
	sdkdiag "github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/mypurecloud/platform-client-sdk-go/v176/platformclientv2"
)

// ============================================================================
// Framework Diagnostic Utility Functions
// These functions provide diagnostic creation and conversion utilities for
// Terraform Plugin Framework resources and data sources
// ============================================================================

// ConvertSDKDiagnosticsToFramework converts SDK v2 diagnostics to Framework diagnostics
// while preserving identical error messages, severity levels, and diagnostic details
func ConvertSDKDiagnosticsToFramework(sdkDiags sdkdiag.Diagnostics) frameworkdiag.Diagnostics {
	var frameworkDiags frameworkdiag.Diagnostics

	for _, sdkDiag := range sdkDiags {
		switch sdkDiag.Severity {
		case sdkdiag.Error:
			frameworkDiags.AddError(sdkDiag.Summary, sdkDiag.Detail)
		case sdkdiag.Warning:
			frameworkDiags.AddWarning(sdkDiag.Summary, sdkDiag.Detail)
		default:
			// Default to error for unknown severity levels to maintain SDK behavior
			frameworkDiags.AddError(sdkDiag.Summary, sdkDiag.Detail)
		}
	}

	return frameworkDiags
}

// BuildFrameworkDiagnosticError creates a Framework diagnostic error
// This mirrors BuildDiagnosticError from util_diagnostics.go but returns Framework diagnostics
func BuildFrameworkDiagnosticError(resourceType string, summary string, err error) frameworkdiag.Diagnostics {
	var frameworkDiags frameworkdiag.Diagnostics

	diagInfo := &detailedDiagnosticInfo{
		ResourceType: resourceType,
		ErrorMessage: fmt.Sprintf("%s", err),
	}
	diagInfoByte, marshalErr := json.Marshal(diagInfo)

	var detail string
	if marshalErr != nil {
		detail = fmt.Sprintf("{'resourceType': '%s', 'details': 'Unable to unmarshal diagnostic info while building diagnostic error'}", resourceType)
	} else {
		detail = string(diagInfoByte)
	}

	frameworkDiags.AddError(summary, detail)
	return frameworkDiags
}

// BuildFrameworkAPIDiagnosticError creates Framework diagnostics from API errors
// This mirrors BuildAPIDiagnosticError from util_diagnostics.go but returns Framework diagnostics
func BuildFrameworkAPIDiagnosticError(resourceType string, summary string, apiResponse *platformclientv2.APIResponse) frameworkdiag.Diagnostics {
	var frameworkDiags frameworkdiag.Diagnostics

	// Checking to make sure we have properly formed response
	if apiResponse == nil {
		err := fmt.Errorf("unable to build a message from the response because the APIResponse does not contain the appropriate data")
		return BuildFrameworkDiagnosticError(resourceType, summary, err)
	}

	// The convertResponseToWrapper() function is used from the util_diagnostics.go file.
	diagInfo := convertResponseToWrapper(resourceType, apiResponse)
	diagInfoByte, err := json.Marshal(diagInfo)

	// Checking to see if we can Marshal the data
	if err != nil {
		err = fmt.Errorf("unable to unmarshal diagnostic info while building diagnostic error. Error: %w", err)
		return BuildFrameworkDiagnosticError(resourceType, summary, err)
	}

	frameworkDiags.AddError(summary, string(diagInfoByte))
	return frameworkDiags
}

// ============================================================================
// SDKv2 Diagnostic Utility Functions for Cache Compatibility
// These functions create SDKv2 diagnostics for use with cache infrastructure
// that still depends on SDKv2. When cache is migrated to Framework, these
// can be deprecated in favor of the Framework versions above.
// ============================================================================

// BuildSDKDiagnosticError creates an SDKv2 diagnostic error
// This is a simple helper for creating SDKv2 diagnostics in Framework code
// that needs to interact with SDKv2-based infrastructure (like caches)
func BuildSDKDiagnosticError(summary string, detail string) sdkdiag.Diagnostics {
	return sdkdiag.Diagnostics{
		sdkdiag.Diagnostic{
			Severity: sdkdiag.Error,
			Summary:  summary,
			Detail:   detail,
		},
	}
}

// BuildSDKDiagnosticWarning creates an SDKv2 diagnostic warning
func BuildSDKDiagnosticWarning(summary string, detail string) sdkdiag.Diagnostics {
	return sdkdiag.Diagnostics{
		sdkdiag.Diagnostic{
			Severity: sdkdiag.Warning,
			Summary:  summary,
			Detail:   detail,
		},
	}
}
