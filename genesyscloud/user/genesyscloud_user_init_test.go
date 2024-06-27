package user

import (
	"sync"
	gcloud "terraform-provider-genesyscloud/genesyscloud"

	authRole "terraform-provider-genesyscloud/genesyscloud/auth_role"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

/*
The genesyscloud_user_init_test.go file is used to initialize the data sources and resources used in testing the user resource
*/

// providerDataSources holds a map of all registered datasources
var providerDataSources map[string]*schema.Resource

// providerResources holds a map of all registered resources
var providerResources map[string]*schema.Resource

type registerTestInstance struct {
	resourceMapMutex   sync.RWMutex
	datasourceMapMutex sync.RWMutex
}

// registerTestResources registers all resources used in the tests
func (r *registerTestInstance) registerTestResources() {
	r.resourceMapMutex.Lock()
	defer r.resourceMapMutex.Unlock()

	providerResources[resourceName] = ResourceUser()
	providerResources["genesyscloud_auth_role"] = authRole.ResourceAuthRole()
	providerResources["genesyscloud_auth_division"] = gcloud.ResourceAuthDivision()
	providerResources["genesyscloud_routing_skill"] = gcloud.ResourceRoutingSkill()
	providerResources["genesyscloud_routing_language"] = gcloud.ResourceRoutingLanguage()
	providerResources["genesyscloud_location"] = gcloud.ResourceLocation()
	providerResources["genesyscloud_routing_utilization_label"] = gcloud.ResourceRoutingUtilizationLabel()

}

// registerTestDataSources registers all data sources used in the tests.
func (r *registerTestInstance) registerTestDataSources() {
	r.datasourceMapMutex.Lock()
	defer r.datasourceMapMutex.Unlock()
	providerDataSources[resourceName] = DataSourceUser()
	providerDataSources["genesyscloud_auth_role"] = authRole.DataSourceAuthRole()
	providerDataSources["genesyscloud_auth_division_home"] = gcloud.DataSourceAuthDivisionHome()
	providerDataSources["genesyscloud_routing_skill"] = gcloud.DataSourceRoutingSkill()
	providerDataSources["genesyscloud_routing_language"] = gcloud.DataSourceRoutingLanguage()
	providerDataSources["genesyscloud_location"] = gcloud.DataSourceLocation()
	providerDataSources["genesyscloud_routing_utilization_label"] = gcloud.DataSourceRoutingUtilizationLabel()
}

// initTestResources initializes all test resources.
func initTestResources() {
	providerResources = make(map[string]*schema.Resource)
	providerDataSources = make(map[string]*schema.Resource)

	regInstance := &registerTestInstance{}

	regInstance.registerTestResources()
	regInstance.registerTestDataSources()
}

// TestMain is a "setup" function called by the testing framework when run the test
func TestMain(m *testing.M) {
	// Run setup function before starting the test suite for user package
	initTestResources()

	// Run the test suite for the user package
	m.Run()
}
