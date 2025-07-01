package guide_jobs

import (
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/guide"
)

/*
The genesyscloud_guide_jobs_init_test.go file is used to initialize the data sources and resources used in testing the guide_jobs resource
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

	providerResources[ResourceType] = ResourceGuideJobs()
	providerResources[guide.ResourceType] = guide.ResourceGuide()

}

// initTestResources initializes all test resources.
func initTestResources() {
	providerResources = make(map[string]*schema.Resource)

	regInstance := &registerTestInstance{}

	regInstance.registerTestResources()
}

// TestMain is a "setup" function called by the testing framework when running the test
func TestMain(m *testing.M) {
	// Run setup function before starting the test suite for guide_jobs package
	initTestResources()

	// Run the test suite for the guide_jobs package
	m.Run()
}
