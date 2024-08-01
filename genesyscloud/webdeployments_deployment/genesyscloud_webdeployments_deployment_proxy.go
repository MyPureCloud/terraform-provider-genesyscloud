package webdeployments_deployment

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

var internalProxy *webDeploymentsProxy

type getAllWebDeploymentsFunc func(ctx context.Context, p *webDeploymentsProxy) (*platformclientv2.Expandablewebdeploymententitylisting, *platformclientv2.APIResponse, error)
type getWebDeploymentsFunc func(ctx context.Context, p *webDeploymentsProxy, deployId string) (*platformclientv2.Webdeployment, *platformclientv2.APIResponse, error)
type createWebdeploymentsFunc func(ctx context.Context, p *webDeploymentsProxy, deployment platformclientv2.Webdeployment) (*platformclientv2.Webdeployment, *platformclientv2.APIResponse, error)
type updateWebdeploymentsFunc func(ctx context.Context, p *webDeploymentsProxy, deploymentId string, deployment platformclientv2.Webdeployment) (*platformclientv2.Webdeployment, *platformclientv2.APIResponse, error)
type deleteWebdeploymentsFunc func(ctx context.Context, p *webDeploymentsProxy, deploymentId string) (*platformclientv2.APIResponse, error)
type determineLatestVersionFunc func(ctx context.Context, p *webDeploymentsProxy, configurationId string) (string, []string, diag.Diagnostics)

type webDeploymentsProxy struct {
	clientConfig      *platformclientv2.Configuration
	webDeploymentsApi *platformclientv2.WebDeploymentsApi

	getAllWebDeploymentsAttr   getAllWebDeploymentsFunc
	getWebDeploymentAttr       getWebDeploymentsFunc
	createWebDeploymentAttr    createWebdeploymentsFunc
	updateWebDeploymentAttr    updateWebdeploymentsFunc
	deleteWebDeploymentAttr    deleteWebdeploymentsFunc
	determineLatestVersionAttr determineLatestVersionFunc
}

func newWebDeploymentsProxy(clientConfig *platformclientv2.Configuration) *webDeploymentsProxy {
	webDeploymentsApi := platformclientv2.NewWebDeploymentsApiWithConfig(clientConfig)

	return &webDeploymentsProxy{
		clientConfig:               clientConfig,
		webDeploymentsApi:          webDeploymentsApi,
		getAllWebDeploymentsAttr:   getAllWebDeploymentsFn,
		getWebDeploymentAttr:       getWebDeploymentsFn,
		createWebDeploymentAttr:    createWebdeploymentsFn,
		updateWebDeploymentAttr:    updateWebdeploymentsFn,
		deleteWebDeploymentAttr:    deleteWebdeploymentsFn,
		determineLatestVersionAttr: determineLatestVersionFn,
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
func (p *webDeploymentsProxy) determineLatestVersion(ctx context.Context, configurationId string) (string, []string, diag.Diagnostics) {
	return p.determineLatestVersionAttr(ctx, p, configurationId)
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

func determineLatestVersionFn(ctx context.Context, p *webDeploymentsProxy, configurationId string) (string, []string, diag.Diagnostics) {
	version := ""
	draft := "DRAFT"
	versionList := []string{}
	err := util.WithRetries(ctx, 30*time.Second, func() *retry.RetryError {
		versions, resp, getErr := p.webDeploymentsApi.GetWebdeploymentsConfigurationVersions(configurationId)
		if getErr != nil {
			if util.IsStatus404(resp) {
				return retry.RetryableError(fmt.Errorf("Failed to determine latest version %s", getErr))
			}
			log.Printf("Failed to determine latest version. Defaulting to DRAFT. Details: %s", getErr)
			version = draft
			return retry.NonRetryableError(fmt.Errorf("Failed to determine latest version %s", getErr))
		}

		maxVersion := 0
		for _, v := range *versions.Entities {
			if *v.Version == draft {
				versionList = append(versionList, *v.Version)
				continue
			}
			APIVersion, err := strconv.Atoi(*v.Version)
			if err != nil {
				log.Printf("Failed to convert version %s to an integer", *v.Version)
			} else {
				versionList = append(versionList, *v.Version)
				if APIVersion > maxVersion {
					maxVersion = APIVersion
				}
			}
		}

		if maxVersion == 0 {
			version = draft
		} else {
			version = strconv.Itoa(maxVersion)
		}

		return nil
	})
	if err != nil {
		return "", nil, err
	}
	return version, versionList, nil
}
