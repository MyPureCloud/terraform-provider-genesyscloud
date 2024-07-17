package routing_language

import (
	"fmt"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "terraform-provider-genesyscloud/genesyscloud/resource_register"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const resourceName = "genesyscloud_routing_language"

// SetRegistrar registers all of the resources, datasources and exporters in the package
func SetRegistrar(regInstance registrar.Registrar) {
	regInstance.RegisterResource(resourceName, ResourceRoutingLanguage())
	regInstance.RegisterExporter(resourceName, RoutingLanguageExporter())
	regInstance.RegisterDataSource(resourceName, DataSourceRoutingLanguage())
}

func ResourceRoutingLanguage() *schema.Resource {
	return &schema.Resource{
		Description: "Genesys Cloud Routing Language",

		CreateContext: provider.CreateWithPooledClient(createRoutingLanguage),
		ReadContext:   provider.ReadWithPooledClient(readRoutingLanguage),
		DeleteContext: provider.DeleteWithPooledClient(deleteRoutingLanguage),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Language name. Changing the language_name attribute will cause the language object to be dropped and recreated with a new ID.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
		},
	}
}

func DataSourceRoutingLanguage() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for Genesys Cloud Routing Languages. Select a language by name.",
		ReadContext: provider.ReadWithPooledClient(dataSourceRoutingLanguageRead),
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Language name.",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}

func RoutingLanguageExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllRoutingLanguages),
	}
}

func GenerateRoutingLanguageResource(
	resourceID string,
	name string) string {
	return fmt.Sprintf(`resource "genesyscloud_routing_language" "%s" {
		name = "%s"
	}
	`, resourceID, name)
}
