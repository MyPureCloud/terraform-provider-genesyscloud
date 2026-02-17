package group_roles

import (
	"log"
	"sync"
	"testing"

	"github.com/mypurecloud/platform-client-sdk-go/v176/platformclientv2"
	authDivision "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/auth_division"
	authRole "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/auth_role"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/group"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/user"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

/*
The genesyscloud_group_roles_init_test.go file is used to initialize the data sources and resources used in testing the group_roles resource
*/

var (
	// providerResources holds a map of all registered resources
	providerResources map[string]*schema.Resource

	// frameworkResources holds a map of all registered Framework resources
	frameworkResources map[string]func() resource.Resource

	// frameworkDataSources holds a map of all registered Framework data sources
	frameworkDataSources map[string]func() datasource.DataSource

	sdkConfig *platformclientv2.Configuration
	authErr   error
)

type registerTestInstance struct {
	resourceMapMutex            sync.RWMutex
	frameworkResourceMapMutex   sync.RWMutex
	frameworkDataSourceMapMutex sync.RWMutex
}

// registerTestResources registers all resources used in the tests
func (r *registerTestInstance) registerTestResources() {
	r.resourceMapMutex.Lock()
	defer r.resourceMapMutex.Unlock()

	providerResources[ResourceType] = ResourceGroupRoles()
	providerResources[group.ResourceType] = group.ResourceGroup()
	providerResources[authRole.ResourceType] = authRole.ResourceAuthRole()
	providerResources[authDivision.ResourceType] = authDivision.ResourceAuthDivision()
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
	sdkConfig, authErr = provider.AuthorizeSdk()
	if authErr != nil {
		log.Fatalf("failed to authorize sdk for the package group_roles: %v", authErr)
	}

	providerResources = make(map[string]*schema.Resource)
	frameworkResources = make(map[string]func() resource.Resource)
	frameworkDataSources = make(map[string]func() datasource.DataSource)

	regInstance := &registerTestInstance{}

	regInstance.registerTestResources()
	regInstance.registerFrameworkTestResources()
	regInstance.registerFrameworkTestDataSources()
}

// TestMain is a "setup" function called by the testing framework when running the test
func TestMain(m *testing.M) {
	// Run setup function before starting the test suite for group_roles package
	initTestResources()

	// Run the test suite for the group_roles package
	m.Run()
}
