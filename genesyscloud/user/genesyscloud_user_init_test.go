package user

import (
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud"
	authDivision "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/auth_division"
	authRole "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/auth_role"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/location"
	routinglanguage "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/routing_language"
	routingSkill "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/routing_skill"
	routingUtilizationLabel "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/routing_utilization_label"
	extensionPool "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/telephony_providers_edges_extension_pool"
)

/*
The genesyscloud_user_init_test.go file is used to initialize the data sources and resources
used in testing the user resource.
*/

// providerDataSources holds a map of all registered SDKv2 datasources,
// should be removed after complete migrtion to plugin Framework
var providerDataSources map[string]*schema.Resource

// providerResources holds a map of all registered SDKv2 resources
// should be removed after complete migrtion to plugin Framework
var providerResources map[string]*schema.Resource

// frameworkResources holds a map of all registered Framework resources
var frameworkResources map[string]func() resource.Resource

// frameworkDataSources holds a map of all registered Framework data sources
var frameworkDataSources map[string]func() datasource.DataSource

type registerTestInstance struct {
	resourceMapMutex            sync.RWMutex
	datasourceMapMutex          sync.RWMutex
	frameworkResourceMapMutex   sync.RWMutex
	frameworkDataSourceMapMutex sync.RWMutex
}

// registerTestResources registers all SDKv2 resources used in the tests
func (r *registerTestInstance) registerTestResources() {
	r.resourceMapMutex.Lock()
	defer r.resourceMapMutex.Unlock()

	// Register SDKv2 resources needed for Framework tests
	providerResources[authRole.ResourceType] = authRole.ResourceAuthRole()
	providerResources[authDivision.ResourceType] = authDivision.ResourceAuthDivision()
	providerResources[location.ResourceType] = location.ResourceLocation()
	providerResources[routingSkill.ResourceType] = routingSkill.ResourceRoutingSkill()
	providerResources[routingUtilizationLabel.ResourceType] = routingUtilizationLabel.ResourceRoutingUtilizationLabel()
	providerResources[extensionPool.ResourceType] = extensionPool.ResourceTelephonyExtensionPool()
}

// registerTestDataSources registers all SDKv2 data sources used in the tests
func (r *registerTestInstance) registerTestDataSources() {
	r.datasourceMapMutex.Lock()
	defer r.datasourceMapMutex.Unlock()

	// Register SDKv2 data sources needed for Framework tests
	providerDataSources[authRole.ResourceType] = authRole.DataSourceAuthRole()
	providerDataSources["genesyscloud_auth_division_home"] = genesyscloud.DataSourceAuthDivisionHome()
	providerDataSources[location.ResourceType] = location.DataSourceLocation()
	providerDataSources[routingSkill.ResourceType] = routingSkill.DataSourceRoutingSkill()
	providerDataSources[routingUtilizationLabel.ResourceType] = routingUtilizationLabel.DataSourceRoutingUtilizationLabel()
}

// registerFrameworkTestResources registers all Framework resources used in the tests
func (r *registerTestInstance) registerFrameworkTestResources() {
	r.frameworkResourceMapMutex.Lock()
	defer r.frameworkResourceMapMutex.Unlock()

	frameworkResources[ResourceType] = NewUserFrameworkResource
	frameworkResources[routinglanguage.ResourceType] = routinglanguage.NewFrameworkRoutingLanguageResource
}

// registerFrameworkTestDataSources registers all Framework data sources used in the tests
func (r *registerTestInstance) registerFrameworkTestDataSources() {
	r.frameworkDataSourceMapMutex.Lock()
	defer r.frameworkDataSourceMapMutex.Unlock()

	frameworkDataSources[ResourceType] = NewUserFrameworkDataSource
	frameworkDataSources[routinglanguage.ResourceType] = routinglanguage.NewFrameworkRoutingLanguageDataSource
}

// initTestResources initializes all test resources and data sources.
func initTestResources() {
	// Initialize both SDKv2 and Framework resources for mixed provider tests
	providerResources = make(map[string]*schema.Resource)
	providerDataSources = make(map[string]*schema.Resource)
	frameworkResources = make(map[string]func() resource.Resource)
	frameworkDataSources = make(map[string]func() datasource.DataSource)

	regInstance := &registerTestInstance{}

	// Register SDKv2 resources and data sources (needed for dependencies)
	regInstance.registerTestResources()
	regInstance.registerTestDataSources()

	// Register Framework resources and data sources
	regInstance.registerFrameworkTestResources()
	regInstance.registerFrameworkTestDataSources()
}

// TestMain is a "setup" function called by the testing framework when run the test
func TestMain(m *testing.M) {
	// Run setup function before starting the test suite for user package
	initTestResources()

	// Run the test suite for the user package
	m.Run()
}
