package dependent_consumers

import (
	"context"
	"fmt"

	"log"
	"strings"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	"terraform-provider-genesyscloud/genesyscloud/util/stringmap"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

type DependentConsumerProxy struct {
	ClientConfig                   *platformclientv2.Configuration
	ArchitectApi                   *platformclientv2.ArchitectApi
	RetrieveDependentConsumersAttr retrieveDependentConsumersFunc
	GetPooledClientAttr            retrievePooledClientFunc
}

var gflow = "genesyscloud_flow"

func (p *DependentConsumerProxy) GetDependentConsumers(ctx context.Context, resourceKeys resourceExporter.ResourceInfo) (resourceExporter.ResourceIDMetaMap, *resourceExporter.DependencyResource, error) {
	return p.RetrieveDependentConsumersAttr(ctx, p, resourceKeys)
}

func (p *DependentConsumerProxy) GetAllWithPooledClient(method provider.GetCustomConfigFunc) (resourceExporter.ResourceIDMetaMap, *resourceExporter.DependencyResource, diag.Diagnostics) {
	return p.GetPooledClientAttr(method)
}

type retrieveDependentConsumersFunc func(ctx context.Context, p *DependentConsumerProxy, resourceKeys resourceExporter.ResourceInfo) (resourceExporter.ResourceIDMetaMap, *resourceExporter.DependencyResource, error)
type retrievePooledClientFunc func(method provider.GetCustomConfigFunc) (resourceExporter.ResourceIDMetaMap, *resourceExporter.DependencyResource, diag.Diagnostics)

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

func retrievePooledClientFn(method provider.GetCustomConfigFunc) (resourceExporter.ResourceIDMetaMap, *resourceExporter.DependencyResource, diag.Diagnostics) {
	resourceFunc := provider.GetAllWithPooledClientCustom(method)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	resources, dependsMap, err := resourceFunc(ctx)
	if err != nil {
		return nil, nil, err
	}
	return resources, dependsMap, err
}

func retrieveDependentConsumersFn(ctx context.Context, p *DependentConsumerProxy, resourceKeys resourceExporter.ResourceInfo) (resourceExporter.ResourceIDMetaMap, *resourceExporter.DependencyResource, error) {
	resourceKey := resourceKeys.State.ID
	resourceName := resourceKeys.Name
	dependsMap := make(map[string][]string)
	architectDependencies := make(map[string][]string)
	dependentResources, dependsMap, cyclicDependsList, err := fetchDepConsumers(ctx, p, resourceKeys.Type, resourceKey, resourceName, make(resourceExporter.ResourceIDMetaMap), dependsMap, architectDependencies, make([]string, 0))

	if err != nil {
		return nil, nil, err
	}
	return dependentResources, &resourceExporter.DependencyResource{
		DependsMap:        buildDependsMap(dependentResources, dependsMap, resourceKey),
		CyclicDependsList: cyclicDependsList,
	}, nil
}

func fetchDepConsumers(ctx context.Context,
	p *DependentConsumerProxy,
	resType string,
	resourceKey string,
	resourceName string,
	resources resourceExporter.ResourceIDMetaMap,
	dependsMap map[string][]string,
	architectDependencies map[string][]string,
	cyclicDependsList []string) (resourceExporter.ResourceIDMetaMap, map[string][]string, []string, error) {
	if resType == gflow {
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
					return nil, nil, nil, err
				}
				log.Printf("Retrieved dependencies for ID %s", resourceKey)

				pageCount = *dependencies.PageCount

				// return empty dependsMap and  resources
				if dependencies.Entities == nil || len(*dependencies.Entities) == 0 {
					return resources, dependsMap, cyclicDependsList, nil
				}

				// iterate dependencies
				if pageCount < 2 {
					resources, dependsMap, cyclicDependsList, err = iterateDependencies(dependencies, resources, dependsMap, ctx, p, resourceKey, architectDependencies, cyclicDependsList, resourceName)
					if err != nil {
						return nil, nil, nil, err
					}
					return resources, dependsMap, cyclicDependsList, nil
				}

				for pageNum := 1; pageNum <= pageCount; pageNum++ {
					dependencies, _, err := p.ArchitectApi.GetArchitectDependencytrackingConsumedresources(resourceKey, *data.PublishedVersion.Id, objectType, nil, pageNum, pageSize)

					if err != nil {
						return nil, nil, nil, err
					}
					if dependencies.Entities == nil || len(*dependencies.Entities) == 0 {
						break
					}
					resources, dependsMap, cyclicDependsList, err = iterateDependencies(dependencies, resources, dependsMap, ctx, p, resourceKey, architectDependencies, cyclicDependsList, resourceName)
					if err != nil {
						return nil, nil, nil, err
					}
				}
			}
		}
	}
	return resources, dependsMap, cyclicDependsList, nil
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
func iterateDependencies(dependencies *platformclientv2.Consumedresourcesentitylisting,
	resources resourceExporter.ResourceIDMetaMap,
	dependsMap map[string][]string,
	ctx context.Context,
	p *DependentConsumerProxy,
	key string,
	architectDependencies map[string][]string,
	cyclicDependsList []string,
	resourceName string) (resourceExporter.ResourceIDMetaMap, map[string][]string, []string, error) {
	var err error
	dependentConsumerMap := SetDependentObjectMaps()
	for _, consumer := range *dependencies.Entities {
		resourceType, exists := getResourceType(consumer, dependentConsumerMap)
		if exists {
			resources, architectDependencies = processResource(consumer, resourceType, resources, architectDependencies, key)
			if resourceType == gflow && *consumer.Id != key {
				if !isDependencyPresent(architectDependencies, *consumer.Id, key) {
					dependsMap, err = fetchAndProcessDependentConsumers(ctx, p, consumer, architectDependencies, dependsMap, cyclicDependsList)
					if err != nil {
						return nil, nil, nil, err
					}
				} else {
					cyclicDependsList = append(cyclicDependsList, gflow+"."+*consumer.Name+" , "+gflow+"."+resourceName)
					log.Printf("cyclic Dependencies Identified %v for %v", cyclicDependsList, *consumer.Name)
					continue
				}
			}
		}
	}
	return resources, dependsMap, cyclicDependsList, nil
}

