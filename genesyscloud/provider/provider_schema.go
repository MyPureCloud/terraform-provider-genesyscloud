package provider

import (
	"fmt"
	"regexp"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

// Provider attribute keys
const (
	AttrTokenPoolSize       = "token_pool_size"
	AttrTokenAcquireTimeout = "token_acquire_timeout"
	AttrTokenInitTimeout    = "token_init_timeout"
	AttrSdkClientPoolDebug  = "sdk_client_pool_debug"
)

func ProviderSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"access_token": {
			Type:        schema.TypeString,
			Optional:    true,
			Sensitive:   true,
			DefaultFunc: schema.EnvDefaultFunc(accessTokenEnvVar, nil),
			Description: fmt.Sprintf("A string that the OAuth client uses to make requests. Can be set with the `%s` environment variable.", accessTokenEnvVar),
		},
		"oauthclient_id": {
			Type:        schema.TypeString,
			Optional:    true,
			DefaultFunc: schema.EnvDefaultFunc(clientIdEnvVar, nil),
			Description: fmt.Sprintf("OAuthClient ID found on the OAuth page of Admin UI. Can be set with the `%s` environment variable.", clientIdEnvVar),
		},
		"oauthclient_secret": {
			Type:        schema.TypeString,
			Optional:    true,
			DefaultFunc: schema.EnvDefaultFunc(clientSecretEnvVar, nil),
			Description: fmt.Sprintf("OAuthClient secret found on the OAuth page of Admin UI. Can be set with the `%s` environment variable.", clientSecretEnvVar),
			Sensitive:   true,
		},
		"aws_region": {
			Type:         schema.TypeString,
			Optional:     true,
			DefaultFunc:  schema.EnvDefaultFunc(regionEnvVar, nil),
			Description:  fmt.Sprintf("AWS region where org exists. e.g. us-east-1. Can be set with the `%s` environment variable.", regionEnvVar),
			ValidateFunc: validation.StringInSlice(getAllowedRegions(), true),
		},
		"sdk_debug": {
			Type:        schema.TypeBool,
			Optional:    true,
			DefaultFunc: schema.EnvDefaultFunc(sdkDebugEnvVar, false),
			Description: fmt.Sprintf("Enables debug tracing in the Genesys Cloud SDK. Output will be written to `sdk_debug_file_path`. Can be set with the `%s` environment variable.", sdkDebugEnvVar),
		},
		"sdk_debug_format": {
			Type:         schema.TypeString,
			Optional:     true,
			DefaultFunc:  schema.EnvDefaultFunc(sdkDebugFormatEnvVar, "Text"),
			Description:  fmt.Sprintf("Specifies the data format of the 'sdk_debug.log'. Only applicable if sdk_debug is true. Can be set with the `%s` environment variable. Default value is Text.", sdkDebugFormatEnvVar),
			ValidateFunc: validation.StringInSlice([]string{"Text", "Json"}, false),
		},
		"sdk_debug_file_path": {
			Type:         schema.TypeString,
			Optional:     true,
			DefaultFunc:  schema.EnvDefaultFunc(sdkDebugFilePathEnvVar, "sdk_debug.log"),
			Description:  fmt.Sprintf("Specifies the file path for the log file. Can be set with the `%s` environment variable. Default value is sdk_debug.log", sdkDebugFilePathEnvVar),
			ValidateFunc: validation.StringDoesNotMatch(regexp.MustCompile("^(|\\s+)$"), "Invalid File path "),
		},
		AttrSdkClientPoolDebug: {
			Type:        schema.TypeBool,
			Optional:    true,
			DefaultFunc: schema.EnvDefaultFunc("GENESYSCLOUD_SDK_CLIENT_POOL_DEBUG", false),
			Description: "Enables debug tracing in the Genesys Cloud SDK client pool. Output will be written to standard log output. Can be set with the `GENESYSCLOUD_SDK_CLIENT_POOL_DEBUG` environment variable.",
		},
		AttrTokenPoolSize: {
			Type:         schema.TypeInt,
			Optional:     true,
			DefaultFunc:  schema.EnvDefaultFunc(tokenPoolSizeEnvVar, DefaultMaxClients),
			Description:  fmt.Sprintf("Max number of OAuth tokens in the token pool. Can be set with the `%s` environment variable.", tokenPoolSizeEnvVar),
			ValidateFunc: validation.IntBetween(MinClients, MaxClients),
		},
		AttrTokenAcquireTimeout: {
			Type:         schema.TypeString,
			Optional:     true,
			DefaultFunc:  schema.EnvDefaultFunc("GENESYSCLOUD_TOKEN_ACQUIRE_TIMEOUT", DefaultAcquireTimeout.String()),
			Description:  "Timeout for acquiring a token from the pool. Can be set with the `GENESYSCLOUD_TOKEN_ACQUIRE_TIMEOUT` environment variable.",
			ValidateFunc: validateDuration,
		},
		AttrTokenInitTimeout: {
			Type:         schema.TypeString,
			Optional:     true,
			DefaultFunc:  schema.EnvDefaultFunc("GENESYSCLOUD_TOKEN_INIT_TIMEOUT", DefaultInitTimeout.String()),
			Description:  "Timeout for initializing the token pool. Can be set with the `GENESYSCLOUD_TOKEN_INIT_TIMEOUT` environment variable.",
			ValidateFunc: validateDuration,
		},
		"log_stack_traces": {
			Type:        schema.TypeBool,
			Optional:    true,
			DefaultFunc: schema.EnvDefaultFunc(logStackTracesEnvVar, false),
			Description: fmt.Sprintf(`If true, stack traces will be logged to a file instead of crashing the provider, whenever possible.
If the stack trace occurs within the create context and before the ID is set in the schema object, then the command will fail with the message
"Root object was present, but now absent." Can be set with the %s environment variable. **WARNING**: This is a debugging feature that may cause your Terraform state to become out of sync with the API.
If you encounter any stack traces, please report them so we can address the underlying issues.`, logStackTracesEnvVar),
		},
		"log_stack_traces_file_path": {
			Type:             schema.TypeString,
			Optional:         true,
			Description:      fmt.Sprintf("Specifies the file path for the stack trace logs. Can be set with the `%s` environment variable. Default value is genesyscloud_stack_traces.log", logStackTracesFilePathEnvVar),
			DefaultFunc:      schema.EnvDefaultFunc(logStackTracesFilePathEnvVar, "genesyscloud_stack_traces.log"),
			ValidateDiagFunc: validateLogFilePath,
		},
		"gateway": {
			Type:     schema.TypeSet,
			Optional: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"port": {
						Type:        schema.TypeString,
						Optional:    true,
						DefaultFunc: schema.EnvDefaultFunc(gatewayPortEnvVar, nil),
						Description: fmt.Sprintf("Port for the gateway can be set with the `%s` environment variable.", gatewayPortEnvVar),
					},
					"host": {
						Type:        schema.TypeString,
						Optional:    true,
						DefaultFunc: schema.EnvDefaultFunc(gatewayHostEnvVar, nil),
						Description: fmt.Sprintf("Host for the gateway can be set with the `%s` environment variable.", gatewayHostEnvVar),
					},
					"protocol": {
						Type:        schema.TypeString,
						Optional:    true,
						DefaultFunc: schema.EnvDefaultFunc(gatewayProtocolEnvVar, nil),
						Description: fmt.Sprintf("Protocol for the gateway can be set with the `%s` environment variable.", gatewayProtocolEnvVar),
					},
					"path_params": {
						Type:     schema.TypeSet,
						Optional: true,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"path_name": {
									Type:        schema.TypeString,
									Required:    true,
									Description: fmt.Sprintf("Path name for Gateway Path Params can be set with the `%s` environment variable.", gatewayPathParamsNameEnvVar),
									DefaultFunc: schema.EnvDefaultFunc(gatewayPathParamsNameEnvVar, nil),
								},
								"path_value": {
									Type:        schema.TypeString,
									Required:    true,
									Description: fmt.Sprintf("Path value for Gateway Path Params can be set with the `%s` environment variable.", gatewayPathParamsValueEnvVar),
									DefaultFunc: schema.EnvDefaultFunc(gatewayPathParamsValueEnvVar, nil),
								},
							},
						},
					},
					"auth": {
						Type:     schema.TypeSet,
						Optional: true,
						MaxItems: 1,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"username": {
									Type:        schema.TypeString,
									Required:    true,
									DefaultFunc: schema.EnvDefaultFunc(gatewayAuthUsernameEnvVar, nil),
									Description: fmt.Sprintf("UserName for the Auth can be set with the `%s` environment variable.", gatewayAuthUsernameEnvVar),
								},
								"password": {
									Type:        schema.TypeString,
									Optional:    true,
									Sensitive:   true,
									DefaultFunc: schema.EnvDefaultFunc(gatewayAuthPasswordEnvVar, nil),
									Description: fmt.Sprintf("Password for the Auth can be set with the `%s` environment variable.", gatewayAuthPasswordEnvVar),
								},
							},
						},
					},
				},
			},
		},
		"proxy": {
			Type:     schema.TypeSet,
			Optional: true,
			MaxItems: 1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"port": {
						Type:        schema.TypeString,
						Optional:    true,
						DefaultFunc: schema.EnvDefaultFunc(proxyPortEnvVar, nil),
						Description: fmt.Sprintf("Port for the proxy can be set with the `%s` environment variable.", proxyPortEnvVar),
					},
					"host": {
						Type:        schema.TypeString,
						Optional:    true,
						DefaultFunc: schema.EnvDefaultFunc(proxyHostEnvVar, nil),
						Description: fmt.Sprintf("Host for the proxy can be set with the `%s` environment variable.", proxyHostEnvVar),
					},
					"protocol": {
						Type:        schema.TypeString,
						Optional:    true,
						DefaultFunc: schema.EnvDefaultFunc(proxyProtocolEnvVar, nil),
						Description: fmt.Sprintf("Protocol for the proxy can be set with the `%s` environment variable.", proxyProtocolEnvVar),
					},
					"auth": {
						Type:     schema.TypeSet,
						Optional: true,
						MaxItems: 1,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"username": {
									Type:        schema.TypeString,
									Required:    true,
									DefaultFunc: schema.EnvDefaultFunc(proxyAuthUsernameEnvVar, nil),
									Description: fmt.Sprintf("UserName for the Auth can be set with the `%s` environment variable.", proxyAuthUsernameEnvVar),
								},
								"password": {
									Type:        schema.TypeString,
									Optional:    true,
									Sensitive:   true,
									DefaultFunc: schema.EnvDefaultFunc(proxyAuthPasswordEnvVar, nil),
									Description: fmt.Sprintf("Password for the Auth can be set with the `%s` environment variable.", proxyAuthPasswordEnvVar),
								},
							},
						},
					},
				},
			},
		},
	}
}

func validateDuration(i interface{}, k string) ([]string, []error) {
	v, ok := i.(string)
	if !ok {
		return nil, []error{fmt.Errorf("expected type of %s to be string", k)}
	}
	_, err := time.ParseDuration(v)
	if err != nil {
		return nil, []error{fmt.Errorf("expected %s to be a valid duration string: %v", k, err)}
	}
	return nil, nil
}
