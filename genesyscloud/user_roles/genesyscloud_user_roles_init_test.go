package user_roles

import (
	"sync"
	"testing"

	gcloud "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud"
	authDivision "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/auth_division"
	authRole "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/auth_role"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/user"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

/*
The genesyscloud_user_roles_init_test.go file is used to initialize the data sources and resources used in testing the user_roles resource
*/

// providerDataSources holds a map of all registered datasources
var providerDataSources map[string]*schema.Resource

// providerResources holds a map of all registered resources
var providerResources map[string]*schema.Resource

// frameworkResources holds a map of all registered Framework resources
var frameworkResources map[string]func() frameworkresource.Resource

// frameworkDataSources holds a map of all registered Framework data sources
var frameworkDataSources map[string]func() datasource.DataSource

type registerTestInstance struct {
	resourceMapMutex            sync.RWMutex
	datasourceMapMutex          sync.RWMutex
	frameworkResourceMapMutex   sync.RWMutex
	frameworkDataSourceMapMutex sync.RWMutex
}

// registerTestResources registers all resources used in the tests
func (r *registerTestInstance) registerTestResources() {
	r.resourceMapMutex.Lock()
	defer r.resourceMapMutex.Unlock()

	providerResources[ResourceType] = ResourceUserRoles()
	// user.ResourceUser() removed - now Framework-only
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

// registerFrameworkTestResources registers all Framework resources used in the tests
func (r *registerTestInstance) registerFrameworkTestResources() {
	r.frameworkResourceMapMutex.Lock()
	defer r.frameworkResourceMapMutex.Unlock()

	frameworkResources[user.ResourceType] = user.NewUserFrameworkResource
}

// registerFrameworkTestDataSources registers all Framework data sources used in the tests
func (r *registerTestInstance) registerFrameworkTestDataSources() {
	r.frameworkDataSourceMapMutex.Lock()
	defer r.frameworkDataSourceMapMutex.Unlock()

	frameworkDataSources[user.ResourceType] = user.NewUserFrameworkDataSource
}

// initTestResources initializes all test resources.
func initTestResources() {
	providerResources = make(map[string]*schema.Resource)
	providerDataSources = make(map[string]*schema.Resource)
	frameworkResources = make(map[string]func() frameworkresource.Resource)
	frameworkDataSources = make(map[string]func() datasource.DataSource)

	regInstance := &registerTestInstance{}

	regInstance.registerTestResources()
	regInstance.registerTestDataSources()
	regInstance.registerFrameworkTestResources()
	regInstance.registerFrameworkTestDataSources()
}

// TestMain is a "setup" function called by the testing framework when run the test
func TestMain(m *testing.M) {
	// Run setup function before starting the test suite for user_roles package
	initTestResources()

	// Run the test suite for the user_roles package
	m.Run()
}
