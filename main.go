package main

import (
	"flag"
	"log"

	"github.com/hashicorp/terraform-plugin-go/tfprotov6/tf6server"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	providerRegistrar "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider_registrar"
)

// Run "go generate" to format example terraform files and generate the docs for the registry/website
//
// If you do not have terraform installed, you can remove the formatting command, but it's suggested to
// ensure the documentation is formatted properly.
//go:generate terraform fmt -recursive ./examples/

// Run the docs generation tool, check its repository for more information on how it works and how docs
// can be customized.
//go:generate go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs
//go:generate go run github.com/mypurecloud/terraform-provider-genesyscloud/apidocs

var (
	// these will be set by goreleaser for released binaries;
	// "0.1.0" is the default for local development
	version string = "0.1.0"
)

func main() {
	var debugMode bool

	flag.BoolVar(&debugMode, "debug", false, "run with managed debugger support")
	flag.Parse()

	providerResources, providerDataSources := providerRegistrar.GetProviderResources()

	// Build a Protocol v6 factory function (SDKv2 upgraded → v6 + PF v6, muxed)
	muxFactoryFuncFunc := provider.New(version, providerResources, providerDataSources)
	muxFactoryFunc, err := muxFactoryFuncFunc()
	if err != nil {
		log.Fatalf("Failed to create muxed provider factory: %v", err)
	}

	// Serve using Protocol v6 server
	var serveOpts []tf6server.ServeOpt
	if debugMode {
		serveOpts = append(serveOpts, tf6server.WithManagedDebug())
		log.Println("⚡ Debug mode enabled (tf6server.WithManagedDebug).")
	}

	// Use your dev address here; adjust as needed for releases.
	const providerAddr = "genesys.com/mypurecloud/genesyscloud"

	if err := tf6server.Serve(providerAddr, muxFactoryFunc, serveOpts...); err != nil {
		log.Fatalf("Provider serve failed: %v", err)
	}
}
