package integration_facebook

import (
	"context"
	"fmt"
	"net/http"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/google/uuid"
	"github.com/mypurecloud/platform-client-sdk-go/v150/platformclientv2"
	"github.com/stretchr/testify/assert"
)

/** Unit Test **/
func TestUnitResourceIntegrationFacebookCreate(t *testing.T) {
	var (
		name                             = "Unit Test Facebook Integration"
		supportedContentId1              = uuid.NewString()
		messagingSettingId1              = uuid.NewString()
		pageId                           = ""
		appId                            = ""
		fId                              = uuid.NewString()
		pageAccessToken1                 = uuid.NewString()
		supportedcontentreference        = &platformclientv2.Supportedcontentreference{Id: &supportedContentId1}
		messagingsettingreference        = &platformclientv2.Messagingsettingreference{Id: &messagingSettingId1}
		messagingsettingrequestreference = &platformclientv2.Messagingsettingrequestreference{Id: &messagingSettingId1}
	)

	fbProxy := &integrationFacebookProxy{}

	fbProxy.getIntegrationFacebookByIdAttr = func(ctx context.Context, p *integrationFacebookProxy, id string) (facebookIntegrationRequest *platformclientv2.Facebookintegration, response *platformclientv2.APIResponse, err error) {
		assert.Equal(t, fId, id)
		integrationFacebookConfig := platformclientv2.Facebookintegration{
			Name:             &name,
			SupportedContent: supportedcontentreference,
			MessagingSetting: messagingsettingreference,
			PageId:           &pageId,
			AppId:            &appId,
		}

		apiResponse := &platformclientv2.APIResponse{StatusCode: http.StatusOK}

		return &integrationFacebookConfig, apiResponse, nil
	}

	fbProxy.createIntegrationFacebookAttr = func(ctx context.Context, p *integrationFacebookProxy, facebookIntegrationRequest *platformclientv2.Facebookintegrationrequest) (*platformclientv2.Facebookintegration, *platformclientv2.APIResponse, error) {
		integrationFacebookProxy := platformclientv2.Facebookintegration{}

		assert.Equal(t, name, *facebookIntegrationRequest.Name, "facebookIntegrationRequest.Name check failed in create createIntegrationFacebookAttr")
		assert.Equal(t, *supportedcontentreference, *facebookIntegrationRequest.SupportedContent, "facebookIntegrationRequest.SupportedContent check failed in create createIntegrationFacebookAttr")
		assert.Equal(t, *messagingsettingrequestreference, *facebookIntegrationRequest.MessagingSetting, "facebookIntegrationRequest.MessagingSetting check failed in create createIntegrationFacebookAttr")
		assert.Equal(t, pageId, *facebookIntegrationRequest.PageId, "facebookIntegrationRequest.PageId check failed in create createIntegrationFacebookAttr")
		assert.Equal(t, appId, *facebookIntegrationRequest.AppId, "facebookIntegrationRequest.AppId check failed in create createIntegrationFacebookAttr")
		assert.Equal(t, pageAccessToken1, *facebookIntegrationRequest.PageAccessToken, "facebookIntegrationRequest.PageAccessToken check failed in create createIntegrationFacebookAttr")

		integrationFacebookProxy.Id = &fId
		integrationFacebookProxy.Name = &name

		return &integrationFacebookProxy, nil, nil
	}

	internalProxy = fbProxy
	defer func() { internalProxy = nil }()

	ctx := context.Background()
	gCloud := &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}

	resourceSchema := ResourceIntegrationFacebook().Schema

	resourceDataMap := buildIntegrationFacebookResourceMap(fId, name, supportedContentId1, messagingSettingId1, pageId, appId, pageAccessToken1)

	//https://github.com/hashicorp/terraform/issues/890
	d := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
	d.SetId(fId)

	diag := createIntegrationFacebook(ctx, d, gCloud)
	assert.Equal(t, false, diag.HasError())
	assert.Equal(t, fId, d.Id())
}

