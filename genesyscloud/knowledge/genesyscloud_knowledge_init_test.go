package knowledge

import (
	"sync"

	gcloud "terraform-provider-genesyscloud/genesyscloud"
	knowledgeDocument "terraform-provider-genesyscloud/genesyscloud/knowledge_document"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var providerResources map[string]*schema.Resource

type registerTestInstance struct {
	resourceMapMutex sync.RWMutex
}

func (r *registerTestInstance) registerTestResources() {

	r.resourceMapMutex.Lock()
	defer r.resourceMapMutex.Unlock()

	providerResources["genesyscloud_knowledge_document_variation"] = ResourceKnowledgeDocumentVariation()
	providerResources["genesyscloud_knowledge_knowledgebase"] = gcloud.ResourceKnowledgeKnowledgebase()
	providerResources[knowledgeDocument.ResourceType] = knowledgeDocument.ResourceKnowledgeDocument()
}

func initTestResources() {
	providerResources = make(map[string]*schema.Resource)
	regInstance := &registerTestInstance{}
	regInstance.registerTestResources()
}

func TestMain(m *testing.M) {
	// Run setup function before starting the test suite for knowledge Package
	initTestResources()
	// Run the test suite for knowledge
	m.Run()
}
