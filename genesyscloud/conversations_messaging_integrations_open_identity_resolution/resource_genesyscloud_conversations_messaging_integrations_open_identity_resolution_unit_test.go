package conversations_messaging_integrations_open_identity_resolution

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v199/platformclientv2"
	open "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/conversations_messaging_integrations_open"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/stretchr/testify/assert"
)

func TestUnitResourceConversationsMessagingIntegrationsOpenIdentityResolutionUpdate(t *testing.T) {
	tOpenIntegrationId := uuid.NewString()

	proxy := &conversationsMessagingIntegrationsOpenIdentityResolutionProxy{}
	proxy.putConversationsMessagingIntegrationsOpenIdentityResolutionAttr = func(ctx context.Context, p *conversationsMessagingIntegrationsOpenIdentityResolutionProxy, openIntegrationId string, config *platformclientv2.Openmessagingidentityresolutionconfig) (*platformclientv2.Openmessagingidentityresolutionconfig, *platformclientv2.APIResponse, error) {
		assert.Equal(t, tOpenIntegrationId, openIntegrationId)
		assert.NotNil(t, config.ResolveIdentities)
		assert.Equal(t, false, *config.ResolveIdentities)

		apiResponse := platformclientv2.APIResponse{StatusCode: http.StatusOK}
		return config, &apiResponse, nil
	}

	proxy.getConversationsMessagingIntegrationsOpenIdentityResolutionAttr = func(ctx context.Context, p *conversationsMessagingIntegrationsOpenIdentityResolutionProxy, openIntegrationId string) (*platformclientv2.Openmessagingidentityresolutionconfig, *platformclientv2.APIResponse, error) {
		resolveIdentities := false
		apiResponse := platformclientv2.APIResponse{StatusCode: http.StatusOK}
		return &platformclientv2.Openmessagingidentityresolutionconfig{
			ResolveIdentities: &resolveIdentities,
		}, &apiResponse, nil
	}

	internalProxy = proxy
	defer func() { internalProxy = nil }()

	ctx := context.Background()
	gcloud := &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}
	resourceSchema := ResourceConversationsMessagingIntegrationsOpenIdentityResolution().Schema
	resourceDataMap := buildIdentityResolutionResourceMap(tOpenIntegrationId, false, "", "")

	d := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
	d.SetId(tOpenIntegrationId)

	diag := updateConversationsMessagingIntegrationsOpenIdentityResolution(ctx, d, gcloud)
	assert.False(t, diag.HasError(), diag)
	assert.Equal(t, tOpenIntegrationId, d.Id())
	assert.Equal(t, tOpenIntegrationId, d.Get("open_integration_id").(string))
}

func TestUnitResourceConversationsMessagingIntegrationsOpenIdentityResolutionRead(t *testing.T) {
	tOpenIntegrationId := uuid.NewString()
	tDivisionId := uuid.NewString()
	tExternalSourceId := uuid.NewString()

	proxy := &conversationsMessagingIntegrationsOpenIdentityResolutionProxy{}
	proxy.getConversationsMessagingIntegrationsOpenIdentityResolutionAttr = func(ctx context.Context, p *conversationsMessagingIntegrationsOpenIdentityResolutionProxy, openIntegrationId string) (*platformclientv2.Openmessagingidentityresolutionconfig, *platformclientv2.APIResponse, error) {
		resolveIdentities := false
		apiResponse := platformclientv2.APIResponse{StatusCode: http.StatusOK}
		return &platformclientv2.Openmessagingidentityresolutionconfig{
			ResolveIdentities: &resolveIdentities,
			Division: &platformclientv2.Writablestarrabledivision{
				Id: &tDivisionId,
			},
			ExternalSource: &platformclientv2.Identityresolutionexternalsource{
				Id: &tExternalSourceId,
			},
		}, &apiResponse, nil
	}

	internalProxy = proxy
	defer func() { internalProxy = nil }()

	ctx := context.Background()
	gcloud := &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}
	resourceSchema := ResourceConversationsMessagingIntegrationsOpenIdentityResolution().Schema
	resourceDataMap := buildIdentityResolutionResourceMap(tOpenIntegrationId, false, tDivisionId, tExternalSourceId)

	d := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
	d.SetId(tOpenIntegrationId)

	diag := readConversationsMessagingIntegrationsOpenIdentityResolution(ctx, d, gcloud)
	assert.False(t, diag.HasError(), diag)
	assert.Equal(t, tOpenIntegrationId, d.Id())
	assert.Equal(t, tOpenIntegrationId, d.Get("open_integration_id").(string))

	assert.Equal(t, false, d.Get("resolve_identities"))
	assert.Equal(t, tDivisionId, d.Get("division_id"))
	assert.Equal(t, tExternalSourceId, d.Get("external_source_id"))
}

