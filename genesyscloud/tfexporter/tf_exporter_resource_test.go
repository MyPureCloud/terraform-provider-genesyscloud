package tfexporter

import (
	"testing"

	gcloud "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud"
	dt "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/architect_datatable"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/architect_datatable_row"
	emergencyGroup "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/architect_emergencygroup"
	flow "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/architect_flow"
	grammar "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/architect_grammar"
	grammarLanguage "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/architect_grammar_language"
	archIvr "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/architect_ivr"
	architectSchedulegroups "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/architect_schedulegroups"
	architectSchedules "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/architect_schedules"
	authDivision "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/auth_division"
	authRole "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/auth_role"
	integrationInstagram "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/conversations_messaging_integrations_instagram"
	cMessagingOpen "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/conversations_messaging_integrations_open"
	cMessagingWhatsapp "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/conversations_messaging_integrations_whatsapp"
	cMessagingSettings "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/conversations_messaging_settings"
	supportedContent "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/conversations_messaging_supportedcontent"
	defaultSupportedContent "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/conversations_messaging_supportedcontent_default"
	employeeperformanceExternalmetricsDefinition "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/employeeperformance_externalmetrics_definitions"
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
	integration "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/integration"
	integrationAction "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/integration_action"
	integrationCred "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/integration_credential"
	integrationFacebook "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/integration_facebook"
	journeyActionMap "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/journey_action_map"
	journeyActionTemplate "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/journey_action_template"
	journeyOutcome "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/journey_outcome"
	journeySegment "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/journey_segment"
	journeyViewSchedule "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/journey_view_schedule"
	knowledgeDocument "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/knowledge_document"
	knowledgeDocumentVariation "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/knowledge_document_variation"
	knowledgeKnowledgebase "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/knowledge_knowledgebase"
	knowledgeLabel "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/knowledge_label"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/oauth_client"
	oAuthSettings "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/organization_authentication_settings"
	oAuthPairing "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/orgauthorization_pairing"
	outboundAttemptLimit "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/outbound_attempt_limit"
	obCallableTimeset "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/outbound_callabletimeset"
	obCallResponseSet "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/outbound_callanalysisresponseset"
	obCampaign "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/outbound_campaign"
	obCampaignRule "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/outbound_campaignrule"
	outboundContactList "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/outbound_contact_list"
	outboundContactListContact "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/outbound_contact_list_contact"
	outboundContactListTemplate "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/outbound_contact_list_template"
	obContactListFilter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/outbound_contactlistfilter"
	obDigitalRuleset "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/outbound_digitalruleset"
	obDncList "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/outbound_dnclist"
	obfst "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/outbound_filespecificationtemplate"
	obMessagingCampaign "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/outbound_messagingcampaign"
	obRuleset "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/outbound_ruleset"
	obSequence "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/outbound_sequence"
	obSettings "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/outbound_settings"
	obw "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/outbound_wrapupcode_mappings"
	pat "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/process_automation_trigger"
	qualityFormsEvaluation "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/quality_forms_evaluation"
	qualityFormsSurvey "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/quality_forms_survey"
	recMediaRetPolicy "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/recording_media_retention_policy"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	respmanagementLibrary "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/responsemanagement_library"
	responsemanagementResponse "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/responsemanagement_response"
	respManagementRespAsset "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/responsemanagement_responseasset"
	routingEmailDomain "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/routing_email_domain"
	routingEmailRoute "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/routing_email_route"
	routinglanguage "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/routing_language"
	routingQueue "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/routing_queue"
	routingQueueConditionalGroupRouting "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/routing_queue_conditional_group_routing"
	routingQueueOutboundEmailAddress "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/routing_queue_outbound_email_address"
	routingSettings "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/routing_settings"
	routingSkill "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/routing_skill"
	routingSkillGroup "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/routing_skill_group"
	routingSmsAddress "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/routing_sms_addresses"
	routingUtilization "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/routing_utilization"
	routingUtilizationLabel "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/routing_utilization_label"
	routingWrapupcode "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/routing_wrapupcode"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/scripts"
	workbin "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/task_management_workbin"
	workitemSchema "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/task_management_workitem_schema"
	worktype "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/task_management_worktype"
	worktypeStatus "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/task_management_worktype_status"
	didPool "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/telephony_providers_edges_did_pool"
	edgeGroup "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/telephony_providers_edges_edge_group"
	edgeExtension "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/telephony_providers_edges_extension_pool"
	phonebaseSettings "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/telephony_providers_edges_phonebasesettings"
	outboundRoute "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/telephony_providers_edges_site_outbound_route"
	edgesTrunk "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/telephony_providers_edges_trunk"
	tbs "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/telephony_providers_edges_trunkbasesettings"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/user"
	userRoles "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/user_roles"
	webdeployConfig "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/webdeployments_configuration"
	webdeployDeploy "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/webdeployments_deployment"

	edgePhone "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/telephony_providers_edges_phone"
	edgeSite "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/telephony_providers_edges_site"

	userPrompt "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/architect_user_prompt"
	externalOrganization "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/external_contacts_organization"
	externalUser "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/external_user"
	knowledgeCategory "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/knowledge_category"
	location "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/location"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// had to do an init here since manual function call in export_test will not work since exporter already loaded
// at ValidateFunc: gcloud.ValidateSubStringInSlice(gcloud.GetAvailableExporterTypes()),
func initTestResources() {
	resourceExporters = make(map[string]*resourceExporter.ResourceExporter)
	providerResources = make(map[string]*schema.Resource)
	providerDataSources = make(map[string]*schema.Resource)

	regInstance := &registerTestInstance{}

	// register exporters first and then resources. Since there is a dependency of exporters on Resources
	regInstance.registerTestExporters()
	regInstance.registerTestResources()
	regInstance.registerTestDataSources()
}

type registerTestInstance struct {
}

func (r *registerTestInstance) registerTestResources() {
	providerResources[oAuthSettings.ResourceType] = oAuthSettings.ResourceOrganizationAuthenticationSettings()
	providerResources[oAuthPairing.ResourceType] = oAuthPairing.ResourceOrgauthorizationPairing()
	providerResources[grammar.ResourceType] = grammar.ResourceArchitectGrammar()
	providerResources[grammarLanguage.ResourceType] = grammarLanguage.ResourceArchitectGrammarLanguage()
	providerResources[dt.ResourceType] = dt.ResourceArchitectDatatable()
	providerResources[architect_datatable_row.ResourceType] = architect_datatable_row.ResourceArchitectDatatableRow()
	providerResources[emergencyGroup.ResourceType] = emergencyGroup.ResourceArchitectEmergencyGroup()
	providerResources[flow.ResourceType] = flow.ResourceArchitectFlow()
	providerResources[flowMilestone.ResourceType] = flowMilestone.ResourceFlowMilestone()
	providerResources[flowOutcome.ResourceType] = flowOutcome.ResourceFlowOutcome()
	providerResources[archIvr.ResourceType] = archIvr.ResourceArchitectIvrConfig()
	providerResources[architectSchedules.ResourceType] = architectSchedules.ResourceArchitectSchedules()
	providerResources[architectSchedulegroups.ResourceType] = architectSchedulegroups.ResourceArchitectSchedulegroups()
	providerResources[userPrompt.ResourceType] = userPrompt.ResourceArchitectUserPrompt()
	providerResources[authRole.ResourceType] = authRole.ResourceAuthRole()
	providerResources[authDivision.ResourceType] = authDivision.ResourceAuthDivision()
	providerResources[employeeperformanceExternalmetricsDefinition.ResourceType] = employeeperformanceExternalmetricsDefinition.ResourceEmployeeperformanceExternalmetricsDefinition()
	providerResources[flowLogLevel.ResourceType] = flowLogLevel.ResourceFlowLoglevel()
	providerResources[group.ResourceType] = group.ResourceGroup()
	providerResources[groupRoles.ResourceType] = groupRoles.ResourceGroupRoles()
	providerResources[idpAdfs.ResourceType] = idpAdfs.ResourceIdpAdfs()
	providerResources[idpGeneric.ResourceType] = idpGeneric.ResourceIdpGeneric()
	providerResources[idpGsuite.ResourceType] = idpGsuite.ResourceIdpGsuite()
	providerResources[idpOkta.ResourceType] = idpOkta.ResourceIdpOkta()
	providerResources[idpOneLogin.ResourceType] = idpOneLogin.ResourceIdpOnelogin()
	providerResources[idpPing.ResourceType] = idpPing.ResourceIdpPing()
	providerResources[idpSalesforce.ResourceType] = idpSalesforce.ResourceIdpSalesforce()
	providerResources[integration.ResourceType] = integration.ResourceIntegration()
	providerResources[integrationAction.ResourceType] = integrationAction.ResourceIntegrationAction()
	providerResources[integrationCred.ResourceType] = integrationCred.ResourceIntegrationCredential()
	providerResources[integrationFacebook.ResourceType] = integrationFacebook.ResourceIntegrationFacebook()
	providerResources[integrationInstagram.ResourceType] = integrationInstagram.ResourceConversationsMessagingIntegrationsInstagram()
	providerResources[cMessagingWhatsapp.ResourceType] = cMessagingWhatsapp.ResourceConversationsMessagingIntegrationsWhatsapp()
	providerResources[journeyActionMap.ResourceType] = journeyActionMap.ResourceJourneyActionMap()
	providerResources[journeyActionTemplate.ResourceType] = journeyActionTemplate.ResourceJourneyActionTemplate()
	providerResources[journeyViewSchedule.ResourceType] = journeyViewSchedule.ResourceJourneyViewSchedule()
	providerResources[knowledgeDocument.ResourceType] = knowledgeDocument.ResourceKnowledgeDocument()
	providerResources[knowledgeLabel.ResourceType] = knowledgeLabel.ResourceKnowledgeLabel()
	providerResources[knowledgeDocumentVariation.ResourceType] = knowledgeDocumentVariation.ResourceKnowledgeDocumentVariation()
	providerResources[knowledgeKnowledgebase.ResourceType] = knowledgeKnowledgebase.ResourceKnowledgeKnowledgebase()
	providerResources[location.ResourceType] = location.ResourceLocation()
	providerResources[recMediaRetPolicy.ResourceType] = recMediaRetPolicy.ResourceMediaRetentionPolicy()
	providerResources[oauth_client.ResourceType] = oauth_client.ResourceOAuthClient()
	providerResources[obSettings.ResourceType] = obSettings.ResourceOutboundSettings()
	providerResources[user.ResourceType] = user.ResourceUser()
	providerResources[respmanagementLibrary.ResourceType] = respmanagementLibrary.ResourceResponsemanagementLibrary()
	providerResources[respManagementRespAsset.ResourceType] = respManagementRespAsset.ResourceResponseManagementResponseAsset()
	providerResources[routingEmailDomain.ResourceType] = routingEmailDomain.ResourceRoutingEmailDomain()
	providerResources[routingEmailRoute.ResourceType] = routingEmailRoute.ResourceRoutingEmailRoute()
	providerResources[routinglanguage.ResourceType] = routinglanguage.ResourceRoutingLanguage()
	providerResources[routingQueue.ResourceType] = routingQueue.ResourceRoutingQueue()
	providerResources[routingQueueConditionalGroupRouting.ResourceType] = routingQueueConditionalGroupRouting.ResourceRoutingQueueConditionalGroupRouting()
	providerResources[routingQueueOutboundEmailAddress.ResourceType] = routingQueueOutboundEmailAddress.ResourceRoutingQueueOutboundEmailAddress()
	providerResources[routingSkill.ResourceType] = routingSkill.ResourceRoutingSkill()
	providerResources[routingSettings.ResourceType] = routingSettings.ResourceRoutingSettings()
	providerResources[routingUtilization.ResourceType] = routingUtilization.ResourceRoutingUtilization()
	providerResources[routingUtilizationLabel.ResourceType] = routingUtilizationLabel.ResourceRoutingUtilizationLabel()
	providerResources[routingWrapupcode.ResourceType] = routingWrapupcode.ResourceRoutingWrapupCode()
	providerResources[edgeExtension.ResourceType] = edgeExtension.ResourceTelephonyExtensionPool()
	providerResources[edgePhone.ResourceType] = edgePhone.ResourcePhone()
	providerResources[edgeSite.ResourceType] = edgeSite.ResourceSite()
	providerResources[outboundRoute.ResourceType] = outboundRoute.ResourceSiteOutboundRoute()
	providerResources[phonebaseSettings.ResourceType] = phonebaseSettings.ResourcePhoneBaseSettings()
	providerResources[tbs.ResourceType] = tbs.ResourceTrunkBaseSettings()
	providerResources[edgesTrunk.ResourceType] = edgesTrunk.ResourceTrunk()
	providerResources[edgeGroup.ResourceType] = edgeGroup.ResourceEdgeGroup()
	providerResources[userRoles.ResourceType] = userRoles.ResourceUserRoles()
	providerResources[webdeployDeploy.ResourceType] = webdeployDeploy.ResourceWebDeployment()
	providerResources[webdeployConfig.ResourceType] = webdeployConfig.ResourceWebDeploymentConfiguration()
	providerResources[pat.ResourceType] = pat.ResourceProcessAutomationTrigger()
	providerResources[outboundAttemptLimit.ResourceType] = outboundAttemptLimit.ResourceOutboundAttemptLimit()
	providerResources[obCallResponseSet.ResourceType] = obCallResponseSet.ResourceOutboundCallanalysisresponseset()
	providerResources[obCallableTimeset.ResourceType] = obCallableTimeset.ResourceOutboundCallabletimeset()
	providerResources[obCampaign.ResourceType] = obCampaign.ResourceOutboundCampaign()
	providerResources[outboundContactList.ResourceType] = outboundContactList.ResourceOutboundContactList()
	providerResources[outboundContactListTemplate.ResourceType] = outboundContactListTemplate.ResourceOutboundContactListTemplate()
	providerResources[outboundContactListContact.ResourceType] = outboundContactListContact.ResourceOutboundContactListContact()
	providerResources[obContactListFilter.ResourceType] = obContactListFilter.ResourceOutboundContactlistfilter()
	providerResources[obMessagingCampaign.ResourceType] = obMessagingCampaign.ResourceOutboundMessagingcampaign()
	providerResources[obSequence.ResourceType] = obSequence.ResourceOutboundSequence()
	providerResources[obDncList.ResourceType] = obDncList.ResourceOutboundDncList()
	providerResources[obCampaignRule.ResourceType] = obCampaignRule.ResourceOutboundCampaignrule()
	providerResources[obDigitalRuleset.ResourceType] = obDigitalRuleset.ResourceOutboundDigitalruleset()
	providerResources[obfst.ResourceType] = obfst.ResourceOutboundFileSpecificationTemplate()
	providerResources[obw.ResourceType] = obw.ResourceOutboundWrapUpCodeMappings()
	providerResources[responsemanagementResponse.ResourceType] = responsemanagementResponse.ResourceResponsemanagementResponse()
	providerResources[routingSmsAddress.ResourceType] = routingSmsAddress.ResourceRoutingSmsAddress()
	providerResources[routingSkillGroup.ResourceType] = routingSkillGroup.ResourceRoutingSkillGroup()
	providerResources[didPool.ResourceType] = didPool.ResourceTelephonyDidPool()
	providerResources[scripts.ResourceType] = scripts.ResourceScript()
	providerResources[workbin.ResourceType] = workbin.ResourceTaskManagementWorkbin()
	providerResources[workitemSchema.ResourceType] = workitemSchema.ResourceTaskManagementWorkitemSchema()
	providerResources[worktype.ResourceType] = worktype.ResourceTaskManagementWorktype()
	providerResources[supportedContent.ResourceType] = supportedContent.ResourceSupportedContent()
	providerResources[cMessagingSettings.ResourceType] = cMessagingSettings.ResourceConversationsMessagingSettings()
	providerResources[defaultSupportedContent.ResourceType] = defaultSupportedContent.ResourceConversationsMessagingSupportedcontentDefault()
	providerResources[worktypeStatus.ResourceType] = worktypeStatus.ResourceTaskManagementWorktypeStatus()
	providerResources[cMessagingOpen.ResourceType] = cMessagingOpen.ResourceConversationsMessagingIntegrationsOpen()
	providerResources[externalOrganization.ResourceType] = externalOrganization.ResourceExternalContactsOrganization()
	providerResources[knowledgeCategory.ResourceType] = knowledgeCategory.ResourceKnowledgeCategory()
	providerResources[journeyOutcome.ResourceType] = journeyOutcome.ResourceJourneyOutcome()
	providerResources[externalUser.ResourceType] = externalUser.ResourceExternalUserIdentity()
	providerResources[journeySegment.ResourceType] = journeySegment.ResourceJourneySegment()
	providerResources[qualityFormsEvaluation.ResourceType] = qualityFormsEvaluation.ResourceEvaluationForm()
	providerResources[knowledgeLabel.ResourceType] = knowledgeLabel.ResourceKnowledgeLabel()
	providerResources[qualityFormsSurvey.ResourceType] = qualityFormsSurvey.ResourceQualityFormsSurvey()
	providerResources[ResourceType] = ResourceTfExport()
}

func (r *registerTestInstance) registerTestExporters() {
	RegisterExporter(journeySegment.ResourceType, journeySegment.JourneySegmentExporter())
	RegisterExporter(architectSchedules.ResourceType, architectSchedules.ArchitectSchedulesExporter())
	RegisterExporter(knowledgeCategory.ResourceType, knowledgeCategory.KnowledgeCategoryExporter())
	RegisterExporter(knowledgeLabel.ResourceType, knowledgeLabel.KnowledgeLabelExporter())
	RegisterExporter(knowledgeCategory.ResourceType, knowledgeCategory.KnowledgeCategoryExporter())
	RegisterExporter(knowledgeLabel.ResourceType, knowledgeLabel.KnowledgeLabelExporter())
	RegisterExporter(knowledgeCategory.ResourceType, knowledgeCategory.KnowledgeCategoryExporter())
	RegisterExporter(knowledgeKnowledgebase.ResourceType, knowledgeKnowledgebase.KnowledgeKnowledgebaseExporter())
	RegisterExporter(journeyViewSchedule.ResourceType, journeyViewSchedule.JourneyViewScheduleExporter())
	RegisterExporter(oAuthSettings.ResourceType, oAuthSettings.OrganizationAuthenticationSettingsExporter())
	RegisterExporter(grammar.ResourceType, grammar.ArchitectGrammarExporter())
	RegisterExporter(grammarLanguage.ResourceType, grammarLanguage.ArchitectGrammarLanguageExporter())
	RegisterExporter(dt.ResourceType, dt.ArchitectDatatableExporter())
	RegisterExporter(architect_datatable_row.ResourceType, architect_datatable_row.ArchitectDatatableRowExporter())
	RegisterExporter(emergencyGroup.ResourceType, emergencyGroup.ArchitectEmergencyGroupExporter())
	RegisterExporter(archIvr.ResourceType, archIvr.ArchitectIvrExporter())
	RegisterExporter(architectSchedulegroups.ResourceType, architectSchedulegroups.ArchitectSchedulegroupsExporter())
	RegisterExporter(userPrompt.ResourceType, userPrompt.ArchitectUserPromptExporter())
	RegisterExporter(authDivision.ResourceType, authDivision.AuthDivisionExporter())
	RegisterExporter(authRole.ResourceType, authRole.AuthRoleExporter())
	RegisterExporter(employeeperformanceExternalmetricsDefinition.ResourceType, employeeperformanceExternalmetricsDefinition.EmployeeperformanceExternalmetricsDefinitionExporter())
	RegisterExporter(flow.ResourceType, flow.ArchitectFlowExporter())
	RegisterExporter(flowLogLevel.ResourceType, flowLogLevel.FlowLogLevelExporter())
	RegisterExporter(flowMilestone.ResourceType, flowMilestone.FlowMilestoneExporter())
	RegisterExporter(flowOutcome.ResourceType, flowOutcome.FlowOutcomeExporter())
	RegisterExporter(group.ResourceType, group.GroupExporter())
	RegisterExporter(groupRoles.ResourceType, groupRoles.GroupRolesExporter())
	RegisterExporter(idpAdfs.ResourceType, idpAdfs.IdpAdfsExporter())
	RegisterExporter(idpGeneric.ResourceType, idpGeneric.IdpGenericExporter())
	RegisterExporter(idpGsuite.ResourceType, idpGsuite.IdpGsuiteExporter())
	RegisterExporter(idpOkta.ResourceType, idpOkta.IdpOktaExporter())
	RegisterExporter(idpOneLogin.ResourceType, idpOneLogin.IdpOneloginExporter())
	RegisterExporter(idpPing.ResourceType, idpPing.IdpPingExporter())
	RegisterExporter(idpSalesforce.ResourceType, idpSalesforce.IdpSalesforceExporter())
	RegisterExporter(integration.ResourceType, integration.IntegrationExporter())
	RegisterExporter(integrationAction.ResourceType, integrationAction.IntegrationActionExporter())
	RegisterExporter(integrationCred.ResourceType, integrationCred.IntegrationCredentialExporter())
	RegisterExporter(integrationFacebook.ResourceType, integrationFacebook.IntegrationFacebookExporter())
	RegisterExporter(integrationInstagram.ResourceType, integrationInstagram.ConversationsMessagingIntegrationsInstagramExporter())
	RegisterExporter(journeyActionMap.ResourceType, journeyActionMap.JourneyActionMapExporter())
	RegisterExporter(journeyActionTemplate.ResourceType, journeyActionTemplate.JourneyActionTemplateExporter())
	RegisterExporter(journeyOutcome.ResourceType, journeyOutcome.JourneyOutcomeExporter())
	RegisterExporter(journeySegment.ResourceType, journeySegment.JourneySegmentExporter())
	RegisterExporter(knowledgeDocument.ResourceType, knowledgeDocument.KnowledgeDocumentExporter())
	RegisterExporter(knowledgeDocumentVariation.ResourceType, knowledgeDocumentVariation.KnowledgeDocumentVariationExporter())
	RegisterExporter(location.ResourceType, location.LocationExporter())
	RegisterExporter(oauth_client.ResourceType, oauth_client.OauthClientExporter())
	RegisterExporter(outboundAttemptLimit.ResourceType, outboundAttemptLimit.OutboundAttemptLimitExporter())
	RegisterExporter(obCallResponseSet.ResourceType, obCallResponseSet.OutboundCallanalysisresponsesetExporter())
	RegisterExporter(obCallableTimeset.ResourceType, obCallableTimeset.OutboundCallableTimesetExporter())
	RegisterExporter(obCampaign.ResourceType, obCampaign.OutboundCampaignExporter())
	RegisterExporter(outboundContactList.ResourceType, outboundContactList.OutboundContactListExporter())
	RegisterExporter(outboundContactListTemplate.ResourceType, outboundContactListTemplate.OutboundContactListTemplateExporter())
	RegisterExporter(obContactListFilter.ResourceType, obContactListFilter.OutboundContactlistfilterExporter())
	RegisterExporter(obMessagingCampaign.ResourceType, obMessagingCampaign.OutboundMessagingcampaignExporter())
	RegisterExporter(obSequence.ResourceType, obSequence.OutboundSequenceExporter())
	RegisterExporter(obDncList.ResourceType, obDncList.OutboundDncListExporter())
	RegisterExporter(obCampaignRule.ResourceType, obCampaignRule.OutboundCampaignruleExporter())
	RegisterExporter(obSettings.ResourceType, obSettings.OutboundSettingsExporter())
	RegisterExporter(obDigitalRuleset.ResourceType, obDigitalRuleset.OutboundDigitalrulesetExporter())
	RegisterExporter(obfst.ResourceType, obfst.OutboundFileSpecificationTemplateExporter())
	RegisterExporter(obw.ResourceType, obw.OutboundWrapupCodeMappingsExporter())
	RegisterExporter(qualityFormsEvaluation.ResourceType, qualityFormsEvaluation.EvaluationFormExporter())
	RegisterExporter(recMediaRetPolicy.ResourceType, recMediaRetPolicy.MediaRetentionPolicyExporter())
	RegisterExporter(respmanagementLibrary.ResourceType, respmanagementLibrary.ResponsemanagementLibraryExporter())
	RegisterExporter(responsemanagementResponse.ResourceType, responsemanagementResponse.ResponsemanagementResponseExporter())
	RegisterExporter(respManagementRespAsset.ResourceType, respManagementRespAsset.ExporterResponseManagementResponseAsset())
	RegisterExporter(routingEmailDomain.ResourceType, routingEmailDomain.RoutingEmailDomainExporter())
	RegisterExporter(routingEmailRoute.ResourceType, routingEmailRoute.RoutingEmailRouteExporter())
	RegisterExporter(routinglanguage.ResourceType, routinglanguage.RoutingLanguageExporter())
	RegisterExporter(routingQueue.ResourceType, routingQueue.RoutingQueueExporter())
	RegisterExporter(routingQueueConditionalGroupRouting.ResourceType, routingQueueConditionalGroupRouting.RoutingQueueConditionalGroupRoutingExporter())
	RegisterExporter(routingQueueOutboundEmailAddress.ResourceType, routingQueueOutboundEmailAddress.OutboundRoutingQueueOutboundEmailAddressExporter())
	RegisterExporter(routingSettings.ResourceType, routingSettings.RoutingSettingsExporter())
	RegisterExporter(routingSkillGroup.ResourceType, routingSkillGroup.ResourceSkillGroupExporter())
	RegisterExporter(routingSkill.ResourceType, routingSkill.RoutingSkillExporter())
	RegisterExporter(routingSmsAddress.ResourceType, routingSmsAddress.RoutingSmsAddressExporter())
	RegisterExporter(routingUtilization.ResourceType, routingUtilization.RoutingUtilizationExporter())
	RegisterExporter(routingUtilizationLabel.ResourceType, routingUtilizationLabel.RoutingUtilizationLabelExporter())
	RegisterExporter(routingWrapupcode.ResourceType, routingWrapupcode.RoutingWrapupCodeExporter())
	RegisterExporter(edgeGroup.ResourceType, edgeGroup.EdgeGroupExporter())
	RegisterExporter(edgeExtension.ResourceType, edgeExtension.TelephonyExtensionPoolExporter())
	RegisterExporter(edgePhone.ResourceType, edgePhone.PhoneExporter())
	RegisterExporter(edgeSite.ResourceType, edgeSite.SiteExporter())
	RegisterExporter(outboundRoute.ResourceType, outboundRoute.SiteExporterOutboundRoute())
	RegisterExporter(phonebaseSettings.ResourceType, phonebaseSettings.PhoneBaseSettingsExporter())
	RegisterExporter(tbs.ResourceType, tbs.TrunkBaseSettingsExporter())
	RegisterExporter(edgesTrunk.ResourceType, edgesTrunk.TrunkExporter())
	RegisterExporter(user.ResourceType, user.UserExporter())
	RegisterExporter(userRoles.ResourceType, userRoles.UserRolesExporter())
	RegisterExporter(webdeployDeploy.ResourceType, webdeployDeploy.WebDeploymentExporter())
	RegisterExporter(webdeployConfig.ResourceType, webdeployConfig.WebDeploymentConfigurationExporter())
	RegisterExporter(pat.ResourceType, pat.ProcessAutomationTriggerExporter())
	RegisterExporter(obRuleset.ResourceType, obRuleset.OutboundRulesetExporter())
	RegisterExporter(didPool.ResourceType, didPool.TelephonyDidPoolExporter())
	RegisterExporter(workbin.ResourceType, workbin.TaskManagementWorkbinExporter())
	RegisterExporter(workitemSchema.ResourceType, workitemSchema.TaskManagementWorkitemSchemaExporter())
	RegisterExporter(worktype.ResourceType, worktype.TaskManagementWorktypeExporter())
	RegisterExporter(cMessagingSettings.ResourceType, cMessagingSettings.ConversationsMessagingSettingsExporter())
	RegisterExporter(worktypeStatus.ResourceType, worktypeStatus.TaskManagementWorktypeStatusExporter())
	RegisterExporter(supportedContent.ResourceType, supportedContent.SupportedContentExporter())
	RegisterExporter(defaultSupportedContent.ResourceType, defaultSupportedContent.ConversationsMessagingSupportedcontentDefaultExporter())
	RegisterExporter(cMessagingOpen.ResourceType, cMessagingOpen.ConversationsMessagingIntegrationsOpenExporter())
	RegisterExporter(scripts.ResourceType, scripts.ExporterScript())
	RegisterExporter(externalOrganization.ResourceType, externalOrganization.ExternalContactsOrganizationExporter())
	RegisterExporter(externalUser.ResourceType, externalUser.ExternalUserIdentityExporter())
	RegisterExporter(qualityFormsSurvey.ResourceType, qualityFormsSurvey.QualityFormsSurveyExporter())

	resourceExporter.SetRegisterExporter(resourceExporters)
}

func (r *registerTestInstance) registerTestDataSources() {
	providerDataSources["genesyscloud_auth_division_home"] = gcloud.DataSourceAuthDivisionHome()
	providerDataSources[scripts.ResourceType] = scripts.DataSourceScript()
	providerDataSources[edgeSite.ResourceType] = edgeSite.DataSourceSite()
	providerDataSources[cMessagingSettings.ResourceType] = cMessagingSettings.DataSourceConversationsMessagingSettings()
	providerDataSources[tbs.ResourceType] = tbs.DataSourceTrunkBaseSettings()
	providerDataSources[routingWrapupcode.ResourceType] = routingWrapupcode.DataSourceRoutingWrapupCode()
	providerDataSources[outboundRoute.ResourceType] = outboundRoute.DataSourceSiteOutboundRoute()
}

func RegisterExporter(exporterResourceType string, resourceExporter *resourceExporter.ResourceExporter) {
	resourceExporters[exporterResourceType] = resourceExporter
}

func TestMain(m *testing.M) {
	// Run setup function before starting the test suite for TfExport
	initTestResources()

	// Run the test suite
	m.Run()
}
