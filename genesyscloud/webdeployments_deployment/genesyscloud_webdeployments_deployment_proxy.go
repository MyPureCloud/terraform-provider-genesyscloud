package webdeployments_deployment

import (
	"context"

	"github.com/mypurecloud/platform-client-sdk-go/v119/platformclientv2"
)

var internalProxy *webDeploymentsProxy

type getAllWebDeploymentsFunc func(ctx context.Context, p *webDeploymentsProxy) (*platformclientv2.Expandablewebdeploymententitylisting, *platformclientv2.APIResponse, error)
type getWebDeploymentsFunc func(ctx context.Context, p *webDeploymentsProxy, deployId string) (*platformclientv2.Webdeployment, *platformclientv2.APIResponse, error)
type createWebdeploymentsFunc func(ctx context.Context, p *webDeploymentsProxy, deployment platformclientv2.Webdeployment) (*platformclientv2.Webdeployment, *platformclientv2.APIResponse, error)
type updateWebdeploymentsFunc func(ctx context.Context, p *webDeploymentsProxy, deploymentId string, deployment platformclientv2.Webdeployment) (*platformclientv2.Webdeployment, *platformclientv2.APIResponse, error)
type deleteWebdeploymentsFunc func(ctx context.Context, p *webDeploymentsProxy, deploymentId string) (*platformclientv2.APIResponse, error)

type webDeploymentsProxy struct {
	clientConfig      *platformclientv2.Configuration
	webDeploymentsApi *platformclientv2.WebDeploymentsApi

	getAllWebDeploymentsAttr getAllWebDeploymentsFunc
	getWebDeploymentAttr     getWebDeploymentsFunc
	createWebDeploymentAttr  createWebdeploymentsFunc
	updateWebDeploymentAttr  updateWebdeploymentsFunc
	deleteWebDeploymentAttr  deleteWebdeploymentsFunc
}

func newWebDeploymentsProxy(clientConfig *platformclientv2.Configuration) *webDeploymentsProxy {
	webDeploymentsApi := platformclientv2.NewWebDeploymentsApiWithConfig(clientConfig)

	return &webDeploymentsProxy{
		clientConfig:             clientConfig,
		webDeploymentsApi:        webDeploymentsApi,
		getAllWebDeploymentsAttr: getAllWebDeploymentsFn,
		getWebDeploymentAttr:     getWebDeploymentsFn,
		createWebDeploymentAttr:  createWebdeploymentsFn,
		updateWebDeploymentAttr:  updateWebdeploymentsFn,
		deleteWebDeploymentAttr:  deleteWebdeploymentsFn,
	}
}

func getWebDeploymentsProxy(clientConfig *platformclientv2.Configuration) *webDeploymentsProxy {
	if internalProxy == nil {
		internalProxy = newWebDeploymentsProxy(clientConfig)
	}
	return internalProxy
}

func (p *webDeploymentsProxy) getWebDeployments(ctx context.Context) (*platformclientv2.Expandablewebdeploymententitylisting, *platformclientv2.APIResponse, error) {
	return p.getAllWebDeploymentsAttr(ctx, p)
}

func (p *webDeploymentsProxy) getWebDeployment(ctx context.Context, deployId string) (*platformclientv2.Webdeployment, *platformclientv2.APIResponse, error) {
	return p.getWebDeploymentAttr(ctx, p, deployId)
}

func (p *webDeploymentsProxy) createWebDeployment(ctx context.Context, deployment platformclientv2.Webdeployment) (*platformclientv2.Webdeployment, *platformclientv2.APIResponse, error) {
	return p.createWebDeploymentAttr(ctx, p, deployment)
}

func (p *webDeploymentsProxy) updateWebDeployment(ctx context.Context, deploymentId string, deployment platformclientv2.Webdeployment) (*platformclientv2.Webdeployment, *platformclientv2.APIResponse, error) {
	return p.updateWebDeploymentAttr(ctx, p, deploymentId, deployment)
}

func (p *webDeploymentsProxy) deleteWebDeployment(ctx context.Context, deploymentId string) (*platformclientv2.APIResponse, error) {
	return p.deleteWebDeploymentAttr(ctx, p, deploymentId)
}

func getAllWebDeploymentsFn(ctx context.Context, p *webDeploymentsProxy) (*platformclientv2.Expandablewebdeploymententitylisting, *platformclientv2.APIResponse, error) {
	return p.webDeploymentsApi.GetWebdeploymentsDeployments([]string{})
}

func getWebDeploymentsFn(ctx context.Context, p *webDeploymentsProxy, deployId string) (*platformclientv2.Webdeployment, *platformclientv2.APIResponse, error) {
	return p.webDeploymentsApi.GetWebdeploymentsDeployment(deployId, []string{})
}

func createWebdeploymentsFn(ctx context.Context, p *webDeploymentsProxy, deployment platformclientv2.Webdeployment) (*platformclientv2.Webdeployment, *platformclientv2.APIResponse, error) {
	return p.webDeploymentsApi.PostWebdeploymentsDeployments(deployment)
}

func updateWebdeploymentsFn(ctx context.Context, p *webDeploymentsProxy, deploymentId string, deployment platformclientv2.Webdeployment) (*platformclientv2.Webdeployment, *platformclientv2.APIResponse, error) {
	return p.webDeploymentsApi.PutWebdeploymentsDeployment(deploymentId, deployment)
}

func deleteWebdeploymentsFn(ctx context.Context, p *webDeploymentsProxy, deploymentId string) (*platformclientv2.APIResponse, error) {
	return p.webDeploymentsApi.DeleteWebdeploymentsDeployment(deploymentId)
}
