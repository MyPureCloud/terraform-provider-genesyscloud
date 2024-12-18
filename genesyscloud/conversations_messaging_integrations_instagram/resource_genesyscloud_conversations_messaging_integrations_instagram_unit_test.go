package conversations_messaging_integrations_instagram

import (
	"context"
	"fmt"
	"net/http"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v149/platformclientv2"
	"github.com/stretchr/testify/assert"
)

func TestUnitIntegrationInstagramCreate(t *testing.T) {
	var (
		name                             = "Unit Test Instagram Integration"
		supportedContentId1              = uuid.NewString()
		messagingSettingId1              = uuid.NewString()
		pageId                           = ""
		appId                            = ""
		instagramId                      = uuid.NewString()
		pageAccessToken1                 = uuid.NewString()
		supportedcontentreference        = &platformclientv2.Supportedcontentreference{Id: &supportedContentId1}
		messagingsettingreference        = &platformclientv2.Messagingsettingreference{Id: &messagingSettingId1}
		messagingsettingrequestreference = &platformclientv2.Messagingsettingrequestreference{Id: &messagingSettingId1}
	)

	instagramProxy := &conversationsMessagingIntegrationsInstagramProxy{}

	instagramProxy.getConversationsMessagingIntegrationsInstagramByIdAttr = func(ctx context.Context, p *conversationsMessagingIntegrationsInstagramProxy, id string) (instagramIntegrationRequest *platformclientv2.Instagramintegration, response *platformclientv2.APIResponse, err error) {
		assert.Equal(t, instagramId, id)
		integrationInstagramConfig := platformclientv2.Instagramintegration{
			Name:             &name,
			SupportedContent: supportedcontentreference,
			MessagingSetting: messagingsettingreference,
			PageId:           &pageId,
			AppId:            &appId,
		}

		apiResponse := &platformclientv2.APIResponse{StatusCode: http.StatusOK}

		return &integrationInstagramConfig, apiResponse, nil
	}

	instagramProxy.createConversationsMessagingIntegrationsInstagramAttr = func(ctx context.Context, p *conversationsMessagingIntegrationsInstagramProxy, instagramIntegrationRequest *platformclientv2.Instagramintegrationrequest) (*platformclientv2.Instagramintegration, *platformclientv2.APIResponse, error) {
		integrationInstagramProxy := platformclientv2.Instagramintegration{}

		assert.Equal(t, name, *instagramIntegrationRequest.Name, "instagramIntegrationRequest.Name check failed in create createConversationsMessagingIntegrationsInstagramAttr")
		assert.Equal(t, *supportedcontentreference, *instagramIntegrationRequest.SupportedContent, "instagramIntegrationRequest.SupportedContent check failed in create createConversationsMessagingIntegrationsInstagramAttr")
		assert.Equal(t, *messagingsettingrequestreference, *instagramIntegrationRequest.MessagingSetting, "instagramIntegrationRequest.MessagingSetting check failed in create createConversationsMessagingIntegrationsInstagramAttr")
		assert.Equal(t, pageId, *instagramIntegrationRequest.PageId, "instagramIntegrationRequest.PageId check failed in create createConversationsMessagingIntegrationsInstagramAttr")
		assert.Equal(t, appId, *instagramIntegrationRequest.AppId, "instagramIntegrationRequest.AppId check failed in create createConversationsMessagingIntegrationsInstagramAttr")
		assert.Equal(t, pageAccessToken1, *instagramIntegrationRequest.PageAccessToken, "instagramIntegrationRequest.PageAccessToken check failed in create createConversationsMessagingIntegrationsInstagramAttr")

		integrationInstagramProxy.Id = &instagramId
		integrationInstagramProxy.Name = &name

		return &integrationInstagramProxy, nil, nil
	}

	internalProxy = instagramProxy
	defer func() { internalProxy = nil }()

	ctx := context.Background()

	gCloud := &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}

	resourceSchema := ResourceConversationsMessagingIntegrationsInstagram().Schema

	resourceDataMap := buildIntegrationInstagramResourceMap(instagramId, name, supportedContentId1, messagingSettingId1, pageId, appId, pageAccessToken1)

	//https://github.com/hashicorp/terraform/issues/890
	d := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
	d.SetId(instagramId)

	diag := createConversationsMessagingIntegrationsInstagram(ctx, d, gCloud)
	assert.Equal(t, false, diag.HasError())
	assert.Equal(t, instagramId, d.Id())
}

