package telephony_providers_edges_edge_group

import (
	"context"
	"fmt"
	"log"

	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

var internalProxy *edgeGroupProxy

type getEdgeGroupByIdFunc func(ctx context.Context, p *edgeGroupProxy, edgeGroupId string) (*platformclientv2.Edgegroup, *platformclientv2.APIResponse, error)
type deleteEdgeGroupFunc func(ctx context.Context, p *edgeGroupProxy, edgeGroupId string) (*platformclientv2.APIResponse, error)
type updateEdgeGroupFunc func(ctx context.Context, p *edgeGroupProxy, edgeGroupId string, body platformclientv2.Edgegroup) (*platformclientv2.Edgegroup, *platformclientv2.APIResponse, error)
type createEdgeGroupFunc func(ctx context.Context, p *edgeGroupProxy, body platformclientv2.Edgegroup) (*platformclientv2.Edgegroup, *platformclientv2.APIResponse, error)
type getAllEdgeGroupsFunc func(ctx context.Context, p *edgeGroupProxy, edgeGroupName string, managed bool) (*[]platformclientv2.Edgegroup, *platformclientv2.APIResponse, error)
type getEdgeGroupByNameFunc func(ctx context.Context, p *edgeGroupProxy, edgeGroupName string, managed bool) (string, bool, *platformclientv2.APIResponse, error)

type edgeGroupProxy struct {
	clientConfig *platformclientv2.Configuration
	edgesApi     *platformclientv2.TelephonyProvidersEdgeApi

	getEdgeGroupByIdAttr   getEdgeGroupByIdFunc
	deleteEdgeGroupAttr    deleteEdgeGroupFunc
	updateEdgeGroupAttr    updateEdgeGroupFunc
	createEdgeGroupAttr    createEdgeGroupFunc
	getAllEdgeGroupsAttr   getAllEdgeGroupsFunc
	getEdgeGroupByNameAttr getEdgeGroupByNameFunc
}

func newEdgeGroupProxy(clientConfig *platformclientv2.Configuration) *edgeGroupProxy {
	edgesApi := platformclientv2.NewTelephonyProvidersEdgeApiWithConfig(clientConfig)

	return &edgeGroupProxy{
		clientConfig: clientConfig,
		edgesApi:     edgesApi,

		getEdgeGroupByIdAttr:   getEdgeGroupByIdFn,
		deleteEdgeGroupAttr:    deleteEdgeGroupFn,
		updateEdgeGroupAttr:    updateEdgeGroupFn,
		createEdgeGroupAttr:    createEdgeGroupFn,
		getAllEdgeGroupsAttr:   getAllEdgeGroupsFn,
		getEdgeGroupByNameAttr: getEdgeGroupByNameFn,
	}
}

func getEdgeGroupProxy(clientConfig *platformclientv2.Configuration) *edgeGroupProxy {
	if internalProxy == nil {
		internalProxy = newEdgeGroupProxy(clientConfig)
	}
	return internalProxy
}

func (p *edgeGroupProxy) getEdgeGroupById(ctx context.Context, edgeGroupId string) (*platformclientv2.Edgegroup, *platformclientv2.APIResponse, error) {
	return p.getEdgeGroupByIdAttr(ctx, p, edgeGroupId)
}

func (p *edgeGroupProxy) deleteEdgeGroup(ctx context.Context, edgeGroupId string) (*platformclientv2.APIResponse, error) {
	return p.deleteEdgeGroupAttr(ctx, p, edgeGroupId)
}

func (p *edgeGroupProxy) updateEdgeGroup(ctx context.Context, edgeGroupId string, body platformclientv2.Edgegroup) (*platformclientv2.Edgegroup, *platformclientv2.APIResponse, error) {
	return p.updateEdgeGroupAttr(ctx, p, edgeGroupId, body)
}

func (p *edgeGroupProxy) createEdgeGroup(ctx context.Context, body platformclientv2.Edgegroup) (*platformclientv2.Edgegroup, *platformclientv2.APIResponse, error) {
	return p.createEdgeGroupAttr(ctx, p, body)
}

func (p *edgeGroupProxy) getAllEdgeGroups(ctx context.Context, edgeGroupName string, managed bool) (*[]platformclientv2.Edgegroup, *platformclientv2.APIResponse, error) {
	return p.getAllEdgeGroupsAttr(ctx, p, edgeGroupName, managed)
}

func (p *edgeGroupProxy) getEdgeGroupByName(ctx context.Context, edgeGroupName string, managed bool) (string, bool, *platformclientv2.APIResponse, error) {
	return p.getEdgeGroupByNameAttr(ctx, p, edgeGroupName, managed)
}

func getEdgeGroupByNameFn(ctx context.Context, p *edgeGroupProxy, edgeGroupName string, managed bool) (string, bool, *platformclientv2.APIResponse, error) {
	var targetEdgeGroup platformclientv2.Edgegroup
	edgeGroups, resp, err := getAllEdgeGroupsFn(ctx, p, edgeGroupName, managed)
	if err != nil {
		return "", true, resp, fmt.Errorf("Error searching Edge Group By Name %s: %s", edgeGroupName, err)
	}
	for _, edgeGroup := range *edgeGroups {
		if *edgeGroup.Name == edgeGroupName {
			log.Printf("Retrieved Edge Group id %s by name %s", *edgeGroup.Id, edgeGroupName)
			targetEdgeGroup = edgeGroup
			return *targetEdgeGroup.Id, false, resp, nil
		}
	}
	return "", true, resp, fmt.Errorf("Unable to find EdgeGroup with name %s", edgeGroupName)
}

func getEdgeGroupByIdFn(ctx context.Context, p *edgeGroupProxy, edgeGroupId string) (*platformclientv2.Edgegroup, *platformclientv2.APIResponse, error) {
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

func updateEdgeGroupFn(ctx context.Context, p *edgeGroupProxy, edgeGroupId string, body platformclientv2.Edgegroup) (*platformclientv2.Edgegroup, *platformclientv2.APIResponse, error) {
	edgeGroup, resp, err := p.edgesApi.PutTelephonyProvidersEdgesEdgegroup(edgeGroupId, body)
	if err != nil {
		return nil, resp, err
	}
	return edgeGroup, resp, nil
}

func createEdgeGroupFn(ctx context.Context, p *edgeGroupProxy, body platformclientv2.Edgegroup) (*platformclientv2.Edgegroup, *platformclientv2.APIResponse, error) {
	edgeGroup, resp, err := p.edgesApi.PostTelephonyProvidersEdgesEdgegroups(body)
	if err != nil {
		return nil, resp, err
	}
	return edgeGroup, resp, nil
}

func getAllEdgeGroupsFn(ctx context.Context, p *edgeGroupProxy, edgeGroupName string, managed bool) (*[]platformclientv2.Edgegroup, *platformclientv2.APIResponse, error) {
	const pageSize = 100
	var allEdgeGroups []platformclientv2.Edgegroup
	var pageNum = 1

	edgeGroups, resp, err := p.edgesApi.GetTelephonyProvidersEdgesEdgegroups(pageSize, pageNum, edgeGroupName, "", managed)
	if err != nil {
		return nil, resp, err
	}
	if edgeGroups.Entities != nil && len(*edgeGroups.Entities) > 0 {
		for _, edgeGroup := range *edgeGroups.Entities {
			if edgeGroup.State != nil && *edgeGroup.State != "deleted" {
				allEdgeGroups = append(allEdgeGroups, edgeGroup)
			}
		}
	}

	for pageNum := 2; pageNum <= *edgeGroups.PageCount; pageNum++ {
		edgeGroups, resp, err := p.edgesApi.GetTelephonyProvidersEdgesEdgegroups(pageSize, pageNum, edgeGroupName, "", managed)
		if err != nil {
			return nil, resp, err
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
	return &allEdgeGroups, resp, nil
}
