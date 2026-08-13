package agentic_virtual_agent_version_publish

import (
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	agenticVirtualAgent "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/agentic_virtual_agent"
	agenticVirtualAgentVersion "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/agentic_virtual_agent_version"
)

/*
   Test initialization for the publish package.
   Registers the publish resource plus its dependencies (agent + version).
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

	providerResources[ResourceType] = ResourceAgenticVirtualAgentVersionPublish()
	providerResources[agenticVirtualAgent.ResourceType] = agenticVirtualAgent.ResourceAgenticVirtualAgent()
	providerResources[agenticVirtualAgentVersion.ResourceType] = agenticVirtualAgentVersion.ResourceAgenticVirtualAgentVersion()
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
