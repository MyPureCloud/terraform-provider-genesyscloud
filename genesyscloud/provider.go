package genesyscloud

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	//"sync"
	"time"
	"os"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/mypurecloud/platform-client-sdk-go/v105/platformclientv2"
)

// var providerResources map[string]*schema.Resource
// var providerDataSources map[string]*schema.Resource

// var resourceMapMutex = sync.RWMutex{}
// var datasourceMapMutex = sync.RWMutex{}

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

	//Registering the schema resources.  We do this in the init to make sure everything is setup before anyone uses a provider.
	// registerResources()
	// registerDataSources()
}

// So I am basically inverting the registration process as we refactor.
// While resources are all in the same package I will continue to register them through the RegisterResource and
// RegisterDataSource.  However, to keep things cleans and avoid circular dependencies, as resources are moved into
// their own packages, I am going to have the individual resources register themselves.
// func RegisterResource(resourceName string, resource *schema.Resource) {
// 	resourceMapMutex.Lock()
// 	providerResources[resourceName] = resource
// 	resourceMapMutex.Unlock()
// }

// func RegisterDataSource(dataSourceName string, datasource *schema.Resource) {
// 	datasourceMapMutex.Lock()
// 	providerDataSources[dataSourceName] = datasource
// 	datasourceMapMutex.Unlock()
// }

