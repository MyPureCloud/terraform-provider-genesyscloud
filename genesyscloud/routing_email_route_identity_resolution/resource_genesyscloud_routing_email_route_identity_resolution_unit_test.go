package routing_email_route_identity_resolution

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v195/platformclientv2"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	routingEmailRoute "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/routing_email_route"
	"github.com/stretchr/testify/assert"
)

func TestUnitResourceRoutingEmailRouteIdentityResolutionUpdate(t *testing.T) {
	tDomainName := uuid.NewString()
	tRouteId := uuid.NewString()

	proxy := &routingEmailRouteIdentityResolutionProxy{}
	proxy.putRoutingEmailRouteIdentityResolutionAttr = func(ctx context.Context, p *routingEmailRouteIdentityResolutionProxy, domainName string, routeId string, config platformclientv2.Routeidentityresolutionconfig) (*platformclientv2.Routeidentityresolutionconfig, *platformclientv2.APIResponse, error) {
		assert.Equal(t, tDomainName, domainName)
		assert.Equal(t, tRouteId, routeId)
		assert.NotNil(t, config.ResolveIdentities)
		assert.Equal(t, false, *config.ResolveIdentities)

		apiResponse := platformclientv2.APIResponse{StatusCode: http.StatusOK}
		return &config, &apiResponse, nil
	}

	proxy.getRoutingEmailRouteIdentityResolutionAttr = func(ctx context.Context, p *routingEmailRouteIdentityResolutionProxy, domainName string, routeId string) (*platformclientv2.Routeidentityresolutionconfig, *platformclientv2.APIResponse, error) {
		resolveIdentities := false
		apiResponse := platformclientv2.APIResponse{StatusCode: http.StatusOK}
		return &platformclientv2.Routeidentityresolutionconfig{
			ResolveIdentities: &resolveIdentities,
		}, &apiResponse, nil
	}

	internalProxy = proxy
	defer func() { internalProxy = nil }()

	ctx := context.Background()
	gcloud := &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}
	resourceSchema := ResourceRoutingEmailRouteIdentityResolution().Schema
	resourceDataMap := buildIdentityResolutionResourceMap(tDomainName, tRouteId, false, "")

	d := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
	d.SetId(tDomainName + "/" + tRouteId)

	diag := updateRoutingEmailRouteIdentityResolution(ctx, d, gcloud)
	assert.False(t, diag.HasError(), diag)
	assert.Equal(t, tDomainName, d.Get("domain_name").(string))
	assert.Equal(t, tRouteId, d.Get("route_id").(string))
	assert.Equal(t, tDomainName+"/"+tRouteId, d.Id())
}

func TestUnitResourceRoutingEmailRouteIdentityResolutionRead(t *testing.T) {
	tDomainName := uuid.NewString()
	tRouteId := uuid.NewString()
	tDivisionId := uuid.NewString()

	proxy := &routingEmailRouteIdentityResolutionProxy{}
	proxy.getRoutingEmailRouteIdentityResolutionAttr = func(ctx context.Context, p *routingEmailRouteIdentityResolutionProxy, domainName string, routeId string) (*platformclientv2.Routeidentityresolutionconfig, *platformclientv2.APIResponse, error) {
		resolveIdentities := false
		apiResponse := platformclientv2.APIResponse{StatusCode: http.StatusOK}
		return &platformclientv2.Routeidentityresolutionconfig{
			ResolveIdentities: &resolveIdentities,
			Division: &platformclientv2.Writablestarrabledivision{
				Id: &tDivisionId,
			},
		}, &apiResponse, nil
	}

	internalProxy = proxy
	defer func() { internalProxy = nil }()

	ctx := context.Background()
	gcloud := &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}
	resourceSchema := ResourceRoutingEmailRouteIdentityResolution().Schema
	resourceDataMap := buildIdentityResolutionResourceMap(tDomainName, tRouteId, false, tDivisionId)

	d := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
	d.SetId(tDomainName + "/" + tRouteId)

	diag := readRoutingEmailRouteIdentityResolution(ctx, d, gcloud)
	assert.False(t, diag.HasError(), diag)
	assert.Equal(t, tDomainName, d.Get("domain_name").(string))
	assert.Equal(t, tRouteId, d.Get("route_id").(string))
	assert.Equal(t, tDomainName+"/"+tRouteId, d.Id())

	assert.Equal(t, false, d.Get("resolve_identities"))
	assert.Equal(t, tDivisionId, d.Get("division_id"))
}

