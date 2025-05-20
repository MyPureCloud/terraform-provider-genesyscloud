package main

import (
	"flag"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	providerRegistrar "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider_registrar"

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
//go:generate go run github.com/mypurecloud/terraform-provider-genesyscloud/apidocs
var (
	// these will be set by the goreleaser configuration
	// to appropriate values for the compiled binary
	// 0.1.0 is the default version for developing locally
	version string = "0.1.0"

	// goreleaser can also pass the specific commit if you want
	// commit  string = ""
)

func main() {
	var debugMode bool

	flag.BoolVar(&debugMode, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	providerResources, providerDataSources := providerRegistrar.GetProviderResources()

	opts := &plugin.ServeOpts{ProviderFunc: provider.New(version, providerResources, providerDataSources)}

	if debugMode {
		opts.Debug = true
		opts.ProviderAddr = "genesys.com/mypurecloud/genesyscloud"
	}
	plugin.Serve(opts)
}
