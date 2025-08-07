package conversations_messaging_integrations_whatsapp

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v165/platformclientv2"
	"github.com/stretchr/testify/assert"
)

/** Unit Test **/
func TestUnitIntegrationWhatsappCreate(t *testing.T) {
	var (
		name                      = "Unit test Whatsapp Integration"
		supportedContentId        = uuid.NewString()
		messagingSettingId        = uuid.NewString()
		whatsappIntegrationId     = uuid.NewString()
		embeddedSignupAccessToken = uuid.NewString()
	)

	whatsappProxy := &conversationsMessagingIntegrationsWhatsappProxy{}

	whatsappProxy.getConversationsMessagingIntegrationsWhatsappByIdAttr = func(ctx context.Context, p *conversationsMessagingIntegrationsWhatsappProxy, id string) (conversationsMessagingIntegrationsWhatsapp *platformclientv2.Whatsappintegration, response *platformclientv2.APIResponse, err error) {
		assert.Equal(t, whatsappIntegrationId, id)

		integrationWhatsappConfig := platformclientv2.Whatsappintegration{
			Name: &name,
			SupportedContent: &platformclientv2.Supportedcontentreference{
				Id: &supportedContentId,
			},
			MessagingSetting: &platformclientv2.Messagingsettingreference{
				Id: &messagingSettingId,
			},
		}

		apiResponse := &platformclientv2.APIResponse{
			StatusCode: http.StatusOK,
		}

		return &integrationWhatsappConfig, apiResponse, nil
	}

	whatsappProxy.createConversationsMessagingIntegrationsWhatsappAttr = func(ctx context.Context, p *conversationsMessagingIntegrationsWhatsappProxy, conversationsMessagingIntegrationsWhatsappRequest *platformclientv2.Whatsappembeddedsignupintegrationrequest) (*platformclientv2.Whatsappintegration, *platformclientv2.APIResponse, error) {
		integrationWhatsappProxy := platformclientv2.Whatsappintegration{}

		assert.Equal(t, name, *conversationsMessagingIntegrationsWhatsappRequest.Name, "conversationsMessagingIntegrationsWhatsappRequest.Name check failed in create createConversationsMessagingIntegrationsWhatsappAttr")
		assert.Equal(t, supportedContentId, *conversationsMessagingIntegrationsWhatsappRequest.SupportedContent.Id, "conversationsMessagingIntegrationsWhatsappRequest.SupportedContent check failed in createConversationsMessagingIntegrationsWhatsappAttr")
		assert.Equal(t, messagingSettingId, *conversationsMessagingIntegrationsWhatsappRequest.MessagingSetting.Id, "conversationsMessagingIntegrationsWhatsappRequest.MessagingSetting check failed in createConversationsMessagingIntegrationsWhatsappAttr")

		integrationWhatsappProxy.Id = &whatsappIntegrationId
		integrationWhatsappProxy.Name = &name

		return &integrationWhatsappProxy, nil, nil
	}

	internalProxy = whatsappProxy
	defer func() {
		internalProxy = nil
	}()

	ctx := context.Background()

	gCloud := &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}

	resourceSchema := ResourceConversationsMessagingIntegrationsWhatsapp().Schema

	resourceDataMap := buildConversationsMessagingIntegrationsWhatsappResourceMap(whatsappIntegrationId, name, supportedContentId, messagingSettingId, embeddedSignupAccessToken)

	d := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
	d.SetId(whatsappIntegrationId)

	diag := createConversationsMessagingIntegrationsWhatsapp(ctx, d, gCloud)
	assert.Equal(t, false, diag.HasError())
	assert.Equal(t, whatsappIntegrationId, d.Id())
}

