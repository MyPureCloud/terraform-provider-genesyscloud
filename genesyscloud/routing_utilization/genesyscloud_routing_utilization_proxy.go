package routing_utilization

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v131/platformclientv2"
)

var internalProxy *routingUtilizationProxy

type getRoutingUtilizationFunc func(ctx context.Context, p *routingUtilizationProxy) (*platformclientv2.APIResponse, error)
type updateRoutingUtilizationFunc func(ctx context.Context, p *routingUtilizationProxy, request *platformclientv2.Utilizationrequest) (*platformclientv2.Utilizationresponse, *platformclientv2.APIResponse, error)
type deleteRoutingUtilizationFunc func(ctx context.Context, p *routingUtilizationProxy) (*platformclientv2.APIResponse, error)

type updateDirectlyFunc func(ctx context.Context, p *routingUtilizationProxy, d *schema.ResourceData, utilizationRequest []interface{}) (*platformclientv2.APIResponse, error)

type routingUtilizationProxy struct {
	clientConfig                 *platformclientv2.Configuration
	routingApi                   *platformclientv2.RoutingApi
	getRoutingUtilizationAttr    getRoutingUtilizationFunc
	updateRoutingUtilizationAttr updateRoutingUtilizationFunc
	deleteRoutingUtilizationAttr deleteRoutingUtilizationFunc

	updateDirectlyAttr updateDirectlyFunc
}

func newRoutingUtilizationProxy(clientConfig *platformclientv2.Configuration) *routingUtilizationProxy {
	api := platformclientv2.NewRoutingApiWithConfig(clientConfig)
	return &routingUtilizationProxy{
		clientConfig:                 clientConfig,
		routingApi:                   api,
		getRoutingUtilizationAttr:    getRoutingUtilizationFn,
		updateRoutingUtilizationAttr: updateRoutingUtilizationFn,
		deleteRoutingUtilizationAttr: deleteRoutingUtilizationFn,

		updateDirectlyAttr: updateDirectlyFn,
	}
}

func getRoutingUtilizationProxy(clientConfig *platformclientv2.Configuration) *routingUtilizationProxy {
	if internalProxy == nil {
		internalProxy = newRoutingUtilizationProxy(clientConfig)
	}
	return internalProxy
}

func (p *routingUtilizationProxy) getRoutingUtilization(ctx context.Context) (*platformclientv2.APIResponse, error) {
	return p.getRoutingUtilizationAttr(ctx, p)
}
func (p *routingUtilizationProxy) updateRoutingUtilization(ctx context.Context, request *platformclientv2.Utilizationrequest) (*platformclientv2.Utilizationresponse, *platformclientv2.APIResponse, error) {
	return p.updateRoutingUtilizationAttr(ctx, p, request)
}
func (p *routingUtilizationProxy) deleteRoutingUtilization(ctx context.Context) (*platformclientv2.APIResponse, error) {
	return p.deleteRoutingUtilizationAttr(ctx, p)
}

func (p *routingUtilizationProxy) updateDirectly(ctx context.Context, d *schema.ResourceData, utilizationRequest []interface{}) (*platformclientv2.APIResponse, error) {
	return p.updateDirectlyAttr(ctx, p, d, utilizationRequest)
}

// Calling the Utilization API directly while the label feature is not available.
// Once it is, this code can go back to using platformclientv2's RoutingApi to make the call.
func getRoutingUtilizationFn(ctx context.Context, p *routingUtilizationProxy) (*platformclientv2.APIResponse, error) {
	apiClient := &p.routingApi.Configuration.APIClient
	path := fmt.Sprintf("%s/api/v2/routing/utilization", p.routingApi.Configuration.BasePath)
	headerParams := buildHeaderParams(p.routingApi)
	resp, err := apiClient.CallAPI(path, "GET", nil, headerParams, nil, nil, "", nil)
	if err != nil {
		return resp, fmt.Errorf("failed to get routing utilization %s ", err)
	}
	return resp, nil
}

func updateRoutingUtilizationFn(ctx context.Context, p *routingUtilizationProxy, utilizationRequest *platformclientv2.Utilizationrequest) (*platformclientv2.Utilizationresponse, *platformclientv2.APIResponse, error) {
	return p.routingApi.PutRoutingUtilization(*utilizationRequest)
}

func deleteRoutingUtilizationFn(ctx context.Context, p *routingUtilizationProxy) (*platformclientv2.APIResponse, error) {
	return p.routingApi.DeleteRoutingUtilization()
}

// If the resource has label(s), calls the Utilization API directly.
// This code can go back to using platformclientv2's RoutingApi to make the call once label utilization is available in platformclientv2's RoutingApi
func updateDirectlyFn(ctx context.Context, p *routingUtilizationProxy, d *schema.ResourceData, utilizationRequest []interface{}) (*platformclientv2.APIResponse, error) {
	apiClient := &p.routingApi.Configuration.APIClient

	path := fmt.Sprintf("%s/api/v2/routing/utilization", p.routingApi.Configuration.BasePath)
	headerParams := buildHeaderParams(p.routingApi)
	requestPayload := make(map[string]interface{})
	requestPayload["utilization"] = buildSdkMediaUtilizations(d)
	requestPayload["labelUtilizations"] = BuildLabelUtilizationsRequest(utilizationRequest)

	resp, err := apiClient.CallAPI(path, "PUT", requestPayload, headerParams, nil, nil, "", nil)
	if err != nil {
		return resp, fmt.Errorf("error updating directly %s", err)
	}
	return resp, nil
}
