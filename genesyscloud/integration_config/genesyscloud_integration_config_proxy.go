package integration_config

import (
	"context"
	"sync"

	"github.com/mypurecloud/platform-client-sdk-go/v195/platformclientv2"
)

// proxy function types
type getIntegrationConfigFunc func(ctx context.Context, p *integrationConfigProxy, integrationId string) (*platformclientv2.Integrationconfiguration, *platformclientv2.APIResponse, error)
type updateIntegrationConfigFunc func(ctx context.Context, p *integrationConfigProxy, integrationId string, config *platformclientv2.Integrationconfiguration) (*platformclientv2.Integrationconfiguration, *platformclientv2.APIResponse, error)

type integrationConfigProxy struct {
	clientConfig                *platformclientv2.Configuration
	integrationsApi             *platformclientv2.IntegrationsApi
	getIntegrationConfigAttr    getIntegrationConfigFunc
	updateIntegrationConfigAttr updateIntegrationConfigFunc
}

var integrationConfigProxyInstance *integrationConfigProxy
var integrationConfigProxyMu sync.Mutex

func newIntegrationConfigProxy(clientConfig *platformclientv2.Configuration) *integrationConfigProxy {
	api := platformclientv2.NewIntegrationsApiWithConfig(clientConfig)
	return &integrationConfigProxy{
		clientConfig:                clientConfig,
		integrationsApi:             api,
		getIntegrationConfigAttr:    getIntegrationConfigFn,
		updateIntegrationConfigAttr: updateIntegrationConfigFn,
	}
}

func getIntegrationConfigProxy(clientConfig *platformclientv2.Configuration) *integrationConfigProxy {
	integrationConfigProxyMu.Lock()
	defer integrationConfigProxyMu.Unlock()

	if integrationConfigProxyInstance == nil {
		integrationConfigProxyInstance = newIntegrationConfigProxy(clientConfig)
	}
	return integrationConfigProxyInstance
}

// Public methods
func (p *integrationConfigProxy) getConfig(ctx context.Context, integrationId string) (*platformclientv2.Integrationconfiguration, *platformclientv2.APIResponse, error) {
	return p.getIntegrationConfigAttr(ctx, p, integrationId)
}

func (p *integrationConfigProxy) updateConfig(ctx context.Context, integrationId string, config *platformclientv2.Integrationconfiguration) (*platformclientv2.Integrationconfiguration, *platformclientv2.APIResponse, error) {
	return p.updateIntegrationConfigAttr(ctx, p, integrationId, config)
}

// Implementation functions
func getIntegrationConfigFn(ctx context.Context, p *integrationConfigProxy, integrationId string) (*platformclientv2.Integrationconfiguration, *platformclientv2.APIResponse, error) {
	return p.integrationsApi.GetIntegrationConfigCurrent(integrationId)
}

func updateIntegrationConfigFn(ctx context.Context, p *integrationConfigProxy, integrationId string, config *platformclientv2.Integrationconfiguration) (*platformclientv2.Integrationconfiguration, *platformclientv2.APIResponse, error) {
	return p.integrationsApi.PutIntegrationConfigCurrent(integrationId, *config)
}
