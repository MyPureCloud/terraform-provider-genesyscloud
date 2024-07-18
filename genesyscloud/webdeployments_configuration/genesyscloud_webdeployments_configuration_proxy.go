package webdeployments_configuration

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

var internalProxy *webDeploymentsConfigurationProxy

type getAllWebDeploymentsConfigurationFunc func(ctx context.Context, p *webDeploymentsConfigurationProxy) (*platformclientv2.Webdeploymentconfigurationversionentitylisting, *platformclientv2.APIResponse, error)
type getWebdeploymentsConfigurationVersionFunc func(ctx context.Context, p *webDeploymentsConfigurationProxy, id string, version string) (*platformclientv2.Webdeploymentconfigurationversion, *platformclientv2.APIResponse, error)
type determineLatestVersionFunc func(ctx context.Context, p *webDeploymentsConfigurationProxy, configurationId string) string
type deleteWebDeploymentConfigurationFunc func(ctx context.Context, p *webDeploymentsConfigurationProxy, configurationId string) (*platformclientv2.APIResponse, error)
type getWebdeploymentsConfigurationVersionsDraftFunc func(ctx context.Context, p *webDeploymentsConfigurationProxy, configurationId string) (*platformclientv2.Webdeploymentconfigurationversion, *platformclientv2.APIResponse, error)
type createWebdeploymentsConfigurationFunc func(ctx context.Context, p *webDeploymentsConfigurationProxy, configurationVersion platformclientv2.Webdeploymentconfigurationversion) (*platformclientv2.Webdeploymentconfigurationversion, *platformclientv2.APIResponse, error)
type createWebdeploymentsConfigurationVersionsDraftPublishFunc func(ctx context.Context, p *webDeploymentsConfigurationProxy, configurationId string) (*platformclientv2.Webdeploymentconfigurationversion, *platformclientv2.APIResponse, error)
type updateWebdeploymentsConfigurationVersionsDraftFunc func(ctx context.Context, p *webDeploymentsConfigurationProxy, configurationId string, configurationVersion platformclientv2.Webdeploymentconfigurationversion) (*platformclientv2.Webdeploymentconfigurationversion, *platformclientv2.APIResponse, error)

func newWebDeploymentsConfigurationProxy(clientConfig *platformclientv2.Configuration) *webDeploymentsConfigurationProxy {
	webDeploymentsConfigurationApi := platformclientv2.NewWebDeploymentsApiWithConfig(clientConfig)

	return &webDeploymentsConfigurationProxy{
		clientConfig:      clientConfig,
		webDeploymentsApi: webDeploymentsConfigurationApi,

		getAllWebDeploymentConfigurationsAttr:                     getAllWebDeploymentsConfigurationFn,
		determineLatestVersionAttr:                                determineLatestVersionFn,
		getWebdeploymentsConfigurationVersionAttr:                 getWebdeploymentsConfigurationVersionFn,
		deleteWebDeploymentConfigurationAttr:                      deleteWebDeploymentConfigurationFn,
		getWebdeploymentsConfigurationVersionsDraftAttr:           getWebdeploymentsConfigurationVersionsDraftFn,
		createWebdeploymentsConfigurationAttr:                     createWebdeploymentsConfigurationFn,
		createWebdeploymentsConfigurationVersionsDraftPublishAttr: createWebdeploymentsConfigurationVersionsDraftPublishFn,
		updateWebdeploymentsConfigurationVersionsDraftAttr:        updateWebdeploymentsConfigurationVersionsDraftFn,
	}
}

func getWebDeploymentConfigurationsProxy(clientConfig *platformclientv2.Configuration) *webDeploymentsConfigurationProxy {
	if internalProxy == nil {
		internalProxy = newWebDeploymentsConfigurationProxy(clientConfig)
	}
	return internalProxy
}

type webDeploymentsConfigurationProxy struct {
	clientConfig      *platformclientv2.Configuration
	webDeploymentsApi *platformclientv2.WebDeploymentsApi

	getAllWebDeploymentConfigurationsAttr                     getAllWebDeploymentsConfigurationFunc
	getWebdeploymentsConfigurationVersionAttr                 getWebdeploymentsConfigurationVersionFunc
	determineLatestVersionAttr                                determineLatestVersionFunc
	deleteWebDeploymentConfigurationAttr                      deleteWebDeploymentConfigurationFunc
	getWebdeploymentsConfigurationVersionsDraftAttr           getWebdeploymentsConfigurationVersionsDraftFunc
	createWebdeploymentsConfigurationAttr                     createWebdeploymentsConfigurationFunc
	createWebdeploymentsConfigurationVersionsDraftPublishAttr createWebdeploymentsConfigurationVersionsDraftPublishFunc
	updateWebdeploymentsConfigurationVersionsDraftAttr        updateWebdeploymentsConfigurationVersionsDraftFunc
}

func (p *webDeploymentsConfigurationProxy) getWebDeploymentsConfiguration(ctx context.Context) (*platformclientv2.Webdeploymentconfigurationversionentitylisting, *platformclientv2.APIResponse, error) {
	return p.getAllWebDeploymentConfigurationsAttr(ctx, p)
}

func (p *webDeploymentsConfigurationProxy) getWebdeploymentsConfigurationVersion(ctx context.Context, id string, version string) (*platformclientv2.Webdeploymentconfigurationversion, *platformclientv2.APIResponse, error) {
	return p.getWebdeploymentsConfigurationVersionAttr(ctx, p, id, version)
}

func (p *webDeploymentsConfigurationProxy) determineLatestVersion(ctx context.Context, configurationId string) string {
	return p.determineLatestVersionAttr(ctx, p, configurationId)
}

