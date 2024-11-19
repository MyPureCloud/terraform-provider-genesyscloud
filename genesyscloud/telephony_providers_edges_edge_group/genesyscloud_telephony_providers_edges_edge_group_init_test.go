package telephony_providers_edges_edge_group

import (
	"sync"

	"terraform-provider-genesyscloud/genesyscloud/location"
	tbs "terraform-provider-genesyscloud/genesyscloud/telephony_provider_edges_trunkbasesettings"
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
	providerResources["genesyscloud_telephony_providers_edges_edge_group"] = ResourceEdgeGroup()
	// external package dependencies for Edges Edge group
	providerResources["genesyscloud_telephony_providers_edges_site"] = edgeSite.ResourceSite()
	providerResources["genesyscloud_telephony_providers_edges_trunkbasesettings"] = tbs.ResourceTrunkBaseSettings()
	providerResources["genesyscloud_location"] = location.ResourceLocation()

}

func (r *registerTestInstance) registerTestDataSources() {

	r.datasourceMapMutex.Lock()
	defer r.datasourceMapMutex.Unlock()
	providerDataSources["genesyscloud_telephony_providers_edges_edge_group"] = DataSourceEdgeGroup()
	// external package dependencies for Edges Edge group
	providerDataSources["genesyscloud_telephony_providers_edges_site"] = edgeSite.DataSourceSite()
	providerDataSources["genesyscloud_telephony_providers_edges_trunkbasesettings"] = tbs.DataSourceTrunkBaseSettings()

}

func initTestresources() {
	providerDataSources = make(map[string]*schema.Resource)
	providerResources = make(map[string]*schema.Resource)
	regInstance := &registerTestInstance{}
	regInstance.registerTestDataSources()
	regInstance.registerTestResources()

}

func TestMain(m *testing.M) {
	// Run setup function before starting the test suite for Edges Edge group
	initTestresources()
	// Run the test suite for outbound
	m.Run()
}
