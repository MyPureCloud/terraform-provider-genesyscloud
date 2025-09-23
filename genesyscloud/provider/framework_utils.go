// Package provider contains utility functions for Framework provider integration.
//
// This file provides helper functions and types that facilitate the integration
// of Plugin Framework resources and data sources with the provider configuration
// and metadata system.
//
// Key utilities:
//   - Provider data extraction and validation
//   - Resource and data source configuration helpers
//   - Type-safe provider metadata access
//
// These utilities ensure consistent provider data handling across all Framework
// resources and data sources in the muxed provider environment.
package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/mypurecloud/platform-client-sdk-go/v165/platformclientv2"
)

// FrameworkProviderData contains the provider configuration data for Framework resources.
// This struct provides a convenient way to access both the Genesys Cloud API client
// configuration and the full provider metadata from Framework resources and data sources.
//
// Fields:
//   - Client: Configured Genesys Cloud API client for making API calls
//   - Meta: Complete provider metadata including version, domain, and configuration
type FrameworkProviderData struct {
	Client *platformclientv2.Configuration
	Meta   *ProviderMeta
}

// GetFrameworkProviderData extracts and validates provider data from Framework resource/datasource configure methods.
// This function handles the type assertion and validation of provider data passed from the Framework provider
// to individual resources and data sources during their configuration phase.
//
// Parameters:
//   - ctx: Context for the operation
//   - req: Configure request from either resource.ConfigureRequest or datasource.ConfigureRequest
//
// Returns:
//   - *FrameworkProviderData: Validated provider data with client and metadata
//   - bool: Success indicator (false if extraction/validation failed)
//
// This function is used internally by ConfigureFrameworkResource and ConfigureFrameworkDataSource.
func GetFrameworkProviderData(ctx context.Context, req interface{}) (*FrameworkProviderData, bool) {
	var providerData interface{}

	switch r := req.(type) {
	case resource.ConfigureRequest:
		providerData = r.ProviderData
	case datasource.ConfigureRequest:
		providerData = r.ProviderData
	default:
		return nil, false
	}

	if providerData == nil {
		return nil, false
	}

	meta, ok := providerData.(*ProviderMeta)
	if !ok {
		return nil, false
	}

	return &FrameworkProviderData{
		Client: meta.ClientConfig,
		Meta:   meta,
	}, true
}

// ConfigureFrameworkResource is a helper function to configure Framework resources with provider data.
// This function should be called from the Configure method of Framework resources to extract
// and validate the provider configuration data.
//
// Parameters:
//   - ctx: Context for the operation
//   - req: Resource configure request from the Framework
//   - resp: Resource configure response to populate with any errors
//
// Returns:
//   - *FrameworkProviderData: Validated provider data, or nil if configuration failed
//
// Usage example:
//
//	func (r *myResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
//	    r.providerData = ConfigureFrameworkResource(ctx, req, resp)
//	}
func ConfigureFrameworkResource(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) *FrameworkProviderData {
	if req.ProviderData == nil {
		return nil
	}

	providerData, ok := GetFrameworkProviderData(ctx, req)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			"Expected *ProviderMeta, got something else. Please report this issue to the provider developers.",
		)
		return nil
	}

	return providerData
}

// ConfigureFrameworkDataSource is a helper function to configure Framework data sources with provider data.
// This function should be called from the Configure method of Framework data sources to extract
// and validate the provider configuration data.
//
// Parameters:
//   - ctx: Context for the operation
//   - req: Data source configure request from the Framework
//   - resp: Data source configure response to populate with any errors
//
// Returns:
//   - *FrameworkProviderData: Validated provider data, or nil if configuration failed
//
// Usage example:
//
//	func (d *myDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
//	    d.providerData = ConfigureFrameworkDataSource(ctx, req, resp)
//	}
func ConfigureFrameworkDataSource(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) *FrameworkProviderData {
	if req.ProviderData == nil {
		return nil
	}

	providerData, ok := GetFrameworkProviderData(ctx, req)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected DataSource Configure Type",
			"Expected *ProviderMeta, got something else. Please report this issue to the provider developers.",
		)
		return nil
	}

	return providerData
}
