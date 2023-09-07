package genesyscloud

import (
	"log"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v105/platformclientv2"
)

var (
	sdkConfig           *platformclientv2.Configuration
	providerDataSources map[string]*schema.Resource
	providerResources   map[string]*schema.Resource
	err                 error
)

type registerTestInstance struct {
	resourceMapMutex   sync.RWMutex
	datasourceMapMutex sync.RWMutex
}

func (r *registerTestInstance) registerTestResources() {

	r.resourceMapMutex.Lock()
	defer r.resourceMapMutex.Unlock()

	providerResources["genesyscloud_telephony_providers_edges_site"] = ResourceSite()
	providerResources["genesyscloud_routing_wrapupcode"] = ResourceRoutingWrapupCode()
	providerResources["genesyscloud_routing_queue"] = ResourceRoutingQueue()
	providerResources["genesyscloud_flow"] = ResourceFlow()
	providerResources["genesyscloud_location"] = ResourceLocation()
	providerResources["genesyscloud_architect_datatable"] = ResourceArchitectDatatable()
	providerResources["genesyscloud_architect_datatable_row"] = ResourceArchitectDatatableRow()
	providerResources["genesyscloud_architect_emergencygroup"] = ResourceArchitectEmergencyGroup()
	providerResources["genesyscloud_flow"] = ResourceFlow()
	providerResources["genesyscloud_flow_milestone"] = ResourceFlowMilestone()
	providerResources["genesyscloud_flow_outcome"] = ResourceFlowOutcome()
	providerResources["genesyscloud_architect_ivr"] = ResourceArchitectIvrConfig()
	providerResources["genesyscloud_architect_schedules"] = ResourceArchitectSchedules()
	providerResources["genesyscloud_architect_schedulegroups"] = ResourceArchitectScheduleGroups()
	providerResources["genesyscloud_architect_user_prompt"] = ResourceArchitectUserPrompt()
	providerResources["genesyscloud_auth_role"] = ResourceAuthRole()
	providerResources["genesyscloud_auth_division"] = ResourceAuthDivision()
	providerResources["genesyscloud_employeeperformance_externalmetrics_definitions"] = ResourceEmployeeperformanceExternalmetricsDefinition()
	providerResources["genesyscloud_group"] = ResourceGroup()
	providerResources["genesyscloud_group_roles"] = ResourceGroupRoles()
	providerResources["genesyscloud_idp_adfs"] = ResourceIdpAdfs()

	providerResources["genesyscloud_routing_wrapupcode"] = ResourceRoutingWrapupCode()
	providerResources["genesyscloud_idp_generic"] = ResourceIdpGeneric()
	providerResources["genesyscloud_idp_gsuite"] = ResourceIdpGsuite()
	providerResources["genesyscloud_idp_okta"] = ResourceIdpOkta()
	providerResources["genesyscloud_idp_onelogin"] = ResourceIdpOnelogin()
	providerResources["genesyscloud_idp_ping"] = ResourceIdpPing()
	providerResources["genesyscloud_idp_salesforce"] = ResourceIdpSalesforce()
	providerResources["genesyscloud_integration"] = ResourceIntegration()
	providerResources["genesyscloud_integration_action"] = ResourceIntegrationAction()
	providerResources["genesyscloud_integration_credential"] = ResourceCredential()
	providerResources["genesyscloud_journey_action_map"] = ResourceJourneyActionMap()
	providerResources["genesyscloud_journey_action_template"] = ResourceJourneyActionTemplate()
	providerResources["genesyscloud_journey_outcome"] = ResourceJourneyOutcome()
	providerResources["genesyscloud_journey_segment"] = ResourceJourneySegment()
	providerResources["genesyscloud_knowledge_knowledgebase"] = ResourceKnowledgeKnowledgebase()
	providerResources["genesyscloud_knowledge_document"] = ResourceKnowledgeDocument()
	providerResources["genesyscloud_knowledge_v1_document"] = ResourceKnowledgeDocumentV1()
	providerResources["genesyscloud_knowledge_document_variation"] = ResourceKnowledgeDocumentVariation()
	providerResources["genesyscloud_knowledge_category"] = ResourceKnowledgeCategory()
	providerResources["genesyscloud_knowledge_v1_category"] = ResourceKnowledgeCategoryV1()
	providerResources["genesyscloud_knowledge_label"] = ResourceKnowledgeLabel()
	providerResources["genesyscloud_location"] = ResourceLocation()
	providerResources["genesyscloud_recording_media_retention_policy"] = ResourceMediaRetentionPolicy()
	providerResources["genesyscloud_oauth_client"] = ResourceOAuthClient()

	providerResources["genesyscloud_orgauthorization_pairing"] = resourceOrgauthorizationPairing()
	providerResources["genesyscloud_quality_forms_evaluation"] = ResourceEvaluationForm()
	providerResources["genesyscloud_quality_forms_survey"] = ResourceSurveyForm()
	providerResources["genesyscloud_responsemanagement_library"] = ResourceResponsemanagementLibrary()
	providerResources["genesyscloud_responsemanagement_response"] = ResourceResponsemanagementResponse()
	providerResources["genesyscloud_responsemanagement_responseasset"] = resourceResponseManagamentResponseAsset()
	providerResources["genesyscloud_routing_email_domain"] = ResourceRoutingEmailDomain()
	providerResources["genesyscloud_routing_email_route"] = ResourceRoutingEmailRoute()
	providerResources["genesyscloud_routing_language"] = ResourceRoutingLanguage()
	providerResources["genesyscloud_routing_queue"] = ResourceRoutingQueue()
	providerResources["genesyscloud_routing_skill"] = ResourceRoutingSkill()
	providerResources["genesyscloud_routing_skill_group"] = ResourceRoutingSkillGroup()
	providerResources["genesyscloud_routing_settings"] = ResourceRoutingSettings()
	providerResources["genesyscloud_routing_utilization"] = ResourceRoutingUtilization()
	providerResources["genesyscloud_routing_wrapupcode"] = ResourceRoutingWrapupCode()
	providerResources["genesyscloud_telephony_providers_edges_did_pool"] = ResourceTelephonyDidPool()
	providerResources["genesyscloud_telephony_providers_edges_edge_group"] = ResourceEdgeGroup()
	providerResources["genesyscloud_telephony_providers_edges_extension_pool"] = ResourceTelephonyExtensionPool()
	providerResources["genesyscloud_telephony_providers_edges_phone"] = ResourcePhone()
	providerResources["genesyscloud_telephony_providers_edges_site"] = ResourceSite()
	providerResources["genesyscloud_telephony_providers_edges_phonebasesettings"] = ResourcePhoneBaseSettings()
	providerResources["genesyscloud_telephony_providers_edges_trunkbasesettings"] = ResourceTrunkBaseSettings()
	providerResources["genesyscloud_telephony_providers_edges_trunk"] = ResourceTrunk()
	providerResources["genesyscloud_user"] = ResourceUser()
	providerResources["genesyscloud_user_roles"] = ResourceUserRoles()
	providerResources["genesyscloud_webdeployments_configuration"] = ResourceWebDeploymentConfiguration()
	providerResources["genesyscloud_webdeployments_deployment"] = ResourceWebDeployment()
	providerResources["genesyscloud_widget_deployment"] = ResourceWidgetDeployment()

}

