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
	"github.com/mypurecloud/platform-client-sdk-go/v91/platformclientv2"
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
			ResourcesMap: map[string]*schema.Resource{
				"genesyscloud_architect_datatable":                             resourceArchitectDatatable(),
				"genesyscloud_architect_datatable_row":                         resourceArchitectDatatableRow(),
				"genesyscloud_architect_emergencygroup":                        resourceArchitectEmergencyGroup(),
				"genesyscloud_flow":                                            resourceFlow(),
				"genesyscloud_flow_milestone":                                  resourceFlowMilestone(),
				"genesyscloud_flow_outcome":                                    resourceFlowOutcome(),
				"genesyscloud_architect_ivr":                                   resourceArchitectIvrConfig(),
				"genesyscloud_architect_schedules":                             resourceArchitectSchedules(),
				"genesyscloud_architect_schedulegroups":                        resourceArchitectScheduleGroups(),
				"genesyscloud_architect_user_prompt":                           resourceArchitectUserPrompt(),
				"genesyscloud_auth_role":                                       resourceAuthRole(),
				"genesyscloud_auth_division":                                   resourceAuthDivision(),
				"genesyscloud_employeeperformance_externalmetrics_definitions": resourceEmployeeperformanceExternalmetricsDefinition(),
				"genesyscloud_group":                                           resourceGroup(),
				"genesyscloud_group_roles":                                     resourceGroupRoles(),
				"genesyscloud_idp_adfs":                                        resourceIdpAdfs(),
				"genesyscloud_idp_generic":                                     resourceIdpGeneric(),
				"genesyscloud_idp_gsuite":                                      resourceIdpGsuite(),
				"genesyscloud_idp_okta":                                        resourceIdpOkta(),
				"genesyscloud_idp_onelogin":                                    resourceIdpOnelogin(),
				"genesyscloud_idp_ping":                                        resourceIdpPing(),
				"genesyscloud_idp_salesforce":                                  resourceIdpSalesforce(),
				"genesyscloud_integration":                                     resourceIntegration(),
				"genesyscloud_integration_action":                              resourceIntegrationAction(),
				"genesyscloud_integration_credential":                          resourceCredential(),
				"genesyscloud_journey_action_map":                              resourceJourneyActionMap(),
				"genesyscloud_journey_outcome":                                 resourceJourneyOutcome(),
				"genesyscloud_journey_segment":                                 resourceJourneySegment(),
				"genesyscloud_knowledge_knowledgebase":                         resourceKnowledgeKnowledgebase(),
				"genesyscloud_knowledge_document":                              resourceKnowledgeDocument(),
				"genesyscloud_knowledge_category":                              resourceKnowledgeCategory(),
				"genesyscloud_location":                                        resourceLocation(),
				"genesyscloud_recording_media_retention_policy":                resourceMediaRetentionPolicy(),
				"genesyscloud_oauth_client":                                    resourceOAuthClient(),
				"genesyscloud_outbound_campaignrule":                           resourceOutboundCampaignRule(),
				"genesyscloud_outbound_attempt_limit":                          resourceOutboundAttemptLimit(),
				"genesyscloud_outbound_callanalysisresponseset":                resourceOutboundCallAnalysisResponseSet(),
				"genesyscloud_outbound_campaign":                               resourceOutboundCampaign(),
				"genesyscloud_outbound_contactlistfilter":                      resourceOutboundContactListFilter(),
				"genesyscloud_outbound_callabletimeset":                        resourceOutboundCallabletimeset(),
				"genesyscloud_outbound_contact_list":                           resourceOutboundContactList(),
				"genesyscloud_outbound_ruleset":                                resourceOutboundRuleset(),
				"genesyscloud_outbound_messagingcampaign":                      resourceOutboundMessagingCampaign(),
				"genesyscloud_outbound_sequence":                               resourceOutboundSequence(),
				"genesyscloud_outbound_settings":                               resourceOutboundSettings(),
				"genesyscloud_outbound_wrapupcodemappings":                     resourceOutboundWrapUpCodeMappings(),
				"genesyscloud_outbound_dnclist":                                resourceOutboundDncList(),
				"genesyscloud_orgauthorization_pairing":                        resourceOrgauthorizationPairing(),
				"genesyscloud_processautomation_trigger":                       resourceProcessAutomationTrigger(),
				"genesyscloud_quality_forms_evaluation":                        resourceEvaluationForm(),
				"genesyscloud_quality_forms_survey":                            resourceSurveyForm(),
				"genesyscloud_responsemanagement_library":                      resourceResponsemanagementLibrary(),
				"genesyscloud_responsemanagement_responseasset":                resourceResponseManagamentResponseAsset(),
				"genesyscloud_routing_email_domain":                            resourceRoutingEmailDomain(),
				"genesyscloud_routing_email_route":                             resourceRoutingEmailRoute(),
				"genesyscloud_routing_language":                                resourceRoutingLanguage(),
				"genesyscloud_routing_queue":                                   resourceRoutingQueue(),
				"genesyscloud_routing_skill":                                   resourceRoutingSkill(),
				"genesyscloud_routing_skill_group":                             resourceRoutingSkillGroup(),
				"genesyscloud_routing_settings":                                resourceRoutingSettings(),
				"genesyscloud_routing_utilization":                             resourceRoutingUtilization(),
				"genesyscloud_routing_wrapupcode":                              resourceRoutingWrapupCode(),
				"genesyscloud_telephony_providers_edges_did_pool":              resourceTelephonyDidPool(),
				"genesyscloud_telephony_providers_edges_edge_group":            resourceEdgeGroup(),
				"genesyscloud_telephony_providers_edges_extension_pool":        resourceTelephonyExtensionPool(),
				"genesyscloud_telephony_providers_edges_phone":                 resourcePhone(),
				"genesyscloud_telephony_providers_edges_site":                  resourceSite(),
				"genesyscloud_telephony_providers_edges_phonebasesettings":     resourcePhoneBaseSettings(),
				"genesyscloud_telephony_providers_edges_trunkbasesettings":     resourceTrunkBaseSettings(),
				"genesyscloud_telephony_providers_edges_trunk":                 resourceTrunk(),
				"genesyscloud_tf_export":                                       resourceTfExport(),
				"genesyscloud_user":                                            resourceUser(),
				"genesyscloud_user_roles":                                      resourceUserRoles(),
				"genesyscloud_webdeployments_configuration":                    resourceWebDeploymentConfiguration(),
				"genesyscloud_webdeployments_deployment":                       resourceWebDeployment(),
				"genesyscloud_widget_deployment":                               resourceWidgetDeployment(),
			},
			DataSourcesMap: map[string]*schema.Resource{
				"genesyscloud_architect_datatable":                             dataSourceArchitectDatatable(),
				"genesyscloud_architect_ivr":                                   dataSourceArchitectIvr(),
				"genesyscloud_architect_emergencygroup":                        dataSourceArchitectEmergencyGroup(),
				"genesyscloud_architect_schedules":                             dataSourceSchedule(),
				"genesyscloud_architect_schedulegroups":                        dataSourceArchitectScheduleGroups(),
				"genesyscloud_architect_user_prompt":                           dataSourceUserPrompt(),
				"genesyscloud_auth_role":                                       dataSourceAuthRole(),
				"genesyscloud_auth_division":                                   dataSourceAuthDivision(),
				"genesyscloud_auth_division_home":                              dataSourceAuthDivisionHome(),
				"genesyscloud_employeeperformance_externalmetrics_definitions": dataSourceEmployeeperformanceExternalmetricsDefinition(),
				"genesyscloud_flow":                                            dataSourceFlow(),
				"genesyscloud_flow_milestone":                                  dataSourceFlowMilestone(),
				"genesyscloud_flow_outcome":                                    dataSourceFlowOutcome(),
				"genesyscloud_group":                                           dataSourceGroup(),
				"genesyscloud_integration":                                     dataSourceIntegration(),
				"genesyscloud_integration_action":                              dataSourceIntegrationAction(),
				"genesyscloud_integration_credential":                          dataSourceIntegrationCredential(),
				"genesyscloud_journey_action_map":                              dataSourceJourneyActionMap(),
				"genesyscloud_journey_outcome":                                 dataSourceJourneyOutcome(),
				"genesyscloud_journey_segment":                                 dataSourceJourneySegment(),
				"genesyscloud_knowledge_knowledgebase":                         dataSourceKnowledgeKnowledgebase(),
				"genesyscloud_location":                                        dataSourceLocation(),
				"genesyscloud_oauth_client":                                    dataSourceOAuthClient(),
				"genesyscloud_processautomation_trigger":                       dataSourceProcessAutomationTrigger(),
				"genesyscloud_organizations_me":                                dataSourceOrganizationsMe(),
				"genesyscloud_outbound_attempt_limit":                          dataSourceOutboundAttemptLimit(),
				"genesyscloud_outbound_callanalysisresponseset":                dataSourceOutboundCallAnalysisResponseSet(),
				"genesyscloud_outbound_campaign":                               dataSourceOutboundCampaign(),
				"genesyscloud_outbound_campaignrule":                           dataSourceOutboundCampaignRule(),
				"genesyscloud_outbound_callabletimeset":                        dataSourceOutboundCallabletimeset(),
				"genesyscloud_outbound_contact_list":                           dataSourceOutboundContactList(),
				"genesyscloud_outbound_messagingcampaign":                      dataSourceOutboundMessagingcampaign(),
				"genesyscloud_outbound_contactlistfilter":                      dataSourceOutboundContactListFilter(),
				"genesyscloud_outbound_ruleset":                                dataSourceOutboundRuleset(),
				"genesyscloud_outbound_sequence":                               dataSourceOutboundSequence(),
				"genesyscloud_outbound_dnclist":                                dataSourceOutboundDncList(),
				"genesyscloud_quality_forms_evaluation":                        dataSourceQualityFormsEvaluations(),
				"genesyscloud_quality_forms_survey":                            dataSourceQualityFormsSurvey(),
				"genesyscloud_recording_media_retention_policy":                dataSourceRecordingMediaRetentionPolicy(),
				"genesyscloud_responsemanagement_library":                      dataSourceResponsemanagementLibrary(),
				"genesyscloud_responsemanagement_responseasset":                dataSourceResponseManagamentResponseAsset(),
				"genesyscloud_routing_language":                                dataSourceRoutingLanguage(),
				"genesyscloud_routing_queue":                                   dataSourceRoutingQueue(),
				"genesyscloud_routing_settings":                                dataSourceRoutingSettings(),
				"genesyscloud_routing_skill":                                   dataSourceRoutingSkill(),
				"genesyscloud_routing_skill_group":                             dataSourceRoutingSkillGroup(),
				"genesyscloud_routing_email_domain":                            dataSourceRoutingEmailDomain(),
				"genesyscloud_routing_wrapupcode":                              dataSourceRoutingWrapupcode(),
				"genesyscloud_script":                                          dataSourceScript(),
				"genesyscloud_station":                                         dataSourceStation(),
				"genesyscloud_user":                                            dataSourceUser(),
				"genesyscloud_telephony_providers_edges_did":                   dataSourceDid(),
				"genesyscloud_telephony_providers_edges_did_pool":              dataSourceDidPool(),
				"genesyscloud_telephony_providers_edges_edge_group":            dataSourceEdgeGroup(),
				"genesyscloud_telephony_providers_edges_extension_pool":        dataSourceExtensionPool(),
				"genesyscloud_telephony_providers_edges_site":                  dataSourceSite(),
				"genesyscloud_telephony_providers_edges_linebasesettings":      dataSourceLineBaseSettings(),
				"genesyscloud_telephony_providers_edges_phone":                 dataSourcePhone(),
				"genesyscloud_telephony_providers_edges_phonebasesettings":     dataSourcePhoneBaseSettings(),
				"genesyscloud_telephony_providers_edges_trunk":                 dataSourceTrunk(),
				"genesyscloud_telephony_providers_edges_trunkbasesettings":     dataSourceTrunkBaseSettings(),
				"genesyscloud_webdeployments_configuration":                    dataSourceWebDeploymentsConfiguration(),
				"genesyscloud_webdeployments_deployment":                       dataSourceWebDeploymentsDeployment(),
				"genesyscloud_widget_deployment":                               dataSourceWidgetDeployments(),
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
