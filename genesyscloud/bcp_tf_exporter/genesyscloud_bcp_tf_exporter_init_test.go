package bcp_tf_exporter

import (
	"testing"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/architect_flow"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/group"
	registrar "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_register"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/user"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

/*
   The genesyscloud_bcp_tf_exporter_init_test.go file is used to initialize the resources
   used in testing the bcp_tf_exporter resource.
*/

// initTestResources initializes all test resources and data sources.
func initTestResources() {

	resources := make(map[string]*schema.Resource)
	resources[ResourceType] = ResourceBcpTfExporter()
	resources[user.ResourceType] = user.ResourceUser()
	resources[group.ResourceType] = group.ResourceGroup()
	resources[architect_flow.ResourceType] = architect_flow.ResourceArchitectFlow()

	registrar.SetResources(resources, make(map[string]*schema.Resource))
}

// TestMain is a "setup" function called by the testing framework when run the test
func TestMain(m *testing.M) {
	// Run setup function before starting the test suite for the auth_role package
	initTestResources()

	// Run the test suite for the auth_role package
	m.Run()
}
