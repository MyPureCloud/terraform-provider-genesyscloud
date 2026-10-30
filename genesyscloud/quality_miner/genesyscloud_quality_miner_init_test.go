package quality_miner

import (
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

/*
The genesyscloud_quality_miner_init_test.go file is used to initialize the data sources and resources
used in testing the quality_miner resource.
*/

var providerDataSources map[string]*schema.Resource

var providerResources map[string]*schema.Resource

type registerTestInstance struct {
	resourceMapMutex   sync.RWMutex
	dataSourceMapMutex sync.RWMutex
}

func (r *registerTestInstance) registerTestResources() {
	r.resourceMapMutex.Lock()
	defer r.resourceMapMutex.Unlock()

	providerResources[ResourceType] = ResourceQualityMiner()
}

func (r *registerTestInstance) registerTestDataSources() {
	r.dataSourceMapMutex.Lock()
	defer r.dataSourceMapMutex.Unlock()

	providerDataSources[ResourceType] = DataSourceQualityMiner()
}

func initTestResources() {
	providerDataSources = make(map[string]*schema.Resource)
	providerResources = make(map[string]*schema.Resource)

	regInstance := &registerTestInstance{}

	regInstance.registerTestResources()
	regInstance.registerTestDataSources()
}

func TestMain(m *testing.M) {
	initTestResources()
	m.Run()
}