func TestUnitIntegrationInstagramRead(t *testing.T) {
	var (
		name                      = "Unit Test Instagram Integration"
		supportedContentId1       = uuid.NewString()
		messagingSettingId1       = uuid.NewString()
		pageId                    = ""
		appId                     = ""
		instagramId               = uuid.NewString()
		pageAccessToken1          = uuid.NewString()
		supportedcontentreference = &platformclientv2.Supportedcontentreference{Id: &supportedContentId1}
		messagingsettingreference = &platformclientv2.Messagingsettingreference{Id: &messagingSettingId1}
	)

	instagramProxy := &conversationsMessagingIntegrationsInstagramProxy{}

	instagramProxy.getConversationsMessagingIntegrationsInstagramByIdAttr = func(ctx context.Context, p *conversationsMessagingIntegrationsInstagramProxy, id string) (instagramIntegrationRequest *platformclientv2.Instagramintegration, response *platformclientv2.APIResponse, err error) {
		assert.Equal(t, instagramId, id)
		integrationInstagramConfig := platformclientv2.Instagramintegration{
			Name:             &name,
			SupportedContent: supportedcontentreference,
			MessagingSetting: messagingsettingreference,
			PageId:           &pageId,
			AppId:            &appId,
		}

		apiResponse := &platformclientv2.APIResponse{StatusCode: http.StatusOK}

		return &integrationInstagramConfig, apiResponse, nil
	}

	internalProxy = instagramProxy
	defer func() { internalProxy = nil }()

	ctx := context.Background()

	gCloud := &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}

	resourceSchema := ResourceConversationsMessagingIntegrationsInstagram().Schema

	resourceDataMap := buildIntegrationInstagramResourceMap(instagramId, name, supportedContentId1, messagingSettingId1, pageId, appId, pageAccessToken1)

	//https://github.com/hashicorp/terraform/issues/890
	d := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
	d.SetId(instagramId)

	scReference := d.Get("supported_content_id").(string)
	msReference := d.Get("messaging_setting_id").(string)

	diag := readConversationsMessagingIntegrationsInstagram(ctx, d, gCloud)
	assert.Equal(t, false, diag.HasError())
	assert.Equal(t, instagramId, d.Id())
	assert.Equal(t, name, d.Get("name").(string))
	assert.Equal(t, supportedcontentreference, &platformclientv2.Supportedcontentreference{Id: &scReference})
	assert.Equal(t, messagingsettingreference, &platformclientv2.Messagingsettingreference{Id: &msReference})
	assert.Equal(t, pageAccessToken1, d.Get("page_access_token").(string))
	assert.Equal(t, pageId, d.Get("page_id").(string))
	assert.Equal(t, appId, d.Get("app_id").(string))
}

func TestUnitIntegrationInstagramDelete(t *testing.T) {
	var (
		name                = "Unit Test Instagram Integration"
		supportedContentId1 = uuid.NewString()
		messagingSettingId1 = uuid.NewString()
		pageId              = ""
		appId               = ""
		instagramId         = uuid.NewString()
		pageAccessToken1    = uuid.NewString()
	)

	instagramProxy := &conversationsMessagingIntegrationsInstagramProxy{}

	instagramProxy.deleteConversationsMessagingIntegrationsInstagramAttr = func(ctx context.Context, p *conversationsMessagingIntegrationsInstagramProxy, id string) (response *platformclientv2.APIResponse, err error) {
		assert.Equal(t, instagramId, id)
		apiResponse := &platformclientv2.APIResponse{StatusCode: http.StatusNoContent}
		return apiResponse, nil
	}

	instagramProxy.getConversationsMessagingIntegrationsInstagramByIdAttr = func(ctx context.Context, p *conversationsMessagingIntegrationsInstagramProxy, id string) (instagramIntegrationRequest *platformclientv2.Instagramintegration, response *platformclientv2.APIResponse, err error) {
		assert.Equal(t, instagramId, id)
		apiResponse := &platformclientv2.APIResponse{StatusCode: http.StatusNotFound}
		return nil, apiResponse, fmt.Errorf("Not found")
	}

	internalProxy = instagramProxy
	defer func() { internalProxy = nil }()

	ctx := context.Background()

	gCloud := &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}

	resourceSchema := ResourceConversationsMessagingIntegrationsInstagram().Schema

	resourceDataMap := buildIntegrationInstagramResourceMap(instagramId, name, supportedContentId1, messagingSettingId1, pageId, appId, pageAccessToken1)

	//https://github.com/hashicorp/terraform/issues/890
	d := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
	d.SetId(instagramId)

	diag := deleteConversationsMessagingIntegrationsInstagram(ctx, d, gCloud)
	assert.Nil(t, diag)
	assert.Equal(t, instagramId, d.Id())
}

