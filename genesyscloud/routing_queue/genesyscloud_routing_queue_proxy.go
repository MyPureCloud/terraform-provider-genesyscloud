package routing_queue

import (
	"context"
	"fmt"
	rc "terraform-provider-genesyscloud/genesyscloud/resource_cache"

	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

/*
The genesyscloud_routing_queue_proxy.go file contains the proxy structures and methods that interact
with the Genesys Cloud SDK. We use composition here for each function on the proxy so individual functions can be stubbed
out during testing.
*/

// internalProxy holds a proxy instance that can be used throughout the package
var internalProxy *RoutingQueueProxy

// Type definitions for each func on our proxy so we can easily mock them out later
type getAllRoutingQueuesFunc func(ctx context.Context, p *RoutingQueueProxy) (*[]platformclientv2.Queue, *platformclientv2.APIResponse, error)
type getRoutingQueueByIdFunc func(ctx context.Context, p *RoutingQueueProxy, queueId string) (*platformclientv2.Queue, *platformclientv2.APIResponse, error)
type getRoutingQueueWrapupCodeIdsFunc func(ctx context.Context, p *RoutingQueueProxy, queueId string) ([]string, *platformclientv2.APIResponse, error)

// RoutingQueueProxy contains all the methods that call genesys cloud APIs.
type RoutingQueueProxy struct {
	clientConfig                     *platformclientv2.Configuration
	routingApi                       *platformclientv2.RoutingApi
	getAllRoutingQueuesAttr          getAllRoutingQueuesFunc
	getRoutingQueueByIdAttr          getRoutingQueueByIdFunc
	getRoutingQueueWrapupCodeIdsAttr getRoutingQueueWrapupCodeIdsFunc
	RoutingQueueCache                rc.CacheInterface[platformclientv2.Queue]
}

// newRoutingQueuesProxy initializes the routing queue proxy with all the data needed to communicate with Genesys Cloud
func newRoutingQueuesProxy(clientConfig *platformclientv2.Configuration) *RoutingQueueProxy {
	api := platformclientv2.NewRoutingApiWithConfig(clientConfig)
	routingQueueCache := rc.NewResourceCache[platformclientv2.Queue]()

	return &RoutingQueueProxy{
		clientConfig:                     clientConfig,
		routingApi:                       api,
		getAllRoutingQueuesAttr:          getAllRoutingQueuesFn,
		getRoutingQueueByIdAttr:          getRoutingQueueByIdFn,
		getRoutingQueueWrapupCodeIdsAttr: getRoutingQueueWrapupCodeIdsFn,
		RoutingQueueCache:                routingQueueCache,
	}
}

// GetRoutingQueueProxy acts as a singleton to for the internalProxy.  It also ensures
// that we can still proxy our tests by directly setting internalProxy package variable
func GetRoutingQueueProxy(clientConfig *platformclientv2.Configuration) *RoutingQueueProxy {
	if internalProxy == nil {
		internalProxy = newRoutingQueuesProxy(clientConfig)
	}
	return internalProxy
}

// GetAllRoutingQueues retrieves all Genesys Cloud routing queues
func (p *RoutingQueueProxy) GetAllRoutingQueues(ctx context.Context) (*[]platformclientv2.Queue, *platformclientv2.APIResponse, error) {
	return p.getAllRoutingQueuesAttr(ctx, p)
}

// getRoutingQueueById returns a single Genesys Cloud Routing Queue by ID
func (p *RoutingQueueProxy) getRoutingQueueById(ctx context.Context, queueId string) (*platformclientv2.Queue, *platformclientv2.APIResponse, error) {
	return p.getRoutingQueueByIdAttr(ctx, p, queueId)
}

// getRoutingQueueWrapupCodeIds returns a list of routing queue wrapup code ids
func (p *RoutingQueueProxy) getRoutingQueueWrapupCodeIds(ctx context.Context, queueId string) ([]string, *platformclientv2.APIResponse, error) {
	return p.getRoutingQueueWrapupCodeIdsAttr(ctx, p, queueId)
}

// getAllRoutingQueuesFn is the implementation for retrieving all routing queues in Genesys Cloud
func getAllRoutingQueuesFn(ctx context.Context, p *RoutingQueueProxy) (*[]platformclientv2.Queue, *platformclientv2.APIResponse, error) {
	var allQueues []platformclientv2.Queue
	const pageSize = 100

	queues, resp, getErr := p.routingApi.GetRoutingQueues(1, pageSize, "", "", nil, nil, nil, "", false)
	if getErr != nil {
		return nil, resp, fmt.Errorf("failed to get first page of queues: %v", getErr)
	}

	// Check if the routing queue cache is populated with all the data, if it is, return that instead
	// If the size of the cache is the same as the total number of queues, the cache is up-to-date
	if rc.GetCacheSize(p.RoutingQueueCache) == *queues.Total && rc.GetCacheSize(p.RoutingQueueCache) != 0 {
		return rc.GetCache(p.RoutingQueueCache), nil, nil
	} else if rc.GetCacheSize(p.RoutingQueueCache) != *queues.Total && rc.GetCacheSize(p.RoutingQueueCache) != 0 {
		// The cache is populated but not with the right data, clear the cache so it can be re populated
		p.RoutingQueueCache = rc.NewResourceCache[platformclientv2.Queue]()
	}

	if queues.Entities == nil || len(*queues.Entities) == 0 {
		return &allQueues, resp, nil
	}

	allQueues = append(allQueues, *queues.Entities...)

	for pageNum := 2; pageNum <= *queues.PageCount; pageNum++ {
		queues, resp, getErr := p.routingApi.GetRoutingQueues(pageNum, pageSize, "", "", nil, nil, nil, "", false)
		if getErr != nil {
			return nil, resp, fmt.Errorf("failed to get page of queues: %v", getErr)
		}

		if queues.Entities == nil || len(*queues.Entities) == 0 {
			break
		}

		allQueues = append(allQueues, *queues.Entities...)
	}

	for _, queue := range allQueues {
		rc.SetCache(p.RoutingQueueCache, *queue.Id, queue)
	}

	return &allQueues, resp, nil
}

// getRoutingQueueByIdFn is the implementation for retrieving a routing queues in Genesys Cloud
func getRoutingQueueByIdFn(ctx context.Context, p *RoutingQueueProxy, queueId string) (*platformclientv2.Queue, *platformclientv2.APIResponse, error) {
	queue := rc.GetCacheItem(p.RoutingQueueCache, queueId)
	if queue != nil {
		return queue, nil, nil
	}

	queue, resp, err := p.routingApi.GetRoutingQueue(queueId)
	if err != nil {
		return nil, resp, fmt.Errorf("failed to retrieve routing queue by id %s: %s", queueId, err)
	}

	return queue, resp, nil
}

func getRoutingQueueWrapupCodeIdsFn(ctx context.Context, p *RoutingQueueProxy, queueId string) ([]string, *platformclientv2.APIResponse, error) {
	var codeIds []string
	const pageSize = 100

	codes, resp, err := p.routingApi.GetRoutingQueueWrapupcodes(queueId, pageSize, 1)
	if err != nil {
		return nil, resp, fmt.Errorf("failed to first page of wrapup codes for queue %s: %s", queueId, err)
	}
	if codes == nil || codes.Entities == nil || len(*codes.Entities) == 0 {
		return codeIds, resp, nil
	}

	for _, code := range *codes.Entities {
		codeIds = append(codeIds, *code.Id)
	}

	for pageNum := 2; pageNum <= *codes.PageCount; pageNum++ {
		codes, resp, err := p.routingApi.GetRoutingQueueWrapupcodes(queueId, pageSize, pageNum)
		if err != nil {
			return nil, resp, fmt.Errorf("failed to page of wrapup codes for queue %s: %s", queueId, err)
		}
		if codes == nil || codes.Entities != nil || len(*codes.Entities) == 0 {
			break
		}
		for _, code := range *codes.Entities {
			codeIds = append(codeIds, *code.Id)
		}
	}

	return codeIds, resp, nil
}