func TestUnitIntegrationWhatsappRead(t *testing.T) {
	var (
		name                      = "Unit test Whatsapp Integration"
		supportedContentId        = uuid.NewString()
		messagingSettingId        = uuid.NewString()
		whatsappIntegrationId     = uuid.NewString()
		embeddedSignupAccessToken = uuid.NewString()
	)

	whatsappProxy := &conversationsMessagingIntegrationsWhatsappProxy{}

	whatsappProxy.getConversationsMessagingIntegrationsWhatsappByIdAttr = func(ctx context.Context, p *conversationsMessagingIntegrationsWhatsappProxy, id string) (conversationsMessagingIntegrationsWhatsapp *platformclientv2.Whatsappintegration, response *platformclientv2.APIResponse, err error) {
		assert.Equal(t, whatsappIntegrationId, id)

		integrationWhatsappConfig := platformclientv2.Whatsappintegration{
			Name: &name,
			SupportedContent: &platformclientv2.Supportedcontentreference{
				Id: &supportedContentId,
			},
			MessagingSetting: &platformclientv2.Messagingsettingreference{
				Id: &messagingSettingId,
			},
		}

		apiResponse := &platformclientv2.APIResponse{
			StatusCode: http.StatusOK,
		}

		return &integrationWhatsappConfig, apiResponse, nil
	}

	internalProxy = whatsappProxy
	defer func() {
		internalProxy = nil
	}()

	ctx := context.Background()

	gCloud := &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}

	resourceSchema := ResourceConversationsMessagingIntegrationsWhatsapp().Schema

	resourceDataMap := buildConversationsMessagingIntegrationsWhatsappResourceMap(whatsappIntegrationId, name, supportedContentId, messagingSettingId, embeddedSignupAccessToken)

	d := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
	d.SetId(whatsappIntegrationId)

	diag := readConversationsMessagingIntegrationsWhatsapp(ctx, d, gCloud)
	assert.Equal(t, false, diag.HasError())
	assert.Equal(t, whatsappIntegrationId, d.Id())
	assert.Equal(t, name, d.Get("name").(string))
	assert.Equal(t, supportedContentId, d.Get("supported_content_id").(string))
	assert.Equal(t, messagingSettingId, d.Get("messaging_setting_id").(string))
}

func TestUnitIntegrationWhatsappDelete(t *testing.T) {
	var (
		whatsappIntegrationId     = uuid.NewString()
		name                      = "Unit test Whatsapp Integration"
		supportedContentId        = uuid.NewString()
		messagingSettingId        = uuid.NewString()
		embeddedSignupAccessToken = uuid.NewString()
	)

	whatsappProxy := &conversationsMessagingIntegrationsWhatsappProxy{}

	whatsappProxy.deleteConversationsMessagingIntegrationsWhatsappAttr = func(ctx context.Context, p *conversationsMessagingIntegrationsWhatsappProxy, id string) (*platformclientv2.APIResponse, error) {
		assert.Equal(t, whatsappIntegrationId, id)

		apiResponse := &platformclientv2.APIResponse{
			StatusCode: http.StatusNoContent,
		}

		return apiResponse, nil
	}

	whatsappProxy.getConversationsMessagingIntegrationsWhatsappByIdAttr = func(ctx context.Context, p *conversationsMessagingIntegrationsWhatsappProxy, id string) (conversationsMessagingIntegrationsWhatsapp *platformclientv2.Whatsappintegration, response *platformclientv2.APIResponse, err error) {
		assert.Equal(t, whatsappIntegrationId, id)
		apiResponse := &platformclientv2.APIResponse{
			StatusCode: http.StatusNotFound,
		}

		return nil, apiResponse, fmt.Errorf("Not found")
	}

	internalProxy = whatsappProxy
	defer func() {
		internalProxy = nil
	}()

	ctx := context.Background()

	gCloud := &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}

	resourceSchema := ResourceConversationsMessagingIntegrationsWhatsapp().Schema

	resourceDataMap := buildConversationsMessagingIntegrationsWhatsappResourceMap(whatsappIntegrationId, name, supportedContentId, messagingSettingId, embeddedSignupAccessToken)

	d := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
	d.SetId(whatsappIntegrationId)

	diag := deleteConversationsMessagingIntegrationsWhatsapp(ctx, d, gCloud)
	assert.Nil(t, diag)
	assert.Equal(t, whatsappIntegrationId, d.Id())
}

