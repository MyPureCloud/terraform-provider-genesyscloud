package agentic_virtual_agent_version

import (
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	agenticVirtualAgent "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/agentic_virtual_agent"
	integration "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/integration"
	integrationAction "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/integration_action"
	knowledgeKnowledgebase "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/knowledge_knowledgebase"
)

/*
   Test initialization for the agentic_virtual_agent_version package.
   Registers both the version resource and the base agent resource (needed as a dependency).
*/

var providerDataSources map[string]*schema.Resource
var providerResources map[string]*schema.Resource

type registerTestInstance struct {
	resourceMapMutex   sync.RWMutex
	datasourceMapMutex sync.RWMutex
}

func (r *registerTestInstance) registerTestResources() {
	r.resourceMapMutex.Lock()
	defer r.resourceMapMutex.Unlock()

	providerResources[ResourceType] = ResourceAgenticVirtualAgentVersion()
	// Register the base agent resource as a dependency
	providerResources[agenticVirtualAgent.ResourceType] = agenticVirtualAgent.ResourceAgenticVirtualAgent()
	// Register the knowledge knowledgebase resource as a dependency for tool target tests
	providerResources[knowledgeKnowledgebase.ResourceType] = knowledgeKnowledgebase.ResourceKnowledgeKnowledgebase()
	// Register integration + integration_action as dependencies for DataAction tool tests
	providerResources[integration.ResourceType] = integration.ResourceIntegration()
	providerResources[integrationAction.ResourceType] = integrationAction.ResourceIntegrationAction()
}

func (r *registerTestInstance) registerTestDataSources() {
	r.datasourceMapMutex.Lock()
	defer r.datasourceMapMutex.Unlock()

	providerDataSources[agenticVirtualAgent.ResourceType] = agenticVirtualAgent.DataSourceAgenticVirtualAgent()
}

func initTestResources() {
	providerDataSources = make(map[string]*schema.Resource)
	providerResources = make(map[string]*schema.Resource)

	regInstance := &registerTestInstance{}

	regInstance.registerTestResources()
	regInstance.registerTestDataSources()
}

func TestMain(m *testing.M) {
	initTestResources()
	m.Run()
}
