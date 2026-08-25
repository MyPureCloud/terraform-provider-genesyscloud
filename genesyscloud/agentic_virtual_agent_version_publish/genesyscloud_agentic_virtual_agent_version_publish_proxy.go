package agentic_virtual_agent_version_publish

import (
	"context"

	customapi "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/custom_api_client"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"

	"github.com/mypurecloud/platform-client-sdk-go/v195/platformclientv2"
)

/*
   genesyscloud_agentic_virtual_agent_version_publish_proxy.go contains the proxy for
   managing AVA version publish jobs.

   API Endpoints:
   - POST  /api/v2/agentic/virtualagents/{agentId}/versions/{versionId}/jobs     → Create publish job
   - GET   /api/v2/agentic/virtualagents/{agentId}/versions/{versionId}/jobs/{jobId} → Poll job status
   - GET   /api/v2/agentic/virtualagents/{agentId}/versions/{versionId}           → Read version status
*/

const basePath = "/api/v2/agentic/virtualagents"

var internalProxy *publishProxy

// Function type definitions
type createPublishJobFunc func(ctx context.Context, p *publishProxy, agentId, versionId string, req *PublishJobRequest) (*PublishJobResponse, *platformclientv2.APIResponse, error)
type getPublishJobStatusFunc func(ctx context.Context, p *publishProxy, agentId, versionId, jobId string) (*PublishJobResponse, *platformclientv2.APIResponse, error)
type getVersionStatusFunc func(ctx context.Context, p *publishProxy, agentId, versionId string) (*VersionStatusResponse, *platformclientv2.APIResponse, error)

// publishProxy holds the API client and function attributes.
type publishProxy struct {
	clientConfig            *platformclientv2.Configuration
	customApiClient         *customapi.Client
	createPublishJobAttr    createPublishJobFunc
	getPublishJobStatusAttr getPublishJobStatusFunc
	getVersionStatusAttr    getVersionStatusFunc
}

// newPublishProxy initializes the proxy.
func newPublishProxy(clientConfig *platformclientv2.Configuration) *publishProxy {
	return &publishProxy{
		clientConfig:            clientConfig,
		customApiClient:         customapi.NewClient(clientConfig, ResourceType),
		createPublishJobAttr:    createPublishJobFn,
		getPublishJobStatusAttr: getPublishJobStatusFn,
		getVersionStatusAttr:    getVersionStatusFn,
	}
}

// getPublishProxy returns the singleton proxy instance.
func getPublishProxy(clientConfig *platformclientv2.Configuration) *publishProxy {
	if internalProxy == nil {
		internalProxy = newPublishProxy(clientConfig)
	}
	return internalProxy
}

// Public proxy methods

func (p *publishProxy) createPublishJob(ctx context.Context, agentId, versionId string, req *PublishJobRequest) (*PublishJobResponse, *platformclientv2.APIResponse, error) {
	return p.createPublishJobAttr(ctx, p, agentId, versionId, req)
}

func (p *publishProxy) getPublishJobStatus(ctx context.Context, agentId, versionId, jobId string) (*PublishJobResponse, *platformclientv2.APIResponse, error) {
	return p.getPublishJobStatusAttr(ctx, p, agentId, versionId, jobId)
}

func (p *publishProxy) getVersionStatus(ctx context.Context, agentId, versionId string) (*VersionStatusResponse, *platformclientv2.APIResponse, error) {
	return p.getVersionStatusAttr(ctx, p, agentId, versionId)
}

// Implementation functions

func createPublishJobFn(ctx context.Context, p *publishProxy, agentId, versionId string, req *PublishJobRequest) (*PublishJobResponse, *platformclientv2.APIResponse, error) {
	ctx = provider.EnsureResourceContext(ctx, ResourceType)
	path := basePath + "/" + agentId + "/versions/" + versionId + "/jobs"
	return customapi.Do[PublishJobResponse](ctx, p.customApiClient, customapi.MethodPost, path, req, nil)
}

func getPublishJobStatusFn(ctx context.Context, p *publishProxy, agentId, versionId, jobId string) (*PublishJobResponse, *platformclientv2.APIResponse, error) {
	ctx = provider.EnsureResourceContext(ctx, ResourceType)
	path := basePath + "/" + agentId + "/versions/" + versionId + "/jobs/" + jobId
	return customapi.Do[PublishJobResponse](ctx, p.customApiClient, customapi.MethodGet, path, nil, nil)
}

func getVersionStatusFn(ctx context.Context, p *publishProxy, agentId, versionId string) (*VersionStatusResponse, *platformclientv2.APIResponse, error) {
	ctx = provider.EnsureResourceContext(ctx, ResourceType)
	path := basePath + "/" + agentId + "/versions/" + versionId
	return customapi.Do[VersionStatusResponse](ctx, p.customApiClient, customapi.MethodGet, path, nil, nil)
}
