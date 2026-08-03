package conversations_messaging_integrations_apple

import (
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
)

/*
The genesyscloud_apple_integration_init_test.go file is used to initialize the data sources and resources
used in testing the apple_integration package.
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

	providerResources[ResourceType] = ResourceConversationsMessagingIntegrationsApple()
}

// registerTestDataSources registers all data sources used in the tests
func (r *registerTestInstance) registerTestDataSources() {
	r.datasourceMapMutex.Lock()
	defer r.datasourceMapMutex.Unlock()

	providerDataSources[ResourceType] = DataSourceConversationsMessagingIntegrationsApple()
}

// initTestResources initializes all test resources and data sources
func initTestResources() {
	sdkConfig, err := provider.AuthorizeSdk()
	if err != nil {
		return
	}

	providerDataSources = make(map[string]*schema.Resource)
	providerResources = make(map[string]*schema.Resource)

	regInstance := &registerTestInstance{}
	regInstance.registerTestDataSources()
	regInstance.registerTestResources()

	// Set the internal proxy for testing
	internalProxy = newConversationsMessagingIntegrationsAppleProxy(sdkConfig)
}

// TestMain runs the test suite for the apple_integration package
func TestMain(m *testing.M) {
	// Run setup function before starting the test suite
	initTestResources()
	// Run the test suite
	m.Run()
}
