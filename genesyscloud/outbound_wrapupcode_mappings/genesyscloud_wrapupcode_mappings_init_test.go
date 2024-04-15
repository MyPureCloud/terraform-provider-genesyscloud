package outbound_wrapupcode_mappings

import (
	"sync"
	"testing"

	gcloud "terraform-provider-genesyscloud/genesyscloud"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// providerResources holds a map of all registered datasources.
var providerResources map[string]*schema.Resource
var providerDataSources map[string]*schema.Resource

type registerTestInstance struct {
	resourceMapMutex sync.RWMutex
}

// registerTestResources registers all resources used in the tests
func (r *registerTestInstance) registerTestResources() {
	r.resourceMapMutex.Lock()
	defer r.resourceMapMutex.Unlock()

	providerResources[resourceName] = ResourceOutboundWrapUpCodeMappings()
	providerResources["genesyscloud_routing_wrapupcode"] = gcloud.ResourceRoutingWrapupCode()

}

// initTestResources initializes all test resources and data sources.
func initTestResources() {
	providerResources = make(map[string]*schema.Resource)

	regInstance := &registerTestInstance{}

	regInstance.registerTestResources()
}

// initTestDataSources is used to initialize data sources used in the test code.  There are no data sources associated with genesyscloud_wrapupcode_mappings resources.
func initTestDataSources() {
	providerDataSources = make(map[string]*schema.Resource) //Keep this here or Null Pointers will abound
}

// TestMain is a "setup" function called by the testing framework when run the
func TestMain(m *testing.M) {
	initTestResources()
	initTestDataSources()

	m.Run()
}
