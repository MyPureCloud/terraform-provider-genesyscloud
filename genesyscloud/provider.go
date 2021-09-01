package genesyscloud

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/mypurecloud/platform-client-sdk-go/v53/platformclientv2"
)

func init() {
	// Set descriptions to support markdown syntax, this will be used in document generation
	// and the language server.
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
func New(version string) func() *schema.Provider {
	return func() *schema.Provider {
		return &schema.Provider{
			Schema: map[string]*schema.Schema{
				"oauthclient_id": {
					Type:        schema.TypeString,
					Required:    true,
					DefaultFunc: schema.EnvDefaultFunc("GENESYSCLOUD_OAUTHCLIENT_ID", nil),
					Description: "OAuthClient ID found on the OAuth page of Admin UI. Can be set with the `GENESYSCLOUD_OAUTHCLIENT_ID` environment variable.",
				},
				"oauthclient_secret": {
					Type:        schema.TypeString,
					Required:    true,
					DefaultFunc: schema.EnvDefaultFunc("GENESYSCLOUD_OAUTHCLIENT_SECRET", nil),
					Description: "OAuthClient secret found on the OAuth page of Admin UI. Can be set with the `GENESYSCLOUD_OAUTHCLIENT_SECRET` environment variable.",
					Sensitive:   true,
				},
				"aws_region": {
					Type:         schema.TypeString,
					Required:     true,
					DefaultFunc:  schema.EnvDefaultFunc("GENESYSCLOUD_REGION", nil),
					Description:  "AWS region where org exists. e.g. us-east-1. Can be set with the `GENESYSCLOUD_REGION` environment variable.",
					ValidateFunc: validation.StringInSlice(getAllowedRegions(), true),
				},
				"sdk_debug": {
					Type:        schema.TypeBool,
					Optional:    true,
					DefaultFunc: schema.EnvDefaultFunc("GENESYSCLOUD_SDK_DEBUG", false),
					Description: "Enables debug tracing in the Genesys Cloud SDK. Output will be written to the local file 'sdk_debug.log'.",
				},
				"token_pool_size": {
					Type:         schema.TypeInt,
					Optional:     true,
					DefaultFunc:  schema.EnvDefaultFunc("GENESYSCLOUD_TOKEN_POOL_SIZE", 10),
					Description:  "Max number of OAuth tokens in the token pool. Can be set with the `GENESYSCLOUD_TOKEN_POOL_SIZE` environment variable.",
					ValidateFunc: validation.IntBetween(1, 20),
				},
			},
			ResourcesMap: map[string]*schema.Resource{
				"genesyscloud_architect_datatable":             resourceArchitectDatatable(),
				"genesyscloud_architect_datatable_row":         resourceArchitectDatatableRow(),
				"genesyscloud_architect_schedules":				resourceArchitectSchedules(),
				"genesyscloud_auth_role":                       resourceAuthRole(),
				"genesyscloud_auth_division":                   resourceAuthDivision(),
				"genesyscloud_group":                           resourceGroup(),
				"genesyscloud_group_roles":                     resourceGroupRoles(),
				"genesyscloud_idp_adfs":                        resourceIdpAdfs(),
				"genesyscloud_idp_generic":                     resourceIdpGeneric(),
				"genesyscloud_idp_gsuite":                      resourceIdpGsuite(),
				"genesyscloud_idp_okta":                        resourceIdpOkta(),
				"genesyscloud_idp_onelogin":                    resourceIdpOnelogin(),
				"genesyscloud_idp_ping":                        resourceIdpPing(),
				"genesyscloud_idp_salesforce":                  resourceIdpSalesforce(),
				"genesyscloud_integration":                     resourceIntegration(),
				"genesyscloud_integration_action":              resourceIntegrationAction(),
				"genesyscloud_integration_credential":          resourceCredential(),
				"genesyscloud_location":                        resourceLocation(),
				"genesyscloud_oauth_client":                    resourceOAuthClient(),
				"genesyscloud_routing_email_domain":            resourceRoutingEmailDomain(),
				"genesyscloud_routing_email_route":             resourceRoutingEmailRoute(),
				"genesyscloud_routing_language":                resourceRoutingLanguage(),
				"genesyscloud_routing_queue":                   resourceRoutingQueue(),
				"genesyscloud_routing_skill":                   resourceRoutingSkill(),
				"genesyscloud_routing_utilization":             resourceRoutingUtilization(),
				"genesyscloud_routing_wrapupcode":              resourceRoutingWrapupCode(),
				"genesyscloud_telephony_providers_edges_phone": resourcePhone(),
				"genesyscloud_tf_export":                       resourceTfExport(),
				"genesyscloud_user":                            resourceUser(),
				"genesyscloud_user_roles":                      resourceUserRoles(),
			},
			DataSourcesMap: map[string]*schema.Resource{
				"genesyscloud_auth_role":            dataSourceAuthRole(),
				"genesyscloud_auth_division":        dataSourceAuthDivision(),
				"genesyscloud_flow":                 dataSourceFlow(),
				"genesyscloud_routing_language":     dataSourceRoutingLanguage(),
				"genesyscloud_routing_skill":        dataSourceRoutingSkill(),
				"genesyscloud_routing_email_domain": dataSourceRoutingEmailDomain(),
				"genesyscloud_script":               dataSourceScript(),
				"genesyscloud_user":                 dataSourceUser(),
			},
			ConfigureContextFunc: configure(version),
		}
	}
}

type providerMeta struct {
	Version      string
	ClientConfig *platformclientv2.Configuration
	Domain       string
}

func configure(version string) schema.ConfigureContextFunc {
	return func(context context.Context, data *schema.ResourceData) (interface{}, diag.Diagnostics) {
		// Initialize the SDK Client pool
		err := InitSDKClientPool(data.Get("token_pool_size").(int), version, data)
		if err != nil {
			return nil, err
		}
		return &providerMeta{
			Version:      version,
			ClientConfig: platformclientv2.GetDefaultConfiguration(),
			Domain:       getRegionDomain(data.Get("aws_region").(string)),
		}, nil
	}
}

func getRegionMap() map[string]string {
	return map[string]string{
		"dca":            "inindca.com",
		"tca":            "inintca.com",
		"us-east-1":      "mypurecloud.com",
		"us-west-2":      "usw2.pure.cloud",
		"eu-west-1":      "mypurecloud.ie",
		"eu-west-2":      "euw2.pure.cloud",
		"ap-southeast-2": "mypurecloud.com.au",
		"ap-northeast-1": "mypurecloud.jp",
		"eu-central-1":   "mypurecloud.de",
		"ca-central-1":   "cac1.pure.cloud",
		"ap-northeast-2": "apne2.pure.cloud",
		"ap-south-1":     "aps1.pure.cloud",
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

func getRegionBasePath(region string) string {
	return "https://api." + getRegionDomain(region)
}

func initClientConfig(data *schema.ResourceData, version string, config *platformclientv2.Configuration) diag.Diagnostics {
	oauthclientID := data.Get("oauthclient_id").(string)
	oauthclientSecret := data.Get("oauthclient_secret").(string)
	basePath := getRegionBasePath(data.Get("aws_region").(string))

	config.BasePath = basePath
	if data.Get("sdk_debug").(bool) {
		config.LoggingConfiguration = &platformclientv2.LoggingConfiguration{
			LogLevel:        platformclientv2.LDebug,
			LogRequestBody:  true,
			LogResponseBody: true,
		}
		config.LoggingConfiguration.SetLogFormat(platformclientv2.Text)
		config.LoggingConfiguration.SetLogFilePath("sdk_debug.log")
	}
	config.AddDefaultHeader("User-Agent", "GC Terraform Provider/"+version)
	config.RetryConfiguration = &platformclientv2.RetryConfiguration{
		RetryWaitMin: time.Second * 1,
		RetryWaitMax: time.Second * 30,
		RetryMax:     20,
		RequestLogHook: func(request *http.Request, count int) {
			if count > 0 && request != nil {
				log.Printf("Retry #%d for %s %s%s", count, request.Method, request.Host, request.RequestURI)
			}
		},
	}

	err := config.AuthorizeClientCredentials(oauthclientID, oauthclientSecret)
	if err != nil {
		return diag.Errorf("Failed to authorize Genesys Cloud client credentials: %v", err)
	}
	log.Printf("Initialized Go SDK Client. Debug=%t", data.Get("sdk_debug").(bool))
	return nil
}
