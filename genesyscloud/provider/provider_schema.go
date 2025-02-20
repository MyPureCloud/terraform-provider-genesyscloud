package provider

import (
	"fmt"
	"regexp"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

const (
	// Provider environment variables
	logStackTracesEnvVar         = "GENESYSCLOUD_LOG_STACK_TRACES"
	logStackTracesFilePathEnvVar = "GENESYSCLOUD_LOG_STACK_TRACES_FILE_PATH"

	// Provider attribute keys
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
			DefaultFunc: schema.EnvDefaultFunc("GENESYSCLOUD_ACCESS_TOKEN", nil),
			Description: "A string that the OAuth client uses to make requests. Can be set with the `GENESYSCLOUD_ACCESS_TOKEN` environment variable.",
		},
		"oauthclient_id": {
			Type:        schema.TypeString,
			Optional:    true,
			DefaultFunc: schema.EnvDefaultFunc("GENESYSCLOUD_OAUTHCLIENT_ID", nil),
			Description: "OAuthClient ID found on the OAuth page of Admin UI. Can be set with the `GENESYSCLOUD_OAUTHCLIENT_ID` environment variable.",
		},
		"oauthclient_secret": {
			Type:        schema.TypeString,
			Optional:    true,
			DefaultFunc: schema.EnvDefaultFunc("GENESYSCLOUD_OAUTHCLIENT_SECRET", nil),
			Description: "OAuthClient secret found on the OAuth page of Admin UI. Can be set with the `GENESYSCLOUD_OAUTHCLIENT_SECRET` environment variable.",
			Sensitive:   true,
		},
		"aws_region": {
			Type:         schema.TypeString,
			Optional:     true,
			DefaultFunc:  schema.EnvDefaultFunc("GENESYSCLOUD_REGION", nil),
			Description:  "AWS region where org exists. e.g. us-east-1. Can be set with the `GENESYSCLOUD_REGION` environment variable.",
			ValidateFunc: validation.StringInSlice(getAllowedRegions(), true),
		},
		"sdk_debug": {
			Type:        schema.TypeBool,
			Optional:    true,
			DefaultFunc: schema.EnvDefaultFunc("GENESYSCLOUD_SDK_DEBUG", false),
			Description: "Enables debug tracing in the Genesys Cloud SDK. Output will be written to the local file 'sdk_debug.log'. Can be set with the `GENESYSCLOUD_SDK_DEBUG` environment variable.",
		},
		"sdk_debug_format": {
			Type:         schema.TypeString,
			Optional:     true,
			DefaultFunc:  schema.EnvDefaultFunc("GENESYSCLOUD_SDK_DEBUG_FORMAT", "Text"),
			Description:  "Specifies the data format of the 'sdk_debug.log'. Only applicable if sdk_debug is true. Can be set with the `GENESYSCLOUD_SDK_DEBUG_FORMAT` environment variable. Default value is Text.",
			ValidateFunc: validation.StringInSlice([]string{"Text", "Json"}, false),
		},
		"sdk_debug_file_path": {
			Type:         schema.TypeString,
			Optional:     true,
			DefaultFunc:  schema.EnvDefaultFunc("GENESYSCLOUD_SDK_DEBUG_FILE_PATH", "sdk_debug.log"),
			Description:  "Specifies the file path for the log file. Can be set with the `GENESYSCLOUD_SDK_DEBUG_FILE_PATH` environment variable. Default value is sdk_debug.log",
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
			DefaultFunc:  schema.EnvDefaultFunc("GENESYSCLOUD_TOKEN_POOL_SIZE", DefaultMaxClients),
			Description:  "Max number of OAuth tokens in the token pool. Can be set with the `GENESYSCLOUD_TOKEN_POOL_SIZE` environment variable.",
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
						DefaultFunc: schema.EnvDefaultFunc("GENESYSCLOUD_GATEWAY_PORT", nil),
						Description: "Port for the gateway can be set with the `GENESYSCLOUD_GATEWAY_PORT` environment variable.",
					},
					"host": {
						Type:        schema.TypeString,
						Optional:    true,
						DefaultFunc: schema.EnvDefaultFunc("GENESYSCLOUD_GATEWAY_HOST", nil),
						Description: "Host for the gateway can be set with the `GENESYSCLOUD_GATEWAY_HOST` environment variable.",
					},
					"protocol": {
						Type:        schema.TypeString,
						Optional:    true,
						DefaultFunc: schema.EnvDefaultFunc("GENESYSCLOUD_GATEWAY_PROTOCOL", nil),
						Description: "Protocol for the gateway can be set with the `GENESYSCLOUD_GATEWAY_PROTOCOL` environment variable.",
					},
					"path_params": {
						Type:     schema.TypeSet,
						Optional: true,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"path_name": {
									Type:        schema.TypeString,
									Required:    true,
									Description: "Path name for Gateway Path Params can be set with the `GENESYSCLOUD_GATEWAY_PATH_NAME` environment variable.",
									DefaultFunc: schema.EnvDefaultFunc("GENESYSCLOUD_GATEWAY_PATH_NAME", nil),
								},
								"path_value": {
									Type:        schema.TypeString,
									Required:    true,
									Description: "Path value for Gateway Path Params can be set with the `GENESYSCLOUD_GATEWAY_PATH_VALUE` environment variable.",
									DefaultFunc: schema.EnvDefaultFunc("GENESYSCLOUD_GATEWAY_PATH_VALUE", nil),
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
									Optional:    true,
									DefaultFunc: schema.EnvDefaultFunc("GENESYSCLOUD_GATEWAY_AUTH_USERNAME", nil),
									Description: "UserName for the Auth can be set with the `GENESYSCLOUD_PROXY_AUTH_USERNAME` environment variable.",
								},
								"password": {
									Type:        schema.TypeString,
									Optional:    true,
									DefaultFunc: schema.EnvDefaultFunc("GENESYSCLOUD_GATEWAY_AUTH_PASSWORD", nil),
									Description: "Password for the Auth can be set with the `GENESYSCLOUD_PROXY_AUTH_PASSWORD` environment variable.",
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
						DefaultFunc: schema.EnvDefaultFunc("GENESYSCLOUD_PROXY_PORT", nil),
						Description: "Port for the proxy can be set with the `GENESYSCLOUD_PROXY_PORT` environment variable.",
					},
					"host": {
						Type:        schema.TypeString,
						Optional:    true,
						DefaultFunc: schema.EnvDefaultFunc("GENESYSCLOUD_PROXY_HOST", nil),
						Description: "Host for the proxy can be set with the `GENESYSCLOUD_PROXY_HOST` environment variable.",
					},
					"protocol": {
						Type:        schema.TypeString,
						Optional:    true,
						DefaultFunc: schema.EnvDefaultFunc("GENESYSCLOUD_PROXY_PROTOCOL", nil),
						Description: "Protocol for the proxy can be set with the `GENESYSCLOUD_PROXY_PROTOCOL` environment variable.",
					},
					"auth": {
						Type:     schema.TypeSet,
						Optional: true,
						MaxItems: 1,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"username": {
									Type:        schema.TypeString,
									Optional:    true,
									DefaultFunc: schema.EnvDefaultFunc("GENESYSCLOUD_PROXY_AUTH_USERNAME", nil),
									Description: "UserName for the Auth can be set with the `GENESYSCLOUD_PROXY_AUTH_USERNAME` environment variable.",
								},
								"password": {
									Type:        schema.TypeString,
									Optional:    true,
									DefaultFunc: schema.EnvDefaultFunc("GENESYSCLOUD_PROXY_AUTH_PASSWORD", nil),
									Description: "Password for the Auth can be set with the `GENESYSCLOUD_PROXY_AUTH_PASSWORD` environment variable.",
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
