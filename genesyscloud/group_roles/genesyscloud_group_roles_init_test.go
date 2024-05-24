package group_roles

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"sync"
	gcloud "terraform-provider-genesyscloud/genesyscloud"
	authRole "terraform-provider-genesyscloud/genesyscloud/auth_role"
	"terraform-provider-genesyscloud/genesyscloud/group"
	"testing"
)

/*
The genesyscloud_group_roles_init_test.go file is used to initialize the data sources and resources used in testing the group_roles resource
*/

// providerResources holds a map of all registered resources
var providerResources map[string]*schema.Resource

type registerTestInstance struct {
	resourceMapMutex sync.RWMutex
}

// registerTestResources registers all resources used in the tests
func (r *registerTestInstance) registerTestResources() {
	r.resourceMapMutex.Lock()
	defer r.resourceMapMutex.Unlock()

	providerResources["genesyscloud_group_roles"] = ResourceGroupRoles()
	providerResources["genesyscloud_user"] = gcloud.ResourceUser()
	providerResources["genesyscloud_group"] = group.ResourceGroup()
	providerResources["genesyscloud_auth_role"] = authRole.ResourceAuthRole()
	providerResources["genesyscloud_auth_division"] = gcloud.ResourceAuthDivision()
}

// initTestResources initializes all test resources.
func initTestResources() {
	providerResources = make(map[string]*schema.Resource)

	regInstance := &registerTestInstance{}

	regInstance.registerTestResources()
}

// TestMain is a "setup" function called by the testing framework when running the test
func TestMain(m *testing.M) {
	// Run setup function before starting the test suite for group_roles package
	initTestResources()

	// Run the test suite for the group_roles package
	m.Run()
}
