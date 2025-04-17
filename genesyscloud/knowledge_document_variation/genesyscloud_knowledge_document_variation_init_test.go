package knowledge_document_variation

import (
	knowledgeDocument "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/knowledge_document"
	knowledgeKnowledgebase "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/knowledge_knowledgebase"
	"sync"

	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// providerResources holds a map of all registered resources
var providerResources map[string]*schema.Resource
var providerDataSources map[string]*schema.Resource

type registerTestInstance struct {
	resourceMapMutex   sync.RWMutex
	datasourceMapMutex sync.RWMutex
}

// registerTestResources registers all resources used in the tests
func (r *registerTestInstance) registerTestResources() {
	r.resourceMapMutex.Lock()
	defer r.resourceMapMutex.Unlock()

	providerResources[ResourceType] = ResourceKnowledgeDocumentVariation()
	providerResources[knowledgeKnowledgebase.ResourceType] = knowledgeKnowledgebase.ResourceKnowledgeKnowledgebase()
	providerResources[knowledgeDocument.ResourceType] = knowledgeDocument.ResourceKnowledgeDocument()
}

// registerTestDataSources registers all data sources used in the tests.
func (r *registerTestInstance) registerTestDataSources() {
	r.datasourceMapMutex.Lock()
	defer r.datasourceMapMutex.Unlock()

	providerDataSources[ResourceType] = dataSourceKnowledgeDocumentVariation()
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
	initTestResources()

	// Run the test suite for the package
	m.Run()
}
