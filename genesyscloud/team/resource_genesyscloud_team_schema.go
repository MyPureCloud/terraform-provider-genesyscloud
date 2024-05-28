package team

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "terraform-provider-genesyscloud/genesyscloud/resource_register"
)

/*
resource_genesycloud_team_schema.go holds four functions within it:

1.  The registration code that registers the Datasource, Resource and Exporter for the package.
2.  The resource schema definitions for the team resource.
3.  The datasource schema definitions for the team datasource.
4.  The resource exporter configuration for the team exporter.
*/
const resourceName = "genesyscloud_team"

// SetRegistrar registers all of the resources, datasources and exporters in the package
func SetRegistrar(regInstance registrar.Registrar) {
	regInstance.RegisterResource(resourceName, ResourceTeam())
	regInstance.RegisterDataSource(resourceName, DataSourceTeam())
	regInstance.RegisterExporter(resourceName, TeamExporter())
}

// ResourceTeam registers the genesyscloud_team resource with Terraform
func ResourceTeam() *schema.Resource {
	return &schema.Resource{
		Description: `Genesys Cloud team`,

		CreateContext: provider.CreateWithPooledClient(createTeam),
		ReadContext:   provider.ReadWithPooledClient(readTeam),
		UpdateContext: provider.UpdateWithPooledClient(updateTeam),
		DeleteContext: provider.DeleteWithPooledClient(deleteTeam),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "The team name",
				Required:    true,
				Type:        schema.TypeString,
			},
			"division_id": {
				Description: "The division to which this entity belongs.",
				Required:    true,
				Type:        schema.TypeString,
			},
			"description": {
				Description: "Team information.",
				Optional:    true,
				Type:        schema.TypeString,
			},
			`member_ids`: {
				Description: "IDs of members assigned to the team. If not set, this resource will not manage group members.",
				Optional:    true,
				Computed:    true,
				Type:        schema.TypeSet,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

// TeamExporter returns the resourceExporter object used to hold the genesyscloud_team exporter's config
func TeamExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllAuthTeams),
		RefAttrs: map[string]*resourceExporter.RefAttrSettings{
			"division_id": {RefType: "genesyscloud_auth_division"},
			"member_ids":  {RefType: "genesyscloud_user"},
		},
	}
}

// DataSourceTeam registers the genesyscloud_team data source
func DataSourceTeam() *schema.Resource {
	return &schema.Resource{
		Description: `Genesys Cloud team data source. Select an team by name`,
		ReadContext: provider.ReadWithPooledClient(dataSourceTeamRead),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Description: `team name`,
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}
