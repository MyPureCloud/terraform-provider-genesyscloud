package conversations_messaging_integrations_instagram_identity_resolution

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v199/platformclientv2"
	instagram "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/conversations_messaging_integrations_instagram"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/stretchr/testify/assert"
)

func TestUnitResourceConversationsMessagingIntegrationsInstagramIdentityResolutionUpdate(t *testing.T) {
	tInstagramIntegrationId := uuid.NewString()

	proxy := &conversationsMessagingIntegrationsInstagramIdentityResolutionProxy{}
	proxy.putConversationsMessagingIntegrationsInstagramIdentityResolutionAttr = func(ctx context.Context, p *conversationsMessagingIntegrationsInstagramIdentityResolutionProxy, instagramIntegrationId string, config *platformclientv2.Instagramidentityresolutionconfig) (*platformclientv2.Instagramidentityresolutionconfig, *platformclientv2.APIResponse, error) {
		assert.Equal(t, tInstagramIntegrationId, instagramIntegrationId)
		assert.NotNil(t, config.ResolveIdentities)
		assert.Equal(t, false, *config.ResolveIdentities)

		apiResponse := platformclientv2.APIResponse{StatusCode: http.StatusOK}
		return config, &apiResponse, nil
	}

	proxy.getConversationsMessagingIntegrationsInstagramIdentityResolutionAttr = func(ctx context.Context, p *conversationsMessagingIntegrationsInstagramIdentityResolutionProxy, instagramIntegrationId string) (*platformclientv2.Instagramidentityresolutionconfig, *platformclientv2.APIResponse, error) {
		resolveIdentities := false
		apiResponse := platformclientv2.APIResponse{StatusCode: http.StatusOK}
		return &platformclientv2.Instagramidentityresolutionconfig{
			ResolveIdentities: &resolveIdentities,
		}, &apiResponse, nil
	}

	internalProxy = proxy
	defer func() { internalProxy = nil }()

	ctx := context.Background()
	gcloud := &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}
	resourceSchema := ResourceConversationsMessagingIntegrationsInstagramIdentityResolution().Schema
	resourceDataMap := buildIdentityResolutionResourceMap(tInstagramIntegrationId, false, "")

	d := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
	d.SetId(tInstagramIntegrationId)

	diag := updateConversationsMessagingIntegrationsInstagramIdentityResolution(ctx, d, gcloud)
	assert.False(t, diag.HasError(), diag)
	assert.Equal(t, tInstagramIntegrationId, d.Id())
	assert.Equal(t, tInstagramIntegrationId, d.Get("instagram_integration_id").(string))
}

