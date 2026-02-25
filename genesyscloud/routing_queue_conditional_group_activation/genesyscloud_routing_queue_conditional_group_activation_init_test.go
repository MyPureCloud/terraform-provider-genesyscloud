package routing_queue_conditional_group_activation

import (
	"sync"
	"testing"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/group"
	routingQueue "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/routing_queue"
	routingSkillGroup "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/routing_skill_group"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/user"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var providerResources map[string]*schema.Resource

type registerTestInstance struct {
	resourceMapMutex sync.RWMutex
}

func (r *registerTestInstance) registerTestResources() {
	r.resourceMapMutex.Lock()
	defer r.resourceMapMutex.Unlock()

	providerResources[ResourceType] = ResourceRoutingQueueConditionalGroupActivation()
	providerResources[routingQueue.ResourceType] = routingQueue.ResourceRoutingQueue()
	providerResources[routingSkillGroup.ResourceType] = routingSkillGroup.ResourceRoutingSkillGroup()
	providerResources[user.ResourceType] = user.ResourceUser()
	providerResources[group.ResourceType] = group.ResourceGroup()
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
