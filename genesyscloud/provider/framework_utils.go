package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/mypurecloud/platform-client-sdk-go/v165/platformclientv2"
)

// FrameworkProviderData contains the provider configuration data for Framework resources
type FrameworkProviderData struct {
	Client *platformclientv2.Configuration
	Meta   *ProviderMeta
}

// GetFrameworkProviderData extracts provider data from Framework resource/datasource configure methods
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

// ConfigureFrameworkResource is a helper function to configure Framework resources
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

// ConfigureFrameworkDataSource is a helper function to configure Framework data sources
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
