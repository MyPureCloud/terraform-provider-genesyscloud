package routing_queue

import (
	"context"
	"fmt"
	"github.com/mypurecloud/platform-client-sdk-go/v125/platformclientv2"
	"log"
	rc "terraform-provider-genesyscloud/genesyscloud/resource_cache"
)

/*
The genesyscloud_routing_queue_proxy.go file contains the proxy structures and methods that interact
with the Genesys Cloud SDK. We use composition here for each function on the proxy so individual functions can be stubbed
out during testing.
*/

// internalProxy holds a proxy instance that can be used throughout the package
var internalProxy *routingQueueProxy

// Type definitions for each func on our proxy so we can easily mock them out later
type getAllRoutingQueuesFunc func(ctx context.Context, p *routingQueueProxy) (*[]platformclientv2.Queue, *platformclientv2.APIResponse, error)
type getRoutingQueueByIdFunc func(ctx context.Context, p *routingQueueProxy, queueId string) (*platformclientv2.Queue, *platformclientv2.APIResponse, error)

// routingQueueProxy contains all the methods that call genesys cloud APIs.
type routingQueueProxy struct {
	clientConfig            *platformclientv2.Configuration
	routingApi              *platformclientv2.RoutingApi
	getAllRoutingQueuesAttr getAllRoutingQueuesFunc
	getRoutingQueueByIdAttr getRoutingQueueByIdFunc
	routingQueueCache       rc.CacheInterface[platformclientv2.Queue]
}

// newRoutingQueuesProxy initializes the routing queue proxy with all the data needed to communicate with Genesys Cloud
func newRoutingQueuesProxy(clientConfig *platformclientv2.Configuration) *routingQueueProxy {
	api := platformclientv2.NewRoutingApiWithConfig(clientConfig)
	routingQueueCache := rc.NewResourceCache[platformclientv2.Queue]()

	return &routingQueueProxy{
		clientConfig:            clientConfig,
		routingApi:              api,
		getAllRoutingQueuesAttr: getAllRoutingQueuesFn,
		getRoutingQueueByIdAttr: getRoutingQueueByIdFn,
		routingQueueCache:       routingQueueCache,
	}
}

// getRoutingQueueProxy acts as a singleton to for the internalProxy.  It also ensures
// that we can still proxy our tests by directly setting internalProxy package variable
func getRoutingQueueProxy(clientConfig *platformclientv2.Configuration) *routingQueueProxy {
	if internalProxy == nil {
		internalProxy = newRoutingQueuesProxy(clientConfig)
	}
	return internalProxy
}

// getAllRoutingQueues retrieves all Genesys Cloud routing queues
func (p *routingQueueProxy) getAllRoutingQueues(ctx context.Context) (*[]platformclientv2.Queue, *platformclientv2.APIResponse, error) {
	return p.getAllRoutingQueuesAttr(ctx, p)
}

// getArchitectGrammarById returns a single Genesys Cloud Architect Grammar by ID
func (p *routingQueueProxy) getRoutingQueueById(ctx context.Context, queueId string) (*platformclientv2.Queue, *platformclientv2.APIResponse, error) {
	return p.getRoutingQueueByIdAttr(ctx, p, queueId)
}

// getAllRoutingQueuesFn is the implementation for retrieving all routing queues in Genesys Cloud
func getAllRoutingQueuesFn(ctx context.Context, p *routingQueueProxy) (*[]platformclientv2.Queue, *platformclientv2.APIResponse, error) {
	var allQueues []platformclientv2.Queue
	const pageSize = 100

	log.Println("Api call")
	queues, resp, getErr := p.routingApi.GetRoutingQueues(1, pageSize, "", "", nil, nil, nil, false)
	if getErr != nil {
		return nil, resp, fmt.Errorf("failed to get first page of queues: %v", getErr)
	}
	if queues.Entities == nil || len(*queues.Entities) == 0 {
		return &allQueues, resp, nil
	}

	allQueues = append(allQueues, *queues.Entities...)

	for pageNum := 2; pageNum <= *queues.PageCount; pageNum++ {
		log.Println("Api call")
		queues, resp, getErr := p.routingApi.GetRoutingQueues(pageNum, pageSize, "", "", nil, nil, nil, false)
		if getErr != nil {
			return nil, resp, fmt.Errorf("failed to get page of queues: %v", getErr)
		}

		if queues.Entities == nil || len(*queues.Entities) == 0 {
			break
		}

		allQueues = append(allQueues, *queues.Entities...)
	}

	for _, queue := range allQueues {
		rc.SetCache(p.routingQueueCache, *queue.Id, queue)
	}

	return &allQueues, resp, nil
}

func getRoutingQueueByIdFn(ctx context.Context, p *routingQueueProxy, queueId string) (*platformclientv2.Queue, *platformclientv2.APIResponse, error) {
	queue := rc.GetCache(p.routingQueueCache, queueId)
	if queue != nil {
		return queue, nil, nil
	}

	log.Println("Api call")
	queue, resp, err := p.routingApi.GetRoutingQueue(queueId)
	if err != nil {
		return nil, resp, fmt.Errorf("failed to retrieve routing queue by id %s: %s", queueId, err)
	}

	return queue, resp, nil
}
