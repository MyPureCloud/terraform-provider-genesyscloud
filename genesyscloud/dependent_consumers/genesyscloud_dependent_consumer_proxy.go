package dependent_consumers

import (
	"context"
	"fmt"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"

	"log"
	"strings"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/stringmap"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/mypurecloud/platform-client-sdk-go/v179/platformclientv2"
)

type DependentConsumerProxy struct {
	ClientConfig                   *platformclientv2.Configuration
	ArchitectApi                   *platformclientv2.ArchitectApi
	RetrieveDependentConsumersAttr retrieveDependentConsumersFunc
	GetPooledClientAttr            retrievePooledClientFunc
}

var gflow = "genesyscloud_flow"

func (p *DependentConsumerProxy) GetDependentConsumers(ctx context.Context, resourceKeys resourceExporter.ResourceInfo, totalFlowResources []string) (resourceExporter.ResourceIDMetaMap, *resourceExporter.DependencyResource, []string, error) {
	return p.RetrieveDependentConsumersAttr(ctx, p, resourceKeys, totalFlowResources)
}

func (p *DependentConsumerProxy) GetAllWithPooledClient(method provider.GetCustomConfigFunc) (resourceExporter.ResourceIDMetaMap, *resourceExporter.DependencyResource, []string, diag.Diagnostics) {
	return p.GetPooledClientAttr(method)
}

type retrieveDependentConsumersFunc func(ctx context.Context, p *DependentConsumerProxy, resourceKeys resourceExporter.ResourceInfo, totalFlowResources []string) (resourceExporter.ResourceIDMetaMap, *resourceExporter.DependencyResource, []string, error)
type retrievePooledClientFunc func(method provider.GetCustomConfigFunc) (resourceExporter.ResourceIDMetaMap, *resourceExporter.DependencyResource, []string, diag.Diagnostics)

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

func retrievePooledClientFn(method provider.GetCustomConfigFunc) (resourceExporter.ResourceIDMetaMap, *resourceExporter.DependencyResource, []string, diag.Diagnostics) {
	resourceFunc := provider.GetAllWithPooledClientCustom(method)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	resources, dependsMap, totalFlowResources, err := resourceFunc(ctx)
	if err != nil {
		return nil, nil, totalFlowResources, err
	}
	return resources, dependsMap, totalFlowResources, err
}

func retrieveDependentConsumersFn(ctx context.Context, p *DependentConsumerProxy, resourceKeys resourceExporter.ResourceInfo, totalFlowResources []string) (resourceExporter.ResourceIDMetaMap, *resourceExporter.DependencyResource, []string, error) {
	resourceKey := resourceKeys.State.ID
	resourceLabel := resourceKeys.BlockLabel
	log.Printf("[DEBUG_ENTRY] retrieveDependentConsumersFn: START processing resourceKey=%s, resourceLabel=%s, resourceType=%s, totalFlowResourcesCount=%d", resourceKey, resourceLabel, resourceKeys.Type, len(totalFlowResources))
	dependsMap := make(map[string][]string)
	architectDependencies := make(map[string][]string)
	dependentResources, dependsMap, cyclicDependsList, err, totalFlowResources := fetchDepConsumers(ctx, p, resourceKeys.Type, resourceKey, resourceLabel, make(resourceExporter.ResourceIDMetaMap), dependsMap, architectDependencies, make([]string, 0), totalFlowResources)

	if err != nil {
		return nil, nil, totalFlowResources, err
	}
	finalDependsMap := buildDependsMap(dependentResources, dependsMap, resourceKey)
	log.Printf("[DEBUG_ENTRY] retrieveDependentConsumersFn: END processing resourceKey=%s, resourceLabel=%s, dependentResourcesCount=%d, finalDependsMapForKey=%v, totalFlowResourcesCount=%d", resourceKey, resourceLabel, len(dependentResources), finalDependsMap[resourceKey], len(totalFlowResources))
	return dependentResources, &resourceExporter.DependencyResource{
		DependsMap:        finalDependsMap,
		CyclicDependsList: cyclicDependsList,
	}, totalFlowResources, nil
}