func TestUnitResourceIntegrationFacebookRead(t *testing.T) {
	var (
		name                      = "Unit Test Facebook Integration"
		supportedContentId1       = uuid.NewString()
		messagingSettingId1       = uuid.NewString()
		pageId                    = ""
		appId                     = ""
		fId                       = uuid.NewString()
		pageAccessToken1          = uuid.NewString()
		supportedcontentreference = &platformclientv2.Supportedcontentreference{Id: &supportedContentId1}
		messagingsettingreference = &platformclientv2.Messagingsettingreference{Id: &messagingSettingId1}
	)

	fbProxy := &integrationFacebookProxy{}

	fbProxy.getIntegrationFacebookByIdAttr = func(ctx context.Context, p *integrationFacebookProxy, id string) (facebookIntegrationRequest *platformclientv2.Facebookintegration, response *platformclientv2.APIResponse, err error) {
		assert.Equal(t, fId, id)
		integrationFacebookConfig := platformclientv2.Facebookintegration{
			Name:             &name,
			SupportedContent: supportedcontentreference,
			MessagingSetting: messagingsettingreference,
			PageId:           &pageId,
			AppId:            &appId,
		}

		apiResponse := &platformclientv2.APIResponse{StatusCode: http.StatusOK}

		return &integrationFacebookConfig, apiResponse, nil
	}

	internalProxy = fbProxy
	defer func() { internalProxy = nil }()

	ctx := context.Background()
	gCloud := &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}

	resourceSchema := ResourceIntegrationFacebook().Schema

	resourceDataMap := buildIntegrationFacebookResourceMap(fId, name, supportedContentId1, messagingSettingId1, pageId, appId, pageAccessToken1)

	//https://github.com/hashicorp/terraform/issues/890
	d := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
	d.SetId(fId)

	scReference := d.Get("supported_content_id").(string)
	msReference := d.Get("messaging_setting_id").(string)
	diag := readIntegrationFacebook(ctx, d, gCloud)
	assert.Equal(t, false, diag.HasError())
	assert.Equal(t, fId, d.Id())
	assert.Equal(t, name, d.Get("name").(string))
	assert.Equal(t, supportedcontentreference, &platformclientv2.Supportedcontentreference{Id: &scReference})
	assert.Equal(t, messagingsettingreference, &platformclientv2.Messagingsettingreference{Id: &msReference})
	assert.Equal(t, pageAccessToken1, d.Get("page_access_token").(string))
	assert.Equal(t, pageId, d.Get("page_id").(string))
	assert.Equal(t, appId, d.Get("app_id").(string))
}

func TestUnitResourceIntegrationFacebookDelete(t *testing.T) {
	var (
		name                = "Unit Test Facebook Integration"
		supportedContentId1 = uuid.NewString()
		messagingSettingId1 = uuid.NewString()
		pageId              = ""
		appId               = ""
		fId                 = uuid.NewString()
		pageAccessToken1    = uuid.NewString()
	)

	fbProxy := &integrationFacebookProxy{}

	fbProxy.deleteIntegrationFacebookAttr = func(ctx context.Context, p *integrationFacebookProxy, id string) (response *platformclientv2.APIResponse, err error) {
		assert.Equal(t, fId, id)

		apiResponse := &platformclientv2.APIResponse{StatusCode: http.StatusNoContent}
		return apiResponse, nil
	}

	fbProxy.getIntegrationFacebookByIdAttr = func(ctx context.Context, p *integrationFacebookProxy, id string) (facebookIntegrationRequest *platformclientv2.Facebookintegration, response *platformclientv2.APIResponse, err error) {
		assert.Equal(t, fId, id)

		apiResponse := &platformclientv2.APIResponse{StatusCode: http.StatusNotFound}

		return nil, apiResponse, fmt.Errorf("not found")
	}

	internalProxy = fbProxy
	defer func() { internalProxy = nil }()

	ctx := context.Background()
	gCloud := &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}

	resourceSchema := ResourceIntegrationFacebook().Schema

	resourceDataMap := buildIntegrationFacebookResourceMap(fId, name, supportedContentId1, messagingSettingId1, pageId, appId, pageAccessToken1)

	//https://github.com/hashicorp/terraform/issues/890
	d := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
	d.SetId(fId)

	diag := deleteIntegrationFacebook(ctx, d, gCloud)
	assert.Nil(t, diag)
	assert.Equal(t, fId, d.Id())
}

