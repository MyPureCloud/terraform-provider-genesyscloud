package routing_utilization_label

import (
	"context"
	"fmt"
	"github.com/mypurecloud/platform-client-sdk-go/v131/platformclientv2"
	rc "terraform-provider-genesyscloud/genesyscloud/resource_cache"
)

var internalProxy *routingUtilizationProxy

type getAllRoutingUtilizationLabelsFunc func(ctx context.Context, p *routingUtilizationProxy, name string) (*[]platformclientv2.Utilizationlabel, *platformclientv2.APIResponse, error)
type createRoutingUtilizationLabelFunc func(ctx context.Context, p *routingUtilizationProxy, req *platformclientv2.Createutilizationlabelrequest) (*platformclientv2.Utilizationlabel, *platformclientv2.APIResponse, error)
type getRoutingUtilizationLabelFunc func(ctx context.Context, p *routingUtilizationProxy, id string) (*platformclientv2.Utilizationlabel, *platformclientv2.APIResponse, error)
type getRoutingUtilizationLabelByNameFunc func(ctx context.Context, p *routingUtilizationProxy, name string) (*platformclientv2.Utilizationlabel, *platformclientv2.APIResponse, error)
type updateRoutingUtilizationLabelFunc func(ctx context.Context, p *routingUtilizationProxy, id string, updateutilizationlabelrequest *platformclientv2.Updateutilizationlabelrequest) (*platformclientv2.Utilizationlabel, *platformclientv2.APIResponse, error)
type deleteRoutingUtilizationLabelFunc func(ctx context.Context, p *routingUtilizationProxy, id string, forceDelete bool) (*platformclientv2.APIResponse, error)

type routingUtilizationProxy struct {
	clientConfig                         *platformclientv2.Configuration
	routingApi                           *platformclientv2.RoutingApi
	getAllRoutingUtilizationLabelsAttr   getAllRoutingUtilizationLabelsFunc
	createRoutingUtilizationLabelAttr    createRoutingUtilizationLabelFunc
	getRoutingUtilizationLabelAttr       getRoutingUtilizationLabelFunc
	getRoutingUtilizationLabelByNameAttr getRoutingUtilizationLabelByNameFunc
	updateRoutingUtilizationLabelAttr    updateRoutingUtilizationLabelFunc
	deleteRoutingUtilizationLabelAttr    deleteRoutingUtilizationLabelFunc
	routingCache                         rc.CacheInterface[platformclientv2.Utilizationlabel]
}

func newRoutingUtilizationProxy(clientConfig *platformclientv2.Configuration) *routingUtilizationProxy {
	api := platformclientv2.NewRoutingApiWithConfig(clientConfig)
	routingCache := rc.NewResourceCache[platformclientv2.Utilizationlabel]()
	return &routingUtilizationProxy{
		clientConfig:                         clientConfig,
		routingApi:                           api,
		getAllRoutingUtilizationLabelsAttr:   getAllRoutingUtilizationLabelsFn,
		createRoutingUtilizationLabelAttr:    createRoutingUtilizationLabelFn,
		getRoutingUtilizationLabelAttr:       getRoutingUtilizationLabelFn,
		getRoutingUtilizationLabelByNameAttr: getRoutingUtilizationLabelByNameFn,
		updateRoutingUtilizationLabelAttr:    updateRoutingUtilizationLabelFn,
		deleteRoutingUtilizationLabelAttr:    deleteRoutingUtilizationLabelFn,
		routingCache:                         routingCache,
	}
}

func getRoutingUtilizationProxy(clientConfig *platformclientv2.Configuration) *routingUtilizationProxy {
	if internalProxy == nil {
		internalProxy = newRoutingUtilizationProxy(clientConfig)
	}
	return internalProxy
}

func (p *routingUtilizationProxy) getAllRoutingUtilizationLabels(ctx context.Context, name string) (*[]platformclientv2.Utilizationlabel, *platformclientv2.APIResponse, error) {
	return p.getAllRoutingUtilizationLabelsAttr(ctx, p, name)
}

func (p *routingUtilizationProxy) createRoutingUtilizationLabel(ctx context.Context, req *platformclientv2.Createutilizationlabelrequest) (*platformclientv2.Utilizationlabel, *platformclientv2.APIResponse, error) {
	return p.createRoutingUtilizationLabelAttr(ctx, p, req)
}

