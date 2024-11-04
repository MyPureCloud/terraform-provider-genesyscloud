package telephony_provider_edges_trunkbasesettings

import (
	"sync"

	"terraform-provider-genesyscloud/genesyscloud/location"
	edgeSite "terraform-provider-genesyscloud/genesyscloud/telephony_providers_edges_site"
	edgeTrunk "terraform-provider-genesyscloud/genesyscloud/telephony_providers_edges_trunk"
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

	providerResources["genesyscloud_telephony_providers_edges_trunkbasesettings"] = ResourceTrunkBaseSettings()
	providerResources["genesyscloud_telephony_providers_edges_trunk"] = edgeTrunk.ResourceTrunk()
	// external package dependencies for outbound
	providerResources["genesyscloud_telephony_providers_edges_site"] = edgeSite.ResourceSite()
	providerResources["genesyscloud_location"] = location.ResourceLocation()

}

func (r *registerTestInstance) registerTestDataSources() {

	r.datasourceMapMutex.Lock()
	defer r.datasourceMapMutex.Unlock()
	providerDataSources["genesyscloud_telephony_providers_edges_trunkbasesettings"] = DataSourceTrunkBaseSettings()
	providerDataSources["genesyscloud_telephony_providers_edges_trunk"] = edgeTrunk.DataSourceTrunk()
	// external package dependencies for outbound
	providerDataSources["genesyscloud_telephony_providers_edges_site"] = edgeSite.DataSourceSite()

}

func initTestResources() {
	providerDataSources = make(map[string]*schema.Resource)
	providerResources = make(map[string]*schema.Resource)
	regInstance := &registerTestInstance{}
	regInstance.registerTestDataSources()
	regInstance.registerTestResources()
}

func TestMain(m *testing.M) {
	// Run setup function before starting the test suite for Outbound Package
	initTestResources()
	// Run the test suite for outbound
	m.Run()
}
