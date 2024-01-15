package team

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	gcloud "terraform-provider-genesyscloud/genesyscloud"
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

		CreateContext: gcloud.CreateWithPooledClient(createTeam),
		ReadContext:   gcloud.ReadWithPooledClient(readTeam),
		UpdateContext: gcloud.UpdateWithPooledClient(updateTeam),
		DeleteContext: gcloud.DeleteWithPooledClient(deleteTeam),
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
				Description: `Specifies the members, No modifications to members will be made if not set. If empty all members will be deleted. If populated, only the populated members will be retained`,
				Optional:    true,
				Type:        schema.TypeList,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

// TeamExporter returns the resourceExporter object used to hold the genesyscloud_team exporter's config
func TeamExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: gcloud.GetAllWithPooledClient(getAllAuthTeams),
		RefAttrs: map[string]*resourceExporter.RefAttrSettings{
			"division_id": {RefType: "genesyscloud_auth_division"},
		},
	}
}

// DataSourceTeam registers the genesyscloud_team data source
func DataSourceTeam() *schema.Resource {
	return &schema.Resource{
		Description: `Genesys Cloud team data source. Select an team by name`,
		ReadContext: gcloud.ReadWithPooledClient(dataSourceTeamRead),
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
