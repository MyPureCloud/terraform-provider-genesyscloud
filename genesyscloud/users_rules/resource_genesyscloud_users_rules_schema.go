package users_rules

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_register"
)

/*
ResourceName is defined in this file along with four functions:

1.  The registration code that registers the Datasource, Resource and Exporter for the package.
2.  The resource schema definitions for the users_rules resource.
3.  The datasource schema definitions for the users_rules datasource.
4.  The resource exporter configuration for the users_rules exporter.
*/
const ResourceName = "genesyscloud_users_rules"
const ResourceType = ResourceName

// SetRegistrar registers all the resources, datasources and exporters in the package
func SetRegistrar(regInstance registrar.Registrar) {
	regInstance.RegisterResource(ResourceName, ResourceUsersRules())
	regInstance.RegisterDataSource(ResourceName, DataSourceUsersRules())
	regInstance.RegisterExporter(ResourceName, UsersRulesExporter())
}

// ResourceUsersRules registers the genesyscloud_users_rules resource with Terraform
func ResourceUsersRules() *schema.Resource {
	userRulesValueResource := &schema.Resource{
		Schema: map[string]*schema.Schema{
			"context_id": {
				Description: `The contextId for this group`,
				Optional:    true,
				Type:        schema.TypeString,
			},
			"ids": {
				Description: `The ids to select for this group`,
				Required:    true,
				Type:        schema.TypeList,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
		},
	}

	userRulesGroupItemResource := &schema.Resource{
		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeString,
				Description: "The internal ID for this group",
				Optional:    true,
				Computed:    true,
			},
			"operator": {
				Type:         schema.TypeString,
				Description:  "The operator for this group",
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"And", "Not"}, false),
			},
			"container": {
				Type:        schema.TypeString,
				Description: "The container that the ids belong to",
				Required:    true,
				ValidateFunc: validation.StringInSlice([]string{
					"AcdSkill",
					"BusinessUnit",
					"DirectoryGroup",
					"Division",
					"Language",
					"Location",
					"ManagementUnit",
					"Queue",
					"ReportsTo",
					"StaffingGroup",
					"Team",
					"User",
				}, false),
			},
			"values": {
				Type:        schema.TypeList,
				Description: "The ids and contextIds to select for this group",
				Required:    true,
				Elem:        userRulesValueResource,
			},
		},
	}

	userRulesCriteriaResource := &schema.Resource{
		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeString,
				Description: "The internal ID for this criteria",
				Optional:    true,
				Computed:    true,
			},
			"operator": {
				Type:         schema.TypeString,
				Description:  "The operator for this criteria",
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"Or"}, false),
			},
			"group": {
				Type:        schema.TypeList,
				Description: "The user selection groups for this criteria",
				Required:    true,
				Elem:        userRulesGroupItemResource,
			},
		},
	}

	return &schema.Resource{
		Description: `Genesys Cloud users rules`,

		CreateContext: provider.CreateWithPooledClient(createUsersRules),
		ReadContext:   provider.ReadWithPooledClient(readUsersRules),
		UpdateContext: provider.UpdateWithPooledClient(updateUsersRules),
		DeleteContext: provider.DeleteWithPooledClient(deleteUsersRules),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,

		Schema: map[string]*schema.Schema{
			"name": {
				Description: "The name of the rule",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"description": {
				Description: "The description of the rule",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"type": {
				Description:  "The type of the rule",
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"Learning", "ActivityPlan"}, false),
			},
			"criteria": {
				Type:        schema.TypeList,
				Description: "The criteria for the rule",
				Optional:    true,
				Elem:        userRulesCriteriaResource,
			},
		},
	}
}

// UsersRulesExporter returns the resourceExporter object used to hold the genesyscloud_users_rules exporter's config
func UsersRulesExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllAuthUsersRules),
	}
}

// DataSourceUsersRules registers the genesyscloud_users_rules data source
func DataSourceUsersRules() *schema.Resource {
	return &schema.Resource{
		Description: `Data source for Genesys Cloud Users Rules. Select a Users Rule by name.`,
		ReadContext: provider.ReadWithPooledClient(dataSourceUsersRulesRead),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Description: `Users Rule name.`,
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}