func getResourceType(consumer platformclientv2.Dependency, dependentConsumerMap map[string]string) (string, bool) {
	resourceType, exists := dependentConsumerMap[*consumer.VarType]
	return resourceType, exists
}

func getResourceFilter(consumer platformclientv2.Dependency, resourceType string) string {
	return resourceType + "::::" + *consumer.Id
}

func processResource(consumer platformclientv2.Dependency, resourceType string, resources resourceExporter.ResourceIDMetaMap, architectDependencies map[string][]string, key string) (resourceExporter.ResourceIDMetaMap, map[string][]string) {
	resourceFilter := getResourceFilter(consumer, resourceType)
	if _, resourceExists := resources[*consumer.Id]; !resourceExists {
		resources[*consumer.Id] = &resourceExporter.ResourceMeta{Name: resourceFilter}
		if architectDependencies[key] != nil {
			architectDependencies[key] = append(architectDependencies[key], *consumer.Id)
		} else {
			architectDependencies[key] = []string{*consumer.Id}
		}
	}
	return resources, architectDependencies
}

func isDependencyPresent(architectDependencies map[string][]string, consumerId, key string) bool {
	return searchForKeyValue(architectDependencies, consumerId, key)
}

func fetchAndProcessDependentConsumers(ctx context.Context,
	p *DependentConsumerProxy,
	consumer platformclientv2.Dependency,
	architectDependencies map[string][]string,
	dependsMap map[string][]string,
	cyclicDependsList []string) (map[string][]string, error) {
	innerDependentResources, innerDependsMap, cyclicDependsList, err := fetchDepConsumers(ctx, p, *consumer.VarType, *consumer.Id, *consumer.Name, make(resourceExporter.ResourceIDMetaMap), make(map[string][]string), architectDependencies, cyclicDependsList)
	dependsMap = stringmap.MergeMaps(dependsMap, buildDependsMap(innerDependentResources, innerDependsMap, *consumer.Id))
	return dependsMap, err
}

func searchForKeyValue(m map[string][]string, key, value string) bool {
	if stringsList, ok := m[key]; ok {
		return stringInSlice(value, stringsList)
	}
	return false
}
func stringInSlice(str string, list []string) bool {
	for _, v := range list {
		if v == str {
			return true
		}
	}
	return false
}
