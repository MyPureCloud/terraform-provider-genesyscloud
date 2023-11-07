package dependent_consumers

import (
	"context"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"

	"github.com/mypurecloud/platform-client-sdk-go/v115/platformclientv2"
)

type dependentConsumerProxy struct {
	clientConfig                   *platformclientv2.Configuration
	architectApi                   *platformclientv2.ArchitectApi
	retrieveDependentConsumersAttr retrieveDependentConsumersFunc
}

func (p *dependentConsumerProxy) GetDependentConsumers(ctx context.Context, resourceKeys resourceExporter.ResourceInfo) (resourceExporter.ResourceIDMetaMap, error) {
	return p.retrieveDependentConsumersAttr(ctx, p, resourceKeys)
}

type retrieveDependentConsumersFunc func(ctx context.Context, p *dependentConsumerProxy, resourceKeys resourceExporter.ResourceInfo) (resourceExporter.ResourceIDMetaMap, error)

var internalProxy *dependentConsumerProxy

// getDependentConsumerProxy acts as a singleton to for the internalProxy.
func GetDependentConsumerProxy(clientConfig *platformclientv2.Configuration) *dependentConsumerProxy {
	if internalProxy == nil {
		internalProxy = newDependentConsumerProxy(clientConfig)
	}

	return internalProxy
}

// newDependentConsumerProxy initializes the ruleset proxy with all of the data needed to communicate with Genesys Cloud
func newDependentConsumerProxy(clientConfig *platformclientv2.Configuration) *dependentConsumerProxy {
	api := platformclientv2.NewArchitectApiWithConfig(clientConfig)
	return &dependentConsumerProxy{
		clientConfig:                   clientConfig,
		architectApi:                   api,
		retrieveDependentConsumersAttr: retrieveDependentConsumersFn,
	}
}

func retrieveDependentConsumersFn(ctx context.Context, p *dependentConsumerProxy, resourceKeys resourceExporter.ResourceInfo) (resourceExporter.ResourceIDMetaMap, error) {
	resourceKey := resourceKeys.State.ID
	resourceType := resourceKeys.Type
	resources := make(resourceExporter.ResourceIDMetaMap)
	dependentConsumerMap := SetDependentObjectMaps()
	objectType, exists := dependentConsumerMap[resourceType]
	if exists {
		for pageNum := 1; ; pageNum++ {
			const pageSize = 100
			dependencies, _, err := p.architectApi.GetArchitectDependencytrackingConsumingresources(resourceKey, objectType, nil, "", pageNum, pageSize, "")
			if err != nil {
				return nil, err
			}
			if dependencies.Entities == nil || len(*dependencies.Entities) == 0 {
				break
			}

			for _, consumer := range *dependencies.Entities {
				resources[*consumer.Id] = &resourceExporter.ResourceMeta{Name: *consumer.Name}
			}
		}
	}

	return resources, nil
}
