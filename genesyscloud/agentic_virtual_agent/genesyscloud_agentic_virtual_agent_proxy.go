package agentic_virtual_agent

import (
	"context"
	"fmt"

	customapi "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/custom_api_client"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"

	"github.com/mypurecloud/platform-client-sdk-go/v193/platformclientv2"

	rc "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_cache"
)

/*
   genesyscloud_agentic_virtual_agent_proxy.go contains the proxy structures and methods
   that interact with the Genesys Cloud API using the custom API client.

   Since the SDK generator does not support the discriminator/polymorphism pattern
   used by the AVA APIs, we use the custom_api_client package to make direct REST calls.
   This is the same pattern used by the Guides resource.
*/

const basePath = "/api/v2/agentic/virtualagents"

var internalProxy *agenticVirtualAgentProxy

// Function type definitions for proxy methods (allows stubbing in tests)
type createAgenticVirtualAgentFunc func(ctx context.Context, p *agenticVirtualAgentProxy, agent *AgenticVirtualAgentCreate) (*AgenticVirtualAgent, *platformclientv2.APIResponse, error)
type getAgenticVirtualAgentByIdFunc func(ctx context.Context, p *agenticVirtualAgentProxy, id string) (*AgenticVirtualAgent, *platformclientv2.APIResponse, error)
type getAgenticVirtualAgentByNameFunc func(ctx context.Context, p *agenticVirtualAgentProxy, name string) (string, bool, *platformclientv2.APIResponse, error)
type getAllAgenticVirtualAgentsFunc func(ctx context.Context, p *agenticVirtualAgentProxy, name string) (*[]AgenticVirtualAgent, *platformclientv2.APIResponse, error)
type updateAgenticVirtualAgentFunc func(ctx context.Context, p *agenticVirtualAgentProxy, id string, agent *AgenticVirtualAgentUpdate) (*AgenticVirtualAgent, *platformclientv2.APIResponse, error)
type deleteAgenticVirtualAgentFunc func(ctx context.Context, p *agenticVirtualAgentProxy, id string) (*AgenticVirtualAgentJob, *platformclientv2.APIResponse, error)
type getDeleteJobStatusFunc func(ctx context.Context, p *agenticVirtualAgentProxy, agentId string, jobId string) (*AgenticVirtualAgentJob, *platformclientv2.APIResponse, error)

// agenticVirtualAgentProxy holds all the methods and clients needed to interact with the API.
type agenticVirtualAgentProxy struct {
	clientConfig                      *platformclientv2.Configuration
	customApiClient                   *customapi.Client
	createAgenticVirtualAgentAttr     createAgenticVirtualAgentFunc
	getAgenticVirtualAgentByIdAttr    getAgenticVirtualAgentByIdFunc
	getAgenticVirtualAgentByNameAttr  getAgenticVirtualAgentByNameFunc
	getAllAgenticVirtualAgentsAttr    getAllAgenticVirtualAgentsFunc
	updateAgenticVirtualAgentAttr     updateAgenticVirtualAgentFunc
	deleteAgenticVirtualAgentAttr     deleteAgenticVirtualAgentFunc
	getDeleteJobStatusAttr            getDeleteJobStatusFunc
	agentCache                        rc.CacheInterface[AgenticVirtualAgent]
}

// newAgenticVirtualAgentProxy initializes the proxy with all function implementations.
func newAgenticVirtualAgentProxy(clientConfig *platformclientv2.Configuration) *agenticVirtualAgentProxy {
	agentCache := rc.NewResourceCache[AgenticVirtualAgent]()
	return &agenticVirtualAgentProxy{
		clientConfig:                      clientConfig,
		customApiClient:                   customapi.NewClient(clientConfig, ResourceType),
		createAgenticVirtualAgentAttr:     createAgenticVirtualAgentFn,
		getAgenticVirtualAgentByIdAttr:    getAgenticVirtualAgentByIdFn,
		getAgenticVirtualAgentByNameAttr:  getAgenticVirtualAgentByNameFn,
		getAllAgenticVirtualAgentsAttr:    getAllAgenticVirtualAgentsFn,
		updateAgenticVirtualAgentAttr:     updateAgenticVirtualAgentFn,
		deleteAgenticVirtualAgentAttr:     deleteAgenticVirtualAgentFn,
		getDeleteJobStatusAttr:            getDeleteJobStatusFn,
		agentCache:                        agentCache,
	}
}

