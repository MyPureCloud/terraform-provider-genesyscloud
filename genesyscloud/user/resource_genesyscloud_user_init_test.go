package user

import (
	"sync"
	"testing"

	gcloud "terraform-provider-genesyscloud/genesyscloud"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

/*
Initializes and registers datasources and resources for the scripts test package
*/
var providerDataSources map[string]*schema.Resource
var providerResources map[string]*schema.Resource

type registerTestInstance struct {
	resourceMapMutex   sync.RWMutex
	datasourceMapMutex sync.RWMutex
}

// registerTestResource registers the CX as Code resources used in test
func (r *registerTestInstance) registerTestResources() {
	r.resourceMapMutex.Lock()
	defer r.resourceMapMutex.Unlock()

	providerResources["genesyscloud_user"] = ResourceUser()
	providerResources["genesyscloud_routing_skill"] = gcloud.ResourceRoutingSkill()
	providerResources["genesyscloud_routing_language"] = gcloud.ResourceRoutingLanguage()
	providerResources["genesyscloud_location"] = gcloud.ResourceLocation()

}

// registerTestResource registers the CX as Code resources used in test
func (r *registerTestInstance) registerTestDataSources() {

	r.datasourceMapMutex.Lock()
	defer r.datasourceMapMutex.Unlock()

	providerDataSources["genesyscloud_user"] = DataSourceUser()

}

// initTestresources initializes all the data sources and resources
func initTestresources() {
	providerDataSources = make(map[string]*schema.Resource)
	providerResources = make(map[string]*schema.Resource)

	reg_instance := &registerTestInstance{}

	reg_instance.registerTestDataSources()
	reg_instance.registerTestResources()

}

// TestMain acts as a setup class that gets call before the tests cases for the resiyrce ryb'
func TestMain(m *testing.M) {

	// Run setup function before starting the test suite for Outbound Package
	initTestresources()

	// Run the test suite for outbound
	m.Run()
}