func TestUnitIntegrationWhatsappUpdate(t *testing.T) {
	var (
		whatsappIntegrationId     = uuid.NewString()
		wName                     = "Updated unit test Whatsapp Integration"
		supportedContentId        = uuid.NewString()
		messagingSettingId        = uuid.NewString()
		embeddedSignupAccessToken = uuid.NewString()
	)

	whatsappProxy := &conversationsMessagingIntegrationsWhatsappProxy{}

	whatsappProxy.getConversationsMessagingIntegrationsWhatsappByIdAttr = func(ctx context.Context, p *conversationsMessagingIntegrationsWhatsappProxy, id string) (conversationsMessagingIntegrationsWhatsapp *platformclientv2.Whatsappintegration, response *platformclientv2.APIResponse, err error) {
		assert.Equal(t, whatsappIntegrationId, id)

		integrationWhatsappConfig := platformclientv2.Whatsappintegration{
			Id:   &whatsappIntegrationId,
			Name: &wName,
			SupportedContent: &platformclientv2.Supportedcontentreference{
				Id: &supportedContentId,
			},
			MessagingSetting: &platformclientv2.Messagingsettingreference{
				Id: &messagingSettingId,
			},
		}

		apiResponse := &platformclientv2.APIResponse{
			StatusCode: http.StatusOK,
		}

		return &integrationWhatsappConfig, apiResponse, nil
	}

	whatsappProxy.updateConversationsMessagingIntegrationsWhatsappAttr = func(ctx context.Context, p *conversationsMessagingIntegrationsWhatsappProxy, id string, conversationsMessagingIntegrationsWhatsapp *platformclientv2.Whatsappintegrationupdaterequest) (*platformclientv2.Whatsappintegration, *platformclientv2.APIResponse, error) {

		assert.Equal(t, wName, *conversationsMessagingIntegrationsWhatsapp.Name, "conversationsMessagingIntegrationsWhatsapp.Name check failed in updateConversationsMessagingIntegrationsWhatsappAttr")
		assert.Equal(t, supportedContentId, *conversationsMessagingIntegrationsWhatsapp.SupportedContent.Id, "conversationsMessagingIntegrationsWhatsapp.SupportedContent check failed in updateConversationsMessagingIntegrationsWhatsappAttr")
		assert.Equal(t, messagingSettingId, *conversationsMessagingIntegrationsWhatsapp.MessagingSetting.Id, "conversationsMessagingIntegrationsWhatsapp.MessagingSetting check failed in updateConversationsMessagingIntegrationsWhatsappAttr")

		integrationWhatsappConfig := platformclientv2.Whatsappintegration{
			Id:   &whatsappIntegrationId,
			Name: conversationsMessagingIntegrationsWhatsapp.Name,
			SupportedContent: &platformclientv2.Supportedcontentreference{
				Id: conversationsMessagingIntegrationsWhatsapp.SupportedContent.Id,
			},
			MessagingSetting: &platformclientv2.Messagingsettingreference{
				Id: conversationsMessagingIntegrationsWhatsapp.MessagingSetting.Id,
			},
		}

		apiResponse := &platformclientv2.APIResponse{
			StatusCode: http.StatusOK,
		}

		return &integrationWhatsappConfig, apiResponse, nil
	}

	internalProxy = whatsappProxy
	defer func() { internalProxy = nil }()

	ctx := context.Background()
	gCloud := &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}

	resourceSchema := ResourceConversationsMessagingIntegrationsWhatsapp().Schema

	resourceDataMap := buildConversationsMessagingIntegrationsWhatsappResourceMap(whatsappIntegrationId, wName, supportedContentId, messagingSettingId, embeddedSignupAccessToken)

	d := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
	d.SetId(whatsappIntegrationId)

	diag := updateConversationsMessagingIntegrationsWhatsapp(ctx, d, gCloud)
	assert.Equal(t, false, diag.HasError())
	assert.Equal(t, whatsappIntegrationId, d.Id())
	assert.Equal(t, wName, d.Get("name").(string))
}