func TestUnitResourceIntegrationFacebookUpdate(t *testing.T) {
	var (
		name                             = "Unit Test Facebook Integration"
		supportedContentId1              = uuid.NewString()
		messagingSettingId1              = uuid.NewString()
		pageId                           = ""
		appId                            = ""
		fId                              = uuid.NewString()
		pageAccessToken1                 = uuid.NewString()
		supportedcontentreference        = &platformclientv2.Supportedcontentreference{Id: &supportedContentId1}
		messagingsettingreference        = &platformclientv2.Messagingsettingreference{Id: &messagingSettingId1}
		messagingsettingrequestreference = &platformclientv2.Messagingsettingrequestreference{Id: &messagingSettingId1}
	)

	fbProxy := &integrationFacebookProxy{}

	fbProxy.getIntegrationFacebookByIdAttr = func(ctx context.Context, p *integrationFacebookProxy, id string) (facebookIntegrationRequest *platformclientv2.Facebookintegration, response *platformclientv2.APIResponse, err error) {
		assert.Equal(t, fId, id)
		integrationFacebookConfig := platformclientv2.Facebookintegration{
			Name:             &name,
			SupportedContent: supportedcontentreference,
			MessagingSetting: messagingsettingreference,
			PageId:           &pageId,
			AppId:            &appId,
		}

		apiResponse := &platformclientv2.APIResponse{StatusCode: http.StatusOK}

		return &integrationFacebookConfig, apiResponse, nil
	}

	fbProxy.updateIntegrationFacebookAttr = func(ctx context.Context, p *integrationFacebookProxy, id string, facebookIntegrationRequest *platformclientv2.Facebookintegrationupdaterequest) (*platformclientv2.Facebookintegration, *platformclientv2.APIResponse, error) {
		fbIntegration := platformclientv2.Facebookintegration{}
		assert.Equal(t, name, *facebookIntegrationRequest.Name, "facebookIntegrationRequest.Name check failed in create createIntegrationFacebookAttr")
		assert.Equal(t, *supportedcontentreference, *facebookIntegrationRequest.SupportedContent, "facebookIntegrationRequest.SupportedContent check failed in create createIntegrationFacebookAttr")
		assert.Equal(t, *messagingsettingrequestreference, *facebookIntegrationRequest.MessagingSetting, "facebookIntegrationRequest.MessagingSetting check failed in create createIntegrationFacebookAttr")
		assert.Equal(t, pageAccessToken1, *facebookIntegrationRequest.PageAccessToken, "facebookIntegrationRequest.PageAccessToken check failed in create createIntegrationFacebookAttr")

		fbIntegration.Id = &fId
		fbIntegration.Name = &name

		return &fbIntegration, nil, nil
	}

	internalProxy = fbProxy
	defer func() { internalProxy = nil }()

	ctx := context.Background()
	gCloud := &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}

	resourceSchema := ResourceIntegrationFacebook().Schema

	resourceDataMap := buildIntegrationFacebookResourceMap(fId, name, supportedContentId1, messagingSettingId1, pageId, appId, pageAccessToken1)

	//https://github.com/hashicorp/terraform/issues/890
	d := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
	d.SetId(fId)

	scReference := d.Get("supported_content_id").(string)
	diag := updateIntegrationFacebook(ctx, d, gCloud)
	assert.Equal(t, false, diag.HasError())
	assert.Equal(t, fId, d.Id())
	assert.Equal(t, supportedcontentreference, &platformclientv2.Supportedcontentreference{Id: &scReference})
}

func buildIntegrationFacebookResourceMap(fId string, name string, supportedContentId string, messagingSettingId string, pageId string, appId string, pageAccessToken string) map[string]interface{} {
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
