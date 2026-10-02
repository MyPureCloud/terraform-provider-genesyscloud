package integration_config

import (
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/integration"
)

/*
The genesyscloud_integration_config_init_test.go file is used to initialize the resources
used in testing the integration_config resource.
*/

var providerResources map[string]*schema.Resource

type registerTestInstance struct {
	resourceMapMutex sync.RWMutex
}

func (r *registerTestInstance) registerTestResources() {
	r.resourceMapMutex.Lock()
	defer r.resourceMapMutex.Unlock()

	providerResources[ResourceType] = ResourceIntegrationConfig()
	providerResources[integration.ResourceType] = integration.ResourceIntegration()
}

func initTestResources() {
	providerResources = make(map[string]*schema.Resource)

	regInstance := &registerTestInstance{}
	regInstance.registerTestResources()
}

func TestMain(m *testing.M) {
	initTestResources()
	m.Run()
}
