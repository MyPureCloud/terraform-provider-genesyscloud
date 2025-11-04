package apple_integration

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v171/platformclientv2"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/stretchr/testify/assert"
)

func TestUnitAppleIntegrationCreate(t *testing.T) {
	var (
		name                  = "Unit test Apple Integration"
		messagesForBusinessId = "test-business-" + uuid.NewString()
		appleIntegrationId    = uuid.NewString()
		businessName          = "Test Business"
		logoUrl               = "https://logo.url"
	)

	appleProxy := &appleIntegrationProxy{}

	appleProxy.getAppleIntegrationByIdAttr = func(ctx context.Context, p *appleIntegrationProxy, id string) (*platformclientv2.Appleintegration, *platformclientv2.APIResponse, error) {
		assert.Equal(t, appleIntegrationId, id)

		integrationConfig := platformclientv2.Appleintegration{
			Id:                    &appleIntegrationId,
			Name:                  &name,
			MessagesForBusinessId: &messagesForBusinessId,
			BusinessName:          &businessName,
			LogoUrl:               &logoUrl,
		}

		apiResponse := &platformclientv2.APIResponse{
			StatusCode: http.StatusOK,
		}

		return &integrationConfig, apiResponse, nil
	}

	appleProxy.createAppleIntegrationAttr = func(ctx context.Context, p *appleIntegrationProxy, appleIntegration *platformclientv2.Appleintegration) (*platformclientv2.Appleintegration, *platformclientv2.APIResponse, error) {
		assert.Equal(t, name, *appleIntegration.Name)
		assert.Equal(t, messagesForBusinessId, *appleIntegration.MessagesForBusinessId)
		assert.Equal(t, businessName, *appleIntegration.BusinessName)
		assert.Equal(t, logoUrl, *appleIntegration.LogoUrl)

		integrationResponse := platformclientv2.Appleintegration{
			Id:                    &appleIntegrationId,
			Name:                  appleIntegration.Name,
			MessagesForBusinessId: appleIntegration.MessagesForBusinessId,
			BusinessName:          appleIntegration.BusinessName,
			LogoUrl:               appleIntegration.LogoUrl,
		}

		return &integrationResponse, nil, nil
	}

	internalProxy = appleProxy
	defer func() { internalProxy = nil }()

	ctx := context.Background()
	gCloud := &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}

	resourceSchema := ResourceAppleIntegration().Schema
	resourceDataMap := buildAppleIntegrationResourceMap(appleIntegrationId, name, messagesForBusinessId, businessName, logoUrl)

	d := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
	d.SetId(appleIntegrationId)

	diag := createAppleIntegration(ctx, d, gCloud)
	assert.Equal(t, false, diag.HasError())
	assert.Equal(t, appleIntegrationId, d.Id())
}

func TestUnitAppleIntegrationRead(t *testing.T) {
	var (
		name                  = "Unit test Apple Integration"
		messagesForBusinessId = "test-business-" + uuid.NewString()
		appleIntegrationId    = uuid.NewString()
		businessName          = "Test Business"
		logoUrl               = "https://logo.url"
	)

	appleProxy := &appleIntegrationProxy{}

	appleProxy.getAppleIntegrationByIdAttr = func(ctx context.Context, p *appleIntegrationProxy, id string) (*platformclientv2.Appleintegration, *platformclientv2.APIResponse, error) {
		assert.Equal(t, appleIntegrationId, id)

		integrationConfig := platformclientv2.Appleintegration{
			Id:                    &appleIntegrationId,
			Name:                  &name,
			MessagesForBusinessId: &messagesForBusinessId,
			BusinessName:          &businessName,
			LogoUrl:               &logoUrl,
		}

		apiResponse := &platformclientv2.APIResponse{
			StatusCode: http.StatusOK,
		}

		return &integrationConfig, apiResponse, nil
	}

	internalProxy = appleProxy
	defer func() { internalProxy = nil }()

	ctx := context.Background()
	gCloud := &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}

	resourceSchema := ResourceAppleIntegration().Schema
	resourceDataMap := buildAppleIntegrationResourceMap(appleIntegrationId, name, messagesForBusinessId, businessName, logoUrl)

	d := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
	d.SetId(appleIntegrationId)

	diag := readAppleIntegration(ctx, d, gCloud)
	assert.Equal(t, false, diag.HasError())
	assert.Equal(t, appleIntegrationId, d.Id())
	assert.Equal(t, name, d.Get("name").(string))
	assert.Equal(t, messagesForBusinessId, d.Get("messages_for_business_id").(string))
	assert.Equal(t, businessName, d.Get("business_name").(string))
	assert.Equal(t, logoUrl, d.Get("logo_url").(string))
}

