package business_rules_decision_table

import (
	"log"
	"sync"
	"testing"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud"
	authDivision "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/auth_division"
	businessRulesSchema "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/business_rules_schema"
	routingQueue "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/routing_queue"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v178/platformclientv2"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
)

// providerDataSources holds a map of all registered datasources
var providerDataSources map[string]*schema.Resource

// providerResources holds a map of all registered resources
var providerResources map[string]*schema.Resource

var (
	sdkConfig *platformclientv2.Configuration
	authErr   error
)

type registerTestInstance struct {
	resourceMapMutex   sync.RWMutex
	datasourceMapMutex sync.RWMutex
}

// registerTestResources registers all resources used in the tests
func (r *registerTestInstance) registerTestResources() {
	r.resourceMapMutex.Lock()
	defer r.resourceMapMutex.Unlock()
	providerResources[ResourceType] = ResourceBusinessRulesDecisionTable()
	providerResources[authDivision.ResourceType] = authDivision.ResourceAuthDivision()
	providerResources[businessRulesSchema.ResourceType] = businessRulesSchema.ResourceBusinessRulesSchema()
	providerResources[routingQueue.ResourceType] = routingQueue.ResourceRoutingQueue()
}

// registerTestDataSources registers all data sources used in the tests.
func (r *registerTestInstance) registerTestDataSources() {
	r.datasourceMapMutex.Lock()
	defer r.datasourceMapMutex.Unlock()
	providerDataSources[ResourceType] = DataSourceBusinessRulesDecisionTable()
	providerDataSources["genesyscloud_auth_division_home"] = genesyscloud.DataSourceAuthDivisionHome()
	providerDataSources["genesyscloud_routing_queue"] = routingQueue.DataSourceRoutingQueue()
}

// initTestResources initializes all test_data resources and data sources.
func initTestResources() {
	sdkConfig, authErr = provider.AuthorizeSdk()
	if authErr != nil {
		log.Fatalf("failed to authorize sdk for package business_rules_decision_table: %v", authErr)
	}

	providerDataSources = make(map[string]*schema.Resource)
	providerResources = make(map[string]*schema.Resource)

	regInstance := &registerTestInstance{}

	regInstance.registerTestResources()
	regInstance.registerTestDataSources()
}

// TestMain is a "setup" function called by the testing framework when run the test_data
func TestMain(m *testing.M) {
	// Run setup function before starting the test_data suite for the package
	initTestResources()

	// Run the test_data suite for the business_rules_decision_table package
	m.Run()
}