func TestUnitResourceConversationsMessagingIntegrationsOpenIdentityResolutionDelete(t *testing.T) {
	tOpenIntegrationId := uuid.NewString()

	proxy := &conversationsMessagingIntegrationsOpenIdentityResolutionProxy{}
	proxy.getConversationsMessagingIntegrationsOpenIdentityResolutionAttr = func(ctx context.Context, p *conversationsMessagingIntegrationsOpenIdentityResolutionProxy, openIntegrationId string) (*platformclientv2.Openmessagingidentityresolutionconfig, *platformclientv2.APIResponse, error) {
		resolveIdentities := false
		apiResponse := platformclientv2.APIResponse{StatusCode: http.StatusOK}
		return &platformclientv2.Openmessagingidentityresolutionconfig{ResolveIdentities: &resolveIdentities}, &apiResponse, nil
	}

	proxy.putConversationsMessagingIntegrationsOpenIdentityResolutionAttr = func(ctx context.Context, p *conversationsMessagingIntegrationsOpenIdentityResolutionProxy, openIntegrationId string, config *platformclientv2.Openmessagingidentityresolutionconfig) (*platformclientv2.Openmessagingidentityresolutionconfig, *platformclientv2.APIResponse, error) {
		assert.Equal(t, tOpenIntegrationId, openIntegrationId)
		assert.NotNil(t, config.ResolveIdentities)
		assert.Equal(t, true, *config.ResolveIdentities)
		assert.Nil(t, config.Division)
		assert.Nil(t, config.ExternalSource)

		apiResponse := platformclientv2.APIResponse{StatusCode: http.StatusOK}
		return config, &apiResponse, nil
	}

	internalProxy = proxy
	defer func() { internalProxy = nil }()

	ctx := context.Background()
	gcloud := &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}
	resourceSchema := ResourceConversationsMessagingIntegrationsOpenIdentityResolution().Schema
	resourceDataMap := buildIdentityResolutionResourceMap(tOpenIntegrationId, false, "", "")

	d := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
	d.SetId(tOpenIntegrationId)

	diag := deleteConversationsMessagingIntegrationsOpenIdentityResolution(ctx, d, gcloud)
	assert.False(t, diag.HasError(), diag)
}

