package genesyscloud

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/mypurecloud/platform-client-sdk-go/v92/platformclientv2"
)

var providerResources map[string]*schema.Resource
var providerDataSources map[string]*schema.Resource

var resourceMapMutex = sync.RWMutex{}
var datasourceMapMutex = sync.RWMutex{}

func init() {
	// Set descriptions to support markdown syntax, this will be used in document generation
	// and the language server.
	providerResources = make(map[string]*schema.Resource)
	providerDataSources = make(map[string]*schema.Resource)
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
	registerResources()
	registerDataSources()
}

// So I am basically inverting the registration process as we refactor.
// While resources are all in the same package I will continue to register them through the RegisterResource and
// RegisterDataSource.  However, to keep things cleans and avoid circular dependencies, as resources are moved into
// their own packages, I am going to have the individual resources register themselves.
func RegisterResource(resourceName string, resource *schema.Resource) {
	resourceMapMutex.Lock()
	providerResources[resourceName] = resource
	resourceMapMutex.Unlock()
}

func RegisterDataSource(dataSourceName string, datasource *schema.Resource) {
	datasourceMapMutex.Lock()
	providerDataSources[dataSourceName] = datasource
	datasourceMapMutex.Unlock()
}

func registerResources() {
	RegisterResource("genesyscloud_architect_datatable", resourceArchitectDatatable())
	RegisterResource("genesyscloud_architect_datatable_row", resourceArchitectDatatableRow())
	RegisterResource("genesyscloud_architect_emergencygroup", resourceArchitectEmergencyGroup())
	RegisterResource("genesyscloud_flow", resourceFlow())
	RegisterResource("genesyscloud_flow_milestone", resourceFlowMilestone())
	RegisterResource("genesyscloud_flow_outcome", resourceFlowOutcome())
	RegisterResource("genesyscloud_architect_ivr", resourceArchitectIvrConfig())
	RegisterResource("genesyscloud_architect_schedules", resourceArchitectSchedules())
	RegisterResource("genesyscloud_architect_schedulegroups", resourceArchitectScheduleGroups())
	RegisterResource("genesyscloud_architect_user_prompt", resourceArchitectUserPrompt())
	RegisterResource("genesyscloud_auth_role", resourceAuthRole())
	RegisterResource("genesyscloud_auth_division", resourceAuthDivision())
	RegisterResource("genesyscloud_employeeperformance_externalmetrics_definitions", resourceEmployeeperformanceExternalmetricsDefinition())
	RegisterResource("genesyscloud_group", resourceGroup())
	RegisterResource("genesyscloud_group_roles", resourceGroupRoles())
	RegisterResource("genesyscloud_idp_adfs", resourceIdpAdfs())
	RegisterResource("genesyscloud_idp_generic", resourceIdpGeneric())
	RegisterResource("genesyscloud_idp_gsuite", resourceIdpGsuite())
	RegisterResource("genesyscloud_idp_okta", resourceIdpOkta())
	RegisterResource("genesyscloud_idp_onelogin", resourceIdpOnelogin())
	RegisterResource("genesyscloud_idp_ping", resourceIdpPing())
	RegisterResource("genesyscloud_idp_salesforce", resourceIdpSalesforce())
	RegisterResource("genesyscloud_integration", resourceIntegration())
	RegisterResource("genesyscloud_integration_action", resourceIntegrationAction())
	RegisterResource("genesyscloud_integration_credential", resourceCredential())
	RegisterResource("genesyscloud_journey_action_map", resourceJourneyActionMap())
	RegisterResource("genesyscloud_journey_action_template", resourceJourneyActionTemplate())
	RegisterResource("genesyscloud_journey_outcome", resourceJourneyOutcome())
	RegisterResource("genesyscloud_journey_segment", resourceJourneySegment())
	RegisterResource("genesyscloud_knowledge_knowledgebase", resourceKnowledgeKnowledgebase())
	RegisterResource("genesyscloud_knowledge_document", resourceKnowledgeDocument())
	RegisterResource("genesyscloud_knowledge_category", resourceKnowledgeCategory())
	RegisterResource("genesyscloud_location", resourceLocation())
	RegisterResource("genesyscloud_recording_media_retention_policy", resourceMediaRetentionPolicy())
	RegisterResource("genesyscloud_oauth_client", resourceOAuthClient())
	RegisterResource("genesyscloud_outbound_campaignrule", resourceOutboundCampaignRule())
	RegisterResource("genesyscloud_outbound_attempt_limit", resourceOutboundAttemptLimit())
	RegisterResource("genesyscloud_outbound_callanalysisresponseset", resourceOutboundCallAnalysisResponseSet())
	RegisterResource("genesyscloud_outbound_campaign", resourceOutboundCampaign())
	RegisterResource("genesyscloud_outbound_contactlistfilter", resourceOutboundContactListFilter())
	RegisterResource("genesyscloud_outbound_callabletimeset", resourceOutboundCallabletimeset())
	RegisterResource("genesyscloud_outbound_contact_list", resourceOutboundContactList())
	RegisterResource("genesyscloud_outbound_ruleset", resourceOutboundRuleset())
	RegisterResource("genesyscloud_outbound_messagingcampaign", resourceOutboundMessagingCampaign())
	RegisterResource("genesyscloud_outbound_sequence", resourceOutboundSequence())
	RegisterResource("genesyscloud_outbound_settings", resourceOutboundSettings())
	RegisterResource("genesyscloud_outbound_wrapupcodemappings", resourceOutboundWrapUpCodeMappings())
	RegisterResource("genesyscloud_outbound_dnclist", resourceOutboundDncList())
	RegisterResource("genesyscloud_orgauthorization_pairing", resourceOrgauthorizationPairing())
	RegisterResource("genesyscloud_processautomation_trigger", resourceProcessAutomationTrigger())
	RegisterResource("genesyscloud_quality_forms_evaluation", resourceEvaluationForm())
	RegisterResource("genesyscloud_quality_forms_survey", resourceSurveyForm())
	RegisterResource("genesyscloud_responsemanagement_library", resourceResponsemanagementLibrary())
	RegisterResource("genesyscloud_responsemanagement_response", resourceResponsemanagementResponse())
	RegisterResource("genesyscloud_responsemanagement_responseasset", resourceResponseManagamentResponseAsset())
	RegisterResource("genesyscloud_routing_email_domain", resourceRoutingEmailDomain())
	RegisterResource("genesyscloud_routing_email_route", resourceRoutingEmailRoute())
	RegisterResource("genesyscloud_routing_language", resourceRoutingLanguage())
	RegisterResource("genesyscloud_routing_queue", resourceRoutingQueue())
	RegisterResource("genesyscloud_routing_skill", resourceRoutingSkill())
	RegisterResource("genesyscloud_routing_skill_group", resourceRoutingSkillGroup())
	RegisterResource("genesyscloud_routing_settings", resourceRoutingSettings())
	RegisterResource("genesyscloud_routing_utilization", resourceRoutingUtilization())
	RegisterResource("genesyscloud_routing_wrapupcode", resourceRoutingWrapupCode())
	RegisterResource("genesyscloud_telephony_providers_edges_did_pool", resourceTelephonyDidPool())
	RegisterResource("genesyscloud_telephony_providers_edges_edge_group", resourceEdgeGroup())
	RegisterResource("genesyscloud_telephony_providers_edges_extension_pool", resourceTelephonyExtensionPool())
	RegisterResource("genesyscloud_telephony_providers_edges_phone", resourcePhone())
	RegisterResource("genesyscloud_telephony_providers_edges_site", resourceSite())
	RegisterResource("genesyscloud_telephony_providers_edges_phonebasesettings", resourcePhoneBaseSettings())
	RegisterResource("genesyscloud_telephony_providers_edges_trunkbasesettings", resourceTrunkBaseSettings())
	RegisterResource("genesyscloud_telephony_providers_edges_trunk", resourceTrunk())
	RegisterResource("genesyscloud_user", resourceUser())
	RegisterResource("genesyscloud_user_roles", resourceUserRoles())
	RegisterResource("genesyscloud_webdeployments_configuration", resourceWebDeploymentConfiguration())
	RegisterResource("genesyscloud_webdeployments_deployment", resourceWebDeployment())
	RegisterResource("genesyscloud_widget_deployment", resourceWidgetDeployment())
}

