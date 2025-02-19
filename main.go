package main

import (
	"context"
	"flag"
	"log"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	providerRegistrar "terraform-provider-genesyscloud/genesyscloud/provider_registrar"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6/tf6server"
	"github.com/hashicorp/terraform-plugin-mux/tf5to6server"
	"github.com/hashicorp/terraform-plugin-mux/tf6muxserver"
)

// Run "go generate" to format example terraform files and generate the docs for the registry/website

// If you do not have terraform installed, you can remove the formatting command, but its suggested to
// ensure the documentation is formatted properly.
//go:generate terraform fmt -recursive ./examples/

// Run the docs generation tool, check its repository for more information on how it works and how docs
// can be customized.
//
//go:generate go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs
//go:generate go run terraform-provider-genesyscloud/apidocs
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
	var ctx = context.Background()

	flag.BoolVar(&debugMode, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	providerResources, providerDataSources := providerRegistrar.GetProviderResources()

	upgradedSdkProvider, err := tf5to6server.UpgradeServer(ctx, provider.New(version, providerResources, providerDataSources)().GRPCProvider)
	if err != nil {
		log.Fatal(err)
	}

	frameworkProvider := providerserver.NewProtocol6(provider.NewFrameWorkProvider(version)())

	bothServers := []func() tfprotov6.ProviderServer{
		frameworkProvider,
		func() tfprotov6.ProviderServer { return upgradedSdkProvider },
	}

	muxServer, err := tf6muxserver.NewMuxServer(ctx, bothServers...)
	if err != nil {
		log.Fatalf("failed to create mux server: %s", err.Error())
	}

	var serveOpts []tf6server.ServeOpt

	if debugMode {
		serveOpts = append(serveOpts, tf6server.WithManagedDebug())
	}

	err = tf6server.Serve(
		"registry.terraform.io/mypurecloud/genesyscloud",
		muxServer.ProviderServer,
		serveOpts...,
	)
	if err != nil {
		log.Fatal(err)
	}

	//opts := &plugin.ServeOpts{ProviderFunc: provider.New(version, providerResources, providerDataSources)}
	//
	//if debugMode {
	//	opts.Debug = true
	//	opts.ProviderAddr = "genesys.com/mypurecloud/genesyscloud"
	//}
	//plugin.Serve(opts)
}
