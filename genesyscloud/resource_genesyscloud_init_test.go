package genesyscloud

import (
	"log"
	"sync"
	"terraform-provider-genesyscloud/genesyscloud/architect_flow"
	archScheduleGroup "terraform-provider-genesyscloud/genesyscloud/architect_schedulegroups"
	"terraform-provider-genesyscloud/genesyscloud/group"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v125/platformclientv2"
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

	providerResources["genesyscloud_flow"] = architect_flow.ResourceArchitectFlow()
	providerResources["genesyscloud_routing_queue"] = ResourceRoutingQueue()
	providerResources["genesyscloud_group"] = group.ResourceGroup()
	providerResources["genesyscloud_location"] = ResourceLocation()
	providerResources["genesyscloud_architect_schedules"] = ResourceArchitectSchedules()
	providerResources["genesyscloud_architect_user_prompt"] = ResourceArchitectUserPrompt()
	providerResources["genesyscloud_auth_division"] = ResourceAuthDivision()
	providerResources["genesyscloud_idp_adfs"] = ResourceIdpAdfs()
	providerResources["genesyscloud_idp_generic"] = ResourceIdpGeneric()
	providerResources["genesyscloud_idp_gsuite"] = ResourceIdpGsuite()
	providerResources["genesyscloud_idp_okta"] = ResourceIdpOkta()
	providerResources["genesyscloud_idp_onelogin"] = ResourceIdpOnelogin()
	providerResources["genesyscloud_idp_ping"] = ResourceIdpPing()
	providerResources["genesyscloud_idp_salesforce"] = ResourceIdpSalesforce()
	providerResources["genesyscloud_journey_action_map"] = ResourceJourneyActionMap()
	providerResources["genesyscloud_journey_action_template"] = ResourceJourneyActionTemplate()
	providerResources["genesyscloud_journey_outcome"] = ResourceJourneyOutcome()
	providerResources["genesyscloud_journey_segment"] = ResourceJourneySegment()
	providerResources["genesyscloud_knowledge_knowledgebase"] = ResourceKnowledgeKnowledgebase()
	providerResources["genesyscloud_knowledge_document"] = ResourceKnowledgeDocument()
	providerResources["genesyscloud_knowledge_document_variation"] = ResourceKnowledgeDocumentVariation()
	providerResources["genesyscloud_knowledge_category"] = ResourceKnowledgeCategory()
	providerResources["genesyscloud_knowledge_label"] = ResourceKnowledgeLabel()
	providerResources["genesyscloud_location"] = ResourceLocation()
	providerResources["genesyscloud_orgauthorization_pairing"] = resourceOrgauthorizationPairing()
	providerResources["genesyscloud_quality_forms_evaluation"] = ResourceEvaluationForm()
	providerResources["genesyscloud_quality_forms_survey"] = ResourceSurveyForm()
	providerResources["genesyscloud_routing_email_domain"] = ResourceRoutingEmailDomain()
	providerResources["genesyscloud_routing_language"] = ResourceRoutingLanguage()
	providerResources["genesyscloud_routing_queue"] = ResourceRoutingQueue()
	providerResources["genesyscloud_routing_skill"] = ResourceRoutingSkill()
	providerResources["genesyscloud_routing_skill_group"] = ResourceRoutingSkillGroup()
	providerResources["genesyscloud_routing_settings"] = ResourceRoutingSettings()
	providerResources["genesyscloud_routing_utilization"] = ResourceRoutingUtilization()
	providerResources["genesyscloud_routing_utilization_label"] = ResourceRoutingUtilizationLabel()
	providerResources["genesyscloud_routing_wrapupcode"] = ResourceRoutingWrapupCode()
	providerResources["genesyscloud_user"] = ResourceUser()
	providerResources["genesyscloud_widget_deployment"] = ResourceWidgetDeployment()
	providerResources["genesyscloud_architect_schedulegroups"] = archScheduleGroup.ResourceArchitectSchedulegroups()
}

func (r *registerTestInstance) registerTestDataSources() {

	r.datasourceMapMutex.Lock()
	defer r.datasourceMapMutex.Unlock()

	providerDataSources["genesyscloud_flow"] = architect_flow.DataSourceArchitectFlow()
	providerDataSources["genesyscloud_group"] = group.DataSourceGroup()
	providerDataSources["genesyscloud_routing_wrapupcode"] = DataSourceRoutingWrapupcode()
	providerDataSources["genesyscloud_routing_queue"] = DataSourceRoutingQueue()
	providerDataSources["genesyscloud_location"] = DataSourceLocation()
	providerDataSources["genesyscloud_auth_division_home"] = DataSourceAuthDivisionHome()
	providerDataSources["genesyscloud_architect_schedules"] = DataSourceSchedule()
	providerDataSources["genesyscloud_architect_user_prompt"] = dataSourceUserPrompt()
	providerDataSources["genesyscloud_auth_division"] = dataSourceAuthDivision()
	providerDataSources["genesyscloud_auth_division_home"] = DataSourceAuthDivisionHome()
	providerDataSources["genesyscloud_journey_action_map"] = dataSourceJourneyActionMap()
	providerDataSources["genesyscloud_journey_action_template"] = dataSourceJourneyActionTemplate()
	providerDataSources["genesyscloud_journey_outcome"] = dataSourceJourneyOutcome()
	providerDataSources["genesyscloud_journey_segment"] = dataSourceJourneySegment()
	providerDataSources["genesyscloud_knowledge_knowledgebase"] = dataSourceKnowledgeKnowledgebase()
	providerDataSources["genesyscloud_knowledge_category"] = dataSourceKnowledgeCategory()
	providerDataSources["genesyscloud_knowledge_label"] = dataSourceKnowledgeLabel()
	providerDataSources["genesyscloud_location"] = DataSourceLocation()
	providerDataSources["genesyscloud_organizations_me"] = DataSourceOrganizationsMe()
	providerDataSources["genesyscloud_quality_forms_evaluation"] = DataSourceQualityFormsEvaluations()
	providerDataSources["genesyscloud_quality_forms_survey"] = dataSourceQualityFormsSurvey()
	providerDataSources["genesyscloud_routing_language"] = dataSourceRoutingLanguage()
	providerDataSources["genesyscloud_routing_queue"] = DataSourceRoutingQueue()
	providerDataSources["genesyscloud_routing_settings"] = dataSourceRoutingSettings()
	providerDataSources["genesyscloud_routing_skill"] = dataSourceRoutingSkill()
	providerDataSources["genesyscloud_routing_skill_group"] = dataSourceRoutingSkillGroup()
	providerDataSources["genesyscloud_routing_email_domain"] = DataSourceRoutingEmailDomain()
	providerDataSources["genesyscloud_routing_utilization_label"] = dataSourceRoutingUtilizationLabel()
	providerDataSources["genesyscloud_routing_wrapupcode"] = DataSourceRoutingWrapupcode()
	providerDataSources["genesyscloud_user"] = DataSourceUser()
	providerDataSources["genesyscloud_widget_deployment"] = dataSourceWidgetDeployments()

}

func initTestResources() {
	if sdkConfig, err = provider.AuthorizeSdk(); err != nil {
		log.Fatal(err)
	}
	providerDataSources = make(map[string]*schema.Resource)
	providerResources = make(map[string]*schema.Resource)
	regInstance := &registerTestInstance{}
	regInstance.registerTestDataSources()
	regInstance.registerTestResources()
}

func TestMain(m *testing.M) {
	// Run setup function before starting the test suite for resources in GenesysCloud Parent Package.
	initTestResources()
	// Run the test suite
	m.Run()
}