// getAgenticVirtualAgentProxy returns the singleton proxy instance.
func getAgenticVirtualAgentProxy(clientConfig *platformclientv2.Configuration) *agenticVirtualAgentProxy {
	if internalProxy == nil {
		internalProxy = newAgenticVirtualAgentProxy(clientConfig)
	}
	return internalProxy
}

// Public proxy methods (delegate to function attributes for testability)

func (p *agenticVirtualAgentProxy) createAgenticVirtualAgent(ctx context.Context, agent *AgenticVirtualAgentCreate) (*AgenticVirtualAgent, *platformclientv2.APIResponse, error) {
	return p.createAgenticVirtualAgentAttr(ctx, p, agent)
}

func (p *agenticVirtualAgentProxy) getAgenticVirtualAgentById(ctx context.Context, id string) (*AgenticVirtualAgent, *platformclientv2.APIResponse, error) {
	return p.getAgenticVirtualAgentByIdAttr(ctx, p, id)
}

func (p *agenticVirtualAgentProxy) getAgenticVirtualAgentByName(ctx context.Context, name string) (string, bool, *platformclientv2.APIResponse, error) {
	return p.getAgenticVirtualAgentByNameAttr(ctx, p, name)
}

func (p *agenticVirtualAgentProxy) getAllAgenticVirtualAgents(ctx context.Context, name string) (*[]AgenticVirtualAgent, *platformclientv2.APIResponse, error) {
	return p.getAllAgenticVirtualAgentsAttr(ctx, p, name)
}

func (p *agenticVirtualAgentProxy) updateAgenticVirtualAgent(ctx context.Context, id string, agent *AgenticVirtualAgentUpdate) (*AgenticVirtualAgent, *platformclientv2.APIResponse, error) {
	return p.updateAgenticVirtualAgentAttr(ctx, p, id, agent)
}

func (p *agenticVirtualAgentProxy) deleteAgenticVirtualAgent(ctx context.Context, id string) (*AgenticVirtualAgentJob, *platformclientv2.APIResponse, error) {
	return p.deleteAgenticVirtualAgentAttr(ctx, p, id)
}

func (p *agenticVirtualAgentProxy) getDeleteJobStatus(ctx context.Context, agentId string, jobId string) (*AgenticVirtualAgentJob, *platformclientv2.APIResponse, error) {
	return p.getDeleteJobStatusAttr(ctx, p, agentId, jobId)
}

// Implementation functions

func createAgenticVirtualAgentFn(ctx context.Context, p *agenticVirtualAgentProxy, agent *AgenticVirtualAgentCreate) (*AgenticVirtualAgent, *platformclientv2.APIResponse, error) {
	ctx = provider.EnsureResourceContext(ctx, ResourceType)
	return customapi.Do[AgenticVirtualAgent](ctx, p.customApiClient, customapi.MethodPost, basePath, agent, nil)
}

func getAgenticVirtualAgentByIdFn(ctx context.Context, p *agenticVirtualAgentProxy, id string) (*AgenticVirtualAgent, *platformclientv2.APIResponse, error) {
	ctx = provider.EnsureResourceContext(ctx, ResourceType)
	if agent := rc.GetCacheItem(p.agentCache, id); agent != nil {
		return agent, nil, nil
	}
	return customapi.Do[AgenticVirtualAgent](ctx, p.customApiClient, customapi.MethodGet, basePath+"/"+id, nil, nil)
}

