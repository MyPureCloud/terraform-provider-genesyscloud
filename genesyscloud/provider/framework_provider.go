// Package provider implements the Genesys Cloud Terraform Framework provider.
//
// This file contains the Plugin Framework provider implementation that works alongside
// the SDKv2 provider through muxing. The Framework provider handles resources that have
// been migrated from SDKv2 to the modern Plugin Framework architecture.
//
// Key Features:
//   - Native Protocol v6 support
//   - Type-safe configuration handling
//   - Shared provider metadata with SDKv2 provider
//   - Schema alignment for muxing compatibility
//
// The provider schema is intentionally aligned with the SDKv2 provider to ensure
// muxing compatibility. Any schema changes must be synchronized between both providers.
package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/mypurecloud/platform-client-sdk-go/v165/platformclientv2"
)

// Ensure GenesysCloudFrameworkProvider satisfies various provider interfaces.
var _ provider.Provider = &GenesysCloudFrameworkProvider{}

// GenesysCloudFrameworkProvider defines the provider implementation for the Plugin Framework.
type GenesysCloudFrameworkProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance testing.
	version string

	// Framework resources and data sources injected as dependencies
	frameworkResources   map[string]func() resource.Resource
	frameworkDataSources map[string]func() datasource.DataSource
}

// GenesysCloudFrameworkProviderModel describes the provider data model for the Framework provider.
// This model must be kept in sync with the SDKv2 provider schema to ensure muxing compatibility.
// All field names, types, and attributes must match exactly between providers.
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

// NewFrameworkProvider creates a factory function for the Framework provider.
// This provider handles resources and data sources that have been migrated from SDKv2
// to the Plugin Framework architecture.
//
// Parameters:
//   - version: Provider version string
//   - frameworkResources: Map of Framework resource names to factory functions
//   - frameworkDataSources: Map of Framework data source names to factory functions
//
// Returns:
//   - A factory function that creates a new Framework provider instance
//
// The provider is designed to work in a muxed environment alongside the SDKv2 provider,
// allowing for gradual migration of resources while maintaining backward compatibility.
func NewFrameworkProvider(version string, frameworkResources map[string]func() resource.Resource, frameworkDataSources map[string]func() datasource.DataSource) func() provider.Provider {
	return func() provider.Provider {
		return &GenesysCloudFrameworkProvider{
			version:              version,
			frameworkResources:   frameworkResources,
			frameworkDataSources: frameworkDataSources,
		}
	}
}

func (p *GenesysCloudFrameworkProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "genesyscloud"
	resp.Version = p.version
}

