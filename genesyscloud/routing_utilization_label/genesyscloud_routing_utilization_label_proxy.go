package routing_utilization_label

import (
	"context"
	"fmt"
	"log"
	rc "terraform-provider-genesyscloud/genesyscloud/resource_cache"

	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

var internalProxy *routingUtilizationLabelProxy

type getAllRoutingUtilizationLabelsFunc func(ctx context.Context, p *routingUtilizationLabelProxy, name string) (*[]platformclientv2.Utilizationlabel, *platformclientv2.APIResponse, error)
type createRoutingUtilizationLabelFunc func(ctx context.Context, p *routingUtilizationLabelProxy, req *platformclientv2.Createutilizationlabelrequest) (*platformclientv2.Utilizationlabel, *platformclientv2.APIResponse, error)
type getRoutingUtilizationLabelFunc func(ctx context.Context, p *routingUtilizationLabelProxy, id string) (*platformclientv2.Utilizationlabel, *platformclientv2.APIResponse, error)
type getRoutingUtilizationLabelByNameFunc func(ctx context.Context, p *routingUtilizationLabelProxy, name string) (*platformclientv2.Utilizationlabel, bool, *platformclientv2.APIResponse, error)
type updateRoutingUtilizationLabelFunc func(ctx context.Context, p *routingUtilizationLabelProxy, id string, updateutilizationlabelrequest *platformclientv2.Updateutilizationlabelrequest) (*platformclientv2.Utilizationlabel, *platformclientv2.APIResponse, error)
type deleteRoutingUtilizationLabelFunc func(ctx context.Context, p *routingUtilizationLabelProxy, id string, forceDelete bool) (*platformclientv2.APIResponse, error)

type routingUtilizationLabelProxy struct {
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

func newRoutingUtilizationLabelProxy(clientConfig *platformclientv2.Configuration) *routingUtilizationLabelProxy {
	api := platformclientv2.NewRoutingApiWithConfig(clientConfig)
	routingCache := rc.NewResourceCache[platformclientv2.Utilizationlabel]()
	return &routingUtilizationLabelProxy{
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

func getRoutingUtilizationLabelProxy(clientConfig *platformclientv2.Configuration) *routingUtilizationLabelProxy {
	if internalProxy == nil {
		internalProxy = newRoutingUtilizationLabelProxy(clientConfig)
	}
	return internalProxy
}

func (p *routingUtilizationLabelProxy) getAllRoutingUtilizationLabels(ctx context.Context, name string) (*[]platformclientv2.Utilizationlabel, *platformclientv2.APIResponse, error) {
	return p.getAllRoutingUtilizationLabelsAttr(ctx, p, name)
}

func (p *routingUtilizationLabelProxy) createRoutingUtilizationLabel(ctx context.Context, req *platformclientv2.Createutilizationlabelrequest) (*platformclientv2.Utilizationlabel, *platformclientv2.APIResponse, error) {
	return p.createRoutingUtilizationLabelAttr(ctx, p, req)
}

func (p *routingUtilizationLabelProxy) getRoutingUtilizationLabel(ctx context.Context, id string) (*platformclientv2.Utilizationlabel, *platformclientv2.APIResponse, error) {
	return p.getRoutingUtilizationLabelAttr(ctx, p, id)
}

func (p *routingUtilizationLabelProxy) getRoutingUtilizationLabelByName(ctx context.Context, name string) (*platformclientv2.Utilizationlabel, bool, *platformclientv2.APIResponse, error) {
	return p.getRoutingUtilizationLabelByNameAttr(ctx, p, name)
}

func (p *routingUtilizationLabelProxy) updateRoutingUtilizationLabel(ctx context.Context, id string, req *platformclientv2.Updateutilizationlabelrequest) (*platformclientv2.Utilizationlabel, *platformclientv2.APIResponse, error) {
	return p.updateRoutingUtilizationLabelAttr(ctx, p, id, req)
}

func (p *routingUtilizationLabelProxy) deleteRoutingUtilizationLabel(ctx context.Context, id string, forceDelete bool) (*platformclientv2.APIResponse, error) {
	return p.deleteRoutingUtilizationLabelAttr(ctx, p, id, forceDelete)
}

func getAllRoutingUtilizationLabelsFn(_ context.Context, p *routingUtilizationLabelProxy, name string) (*[]platformclientv2.Utilizationlabel, *platformclientv2.APIResponse, error) {
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

func createRoutingUtilizationLabelFn(_ context.Context, p *routingUtilizationLabelProxy, req *platformclientv2.Createutilizationlabelrequest) (*platformclientv2.Utilizationlabel, *platformclientv2.APIResponse, error) {
	return p.routingApi.PostRoutingUtilizationLabels(*req)
}

func getRoutingUtilizationLabelFn(_ context.Context, p *routingUtilizationLabelProxy, id string) (*platformclientv2.Utilizationlabel, *platformclientv2.APIResponse, error) {
	if label := rc.GetCacheItem(p.routingCache, id); label != nil {
		return label, nil, nil
	}
	return p.routingApi.GetRoutingUtilizationLabel(id)
}

func getRoutingUtilizationLabelByNameFn(ctx context.Context, p *routingUtilizationLabelProxy, name string) (*platformclientv2.Utilizationlabel, bool, *platformclientv2.APIResponse, error) {
	labels, resp, err := getAllRoutingUtilizationLabelsFn(ctx, p, name)
	if err != nil {
		return nil, false, resp, fmt.Errorf("error retrieving routing utilization label by name %s", err)
	}

	if labels == nil || len(*labels) == 0 {
		return nil, true, resp, fmt.Errorf("no routing utilization labels found with name %s", name)
	}

	for _, label := range *labels {
		if *label.Name == name {
			log.Printf("Retrieved routing utilization label %s by name %s", *label.Id, name)
			return &label, false, resp, nil
		}
	}
	return nil, true, resp, fmt.Errorf("no routing utilization label found with name: %s", name)
}

func updateRoutingUtilizationLabelFn(_ context.Context, p *routingUtilizationLabelProxy, id string, req *platformclientv2.Updateutilizationlabelrequest) (*platformclientv2.Utilizationlabel, *platformclientv2.APIResponse, error) {
	return p.routingApi.PutRoutingUtilizationLabel(id, *req)
}

func deleteRoutingUtilizationLabelFn(_ context.Context, p *routingUtilizationLabelProxy, id string, forceDelete bool) (*platformclientv2.APIResponse, error) {
	return p.routingApi.DeleteRoutingUtilizationLabel(id, forceDelete)
}
