package routing_language

// @team: Assignment
// @chat: #genesys-cloud-acd-routing
// @description: Routing configuration service for queues, skills, wrapup codes, and utilization settings. Manages how contacts are distributed to agents based on skills, capacity, and routing rules across all interaction channels.

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/mypurecloud/platform-client-sdk-go/v176/platformclientv2"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_register"
)

const ResourceType = "genesyscloud_routing_language"

// SetRegistrar registers all of the resources, datasources and exporters in the package
func SetRegistrar(regInstance registrar.Registrar) {
	// Framework-only registration (SDKv2 removed)
	regInstance.RegisterFrameworkResource(ResourceType, NewFrameworkRoutingLanguageResource)
	regInstance.RegisterFrameworkDataSource(ResourceType, NewFrameworkRoutingLanguageDataSource)
	regInstance.RegisterExporter(ResourceType, RoutingLanguageExporter())
}

// SDKv2 resource and data source functions removed - Framework-only migration

func RoutingLanguageExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(GetAllRoutingLanguages),
	}
}

// GetAllRoutingLanguages retrieves all routing languages for export using the proxy
func GetAllRoutingLanguages(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	proxy := getRoutingLanguageProxy(clientConfig)
	languages, _, err := proxy.getAllRoutingLanguages(ctx, "")
	if err != nil {
		return nil, diag.Errorf("Failed to get routing languages for export: %v", err)
	}

	if languages == nil {
		return resourceExporter.ResourceIDMetaMap{}, nil
	}

	exportMap := make(resourceExporter.ResourceIDMetaMap)
	for _, language := range *languages {
		exportMap[*language.Id] = &resourceExporter.ResourceMeta{
			BlockLabel: *language.Name,
		}
	}
	return exportMap, nil
}

func GenerateRoutingLanguageResource(
	resourceLabel string,
	name string) string {
	return fmt.Sprintf(`resource "genesyscloud_routing_language" "%s" {
		name = "%s"
	}
	`, resourceLabel, name)
}
