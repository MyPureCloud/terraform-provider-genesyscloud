package telephony_providers_edges_trunk

import (
	"sync"
	"terraform-provider-genesyscloud/genesyscloud/location"
	tbs "terraform-provider-genesyscloud/genesyscloud/telephony_provider_edges_trunkbasesettings"
	edgeGroup "terraform-provider-genesyscloud/genesyscloud/telephony_providers_edges_edge_group"
	edgeSite "terraform-provider-genesyscloud/genesyscloud/telephony_providers_edges_site"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var providerDataSources map[string]*schema.Resource
var providerResources map[string]*schema.Resource

type registerTestInstance struct {
	resourceMapMutex   sync.RWMutex
	datasourceMapMutex sync.RWMutex
}

func (r *registerTestInstance) registerTestResources() {

	r.resourceMapMutex.Lock()
	defer r.resourceMapMutex.Unlock()

	providerResources[tbs.ResourceType] = tbs.ResourceTrunkBaseSettings()
	providerResources[ResourceType] = ResourceTrunk()

	// external package dependencies
	providerResources[edgeSite.ResourceType] = edgeSite.ResourceSite()

	providerResources[location.ResourceType] = location.ResourceLocation()
	providerResources[edgeGroup.ResourceType] = edgeGroup.ResourceEdgeGroup()

}

func (r *registerTestInstance) registerTestDataSources() {

	r.datasourceMapMutex.Lock()
	defer r.datasourceMapMutex.Unlock()

	providerDataSources[tbs.ResourceType] = tbs.DataSourceTrunkBaseSettings()
	providerDataSources[ResourceType] = DataSourceTrunk()
	// external package dependencies
	providerDataSources[edgeSite.ResourceType] = edgeSite.DataSourceSite()
	providerDataSources[edgeGroup.ResourceType] = edgeGroup.DataSourceEdgeGroup()

}

func initTestResources() {
	providerDataSources = make(map[string]*schema.Resource)
	providerResources = make(map[string]*schema.Resource)

	regInstance := &registerTestInstance{}

	regInstance.registerTestDataSources()
	regInstance.registerTestResources()
}

func TestMain(m *testing.M) {
	// Run setup function before starting the test suite
	initTestResources()

	// Run the test suite
	m.Run()
}
