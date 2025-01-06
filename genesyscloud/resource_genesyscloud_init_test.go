package genesyscloud

import (
	"log"
	"sync"
	"terraform-provider-genesyscloud/genesyscloud/architect_flow"
	archScheduleGroup "terraform-provider-genesyscloud/genesyscloud/architect_schedulegroups"
	architectSchedules "terraform-provider-genesyscloud/genesyscloud/architect_schedules"
	authDivision "terraform-provider-genesyscloud/genesyscloud/auth_division"
	cMessagingSettings "terraform-provider-genesyscloud/genesyscloud/conversations_messaging_settings"
	"terraform-provider-genesyscloud/genesyscloud/group"
	"terraform-provider-genesyscloud/genesyscloud/location"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	routingEmailDomain "terraform-provider-genesyscloud/genesyscloud/routing_email_domain"
	routinglanguage "terraform-provider-genesyscloud/genesyscloud/routing_language"
	routingQueue "terraform-provider-genesyscloud/genesyscloud/routing_queue"
	routingSettings "terraform-provider-genesyscloud/genesyscloud/routing_settings"
	routingSkill "terraform-provider-genesyscloud/genesyscloud/routing_skill"
	routingSkillGroup "terraform-provider-genesyscloud/genesyscloud/routing_skill_group"
	routingUtilization "terraform-provider-genesyscloud/genesyscloud/routing_utilization"
	routingUtilizationLabel "terraform-provider-genesyscloud/genesyscloud/routing_utilization_label"
	routingWrapupCode "terraform-provider-genesyscloud/genesyscloud/routing_wrapupcode"
	extensionPool "terraform-provider-genesyscloud/genesyscloud/telephony_providers_edges_extension_pool"
	"terraform-provider-genesyscloud/genesyscloud/user"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v150/platformclientv2"
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

	providerResources[architect_flow.ResourceType] = architect_flow.ResourceArchitectFlow()
	providerResources[group.ResourceType] = group.ResourceGroup()
	providerResources[routingQueue.ResourceType] = routingQueue.ResourceRoutingQueue()
	providerResources[location.ResourceType] = location.ResourceLocation()
	providerResources[authDivision.ResourceType] = authDivision.ResourceAuthDivision()
	providerResources["genesyscloud_journey_outcome"] = ResourceJourneyOutcome()
	providerResources["genesyscloud_journey_segment"] = ResourceJourneySegment()
	providerResources["genesyscloud_knowledge_knowledgebase"] = ResourceKnowledgeKnowledgebase()
	providerResources["genesyscloud_quality_forms_evaluation"] = ResourceEvaluationForm()
	providerResources["genesyscloud_quality_forms_survey"] = ResourceSurveyForm()
	providerResources["genesyscloud_widget_deployment"] = ResourceWidgetDeployment()
	providerResources[user.ResourceType] = user.ResourceUser()
	providerResources[routinglanguage.ResourceType] = routinglanguage.ResourceRoutingLanguage()
	providerResources[routingEmailDomain.ResourceType] = routingEmailDomain.ResourceRoutingEmailDomain()
	providerResources[routingSkillGroup.ResourceType] = routingSkillGroup.ResourceRoutingSkillGroup()
	providerResources[routingSkill.ResourceType] = routingSkill.ResourceRoutingSkill()
	providerResources[routingSettings.ResourceType] = routingSettings.ResourceRoutingSettings()
	providerResources[routingUtilization.ResourceType] = routingUtilization.ResourceRoutingUtilization()
	providerResources[routingWrapupCode.ResourceType] = routingWrapupCode.ResourceRoutingWrapupCode()
	providerResources[archScheduleGroup.ResourceType] = archScheduleGroup.ResourceArchitectSchedulegroups()
	providerResources[architectSchedules.ResourceType] = architectSchedules.ResourceArchitectSchedules()
	providerResources[routingUtilizationLabel.ResourceType] = routingUtilizationLabel.ResourceRoutingUtilizationLabel()
	providerResources[cMessagingSettings.ResourceType] = cMessagingSettings.ResourceConversationsMessagingSettings()
	providerResources[extensionPool.ResourceType] = extensionPool.ResourceTelephonyExtensionPool()
}

func (r *registerTestInstance) registerTestDataSources() {

	r.datasourceMapMutex.Lock()
	defer r.datasourceMapMutex.Unlock()

	providerDataSources[architect_flow.ResourceType] = architect_flow.DataSourceArchitectFlow()
	providerDataSources[group.ResourceType] = group.DataSourceGroup()
	providerDataSources[routingQueue.ResourceType] = routingQueue.DataSourceRoutingQueue()
	providerDataSources[location.ResourceType] = location.DataSourceLocation()
	providerDataSources[authDivision.ResourceType] = authDivision.DataSourceAuthDivision()
	providerDataSources[user.ResourceType] = user.DataSourceUser()
	providerDataSources[routingSkill.ResourceType] = routingSkill.DataSourceRoutingSkill()
	providerDataSources[routingEmailDomain.ResourceType] = routingEmailDomain.DataSourceRoutingEmailDomain()
	providerDataSources[routingSkillGroup.ResourceType] = routingSkillGroup.DataSourceRoutingSkillGroup()
	providerDataSources[routingWrapupCode.ResourceType] = routingWrapupCode.DataSourceRoutingWrapupCode()
	providerDataSources[routingUtilizationLabel.ResourceType] = routingUtilizationLabel.DataSourceRoutingUtilizationLabel()
	providerDataSources[cMessagingSettings.ResourceType] = cMessagingSettings.DataSourceConversationsMessagingSettings()
	providerDataSources["genesyscloud_journey_outcome"] = dataSourceJourneyOutcome()
	providerDataSources["genesyscloud_journey_segment"] = dataSourceJourneySegment()
	providerDataSources["genesyscloud_knowledge_knowledgebase"] = dataSourceKnowledgeKnowledgebase()
	providerDataSources["genesyscloud_organizations_me"] = DataSourceOrganizationsMe()
	providerDataSources["genesyscloud_quality_forms_evaluation"] = DataSourceQualityFormsEvaluations()
	providerDataSources["genesyscloud_quality_forms_survey"] = dataSourceQualityFormsSurvey()
	providerDataSources["genesyscloud_auth_division_home"] = DataSourceAuthDivisionHome()
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