func TestUnitIntegrationInstagramUpdate(t *testing.T) {
	var (
		name                             = "Unit Test Instagram Integration"
		supportedContentId1              = uuid.NewString()
		messagingSettingId1              = uuid.NewString()
		pageId                           = ""
		appId                            = ""
		instagramId                      = uuid.NewString()
		pageAccessToken1                 = uuid.NewString()
		supportedcontentreference        = &platformclientv2.Supportedcontentreference{Id: &supportedContentId1}
		messagingsettingreference        = &platformclientv2.Messagingsettingreference{Id: &messagingSettingId1}
		messagingsettingrequestreference = &platformclientv2.Messagingsettingrequestreference{Id: &messagingSettingId1}
	)

	instagramProxy := &conversationsMessagingIntegrationsInstagramProxy{}

	instagramProxy.getConversationsMessagingIntegrationsInstagramByIdAttr = func(ctx context.Context, p *conversationsMessagingIntegrationsInstagramProxy, id string) (instagramIntegrationRequest *platformclientv2.Instagramintegration, response *platformclientv2.APIResponse, err error) {
		assert.Equal(t, instagramId, id)
		integrationInstagramConfig := platformclientv2.Instagramintegration{
			Name:             &name,
			SupportedContent: supportedcontentreference,
			MessagingSetting: messagingsettingreference,
			PageId:           &pageId,
			AppId:            &appId,
		}

		apiResponse := &platformclientv2.APIResponse{StatusCode: http.StatusOK}

		return &integrationInstagramConfig, apiResponse, nil
	}

	instagramProxy.updateConversationsMessagingIntegrationsInstagramAttr = func(ctx context.Context, p *conversationsMessagingIntegrationsInstagramProxy, id string, instagramIntegrationRequest *platformclientv2.Instagramintegrationupdaterequest) (*platformclientv2.Instagramintegration, *platformclientv2.APIResponse, error) {
		integrationInstagramProxy := platformclientv2.Instagramintegration{}

		assert.Equal(t, name, *instagramIntegrationRequest.Name, "instagramIntegrationRequest.Name check failed in create createConversationsMessagingIntegrationsInstagramAttr")
		assert.Equal(t, *supportedcontentreference, *instagramIntegrationRequest.SupportedContent, "instagramIntegrationRequest.SupportedContent check failed in create createConversationsMessagingIntegrationsInstagramAttr")
		assert.Equal(t, *messagingsettingrequestreference, *instagramIntegrationRequest.MessagingSetting, "instagramIntegrationRequest.MessagingSetting check failed in create createConversationsMessagingIntegrationsInstagramAttr")
		assert.Equal(t, pageAccessToken1, *instagramIntegrationRequest.PageAccessToken, "instagramIntegrationRequest.PageAccessToken check failed in create createConversationsMessagingIntegrationsInstagramAttr")

		integrationInstagramProxy.Id = &instagramId
		integrationInstagramProxy.Name = &name

		return &integrationInstagramProxy, nil, nil
	}

	internalProxy = instagramProxy
	defer func() { internalProxy = nil }()

	ctx := context.Background()

	gCloud := &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}

	resourceSchema := ResourceConversationsMessagingIntegrationsInstagram().Schema

	resourceDataMap := buildIntegrationInstagramResourceMap(instagramId, name, supportedContentId1, messagingSettingId1, pageId, appId, pageAccessToken1)

	//https://github.com/hashicorp/terraform/issues/890
	d := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
	d.SetId(instagramId)

	scReference := d.Get("supported_content_id").(string)

	diag := updateConversationsMessagingIntegrationsInstagram(ctx, d, gCloud)
	assert.Equal(t, false, diag.HasError())
	assert.Equal(t, instagramId, d.Id())
	assert.Equal(t, supportedcontentreference, &platformclientv2.Supportedcontentreference{Id: &scReference})
}

func buildIntegrationInstagramResourceMap(fId string, name string, supportedContentId string, messagingSettingId string, pageId string, appId string, pageAccessToken string) map[string]interface{} {
	resourceDataMap := map[string]interface{}{
		"id":                   fId,
		"name":                 name,
		"supported_content_id": supportedContentId,
		"messaging_setting_id": messagingSettingId,
		"page_access_token":    pageAccessToken,
		"page_id":              pageId,
		"app_id":               appId,
	}
	return resourceDataMap
}
