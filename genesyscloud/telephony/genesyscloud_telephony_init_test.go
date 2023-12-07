package telephony

import (
	"sync"
	gcloud "terraform-provider-genesyscloud/genesyscloud"

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

	providerResources["genesyscloud_telephony_providers_edges_trunkbasesettings"] = ResourceTrunkBaseSettings()
	providerResources["genesyscloud_telephony_providers_edges_edge_group"] = ResourceEdgeGroup()
	providerResources["genesyscloud_telephony_providers_edges_trunk"] = ResourceTrunk()

	// external package dependencies for outbound
	providerResources["genesyscloud_telephony_providers_edges_site"] = edgeSite.ResourceSite()

	providerResources["genesyscloud_location"] = gcloud.ResourceLocation()

}

func (r *registerTestInstance) registerTestDataSources() {

	r.datasourceMapMutex.Lock()
	defer r.datasourceMapMutex.Unlock()

	providerDataSources["genesyscloud_telephony_providers_edges_trunkbasesettings"] = dataSourceTrunkBaseSettings()
	providerDataSources["genesyscloud_telephony_providers_edges_edge_group"] = dataSourceEdgeGroup()
	providerDataSources["genesyscloud_telephony_providers_edges_trunk"] = dataSourceTrunk()
	// external package dependencies for outbound
	providerDataSources["genesyscloud_telephony_providers_edges_site"] = edgeSite.DataSourceSite()

}

func initTestresources() {
	providerDataSources = make(map[string]*schema.Resource)
	providerResources = make(map[string]*schema.Resource)

	reg_instance := &registerTestInstance{}

	reg_instance.registerTestDataSources()
	reg_instance.registerTestResources()

}

func TestMain(m *testing.M) {

	// Run setup function before starting the test suite for Outbound Package
	initTestresources()

	// Run the test suite for outbound
	m.Run()
}
