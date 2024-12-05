package group_roles

import (
	"sync"
	authDivision "terraform-provider-genesyscloud/genesyscloud/auth_division"
	authRole "terraform-provider-genesyscloud/genesyscloud/auth_role"
	"terraform-provider-genesyscloud/genesyscloud/group"
	"terraform-provider-genesyscloud/genesyscloud/user"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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

	providerResources[ResourceType] = ResourceGroupRoles()
	providerResources[user.ResourceType] = user.ResourceUser()
	providerResources[group.ResourceType] = group.ResourceGroup()
	providerResources[authRole.ResourceType] = authRole.ResourceAuthRole()
	providerResources[authDivision.ResourceType] = authDivision.ResourceAuthDivision()
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
