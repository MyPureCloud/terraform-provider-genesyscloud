package dependent_consumers

import (
	"context"
	gcloud "terraform-provider-genesyscloud/genesyscloud"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/mypurecloud/platform-client-sdk-go/v115/platformclientv2"
)

type DependentConsumerProxy struct {
	ClientConfig                   *platformclientv2.Configuration
	ArchitectApi                   *platformclientv2.ArchitectApi
	RetrieveDependentConsumersAttr retrieveDependentConsumersFunc
	GetPooledClientAttr            retrievePooledClientFunc
}

func (p *DependentConsumerProxy) GetDependentConsumers(ctx context.Context, resourceKeys resourceExporter.ResourceInfo) (resourceExporter.ResourceIDMetaMap, error) {
	return p.RetrieveDependentConsumersAttr(ctx, p, resourceKeys)
}

func (p *DependentConsumerProxy) GetAllWithPooledClient(method gcloud.GetAllConfigFunc) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	return p.GetPooledClientAttr(method)
}

type retrieveDependentConsumersFunc func(ctx context.Context, p *DependentConsumerProxy, resourceKeys resourceExporter.ResourceInfo) (resourceExporter.ResourceIDMetaMap, error)
type retrievePooledClientFunc func(method gcloud.GetAllConfigFunc) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics)

var InternalProxy *DependentConsumerProxy

// getDependentConsumerProxy acts as a singleton to for the InternalProxy.
func GetDependentConsumerProxy(ClientConfig *platformclientv2.Configuration) *DependentConsumerProxy {
	return newDependentConsumerProxy(ClientConfig)
}

// newDependentConsumerProxy initializes the ruleset proxy with all of the data needed to communicate with Genesys Cloud
func newDependentConsumerProxy(ClientConfig *platformclientv2.Configuration) *DependentConsumerProxy {
	if InternalProxy == nil {
		InternalProxy = &DependentConsumerProxy{
			GetPooledClientAttr: retrievePooledClientFn,
		}
	}

	if ClientConfig != nil {
		api := platformclientv2.NewArchitectApiWithConfig(ClientConfig)
		InternalProxy.ClientConfig = ClientConfig
		InternalProxy.ArchitectApi = api
		InternalProxy.RetrieveDependentConsumersAttr = retrieveDependentConsumersFn
	}

	return InternalProxy
}

func retrievePooledClientFn(method gcloud.GetAllConfigFunc) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	resourcefunc := gcloud.GetAllWithPooledClient(method)
	ctx, _ := context.WithCancel(context.Background())
	resources, err := resourcefunc(ctx)
	if err != nil {
		return nil, err
	}
	return resources, err
}

func retrieveDependentConsumersFn(ctx context.Context, p *DependentConsumerProxy, resourceKeys resourceExporter.ResourceInfo) (resourceExporter.ResourceIDMetaMap, error) {
	resourceKey := resourceKeys.State.ID
	resourceType := resourceKeys.Type
	resources := make(resourceExporter.ResourceIDMetaMap)
	dependentConsumerMap := SetDependentObjectMaps()
	objectType, exists := dependentConsumerMap[resourceType]
	if exists {
		pageCount := 1
		for pageNum := 1; pageNum <= pageCount; pageNum++ {
			const pageSize = 100
			dependencies, _, err := p.ArchitectApi.GetArchitectDependencytrackingConsumingresources(resourceKey, objectType, nil, "", pageNum, pageSize, "")
			if err != nil {
				return nil, err
			}
			if dependencies.Entities == nil || len(*dependencies.Entities) == 0 {
				break
			}

			for _, consumer := range *dependencies.Entities {
				resources[*consumer.Id] = &resourceExporter.ResourceMeta{Name: *consumer.Name}
			}
			pageCount = *dependencies.PageCount
		}
	}
	return resources, nil
}
