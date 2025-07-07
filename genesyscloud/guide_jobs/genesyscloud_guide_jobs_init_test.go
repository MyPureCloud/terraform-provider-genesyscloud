package guide_jobs

import (
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/guide"
)

/*
The genesyscloud_guide_jobs_init_test.go file is used to initialize the data sources and resources used in testing the guide_jobs resource
*/

var providerResources map[string]*schema.Resource

type registerTestInstance struct {
	resourceMapMutex sync.RWMutex
}

func (r *registerTestInstance) registerTestResources() {
	r.resourceMapMutex.Lock()
	defer r.resourceMapMutex.Unlock()

	providerResources[ResourceType] = ResourceGuideJobs()
	providerResources[guide.ResourceType] = guide.ResourceGuide()

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