func TestUnitAppleIntegrationUpdate(t *testing.T) {
	var (
		appleIntegrationId    = uuid.NewString()
		updatedName           = "Updated unit test Apple Integration"
		messagesForBusinessId = "test-business-" + uuid.NewString()
		businessName          = "Updated Business"
		logoUrl               = "https://updated-logo.url"
	)

	appleProxy := &appleIntegrationProxy{}

	appleProxy.getAppleIntegrationByIdAttr = func(ctx context.Context, p *appleIntegrationProxy, id string) (*platformclientv2.Appleintegration, *platformclientv2.APIResponse, error) {
		assert.Equal(t, appleIntegrationId, id)

		integrationConfig := platformclientv2.Appleintegration{
			Id:                    &appleIntegrationId,
			Name:                  &updatedName,
			MessagesForBusinessId: &messagesForBusinessId,
			BusinessName:          &businessName,
			LogoUrl:               &logoUrl,
		}

		apiResponse := &platformclientv2.APIResponse{
			StatusCode: http.StatusOK,
		}

		return &integrationConfig, apiResponse, nil
	}

	appleProxy.updateAppleIntegrationAttr = func(ctx context.Context, p *appleIntegrationProxy, id string, appleIntegration *platformclientv2.Appleintegration) (*platformclientv2.Appleintegration, *platformclientv2.APIResponse, error) {
		assert.Equal(t, updatedName, *appleIntegration.Name)
		assert.Equal(t, messagesForBusinessId, *appleIntegration.MessagesForBusinessId)
		assert.Equal(t, businessName, *appleIntegration.BusinessName)
		assert.Equal(t, logoUrl, *appleIntegration.LogoUrl)

		integrationConfig := platformclientv2.Appleintegration{
			Id:                    &appleIntegrationId,
			Name:                  appleIntegration.Name,
			MessagesForBusinessId: appleIntegration.MessagesForBusinessId,
			BusinessName:          appleIntegration.BusinessName,
			LogoUrl:               appleIntegration.LogoUrl,
		}

		apiResponse := &platformclientv2.APIResponse{
			StatusCode: http.StatusOK,
		}

		return &integrationConfig, apiResponse, nil
	}

	internalProxy = appleProxy
	defer func() { internalProxy = nil }()

	ctx := context.Background()
	gCloud := &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}

	resourceSchema := ResourceAppleIntegration().Schema
	resourceDataMap := buildAppleIntegrationResourceMap(appleIntegrationId, updatedName, messagesForBusinessId, businessName, logoUrl)

	d := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
	d.SetId(appleIntegrationId)

	diag := updateAppleIntegration(ctx, d, gCloud)
	assert.Equal(t, false, diag.HasError())
	assert.Equal(t, appleIntegrationId, d.Id())
	assert.Equal(t, updatedName, d.Get("name").(string))
}

