package routing_queue

import (
	"context"
	"fmt"
	"log"
	rc "terraform-provider-genesyscloud/genesyscloud/resource_cache"

	"github.com/mypurecloud/platform-client-sdk-go/v146/platformclientv2"
)

/*
The genesyscloud_routing_queue_proxy.go file contains the proxy structures and methods that interact
with the Genesys Cloud SDK. We use composition here for each function on the proxy so individual functions can be stubbed
out during testing.
*/

var routingQueueCache = rc.NewResourceCache[platformclientv2.Queue]()
var internalProxy *RoutingQueueProxy

type GetAllRoutingQueuesFunc func(ctx context.Context, p *RoutingQueueProxy, name string) (*[]platformclientv2.Queue, *platformclientv2.APIResponse, error)
type createRoutingQueueFunc func(ctx context.Context, p *RoutingQueueProxy, createReq *platformclientv2.Createqueuerequest) (*platformclientv2.Queue, *platformclientv2.APIResponse, error)
type getRoutingQueueByIdFunc func(ctx context.Context, p *RoutingQueueProxy, queueId string, checkCache bool) (*platformclientv2.Queue, *platformclientv2.APIResponse, error)
type getRoutingQueueByNameFunc func(ctx context.Context, p *RoutingQueueProxy, name string) (string, *platformclientv2.APIResponse, bool, error)
type updateRoutingQueueFunc func(ctx context.Context, p *RoutingQueueProxy, queueId string, updateReq *platformclientv2.Queuerequest) (*platformclientv2.Queue, *platformclientv2.APIResponse, error)
type deleteRoutingQueueFunc func(ctx context.Context, p *RoutingQueueProxy, queueId string, forceDelete bool) (*platformclientv2.APIResponse, error)

type getAllRoutingQueueWrapupCodesFunc func(ctx context.Context, p *RoutingQueueProxy, queueId string) (*[]platformclientv2.Wrapupcode, *platformclientv2.APIResponse, error)
type createRoutingQueueWrapupCodeFunc func(ctx context.Context, p *RoutingQueueProxy, queueId string, body []platformclientv2.Wrapupcodereference) ([]platformclientv2.Wrapupcode, *platformclientv2.APIResponse, error)
type deleteRoutingQueueWrapupCodeFunc func(ctx context.Context, p *RoutingQueueProxy, queueId, codeId string) (*platformclientv2.APIResponse, error)

type addOrRemoveMembersFunc func(ctx context.Context, p *RoutingQueueProxy, queueId string, body []platformclientv2.Writableentity, delete bool) (*platformclientv2.APIResponse, error)
type updateRoutingQueueMemberFunc func(ctx context.Context, p *RoutingQueueProxy, queueId, userId string, body platformclientv2.Queuemember) (*platformclientv2.APIResponse, error)

// RoutingQueueProxy contains all the methods that call genesys cloud APIs.
type RoutingQueueProxy struct {
	clientConfig *platformclientv2.Configuration
	routingApi   *platformclientv2.RoutingApi

	GetAllRoutingQueuesAttr   GetAllRoutingQueuesFunc
	createRoutingQueueAttr    createRoutingQueueFunc
	getRoutingQueueByIdAttr   getRoutingQueueByIdFunc
	getRoutingQueueByNameAttr getRoutingQueueByNameFunc
	updateRoutingQueueAttr    updateRoutingQueueFunc
	deleteRoutingQueueAttr    deleteRoutingQueueFunc

	getAllRoutingQueueWrapupCodesAttr getAllRoutingQueueWrapupCodesFunc
	createRoutingQueueWrapupCodeAttr  createRoutingQueueWrapupCodeFunc
	deleteRoutingQueueWrapupCodeAttr  deleteRoutingQueueWrapupCodeFunc

	addOrRemoveMembersAttr       addOrRemoveMembersFunc
	updateRoutingQueueMemberAttr updateRoutingQueueMemberFunc

	RoutingQueueCache rc.CacheInterface[platformclientv2.Queue]
	wrapupCodeCache   rc.CacheInterface[platformclientv2.Wrapupcode]
}

