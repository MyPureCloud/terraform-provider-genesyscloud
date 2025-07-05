package provider

import (
	"context"
	"fmt"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/platform"
	customvalidators "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider/framework_custom_validators"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/routing_wrapupcode_v2"
	prl "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/panic_recovery_logger"
	"log"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/mypurecloud/platform-client-sdk-go/v157/platformclientv2"
)

var (
	// ensure the implementation satisfies the expected interfaces
	_ provider.Provider = &GenesysCloudProvider{}
)

type GenesysCloudProvider struct {
	Meta          ProviderMeta
	Version       string
	SdkClientPool SDKClientPool

	AttributeEnvValues     *providerEnvVars
	TokenPoolSize          int32
	LogStackTraces         bool
	LogStackTracesFilePath string

	AuthDetails  *AuthInfo
	SdkDebugInfo *SdkDebugInfo
	Proxy        *Proxy
	Gateway      *Gateway
}

type AuthInfo struct {
	AccessToken  string
	ClientId     string
	ClientSecret string
	Region       string
}

type SdkDebugInfo struct {
	DebugEnabled bool
	Format       string
	FilePath     string
}

type Proxy struct {
	Port     string
	Host     string
	Protocol string
	Auth     *Auth
}

type Auth struct {
	Username string
	Password string
}

type Gateway struct {
	Port       string
	Host       string
	Protocol   string
	PathParams []PathParam
	Auth       *Auth
}

type PathParam struct {
	PathName  string
	PathValue string
}

func NewFrameWorkProvider(version string) func() provider.Provider {
	return func() provider.Provider {
		return &GenesysCloudProvider{
			Version: version,
		}
	}
}

func (f *GenesysCloudProvider) Metadata(_ context.Context, request provider.MetadataRequest, response *provider.MetadataResponse) {
	response.TypeName = "genesyscloud"
}

