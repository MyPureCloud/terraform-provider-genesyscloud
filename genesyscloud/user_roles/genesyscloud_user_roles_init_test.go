package user_roles

import (
	gcloud "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud"
	authDivision "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/auth_division"
	authRole "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/auth_role"
	"sync"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/user"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

/*
The genesyscloud_user_roles_init_test.go file is used to initialize the data sources and resources used in testing the user_roles resource
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

	providerResources[ResourceType] = ResourceUserRoles()
	providerResources[user.ResourceType] = user.ResourceUser()
	providerResources[authRole.ResourceType] = authRole.ResourceAuthRole()
	providerResources[authDivision.ResourceType] = authDivision.ResourceAuthDivision()
}

// registerTestDataSources registers all data sources used in the tests.
func (r *registerTestInstance) registerTestDataSources() {
	r.datasourceMapMutex.Lock()
	defer r.datasourceMapMutex.Unlock()

	providerDataSources[authRole.ResourceType] = authRole.DataSourceAuthRole()
	providerDataSources["genesyscloud_auth_division_home"] = gcloud.DataSourceAuthDivisionHome()

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
	// Run setup function before starting the test suite for user_roles package
	initTestResources()

	// Run the test suite for the user_roles package
	m.Run()
}