func (p *webDeploymentsConfigurationProxy) deleteWebDeploymentConfiguration(ctx context.Context, configurationId string) (*platformclientv2.APIResponse, error) {
	return p.deleteWebDeploymentConfigurationAttr(ctx, p, configurationId)
}

func (p *webDeploymentsConfigurationProxy) getWebdeploymentsConfigurationVersionsDraft(ctx context.Context, configurationId string) (*platformclientv2.Webdeploymentconfigurationversion, *platformclientv2.APIResponse, error) {
	return p.getWebdeploymentsConfigurationVersionsDraftAttr(ctx, p, configurationId)
}

func (p *webDeploymentsConfigurationProxy) createWebdeploymentsConfiguration(ctx context.Context, configurationVersion platformclientv2.Webdeploymentconfigurationversion) (*platformclientv2.Webdeploymentconfigurationversion, *platformclientv2.APIResponse, error) {
	return p.createWebdeploymentsConfigurationAttr(ctx, p, configurationVersion)
}

func (p *webDeploymentsConfigurationProxy) createWebdeploymentsConfigurationVersionsDraftPublish(ctx context.Context, configurationId string) (*platformclientv2.Webdeploymentconfigurationversion, *platformclientv2.APIResponse, error) {
	return p.createWebdeploymentsConfigurationVersionsDraftPublishAttr(ctx, p, configurationId)
}

func (p *webDeploymentsConfigurationProxy) updateWebdeploymentsConfigurationVersionsDraft(ctx context.Context, configurationId string, configurationVersion platformclientv2.Webdeploymentconfigurationversion) (*platformclientv2.Webdeploymentconfigurationversion, *platformclientv2.APIResponse, error) {
	return p.updateWebdeploymentsConfigurationVersionsDraftAttr(ctx, p, configurationId, configurationVersion)
}

func getAllWebDeploymentsConfigurationFn(ctx context.Context, p *webDeploymentsConfigurationProxy) (*platformclientv2.Webdeploymentconfigurationversionentitylisting, *platformclientv2.APIResponse, error) {
	configurations, resp, getErr := p.webDeploymentsApi.GetWebdeploymentsConfigurations(false)

	if getErr != nil {
		return nil, resp, fmt.Errorf("Failed to get web deployment configurations: %v", getErr)
	}
	return configurations, resp, nil
}

func getWebdeploymentsConfigurationVersionFn(ctx context.Context, p *webDeploymentsConfigurationProxy, id string, version string) (*platformclientv2.Webdeploymentconfigurationversion, *platformclientv2.APIResponse, error) {
	return p.webDeploymentsApi.GetWebdeploymentsConfigurationVersion(id, version)
}

func determineLatestVersionFn(ctx context.Context, p *webDeploymentsConfigurationProxy, configurationId string) string {
	version := ""
	draft := "DRAFT"
	_ = util.WithRetries(ctx, 30*time.Second, func() *retry.RetryError {
		versions, resp, getErr := p.webDeploymentsApi.GetWebdeploymentsConfigurationVersions(configurationId)
		if getErr != nil {
			if util.IsStatus404(resp) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(resourceName, fmt.Sprintf("Failed to determine latest version | error: %s", getErr), resp))
			}
			log.Printf("Failed to determine latest version. Defaulting to DRAFT. Details: %s", getErr)
			version = draft
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(resourceName, fmt.Sprintf("Failed to determine latest version | error: %s", getErr), resp))
		}

		maxVersion := 0
		for _, v := range *versions.Entities {
			if *v.Version == draft {
				continue
			}
			APIVersion, err := strconv.Atoi(*v.Version)
			if err != nil {
				log.Printf("Failed to convert version %s to an integer", *v.Version)
			} else {
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

	return version
}

func deleteWebDeploymentConfigurationFn(ctx context.Context, p *webDeploymentsConfigurationProxy, configurationId string) (*platformclientv2.APIResponse, error) {
	return p.webDeploymentsApi.DeleteWebdeploymentsConfiguration(configurationId)
}

func getWebdeploymentsConfigurationVersionsDraftFn(ctx context.Context, p *webDeploymentsConfigurationProxy, configurationId string) (*platformclientv2.Webdeploymentconfigurationversion, *platformclientv2.APIResponse, error) {
	return p.webDeploymentsApi.GetWebdeploymentsConfigurationVersionsDraft(configurationId)
}

func createWebdeploymentsConfigurationFn(ctx context.Context, p *webDeploymentsConfigurationProxy, configurationVersion platformclientv2.Webdeploymentconfigurationversion) (*platformclientv2.Webdeploymentconfigurationversion, *platformclientv2.APIResponse, error) {
	return p.webDeploymentsApi.PostWebdeploymentsConfigurations(configurationVersion)
}

func createWebdeploymentsConfigurationVersionsDraftPublishFn(ctx context.Context, p *webDeploymentsConfigurationProxy, configurationId string) (*platformclientv2.Webdeploymentconfigurationversion, *platformclientv2.APIResponse, error) {
	return p.webDeploymentsApi.PostWebdeploymentsConfigurationVersionsDraftPublish(configurationId)
}

func updateWebdeploymentsConfigurationVersionsDraftFn(ctx context.Context, p *webDeploymentsConfigurationProxy, configurationId string, configurationVersion platformclientv2.Webdeploymentconfigurationversion) (*platformclientv2.Webdeploymentconfigurationversion, *platformclientv2.APIResponse, error) {
	return p.webDeploymentsApi.PutWebdeploymentsConfigurationVersionsDraft(configurationId, configurationVersion)
}
