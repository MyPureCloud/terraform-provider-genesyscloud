package Registrar
import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	resource_exporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
)

// multipel interfaces
type Registrar interface {
	RegisterResource(resourceName string, resource *schema.Resource)
	RegisterDataSource(dataSourceName string, datasource *schema.Resource)
	RegisterExporter(exporterName string, resourceExporter *resource_exporter.ResourceExporter) 
}

type TestRegistrar interface {
	RegisterResourcesAndDataSources()  (map[string]*schema.Resource, map[string]*schema.Resource)
} 

var providerResources map[string]*schema.Resource
var providerDataSources map[string]*schema.Resource

 

func SetResources(resources map[string]*schema.Resource, dataSources map[string]*schema.Resource) {
	providerResources = resources
	providerDataSources = dataSources
}

func GetResources() (map[string]*schema.Resource, map[string]*schema.Resource) {
	return providerResources, providerDataSources
}





// type TestRegistrar interface {
// 	SetTestRegistrar() 
// } 


// package resource_registrar
// import (
// 	_ "terraform-provider-genesyscloud/genesyscloud/tfexporter"
	
// )

// var providerResources map[string]*schema.Resource
// var providerDataSources map[string]*schema.Resource

// func SetResources(resources map[string]*schema.Resource, dataSources map[string]*schema.Resource) {
// 	providerResources := resources
// 	providerDataSources := dataSources
// }

// func GetResources() map[string]*schema.Resource, map[string]*schema.Resource {
// 	retrun providerResources, providerDataSources
// }