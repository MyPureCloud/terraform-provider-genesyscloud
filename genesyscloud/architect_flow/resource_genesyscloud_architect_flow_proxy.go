package architect_flow

import (
	"context"
	"fmt"
	"github.com/mypurecloud/platform-client-sdk-go/v125/platformclientv2"
	"log"
)

var internalProxy *architectFlowProxy

type getArchitectFunc func(context.Context, *architectFlowProxy, string) (*platformclientv2.Flow, *platformclientv2.APIResponse, error)
type forceUnlockFlowFunc func(context.Context, *architectFlowProxy, string) error
type deleteArchitectFlowFunc func(context.Context, *architectFlowProxy, string) (*platformclientv2.APIResponse, error)
type createArchitectFlowJobsFunc func(context.Context, *architectFlowProxy) (*platformclientv2.Registerarchitectjobresponse, *platformclientv2.APIResponse, error)
type getArchitectFlowJobsFunc func(context.Context, *architectFlowProxy, string) (*platformclientv2.Architectjobstateresponse, *platformclientv2.APIResponse, error)
type getAllArchitectFlowsFunc func(context.Context, *architectFlowProxy) (*[]platformclientv2.Flow, error)

type architectFlowProxy struct {
	clientConfig *platformclientv2.Configuration
	api          *platformclientv2.ArchitectApi

	getArchitectFlowAttr        getArchitectFunc
	getAllArchitectFlowsAttr    getAllArchitectFlowsFunc
	forceUnlockFlowAttr         forceUnlockFlowFunc
	deleteArchitectFlowAttr     deleteArchitectFlowFunc
	createArchitectFlowJobsAttr createArchitectFlowJobsFunc
	getArchitectFlowJobsAttr    getArchitectFlowJobsFunc
}

func newArchitectFlowProxy(clientConfig *platformclientv2.Configuration) *architectFlowProxy {
	api := platformclientv2.NewArchitectApiWithConfig(clientConfig)
	return &architectFlowProxy{
		clientConfig: clientConfig,
		api:          api,

		getArchitectFlowAttr:        getArchitectFlowFn,
		getAllArchitectFlowsAttr:    getAllArchitectFlowsFn,
		forceUnlockFlowAttr:         forceUnlockFlowFn,
		deleteArchitectFlowAttr:     deleteArchitectFlowFn,
		createArchitectFlowJobsAttr: createArchitectFlowJobsFn,
		getArchitectFlowJobsAttr:    getArchitectFlowJobsFn,
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

func (a *architectFlowProxy) ForceUnlockFlow(ctx context.Context, id string) error {
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

func (a *architectFlowProxy) GetAllFlows(ctx context.Context) (*[]platformclientv2.Flow, error) {
	return a.getAllArchitectFlowsAttr(ctx, a)
}

func getArchitectFlowFn(ctx context.Context, p *architectFlowProxy, id string) (*platformclientv2.Flow, *platformclientv2.APIResponse, error) {
	return p.api.GetFlow(id, false)
}

func forceUnlockFlowFn(ctx context.Context, p *architectFlowProxy, flowId string) error {
	log.Printf("Attempting to perform an unlock on flow: %s", flowId)
	_, _, err := p.api.PostFlowsActionsUnlock(flowId)

	if err != nil {
		return err
	}
	return nil
}

func deleteArchitectFlowFn(ctx context.Context, p *architectFlowProxy, flowId string) (*platformclientv2.APIResponse, error) {
	return p.api.DeleteFlow(flowId)
}

func createArchitectFlowJobsFn(ctx context.Context, p *architectFlowProxy) (*platformclientv2.Registerarchitectjobresponse, *platformclientv2.APIResponse, error) {
	return p.api.PostFlowsJobs()
}

func getArchitectFlowJobsFn(ctx context.Context, p *architectFlowProxy, jobId string) (*platformclientv2.Architectjobstateresponse, *platformclientv2.APIResponse, error) {
	return p.api.GetFlowsJob(jobId, []string{"messages"})
}

func getAllArchitectFlowsFn(ctx context.Context, p *architectFlowProxy) (*[]platformclientv2.Flow, error) {
	const pageSize = 100
	var totalFlows []platformclientv2.Flow

	flows, _, err := p.api.GetFlows(nil, 1, pageSize, "", "", nil, "", "", "", "", "", "", "", "", false, true, "", "", nil)
	if err != nil {
		return nil, fmt.Errorf("Failed to get page of flows: %v", err)
	}

	for _, flow := range *flows.Entities {
		totalFlows = append(totalFlows, flow)
	}

	for pageNum := 2; pageNum <= *flows.PageCount; pageNum++ {
		flows, _, err := p.api.GetFlows(nil, pageNum, pageSize, "", "", nil, "", "", "", "", "", "", "", "", false, true, "", "", nil)
		if err != nil {
			return nil, fmt.Errorf("Failed to get page %d of flows: %v", pageNum, err)
		}
		for _, flow := range *flows.Entities {
			totalFlows = append(totalFlows, flow)
		}
	}

	return &totalFlows, nil
}
