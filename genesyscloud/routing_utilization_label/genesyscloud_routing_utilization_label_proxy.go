package routing_utilization_label

import (
<<<<<<< HEAD
	"context"
	"fmt"
	"github.com/mypurecloud/platform-client-sdk-go/v131/platformclientv2"
=======
	"fmt"
	"github.com/mypurecloud/platform-client-sdk-go/v130/platformclientv2"
	"log"
>>>>>>> f33044e5 (refactor routing utilization label)
	rc "terraform-provider-genesyscloud/genesyscloud/resource_cache"
)

var internalProxy *routingUtilizationProxy

<<<<<<< HEAD
type getAllRoutingUtilizationLabelsFunc func(ctx context.Context, p *routingUtilizationProxy, name string) (*[]platformclientv2.Utilizationlabel, *platformclientv2.APIResponse, error)
type createRoutingUtilizationLabelFunc func(ctx context.Context, p *routingUtilizationProxy, req *platformclientv2.Createutilizationlabelrequest) (*platformclientv2.Utilizationlabel, *platformclientv2.APIResponse, error)
type getRoutingUtilizationLabelFunc func(ctx context.Context, p *routingUtilizationProxy, id string) (*platformclientv2.Utilizationlabel, *platformclientv2.APIResponse, error)
type getRoutingUtilizationLabelByNameFunc func(ctx context.Context, p *routingUtilizationProxy, name string) (*platformclientv2.Utilizationlabel, *platformclientv2.APIResponse, error)
type updateRoutingUtilizationLabelFunc func(ctx context.Context, p *routingUtilizationProxy, id string, updateutilizationlabelrequest *platformclientv2.Updateutilizationlabelrequest) (*platformclientv2.Utilizationlabel, *platformclientv2.APIResponse, error)
type deleteRoutingUtilizationLabelFunc func(ctx context.Context, p *routingUtilizationProxy, id string, forceDelete bool) (*platformclientv2.APIResponse, error)
=======
type getAllRoutingUtilizationLabelsFunc func(p *routingUtilizationProxy, name string) (*[]platformclientv2.Utilizationlabel, *platformclientv2.APIResponse, error)
type createRoutingUtilizationLabelFunc func(p *routingUtilizationProxy, req *platformclientv2.Createutilizationlabelrequest) (*platformclientv2.Utilizationlabel, *platformclientv2.APIResponse, error)
type getRoutingUtilizationLabelFunc func(p *routingUtilizationProxy, id string) (*platformclientv2.Utilizationlabel, *platformclientv2.APIResponse, error)
type getRoutingUtilizationLabelByNameFunc func(p *routingUtilizationProxy, name string) (*platformclientv2.Utilizationlabel, *platformclientv2.APIResponse, error)
type updateRoutingUtilizationLabelFunc func(p *routingUtilizationProxy, id string, updateutilizationlabelrequest *platformclientv2.Updateutilizationlabelrequest) (*platformclientv2.Utilizationlabel, *platformclientv2.APIResponse, error)
type deleteRoutingUtilizationLabelFunc func(p *routingUtilizationProxy, id string, forceDelete bool) (*platformclientv2.APIResponse, error)
>>>>>>> f33044e5 (refactor routing utilization label)

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

<<<<<<< HEAD
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
=======
func (p *routingUtilizationProxy) getAllRoutingUtilizationLabels(name string) (*[]platformclientv2.Utilizationlabel, *platformclientv2.APIResponse, error) {
	return p.getAllRoutingUtilizationLabelsAttr(p, name)
}

func (p *routingUtilizationProxy) createRoutingUtilizationLabel(req *platformclientv2.Createutilizationlabelrequest) (*platformclientv2.Utilizationlabel, *platformclientv2.APIResponse, error) {
	return p.createRoutingUtilizationLabelAttr(p, req)
}

func (p *routingUtilizationProxy) getRoutingUtilizationLabel(id string) (*platformclientv2.Utilizationlabel, *platformclientv2.APIResponse, error) {
	return p.getRoutingUtilizationLabelAttr(p, id)
}

func (p *routingUtilizationProxy) getRoutingUtilizationLabelByName(name string) (*platformclientv2.Utilizationlabel, *platformclientv2.APIResponse, error) {
	return p.getRoutingUtilizationLabelByNameAttr(p, name)
}

func (p *routingUtilizationProxy) updateRoutingUtilizationLabel(id string, req *platformclientv2.Updateutilizationlabelrequest) (*platformclientv2.Utilizationlabel, *platformclientv2.APIResponse, error) {
	return p.updateRoutingUtilizationLabelAttr(p, id, req)
}

