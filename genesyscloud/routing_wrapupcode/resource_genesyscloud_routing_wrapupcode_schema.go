package routing_wrapupcode

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
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
)

const ResourceType = "genesyscloud_routing_wrapupcode"

// SetRegistrar registers all of the resources, datasources and exporters in the package
func SetRegistrar(regInstance registrar.Registrar) {
	regInstance.RegisterFrameworkResource(ResourceType, NewRoutingWrapupcodeFrameworkResource)
	regInstance.RegisterFrameworkDataSource(ResourceType, NewRoutingWrapupcodeFrameworkDataSource)
	regInstance.RegisterExporter(ResourceType, RoutingWrapupcodeExporter())
}

// RoutingWrapupcodeResourceSchema returns the schema for the routing wrapupcode resource
func RoutingWrapupcodeResourceSchema() schema.Schema {
	return schema.Schema{
		Description: "Genesys Cloud Routing Wrapup Code",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The globally unique identifier for the wrapup code.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Wrapup Code name.",
				Required:    true,
			},
			"division_id": schema.StringAttribute{
				Description: "The division to which this routing wrapupcode will belong.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"description": schema.StringAttribute{
				Description: "The wrap-up code description.",
				Optional:    true,
			},
		},
	}
}

// RoutingWrapupcodeDataSourceSchema returns the schema for the routing wrapupcode data source
func RoutingWrapupcodeDataSourceSchema() datasourceschema.Schema {
	return datasourceschema.Schema{
		Description: "Data source for Genesys Cloud Wrap-up Code. Select a wrap-up code by name",
		Attributes: map[string]datasourceschema.Attribute{
			"id": datasourceschema.StringAttribute{
				Description: "The globally unique identifier for the wrapup code.",
				Computed:    true,
			},
			"name": datasourceschema.StringAttribute{
				Description: "Wrap-up code name.",
				Required:    true,
			},
		},
	}
}

func RoutingWrapupcodeExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(GetAllRoutingWrapupcodesSDK),
		RefAttrs: map[string]*resourceExporter.RefAttrSettings{
			"division_id": {RefType: "genesyscloud_auth_division"},
		},
	}
}

// GenerateRoutingWrapupcodeResource generates a routing wrapupcode resource for cross-package testing
// This function is used by other packages that need to create routing wrapupcode resources in their tests
func GenerateRoutingWrapupcodeResource(resourceLabel string, name string, divisionId string, description string) string {
	divisionIdAttr := ""
	if divisionId != util.NullValue {
		divisionIdAttr = fmt.Sprintf(`
		division_id = %s`, divisionId)
	}

	descriptionAttr := ""
	if description != "" {
		descriptionAttr = fmt.Sprintf(`
		description = "%s"`, description)
	}

	return fmt.Sprintf(`resource "genesyscloud_routing_wrapupcode" "%s" {
		name = "%s"%s%s
	}
	`, resourceLabel, name, divisionIdAttr, descriptionAttr)
}