func getAllAgenticVirtualAgentsFn(ctx context.Context, p *agenticVirtualAgentProxy, name string) (*[]AgenticVirtualAgent, *platformclientv2.APIResponse, error) {
	ctx = provider.EnsureResourceContext(ctx, ResourceType)

	var allAgents []AgenticVirtualAgent
	queryParams := customapi.NewQueryParams(map[string]string{"pageSize": "100", "pageNumber": "1"})
	if name != "" {
		queryParams.Set("name", name)
	}

	agents, resp, err := customapi.Do[AgenticVirtualAgentEntityListing](ctx, p.customApiClient, customapi.MethodGet, basePath, nil, queryParams)
	if err != nil {
		return nil, resp, err
	}
	if agents.Entities == nil {
		return &allAgents, resp, nil
	}
	allAgents = append(allAgents, *agents.Entities...)

	for pageNum := 2; pageNum <= *agents.PageCount; pageNum++ {
		queryParams.Set("pageNumber", fmt.Sprintf("%v", pageNum))
		pageAgents, resp, err := customapi.Do[AgenticVirtualAgentEntityListing](ctx, p.customApiClient, customapi.MethodGet, basePath, nil, queryParams)
		if err != nil {
			return nil, resp, err
		}
		if pageAgents.Entities != nil {
			allAgents = append(allAgents, *pageAgents.Entities...)
		}
	}

	for _, agent := range allAgents {
		rc.SetCache(p.agentCache, *agent.Id, agent)
	}
	return &allAgents, resp, nil
}

func getAgenticVirtualAgentByNameFn(ctx context.Context, p *agenticVirtualAgentProxy, name string) (string, bool, *platformclientv2.APIResponse, error) {
	ctx = provider.EnsureResourceContext(ctx, ResourceType)

	agents, resp, err := getAllAgenticVirtualAgentsFn(ctx, p, name)
	if err != nil {
		return "", false, resp, err
	}

	if agents == nil || len(*agents) == 0 {
		return "", true, resp, fmt.Errorf("no agentic virtual agent found with name: %s", name)
	}

	for _, agent := range *agents {
		if agent.Name != nil && *agent.Name == name {
			if agent.Id != nil {
				return *agent.Id, false, resp, nil
			}
			return "", false, resp, fmt.Errorf("agentic virtual agent found but has nil ID: %s", name)
		}
	}

	return "", false, resp, fmt.Errorf("unable to find agentic virtual agent with name %s", name)
}

func updateAgenticVirtualAgentFn(ctx context.Context, p *agenticVirtualAgentProxy, id string, agent *AgenticVirtualAgentUpdate) (*AgenticVirtualAgent, *platformclientv2.APIResponse, error) {
	ctx = provider.EnsureResourceContext(ctx, ResourceType)
	return customapi.Do[AgenticVirtualAgent](ctx, p.customApiClient, customapi.MethodPatch, basePath+"/"+id, agent, nil)
}

func deleteAgenticVirtualAgentFn(ctx context.Context, p *agenticVirtualAgentProxy, id string) (*AgenticVirtualAgentJob, *platformclientv2.APIResponse, error) {
	ctx = provider.EnsureResourceContext(ctx, ResourceType)
	return customapi.Do[AgenticVirtualAgentJob](ctx, p.customApiClient, customapi.MethodDelete, basePath+"/"+id+"/jobs", nil, nil)
}

func getDeleteJobStatusFn(ctx context.Context, p *agenticVirtualAgentProxy, agentId string, jobId string) (*AgenticVirtualAgentJob, *platformclientv2.APIResponse, error) {
	ctx = provider.EnsureResourceContext(ctx, ResourceType)
	return customapi.Do[AgenticVirtualAgentJob](ctx, p.customApiClient, customapi.MethodGet, basePath+"/"+agentId+"/jobs/"+jobId, nil, nil)
}
