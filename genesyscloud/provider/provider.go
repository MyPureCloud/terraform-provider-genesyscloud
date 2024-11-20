package provider

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/mypurecloud/platform-client-sdk-go/v146/platformclientv2"
)

var orgDefaultCountryCode string

func init() {
	// Set descriptions to support markdown syntax, this will be used in document generation
	// and the language server.
	// providerResources = make(map[string]*schema.Resource)
	// providerDataSources = make(map[string]*schema.Resource)
	schema.DescriptionKind = schema.StringMarkdown

	// Customize the content of descriptions when output.
	schema.SchemaDescriptionBuilder = func(s *schema.Schema) string {
		desc := s.Description
		if s.Default != nil {
			desc += fmt.Sprintf(" Defaults to `%v`.", s.Default)
		}
		return strings.TrimSpace(desc)
	}

}

// New initializes the provider schema
func New(version string, providerResources map[string]*schema.Resource, providerDataSources map[string]*schema.Resource) func() *schema.Provider {
	return func() *schema.Provider {

		/*
		   The next two lines are important.  We have to make sure the Terraform provider has their own deep copies of the resource
		   and data source maps.  If you do not do a deep copy and try to pass in the original maps, you open yourself up to race conditions
		   because they map are being read and written to concurrently.
		*/
		copiedResources := make(map[string]*schema.Resource)
		for k, v := range providerResources {
			copiedResources[k] = v
		}

		copiedDataSources := make(map[string]*schema.Resource)
		for k, v := range providerDataSources {
			copiedDataSources[k] = v
		}

		return &schema.Provider{
			Schema: map[string]*schema.Schema{
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
				"token_pool_size": {
					Type:         schema.TypeInt,
					Optional:     true,
					DefaultFunc:  schema.EnvDefaultFunc("GENESYSCLOUD_TOKEN_POOL_SIZE", 10),
					Description:  "Max number of OAuth tokens in the token pool. Can be set with the `GENESYSCLOUD_TOKEN_POOL_SIZE` environment variable.",
					ValidateFunc: validation.IntBetween(1, 20),
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
			},
			ResourcesMap:         copiedResources,
			DataSourcesMap:       copiedDataSources,
			ConfigureContextFunc: configure(version),
		}
	}
}

type ProviderMeta struct {
	Version      string
	ClientConfig *platformclientv2.Configuration
	Domain       string
	Organization *platformclientv2.Organization
}

func configure(version string) schema.ConfigureContextFunc {
	return func(context context.Context, data *schema.ResourceData) (interface{}, diag.Diagnostics) {
		err := InitSDKClientPool(data.Get("token_pool_size").(int), version, data)
		if err != nil {
			return nil, err
		}

		defaultConfig := platformclientv2.GetDefaultConfiguration()

		currentOrg, err := getOrganizationMe(defaultConfig)
		if err != nil {
			return nil, err
		}
		orgDefaultCountryCode = *currentOrg.DefaultCountryCode

		return &ProviderMeta{
			Version:      version,
			ClientConfig: defaultConfig,
			Domain:       getRegionDomain(data.Get("aws_region").(string)),
			Organization: currentOrg,
		}, nil
	}
}

func getOrganizationMe(defaultConfig *platformclientv2.Configuration) (*platformclientv2.Organization, diag.Diagnostics) {
	orgApiClient := platformclientv2.NewOrganizationApiWithConfig(defaultConfig)
	me, _, err := orgApiClient.GetOrganizationsMe()
	if err != nil {
		return nil, diag.FromErr(err)
	}
	return me, nil
}

func getRegionMap() map[string]string {
	return map[string]string{
		"dca":            "inindca.com",
		"tca":            "inintca.com",
		"us-east-1":      "mypurecloud.com",
		"us-east-2":      "use2.us-gov-pure.cloud",
		"us-west-2":      "usw2.pure.cloud",
		"eu-west-1":      "mypurecloud.ie",
		"eu-west-2":      "euw2.pure.cloud",
		"ap-southeast-2": "mypurecloud.com.au",
		"ap-northeast-1": "mypurecloud.jp",
		"eu-central-1":   "mypurecloud.de",
		"ca-central-1":   "cac1.pure.cloud",
		"ap-northeast-2": "apne2.pure.cloud",
		"ap-south-1":     "aps1.pure.cloud",
		"sa-east-1":      "sae1.pure.cloud",
		"ap-northeast-3": "apne3.pure.cloud",
		"eu-central-2":   "euc2.pure.cloud",
		"me-central-1":   "mec1.pure.cloud",
	}
}

func getAllowedRegions() []string {
	regionMap := getRegionMap()
	regionKeys := make([]string, 0, len(regionMap))
	for k := range regionMap {
		regionKeys = append(regionKeys, k)
	}
	return regionKeys
}

func getRegionDomain(region string) string {
	return getRegionMap()[strings.ToLower(region)]
}

func GetRegionBasePath(region string) string {
	return "https://api." + getRegionDomain(region)
}

func InitClientConfig(data *schema.ResourceData, version string, config *platformclientv2.Configuration) diag.Diagnostics {
	accessToken := data.Get("access_token").(string)
	oauthclientID := data.Get("oauthclient_id").(string)
	oauthclientSecret := data.Get("oauthclient_secret").(string)
	basePath := GetRegionBasePath(data.Get("aws_region").(string))
	config.BasePath = basePath

	diagErr := setUpSDKLogging(data, config)
	if diagErr != nil {
		return diagErr
	}

	setupProxy(data, config)
	setupGateway(data, config)

	config.AddDefaultHeader("User-Agent", "GC Terraform Provider/"+version)
	config.RetryConfiguration = &platformclientv2.RetryConfiguration{
		RetryWaitMin: time.Second * 1,
		RetryWaitMax: time.Second * 30,
		RetryMax:     20,
		RequestLogHook: func(request *http.Request, count int) {
			sdkDebugRequest := newSDKDebugRequest(request, count)
			request.Header.Set("TF-Correlation-Id", sdkDebugRequest.TransactionId)
			err, jsonStr := sdkDebugRequest.ToJSON()

			if err != nil {
				log.Printf("WARNING: Unable to log RequestLogHook: %s", err)
			}
			log.Printf(jsonStr)
		},
		ResponseLogHook: func(response *http.Response) {
			sdkDebugResponse := newSDKDebugResponse(response)
			err, jsonStr := sdkDebugResponse.ToJSON()

			if err != nil {
				log.Printf("WARNING: Unable to log ResponseLogHook: %s", err)
			}
			log.Printf(jsonStr)
		},
	}

	if accessToken != "" {
		log.Print("Setting access token set on configuration instance.")
		config.AccessToken = accessToken
	} else {
		config.AutomaticTokenRefresh = true // Enable automatic token refreshing

		return withRetries(context.Background(), time.Minute, func() *retry.RetryError {
			err := config.AuthorizeClientCredentials(oauthclientID, oauthclientSecret)
			if err != nil {
				if !strings.Contains(err.Error(), "Auth Error: 400 - invalid_request (rate limit exceeded;") {
					return retry.NonRetryableError(fmt.Errorf("failed to authorize Genesys Cloud client credentials: %v", err))
				}
				return retry.RetryableError(fmt.Errorf("exhausted retries on Genesys Cloud client credentials. %v", err))
			}

			return nil
		})
	}

	log.Printf("Initialized Go SDK Client. Debug=%t", data.Get("sdk_debug").(bool))
	return nil
}

func withRetries(ctx context.Context, timeout time.Duration, method func() *retry.RetryError) diag.Diagnostics {
	err := diag.FromErr(retry.RetryContext(ctx, timeout, method))
	if err != nil && strings.Contains(fmt.Sprintf("%v", err), "timeout while waiting for state to become") {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		return withRetries(ctx, timeout, method)
	}
	return err
}

func setUpSDKLogging(data *schema.ResourceData, config *platformclientv2.Configuration) diag.Diagnostics {
	sdkDebugFilePath := data.Get("sdk_debug_file_path").(string)
	if data.Get("sdk_debug").(bool) {
		config.LoggingConfiguration = &platformclientv2.LoggingConfiguration{
			LogLevel:        platformclientv2.LTrace,
			LogRequestBody:  true,
			LogResponseBody: true,
		}
		config.LoggingConfiguration.SetLogToConsole(false)
		config.LoggingConfiguration.SetLogFilePath(sdkDebugFilePath)

		dir, _ := filepath.Split(sdkDebugFilePath)
		if err := os.MkdirAll(dir, os.ModePerm); os.IsExist(err) {
			return diag.Errorf("error while creating filepath for %s: %s", sdkDebugFilePath, err)
		}

		if format := data.Get("sdk_debug_format"); format == "Json" {
			config.LoggingConfiguration.SetLogFormat(platformclientv2.JSON)
		} else {
			config.LoggingConfiguration.SetLogFormat(platformclientv2.Text)
		}
	}
	return nil
}

func setupProxy(data *schema.ResourceData, config *platformclientv2.Configuration) {
	proxySet := data.Get("proxy").(*schema.Set)
	for _, proxyObj := range proxySet.List() {
		proxy := proxyObj.(map[string]interface{})

		// Retrieve the values of the `host`, `port`, and `protocol` attributes
		host := proxy["host"].(string)
		port := proxy["port"].(string)
		protocol := proxy["protocol"].(string)

		config.ProxyConfiguration = &platformclientv2.ProxyConfiguration{}

		config.ProxyConfiguration.Host = host
		config.ProxyConfiguration.Port = port
		config.ProxyConfiguration.Protocol = protocol

		authSet := proxy["auth"].(*schema.Set)
		authList := authSet.List()

		for _, authElement := range authList {
			auth := authElement.(map[string]interface{})
			username := auth["username"].(string)
			password := auth["password"].(string)
			config.ProxyConfiguration.Auth = &platformclientv2.Auth{}

			config.ProxyConfiguration.Auth.UserName = username
			config.ProxyConfiguration.Auth.Password = password
		}
	}
}

func setupGateway(data *schema.ResourceData, config *platformclientv2.Configuration) {
	gatewaySet := data.Get("gateway").(*schema.Set)
	for _, gatewayObj := range gatewaySet.List() {
		gateway := gatewayObj.(map[string]interface{})

		// Retrieve the values of the `host`, `port`, and `protocol` attributes
		host := gateway["host"].(string)
		port := gateway["port"].(string)
		protocol := gateway["protocol"].(string)
		config.GateWayConfiguration = &platformclientv2.GateWayConfiguration{}

		config.GateWayConfiguration.Host = host
		config.GateWayConfiguration.Port = port
		config.GateWayConfiguration.Protocol = protocol

		paramSet := gateway["path_params"].(*schema.Set)
		paramList := paramSet.List()

		for _, paramElement := range paramList {
			param := paramElement.(map[string]interface{})

			pathName := param["path_name"].(string)
			pathValue := param["path_value"].(string)

			config.GateWayConfiguration.PathParams = append(config.GateWayConfiguration.PathParams, &platformclientv2.PathParams{
				PathName:  pathName,
				PathValue: pathValue,
			})
		}

		authSet := gateway["auth"].(*schema.Set)
		authList := authSet.List()

		for _, authElement := range authList {
			auth := authElement.(map[string]interface{})
			username := auth["username"].(string)
			password := auth["password"].(string)
			config.GateWayConfiguration.Auth = &platformclientv2.Auth{}

			config.GateWayConfiguration.Auth.UserName = username
			config.GateWayConfiguration.Auth.Password = password
		}
	}
}

func AuthorizeSdk() (*platformclientv2.Configuration, error) {
	// Create new config
	sdkConfig := platformclientv2.GetDefaultConfiguration()

	v, exists := os.LookupEnv("TF_UNIT")
	if exists && v != "" {
		log.Printf("TF_UNIT environment is set.  No authorization of the SDK has occurred")
		return sdkConfig, nil
	}

	sdkConfig.BasePath = GetRegionBasePath(os.Getenv("GENESYSCLOUD_REGION"))

	diagErr := withRetries(context.Background(), time.Minute, func() *retry.RetryError {
		err := sdkConfig.AuthorizeClientCredentials(os.Getenv("GENESYSCLOUD_OAUTHCLIENT_ID"), os.Getenv("GENESYSCLOUD_OAUTHCLIENT_SECRET"))
		if err != nil {
			if !strings.Contains(err.Error(), "Auth Error: 400 - invalid_request (rate limit exceeded;") {
				return retry.NonRetryableError(fmt.Errorf("failed to authorize Genesys Cloud client credentials: %v", err))
			}
			return retry.RetryableError(fmt.Errorf("exhausted retries on Genesys Cloud client credentials. %v", err))
		}

		return nil
	})
	if diagErr != nil {
		return sdkConfig, fmt.Errorf("%v", diagErr)
	}

	return sdkConfig, nil
}
