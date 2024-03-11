package routing_email_route

import (
	"context"
	"fmt"
	"github.com/mypurecloud/platform-client-sdk-go/v123/platformclientv2"
	"log"
)

/*
The genesyscloud_routing_email_route_proxy.go file contains the proxy structures and methods that interact
with the Genesys Cloud SDK. We use composition here for each function on the proxy so individual functions can be stubbed
out during testing.
*/

// internalProxy holds a proxy instance that can be used throughout the package
var internalProxy *routingEmailRouteProxy

// Type definitions for each func on our proxy so we can easily mock them out later
type createRoutingEmailRouteFunc func(ctx context.Context, p *routingEmailRouteProxy, inboundRoute *platformclientv2.Inboundroute) (*platformclientv2.Inboundroute, error)
type getAllRoutingEmailRouteFunc func(ctx context.Context, p *routingEmailRouteProxy) (*[]platformclientv2.Inboundroute, error)
type getRoutingEmailRouteIdByNameFunc func(ctx context.Context, p *routingEmailRouteProxy, name string) (id string, retryable bool, err error)
type getRoutingEmailRouteByIdFunc func(ctx context.Context, p *routingEmailRouteProxy, id string) (inboundRoute *platformclientv2.Inboundroute, responseCode int, err error)
type updateRoutingEmailRouteFunc func(ctx context.Context, p *routingEmailRouteProxy, id string, domainId string, inboundRoute *platformclientv2.Inboundroute) (*platformclientv2.Inboundroute, int, error)
type deleteRoutingEmailRouteFunc func(ctx context.Context, p *routingEmailRouteProxy, domainId string, id string) (responseCode int, err error)

// routingEmailRouteProxy contains all of the methods that call genesys cloud APIs.
type routingEmailRouteProxy struct {
	clientConfig                     *platformclientv2.Configuration
	routingApi                       *platformclientv2.RoutingApi
	createRoutingEmailRouteAttr      createRoutingEmailRouteFunc
	getAllRoutingEmailRouteAttr      getAllRoutingEmailRouteFunc
	getRoutingEmailRouteIdByNameAttr getRoutingEmailRouteIdByNameFunc
	getRoutingEmailRouteByIdAttr     getRoutingEmailRouteByIdFunc
	updateRoutingEmailRouteAttr      updateRoutingEmailRouteFunc
	deleteRoutingEmailRouteAttr      deleteRoutingEmailRouteFunc
}

// newRoutingEmailRouteProxy initializes the routing email route proxy with all of the data needed to communicate with Genesys Cloud
func newRoutingEmailRouteProxy(clientConfig *platformclientv2.Configuration) *routingEmailRouteProxy {
	api := platformclientv2.NewRoutingApiWithConfig(clientConfig)
	return &routingEmailRouteProxy{
		clientConfig:                     clientConfig,
		routingApi:                       api,
		createRoutingEmailRouteAttr:      createRoutingEmailRouteFn,
		getAllRoutingEmailRouteAttr:      getAllRoutingEmailRouteFn,
		getRoutingEmailRouteIdByNameAttr: getRoutingEmailRouteIdByNameFn,
		getRoutingEmailRouteByIdAttr:     getRoutingEmailRouteByIdFn,
		updateRoutingEmailRouteAttr:      updateRoutingEmailRouteFn,
		deleteRoutingEmailRouteAttr:      deleteRoutingEmailRouteFn,
	}
}

// getRoutingEmailRouteProxy acts as a singleton to for the internalProxy.  It also ensures
// that we can still proxy our tests by directly setting internalProxy package variable
func getRoutingEmailRouteProxy(clientConfig *platformclientv2.Configuration) *routingEmailRouteProxy {
	if internalProxy == nil {
		internalProxy = newRoutingEmailRouteProxy(clientConfig)
	}

	return internalProxy
}

// createRoutingEmailRoute creates a Genesys Cloud routing email route
func (p *routingEmailRouteProxy) createRoutingEmailRoute(ctx context.Context, routingEmailRoute *platformclientv2.Inboundroute) (*platformclientv2.Inboundroute, error) {
	return p.createRoutingEmailRouteAttr(ctx, p, routingEmailRoute)
}

// getRoutingEmailRoute retrieves all Genesys Cloud routing email route
func (p *routingEmailRouteProxy) getAllRoutingEmailRoute(ctx context.Context) (*[]platformclientv2.Inboundroute, error) {
	return p.getAllRoutingEmailRouteAttr(ctx, p)
}

// getRoutingEmailRouteIdByName returns a single Genesys Cloud routing email route by a name
func (p *routingEmailRouteProxy) getRoutingEmailRouteIdByName(ctx context.Context, name string) (id string, retryable bool, err error) {
	return p.getRoutingEmailRouteIdByNameAttr(ctx, p, name)
}

// getRoutingEmailRouteById returns a single Genesys Cloud routing email route by Id
func (p *routingEmailRouteProxy) getRoutingEmailRouteById(ctx context.Context, id string) (routingEmailRoute *platformclientv2.Inboundroute, statusCode int, err error) {
	return p.getRoutingEmailRouteByIdAttr(ctx, p, id)
}

// updateRoutingEmailRoute updates a Genesys Cloud routing email route
func (p *routingEmailRouteProxy) updateRoutingEmailRoute(ctx context.Context, id string, domainId string, routingEmailRoute *platformclientv2.Inboundroute) (*platformclientv2.Inboundroute, int, error) {
	return p.updateRoutingEmailRouteAttr(ctx, p, id, domainId, routingEmailRoute)
}

