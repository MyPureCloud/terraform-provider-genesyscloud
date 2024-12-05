package auth_division

import (
	"terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "terraform-provider-genesyscloud/genesyscloud/resource_register"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const ResourceType = "genesyscloud_auth_division"

// SetRegistrar registers all of the resources, datasources and exporters in the package
func SetRegistrar(regInstance registrar.Registrar) {
	regInstance.RegisterResource(ResourceType, ResourceAuthDivision())
	regInstance.RegisterDataSource(ResourceType, DataSourceAuthDivision())
	regInstance.RegisterExporter(ResourceType, AuthDivisionExporter())
}

func ResourceAuthDivision() *schema.Resource {
	return &schema.Resource{
		Description: "Genesys Cloud Authorization Division",

		CreateContext: provider.CreateWithPooledClient(createAuthDivision),
		ReadContext:   provider.ReadWithPooledClient(readAuthDivision),
		UpdateContext: provider.UpdateWithPooledClient(updateAuthDivision),
		DeleteContext: provider.DeleteWithPooledClient(deleteAuthDivision),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Division name.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"description": {
				Description: "Division description.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"home": {
				Description: "True if this is the home division. This can be set to manage the pre-existing home division.  Note: If name attribute is changed, this will cause the auth_division to be dropped and recreated. This will generate a new ID the division.  Existing objects with the old division will not be migrated to the new division",
				Type:        schema.TypeBool,
				Optional:    true,
				ForceNew:    true,
			},
		},
	}
}

func DataSourceAuthDivision() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for Genesys Cloud Divisions. Select a division by name.",
		ReadContext: provider.ReadWithPooledClient(dataSourceAuthDivisionRead),
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Division name.",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}

func AuthDivisionExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllAuthDivisions),
		RefAttrs:         map[string]*resourceExporter.RefAttrSettings{}, // No references
	}
}
