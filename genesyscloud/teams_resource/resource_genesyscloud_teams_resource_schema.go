package teams_resource

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	gcloud "terraform-provider-genesyscloud/genesyscloud"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "terraform-provider-genesyscloud/genesyscloud/resource_register"
)

/*
resource_genesycloud_teams_resource_schema.go holds four functions within it:

1.  The registration code that registers the Datasource, Resource and Exporter for the package.
2.  The resource schema definitions for the teams_resource resource.
3.  The datasource schema definitions for the teams_resource datasource.
4.  The resource exporter configuration for the teams_resource exporter.
*/
const resourceName = "genesyscloud_teams_resource"

// SetRegistrar registers all of the resources, datasources and exporters in the package
func SetRegistrar(regInstance registrar.Registrar) {
	regInstance.RegisterResource(resourceName, ResourceTeamsResource())
	regInstance.RegisterDataSource(resourceName, DataSourceTeamsResource())
	regInstance.RegisterExporter(resourceName, TeamsResourceExporter())
}

// ResourceTeamsResource registers the genesyscloud_teams_resource resource with Terraform
func ResourceTeamsResource() *schema.Resource {
	return &schema.Resource{
		Description: `Genesys Cloud teams resource`,

		CreateContext: gcloud.CreateWithPooledClient(createTeamsResource),
		ReadContext:   gcloud.ReadWithPooledClient(readTeamsResource),
		UpdateContext: gcloud.UpdateWithPooledClient(updateTeamsResource),
		DeleteContext: gcloud.DeleteWithPooledClient(deleteTeamsResource),
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
				Optional:    true,
				Type:        schema.TypeString,
			},
			"description": {
				Description: "Team information.",
				Optional:    true,
				Type:        schema.TypeString,
			},
			"member_count": {
				Description: "Number of members in a team",
				Optional:    true,
				Type:        schema.TypeInt,
			},
		},
	}
}

// TeamsResourceExporter returns the resourceExporter object used to hold the genesyscloud_teams_resource exporter's config
func TeamsResourceExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: gcloud.GetAllWithPooledClient(getAllAuthTeamsResources),
		RefAttrs:         map[string]*resourceExporter.RefAttrSettings{
			// TODO: Add any reference attributes here
		},
	}
}

// DataSourceTeamsResource registers the genesyscloud_teams_resource data source
func DataSourceTeamsResource() *schema.Resource {
	return &schema.Resource{
		Description: `Genesys Cloud teams resource data source. Select an teams resource by name`,
		ReadContext: gcloud.ReadWithPooledClient(dataSourceTeamsResourceRead),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Description: `teams resource name`,
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}
