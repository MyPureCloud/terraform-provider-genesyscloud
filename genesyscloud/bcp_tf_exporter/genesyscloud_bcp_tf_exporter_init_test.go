package bcp_tf_exporter

import (
	"sync"
	"testing"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/architect_flow"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/group"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/user"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

/*
   The genesyscloud_bcp_tf_exporter_init_test.go file is used to initialize the resources
   used in testing the bcp_tf_exporter resource.
*/

type registerTestInstance struct {
	resourceMapMutex   sync.RWMutex
	datasourceMapMutex sync.RWMutex
}

// registerTestResources registers all resources used in the tests
func (r *registerTestInstance) registerTestResources() {
	r.resourceMapMutex.Lock()
	defer r.resourceMapMutex.Unlock()

	providerResources[ResourceType] = ResourceBcpTfExporter()
	providerResources[user.ResourceType] = user.ResourceUser()
	providerResources[group.ResourceType] = group.ResourceGroup()
	providerResources[architect_flow.ResourceType] = architect_flow.ResourceArchitectFlow()
}

// initTestResources initializes all test resources and data sources.
func initTestResources() {
	providerResources = make(map[string]*schema.Resource)

	regInstance := &registerTestInstance{}

	regInstance.registerTestResources()

}

// TestMain is a "setup" function called by the testing framework when run the test
func TestMain(m *testing.M) {
	// Run setup function before starting the test suite for the auth_role package
	initTestResources()

	// Run the test suite for the auth_role package
	m.Run()
}
