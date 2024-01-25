package tfexporter

import (
	gcloud "terraform-provider-genesyscloud/genesyscloud"
	dt "terraform-provider-genesyscloud/genesyscloud/architect_datatable"
	"terraform-provider-genesyscloud/genesyscloud/architect_datatable_row"
	emergencyGroup "terraform-provider-genesyscloud/genesyscloud/architect_emergencygroup"
	grammar "terraform-provider-genesyscloud/genesyscloud/architect_grammar"
	grammarLanguage "terraform-provider-genesyscloud/genesyscloud/architect_grammar_language"
	archIvr "terraform-provider-genesyscloud/genesyscloud/architect_ivr"
	flowMilestone "terraform-provider-genesyscloud/genesyscloud/flow_milestone"
	flowOutcome "terraform-provider-genesyscloud/genesyscloud/flow_outcome"
	integration "terraform-provider-genesyscloud/genesyscloud/integration"
	integrationAction "terraform-provider-genesyscloud/genesyscloud/integration_action"
	integrationCred "terraform-provider-genesyscloud/genesyscloud/integration_credential"
	ob "terraform-provider-genesyscloud/genesyscloud/outbound"
	outboundAttemptLimit "terraform-provider-genesyscloud/genesyscloud/outbound_attempt_limit"
	obCampaign "terraform-provider-genesyscloud/genesyscloud/outbound_campaign"
	obCampaignRule "terraform-provider-genesyscloud/genesyscloud/outbound_campaignrule"
	outboundContactList "terraform-provider-genesyscloud/genesyscloud/outbound_contact_list"
	obRuleset "terraform-provider-genesyscloud/genesyscloud/outbound_ruleset"
	obSequence "terraform-provider-genesyscloud/genesyscloud/outbound_sequence"
	obw "terraform-provider-genesyscloud/genesyscloud/outbound_wrapupcode_mappings"
	pat "terraform-provider-genesyscloud/genesyscloud/process_automation_trigger"
	recMediaRetPolicy "terraform-provider-genesyscloud/genesyscloud/recording_media_retention_policy"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	routingSmsAddress "terraform-provider-genesyscloud/genesyscloud/routing_sms_addresses"
	workbin "terraform-provider-genesyscloud/genesyscloud/task_management_workbin"
	workitemSchema "terraform-provider-genesyscloud/genesyscloud/task_management_workitem_schema"
	worktype "terraform-provider-genesyscloud/genesyscloud/task_management_worktype"
	telephony "terraform-provider-genesyscloud/genesyscloud/telephony"
	didPool "terraform-provider-genesyscloud/genesyscloud/telephony_providers_edges_did_pool"
	edgeGroup "terraform-provider-genesyscloud/genesyscloud/telephony_providers_edges_edge_group"
	edgeExtension "terraform-provider-genesyscloud/genesyscloud/telephony_providers_edges_extension_pool"
	phonebaseSettings "terraform-provider-genesyscloud/genesyscloud/telephony_providers_edges_phonebasesettings"
	edgesTrunk "terraform-provider-genesyscloud/genesyscloud/telephony_providers_edges_trunk"
	webdeployConfig "terraform-provider-genesyscloud/genesyscloud/webdeployments_configuration"
	webdeployDeploy "terraform-provider-genesyscloud/genesyscloud/webdeployments_deployment"

	"testing"

	edgePhone "terraform-provider-genesyscloud/genesyscloud/telephony_providers_edges_phone"
	edgeSite "terraform-provider-genesyscloud/genesyscloud/telephony_providers_edges_site"

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
	providerResources["genesyscloud_architect_grammar"] = grammar.ResourceArchitectGrammar()
	providerResources["genesyscloud_architect_grammar_language"] = grammarLanguage.ResourceArchitectGrammarLanguage()
	providerResources["genesyscloud_architect_datatable"] = dt.ResourceArchitectDatatable()
	providerResources["genesyscloud_architect_datatable_row"] = architect_datatable_row.ResourceArchitectDatatableRow()
	providerResources["genesyscloud_architect_emergencygroup"] = emergencyGroup.ResourceArchitectEmergencyGroup()
	providerResources["genesyscloud_flow"] = gcloud.ResourceFlow()
	providerResources["genesyscloud_flow_milestone"] = flowMilestone.ResourceFlowMilestone()
	providerResources["genesyscloud_flow_outcome"] = flowOutcome.ResourceFlowOutcome()
	providerResources["genesyscloud_architect_ivr"] = archIvr.ResourceArchitectIvrConfig()
	providerResources["genesyscloud_architect_schedules"] = gcloud.ResourceArchitectSchedules()
	providerResources["genesyscloud_architect_schedulegroups"] = gcloud.ResourceArchitectScheduleGroups()
	providerResources["genesyscloud_architect_user_prompt"] = gcloud.ResourceArchitectUserPrompt()
	providerResources["genesyscloud_auth_role"] = gcloud.ResourceAuthRole()
	providerResources["genesyscloud_auth_division"] = gcloud.ResourceAuthDivision()
	providerResources["genesyscloud_employeeperformance_externalmetrics_definitions"] = gcloud.ResourceEmployeeperformanceExternalmetricsDefinition()
	providerResources["genesyscloud_group"] = gcloud.ResourceGroup()
	providerResources["genesyscloud_group_roles"] = gcloud.ResourceGroupRoles()
	providerResources["genesyscloud_idp_adfs"] = gcloud.ResourceIdpAdfs()
	providerResources["genesyscloud_idp_generic"] = gcloud.ResourceIdpGeneric()
	providerResources["genesyscloud_idp_gsuite"] = gcloud.ResourceIdpGsuite()
	providerResources["genesyscloud_idp_okta"] = gcloud.ResourceIdpOkta()
	providerResources["genesyscloud_idp_onelogin"] = gcloud.ResourceIdpOnelogin()
	providerResources["genesyscloud_idp_ping"] = gcloud.ResourceIdpPing()
	providerResources["genesyscloud_idp_salesforce"] = gcloud.ResourceIdpSalesforce()
	providerResources["genesyscloud_integration"] = integration.ResourceIntegration()
	providerResources["genesyscloud_integration_action"] = integrationAction.ResourceIntegrationAction()
	providerResources["genesyscloud_integration_credential"] = integrationCred.ResourceIntegrationCredential()
	providerResources["genesyscloud_journey_action_map"] = gcloud.ResourceJourneyActionMap()
	providerResources["genesyscloud_journey_action_template"] = gcloud.ResourceJourneyActionTemplate()
	providerResources["genesyscloud_journey_outcome"] = gcloud.ResourceJourneyOutcome()
	providerResources["genesyscloud_journey_segment"] = gcloud.ResourceJourneySegment()
	providerResources["genesyscloud_knowledge_knowledgebase"] = gcloud.ResourceKnowledgeKnowledgebase()
	providerResources["genesyscloud_knowledge_document"] = gcloud.ResourceKnowledgeDocument()
	providerResources["genesyscloud_knowledge_document_variation"] = gcloud.ResourceKnowledgeDocumentVariation()
	providerResources["genesyscloud_knowledge_category"] = gcloud.ResourceKnowledgeCategory()
	providerResources["genesyscloud_knowledge_label"] = gcloud.ResourceKnowledgeLabel()
	providerResources["genesyscloud_location"] = gcloud.ResourceLocation()
	providerResources["genesyscloud_recording_media_retention_policy"] = recMediaRetPolicy.ResourceMediaRetentionPolicy()
	providerResources["genesyscloud_oauth_client"] = gcloud.ResourceOAuthClient()
	providerResources["genesyscloud_outbound_settings"] = ob.ResourceOutboundSettings()
	providerResources["genesyscloud_quality_forms_evaluation"] = gcloud.ResourceEvaluationForm()
	providerResources["genesyscloud_user"] = gcloud.ResourceUser()
	providerResources["genesyscloud_responsemanagement_library"] = gcloud.ResourceResponsemanagementLibrary()
	providerResources["genesyscloud_routing_email_domain"] = gcloud.ResourceRoutingEmailDomain()
	providerResources["genesyscloud_routing_email_route"] = gcloud.ResourceRoutingEmailRoute()
	providerResources["genesyscloud_routing_language"] = gcloud.ResourceRoutingLanguage()
	providerResources["genesyscloud_routing_queue"] = gcloud.ResourceRoutingQueue()
	providerResources["genesyscloud_routing_skill"] = gcloud.ResourceRoutingSkill()
	providerResources["genesyscloud_routing_settings"] = gcloud.ResourceRoutingSettings()
	providerResources["genesyscloud_routing_utilization"] = gcloud.ResourceRoutingUtilization()

	providerResources["genesyscloud_routing_wrapupcode"] = gcloud.ResourceRoutingWrapupCode()
	providerResources["genesyscloud_telephony_providers_edges_extension_pool"] = edgeExtension.ResourceTelephonyExtensionPool()
	providerResources["genesyscloud_telephony_providers_edges_phone"] = edgePhone.ResourcePhone()
	providerResources["genesyscloud_telephony_providers_edges_site"] = edgeSite.ResourceSite()
	providerResources["genesyscloud_telephony_providers_edges_phonebasesettings"] = phonebaseSettings.ResourcePhoneBaseSettings()
	providerResources["genesyscloud_telephony_providers_edges_trunkbasesettings"] = telephony.ResourceTrunkBaseSettings()
	providerResources["genesyscloud_telephony_providers_edges_trunk"] = edgesTrunk.ResourceTrunk()
	providerResources["genesyscloud_telephony_providers_edges_edge_group"] = edgeGroup.ResourceEdgeGroup()

	providerResources["genesyscloud_user_roles"] = gcloud.ResourceUserRoles()
	providerResources["genesyscloud_webdeployments_deployment"] = webdeployDeploy.ResourceWebDeployment()
	providerResources["genesyscloud_webdeployments_configuration"] = webdeployConfig.ResourceWebDeploymentConfiguration()
	providerResources["genesyscloud_widget_deployment"] = gcloud.ResourceWidgetDeployment()
	providerResources["genesyscloud_processautomation_trigger"] = pat.ResourceProcessAutomationTrigger()

	providerResources["genesyscloud_outbound_attempt_limit"] = outboundAttemptLimit.ResourceOutboundAttemptLimit()
	providerResources["genesyscloud_outbound_callanalysisresponseset"] = ob.ResourceOutboundCallAnalysisResponseSet()
	providerResources["genesyscloud_outbound_callabletimeset"] = ob.ResourceOutboundCallabletimeset()
	providerResources["genesyscloud_outbound_campaign"] = obCampaign.ResourceOutboundCampaign()
	providerResources["genesyscloud_outbound_contact_list"] = outboundContactList.ResourceOutboundContactList()
	providerResources["genesyscloud_outbound_contactlistfilter"] = ob.ResourceOutboundContactListFilter()
	providerResources["genesyscloud_outbound_messagingcampaign"] = ob.ResourceOutboundMessagingCampaign()
	providerResources["genesyscloud_outbound_sequence"] = obSequence.ResourceOutboundSequence()
	providerResources["genesyscloud_outbound_dnclist"] = ob.ResourceOutboundDncList()
	providerResources["genesyscloud_outbound_campaignrule"] = obCampaignRule.ResourceOutboundCampaignrule()
	providerResources["genesyscloud_outbound_wrapupcodemappings"] = obw.ResourceOutboundWrapUpCodeMappings()
	providerResources["genesyscloud_quality_forms_survey"] = gcloud.ResourceSurveyForm()
	providerResources["genesyscloud_responsemanagement_response"] = gcloud.ResourceResponsemanagementResponse()
	providerResources["genesyscloud_routing_sms_address"] = routingSmsAddress.ResourceRoutingSmsAddress()
	providerResources["genesyscloud_routing_skill_group"] = gcloud.ResourceRoutingSkillGroup()
	providerResources["genesyscloud_telephony_providers_edges_did_pool"] = didPool.ResourceTelephonyDidPool()

	providerResources["genesyscloud_task_management_workbin"] = workbin.ResourceTaskManagementWorkbin()
	providerResources["genesyscloud_task_management_workitem_schema"] = workitemSchema.ResourceTaskManagementWorkitemSchema()
	providerResources["genesyscloud_task_management_worktype"] = worktype.ResourceTaskManagementWorktype()

	providerResources["genesyscloud_tf_export"] = ResourceTfExport()
}

