package routing_language

// @team: Assignment
// @chat: #genesys-cloud-acd-routing
// @description: Routing configuration service for queues, skills, wrapup codes, and utilization settings. Manages how contacts are distributed to agents based on skills, capacity, and routing rules across all interaction channels.

import (
	"fmt"

	datasourceschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_register"
)

const ResourceType = "genesyscloud_routing_language"

// SetRegistrar registers all of the resources, datasources and exporters in the package
func SetRegistrar(regInstance registrar.Registrar) {
	regInstance.RegisterFrameworkResource(ResourceType, NewFrameworkRoutingLanguageResource)
	regInstance.RegisterFrameworkDataSource(ResourceType, NewFrameworkRoutingLanguageDataSource)
	regInstance.RegisterExporter(ResourceType, RoutingLanguageExporter())
}

// RoutingLanguageResourceSchema returns the schema for the routing language resource
func RoutingLanguageResourceSchema() schema.Schema {
	return schema.Schema{
		Description: "Genesys Cloud Routing Language",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of the routing language.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Language name. Changing the language_name attribute will cause the language object to be dropped and recreated with a new ID.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

// RoutingLanguageDataSourceSchema returns the schema for the routing language data source
func RoutingLanguageDataSourceSchema() datasourceschema.Schema {
	return datasourceschema.Schema{
		Description: "Data source for Genesys Cloud Routing Languages. Select a language by name.",
		Attributes: map[string]datasourceschema.Attribute{
			"id": datasourceschema.StringAttribute{
				Description: "The ID of the routing language.",
				Computed:    true,
			},
			"name": datasourceschema.StringAttribute{
				Description: "Language name.",
				Required:    true,
			},
		},
	}
}

// SDKv2 resource and data source functions removed - Framework-only migration

func RoutingLanguageExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(GetAllRoutingLanguagesSDK),
	}
}

func GenerateRoutingLanguageResource(
	resourceLabel string,
	name string) string {
	return fmt.Sprintf(`resource "genesyscloud_routing_language" "%s" {
		name = "%s"
	}
	`, resourceLabel, name)
}
