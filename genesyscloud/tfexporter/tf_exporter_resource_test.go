package tfexporter

import (
	gcloud "terraform-provider-genesyscloud/genesyscloud"
	dt "terraform-provider-genesyscloud/genesyscloud/architect_datatable"
	"terraform-provider-genesyscloud/genesyscloud/architect_datatable_row"
	emergencyGroup "terraform-provider-genesyscloud/genesyscloud/architect_emergencygroup"
	flow "terraform-provider-genesyscloud/genesyscloud/architect_flow"
	flowLogLevel "terraform-provider-genesyscloud/genesyscloud/flow_loglevel"
	journeyActionTemplate "terraform-provider-genesyscloud/genesyscloud/journey_action_template"
	knowledge "terraform-provider-genesyscloud/genesyscloud/knowledge"
	knowledgeDocument "terraform-provider-genesyscloud/genesyscloud/knowledge_document"
	knowledgeLabel "terraform-provider-genesyscloud/genesyscloud/knowledge_label"
	routingWrapupcode "terraform-provider-genesyscloud/genesyscloud/routing_wrapupcode"
	outboundRoute "terraform-provider-genesyscloud/genesyscloud/telephony_providers_edges_site_outbound_route"

	grammar "terraform-provider-genesyscloud/genesyscloud/architect_grammar"
	grammarLanguage "terraform-provider-genesyscloud/genesyscloud/architect_grammar_language"
	archIvr "terraform-provider-genesyscloud/genesyscloud/architect_ivr"
	architectSchedulegroups "terraform-provider-genesyscloud/genesyscloud/architect_schedulegroups"
	architectSchedules "terraform-provider-genesyscloud/genesyscloud/architect_schedules"
	authDivision "terraform-provider-genesyscloud/genesyscloud/auth_division"
	authRole "terraform-provider-genesyscloud/genesyscloud/auth_role"
	integrationInstagram "terraform-provider-genesyscloud/genesyscloud/conversations_messaging_integrations_instagram"
	cMessagingOpen "terraform-provider-genesyscloud/genesyscloud/conversations_messaging_integrations_open"
	cMessagingSettings "terraform-provider-genesyscloud/genesyscloud/conversations_messaging_settings"
	supportedContent "terraform-provider-genesyscloud/genesyscloud/conversations_messaging_supportedcontent"
	defaultSupportedContent "terraform-provider-genesyscloud/genesyscloud/conversations_messaging_supportedcontent_default"
	employeeperformanceExternalmetricsDefinition "terraform-provider-genesyscloud/genesyscloud/employeeperformance_externalmetrics_definitions"
	flowMilestone "terraform-provider-genesyscloud/genesyscloud/flow_milestone"
	flowOutcome "terraform-provider-genesyscloud/genesyscloud/flow_outcome"
	"terraform-provider-genesyscloud/genesyscloud/group"
	groupRoles "terraform-provider-genesyscloud/genesyscloud/group_roles"
	idpAdfs "terraform-provider-genesyscloud/genesyscloud/idp_adfs"
	idpGeneric "terraform-provider-genesyscloud/genesyscloud/idp_generic"
	idpGsuite "terraform-provider-genesyscloud/genesyscloud/idp_gsuite"
	idpOkta "terraform-provider-genesyscloud/genesyscloud/idp_okta"
	idpOneLogin "terraform-provider-genesyscloud/genesyscloud/idp_onelogin"
	idpPing "terraform-provider-genesyscloud/genesyscloud/idp_ping"
	idpSalesforce "terraform-provider-genesyscloud/genesyscloud/idp_salesforce"
	integration "terraform-provider-genesyscloud/genesyscloud/integration"
	integrationAction "terraform-provider-genesyscloud/genesyscloud/integration_action"
	integrationCred "terraform-provider-genesyscloud/genesyscloud/integration_credential"
	integrationFacebook "terraform-provider-genesyscloud/genesyscloud/integration_facebook"
	journeyActionMap "terraform-provider-genesyscloud/genesyscloud/journey_action_map"
	"terraform-provider-genesyscloud/genesyscloud/oauth_client"
	oAuthSettings "terraform-provider-genesyscloud/genesyscloud/organization_authentication_settings"
	oAuthPairing "terraform-provider-genesyscloud/genesyscloud/orgauthorization_pairing"
	ob "terraform-provider-genesyscloud/genesyscloud/outbound"
	outboundAttemptLimit "terraform-provider-genesyscloud/genesyscloud/outbound_attempt_limit"
	obCallableTimeset "terraform-provider-genesyscloud/genesyscloud/outbound_callabletimeset"
	obCallResponseSet "terraform-provider-genesyscloud/genesyscloud/outbound_callanalysisresponseset"
	obCampaign "terraform-provider-genesyscloud/genesyscloud/outbound_campaign"
	obCampaignRule "terraform-provider-genesyscloud/genesyscloud/outbound_campaignrule"
	outboundContactList "terraform-provider-genesyscloud/genesyscloud/outbound_contact_list"
	outboundContactListContact "terraform-provider-genesyscloud/genesyscloud/outbound_contact_list_contact"
	outboundContactListTemplate "terraform-provider-genesyscloud/genesyscloud/outbound_contact_list_template"
	obContactListFilter "terraform-provider-genesyscloud/genesyscloud/outbound_contactlistfilter"
	obDigitalRuleset "terraform-provider-genesyscloud/genesyscloud/outbound_digitalruleset"
	obDncList "terraform-provider-genesyscloud/genesyscloud/outbound_dnclist"
	obfst "terraform-provider-genesyscloud/genesyscloud/outbound_filespecificationtemplate"
	obRuleset "terraform-provider-genesyscloud/genesyscloud/outbound_ruleset"
	obSequence "terraform-provider-genesyscloud/genesyscloud/outbound_sequence"
	obSettings "terraform-provider-genesyscloud/genesyscloud/outbound_settings"
	obw "terraform-provider-genesyscloud/genesyscloud/outbound_wrapupcode_mappings"
	pat "terraform-provider-genesyscloud/genesyscloud/process_automation_trigger"
	recMediaRetPolicy "terraform-provider-genesyscloud/genesyscloud/recording_media_retention_policy"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	respmanagementLibrary "terraform-provider-genesyscloud/genesyscloud/responsemanagement_library"
	responsemanagementResponse "terraform-provider-genesyscloud/genesyscloud/responsemanagement_response"
	respManagementRespAsset "terraform-provider-genesyscloud/genesyscloud/responsemanagement_responseasset"
	routingEmailDomain "terraform-provider-genesyscloud/genesyscloud/routing_email_domain"
	routingEmailRoute "terraform-provider-genesyscloud/genesyscloud/routing_email_route"
	routinglanguage "terraform-provider-genesyscloud/genesyscloud/routing_language"
	routingQueue "terraform-provider-genesyscloud/genesyscloud/routing_queue"
	routingQueueConditionalGroupRouting "terraform-provider-genesyscloud/genesyscloud/routing_queue_conditional_group_routing"
	routingQueueOutboundEmailAddress "terraform-provider-genesyscloud/genesyscloud/routing_queue_outbound_email_address"
	routingSettings "terraform-provider-genesyscloud/genesyscloud/routing_settings"
	routingSkill "terraform-provider-genesyscloud/genesyscloud/routing_skill"
	routingSkillGroup "terraform-provider-genesyscloud/genesyscloud/routing_skill_group"
	routingSmsAddress "terraform-provider-genesyscloud/genesyscloud/routing_sms_addresses"
	routingUtilization "terraform-provider-genesyscloud/genesyscloud/routing_utilization"
	routingUtilizationLabel "terraform-provider-genesyscloud/genesyscloud/routing_utilization_label"
	"terraform-provider-genesyscloud/genesyscloud/scripts"
	workbin "terraform-provider-genesyscloud/genesyscloud/task_management_workbin"
	workitemSchema "terraform-provider-genesyscloud/genesyscloud/task_management_workitem_schema"
	worktype "terraform-provider-genesyscloud/genesyscloud/task_management_worktype"
	worktypeStatus "terraform-provider-genesyscloud/genesyscloud/task_management_worktype_status"
	tbs "terraform-provider-genesyscloud/genesyscloud/telephony_provider_edges_trunkbasesettings"
	didPool "terraform-provider-genesyscloud/genesyscloud/telephony_providers_edges_did_pool"
	edgeGroup "terraform-provider-genesyscloud/genesyscloud/telephony_providers_edges_edge_group"
	edgeExtension "terraform-provider-genesyscloud/genesyscloud/telephony_providers_edges_extension_pool"
	phonebaseSettings "terraform-provider-genesyscloud/genesyscloud/telephony_providers_edges_phonebasesettings"
	edgesTrunk "terraform-provider-genesyscloud/genesyscloud/telephony_providers_edges_trunk"
	"terraform-provider-genesyscloud/genesyscloud/user"
	userRoles "terraform-provider-genesyscloud/genesyscloud/user_roles"
	webdeployConfig "terraform-provider-genesyscloud/genesyscloud/webdeployments_configuration"
	webdeployDeploy "terraform-provider-genesyscloud/genesyscloud/webdeployments_deployment"
	"testing"

	edgePhone "terraform-provider-genesyscloud/genesyscloud/telephony_providers_edges_phone"
	edgeSite "terraform-provider-genesyscloud/genesyscloud/telephony_providers_edges_site"

	userPrompt "terraform-provider-genesyscloud/genesyscloud/architect_user_prompt"
	externalOrganization "terraform-provider-genesyscloud/genesyscloud/external_contacts_organization"
	knowledgeCategory "terraform-provider-genesyscloud/genesyscloud/knowledge_category"
	location "terraform-provider-genesyscloud/genesyscloud/location"

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
	providerResources[journeyActionMap.ResourceType] = journeyActionMap.ResourceJourneyActionMap()
	providerResources[journeyActionTemplate.ResourceType] = journeyActionTemplate.ResourceJourneyActionTemplate()
	providerResources[knowledgeDocument.ResourceType] = knowledgeDocument.ResourceKnowledgeDocument()
	providerResources[knowledgeLabel.ResourceType] = knowledgeLabel.ResourceKnowledgeLabel()
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
	providerResources[ob.ResourceType] = ob.ResourceOutboundMessagingCampaign()
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

	providerResources["genesyscloud_quality_forms_survey"] = gcloud.ResourceSurveyForm()
	providerResources["genesyscloud_quality_forms_evaluation"] = gcloud.ResourceEvaluationForm()
	providerResources["genesyscloud_knowledge_document_variation"] = knowledge.ResourceKnowledgeDocumentVariation()
	providerResources["genesyscloud_widget_deployment"] = gcloud.ResourceWidgetDeployment()
	providerResources["genesyscloud_knowledge_label"] = knowledgeLabel.ResourceKnowledgeLabel()
	providerResources["genesyscloud_journey_outcome"] = gcloud.ResourceJourneyOutcome()
	providerResources["genesyscloud_journey_segment"] = gcloud.ResourceJourneySegment()
	providerResources["genesyscloud_knowledge_knowledgebase"] = gcloud.ResourceKnowledgeKnowledgebase()
	providerResources["genesyscloud_tf_export"] = ResourceTfExport()
}

