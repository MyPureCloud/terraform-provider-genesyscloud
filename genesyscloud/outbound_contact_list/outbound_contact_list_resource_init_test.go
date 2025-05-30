package outbound_contact_list

import (
	gcloud "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud"
	obAttemptLimit "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/outbound_attempt_limit"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var providerDataSources map[string]*schema.Resource
var providerResources map[string]*schema.Resource

func initTestResources() {
	providerDataSources = make(map[string]*schema.Resource)
	providerResources = make(map[string]*schema.Resource)

	regInstance := &registerTestInstance{}

	regInstance.registerTestResources()
	regInstance.registerTestDataSources()
}

type registerTestInstance struct {
	resourceMapMutex   sync.RWMutex
	datasourceMapMutex sync.RWMutex
}

func (r *registerTestInstance) registerTestResources() {
	r.resourceMapMutex.Lock()
	defer r.resourceMapMutex.Unlock()

	providerResources[ResourceType] = ResourceOutboundContactList()
	providerResources[obAttemptLimit.ResourceType] = obAttemptLimit.ResourceOutboundAttemptLimit()
}

func (r *registerTestInstance) registerTestDataSources() {
	r.datasourceMapMutex.Lock()
	defer r.datasourceMapMutex.Unlock()

	providerDataSources[ResourceType] = DataSourceOutboundContactList()
	providerDataSources[obAttemptLimit.ResourceType] = obAttemptLimit.DataSourceOutboundAttemptLimit()
	providerDataSources["genesyscloud_auth_division_home"] = gcloud.DataSourceAuthDivisionHome()
	providerDataSources["genesyscloud_auth_division_home"] = gcloud.DataSourceAuthDivisionHome()
}

func TestMain(m *testing.M) {
	// Run setup function before starting the test suite
	initTestResources()

	// Run the test suite for outbound ruleset
	m.Run()
}
