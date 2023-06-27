package main

import (
	"flag"
	"sync"
	"log"

	gcloud "terraform-provider-genesyscloud/genesyscloud"
	
	obs "terraform-provider-genesyscloud/genesyscloud/outbound_ruleset"
	pat "terraform-provider-genesyscloud/genesyscloud/process_automation_trigger"
	tfexp "terraform-provider-genesyscloud/genesyscloud/tfexporter"
	registrar "terraform-provider-genesyscloud/genesyscloud/Registrar"
	ob "terraform-provider-genesyscloud/genesyscloud/outbound"
	resource_exporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	ob_attempt_limit "terraform-provider-genesyscloud/genesyscloud/outbound_attempt_limit"
	ob_contact_list "terraform-provider-genesyscloud/genesyscloud/outbound_contact_list"

	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Run "go generate" to format example terraform files and generate the docs for the registry/website

// If you do not have terraform installed, you can remove the formatting command, but its suggested to
// ensure the documentation is formatted properly.
//go:generate terraform fmt -recursive ./examples/

// Run the docs generation tool, check its repository for more information on how it works and how docs
// can be customized.
//
//go:generate go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs
//go:generate git restore docs/index.md
//go:generate go run terraform-provider-genesyscloud/apidocs
var (
	// these will be set by the goreleaser configuration
	// to appropriate values for the compiled binary
	// 0.1.0 is the default version for developing locally
	version string = "0.1.0"

	// goreleaser can also pass the specific commit if you want
	// commit  string = ""
)

// var resourceMapMutex = sync.RWMutex{}
// var datasourceMapMutex = sync.RWMutex{}

var providerResources map[string]*schema.Resource
var providerDataSources map[string]*schema.Resource
var resourceExporters map[string]*resource_exporter.ResourceExporter

func main() {
	var debugMode bool

	//flag.BoolVar(&debugMode, "debug,false, "set to true to run the provider with support for debuggers like delve")

	//flag.BoolVar(&debugMode, "debug,false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	

	providerResources = make(map[string]*schema.Resource)
	providerDataSources = make(map[string]*schema.Resource)
	resourceExporters  = make(map[string]*resource_exporter.ResourceExporter)
	

	registerResources()
	//GetResourceExporters()

	opts := &plugin.ServeOpts{ProviderFunc: gcloud.New(version, providerResources, providerDataSources )}

	log.Printf("resource registration sss")
	tfexp.GetRegistrarresources()
	//gcloud.SetRegisterExporter(resourceExporters)


	if debugMode {
		opts.Debug = true
		opts.ProviderAddr = "registry.terraform.io/mypurecloud/genesyscloud"
	}
	plugin.Serve(opts)
}

//for practical use cases

// add mutex to struct
// a pointer. insiatitate as pointer and pass them
type RegisterInstance struct{
	resourceMapMutex sync.RWMutex
	datasourceMapMutex sync.RWMutex
	exporterMapMutex sync.RWMutex
}

// func (RegisterInstance) RegisterResource(resourceName string, resource *schema.Resource) {
//      RegisterResource(resourceName, resource)
// }

// func (RegisterInstance) RegisterDataSource(dataSourceName string, datasource *schema.Resource) {
// 	RegisterDataSource(dataSourceName, datasource)
// }

// func RegisterExporter(exporterName string, resourceExporter *resource_exporter.ResourceExporter) {
// 	//resourceMapMutex.Lock()
// 	resourceExporters[exporterName] = resourceExporter
// 	//resourceMapMutex.Unlock()
// }

func registerResources() (map[string]*schema.Resource, map[string]*schema.Resource) {
	log.Println(providerResources)
	log.Println(providerDataSources)

	i := &RegisterInstance{}

	pat.SetRegistrar(i)
	obs.SetRegistrar(i)
	tfexp.SetRegistrar(i)
	ob.SetRegistrar(i)
	gcloud.SetRegistrar(i)
	ob_attempt_limit.SetRegistrar(i)
	ob_contact_list.SetRegistrar(i)

	// registerGcloudResources()
	// registerGcloudSources()
	

	log.Println(providerResources)
	log.Println(providerDataSources)

	registrar.SetResources(providerResources, providerDataSources)
	return providerResources,providerDataSources
}




func (r *RegisterInstance) RegisterResource(resourceName string, resource *schema.Resource) {
	r.resourceMapMutex.Lock()
	providerResources[resourceName] = resource
	r.resourceMapMutex.Unlock()
}

func (r *RegisterInstance) RegisterDataSource(dataSourceName string, datasource *schema.Resource) {
	r.datasourceMapMutex.Lock()
	providerDataSources[dataSourceName] = datasource
	r.datasourceMapMutex.Unlock()
}

func (r *RegisterInstance) RegisterExporter(exporterName string, resourceExporter *resource_exporter.ResourceExporter) {
	r.exporterMapMutex.Lock()
	resourceExporters[exporterName] = resourceExporter
	r.exporterMapMutex.Unlock()
}

// func registerGcloudResources() {
// 	log.Printf("resource registration started")
// 	i := &RegisterInstance{}
// 	i.RegisterResource("genesyscloud_architect_datatable", gcloud.ResourceArchitectDatatable())
// }

// func registerGcloudSources() {
// 	i := &RegisterInstance{}
// 	i.RegisterDataSource("genesyscloud_architect_datatable", gcloud.DataSourceArchitectDatatable())
// }






// for Testing

// type TestRegisterInstance struct{}

// func (TestRegisterInstance) RegisterResourcesAndDataSources()  (map[string]*schema.Resource, map[string]*schema.Resource) {
// 	RegisterResourcesAndDataSources()  (map[string]*schema.Resource, map[string]*schema.Resource)
// }


// func GetResourceExporters() map[string]*resource_exporter.ResourceExporter {

	


	
// 	//RegisterExporter("genesyscloud_processautomation_trigger", pat.ProcessAutomationTriggerExporter())
// 	//RegisterExporter("genesyscloud_outbound_ruleset", obs.OutboundRulesetExporter())

// 	// l.RegisterDataSource("genesyscloud_telephony_providers_edges_site",  gcloud.DataSourceSite())
// 	// l.RegisterResource("genesyscloud_telephony_providers_edges_site",  gcloud.ResourceSite())
// 	// l.RegisterDataSource("genesyscloud_routing_wrapupcode",  gcloud.DataSourceRoutingWrapupcode())
//     // l.RegisterResource("genesyscloud_routing_wrapupcode",  gcloud.ResourceRoutingWrapupCode())
// 	// l.RegisterDataSource("genesyscloud_routing_queue",  gcloud.DataSourceRoutingQueue())
// 	// l.RegisterResource("genesyscloud_routing_queue",  gcloud.ResourceRoutingQueue())
// 	// l.RegisterResource("genesyscloud_flow",  gcloud.ResourceFlow())
// 	// l.RegisterDataSource("genesyscloud_flow",  gcloud.DataSourceFlow())
// 	// l.RegisterDataSource("genesyscloud_location",  gcloud.DataSourceLocation())
// 	// l.RegisterResource("genesyscloud_location",  gcloud.ResourceLocation())
// 	// l.RegisterDataSource("genesyscloud_auth_division_home",  gcloud.DataSourceAuthDivisionHome())
	
// 	return resourceExporters
// }
