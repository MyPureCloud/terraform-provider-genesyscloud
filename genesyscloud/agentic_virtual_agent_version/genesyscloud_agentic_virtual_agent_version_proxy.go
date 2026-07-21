package agentic_virtual_agent_version

import (
	"context"

	customapi "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/custom_api_client"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"

	"github.com/mypurecloud/platform-client-sdk-go/v193/platformclientv2"
)

/*
   genesyscloud_agentic_virtual_agent_version_proxy.go contains the proxy structures and methods
   that interact with the Genesys Cloud API for AVA versions.

   API Endpoints:
   - POST   /api/v2/agentic/virtualagents/{virtualAgentId}/versions           → Create version
   - GET    /api/v2/agentic/virtualagents/{virtualAgentId}/versions/{versionId} → Read version
   - PATCH  /api/v2/agentic/virtualagents/{virtualAgentId}/versions/{versionId} → Update version
   - DELETE → Does not exist. Version destroy is state-only (no-op).
*/

const ResourceType = "genesyscloud_agentic_virtual_agent_version"
const basePath = "/api/v2/agentic/virtualagents"

var internalProxy *agenticVirtualAgentVersionProxy

// Function type definitions
type createVersionFunc func(ctx context.Context, p *agenticVirtualAgentVersionProxy, agentId string, version *AgenticVirtualAgentVersionCreate) (*AgenticVirtualAgentVersionResponse, *platformclientv2.APIResponse, error)
type getVersionByIdFunc func(ctx context.Context, p *agenticVirtualAgentVersionProxy, agentId string, versionId string) (*AgenticVirtualAgentVersionResponse, *platformclientv2.APIResponse, error)
type updateVersionFunc func(ctx context.Context, p *agenticVirtualAgentVersionProxy, agentId string, versionId string, version *AgenticVirtualAgentVersionUpdate) (*AgenticVirtualAgentVersionResponse, *platformclientv2.APIResponse, error)

// agenticVirtualAgentVersionProxy holds all the methods and clients needed to interact with the API.
type agenticVirtualAgentVersionProxy struct {
	clientConfig       *platformclientv2.Configuration
	customApiClient    *customapi.Client
	createVersionAttr  createVersionFunc
	getVersionByIdAttr getVersionByIdFunc
	updateVersionAttr  updateVersionFunc
}

// newAgenticVirtualAgentVersionProxy initializes the proxy.
func newAgenticVirtualAgentVersionProxy(clientConfig *platformclientv2.Configuration) *agenticVirtualAgentVersionProxy {
	return &agenticVirtualAgentVersionProxy{
		clientConfig:       clientConfig,
		customApiClient:    customapi.NewClient(clientConfig, ResourceType),
		createVersionAttr:  createVersionFn,
		getVersionByIdAttr: getVersionByIdFn,
		updateVersionAttr:  updateVersionFn,
	}
}

// getAgenticVirtualAgentVersionProxy returns the singleton proxy instance.
func getAgenticVirtualAgentVersionProxy(clientConfig *platformclientv2.Configuration) *agenticVirtualAgentVersionProxy {
	if internalProxy == nil {
		internalProxy = newAgenticVirtualAgentVersionProxy(clientConfig)
	}
	return internalProxy
}

// Public proxy methods

func (p *agenticVirtualAgentVersionProxy) createVersion(ctx context.Context, agentId string, version *AgenticVirtualAgentVersionCreate) (*AgenticVirtualAgentVersionResponse, *platformclientv2.APIResponse, error) {
	return p.createVersionAttr(ctx, p, agentId, version)
}

func (p *agenticVirtualAgentVersionProxy) getVersionById(ctx context.Context, agentId string, versionId string) (*AgenticVirtualAgentVersionResponse, *platformclientv2.APIResponse, error) {
	return p.getVersionByIdAttr(ctx, p, agentId, versionId)
}

func (p *agenticVirtualAgentVersionProxy) updateVersion(ctx context.Context, agentId string, versionId string, version *AgenticVirtualAgentVersionUpdate) (*AgenticVirtualAgentVersionResponse, *platformclientv2.APIResponse, error) {
	return p.updateVersionAttr(ctx, p, agentId, versionId, version)
}

// Implementation functions

func createVersionFn(ctx context.Context, p *agenticVirtualAgentVersionProxy, agentId string, version *AgenticVirtualAgentVersionCreate) (*AgenticVirtualAgentVersionResponse, *platformclientv2.APIResponse, error) {
	ctx = provider.EnsureResourceContext(ctx, ResourceType)
	path := basePath + "/" + agentId + "/versions"
	return customapi.Do[AgenticVirtualAgentVersionResponse](ctx, p.customApiClient, customapi.MethodPost, path, version, nil)
}

func getVersionByIdFn(ctx context.Context, p *agenticVirtualAgentVersionProxy, agentId string, versionId string) (*AgenticVirtualAgentVersionResponse, *platformclientv2.APIResponse, error) {
	ctx = provider.EnsureResourceContext(ctx, ResourceType)
	path := basePath + "/" + agentId + "/versions/" + versionId
	return customapi.Do[AgenticVirtualAgentVersionResponse](ctx, p.customApiClient, customapi.MethodGet, path, nil, nil)
}

func updateVersionFn(ctx context.Context, p *agenticVirtualAgentVersionProxy, agentId string, versionId string, version *AgenticVirtualAgentVersionUpdate) (*AgenticVirtualAgentVersionResponse, *platformclientv2.APIResponse, error) {
	ctx = provider.EnsureResourceContext(ctx, ResourceType)
	path := basePath + "/" + agentId + "/versions/" + versionId
	return customapi.Do[AgenticVirtualAgentVersionResponse](ctx, p.customApiClient, customapi.MethodPatch, path, version, nil)
}
