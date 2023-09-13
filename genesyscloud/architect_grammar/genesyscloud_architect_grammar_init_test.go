package architect_grammar

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"sync"
	"testing"
)

/*
   The genesyscloud_architect_grammar_init_test.go file is used to initialize the data sources and resources
   used in testing the architect_grammar resource.
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
	providerResources["genesyscloud_architect_grammar"] = ResourceArchitectGrammar()
}

// registerTestDataSources registers all data sources used in the tests.
func (r *registerTestInstance) registerTestDataSources() {
	providerDataSources["genesyscloud_architect_grammar"] = DataSourceArchitectGrammar()
}

// initTestresources initializes all test resources and data sources.
func initTestresources() {
	providerDataSources = make(map[string]*schema.Resource)
	providerResources = make(map[string]*schema.Resource)

	reg_instance := &registerTestInstance{}

	reg_instance.registerTestResources()
	reg_instance.registerTestDataSources()
}

// TestMain is a "setup" function called by the testing framework when run the test
func TestMain(m *testing.M) {
	// Run setup function before starting the test suite for the architect_grammar package
	initTestresources()

	// Run the test suite for the architect_grammar package
	m.Run()
}