// func registerResources() {
// 	log.Printf("resource registration started")
// 	RegisterResource("genesyscloud_architect_datatable", ResourceArchitectDatatable())
// 	RegisterResource("genesyscloud_architect_datatable_row", ResourceArchitectDatatableRow())
// 	RegisterResource("genesyscloud_architect_emergencygroup", ResourceArchitectEmergencyGroup())
// 	RegisterResource("genesyscloud_flow", ResourceFlow())
// 	RegisterResource("genesyscloud_flow_milestone", ResourceFlowMilestone())
// 	RegisterResource("genesyscloud_flow_outcome", ResourceFlowOutcome())
// 	RegisterResource("genesyscloud_architect_ivr", ResourceArchitectIvrConfig())
// 	RegisterResource("genesyscloud_architect_schedules", ResourceArchitectSchedules())
// 	RegisterResource("genesyscloud_architect_schedulegroups", ResourceArchitectScheduleGroups())
// 	RegisterResource("genesyscloud_architect_user_prompt", ResourceArchitectUserPrompt())
// 	RegisterResource("genesyscloud_auth_role", ResourceAuthRole())
// 	RegisterResource("genesyscloud_auth_division", ResourceAuthDivision())
// 	RegisterResource("genesyscloud_employeeperformance_externalmetrics_definitions", ResourceEmployeeperformanceExternalmetricsDefinition())
// 	RegisterResource("genesyscloud_externalcontacts_contact", ResourceExternalContact())
// 	RegisterResource("genesyscloud_group", ResourceGroup())
// 	RegisterResource("genesyscloud_group_roles", ResourceGroupRoles())
// 	RegisterResource("genesyscloud_idp_adfs", ResourceIdpAdfs())
// 	RegisterResource("genesyscloud_idp_generic", ResourceIdpGeneric())
// 	RegisterResource("genesyscloud_idp_gsuite", ResourceIdpGsuite())
// 	RegisterResource("genesyscloud_idp_okta", ResourceIdpOkta())
// 	RegisterResource("genesyscloud_idp_onelogin", ResourceIdpOnelogin())
// 	RegisterResource("genesyscloud_idp_ping", ResourceIdpPing())
// 	RegisterResource("genesyscloud_idp_salesforce", ResourceIdpSalesforce())
// 	RegisterResource("genesyscloud_integration", ResourceIntegration())
// 	RegisterResource("genesyscloud_integration_action", ResourceIntegrationAction())
// 	RegisterResource("genesyscloud_integration_credential", ResourceCredential())
// 	RegisterResource("genesyscloud_journey_action_map", ResourceJourneyActionMap())
// 	RegisterResource("genesyscloud_journey_action_template", ResourceJourneyActionTemplate())
// 	RegisterResource("genesyscloud_journey_outcome", ResourceJourneyOutcome())
// 	RegisterResource("genesyscloud_journey_segment", ResourceJourneySegment())
// 	RegisterResource("genesyscloud_knowledge_knowledgebase", ResourceKnowledgeKnowledgebase())
// 	RegisterResource("genesyscloud_knowledge_document", ResourceKnowledgeDocument())
// 	RegisterResource("genesyscloud_knowledge_v1_document", ResourceKnowledgeDocumentV1())
// 	RegisterResource("genesyscloud_knowledge_document_variation", ResourceKnowledgeDocumentVariation())
// 	RegisterResource("genesyscloud_knowledge_category", ResourceKnowledgeCategory())
// 	RegisterResource("genesyscloud_knowledge_v1_category", ResourceKnowledgeCategoryV1())
// 	RegisterResource("genesyscloud_knowledge_label", ResourceKnowledgeLabel())
// 	RegisterResource("genesyscloud_location", ResourceLocation())
// 	RegisterResource("genesyscloud_recording_media_retention_policy", ResourceMediaRetentionPolicy())
// 	RegisterResource("genesyscloud_oauth_client", ResourceOAuthClient())
// 	RegisterResource("genesyscloud_outbound_campaignrule", resourceOutboundCampaignRule())
// 	RegisterResource("genesyscloud_outbound_attempt_limit", ResourceOutboundAttemptLimit())
// 	RegisterResource("genesyscloud_outbound_callanalysisresponseset", resourceOutboundCallAnalysisResponseSet())
// 	RegisterResource("genesyscloud_outbound_campaign", resourceOutboundCampaign())
// 	RegisterResource("genesyscloud_outbound_contactlistfilter", resourceOutboundContactListFilter())
// 	RegisterResource("genesyscloud_outbound_callabletimeset", resourceOutboundCallabletimeset())
// 	RegisterResource("genesyscloud_outbound_contact_list", ResourceOutboundContactList())
// 	RegisterResource("genesyscloud_outbound_messagingcampaign", resourceOutboundMessagingCampaign())
// 	RegisterResource("genesyscloud_outbound_sequence", resourceOutboundSequence())
// 	RegisterResource("genesyscloud_outbound_settings", ResourceOutboundSettings())
// 	RegisterResource("genesyscloud_outbound_wrapupcodemappings", resourceOutboundWrapUpCodeMappings())
// 	RegisterResource("genesyscloud_outbound_dnclist", resourceOutboundDncList())
// 	RegisterResource("genesyscloud_orgauthorization_pairing", resourceOrgauthorizationPairing())
// 	RegisterResource("genesyscloud_quality_forms_evaluation", ResourceEvaluationForm())
// 	RegisterResource("genesyscloud_quality_forms_survey", resourceSurveyForm())
// 	RegisterResource("genesyscloud_responsemanagement_library", ResourceResponsemanagementLibrary())
// 	RegisterResource("genesyscloud_responsemanagement_response", resourceResponsemanagementResponse())
// 	RegisterResource("genesyscloud_responsemanagement_responseasset", resourceResponseManagamentResponseAsset())
// 	RegisterResource("genesyscloud_routing_email_domain", ResourceRoutingEmailDomain())
// 	RegisterResource("genesyscloud_routing_email_route", ResourceRoutingEmailRoute())
// 	RegisterResource("genesyscloud_routing_language", ResourceRoutingLanguage())
// 	RegisterResource("genesyscloud_routing_queue", ResourceRoutingQueue())
// 	RegisterResource("genesyscloud_routing_skill", ResourceRoutingSkill())
// 	RegisterResource("genesyscloud_routing_skill_group", resourceRoutingSkillGroup())
// 	RegisterResource("genesyscloud_routing_sms_address", resourceRoutingSmsAddress())
// 	RegisterResource("genesyscloud_routing_settings", ResourceRoutingSettings())
// 	RegisterResource("genesyscloud_routing_utilization", ResourceRoutingUtilization())
// 	RegisterResource("genesyscloud_routing_wrapupcode", ResourceRoutingWrapupCode())
// 	RegisterResource("genesyscloud_script", resourceScript())
// 	RegisterResource("genesyscloud_telephony_providers_edges_did_pool", ResourceTelephonyDidPool())
// 	RegisterResource("genesyscloud_telephony_providers_edges_edge_group", ResourceEdgeGroup())
// 	RegisterResource("genesyscloud_telephony_providers_edges_extension_pool", ResourceTelephonyExtensionPool())
// 	RegisterResource("genesyscloud_telephony_providers_edges_phone", ResourcePhone())
// 	RegisterResource("genesyscloud_telephony_providers_edges_site", ResourceSite())
// 	RegisterResource("genesyscloud_telephony_providers_edges_phonebasesettings", ResourcePhoneBaseSettings())
// 	RegisterResource("genesyscloud_telephony_providers_edges_trunkbasesettings", ResourceTrunkBaseSettings())
// 	RegisterResource("genesyscloud_telephony_providers_edges_trunk", ResourceTrunk())
// 	RegisterResource("genesyscloud_user", ResourceUser())
// 	RegisterResource("genesyscloud_user_roles", ResourceUserRoles())
// 	RegisterResource("genesyscloud_webdeployments_configuration", ResourceWebDeploymentConfiguration())
// 	RegisterResource("genesyscloud_webdeployments_deployment", ResourceWebDeployment())
// 	RegisterResource("genesyscloud_widget_deployment", ResourceWidgetDeployment())
// }