func TestUnitGetAllConversationsMessagingIntegrationsIdentityResolution(t *testing.T) {
	defaultOpenIntegrationId := uuid.NewString()
	defaultOpenIntegrationName := "Default Open Integration"
	customOpenIntegrationId := uuid.NewString()
	customOpenIntegrationName := "Custom Open Integration"
	notFoundOpenIntegrationId := uuid.NewString()
	notFoundOpenIntegrationName := "Not Found Open Integration"

	resolveTrue := true
	resolveFalse := false
	divisionId := uuid.NewString()
	externalSourceId := uuid.NewString()

	openProxy := &open.ConversationsMessagingIntegrationsOpenProxy{}
	openProxy.GetAllConversationsMessagingIntegrationsOpenAttr = func(_ context.Context, _ *open.ConversationsMessagingIntegrationsOpenProxy) (*[]platformclientv2.Openintegration, *platformclientv2.APIResponse, error) {
		return &[]platformclientv2.Openintegration{
			{Id: &defaultOpenIntegrationId, Name: &defaultOpenIntegrationName},
			{Id: &customOpenIntegrationId, Name: &customOpenIntegrationName},
			{Id: &notFoundOpenIntegrationId, Name: &notFoundOpenIntegrationName},
		}, nil, nil
	}

	proxy := &conversationsMessagingIntegrationsOpenIdentityResolutionProxy{
		openIntegrationProxy: openProxy,
	}
	proxy.getConversationsMessagingIntegrationsOpenIdentityResolutionAttr = func(ctx context.Context, p *conversationsMessagingIntegrationsOpenIdentityResolutionProxy, openIntegrationId string) (*platformclientv2.Openmessagingidentityresolutionconfig, *platformclientv2.APIResponse, error) {
		switch openIntegrationId {
		case defaultOpenIntegrationId:
			apiResponse := platformclientv2.APIResponse{StatusCode: http.StatusOK}
			return &platformclientv2.Openmessagingidentityresolutionconfig{
				ResolveIdentities: &resolveTrue,
			}, &apiResponse, nil
		case customOpenIntegrationId:
			apiResponse := platformclientv2.APIResponse{StatusCode: http.StatusOK}
			return &platformclientv2.Openmessagingidentityresolutionconfig{
				ResolveIdentities: &resolveFalse,
				Division: &platformclientv2.Writablestarrabledivision{
					Id: &divisionId,
				},
				ExternalSource: &platformclientv2.Identityresolutionexternalsource{
					Id: &externalSourceId,
				},
			}, &apiResponse, nil
		case notFoundOpenIntegrationId:
			apiResponse := platformclientv2.APIResponse{StatusCode: http.StatusNotFound}
			return nil, &apiResponse, fmt.Errorf("not found")
		default:
			t.Fatalf("unexpected Open Integration ID %s", openIntegrationId)
			return nil, nil, nil
		}
	}

	internalProxy = proxy
	defer func() { internalProxy = nil }()

	ctx := context.Background()
	resources, diag := getAllConversationsMessagingIntegrationsOpenIdentityResolution(ctx, &platformclientv2.Configuration{})

	assert.False(t, diag.HasError())
	assert.Len(t, resources, 1)
	assert.Contains(t, resources, customOpenIntegrationId)
	assert.Equal(t, customOpenIntegrationName+"-identity-resolution", resources[customOpenIntegrationId].BlockLabel)
}

func TestUnitGetAllConversationsMessagingIntegrationsOpenIdentityResolutionListError(t *testing.T) {
	openProxy := &open.ConversationsMessagingIntegrationsOpenProxy{}
	openProxy.GetAllConversationsMessagingIntegrationsOpenAttr = func(_ context.Context, _ *open.ConversationsMessagingIntegrationsOpenProxy) (*[]platformclientv2.Openintegration, *platformclientv2.APIResponse, error) {
		return nil, nil, fmt.Errorf("mock list error")
	}

	proxy := &conversationsMessagingIntegrationsOpenIdentityResolutionProxy{
		openIntegrationProxy: openProxy,
	}

	internalProxy = proxy
	defer func() { internalProxy = nil }()

	ctx := context.Background()
	resources, diag := getAllConversationsMessagingIntegrationsOpenIdentityResolution(ctx, &platformclientv2.Configuration{})

	assert.True(t, diag.HasError())
	assert.Nil(t, resources)
}

func buildIdentityResolutionResourceMap(openIntegrationId string, resolveIdentities bool, divisionId string, externalSourceId string) map[string]interface{} {
	return map[string]interface{}{
		"open_integration_id": openIntegrationId,
		"resolve_identities":  resolveIdentities,
		"division_id":         divisionId,
		"external_source_id":  externalSourceId,
	}
}
