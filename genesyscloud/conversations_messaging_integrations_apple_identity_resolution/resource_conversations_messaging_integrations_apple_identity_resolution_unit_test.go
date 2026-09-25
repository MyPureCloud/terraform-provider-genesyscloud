package conversations_messaging_integrations_apple_identity_resolution

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v199/platformclientv2"
	apple "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/conversations_messaging_integrations_apple"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/stretchr/testify/assert"
)

func TestUnitResourceConversationsMessagingIntegrationsAppleIdentityResolutionUpdate(t *testing.T) {
	tAppleIntegrationId := uuid.NewString()

	proxy := &conversationsMessagingIntegrationsAppleIdentityResolutionProxy{}
	proxy.putConversationsMessagingIntegrationsAppleIdentityResolutionAttr = func(ctx context.Context, p *conversationsMessagingIntegrationsAppleIdentityResolutionProxy, appleIntegrationId string, config *platformclientv2.Appleidentityresolutionconfig) (*platformclientv2.Appleidentityresolutionconfig, *platformclientv2.APIResponse, error) {
		assert.Equal(t, tAppleIntegrationId, appleIntegrationId)
		assert.NotNil(t, config.ResolveIdentities)
		assert.Equal(t, false, *config.ResolveIdentities)

		apiResponse := platformclientv2.APIResponse{StatusCode: http.StatusOK}
		return config, &apiResponse, nil
	}

	proxy.getConversationsMessagingIntegrationsAppleIdentityResolutionAttr = func(ctx context.Context, p *conversationsMessagingIntegrationsAppleIdentityResolutionProxy, appleIntegrationId string) (*platformclientv2.Appleidentityresolutionconfig, *platformclientv2.APIResponse, error) {
		resolveIdentities := false
		apiResponse := platformclientv2.APIResponse{StatusCode: http.StatusOK}
		return &platformclientv2.Appleidentityresolutionconfig{
			ResolveIdentities: &resolveIdentities,
		}, &apiResponse, nil
	}

	internalProxy = proxy
	defer func() { internalProxy = nil }()

	ctx := context.Background()
	gcloud := &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}
	resourceSchema := ResourceConversationsMessagingIntegrationsAppleIdentityResolution().Schema
	resourceDataMap := buildIdentityResolutionResourceMap(tAppleIntegrationId, false, "")

	d := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
	d.SetId(tAppleIntegrationId)

	diag := updateConversationsMessagingIntegrationsAppleIdentityResolution(ctx, d, gcloud)
	assert.False(t, diag.HasError(), diag)
	assert.Equal(t, tAppleIntegrationId, d.Id())
	assert.Equal(t, tAppleIntegrationId, d.Get("apple_integration_id").(string))
}