func TestUnitAppleIntegrationDelete(t *testing.T) {
	var (
		appleIntegrationId    = uuid.NewString()
		name                  = "Unit test Apple Integration"
		messagesForBusinessId = "test-business-" + uuid.NewString()
		businessName          = "Test Business"
		logoUrl               = "https://logo.url"
	)

	appleProxy := &appleIntegrationProxy{}

	appleProxy.deleteAppleIntegrationAttr = func(ctx context.Context, p *appleIntegrationProxy, id string) (*platformclientv2.APIResponse, error) {
		assert.Equal(t, appleIntegrationId, id)

		apiResponse := &platformclientv2.APIResponse{
			StatusCode: http.StatusNoContent,
		}

		return apiResponse, nil
	}

	appleProxy.getAppleIntegrationByIdAttr = func(ctx context.Context, p *appleIntegrationProxy, id string) (*platformclientv2.Appleintegration, *platformclientv2.APIResponse, error) {
		assert.Equal(t, appleIntegrationId, id)
		apiResponse := &platformclientv2.APIResponse{
			StatusCode: http.StatusNotFound,
		}

		return nil, apiResponse, fmt.Errorf("Not found")
	}

	internalProxy = appleProxy
	defer func() { internalProxy = nil }()

	ctx := context.Background()
	gCloud := &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}

	resourceSchema := ResourceAppleIntegration().Schema
	resourceDataMap := buildAppleIntegrationResourceMap(appleIntegrationId, name, messagesForBusinessId, businessName, logoUrl)

	d := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
	d.SetId(appleIntegrationId)

	diag := deleteAppleIntegration(ctx, d, gCloud)
	assert.Nil(t, diag)
	assert.Equal(t, appleIntegrationId, d.Id())
}

func TestUnitGetAllAppleIntegrations(t *testing.T) {
	var (
		id1   = uuid.NewString()
		name1 = "Apple Integration 1"
		id2   = uuid.NewString()
		name2 = "Apple Integration 2"
	)

	appleProxy := &appleIntegrationProxy{}
	appleProxy.getAllAppleIntegrationAttr = func(ctx context.Context, p *appleIntegrationProxy) (*[]platformclientv2.Appleintegration, *platformclientv2.APIResponse, error) {
		return &[]platformclientv2.Appleintegration{
			{
				Id:   &id1,
				Name: &name1,
			},
			{
				Id:   &id2,
				Name: &name2,
			},
		}, nil, nil
	}

	internalProxy = appleProxy
	defer func() { internalProxy = nil }()

	ctx := context.Background()
	clientConfig := &platformclientv2.Configuration{}

	integrations, diag := getAllAppleIntegrations(ctx, clientConfig)
	assert.Nil(t, diag)
	assert.Equal(t, 2, len(integrations))
	assert.Contains(t, integrations, id1)
	assert.Contains(t, integrations, id2)
}

func TestUnitDataSourceAppleIntegrationRead(t *testing.T) {
	targetId := uuid.NewString()
	targetName := "MyTargetAppleIntegration"
	appleProxy := &appleIntegrationProxy{}
	appleProxy.getAppleIntegrationIdByNameAttr = func(ctx context.Context, p *appleIntegrationProxy, name string) (string, *platformclientv2.APIResponse, bool, error) {
		assert.Equal(t, targetName, name)
		return targetId, nil, false, nil
	}
	appleProxy.getAllAppleIntegrationAttr = func(ctx context.Context, p *appleIntegrationProxy) (*[]platformclientv2.Appleintegration, *platformclientv2.APIResponse, error) {
		return &[]platformclientv2.Appleintegration{
			{
				Id:   &targetId,
				Name: &targetName,
			},
		}, nil, nil
	}
	internalProxy = appleProxy
	defer func() { internalProxy = nil }()
	ctx := context.Background()
	gcloud := &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}

	resourceSchema := DataSourceAppleIntegration().Schema

	resourceDataMap := map[string]interface{}{
		"name": targetName,
	}

	d := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
	dataSourceAppleIntegrationRead(ctx, d, gcloud)
	assert.Equal(t, targetId, d.Id())
}

func buildAppleIntegrationResourceMap(id, name, messagesForBusinessId, businessName, logoUrl string) map[string]interface{} {
	return map[string]interface{}{
		"id":                       id,
		"name":                     name,
		"messages_for_business_id": messagesForBusinessId,
		"business_name":            businessName,
		"logo_url":                 logoUrl,
	}
}
