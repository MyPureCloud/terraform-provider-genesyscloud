package telephony_providers_edges_edge_group

import (
	"context"

	"github.com/mypurecloud/platform-client-sdk-go/v121/platformclientv2"
)

var internalProxy *edgeGroupProxy

type getEdgeGroupFunc func(ctx context.Context, p *edgeGroupProxy, edgeGroupId string) (*platformclientv2.Edgegroup, *platformclientv2.APIResponse, error)
type deleteEdgeGroupFunc func(ctx context.Context, p *edgeGroupProxy, edgeGroupId string) (*platformclientv2.APIResponse, error)
type putEdgeGroupFunc func(ctx context.Context, p *edgeGroupProxy, edgeGroupId string, body platformclientv2.Edgegroup) (*platformclientv2.Edgegroup, *platformclientv2.APIResponse, error)
type postEdgeGroupFunc func(ctx context.Context, p *edgeGroupProxy, body platformclientv2.Edgegroup) (*platformclientv2.Edgegroup, *platformclientv2.APIResponse, error)
type getAllEdgeGroupsFunc func(ctx context.Context, p *edgeGroupProxy) (*[]platformclientv2.Edgegroup, error)

type edgeGroupProxy struct {
	clientConfig *platformclientv2.Configuration
	edgesApi     *platformclientv2.TelephonyProvidersEdgeApi

	getEdgeGroupAttr     getEdgeGroupFunc
	deleteEdgeGroupAttr  deleteEdgeGroupFunc
	putEdgeGroupAttr     putEdgeGroupFunc
	postEdgeGroupAttr    postEdgeGroupFunc
	getAllEdgeGroupsAttr getAllEdgeGroupsFunc
}

func newEdgeGroupProxy(clientConfig *platformclientv2.Configuration) *edgeGroupProxy {
	edgesApi := platformclientv2.NewTelephonyProvidersEdgeApiWithConfig(clientConfig)

	return &edgeGroupProxy{
		clientConfig: clientConfig,
		edgesApi:     edgesApi,

		getEdgeGroupAttr:     getEdgeGroupFn,
		deleteEdgeGroupAttr:  deleteEdgeGroupFn,
		putEdgeGroupAttr:     putEdgeGroupFn,
		postEdgeGroupAttr:    postEdgeGroupFn,
		getAllEdgeGroupsAttr: getAllEdgeGroupsFn,
	}
}

func getEdgeGroupProxy(clientConfig *platformclientv2.Configuration) *edgeGroupProxy {
	if internalProxy == nil {
		internalProxy = newEdgeGroupProxy(clientConfig)
	}
	return internalProxy
}

func (p *edgeGroupProxy) getEdgeGroup(ctx context.Context, edgeGroupId string) (*platformclientv2.Edgegroup, *platformclientv2.APIResponse, error) {
	return p.getEdgeGroupAttr(ctx, p, edgeGroupId)
}

func (p *edgeGroupProxy) deleteEdgeGroup(ctx context.Context, edgeGroupId string) (*platformclientv2.APIResponse, error) {
	return p.deleteEdgeGroupAttr(ctx, p, edgeGroupId)
}

func (p *edgeGroupProxy) putEdgeGroup(ctx context.Context, edgeGroupId string, body platformclientv2.Edgegroup) (*platformclientv2.Edgegroup, *platformclientv2.APIResponse, error) {
	return p.putEdgeGroupAttr(ctx, p, edgeGroupId, body)
}

func (p *edgeGroupProxy) postEdgeGroup(ctx context.Context, body platformclientv2.Edgegroup) (*platformclientv2.Edgegroup, *platformclientv2.APIResponse, error) {
	return p.postEdgeGroupAttr(ctx, p, body)
}

func (p *edgeGroupProxy) getAllEdgeGroups(ctx context.Context) (*[]platformclientv2.Edgegroup, error) {
	return p.getAllEdgeGroupsAttr(ctx, p)
}

func getEdgeGroupFn(ctx context.Context, p *edgeGroupProxy, edgeGroupId string) (*platformclientv2.Edgegroup, *platformclientv2.APIResponse, error) {
	edgeGroup, resp, err := p.edgesApi.GetTelephonyProvidersEdgesEdgegroup(edgeGroupId, nil)
	if err != nil {
		return nil, resp, err
	}
	return edgeGroup, resp, nil
}

func deleteEdgeGroupFn(ctx context.Context, p *edgeGroupProxy, edgeGroupId string) (*platformclientv2.APIResponse, error) {
	resp, err := p.edgesApi.DeleteTelephonyProvidersEdgesEdgegroup(edgeGroupId)
	return resp, err
}

func putEdgeGroupFn(ctx context.Context, p *edgeGroupProxy, edgeGroupId string, body platformclientv2.Edgegroup) (*platformclientv2.Edgegroup, *platformclientv2.APIResponse, error) {
	edgeGroup, resp, err := p.edgesApi.PutTelephonyProvidersEdgesEdgegroup(edgeGroupId, body)
	if err != nil {
		return nil, resp, err
	}
	return edgeGroup, resp, nil
}

func postEdgeGroupFn(ctx context.Context, p *edgeGroupProxy, body platformclientv2.Edgegroup) (*platformclientv2.Edgegroup, *platformclientv2.APIResponse, error) {
	edgeGroup, resp, err := p.edgesApi.PostTelephonyProvidersEdgesEdgegroups(body)
	if err != nil {
		return nil, resp, err
	}
	return edgeGroup, resp, nil
}

func getAllEdgeGroupsFn(ctx context.Context, p *edgeGroupProxy) (*[]platformclientv2.Edgegroup, error) {
	const pageSize = 100
	var allEdgeGroups []platformclientv2.Edgegroup
	for pageNum := 1; ; pageNum++ {

		edgeGroups, _, err := p.edgesApi.GetTelephonyProvidersEdgesEdgegroups(pageSize, pageNum, "", "", false)
		if err != nil {
			return nil, err
		}
		if edgeGroups.Entities == nil || len(*edgeGroups.Entities) == 0 {
			break
		}
		for _, edgeGroup := range *edgeGroups.Entities {
			if edgeGroup.State != nil && *edgeGroup.State != "deleted" {
				allEdgeGroups = append(allEdgeGroups, edgeGroup)
			}
		}
	}
	return &allEdgeGroups, nil
}