// deleteRoutingEmailRoute deletes a Genesys Cloud routing email route by Id
func (p *routingEmailRouteProxy) deleteRoutingEmailRoute(ctx context.Context, domainId string, id string) (statusCode int, err error) {
	return p.deleteRoutingEmailRouteAttr(ctx, p, domainId, id)
}

// createRoutingEmailRouteFn is an implementation function for creating a Genesys Cloud routing email route
func createRoutingEmailRouteFn(ctx context.Context, p *routingEmailRouteProxy, routingEmailRoute *platformclientv2.Inboundroute) (*platformclientv2.Inboundroute, error) {
	inboundRoute, _, err := p.routingApi.PostRoutingEmailDomainRoutes(*routingEmailRoute)
	if err != nil {
		return nil, fmt.Errorf("Failed to create routing email route: %s", err)
	}

	return inboundRoute, nil
}

// getAllRoutingEmailRouteFn is the implementation for retrieving all routing email route in Genesys Cloud
func getAllRoutingEmailRouteFn(ctx context.Context, p *routingEmailRouteProxy) (*[]platformclientv2.Inboundroute, error) {
	var allInboundRoutes []platformclientv2.Inboundroute
	const pageSize = 100

	for pageNum := 1; ; pageNum++ {
		domains, resp, getErr := p.routingApi.GetRoutingEmailDomains(pageSize, pageNum, false, "")
		if getErr != nil {
			return nil, fmt.Errorf("Failed to get routing email domains: %v %s", resp, getErr)
		}
		if domains.Entities == nil || len(*domains.Entities) == 0 {
			return &allInboundRoutes, nil
		}

	}

	inboundRoutes, _, err := p.routingApi.GetRoutingEmailDomainRoutes()
	if err != nil {
		return nil, fmt.Errorf("Failed to get inbound route: %v", err)
	}
	if inboundRoutes.Entities == nil || len(*inboundRoutes.Entities) == 0 {
		return &allInboundRoutes, nil
	}
	for _, inboundRoute := range *inboundRoutes.Entities {
		allInboundRoutes = append(allInboundRoutes, inboundRoute)
	}

	for pageNum := 2; pageNum <= *inboundRoutes.PageCount; pageNum++ {
		inboundRoutes, _, err := p.routingApi.GetRoutingEmailDomainRoutes()
		if err != nil {
			return nil, fmt.Errorf("Failed to get inbound route: %v", err)
		}

		if inboundRoutes.Entities == nil || len(*inboundRoutes.Entities) == 0 {
			break
		}

		for _, inboundRoute := range *inboundRoutes.Entities {
			allInboundRoutes = append(allInboundRoutes, inboundRoute)
		}
	}

	return &allInboundRoutes, nil
}

// getRoutingEmailRouteIdByNameFn is an implementation of the function to get a Genesys Cloud routing email route by name
func getRoutingEmailRouteIdByNameFn(ctx context.Context, p *routingEmailRouteProxy, name string) (id string, retryable bool, err error) {
	inboundRoutes, _, err := p.routingApi.GetRoutingEmailDomainRoutes()
	if err != nil {
		return "", false, err
	}

	if inboundRoutes.Entities == nil || len(*inboundRoutes.Entities) == 0 {
		return "", true, fmt.Errorf("No routing email route found with name %s", name)
	}

	for _, inboundRoute := range *inboundRoutes.Entities {
		if *inboundRoute.Name == name {
			log.Printf("Retrieved the routing email route id %s by name %s", *inboundRoute.Id, name)
			return *inboundRoute.Id, false, nil
		}
	}

	return "", true, fmt.Errorf("Unable to find routing email route with name %s", name)
}

// getRoutingEmailRouteByIdFn is an implementation of the function to get a Genesys Cloud routing email route by Id
func getRoutingEmailRouteByIdFn(ctx context.Context, p *routingEmailRouteProxy, id string) (routingEmailRoute *platformclientv2.Inboundroute, statusCode int, err error) {
	inboundRoute, resp, err := p.routingApi.GetRoutingEmailDomainRoute(id)
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("Failed to retrieve routing email route by id %s: %s", id, err)
	}

	return inboundRoute, resp.StatusCode, nil
}

// updateRoutingEmailRouteFn is an implementation of the function to update a Genesys Cloud routing email route
func updateRoutingEmailRouteFn(ctx context.Context, p *routingEmailRouteProxy, id string, domainId string, routingEmailRoute *platformclientv2.Inboundroute) (*platformclientv2.Inboundroute, int, error) {
	inboundRoute, resp, err := p.routingApi.PutRoutingEmailDomainRoute(domainId, id, *routingEmailRoute)
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("Failed to update routing email route: %s", err)
	}
	return inboundRoute, resp.StatusCode, nil
}

// deleteRoutingEmailRouteFn is an implementation function for deleting a Genesys Cloud routing email route
func deleteRoutingEmailRouteFn(ctx context.Context, p *routingEmailRouteProxy, domainId string, id string) (statusCode int, err error) {
	resp, err := p.routingApi.DeleteRoutingEmailDomainRoute(domainId, id)
	if err != nil {
		return resp.StatusCode, fmt.Errorf("Failed to delete routing email route: %s", err)
	}
	return resp.StatusCode, nil
}
