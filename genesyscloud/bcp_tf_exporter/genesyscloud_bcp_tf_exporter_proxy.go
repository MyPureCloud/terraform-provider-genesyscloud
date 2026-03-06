package bcp_tf_exporter

import (
	"context"

	dependentconsumers "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/dependent_consumers"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/mypurecloud/platform-client-sdk-go/v179/platformclientv2"
)

type BcpExporterProxy struct {
	ClientConfig               *platformclientv2.Configuration
	GetFlowDependenciesAttr    getFlowDependenciesFunc
	GetAllWithPooledClientAttr getPooledClientFunc
}

type getFlowDependenciesFunc func(ctx context.Context, p *BcpExporterProxy, resourceInfo resourceExporter.ResourceInfo) (resourceExporter.ResourceIDMetaMap, *resourceExporter.DependencyResource, error)
type getPooledClientFunc func(ctx context.Context, method provider.GetCustomConfigFunc) (resourceExporter.ResourceIDMetaMap, *resourceExporter.DependencyResource, []string, diag.Diagnostics)

var InternalProxy *BcpExporterProxy

func GetBcpExporterProxy(ClientConfig *platformclientv2.Configuration) *BcpExporterProxy {
	return newBcpExporterProxy(ClientConfig)
}

func newBcpExporterProxy(ClientConfig *platformclientv2.Configuration) *BcpExporterProxy {
	if InternalProxy == nil {
		InternalProxy = &BcpExporterProxy{
			GetAllWithPooledClientAttr: getPooledClientFn,
		}
	}

	if ClientConfig != nil {
		InternalProxy.ClientConfig = ClientConfig
		InternalProxy.GetFlowDependenciesAttr = getFlowDependenciesFn
	}
	return InternalProxy
}

func (p *BcpExporterProxy) GetFlowDependencies(ctx context.Context, resourceInfo resourceExporter.ResourceInfo) (resourceExporter.ResourceIDMetaMap, *resourceExporter.DependencyResource, error) {
	return p.GetFlowDependenciesAttr(ctx, p, resourceInfo)
}

func (p *BcpExporterProxy) GetAllWithPooledClient(ctx context.Context, method provider.GetCustomConfigFunc) (resourceExporter.ResourceIDMetaMap, *resourceExporter.DependencyResource, []string, diag.Diagnostics) {
	return p.GetAllWithPooledClientAttr(ctx, method)
}

func getPooledClientFn(ctx context.Context, method provider.GetCustomConfigFunc) (resourceExporter.ResourceIDMetaMap, *resourceExporter.DependencyResource, []string, diag.Diagnostics) {
	resourceFunc := provider.GetAllWithPooledClientCustom(method)
	// Pass the context through - don't replace it
	resources, dependsMap, totalFlowResources, err := resourceFunc(ctx)
	if err != nil {
		return nil, nil, totalFlowResources, err
	}
	return resources, dependsMap, totalFlowResources, err
}

func getFlowDependenciesFn(ctx context.Context, p *BcpExporterProxy, resourceInfo resourceExporter.ResourceInfo) (resourceExporter.ResourceIDMetaMap, *resourceExporter.DependencyResource, error) {
	// Don't use pooled client - the pooled client's HTTP client has a canceled context
	// Create a completely fresh API instance with the same auth token
	dependentconsumers.InternalProxy = nil
	
	// Create fresh config and copy only the essential auth settings
	freshConfig := platformclientv2.NewConfiguration()
	freshConfig.BasePath = p.ClientConfig.BasePath
	freshConfig.AccessToken = p.ClientConfig.AccessToken
	freshConfig.DefaultHeader = make(map[string]string)
	for k, v := range p.ClientConfig.DefaultHeader {
		freshConfig.DefaultHeader[k] = v
	}
	
	depConsumerProxy := dependentconsumers.GetDependentConsumerProxy(freshConfig)
	resources, dependsMap, _, err := depConsumerProxy.GetDependentConsumers(ctx, resourceInfo, []string{})
	if err != nil {
		return nil, nil, err
	}
	return resources, dependsMap, nil
}
