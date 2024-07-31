package architect_flow

import (
	"context"
	"fmt"
	"log"
	rc "terraform-provider-genesyscloud/genesyscloud/resource_cache"

	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

var internalProxy *architectFlowProxy

type getArchitectFunc func(context.Context, *architectFlowProxy, string) (*platformclientv2.Flow, *platformclientv2.APIResponse, error)
type forceUnlockFlowFunc func(context.Context, *architectFlowProxy, string) (*platformclientv2.APIResponse, error)
type deleteArchitectFlowFunc func(context.Context, *architectFlowProxy, string) (*platformclientv2.APIResponse, error)
type createArchitectFlowJobsFunc func(context.Context, *architectFlowProxy) (*platformclientv2.Registerarchitectjobresponse, *platformclientv2.APIResponse, error)
type getArchitectFlowJobsFunc func(context.Context, *architectFlowProxy, string) (*platformclientv2.Architectjobstateresponse, *platformclientv2.APIResponse, error)
type getAllArchitectFlowsFunc func(context.Context, *architectFlowProxy) (*[]platformclientv2.Flow, *platformclientv2.APIResponse, error)

type architectFlowProxy struct {
	clientConfig *platformclientv2.Configuration
	api          *platformclientv2.ArchitectApi

	getArchitectFlowAttr        getArchitectFunc
	getAllArchitectFlowsAttr    getAllArchitectFlowsFunc
	forceUnlockFlowAttr         forceUnlockFlowFunc
	deleteArchitectFlowAttr     deleteArchitectFlowFunc
	createArchitectFlowJobsAttr createArchitectFlowJobsFunc
	getArchitectFlowJobsAttr    getArchitectFlowJobsFunc

	flowCache rc.CacheInterface[platformclientv2.Flow]
}

func newArchitectFlowProxy(clientConfig *platformclientv2.Configuration) *architectFlowProxy {
	api := platformclientv2.NewArchitectApiWithConfig(clientConfig)
	flowCache := rc.NewResourceCache[platformclientv2.Flow]()
	return &architectFlowProxy{
		clientConfig: clientConfig,
		api:          api,

		getArchitectFlowAttr:        getArchitectFlowFn,
		getAllArchitectFlowsAttr:    getAllArchitectFlowsFn,
		forceUnlockFlowAttr:         forceUnlockFlowFn,
		deleteArchitectFlowAttr:     deleteArchitectFlowFn,
		createArchitectFlowJobsAttr: createArchitectFlowJobsFn,
		getArchitectFlowJobsAttr:    getArchitectFlowJobsFn,
		flowCache:                   flowCache,
	}
}

func getArchitectFlowProxy(clientConfig *platformclientv2.Configuration) *architectFlowProxy {
	if internalProxy == nil {
		internalProxy = newArchitectFlowProxy(clientConfig)
	}
	return internalProxy
}

func (a *architectFlowProxy) GetFlow(ctx context.Context, id string) (*platformclientv2.Flow, *platformclientv2.APIResponse, error) {
	return a.getArchitectFlowAttr(ctx, a, id)
}

func (a *architectFlowProxy) ForceUnlockFlow(ctx context.Context, id string) (*platformclientv2.APIResponse, error) {
	return a.forceUnlockFlowAttr(ctx, a, id)
}

func (a *architectFlowProxy) DeleteFlow(ctx context.Context, id string) (*platformclientv2.APIResponse, error) {
	return a.deleteArchitectFlowAttr(ctx, a, id)
}

func (a *architectFlowProxy) CreateFlowsDeployJob(ctx context.Context) (*platformclientv2.Registerarchitectjobresponse, *platformclientv2.APIResponse, error) {
	return a.createArchitectFlowJobsAttr(ctx, a)
}

func (a *architectFlowProxy) GetFlowsDeployJob(ctx context.Context, jobId string) (*platformclientv2.Architectjobstateresponse, *platformclientv2.APIResponse, error) {
	return a.getArchitectFlowJobsAttr(ctx, a, jobId)
}

func (a *architectFlowProxy) GetAllFlows(ctx context.Context) (*[]platformclientv2.Flow, *platformclientv2.APIResponse, error) {
	return a.getAllArchitectFlowsAttr(ctx, a)
}

func getArchitectFlowFn(_ context.Context, p *architectFlowProxy, id string) (*platformclientv2.Flow, *platformclientv2.APIResponse, error) {
	flow := rc.GetCacheItem(p.flowCache, id)
	if flow != nil {
		return flow, nil, nil
	}
	return p.api.GetFlow(id, false)
}

func forceUnlockFlowFn(_ context.Context, p *architectFlowProxy, flowId string) (*platformclientv2.APIResponse, error) {
	log.Printf("Attempting to perform an unlock on flow: %s", flowId)
	_, resp, err := p.api.PostFlowsActionsUnlock(flowId)
	return resp, err
}

func deleteArchitectFlowFn(_ context.Context, p *architectFlowProxy, flowId string) (*platformclientv2.APIResponse, error) {
	return p.api.DeleteFlow(flowId)
}

func createArchitectFlowJobsFn(_ context.Context, p *architectFlowProxy) (*platformclientv2.Registerarchitectjobresponse, *platformclientv2.APIResponse, error) {
	return p.api.PostFlowsJobs()
}

func getArchitectFlowJobsFn(_ context.Context, p *architectFlowProxy, jobId string) (*platformclientv2.Architectjobstateresponse, *platformclientv2.APIResponse, error) {
	return p.api.GetFlowsJob(jobId, []string{"messages"})
}

func getAllArchitectFlowsFn(ctx context.Context, p *architectFlowProxy) (*[]platformclientv2.Flow, *platformclientv2.APIResponse, error) {
	const pageSize = 100
	var totalFlows []platformclientv2.Flow

	flows, resp, err := p.api.GetFlows(nil, 1, pageSize, "", "", nil, "", "", "", "", "", "", "", "", false, true, "", "", nil)
	if err != nil {
		return nil, resp, fmt.Errorf("Failed to get page of flows: %v %v", err, resp)
	}
	if flows.Entities == nil || len(*flows.Entities) == 0 {
		return &totalFlows, nil, nil
	}

	totalFlows = append(totalFlows, *flows.Entities...)

	for pageNum := 2; pageNum <= *flows.PageCount; pageNum++ {
		flows, resp, err := p.api.GetFlows(nil, pageNum, pageSize, "", "", nil, "", "", "", "", "", "", "", "", false, true, "", "", nil)
		if err != nil {
			return nil, resp, fmt.Errorf("failed to get page %d of flows: %v", pageNum, err)
		}
		if flows.Entities == nil || len(*flows.Entities) == 0 {
			break
		}
		totalFlows = append(totalFlows, *flows.Entities...)
	}

	for _, flow := range totalFlows {
		rc.SetCache(p.flowCache, *flow.Id, flow)
	}

	return &totalFlows, nil, nil
}
