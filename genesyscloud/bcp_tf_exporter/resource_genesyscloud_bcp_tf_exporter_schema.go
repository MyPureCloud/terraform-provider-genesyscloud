package bcp_tf_exporter

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_register"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/validators"
)

const ResourceType = "genesyscloud_bcp_tf_exporter"

func SetRegistrar(l registrar.Registrar) {
	l.RegisterResource(ResourceType, ResourceBcpTfExporter())
}

func ResourceBcpTfExporter() *schema.Resource {
	return &schema.Resource{
		// INTERNAL TOOL ONLY - NOT FOR PUBLIC DOCUMENTATION
		//
		// This resource exports Genesys Cloud resources in a JSON format suitable for Business Continuity Planning (BCP).
		// The export includes resource IDs, names, and their dependencies to help understand resource relationships
		// and plan for disaster recovery scenarios.
		//
		// Usage Examples:
		//
		// Basic export (all resources):
		//   resource "genesyscloud_bcp_tf_exporter" "example" {
		//     directory = "./bcp_exports"
		//     filename  = "production_resources.json"
		//   }
		//
		// Export specific resource types:
		//   resource "genesyscloud_bcp_tf_exporter" "users_and_groups" {
		//     directory = "./bcp_exports"
		//     filename  = "users_groups.json"
		//     include_filter_resources = [
		//       "genesyscloud_user",
		//       "genesyscloud_group",
		//       "genesyscloud_auth_role"
		//     ]
		//   }
		//
		// Exclude specific resource types:
		//   resource "genesyscloud_bcp_tf_exporter" "no_flows" {
		//     directory = "./bcp_exports"
		//     filename  = "resources_no_flows.json"
		//     exclude_filter_resources = [
		//       "genesyscloud_flow",
		//       "genesyscloud_architect_datatable"
		//     ]
		//   }
		//
		// Output JSON Structure:
		// {
		//   "genesyscloud_user": [
		//     {
		//       "id": "12345678-1234-1234-1234-123456789012",
		//       "name": "john_doe",
		//       "dependencies": []
		//     }
		//   ],
		//   "genesyscloud_group": [
		//     {
		//       "id": "87654321-4321-4321-4321-210987654321",
		//       "name": "support_team",
		//       "dependencies": [
		//         "genesyscloud_user::12345678-1234-1234-1234-123456789012"
		//       ]
		//     }
		//   ]
		// }
		//
		// Dependency Format: "resource_type::resource_id"
		// - For flows: Uses dependent consumers API to get actual dependencies
		// - For other resources: Extracts dependencies from resource state using RefAttrs configuration
		//
		// Use Cases:
		// - Disaster recovery planning
		// - Resource dependency mapping
		// - Cross-org resource correlation
		// - Backup and restore planning
		Description: "Genesys Cloud BCP Resource to export resource IDs, names, and dependencies in JSON format.",

		CreateWithoutTimeout: createBcpTfExporter,
		ReadWithoutTimeout:   readBcpTfExporter,
		DeleteContext:        deleteBcpTfExporter,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"directory": {
				// Directory where the JSON export file will be written.
				// Defaults to "./genesyscloud" if not specified.
				// The directory will be created if it doesn't exist.
				Description: "Directory where the JSON export file will be written.",
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "./genesyscloud",
				ForceNew:    true,
			},
			"filename": {
				// Name of the JSON export file.
				// Defaults to "bcp_export.json" if not specified.
				// Should include .json extension for clarity.
				Description: "Name of the JSON export file.",
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "bcp_export.json",
				ForceNew:    true,
			},
			"log_permissions_filename": {
				// Name of the JSON export file for outputting permissions errors.
				// If not specified, no separate permissions file will be created.
				Description: "Name of the file to output permissions errors",
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
			},
			"include_filter_resources": {
				// Include only resources that match the specified resource types.
				// Use exact resource type names (e.g., "genesyscloud_user", "genesyscloud_group").
				// Cannot be used together with exclude_filter_resources.
				// If not specified, all available resource types will be exported.
				Description: "Include only resources that match either a resource type or a resource type::regular expression.",
				Type:        schema.TypeList,
				Optional:    true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validators.ValidateSubStringInSlice(resourceExporter.GetAvailableExporterTypes()),
				},
				ConflictsWith: []string{"exclude_filter_resources"},
				ForceNew:      true,
			},
			"exclude_filter_resources": {
				// Exclude resources that match the specified resource types.
				// Use exact resource type names (e.g., "genesyscloud_flow", "genesyscloud_architect_datatable").
				// Cannot be used together with include_filter_resources.
				// All other available resource types will be exported.
				Description: "Exclude resources that match either a resource type or a resource type::regular expression.",
				Type:        schema.TypeList,
				Optional:    true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validators.ValidateSubStringInSlice(resourceExporter.GetAvailableExporterTypes()),
				},
				ForceNew:      true,
				ConflictsWith: []string{"include_filter_resources"},
			},
		},
	}
}
