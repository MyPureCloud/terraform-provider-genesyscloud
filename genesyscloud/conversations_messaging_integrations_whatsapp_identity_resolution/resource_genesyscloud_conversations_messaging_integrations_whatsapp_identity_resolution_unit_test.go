package conversations_messaging_integrations_whatsapp_identity_resolution

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v199/platformclientv2"
	whatsapp "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/conversations_messaging_integrations_whatsapp"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/stretchr/testify/assert"
)

func TestUnitResourceConversationsMessagingIntegrationsWhatsappIdentityResolutionUpdate(t *testing.T) {
	tWhatsappIntegrationId := uuid.NewString()

	proxy := &conversationsMessagingIntegrationsWhatsappIdentityResolutionProxy{}
	proxy.putConversationsMessagingIntegrationsWhatsappIdentityResolutionAttr = func(ctx context.Context, p *conversationsMessagingIntegrationsWhatsappIdentityResolutionProxy, whatsappIntegrationId string, config *platformclientv2.Whatsappidentityresolutionconfig) (*platformclientv2.Whatsappidentityresolutionconfig, *platformclientv2.APIResponse, error) {
		assert.Equal(t, tWhatsappIntegrationId, whatsappIntegrationId)
		assert.NotNil(t, config.ResolveIdentities)
		assert.Equal(t, false, *config.ResolveIdentities)

		apiResponse := platformclientv2.APIResponse{StatusCode: http.StatusOK}
		return config, &apiResponse, nil
	}

	proxy.getConversationsMessagingIntegrationsWhatsappIdentityResolutionAttr = func(ctx context.Context, p *conversationsMessagingIntegrationsWhatsappIdentityResolutionProxy, whatsappIntegrationId string) (*platformclientv2.Whatsappidentityresolutionconfig, *platformclientv2.APIResponse, error) {
		resolveIdentities := false
		apiResponse := platformclientv2.APIResponse{StatusCode: http.StatusOK}
		return &platformclientv2.Whatsappidentityresolutionconfig{
			ResolveIdentities: &resolveIdentities,
		}, &apiResponse, nil
	}

	internalProxy = proxy
	defer func() { internalProxy = nil }()

	ctx := context.Background()
	gcloud := &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}
	resourceSchema := ResourceConversationsMessagingIntegrationsWhatsappIdentityResolution().Schema
	resourceDataMap := buildIdentityResolutionResourceMap(tWhatsappIntegrationId, false, "")

	d := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
	d.SetId(tWhatsappIntegrationId)

	diag := updateConversationsMessagingIntegrationsWhatsappIdentityResolution(ctx, d, gcloud)
	assert.False(t, diag.HasError(), diag)
	assert.Equal(t, tWhatsappIntegrationId, d.Id())
	assert.Equal(t, tWhatsappIntegrationId, d.Get("whatsapp_integration_id").(string))
}

