package conversations_messaging_integrations_instagram

import (
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	cmMessagingSetting "terraform-provider-genesyscloud/genesyscloud/conversations_messaging_settings"
	cmSupportedContent "terraform-provider-genesyscloud/genesyscloud/conversations_messaging_supportedcontent"
)

/*
   The genesyscloud_conversations_messaging_integrations_instagram_init_test.go file is used to initialize the data sources and resources
   used in testing the conversations_messaging_integrations_instagram resource.
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

	providerResources[ResourceType] = ResourceConversationsMessagingIntegrationsInstagram()
	providerResources[cmSupportedContent.ResourceType] = cmSupportedContent.ResourceSupportedContent()
	providerResources[cmMessagingSetting.ResourceType] = cmMessagingSetting.ResourceConversationsMessagingSettings()
}

// registerTestDataSources registers all data sources used in the tests.
func (r *registerTestInstance) registerTestDataSources() {
	r.datasourceMapMutex.Lock()
	defer r.datasourceMapMutex.Unlock()

	providerDataSources[ResourceType] = DataSourceConversationsMessagingIntegrationsInstagram()
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
	// Run setup function before starting the test suite for the conversations_messaging_integrations_instagram package
	initTestResources()

	// Run the test suite for the conversations_messaging_integrations_instagram package
	m.Run()
}