// func registerDataSources() {
// 	RegisterDataSource("genesyscloud_architect_datatable", DataSourceArchitectDatatable())
// 	RegisterDataSource("genesyscloud_architect_ivr", dataSourceArchitectIvr())
// 	RegisterDataSource("genesyscloud_architect_emergencygroup", dataSourceArchitectEmergencyGroup())
// 	RegisterDataSource("genesyscloud_architect_schedules", dataSourceSchedule())
// 	RegisterDataSource("genesyscloud_architect_schedulegroups", dataSourceArchitectScheduleGroups())
// 	RegisterDataSource("genesyscloud_architect_user_prompt", dataSourceUserPrompt())
// 	RegisterDataSource("genesyscloud_auth_role", dataSourceAuthRole())
// 	RegisterDataSource("genesyscloud_auth_division", dataSourceAuthDivision())
// 	RegisterDataSource("genesyscloud_auth_division_home", DataSourceAuthDivisionHome())
// 	RegisterDataSource("genesyscloud_employeeperformance_externalmetrics_definitions", dataSourceEmployeeperformanceExternalmetricsDefinition())
// 	RegisterDataSource("genesyscloud_externalcontacts_contact", dataSourceExternalContactsContact())
// 	RegisterDataSource("genesyscloud_flow", DataSourceFlow())
// 	RegisterDataSource("genesyscloud_flow_milestone", dataSourceFlowMilestone())
// 	RegisterDataSource("genesyscloud_flow_outcome", dataSourceFlowOutcome())
// 	RegisterDataSource("genesyscloud_group", dataSourceGroup())
// 	RegisterDataSource("genesyscloud_integration", dataSourceIntegration())
// 	RegisterDataSource("genesyscloud_integration_action", dataSourceIntegrationAction())
// 	RegisterDataSource("genesyscloud_integration_credential", dataSourceIntegrationCredential())
// 	RegisterDataSource("genesyscloud_journey_action_map", dataSourceJourneyActionMap())
// 	RegisterDataSource("genesyscloud_journey_action_template", dataSourceJourneyActionTemplate())
// 	RegisterDataSource("genesyscloud_journey_outcome", dataSourceJourneyOutcome())
// 	RegisterDataSource("genesyscloud_journey_segment", dataSourceJourneySegment())
// 	RegisterDataSource("genesyscloud_knowledge_knowledgebase", dataSourceKnowledgeKnowledgebase())
// 	RegisterDataSource("genesyscloud_knowledge_category", dataSourceKnowledgeCategory())
// 	RegisterDataSource("genesyscloud_knowledge_label", dataSourceKnowledgeLabel())
// 	RegisterDataSource("genesyscloud_location", DataSourceLocation())
// 	RegisterDataSource("genesyscloud_oauth_client", dataSourceOAuthClient())
// 	RegisterDataSource("genesyscloud_organizations_me", dataSourceOrganizationsMe())
// 	RegisterDataSource("genesyscloud_outbound_attempt_limit", DataSourceOutboundAttemptLimit())
// 	RegisterDataSource("genesyscloud_outbound_callanalysisresponseset", dataSourceOutboundCallAnalysisResponseSet())
// 	RegisterDataSource("genesyscloud_outbound_campaign", dataSourceOutboundCampaign())
// 	RegisterDataSource("genesyscloud_outbound_campaignrule", dataSourceOutboundCampaignRule())
// 	RegisterDataSource("genesyscloud_outbound_callabletimeset", dataSourceOutboundCallabletimeset())
// 	RegisterDataSource("genesyscloud_outbound_contact_list", DataSourceOutboundContactList())
// 	RegisterDataSource("genesyscloud_outbound_messagingcampaign", dataSourceOutboundMessagingcampaign())
// 	RegisterDataSource("genesyscloud_outbound_contactlistfilter", dataSourceOutboundContactListFilter())
// 	RegisterDataSource("genesyscloud_outbound_sequence", dataSourceOutboundSequence())
// 	RegisterDataSource("genesyscloud_outbound_dnclist", dataSourceOutboundDncList())
// 	RegisterDataSource("genesyscloud_quality_forms_evaluation", dataSourceQualityFormsEvaluations())
// 	RegisterDataSource("genesyscloud_quality_forms_survey", dataSourceQualityFormsSurvey())
// 	RegisterDataSource("genesyscloud_recording_media_retention_policy", dataSourceRecordingMediaRetentionPolicy())
// 	RegisterDataSource("genesyscloud_responsemanagement_library", dataSourceResponsemanagementLibrary())
// 	RegisterDataSource("genesyscloud_responsemanagement_response", dataSourceResponsemanagementResponse())
// 	RegisterDataSource("genesyscloud_responsemanagement_responseasset", dataSourceResponseManagamentResponseAsset())
// 	RegisterDataSource("genesyscloud_routing_language", dataSourceRoutingLanguage())
// 	RegisterDataSource("genesyscloud_routing_queue", DataSourceRoutingQueue())
// 	RegisterDataSource("genesyscloud_routing_settings", dataSourceRoutingSettings())
// 	RegisterDataSource("genesyscloud_routing_skill", dataSourceRoutingSkill())
// 	RegisterDataSource("genesyscloud_routing_skill_group", dataSourceRoutingSkillGroup())
// 	RegisterDataSource("genesyscloud_routing_sms_address", dataSourceRoutingSmsAddress())
// 	RegisterDataSource("genesyscloud_routing_email_domain", dataSourceRoutingEmailDomain())
// 	RegisterDataSource("genesyscloud_routing_wrapupcode", DataSourceRoutingWrapupcode())
//  RegisterResource("genesyscloud_routing_wrapupcode", ResourceRoutingWrapupCode())
// 	RegisterDataSource("genesyscloud_script", dataSourceScript())
// 	RegisterDataSource("genesyscloud_station", dataSourceStation())
// 	RegisterDataSource("genesyscloud_user", dataSourceUser())
// 	RegisterDataSource("genesyscloud_telephony_providers_edges_did", dataSourceDid())
// 	RegisterDataSource("genesyscloud_telephony_providers_edges_did_pool", dataSourceDidPool())
// 	RegisterDataSource("genesyscloud_telephony_providers_edges_edge_group", dataSourceEdgeGroup())
// 	RegisterDataSource("genesyscloud_telephony_providers_edges_extension_pool", dataSourceExtensionPool())
// 	RegisterDataSource("genesyscloud_telephony_providers_edges_site", DataSourceSite())
// 	RegisterDataSource("genesyscloud_telephony_providers_edges_linebasesettings", dataSourceLineBaseSettings())
// 	RegisterDataSource("genesyscloud_telephony_providers_edges_phone", dataSourcePhone())
// 	RegisterDataSource("genesyscloud_telephony_providers_edges_phonebasesettings", dataSourcePhoneBaseSettings())
// 	RegisterDataSource("genesyscloud_telephony_providers_edges_trunk", dataSourceTrunk())
// 	RegisterDataSource("genesyscloud_telephony_providers_edges_trunkbasesettings", dataSourceTrunkBaseSettings())
// 	RegisterDataSource("genesyscloud_webdeployments_configuration", dataSourceWebDeploymentsConfiguration())
// 	RegisterDataSource("genesyscloud_webdeployments_deployment", dataSourceWebDeploymentsDeployment())
// 	RegisterDataSource("genesyscloud_widget_deployment", dataSourceWidgetDeployments())
// 	log.Printf("resource registration ended")
// }