func (p *routingUtilizationProxy) getRoutingUtilizationLabel(ctx context.Context, id string) (*platformclientv2.Utilizationlabel, *platformclientv2.APIResponse, error) {
	return p.getRoutingUtilizationLabelAttr(ctx, p, id)
}

func (p *routingUtilizationProxy) getRoutingUtilizationLabelByName(ctx context.Context, name string) (*platformclientv2.Utilizationlabel, *platformclientv2.APIResponse, error) {
	return p.getRoutingUtilizationLabelByNameAttr(ctx, p, name)
}

func (p *routingUtilizationProxy) updateRoutingUtilizationLabel(ctx context.Context, id string, req *platformclientv2.Updateutilizationlabelrequest) (*platformclientv2.Utilizationlabel, *platformclientv2.APIResponse, error) {
	return p.updateRoutingUtilizationLabelAttr(ctx, p, id, req)
}

func (p *routingUtilizationProxy) deleteRoutingUtilizationLabel(ctx context.Context, id string, forceDelete bool) (*platformclientv2.APIResponse, error) {
	return p.deleteRoutingUtilizationLabelAttr(ctx, p, id, forceDelete)
}

func getAllRoutingUtilizationLabelsFn(_ context.Context, p *routingUtilizationProxy, name string) (*[]platformclientv2.Utilizationlabel, *platformclientv2.APIResponse, error) {
	var allUtilizationLabels []platformclientv2.Utilizationlabel
	const pageSize = 100

	labels, resp, err := p.routingApi.GetRoutingUtilizationLabels(100, 1, "", name)
	if err != nil {
		return nil, resp, fmt.Errorf("failed to get routing utilization labels | error: %s", err)
	}

	if labels.Entities == nil || len(*labels.Entities) == 0 {
		return &allUtilizationLabels, resp, nil
	}
	allUtilizationLabels = append(allUtilizationLabels, *labels.Entities...)

	for pageNum := 2; pageNum <= *labels.PageCount; pageNum++ {
		labels, resp, err := p.routingApi.GetRoutingUtilizationLabels(pageSize, pageNum, "", name)
		if err != nil {
			return nil, resp, fmt.Errorf("failed to get routing utilization labels | error: %s", err)
		}

		if labels.Entities == nil || len(*labels.Entities) == 0 {
			break
		}
		allUtilizationLabels = append(allUtilizationLabels, *labels.Entities...)
	}

	for _, label := range allUtilizationLabels {
		rc.SetCache(p.routingCache, *label.Id, label)
	}

	return &allUtilizationLabels, resp, nil
}

func createRoutingUtilizationLabelFn(_ context.Context, p *routingUtilizationProxy, req *platformclientv2.Createutilizationlabelrequest) (*platformclientv2.Utilizationlabel, *platformclientv2.APIResponse, error) {
	return p.routingApi.PostRoutingUtilizationLabels(*req)
}

func getRoutingUtilizationLabelFn(_ context.Context, p *routingUtilizationProxy, id string) (*platformclientv2.Utilizationlabel, *platformclientv2.APIResponse, error) {
	if label := rc.GetCacheItem(p.routingCache, id); label != nil {
		return label, nil, nil
	}
	return p.routingApi.GetRoutingUtilizationLabel(id)
}

func getRoutingUtilizationLabelByNameFn(ctx context.Context, p *routingUtilizationProxy, name string) (*platformclientv2.Utilizationlabel, *platformclientv2.APIResponse, error) {
	labels, resp, err := getAllRoutingUtilizationLabelsFn(ctx, p, name)
	if err != nil {
		return nil, resp, fmt.Errorf("error retrieving routing utilization label by name %s", err)
	}

	for _, label := range *labels {
		if *label.Name == name {
			return &label, resp, nil
		}
	}
	return nil, resp, fmt.Errorf("no routing utilization label found with name: %s", name)
}

func updateRoutingUtilizationLabelFn(_ context.Context, p *routingUtilizationProxy, id string, req *platformclientv2.Updateutilizationlabelrequest) (*platformclientv2.Utilizationlabel, *platformclientv2.APIResponse, error) {
	return p.routingApi.PutRoutingUtilizationLabel(id, *req)
}

func deleteRoutingUtilizationLabelFn(_ context.Context, p *routingUtilizationProxy, id string, forceDelete bool) (*platformclientv2.APIResponse, error) {
	return p.routingApi.DeleteRoutingUtilizationLabel(id, forceDelete)
}
