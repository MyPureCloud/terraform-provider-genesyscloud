package conversations_messaging_supportedcontent_default

import (
	"sync"
	"testing"

	cmSupportedContent "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/conversations_messaging_supportedcontent"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

/*
   The genesyscloud_conversations_messaging_supportedcontent_default_init_test.go file is used to initialize the data sources and resources
   used in testing the conversations_messaging_supportedcontent_default resource.
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

	providerResources[ResourceType] = ResourceConversationsMessagingSupportedcontentDefault()
	providerResources[cmSupportedContent.ResourceType] = cmSupportedContent.ResourceSupportedContent()
}

// initTestResources initializes all test resources and data sources.
func initTestResources() {
	providerResources = make(map[string]*schema.Resource)

	regInstance := &registerTestInstance{}

	regInstance.registerTestResources()
}

// TestMain is a "setup" function called by the testing framework when run the test
func TestMain(m *testing.M) {
	// Run setup function before starting the test suite for the conversations_messaging_supportedcontent_default package
	initTestResources()

	// Run the test suite for the conversations_messaging_supportedcontent_default package
	m.Run()
}
