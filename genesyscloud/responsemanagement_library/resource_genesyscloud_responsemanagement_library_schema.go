package responsemanagement_library

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "terraform-provider-genesyscloud/genesyscloud/resource_register"
)

/*
resource_genesycloud_responsemanagement_library_schema.go holds four functions within it:

1.  The registration code that registers the Datasource, Resource and Exporter for the package.
2.  The resource schema definitions for the responsemanagement_library resource.
3.  The datasource schema definitions for the responsemanagement_library datasource.
4.  The resource exporter configuration for the responsemanagement_library exporter.
*/
const resourceName = "genesyscloud_responsemanagement_library"

// SetRegistrar registers all of the resources, datasources and exporters in the package
func SetRegistrar(regInstance registrar.Registrar) {
	regInstance.RegisterResource(resourceName, ResourceResponsemanagementLibrary())
	regInstance.RegisterDataSource(resourceName, DataSourceResponsemanagementLibrary())
	regInstance.RegisterExporter(resourceName, ResponsemanagementLibraryExporter())
}

// ResourceResponsemanagementLibrary registers the genesyscloud_responsemanagement_library resource with Terraform
func ResourceResponsemanagementLibrary() *schema.Resource {
	return &schema.Resource{
		Description: `Genesys Cloud responsemanagement library`,

		CreateContext: provider.CreateWithPooledClient(createResponsemanagementLibrary),
		ReadContext:   provider.ReadWithPooledClient(readResponsemanagementLibrary),
		UpdateContext: provider.UpdateWithPooledClient(updateResponsemanagementLibrary),
		DeleteContext: provider.DeleteWithPooledClient(deleteResponsemanagementLibrary),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			`name`: {
				Description: `The library name.`,
				Required:    true,
				Type:        schema.TypeString,
			},
		},
	}
}

// ResponsemanagementLibraryExporter returns the resourceExporter object used to hold the genesyscloud_responsemanagement_library exporter's config
func ResponsemanagementLibraryExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllAuthResponsemanagementLibrarys),
	}
}

// DataSourceResponsemanagementLibrary registers the genesyscloud_responsemanagement_library data source
func DataSourceResponsemanagementLibrary() *schema.Resource {
	return &schema.Resource{
		Description: `Data source for Genesys Cloud Responsemanagement Library. Select a Responsemanagement Library by name.`,
		ReadContext: provider.ReadWithPooledClient(dataSourceResponsemanagementLibraryRead),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Description: `Responsemanagement Library name.`,
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}

func GenerateResponseManagementLibraryResource(
	resourceId string,
	name string) string {
	return fmt.Sprintf(`
		resource "genesyscloud_responsemanagement_library" "%s" {
			name = "%s"
		}
	`, resourceId, name)
}