func (f *GenesysCloudProvider) Schema(_ context.Context, request provider.SchemaRequest, response *provider.SchemaResponse) {
	response.Schema = schema.Schema{
		Blocks: map[string]schema.Block{
			"gateway": schema.SetNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"auth": schema.SetNestedBlock{
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"username": schema.StringAttribute{
										Required:            true,
										MarkdownDescription: fmt.Sprintf("UserName for the Auth can be set with the `%s` environment variable.", gatewayAuthUsernameEnvVar),
									},
									"password": schema.StringAttribute{
										Optional:            true,
										Sensitive:           true,
										MarkdownDescription: fmt.Sprintf("Password for the Auth can be set with the `%s` environment variable.", gatewayAuthPasswordEnvVar),
									},
								},
							},
						},
						"path_params": schema.SetNestedBlock{
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"path_name": schema.StringAttribute{
										Required:            true,
										MarkdownDescription: fmt.Sprintf("Path name for Gateway Path Params can be set with the `%s` environment variable.", gatewayPathParamsNameEnvVar),
									},
									"path_value": schema.StringAttribute{
										Required:            true,
										MarkdownDescription: fmt.Sprintf("Path value for Gateway Path Params can be set with the `%s` environment variable.", gatewayPathParamsValueEnvVar),
									},
								},
							},
						},
					},
					Attributes: map[string]schema.Attribute{
						"port": schema.StringAttribute{
							Optional:            true,
							MarkdownDescription: fmt.Sprintf("Port for the gateway can be set with the `%s` environment variable.", gatewayPortEnvVar),
						},
						"host": schema.StringAttribute{
							Optional:            true,
							MarkdownDescription: fmt.Sprintf("Host for the gateway can be set with the `%s` environment variable.", gatewayHostEnvVar),
						},
						"protocol": schema.StringAttribute{
							Optional:            true,
							MarkdownDescription: fmt.Sprintf("Protocol for the gateway can be set with the `%s` environment variable.", gatewayProtocolEnvVar),
						},
					},
				},
			},
			"proxy": schema.SetNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"auth": schema.SetNestedBlock{
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"username": schema.StringAttribute{
										Required:            true,
										MarkdownDescription: fmt.Sprintf("UserName for the Auth can be set with the `%s` environment variable.", proxyAuthUsernameEnvVar),
									},
									"password": schema.StringAttribute{
										Optional:            true,
										MarkdownDescription: fmt.Sprintf("Password for the Auth can be set with the `%s` environment variable.", proxyAuthPasswordEnvVar),
										Sensitive:           true,
									},
								},
							},
						},
					},
					Attributes: map[string]schema.Attribute{
						"port": schema.StringAttribute{
							Optional:            true,
							MarkdownDescription: fmt.Sprintf("Port for the proxy can be set with the `%s` environment variable.", proxyPortEnvVar),
						},
						"host": schema.StringAttribute{
							Optional:            true,
							MarkdownDescription: fmt.Sprintf("Host for the proxy can be set with the `%s` environment variable.", proxyHostEnvVar),
						},
						"protocol": schema.StringAttribute{
							Optional:            true,
							MarkdownDescription: fmt.Sprintf("Protocol for the proxy can be set with the `%s` environment variable.", proxyProtocolEnvVar),
						},
					},
				},
			},
		},
		Attributes: map[string]schema.Attribute{
			"access_token": schema.StringAttribute{
				Optional:            true,
				Sensitive:           true,
				MarkdownDescription: fmt.Sprintf("A string that the OAuth client uses to make requests. Can be set with the `%s` environment variable.", accessTokenEnvVar),
			},
			"oauthclient_id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: fmt.Sprintf("OAuthClient ID found on the OAuth page of Admin UI. Can be set with the `%s` environment variable.", clientIdEnvVar),
			},
			"oauthclient_secret": schema.StringAttribute{
				Optional:            true,
				Sensitive:           true,
				MarkdownDescription: fmt.Sprintf("OAuthClient secret found on the OAuth page of Admin UI. Can be set with the `%s` environment variable.", clientSecretEnvVar),
			},
			"aws_region": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: fmt.Sprintf("AWS region where org exists. e.g. us-east-1. Can be set with the `%s` environment variable.", regionEnvVar),
				Validators: []validator.String{
					stringvalidator.OneOfCaseInsensitive(getAllowedRegions()...),
				},
			},
			"sdk_debug": schema.BoolAttribute{
				Optional:            true,
				MarkdownDescription: fmt.Sprintf("Enables debug tracing in the Genesys Cloud SDK. Output will be written to `sdk_debug_file_path`. Can be set with the `%s` environment variable.", sdkDebugEnvVar),
			},
			"sdk_debug_format": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: fmt.Sprintf("Specifies the data format of the 'sdk_debug.log'. Only applicable if sdk_debug is true. Can be set with the `%s` environment variable. Default value is %s.", sdkDebugFormatEnvVar, sdkDebugFormatDefaultValue),
				Validators: []validator.String{
					stringvalidator.OneOf("Text", "Json"),
				},
			},
			AttrSdkClientPoolDebug: schema.BoolAttribute{
				Optional:            true,
				MarkdownDescription: "Enables debug tracing in the Genesys Cloud SDK client pool. Output will be written to standard log output. Can be set with the `GENESYSCLOUD_SDK_CLIENT_POOL_DEBUG` environment variable.",
			},
			AttrTokenAcquireTimeout: schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Timeout for acquiring a token from the pool. Can be set with the `GENESYSCLOUD_TOKEN_ACQUIRE_TIMEOUT` environment variable.",
				Validators: []validator.String{
					customvalidators.ValidateDuration(),
				},
			},
			"sdk_debug_file_path": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: fmt.Sprintf("Specifies the file path for the log file. Can be set with the `%s` environment variable. Default value is %s", sdkDebugFilePathEnvVar, sdkDebugFilePathDefaultValue),
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(`^\S+$`), "Invalid file path."),
				},
			},
			"token_pool_size": schema.Int32Attribute{
				Optional:            true,
				MarkdownDescription: fmt.Sprintf("Max number of OAuth tokens in the token pool. Can be set with the `%s` environment variable.", tokenPoolSizeEnvVar),
				Validators: []validator.Int32{
					int32validator.Between(1, 20),
				},
			},
			AttrTokenInitTimeout: schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: fmt.Sprintf("Timeout for initializing the token pool. Can be set with the `%s` environment variable.", tokenInitTimeoutEnvVar),
				Validators: []validator.String{
					customvalidators.ValidateDuration(),
				},
			},
			"log_stack_traces": schema.BoolAttribute{
				Optional: true,
				MarkdownDescription: fmt.Sprintf(`If true, stack traces will be logged to a file instead of crashing the provider, whenever possible.
If the stack trace occurs within the create context and before the ID is set in the schema object, then the command will fail with the message
"Root object was present, but now absent." Can be set with the %s environment variable. **WARNING**: This is a debugging feature that may cause your Terraform state to become out of sync with the API.
If you encounter any stack traces, please report them so we can address the underlying issues.`, logStackTracesEnvVar),
			},
			"log_stack_traces_file_path": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: fmt.Sprintf("Specifies the file path for the stack trace logs. Can be set with the `%s` environment variable. Default value is genesyscloud_stack_traces.log", logStackTracesFilePathEnvVar),
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(`^\S+\.log$`), "File path cannot be an empty string, contain whitespaces, and must end with the .log extension."),
				},
			},
		},
	}
}

