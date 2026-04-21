package outbound_messagingcampaign

import (
	"context"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v186/platformclientv2"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/stretchr/testify/assert"
)

func TestUnitBuildWhatsAppConfigs(t *testing.T) {
	t.Parallel()

	t.Run("returns nil for nil input", func(t *testing.T) {
		result := buildWhatsAppConfigs(nil)
		assert.Nil(t, result)
	})

	t.Run("returns nil for empty set", func(t *testing.T) {
		emptySet := schema.NewSet(schema.HashResource(whatsAppConfigResource), []interface{}{})
		result := buildWhatsAppConfigs(emptySet)
		assert.Nil(t, result)
	})

	t.Run("builds config with all fields", func(t *testing.T) {
		configMap := map[string]interface{}{
			"whats_app_columns":        []interface{}{"col1", "col2"},
			"whats_app_integration_id": "integration-123",
			"content_template_id":      "template-456",
		}
		configSet := schema.NewSet(schema.HashResource(whatsAppConfigResource), []interface{}{configMap})

		result := buildWhatsAppConfigs(configSet)

		assert.NotNil(t, result)
		assert.NotNil(t, result.WhatsAppColumns)
		assert.Equal(t, []string{"col1", "col2"}, *result.WhatsAppColumns)
		assert.NotNil(t, result.WhatsAppIntegration)
		assert.Equal(t, "integration-123", *result.WhatsAppIntegration.Id)
		assert.NotNil(t, result.ContentTemplate)
		assert.Equal(t, "template-456", *result.ContentTemplate.Id)
	})
}

func TestUnitFlattenWhatsAppConfigs(t *testing.T) {
	t.Parallel()

	t.Run("returns nil for nil input", func(t *testing.T) {
		result := flattenWhatsAppConfigs(nil)
		assert.Nil(t, result)
	})

	t.Run("flattens config with all fields", func(t *testing.T) {
		columns := []string{"col1", "col2"}
		config := &platformclientv2.Whatsappconfig{
			WhatsAppColumns:     &columns,
			WhatsAppIntegration: &platformclientv2.Addressableentityref{Id: platformclientv2.String("integration-123")},
			ContentTemplate:     &platformclientv2.Domainentityref{Id: platformclientv2.String("template-456")},
		}

		result := flattenWhatsAppConfigs(config)

		assert.NotNil(t, result)
		assert.Equal(t, 1, result.Len())
		resultMap := result.List()[0].(map[string]interface{})
		assert.Equal(t, "integration-123", resultMap["whats_app_integration_id"])
		assert.Equal(t, "template-456", resultMap["content_template_id"])
	})

	t.Run("handles nil nested fields", func(t *testing.T) {
		config := &platformclientv2.Whatsappconfig{}

		result := flattenWhatsAppConfigs(config)

		assert.NotNil(t, result)
		assert.Equal(t, 1, result.Len())
	})
}

func TestUnitBuildAndFlattenWhatsAppConfigsRoundTrip(t *testing.T) {
	t.Parallel()

	configMap := map[string]interface{}{
		"whats_app_columns":        []interface{}{"phoneCol"},
		"whats_app_integration_id": "int-abc",
		"content_template_id":      "tmpl-xyz",
	}
	configSet := schema.NewSet(schema.HashResource(whatsAppConfigResource), []interface{}{configMap})

	built := buildWhatsAppConfigs(configSet)
	flattened := flattenWhatsAppConfigs(built)

	assert.NotNil(t, flattened)
	assert.Equal(t, 1, flattened.Len())
	resultMap := flattened.List()[0].(map[string]interface{})
	assert.Equal(t, "int-abc", resultMap["whats_app_integration_id"])
	assert.Equal(t, "tmpl-xyz", resultMap["content_template_id"])
}