func (r *registerTestInstance) registerTestExporters() {
	RegisterExporter("genesyscloud_organization_authentication_settings", oAuthSettings.OrganizationAuthenticationSettingsExporter())
	RegisterExporter("genesyscloud_architect_grammar", grammar.ArchitectGrammarExporter())
	RegisterExporter("genesyscloud_architect_grammar_language", grammarLanguage.ArchitectGrammarLanguageExporter())
	RegisterExporter("genesyscloud_architect_datatable", dt.ArchitectDatatableExporter())
	RegisterExporter("genesyscloud_architect_datatable_row", architect_datatable_row.ArchitectDatatableRowExporter())
	RegisterExporter("genesyscloud_architect_emergencygroup", emergencyGroup.ArchitectEmergencyGroupExporter())
	RegisterExporter("genesyscloud_architect_ivr", archIvr.ArchitectIvrExporter())
	RegisterExporter("genesyscloud_architect_schedules", architectSchedules.ArchitectSchedulesExporter())
	RegisterExporter("genesyscloud_architect_schedulegroups", architectSchedulegroups.ArchitectSchedulegroupsExporter())
	RegisterExporter("genesyscloud_architect_user_prompt", userPrompt.ArchitectUserPromptExporter())
	RegisterExporter("genesyscloud_auth_division", authDivision.AuthDivisionExporter())
	RegisterExporter("genesyscloud_auth_role", authRole.AuthRoleExporter())
	RegisterExporter("genesyscloud_employeeperformance_externalmetrics_definitions", employeeperformanceExternalmetricsDefinition.EmployeeperformanceExternalmetricsDefinitionExporter())
	RegisterExporter("genesyscloud_flow", flow.ArchitectFlowExporter())
	RegisterExporter("genesyscloud_flow_loglevel", flowLogLevel.FlowLogLevelExporter())
	RegisterExporter("genesyscloud_flow_milestone", flowMilestone.FlowMilestoneExporter())
	RegisterExporter("genesyscloud_flow_outcome", flowOutcome.FlowOutcomeExporter())
	RegisterExporter("genesyscloud_group", group.GroupExporter())
	RegisterExporter("genesyscloud_group_roles", groupRoles.GroupRolesExporter())
	RegisterExporter("genesyscloud_idp_adfs", idpAdfs.IdpAdfsExporter())
	RegisterExporter("genesyscloud_idp_generic", idpGeneric.IdpGenericExporter())
	RegisterExporter("genesyscloud_idp_gsuite", idpGsuite.IdpGsuiteExporter())
	RegisterExporter("genesyscloud_idp_okta", idpOkta.IdpOktaExporter())
	RegisterExporter("genesyscloud_idp_onelogin", idpOneLogin.IdpOneloginExporter())
	RegisterExporter("genesyscloud_idp_ping", idpPing.IdpPingExporter())
	RegisterExporter("genesyscloud_idp_salesforce", idpSalesforce.IdpSalesforceExporter())
	RegisterExporter("genesyscloud_integration", integration.IntegrationExporter())
	RegisterExporter("genesyscloud_integration_action", integrationAction.IntegrationActionExporter())
	RegisterExporter("genesyscloud_integration_credential", integrationCred.IntegrationCredentialExporter())
	RegisterExporter("genesyscloud_integration_facebook", integrationFacebook.IntegrationFacebookExporter())
	RegisterExporter("genesyscloud_conversations_messaging_integrations_instagram", integrationInstagram.ConversationsMessagingIntegrationsInstagramExporter())
	RegisterExporter("genesyscloud_journey_action_map", journeyActionMap.JourneyActionMapExporter())
	RegisterExporter("genesyscloud_journey_action_template", journeyActionTemplate.JourneyActionTemplateExporter())
	RegisterExporter("genesyscloud_journey_outcome", gcloud.JourneyOutcomeExporter())
	RegisterExporter("genesyscloud_journey_segment", gcloud.JourneySegmentExporter())
	RegisterExporter("genesyscloud_knowledge_knowledgebase", gcloud.KnowledgeKnowledgebaseExporter())
	RegisterExporter("genesyscloud_knowledge_document", knowledgeDocument.KnowledgeDocumentExporter())
	RegisterExporter(knowledgeCategory.ResourceType, knowledgeCategory.KnowledgeCategoryExporter())
	RegisterExporter("genesyscloud_location", location.LocationExporter())
	RegisterExporter("genesyscloud_oauth_client", oauth_client.OauthClientExporter())
	RegisterExporter("genesyscloud_outbound_attempt_limit", outboundAttemptLimit.OutboundAttemptLimitExporter())
	RegisterExporter("genesyscloud_outbound_callanalysisresponseset", obCallResponseSet.OutboundCallanalysisresponsesetExporter())
	RegisterExporter("genesyscloud_outbound_callabletimeset", obCallableTimeset.OutboundCallableTimesetExporter())
	RegisterExporter("genesyscloud_outbound_campaign", obCampaign.OutboundCampaignExporter())
	RegisterExporter("genesyscloud_outbound_contact_list", outboundContactList.OutboundContactListExporter())
	RegisterExporter("genesyscloud_outbound_contact_list_template", outboundContactListTemplate.OutboundContactListTemplateExporter())
	RegisterExporter("genesyscloud_outbound_contact_list_contact", outboundContactListContact.ContactExporter())
	RegisterExporter("genesyscloud_outbound_contactlistfilter", obContactListFilter.OutboundContactlistfilterExporter())
	RegisterExporter("genesyscloud_outbound_messagingcampaign", ob.OutboundMessagingcampaignExporter())
	RegisterExporter("genesyscloud_outbound_sequence", obSequence.OutboundSequenceExporter())
	RegisterExporter("genesyscloud_outbound_dnclist", obDncList.OutboundDncListExporter())
	RegisterExporter("genesyscloud_outbound_campaignrule", obCampaignRule.OutboundCampaignruleExporter())
	RegisterExporter("genesyscloud_outbound_settings", obSettings.OutboundSettingsExporter())
	RegisterExporter("genesyscloud_outbound_digitalruleset", obDigitalRuleset.OutboundDigitalrulesetExporter())
	RegisterExporter("genesyscloud_outbound_filespecificationtemplate", obfst.OutboundFileSpecificationTemplateExporter())
	RegisterExporter("genesyscloud_outbound_wrapupcodemappings", obw.OutboundWrapupCodeMappingsExporter())
	RegisterExporter("genesyscloud_quality_forms_evaluation", gcloud.EvaluationFormExporter())
	RegisterExporter("genesyscloud_quality_forms_survey", gcloud.SurveyFormExporter())
	RegisterExporter("genesyscloud_recording_media_retention_policy", recMediaRetPolicy.MediaRetentionPolicyExporter())
	RegisterExporter("genesyscloud_responsemanagement_library", respmanagementLibrary.ResponsemanagementLibraryExporter())
	RegisterExporter("genesyscloud_responsemanagement_response", responsemanagementResponse.ResponsemanagementResponseExporter())
	RegisterExporter("genesyscloud_responsemanagement_responseasset", respManagementRespAsset.ExporterResponseManagementResponseAsset())
	RegisterExporter("genesyscloud_routing_email_domain", routingEmailDomain.RoutingEmailDomainExporter())
	RegisterExporter("genesyscloud_routing_email_route", routingEmailRoute.RoutingEmailRouteExporter())
	RegisterExporter("genesyscloud_routing_language", routinglanguage.RoutingLanguageExporter())
	RegisterExporter("genesyscloud_routing_queue", routingQueue.RoutingQueueExporter())
	RegisterExporter("genesyscloud_routing_queue_conditional_group_routing", routingQueueConditionalGroupRouting.RoutingQueueConditionalGroupRoutingExporter())
	RegisterExporter("genesyscloud_routing_queue_outbound_email_address", routingQueueOutboundEmailAddress.OutboundRoutingQueueOutboundEmailAddressExporter())
	RegisterExporter("genesyscloud_routing_settings", routingSettings.RoutingSettingsExporter())
	RegisterExporter("genesyscloud_routing_skill_group", routingSkillGroup.ResourceSkillGroupExporter())
	RegisterExporter("genesyscloud_routing_skill", routingSkill.RoutingSkillExporter())
	RegisterExporter("genesyscloud_routing_sms_address", routingSmsAddress.RoutingSmsAddressExporter())
	RegisterExporter("genesyscloud_routing_utilization", routingUtilization.RoutingUtilizationExporter())
	RegisterExporter("genesyscloud_routing_utilization_label", routingUtilizationLabel.RoutingUtilizationLabelExporter())
	RegisterExporter("genesyscloud_routing_wrapupcode", routingWrapupcode.RoutingWrapupCodeExporter())
	RegisterExporter("genesyscloud_telephony_providers_edges_edge_group", edgeGroup.EdgeGroupExporter())
	RegisterExporter("genesyscloud_telephony_providers_edges_extension_pool", edgeExtension.TelephonyExtensionPoolExporter())
	RegisterExporter("genesyscloud_telephony_providers_edges_phone", edgePhone.PhoneExporter())
	RegisterExporter("genesyscloud_telephony_providers_edges_site", edgeSite.SiteExporter())
	RegisterExporter("genesyscloud_telephony_providers_edges_site_outbound_route", outboundRoute.SiteExporterOutboundRoute())
	RegisterExporter("genesyscloud_telephony_providers_edges_phonebasesettings", phonebaseSettings.PhoneBaseSettingsExporter())
	RegisterExporter("genesyscloud_telephony_providers_edges_trunkbasesettings", tbs.TrunkBaseSettingsExporter())
	RegisterExporter("genesyscloud_telephony_providers_edges_trunk", edgesTrunk.TrunkExporter())
	RegisterExporter("genesyscloud_user", user.UserExporter())
	RegisterExporter("genesyscloud_user_roles", userRoles.UserRolesExporter())
	RegisterExporter("genesyscloud_webdeployments_deployment", webdeployDeploy.WebDeploymentExporter())
	RegisterExporter("genesyscloud_webdeployments_configuration", webdeployConfig.WebDeploymentConfigurationExporter())
	RegisterExporter("genesyscloud_widget_deployment", gcloud.WidgetDeploymentExporter())

	RegisterExporter("genesyscloud_knowledge_document_variation", knowledge.KnowledgeDocumentVariationExporter())
	RegisterExporter(knowledgeLabel.ResourceType, knowledgeLabel.KnowledgeLabelExporter())

	RegisterExporter("genesyscloud_processautomation_trigger", pat.ProcessAutomationTriggerExporter())
	RegisterExporter("genesyscloud_outbound_ruleset", obRuleset.OutboundRulesetExporter())
	RegisterExporter("genesyscloud_telephony_providers_edges_did_pool", didPool.TelephonyDidPoolExporter())

	RegisterExporter("genesyscloud_task_management_workbin", workbin.TaskManagementWorkbinExporter())
	RegisterExporter("genesyscloud_task_management_workitem_schema", workbin.TaskManagementWorkbinExporter())
	RegisterExporter("genesyscloud_task_management_worktype", worktype.TaskManagementWorktypeExporter())
	RegisterExporter("genesyscloud_conversations_messaging_settings", cMessagingSettings.ConversationsMessagingSettingsExporter())
	RegisterExporter("genesyscloud_task_management_worktype_status", worktypeStatus.TaskManagementWorktypeStatusExporter())

	RegisterExporter("genesyscloud_conversations_messaging_supportedcontent", supportedContent.SupportedContentExporter())
	RegisterExporter("genesyscloud_conversations_messaging_supportedcontent_default", defaultSupportedContent.ConversationsMessagingSupportedcontentDefaultExporter())
	RegisterExporter("genesyscloud_conversations_messaging_integrations_open", cMessagingOpen.ConversationsMessagingIntegrationsOpenExporter())
	RegisterExporter("genesyscloud_script", scripts.ExporterScript())
	RegisterExporter("genesyscloud_externalcontacts_organization", externalOrganization.ExternalContactsOrganizationExporter())
	resourceExporter.SetRegisterExporter(resourceExporters)
}

func (r *registerTestInstance) registerTestDataSources() {
	providerDataSources["genesyscloud_auth_division_home"] = gcloud.DataSourceAuthDivisionHome()
	providerDataSources[scripts.ResourceType] = scripts.DataSourceScript()
	providerDataSources[edgeSite.ResourceType] = edgeSite.DataSourceSite()
	providerDataSources[cMessagingSettings.ResourceType] = cMessagingSettings.DataSourceConversationsMessagingSettings()
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