func (f *GenesysCloudProvider) Configure(ctx context.Context, request provider.ConfigureRequest, response *provider.ConfigureResponse) {
	var data GenesysCloudProviderModel

	response.Diagnostics.Append(request.Config.Get(ctx, &data)...)

	platformInstance := platform.GetPlatform()
	platformValidationErr := platformInstance.Validate()
	if platformValidationErr != nil {
		log.Printf("%v error during platform validation switching to defaults", platformValidationErr)
	}

	providerSourceRegistry := getRegistry(&platformInstance, f.Version)

	frameworkLog("Configuring provider schema attributes")
	f.AttributeEnvValues = readProviderEnvVars()
	f.configureAuthInfo(data)
	f.configureSdkDebugInfo(data)
	f.configureRootAttributes(data)
	f.configureProxyAttributes(data)
	f.configureGatewayAttributes(data)

	frameworkLog("Initialising SDK client pool")
	err := f.InitSDKClientPool()
	if err.HasError() {
		response.Diagnostics.AddError(fmt.Sprintf("%v", err), "Failed to init SDK client pool")
	}

	frameworkLog("Establishing current org")
	defaultConfig := platformclientv2.GetDefaultConfiguration()
	currentOrg, getOrgMeErr := getOrganizationMe(defaultConfig)
	if getOrgMeErr != nil { // plugin sdk diagnostic error
		response.Diagnostics.AddError(fmt.Sprintf("%v", getOrgMeErr), "Failed to establish current organisation.")
	}

	prl.InitPanicRecoveryLoggerInstance(data.LogStackTraces.ValueBool(), data.LogStackTracesFilePath.ValueString())

	meta := ProviderMeta{
		Version:      f.Version,
		Platform:     &platformInstance,
		Registry:     providerSourceRegistry,
		ClientConfig: defaultConfig,
		Domain:       getRegionDomain(f.AuthDetails.Region),
		Organization: currentOrg,
	}

	if currentOrg != nil && currentOrg.DefaultCountryCode != nil {
		meta.DefaultCountryCode = *currentOrg.DefaultCountryCode
	}

	f.Meta = meta

	response.ResourceData = &f
}

func (f *GenesysCloudProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		// TODO: add a datasource
	}
}

func (f *GenesysCloudProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		routing_wrapupcode_v2.NewWrapupCodeResource,
	}
}

func frameworkLog(s string) {
	const logPrefix = "(Framework) "
	log.Println(logPrefix, s)
}