func (r *registerTestInstance) registerTestExporters() {
	RegisterExporter("genesyscloud_architect_grammar", grammar.ArchitectGrammarExporter())
	RegisterExporter("genesyscloud_architect_grammar_language", grammarLanguage.ArchitectGrammarLanguageExporter())
	RegisterExporter("genesyscloud_architect_datatable", dt.ArchitectDatatableExporter())
	RegisterExporter("genesyscloud_architect_datatable_row", architect_datatable_row.ArchitectDatatableRowExporter())
	RegisterExporter("genesyscloud_architect_emergencygroup", emergencyGroup.ArchitectEmergencyGroupExporter())
	RegisterExporter("genesyscloud_architect_ivr", archIvr.ArchitectIvrExporter())
	RegisterExporter("genesyscloud_architect_schedules", gcloud.ArchitectSchedulesExporter())
	RegisterExporter("genesyscloud_architect_schedulegroups", gcloud.ArchitectScheduleGroupsExporter())
	RegisterExporter("genesyscloud_architect_user_prompt", gcloud.ArchitectUserPromptExporter())
	RegisterExporter("genesyscloud_auth_division", gcloud.AuthDivisionExporter())
	RegisterExporter("genesyscloud_auth_role", gcloud.AuthRoleExporter())
	RegisterExporter("genesyscloud_employeeperformance_externalmetrics_definitions", gcloud.EmployeeperformanceExternalmetricsDefinitionExporter())
	RegisterExporter("genesyscloud_flow", gcloud.FlowExporter())
	RegisterExporter("genesyscloud_flow_milestone", flowMilestone.FlowMilestoneExporter())
	RegisterExporter("genesyscloud_flow_outcome", flowOutcome.FlowOutcomeExporter())
	RegisterExporter("genesyscloud_group", gcloud.GroupExporter())
	RegisterExporter("genesyscloud_group_roles", gcloud.GroupRolesExporter())
	RegisterExporter("genesyscloud_idp_adfs", gcloud.IdpAdfsExporter())
	RegisterExporter("genesyscloud_idp_generic", gcloud.IdpGenericExporter())
	RegisterExporter("genesyscloud_idp_gsuite", gcloud.IdpGsuiteExporter())
	RegisterExporter("genesyscloud_idp_okta", gcloud.IdpOktaExporter())
	RegisterExporter("genesyscloud_idp_onelogin", gcloud.IdpOneloginExporter())
	RegisterExporter("genesyscloud_idp_ping", gcloud.IdpPingExporter())
	RegisterExporter("genesyscloud_idp_salesforce", gcloud.IdpSalesforceExporter())
	RegisterExporter("genesyscloud_integration", integration.IntegrationExporter())
	RegisterExporter("genesyscloud_integration_action", integrationAction.IntegrationActionExporter())
	RegisterExporter("genesyscloud_integration_credential", integrationCred.IntegrationCredentialExporter())
	RegisterExporter("genesyscloud_journey_action_map", gcloud.JourneyActionMapExporter())
	RegisterExporter("genesyscloud_journey_action_template", gcloud.JourneyActionTemplateExporter())
	RegisterExporter("genesyscloud_journey_outcome", gcloud.JourneyOutcomeExporter())
	RegisterExporter("genesyscloud_journey_segment", gcloud.JourneySegmentExporter())
	RegisterExporter("genesyscloud_knowledge_knowledgebase", gcloud.KnowledgeKnowledgebaseExporter())
	RegisterExporter("genesyscloud_knowledge_document", gcloud.KnowledgeDocumentExporter())
	RegisterExporter("genesyscloud_knowledge_category", gcloud.KnowledgeCategoryExporter())
	RegisterExporter("genesyscloud_location", gcloud.LocationExporter())
	RegisterExporter("genesyscloud_oauth_client", gcloud.OauthClientExporter())
	RegisterExporter("genesyscloud_outbound_attempt_limit", outboundAttemptLimit.OutboundAttemptLimitExporter())
	RegisterExporter("genesyscloud_outbound_callanalysisresponseset", ob.OutboundCallAnalysisResponseSetExporter())
	RegisterExporter("genesyscloud_outbound_callabletimeset", ob.OutboundCallableTimesetExporter())
	RegisterExporter("genesyscloud_outbound_campaign", obCampaign.OutboundCampaignExporter())
	RegisterExporter("genesyscloud_outbound_contact_list", outboundContactList.OutboundContactListExporter())
	RegisterExporter("genesyscloud_outbound_contactlistfilter", ob.OutboundContactListFilterExporter())
	RegisterExporter("genesyscloud_outbound_messagingcampaign", ob.OutboundMessagingcampaignExporter())
	RegisterExporter("genesyscloud_outbound_sequence", obSequence.OutboundSequenceExporter())
	RegisterExporter("genesyscloud_outbound_dnclist", ob.OutboundDncListExporter())
	RegisterExporter("genesyscloud_outbound_campaignrule", obCampaignRule.OutboundCampaignruleExporter())
	RegisterExporter("genesyscloud_outbound_settings", ob.OutboundSettingsExporter())
	RegisterExporter("genesyscloud_outbound_wrapupcodemappings", obw.OutboundWrapupCodeMappingsExporter())
	RegisterExporter("genesyscloud_quality_forms_evaluation", gcloud.EvaluationFormExporter())
	RegisterExporter("genesyscloud_quality_forms_survey", gcloud.SurveyFormExporter())
	RegisterExporter("genesyscloud_recording_media_retention_policy", recMediaRetPolicy.MediaRetentionPolicyExporter())
	RegisterExporter("genesyscloud_responsemanagement_library", gcloud.ResponsemanagementLibraryExporter())
	RegisterExporter("genesyscloud_responsemanagement_response", gcloud.ResponsemanagementResponseExporter())
	RegisterExporter("genesyscloud_routing_email_domain", gcloud.RoutingEmailDomainExporter())
	RegisterExporter("genesyscloud_routing_email_route", gcloud.RoutingEmailRouteExporter())
	RegisterExporter("genesyscloud_routing_language", gcloud.RoutingLanguageExporter())
	RegisterExporter("genesyscloud_routing_queue", gcloud.RoutingQueueExporter())
	RegisterExporter("genesyscloud_routing_settings", gcloud.RoutingSettingsExporter())
	RegisterExporter("genesyscloud_routing_skill", gcloud.RoutingSkillExporter())
	RegisterExporter("genesyscloud_routing_skill_group", gcloud.ResourceSkillGroupExporter())
	RegisterExporter("genesyscloud_routing_sms_address", routingSmsAddress.RoutingSmsAddressExporter())
	RegisterExporter("genesyscloud_routing_utilization", gcloud.RoutingUtilizationExporter())
	RegisterExporter("genesyscloud_routing_wrapupcode", gcloud.RoutingWrapupCodeExporter())
	RegisterExporter("genesyscloud_telephony_providers_edges_edge_group", edgeGroup.EdgeGroupExporter())
	RegisterExporter("genesyscloud_telephony_providers_edges_extension_pool", edgeExtension.TelephonyExtensionPoolExporter())
	RegisterExporter("genesyscloud_telephony_providers_edges_phone", edgePhone.PhoneExporter())
	RegisterExporter("genesyscloud_telephony_providers_edges_site", edgeSite.SiteExporter())
	RegisterExporter("genesyscloud_telephony_providers_edges_phonebasesettings", phonebaseSettings.PhoneBaseSettingsExporter())
	RegisterExporter("genesyscloud_telephony_providers_edges_trunkbasesettings", telephony.TrunkBaseSettingsExporter())
	RegisterExporter("genesyscloud_telephony_providers_edges_trunk", edgesTrunk.TrunkExporter())
	RegisterExporter("genesyscloud_user", gcloud.UserExporter())
	RegisterExporter("genesyscloud_user_roles", gcloud.UserRolesExporter())
	RegisterExporter("genesyscloud_webdeployments_deployment", webdeployDeploy.WebDeploymentExporter())
	RegisterExporter("genesyscloud_webdeployments_configuration", webdeployConfig.WebDeploymentConfigurationExporter())
	RegisterExporter("genesyscloud_widget_deployment", gcloud.WidgetDeploymentExporter())

	RegisterExporter("genesyscloud_knowledge_document_variation", gcloud.KnowledgeDocumentVariationExporter())
	RegisterExporter("genesyscloud_knowledge_label", gcloud.KnowledgeLabelExporter())

	RegisterExporter("genesyscloud_processautomation_trigger", pat.ProcessAutomationTriggerExporter())
	RegisterExporter("genesyscloud_outbound_ruleset", obRuleset.OutboundRulesetExporter())
	RegisterExporter("genesyscloud_telephony_providers_edges_did_pool", didPool.TelephonyDidPoolExporter())

	RegisterExporter("genesyscloud_task_management_workbin", workbin.TaskManagementWorkbinExporter())
	RegisterExporter("genesyscloud_task_management_workitem_schema", workbin.TaskManagementWorkbinExporter())
	RegisterExporter("genesyscloud_task_management_worktype", worktype.TaskManagementWorktypeExporter())

	resourceExporter.SetRegisterExporter(resourceExporters)
}

func (r *registerTestInstance) registerTestDataSources() {
	providerDataSources["genesyscloud_auth_division_home"] = gcloud.DataSourceAuthDivisionHome()
}

func RegisterExporter(exporterName string, resourceExporter *resourceExporter.ResourceExporter) {
	resourceExporters[exporterName] = resourceExporter
}

func TestMain(m *testing.M) {
	// Run setup function before starting the test suite for TfExport
	initTestResources()

	// Run the test suite
	m.Run()

}