// New initializes the provider schema
func New(version string,providerResources map[string]*schema.Resource, providerDataSources map[string]*schema.Resource) func() *schema.Provider {
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
}

func configure(version string) schema.ConfigureContextFunc {
	return func(context context.Context, data *schema.ResourceData) (interface{}, diag.Diagnostics) {
		// Initialize a single client if we have an access token
		accessToken := data.Get("access_token").(string)
		if accessToken != "" {
			once.Do(func() {
				sdkConfig := platformclientv2.GetDefaultConfiguration()
				_ = initClientConfig(data, version, sdkConfig)

				sdkClientPool = &SDKClientPool{
					pool: make(chan *platformclientv2.Configuration, 1),
				}
				sdkClientPool.pool <- sdkConfig
			})
		} else {
			// Initialize the SDK Client pool
			err := InitSDKClientPool(data.Get("token_pool_size").(int), version, data)
			if err != nil {
				return nil, err
			}
		}
		return &ProviderMeta{
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

func initClientConfig(data *schema.ResourceData, version string, config *platformclientv2.Configuration) diag.Diagnostics {
	accessToken := data.Get("access_token").(string)
	oauthclientID := data.Get("oauthclient_id").(string)
	oauthclientSecret := data.Get("oauthclient_secret").(string)
	basePath := GetRegionBasePath(data.Get("aws_region").(string))

	config.BasePath = basePath
	if data.Get("sdk_debug").(bool) {
		config.LoggingConfiguration = &platformclientv2.LoggingConfiguration{
			LogLevel:        platformclientv2.LTrace,
			LogRequestBody:  true,
			LogResponseBody: true,
		}
		config.LoggingConfiguration.SetLogToConsole(false)
		config.LoggingConfiguration.SetLogFormat(platformclientv2.Text)
		config.LoggingConfiguration.SetLogFilePath("sdk_debug.log")
	}

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

	config.AddDefaultHeader("User-Agent", "GC Terraform Provider/"+version)
	config.RetryConfiguration = &platformclientv2.RetryConfiguration{
		RetryWaitMin: time.Second * 1,
		RetryWaitMax: time.Second * 30,
		RetryMax:     20,
		RequestLogHook: func(request *http.Request, count int) {
			if count > 0 && request != nil {
				log.Printf("Retry #%d for %s %s", count, request.Method, request.URL)
			}
		},
		ResponseLogHook: func(response *http.Response) {
			if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
				log.Printf("Response %s", response.Status)
			}
		},
	}

	if accessToken != "" {
		log.Print("Setting access token set on configuration instance.")
		config.AccessToken = accessToken
	} else {
		err := config.AuthorizeClientCredentials(oauthclientID, oauthclientSecret)
		if err != nil {
			return diag.Errorf("Failed to authorize Genesys Cloud client credentials: %v", err)
		}
	}

	log.Printf("Initialized Go SDK Client. Debug=%t", data.Get("sdk_debug").(bool))
	return nil
}

func AuthorizeSdk() error {
	// Create new config
	sdkConfig := platformclientv2.GetDefaultConfiguration()

	sdkConfig.BasePath = GetRegionBasePath(os.Getenv("GENESYSCLOUD_REGION"))

	err := sdkConfig.AuthorizeClientCredentials(os.Getenv("GENESYSCLOUD_OAUTHCLIENT_ID"), os.Getenv("GENESYSCLOUD_OAUTHCLIENT_SECRET"))
	if err != nil {
		return err
	}

	return nil
}
