package provider_registrar

import (
	"sync"

	gcloud "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud"
	dt "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/architect_datatable"
	dtr "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/architect_datatable_row"
	emergencyGroup "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/architect_emergencygroup"
	flow "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/architect_flow"
	grammar "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/architect_grammar"
	grammarLanguage "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/architect_grammar_language"
	archIvr "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/architect_ivr"
	architectSchedulegroups "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/architect_schedulegroups"
	architectSchedules "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/architect_schedules"
	userPrompt "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/architect_user_prompt"
	authDivision "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/auth_division"
	authRole "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/auth_role"
	authorizatioProduct "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/authorization_product"
	integrationInstagram "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/conversations_messaging_integrations_instagram"
	cMessagingOpen "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/conversations_messaging_integrations_open"
	cMessagingWhatsapp "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/conversations_messaging_integrations_whatsapp"
	cMessageSettings "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/conversations_messaging_settings"
	cMessageSettingsDefault "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/conversations_messaging_settings_default"
	supportedContent "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/conversations_messaging_supportedcontent"
	cmSupportedContentDefault "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/conversations_messaging_supportedcontent_default"
	employeeperformanceExternalmetricsDefinition "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/employeeperformance_externalmetrics_definitions"
	externalContacts "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/external_contacts"
	externalSource "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/external_contacts_external_source"
	externalOrganization "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/external_contacts_organization"
	externalUser "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/external_user"
	flowLogLevel "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/flow_loglevel"
	flowMilestone "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/flow_milestone"
	flowOutcome "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/flow_outcome"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/group"
	groupRoles "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/group_roles"
	idpAdfs "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/idp_adfs"
	idpGeneric "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/idp_generic"
	idpGsuite "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/idp_gsuite"
	idpOkta "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/idp_okta"
	idpOneLogin "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/idp_onelogin"
	idpPing "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/idp_ping"
	idpSalesforce "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/idp_salesforce"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/integration"
	integrationAction "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/integration_action"
	integrationCred "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/integration_credential"
	integrationCustomAuth "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/integration_custom_auth_action"
	integrationFacebook "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/integration_facebook"
	journeyActionMap "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/journey_action_map"
	journeyActionTemplate "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/journey_action_template"
	journeyOutcome "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/journey_outcome"
	journeyOutcomePredictor "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/journey_outcome_predictor"
	journeySegment "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/journey_segment"
	journeyViewSchedule "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/journey_view_schedule"
	journeyViews "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/journey_views"
	knowledgeCategory "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/knowledge_category"
	knowledgeDocument "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/knowledge_document"
	knowledgeDocumentVariation "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/knowledge_document_variation"
	knowledgeKnowledgebase "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/knowledge_knowledgebase"
	knowledgeLabel "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/knowledge_label"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/location"
	oauth "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/oauth_client"
	oAuthSettings "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/organization_authentication_settings"
	oPresenceDefinition "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/organization_presence_definition"
	oAuthPairing "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/orgauthorization_pairing"
	obAttemptLimit "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/outbound_attempt_limit"
	obCallableTimeset "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/outbound_callabletimeset"
	obCallResponseSet "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/outbound_callanalysisresponseset"
	obCampaign "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/outbound_campaign"
	obCampaignRule "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/outbound_campaignrule"
	obContactList "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/outbound_contact_list"
	outboundContactListContact "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/outbound_contact_list_contact"
	obContactListTemplate "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/outbound_contact_list_template"
	obContactListFilter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/outbound_contactlistfilter"
	obDigitalRuleSet "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/outbound_digitalruleset"
	obDncList "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/outbound_dnclist"
	obfst "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/outbound_filespecificationtemplate"
	obMessagingCampaign "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/outbound_messagingcampaign"
	obs "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/outbound_ruleset"
	obSequence "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/outbound_sequence"
	obSettings "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/outbound_settings"
	obwm "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/outbound_wrapupcode_mappings"
	pat "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/process_automation_trigger"
	qualityFormsEvaluation "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/quality_forms_evaluation"
	qualityFormsSurvey "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/quality_forms_survey"
	recMediaRetPolicy "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/recording_media_retention_policy"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_register"
	respmanagementLibrary "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/responsemanagement_library"
	responsemanagementResponse "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/responsemanagement_response"
	responsemanagementResponseasset "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/responsemanagement_responseasset"
	routingEmailDomain "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/routing_email_domain"
	routingEmailRoute "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/routing_email_route"
	routingLanguage "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/routing_language"
	routingQueue "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/routing_queue"
	routingQueueConditionalGroupRouting "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/routing_queue_conditional_group_routing"
	routingQueueOutboundEmailAddress "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/routing_queue_outbound_email_address"
	routingSettings "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/routing_settings"
	routingSkill "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/routing_skill"
	routingSkillGroup "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/routing_skill_group"
	smsAddresses "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/routing_sms_addresses"
	routingUtilization "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/routing_utilization"
	routingUtilizationLabel "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/routing_utilization_label"
	routingWrapupcode "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/routing_wrapupcode"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/scripts"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/station"
	workbin "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/task_management_workbin"
	workitem "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/task_management_workitem"
	workitemSchema "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/task_management_workitem_schema"
	worktype "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/task_management_worktype"
	workitemDateBasedRule "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/task_management_worktype_flow_datebased_rule"
	workitemOnAttributeChangeRule "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/task_management_worktype_flow_onattributechange_rule"
	workitemOnCreateRule "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/task_management_worktype_flow_oncreate_rule"
	worktypeStatus "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/task_management_worktype_status"
	worktypeStatusTransition "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/task_management_worktype_status_transition"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/team"
	did "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/telephony_providers_edges_did"
	didPool "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/telephony_providers_edges_did_pool"
	edgeGroup "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/telephony_providers_edges_edge_group"
	extPool "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/telephony_providers_edges_extension_pool"
	lineBaseSettings "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/telephony_providers_edges_linebasesettings"
	edgePhone "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/telephony_providers_edges_phone"
	phoneBaseSettings "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/telephony_providers_edges_phonebasesettings"
	edgeSite "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/telephony_providers_edges_site"
	siteOutboundRoutes "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/telephony_providers_edges_site_outbound_route"
	edgesTrunk "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/telephony_providers_edges_trunk"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/telephony_providers_edges_trunkbasesettings"
	tfexp "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/tfexporter"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/user"
	userRoles "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/user_roles"
	webDeployConfig "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/webdeployments_configuration"
	webDeployDeploy "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/webdeployments_deployment"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

/*
Package provider_registrar handles the registration and management of Terraform provider resources,
data sources, and resource exporters for the Genesys Cloud provider.

Key Components:

1. Package-level variables:
   - providerResources: Maps resource names to schema.Resource definitions
   - providerDataSources: Maps data source names to schema.Resource definitions
   - resourceExporters: Maps resource names to ResourceExporter definitions

2. RegisterInstance struct:
   Thread-safe structure that manages resource registration using mutex locks:
   - resourceMapMutex: Controls access to provider resources
   - datasourceMapMutex: Controls access to data sources
   - exporterMapMutex: Controls access to resource exporters

3. Main Functions:
   - GetProviderResources(): Returns registered resources and data sources
   - resourceMapsAreRegistered(): Validates if resources are properly registered
   - registerResources(): Initializes and registers all Genesys Cloud resources

Each resource is registered using the SetRegistrar method with a RegisterInstance.

Note: The package resource_register cannot take on this responsibility without incurring
circular dependency issues with the package tf_exporter
*/

var (
	providerResources   = make(map[string]*schema.Resource)
	providerDataSources = make(map[string]*schema.Resource)
	resourceExporters   = make(map[string]*resourceExporter.ResourceExporter)
)

type RegisterInstance struct {
	resourceMapMutex   sync.RWMutex
	datasourceMapMutex sync.RWMutex
	exporterMapMutex   sync.RWMutex
}

func GetProviderResources() (resources map[string]*schema.Resource, datasources map[string]*schema.Resource) {
	if !resourceMapsAreRegistered() {
		registerResources()
	}
	return providerResources, providerDataSources
}

func GetResourceExporters() (exporters map[string]*resourceExporter.ResourceExporter) {
	if !resourceMapsAreRegistered() {
		registerResources()
	}
	return resourceExporters
}

func resourceMapsAreRegistered() bool {
	if providerResources == nil || providerDataSources == nil || resourceExporters == nil {
		return false
	}
	if len(providerResources) == 0 || len(providerDataSources) == 0 || len(resourceExporters) == 0 {
		return false
	}
	return true
}

func registerResources() {
	regInstance := &RegisterInstance{}
	authRole.SetRegistrar(regInstance)                                     //Registering auth_role
	authDivision.SetRegistrar(regInstance)                                 //Registering auth_division
	oauth.SetRegistrar(regInstance)                                        //Registering oauth_client
	dt.SetRegistrar(regInstance)                                           //Registering architect data table
	dtr.SetRegistrar(regInstance)                                          //Registering architect data table row
	emergencyGroup.SetRegistrar(regInstance)                               //Registering architect emergency group
	architectSchedulegroups.SetRegistrar(regInstance)                      //Registering architect schedule groups
	architectSchedules.SetRegistrar(regInstance)                           //Registering architect schedules
	employeeperformanceExternalmetricsDefinition.SetRegistrar(regInstance) //Registering employee performance external metrics definitions
	grammar.SetRegistrar(regInstance)                                      //Registering architect grammar
	grammarLanguage.SetRegistrar(regInstance)                              //Registering architect grammar language
	groupRoles.SetRegistrar(regInstance)                                   //Registering group roles
	edgePhone.SetRegistrar(regInstance)                                    //Registering telephony  providers edges phone
	edgeSite.SetRegistrar(regInstance)                                     //Registering telephony providers edges site
	siteOutboundRoutes.SetRegistrar(regInstance)                           //Registering telephony providers edges site outbound routes
	flow.SetRegistrar(regInstance)                                         //Registering architect flow
	flowLogLevel.SetRegistrar(regInstance)                                 //Registering flow log Level
	flowMilestone.SetRegistrar(regInstance)                                //Registering flow milestone
	flowOutcome.SetRegistrar(regInstance)                                  //Registering flow outcome
	station.SetRegistrar(regInstance)                                      //Registering station
	pat.SetRegistrar(regInstance)                                          //Registering process automation triggers
	obs.SetRegistrar(regInstance)                                          //Resistering outbound ruleset
	obMessagingCampaign.SetRegistrar(regInstance)                          //Registering outbound messaging campaign
	obwm.SetRegistrar(regInstance)                                         //Registering outbound wrapup code mappings
	oAuthSettings.SetRegistrar(regInstance)                                //Registering organization authentication settings
	oPresenceDefinition.SetRegistrar(regInstance)                          //Registering organization presence definition
	gcloud.SetRegistrar(regInstance)                                       //Registering genesyscloud
	obAttemptLimit.SetRegistrar(regInstance)                               //Registering outbound attempt limit
	obCallableTimeset.SetRegistrar(regInstance)                            //Registering outbound callable timeset
	obCampaign.SetRegistrar(regInstance)                                   //Registering outbound campaign
	obContactList.SetRegistrar(regInstance)                                //Registering outbound contact list
	obContactListFilter.SetRegistrar(regInstance)                          //Registering outbound contact list filter
	obContactListTemplate.SetRegistrar(regInstance)                        //Registering outbound contact list template
	obSequence.SetRegistrar(regInstance)                                   //Registering outbound sequence
	obCampaignRule.SetRegistrar(regInstance)                               //Registering outbound campaignrule
	obSettings.SetRegistrar(regInstance)                                   //Registering outbound settings
	obCallResponseSet.SetRegistrar(regInstance)                            //Registering outbound call analysis response set
	obCampaign.SetRegistrar(regInstance)                                   //Registering outbound campaign
	obfst.SetRegistrar(regInstance)                                        //Registering outbound file specification template
	obDncList.SetRegistrar(regInstance)                                    //Registering outbound dnclist
	obDigitalRuleSet.SetRegistrar(regInstance)                             //Registering outbound digital ruleset
	oAuthPairing.SetRegistrar(regInstance)                                 //Registering orgauthorization pairing
	scripts.SetRegistrar(regInstance)                                      //Registering Scripts
	smsAddresses.SetRegistrar(regInstance)                                 //Registering routing sms addresses
	idpAdfs.SetRegistrar(regInstance)                                      //Registering idp adfs
	idpSalesforce.SetRegistrar(regInstance)                                //Registering idp salesforce
	idpOkta.SetRegistrar(regInstance)                                      //Registering idp okta
	idpOneLogin.SetRegistrar(regInstance)                                  //Registering idp onelogin
	idpGeneric.SetRegistrar(regInstance)                                   //Registering idp generic
	idpPing.SetRegistrar(regInstance)                                      //Registering idp ping
	idpGsuite.SetRegistrar(regInstance)                                    //Registering idp gsuite
	integration.SetRegistrar(regInstance)                                  //Registering integrations
	integrationCustomAuth.SetRegistrar(regInstance)                        //Registering integrations custom auth actions
	integrationAction.SetRegistrar(regInstance)                            //Registering integrations actions
	integrationCred.SetRegistrar(regInstance)                              //Registering integrations credentials
	integrationFacebook.SetRegistrar(regInstance)                          //Registering integrations Facebook
	integrationInstagram.SetRegistrar(regInstance)                         //Registering integrations Instagram
	recMediaRetPolicy.SetRegistrar(regInstance)                            //Registering recording media retention policies
	responsemanagementResponse.SetRegistrar(regInstance)                   //Registering responsemanagement responses
	responsemanagementResponseasset.SetRegistrar(regInstance)              //Registering responsemanagement response asset
	respmanagementLibrary.SetRegistrar(regInstance)                        //Registering responsemanagement library
	routingEmailRoute.SetRegistrar(regInstance)                            //Registering routing email route
	did.SetRegistrar(regInstance)                                          //Registering telephony did
	didPool.SetRegistrar(regInstance)                                      //Registering telephony did pools
	archIvr.SetRegistrar(regInstance)                                      //Registering architect ivr
	workbin.SetRegistrar(regInstance)                                      //Registering task management workbin
	workitemSchema.SetRegistrar(regInstance)                               //Registering task management workitem schema
	worktype.SetRegistrar(regInstance)                                     //Registering task management worktype
	worktypeStatus.SetRegistrar(regInstance)                               //Registering task management worktype status
	worktypeStatusTransition.SetRegistrar(regInstance)                     //Registering task management worktype status Transition
	workitem.SetRegistrar(regInstance)                                     //Registering task management workitem
	workitemOnCreateRule.SetRegistrar(regInstance)                         //Registering task management oncreate rule
	workitemOnAttributeChangeRule.SetRegistrar(regInstance)                //Registering task management onattributechange rule
	workitemDateBasedRule.SetRegistrar(regInstance)                        //Registering task management datebased rule
	externalContacts.SetRegistrar(regInstance)                             //Registering external contacts
	externalUser.SetRegistrar(regInstance)                                 //Registering external user identity
	team.SetRegistrar(regInstance)                                         //Registering team
	telephony_providers_edges_trunkbasesettings.SetRegistrar(regInstance)  //Registering telephony_providers_edges_trunkbasesettings package
	edgeGroup.SetRegistrar(regInstance)                                    //Registering edges edge group
	webDeployConfig.SetRegistrar(regInstance)                              //Registering webdeployments_config
	webDeployDeploy.SetRegistrar(regInstance)                              //Registering webdeployments_deploy
	authorizatioProduct.SetRegistrar(regInstance)                          //Registering Authorization Product
	extPool.SetRegistrar(regInstance)                                      //Registering Extension Pool
	phoneBaseSettings.SetRegistrar(regInstance)                            //Registering Phone Base Settings
	lineBaseSettings.SetRegistrar(regInstance)                             //Registering Line Base Settings
	edgesTrunk.SetRegistrar(regInstance)                                   //Registering Edges Trunk Settings
	resourceExporter.SetRegisterExporter(resourceExporters)                //Registering register exporters
	userRoles.SetRegistrar(regInstance)                                    //Registering user roles
	user.SetRegistrar(regInstance)                                         //Registering user
	journeyOutcomePredictor.SetRegistrar(regInstance)                      //Registering journey outcome predictor
	journeyActionTemplate.SetRegistrar(regInstance)                        //Registering journey action template
	journeyOutcome.SetRegistrar(regInstance)                               //Registering journey outcome
	group.SetRegistrar(regInstance)                                        //Registering group
	userPrompt.SetRegistrar(regInstance)                                   //Registering user prompt
	routingQueue.SetRegistrar(regInstance)                                 //Registering routing queue
	routingQueueConditionalGroupRouting.SetRegistrar(regInstance)          //Registering routing queue conditional group routing
	routingQueueOutboundEmailAddress.SetRegistrar(regInstance)             //Registering routing queue outbound email address
	outboundContactListContact.SetRegistrar(regInstance)                   //Registering outbound contact list contact
	routingSettings.SetRegistrar(regInstance)                              //Registering routing Settings
	routingUtilization.SetRegistrar(regInstance)                           //Registering routing utilization
	routingUtilizationLabel.SetRegistrar(regInstance)                      //Registering routing utilization label
	journeyViews.SetRegistrar(regInstance)                                 //Registering journey views
	journeyViewSchedule.SetRegistrar(regInstance)                          //Registering journey view schedule
	journeySegment.SetRegistrar(regInstance)                               //Registering journey Segment
	journeyActionMap.SetRegistrar(regInstance)                             //Registering journey Action Map
	routingWrapupcode.SetRegistrar(regInstance)                            //Registering routing wrapupcode
	routingLanguage.SetRegistrar(regInstance)                              //Registering Routing Language
	routingEmailDomain.SetRegistrar(regInstance)                           //Registering Routing Email Domain
	supportedContent.SetRegistrar(regInstance)                             //Registering Supported Content
	routingSkill.SetRegistrar(regInstance)                                 //Registering Routing Skill
	cMessageSettings.SetRegistrar(regInstance)                             //Registering conversations messaging settings
	routingSkillGroup.SetRegistrar(regInstance)                            //Registering routing skill group
	cMessageSettingsDefault.SetRegistrar(regInstance)                      //Registering conversations messaging settings default
	cmSupportedContentDefault.SetRegistrar(regInstance)                    //Registering conversations supported content default
	cMessagingOpen.SetRegistrar(regInstance)                               //Registering conversations messaging open
	cMessagingWhatsapp.SetRegistrar(regInstance)                           //Registering conversations messaging whatsapp
	location.SetRegistrar(regInstance)                                     //Registering location
	knowledgeDocument.SetRegistrar(regInstance)                            //Registering knowledge document
	knowledgeDocumentVariation.SetRegistrar(regInstance)                   //Registering knowledge document variation
	externalOrganization.SetRegistrar(regInstance)                         //Registering external organization
	externalSource.SetRegistrar(regInstance)                               //Registering external source
	knowledgeCategory.SetRegistrar(regInstance)                            //Registering knowledge category
	knowledgeLabel.SetRegistrar(regInstance)                               //Registering Knowledge Label
	knowledgeKnowledgebase.SetRegistrar(regInstance)                       //Registering Knowledge base
	qualityFormsEvaluation.SetRegistrar(regInstance)                       //Registering quality forms evaluation
	qualityFormsSurvey.SetRegistrar(regInstance)                           //Registering quality forms survey
	// setting resources for Use cases  like TF export where provider is used in resource classes.
	tfexp.SetRegistrar(regInstance) //Registering tf exporter
	registrar.SetResources(providerResources, providerDataSources)
}

func (r *RegisterInstance) RegisterResource(resourceType string, resource *schema.Resource) {
	r.resourceMapMutex.Lock()
	defer r.resourceMapMutex.Unlock()
	providerResources[resourceType] = resource
}

func (r *RegisterInstance) RegisterDataSource(dataSourceType string, datasource *schema.Resource) {
	r.datasourceMapMutex.Lock()
	defer r.datasourceMapMutex.Unlock()
	providerDataSources[dataSourceType] = datasource
}

func (r *RegisterInstance) RegisterExporter(exporterName string, resourceExporter *resourceExporter.ResourceExporter) {
	r.exporterMapMutex.Lock()
	defer r.exporterMapMutex.Unlock()
	resourceExporters[exporterName] = resourceExporter
}
