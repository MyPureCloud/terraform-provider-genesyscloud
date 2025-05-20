package genesyscloud

import (
	"log"
	"sync"
	"testing"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/architect_flow"
	archScheduleGroup "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/architect_schedulegroups"
	architectSchedules "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/architect_schedules"
	authDivision "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/auth_division"
	cMessagingSettings "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/conversations_messaging_settings"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/group"
	journeyOutcome "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/journey_outcome"
	journeySegment "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/journey_segment"
	knowledgeKnowledgebase "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/knowledge_knowledgebase"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/location"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	routingEmailDomain "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/routing_email_domain"
	routinglanguage "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/routing_language"
	routingQueue "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/routing_queue"
	routingSettings "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/routing_settings"
	routingSkill "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/routing_skill"
	routingSkillGroup "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/routing_skill_group"
	routingUtilization "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/routing_utilization"
	routingUtilizationLabel "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/routing_utilization_label"
	routingWrapupCode "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/routing_wrapupcode"
	extensionPool "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/telephony_providers_edges_extension_pool"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/user"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v157/platformclientv2"
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
	providerResources[journeySegment.ResourceType] = journeySegment.ResourceJourneySegment()
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
	providerResources[journeyOutcome.ResourceType] = journeyOutcome.ResourceJourneyOutcome()
	providerResources[knowledgeKnowledgebase.ResourceType] = knowledgeKnowledgebase.ResourceKnowledgeKnowledgebase()
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
	providerDataSources["genesyscloud_organizations_me"] = DataSourceOrganizationsMe()
	providerDataSources["genesyscloud_auth_division_home"] = DataSourceAuthDivisionHome()
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
