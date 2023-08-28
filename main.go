package main

import (
	"flag"
	"sync"

	gcloud "terraform-provider-genesyscloud/genesyscloud"
	externalContacts "terraform-provider-genesyscloud/genesyscloud/external_contacts"
	ob "terraform-provider-genesyscloud/genesyscloud/outbound"
	obAttemptLimit "terraform-provider-genesyscloud/genesyscloud/outbound_attempt_limit"
	obContactList "terraform-provider-genesyscloud/genesyscloud/outbound_contact_list"
	obRuleset "terraform-provider-genesyscloud/genesyscloud/outbound_ruleset"
	pat "terraform-provider-genesyscloud/genesyscloud/process_automation_trigger"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "terraform-provider-genesyscloud/genesyscloud/resource_register"
	"terraform-provider-genesyscloud/genesyscloud/scripts"
	simpleRoutingQueue "terraform-provider-genesyscloud/genesyscloud/simple_routing_queue"
	tfexp "terraform-provider-genesyscloud/genesyscloud/tfexporter"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
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

var providerResources map[string]*schema.Resource
var providerDataSources map[string]*schema.Resource
var resourceExporters map[string]*resourceExporter.ResourceExporter

func main() {
	var debugMode bool

	flag.BoolVar(&debugMode, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	providerResources = make(map[string]*schema.Resource)
	providerDataSources = make(map[string]*schema.Resource)
	resourceExporters = make(map[string]*resourceExporter.ResourceExporter)

	registerResources()

	opts := &plugin.ServeOpts{ProviderFunc: gcloud.New(version, providerResources, providerDataSources)}

	if debugMode {
		opts.Debug = true
		opts.ProviderAddr = "registry.terraform.io/mypurecloud/genesyscloud"
	}
	plugin.Serve(opts)
}

type RegisterInstance struct {
	resourceMapMutex   sync.RWMutex
	datasourceMapMutex sync.RWMutex
	exporterMapMutex   sync.RWMutex
}

func registerResources() {

	reg_instance := &RegisterInstance{}

	pat.SetRegistrar(reg_instance)

	ob.SetRegistrar(reg_instance)
	gcloud.SetRegistrar(reg_instance)
	obAttemptLimit.SetRegistrar(reg_instance)
	obContactList.SetRegistrar(reg_instance)
	obRuleset.SetRegistrar(reg_instance)
	scripts.SetRegistrar(reg_instance)
	externalContacts.SetRegistrar(reg_instance)
	simpleRoutingQueue.SetRegistrar(reg_instance)
	resourceExporter.SetRegisterExporter(resourceExporters)

	// setting resources for Use cases  like TF export where provider is used in resource classes.
	//tfexp.GetRegistrarresources()
	tfexp.SetRegistrar(reg_instance)
	registrar.SetResources(providerResources, providerDataSources)

}

func (r *RegisterInstance) RegisterResource(resourceName string, resource *schema.Resource) {
	r.resourceMapMutex.Lock()
	defer r.resourceMapMutex.Unlock()
	providerResources[resourceName] = resource
}

func (r *RegisterInstance) RegisterDataSource(dataSourceName string, datasource *schema.Resource) {
	r.datasourceMapMutex.Lock()
	defer r.datasourceMapMutex.Unlock()
	providerDataSources[dataSourceName] = datasource
}

func (r *RegisterInstance) RegisterExporter(exporterName string, resourceExporter *resourceExporter.ResourceExporter) {
	r.exporterMapMutex.Lock()
	defer r.exporterMapMutex.Unlock()
	resourceExporters[exporterName] = resourceExporter
}
