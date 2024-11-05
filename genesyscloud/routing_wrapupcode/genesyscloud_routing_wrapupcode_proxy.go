package routing_wrapupcode

import (
	"context"
	"fmt"
	"log"
	rc "terraform-provider-genesyscloud/genesyscloud/resource_cache"

	"github.com/mypurecloud/platform-client-sdk-go/v146/platformclientv2"
)

/*
The file genesyscloud_routing_wrapupcode_proxy.go manages the interaction between our software and
the Genesys Cloud SDK. Within this file, we define proxy structures and methods.
We employ a technique called composition for each function on the proxy. This means that each function
is built by combining smaller, independent parts. One advantage of this approach is that it allows us
to isolate and test individual functions more easily. For testing purposes, we can replace or
simulate these smaller parts, known as stubs, to ensure that each function behaves correctly in different scenarios.
*/

// internalProxy holds a proxy instance that can be used throughout the package
var internalProxy *routingWrapupcodeProxy

// Type definitions for each func on our proxy so we can easily mock them out later
type createRoutingWrapupcodeFunc func(ctx context.Context, p *routingWrapupcodeProxy, wrapupcode *platformclientv2.Wrapupcoderequest) (*platformclientv2.Wrapupcode, *platformclientv2.APIResponse, error)
type getAllRoutingWrapupcodeFunc func(ctx context.Context, p *routingWrapupcodeProxy) (*[]platformclientv2.Wrapupcode, *platformclientv2.APIResponse, error)
type getRoutingWrapupcodeIdByNameFunc func(ctx context.Context, p *routingWrapupcodeProxy, name string) (id string, retryable bool, response *platformclientv2.APIResponse, err error)
type getRoutingWrapupcodeByIdFunc func(ctx context.Context, p *routingWrapupcodeProxy, id string) (wrapupcode *platformclientv2.Wrapupcode, response *platformclientv2.APIResponse, err error)
type updateRoutingWrapupcodeFunc func(ctx context.Context, p *routingWrapupcodeProxy, id string, wrapupcode *platformclientv2.Wrapupcoderequest) (*platformclientv2.Wrapupcode, *platformclientv2.APIResponse, error)
type deleteRoutingWrapupcodeFunc func(ctx context.Context, p *routingWrapupcodeProxy, id string) (*platformclientv2.APIResponse, error)

/*
The routingWrapupcodeProxy struct holds all the methods responsible for making calls to
the Genesys Cloud APIs. This means that within this struct, you'll find all the functions designed
to interact directly with the various features and services offered by Genesys Cloud,
enabling this terraform provider software to perform tasks like retrieving data, updating information,
or triggering actions within the Genesys Cloud environment.
*/
type routingWrapupcodeProxy struct {
	clientConfig                     *platformclientv2.Configuration
	routingApi                       *platformclientv2.RoutingApi
	createRoutingWrapupcodeAttr      createRoutingWrapupcodeFunc
	getAllRoutingWrapupcodeAttr      getAllRoutingWrapupcodeFunc
	getRoutingWrapupcodeIdByNameAttr getRoutingWrapupcodeIdByNameFunc
	getRoutingWrapupcodeByIdAttr     getRoutingWrapupcodeByIdFunc
	updateRoutingWrapupcodeAttr      updateRoutingWrapupcodeFunc
	deleteRoutingWrapupcodeAttr      deleteRoutingWrapupcodeFunc
	routingWrapupcodesCache          rc.CacheInterface[platformclientv2.Wrapupcode] //Define the cache for routing wrapupcode resource
}

/*
The function newRoutingWrapupcodeProxy sets up the routing wrapupcodes proxy by providing it
with all the necessary information to communicate effectively with Genesys Cloud.
This includes configuring the proxy with the required data and settings so that it can interact
seamlessly with the Genesys Cloud platform.
*/
func newRoutingWrapupcodeProxy(clientConfig *platformclientv2.Configuration) *routingWrapupcodeProxy {
	api := platformclientv2.NewRoutingApiWithConfig(clientConfig)                 // NewArchitectApiWithConfig creates an Genesyc Cloud API instance using the provided configuration
	routingWrapupcodesCache := rc.NewResourceCache[platformclientv2.Wrapupcode]() // Create Cache for routing wrapupcode resource
	return &routingWrapupcodeProxy{
		clientConfig:                     clientConfig,
		routingApi:                       api,
		routingWrapupcodesCache:          routingWrapupcodesCache,
		createRoutingWrapupcodeAttr:      createRoutingWrapupcodeFn,
		getAllRoutingWrapupcodeAttr:      getAllRoutingWrapupcodeFn,
		getRoutingWrapupcodeIdByNameAttr: getRoutingWrapupcodeIdByNameFn,
		getRoutingWrapupcodeByIdAttr:     getRoutingWrapupcodeByIdFn,
		updateRoutingWrapupcodeAttr:      updateRoutingWrapupcodeFn,
		deleteRoutingWrapupcodeAttr:      deleteRoutingWrapupcodeFn,
	}
}

