package architect_grammar_language

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"sync"
	architectGrammar "terraform-provider-genesyscloud/genesyscloud/architect_grammar"
	"testing"
)

/*
   The genesyscloud_architect_grammar_language_init_test.go file is used to initialize the resource
   used in testing the architect_grammar_language resource.
*/

// providerDataSources holds a map of all registered data sources
var providerDataSources map[string]*schema.Resource

// providerResources holds a map of all registered resources
var providerResources map[string]*schema.Resource

type registerTestInstance struct {
	resourceMapMutex sync.RWMutex
}

// registerTestResources registers all resources used in the tests
func (r *registerTestInstance) registerTestResources() {
	r.resourceMapMutex.Lock()
	defer r.resourceMapMutex.Unlock()

	providerResources["genesyscloud_architect_grammar_language"] = ResourceArchitectGrammarLanguage()
	providerResources["genesyscloud_architect_grammar"] = architectGrammar.ResourceArchitectGrammar()
}

// initTestResources initializes all test_data resources and data sources.
func initTestResources() {
	providerDataSources = make(map[string]*schema.Resource)
	providerResources = make(map[string]*schema.Resource)

	regInstance := &registerTestInstance{}

	regInstance.registerTestResources()
}

// TestMain is a "setup" function called by the testing framework when run the test_data
func TestMain(m *testing.M) {
	// Run setup function before starting the test_data suite for the architect_grammar_language package
	initTestResources()

	// Run the test_data suite for the architect_grammar_language package
	m.Run()
}