func fetchDepConsumers(ctx context.Context,
	p *DependentConsumerProxy,
	resType string,
	resourceKey string,
	resourceLabel string,
	resources resourceExporter.ResourceIDMetaMap,
	dependsMap map[string][]string,
	architectDependencies map[string][]string,
	cyclicDependsList []string,
	totalFlowResources []string) (resourceExporter.ResourceIDMetaMap, map[string][]string, []string, error, []string) {
	if resType == gflow {
		alreadyProcessed := util.StringExists(resourceKey, totalFlowResources)
		log.Printf("[DEBUG_FLOW] fetchDepConsumers called: resourceKey=%s, resourceLabel=%s, alreadyInTotalFlowResources=%v, totalFlowResourcesCount=%d", resourceKey, resourceLabel, alreadyProcessed, len(totalFlowResources))
		if alreadyProcessed {
			log.Printf("[DEBUG_FLOW] SKIPPING flow %s (%s) - already in totalFlowResources", resourceKey, resourceLabel)
		}
	}
	if resType == gflow && !util.StringExists(resourceKey, totalFlowResources) {
		// Fetches MetaData for the Flow
		data, _, err := p.ArchitectApi.GetFlow(resourceKey, false)
		if err != nil {
			log.Printf("Error calling GetFlow: %v\n", err)
		}
		// Fetch Dependent Consumed Resources only for Published Versions
		if data != nil && data.PublishedVersion != nil && data.PublishedVersion.Id != nil {
			flowTypeObjectMaps := SetFlowTypeObjectMaps()
			objectType, flowTypeExists := flowTypeObjectMaps[*data.VarType]
			log.Printf("[DEBUG_FLOW] Flow %s (%s): VarType=%s, flowTypeExists=%v, objectType=%s", resourceKey, resourceLabel, *data.VarType, flowTypeExists, objectType)
			if flowTypeExists {
				// Mark this flow as being processed EARLY to prevent re-entry during recursive calls
				// Only add after confirming: flow exists, has published version, and has valid flow type
				log.Printf("[DEBUG_FLOW] ADDING flow %s (%s) to totalFlowResources (before: %d items)", resourceKey, resourceLabel, len(totalFlowResources))
				totalFlowResources = append(totalFlowResources, resourceKey)
				pageCount := 1
				const pageSize = 100
				dependencies, _, err := p.ArchitectApi.GetArchitectDependencytrackingConsumedresources(resourceKey, *data.PublishedVersion.Id, objectType, nil, pageCount, pageSize)
				if err != nil {
					return nil, nil, nil, err, totalFlowResources
				}
				log.Printf("Retrieved dependencies for ID %s", resourceKey)

				pageCount = *dependencies.PageCount

				// return empty dependsMap and  resources
				if dependencies.Entities == nil || len(*dependencies.Entities) == 0 {
					log.Printf("Retrieved dependencies for ID  noresult %v, resourceKey %s, length %d", resources, resourceKey, len(resources))
					return resources, dependsMap, cyclicDependsList, nil, totalFlowResources
				}

				// iterate dependencies
				if pageCount < 2 {
					resources, dependsMap, cyclicDependsList, totalFlowResources, err = iterateDependencies(dependencies, resources, dependsMap, ctx, p, resourceKey, architectDependencies, cyclicDependsList, resourceLabel, totalFlowResources)
					if err != nil {
						return nil, nil, nil, err, totalFlowResources
					}
					log.Printf("Retrieved dependencies for resourceKey %s, resources %v, length %d", resourceKey, resources, len(resources))
					return resources, dependsMap, cyclicDependsList, nil, totalFlowResources
				}

				for pageNum := 1; pageNum <= pageCount; pageNum++ {
					dependencies, _, err := p.ArchitectApi.GetArchitectDependencytrackingConsumedresources(resourceKey, *data.PublishedVersion.Id, objectType, nil, pageNum, pageSize)

					if err != nil {
						return nil, nil, nil, err, totalFlowResources
					}
					if dependencies.Entities == nil || len(*dependencies.Entities) == 0 {
						break
					}
					resources, dependsMap, cyclicDependsList, totalFlowResources, err = iterateDependencies(dependencies, resources, dependsMap, ctx, p, resourceKey, architectDependencies, cyclicDependsList, resourceLabel, totalFlowResources)
					if err != nil {
						return nil, nil, nil, err, totalFlowResources
					}
				}
			}
		}
	}
	log.Printf("Retrieved dependencies for ID %v, resourceKey %s, length %d", resources, resourceKey, len(resources))
	return resources, dependsMap, cyclicDependsList, nil, totalFlowResources
}