func TestUnitMessagingCampaignWhatsAppCreate(t *testing.T) {
	var (
		campaignId      = uuid.NewString()
		campaignName    = "Test WhatsApp Campaign"
		contactListId   = uuid.NewString()
		integrationId   = uuid.NewString()
		templateId      = uuid.NewString()
		messagesPerMin  = 10
		campaignStatus  = "off"
		whatsAppColumns = []string{"whatsappCol"}
		version         = 1
	)

	proxy := &outboundMessagingcampaignProxy{}

	proxy.createOutboundMessagingcampaignAttr = func(ctx context.Context, p *outboundMessagingcampaignProxy, campaign *platformclientv2.Messagingcampaign) (*platformclientv2.Messagingcampaign, *platformclientv2.APIResponse, error) {
		assert.Equal(t, campaignName, *campaign.Name)
		assert.NotNil(t, campaign.WhatsAppConfig)
		assert.Equal(t, integrationId, *campaign.WhatsAppConfig.WhatsAppIntegration.Id)
		assert.Equal(t, templateId, *campaign.WhatsAppConfig.ContentTemplate.Id)
		assert.Equal(t, whatsAppColumns, *campaign.WhatsAppConfig.WhatsAppColumns)

		campaign.Id = &campaignId
		campaign.Version = &version
		return campaign, nil, nil
	}

	proxy.getOutboundMessagingcampaignByIdAttr = func(ctx context.Context, p *outboundMessagingcampaignProxy, id string) (*platformclientv2.Messagingcampaign, *platformclientv2.APIResponse, error) {
		assert.Equal(t, campaignId, id)
		return &platformclientv2.Messagingcampaign{
			Id:                &campaignId,
			Name:              &campaignName,
			CampaignStatus:    &campaignStatus,
			MessagesPerMinute: &messagesPerMin,
			ContactList:       &platformclientv2.Domainentityref{Id: &contactListId},
			Version:           &version,
			WhatsAppConfig: &platformclientv2.Whatsappconfig{
				WhatsAppColumns:     &whatsAppColumns,
				WhatsAppIntegration: &platformclientv2.Addressableentityref{Id: &integrationId},
				ContentTemplate:     &platformclientv2.Domainentityref{Id: &templateId},
			},
		}, &platformclientv2.APIResponse{StatusCode: http.StatusOK}, nil
	}

	internalProxy = proxy
	defer func() { internalProxy = nil }()

	ctx := context.Background()
	gCloud := &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}

	resourceSchema := ResourceOutboundMessagingcampaign().Schema
	resourceDataMap := buildMessagingCampaignWhatsAppResourceMap(campaignName, contactListId, campaignStatus, messagesPerMin, integrationId, templateId, whatsAppColumns)

	d := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
	d.SetId(campaignId)

	diag := createOutboundMessagingcampaign(ctx, d, gCloud)
	assert.Equal(t, false, diag.HasError())
	assert.Equal(t, campaignId, d.Id())
}

func TestUnitMessagingCampaignWhatsAppRead(t *testing.T) {
	var (
		campaignId      = uuid.NewString()
		campaignName    = "Test WhatsApp Campaign"
		contactListId   = uuid.NewString()
		integrationId   = uuid.NewString()
		templateId      = uuid.NewString()
		messagesPerMin  = 10
		campaignStatus  = "off"
		whatsAppColumns = []string{"whatsappCol"}
		version         = 1
	)

	proxy := &outboundMessagingcampaignProxy{}

	proxy.getOutboundMessagingcampaignByIdAttr = func(ctx context.Context, p *outboundMessagingcampaignProxy, id string) (*platformclientv2.Messagingcampaign, *platformclientv2.APIResponse, error) {
		assert.Equal(t, campaignId, id)
		return &platformclientv2.Messagingcampaign{
			Id:                &campaignId,
			Name:              &campaignName,
			CampaignStatus:    &campaignStatus,
			MessagesPerMinute: &messagesPerMin,
			ContactList:       &platformclientv2.Domainentityref{Id: &contactListId},
			Version:           &version,
			WhatsAppConfig: &platformclientv2.Whatsappconfig{
				WhatsAppColumns:     &whatsAppColumns,
				WhatsAppIntegration: &platformclientv2.Addressableentityref{Id: &integrationId},
				ContentTemplate:     &platformclientv2.Domainentityref{Id: &templateId},
			},
		}, &platformclientv2.APIResponse{StatusCode: http.StatusOK}, nil
	}

	internalProxy = proxy
	defer func() { internalProxy = nil }()

	ctx := context.Background()
	gCloud := &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}

	resourceSchema := ResourceOutboundMessagingcampaign().Schema
	resourceDataMap := buildMessagingCampaignWhatsAppResourceMap(campaignName, contactListId, campaignStatus, messagesPerMin, integrationId, templateId, whatsAppColumns)

	d := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
	d.SetId(campaignId)

	diag := readOutboundMessagingcampaign(ctx, d, gCloud)
	assert.Equal(t, false, diag.HasError())
	assert.Equal(t, campaignId, d.Id())
	assert.Equal(t, campaignName, d.Get("name").(string))
	assert.Equal(t, campaignStatus, d.Get("campaign_status").(string))
	assert.Equal(t, messagesPerMin, d.Get("messages_per_minute").(int))
}

