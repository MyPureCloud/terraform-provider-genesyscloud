package routing_email_route

import (
	"context"
	"fmt"
	"log"
	"reflect"

	"github.com/mypurecloud/platform-client-sdk-go/v157/platformclientv2"
)

/*
The genesyscloud_routing_email_route_proxy.go file contains the proxy structures and methods that interact
with the Genesys Cloud SDK. We use composition here for each function on the proxy so individual functions can be stubbed
out during testing.
*/

// internalProxy holds a proxy instance that can be used throughout the package
var internalProxy *routingEmailRouteProxy

// Type definitions for each func on our proxy so we can easily mock them out later
type createRoutingEmailRouteFunc func(ctx context.Context, p *routingEmailRouteProxy, domainId string, inboundRoute *platformclientv2.Inboundroute) (*platformclientv2.Inboundroute, *platformclientv2.APIResponse, error)
type getAllRoutingEmailRouteFunc func(ctx context.Context, p *routingEmailRouteProxy, domainId string, name string) (*map[string][]platformclientv2.Inboundroute, *platformclientv2.APIResponse, error)
type getRoutingEmailRouteIdByPatternFunc func(ctx context.Context, p *routingEmailRouteProxy, pattern string, domainId string) (id string, retryable bool, response *platformclientv2.APIResponse, err error)
type getRoutingEmailRouteByIdFunc func(ctx context.Context, p *routingEmailRouteProxy, domainId string, id string) (inboundRoute *platformclientv2.Inboundroute, response *platformclientv2.APIResponse, err error)
type updateRoutingEmailRouteFunc func(ctx context.Context, p *routingEmailRouteProxy, id string, domainId string, inboundRoute *platformclientv2.Inboundroute) (*platformclientv2.Inboundroute, *platformclientv2.APIResponse, error)
type deleteRoutingEmailRouteFunc func(ctx context.Context, p *routingEmailRouteProxy, domainId string, id string) (response *platformclientv2.APIResponse, err error)

// routingEmailRouteProxy contains all methods that call genesys cloud APIs.
type routingEmailRouteProxy struct {
	clientConfig                        *platformclientv2.Configuration
	routingApi                          *platformclientv2.RoutingApi
	createRoutingEmailRouteAttr         createRoutingEmailRouteFunc
	getAllRoutingEmailRouteAttr         getAllRoutingEmailRouteFunc
	getRoutingEmailRouteIdByPatternAttr getRoutingEmailRouteIdByPatternFunc
	getRoutingEmailRouteByIdAttr        getRoutingEmailRouteByIdFunc
	updateRoutingEmailRouteAttr         updateRoutingEmailRouteFunc
	deleteRoutingEmailRouteAttr         deleteRoutingEmailRouteFunc
}