/*
The function getRoutingWrapupcodeProxy serves a dual purpose: first, it functions as a singleton for
the internalProxy, meaning it ensures that only one instance of the internalProxy exists. Second,
it enables us to proxy our tests by allowing us to directly set the internalProxy package variable.
This ensures consistency and control in managing the internalProxy across our codebase, while also
facilitating efficient testing by providing a straightforward way to substitute the proxy for testing purposes.
*/
func getRoutingWrapupcodeProxy(clientConfig *platformclientv2.Configuration) *routingWrapupcodeProxy {
	if internalProxy == nil {
		internalProxy = newRoutingWrapupcodeProxy(clientConfig)
	}
	return internalProxy
}

// createRoutingWrapupcode creates a Genesys Cloud routing wrapupcodes
func (p *routingWrapupcodeProxy) createRoutingWrapupcode(ctx context.Context, routingWrapupcode *platformclientv2.Wrapupcoderequest) (*platformclientv2.Wrapupcode, *platformclientv2.APIResponse, error) {
	return p.createRoutingWrapupcodeAttr(ctx, p, routingWrapupcode)
}

// getRoutingWrapupcode retrieves all Genesys Cloud routing wrapupcodes
func (p *routingWrapupcodeProxy) getAllRoutingWrapupcode(ctx context.Context) (*[]platformclientv2.Wrapupcode, *platformclientv2.APIResponse, error) {
	return p.getAllRoutingWrapupcodeAttr(ctx, p)
}

// getRoutingWrapupcodeIdByName returns a single Genesys Cloud routing wrapupcodes by a name
func (p *routingWrapupcodeProxy) getRoutingWrapupcodeIdByName(ctx context.Context, name string) (id string, retryable bool, response *platformclientv2.APIResponse, err error) {
	return p.getRoutingWrapupcodeIdByNameAttr(ctx, p, name)
}

// getRoutingWrapupcodeById returns a single Genesys Cloud routing wrapupcodes by Id
func (p *routingWrapupcodeProxy) getRoutingWrapupcodeById(ctx context.Context, id string) (routingWrapupcode *platformclientv2.Wrapupcode, response *platformclientv2.APIResponse, err error) {
	if wrapupcode := rc.GetCacheItem(p.routingWrapupcodesCache, id); wrapupcode != nil { // Get the wrapupcode from the cache, if not there in the cache then call p.getRoutingWrapupcodeByIdAttr()
		return wrapupcode, nil, nil
	}
	return p.getRoutingWrapupcodeByIdAttr(ctx, p, id)
}

// updateRoutingWrapupcode updates a Genesys Cloud routing wrapupcodes
func (p *routingWrapupcodeProxy) updateRoutingWrapupcode(ctx context.Context, id string, routingWrapupcode *platformclientv2.Wrapupcoderequest) (*platformclientv2.Wrapupcode, *platformclientv2.APIResponse, error) {
	return p.updateRoutingWrapupcodeAttr(ctx, p, id, routingWrapupcode)
}

// deleteRoutingWrapupcode deletes a Genesys Cloud routing wrapupcodes by Id
func (p *routingWrapupcodeProxy) deleteRoutingWrapupcode(ctx context.Context, id string) (*platformclientv2.APIResponse, error) {
	return p.deleteRoutingWrapupcodeAttr(ctx, p, id)
}

// createRoutingWrapupcodeFn is an implementation function for creating a Genesys Cloud routing wrapupcodes
func createRoutingWrapupcodeFn(ctx context.Context, p *routingWrapupcodeProxy, routingWrapupcode *platformclientv2.Wrapupcoderequest) (*platformclientv2.Wrapupcode, *platformclientv2.APIResponse, error) {
	return p.routingApi.PostRoutingWrapupcodes(*routingWrapupcode)
}

