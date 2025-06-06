package outbound_callanalysisresponseset

import (
	gcloud "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud"
	flow "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/architect_flow"
	authDivision "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/auth_division"
	obContactList "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/outbound_contact_list"
	routingWrapupcode "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/routing_wrapupcode"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

/*
   The genesyscloud_outbound_callanalysisresponseset_init_test.go file is used to initialize the data sources and resources
   used in testing the outbound_callanalysisresponseset resource.
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

	providerResources[ResourceType] = ResourceOutboundCallanalysisresponseset()
	providerResources[obContactList.ResourceType] = obContactList.ResourceOutboundContactList()
	providerResources[flow.ResourceType] = flow.ResourceArchitectFlow()
	providerResources[routingWrapupcode.ResourceType] = routingWrapupcode.ResourceRoutingWrapupCode()
	providerResources[authDivision.ResourceType] = authDivision.ResourceAuthDivision()
}

// registerTestDataSources registers all data sources used in the tests.
func (r *registerTestInstance) registerTestDataSources() {
	r.datasourceMapMutex.Lock()
	defer r.datasourceMapMutex.Unlock()

	providerDataSources[ResourceType] = DataSourceOutboundCallanalysisresponseset()
	providerDataSources["genesyscloud_auth_division_home"] = gcloud.DataSourceAuthDivisionHome()
}

// initTestResources initializes all test resources and data sources.
func initTestResources() {
	providerDataSources = make(map[string]*schema.Resource)
	providerResources = make(map[string]*schema.Resource)

	regInstance := &registerTestInstance{}

	regInstance.registerTestResources()
	regInstance.registerTestDataSources()
}

// TestMain is a "setup" function called by the testing framework when run the test
func TestMain(m *testing.M) {
	// Run setup function before starting the test suite for the outbound_callanalysisresponseset package
	initTestResources()

	// Run the test suite for the outbound_callanalysisresponseset package
	m.Run()
}
