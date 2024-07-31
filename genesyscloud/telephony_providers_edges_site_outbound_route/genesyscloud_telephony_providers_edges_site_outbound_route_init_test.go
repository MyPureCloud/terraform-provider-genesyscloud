package telephony_providers_edges_site_outbound_route

import (
	"log"
	"sync"
	gcloud "terraform-provider-genesyscloud/genesyscloud"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/telephony"
	"terraform-provider-genesyscloud/genesyscloud/telephony_providers_edges_site"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

/*
   The genesyscloud_telephony_providers_edges_site_outbound_routes.go file is used to initialize the data sources and resources
   used in testing the edges site outbound routes.

   Please make sure you register ALL resources and data sources your test cases will use.
*/

// used in sdk authorization for tests
var (
	sdkConfig *platformclientv2.Configuration
	authErr   error
)

// providerResources holds a map of all registered sites
var providerResources map[string]*schema.Resource

type registerTestInstance struct {
	resourceMapMutex sync.RWMutex
}

// registerTestResources registers all resources used in the tests
func (r *registerTestInstance) registerTestResources() {
	r.resourceMapMutex.Lock()
	defer r.resourceMapMutex.Unlock()

	providerResources[resourceName] = ResourceSiteOutboundRoute()
	providerResources["genesyscloud_telephony_providers_edges_trunkbasesettings"] = telephony.ResourceTrunkBaseSettings()
	providerResources["genesyscloud_location"] = gcloud.ResourceLocation()
	providerResources["genesyscloud_telephony_providers_edges_site"] = telephony_providers_edges_site.ResourceSite()
}

// initTestResources initializes all test resources and data sources.
func initTestResources() {
	sdkConfig, authErr = provider.AuthorizeSdk()
	if authErr != nil {
		log.Fatalf("failed to authorize sdk: %v", authErr)
	}

	providerResources = make(map[string]*schema.Resource)

	regInstance := &registerTestInstance{}

	regInstance.registerTestResources()
}

// TestMain is a "setup" function called by the testing framework when run the test
func TestMain(m *testing.M) {
	// Run setup function before starting the test suite for telephony_providers_edges_site_outbound_route package
	initTestResources()

	// Run the test suite for the telephony_providers_edges_site_outbound_route package
	m.Run()
}