func TestUnitResourceRoutingEmailRouteIdentityResolutionDelete(t *testing.T) {
	tDomainName := uuid.NewString()
	tRouteId := uuid.NewString()

	proxy := &routingEmailRouteIdentityResolutionProxy{}
	proxy.getRoutingEmailRouteByIdAttr = func(ctx context.Context, p *routingEmailRouteIdentityResolutionProxy, domainName string, routeId string) (*platformclientv2.Inboundroute, *platformclientv2.APIResponse, error) {
		apiResponse := platformclientv2.APIResponse{StatusCode: http.StatusOK}
		return &platformclientv2.Inboundroute{
			Id: &tRouteId,
		}, &apiResponse, nil
	}

	proxy.putRoutingEmailRouteIdentityResolutionAttr = func(ctx context.Context, p *routingEmailRouteIdentityResolutionProxy, domainName string, routeId string, config platformclientv2.Routeidentityresolutionconfig) (*platformclientv2.Routeidentityresolutionconfig, *platformclientv2.APIResponse, error) {
		assert.Equal(t, tDomainName, domainName)
		assert.Equal(t, tRouteId, routeId)
		assert.NotNil(t, config.ResolveIdentities)
		assert.Equal(t, true, *config.ResolveIdentities)
		assert.Nil(t, config.Division)

		apiResponse := platformclientv2.APIResponse{StatusCode: http.StatusOK}
		return &config, &apiResponse, nil
	}

	internalProxy = proxy
	defer func() { internalProxy = nil }()

	ctx := context.Background()
	gcloud := &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}
	resourceSchema := ResourceRoutingEmailRouteIdentityResolution().Schema
	resourceDataMap := buildIdentityResolutionResourceMap(tDomainName, tRouteId, false, "")

	d := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
	d.SetId(tDomainName + "/" + tRouteId)

	diag := deleteRoutingEmailRouteIdentityResolution(ctx, d, gcloud)
	assert.False(t, diag.HasError(), diag)
}

func TestUnitGetAllRoutingEmailRouteIdentityResolution(t *testing.T) {
	defaultDomainName := uuid.NewString()
	defaultRouteId := uuid.NewString()
	defaultRouteName := "Default Route"
	customDomainName := uuid.NewString()
	customRouteId := uuid.NewString()
	customRouteName := "Custom Route"
	notFoundDomainName := uuid.NewString()
	notFoundRouteId := uuid.NewString()
	notFoundRouteName := "Not Found Route"

	resolveTrue := true
	resolveFalse := false
	divisionId := uuid.NewString()

	routeProxy := &routingEmailRoute.RoutingEmailRouteProxy{}
	routeProxy.GetAllRoutingEmailRouteAttr = func(ctx context.Context, _ *routingEmailRoute.RoutingEmailRouteProxy, domainId string, pattern string) (*map[string][]platformclientv2.Inboundroute, *platformclientv2.APIResponse, error) {
		return &map[string][]platformclientv2.Inboundroute{
			defaultDomainName:  {{Id: &defaultRouteId, Name: &defaultRouteName}},
			customDomainName:   {{Id: &customRouteId, Name: &customRouteName}},
			notFoundDomainName: {{Id: &notFoundRouteId, Name: &notFoundRouteName}},
		}, nil, nil
	}

	proxy := &routingEmailRouteIdentityResolutionProxy{
		routingEmailRouteProxy: routeProxy,
	}
	proxy.getRoutingEmailRouteIdentityResolutionAttr = func(ctx context.Context, p *routingEmailRouteIdentityResolutionProxy, domainName string, routeId string) (*platformclientv2.Routeidentityresolutionconfig, *platformclientv2.APIResponse, error) {
		switch routeId {
		case defaultRouteId:
			apiResponse := platformclientv2.APIResponse{StatusCode: http.StatusOK}
			return &platformclientv2.Routeidentityresolutionconfig{
				ResolveIdentities: &resolveTrue,
			}, &apiResponse, nil
		case customRouteId:
			apiResponse := platformclientv2.APIResponse{StatusCode: http.StatusOK}
			return &platformclientv2.Routeidentityresolutionconfig{
				ResolveIdentities: &resolveFalse,
				Division: &platformclientv2.Writablestarrabledivision{
					Id: &divisionId,
				},
			}, &apiResponse, nil
		case notFoundRouteId:
			apiResponse := platformclientv2.APIResponse{StatusCode: http.StatusNotFound}
			return nil, &apiResponse, fmt.Errorf("not found")
		default:
			t.Fatalf("unexpected Route ID %s", routeId)
			return nil, nil, nil
		}
	}

	internalProxy = proxy
	defer func() { internalProxy = nil }()

	ctx := context.Background()
	resources, diag := getAllRoutingEmailRouteIdentityResolution(ctx, &platformclientv2.Configuration{})

	assert.False(t, diag.HasError())
	assert.Len(t, resources, 1)

	customKey := customDomainName + "/" + customRouteId
	assert.Contains(t, resources, customKey)
	assert.Equal(t, customRouteName+"-identity-resolution", resources[customKey].BlockLabel)
}

func TestUnitGetAllRoutingEmailRouteIdentityResolutionListError(t *testing.T) {
	routeProxy := &routingEmailRoute.RoutingEmailRouteProxy{}
	routeProxy.GetAllRoutingEmailRouteAttr = func(ctx context.Context, _ *routingEmailRoute.RoutingEmailRouteProxy, domainId string, pattern string) (*map[string][]platformclientv2.Inboundroute, *platformclientv2.APIResponse, error) {
		return nil, nil, fmt.Errorf("mock list error")
	}

	proxy := &routingEmailRouteIdentityResolutionProxy{
		routingEmailRouteProxy: routeProxy,
	}

	internalProxy = proxy
	defer func() { internalProxy = nil }()

	ctx := context.Background()
	resources, diag := getAllRoutingEmailRouteIdentityResolution(ctx, &platformclientv2.Configuration{})

	assert.True(t, diag.HasError())
	assert.Nil(t, resources)
}

func buildIdentityResolutionResourceMap(domainName, routeId string, resolveIdentities bool, divisionId string) map[string]interface{} {
	return map[string]interface{}{
		"domain_name":        domainName,
		"route_id":           routeId,
		"resolve_identities": resolveIdentities,
		"division_id":        divisionId,
	}
}