// getRoutingWrapupcodeByIdFn is an implementation of the function to get a Genesys Cloud routing wrapupcodes by Id
func getRoutingWrapupcodeByIdFn(ctx context.Context, p *routingWrapupcodeProxy, id string) (routingWrapupcode *platformclientv2.Wrapupcode, response *platformclientv2.APIResponse, err error) {
	return p.routingApi.GetRoutingWrapupcode(id)
}

// updateRoutingWrapupcodeFn is an implementation of the function to update a Genesys Cloud routing wrapupcodes
func updateRoutingWrapupcodeFn(ctx context.Context, p *routingWrapupcodeProxy, id string, routingWrapupcode *platformclientv2.Wrapupcoderequest) (*platformclientv2.Wrapupcode, *platformclientv2.APIResponse, error) {
	return p.routingApi.PutRoutingWrapupcode(id, *routingWrapupcode)
}

// deleteRoutingWrapupcodeFn is an implementation function for deleting a Genesys Cloud routing wrapupcodes
func deleteRoutingWrapupcodeFn(ctx context.Context, p *routingWrapupcodeProxy, id string) (*platformclientv2.APIResponse, error) {
	resp, err := p.routingApi.DeleteRoutingWrapupcode(id)
	if err != nil {
		return resp, err
	}
	rc.DeleteCacheItem(p.routingWrapupcodesCache, id)
	return nil, nil
}

// getAllRoutingWrapupcodeFn is the implementation for retrieving all routing wrapupcodes in Genesys Cloud
func getAllRoutingWrapupcodeFn(ctx context.Context, p *routingWrapupcodeProxy) (*[]platformclientv2.Wrapupcode, *platformclientv2.APIResponse, error) {
	var allWrapupcodes []platformclientv2.Wrapupcode
	const pageSize = 100

	wrapupcodes, apiResponse, err := p.routingApi.GetRoutingWrapupcodes(pageSize, 1, "", "", "", []string{}, []string{})
	if err != nil {
		return nil, apiResponse, fmt.Errorf("Failed to get routing wrapupcode : %v", err)
	}

	if wrapupcodes == nil || wrapupcodes.Entities == nil || len(*wrapupcodes.Entities) == 0 {
		return &allWrapupcodes, apiResponse, nil
	}

	allWrapupcodes = append(allWrapupcodes, *wrapupcodes.Entities...)

	for pageNum := 2; pageNum <= *wrapupcodes.PageCount; pageNum++ {
		wrapupcodes, apiResponse, err := p.routingApi.GetRoutingWrapupcodes(pageSize, pageNum, "", "", "", []string{}, []string{})
		if err != nil {
			return nil, apiResponse, fmt.Errorf("Failed to get routing wrapupcode : %v", err)
		}

		if wrapupcodes == nil || wrapupcodes.Entities == nil || len(*wrapupcodes.Entities) == 0 {
			break
		}

		allWrapupcodes = append(allWrapupcodes, *wrapupcodes.Entities...)
	}

	// Cache the routing wrapupcodes resource into the p.routingWrapupcodesCache for later use
	for _, wrapupcode := range allWrapupcodes {
		rc.SetCache(p.routingWrapupcodesCache, *wrapupcode.Id, wrapupcode)
	}

	return &allWrapupcodes, apiResponse, nil
}

// getRoutingWrapupcodeIdByNameFn is an implementation of the function to get a Genesys Cloud routing wrapupcodes by name
func getRoutingWrapupcodeIdByNameFn(ctx context.Context, p *routingWrapupcodeProxy, name string) (id string, retryable bool, response *platformclientv2.APIResponse, err error) {
	wrapupcodes, apiResponse, err := getAllRoutingWrapupcodeFn(ctx, p)
	if err != nil {
		return "", false, apiResponse, err
	}

	if wrapupcodes == nil || len(*wrapupcodes) == 0 {
		return "", true, apiResponse, fmt.Errorf("No routing wrapupcodes found with name %s", name)
	}

	for _, wrapupcode := range *wrapupcodes {
		if *wrapupcode.Name == name {
			log.Printf("Retrieved the routing wrapupcodes id %s by name %s", *wrapupcode.Id, name)
			return *wrapupcode.Id, false, apiResponse, nil
		}
	}

	return "", true, apiResponse, fmt.Errorf("Unable to find routing wrapupcodes with name %s", name)
}
