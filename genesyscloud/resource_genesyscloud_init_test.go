package genesyscloud

import (
	"log"
	"sync"
	"terraform-provider-genesyscloud/genesyscloud/architect_flow"
	archScheduleGroup "terraform-provider-genesyscloud/genesyscloud/architect_schedulegroups"
	architectSchedules "terraform-provider-genesyscloud/genesyscloud/architect_schedules"
	cMessagingSettings "terraform-provider-genesyscloud/genesyscloud/conversations_messaging_settings"
	"terraform-provider-genesyscloud/genesyscloud/group"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	routingEmailDomain "terraform-provider-genesyscloud/genesyscloud/routing_email_domain"
	routinglanguage "terraform-provider-genesyscloud/genesyscloud/routing_language"
	routingQueue "terraform-provider-genesyscloud/genesyscloud/routing_queue"
	routingSettings "terraform-provider-genesyscloud/genesyscloud/routing_settings"
	routingUtilization "terraform-provider-genesyscloud/genesyscloud/routing_utilization"
	routingUtilizationLabel "terraform-provider-genesyscloud/genesyscloud/routing_utilization_label"
	routingSkill "terraform-provider-genesyscloud/genesyscloud/routing_skill"
	extensionPool "terraform-provider-genesyscloud/genesyscloud/telephony_providers_edges_extension_pool"
	routingSkillGroup "terraform-provider-genesyscloud/genesyscloud/routing_skill_group"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

var (
	sdkConfig           *platformclientv2.Configuration
	providerDataSources map[string]*schema.Resource
	providerResources   map[string]*schema.Resource
	err                 error
	mu                  sync.Mutex
)

type registerTestInstance struct {
	resourceMapMutex   sync.RWMutex
	datasourceMapMutex sync.RWMutex
}

func (r *registerTestInstance) registerTestResources() {

	r.resourceMapMutex.Lock()
	defer r.resourceMapMutex.Unlock()

	providerResources["genesyscloud_flow"] = architect_flow.ResourceArchitectFlow()
	providerResources["genesyscloud_group"] = group.ResourceGroup()
	providerResources["genesyscloud_routing_queue"] = routingQueue.ResourceRoutingQueue()
	providerResources["genesyscloud_location"] = ResourceLocation()
	providerResources["genesyscloud_auth_division"] = ResourceAuthDivision()
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
	providerResources["genesyscloud_quality_forms_evaluation"] = ResourceEvaluationForm()
	providerResources["genesyscloud_quality_forms_survey"] = ResourceSurveyForm()
	providerResources["genesyscloud_routing_language"] = routinglanguage.ResourceRoutingLanguage()
	providerResources["genesyscloud_routing_email_domain"] = routingEmailDomain.ResourceRoutingEmailDomain()
	providerResources["genesyscloud_routing_skill_group"] = routingSkillGroup.ResourceRoutingSkillGroup()
	providerResources["genesyscloud_routing_skill"] = routingSkill.ResourceRoutingSkill()
	providerResources["genesyscloud_routing_settings"] = routingSettings.ResourceRoutingSettings()
	providerResources["genesyscloud_routing_utilization"] = routingUtilization.ResourceRoutingUtilization()
	providerResources["genesyscloud_routing_wrapupcode"] = ResourceRoutingWrapupCode()
	providerResources["genesyscloud_user"] = ResourceUser()
	providerResources["genesyscloud_widget_deployment"] = ResourceWidgetDeployment()
	providerResources["genesyscloud_architect_schedulegroups"] = archScheduleGroup.ResourceArchitectSchedulegroups()
	providerResources["genesyscloud_architect_schedules"] = architectSchedules.ResourceArchitectSchedules()
	providerResources["genesyscloud_routing_utilization_label"] = routingUtilizationLabel.ResourceRoutingUtilizationLabel()
	providerResources["genesyscloud_conversations_messaging_settings"] = cMessagingSettings.ResourceConversationsMessagingSettings()
	providerResources["genesyscloud_telephony_providers_edges_extension_pool"] = extensionPool.ResourceTelephonyExtensionPool()
}

func (r *registerTestInstance) registerTestDataSources() {

	r.datasourceMapMutex.Lock()
	defer r.datasourceMapMutex.Unlock()

	providerDataSources["genesyscloud_flow"] = architect_flow.DataSourceArchitectFlow()
	providerDataSources["genesyscloud_group"] = group.DataSourceGroup()
	providerDataSources["genesyscloud_routing_wrapupcode"] = DataSourceRoutingWrapupcode()
	providerDataSources["genesyscloud_routing_queue"] = routingQueue.DataSourceRoutingQueue()
	providerDataSources["genesyscloud_location"] = DataSourceLocation()
	providerDataSources["genesyscloud_auth_division_home"] = DataSourceAuthDivisionHome()
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
	providerDataSources["genesyscloud_routing_language"] = routinglanguage.DataSourceRoutingLanguage()
	providerDataSources["genesyscloud_routing_skill"] = routingSkill.DataSourceRoutingSkill()
	providerDataSources["genesyscloud_routing_email_domain"] = routingEmailDomain.DataSourceRoutingEmailDomain()
	providerDataSources["genesyscloud_routing_skill_group"] = routingSkillGroup.DataSourceRoutingSkillGroup()
	providerDataSources["genesyscloud_routing_wrapupcode"] = DataSourceRoutingWrapupcode()
	providerDataSources["genesyscloud_user"] = DataSourceUser()
	providerDataSources["genesyscloud_widget_deployment"] = dataSourceWidgetDeployments()
	providerDataSources["genesyscloud_routing_utilization_label"] = routingUtilizationLabel.DataSourceRoutingUtilizationLabel()
	providerDataSources["genesyscloud_conversations_messaging_settings"] = cMessagingSettings.DataSourceConversationsMessagingSettings()

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