func registerDataSources() {
	RegisterDataSource("genesyscloud_architect_datatable", dataSourceArchitectDatatable())
	RegisterDataSource("genesyscloud_architect_ivr", dataSourceArchitectIvr())
	RegisterDataSource("genesyscloud_architect_emergencygroup", dataSourceArchitectEmergencyGroup())
	RegisterDataSource("genesyscloud_architect_schedules", dataSourceSchedule())
	RegisterDataSource("genesyscloud_architect_schedulegroups", dataSourceArchitectScheduleGroups())
	RegisterDataSource("genesyscloud_architect_user_prompt", dataSourceUserPrompt())
	RegisterDataSource("genesyscloud_auth_role", dataSourceAuthRole())
	RegisterDataSource("genesyscloud_auth_division", dataSourceAuthDivision())
	RegisterDataSource("genesyscloud_auth_division_home", dataSourceAuthDivisionHome())
	RegisterDataSource("genesyscloud_employeeperformance_externalmetrics_definitions", dataSourceEmployeeperformanceExternalmetricsDefinition())
	RegisterDataSource("genesyscloud_flow", dataSourceFlow())
	RegisterDataSource("genesyscloud_flow_milestone", dataSourceFlowMilestone())
	RegisterDataSource("genesyscloud_flow_outcome", dataSourceFlowOutcome())
	RegisterDataSource("genesyscloud_group", dataSourceGroup())
	RegisterDataSource("genesyscloud_integration", dataSourceIntegration())
	RegisterDataSource("genesyscloud_integration_action", dataSourceIntegrationAction())
	RegisterDataSource("genesyscloud_integration_credential", dataSourceIntegrationCredential())
	RegisterDataSource("genesyscloud_journey_action_map", dataSourceJourneyActionMap())
	RegisterDataSource("genesyscloud_journey_action_template", dataSourceJourneyActionTemplate())
	RegisterDataSource("genesyscloud_journey_outcome", dataSourceJourneyOutcome())
	RegisterDataSource("genesyscloud_journey_segment", dataSourceJourneySegment())
	RegisterDataSource("genesyscloud_knowledge_knowledgebase", dataSourceKnowledgeKnowledgebase())
	RegisterDataSource("genesyscloud_location", dataSourceLocation())
	RegisterDataSource("genesyscloud_oauth_client", dataSourceOAuthClient())
	RegisterDataSource("genesyscloud_processautomation_trigger", dataSourceProcessAutomationTrigger())
	RegisterDataSource("genesyscloud_organizations_me", dataSourceOrganizationsMe())
	RegisterDataSource("genesyscloud_outbound_attempt_limit", dataSourceOutboundAttemptLimit())
	RegisterDataSource("genesyscloud_outbound_callanalysisresponseset", dataSourceOutboundCallAnalysisResponseSet())
	RegisterDataSource("genesyscloud_outbound_campaign", dataSourceOutboundCampaign())
	RegisterDataSource("genesyscloud_outbound_campaignrule", dataSourceOutboundCampaignRule())
	RegisterDataSource("genesyscloud_outbound_callabletimeset", dataSourceOutboundCallabletimeset())
	RegisterDataSource("genesyscloud_outbound_contact_list", dataSourceOutboundContactList())
	RegisterDataSource("genesyscloud_outbound_messagingcampaign", dataSourceOutboundMessagingcampaign())
	RegisterDataSource("genesyscloud_outbound_contactlistfilter", dataSourceOutboundContactListFilter())
	RegisterDataSource("genesyscloud_outbound_ruleset", dataSourceOutboundRuleset())
	RegisterDataSource("genesyscloud_outbound_sequence", dataSourceOutboundSequence())
	RegisterDataSource("genesyscloud_outbound_dnclist", dataSourceOutboundDncList())
	RegisterDataSource("genesyscloud_quality_forms_evaluation", dataSourceQualityFormsEvaluations())
	RegisterDataSource("genesyscloud_quality_forms_survey", dataSourceQualityFormsSurvey())
	RegisterDataSource("genesyscloud_recording_media_retention_policy", dataSourceRecordingMediaRetentionPolicy())
	RegisterDataSource("genesyscloud_responsemanagement_library", dataSourceResponsemanagementLibrary())
	RegisterDataSource("genesyscloud_responsemanagement_response", dataSourceResponsemanagementResponse())
	RegisterDataSource("genesyscloud_responsemanagement_responseasset", dataSourceResponseManagamentResponseAsset())
	RegisterDataSource("genesyscloud_routing_language", dataSourceRoutingLanguage())
	RegisterDataSource("genesyscloud_routing_queue", dataSourceRoutingQueue())
	RegisterDataSource("genesyscloud_routing_settings", dataSourceRoutingSettings())
	RegisterDataSource("genesyscloud_routing_skill", dataSourceRoutingSkill())
	RegisterDataSource("genesyscloud_routing_skill_group", dataSourceRoutingSkillGroup())
	RegisterDataSource("genesyscloud_routing_email_domain", dataSourceRoutingEmailDomain())
	RegisterDataSource("genesyscloud_routing_wrapupcode", dataSourceRoutingWrapupcode())
	RegisterDataSource("genesyscloud_script", dataSourceScript())
	RegisterDataSource("genesyscloud_station", dataSourceStation())
	RegisterDataSource("genesyscloud_user", dataSourceUser())
	RegisterDataSource("genesyscloud_telephony_providers_edges_did", dataSourceDid())
	RegisterDataSource("genesyscloud_telephony_providers_edges_did_pool", dataSourceDidPool())
	RegisterDataSource("genesyscloud_telephony_providers_edges_edge_group", dataSourceEdgeGroup())
	RegisterDataSource("genesyscloud_telephony_providers_edges_extension_pool", dataSourceExtensionPool())
	RegisterDataSource("genesyscloud_telephony_providers_edges_site", dataSourceSite())
	RegisterDataSource("genesyscloud_telephony_providers_edges_linebasesettings", dataSourceLineBaseSettings())
	RegisterDataSource("genesyscloud_telephony_providers_edges_phone", dataSourcePhone())
	RegisterDataSource("genesyscloud_telephony_providers_edges_phonebasesettings", dataSourcePhoneBaseSettings())
	RegisterDataSource("genesyscloud_telephony_providers_edges_trunk", dataSourceTrunk())
	RegisterDataSource("genesyscloud_telephony_providers_edges_trunkbasesettings", dataSourceTrunkBaseSettings())
	RegisterDataSource("genesyscloud_webdeployments_configuration", dataSourceWebDeploymentsConfiguration())
	RegisterDataSource("genesyscloud_webdeployments_deployment", dataSourceWebDeploymentsDeployment())
	RegisterDataSource("genesyscloud_widget_deployment", dataSourceWidgetDeployments())
}

// New initializes the provider schema
func New(version string) func() *schema.Provider {
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

func getRegionBasePath(region string) string {
	return "https://api." + getRegionDomain(region)
}

func initClientConfig(data *schema.ResourceData, version string, config *platformclientv2.Configuration) diag.Diagnostics {
	accessToken := data.Get("access_token").(string)
	oauthclientID := data.Get("oauthclient_id").(string)
	oauthclientSecret := data.Get("oauthclient_secret").(string)
	basePath := getRegionBasePath(data.Get("aws_region").(string))

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
