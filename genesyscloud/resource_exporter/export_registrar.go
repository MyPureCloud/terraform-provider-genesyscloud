package resource_exporter
import (
	//gcloud "terraform-provider-genesyscloud/genesyscloud"
)

// // multipel interfaces
// type Registrar interface {
// 	RegisterResource(resourceName string, resource *schema.Resource)
// 	RegisterDataSource(dataSourceName string, datasource *schema.Resource)
// }

//var resourceExporters map[string]*resource_exporter.ResourceExporter

 


func GetResources() (map[string]*ResourceExporter) {
	return resourceExporters
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