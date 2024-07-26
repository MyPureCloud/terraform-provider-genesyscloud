package telephony_providers_edges_trunk

import (
	"sync"
	gcloud "terraform-provider-genesyscloud/genesyscloud"
	telephony "terraform-provider-genesyscloud/genesyscloud/telephony"
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

	providerResources["genesyscloud_telephony_providers_edges_trunkbasesettings"] = telephony.ResourceTrunkBaseSettings()
	providerResources[resourceName] = ResourceTrunk()

	// external package dependencies
	providerResources["genesyscloud_telephony_providers_edges_site"] = edgeSite.ResourceSite()

	providerResources["genesyscloud_location"] = gcloud.ResourceLocation()
	providerResources["genesyscloud_telephony_providers_edges_edge_group"] = edgeGroup.ResourceEdgeGroup()

}

func (r *registerTestInstance) registerTestDataSources() {

	r.datasourceMapMutex.Lock()
	defer r.datasourceMapMutex.Unlock()

	providerDataSources["genesyscloud_telephony_providers_edges_trunkbasesettings"] = telephony.DataSourceTrunkBaseSettings()
	providerDataSources[resourceName] = DataSourceTrunk()
	// external package dependencies
	providerDataSources["genesyscloud_telephony_providers_edges_site"] = edgeSite.DataSourceSite()
	providerDataSources["genesyscloud_telephony_providers_edges_edge_group"] = edgeGroup.DataSourceEdgeGroup()

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
