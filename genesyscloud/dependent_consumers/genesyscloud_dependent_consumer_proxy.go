package dependent_consumers

import (
	"context"
	"fmt"
	"log"
	"strings"
	gcloud "terraform-provider-genesyscloud/genesyscloud"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	"terraform-provider-genesyscloud/genesyscloud/util/stringmap"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/mypurecloud/platform-client-sdk-go/v119/platformclientv2"
)

type DependentConsumerProxy struct {
	ClientConfig                   *platformclientv2.Configuration
	ArchitectApi                   *platformclientv2.ArchitectApi
	RetrieveDependentConsumersAttr retrieveDependentConsumersFunc
	GetPooledClientAttr            retrievePooledClientFunc
}

func (p *DependentConsumerProxy) GetDependentConsumers(ctx context.Context, resourceKeys resourceExporter.ResourceInfo) (resourceExporter.ResourceIDMetaMap, map[string][]string, error) {
	return p.RetrieveDependentConsumersAttr(ctx, p, resourceKeys)
}

func (p *DependentConsumerProxy) GetAllWithPooledClient(method gcloud.GetCustomConfigFunc) (resourceExporter.ResourceIDMetaMap, map[string][]string, diag.Diagnostics) {
	return p.GetPooledClientAttr(method)
}

type retrieveDependentConsumersFunc func(ctx context.Context, p *DependentConsumerProxy, resourceKeys resourceExporter.ResourceInfo) (resourceExporter.ResourceIDMetaMap, map[string][]string, error)
type retrievePooledClientFunc func(method gcloud.GetCustomConfigFunc) (resourceExporter.ResourceIDMetaMap, map[string][]string, diag.Diagnostics)

var InternalProxy *DependentConsumerProxy

// GetDependentConsumerProxy acts as a singleton to for the InternalProxy.
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

func retrievePooledClientFn(method gcloud.GetCustomConfigFunc) (resourceExporter.ResourceIDMetaMap, map[string][]string, diag.Diagnostics) {
	resourceFunc := gcloud.GetAllWithPooledClientCustom(method)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	resources, dependsMap, err := resourceFunc(ctx)
	if err != nil {
		return nil, nil, err
	}
	return resources, dependsMap, err
}

func retrieveDependentConsumersFn(ctx context.Context, p *DependentConsumerProxy, resourceKeys resourceExporter.ResourceInfo) (resourceExporter.ResourceIDMetaMap, map[string][]string, error) {
	resourceKey := resourceKeys.State.ID
	dependsMap := make(map[string][]string)
	dependentResources, dependsMap, err := fetchDepConsumers(ctx, p, resourceKeys.Type, resourceKey, make(resourceExporter.ResourceIDMetaMap), dependsMap)
	if err != nil {
		return nil, nil, err
	}
	return dependentResources, buildDependsMap(dependentResources, dependsMap, resourceKey), nil
}

func fetchDepConsumers(ctx context.Context, p *DependentConsumerProxy, resType string, resourceKey string, resources resourceExporter.ResourceIDMetaMap, dependsMap map[string][]string) (resourceExporter.ResourceIDMetaMap, map[string][]string, error) {
	if resType == "genesyscloud_flow" {
		// Fetches MetaData for the Flow
		data, _, err := p.ArchitectApi.GetFlow(resourceKey, false)
		if err != nil {
			log.Printf("Error calling GetFlow: %v\n", err)
		}
		// Fetch Dependent Consumed Resources only for Published Versions
		if data != nil && data.PublishedVersion != nil && data.PublishedVersion.Id != nil {
			flowTypeObjectMaps := SetFlowTypeObjectMaps()
			objectType, flowTypeExists := flowTypeObjectMaps[*data.VarType]
			if flowTypeExists {
				pageCount := 1
				const pageSize = 100
				dependencies, _, err := p.ArchitectApi.GetArchitectDependencytrackingConsumedresources(resourceKey, *data.PublishedVersion.Id, objectType, nil, pageCount, pageSize)

				if err != nil {
					return nil, nil, err
				}

				pageCount = *dependencies.PageCount

				// return empty dependsMap and  resources
				if dependencies.Entities == nil || len(*dependencies.Entities) == 0 {
					return resources, dependsMap, nil
				}

				// iterate dependencies
				if pageCount < 2 {
					resources, dependsMap, err = iterateDependencies(dependencies, resources, dependsMap, ctx, p)
					log.Printf("dependencies GetArchitectDependencytrackingConsumedresources: %v\n", resources)
					if err != nil {
						return nil, nil, err
					}
					return resources, dependsMap, nil
				}

				for pageNum := 1; pageNum <= pageCount; pageNum++ {
					dependencies, _, err := p.ArchitectApi.GetArchitectDependencytrackingConsumedresources(resourceKey, *data.PublishedVersion.Id, objectType, nil, pageNum, pageSize)
					if err != nil {
						return nil, nil, err
					}
					if dependencies.Entities == nil || len(*dependencies.Entities) == 0 {
						break
					}
					resources, dependsMap, err = iterateDependencies(dependencies, resources, dependsMap, ctx, p)
					if err != nil {
						return nil, nil, err
					}
				}
			}
		}
	}
	return resources, dependsMap, nil
}

func buildDependsMap(resources resourceExporter.ResourceIDMetaMap, dependsMap map[string][]string, id string) map[string][]string {
	dependsList := make([]string, 0)
	for depId, meta := range resources {
		resource := strings.Split(meta.Name, "::::")
		if id != depId {
			dependsList = append(dependsList, fmt.Sprintf("%s.%s", resource[0], depId))
		}
	}
	dependsMap[id] = dependsList
	return dependsMap
}

// This private function includes iteration of the dependent Consumers and build DependsList for each Resource
// This also checks for dependent flows and again export those dependencies
func iterateDependencies(dependencies *platformclientv2.Consumedresourcesentitylisting, resources resourceExporter.ResourceIDMetaMap, dependsMap map[string][]string, ctx context.Context, p *DependentConsumerProxy) (resourceExporter.ResourceIDMetaMap, map[string][]string, error) {
	dependentConsumerMap := SetDependentObjectMaps()
	for _, consumer := range *dependencies.Entities {
		resourceType, exists := dependentConsumerMap[*consumer.VarType]
		if exists {
			resourceFilter := resourceType + "::::" + *consumer.Name
			if _, resourceExists := resources[*consumer.Id]; !resourceExists {
				resources[*consumer.Id] = &resourceExporter.ResourceMeta{Name: resourceFilter}
				if resourceType == "genesyscloud_flow" {
					innerDependentResources, innerDependsMap, err := fetchDepConsumers(ctx, p, resourceType, *consumer.Id, resources, dependsMap)
					dependsMap = stringmap.MergeMaps(dependsMap, buildDependsMap(innerDependentResources, innerDependsMap, *consumer.Id))
					if err != nil {
						return nil, nil, err
					}
				}
			}
		}
	}
	return resources, dependsMap, nil
}