// Schema defines the provider schema for the Framework provider.
// This schema is intentionally kept identical to the SDKv2 provider schema to ensure
// muxing compatibility. Any changes here must be synchronized with the SDKv2 provider.
//
// Key alignment requirements:
//   - All attribute names, types, and properties must match exactly
//   - Sensitive field settings must be identical
//   - Environment variable references must be consistent
//   - Block descriptions must match (currently empty for both providers)
func (p *GenesysCloudFrameworkProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "", // Intentionally empty to match SDKv2 provider
		Attributes: map[string]schema.Attribute{
			"access_token": schema.StringAttribute{
				Optional:    true,
				Sensitive:   false,
				Description: "A string that the OAuth client uses to make requests. Can be set with the `GENESYSCLOUD_ACCESS_TOKEN` environment variable.",
			},
			"oauthclient_id": schema.StringAttribute{
				Optional:    true,
				Description: "OAuthClient ID found on the OAuth page of Admin UI. Can be set with the `GENESYSCLOUD_OAUTHCLIENT_ID` environment variable.",
			},
			"oauthclient_secret": schema.StringAttribute{
				Optional:    true,
				Sensitive:   false,
				Description: "OAuthClient secret found on the OAuth page of Admin UI. Can be set with the `GENESYSCLOUD_OAUTHCLIENT_SECRET` environment variable.",
			},
			"aws_region": schema.StringAttribute{
				Optional:    true,
				Description: "AWS region where org exists. e.g. us-east-1. Can be set with the `GENESYSCLOUD_REGION` environment variable.",
			},
			"sdk_debug": schema.BoolAttribute{
				Optional:    true,
				Description: "Enables debug tracing in the Genesys Cloud SDK. Output will be written to the local file 'sdk_debug.log'. Can be set with the `GENESYSCLOUD_SDK_DEBUG` environment variable.",
			},
			"sdk_debug_format": schema.StringAttribute{
				Optional:    true,
				Description: "Specifies the data format of the 'sdk_debug.log'. Only applicable if sdk_debug is true. Can be set with the `GENESYSCLOUD_SDK_DEBUG_FORMAT` environment variable. Default value is Text.",
			},
			"sdk_debug_file_path": schema.StringAttribute{
				Optional:    true,
				Description: "Specifies the file path for the log file. Can be set with the `GENESYSCLOUD_SDK_DEBUG_FILE_PATH` environment variable. Default value is sdk_debug.log",
			},
			"sdk_client_pool_debug": schema.BoolAttribute{
				Optional:    true,
				Description: "Enables debug tracing in the Genesys Cloud SDK client pool. Output will be written to standard log output. Can be set with the `GENESYSCLOUD_SDK_CLIENT_POOL_DEBUG` environment variable.",
			},
			"token_pool_size": schema.Int64Attribute{
				Optional:    true,
				Description: "Max number of OAuth tokens in the token pool. Can be set with the `GENESYSCLOUD_TOKEN_POOL_SIZE` environment variable.",
			},
			"token_acquire_timeout": schema.StringAttribute{
				Optional:    true,
				Description: "Timeout for acquiring a token from the pool. Can be set with the `GENESYSCLOUD_TOKEN_ACQUIRE_TIMEOUT` environment variable.",
			},
			"token_init_timeout": schema.StringAttribute{
				Optional:    true,
				Description: "Timeout for initializing the token pool. Can be set with the `GENESYSCLOUD_TOKEN_INIT_TIMEOUT` environment variable.",
			},
			"log_stack_traces": schema.BoolAttribute{
				Optional:    true,
				Description: "If true, stack traces will be logged to a file instead of crashing the provider, whenever possible.\nIf the stack trace occurs within the create context and before the ID is set in the schema object, then the command will fail with the message\n\"Root object was present, but now absent.\" Can be set with the GENESYSCLOUD_LOG_STACK_TRACES environment variable. **WARNING**: This is a debugging feature that may cause your Terraform state to become out of sync with the API.\nIf you encounter any stack traces, please report them so we can address the underlying issues.",
			},
			"log_stack_traces_file_path": schema.StringAttribute{
				Optional:    true,
				Description: "Specifies the file path for the stack trace logs. Can be set with the `GENESYSCLOUD_LOG_STACK_TRACES_FILE_PATH` environment variable. Default value is genesyscloud_stack_traces.log",
			},
		},
		Blocks: map[string]schema.Block{
			"gateway": schema.SetNestedBlock{
				Description: "",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"port": schema.StringAttribute{
							Optional:    true,
							Description: "Port for the gateway can be set with the `GENESYSCLOUD_GATEWAY_PORT` environment variable.",
						},
						"host": schema.StringAttribute{
							Optional:    true,
							Description: "Host for the gateway can be set with the `GENESYSCLOUD_GATEWAY_HOST` environment variable.",
						},
						"protocol": schema.StringAttribute{
							Optional:    true,
							Description: "Protocol for the gateway can be set with the `GENESYSCLOUD_GATEWAY_PROTOCOL` environment variable.",
						},
					},
					Blocks: map[string]schema.Block{
						"path_params": schema.SetNestedBlock{
							Description: "",
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"path_name": schema.StringAttribute{
										Required:    true,
										Description: "Path name for Gateway Path Params can be set with the `GENESYSCLOUD_GATEWAY_PATH_NAME` environment variable.",
									},
									"path_value": schema.StringAttribute{
										Required:    true,
										Description: "Path value for Gateway Path Params can be set with the `GENESYSCLOUD_GATEWAY_PATH_VALUE` environment variable.",
									},
								},
							},
						},
						"auth": schema.SetNestedBlock{
							Description: "",
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"username": schema.StringAttribute{
										Optional:    true,
										Description: "UserName for the Auth can be set with the `GENESYSCLOUD_PROXY_AUTH_USERNAME` environment variable.",
									},
									"password": schema.StringAttribute{
										Optional:    true,
										Sensitive:   false,
										Description: "Password for the Auth can be set with the `GENESYSCLOUD_PROXY_AUTH_PASSWORD` environment variable.",
									},
								},
							},
						},
					},
				},
			},
			"proxy": schema.SetNestedBlock{
				Description: "",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"port": schema.StringAttribute{
							Optional:    true,
							Description: "Port for the proxy can be set with the `GENESYSCLOUD_PROXY_PORT` environment variable.",
						},
						"host": schema.StringAttribute{
							Optional:    true,
							Description: "Host for the proxy can be set with the `GENESYSCLOUD_PROXY_HOST` environment variable.",
						},
						"protocol": schema.StringAttribute{
							Optional:    true,
							Description: "Protocol for the proxy can be set with the `GENESYSCLOUD_PROXY_PROTOCOL` environment variable.",
						},
					},
					Blocks: map[string]schema.Block{
						"auth": schema.SetNestedBlock{
							Description: "",
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"username": schema.StringAttribute{
										Optional:    true,
										Description: "UserName for the Auth can be set with the `GENESYSCLOUD_PROXY_AUTH_USERNAME` environment variable.",
									},
									"password": schema.StringAttribute{
										Optional:    true,
										Sensitive:   false,
										Description: "Password for the Auth can be set with the `GENESYSCLOUD_PROXY_AUTH_PASSWORD` environment variable.",
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

// Configure configures the Framework provider with the given configuration.
// This method handles authentication, API client setup, and provider metadata creation.
// It attempts to share configuration with the SDKv2 provider when possible to avoid
// duplicate authentication and API client creation in muxed environments.
//
// Configuration priority:
//  1. Explicit configuration values
//  2. Environment variables
//  3. Shared metadata from SDKv2 provider (if available)
//
// The configured provider metadata is made available to Framework resources and data sources.
func (p *GenesysCloudFrameworkProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data GenesysCloudFrameworkProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Get configuration values with environment variable fallbacks
	accessToken := getStringValue(data.AccessToken, "GENESYSCLOUD_ACCESS_TOKEN")
	oauthClientID := getStringValue(data.OAuthClientID, "GENESYSCLOUD_OAUTHCLIENT_ID")
	oauthClientSecret := getStringValue(data.OAuthClientSecret, "GENESYSCLOUD_OAUTHCLIENT_SECRET")
	awsRegion := getStringValue(data.AWSRegion, "GENESYSCLOUD_REGION")

	// Validate required configuration
	if accessToken == "" && (oauthClientID == "" || oauthClientSecret == "") {
		resp.Diagnostics.AddError(
			"Missing Authentication Configuration",
			"Either access_token or both oauthclient_id and oauthclient_secret must be provided",
		)
		return
	}

	if awsRegion == "" {
		resp.Diagnostics.AddError(
			"Missing Region Configuration",
			"aws_region must be provided",
		)
		return
	}

	// Create Genesys Cloud configuration
	config := platformclientv2.GetDefaultConfiguration()
	config.BasePath = GetRegionBasePath(awsRegion)

	// Configure authentication
	if accessToken != "" {
		config.AccessToken = accessToken
	} else {
		config.AutomaticTokenRefresh = true
		err := config.AuthorizeClientCredentials(oauthClientID, oauthClientSecret)
		if err != nil {
			resp.Diagnostics.AddError(
				"Authentication Failed",
				"Failed to authorize with Genesys Cloud: "+err.Error(),
			)
			return
		}
	}

	// Create or use shared provider meta
	providerMeta := FrameworkProviderMeta(p.version, config, getRegionDomain(awsRegion))

	// If we created new meta (not shared), update it with our config
	if !IsSharedMetaAvailable() {
		providerMeta.Version = p.version
		providerMeta.ClientConfig = config
		providerMeta.Domain = getRegionDomain(awsRegion)
	}

	// Store the configuration for Framework resources to use
	resp.DataSourceData = providerMeta
	resp.ResourceData = providerMeta
}

func (p *GenesysCloudFrameworkProvider) Resources(ctx context.Context) []func() resource.Resource {
	// Convert map to slice of factory functions
	var resourceFactories []func() resource.Resource
	for _, factory := range p.frameworkResources {
		resourceFactories = append(resourceFactories, factory)
	}
	return resourceFactories
}

func (p *GenesysCloudFrameworkProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	// Convert map to slice of factory functions
	var dataSourceFactories []func() datasource.DataSource
	for _, factory := range p.frameworkDataSources {
		dataSourceFactories = append(dataSourceFactories, factory)
	}
	return dataSourceFactories
}

// getStringValue gets a string value from a Framework types.String with environment variable fallback
func getStringValue(attr types.String, envVar string) string {
	if !attr.IsNull() && !attr.IsUnknown() {
		return attr.ValueString()
	}
	return os.Getenv(envVar)
}
