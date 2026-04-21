package case_management_stageplan

import (
	"context"
	"fmt"
	"sort"

	"github.com/mypurecloud/platform-client-sdk-go/v186/platformclientv2"
)

/*
The genesyscloud_case_management_stageplan_proxy.go file contains the proxy structures and methods that interact
with the Genesys Cloud SDK.
*/

const caseplanAPIVersionLatest = "latest"

// internalProxy holds a proxy instance that can be used throughout the package
var internalProxy *caseManagementStageplanProxy

type listStageplansForCaseplanFunc func(ctx context.Context, p *caseManagementStageplanProxy, caseplanID string) ([]platformclientv2.Stageplan, *platformclientv2.APIResponse, error)
type getCaseManagementStageplanFunc func(ctx context.Context, p *caseManagementStageplanProxy, caseplanID, stageplanID string) (*platformclientv2.Stageplan, *platformclientv2.APIResponse, error)
type patchCaseManagementStageplanFunc func(ctx context.Context, p *caseManagementStageplanProxy, caseplanID, stageplanID string, body platformclientv2.Stageplanupdate) (*platformclientv2.Stageplan, *platformclientv2.APIResponse, error)

type caseManagementStageplanProxy struct {
	clientConfig                     *platformclientv2.Configuration
	caseManagementApi                *platformclientv2.CaseManagementApi
	listStageplansForCaseplanAttr    listStageplansForCaseplanFunc
	getCaseManagementStageplanAttr   getCaseManagementStageplanFunc
	patchCaseManagementStageplanAttr patchCaseManagementStageplanFunc
}

func newCaseManagementStageplanProxy(clientConfig *platformclientv2.Configuration) *caseManagementStageplanProxy {
	api := platformclientv2.NewCaseManagementApiWithConfig(clientConfig)
	return &caseManagementStageplanProxy{
		clientConfig:                     clientConfig,
		caseManagementApi:                api,
		listStageplansForCaseplanAttr:    listStageplansForCaseplanFn,
		getCaseManagementStageplanAttr:   getCaseManagementStageplanFn,
		patchCaseManagementStageplanAttr: patchCaseManagementStageplanFn,
	}
}

func getCaseManagementStageplanProxy(clientConfig *platformclientv2.Configuration) *caseManagementStageplanProxy {
	if internalProxy == nil {
		internalProxy = newCaseManagementStageplanProxy(clientConfig)
	}
	return internalProxy
}

func (p *caseManagementStageplanProxy) listStageplansForCaseplan(ctx context.Context, caseplanID string) ([]platformclientv2.Stageplan, *platformclientv2.APIResponse, error) {
	return p.listStageplansForCaseplanAttr(ctx, p, caseplanID)
}

func (p *caseManagementStageplanProxy) getCaseManagementStageplan(ctx context.Context, caseplanID, stageplanID string) (*platformclientv2.Stageplan, *platformclientv2.APIResponse, error) {
	return p.getCaseManagementStageplanAttr(ctx, p, caseplanID, stageplanID)
}

func (p *caseManagementStageplanProxy) patchCaseManagementStageplan(ctx context.Context, caseplanID, stageplanID string, body platformclientv2.Stageplanupdate) (*platformclientv2.Stageplan, *platformclientv2.APIResponse, error) {
	return p.patchCaseManagementStageplanAttr(ctx, p, caseplanID, stageplanID, body)
}

func listStageplansForCaseplanFn(ctx context.Context, p *caseManagementStageplanProxy, caseplanID string) ([]platformclientv2.Stageplan, *platformclientv2.APIResponse, error) {
	var combined []platformclientv2.Stageplan
	after := ""
	var lastResp *platformclientv2.APIResponse
	for {
		listing, resp, err := p.caseManagementApi.GetCasemanagementCaseplanVersionStageplans(caseplanID, caseplanAPIVersionLatest, "", after, "100", nil)
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

func getCaseManagementStageplanFn(ctx context.Context, p *caseManagementStageplanProxy, caseplanID, stageplanID string) (*platformclientv2.Stageplan, *platformclientv2.APIResponse, error) {
	return p.caseManagementApi.GetCasemanagementCaseplanVersionStageplan(caseplanID, caseplanAPIVersionLatest, stageplanID, nil)
}

func patchCaseManagementStageplanFn(ctx context.Context, p *caseManagementStageplanProxy, caseplanID, stageplanID string, body platformclientv2.Stageplanupdate) (*platformclientv2.Stageplan, *platformclientv2.APIResponse, error) {
	return p.caseManagementApi.PatchCasemanagementCaseplanStageplan(caseplanID, stageplanID, body)
}

// resolveStageplanByOrdinal lists stageplans for the caseplan (version "latest"), sorts by name
// (platform defaults "Stage 1".."Stage 3"; identical timestamps make name the stable key), then returns
// the stage at index stageNumber-1 (stageNumber in 1..3).
func resolveStageplanByOrdinal(ctx context.Context, p *caseManagementStageplanProxy, caseplanID string, stageNumber int) (*platformclientv2.Stageplan, *platformclientv2.APIResponse, error) {
	if stageNumber < 1 || stageNumber > 3 {
		return nil, nil, fmt.Errorf("stage_number must be between 1 and 3, got %d", stageNumber)
	}
	stages, resp, err := p.listStageplansForCaseplan(ctx, caseplanID)
	if err != nil {
		return nil, resp, err
	}
	if len(stages) < stageNumber {
		return nil, resp, fmt.Errorf("caseplan %s has %d stageplan(s); expected at least %d", caseplanID, len(stages), stageNumber)
	}
	sortStageplansByName(stages)
	chosen := stages[stageNumber-1]
	if chosen.Id == nil || *chosen.Id == "" {
		return nil, resp, fmt.Errorf("resolved stageplan at position %d has no id", stageNumber)
	}
	return &chosen, resp, nil
}

// ResolveStageplanForCaseplanOrdinal lists stageplans under caseplan version "latest" and returns the stage at 1-based index (1..3).
func ResolveStageplanForCaseplanOrdinal(ctx context.Context, clientConfig *platformclientv2.Configuration, caseplanID string, stageNumber int) (*platformclientv2.Stageplan, *platformclientv2.APIResponse, error) {
	return resolveStageplanByOrdinal(ctx, getCaseManagementStageplanProxy(clientConfig), caseplanID, stageNumber)
}

func sortStageplansByName(stages []platformclientv2.Stageplan) {
	sort.Slice(stages, func(i, j int) bool {
		ni, nj := stringFromPtr(stages[i].Name), stringFromPtr(stages[j].Name)
		if ni != nj {
			return ni < nj
		}
		return stringFromPtr(stages[i].Id) < stringFromPtr(stages[j].Id)
	})
}

func stringFromPtr(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}