func TestUnitResourceConversationsMessagingIntegrationsWhatsappIdentityResolutionRead(t *testing.T) {
	tWhatsappIntegrationId := uuid.NewString()
	tDivisionId := uuid.NewString()

	proxy := &conversationsMessagingIntegrationsWhatsappIdentityResolutionProxy{}
	proxy.getConversationsMessagingIntegrationsWhatsappIdentityResolutionAttr = func(ctx context.Context, p *conversationsMessagingIntegrationsWhatsappIdentityResolutionProxy, whatsappIntegrationId string) (*platformclientv2.Whatsappidentityresolutionconfig, *platformclientv2.APIResponse, error) {
		resolveIdentities := false
		apiResponse := platformclientv2.APIResponse{StatusCode: http.StatusOK}
		return &platformclientv2.Whatsappidentityresolutionconfig{
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
	resourceSchema := ResourceConversationsMessagingIntegrationsWhatsappIdentityResolution().Schema
	resourceDataMap := buildIdentityResolutionResourceMap(tWhatsappIntegrationId, false, tDivisionId)

	d := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
	d.SetId(tWhatsappIntegrationId)

	diag := readConversationsMessagingIntegrationsWhatsappIdentityResolution(ctx, d, gcloud)
	assert.False(t, diag.HasError(), diag)
	assert.Equal(t, tWhatsappIntegrationId, d.Id())
	assert.Equal(t, tWhatsappIntegrationId, d.Get("whatsapp_integration_id").(string))

	assert.Equal(t, false, d.Get("resolve_identities"))
	assert.Equal(t, tDivisionId, d.Get("division_id"))
}

func TestUnitResourceConversationsMessagingIntegrationsWhatsappIdentityResolutionDelete(t *testing.T) {
	tWhatsappIntegrationId := uuid.NewString()

	proxy := &conversationsMessagingIntegrationsWhatsappIdentityResolutionProxy{}
	proxy.getConversationsMessagingIntegrationsWhatsappIdentityResolutionAttr = func(ctx context.Context, p *conversationsMessagingIntegrationsWhatsappIdentityResolutionProxy, whatsappIntegrationId string) (*platformclientv2.Whatsappidentityresolutionconfig, *platformclientv2.APIResponse, error) {
		resolveIdentities := false
		apiResponse := platformclientv2.APIResponse{StatusCode: http.StatusOK}
		return &platformclientv2.Whatsappidentityresolutionconfig{ResolveIdentities: &resolveIdentities}, &apiResponse, nil
	}

	proxy.putConversationsMessagingIntegrationsWhatsappIdentityResolutionAttr = func(ctx context.Context, p *conversationsMessagingIntegrationsWhatsappIdentityResolutionProxy, whatsappIntegrationId string, config *platformclientv2.Whatsappidentityresolutionconfig) (*platformclientv2.Whatsappidentityresolutionconfig, *platformclientv2.APIResponse, error) {
		assert.Equal(t, tWhatsappIntegrationId, whatsappIntegrationId)
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
	resourceSchema := ResourceConversationsMessagingIntegrationsWhatsappIdentityResolution().Schema
	resourceDataMap := buildIdentityResolutionResourceMap(tWhatsappIntegrationId, false, "")

	d := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
	d.SetId(tWhatsappIntegrationId)

	diag := deleteConversationsMessagingIntegrationsWhatsappIdentityResolution(ctx, d, gcloud)
	assert.False(t, diag.HasError(), diag)
}

func TestUnitGetAllArchitectIvrIdentityResolution(t *testing.T) {
	defaultWhatsappIntegrationId := uuid.NewString()
	defaultWhatsappIntegrationName := "Default Whatsapp Integration"
	customWhatsappIntegrationId := uuid.NewString()
	customWhatsappIntegrationName := "Custom Whatsapp Integration"
	notFoundWhatsappIntegrationId := uuid.NewString()
	notFoundWhatsappIntegrationName := "Not Found Whatsapp Integration"

	resolveTrue := true
	resolveFalse := false
	divisionId := uuid.NewString()

	whatsappProxy := &whatsapp.ConversationsMessagingIntegrationsWhatsappProxy{}
	whatsappProxy.GetAllConversationsMessagingIntegrationsWhatsappAttr = func(_ context.Context, _ *whatsapp.ConversationsMessagingIntegrationsWhatsappProxy) (*[]platformclientv2.Whatsappintegration, *platformclientv2.APIResponse, error) {
		return &[]platformclientv2.Whatsappintegration{
			{Id: &defaultWhatsappIntegrationId, Name: &defaultWhatsappIntegrationName},
			{Id: &customWhatsappIntegrationId, Name: &customWhatsappIntegrationName},
			{Id: &notFoundWhatsappIntegrationId, Name: &notFoundWhatsappIntegrationName},
		}, nil, nil
	}

	proxy := &conversationsMessagingIntegrationsWhatsappIdentityResolutionProxy{
		whatsappIntegrationProxy: whatsappProxy,
	}
	proxy.getConversationsMessagingIntegrationsWhatsappIdentityResolutionAttr = func(ctx context.Context, p *conversationsMessagingIntegrationsWhatsappIdentityResolutionProxy, whatsappIntegrationId string) (*platformclientv2.Whatsappidentityresolutionconfig, *platformclientv2.APIResponse, error) {
		switch whatsappIntegrationId {
		case defaultWhatsappIntegrationId:
			apiResponse := platformclientv2.APIResponse{StatusCode: http.StatusOK}
			return &platformclientv2.Whatsappidentityresolutionconfig{
				ResolveIdentities: &resolveTrue,
			}, &apiResponse, nil
		case customWhatsappIntegrationId:
			apiResponse := platformclientv2.APIResponse{StatusCode: http.StatusOK}
			return &platformclientv2.Whatsappidentityresolutionconfig{
				ResolveIdentities: &resolveFalse,
				Division: &platformclientv2.Writablestarrabledivision{
					Id: &divisionId,
				},
			}, &apiResponse, nil
		case notFoundWhatsappIntegrationId:
			apiResponse := platformclientv2.APIResponse{StatusCode: http.StatusNotFound}
			return nil, &apiResponse, fmt.Errorf("not found")
		default:
			t.Fatalf("unexpected Whatsapp Integration ID %s", whatsappIntegrationId)
			return nil, nil, nil
		}
	}

	internalProxy = proxy
	defer func() { internalProxy = nil }()

	ctx := context.Background()
	resources, diag := getAllConversationsMessagingIntegrationsWhatsappIdentityResolution(ctx, &platformclientv2.Configuration{})

	assert.False(t, diag.HasError())
	assert.Len(t, resources, 1)
	assert.Contains(t, resources, customWhatsappIntegrationId)
	assert.Equal(t, customWhatsappIntegrationName+"-identity-resolution", resources[customWhatsappIntegrationId].BlockLabel)
}

func TestUnitGetAllConversationsMessagingIntegrationsWhatsappIdentityResolutionListError(t *testing.T) {
	whatsappProxy := &whatsapp.ConversationsMessagingIntegrationsWhatsappProxy{}
	whatsappProxy.GetAllConversationsMessagingIntegrationsWhatsappAttr = func(_ context.Context, _ *whatsapp.ConversationsMessagingIntegrationsWhatsappProxy) (*[]platformclientv2.Whatsappintegration, *platformclientv2.APIResponse, error) {
		return nil, nil, fmt.Errorf("mock list error")
	}

	proxy := &conversationsMessagingIntegrationsWhatsappIdentityResolutionProxy{
		whatsappIntegrationProxy: whatsappProxy,
	}

	internalProxy = proxy
	defer func() { internalProxy = nil }()

	ctx := context.Background()
	resources, diag := getAllConversationsMessagingIntegrationsWhatsappIdentityResolution(ctx, &platformclientv2.Configuration{})

	assert.True(t, diag.HasError())
	assert.Nil(t, resources)
}

func buildIdentityResolutionResourceMap(whatsappIntegrationId string, resolveIdentities bool, divisionId string) map[string]interface{} {
	return map[string]interface{}{
		"whatsapp_integration_id": whatsappIntegrationId,
		"resolve_identities":      resolveIdentities,
		"division_id":             divisionId,
	}
}