func buildDependsMap(resources resourceExporter.ResourceIDMetaMap, dependsMap map[string][]string, id string) map[string][]string {
	dependsList := make([]string, 0)
	for depId, meta := range resources {
		resource := strings.Split(meta.BlockLabel, "::::")
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
	resourceLabel string,
	totalFlowResources []string) (resourceExporter.ResourceIDMetaMap, map[string][]string, []string, []string, error) {
	var err error
	dependentConsumerMap := SetDependentObjectMaps()
	log.Printf("[DEBUG_DEPS] iterateDependencies for flow key=%s, label=%s, entityCount=%d", key, resourceLabel, len(*dependencies.Entities))
	for _, consumer := range *dependencies.Entities {
		resourceType, exists := getResourceType(consumer, dependentConsumerMap)
		log.Printf("[DEBUG_DEPS] Processing consumer: Id=%s, Name=%s, VarType=%s, mappedResourceType=%s, exists=%v (parent flow: %s)",
			*consumer.Id, *consumer.Name, *consumer.VarType, resourceType, exists, resourceLabel)
		if exists {
			resources, architectDependencies = processResource(consumer, resourceType, resources, architectDependencies, key)
			log.Printf("[DEBUG_DEPS] Added resource %s.%s as dependency of %s (total resources now: %d)", resourceType, *consumer.Id, resourceLabel, len(resources))
			if resourceType == gflow && *consumer.Id != key {
				if !isDependencyPresent(architectDependencies, *consumer.Id, key) {
					resources, dependsMap, totalFlowResources, err = fetchAndProcessDependentConsumers(ctx, p, consumer, architectDependencies, resources, dependsMap, cyclicDependsList, totalFlowResources, resourceType)
					if err != nil {
						return nil, nil, nil, totalFlowResources, err
					}
				} else {
					cyclicDependsList = append(cyclicDependsList, gflow+"."+*consumer.Name+" , "+gflow+"."+resourceLabel)
					log.Printf("cyclic Dependencies Identified %v for %v", cyclicDependsList, *consumer.Name)
					continue
				}
			}
		}
	}
	return resources, dependsMap, cyclicDependsList, totalFlowResources, nil
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
		resources[*consumer.Id] = &resourceExporter.ResourceMeta{BlockLabel: resourceFilter}
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
	resources resourceExporter.ResourceIDMetaMap,
	dependsMap map[string][]string,
	cyclicDependsList []string,
	totalFlowResources []string,
	resourceType string) (resourceExporter.ResourceIDMetaMap, map[string][]string, []string, error) {
	log.Printf("[DEBUG_RECURSIVE] fetchAndProcessDependentConsumers: Starting recursive call for consumer Id=%s, Name=%s, totalFlowResourcesCount=%d", *consumer.Id, *consumer.Name, len(totalFlowResources))
	innerDependentResources, innerDependsMap, cyclicDependsList, err, totalFlowResources := fetchDepConsumers(ctx, p, resourceType, *consumer.Id, *consumer.Name, make(resourceExporter.ResourceIDMetaMap), make(map[string][]string), architectDependencies, cyclicDependsList, totalFlowResources)
	log.Printf("[DEBUG_RECURSIVE] fetchAndProcessDependentConsumers: Completed for consumer Id=%s, Name=%s, innerResourcesCount=%d, totalFlowResourcesCount=%d", *consumer.Id, *consumer.Name, len(innerDependentResources), len(totalFlowResources))

	// Merge inner resources back to parent's resources map so dependencies are properly propagated
	for id, meta := range innerDependentResources {
		if _, exists := resources[id]; !exists {
			resources[id] = meta
			log.Printf("[DEBUG_RECURSIVE] Merged inner resource %s back to parent resources", id)
		}
	}

	dependsMap = stringmap.MergeMaps(dependsMap, buildDependsMap(innerDependentResources, innerDependsMap, *consumer.Id))
	return resources, dependsMap, totalFlowResources, err
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