func TestUnitIntegrationWhatsappActivate(t *testing.T) {
	var (
		whatsappIntegrationId     = uuid.NewString()
		wName                     = "Unit test Whatsapp Integration"
		supportedContentId        = uuid.NewString()
		messagingSettingId        = uuid.NewString()
		embeddedSignupAccessToken = uuid.NewString()
		phoneNumber               = "+13172222222"
		pin                       = "1234"
	)

	whatsappProxy := &conversationsMessagingIntegrationsWhatsappProxy{}

	whatsappProxy.getConversationsMessagingIntegrationsWhatsappByIdAttr = func(ctx context.Context, p *conversationsMessagingIntegrationsWhatsappProxy, id string) (conversationsMessagingIntegrationsWhatsapp *platformclientv2.Whatsappintegration, response *platformclientv2.APIResponse, err error) {
		assert.Equal(t, whatsappIntegrationId, id)

		integrationWhatsappConfig := platformclientv2.Whatsappintegration{
			Id:   &whatsappIntegrationId,
			Name: &wName,
			SupportedContent: &platformclientv2.Supportedcontentreference{
				Id: &supportedContentId,
			},
			MessagingSetting: &platformclientv2.Messagingsettingreference{
				Id: &messagingSettingId,
			},
		}

		apiResponse := &platformclientv2.APIResponse{
			StatusCode: http.StatusOK,
		}

		return &integrationWhatsappConfig, apiResponse, nil
	}

	whatsappProxy.updateConversationsMessagingIntegrationsWhatsappEmbeddedSignupAttr = func(ctx context.Context, p *conversationsMessagingIntegrationsWhatsappProxy, id string, conversationsMessagingIntegrationsWhatsapp *platformclientv2.Whatsappembeddedsignupintegrationactivationrequest) (*platformclientv2.Whatsappintegration, *platformclientv2.APIResponse, error) {
		assert.Equal(t, phoneNumber, *conversationsMessagingIntegrationsWhatsapp.PhoneNumber, "conversationsMessagingIntegrationsWhatsapp.PhoneNumber check failed in updateConversationsMessagingIntegrationsWhatsappEmbeddedSignupAttr")
		assert.Equal(t, pin, *conversationsMessagingIntegrationsWhatsapp.Pin, "conversationsMessagingIntegrationsWhatsapp.Pin check failed in updateConversationsMessagingIntegrationsWhatsappEmbeddedSignupAttr")

		integrationWhatsappConfig := platformclientv2.Whatsappintegration{
			Id:   &whatsappIntegrationId,
			Name: &wName,
			SupportedContent: &platformclientv2.Supportedcontentreference{
				Id: &supportedContentId,
			},
			MessagingSetting: &platformclientv2.Messagingsettingreference{
				Id: &messagingSettingId,
			},
		}

		apiResponse := &platformclientv2.APIResponse{
			StatusCode: http.StatusOK,
		}

		return &integrationWhatsappConfig, apiResponse, nil
	}

	internalProxy = whatsappProxy
	defer func() { internalProxy = nil }()

	ctx := context.Background()
	gCloud := &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}

	resourceSchema := ResourceConversationsMessagingIntegrationsWhatsapp().Schema
	activateResourceDataMap := buildConversationsMessagingIntegrationsWhatsappActivateResourceMap(whatsappIntegrationId, wName, supportedContentId, messagingSettingId, embeddedSignupAccessToken, phoneNumber, pin)

	d := schema.TestResourceDataRaw(t, resourceSchema, activateResourceDataMap)
	d.SetId(whatsappIntegrationId)

	diag := activateConversationsMessagingIntegrationsWhatsapp(ctx, d, gCloud)
	assert.Equal(t, false, diag.HasError())
	assert.Equal(t, whatsappIntegrationId, d.Id())
}

func buildConversationsMessagingIntegrationsWhatsappResourceMap(wId string, name string, supportedContentId string, messagingSettingId string, embeddedSignupAccessToken string) map[string]interface{} {
	resourceDataMap := map[string]interface{}{
		"id":                           wId,
		"name":                         name,
		"supported_content_id":         supportedContentId,
		"messaging_setting_id":         messagingSettingId,
		"embedded_signup_access_token": embeddedSignupAccessToken,
	}
	return resourceDataMap
}

func buildConversationsMessagingIntegrationsWhatsappActivateResourceMap(wId string, name string, supportedContentId string, messagingSettingId string, embeddedSignupAccessToken string, phoneNumber string, pin string) map[string]interface{} {
	resourceDataMap := map[string]interface{}{
		"id":                           wId,
		"name":                         name,
		"supported_content_id":         supportedContentId,
		"messaging_setting_id":         messagingSettingId,
		"embedded_signup_access_token": embeddedSignupAccessToken,
		"activate_whatsapp":            flattenActivateWhatsapp(phoneNumber, pin),
	}
	return resourceDataMap
}