// newRoutingQueuesProxy initializes the routing queue proxy with all the data needed to communicate with Genesys Cloud
func newRoutingQueuesProxy(clientConfig *platformclientv2.Configuration) *RoutingQueueProxy {
	api := platformclientv2.NewRoutingApiWithConfig(clientConfig)
	wrapupCodeCache := rc.NewResourceCache[platformclientv2.Wrapupcode]()

	return &RoutingQueueProxy{
		clientConfig: clientConfig,
		routingApi:   api,

		GetAllRoutingQueuesAttr:   GetAllRoutingQueuesFn,
		createRoutingQueueAttr:    createRoutingQueueFn,
		getRoutingQueueByIdAttr:   getRoutingQueueByIdFn,
		getRoutingQueueByNameAttr: getRoutingQueueByNameFn,
		updateRoutingQueueAttr:    updateRoutingQueueFn,
		deleteRoutingQueueAttr:    deleteRoutingQueueFn,

		getAllRoutingQueueWrapupCodesAttr: getAllRoutingQueueWrapupCodesFn,
		createRoutingQueueWrapupCodeAttr:  createRoutingQueueWrapupCodeFn,
		deleteRoutingQueueWrapupCodeAttr:  deleteRoutingQueueWrapupCodeFn,

		addOrRemoveMembersAttr:       addOrRemoveMembersFn,
		updateRoutingQueueMemberAttr: updateRoutingQueueMemberFn,

		RoutingQueueCache: routingQueueCache,
		wrapupCodeCache:   wrapupCodeCache,
	}
}

// GetRoutingQueueProxy returns an instance of our proxy
func GetRoutingQueueProxy(clientConfig *platformclientv2.Configuration) *RoutingQueueProxy {
	// continue with singleton approach if unit tests are running
	if isRoutingQueueUnitTestsActive() {
		return internalProxy
	}
	return newRoutingQueuesProxy(clientConfig)
}

func (p *RoutingQueueProxy) GetAllRoutingQueues(ctx context.Context, name string) (*[]platformclientv2.Queue, *platformclientv2.APIResponse, error) {
	return p.GetAllRoutingQueuesAttr(ctx, p, name)
}

func (p *RoutingQueueProxy) createRoutingQueue(ctx context.Context, createReq *platformclientv2.Createqueuerequest) (*platformclientv2.Queue, *platformclientv2.APIResponse, error) {
	return p.createRoutingQueueAttr(ctx, p, createReq)
}

func (p *RoutingQueueProxy) getRoutingQueueById(ctx context.Context, queueId string, checkCache bool) (*platformclientv2.Queue, *platformclientv2.APIResponse, error) {
	return p.getRoutingQueueByIdAttr(ctx, p, queueId, checkCache)
}

func (p *RoutingQueueProxy) getRoutingQueueByName(ctx context.Context, name string) (string, *platformclientv2.APIResponse, bool, error) {
	return p.getRoutingQueueByNameAttr(ctx, p, name)
}

func (p *RoutingQueueProxy) updateRoutingQueue(ctx context.Context, queueId string, updateReq *platformclientv2.Queuerequest) (*platformclientv2.Queue, *platformclientv2.APIResponse, error) {
	return p.updateRoutingQueueAttr(ctx, p, queueId, updateReq)
}

func (p *RoutingQueueProxy) deleteRoutingQueue(ctx context.Context, queueId string, forceDelete bool) (*platformclientv2.APIResponse, error) {
	return p.deleteRoutingQueueAttr(ctx, p, queueId, forceDelete)
}

func (p *RoutingQueueProxy) getAllRoutingQueueWrapupCodes(ctx context.Context, queueId string) (*[]platformclientv2.Wrapupcode, *platformclientv2.APIResponse, error) {
	return p.getAllRoutingQueueWrapupCodesAttr(ctx, p, queueId)
}

