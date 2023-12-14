package members

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	gcloud "terraform-provider-genesyscloud/genesyscloud"
	//resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "terraform-provider-genesyscloud/genesyscloud/resource_register"
)

var (
	//required for read
	userReferenceWithnameResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			`member_id`: {
				Description: `Response text content.`,
				Optional:    true,
				Type:        schema.TypeString,
			},
			`member_name`: {
				Description: `Response text content type.`,
				Optional:    true,
				Type:        schema.TypeString,
			},
		},
	}
)
/*
resource_genesycloud_members_schema.go holds four functions within it:

1.  The registration code that registers the Datasource, Resource and Exporter for the package.
2.  The resource schema definitions for the members resource.
3.  The datasource schema definitions for the members datasource.
4.  The resource exporter configuration for the members exporter.
*/
const resourceName = "genesyscloud_members"

// SetRegistrar registers all of the resources, datasources and exporters in the package
func SetRegistrar(regInstance registrar.Registrar) {
	regInstance.RegisterResource(resourceName, ResourceMembers())
}

// ResourceMembers registers the genesyscloud_members resource with Terraform
func ResourceMembers() *schema.Resource {
	return &schema.Resource{
		Description: `Genesys Cloud members`,

		CreateContext: gcloud.CreateWithPooledClient(createMembers),
		ReadContext:   gcloud.ReadWithPooledClient(readMembers),
		DeleteContext: gcloud.DeleteWithPooledClient(deleteMembers),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			`team_id`: {
				Description: `Specifies the team id to which the members belong to.`,
				Required:    true,
				Type:        schema.TypeString,
			},
			`member_ids`: {
				Description: `Specifies the members`,
				Optional:    true,
				Type:        schema.TypeList,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			``
		},
	}
}
