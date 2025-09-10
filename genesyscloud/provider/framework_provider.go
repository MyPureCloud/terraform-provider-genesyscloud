package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure GenesysCloudFrameworkProvider satisfies various provider interfaces.
var _ provider.Provider = &GenesysCloudFrameworkProvider{}

// GenesysCloudFrameworkProvider defines the provider implementation for the Plugin Framework.
type GenesysCloudFrameworkProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance testing.
	version string
}

// GenesysCloudFrameworkProviderModel describes the provider data model.
type GenesysCloudFrameworkProviderModel struct {
	AccessToken            types.String `tfsdk:"access_token"`
	OAuthClientID          types.String `tfsdk:"oauthclient_id"`
	OAuthClientSecret      types.String `tfsdk:"oauthclient_secret"`
	AWSRegion              types.String `tfsdk:"aws_region"`
	SDKDebug               types.Bool   `tfsdk:"sdk_debug"`
	SDKDebugFormat         types.String `tfsdk:"sdk_debug_format"`
	SDKDebugFilePath       types.String `tfsdk:"sdk_debug_file_path"`
	SDKClientPoolDebug     types.Bool   `tfsdk:"sdk_client_pool_debug"`
	TokenPoolSize          types.Int64  `tfsdk:"token_pool_size"`
	TokenAcquireTimeout    types.String `tfsdk:"token_acquire_timeout"`
	TokenInitTimeout       types.String `tfsdk:"token_init_timeout"`
	LogStackTraces         types.Bool   `tfsdk:"log_stack_traces"`
	LogStackTracesFilePath types.String `tfsdk:"log_stack_traces_file_path"`
	Gateway                types.Set    `tfsdk:"gateway"`
	Proxy                  types.Set    `tfsdk:"proxy"`
}

func NewFrameworkProvider(version string) func() provider.Provider {
	return func() provider.Provider {
		return &GenesysCloudFrameworkProvider{
			version: version,
		}
	}
}

func (p *GenesysCloudFrameworkProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "genesyscloud"
	resp.Version = p.version
}

func (p *GenesysCloudFrameworkProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	// For now, return an empty schema. This ensures the Framework provider doesn't conflict
	// with the SDKv2 provider schema during muxing. All provider configuration will be
	// handled by the SDKv2 provider until we start migrating resources.
	resp.Schema = schema.Schema{}
}

func (p *GenesysCloudFrameworkProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data GenesysCloudFrameworkProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// For now, we don't need to configure anything since no resources are migrated yet
	// This will be expanded when we start migrating resources
}

func (p *GenesysCloudFrameworkProvider) Resources(ctx context.Context) []func() resource.Resource {
	// Initially empty - resources will be added here as they are migrated from SDKv2
	return []func() resource.Resource{}
}

func (p *GenesysCloudFrameworkProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	// Initially empty - data sources will be added here as they are migrated from SDKv2
	return []func() datasource.DataSource{}
}
