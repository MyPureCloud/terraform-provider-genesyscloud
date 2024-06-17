package routing_utilization_label

import (
	"fmt"
	"github.com/mypurecloud/platform-client-sdk-go/v130/platformclientv2"
	"log"
	rc "terraform-provider-genesyscloud/genesyscloud/resource_cache"
)

var internalProxy *routingUtilizationProxy

type getAllRoutingUtilizationLabelsFunc func(p *routingUtilizationProxy, name string) (*[]platformclientv2.Utilizationlabel, *platformclientv2.APIResponse, error)
type createRoutingUtilizationLabelFunc func(p *routingUtilizationProxy, req *platformclientv2.Createutilizationlabelrequest) (*platformclientv2.Utilizationlabel, *platformclientv2.APIResponse, error)
type getRoutingUtilizationLabelFunc func(p *routingUtilizationProxy, id string) (*platformclientv2.Utilizationlabel, *platformclientv2.APIResponse, error)
type getRoutingUtilizationLabelByNameFunc func(p *routingUtilizationProxy, name string) (*platformclientv2.Utilizationlabel, *platformclientv2.APIResponse, error)
type updateRoutingUtilizationLabelFunc func(p *routingUtilizationProxy, id string, updateutilizationlabelrequest *platformclientv2.Updateutilizationlabelrequest) (*platformclientv2.Utilizationlabel, *platformclientv2.APIResponse, error)
type deleteRoutingUtilizationLabelFunc func(p *routingUtilizationProxy, id string, forceDelete bool) (*platformclientv2.APIResponse, error)

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
	var allUtilizationLabels []platformclientv2.Utilizationlabel
	const pageSize = 100

	labels, resp, err := p.routingApi.GetRoutingUtilizationLabels(100, 1, "", name)
	if err != nil {
		return nil, resp, fmt.Errorf("failed to get routing utilization labels | error: %s", err)
	}

	if name != "" {
		log.Println("here: ", labels)
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

func createRoutingUtilizationLabelFn(p *routingUtilizationProxy, req *platformclientv2.Createutilizationlabelrequest) (*platformclientv2.Utilizationlabel, *platformclientv2.APIResponse, error) {
	label, resp, err := p.routingApi.PostRoutingUtilizationLabels(*req)
	if err != nil {
		return nil, resp, fmt.Errorf("failed to create routing utilization label | error: %s", err)
	}
	return label, resp, nil
}

func getRoutingUtilizationLabelFn(p *routingUtilizationProxy, id string) (*platformclientv2.Utilizationlabel, *platformclientv2.APIResponse, error) {
	if label := rc.GetCacheItem(p.routingCache, id); label != nil {
		return label, nil, nil
	}

	label, resp, err := p.routingApi.GetRoutingUtilizationLabel(id)
	if err != nil {
		return nil, resp, fmt.Errorf("failed to get routing utilization label: %s | error: %s", id, err)
	}
	return label, resp, nil
}

func getRoutingUtilizationLabelByNameFn(p *routingUtilizationProxy, name string) (*platformclientv2.Utilizationlabel, *platformclientv2.APIResponse, error) {
	labels, resp, err := getAllRoutingUtilizationLabelsFn(p, name)
	if err != nil {
		return nil, resp, fmt.Errorf("error retrieving routing utilization label by name %s", err)
	}

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
}