func (p *routingUtilizationProxy) deleteRoutingUtilizationLabel(id string, forceDelete bool) (*platformclientv2.APIResponse, error) {
	return p.deleteRoutingUtilizationLabelAttr(p, id, forceDelete)
}

func getAllRoutingUtilizationLabelsFn(p *routingUtilizationProxy, name string) (*[]platformclientv2.Utilizationlabel, *platformclientv2.APIResponse, error) {
>>>>>>> f33044e5 (refactor routing utilization label)
	var allUtilizationLabels []platformclientv2.Utilizationlabel
	const pageSize = 100

	labels, resp, err := p.routingApi.GetRoutingUtilizationLabels(100, 1, "", name)
	if err != nil {
		return nil, resp, fmt.Errorf("failed to get routing utilization labels | error: %s", err)
	}

<<<<<<< HEAD
=======
	if name != "" {
		log.Println("here: ", labels)
	}

>>>>>>> f33044e5 (refactor routing utilization label)
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

<<<<<<< HEAD
func createRoutingUtilizationLabelFn(_ context.Context, p *routingUtilizationProxy, req *platformclientv2.Createutilizationlabelrequest) (*platformclientv2.Utilizationlabel, *platformclientv2.APIResponse, error) {
	return p.routingApi.PostRoutingUtilizationLabels(*req)
}

func getRoutingUtilizationLabelFn(_ context.Context, p *routingUtilizationProxy, id string) (*platformclientv2.Utilizationlabel, *platformclientv2.APIResponse, error) {
=======
func createRoutingUtilizationLabelFn(p *routingUtilizationProxy, req *platformclientv2.Createutilizationlabelrequest) (*platformclientv2.Utilizationlabel, *platformclientv2.APIResponse, error) {
	label, resp, err := p.routingApi.PostRoutingUtilizationLabels(*req)
	if err != nil {
		return nil, resp, fmt.Errorf("failed to create routing utilization label | error: %s", err)
	}
	return label, resp, nil
}

func getRoutingUtilizationLabelFn(p *routingUtilizationProxy, id string) (*platformclientv2.Utilizationlabel, *platformclientv2.APIResponse, error) {
>>>>>>> f33044e5 (refactor routing utilization label)
	if label := rc.GetCacheItem(p.routingCache, id); label != nil {
		return label, nil, nil
	}

<<<<<<< HEAD
	return p.routingApi.GetRoutingUtilizationLabel(id)
}

func getRoutingUtilizationLabelByNameFn(ctx context.Context, p *routingUtilizationProxy, name string) (*platformclientv2.Utilizationlabel, *platformclientv2.APIResponse, error) {
	labels, resp, err := getAllRoutingUtilizationLabelsFn(ctx, p, name)
=======
	label, resp, err := p.routingApi.GetRoutingUtilizationLabel(id)
	if err != nil {
		return nil, resp, fmt.Errorf("failed to get routing utilization label: %s | error: %s", id, err)
	}
	return label, resp, nil
}

func getRoutingUtilizationLabelByNameFn(p *routingUtilizationProxy, name string) (*platformclientv2.Utilizationlabel, *platformclientv2.APIResponse, error) {
	labels, resp, err := getAllRoutingUtilizationLabelsFn(p, name)
>>>>>>> f33044e5 (refactor routing utilization label)
	if err != nil {
		return nil, resp, fmt.Errorf("error retrieving routing utilization label by name %s", err)
	}

<<<<<<< HEAD
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
=======
	label := (*labels)[0]
	return &label, resp, nil
}

func updateRoutingUtilizationLabelFn(p *routingUtilizationProxy, id string, req *platformclientv2.Updateutilizationlabelrequest) (*platformclientv2.Utilizationlabel, *platformclientv2.APIResponse, error) {
	label, resp, err := p.routingApi.PutRoutingUtilizationLabel(id, *req)
	if err != nil {
		return nil, resp, fmt.Errorf("failed to update routing utilization label %s | error: %s", id, err)
	}
	return label, resp, nil
}

func deleteRoutingUtilizationLabelFn(p *routingUtilizationProxy, id string, forceDelete bool) (*platformclientv2.APIResponse, error) {
	resp, err := p.routingApi.DeleteRoutingUtilizationLabel(id, forceDelete)
	if err != nil {
		return resp, fmt.Errorf("failed to delete routing utilization label: %s | error: %s", id, err)
	}
	return resp, nil
>>>>>>> f33044e5 (refactor routing utilization label)
}
