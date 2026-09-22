package webdeployments_deployment_identity_resolution

import (
	"context"

	"github.com/mypurecloud/platform-client-sdk-go/v195/platformclientv2"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	webDeploymentsDeployment "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/webdeployments_deployment"
)

var internalProxy *webDeploymentIdentityResolutionProxy

type getWebDeploymentIdentityResolutionFunc func(ctx context.Context, p *webDeploymentIdentityResolutionProxy, deploymentId string) (*platformclientv2.Deploymentidentityresolutionconfig, *platformclientv2.APIResponse, error)
type putWebDeploymentIdentityResolutionFunc func(ctx context.Context, p *webDeploymentIdentityResolutionProxy, deploymentId string, config platformclientv2.Deploymentidentityresolutionconfig) (*platformclientv2.Deploymentidentityresolutionconfig, *platformclientv2.APIResponse, error)

type webDeploymentIdentityResolutionProxy struct {
	clientConfig      *platformclientv2.Configuration
	webDeploymentsApi *platformclientv2.WebDeploymentsApi

	getWebDeploymentIdentityResolutionAttr getWebDeploymentIdentityResolutionFunc
	putWebDeploymentIdentityResolutionAttr putWebDeploymentIdentityResolutionFunc

	webDeploymentsProxy *webDeploymentsDeployment.WebDeploymentsProxy
}

func newWebDeploymentIdentityResolutionProxy(clientConfig *platformclientv2.Configuration) *webDeploymentIdentityResolutionProxy {
	api := platformclientv2.NewWebDeploymentsApiWithConfig(clientConfig)
	webDeploymentsProxy := webDeploymentsDeployment.GetWebDeploymentsProxy(clientConfig)

	return &webDeploymentIdentityResolutionProxy{
		clientConfig:                           clientConfig,
		webDeploymentsApi:                      api,
		getWebDeploymentIdentityResolutionAttr: getWebDeploymentIdentityResolutionFn,
		putWebDeploymentIdentityResolutionAttr: putWebDeploymentIdentityResolutionFn,
		webDeploymentsProxy:                    webDeploymentsProxy,
	}
}

func getWebDeploymentIdentityResolutionProxy(clientConfig *platformclientv2.Configuration) *webDeploymentIdentityResolutionProxy {
	if internalProxy == nil {
		internalProxy = newWebDeploymentIdentityResolutionProxy(clientConfig)
	}
	return internalProxy
}

func (p *webDeploymentIdentityResolutionProxy) getWebDeploymentIdentityResolution(ctx context.Context, deploymentId string) (*platformclientv2.Deploymentidentityresolutionconfig, *platformclientv2.APIResponse, error) {
	return p.getWebDeploymentIdentityResolutionAttr(ctx, p, deploymentId)
}

func (p *webDeploymentIdentityResolutionProxy) putWebDeploymentIdentityResolution(ctx context.Context, deploymentId string, config platformclientv2.Deploymentidentityresolutionconfig) (*platformclientv2.Deploymentidentityresolutionconfig, *platformclientv2.APIResponse, error) {
	return p.putWebDeploymentIdentityResolutionAttr(ctx, p, deploymentId, config)
}

func getWebDeploymentIdentityResolutionFn(ctx context.Context, p *webDeploymentIdentityResolutionProxy, deploymentId string) (*platformclientv2.Deploymentidentityresolutionconfig, *platformclientv2.APIResponse, error) {
	// Set resource context for SDK debug logging
	ctx = provider.EnsureResourceContext(ctx, ResourceType)
	return p.webDeploymentsApi.GetWebdeploymentsDeploymentIdentityresolution(deploymentId)
}

func putWebDeploymentIdentityResolutionFn(ctx context.Context, p *webDeploymentIdentityResolutionProxy, deploymentId string, config platformclientv2.Deploymentidentityresolutionconfig) (*platformclientv2.Deploymentidentityresolutionconfig, *platformclientv2.APIResponse, error) {
	// Set resource context for SDK debug logging
	ctx = provider.EnsureResourceContext(ctx, ResourceType)
	return p.webDeploymentsApi.PutWebdeploymentsDeploymentIdentityresolution(deploymentId, config)
}
