package bcp_tf_exporter

import (
	"context"

	dependentconsumers "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/dependent_consumers"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/mypurecloud/platform-client-sdk-go/v192/platformclientv2"
)

type bcpExporterProxy struct {
	clientConfig               *platformclientv2.Configuration
	getFlowDependenciesAttr    getFlowDependenciesFunc
	getAllWithPooledClientAttr getPooledClientFunc
}

type getFlowDependenciesFunc func(ctx context.Context, p *bcpExporterProxy, resourceInfo resourceExporter.ResourceInfo) (resourceExporter.ResourceIDMetaMap, *resourceExporter.DependencyResource, error)
type getPooledClientFunc func(ctx context.Context, method provider.GetCustomConfigFunc) (resourceExporter.ResourceIDMetaMap, *resourceExporter.DependencyResource, []string, diag.Diagnostics)

var internalProxy *bcpExporterProxy

func getBcpExporterProxy(ClientConfig *platformclientv2.Configuration) *bcpExporterProxy {
	if internalProxy == nil {
		internalProxy = newBcpExporterProxy(ClientConfig)
	}
	return internalProxy
}

func newBcpExporterProxy(clientConfig *platformclientv2.Configuration) *bcpExporterProxy {
	return &bcpExporterProxy{
		clientConfig:               clientConfig,
		getAllWithPooledClientAttr: getPooledClientFn,
		getFlowDependenciesAttr:    getFlowDependenciesFn,
	}
}

func (p *bcpExporterProxy) GetFlowDependencies(ctx context.Context, resourceInfo resourceExporter.ResourceInfo) (resourceExporter.ResourceIDMetaMap, *resourceExporter.DependencyResource, error) {
	return p.getFlowDependenciesAttr(ctx, p, resourceInfo)
}

func (p *bcpExporterProxy) GetAllWithPooledClient(ctx context.Context, method provider.GetCustomConfigFunc) (resourceExporter.ResourceIDMetaMap, *resourceExporter.DependencyResource, []string, diag.Diagnostics) {
	return p.getAllWithPooledClientAttr(ctx, method)
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

func getFlowDependenciesFn(ctx context.Context, p *bcpExporterProxy, resourceInfo resourceExporter.ResourceInfo) (resourceExporter.ResourceIDMetaMap, *resourceExporter.DependencyResource, error) {
	// Create fresh config to avoid pooled client's canceled context issue
	// Only copy essential auth settings for efficiency
	freshConfig := platformclientv2.NewConfiguration()
	freshConfig.BasePath = p.clientConfig.BasePath
	freshConfig.AccessToken = p.clientConfig.AccessToken
	freshConfig.DefaultHeader = p.clientConfig.DefaultHeader // Direct assignment is safe for read-only use

	depConsumerProxy := dependentconsumers.GetDependentConsumerProxy(freshConfig)
	resources, dependsMap, _, err := depConsumerProxy.GetDependentConsumers(ctx, resourceInfo, []string{})
	if err != nil {
		return nil, nil, err
	}
	return resources, dependsMap, nil
}