// newRoutingEmailRouteProxy initializes the routing email route proxy with all data needed to communicate with Genesys Cloud
func newRoutingEmailRouteProxy(clientConfig *platformclientv2.Configuration) *routingEmailRouteProxy {
	api := platformclientv2.NewRoutingApiWithConfig(clientConfig)
	return &routingEmailRouteProxy{
		clientConfig:                        clientConfig,
		routingApi:                          api,
		createRoutingEmailRouteAttr:         createRoutingEmailRouteFn,
		getAllRoutingEmailRouteAttr:         getAllRoutingEmailRouteFn,
		getRoutingEmailRouteIdByPatternAttr: getRoutingEmailRouteIdByPatternFn,
		getRoutingEmailRouteByIdAttr:        getRoutingEmailRouteByIdFn,
		updateRoutingEmailRouteAttr:         updateRoutingEmailRouteFn,
		deleteRoutingEmailRouteAttr:         deleteRoutingEmailRouteFn,
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
func (p *routingEmailRouteProxy) createRoutingEmailRoute(ctx context.Context, domainId string, routingEmailRoute *platformclientv2.Inboundroute) (*platformclientv2.Inboundroute, *platformclientv2.APIResponse, error) {
	return p.createRoutingEmailRouteAttr(ctx, p, domainId, routingEmailRoute)
}

// getRoutingEmailRoute retrieves all Genesys Cloud routing email route
func (p *routingEmailRouteProxy) getAllRoutingEmailRoute(ctx context.Context, domainId string, name string) (*map[string][]platformclientv2.Inboundroute, *platformclientv2.APIResponse, error) {
	return p.getAllRoutingEmailRouteAttr(ctx, p, domainId, name)
}

// getRoutingEmailRouteIdByName returns a single Genesys Cloud routing email route by a pattern
func (p *routingEmailRouteProxy) getRoutingEmailRouteIdByPattern(ctx context.Context, pattern string, domainId string) (id string, retryable bool, response *platformclientv2.APIResponse, err error) {
	return p.getRoutingEmailRouteIdByPatternAttr(ctx, p, pattern, domainId)
}

// getRoutingEmailRouteById returns a single Genesys Cloud routing email route by Id
func (p *routingEmailRouteProxy) getRoutingEmailRouteById(ctx context.Context, domainId string, id string) (routingEmailRoute *platformclientv2.Inboundroute, response *platformclientv2.APIResponse, err error) {
	return p.getRoutingEmailRouteByIdAttr(ctx, p, domainId, id)
}

// updateRoutingEmailRoute updates a Genesys Cloud routing email route
func (p *routingEmailRouteProxy) updateRoutingEmailRoute(ctx context.Context, id string, domainId string, routingEmailRoute *platformclientv2.Inboundroute) (*platformclientv2.Inboundroute, *platformclientv2.APIResponse, error) {
	return p.updateRoutingEmailRouteAttr(ctx, p, id, domainId, routingEmailRoute)
}

// deleteRoutingEmailRoute deletes a Genesys Cloud routing email route by Id
func (p *routingEmailRouteProxy) deleteRoutingEmailRoute(ctx context.Context, domainId string, id string) (response *platformclientv2.APIResponse, err error) {
	return p.deleteRoutingEmailRouteAttr(ctx, p, domainId, id)
}

func getAllRoutingEmailRouteByDomainIdFn(_ context.Context, p *routingEmailRouteProxy, domains []platformclientv2.Inbounddomain, pattern string) (*map[string][]platformclientv2.Inboundroute, *platformclientv2.APIResponse, error) {
	const pageSize = 100
	var allInboundRoutes = make(map[string][]platformclientv2.Inboundroute)
	var apiResponse *platformclientv2.APIResponse
	for _, domain := range domains {

		var allDomainRoutes = make([]platformclientv2.Inboundroute, 0)

		for pageNum := 1; ; pageNum++ {
			routes, resp, err := p.routingApi.GetRoutingEmailDomainRoutes(*domain.Id, pageSize, pageNum, pattern)
			if err != nil {
				apiResponse = resp
				return nil, apiResponse, fmt.Errorf("failed to get routing email route: %s", err.Error())
			}
			if routes.Entities == nil || len(*routes.Entities) == 0 {
				break
			}
			allDomainRoutes = append(allDomainRoutes, *routes.Entities...)
		}
		allInboundRoutes[*domain.Id] = allDomainRoutes
	}
	return &allInboundRoutes, apiResponse, nil
}

// getAllRoutingEmailRouteFn is the implementation for retrieving all routing email route in Genesys Cloud
func getAllRoutingEmailRouteFn(ctx context.Context, p *routingEmailRouteProxy, domainId, pattern string) (*map[string][]platformclientv2.Inboundroute, *platformclientv2.APIResponse, error) {
	const pageSize = 100
	var apiResponse *platformclientv2.APIResponse

	var allDomains = make([]platformclientv2.Inbounddomain, 0)
	for pageNum := 1; ; pageNum++ {
		domains, resp, err := p.routingApi.GetRoutingEmailDomains(pageSize, pageNum, false, domainId)
		apiResponse = resp
		if err != nil {
			return nil, resp, fmt.Errorf("failed to get routing email domains: %s", err.Error())
		}
		if domains.Entities == nil || len(*domains.Entities) == 0 {
			break
		}
		allDomains = append(allDomains, *domains.Entities...)
	}

	if len(allDomains) == 0 {
		return nil, apiResponse, nil
	}

	// Get all routes for each domain
	routes, resp, err := getAllRoutingEmailRouteByDomainIdFn(ctx, p, allDomains, pattern)
	if err != nil {
		return nil, resp, fmt.Errorf("failed to get routing email domains: %s", err.Error())
	}

	if routes == nil {
		log.Printf("No routing email routes found. domainId: '%s', pattern: '%s'", domainId, pattern)
		return nil, resp, nil
	}

	log.Printf("Returning routes for domains: %v", reflect.ValueOf(*routes).MapKeys())
	return routes, resp, nil
}

// createRoutingEmailRouteFn is an implementation function for creating a Genesys Cloud routing email route
func createRoutingEmailRouteFn(_ context.Context, p *routingEmailRouteProxy, domainId string, routingEmailRoute *platformclientv2.Inboundroute) (*platformclientv2.Inboundroute, *platformclientv2.APIResponse, error) {
	inboundRoute, resp, err := p.routingApi.PostRoutingEmailDomainRoutes(domainId, *routingEmailRoute)
	if err != nil {
		return nil, resp, fmt.Errorf("failed to create routing email route: %s", err)
	}
	return inboundRoute, resp, nil
}

// updateRoutingEmailRouteFn is an implementation of the function to update a Genesys Cloud routing email route
func updateRoutingEmailRouteFn(_ context.Context, p *routingEmailRouteProxy, id string, domainId string, routingEmailRoute *platformclientv2.Inboundroute) (*platformclientv2.Inboundroute, *platformclientv2.APIResponse, error) {
	inboundRoute, resp, err := p.routingApi.PutRoutingEmailDomainRoute(domainId, id, *routingEmailRoute)
	if err != nil {
		return nil, resp, fmt.Errorf("failed to update routing email route: %s", err)
	}
	return inboundRoute, resp, nil
}

// deleteRoutingEmailRouteFn is an implementation function for deleting a Genesys Cloud routing email route
func deleteRoutingEmailRouteFn(_ context.Context, p *routingEmailRouteProxy, domainId string, id string) (*platformclientv2.APIResponse, error) {
	resp, err := p.routingApi.DeleteRoutingEmailDomainRoute(domainId, id)
	if err != nil {
		return resp, fmt.Errorf("failed to delete routing email route: %s", err)
	}
	return resp, nil
}

// getRoutingEmailRouteByIdFn is an implementation of the function to get a Genesys Cloud routing email route by Id
func getRoutingEmailRouteByIdFn(_ context.Context, p *routingEmailRouteProxy, domainId string, id string) (*platformclientv2.Inboundroute, *platformclientv2.APIResponse, error) {
	inboundRoute, resp, err := p.routingApi.GetRoutingEmailDomainRoute(domainId, id)
	if err != nil {
		return nil, resp, fmt.Errorf("failed to retrieve routing email route by id %s: %s", id, err)
	}
	return inboundRoute, resp, nil
}

// getRoutingEmailRouteIdByNameFn is an implementation of the function to get a Genesys Cloud routing email route by name
func getRoutingEmailRouteIdByPatternFn(ctx context.Context, p *routingEmailRouteProxy, pattern string, domainId string) (string, bool, *platformclientv2.APIResponse, error) {
	inboundRoutesMap, resp, err := getAllRoutingEmailRouteFn(ctx, p, domainId, pattern)
	if err != nil {
		return "", false, resp, err
	}

	if inboundRoutesMap == nil || len(*inboundRoutesMap) == 0 {
		return "", true, resp, fmt.Errorf("no routing email route found with pattern %s", pattern)
	}

	for _, inboundRoutes := range *inboundRoutesMap {
		for _, inboundRoute := range inboundRoutes {
			if *inboundRoute.Pattern == pattern {
				log.Printf("Retrieved the routing email route id %s by pattern %s for DomainID %s", *inboundRoute.Id, pattern, domainId)
				return *inboundRoute.Id, false, resp, nil
			}
		}
	}

	return "", true, resp, fmt.Errorf("unable to find routing email route with name %s", pattern)
}
