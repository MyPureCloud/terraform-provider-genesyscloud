package tfexporter

import (
	gcloud "terraform-provider-genesyscloud/genesyscloud"
	dt "terraform-provider-genesyscloud/genesyscloud/architect_datatable"
	"terraform-provider-genesyscloud/genesyscloud/architect_datatable_row"
	emergencyGroup "terraform-provider-genesyscloud/genesyscloud/architect_emergencygroup"
	flow "terraform-provider-genesyscloud/genesyscloud/architect_flow"
	flowLogLevel "terraform-provider-genesyscloud/genesyscloud/flow_loglevel"
	journeyActionTemplate "terraform-provider-genesyscloud/genesyscloud/journey_action_template"
	journeyOutcome "terraform-provider-genesyscloud/genesyscloud/journey_outcome"
	journeySegment "terraform-provider-genesyscloud/genesyscloud/journey_segment"
	journeyViewSchedule "terraform-provider-genesyscloud/genesyscloud/journey_view_schedule"
	knowledgeDocument "terraform-provider-genesyscloud/genesyscloud/knowledge_document"
	knowledgeDocumentVariation "terraform-provider-genesyscloud/genesyscloud/knowledge_document_variation"
	knowledgeKnowledgebase "terraform-provider-genesyscloud/genesyscloud/knowledge_knowledgebase"
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
	obMessagingCampaign "terraform-provider-genesyscloud/genesyscloud/outbound_messagingcampaign"
	obRuleset "terraform-provider-genesyscloud/genesyscloud/outbound_ruleset"
	obSequence "terraform-provider-genesyscloud/genesyscloud/outbound_sequence"
	obSettings "terraform-provider-genesyscloud/genesyscloud/outbound_settings"
	obw "terraform-provider-genesyscloud/genesyscloud/outbound_wrapupcode_mappings"
	pat "terraform-provider-genesyscloud/genesyscloud/process_automation_trigger"
	qualityFormsEvaluation "terraform-provider-genesyscloud/genesyscloud/quality_forms_evaluation"
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
	didPool "terraform-provider-genesyscloud/genesyscloud/telephony_providers_edges_did_pool"
	edgeGroup "terraform-provider-genesyscloud/genesyscloud/telephony_providers_edges_edge_group"
	edgeExtension "terraform-provider-genesyscloud/genesyscloud/telephony_providers_edges_extension_pool"
	phonebaseSettings "terraform-provider-genesyscloud/genesyscloud/telephony_providers_edges_phonebasesettings"
	edgesTrunk "terraform-provider-genesyscloud/genesyscloud/telephony_providers_edges_trunk"
	tbs "terraform-provider-genesyscloud/genesyscloud/telephony_providers_edges_trunkbasesettings"
	"terraform-provider-genesyscloud/genesyscloud/user"
	userRoles "terraform-provider-genesyscloud/genesyscloud/user_roles"
	webdeployConfig "terraform-provider-genesyscloud/genesyscloud/webdeployments_configuration"
	webdeployDeploy "terraform-provider-genesyscloud/genesyscloud/webdeployments_deployment"
	"testing"

	edgePhone "terraform-provider-genesyscloud/genesyscloud/telephony_providers_edges_phone"
	edgeSite "terraform-provider-genesyscloud/genesyscloud/telephony_providers_edges_site"

	userPrompt "terraform-provider-genesyscloud/genesyscloud/architect_user_prompt"
	externalOrganization "terraform-provider-genesyscloud/genesyscloud/external_contacts_organization"
	externalUser "terraform-provider-genesyscloud/genesyscloud/external_user"
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
	providerResources["genesyscloud_quality_forms_survey"] = gcloud.ResourceSurveyForm()
	providerResources["genesyscloud_widget_deployment"] = gcloud.ResourceWidgetDeployment()
	providerResources["genesyscloud_knowledge_label"] = knowledgeLabel.ResourceKnowledgeLabel()
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
	RegisterExporter("genesyscloud_widget_deployment", gcloud.WidgetDeploymentExporter())
	RegisterExporter("genesyscloud_quality_forms_survey", gcloud.SurveyFormExporter())
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
