package provider

import "github.com/hashicorp/terraform-plugin-framework/types"

type GenesysCloudProviderModel struct {
	AccessToken            types.String   `tfsdk:"access_token"`
	OAuthClientId          types.String   `tfsdk:"oauthclient_id"`
	OAuthClientSecret      types.String   `tfsdk:"oauthclient_secret"`
	AwsRegion              types.String   `tfsdk:"aws_region"`
	SdkDebug               types.Bool     `tfsdk:"sdk_debug"`
	SdkDebugFormat         types.String   `tfsdk:"sdk_debug_format"`
	SdkDebugFilePath       types.String   `tfsdk:"sdk_debug_file_path"`
	TokenPoolSize          types.Int32    `tfsdk:"token_pool_size"`
	LogStackTraces         types.Bool     `tfsdk:"log_stack_traces"`
	LogStackTracesFilePath types.String   `tfsdk:"log_stack_traces_file_path"`
	Gateway                []GatewayModel `tfsdk:"gateway"`
	Proxy                  []ProxyModel   `tfsdk:"proxy"`
}

type GatewayModel struct {
	Port       types.String     `tfsdk:"port"`
	Host       types.String     `tfsdk:"host"`
	Protocol   types.String     `tfsdk:"protocol"`
	PathParams []PathParamModel `tfsdk:"path_params"`
	Auth       []AuthModel      `tfsdk:"auth"`
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
	Port     types.String `tfsdk:"port"`
	Host     types.String `tfsdk:"host"`
	Protocol types.String `tfsdk:"protocol"`
	Auth     []AuthModel  `tfsdk:"auth"`
}