func (r *registerTestInstance) registerTestDataSources() {

	r.datasourceMapMutex.Lock()
	defer r.datasourceMapMutex.Unlock()

	providerDataSources["genesyscloud_telephony_providers_edges_site"] = DataSourceSite()
	providerDataSources["genesyscloud_routing_wrapupcode"] = DataSourceRoutingWrapupcode()
	providerDataSources["genesyscloud_routing_queue"] = DataSourceRoutingQueue()
	providerDataSources["genesyscloud_flow"] = DataSourceFlow()
	providerDataSources["genesyscloud_location"] = DataSourceLocation()
	providerDataSources["genesyscloud_auth_division_home"] = DataSourceAuthDivisionHome()

	providerDataSources["genesyscloud_architect_datatable"] = DataSourceArchitectDatatable()
	providerDataSources["genesyscloud_architect_ivr"] = DataSourceArchitectIvr()
	providerDataSources["genesyscloud_architect_emergencygroup"] = DataSourceArchitectEmergencyGroup()
	providerDataSources["genesyscloud_architect_schedules"] = DataSourceSchedule()
	providerDataSources["genesyscloud_architect_schedulegroups"] = DataSourceArchitectScheduleGroups()
	providerDataSources["genesyscloud_architect_user_prompt"] = dataSourceUserPrompt()
	providerDataSources["genesyscloud_auth_role"] = dataSourceAuthRole()
	providerDataSources["genesyscloud_auth_division"] = dataSourceAuthDivision()
	providerDataSources["genesyscloud_auth_division_home"] = DataSourceAuthDivisionHome()
	providerDataSources["genesyscloud_employeeperformance_externalmetrics_definitions"] = dataSourceEmployeeperformanceExternalmetricsDefinition()
	providerDataSources["genesyscloud_flow"] = DataSourceFlow()
	providerDataSources["genesyscloud_flow_milestone"] = dataSourceFlowMilestone()
	providerDataSources["genesyscloud_flow_outcome"] = dataSourceFlowOutcome()
	providerDataSources["genesyscloud_group"] = dataSourceGroup()
	providerDataSources["genesyscloud_integration"] = dataSourceIntegration()
	providerDataSources["genesyscloud_integration_action"] = dataSourceIntegrationAction()
	providerDataSources["genesyscloud_integration_credential"] = dataSourceIntegrationCredential()
	providerDataSources["genesyscloud_journey_action_map"] = dataSourceJourneyActionMap()
	providerDataSources["genesyscloud_journey_action_template"] = dataSourceJourneyActionTemplate()
	providerDataSources["genesyscloud_journey_outcome"] = dataSourceJourneyOutcome()
	providerDataSources["genesyscloud_journey_segment"] = dataSourceJourneySegment()
	providerDataSources["genesyscloud_knowledge_knowledgebase"] = dataSourceKnowledgeKnowledgebase()
	providerDataSources["genesyscloud_knowledge_category"] = dataSourceKnowledgeCategory()
	providerDataSources["genesyscloud_knowledge_label"] = dataSourceKnowledgeLabel()
	providerDataSources["genesyscloud_location"] = DataSourceLocation()
	providerDataSources["genesyscloud_oauth_client"] = dataSourceOAuthClient()
	providerDataSources["genesyscloud_organizations_me"] = dataSourceOrganizationsMe()

	providerDataSources["genesyscloud_quality_forms_evaluation"] = dataSourceQualityFormsEvaluations()
	providerDataSources["genesyscloud_quality_forms_survey"] = dataSourceQualityFormsSurvey()
	providerDataSources["genesyscloud_recording_media_retention_policy"] = dataSourceRecordingMediaRetentionPolicy()
	providerDataSources["genesyscloud_responsemanagement_library"] = dataSourceResponsemanagementLibrary()
	providerDataSources["genesyscloud_responsemanagement_response"] = dataSourceResponsemanagementResponse()
	providerDataSources["genesyscloud_responsemanagement_responseasset"] = dataSourceResponseManagamentResponseAsset()
	providerDataSources["genesyscloud_routing_language"] = dataSourceRoutingLanguage()
	providerDataSources["genesyscloud_routing_queue"] = DataSourceRoutingQueue()
	providerDataSources["genesyscloud_routing_settings"] = dataSourceRoutingSettings()
	providerDataSources["genesyscloud_routing_skill"] = dataSourceRoutingSkill()
	providerDataSources["genesyscloud_routing_skill_group"] = dataSourceRoutingSkillGroup()
	providerDataSources["genesyscloud_routing_email_domain"] = dataSourceRoutingEmailDomain()
	providerDataSources["genesyscloud_routing_wrapupcode"] = DataSourceRoutingWrapupcode()

	providerDataSources["genesyscloud_station"] = dataSourceStation()
	providerDataSources["genesyscloud_user"] = dataSourceUser()
	providerDataSources["genesyscloud_telephony_providers_edges_did"] = dataSourceDid()
	providerDataSources["genesyscloud_telephony_providers_edges_did_pool"] = dataSourceDidPool()
	providerDataSources["genesyscloud_telephony_providers_edges_edge_group"] = dataSourceEdgeGroup()
	providerDataSources["genesyscloud_telephony_providers_edges_extension_pool"] = dataSourceExtensionPool()
	providerDataSources["genesyscloud_telephony_providers_edges_site"] = DataSourceSite()
	providerDataSources["genesyscloud_telephony_providers_edges_linebasesettings"] = dataSourceLineBaseSettings()
	providerDataSources["genesyscloud_telephony_providers_edges_phone"] = dataSourcePhone()
	providerDataSources["genesyscloud_telephony_providers_edges_phonebasesettings"] = dataSourcePhoneBaseSettings()
	providerDataSources["genesyscloud_telephony_providers_edges_trunk"] = dataSourceTrunk()
	providerDataSources["genesyscloud_telephony_providers_edges_trunkbasesettings"] = dataSourceTrunkBaseSettings()
	providerDataSources["genesyscloud_webdeployments_configuration"] = dataSourceWebDeploymentsConfiguration()
	providerDataSources["genesyscloud_webdeployments_deployment"] = dataSourceWebDeploymentsDeployment()
	providerDataSources["genesyscloud_widget_deployment"] = dataSourceWidgetDeployments()

}

func init_test_resources() {

	if sdkConfig, err = AuthorizeSdk(); err != nil {
		log.Fatal(err)
	}

	providerDataSources = make(map[string]*schema.Resource)
	providerResources = make(map[string]*schema.Resource)

	reg_instance := &registerTestInstance{}

	reg_instance.registerTestDataSources()
	reg_instance.registerTestResources()

}

func TestMain(m *testing.M) {
	// Run setup function before starting the test suite for resources in GenesysCloud Parent Package.
	init_test_resources()

	// Run the test suite
	m.Run()

}
