package telephony_providers_edges_edge_group

import (
	"sync"
	gcloud "terraform-provider-genesyscloud/genesyscloud"

	telephony "terraform-provider-genesyscloud/genesyscloud/telephony"
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

	// external package dependencies for outbound
	providerResources["genesyscloud_telephony_providers_edges_site"] = edgeSite.ResourceSite()
	providerResources["genesyscloud_telephony_providers_edges_trunkbasesettings"] = telephony.ResourceTrunkBaseSettings()

	providerResources["genesyscloud_location"] = gcloud.ResourceLocation()

}

func (r *registerTestInstance) registerTestDataSources() {

	r.datasourceMapMutex.Lock()
	defer r.datasourceMapMutex.Unlock()

	providerDataSources["genesyscloud_telephony_providers_edges_edge_group"] = DataSourceEdgeGroup()
	// external package dependencies for outbound
	providerDataSources["genesyscloud_telephony_providers_edges_site"] = edgeSite.DataSourceSite()
	providerDataSources["genesyscloud_telephony_providers_edges_trunkbasesettings"] = telephony.DataSourceTrunkBaseSettings()

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
