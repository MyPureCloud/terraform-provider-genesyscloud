package routing_email_domain

import (
	"context"
	"fmt"
	"log"
	rc "terraform-provider-genesyscloud/genesyscloud/resource_cache"

	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

var internalProxy *routingEmailDomainProxy

type getAllRoutingEmailDomainsFunc func(ctx context.Context, p *routingEmailDomainProxy) (*[]platformclientv2.Inbounddomain, *platformclientv2.APIResponse, error)
type createRoutingEmailDomainFunc func(ctx context.Context, p *routingEmailDomainProxy, inboundDomain *platformclientv2.Inbounddomain) (*platformclientv2.Inbounddomain, *platformclientv2.APIResponse, error)
type getRoutingEmailDomainByIdFunc func(ctx context.Context, p *routingEmailDomainProxy, id string) (*platformclientv2.Inbounddomain, *platformclientv2.APIResponse, error)
type getRoutingEmailDomainIdByNameFunc func(ctx context.Context, p *routingEmailDomainProxy, name string) (string, *platformclientv2.APIResponse, bool, error)
type updateRoutingEmailDomainFunc func(ctx context.Context, p *routingEmailDomainProxy, id string, inboundDomain *platformclientv2.Inbounddomainpatchrequest) (*platformclientv2.Inbounddomain, *platformclientv2.APIResponse, error)
type deleteRoutingEmailDomainFunc func(ctx context.Context, p *routingEmailDomainProxy, id string) (*platformclientv2.APIResponse, error)

// routingEmailDomainProxy contains all of the methods that call genesys cloud APIs.
type routingEmailDomainProxy struct {
	clientConfig                      *platformclientv2.Configuration
	routingApi                        *platformclientv2.RoutingApi
	createRoutingEmailDomainAttr      createRoutingEmailDomainFunc
	getAllRoutingEmailDomainsAttr     getAllRoutingEmailDomainsFunc
	getRoutingEmailDomainIdByNameAttr getRoutingEmailDomainIdByNameFunc
	getRoutingEmailDomainByIdAttr     getRoutingEmailDomainByIdFunc
	updateRoutingEmailDomainAttr      updateRoutingEmailDomainFunc
	deleteRoutingEmailDomainAttr      deleteRoutingEmailDomainFunc
	routingEmailDomainCache           rc.CacheInterface[platformclientv2.Inbounddomain]
}

// newRoutingEmailDomainProxy initializes the routing email domain proxy with all of the data needed to communicate with Genesys Cloud
func newRoutingEmailDomainProxy(clientConfig *platformclientv2.Configuration) *routingEmailDomainProxy {
	api := platformclientv2.NewRoutingApiWithConfig(clientConfig)
	routingEmailDomainCache := rc.NewResourceCache[platformclientv2.Inbounddomain]()
	return &routingEmailDomainProxy{
		clientConfig:                      clientConfig,
		routingApi:                        api,
		createRoutingEmailDomainAttr:      createRoutingEmailDomainFn,
		getAllRoutingEmailDomainsAttr:     getAllRoutingEmailDomainsFn,
		getRoutingEmailDomainIdByNameAttr: getRoutingEmailDomainIdByNameFn,
		getRoutingEmailDomainByIdAttr:     getRoutingEmailDomainByIdFn,
		updateRoutingEmailDomainAttr:      updateRoutingEmailDomainFn,
		deleteRoutingEmailDomainAttr:      deleteRoutingEmailDomainFn,
		routingEmailDomainCache:           routingEmailDomainCache,
	}
}

// getRoutingEmailDomainProxy acts as a singleton to for the internalProxy.  It also ensures
// that we can still proxy our tests by directly setting internalProxy package variable
func getRoutingEmailDomainProxy(clientConfig *platformclientv2.Configuration) *routingEmailDomainProxy {
	if internalProxy == nil {
		internalProxy = newRoutingEmailDomainProxy(clientConfig)
	}
	return internalProxy
}

func (p *routingEmailDomainProxy) getAllRoutingEmailDomains(ctx context.Context) (*[]platformclientv2.Inbounddomain, *platformclientv2.APIResponse, error) {
	return p.getAllRoutingEmailDomainsAttr(ctx, p)
}

// createRoutingEmailDomain creates a Genesys Cloud routing email domain
func (p *routingEmailDomainProxy) createRoutingEmailDomain(ctx context.Context, routingEmailDomain *platformclientv2.Inbounddomain) (*platformclientv2.Inbounddomain, *platformclientv2.APIResponse, error) {
	return p.createRoutingEmailDomainAttr(ctx, p, routingEmailDomain)
}

// getRoutingEmailDomainById returns a single Genesys Cloud routing email domain by Id
func (p *routingEmailDomainProxy) getRoutingEmailDomainById(ctx context.Context, id string) (*platformclientv2.Inbounddomain, *platformclientv2.APIResponse, error) {
	return p.getRoutingEmailDomainByIdAttr(ctx, p, id)
}

// getRoutingEmailDomainIdByName returns a single Genesys Cloud routing email domain by a name
func (p *routingEmailDomainProxy) getRoutingEmailDomainIdByName(ctx context.Context, name string) (string, *platformclientv2.APIResponse, bool, error) {
	return p.getRoutingEmailDomainIdByNameAttr(ctx, p, name)
}

// updateRoutingEmailDomain updates a Genesys Cloud routing email domain
func (p *routingEmailDomainProxy) updateRoutingEmailDomain(ctx context.Context, id string, routingEmailDomain *platformclientv2.Inbounddomainpatchrequest) (*platformclientv2.Inbounddomain, *platformclientv2.APIResponse, error) {
	return p.updateRoutingEmailDomainAttr(ctx, p, id, routingEmailDomain)
}

// deleteRoutingEmailDomain deletes a Genesys Cloud routing email domain by Id
func (p *routingEmailDomainProxy) deleteRoutingEmailDomain(ctx context.Context, id string) (*platformclientv2.APIResponse, error) {
	return p.deleteRoutingEmailDomainAttr(ctx, p, id)
}

func getAllRoutingEmailDomainsFn(ctx context.Context, p *routingEmailDomainProxy) (*[]platformclientv2.Inbounddomain, *platformclientv2.APIResponse, error) {
	var (
		allDomains []platformclientv2.Inbounddomain
		pageSize   = 100
		response   *platformclientv2.APIResponse
	)

	domains, resp, err := p.routingApi.GetRoutingEmailDomains(pageSize, 1, false, "")
	if err != nil {
		return nil, resp, fmt.Errorf("failed to get routing email domains error: %s", err)
	}

	if domains.Entities == nil || len(*domains.Entities) == 0 {
		return &allDomains, resp, nil
	}
	allDomains = append(allDomains, *domains.Entities...)

	for pageNum := 2; pageNum <= *domains.PageCount; pageNum++ {
		domains, resp, err := p.routingApi.GetRoutingEmailDomains(pageSize, pageNum, false, "")
		if err != nil {
			return nil, resp, fmt.Errorf("failed to get routing email domains error: %s", err)
		}

		response = resp
		if domains.Entities == nil || len(*domains.Entities) == 0 {
			return &allDomains, resp, nil
		}
		allDomains = append(allDomains, *domains.Entities...)
	}

	for _, domain := range allDomains {
		rc.SetCache(p.routingEmailDomainCache, *domain.Id, domain)
	}
	return &allDomains, response, nil
}

func createRoutingEmailDomainFn(ctx context.Context, p *routingEmailDomainProxy, routingEmailDomain *platformclientv2.Inbounddomain) (*platformclientv2.Inbounddomain, *platformclientv2.APIResponse, error) {
	return p.routingApi.PostRoutingEmailDomains(*routingEmailDomain)
}

func getRoutingEmailDomainByIdFn(ctx context.Context, p *routingEmailDomainProxy, id string) (*platformclientv2.Inbounddomain, *platformclientv2.APIResponse, error) {
	if domain := rc.GetCacheItem(p.routingEmailDomainCache, id); domain != nil {
		return domain, nil, nil
	}
	return p.routingApi.GetRoutingEmailDomain(id)
}

func getRoutingEmailDomainIdByNameFn(ctx context.Context, p *routingEmailDomainProxy, name string) (string, *platformclientv2.APIResponse, bool, error) {
	domains, resp, err := getAllRoutingEmailDomainsFn(ctx, p)
	if err != nil {
		return "", resp, false, err
	}

	if domains == nil || len(*domains) == 0 {
		return "", resp, true, fmt.Errorf("no routing email domain found with name %s", name)
	}

	for _, domain := range *domains {
		if *domain.Id == name {
			log.Printf("retrieved the routing email domain id %s by name %s", *domain.Id, name)
			return *domain.Id, resp, false, nil
		}
	}

	return "", resp, true, fmt.Errorf("unable to find routing email domain with name %s", name)
}

func updateRoutingEmailDomainFn(ctx context.Context, p *routingEmailDomainProxy, id string, routingEmailDomainReq *platformclientv2.Inbounddomainpatchrequest) (*platformclientv2.Inbounddomain, *platformclientv2.APIResponse, error) {
	return p.routingApi.PatchRoutingEmailDomain(id, *routingEmailDomainReq)
}

func deleteRoutingEmailDomainFn(ctx context.Context, p *routingEmailDomainProxy, id string) (*platformclientv2.APIResponse, error) {
	return p.routingApi.DeleteRoutingEmailDomain(id)
}
