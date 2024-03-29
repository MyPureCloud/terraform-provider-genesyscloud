package integration_credential

import (
	"sync"
	"terraform-provider-genesyscloud/genesyscloud/auth_role"
	oauth "terraform-provider-genesyscloud/genesyscloud/oauth_client"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

/*
   The genesyscloud_integration_credential_init_test.go file is used to initialize the data sources and resources
   used in testing the integration credential resource.

   Please make sure you register ALL resources and data sources your test cases will use.
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

	providerResources["genesyscloud_integration_credential"] = ResourceIntegrationCredential()
	providerResources["genesyscloud_oauth_client"] = oauth.ResourceOAuthClient()
}

// registerTestDataSources registers all data sources used in the tests.
func (r *registerTestInstance) registerTestDataSources() {
	r.datasourceMapMutex.Lock()
	defer r.datasourceMapMutex.Unlock()

	providerDataSources["genesyscloud_integration_credential"] = DataSourceIntegrationCredential()
	providerDataSources["genesyscloud_auth_role"] = auth_role.DataSourceAuthRole()
}

// initTestResources initializes all test resources and data sources.
func initTestResources() {
	providerDataSources = make(map[string]*schema.Resource)
	providerResources = make(map[string]*schema.Resource)

	regInstance := &registerTestInstance{}

	regInstance.registerTestDataSources()
	regInstance.registerTestResources()
}

// TestMain is a "setup" function called by the testing framework when run the test
func TestMain(m *testing.M) {
	// Run setup function before starting the test suite for integration_credential package
	initTestResources()

	// Run the test suite for the integration_credential package
	m.Run()
}
