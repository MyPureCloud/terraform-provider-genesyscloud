package provider

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"terraform-provider-genesyscloud/genesyscloud/platform"

	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/mypurecloud/platform-client-sdk-go/v150/platformclientv2"
)

var (
	// ensure the implementation satisfies the expected interfaces
	_ provider.Provider = &GenesysCloudProvider{}
)

type GenesysCloudProvider struct {
	Meta          ProviderMeta
	Version       string
	SdkClientPool SDKClientPool

	AuthDetails *AuthDetails
}

type AuthDetails struct {
	AccessToken  string
	ClientId     string
	ClientSecret string
	Region       string
}

type GenesysCloudProviderModel struct {
	AccessToken            types.String  `tfsdk:"access_token"`
	OAuthClientId          types.String  `tfsdk:"oauthclient_id"`
	OAuthClientSecret      types.String  `tfsdk:"oauthclient_secret"`
	AwsRegion              types.String  `tfsdk:"aws_region"`
	SdkDebug               types.Bool    `tfsdk:"sdk_debug"`
	SdkDebugFormat         types.String  `tfsdk:"sdk_debug_format"`
	SdkDebugFilePath       types.String  `tfsdk:"sdk_debug_file_path"`
	TokenPoolSize          types.Int32   `tfsdk:"token_pool_size"`
	LogStackTraces         types.Bool    `tfsdk:"log_stack_traces"`
	LogStackTracesFilePath types.String  `tfsdk:"log_stack_traces_file_path"`
	Gateway                *GatewayModel `tfsdk:"gateway"`
	Proxy                  *ProxyModel   `tfsdk:"proxy"`
}

type GatewayModel struct {
	Port       types.String     `tfsdk:"port"`
	Host       types.String     `tfsdk:"host"`
	Protocol   types.String     `tfsdk:"protocol"`
	PathParams []PathParamModel `tfsdk:"path_params"`
	Auth       *AuthModel       `tfsdk:"auth"`
}

type PathParamModel struct {
	PathName  types.String `tfsdk:"path_name"`
	PathValue types.String `tfsdk:"path_value"`
}

type AuthModel struct {
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
}

type ProxyModel struct {
	Port     types.String    `tfsdk:"port"`
	Host     types.String    `tfsdk:"host"`
	Protocol types.String    `tfsdk:"protocol"`
	Auth     *ProxyAuthModel `tfsdk:"auth"`
}

type ProxyAuthModel struct {
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
}

func NewFrameWorkProvider(version string) func() provider.Provider {
	return func() provider.Provider {
		return &GenesysCloudProvider{
			Version: version,
		}
	}
}

func (f GenesysCloudProvider) Metadata(ctx context.Context, request provider.MetadataRequest, response *provider.MetadataResponse) {
	//TODO implement me
	panic("implement me")
}

func (f GenesysCloudProvider) Schema(_ context.Context, request provider.SchemaRequest, response *provider.SchemaResponse) {
	response.Schema = schema.Schema{
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
				MarkdownDescription: fmt.Sprintf("AWS region where org exists. e.g. us-east-1. Can be set with the `%s` environment variable. Defaults to \"%s\"", regionEnvVar, awsRegionDefaultValue),
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
				MarkdownDescription: fmt.Sprintf("Specifies the data format of the 'sdk_debug.log'. Only applicable if sdk_debug is true. Can be set with the `%s` environment variable. Default value is Text.", sdkDebugFormatEnvVar),
				Validators: []validator.String{
					stringvalidator.OneOf("Text", "Json"),
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
			"gateway": schema.SingleNestedAttribute{
				Optional: true,
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
					"path_params": schema.SetNestedAttribute{
						Optional: true,
						NestedObject: schema.NestedAttributeObject{
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
					"auth": schema.SingleNestedAttribute{
						Optional: true,
						Attributes: map[string]schema.Attribute{
							"username": schema.StringAttribute{
								Optional:            true,
								MarkdownDescription: fmt.Sprintf("UserName for the Auth can be set with the `%s` environment variable.", gatewayAuthUsernameEnvVar),
							},
							"password": schema.StringAttribute{
								Optional:            false,
								Sensitive:           true,
								MarkdownDescription: fmt.Sprintf("Password for the Auth can be set with the `%s` environment variable.", gatewayAuthPasswordEnvVar),
							},
						},
					},
				},
			},
			"proxy": schema.SingleNestedAttribute{
				Optional: true,
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
					"auth": schema.SingleNestedAttribute{
						Optional: true,
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
		},
	}
}

func (f GenesysCloudProvider) Configure(ctx context.Context, request provider.ConfigureRequest, response *provider.ConfigureResponse) {
	var data GenesysCloudProviderModel

	// TODO: read all env variables
	providerEnvValues := readProviderEnvVars()

	response.Diagnostics.Append(request.Config.Get(ctx, &data)...)

	// TODO: use data values if env vars are not set
	var authDetails AuthDetails

	if data.AccessToken.ValueString() != "" {
		authDetails.AccessToken = data.AccessToken.ValueString()
	} else if providerEnvValues.accessToken != "" {
		authDetails.AccessToken = providerEnvValues.accessToken
	}

	if data.OAuthClientId.ValueString() != "" {
		authDetails.AccessToken = data.OAuthClientId.ValueString()
	} else if providerEnvValues.clientId != "" {
		authDetails.AccessToken = providerEnvValues.clientId
	}

	if data.OAuthClientSecret.ValueString() != "" {
		authDetails.ClientSecret = data.OAuthClientSecret.ValueString()
	} else if providerEnvValues.clientSecret != "" {
		authDetails.ClientSecret = providerEnvValues.clientSecret
	}

	if data.AwsRegion.ValueString() != "" {
		authDetails.Region = data.AwsRegion.ValueString()
	} else if providerEnvValues.region != "" {
		authDetails.Region = providerEnvValues.region
	} else {
		authDetails.Region = awsRegionDefaultValue
	}

	f.AuthDetails = &authDetails

	platformInstance := platform.GetPlatform()
	platformValidationErr := platformInstance.Validate()
	if platformValidationErr != nil {
		log.Printf("%v error during platform validation switching to defaults", platformValidationErr)
	}

	providerSourceRegistry := getRegistry(&platformInstance, f.Version)

	err := f.InitSDKClientPool(data)
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("%v", err), "Failed to init SDK client pool")
		return
	}

	defaultConfig := platformclientv2.GetDefaultConfiguration()

	currentOrg, getOrgMeErr := getOrganizationMe(defaultConfig)
	if getOrgMeErr != nil {
		response.Diagnostics.AddError(fmt.Sprintf("%v", getOrgMeErr), "Failed to establish current organisation.")
	}

	// probably not necessary because this is being called in the Plugin SDK configure function
	//prl.InitPanicRecoveryLoggerInstance(data.LogStackTraces.ValueBool(), data.LogStackTracesFilePath.ValueString())

	meta := &ProviderMeta{
		Version:            f.Version,
		Platform:           &platformInstance,
		Registry:           providerSourceRegistry,
		ClientConfig:       defaultConfig,
		Domain:             getRegionDomain(data.AwsRegion.ValueString()),
		Organization:       currentOrg,
		DefaultCountryCode: *currentOrg.DefaultCountryCode,
	}

	f.Meta = *meta
}

func (f GenesysCloudProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	//TODO implement me
	panic("implement me")
}

func (f GenesysCloudProvider) Resources(ctx context.Context) []func() resource.Resource {
	//TODO implement me
	panic("implement me")
}