func (p *RoutingQueueProxy) createRoutingQueueWrapupCode(ctx context.Context, queueId string, body []platformclientv2.Wrapupcodereference) ([]platformclientv2.Wrapupcode, *platformclientv2.APIResponse, error) {
	return p.createRoutingQueueWrapupCodeAttr(ctx, p, queueId, body)
}

func (p *RoutingQueueProxy) deleteRoutingQueueWrapupCode(ctx context.Context, queueId, codeId string) (*platformclientv2.APIResponse, error) {
	return p.deleteRoutingQueueWrapupCodeAttr(ctx, p, queueId, codeId)
}

func (p *RoutingQueueProxy) addOrRemoveMembers(ctx context.Context, queueId string, body []platformclientv2.Writableentity, delete bool) (*platformclientv2.APIResponse, error) {
	return p.addOrRemoveMembersAttr(ctx, p, queueId, body, delete)
}

func (p *RoutingQueueProxy) updateRoutingQueueMember(ctx context.Context, queueId, userId string, body platformclientv2.Queuemember) (*platformclientv2.APIResponse, error) {
	return p.updateRoutingQueueMemberAttr(ctx, p, queueId, userId, body)
}

// GetAllRoutingQueuesFn is the implementation for retrieving all routing queues in Genesys Cloud
func GetAllRoutingQueuesFn(ctx context.Context, p *RoutingQueueProxy, name string) (*[]platformclientv2.Queue, *platformclientv2.APIResponse, error) {
	var allQueues []platformclientv2.Queue
	const pageSize = 100

	queues, resp, getErr := p.routingApi.GetRoutingQueues(1, pageSize, "", name, nil, nil, nil, "", false)
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
		queues, resp, getErr := p.routingApi.GetRoutingQueues(pageNum, pageSize, "", name, nil, nil, nil, "", false)
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

func createRoutingQueueFn(ctx context.Context, p *RoutingQueueProxy, createReq *platformclientv2.Createqueuerequest) (*platformclientv2.Queue, *platformclientv2.APIResponse, error) {
	return p.routingApi.PostRoutingQueues(*createReq)
}

// getRoutingQueueByIdFn is the implementation for retrieving a routing queues in Genesys Cloud
func getRoutingQueueByIdFn(ctx context.Context, p *RoutingQueueProxy, queueId string, checkCache bool) (*platformclientv2.Queue, *platformclientv2.APIResponse, error) {
	if checkCache {
		queue := rc.GetCacheItem(p.RoutingQueueCache, queueId)
		if queue != nil {
			return queue, nil, nil
		}
	}
	return p.routingApi.GetRoutingQueue(queueId)
}

func getRoutingQueueByNameFn(ctx context.Context, p *RoutingQueueProxy, name string) (string, *platformclientv2.APIResponse, bool, error) {
	queues, resp, err := p.GetAllRoutingQueues(ctx, name)
	if err != nil {
		return "", resp, false, err
	}

	if queues == nil || len(*queues) == 0 {
		return "", resp, true, fmt.Errorf("no routing queue found with name %s", name)
	}

	for _, queue := range *queues {
		if *queue.Name == name {
			log.Printf("Retrieved the routing queue id %s by name %s", *queue.Id, name)
			return *queue.Id, resp, false, nil
		}
	}
	return "", resp, true, fmt.Errorf("unable to find routing queue with name %s", name)
}

func updateRoutingQueueFn(ctx context.Context, p *RoutingQueueProxy, queueId string, updateReq *platformclientv2.Queuerequest) (*platformclientv2.Queue, *platformclientv2.APIResponse, error) {
	return p.routingApi.PutRoutingQueue(queueId, *updateReq)
}

func deleteRoutingQueueFn(ctx context.Context, p *RoutingQueueProxy, queueID string, forceDelete bool) (*platformclientv2.APIResponse, error) {
	resp, err := p.routingApi.DeleteRoutingQueue(queueID, forceDelete)
	if err != nil {
		return resp, err
	}
	rc.DeleteCacheItem(p.RoutingQueueCache, queueID)
	return resp, nil
}

func getAllRoutingQueueWrapupCodesFn(ctx context.Context, p *RoutingQueueProxy, queueId string) (*[]platformclientv2.Wrapupcode, *platformclientv2.APIResponse, error) {
	var allWrapupcodes []platformclientv2.Wrapupcode
	const pageSize = 100

	wrapupcodes, apiResponse, err := p.routingApi.GetRoutingQueueWrapupcodes(queueId, pageSize, 1)
	if err != nil {
		return nil, apiResponse, fmt.Errorf("failed to get routing wrapupcode : %v", err)
	}

	if wrapupcodes.Total != nil {
		if rc.GetCacheSize(p.wrapupCodeCache) == *wrapupcodes.Total && rc.GetCacheSize(p.wrapupCodeCache) != 0 {
			return rc.GetCache(p.wrapupCodeCache), nil, nil
		} else if rc.GetCacheSize(p.wrapupCodeCache) != *wrapupcodes.Total && rc.GetCacheSize(p.wrapupCodeCache) != 0 {
			// The cache is populated but not with the right data, clear the cache so it can be re populated
			p.wrapupCodeCache = rc.NewResourceCache[platformclientv2.Wrapupcode]()
		}
	}

	if wrapupcodes == nil || wrapupcodes.Entities == nil || len(*wrapupcodes.Entities) == 0 {
		return &allWrapupcodes, apiResponse, nil
	}

	allWrapupcodes = append(allWrapupcodes, *wrapupcodes.Entities...)

	for pageNum := 2; pageNum <= *wrapupcodes.PageCount; pageNum++ {
		wrapupcodes, apiResponse, err := p.routingApi.GetRoutingQueueWrapupcodes(queueId, pageSize, pageNum)
		if err != nil {
			return nil, apiResponse, fmt.Errorf("failed to get routing wrapupcode : %v", err)
		}

		if wrapupcodes == nil || wrapupcodes.Entities == nil || len(*wrapupcodes.Entities) == 0 {
			break
		}

		allWrapupcodes = append(allWrapupcodes, *wrapupcodes.Entities...)
	}

	// Cache the routing wrapupcodes resource into the p.routingWrapupcodesCache for later use
	for _, wrapupcode := range allWrapupcodes {
		rc.SetCache(p.wrapupCodeCache, *wrapupcode.Id, wrapupcode)
	}

	return &allWrapupcodes, apiResponse, nil
}

func createRoutingQueueWrapupCodeFn(ctx context.Context, p *RoutingQueueProxy, queueId string, body []platformclientv2.Wrapupcodereference) ([]platformclientv2.Wrapupcode, *platformclientv2.APIResponse, error) {
	return p.routingApi.PostRoutingQueueWrapupcodes(queueId, body)
}

func deleteRoutingQueueWrapupCodeFn(ctx context.Context, p *RoutingQueueProxy, queueId, codeId string) (*platformclientv2.APIResponse, error) {
	return p.routingApi.DeleteRoutingQueueWrapupcode(queueId, codeId)
}

func addOrRemoveMembersFn(ctx context.Context, p *RoutingQueueProxy, queueId string, body []platformclientv2.Writableentity, delete bool) (*platformclientv2.APIResponse, error) {
	return p.routingApi.PostRoutingQueueMembers(queueId, body, delete)
}

func updateRoutingQueueMemberFn(ctx context.Context, p *RoutingQueueProxy, queueId, userId string, body platformclientv2.Queuemember) (*platformclientv2.APIResponse, error) {
	return p.routingApi.PatchRoutingQueueMember(queueId, userId, body)
}
