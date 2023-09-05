package resource_register

import (
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type Registrar interface {
	RegisterResource(resourceName string, resource *schema.Resource)
	RegisterDataSource(dataSourceName string, datasource *schema.Resource)
	RegisterExporter(exporterName string, resourceExporter *resourceExporter.ResourceExporter)
}

// need this for TFexport where Resources are required for provider initialisation.
// NewGenesysCloudResourceExporter

var providerResources map[string]*schema.Resource
var providerDataSources map[string]*schema.Resource

func SetResources(resources map[string]*schema.Resource, dataSources map[string]*schema.Resource) {
	providerResources = resources
	providerDataSources = dataSources
}

func GetResources() (map[string]*schema.Resource, map[string]*schema.Resource) {
	return providerResources, providerDataSources
}