func TestUnitResourceConversationsMessagingIntegrationsInstagramIdentityResolutionRead(t *testing.T) {
	tInstagramIntegrationId := uuid.NewString()
	tDivisionId := uuid.NewString()

	proxy := &conversationsMessagingIntegrationsInstagramIdentityResolutionProxy{}
	proxy.getConversationsMessagingIntegrationsInstagramIdentityResolutionAttr = func(ctx context.Context, p *conversationsMessagingIntegrationsInstagramIdentityResolutionProxy, instagramIntegrationId string) (*platformclientv2.Instagramidentityresolutionconfig, *platformclientv2.APIResponse, error) {
		resolveIdentities := false
		apiResponse := platformclientv2.APIResponse{StatusCode: http.StatusOK}
		return &platformclientv2.Instagramidentityresolutionconfig{
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
	resourceSchema := ResourceConversationsMessagingIntegrationsInstagramIdentityResolution().Schema
	resourceDataMap := buildIdentityResolutionResourceMap(tInstagramIntegrationId, false, tDivisionId)

	d := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
	d.SetId(tInstagramIntegrationId)

	diag := readConversationsMessagingIntegrationsInstagramIdentityResolution(ctx, d, gcloud)
	assert.False(t, diag.HasError(), diag)
	assert.Equal(t, tInstagramIntegrationId, d.Id())
	assert.Equal(t, tInstagramIntegrationId, d.Get("instagram_integration_id").(string))

	assert.Equal(t, false, d.Get("resolve_identities"))
	assert.Equal(t, tDivisionId, d.Get("division_id"))
}

func TestUnitResourceConversationsMessagingIntegrationsInstagramIdentityResolutionDelete(t *testing.T) {
	tInstagramIntegrationId := uuid.NewString()

	proxy := &conversationsMessagingIntegrationsInstagramIdentityResolutionProxy{}
	proxy.getConversationsMessagingIntegrationsInstagramIdentityResolutionAttr = func(ctx context.Context, p *conversationsMessagingIntegrationsInstagramIdentityResolutionProxy, instagramIntegrationId string) (*platformclientv2.Instagramidentityresolutionconfig, *platformclientv2.APIResponse, error) {
		resolveIdentities := false
		apiResponse := platformclientv2.APIResponse{StatusCode: http.StatusOK}
		return &platformclientv2.Instagramidentityresolutionconfig{ResolveIdentities: &resolveIdentities}, &apiResponse, nil
	}

	proxy.putConversationsMessagingIntegrationsInstagramIdentityResolutionAttr = func(ctx context.Context, p *conversationsMessagingIntegrationsInstagramIdentityResolutionProxy, instagramIntegrationId string, config *platformclientv2.Instagramidentityresolutionconfig) (*platformclientv2.Instagramidentityresolutionconfig, *platformclientv2.APIResponse, error) {
		assert.Equal(t, tInstagramIntegrationId, instagramIntegrationId)
		assert.NotNil(t, config.ResolveIdentities)
		assert.Equal(t, true, *config.ResolveIdentities)
		assert.Nil(t, config.Division)

		apiResponse := platformclientv2.APIResponse{StatusCode: http.StatusOK}
		return config, &apiResponse, nil
	}

	internalProxy = proxy
	defer func() { internalProxy = nil }()

	ctx := context.Background()
	gcloud := &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}
	resourceSchema := ResourceConversationsMessagingIntegrationsInstagramIdentityResolution().Schema
	resourceDataMap := buildIdentityResolutionResourceMap(tInstagramIntegrationId, false, "")

	d := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
	d.SetId(tInstagramIntegrationId)

	diag := deleteConversationsMessagingIntegrationsInstagramIdentityResolution(ctx, d, gcloud)
	assert.False(t, diag.HasError(), diag)
}

func TestUnitGetAllArchitectIvrIdentityResolution(t *testing.T) {
	defaultInstagramIntegrationId := uuid.NewString()
	defaultInstagramIntegrationName := "Default Instagram Integration"
	customInstagramIntegrationId := uuid.NewString()
	customInstagramIntegrationName := "Custom Instagram Integration"
	notFoundInstagramIntegrationId := uuid.NewString()
	notFoundInstagramIntegrationName := "Not Found Instagram Integration"

	resolveTrue := true
	resolveFalse := false
	divisionId := uuid.NewString()

	instagramProxy := &instagram.ConversationsMessagingIntegrationsInstagramProxy{}
	instagramProxy.GetAllConversationsMessagingIntegrationsInstagramAttr = func(_ context.Context, _ *instagram.ConversationsMessagingIntegrationsInstagramProxy) (*[]platformclientv2.Instagramintegration, *platformclientv2.APIResponse, error) {
		return &[]platformclientv2.Instagramintegration{
			{Id: &defaultInstagramIntegrationId, Name: &defaultInstagramIntegrationName},
			{Id: &customInstagramIntegrationId, Name: &customInstagramIntegrationName},
			{Id: &notFoundInstagramIntegrationId, Name: &notFoundInstagramIntegrationName},
		}, nil, nil
	}

	proxy := &conversationsMessagingIntegrationsInstagramIdentityResolutionProxy{
		instagramIntegrationProxy: instagramProxy,
	}
	proxy.getConversationsMessagingIntegrationsInstagramIdentityResolutionAttr = func(ctx context.Context, p *conversationsMessagingIntegrationsInstagramIdentityResolutionProxy, instagramIntegrationId string) (*platformclientv2.Instagramidentityresolutionconfig, *platformclientv2.APIResponse, error) {
		switch instagramIntegrationId {
		case defaultInstagramIntegrationId:
			apiResponse := platformclientv2.APIResponse{StatusCode: http.StatusOK}
			return &platformclientv2.Instagramidentityresolutionconfig{
				ResolveIdentities: &resolveTrue,
			}, &apiResponse, nil
		case customInstagramIntegrationId:
			apiResponse := platformclientv2.APIResponse{StatusCode: http.StatusOK}
			return &platformclientv2.Instagramidentityresolutionconfig{
				ResolveIdentities: &resolveFalse,
				Division: &platformclientv2.Writablestarrabledivision{
					Id: &divisionId,
				},
			}, &apiResponse, nil
		case notFoundInstagramIntegrationId:
			apiResponse := platformclientv2.APIResponse{StatusCode: http.StatusNotFound}
			return nil, &apiResponse, fmt.Errorf("not found")
		default:
			t.Fatalf("unexpected Instagram Integration ID %s", instagramIntegrationId)
			return nil, nil, nil
		}
	}

	internalProxy = proxy
	defer func() { internalProxy = nil }()

	ctx := context.Background()
	resources, diag := getAllConversationsMessagingIntegrationsInstagramIdentityResolution(ctx, &platformclientv2.Configuration{})

	assert.False(t, diag.HasError())
	assert.Len(t, resources, 1)
	assert.Contains(t, resources, customInstagramIntegrationId)
	assert.Equal(t, customInstagramIntegrationName+"-identity-resolution", resources[customInstagramIntegrationId].BlockLabel)
}

func TestUnitGetAllConversationsMessagingIntegrationsInstagramIdentityResolutionListError(t *testing.T) {
	instagramProxy := &instagram.ConversationsMessagingIntegrationsInstagramProxy{}
	instagramProxy.GetAllConversationsMessagingIntegrationsInstagramAttr = func(_ context.Context, _ *instagram.ConversationsMessagingIntegrationsInstagramProxy) (*[]platformclientv2.Instagramintegration, *platformclientv2.APIResponse, error) {
		return nil, nil, fmt.Errorf("mock list error")
	}

	proxy := &conversationsMessagingIntegrationsInstagramIdentityResolutionProxy{
		instagramIntegrationProxy: instagramProxy,
	}

	internalProxy = proxy
	defer func() { internalProxy = nil }()

	ctx := context.Background()
	resources, diag := getAllConversationsMessagingIntegrationsInstagramIdentityResolution(ctx, &platformclientv2.Configuration{})

	assert.True(t, diag.HasError())
	assert.Nil(t, resources)
}

func buildIdentityResolutionResourceMap(instagramIntegrationId string, resolveIdentities bool, divisionId string) map[string]interface{} {
	return map[string]interface{}{
		"instagram_integration_id": instagramIntegrationId,
		"resolve_identities":       resolveIdentities,
		"division_id":              divisionId,
	}
}
