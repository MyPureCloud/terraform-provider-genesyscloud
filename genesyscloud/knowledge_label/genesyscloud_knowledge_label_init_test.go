package knowledge_label

import (
	knowledgeKnowledgebase "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/knowledge_knowledgebase"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

/*
The genesyscloud_label_init_test.go file is used to initialize the data sources and resources
used in testing the location resource.
*/

// providerDataSources holds a map of all registered datasources

// providerResources holds a map of all registered resources
var providerResources map[string]*schema.Resource
var providerDataSources map[string]*schema.Resource

type registerTestInstance struct {
	resourceMapMutex sync.RWMutex
}

// registerTestResources registers all resources used in the tests
func (r *registerTestInstance) registerTestResources() {
	r.resourceMapMutex.Lock()
	defer r.resourceMapMutex.Unlock()
	providerResources[knowledgeKnowledgebase.ResourceType] = knowledgeKnowledgebase.ResourceKnowledgeKnowledgebase()
	providerResources[ResourceType] = ResourceKnowledgeLabel()
}

// registerTestDataSources registers all data sources used in the tests
func (r *registerTestInstance) registerTestDataSources() {
	r.resourceMapMutex.Lock()
	defer r.resourceMapMutex.Unlock()
	providerDataSources[ResourceType] = dataSourceKnowledgeLabel()
}

// initTestResources initializes all test resources and data sources.
func initTestResources() {
	providerResources = make(map[string]*schema.Resource)
	providerDataSources = make(map[string]*schema.Resource)
	regInstance := &registerTestInstance{}

	regInstance.registerTestResources()
	regInstance.registerTestDataSources()
}

// TestMain is a "setup" function called by the testing framework when run the test
func TestMain(m *testing.M) {
	// Run setup function before starting the test suite for knowledge label package
	initTestResources()

	// Run the test suite for the knowledge label package
	m.Run()
}
