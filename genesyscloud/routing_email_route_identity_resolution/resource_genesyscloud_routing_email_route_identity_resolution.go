package routing_email_route_identity_resolution

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v195/platformclientv2"
	consistencyChecker "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/consistency_checker"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/constants"
)

func getAllRoutingEmailRouteIdentityResolution(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(resourceExporter.ResourceIDMetaMap)
	proxy := getRoutingEmailRouteIdentityResolutionProxy(clientConfig)

	domains, resp, err := proxy.routingEmailRouteProxy.GetAllRoutingEmailRoute(ctx, "", "")
	if err != nil {
		return nil, util.BuildAPIDiagnosticError(ResourceType, "failed to list routing email routes for identity resolution export", resp)
	}
	if domains == nil {
		return resources, nil
	}

	for domainId, routes := range *domains {
		for _, route := range routes {
			if route.Id == nil || route.Name == nil {
				continue
			}

			config, getResp, getErr := proxy.getRoutingEmailRouteIdentityResolution(ctx, domainId, *route.Id)
			if getErr != nil {
				if util.IsStatus404(getResp) {
					continue
				}
				return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("failed to get identity resolution config for routing email route %s", *route.Id), getResp)
			}

			if isDefaultIdentityResolutionConfig(config) {
				continue
			}

			resources[domainId+"/"+*route.Id] = &resourceExporter.ResourceMeta{BlockLabel: *route.Name + "-identity-resolution"}
		}
	}

	return resources, nil
}

func createRoutingEmailRouteIdentityResolution(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	domainName := d.Get("domain_name").(string)
	routeId := d.Get("route_id").(string)
	log.Printf("creating identity resolution for routing email route %s", routeId)
	d.SetId(domainName + "/" + routeId)

	return updateRoutingEmailRouteIdentityResolution(ctx, d, meta)
}

func readRoutingEmailRouteIdentityResolution(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getRoutingEmailRouteIdentityResolutionProxy(sdkConfig)
	cc := consistencyChecker.NewConsistencyCheck(ctx, d, meta, ResourceRoutingEmailRouteIdentityResolution(), constants.ConsistencyChecks(), ResourceType)

	parts := strings.Split(d.Id(), "/")
	if len(parts) != 2 {
		return diag.Errorf("invalid id: %s", d.Id())
	}
	domainName, routeId := parts[0], parts[1]
	log.Printf("reading identity resolution for routing email route %s/%s", domainName, routeId)

	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		config, resp, getErr := proxy.getRoutingEmailRouteIdentityResolution(ctx, domainName, routeId)
		if getErr != nil {
			if util.IsStatus404(resp) {
				log.Printf("routing email route %s/%s not found, removing identity resolution from state", domainName, routeId)
				d.SetId("")
				return nil
			}
			return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("failed to read identity resolution for routing email route %s/%s: %s", domainName, routeId, getErr), resp))
		}

		_ = d.Set("domain_name", domainName)
		_ = d.Set("route_id", routeId)

		if config.ResolveIdentities != nil {
			_ = d.Set("resolve_identities", *config.ResolveIdentities)
		} else {
			_ = d.Set("resolve_identities", false)
		}
		if config.Division != nil && config.Division.Id != nil && !isUnassignedDivisionId(*config.Division.Id) {
			_ = d.Set("division_id", *config.Division.Id)
		} else {
			_ = d.Set("division_id", "")
		}

		log.Printf("read identity resolution for routing email route %s/%s", domainName, routeId)
		return cc.CheckState(d)
	})
}

func updateRoutingEmailRouteIdentityResolution(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getRoutingEmailRouteIdentityResolutionProxy(sdkConfig)
	parts := strings.Split(d.Id(), "/")
	if len(parts) != 2 {
		return diag.Errorf("invalid id: %s", d.Id())
	}
	domainName, routeId := parts[0], parts[1]

	config, err := buildIdentityResolutionRoutingEmailRouteConfig(d)
	if err != nil {
		return util.BuildDiagnosticError(ResourceType, "failed to build identity resolution config", err)
	}

	log.Printf("updating identity resolution for routing email route %s", routeId)
	_, resp, putErr := proxy.putRoutingEmailRouteIdentityResolution(ctx, domainName, routeId, config)
	if putErr != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("failed to update identity resolution for routing email route %s", routeId), resp)
	}

	log.Printf("updated identity resolution for routing email route %s/%s", domainName, routeId)
	return readRoutingEmailRouteIdentityResolution(ctx, d, meta)
}

func deleteRoutingEmailRouteIdentityResolution(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getRoutingEmailRouteIdentityResolutionProxy(sdkConfig)
	parts := strings.Split(d.Id(), "/")
	if len(parts) != 2 {
		return diag.Errorf("invalid id: %s", d.Id())
	}
	domainName, routeId := parts[0], parts[1]

	log.Printf("resetting identity resolution for routing email route %s to default", routeId)

	_, resp, getErr := proxy.getRoutingEmailRouteById(ctx, domainName, routeId)
	if getErr != nil {
		if util.IsStatus404(resp) {
			log.Printf("parent Email route %s/%s already deleted", domainName, routeId)
			return nil
		}
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("failed to verify routing email route %s/%s before resetting identity resolution", domainName, routeId), resp)
	}

	defaultConfig := buildDefaultIdentityResolutionRoutingEmailRouteConfig()
	_, putResp, putErr := proxy.putRoutingEmailRouteIdentityResolution(ctx, domainName, routeId, defaultConfig)
	if putErr != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("failed to reset identity resolution for routing email route %s", routeId), putResp)
	}

	log.Printf("reset identity resolution for routing email route %s/%s to default", domainName, routeId)
	return nil
}