func TestUnitMessagingCampaignWhatsAppUpdate(t *testing.T) {
	var (
		campaignId      = uuid.NewString()
		campaignName    = "Updated WhatsApp Campaign"
		contactListId   = uuid.NewString()
		integrationId   = uuid.NewString()
		templateId      = uuid.NewString()
		messagesPerMin  = 20
		campaignStatus  = "off"
		whatsAppColumns = []string{"whatsappCol"}
		version         = 2
	)

	proxy := &outboundMessagingcampaignProxy{}

	proxy.getOutboundMessagingcampaignByIdAttr = func(ctx context.Context, p *outboundMessagingcampaignProxy, id string) (*platformclientv2.Messagingcampaign, *platformclientv2.APIResponse, error) {
		return &platformclientv2.Messagingcampaign{
			Id:                &campaignId,
			Name:              &campaignName,
			CampaignStatus:    &campaignStatus,
			MessagesPerMinute: &messagesPerMin,
			ContactList:       &platformclientv2.Domainentityref{Id: &contactListId},
			Version:           &version,
			WhatsAppConfig: &platformclientv2.Whatsappconfig{
				WhatsAppColumns:     &whatsAppColumns,
				WhatsAppIntegration: &platformclientv2.Addressableentityref{Id: &integrationId},
				ContentTemplate:     &platformclientv2.Domainentityref{Id: &templateId},
			},
		}, &platformclientv2.APIResponse{StatusCode: http.StatusOK}, nil
	}

	proxy.updateOutboundMessagingcampaignAttr = func(ctx context.Context, p *outboundMessagingcampaignProxy, id string, campaign *platformclientv2.Messagingcampaign) (*platformclientv2.Messagingcampaign, *platformclientv2.APIResponse, error) {
		assert.Equal(t, campaignId, id)
		assert.NotNil(t, campaign.WhatsAppConfig)
		assert.Equal(t, integrationId, *campaign.WhatsAppConfig.WhatsAppIntegration.Id)
		assert.Equal(t, templateId, *campaign.WhatsAppConfig.ContentTemplate.Id)
		assert.Equal(t, version, *campaign.Version)

		campaign.Id = &campaignId
		return campaign, nil, nil
	}

	internalProxy = proxy
	defer func() { internalProxy = nil }()

	ctx := context.Background()
	gCloud := &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}

	resourceSchema := ResourceOutboundMessagingcampaign().Schema
	resourceDataMap := buildMessagingCampaignWhatsAppResourceMap(campaignName, contactListId, campaignStatus, messagesPerMin, integrationId, templateId, whatsAppColumns)

	d := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
	d.SetId(campaignId)

	diag := updateOutboundMessagingcampaign(ctx, d, gCloud)
	assert.Equal(t, false, diag.HasError())
	assert.Equal(t, campaignId, d.Id())
}

func buildMessagingCampaignWhatsAppResourceMap(name, contactListId, campaignStatus string, messagesPerMin int, integrationId, templateId string, whatsAppColumns []string) map[string]interface{} {
	columns := make([]interface{}, len(whatsAppColumns))
	for i, c := range whatsAppColumns {
		columns[i] = c
	}
	return map[string]interface{}{
		"name":                name,
		"contact_list_id":     contactListId,
		"campaign_status":     campaignStatus,
		"messages_per_minute": messagesPerMin,
		"always_running":      false,
		"whats_app_config": []interface{}{
			map[string]interface{}{
				"whats_app_columns":        columns,
				"whats_app_integration_id": integrationId,
				"content_template_id":      templateId,
			},
		},
	}
}