func TestUnitResourceConversationsMessagingIntegrationsAppleIdentityResolutionRead(t *testing.T) {
	tAppleIntegrationId := uuid.NewString()
	tDivisionId := uuid.NewString()

	proxy := &conversationsMessagingIntegrationsAppleIdentityResolutionProxy{}
	proxy.getConversationsMessagingIntegrationsAppleIdentityResolutionAttr = func(ctx context.Context, p *conversationsMessagingIntegrationsAppleIdentityResolutionProxy, appleIntegrationId string) (*platformclientv2.Appleidentityresolutionconfig, *platformclientv2.APIResponse, error) {
		resolveIdentities := false
		apiResponse := platformclientv2.APIResponse{StatusCode: http.StatusOK}
		return &platformclientv2.Appleidentityresolutionconfig{
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
	resourceSchema := ResourceConversationsMessagingIntegrationsAppleIdentityResolution().Schema
	resourceDataMap := buildIdentityResolutionResourceMap(tAppleIntegrationId, false, tDivisionId)

	d := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
	d.SetId(tAppleIntegrationId)

	diag := readConversationsMessagingIntegrationsAppleIdentityResolution(ctx, d, gcloud)
	assert.False(t, diag.HasError(), diag)
	assert.Equal(t, tAppleIntegrationId, d.Id())
	assert.Equal(t, tAppleIntegrationId, d.Get("apple_integration_id").(string))

	assert.Equal(t, false, d.Get("resolve_identities"))
	assert.Equal(t, tDivisionId, d.Get("division_id"))
}

func TestUnitResourceConversationsMessagingIntegrationsAppleIdentityResolutionDelete(t *testing.T) {
	tAppleIntegrationId := uuid.NewString()

	proxy := &conversationsMessagingIntegrationsAppleIdentityResolutionProxy{}
	proxy.getConversationsMessagingIntegrationsAppleIdentityResolutionAttr = func(ctx context.Context, p *conversationsMessagingIntegrationsAppleIdentityResolutionProxy, appleIntegrationId string) (*platformclientv2.Appleidentityresolutionconfig, *platformclientv2.APIResponse, error) {
		resolveIdentities := false
		apiResponse := platformclientv2.APIResponse{StatusCode: http.StatusOK}
		return &platformclientv2.Appleidentityresolutionconfig{ResolveIdentities: &resolveIdentities}, &apiResponse, nil
	}

	proxy.putConversationsMessagingIntegrationsAppleIdentityResolutionAttr = func(ctx context.Context, p *conversationsMessagingIntegrationsAppleIdentityResolutionProxy, appleIntegrationId string, config *platformclientv2.Appleidentityresolutionconfig) (*platformclientv2.Appleidentityresolutionconfig, *platformclientv2.APIResponse, error) {
		assert.Equal(t, tAppleIntegrationId, appleIntegrationId)
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
	resourceSchema := ResourceConversationsMessagingIntegrationsAppleIdentityResolution().Schema
	resourceDataMap := buildIdentityResolutionResourceMap(tAppleIntegrationId, false, "")

	d := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
	d.SetId(tAppleIntegrationId)

	diag := deleteConversationsMessagingIntegrationsAppleIdentityResolution(ctx, d, gcloud)
	assert.False(t, diag.HasError(), diag)
}

func TestUnitGetAllArchitectIvrIdentityResolution(t *testing.T) {
	defaultAppleIntegrationId := uuid.NewString()
	defaultAppleIntegrationName := "Default Apple Integration"
	customAppleIntegrationId := uuid.NewString()
	customAppleIntegrationName := "Custom Apple Integration"
	notFoundAppleIntegrationId := uuid.NewString()
	notFoundAppleIntegrationName := "Not Found Apple Integration"

	resolveTrue := true
	resolveFalse := false
	divisionId := uuid.NewString()

	appleProxy := &apple.ConversationsMessagingIntegrationsAppleProxy{}
	appleProxy.GetAllConversationsMessagingIntegrationsAppleAttr = func(_ context.Context, _ *apple.ConversationsMessagingIntegrationsAppleProxy) (*[]platformclientv2.Appleintegration, *platformclientv2.APIResponse, error) {
		return &[]platformclientv2.Appleintegration{
			{Id: &defaultAppleIntegrationId, Name: &defaultAppleIntegrationName},
			{Id: &customAppleIntegrationId, Name: &customAppleIntegrationName},
			{Id: &notFoundAppleIntegrationId, Name: &notFoundAppleIntegrationName},
		}, nil, nil
	}

	proxy := &conversationsMessagingIntegrationsAppleIdentityResolutionProxy{
		appleIntegrationProxy: appleProxy,
	}
	proxy.getConversationsMessagingIntegrationsAppleIdentityResolutionAttr = func(ctx context.Context, p *conversationsMessagingIntegrationsAppleIdentityResolutionProxy, appleIntegrationId string) (*platformclientv2.Appleidentityresolutionconfig, *platformclientv2.APIResponse, error) {
		switch appleIntegrationId {
		case defaultAppleIntegrationId:
			apiResponse := platformclientv2.APIResponse{StatusCode: http.StatusOK}
			return &platformclientv2.Appleidentityresolutionconfig{
				ResolveIdentities: &resolveTrue,
			}, &apiResponse, nil
		case customAppleIntegrationId:
			apiResponse := platformclientv2.APIResponse{StatusCode: http.StatusOK}
			return &platformclientv2.Appleidentityresolutionconfig{
				ResolveIdentities: &resolveFalse,
				Division: &platformclientv2.Writablestarrabledivision{
					Id: &divisionId,
				},
			}, &apiResponse, nil
		case notFoundAppleIntegrationId:
			apiResponse := platformclientv2.APIResponse{StatusCode: http.StatusNotFound}
			return nil, &apiResponse, fmt.Errorf("not found")
		default:
			t.Fatalf("unexpected Apple Integration ID %s", appleIntegrationId)
			return nil, nil, nil
		}
	}

	internalProxy = proxy
	defer func() { internalProxy = nil }()

	ctx := context.Background()
	resources, diag := getAllConversationsMessagingIntegrationsAppleIdentityResolution(ctx, &platformclientv2.Configuration{})

	assert.False(t, diag.HasError())
	assert.Len(t, resources, 1)
	assert.Contains(t, resources, customAppleIntegrationId)
	assert.Equal(t, customAppleIntegrationName+"-identity-resolution", resources[customAppleIntegrationId].BlockLabel)
}

func TestUnitGetAllConversationsMessagingIntegrationsAppleIdentityResolutionListError(t *testing.T) {
	appleProxy := &apple.ConversationsMessagingIntegrationsAppleProxy{}
	appleProxy.GetAllConversationsMessagingIntegrationsAppleAttr = func(_ context.Context, _ *apple.ConversationsMessagingIntegrationsAppleProxy) (*[]platformclientv2.Appleintegration, *platformclientv2.APIResponse, error) {
		return nil, nil, fmt.Errorf("mock list error")
	}

	proxy := &conversationsMessagingIntegrationsAppleIdentityResolutionProxy{
		appleIntegrationProxy: appleProxy,
	}

	internalProxy = proxy
	defer func() { internalProxy = nil }()

	ctx := context.Background()
	resources, diag := getAllConversationsMessagingIntegrationsAppleIdentityResolution(ctx, &platformclientv2.Configuration{})

	assert.True(t, diag.HasError())
	assert.Nil(t, resources)
}

func buildIdentityResolutionResourceMap(appleIntegrationId string, resolveIdentities bool, divisionId string) map[string]interface{} {
	return map[string]interface{}{
		"apple_integration_id": appleIntegrationId,
		"resolve_identities":   resolveIdentities,
		"division_id":          divisionId,
	}
}
