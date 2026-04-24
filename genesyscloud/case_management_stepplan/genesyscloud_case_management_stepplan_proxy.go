package case_management_stepplan

import (
	"context"
	"fmt"
	"sort"

	"github.com/mypurecloud/platform-client-sdk-go/v186/platformclientv2"
)

const caseplanAPIVersionLatest = "latest"

var internalProxy *caseManagementStepplanProxy

type listStepplansForStageFunc func(ctx context.Context, p *caseManagementStepplanProxy, caseplanID, stageplanID string) ([]platformclientv2.Stepplan, *platformclientv2.APIResponse, error)
type getCaseManagementStepplanFunc func(ctx context.Context, p *caseManagementStepplanProxy, caseplanID, stageplanID, stepplanID string) (*platformclientv2.Stepplan, *platformclientv2.APIResponse, error)
type patchCaseManagementStepplanFunc func(ctx context.Context, p *caseManagementStepplanProxy, caseplanID, stageplanID, stepplanID string, body platformclientv2.Stepplanupdate) (*platformclientv2.Stepplan, *platformclientv2.APIResponse, error)

type caseManagementStepplanProxy struct {
	clientConfig                    *platformclientv2.Configuration
	caseManagementApi               *platformclientv2.CaseManagementApi
	listStepplansForStageAttr       listStepplansForStageFunc
	getCaseManagementStepplanAttr   getCaseManagementStepplanFunc
	patchCaseManagementStepplanAttr patchCaseManagementStepplanFunc
}

func newCaseManagementStepplanProxy(clientConfig *platformclientv2.Configuration) *caseManagementStepplanProxy {
	api := platformclientv2.NewCaseManagementApiWithConfig(clientConfig)
	return &caseManagementStepplanProxy{
		clientConfig:                    clientConfig,
		caseManagementApi:               api,
		listStepplansForStageAttr:       listStepplansForStageFn,
		getCaseManagementStepplanAttr:   getCaseManagementStepplanFn,
		patchCaseManagementStepplanAttr: patchCaseManagementStepplanFn,
	}
}

func getCaseManagementStepplanProxy(clientConfig *platformclientv2.Configuration) *caseManagementStepplanProxy {
	if internalProxy == nil {
		internalProxy = newCaseManagementStepplanProxy(clientConfig)
	}
	return internalProxy
}

func (p *caseManagementStepplanProxy) listStepplansForStage(ctx context.Context, caseplanID, stageplanID string) ([]platformclientv2.Stepplan, *platformclientv2.APIResponse, error) {
	return p.listStepplansForStageAttr(ctx, p, caseplanID, stageplanID)
}

func (p *caseManagementStepplanProxy) getCaseManagementStepplan(ctx context.Context, caseplanID, stageplanID, stepplanID string) (*platformclientv2.Stepplan, *platformclientv2.APIResponse, error) {
	return p.getCaseManagementStepplanAttr(ctx, p, caseplanID, stageplanID, stepplanID)
}

func (p *caseManagementStepplanProxy) patchCaseManagementStepplan(ctx context.Context, caseplanID, stageplanID, stepplanID string, body platformclientv2.Stepplanupdate) (*platformclientv2.Stepplan, *platformclientv2.APIResponse, error) {
	return p.patchCaseManagementStepplanAttr(ctx, p, caseplanID, stageplanID, stepplanID, body)
}

func listStepplansForStageFn(ctx context.Context, p *caseManagementStepplanProxy, caseplanID, stageplanID string) ([]platformclientv2.Stepplan, *platformclientv2.APIResponse, error) {
	var combined []platformclientv2.Stepplan
	after := ""
	var lastResp *platformclientv2.APIResponse
	for {
		listing, resp, err := p.caseManagementApi.GetCasemanagementCaseplanVersionStageplanStepplans(caseplanID, caseplanAPIVersionLatest, stageplanID, "", after, "100", nil)
		lastResp = resp
		if err != nil {
			return nil, resp, err
		}
		if listing == nil || listing.Entities == nil || len(*listing.Entities) == 0 {
			break
		}
		entities := *listing.Entities
		combined = append(combined, entities...)
		if len(entities) < 100 {
			break
		}
		last := entities[len(entities)-1]
		if last.Id == nil || *last.Id == "" {
			break
		}
		nextAfter := *last.Id
		if nextAfter == after {
			break
		}
		after = nextAfter
	}
	return combined, lastResp, nil
}

func getCaseManagementStepplanFn(ctx context.Context, p *caseManagementStepplanProxy, caseplanID, stageplanID, stepplanID string) (*platformclientv2.Stepplan, *platformclientv2.APIResponse, error) {
	return p.caseManagementApi.GetCasemanagementCaseplanVersionStageplanStepplan(caseplanID, caseplanAPIVersionLatest, stageplanID, stepplanID, nil)
}

func patchCaseManagementStepplanFn(ctx context.Context, p *caseManagementStepplanProxy, caseplanID, stageplanID, stepplanID string, body platformclientv2.Stepplanupdate) (*platformclientv2.Stepplan, *platformclientv2.APIResponse, error) {
	return p.caseManagementApi.PatchCasemanagementCaseplanStageplanStepplan(caseplanID, stageplanID, stepplanID, body)
}

// ResolveSingleStepplanForStage lists stepplans for a stage (version "latest"), expects exactly one, and returns it.
func ResolveSingleStepplanForStage(ctx context.Context, p *caseManagementStepplanProxy, caseplanID, stageplanID string) (*platformclientv2.Stepplan, *platformclientv2.APIResponse, error) {
	steps, resp, err := p.listStepplansForStage(ctx, caseplanID, stageplanID)
	if err != nil {
		return nil, resp, err
	}
	if len(steps) == 0 {
		return nil, resp, fmt.Errorf("no stepplans for caseplan %s stageplan %s", caseplanID, stageplanID)
	}
	sortStepplansByName(steps)
	if len(steps) != 1 {
		return nil, resp, fmt.Errorf("expected exactly 1 stepplan for stageplan %s, found %d", stageplanID, len(steps))
	}
	chosen := steps[0]
	if chosen.Id == nil || *chosen.Id == "" {
		return nil, resp, fmt.Errorf("resolved stepplan has no id")
	}
	return &chosen, resp, nil
}

func sortStepplansByName(steps []platformclientv2.Stepplan) {
	sort.Slice(steps, func(i, j int) bool {
		ni, nj := stringFromPtr(steps[i].Name), stringFromPtr(steps[j].Name)
		if ni != nj {
			return ni < nj
		}
		return stringFromPtr(steps[i].Id) < stringFromPtr(steps[j].Id)
	})
}

func stringFromPtr(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}
